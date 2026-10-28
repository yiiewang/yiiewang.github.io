package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

func TestStreamEcho(t *testing.T) {
	target := startMuxEchoTarget(t)
	sess := newTestSession(t, startMuxServer(t))

	st, err := sess.Open(target)
	if err != nil {
		t.Fatalf("open stream: %v", err)
	}
	defer st.Close()

	const msg = "hello v2"
	if _, err := st.Write([]byte(msg)); err != nil {
		t.Fatalf("write stream: %v", err)
	}

	buf := make([]byte, len(msg))
	readFullTimeout(t, st, buf, 3*time.Second)
	if string(buf) != msg {
		t.Fatalf("echo mismatch: got %q, want %q", buf, msg)
	}
}

// TestStreamResetOnUnreachableTarget 对应 v1 的"目标不可达要快速失败"：
// 服务端 dial 失败 → 回 RST → 客户端 Open 立刻返回错误，而不是挂着。
func TestStreamResetOnUnreachableTarget(t *testing.T) {
	sess := newTestSession(t, startMuxServer(t))

	if _, err := sess.Open(freeMuxAddr(t)); !errors.Is(err, ErrStreamReset) {
		t.Fatalf("expected ErrStreamReset, got %v", err)
	}
}

// TestSessionSurvivesReset 验证单条流被拒不影响整条会话——
// 这是多路复用最容易出错的地方：一条流出事不能连坐其他流。
func TestSessionSurvivesReset(t *testing.T) {
	target := startMuxEchoTarget(t)
	sess := newTestSession(t, startMuxServer(t))

	if _, err := sess.Open(freeMuxAddr(t)); !errors.Is(err, ErrStreamReset) {
		t.Fatalf("expected ErrStreamReset, got %v", err)
	}

	st, err := sess.Open(target)
	if err != nil {
		t.Fatalf("open stream after reset: %v", err)
	}
	defer st.Close()

	const msg = "still alive"
	if _, err := st.Write([]byte(msg)); err != nil {
		t.Fatalf("write stream: %v", err)
	}
	buf := make([]byte, len(msg))
	readFullTimeout(t, st, buf, 3*time.Second)
	if string(buf) != msg {
		t.Fatalf("echo mismatch: got %q", buf)
	}
}

// TestManyStreamsOverOneSession 是 v2 的核心不变量：
// 100 条流跑在同一条 TCP 上，数据不能串。
func TestManyStreamsOverOneSession(t *testing.T) {
	target := startMuxEchoTarget(t)
	sess := newTestSession(t, startMuxServer(t))

	const n = 100
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			msg := fmt.Sprintf("msg-%03d", i)
			st, err := sess.Open(target)
			if err != nil {
				t.Errorf("open stream %d: %v", i, err)
				return
			}
			defer st.Close()

			if _, err := st.Write([]byte(msg)); err != nil {
				t.Errorf("write stream %d: %v", i, err)
				return
			}
			buf := make([]byte, len(msg))
			readFullTimeout(t, st, buf, 5*time.Second)
			if string(buf) != msg {
				t.Errorf("stream %d cross-talk: got %q, want %q", i, buf, msg)
			}
		}(i)
	}
	wg.Wait()
}

// TestStreamHalfClose 验证半关闭能穿透"本地连接 → Stream → 目标"这条链：
// 客户端 CloseWrite 之后，仍要收到目标端补发的响应。
func TestStreamHalfClose(t *testing.T) {
	target := startMuxCollectTarget(t, "bye")
	sess := newTestSession(t, startMuxServer(t))

	st, err := sess.Open(target)
	if err != nil {
		t.Fatalf("open stream: %v", err)
	}
	defer st.Close()

	if _, err := st.Write([]byte("ping")); err != nil {
		t.Fatalf("write stream: %v", err)
	}
	if err := st.CloseWrite(); err != nil {
		t.Fatalf("close write: %v", err)
	}

	buf := make([]byte, 3)
	readFullTimeout(t, st, buf, 3*time.Second)
	if string(buf) != "bye" {
		t.Fatalf("reply mismatch: got %q, want %q", buf, "bye")
	}
}

// TestSessionCloseClosesStreams 验证会话断开时所有流一起收尾——
// 否则每条上层连接都会挂死，这是放大版的僵尸连接。
func TestSessionCloseClosesStreams(t *testing.T) {
	target := startMuxEchoTarget(t)
	sess := newTestSession(t, startMuxServer(t))

	st, err := sess.Open(target)
	if err != nil {
		t.Fatalf("open stream: %v", err)
	}
	if err := sess.Close(errors.New("test shutdown")); err == nil {
		t.Fatalf("expected session close to report cause")
	}

	buf := make([]byte, 1)
	if _, err := st.Read(buf); !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF after session closed, got %v", err)
	}
}

// ---------------- 测试桩 ----------------

// startMuxServer 起一个 v2 隧道服务端，返回它的地址。
func startMuxServer(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen mux server: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go serveSession(newSession(conn, false), testAllowAll(t))
		}
	}()
	return ln.Addr().String()
}

// testAllowAll 给从 v2 复制过来的用例用：它们只关心转发本身，不关心访问控制。
func testAllowAll(t *testing.T) *AllowList {
	t.Helper()

	allow, err := NewAllowList([]string{"0.0.0.0/0:*"})
	if err != nil {
		t.Fatalf("build allow list: %v", err)
	}
	return allow
}

// newTestSession 作为客户端连上隧道服务端，返回会话。
func newTestSession(t *testing.T, serverAddr string) *Session {
	t.Helper()

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		t.Fatalf("dial mux server: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return newSession(conn, true)
}

func startMuxEchoTarget(t *testing.T) string {
	t.Helper()
	return startMuxTarget(t, func(conn net.Conn) {
		defer conn.Close()
		_, _ = io.Copy(conn, conn)
	})
}

// startMuxCollectTarget 起一个"收完再回"的服务，用来验证半关闭。
func startMuxCollectTarget(t *testing.T, reply string) string {
	t.Helper()
	return startMuxTarget(t, func(conn net.Conn) {
		defer conn.Close()
		_, _ = io.Copy(io.Discard, conn)
		_, _ = conn.Write([]byte(reply))
	})
}

func startMuxTarget(t *testing.T, handler func(net.Conn)) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen target: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handler(conn)
		}
	}()
	return ln.Addr().String()
}

// freeMuxAddr 返回一个没人监听的地址，用来构造"目标不可达"。
func freeMuxAddr(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("probe free addr: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	return addr
}

// readFullTimeout 带超时地读满 buf。Stream 不支持 deadline，只能靠 goroutine + 定时器兜底。
func readFullTimeout(t *testing.T, r io.Reader, buf []byte, d time.Duration) {
	t.Helper()

	type result struct {
		n   int
		err error
	}
	ch := make(chan result, 1)
	go func() {
		n, err := io.ReadFull(r, buf)
		ch <- result{n: n, err: err}
	}()

	select {
	case res := <-ch:
		if res.err != nil {
			t.Fatalf("read full: %v", res.err)
		}
	case <-time.After(d):
		t.Fatalf("read full timeout after %v", d)
	}
}
