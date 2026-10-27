// Package main 是 minibridge 的第一阶段：一条转发独占一条 TCP 连接的最小端口转发工具。
//
// 它要做的事等价于 ssh -L 8000:localhost:8000 user@server，但隧道是自己写的，
// 每一层都能看见：握手、拨号、双向拷贝、半关闭。
package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// protocolVersion 是握手协议版本号。v1 里一条隧道连接只服务一条转发，
// 下一步（v2）会把它换成带 streamID 的帧协议，让多条转发复用同一条连接。
const protocolVersion byte = 1

// magic 是每个握手报文开头的 4 字节标识，作用只有一个：**认出对端不是自己人**。
//
// 没有它的时候，客户端把 HTTP 响应的首字节（'H' = 72）当成状态位读，
// 日志里只能打出 status=72 这种要推理半天的输出。
const magic = "MBRD"

// 握手请求头 [magic:4][ver:1][len:2]，握手响应固定 [magic:4][ver:1][status:1]。
const (
	reqHeadLen = 7
	ackLen     = 6
)

// 握手响应状态：服务端 dial 目标地址的成败。
const (
	statusOK         byte = 0
	statusDialFailed byte = 1
)

// maxTargetLen 限制目标地址长度，避免恶意客户端发超长字符串把服务端内存打满。
const maxTargetLen = 255

// dialTimeout 是连接隧道服务端与连接目标地址的超时时间。
const dialTimeout = 5 * time.Second

var (
	errBadVersion    = errors.New("unsupported protocol version")
	errTargetTooLong = errors.New("target address too long")
	errBadMagic      = errors.New("peer is not a minibridge server: bad magic")
)

// badPeerErr 根据抓到的前 4 个字节给出可诊断的错误。
//
// 误连是最常见的运维事故：端口上跑的可能是 HTTP 服务、代理或拦截页，
// 它们回话的方式和自家协议完全不同，报错时把看到的字节带上能省掉一轮抓包。
func badPeerErr(prefix []byte) error {
	if len(prefix) >= 4 && string(prefix[:4]) == "HTTP" {
		return fmt.Errorf("peer is not a minibridge server (looks like HTTP, first bytes %q)", prefix[:4])
	}
	return fmt.Errorf("%w (first bytes %q)", errBadMagic, prefix)
}

// writeHandshake 把目标地址按 [magic:4][ver:1][len:2][target:N] 写入隧道连接。
//
// 隧道服务端必须知道"你要我转发到哪"，这就是握手存在的唯一理由：
// 一条连接只服务一个目标，目标地址只能由客户端在建立连接时告诉服务端。
func writeHandshake(w io.Writer, target string) error {
	if len(target) > maxTargetLen {
		return errTargetTooLong
	}
	buf := make([]byte, reqHeadLen+len(target))
	copy(buf[0:4], magic)
	buf[4] = protocolVersion
	binary.BigEndian.PutUint16(buf[5:reqHeadLen], uint16(len(target)))
	copy(buf[reqHeadLen:], target)
	_, err := w.Write(buf)
	return err
}

// readHandshake 读取握手请求。
//
// 注意这里用 io.ReadFull 而不是 conn.Read：TCP 是字节流，一次 Read 拿到多少字节
// 都不保证，只有 ReadFull 能保证报文头一定读全——这是所有自研协议的第一道坎。
func readHandshake(r io.Reader) (string, error) {
	head := make([]byte, reqHeadLen)
	if _, err := io.ReadFull(r, head); err != nil {
		return "", err
	}
	if string(head[0:4]) != magic {
		return "", badPeerErr(head[0:4])
	}
	if head[4] != protocolVersion {
		return "", errBadVersion
	}
	n := int(binary.BigEndian.Uint16(head[5:reqHeadLen]))
	if n > maxTargetLen {
		return "", errTargetTooLong
	}
	target := make([]byte, n)
	if _, err := io.ReadFull(r, target); err != nil {
		return "", err
	}
	return string(target), nil
}

// writeStatus 回复握手结果，格式 [magic:4][ver:1][status:1]。
func writeStatus(w io.Writer, status byte) error {
	buf := make([]byte, ackLen)
	copy(buf[0:4], magic)
	buf[4] = protocolVersion
	buf[5] = status
	_, err := w.Write(buf)
	return err
}

// readStatus 读取握手响应并校验 magic 与版本，返回状态字节。
func readStatus(r io.Reader) (byte, error) {
	ack := make([]byte, ackLen)
	if _, err := io.ReadFull(r, ack); err != nil {
		return 0, err
	}
	if string(ack[0:4]) != magic {
		return 0, badPeerErr(ack[0:4])
	}
	if ack[4] != protocolVersion {
		return 0, errBadVersion
	}
	return ack[5], nil
}

// serveTunnel 处理一条隧道连接：读握手 → dial 目标 → 双向桥接。
func serveTunnel(conn net.Conn) {
	defer conn.Close()

	target, err := readHandshake(conn)
	if err != nil {
		log.Printf("handshake failed, remote=%s, err=%v", conn.RemoteAddr(), err)
		return
	}

	upstream, err := net.DialTimeout("tcp", target, dialTimeout)
	if err != nil {
		_ = writeStatus(conn, statusDialFailed)
		log.Printf("dial target failed, target=%s, err=%v", target, err)
		return
	}
	defer upstream.Close()

	if err := writeStatus(conn, statusOK); err != nil {
		log.Printf("write handshake ack failed, target=%s, err=%v", target, err)
		return
	}

	in, out, berr := bridge(conn, upstream)
	log.Printf("tunnel finished, target=%s, client_to_target=%d, target_to_client=%d, err=%v", target, in, out, berr)
}

// forwardConn 处理一条本地连接：连隧道服务端 → 发握手 → 双向桥接。
func forwardConn(local net.Conn, serverAddr, target string) {
	defer local.Close()

	remote, err := net.DialTimeout("tcp", serverAddr, dialTimeout)
	if err != nil {
		log.Printf("dial tunnel server failed, server=%s, err=%v", serverAddr, err)
		return
	}
	defer remote.Close()

	if err := writeHandshake(remote, target); err != nil {
		log.Printf("write handshake failed, target=%s, err=%v", target, err)
		return
	}

	// 读握手响应必须有超时：对端不是 minibridge 时，可能永远凑不齐 6 个字节，
	// 没有 deadline 这个 goroutine 就永久挂住了。
	if err := remote.SetReadDeadline(time.Now().Add(dialTimeout)); err != nil {
		log.Printf("set handshake deadline failed, target=%s, err=%v", target, err)
		return
	}
	status, err := readStatus(remote)
	if err != nil {
		log.Printf("tunnel handshake failed, server=%s, target=%s, err=%v", serverAddr, target, err)
		return
	}
	// 转发阶段没有时限，务必把握手用的 deadline 清掉
	if err := remote.SetReadDeadline(time.Time{}); err != nil {
		log.Printf("clear handshake deadline failed, target=%s, err=%v", target, err)
		return
	}
	if status != statusOK {
		log.Printf("tunnel server rejected, target=%s, status=%d", target, status)
		return
	}

	in, out, berr := bridge(local, remote)
	log.Printf("forward finished, target=%s, local_to_remote=%d, remote_to_local=%d, err=%v", target, in, out, berr)
}

// bridge 在两条连接之间双向拷贝字节流，返回两个方向的字节数与第一个真实错误。
//
// 关键设计：某一方向读到 EOF 时，对另一端调用 CloseWrite（半关闭）而不是直接 Close。
// 半关闭让对端能收到 EOF、同时把自己剩下的数据发回来——HTTP keep-alive、
// 数据库驱动等都依赖这个语义。直接 Close 会把"我发完了"变成"我挂了"，
// 对端还没发完的数据就丢了。
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

	closeBoth(a, b)
	return a2b, b2a, firstErr(errA2B, errB2A)
}

// copyAndHalfClose 把 src 的数据拷给 dst，src 结束后对 dst 半关闭。
//
// io.Copy 正常结束（src 读到 EOF）时返回 nil error，所以这里不区分"正常结束"与
// "对端 RST"——两者都要对 dst 做同样的收尾动作。
func copyAndHalfClose(dst, src net.Conn) (int64, error) {
	n, err := io.Copy(dst, src)

	if tcp, ok := dst.(*net.TCPConn); ok {
		_ = tcp.CloseWrite()
	} else {
		_ = dst.Close()
	}
	return n, err
}

// closeBoth 关闭两条连接，重复关闭的错误忽略。
func closeBoth(conns ...net.Conn) {
	for _, c := range conns {
		_ = c.Close()
	}
}

// firstErr 返回第一个既非 nil 也非 io.EOF 的错误。
//
// 转发链路的正常终止在日志里应该表现为 err=<nil>，被 RST 才需要报出来。
func firstErr(errs ...error) error {
	for _, err := range errs {
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
	}
	return nil
}
