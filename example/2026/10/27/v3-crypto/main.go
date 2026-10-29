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

客户端必填：
  -ca          校验服务端证书的 CA PEM
  -cert        客户端证书 PEM（服务端拿 -client-ca 校验它）
  -key         客户端私钥 PEM
  -target      目标地址 host:port（服务端视角）

  server -addr :9000 -cert srv.crt -key srv.key -client-ca ca.crt
  client -listen :8080 -server host:9000 -target 127.0.0.1:8000 \
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
	}

	ln, err := tls.Listen("tcp", *addr, cfg)
	if err != nil {
		log.Fatalf("listen tunnel failed, addr=%s, err=%v", *addr, err)
	}
	log.Printf("crypto tunnel server listening, addr=%s, auth=mtls", *addr)

	serveCrypto(ln, allow)
}

// runClient 启动带 TLS 的转发客户端。
func runClient(args []string) {
	fs := flag.NewFlagSet("client", flag.ExitOnError)
	listen := fs.String("listen", ":8080", "本地监听地址")
	server := fs.String("server", "127.0.0.1:9000", "隧道服务端地址")
	target := fs.String("target", "", "目标地址 host:port（必填，服务端视角）")
	caFile := fs.String("ca", "", "校验服务端证书的 CA PEM（必填）")
	certFile := fs.String("cert", "", "客户端证书 PEM（必填）")
	keyFile := fs.String("key", "", "客户端私钥 PEM（必填）")
	serverName := fs.String("server-name", "", "TLS 校验用的服务名，默认取 -server 的主机部分")
	_ = fs.Parse(args)

	if *target == "" || *caFile == "" || *certFile == "" || *keyFile == "" {
		log.Fatalf("client requires -target, -ca, -cert and -key")
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
	}

	conn, err := tls.Dial("tcp", *server, cfg)
	if err != nil {
		log.Fatalf("dial tunnel server failed, server=%s, err=%v", *server, err)
	}
	sess := newSession(conn, true)

	// 逼出服务端对客户端证书的裁决：TLS 1.3 下 tls.Dial 成功不代表认证通过，
	// 服务端拒收时它已经把连接关了。这里拿不到 PONG 就直接退出，
	// 而不是等到第一条流失败才暴露——那时错误已经埋在一堆转发失败里了。
	if err := sess.Handshake(handshakeTimeout); err != nil {
		log.Fatalf("server rejected this client, check -cert/-key against the server's -client-ca, server=%s, err=%v", *server, err)
	}
	log.Printf("crypto tunnel ready, server=%s, auth=mtls", *server)

	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatalf("listen local failed, addr=%s, err=%v", *listen, err)
	}
	log.Printf("crypto forward listening, listen=%s, target=%s", *listen, *target)

	for {
		local, err := ln.Accept()
		if err != nil {
			log.Printf("accept failed, err=%v", err)
			continue
		}
		go func(local net.Conn) {
			st, err := sess.Open(*target)
			if err != nil {
				log.Printf("open stream failed, target=%s, err=%v", *target, err)
				_ = local.Close()
				return
			}
			in, out, berr := bridge(local, st)
			log.Printf("forward finished, id=%d, target=%s, local_to_remote=%d, remote_to_local=%d, err=%v",
				st.ID(), *target, in, out, berr)
		}(local)
	}
}

// serveCrypto 接受连接并起会话。
//
// 这里没有"认证"这一步，是因为它已经在 TLS 握手里做完了：能 Accept 到
// 的连接必然持有 -client-ca 签发的证书，校验失败的连接被 crypto/tls
// 直接丢掉，根本到不了应用层。这正是 mTLS 相对应用层认证的优势——
// 不存在"认证没过但已经能驱动服务端 dial"的窗口。
func serveCrypto(ln net.Listener, allow *AllowList) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept failed, err=%v", err)
			return
		}
		go serveSession(newSession(conn, false), allow)
	}
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


