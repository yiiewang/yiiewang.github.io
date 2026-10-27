package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// ---------------- 握手协议的单元测试 ----------------

func TestHandshakeRoundTrip(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := writeHandshake(buf, "127.0.0.1:8000"); err != nil {
		t.Fatalf("write handshake: %v", err)
	}

	got, err := readHandshake(buf)
	if err != nil {
		t.Fatalf("read handshake: %v", err)
	}
	if got != "127.0.0.1:8000" {
		t.Fatalf("target mismatch: got %q", got)
	}
}

func TestWriteHandshakeRejectsLongTarget(t *testing.T) {
	err := writeHandshake(new(bytes.Buffer), strings.Repeat("a", maxTargetLen+1))
	if !errors.Is(err, errTargetTooLong) {
		t.Fatalf("expected errTargetTooLong, got %v", err)
	}
}

func TestReadHandshakeRejectsBadVersion(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := writeHandshake(buf, "127.0.0.1:8000"); err != nil {
		t.Fatalf("write handshake: %v", err)
	}
	buf.Bytes()[4] = 0x7f // 篡改版本号（前 4 字节是 magic）

	if _, err := readHandshake(buf); !errors.Is(err, errBadVersion) {
		t.Fatalf("expected errBadVersion, got %v", err)
	}
}

func TestReadHandshakeRejectsBadMagic(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := writeHandshake(buf, "127.0.0.1:8000"); err != nil {
		t.Fatalf("write handshake: %v", err)
	}
	copy(buf.Bytes()[0:4], "XXXX") // 篡改 magic

	if _, err := readHandshake(buf); !errors.Is(err, errBadMagic) {
		t.Fatalf("expected errBadMagic, got %v", err)
	}
}

func TestWriteReadStatusRoundTrip(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := writeStatus(buf, statusOK); err != nil {
		t.Fatalf("write status: %v", err)
	}

	got, err := readStatus(buf)
	if err != nil {
		t.Fatalf("read status: %v", err)
	}
	if got != statusOK {
		t.Fatalf("status mismatch: got %d, want %d", got, statusOK)
	}
}

// TestReadStatusDetectsHTTPPeer 覆盖真实踩过的一次误连：隧道端口上跑的其实是 HTTP 服务，
// 没有 magic 时客户端把响应首字节（'H' = 72）当成状态位读，日志里只剩 status=72 要猜。
func TestReadStatusDetectsHTTPPeer(t *testing.T) {
	got, err := readStatus(bytes.NewReader([]byte("HTTP/1.1 400 Bad Request\r\n\r\n")))
	if err == nil {
		t.Fatalf("expected error for HTTP peer, got status=%d", got)
	}
	if !strings.Contains(err.Error(), "looks like HTTP") {
		t.Fatalf("error should name HTTP explicitly, got %v", err)
	}
}

func TestReadStatusRejectsBadMagic(t *testing.T) {
	if _, err := readStatus(bytes.NewReader([]byte("\x00\x01\x02\x03\x01\x00"))); !errors.Is(err, errBadMagic) {
		t.Fatalf("expected errBadMagic, got %v", err)
	}
}

// dripReader 每次只吐 1 个字节，用来模拟 TCP 拆包。
type dripReader struct {
	data []byte
	off  int
}

func (d *dripReader) Read(p []byte) (int, error) {
	if d.off >= len(d.data) {
		return 0, io.EOF
	}
	n := copy(p, d.data[d.off:d.off+1])
	d.off += n
	return n, nil
}

// TestReadHandshakeAcrossPackets 验证握手读取不会被拆包打穿——
// 这是把 conn.Read 换成 io.ReadFull 才有保证的性质。
func TestReadHandshakeAcrossPackets(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := writeHandshake(buf, "127.0.0.1:9000"); err != nil {
		t.Fatalf("write handshake: %v", err)
	}

	got, err := readHandshake(&dripReader{data: buf.Bytes()})
	if err != nil {
		t.Fatalf("read handshake from dripped reader: %v", err)
	}
	if got != "127.0.0.1:9000" {
		t.Fatalf("target mismatch: got %q", got)
	}
}

// ---------------- 端到端转发测试 ----------------

func TestForwardRoundTrip(t *testing.T) {
	target := startEchoTarget(t)
	local := startTunnelClient(t, startTunnelServer(t), target)

	conn, err := net.Dial("tcp", local)
	if err != nil {
		t.Fatalf("dial local: %v", err)
	}
	defer conn.Close()

	const msg = "hello v1"
	if _, err := conn.Write([]byte(msg)); err != nil {
		t.Fatalf("write local: %v", err)
	}

	buf := make([]byte, len(msg))
	readFullWithDeadline(t, conn, buf)
	if string(buf) != msg {
		t.Fatalf("echo mismatch: got %q, want %q", buf, msg)
	}
}

// TestForwardRejectsUnreachableTarget 验证目标不可达时链路能干净收尾：
// 服务端回 statusDialFailed 并关闭连接，本地应读到 EOF 而不是永远挂着。
func TestForwardRejectsUnreachableTarget(t *testing.T) {
	dead := freeAddr(t)
	local := startTunnelClient(t, startTunnelServer(t), dead)

	conn, err := net.Dial("tcp", local)
	if err != nil {
		t.Fatalf("dial local: %v", err)
	}
	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatalf("set deadline: %v", err)
	}
	var one [1]byte
	if _, err := conn.Read(one[:]); !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF after rejected handshake, got %v", err)
	}
}

// TestForwardDetectsNonMinibridgeServer 验证误连到非 minibridge 服务时能快速失败：
// 客户端应立刻关闭本地连接，而不是挂在那里等一个永远不来的响应。
func TestForwardDetectsNonMinibridgeServer(t *testing.T) {
	bogus := startTarget(t, func(conn net.Conn) {
		defer conn.Close()
		_, _ = io.ReadFull(conn, make([]byte, 1)) // 等客户端先把握手发过来
		_, _ = conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
	})
	local := startTunnelClient(t, bogus, "127.0.0.1:8000")

	conn, err := net.Dial("tcp", local)
	if err != nil {
		t.Fatalf("dial local: %v", err)
	}
	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatalf("set deadline: %v", err)
	}
	var one [1]byte
	if _, err := conn.Read(one[:]); !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF after non-minibridge peer, got %v", err)
	}
}

// TestForwardPropagatesHalfClose 验证半关闭能穿透整条链路：
// 客户端发完数据后 CloseWrite，目标端收完再回复，客户端仍应收到回复。
// 若桥接里用 Close 代替 CloseWrite，这个用例会失败。
func TestForwardPropagatesHalfClose(t *testing.T) {
	target := startCollectTarget(t, "bye")
	local := startTunnelClient(t, startTunnelServer(t), target)

	conn, err := net.Dial("tcp", local)
	if err != nil {
		t.Fatalf("dial local: %v", err)
	}
	defer conn.Close()

	if _, err := conn.Write([]byte("ping")); err != nil {
		t.Fatalf("write local: %v", err)
	}
	tcp, ok := conn.(*net.TCPConn)
	if !ok {
		t.Fatalf("expected *net.TCPConn, got %T", conn)
	}
	if err := tcp.CloseWrite(); err != nil {
		t.Fatalf("close write: %v", err)
	}

	buf := make([]byte, 3)
	readFullWithDeadline(t, conn, buf)
	if string(buf) != "bye" {
		t.Fatalf("reply mismatch: got %q, want %q", buf, "bye")
	}
}

func TestForwardConcurrent(t *testing.T) {
	target := startEchoTarget(t)
	local := startTunnelClient(t, startTunnelServer(t), target)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			msg := fmt.Sprintf("msg-%02d", i)
			conn, err := net.Dial("tcp", local)
			if err != nil {
				t.Errorf("dial local: %v", err)
				return
			}
			defer conn.Close()

			if _, err := conn.Write([]byte(msg)); err != nil {
				t.Errorf("write local: %v", err)
				return
			}
			buf := make([]byte, len(msg))
			if err := conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
				t.Errorf("set deadline: %v", err)
				return
			}
			if _, err := io.ReadFull(conn, buf); err != nil {
				t.Errorf("read local: %v", err)
				return
			}
			if string(buf) != msg {
				t.Errorf("echo mismatch: got %q, want %q", buf, msg)
			}
		}(i)
	}
	wg.Wait()
}

// ---------------- 测试桩 ----------------

// startEchoTarget 起一个回显服务：把收到的字节原样写回，直到对端 EOF。
func startEchoTarget(t *testing.T) string {
	t.Helper()
	return startTarget(t, func(conn net.Conn) {
		defer conn.Close()
		_, _ = io.Copy(conn, conn)
	})
}

// startCollectTarget 起一个「收完再回」的服务：读到 EOF 后回复固定字符串。
func startCollectTarget(t *testing.T, reply string) string {
	t.Helper()
	return startTarget(t, func(conn net.Conn) {
		defer conn.Close()
		_, _ = io.Copy(io.Discard, conn)
		_, _ = conn.Write([]byte(reply))
	})
}

// startTarget 在随机端口上起一个 TCP 服务，每个连接交给 handler 处理。
func startTarget(t *testing.T, handler func(net.Conn)) string {
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

// startTunnelServer 在随机端口起隧道服务端，返回它的地址。
func startTunnelServer(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen tunnel server: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go serveTunnel(conn)
		}
	}()
	return ln.Addr().String()
}

// startTunnelClient 在随机端口起转发客户端，返回本地监听地址。
func startTunnelClient(t *testing.T, serverAddr, target string) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen tunnel client: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go forwardConn(conn, serverAddr, target)
		}
	}()
	return ln.Addr().String()
}

// freeAddr 返回一个当前没人监听的地址，用于构造「目标不可达」场景。
func freeAddr(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("probe free addr: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	return addr
}

// readFullWithDeadline 带超时地读满 buf，避免用例失败时把测试挂死。
func readFullWithDeadline(t *testing.T, conn net.Conn, buf []byte) {
	t.Helper()

	if err := conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatalf("set deadline: %v", err)
	}
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatalf("read full: %v", err)
	}
}
