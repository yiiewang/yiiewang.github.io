package main

import (
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

// pipeDial 用 net.Pipe 伪造"拨号"：客户端侧会话交给 tunnel，服务端侧
// 由测试持有——关掉它就等于 kill 服务端进程。不引真 TLS，重连逻辑本身
// 与传输层无关。
func pipeDial(t *testing.T) (dial func() (*Session, error), kill func() *Session) {
	t.Helper()
	var server atomic.Pointer[Session]
	dial = func() (*Session, error) {
		a, b := net.Pipe()
		client := newSession(a, true)
		server.Store(newSession(b, false))
		if err := client.Handshake(3 * time.Second); err != nil {
			return nil, err
		}
		go client.KeepAlive(5*time.Millisecond, 200*time.Millisecond)
		return client, nil
	}
	kill = func() *Session { return server.Load() }
	return dial, kill
}

// TestTunnelReconnectsAfterDeath 对应 P0-2 验收：服务端死了，
// 客户端必须自动换上一条新会话，不需要任何人工干预。
func TestTunnelReconnectsAfterDeath(t *testing.T) {
	dial, kill := pipeDial(t)

	tun, err := newTunnel(dial)
	if err != nil {
		t.Fatalf("first dial: %v", err)
	}
	go tun.reconnectLoop(dial, time.Millisecond)

	old := tun.current()
	_ = kill().Close(errors.New("server killed")) // 模拟 kill 服务端进程

	// 等待换新会话；毫秒级退避，3 秒绰绰有余
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if cur := tun.current(); cur != old {
			// 新会话必须真的活着，而不是刚建好又死了
			select {
			case <-cur.Done():
				t.Fatalf("new session died immediately")
			default:
				return
			}
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("tunnel did not reconnect within 3s")
}

// TestTunnelGivesUpAfterMaxAttempts 对应 P0-3 验收：对端一直不回来，
// 重连到 maxReconnectAttempts 就放弃并返回错误（main 层转成非零退出）。
// 毫秒级初始退避：整条放弃路径在几百毫秒内跑完。
func TestTunnelGivesUpAfterMaxAttempts(t *testing.T) {
	a, b := net.Pipe()
	first := newSession(a, true)
	tun := &tunnel{sess: first}
	_ = b.Close() // 首个会话立刻死亡，触发重连循环

	select {
	case <-first.Done():
	case <-time.After(time.Second):
		t.Fatalf("first session did not die")
	}

	calls := 0
	dial := func() (*Session, error) {
		calls++
		return nil, errors.New("unreachable")
	}

	err := tun.reconnectLoop(dial, time.Millisecond)
	if err == nil {
		t.Fatalf("expected give-up error, got nil")
	}
	if calls != maxReconnectAttempts {
		t.Fatalf("dial calls: want %d, got %d", maxReconnectAttempts, calls)
	}
}
