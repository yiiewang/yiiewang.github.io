package main

import (
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"net"
	"strings"
	"testing"
	"time"
)

// ---------------- 白名单 ----------------

func TestAllowListResolve(t *testing.T) {
	cases := []struct {
		name   string
		rules  []string
		target string
		wantOK bool
	}{
		{"精确端口匹配", []string{"127.0.0.1:8000"}, "127.0.0.1:8000", true},
		{"端口不匹配", []string{"127.0.0.1:8000"}, "127.0.0.1:9000", false},
		{"任意端口", []string{"127.0.0.1:*"}, "127.0.0.1:9100", true},
		{"CIDR 网段", []string{"10.0.0.0/8:3306"}, "10.2.3.4:3306", true},
		{"网段外被拒", []string{"10.0.0.0/8:3306"}, "192.168.1.1:3306", false},
		{"域名解析后命中", []string{"127.0.0.1:*"}, "localhost:8000", true},
		{"回环之外被拒", []string{"127.0.0.1:*"}, "8.8.8.8:53", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			allow, err := NewAllowList(tc.rules)
			if err != nil {
				t.Fatalf("build allow list: %v", err)
			}
			got, ok := allow.Resolve(tc.target)
			if ok != tc.wantOK {
				t.Fatalf("Resolve(%q) ok=%v, want %v (got addr %q)", tc.target, ok, tc.wantOK, got)
			}
			// 放行时必须返回 IP 形式，杜绝校验与拨号之间的二次解析
			if ok {
				host, _, err := net.SplitHostPort(got)
				if err != nil || net.ParseIP(host) == nil {
					t.Fatalf("Resolve should return IP form, got %q", got)
				}
			}
		})
	}
}

func TestAllowListRejectsDomainRule(t *testing.T) {
	if _, err := NewAllowList([]string{"example.com:80"}); err == nil {
		t.Fatalf("domain rules must be rejected (DNS rebinding)")
	}
}

// ---------------- TLS ----------------

func TestTLSRejectsUntrustedCA(t *testing.T) {
	ca, caKey := testCA(t)
	cfg := testServerTLSConfig(t, ca, caKey)
	addr := startCryptoServer(t, cfg, testAllowAll(t))

	// 用另一家 CA 去校验服务端证书：必须失败
	otherCA, _ := testCA(t)
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		RootCAs:    CertPool(otherCA),
		ServerName: "127.0.0.1",
		MinVersion: tls.VersionTLS13,
	})
	if err == nil {
		conn.Close()
		t.Fatalf("client must reject server certificate signed by unknown CA")
	}
	if !strings.Contains(err.Error(), "certificate") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTLSRejectsPlaintextClient(t *testing.T) {
	ca, caKey := testCA(t)
	cfg := testServerTLSConfig(t, ca, caKey)
	addr := startCryptoServer(t, cfg, testAllowAll(t))

	// 不带 TLS 直接连：服务端等的是 TLS ClientHello，握手必然失败
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	if _, err := conn.Write([]byte("MBRD")); err != nil {
		return // 连接被服务端先关掉也算过关
	}
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 1)
	if _, err := conn.Read(buf); err == nil {
		// 读到东西也不代表握手成功，只有完整握手后才会进入认证阶段
		t.Logf("plaintext client got a byte, but no session is created")
	}
}

func TestMTLSRequiresClientCert(t *testing.T) {
	target := startMuxEchoTarget(t)

	ca, caKey := testCA(t)
	cfg := testServerMTLSConfig(t, ca, caKey)
	addr := startCryptoServer(t, cfg, testAllowAll(t))

	// 没有客户端证书：TLS 1.3 下 tls.Dial **不会**失败——
	// 客户端在收到服务端对证书的裁决之前就认为握手完成了。
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		RootCAs:    CertPool(ca),
		ServerName: "127.0.0.1",
		MinVersion: tls.VersionTLS13,
	})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	// 裁决由 Handshake 逼出来：服务端这时已经把连接关了，拿不到 PONG。
	if err := newSession(conn, true).Handshake(3 * time.Second); err == nil {
		t.Fatalf("server must reject client without certificate")
	}

	// 带上客户端证书：完整走通一次转发才算真的通过
	withCert, err := tls.Dial("tcp", addr, testClientTLSConfig(t, ca, caKey))
	if err != nil {
		t.Fatalf("mTLS dial: %v", err)
	}
	t.Cleanup(func() { withCert.Close() })

	st, err := newSession(withCert, true).Open(target)
	if err != nil {
		t.Fatalf("open stream over mTLS: %v", err)
	}
	defer st.Close()

	const msg = "hello mtls"
	if _, err := st.Write([]byte(msg)); err != nil {
		t.Fatalf("write: %v", err)
	}
	buf := make([]byte, len(msg))
	readFullTimeout(t, st, buf, 3*time.Second)
	if string(buf) != msg {
		t.Fatalf("echo mismatch: got %q", buf)
	}
}

// Handshake 是客户端确认"服务端真的收了我"的唯一手段，成功路径也得钉住。
func TestSessionHandshakeSucceedsOverAcceptedMTLS(t *testing.T) {
	ca, caKey := testCA(t)
	cfg := testServerMTLSConfig(t, ca, caKey)
	addr := startCryptoServer(t, cfg, testAllowAll(t))

	conn, err := tls.Dial("tcp", addr, testClientTLSConfig(t, ca, caKey))
	if err != nil {
		t.Fatalf("mtls dial: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	if err := newSession(conn, true).Handshake(3 * time.Second); err != nil {
		t.Fatalf("handshake over accepted mTLS connection: %v", err)
	}
}

// ---------------- 端到端 ----------------

func TestCryptoEndToEndForward(t *testing.T) {
	target := startMuxEchoTarget(t)

	ca, caKey := testCA(t)
	cfg := testServerMTLSConfig(t, ca, caKey)
	addr := startCryptoServer(t, cfg, testAllowAll(t))

	conn, err := tls.Dial("tcp", addr, testClientTLSConfig(t, ca, caKey))
	if err != nil {
		t.Fatalf("mtls dial: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	sess := newSession(conn, true)
	st, err := sess.Open(target)
	if err != nil {
		t.Fatalf("open stream: %v", err)
	}
	defer st.Close()

	const msg = "hello v3"
	if _, err := st.Write([]byte(msg)); err != nil {
		t.Fatalf("write: %v", err)
	}
	buf := make([]byte, len(msg))
	readFullTimeout(t, st, buf, 3*time.Second)
	if string(buf) != msg {
		t.Fatalf("echo mismatch: got %q", buf)
	}
}

func TestCryptoDeniesTargetOutsideAllowList(t *testing.T) {
	ca, caKey := testCA(t)
	cfg := testServerMTLSConfig(t, ca, caKey)
	// 白名单只允许回环的 8000，目标用别的端口
	allow, err := NewAllowList([]string{"127.0.0.1:8000"})
	if err != nil {
		t.Fatalf("build allow list: %v", err)
	}
	addr := startCryptoServer(t, cfg, allow)

	conn, err := tls.Dial("tcp", addr, testClientTLSConfig(t, ca, caKey))
	if err != nil {
		t.Fatalf("mtls dial: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	sess := newSession(conn, true)
	denied := freeMuxAddr(t) // 一个没人监听、也不在白名单里的地址
	if _, err := sess.Open(denied); err == nil {
		t.Fatalf("server must reject target outside allow list")
	}
}

// ---------------- 测试桩 ----------------

func testCA(t *testing.T) (*x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()

	ca, key, err := GenerateCA("minibridge-test-ca")
	if err != nil {
		t.Fatalf("generate ca: %v", err)
	}
	return ca, key
}

func testServerTLSConfig(t *testing.T, ca *x509.Certificate, caKey *ecdsa.PrivateKey) *tls.Config {
	t.Helper()

	pair, err := IssueCert(ca, caKey, "127.0.0.1", true, []net.IP{net.ParseIP("127.0.0.1")}, nil)
	if err != nil {
		t.Fatalf("issue server cert: %v", err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{pair},
		MinVersion:   tls.VersionTLS13,
	}
}

// testServerMTLSConfig 在普通服务端配置之上要求并校验客户端证书。
func testServerMTLSConfig(t *testing.T, ca *x509.Certificate, caKey *ecdsa.PrivateKey) *tls.Config {
	cfg := testServerTLSConfig(t, ca, caKey)
	cfg.ClientCAs = CertPool(ca)
	cfg.ClientAuth = tls.RequireAndVerifyClientCert
	return cfg
}

// testClientTLSConfig 造出带客户端证书、并校验服务端证书的 tls.Config。
func testClientTLSConfig(t *testing.T, ca *x509.Certificate, caKey *ecdsa.PrivateKey) *tls.Config {
	t.Helper()

	cliCert, err := IssueCert(ca, caKey, "client", false, nil, nil)
	if err != nil {
		t.Fatalf("issue client cert: %v", err)
	}
	return &tls.Config{
		RootCAs:      CertPool(ca),
		ServerName:   "127.0.0.1",
		Certificates: []tls.Certificate{cliCert},
		MinVersion:   tls.VersionTLS13,
	}
}

func startCryptoServer(t *testing.T, cfg *tls.Config, allow *AllowList) string {
	t.Helper()

	ln, err := tls.Listen("tcp", "127.0.0.1:0", cfg)
	if err != nil {
		t.Fatalf("tls listen: %v", err)
	}
	t.Cleanup(func() { ln.Close() })

	go serveCrypto(ln, allow)
	return ln.Addr().String()
}
