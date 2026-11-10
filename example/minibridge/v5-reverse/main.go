// Package main 是 minibridge 第三阶段：给隧道上锁。
//
// 在 v2 的多路复用之上加三件事：
//  1. TLS：加密 + 服务端身份（客户端必须校验证书，绝不能 InsecureSkipVerify）
//  2. 认证：mTLS 客户端证书，握手内完成，不额外发明协议
//  3. 授权：目标白名单，治 SSRF
//
// 顺序是铁律：TLS 握手（含客户端证书校验）→ 才接受 SYN → 才 dial 目标。
// 第二层之所以不做成应用层协议，是因为 TLS 已经把这件事做对了：
// 校验失败根本进不到应用层，省掉一整个"认证没过就别 dial"的窗口。
package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "server":
		runServer(os.Args[2:])
	case "client":
		runClient(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `双向 TLS：两端各持有自己的证书，各校验对方一次。
注意 -cert / -key 同名不同物——服务端那是服务端的，客户端那是客户端的。

服务端必填：
  -cert        服务端证书 PEM（客户端拿 -ca 校验它）
  -key         服务端私钥 PEM
  -client-ca   校验客户端证书的 CA PEM

服务端可选：
  -max-conns   并发隧道连接上限（默认 64），超出直接关闭——防资源耗尽
  -revoked     吊销的客户端证书 serial 清单文件，内容变化自动重载——
               泄漏单张证书不再需要换 CA 重签全链

客户端证书可绑定来源 IP（make certs CLI_IP=...）：整套证书被拷到
别的机器、换个 IP 连上来，服务端在握手后直接断开。

客户端必填：
  -bind        端口映射 [地址:]端口:目标:端口（逗号分隔多条），SSH -L/-R 风格：
               正向=本端监听、服务端拨目标；反向=服务端监听、本端拨目标
  -ca          校验服务端证书的 CA PEM
  -cert        客户端证书 PEM（服务端拿 -client-ca 校验它）
  -key         客户端私钥 PEM

客户端可选：
  -reverse     反向模式开关：bind 在服务端侧、target 在本端
  -allow       反向流的拨号白名单，默认只放行回环——反向转发意味着
               服务端能驱动本端 dial，这道闸门与服务端的 -allow 对称
  -admin       本地管理面（LISTEN <bind> <target> / UNLISTEN <bind> / LIST），
               运行时增删反向转发不用重启；只该绑回环
  -revoked     吊销的服务端证书 serial 清单（对称于服务端 -revoked）

  server -addr :9000 -cert srv.crt -key srv.key -client-ca ca.crt
  # 正向：本机 :8080 == 服务端视角的 127.0.0.1:8000
  client -bind 8080:127.0.0.1:8000 -server host:9000 \
         -ca ca.crt -cert cli.crt -key cli.key
  # 反向：服务端 :18000 == 本端视角的 127.0.0.1:8000
  client -reverse -bind :18000:127.0.0.1:8000 -server host:9000 \
         -ca ca.crt -cert cli.crt -key cli.key
`)
}

// runServer 启动带 TLS 的隧道服务端。
func runServer(args []string) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	addr := fs.String("addr", "127.0.0.1:9000", "隧道监听地址")
	certFile := fs.String("cert", "", "服务端证书 PEM（必填）")
	keyFile := fs.String("key", "", "服务端私钥 PEM（必填）")
	clientCAFile := fs.String("client-ca", "", "校验客户端证书的 CA PEM（必填）")
	allowSpec := fs.String("allow", "127.0.0.1:*", "允许拨号的目标，逗号分隔\n    可选，默认只放行回环")
	maxConns := fs.Int("max-conns", 64, "并发隧道连接上限（DoS 缓解），超出直接关闭")
	revokedFile := fs.String("revoked", "", "吊销的客户端证书 serial 清单文件（每行一个；\n    0x 前缀=十六进制，纯数字=十进制，# 注释）。内容变化自动重载，无需重启")
	_ = fs.Parse(args)

	if *certFile == "" || *keyFile == "" || *clientCAFile == "" {
		log.Fatalf("server requires -cert, -key and -client-ca")
	}

	cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		log.Fatalf("load server cert failed, err=%v", err)
	}

	clientCA, err := os.ReadFile(*clientCAFile)
	if err != nil {
		log.Fatalf("read client ca failed, err=%v", err)
	}

	allow, err := NewAllowList(strings.Split(*allowSpec, ","))
	if err != nil {
		log.Fatalf("parse allow list failed, err=%v", err)
	}

	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		// 这两行就是全部的客户端认证。少了它们，谁都能连。
		ClientCAs:  loadCertPool(clientCA),
		ClientAuth: tls.RequireAndVerifyClientCert,
		MinVersion: tls.VersionTLS13,
		// 单张吊销：链校验之后追加 serial 裁决，被吊销的客户端证书握手即拒
		VerifyPeerCertificate: verifyPeer(NewRevocationList(*revokedFile)),
	}

	ln, err := tls.Listen("tcp", *addr, cfg)
	if err != nil {
		log.Fatalf("listen tunnel failed, addr=%s, err=%v", *addr, err)
	}
	log.Printf("crypto tunnel server listening, addr=%s, auth=mtls", *addr)

	serveCrypto(ln, allow, *maxConns)
}

// runClient 启动转发客户端。正反向共用 -bind（SSH -L/-R 风格），-reverse 只是方向开关：
//
//	正向：-bind 90:127.0.0.1:80   本端监听 :90，服务端拨 127.0.0.1:80
//	反向：-reverse -bind 90:127.0.0.1:80   服务端监听 :90，本端拨 127.0.0.1:80
//
// 数据面两边走的是同一段代码——Session 本来就是对称的（v2 埋的），
// 区别只在谁 Open、谁 Accept。
func runClient(args []string) {
	fs := flag.NewFlagSet("client", flag.ExitOnError)
	server := fs.String("server", "127.0.0.1:9000", "隧道服务端地址")
	bind := fs.String("bind", "", "端口映射 [地址:]端口:目标:端口（必填，逗号分隔多条）：\n    正向=本端监听，服务端拨目标；反向=服务端监听，本端拨目标")
	reverse := fs.Bool("reverse", false, "反向模式：bind 在服务端侧、target 在本端")
	caFile := fs.String("ca", "", "校验服务端证书的 CA PEM（必填）")
	certFile := fs.String("cert", "", "客户端证书 PEM（必填）")
	keyFile := fs.String("key", "", "客户端私钥 PEM（必填）")
	serverName := fs.String("server-name", "", "TLS 校验用的服务名，默认取 -server 的主机部分")
	adminAddr := fs.String("admin", "", "本地管理面地址（LISTEN/UNLISTEN/LIST 行协议），\n    运行时增删反向转发不用重启；只该绑回环")
	clientAllowSpec := fs.String("allow", "127.0.0.1:*", "反向流的拨号白名单，逗号分隔\n    默认只放行回环——反向意味着服务端能驱动本端 dial")
	revokedFile := fs.String("revoked", "", "吊销的服务端证书 serial 清单文件（格式同 server -revoked），\n    对称校验：客户端同样可以单张吊销对端")
	_ = fs.Parse(args)

	if *bind == "" || *caFile == "" || *certFile == "" || *keyFile == "" {
		log.Fatalf("client requires -bind, -ca, -cert and -key")
	}

	pemData, err := os.ReadFile(*caFile)
	if err != nil {
		log.Fatalf("read ca failed, err=%v", err)
	}
	name := *serverName
	if name == "" {
		if host, _, err := net.SplitHostPort(*server); err == nil {
			name = host
		}
	}

	pair, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		log.Fatalf("load client cert failed, err=%v", err)
	}

	cfg := &tls.Config{
		RootCAs:      loadCertPool(pemData),
		ServerName:   name,
		Certificates: []tls.Certificate{pair},
		MinVersion:   tls.VersionTLS13,
		// 单张吊销：对称于服务端，客户端可吊销指定的服务端证书
		VerifyPeerCertificate: verifyPeer(NewRevocationList(*revokedFile)),
	}

	// 端口映射清单：正反向都从 -bind 解析；反向模式下它注册到服务端并
	// 按会话重放——服务端不持久化任何注册状态。正向模式的清单只用于
	// 本地监听（mappings 不进注册表，admin 的 LISTEN 仍可运行时补）。
	mappings, err := parseBindSpec(*bind)
	if err != nil {
		log.Fatalf("parse bind spec failed, err=%v", err)
	}
	reg := newReverseRegistry(nil)
	if *reverse {
		reg = newReverseRegistry(mappings)
	}
	clientAllow, err := NewAllowList(strings.Split(*clientAllowSpec, ","))
	if err != nil {
		log.Fatalf("parse client allow list failed, err=%v", err)
	}

	// dial 建立一条全新会话：首连和重连共用同一套动作，差别只是谁在调它、
	// 失败后由谁处置（首连 Fatal，重连按退避重试）。
	// DialWithDialer 给 TCP 连接和 TLS 握手都挂上 dialTimeout——否则对
	// 一个被防火墙黑洞的地址，重连尝试会无限挂着，退避形同虚设。
	dial := func() (*Session, error) {
		dialer := &net.Dialer{Timeout: dialTimeout}
		conn, err := tls.DialWithDialer(dialer, "tcp", *server, cfg)
		if err != nil {
			return nil, err
		}
		return bootSession(conn, reg, clientAllow)
	}

	tun, err := newTunnel(dial)
	if err != nil {
		log.Fatalf("dial tunnel server failed, server=%s, err=%v", *server, err)
	}
	// 重连循环：会话死亡自动换新；连续失败到上限就非零退出（P0-3），
	// 把"拉起"交给 systemd。
	go func() {
		if err := tun.reconnectLoop(dial, initialBackoff); err != nil {
			log.Fatalf("tunnel gave up reconnecting, exit for supervisor: %v", err)
		}
	}()
	log.Printf("crypto tunnel ready, server=%s, mode=%s, auth=mtls", *server,
		map[bool]string{true: "reverse", false: "forward"}[*reverse])

	// 管理面：运行时增删反向转发。能下 LISTEN 就等于拿到了本机的拨号权，
	// 所以这个地址只该绑回环，绝不该出现在任何监听 0.0.0.0 的配置里。
	// 校验是硬的：非回环直接拒绝启动——这类高危配置不该给"跑起来再说"的机会。
	if *adminAddr != "" {
		if !adminAddrIsLoopback(*adminAddr) {
			log.Fatalf("admin must bind loopback (e.g. 127.0.0.1:9090), got %q: LISTEN equals dial power over this machine", *adminAddr)
		}
		aln, err := net.Listen("tcp", *adminAddr)
		if err != nil {
			log.Fatalf("listen admin failed, addr=%s, err=%v", *adminAddr, err)
		}
		log.Printf("admin listening, addr=%s", *adminAddr)
		go serveAdmin(aln, reg, tun)
	}

	// 反向模式：数据面全在服务端（bind）与本端 Accept 之间，本端没有
	// 用户监听。正向模式：每条映射在本端起一个监听，accept 后开流让
	// 服务端拨号。两种模式的主 goroutine 都只剩"活着"——重连循环已在后台。
	if *reverse {
		select {}
	}

	for _, m := range mappings {
		ln, err := net.Listen("tcp", m.bind)
		if err != nil {
			log.Fatalf("listen local failed, addr=%s, err=%v", m.bind, err)
		}
		log.Printf("crypto forward listening, bind=%s, target=%s", m.bind, m.target)
		go func(m portMapping, ln net.Listener) {
			for {
				local, err := ln.Accept()
				if err != nil {
					log.Printf("accept failed, bind=%s, err=%v", m.bind, err)
					return
				}
				go func(local net.Conn, target string) {
					// 每条新流都取当前会话：重连刚切完的瞬间可能拿到已死的旧会话，
					// Open 会以 ErrSessionClosed 拒绝这条 local，下一条自然走新会话。
					st, err := tun.current().Open(target)
					if err != nil {
						log.Printf("open stream failed, target=%s, err=%v", target, err)
						_ = local.Close()
						return
					}
					in, out, berr := bridge(local, st)
					log.Printf("forward finished, id=%d, target=%s, local_to_remote=%d, remote_to_local=%d, err=%v",
						st.ID(), target, in, out, berr)
				}(local, m.target)
			}
		}(m, ln)
	}
	select {}
}

// serveCrypto 接受连接并起会话。
//
// 这里没有"认证"这一步，是因为它已经在 TLS 握手里做完了：能 Accept 到
// 的连接必然持有 -client-ca 签发的证书，校验失败的连接被 crypto/tls
// 直接丢掉，根本到不了应用层。这正是 mTLS 相对应用层认证的优势——
// 不存在"认证没过但已经能驱动服务端 dial"的窗口。
//
// 客户端证书还可以绑定来源 IP（make certs CLI_IP=...）：整套证书+私钥
// 被拷到别的机器、换个 IP 连上来，verifyClientSourceIP 会直接断开。
// 校验放在 goroutine 里做，单个恶意连接的慢握手不能卡死 accept 循环。
//
// maxConns 是并发会话上限（信号量）：满了直接关连接、连 TLS 握手都不做。
// 这不是完整的 DoS 防御（SYN 洪水归内核/防火墙管），只是把"每连接
// goroutine + 心跳"的资源放大倍数封顶，让应用层耗不垮。
func serveCrypto(ln net.Listener, allow *AllowList, maxConns int) {
	slots := make(chan struct{}, maxConns)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept failed, err=%v", err)
			return
		}
		select {
		case slots <- struct{}{}:
		default:
			log.Printf("tunnel conn refused, remote=%s, conns=%d (limit)", conn.RemoteAddr(), maxConns)
			_ = conn.Close()
			continue
		}
		go func(conn net.Conn) {
			defer func() { <-slots }()
			if err := verifyClientSourceIP(conn); err != nil {
				log.Printf("client ip denied, remote=%s, err=%v", conn.RemoteAddr(), err)
				_ = conn.Close()
				return
			}
			sess := newSession(conn, false)
			// 服务端同样发心跳：会话死了不能等客户端来告诉你。
			// 双端互发互回，KeepAlive 判死即 Close，会话死因由 serveSession 记日志。
			go sess.KeepAlive(keepAliveInterval, keepAliveTimeout)
			go serveSession(sess, allow)
			// 反向转发的控制面：消费这条会话的 LISTEN/UNLISTEN。
			// 会话死亡时，它名下的反向监听由 manager 自己关掉。
			go newReverseManager(sess).run()
		}(conn)
	}
}

// verifyClientSourceIP 比对客户端证书 SAN 里的 IP 绑定与实际来源地址。
//
// mTLS 的标准校验只回答"证书是不是 CA 签的"，不看连接从哪来——IP 绑定
// 是它之外的补充闸门。tls.Listen 返回的连接默认惰性握手，必须先显式
// Handshake 才拿得到对端证书。证书没写 IP SAN 的不受限，老证书不受影响。
func verifyClientSourceIP(conn net.Conn) error {
	tc, ok := conn.(*tls.Conn)
	if !ok {
		return fmt.Errorf("not a tls connection")
	}
	if err := tc.Handshake(); err != nil {
		return fmt.Errorf("handshake: %w", err)
	}
	certs := tc.ConnectionState().PeerCertificates
	if len(certs) == 0 || len(certs[0].IPAddresses) == 0 {
		return nil
	}
	raddr, ok := conn.RemoteAddr().(*net.TCPAddr)
	if !ok {
		return fmt.Errorf("non-tcp remote address %T", conn.RemoteAddr())
	}
	for _, bound := range certs[0].IPAddresses {
		// Equal 把 IPv4-mapped IPv6 与 IPv4 视为同一地址，两侧格式差异不必担心
		if bound.Equal(raddr.IP) {
			return nil
		}
	}
	return fmt.Errorf("cert is bound to %v", certs[0].IPAddresses)
}

// serveSession 处理一条隧道连接上的所有流。
func serveSession(sess *Session, allow *AllowList) {
	for {
		st, err := sess.Accept()
		if err != nil {
			log.Printf("session closed, err=%v", err)
			return
		}
		go handleStream(st, allow)
	}
}

// handleStream 处理一条流：先过白名单，再 dial 目标，最后桥接。
func handleStream(st *Stream, allow *AllowList) {
	// 白名单在 dial 之前：不允许的目标一个包都不往外发。
	target, ok := allow.Resolve(st.Target())
	if !ok {
		log.Printf("target denied, id=%d, target=%s", st.ID(), st.Target())
		_ = st.Reset()
		return
	}

	upstream, err := net.DialTimeout("tcp", target, dialTimeout)
	if err != nil {
		log.Printf("dial target failed, id=%d, target=%s, err=%v", st.ID(), target, err)
		_ = st.Reset()
		return
	}

	if err := st.AcceptOK(); err != nil {
		_ = upstream.Close()
		log.Printf("accept stream failed, id=%d, err=%v", st.ID(), err)
		return
	}

	in, out, berr := bridge(st, upstream)
	log.Printf("stream finished, id=%d, target=%s, client_to_target=%d, target_to_client=%d, err=%v",
		st.ID(), target, in, out, berr)
}

func loadCertPool(pemData []byte) *x509.CertPool {
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pemData) {
		log.Fatalf("no valid certificate found in CA file")
	}
	return pool
}

// adminAddrIsLoopback 校验管理面地址只绑回环。
// 必须显式写全地址（127.0.0.1:9090）：留空 host 的 ":9090" 监听所有接口，同样拒绝。
func adminAddrIsLoopback(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
