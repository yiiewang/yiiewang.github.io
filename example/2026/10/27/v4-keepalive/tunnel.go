package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// v4 的第二块：会话死了自动换一条（P0-2），换不出来就退出进程（P0-3）。
//
// Session 本身不会复活——一条 TLS 连接断了就是断了，唯一的恢复方式是
// 重拨、重跑握手、重建会话。tunnel 持有"当前会话"：转发方永远从
// current() 取最新，切换对它们透明；旧会话上还在跑的流由 Session.Close
// 收尾，新流走新会话。

const (
	// initialBackoff 是会话死亡后第一次重连失败后的等待。
	initialBackoff = time.Second
	// maxBackoff 封顶长故障下的重试节奏，别把对端打爆。
	maxBackoff = 30 * time.Second
	// maxReconnectAttempts 连续失败到这个数就放弃（P0-3）：
	// 与其在进程内静默装死，不如非零退出，把"拉起"交给 systemd——
	// 它会重读配置重跑首连，比内循环重试更干净。
	maxReconnectAttempts = 8
)

// tunnel 持有当前会话并负责死亡后的重建。
type tunnel struct {
	mu   sync.Mutex
	sess *Session
}

// newTunnel 建立首个会话。失败直接返回错误：启动期的问题（证书错、地址
// 错、对端不在线）要立刻暴露给调用方，不在这一层静默重试。
func newTunnel(dial func() (*Session, error)) (*tunnel, error) {
	sess, err := dial()
	if err != nil {
		return nil, err
	}
	return &tunnel{sess: sess}, nil
}

// current 返回当前会话。锁内只做指针读取，不会阻塞转发方。
func (t *tunnel) current() *Session {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.sess
}

func (t *tunnel) set(sess *Session) {
	t.mu.Lock()
	t.sess = sess
	t.mu.Unlock()
}

// reconnectLoop 监听会话死亡并重建，直到成功或连续失败到上限。
//
// 节奏：死亡后立刻试第一把（服务端进程重启这种短故障能秒级恢复），
// 失败再按 initialBackoff 指数退避、封顶 maxBackoff。
// 放弃时返回错误而不是直接 os.Exit：退出动作留给调用方，进程边界
// （退出码、日志措辞）由 main 决定，这里保持可测试。
//
// initial 做成入参是为了测试能用毫秒节奏跑完"连续失败到放弃"的整条路径。
func (t *tunnel) reconnectLoop(dial func() (*Session, error), initial time.Duration) error {
	for {
		sess := t.current()
		<-sess.Done() // 等当前会话死：对端断开、心跳判死、主动 Close 都会响
		// Close 幂等且保留首次原因：这里取回死因只用于日志
		cause := sess.Close(nil)
		log.Printf("tunnel session dead, reconnecting, cause=%v", cause)

		backoff := initial
		for attempt := 1; ; attempt++ {
			sess, err := dial()
			if err == nil {
				t.set(sess)
				log.Printf("tunnel reconnected, attempt=%d", attempt)
				break
			}
			if attempt >= maxReconnectAttempts {
				return fmt.Errorf("reconnect failed after %d attempts: %w", attempt, err)
			}
			log.Printf("tunnel reconnect failed, attempt=%d, backoff=%s, err=%v", attempt, backoff, err)
			time.Sleep(backoff)
			backoff = min(backoff*2, maxBackoff)
		}
	}
}
