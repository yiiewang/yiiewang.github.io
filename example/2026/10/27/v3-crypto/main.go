// Package main 是 minibridge 第三阶段：给隧道上锁。
//
// 在 v2 的多路复用之上加三件事：
//  1. TLS：加密 + 服务端身份（客户端必须校验证书，绝不能 InsecureSkipVerify）
//  2. 认证：可插拔。默认 sshkey（用 OpenSSH 密钥做挑战-响应），可选 mtls（客户端证书）
//  3. 授权：目标白名单，治 SSRF
//
// 顺序是铁律：TLS 握手 → 认证 → 才接受 SYN → 才 dial 目标。
package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
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
	fmt.Fprint(os.Stderr, `用法：
  v3-crypto server -addr :9000 -cert srv.crt -key srv.key -allow 127.0.0.1:8000
  v3-crypto client -listen :8080 -server host:9000 -target 127.0.0.1:8000 \
      -ca ca.crt -identity ~/.ssh/id_ed25519

  认证方式 -auth：sshkey（默认，用 SSH 密钥做挑战-响应）| mtls（客户端证书）
`)
}

// runServer 启动带 TLS 的隧道服务端。
func runServer(args []string) {
	fs := flag.NewFlagSet("server", flag.ExitOnError)
	addr := fs.String("addr", "127.0.0.1:9000", "隧道监听地址")
	certFile := fs.String("cert", "", "服务端证书 PEM（必填）")
	keyFile := fs.String("key", "", "服务端私钥 PEM（必填）")
	authMode := fs.String("auth", "sshkey", "客户端认证方式：sshkey | mtls")
	authorizedKeys := fs.String("authorized-keys", defaultAuthorizedKeys(), "sshkey 模式下的公钥名单")
	clientCAFile := fs.String("client-ca", "", "mtls 模式下校验客户端证书的 CA PEM")
	allowSpec := fs.String("allow", "127.0.0.1:*", "允许拨号的目标，逗号分隔")
	_ = fs.Parse(args)

	if *certFile == "" || *keyFile == "" {
		log.Fatalf("server requires -cert and -key")
	}

	cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		log.Fatalf("load server cert failed, err=%v", err)
	}

	allow, err := NewAllowList(strings.Split(*allowSpec, ","))
	if err != nil {
		log.Fatalf("parse allow list failed, err=%v", err)
	}

	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
	}

	auth, err := newServerAuth(*authMode, cfg, *authorizedKeys, *clientCAFile)
	if err != nil {
		log.Fatalf("setup auth failed, err=%v", err)
	}

	ln, err := tls.Listen("tcp", *addr, cfg)
	if err != nil {
		log.Fatalf("listen tunnel failed, addr=%s, err=%v", *addr, err)
	}
	log.Printf("crypto tunnel server listening, addr=%s, auth=%s", *addr, auth.Name())

	serveCrypto(ln, auth, allow)
}

// newServerAuth 按模式装配认证器。
//
// mtls 模式要把根池和 ClientAuth 写进 tls.Config——
// 少了这两行，x509Auth 那个空实现就等于"谁都能连"。
func newServerAuth(mode string, cfg *tls.Config, authorizedKeys, clientCAFile string) (Authenticator, error) {
	switch mode {
	case "sshkey":
		keys, err := LoadAuthorizedKeys(authorizedKeys)
		if err != nil {
			return nil, err
		}
		log.Printf("loaded %d authorized key(s) from %s", len(keys), authorizedKeys)
		return newSSHKeyServerAuth(keys), nil

	case "mtls":
		if clientCAFile == "" {
			return nil, fmt.Errorf("mtls requires -client-ca")
		}
		pemData, err := os.ReadFile(clientCAFile)
		if err != nil {
			return nil, err
		}
		pool := loadCertPool(pemData)
		cfg.ClientCAs = pool
		cfg.ClientAuth = tls.RequireAndVerifyClientCert
		return x509Auth{}, nil

	default:
		return nil, fmt.Errorf("unknown auth mode %q", mode)
	}
}

// runClient 启动带 TLS 的转发客户端。
func runClient(args []string) {
	fs := flag.NewFlagSet("client", flag.ExitOnError)
	listen := fs.String("listen", ":8080", "本地监听地址")
	server := fs.String("server", "127.0.0.1:9000", "隧道服务端地址")
	target := fs.String("target", "", "目标地址 host:port（必填，服务端视角）")
	caFile := fs.String("ca", "", "校验服务端证书的 CA PEM（必填）")
	authMode := fs.String("auth", "sshkey", "认证方式：sshkey | mtls")
	identity := fs.String("identity", defaultIdentity(), "sshkey 模式下的私钥路径")
	certFile := fs.String("cert", "", "mtls 模式下的客户端证书 PEM")
	keyFile := fs.String("key", "", "mtls 模式下的客户端私钥 PEM")
	serverName := fs.String("server-name", "", "TLS 校验用的服务名，默认取 -server 的主机部分")
	_ = fs.Parse(args)

	if *target == "" || *caFile == "" {
		log.Fatalf("client requires -target and -ca")
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

	cfg := &tls.Config{
		RootCAs:    loadCertPool(pemData),
		ServerName: name,
		MinVersion: tls.VersionTLS13,
	}

	auth, err := newClientAuth(*authMode, cfg, *identity, *certFile, *keyFile)
	if err != nil {
		log.Fatalf("setup auth failed, err=%v", err)
	}

	conn, err := tls.Dial("tcp", *server, cfg)
	if err != nil {
		log.Fatalf("dial tunnel server failed, server=%s, err=%v", *server, err)
	}
	if err := auth.Client(conn); err != nil {
		log.Fatalf("authenticate failed, auth=%s, err=%v", auth.Name(), err)
	}
	log.Printf("crypto tunnel ready, server=%s, auth=%s", *server, auth.Name())

	sess := newSession(conn, true)

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

func newClientAuth(mode string, cfg *tls.Config, identity, certFile, keyFile string) (Authenticator, error) {
	switch mode {
	case "sshkey":
		signer, err := LoadPrivateKey(identity)
		if err != nil {
			return nil, err
		}
		return newSSHKeyClientAuth(signer), nil

	case "mtls":
		if certFile == "" || keyFile == "" {
			return nil, fmt.Errorf("mtls requires -cert and -key")
		}
		pair, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, err
		}
		cfg.Certificates = []tls.Certificate{pair}
		return x509Auth{}, nil

	default:
		return nil, fmt.Errorf("unknown auth mode %q", mode)
	}
}

// serveCrypto 接受连接：先认证，认证通过才起会话。
func serveCrypto(ln net.Listener, auth Authenticator, allow *AllowList) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept failed, err=%v", err)
			return
		}
		go func() {
			if err := auth.Server(conn); err != nil {
				log.Printf("auth failed, remote=%s, err=%v", conn.RemoteAddr(), err)
				_ = conn.Close()
				return
			}
			serveSession(newSession(conn, false), allow)
		}()
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

func defaultAuthorizedKeys() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "authorized_keys"
	}
	return filepath.Join(home, ".ssh", "authorized_keys")
}

func defaultIdentity() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "id_ed25519"
	}
	return filepath.Join(home, ".ssh", "id_ed25519")
}
