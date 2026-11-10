package main

// 单张证书吊销（CRL/OCSP 的极简替代）。
//
// 动机：自建 CA 体系里 Go 标准库只验证"链是否通到 CA"，签发出去的
// 叶子证书在到期前永远有效——泄漏一张就得换 CA 重签全链，代价不成比例。
// 这里补上最后一块：握手时解析对端叶子证书，serial 命中清单即拒绝，
// 签发链有效也挡下。
//
// 形态选择：-revoked 指向一个文本文件，每行一个证书 serial。
//   - 0x 前缀 = 十六进制（证书工具的常见打印形式，冒号分隔自动剥掉）
//   - 纯 0-9  = 十进制
//   - 其余    = 十六进制（openssl 风格 6F:04:A1，去冒号）
//   - # 开头注释，空行忽略；格式错误的行启动即 Fatal——
//     吊销清单静默丢一行等于该证书复活，宁可吵死不可漏挡
//
// 热加载：每次握手读文件比对内容，变化即重载（原子整表替换）。不用
// mtime 判变更——容器文件系统的 mtime 常为秒级粒度，同一秒内两次修改
// 检测不到；清单只有几行，每次握手读一次的成本可以忽略。
// 两个节点、个位数证书的规模，一张文本表比 CRL 的分发/签名/过期
// 管理便宜一个量级，也不给标准库之外引入任何解析信任。

import (
	"crypto/x509"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"sync"
)

// RevocationList 串行吊销表。nil receiver 的所有方法都是安全空操作
// （未启用 -revoked 时零行为变化），因此调用方不需要判空。
type RevocationList struct {
	path string

	mu    sync.Mutex
	raw   string              // 上次成功加载的内容，变更检测用
	known bool                // 是否成功加载过（区分"还没加载"与"加载失败"）
	hex   map[string]struct{} // key: serial 十进制字符串
}

// NewRevocationList 建表但不立即读文件——首次校验时才加载，
// 与热加载路径走同一段代码。
func NewRevocationList(path string) *RevocationList {
	if path == "" {
		return nil
	}
	return &RevocationList{path: path}
}

// IsRevoked 检查前先看文件是否变化，必要时重载。
func (r *RevocationList) IsRevoked(serial *big.Int) bool {
	if r == nil || serial == nil || serial.Sign() <= 0 {
		return false
	}
	r.maybeReload()
	r.mu.Lock()
	_, hit := r.hex[serial.Text(10)]
	r.mu.Unlock()
	return hit
}

// maybeReload 每次校验前读取文件并比对内容，变化即整表重载。
//
// 不用 mtime 判变更：容器 overlayfs 的 mtime 常为秒级粒度，同一秒内
// 两次修改检测不到。清单本身只有几行，每次握手读一次的成本可以忽略，
// 内容比对在文件系统差异下行为一致。
func (r *RevocationList) maybeReload() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.path == "" {
		return
	}
	data, err := os.ReadFile(r.path)
	if err != nil {
		// 文件不可访问：从未成功加载则视为空表；否则保留旧表——
		// 吊销数据在故障时收紧为"维持现状"，绝不放大放行面
		if !r.known {
			log.Printf("revocation list not readable yet, treating as empty: %v", err)
			r.hex = map[string]struct{}{}
			r.known = true
		} else {
			log.Printf("revocation list read failed, keeping last loaded table: %v", err)
		}
		return
	}
	if r.known && string(data) == r.raw {
		return // 内容未变
	}
	table := map[string]struct{}{}
	for _, line := range strings.Split(string(data), "\n") {
		serial, ok := parseSerialLine(line)
		if !ok {
			continue // 注释/空行
		}
		table[serial.Text(10)] = struct{}{}
	}
	r.hex = table
	r.raw = string(data)
	r.known = true
	log.Printf("revocation list loaded, revoked=%d, file=%s", len(table), r.path)
}

// parseSerialLine 解析一行 serial。格式错误 Fatal 而不是跳过：
// 清单是安全收紧面，宁可启动失败被人发现，不可静默漏挡。
func parseSerialLine(line string) (*big.Int, bool) {
	s := strings.TrimSpace(line)
	if s == "" || strings.HasPrefix(s, "#") {
		return nil, false
	}
	s = strings.ReplaceAll(s, ":", "")
	s = strings.TrimSpace(s)

	base := 10
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
		base = 16
	} else if strings.ContainsAny(s, "abcdefABCDEF") {
		base = 16
	}
	n, ok := new(big.Int).SetString(s, base)
	if !ok || n.Sign() <= 0 {
		log.Fatalf("unparseable serial line %q in revoked file (use 0x-prefixed hex or plain decimal)", line)
	}
	return n, true
}

// verifyPeer 生成 tls.Config.VerifyPeerCertificate 回调。
// 回调在标准链校验之后执行，只追加"单张吊销"这一层裁决：
// rawCerts[0] 恒为对端叶子证书，解析失败说明链校验已在标准
// 流程报错，这里返回 nil 不抢它的错误。
func verifyPeer(rl *RevocationList) func([][]byte, [][]*x509.Certificate) error {
	return func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
		if rl == nil || len(rawCerts) == 0 {
			return nil
		}
		leaf, err := x509.ParseCertificate(rawCerts[0])
		if err != nil {
			return nil
		}
		if rl.IsRevoked(leaf.SerialNumber) {
			return fmt.Errorf("peer certificate revoked: serial=%s", leaf.SerialNumber.Text(16))
		}
		return nil
	}
}
