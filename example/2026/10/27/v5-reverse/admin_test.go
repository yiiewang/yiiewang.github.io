package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

// adminConn 是测试用的管理面客户端：一个连接一个 scanner，逐行收发。
type adminConn struct {
	t  *testing.T
	c  net.Conn
	sc *bufio.Scanner
}

func dialAdmin(t *testing.T, addr string) *adminConn {
	t.Helper()
	c, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial admin: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return &adminConn{t: t, c: c, sc: bufio.NewScanner(c)}
}

// cmd 发一条命令，收一行回应（LIST 的多行回应要另外 Scan）。
func (a *adminConn) cmd(line string) string {
	a.t.Helper()
	if _, err := fmt.Fprintf(a.c, "%s\n", line); err != nil {
		a.t.Fatalf("send %q: %v", line, err)
	}
	if !a.sc.Scan() {
		a.t.Fatalf("no reply for %q", line)
	}
	return a.sc.Text()
}

// newAdminStack 起一套最小客户端运行时：明文会话 + 注册表 + tunnel + admin。
func newAdminStack(t *testing.T) (*tunnel, *reverseRegistry, string) {
	t.Helper()
	srv := startReverseServer(t)
	reg := newReverseRegistry(nil)
	allow := testAllowAll(t)

	dial := func() (*Session, error) {
		conn, err := net.Dial("tcp", srv)
		if err != nil {
			return nil, err
		}
		return bootSession(conn, reg, allow)
	}
	tun, err := newTunnel(dial)
	if err != nil {
		t.Fatalf("first dial: %v", err)
	}
	go tun.reconnectLoop(dial, time.Millisecond)

	aln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen admin: %v", err)
	}
	t.Cleanup(func() { _ = aln.Close() })
	go serveAdmin(aln, reg, tun)
	return tun, reg, aln.Addr().String()
}

// echoThrough 连 addr 写 msg，断言原样回来。
func echoThrough(t *testing.T, addr, msg string) {
	t.Helper()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial %s: %v", addr, err)
	}
	defer conn.Close()
	if _, err := conn.Write([]byte(msg)); err != nil {
		t.Fatalf("write: %v", err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	buf := make([]byte, len(msg))
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatalf("read echo: %v", err)
	}
	if string(buf) != msg {
		t.Fatalf("echo mismatch: got %q, want %q", buf, msg)
	}
}

// TestAdminLifecycle：LISTEN 生效 → LIST 可见 → UNLISTEN 后端口不可达。
func TestAdminLifecycle(t *testing.T) {
	echo := startMuxEchoTarget(t)
	_, _, adminAddr := newAdminStack(t)
	admin := dialAdmin(t, adminAddr)

	if got := admin.cmd("BOGUS"); !strings.HasPrefix(got, "ERR") {
		t.Fatalf("bogus command should fail, got %q", got)
	}

	bind := freeMuxAddr(t)
	if got := admin.cmd("LISTEN " + bind + " " + echo); got != "OK listening" {
		t.Fatalf("listen: %q", got)
	}

	// 真通了才算数：连 bind，数据穿过反向流到 echo 再回来
	echoThrough(t, bind, "via admin")

	if got := admin.cmd("LIST"); !strings.Contains(got, "1 mapping") {
		t.Fatalf("list: %q", got)
	}
	if !admin.sc.Scan() {
		t.Fatalf("list entry missing")
	}
	if line := admin.sc.Text(); !strings.Contains(line, bind) || !strings.Contains(line, echo) {
		t.Fatalf("list entry: %q", line)
	}

	if got := admin.cmd("UNLISTEN " + bind); got != "OK removed" {
		t.Fatalf("unlisten: %q", got)
	}
	if _, err := net.Dial("tcp", bind); err == nil {
		t.Fatalf("bind should be unreachable after unlisten")
	}
}

// TestAdminResyncAfterReconnect：会话死亡换新后，注册表整体重放，
// 运行时注册的反向转发不用人工重做——"不用重启"的另一半。
func TestAdminResyncAfterReconnect(t *testing.T) {
	echo := startMuxEchoTarget(t)
	tun, _, adminAddr := newAdminStack(t)
	admin := dialAdmin(t, adminAddr)

	bind := freeMuxAddr(t)
	if got := admin.cmd("LISTEN " + bind + " " + echo); got != "OK listening" {
		t.Fatalf("listen: %q", got)
	}

	// 杀掉当前会话：tunnel 自动换新，bootSession 里重放注册表
	old := tun.current()
	_ = old.Close(nil)

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if tun.current() != old {
			break
		}
		time.Sleep(2 * time.Millisecond)
	}
	if tun.current() == old {
		t.Fatalf("tunnel did not reconnect")
	}

	// 重放后 bind 必须仍然可达（补试窗口 200ms，多等一点）
	time.Sleep(300 * time.Millisecond)
	echoThrough(t, bind, "after reconnect")
}

// TestAdminDeferredRegistration：掉线窗口里注册 → 记账不报错；
// 新会话上线时重放生效。
func TestAdminDeferredRegistration(t *testing.T) {
	echo := startMuxEchoTarget(t)
	srv := startReverseServer(t)
	reg := newReverseRegistry(nil)

	// 第一条会话：建好就杀，制造"掉线窗口"
	conn1, err := net.Dial("tcp", srv)
	if err != nil {
		t.Fatalf("dial first: %v", err)
	}
	sess1, err := bootSession(conn1, reg, testAllowAll(t))
	if err != nil {
		t.Fatalf("boot first session: %v", err)
	}
	tun := &tunnel{sess: sess1}
	_ = sess1.Close(nil)

	bind := freeMuxAddr(t)
	got := execAdmin("LISTEN "+bind+" "+echo, reg, tun)
	if len(got) != 1 || !strings.Contains(got[0], "session down") {
		t.Fatalf("want deferred OK, got %v", got)
	}

	// 新会话上线：bootSession 重放注册表，bind 生效
	conn2, err := net.Dial("tcp", srv)
	if err != nil {
		t.Fatalf("dial second: %v", err)
	}
	if _, err := bootSession(conn2, reg, testAllowAll(t)); err != nil {
		t.Fatalf("boot second session: %v", err)
	}
	echoThrough(t, bind, "deferred registration")
}
