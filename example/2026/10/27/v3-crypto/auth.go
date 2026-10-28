package main

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
)

// Authenticator 是 TLS 握手之后、允许开流之前的身份验证。
//
// 注意它的位置：在 TLS **之后**，在第一个 SYN **之前**。
// 认证没过就绝不能让对端驱动我们 dial 目标——否则等于给陌生人留了一个 SSRF 入口。
type Authenticator interface {
	// Name 返回认证方式的名字，用于配置与日志。
	Name() string
	// Server 在服务端验证客户端身份。
	Server(conn net.Conn) error
	// Client 在客户端完成身份证明。
	Client(conn net.Conn) error
}

const (
	authOK    byte = 0
	authFail  byte = 1
	nonceLen       = 32
	maxSigLen      = 8192
)

var errAuthFailed = errors.New("minibridge: authentication failed")

// sshKeyAuth 用 OpenSSH 密钥做挑战-响应认证。
//
// 它用的不是 SSH 协议，只是 SSH 的**密钥格式与签名算法**：
// 服务端发一段随机 nonce，客户端用私钥签名，服务端拿 authorized_keys 里的公钥验签。
// 挑战-响应这件事本身跟通道用什么加密无关——TLS、裸 TCP 都能跑。
type sshKeyAuth struct {
	keys   []ssh.PublicKey // 服务端：允许的公钥
	signer ssh.Signer      // 客户端：签名的私钥
}

func newSSHKeyServerAuth(keys []ssh.PublicKey) *sshKeyAuth {
	return &sshKeyAuth{keys: keys}
}

func newSSHKeyClientAuth(signer ssh.Signer) *sshKeyAuth {
	return &sshKeyAuth{signer: signer}
}

func (a *sshKeyAuth) Name() string { return "sshkey" }

// Server 发 nonce → 收签名 → 验签 → 回一字节结果。
func (a *sshKeyAuth) Server(conn net.Conn) error {
	nonce := make([]byte, nonceLen)
	if _, err := rand.Read(nonce); err != nil {
		return err
	}
	if _, err := conn.Write(nonce); err != nil {
		return err
	}

	var lenBuf [4]byte
	if _, err := io.ReadFull(conn, lenBuf[:]); err != nil {
		return err
	}
	n := binary.BigEndian.Uint32(lenBuf[:])
	if n == 0 || n > maxSigLen {
		return replyAuth(conn, authFail, fmt.Errorf("invalid signature length %d", n))
	}
	blob := make([]byte, n)
	if _, err := io.ReadFull(conn, blob); err != nil {
		return err
	}

	// ssh.Signature 带 ssh tag，可以直接用 Unmarshal 从 SSH wire format 还原
	// （Format + Blob + rest）。这个包没有提供 ParseSignature，别去找。
	var sig ssh.Signature
	if err := ssh.Unmarshal(blob, &sig); err != nil {
		return replyAuth(conn, authFail, err)
	}
	for _, pk := range a.keys {
		if pk.Verify(nonce, &sig) == nil {
			return replyAuth(conn, authOK, nil)
		}
	}
	return replyAuth(conn, authFail, errAuthFailed)
}

// Client 收 nonce → 签名 → 等结果。
func (a *sshKeyAuth) Client(conn net.Conn) error {
	nonce := make([]byte, nonceLen)
	if _, err := io.ReadFull(conn, nonce); err != nil {
		return err
	}

	sig, err := a.signer.Sign(rand.Reader, nonce)
	if err != nil {
		return err
	}
	blob := ssh.Marshal(sig)
	buf := make([]byte, 4+len(blob))
	binary.BigEndian.PutUint32(buf[:4], uint32(len(blob)))
	copy(buf[4:], blob)
	if _, err := conn.Write(buf); err != nil {
		return err
	}

	var status [1]byte
	if _, err := io.ReadFull(conn, status[:]); err != nil {
		return err
	}
	if status[0] != authOK {
		return errAuthFailed
	}
	return nil
}

// replyAuth 回一字节结果，并把错误返回给调用方（调用方据此决定是否断开）。
func replyAuth(conn net.Conn, status byte, err error) error {
	if _, werr := conn.Write([]byte{status}); werr != nil {
		return werr
	}
	return err
}

// x509Auth 是 mTLS 的实现：客户端证书校验在 TLS 握手内由 crypto/tls 完成，
// 应用层无需再做任何事，所以两个方法都是空实现。
//
// 前提是服务端必须设置 ClientAuth: RequireAndVerifyClientCert 且配好 ClientCAs——
// 否则"什么都不做"就等于没有任何客户端认证。
type x509Auth struct{}

func (x509Auth) Name() string          { return "mtls" }
func (x509Auth) Server(net.Conn) error { return nil }
func (x509Auth) Client(net.Conn) error { return nil }

// LoadAuthorizedKeys 读 OpenSSH 的 authorized_keys 文件。
func LoadAuthorizedKeys(path string) ([]ssh.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var keys []ssh.PublicKey
	for len(data) > 0 {
		pub, _, _, rest, err := ssh.ParseAuthorizedKey(data)
		if err != nil {
			break // 遇到无法解析的行就停：注释行、空行都会到这里
		}
		keys = append(keys, pub)
		data = rest
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("no public key parsed from %s", path)
	}
	return keys, nil
}

// LoadPrivateKey 读 OpenSSH 私钥（ed25519 / rsa / ecdsa）。
//
// 带口令的私钥请先 ssh-add 交给 agent——这里不处理加密私钥，
// 否则要么在代码里存口令，要么每次交互输入，都不是好选择。
func LoadPrivateKey(path string) (ssh.Signer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(data)
}
