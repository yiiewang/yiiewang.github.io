package main

import (
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

// startReverseServer 起一个带反向转发能力的隧道服务端。
// 明文无 TLS：反向数据面与控制面不依赖传输层，测试只关心协议本身。
func startReverseServer(t *testing.T) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen reverse server: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			sess := newSession(conn, false)
			go newReverseManager(sess).run()
			go serveSession(sess, testAllowAll(t))
		}
	}()
	return ln.Addr().String()
}

// newReverseClient 连上服务端并起客户端数据面（处理服务端开来的反向流）。
func newReverseClient(t *testing.T, serverAddr string, allow *AllowList) *Session {
	t.Helper()

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		t.Fatalf("dial reverse server: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	sess := newSession(conn, true)
	go serveSession(sess, allow)
	return sess
}

// TestReverseEcho 是 v5 的核心验收：连服务端上的 bind 端口，
// 数据要穿过反向流打到客户端身边的回显目标再回来。
func TestReverseEcho(t *testing.T) {
	echo := startMuxEchoTarget(t) // "客户端身边"的目标
	sess := newReverseClient(t, startReverseServer(t), testAllowAll(t))

	bind := freeMuxAddr(t)
	if err := sess.ListenRemote(bind, echo, 3*time.Second); err != nil {
		t.Fatalf("listen remote: %v", err)
	}

	conn, err := net.Dial("tcp", bind)
	if err != nil {
		t.Fatalf("dial bind: %v", err)
	}
	defer conn.Close()

	const msg = "hello reverse"
	if _, err := conn.Write([]byte(msg)); err != nil {
		t.Fatalf("write: %v", err)
	}
	buf := make([]byte, len(msg))
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	if _, err := io.ReadFull(conn, buf); err != nil {
		t.Fatalf("read echo: %v", err)
	}
	if string(buf) != msg {
		t.Fatalf("echo mismatch: got %q, want %q", buf, msg)
	}
}

// TestReverseDeniedByClientAllow：反向流的目标不在客户端白名单里，
// 必须被 RST 拒绝——反向转发意味着服务端能驱动客户端 dial，
// 这道闸门是对称 SSRF 防护的另一半。
func TestReverseDeniedByClientAllow(t *testing.T) {
	echo := startMuxEchoTarget(t)

	// 白名单只放行一个跟 echo 无关的端口
	allow, err := NewAllowList([]string{"127.0.0.1:1"})
	if err != nil {
		t.Fatalf("build deny-all allow list: %v", err)
	}
	sess := newReverseClient(t, startReverseServer(t), allow)

	bind := freeMuxAddr(t)
	if err := sess.ListenRemote(bind, echo, 3*time.Second); err != nil {
		t.Fatalf("listen remote: %v", err)
	}

	conn, err := net.Dial("tcp", bind)
	if err != nil {
		t.Fatalf("dial bind: %v", err)
	}
	defer conn.Close()

	_, _ = conn.Write([]byte("ping"))
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	buf := make([]byte, 4)
	if _, err := conn.Read(buf); err == nil {
		t.Fatalf("expected denial to close the connection, got data %q", buf)
	}
}

// TestReverseUnlisten：注销后端口必须真的不可达。
func TestReverseUnlisten(t *testing.T) {
	echo := startMuxEchoTarget(t)
	sess := newReverseClient(t, startReverseServer(t), testAllowAll(t))

	bind := freeMuxAddr(t)
	if err := sess.ListenRemote(bind, echo, 3*time.Second); err != nil {
		t.Fatalf("listen remote: %v", err)
	}

	if err := sess.UnlistenRemote(bind, 3*time.Second); err != nil {
		t.Fatalf("unlisten remote: %v", err)
	}

	if _, err := net.Dial("tcp", bind); err == nil {
		t.Fatalf("bind should be unreachable after unlisten")
	}
}

// TestReverseDuplicateBindRejected：同一 bind 重复注册按错误回应。
func TestReverseDuplicateBindRejected(t *testing.T) {
	echo := startMuxEchoTarget(t)
	sess := newReverseClient(t, startReverseServer(t), testAllowAll(t))

	bind := freeMuxAddr(t)
	if err := sess.ListenRemote(bind, echo, 3*time.Second); err != nil {
		t.Fatalf("first listen remote: %v", err)
	}

	err := sess.ListenRemote(bind, echo, 3*time.Second)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("want duplicate error, got %v", err)
	}
}

// TestReverseListenersDieWithSession：会话死亡必须连带关掉它名下的
// 反向监听——否则端口泄漏，而且没人说得清那个端口还通向哪。
func TestReverseListenersDieWithSession(t *testing.T) {
	a, b := net.Pipe()
	client := newSession(a, true)
	server := newSession(b, false)
	t.Cleanup(func() { _ = client.Close(nil); _ = server.Close(nil) })
	go newReverseManager(server).run()

	bind := freeMuxAddr(t)
	if err := client.ListenRemote(bind, "127.0.0.1:8000", 3*time.Second); err != nil {
		t.Fatalf("listen remote: %v", err)
	}

	_ = server.Close(errors.New("server killed"))

	// 轮询等泄漏的监听被关：能重新 Listen 成功说明端口已释放
	deadline := time.Now().Add(2 * time.Second)
	for {
		ln, err := net.Listen("tcp", bind)
		if err == nil {
			_ = ln.Close()
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("reverse listener leaked after session death: %v", err)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// parseBindSpec：SSH 风格 [地址:]端口:目标:端口，逗号分隔多条。
func TestParseBindSpec(t *testing.T) {
	ms, err := parseBindSpec("90:127.0.0.1:80,:18000:127.0.0.1:8000,10.0.0.5:91:db.local:3306")
	if err != nil {
		t.Fatalf("parse bind spec: %v", err)
	}
	want := []portMapping{
		{bind: ":90", target: "127.0.0.1:80"},
		{bind: ":18000", target: "127.0.0.1:8000"},
		{bind: "10.0.0.5:91", target: "db.local:3306"},
	}
	if len(ms) != len(want) {
		t.Fatalf("want %d mappings, got %d", len(want), len(ms))
	}
	for i, w := range want {
		if ms[i] != w {
			t.Fatalf("mapping[%d] = %+v, want %+v", i, ms[i], w)
		}
	}

	for _, bad := range []string{"", "90", "90:127.0.0.1", ":90::80", "a:b:c:d:e"} {
		if _, err := parseBindSpec(bad); err == nil {
			t.Fatalf("%q must be rejected", bad)
		}
	}
}
