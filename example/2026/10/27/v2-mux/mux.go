package main

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

const (
	// dialTimeout 连接隧道服务端与连接目标地址的超时。
	dialTimeout = 5 * time.Second
	// openTimeout 是发出 SYN 后等待对端 SYNACK 的超时。
	openTimeout = 5 * time.Second
	// streamBacklog 是每条流接收队列的长度，也是背压的闸门（见 Stream.deliver）。
	streamBacklog = 16
	// acceptBacklog 是对端主动开流时的排队长度，v4 反向转发会用它。
	acceptBacklog = 64
)

var (
	// ErrSessionClosed 表示隧道连接已断开。
	ErrSessionClosed = errors.New("minibridge: session closed")
	// ErrStreamReset 表示对端拒绝了这条流（目标不可达等）。
	ErrStreamReset = errors.New("minibridge: stream reset by peer")

	errDeadlineUnsupported = errors.New("minibridge: stream deadlines not supported")
)

// Session 是一条隧道 TCP 连接上的多路复用会话，两端各持有一个。
//
// 它的能力是**对称**的：两端都能 Open（主动开流）和 Accept（接受对端开的流）。
// v2 只有客户端 Open、服务端 Accept；到 v4 做反向转发时，服务端也会 Open，
// 代码不用改结构——这就是现在把对称性做全的原因。
type Session struct {
	conn    net.Conn
	writeMu sync.Mutex // 保证整帧原子写入

	mu      sync.Mutex
	streams map[uint32]*Stream
	nextID  uint32
	closed  bool

	acceptCh  chan *Stream
	doneCh    chan struct{}
	closeOnce sync.Once
	closeErr  error
}

// newSession 在一条已建立的连接上起会话。isClient 决定本端用奇数还是偶数 streamID。
func newSession(conn net.Conn, isClient bool) *Session {
	s := &Session{
		conn:     conn,
		streams:  make(map[uint32]*Stream),
		acceptCh: make(chan *Stream, acceptBacklog),
		doneCh:   make(chan struct{}),
	}
	// 客户端用奇数、服务端用偶数：双方同时 Open 也不会撞号（HTTP/2 同规则）。
	if isClient {
		s.nextID = 1
	} else {
		s.nextID = 2
	}
	go s.readLoop()
	return s
}

// Open 主动开一条流，等价于 v1 的一次握手：告知对端目标地址，等它回应。
func (s *Session) Open(target string) (*Stream, error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, ErrSessionClosed
	}
	st := s.newStreamLocked(s.nextID, target)
	s.nextID += 2
	s.mu.Unlock()

	if err := s.writeFrame(frame{typ: frameSYN, streamID: st.id, payload: []byte(target)}); err != nil {
		st.closeWith(err)
		return nil, err
	}

	select {
	case err := <-st.openCh:
		if err != nil {
			st.closeWith(err)
			return nil, err
		}
		return st, nil
	case <-s.doneCh:
		st.closeWith(ErrSessionClosed)
		return nil, ErrSessionClosed
	case <-time.After(openTimeout):
		err := errors.New("minibridge: open stream timeout")
		st.closeWith(err)
		return nil, err
	}
}

// Accept 等一条对端主动开出来的流。
func (s *Session) Accept() (*Stream, error) {
	select {
	case st := <-s.acceptCh:
		return st, nil
	case <-s.doneCh:
		return nil, ErrSessionClosed
	}
}

// Close 关闭会话，并把所有还活着的流一起收尾——否则它们的上层连接会集体挂死。
func (s *Session) Close(cause error) error {
	s.closeOnce.Do(func() {
		s.mu.Lock()
		s.closed = true
		s.closeErr = cause
		streams := make([]*Stream, 0, len(s.streams))
		for _, st := range s.streams {
			streams = append(streams, st)
		}
		s.streams = make(map[uint32]*Stream)
		s.mu.Unlock()

		close(s.doneCh)
		for _, st := range streams {
			st.closeWith(ErrSessionClosed)
		}
		_ = s.conn.Close()
	})

	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closeErr
}

// writeFrame 写一整帧，调用方不需要自己加锁。
func (s *Session) writeFrame(f frame) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return writeFrame(s.conn, f)
}

// readLoop 是会话唯一的读 goroutine：读帧 → 按 streamID 分发。
func (s *Session) readLoop() {
	var err error
	for {
		f, e := readFrame(s.conn)
		if e != nil {
			err = e
			break
		}
		s.handleFrame(f)
	}
	if err == io.EOF {
		err = nil // 对端正常关闭不是错误
	}
	_ = s.Close(err)
}

func (s *Session) handleFrame(f frame) {
	s.mu.Lock()
	st := s.streams[f.streamID]
	s.mu.Unlock()

	switch f.typ {
	case frameSYN:
		s.handleSYN(f)
	case frameSYNACK:
		if st != nil {
			st.signalOpen(nil)
		}
	case frameRST:
		if st != nil {
			st.signalOpen(ErrStreamReset)
			st.closeRemote()
		}
	case frameFIN:
		if st != nil {
			st.closeRemote() // 对端发完了：我这边读到队列空后返回 EOF
		}
	case frameDATA:
		if st != nil {
			st.deliver(f.payload) // 注意：此时没有持有任何锁，deliver 允许阻塞
		}
	case framePING:
		_ = s.writeFrame(frame{typ: framePONG, streamID: f.streamID})
	}
}

// handleSYN 处理对端主动开流：登记后放进 acceptCh 交给上层。
func (s *Session) handleSYN(f frame) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	if _, exists := s.streams[f.streamID]; exists {
		s.mu.Unlock()
		return
	}
	st := s.newStreamLocked(f.streamID, string(f.payload))
	s.mu.Unlock()

	select {
	case s.acceptCh <- st:
	case <-s.doneCh:
	}
}

// newStreamLocked 新建流并登记，调用方必须持有 s.mu。
func (s *Session) newStreamLocked(id uint32, target string) *Stream {
	st := &Stream{
		id:     id,
		target: target,
		sess:   s,
		recv:   make(chan []byte, streamBacklog),
		openCh: make(chan error, 1),
		finCh:  make(chan struct{}),
		doneCh: make(chan struct{}),
	}
	s.streams[id] = st
	return st
}

func (s *Session) removeStream(id uint32) {
	s.mu.Lock()
	delete(s.streams, id)
	s.mu.Unlock()
}

// Stream 是会话上的一条流，实现 net.Conn——
// 这样 v1 那套 io.Copy 桥接代码可以原样复用，不必为多路复用重写。
type Stream struct {
	id     uint32
	target string
	sess   *Session

	recv   chan []byte
	openCh chan error
	finCh  chan struct{}
	doneCh chan struct{}

	finOnce  sync.Once // 发送 FIN（半关闭）
	recvOnce sync.Once // 收到对端 FIN
	doneOnce sync.Once // 本流彻底结束

	mu       sync.Mutex
	cur      []byte
	closeErr error
}

var _ net.Conn = (*Stream)(nil)

// Target 返回建流时约定的目标地址（从 SYN 的 payload 里来）。
func (st *Stream) Target() string { return st.target }

// ID 返回本流的 streamID。
func (st *Stream) ID() uint32 { return st.id }

// AcceptOK 通知对端"流已建立"，服务端 dial 目标成功后调用。
func (st *Stream) AcceptOK() error {
	return st.sess.writeFrame(frame{typ: frameSYNACK, streamID: st.id})
}

// Reset 拒绝这条流（服务端 dial 失败时调用），对端 Open 会立刻收到错误。
func (st *Stream) Reset() error {
	err := st.sess.writeFrame(frame{typ: frameRST, streamID: st.id})
	st.closeWith(ErrStreamReset)
	return err
}

func (st *Stream) Write(p []byte) (int, error) {
	select {
	case <-st.doneCh:
		return 0, io.ErrClosedPipe
	default:
	}

	sent := 0
	for len(p) > 0 {
		n := len(p)
		if n > maxFramePayload {
			n = maxFramePayload
		}
		if err := st.sess.writeFrame(frame{typ: frameDATA, streamID: st.id, payload: p[:n]}); err != nil {
			return sent, err
		}
		sent += n
		p = p[n:]
	}
	return sent, nil
}

func (st *Stream) Read(p []byte) (int, error) {
	if len(st.cur) == 0 {
		// 先把队列里已到达的数据取干净，避免收到 FIN 后把剩余数据丢掉
		select {
		case b := <-st.recv:
			st.cur = b
		default:
			select {
			case b := <-st.recv:
				st.cur = b
			case <-st.finCh:
				return 0, io.EOF
			case <-st.doneCh:
				return 0, io.EOF
			}
		}
	}
	n := copy(p, st.cur)
	st.cur = st.cur[n:]
	return n, nil
}

// CloseWrite 半关闭：告诉对端"我发完了"，但仍可继续读。
func (st *Stream) CloseWrite() error {
	st.finOnce.Do(func() {
		_ = st.sess.writeFrame(frame{typ: frameFIN, streamID: st.id})
	})
	return nil
}

// Close 彻底结束这条流：补发 FIN、从会话里摘掉。
func (st *Stream) Close() error {
	st.doneOnce.Do(func() {
		_ = st.CloseWrite()
		close(st.doneCh)
		st.sess.removeStream(st.id)
	})
	return nil
}

// deliver 把对端发来的数据放进接收队列。
//
// 队列满时**阻塞**就是背压：压力沿本 goroutine 回到 session 读循环，
// 进而让 TCP 接收窗口关闭、对端发送减速，而不是把内存无上限吃掉。
// 因此这里绝不能持锁阻塞——否则会把整条会话卡死。
func (st *Stream) deliver(p []byte) {
	select {
	case <-st.doneCh:
		return
	default:
	}
	select {
	case st.recv <- p:
	case <-st.doneCh:
	}
}

// closeRemote 收到对端 FIN/RST：读侧结束，但已到达的数据仍然能读完。
func (st *Stream) closeRemote() {
	st.recvOnce.Do(func() {
		close(st.finCh)
	})
}

func (st *Stream) signalOpen(err error) {
	select {
	case st.openCh <- err:
	default:
	}
}

func (st *Stream) closeWith(err error) {
	st.mu.Lock()
	if st.closeErr == nil {
		st.closeErr = err
	}
	st.mu.Unlock()

	st.doneOnce.Do(func() {
		close(st.doneCh)
		st.sess.removeStream(st.id)
	})
}

func (st *Stream) LocalAddr() net.Addr  { return st.sess.conn.LocalAddr() }
func (st *Stream) RemoteAddr() net.Addr { return st.sess.conn.RemoteAddr() }
func (st *Stream) SetDeadline(t time.Time) error {
	return errDeadlineUnsupported
}
func (st *Stream) SetReadDeadline(t time.Time) error {
	return errDeadlineUnsupported
}
func (st *Stream) SetWriteDeadline(t time.Time) error {
	return errDeadlineUnsupported
}

// halfCloser 是"能半关闭"的能力接口。
//
// v1 里直接断言 *net.TCPConn，到了 v2 就行不通了：Stream 不是 TCPConn，
// 但它同样能表达半关闭。按**能力**断言而不是按**具体类型**断言，
// 桥接代码才能同时服务 TCP 连接和 Stream。
type halfCloser interface {
	CloseWrite() error
}

// bridge 在两条"连接"之间双向拷贝，返回两个方向的字节数与第一个真实错误。
func bridge(a, b net.Conn) (a2b, b2a int64, err error) {
	var (
		errA2B, errB2A error
		wg             sync.WaitGroup
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		a2b, errA2B = copyAndHalfClose(b, a)
	}()
	go func() {
		defer wg.Done()
		b2a, errB2A = copyAndHalfClose(a, b)
	}()
	wg.Wait()

	_ = a.Close()
	_ = b.Close()
	return a2b, b2a, firstErr(errA2B, errB2A)
}

// firstErr 过滤掉 io.EOF：转发链路正常结束时不该在日志里报错。
func firstErr(errs ...error) error {
	for _, err := range errs {
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
	}
	return nil
}

func copyAndHalfClose(dst, src net.Conn) (int64, error) {
	n, err := io.Copy(dst, src)

	if hc, ok := dst.(halfCloser); ok {
		_ = hc.CloseWrite() // 我发完了，但你对我的回话我还要听
	} else {
		_ = dst.Close()
	}
	return n, err
}
