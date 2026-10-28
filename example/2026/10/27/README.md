# minibridge：手写一个隧道协议（学习路线）

从 `ssh -L` 出发，一步步造出 VSCode Remote-SSH 那种"一条连接多路复用"的桥接能力。每个阶段一个可运行目录，配一篇博客。

代码在同一个 Go module 下，阶段之间可以复用：

```bash
cd example/2026/10/27
go test -count=1 -race ./...   # 跑全部阶段的测试
```

## 路线图

| 阶段 | 目录 | 状态 | 学到什么 |
|---|---|---|---|
| v1 最简转发 | `v1-forward` | ✅ 完成 | TCP 监听/拨号、握手协议、`io.Copy` 双向桥接、半关闭 |
| v2 帧协议多路复用 | `v2-mux` | ✅ 完成 | 帧格式设计、stream 生命周期、单连接跑多条流 |
| v3 加密与认证 | `v3-crypto` | ✅ 完成 | TLS、证书校验、"先认证后转发"的顺序、认证可插拔、目标白名单 |
| v4 反向转发与控制面 | `v4-control` | 计划 | 反向通道语义、运行时注册/注销端口（v2 已把 Open/Accept 做成对称，接上去即可） |
| v5 生产化 | `v5-ops` | 计划 | 心跳保活、断线重连、端口自动探测、优雅退出 |

## v1：一条转发一条连接

最笨也最清楚的一版：每来一条本地连接，就往服务端新开一条 TCP，转发完就关。

```bash
# 机器 A（能访问目标服务的那台）
go run ./v1-forward server -addr 127.0.0.1:19000

# 机器 B（你的本机）
go run ./v1-forward client -listen 127.0.0.1:18000 \
    -server 127.0.0.1:19000 -target 127.0.0.1:8000

# 访问本机 18000 == 访问机器 A 视角下的 8000
curl http://127.0.0.1:18000/
```

### 文件

| 文件 | 职责 |
|---|---|
| `main.go` | `server` / `client` 两个子命令，只有 flag 解析和 accept 循环 |
| `tunnel.go` | 握手编解码 + `bridge` 双向桥接，全部可测逻辑都在这 |
| `tunnel_test.go` | 握手单元测试 + 端到端转发测试 |

### 验收标准

- 端到端回显：客户端写什么，经隧道原样回来
- 目标不可达：服务端回 `statusDialFailed` 并关闭，本地读到 EOF 而不是挂死
- 误连识别：对端不是 minibridge（比如连到了 HTTP 服务）时报错点名 `looks like HTTP` 并立刻关闭，不挂死
- 半关闭穿透：客户端 `CloseWrite` 后仍能收到目标端补发的响应
- 并发：20 条连接同时转发互不干扰
- `-race` 下全绿

### v1 故意不做的事

- **不加密、不鉴权**：隧道裸奔，公网可用等于开门揖盗（v3 解决）
- **不限制目标地址**：客户端让服务端连哪就连哪，是个 SSRF 面（v3 加白名单）
- **一条转发一条连接**：连接数与转发数 1:1，长连接场景会打爆 fd（v2 解决）

## v2：一条连接跑 N 条流

命令与 v1 完全一致，变的只是"搬运层"：所有转发复用同一条隧道连接。

```bash
go run ./v2-mux server -addr 127.0.0.1:19000
go run ./v2-mux client -listen 127.0.0.1:18000 -server 127.0.0.1:19000 -target 127.0.0.1:8000

curl http://127.0.0.1:18000/
```

### 帧格式

```text
[ len:4 ][ type:1 ][ streamID:4 ][ payload:N ]
└──────── 9 字节固定头 ────────┘
```

| type | 谁发 | payload | 含义 |
|---|---|---|---|
| `SYN` | 发起方 | 目标地址 | 请求开流 |
| `SYNACK` | 接受方 | 空 | 同意 |
| `RST` | 接受方 | 空 | 拒绝（目标不可达） |
| `DATA` | 双方 | 业务字节 | 数据 |
| `FIN` | 双方 | 空 | 半关闭 |
| `PING`/`PONG` | 双方 | 时间戳 | 保活（v5 启用） |

### 文件

| 文件 | 职责 |
|---|---|
| `frame.go` | 帧定义与编解码 |
| `mux.go` | `Session`（一条 TCP）+ `Stream`（实现 `net.Conn`）+ `bridge` |
| `main.go` | `server` / `client` 子命令 |
| `frame_test.go` / `mux_test.go` | 帧单元测试 + 多路复用端到端测试 |

### 验收标准

- 100 条并发流跑单条连接，逐条回显校验 **不串包**
- 目标不可达 → `RST` → `Open` 立刻返回错误，且**不影响会话上其他流**
- 半关闭穿透：`Stream.CloseWrite()` 后仍能收到目标端补发的响应
- 会话断开 → 所有流一起收到 EOF，不留一堆挂死的连接
- `-race` 下全绿

### 三个设计要点

1. **streamID 分奇偶**：客户端奇数、服务端偶数，双方同时 Open 也不撞号（HTTP/2 同规则）
2. **背压靠阻塞**：`Stream` 的接收队列满了就阻塞投递，压力沿读循环回传到 TCP 窗口，而不是把内存吃光——所以投递时绝不能持锁
3. **按能力断言而非按类型**：`bridge` 用 `interface{ CloseWrite() error }` 收尾，同一段桥接代码既能服务 TCP 连接，也能服务 `Stream`

### v2 故意不做的事

- 仍无加密与鉴权（v3）
- 仍无目标白名单（v3）
- 客户端只建一条会话，断了不会自动重连（v5）
- 接收队列是固定条数，没有窗口通告这种真正的流控（v5）

## v3：给隧道上锁

在 v2 的多路复用之上加三层，顺序不能变：

```text
TCP → TLS 握手（客户端校验服务端证书）→ 认证（可插拔）→ 才接受 SYN → 才 dial 目标
```

```bash
# 生成 CA 与服务端证书（示例用自签；生产请用企业 CA 或 Let's Encrypt）
# 见 certs.go 的 GenerateCA / IssueCert

go run ./v3-crypto server -addr 127.0.0.1:9000 -cert srv.crt -key srv.key -allow 127.0.0.1:8000
go run ./v3-crypto client -listen :8080 -server 127.0.0.1:9000 \
    -target 127.0.0.1:8000 -ca ca.crt -identity ~/.ssh/id_ed25519
```

### 三层各管什么

| 层 | 解决什么 | 关键约束 |
|---|---|---|
| **TLS** | 加密 + 服务端身份 | 客户端必须指定 `RootCAs` + `ServerName`，**绝不能用 `InsecureSkipVerify`**（那样等于没加密） |
| **认证** | 客户端身份 | 可插拔：`sshkey`（默认，挑战-响应，复用 OpenSSH 密钥）或 `mtls`（客户端证书） |
| **授权** | 能拨号到哪 | 目标白名单，规则只收 IP/CIDR，`Resolve` 返回 IP 形式地址 |

### 认证为什么可插拔

```go
type Authenticator interface {
	Name() string
	Server(conn net.Conn) error  // 服务端验证客户端
	Client(conn net.Conn) error  // 客户端证明身份
}
```

| 实现 | 做法 |
|---|---|
| `sshKeyAuth` | 服务端发 32 字节 nonce；客户端用私钥签名；服务端拿 `authorized_keys` 验签 |
| `x509Auth` | **空实现**——mTLS 的校验在 TLS 握手内由 `crypto/tls` 完成，应用层无事可做 |

两种认证的**插入点不同**：mTLS 在 TLS 握手内，SSH 密钥必须在应用层自己做挑战-响应——因为 `crypto/tls` 只认 X.509，SSH 公钥塞不进 TLS 握手。

### 文件

| 文件 | 职责 |
|---|---|
| `certs.go` | 自签 CA + 证书签发（示例/测试用） |
| `allow.go` | 目标白名单：解析规则、返回 IP 形式地址 |
| `auth.go` | `Authenticator` 接口 + 两种实现 + SSH 密钥加载 |
| `main.go` | `server` / `client`，串起 TLS → 认证 → 会话 |
| `crypto_test.go` | 白名单、SSH 密钥认证、TLS 校验、mTLS、端到端 |
| `frame.go` / `mux.go` | 与 v2 相同的多路复用核心（各阶段保持自包含） |

### 验收标准

- 服务端证书由陌生 CA 签发 → 客户端连不上
- 明文客户端（不走 TLS）→ 建不了会话
- `mTLS` 模式下不带客户端证书 → 首次读写必失败（TLS 1.3 的坑：裁决不在 `Dial` 时暴露）
- `sshkey` 模式下用错钥匙 → 认证失败，连接断开
- 白名单外的目标 → 直接 `RST`，一个包都不往外发
- `-race` 下全绿

### v3 故意不做的事

- **自签证书**：只为示例可复现；生产必须换企业 CA 或 ACME
- **不支持带口令的私钥**：请先 `ssh-add` 交给 agent，不在代码里存口令
- **无吊销机制**：`authorized_keys` 改了要重连才生效（v5 加周期性重校验）
- 客户端仍不会自动重连（v5）
