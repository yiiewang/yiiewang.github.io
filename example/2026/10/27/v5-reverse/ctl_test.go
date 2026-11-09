package main

import (
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

// serveCtl 是测试里的服务端控制面桩：收一条、验一条、回一条。
// 真实的 reverse manager（服务端应用层）遵循完全相同的节奏。
func serveCtl(t *testing.T, sess *Session, handle func(ctlMsg) error) {
	t.Helper()
	go func() {
		for {
			select {
			case msg := <-sess.Ctl():
				sess.ReplyCtl(msg, handle(msg))
			case <-sess.Done():
				return
			}
		}
	}()
}

func newCtlPair(t *testing.T) (client, server *Session) {
	t.Helper()
	a, b := net.Pipe()
	client = newSession(a, true)
	server = newSession(b, false)
	t.Cleanup(func() { _ = client.Close(nil); _ = server.Close(nil) })
	return client, server
}

// TestCtlListenRoundTrip：LISTEN 的 payload 要原样穿过两端的编解码。
func TestCtlListenRoundTrip(t *testing.T) {
	client, server := newCtlPair(t)

	serveCtl(t, server, func(msg ctlMsg) error {
		if msg.unlisten {
			t.Errorf("want LISTEN, got UNLISTEN")
		}
		if msg.bind != "127.0.0.1:19001" || msg.target != "127.0.0.1:8000" {
			t.Errorf("payload mismatch: bind=%q target=%q", msg.bind, msg.target)
		}
		return nil
	})

	if err := client.ListenRemote("127.0.0.1:19001", "127.0.0.1:8000", 3*time.Second); err != nil {
		t.Fatalf("listen remote: %v", err)
	}
}

// TestCtlErrorPropagates：服务端注册失败（如端口被占），错误要回到请求方。
func TestCtlErrorPropagates(t *testing.T) {
	client, server := newCtlPair(t)

	serveCtl(t, server, func(ctlMsg) error {
		return errors.New("address already in use")
	})

	err := client.ListenRemote("127.0.0.1:19001", "127.0.0.1:8000", 3*time.Second)
	if err == nil || err.Error() != "address already in use" {
		t.Fatalf("want propagated error, got %v", err)
	}
}

// TestCtlUnlisten：UNLISTEN 只带 bind，不带 target。
func TestCtlUnlisten(t *testing.T) {
	client, server := newCtlPair(t)

	serveCtl(t, server, func(msg ctlMsg) error {
		if !msg.unlisten {
			t.Errorf("want UNLISTEN, got LISTEN")
		}
		if msg.bind != "127.0.0.1:19001" || msg.target != "" {
			t.Errorf("payload mismatch: bind=%q target=%q", msg.bind, msg.target)
		}
		return nil
	})

	if err := client.UnlistenRemote("127.0.0.1:19001", 3*time.Second); err != nil {
		t.Fatalf("unlisten remote: %v", err)
	}
}

// TestCtlRequestTimeout：对端应用层不回应（不消费 Ctl 或处理挂死），
// 请求必须超时返回，而不是永远挂着。
func TestCtlRequestTimeout(t *testing.T) {
	client, _ := newCtlPair(t) // 服务端故意不接控制面

	if err := client.ListenRemote("127.0.0.1:19001", "127.0.0.1:8000", 50*time.Millisecond); err == nil {
		t.Fatalf("expected timeout error, got nil")
	}
}

// TestCtlStaleResultCleared：上一次超时后的迟到回应，
// 不能被下一次请求冒充成自己的结果。
func TestCtlStaleResultCleared(t *testing.T) {
	client, server := newCtlPair(t)

	var calls atomic.Int32
	serveCtl(t, server, func(ctlMsg) error {
		if calls.Add(1) == 1 {
			time.Sleep(150 * time.Millisecond) // 第一条故意迟到
			return errors.New("late failure")  // 这个错误不该被第二次请求看到
		}
		return nil
	})

	// 第一次：超时收场
	if err := client.ListenRemote("127.0.0.1:19001", "127.0.0.1:8000", 30*time.Millisecond); err == nil {
		t.Fatalf("first request should timeout")
	}
	// 第二次：必须拿到自己的 ok，而不是残留的 late failure
	if err := client.ListenRemote("127.0.0.1:19002", "127.0.0.1:8000", 3*time.Second); err != nil {
		t.Fatalf("second request got stale result: %v", err)
	}
}
