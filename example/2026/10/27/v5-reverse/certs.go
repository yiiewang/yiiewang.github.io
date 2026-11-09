package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// 本文件只做一件事：造出自签 CA 与叶子证书，供示例和测试使用。
//
// 生产环境的正确做法不是自签，而是：
//   - 服务端证书：企业 CA 签发，或 Let's Encrypt（ACME）
//   - 客户端证书：企业 CA 统一签发（mTLS 场景）
// 自签的差别只在"谁来证明这张证书可信"，协议本身完全一样。

const certValidYears = 1

// GenerateCA 生成一张自签根证书。
func GenerateCA(cn string) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: cn},
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().AddDate(certValidYears, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}
	ca, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, nil, err
	}
	return ca, key, nil
}

// IssueCert 用 CA 签发一张叶子证书。
//
// ExtKeyUsage 决定这张证书能做什么：服务端要 ServerAuth，客户端要 ClientAuth。
// 少了这一项，TLS 握手会以 "certificate specifies an incompatible key usage" 失败。
func IssueCert(ca *x509.Certificate, caKey *ecdsa.PrivateKey, cn string, server bool, ips []net.IP, dnsNames []string) (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	usage := x509.ExtKeyUsageClientAuth
	if server {
		usage = x509.ExtKeyUsageServerAuth
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 62))
	if err != nil {
		return tls.Certificate{}, err
	}

	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now().Add(-time.Minute),
		NotAfter:     time.Now().AddDate(certValidYears, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{usage},
		IPAddresses:  ips,
		DNSNames:     dnsNames,
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, ca, &key.PublicKey, caKey)
	if err != nil {
		return tls.Certificate{}, err
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return tls.Certificate{}, err
	}

	pair, err := tls.X509KeyPair(
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}),
	)
	return pair, err
}

// CertPool 把 CA 装成校验用的根池。
func CertPool(ca *x509.Certificate) *x509.CertPool {
	pool := x509.NewCertPool()
	pool.AddCert(ca)
	return pool
}

// WriteCertAndKey 把证书与私钥写成 PEM 文件，供命令行使用。
func WriteCertAndKey(dir, name string, pair tls.Certificate) (certPath, keyPath string, err error) {
	if len(pair.Certificate) == 0 {
		return "", "", fmt.Errorf("empty certificate")
	}
	key, ok := pair.PrivateKey.(*ecdsa.PrivateKey)
	if !ok {
		return "", "", fmt.Errorf("unsupported private key type %T", pair.PrivateKey)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return "", "", err
	}

	certPath = filepath.Join(dir, name+".crt")
	keyPath = filepath.Join(dir, name+".key")
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: pair.Certificate[0]})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	if err := os.WriteFile(certPath, certPEM, 0o644); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		return "", "", err
	}
	return certPath, keyPath, nil
}
