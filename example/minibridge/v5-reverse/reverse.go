package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

// reverse.go 是 v5 的应用层：反向转发。
//
// 数据面零新协议——Session 本来就是对称的（v2 埋的）：
// 服务端在 bind 上 accept 用户连接，sess.Open(target) 把流开回客户端，
// 客户端 Accept 后按 target 拨号再桥接。用的正是正向转发那同一段代码。
//
// 控制面三个帧（LISTEN/UNLISTEN/CTLRESULT）负责注册与注销：
// 客户端持声明式的映射清单，每次（重）连成功后整体重放——控制面状态
// 是 per-session 的，会话一死，服务端就把它名下的监听全部关掉。

// portMapping 是一条端口映射的声明：bind 监听用户接入，target 是流量的最终去向。
// 正反向共用这一对概念，-reverse 只决定两端谁当 bind、谁当 target：
//
//	正向：bind 在本端 net.Listen，target 由服务端拨（服务端视角）
//	反向：bind 注册到服务端，target 由本端拨（本端视角）
type portMapping struct {
	bind   string // 用户接入的监听地址
	target string // 回程目标（拨号方视角）
}

// parseBindSpec 解析 -bind 的值，SSH -L/-R 风格，逗号分隔多条：
//
//	[bind地址:]端口:目标host:目标端口
//	90:127.0.0.1:80          本端（正向）或服务端（反向）监听 :90，回程 127.0.0.1:80
//	:90:127.0.0.1:80         同上，显式给出 bind 地址（空 = 所有接口）
//
// 3 段式省略 bind 地址（等价 ":port"），4 段式的第一段可以是 IP 或留空。
// IPv6 不支持——bind/target 的冒号会歧义，demo 不为它增加复杂度。
func parseBindSpec(spec string) ([]portMapping, error) {
	if spec == "" {
		return nil, fmt.Errorf("empty bind spec")
	}
	var out []portMapping
	for _, item := range strings.Split(spec, ",") {
		parts := strings.Split(item, ":")
		var bind string
		switch len(parts) {
		case 3: // port:host:hostport
			bind = ":" + parts[0]
		case 4: // bindaddr:port:host:hostport
			bind = parts[0] + ":" + parts[1]
		default:
			return nil, fmt.Errorf("malformed bind item %q, want [addr:]port:host:port", item)
		}
		target := parts[len(parts)-2] + ":" + parts[len(parts)-1]
		if parts[len(parts)-2] == "" || parts[len(parts)-1] == "" || parts[0] == "" && len(parts) == 3 {
			return nil, fmt.Errorf("malformed bind item %q", item)
		}
		out = append(out, portMapping{bind: bind, target: target})
	}
	return out, nil
}

// reverseListener 是服务端上的一个反向监听。
type reverseListener struct {
	ln     net.Listener
	target string // 回程目标（客户端视角）
}

// reverseManager 管理一条会话名下的全部反向监听。
//
// 生命周期：run() 消费会话的控制请求；会话死亡时把它名下的监听
// 全部关掉——否则端口泄漏，而且谁也说不清那个端口还通向哪。
type reverseManager struct {
	sess      *Session
	mu        sync.Mutex
	listeners map[string]*reverseListener
}

func newReverseManager(sess *Session) *reverseManager {
	return &reverseManager{
		sess:      sess,
		listeners: make(map[string]*reverseListener),
	}
}

// run 是反向监听的控制循环：收一条、办一条、回一条。
// 一个 bind 只属于一个会话，后来者会拿到 address already in use。
func (rm *reverseManager) run() {
	go func() {
		<-rm.sess.Done()
		rm.closeAll()
	}()

	for {
		select {
		case msg := <-rm.sess.Ctl():
			var err error
			if msg.unlisten {
				err = rm.unlisten(msg.bind)
			} else {
				err = rm.listen(msg.bind, msg.target)
			}
			rm.sess.ReplyCtl(msg, err)
		case <-rm.sess.Done():
			return
		}
	}
}

// listen 在 bind 上建监听。同一 bind 重复注册按错误回应——
// 控制请求在 run() 里是串行处理的，不存在并发插队。
func (rm *reverseManager) listen(bind, target string) error {
	rm.mu.Lock()
	_, dup := rm.listeners[bind]
	rm.mu.Unlock()
	if dup {
		return fmt.Errorf("reverse listener already exists: %s", bind)
	}

	ln, err := net.Listen("tcp", bind)
	if err != nil {
		return err
	}

	rm.mu.Lock()
	rm.listeners[bind] = &reverseListener{ln: ln, target: target}
	rm.mu.Unlock()

	log.Printf("reverse listening, bind=%s, target=%s", bind, target)
	go rm.acceptLoop(bind, ln, target)
	return nil
}

// unlisten 注销 bind 上的监听。Close 会唤醒 acceptLoop，让它自然退出。
func (rm *reverseManager) unlisten(bind string) error {
	rm.mu.Lock()
	rl, ok := rm.listeners[bind]
	delete(rm.listeners, bind)
	rm.mu.Unlock()
	if !ok {
		return fmt.Errorf("no reverse listener on %s", bind)
	}
	return rl.ln.Close()
}

// acceptLoop 收用户连接。监听被关（UNLISTEN 或会话死亡）时 Accept 报错，直接退。
func (rm *reverseManager) acceptLoop(bind string, ln net.Listener, target string) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		go rm.bridge(bind, conn, target)
	}
}

// bridge 把一条用户连接经反向流接到客户端身边的目标。
//
// 与正向唯一的区别是方向：这里由服务端 Open、客户端 Accept——
// 帧和语义与 v2 完全相同，对称性在这里兑现。
func (rm *reverseManager) bridge(bind string, conn net.Conn, target string) {
	st, err := rm.sess.Open(target)
	if err != nil {
		log.Printf("reverse open failed, bind=%s, target=%s, err=%v", bind, target, err)
		_ = conn.Close()
		return
	}
	log.Printf("reverse accepted, bind=%s, remote=%s, id=%d", bind, conn.RemoteAddr(), st.ID())

	in, out, berr := bridge(conn, st)
	log.Printf("reverse finished, bind=%s, target=%s, user_to_target=%d, target_to_user=%d, err=%v",
		bind, target, in, out, berr)
}

// closeAll 关掉本会话名下的全部反向监听（会话死亡时调用）。
// 拿锁 → 换走整张表 → 放锁 → 再关：不持锁做 I/O。
func (rm *reverseManager) closeAll() {
	rm.mu.Lock()
	listeners := rm.listeners
	rm.listeners = make(map[string]*reverseListener)
	rm.mu.Unlock()
	for _, rl := range listeners {
		_ = rl.ln.Close()
	}
}
