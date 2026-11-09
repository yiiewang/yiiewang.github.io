package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

// admin.go 是 v5 的运行时控制面：声明式注册表 + 本地管理接口。
//
// 设计立场：客户端是反向转发唯一的事实源。服务端的监听只是"会话资产"，
// 会话一死就清零；注册表在客户端，每次（重）连成功后整体重放。
// 于是"运行时注册/注销"就是改注册表 + 尽力应用到当前会话——不需要
// 任何持久化协议，也不怕两端状态漂移：漂移了就整体重放一遍。

// reverseRetryDelay 是注册重放失败后的补试间隔。最常见的失败是撞上
// 旧会话的监听还没释放——会话死亡的清理和重连的注册在赛跑。
const reverseRetryDelay = 200 * time.Millisecond

// reverseRegistry 是反向转发的声明式清单。
// 静态 -reverse 提供初值，admin 接口运行时增删，dial 成功后整体重放。
type reverseRegistry struct {
	mu       sync.Mutex
	mappings []reverseMapping
}

func newReverseRegistry(initial []reverseMapping) *reverseRegistry {
	return &reverseRegistry{mappings: append([]reverseMapping(nil), initial...)}
}

// snapshot 返回当前清单的副本——重放和 LIST 都靠它，不暴露内部切片。
func (r *reverseRegistry) snapshot() []reverseMapping {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]reverseMapping(nil), r.mappings...)
}

func (r *reverseRegistry) containsLocked(bind string) bool {
	for _, m := range r.mappings {
		if m.bind == bind {
			return true
		}
	}
	return false
}

func (r *reverseRegistry) removeLocked(bind string) bool {
	for i, m := range r.mappings {
		if m.bind == bind {
			r.mappings = append(r.mappings[:i], r.mappings[i+1:]...)
			return true
		}
	}
	return false
}

// register 把映射加进清单并尽力立即应用到当前会话。
//
// 会话不在（掉线/重连中）就先记账，下一次 dial 成功时重放；
// 立刻失败（如端口被占）则回滚——不留期望与现实不符的条目。
func (r *reverseRegistry) register(bind, target string, tun *tunnel) []string {
	r.mu.Lock()
	if r.containsLocked(bind) {
		r.mu.Unlock()
		return []string{fmt.Sprintf("ERR bind already registered: %s", bind)}
	}
	r.mappings = append(r.mappings, reverseMapping{bind: bind, target: target})
	r.mu.Unlock()

	sess := tun.current()
	select {
	case <-sess.Done():
		return []string{"OK registered (session down, applied on next connect)"}
	default:
	}
	if err := sess.ListenRemote(bind, target, ctlTimeout); err != nil {
		// 刚好撞上换会话的窗口：按掉线处理，交给重放
		select {
		case <-sess.Done():
			return []string{"OK registered (session down, applied on next connect)"}
		default:
		}
		r.mu.Lock()
		r.removeLocked(bind)
		r.mu.Unlock()
		return []string{"ERR " + err.Error()}
	}
	return []string{"OK listening"}
}

// unregister 从清单摘除，并尽力撤掉服务端上的监听。
//
// 服务端没有这个监听不算错——清单才是事实源；真错误只记日志，
// 那个监听最迟随会话死亡一起消失。
func (r *reverseRegistry) unregister(bind string, tun *tunnel) []string {
	r.mu.Lock()
	found := r.removeLocked(bind)
	r.mu.Unlock()
	if !found {
		return []string{"ERR no such bind: " + bind}
	}

	sess := tun.current()
	select {
	case <-sess.Done():
		return []string{"OK removed (session down, server listener dies with it)"}
	default:
	}
	if err := sess.UnlistenRemote(bind, ctlTimeout); err != nil {
		log.Printf("reverse unlisten, bind=%s, err=%v", bind, err)
	}
	return []string{"OK removed"}
}

func (r *reverseRegistry) list() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.mappings) == 0 {
		return []string{"OK 0 mapping"}
	}
	out := []string{fmt.Sprintf("OK %d mapping(s)", len(r.mappings))}
	for _, m := range r.mappings {
		out = append(out, fmt.Sprintf("%s -> %s", m.bind, m.target))
	}
	return out
}

// bootSession 把一条裸连接组装成完整的客户端会话：
// 逼出裁决 → 心跳 → 反向数据面 → 反向注册重放。
// dial 只是它加一层 TLS；测试用明文连接走同一段逻辑。
func bootSession(conn net.Conn, reg *reverseRegistry, allow *AllowList) (*Session, error) {
	sess := newSession(conn, true)

	// 逼出服务端对客户端证书的裁决：TLS 1.3 下 tls.Dial 成功不代表
	// 认证通过，服务端拒收时它已经把连接关了。拿不到 PONG 就算拨号
	// 失败，交给调用方按退避重试，而不是等第一条流失败才暴露。
	if err := sess.Handshake(handshakeTimeout); err != nil {
		_ = conn.Close()
		return nil, err
	}

	// 心跳必须排在 Handshake 之后（KeepAlive 注释里写了为什么）。
	go sess.KeepAlive(keepAliveInterval, keepAliveTimeout)

	// 反向数据面：对端（服务端）开来的流，走与正向完全相同的
	// 白名单 → dial → 桥接——serveSession 两端共用同一段代码。
	go serveSession(sess, allow)

	// 反向注册：控制面状态是 per-session 的，每次（重）连成功后
	// 整体重放声明式清单。失败补试一把（多半是撞上旧会话的清理），
	// 再失败只记日志，不拖垮正向转发。
	for _, m := range reg.snapshot() {
		err := sess.ListenRemote(m.bind, m.target, ctlTimeout)
		if err != nil {
			time.Sleep(reverseRetryDelay)
			err = sess.ListenRemote(m.bind, m.target, ctlTimeout)
		}
		if err != nil {
			log.Printf("reverse listen failed, bind=%s, target=%s, err=%v", m.bind, m.target, err)
		} else {
			log.Printf("reverse listening, bind=%s, target=%s", m.bind, m.target)
		}
	}
	return sess, nil
}

// serveAdmin 是本地管理面：行协议，一次一行，逐行回应。
// 只该绑回环——能下 LISTEN 命令就等于拿到了本机的拨号权。
func serveAdmin(ln net.Listener, reg *reverseRegistry, tun *tunnel) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		go handleAdminConn(conn, reg, tun)
	}
}

func handleAdminConn(conn net.Conn, reg *reverseRegistry, tun *tunnel) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		for _, line := range execAdmin(scanner.Text(), reg, tun) {
			fmt.Fprintf(conn, "%s\n", line)
		}
	}
}

// execAdmin 解析并执行一条管理命令，返回回应行。
func execAdmin(line string, reg *reverseRegistry, tun *tunnel) []string {
	fields := strings.Fields(line)
	switch {
	case len(fields) == 3 && fields[0] == "LISTEN":
		return reg.register(fields[1], fields[2], tun)
	case len(fields) == 2 && fields[0] == "UNLISTEN":
		return reg.unregister(fields[1], tun)
	case len(fields) == 1 && fields[0] == "LIST":
		return reg.list()
	default:
		return []string{"ERR usage: LISTEN <bind> <target> | UNLISTEN <bind> | LIST"}
	}
}
