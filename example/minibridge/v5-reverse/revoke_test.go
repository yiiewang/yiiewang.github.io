package main

// 单张吊销（-revoked）测试。
//
// 断言模式沿用 crypto_test.go 的经验：TLS 1.3 下 tls.Dial 不报错，
// 服务端的裁决要靠 newSession(...).Handshake 逼出来。

import (
	"crypto/tls"
	"crypto/x509"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// leafSerial 从 tls.Config 里取出（第一张证书的）叶子证书 serial。
func leafSerial(t *testing.T, cfg *tls.Config) *big.Int {
	t.Helper()
	if len(cfg.Certificates) == 0 || len(cfg.Certificates[0].Certificate) == 0 {
		t.Fatalf("config carries no leaf certificate")
	}
	cert, err := x509.ParseCertificate(cfg.Certificates[0].Certificate[0])
	if err != nil {
		t.Fatalf("parse leaf certificate: %v", err)
	}
	return cert.SerialNumber
}

// writeRevokedFile 覆盖写吊销清单文件。两次调用之间由调用方 sleep
// 保证 mtime 前进，触发热加载。
func writeRevokedFile(t *testing.T, path string, serials ...string) {
	t.Helper()
	content := "# revoked serials (test)\n"
	for _, s := range serials {
		content += s + "\n"
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write revoked file: %v", err)
	}
}

// ---------------- 解析 ----------------

func TestParseSerialLine(t *testing.T) {
	cases := []struct {
		name    string
		line    string
		want    uint64 // 期望的十进制值；0 表示该行应被忽略/报错
		wantOK  bool
		wantHex bool // 是否按十六进制解析
	}{
		{"十进制", "12345", 12345, true, false},
		{"0x 十六进制", "0x6F04A1", 0x6F04A1, true, true},
		{"openssl 冒号 hex", "6F:04:A1", 0x6F04A1, true, true},
		{"大写十六进制", "6F04A1", 0x6F04A1, true, true},
		{"注释行", "# revoked: old cert", 0, false, false},
		{"空行", "   ", 0, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := parseSerialLine(tc.line)
			if ok != tc.wantOK {
				t.Fatalf("parseSerialLine(%q) ok=%v, want %v", tc.line, ok, tc.wantOK)
			}
			if !ok {
				return
			}
			if got.Uint64() != tc.want {
				t.Fatalf("parseSerialLine(%q) = %d, want %d", tc.line, got.Uint64(), tc.want)
			}
		})
	}
}

// ---------------- 端到端：服务端吊销客户端证书 ----------------

func TestRevokedClientCertRejected(t *testing.T) {
	ca, caKey := testCA(t)
	cfg := testServerMTLSConfig(t, ca, caKey)

	cliCfg := testClientTLSConfig(t, ca, caKey)
	cliSerial := leafSerial(t, cliCfg)

	revokedPath := filepath.Join(t.TempDir(), "revoked.txt")
	// 名单里放一个无关 serial：合法客户端必须照常通过（回归对照）
	writeRevokedFile(t, revokedPath, "0x"+big.NewInt(1).Text(16))
	cfg.VerifyPeerCertificate = verifyPeer(NewRevocationList(revokedPath))
	addr := startCryptoServer(t, cfg, testAllowAll(t))

	// 干净名单：握手必须成功
	okConn, err := tls.Dial("tcp", addr, cliCfg)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { okConn.Close() })
	if err := newSession(okConn, true).Handshake(3 * time.Second); err != nil {
		t.Fatalf("cert not in list must pass: %v", err)
	}

	// 写入真实 serial（0x 十六进制形式，与运维手填的习惯一致）
	writeRevokedFile(t, revokedPath, "0x"+cliSerial.Text(16))

	badConn, err := tls.Dial("tcp", addr, cliCfg)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { badConn.Close() })
	if err := newSession(badConn, true).Handshake(3 * time.Second); err == nil {
		t.Fatalf("revoked client certificate must be rejected")
	} else {
		t.Logf("rejection as expected: %v", err)
	}
}

// ---------------- 端到端：客户端吊销服务端证书（对称） ----------------

func TestRevokedServerCertRejected(t *testing.T) {
	ca, caKey := testCA(t)
	srvCfg := testServerMTLSConfig(t, ca, caKey)
	addr := startCryptoServer(t, srvCfg, testAllowAll(t))

	cliCfg := testClientTLSConfig(t, ca, caKey)
	srvSerial := leafSerial(t, srvCfg)

	revokedPath := filepath.Join(t.TempDir(), "revoked.txt")
	writeRevokedFile(t, revokedPath, "# empty")
	cliCfg.VerifyPeerCertificate = verifyPeer(NewRevocationList(revokedPath))

	// 干净名单：正常握手
	okConn, err := tls.Dial("tcp", addr, cliCfg)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { okConn.Close() })
	if err := newSession(okConn, true).Handshake(3 * time.Second); err != nil {
		t.Fatalf("server cert not in list must pass: %v", err)
	}

	// 吊销服务端证书（openssl 冒号 hex 也得认）
	hexStr := srvSerial.Text(16)
	colonHex := ""
	for i := 0; i+2 <= len(hexStr); i += 2 {
		if colonHex != "" {
			colonHex += ":"
		}
		colonHex += hexStr[i : i+2]
	}
	writeRevokedFile(t, revokedPath, colonHex)

	// 客户端侧的吊销裁决在本地握手内同步执行：多数情况下 Dial 直接报错；
	// 万一握手完成了，也要靠 Handshake 逼出服务端的关闭
	badConn, err := tls.Dial("tcp", addr, cliCfg)
	if err == nil {
		t.Cleanup(func() { badConn.Close() })
		if herr := newSession(badConn, true).Handshake(3 * time.Second); herr == nil {
			t.Fatalf("revoked server certificate must be rejected")
		} else {
			t.Logf("rejection as expected (via handshake): %v", herr)
		}
	} else {
		t.Logf("rejection as expected (at dial): %v", err)
	}
}

// ---------------- 热加载单元级：mtime 变化翻转 IsRevoked ----------------

func TestRevocationListHotReload(t *testing.T) {
	revokedPath := filepath.Join(t.TempDir(), "revoked.txt")
	writeRevokedFile(t, revokedPath, "# empty")
	rl := NewRevocationList(revokedPath)

	probe := &big.Int{}
	probe.SetString("6F04A1", 16)
	if rl.IsRevoked(probe) {
		t.Fatalf("empty list must not revoke anything")
	}

	writeRevokedFile(t, revokedPath, "6F:04:A1")
	if !rl.IsRevoked(probe) {
		t.Fatalf("after reload the serial must be revoked")
	}

	// 清空名单：恢复放行（误吊销的可逆性）
	writeRevokedFile(t, revokedPath, "# cleared")
	if rl.IsRevoked(probe) {
		t.Fatalf("cleared list must not revoke anything")
	}
}
