# minibridge：手写一个隧道协议（学习路线）

从 `ssh -L` 出发，一步步造出 VSCode Remote-SSH 那种"一条连接多路复用"的桥接能力。每个阶段一个可运行目录，配一篇博客。

代码在同一个 Go module 下，阶段之间可以复用：

```bash
cd example/minibridge
go test -count=1 -race ./...   # 跑全部阶段的测试
```

## 路线图

| 阶段 | 目录 | 状态 | 学到什么 |
|---|---|---|---|
| v1 最简转发 | `v1-forward` | ✅ 完成 | TCP 监听/拨号、握手协议、`io.Copy` 双向桥接、半关闭 |
| v2 帧协议多路复用 | `v2-mux` | ✅ 完成 | 帧格式设计、stream 生命周期、单连接跑多条流 |
| v3 加密与认证 | `v3-crypto` | ✅ 完成 | TLS、双向证书校验（mTLS）、"先认证后转发"的顺序、目标白名单 |
| v4 心跳与重连 | `v4-keepalive` | ✅ 完成 | 应用层心跳判死、指数退避重连、连续失败退出交给 systemd |
| v5 反向转发 | `v5-reverse` | ✅ 完成 | 反向通道语义、声明式注册表、运行时管理面（admin） |

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
| `PING`/`PONG` | 双方 | 时间戳 | 保活（v4 启用） |

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
- 客户端只建一条会话，断了不会自动重连（v4）
- 接收队列是固定条数，没有窗口通告这种真正的流控（未做）

## v3：给隧道上锁

在 v2 的多路复用之上加三层，顺序不能变：

```text
TCP → TLS 握手（客户端校验服务端证书 + 服务端校验客户端证书）→ 才接受 SYN → 才 dial 目标
```

```bash
# 生成 CA、服务端证书、客户端证书（示例用自签；生产请用企业 CA 或 ACME）
# 在 minibridge/ 下执行，证书落在 minibridge/certs/（已 gitignore），三个版本共用
make certs
# 可选：客户端证书绑定来源 IP——整套证书+私钥被拷到别的机器（来源 IP
# 不在证书 SAN 里）就被服务端断开；留空不绑定，老证书不受影响
make certs CLI_IP=10.0.0.5

go run ./v3-crypto server -addr 127.0.0.1:9000 \
    -cert certs/srv.crt -key certs/srv.key \
    -client-ca certs/ca.crt -allow 127.0.0.1:8000
go run ./v3-crypto client -listen :8080 -server 127.0.0.1:9000 \
    -target 127.0.0.1:8000 -ca certs/ca.crt \
    -cert certs/cli.crt -key certs/cli.key
```

注意 `-cert` / `-key` 两端同名不同物：服务端那是**服务端自己的**，客户端那是**客户端自己的**。

### 三层各管什么

| 层 | 解决什么 | 关键约束 |
|---|---|---|
| **TLS** | 加密 + 服务端身份 | 客户端必须指定 `RootCAs` + `ServerName`，**绝不能用 `InsecureSkipVerify`**（那样等于没加密） |
| **认证** | 客户端身份 | mTLS：`ClientAuth: RequireAndVerifyClientCert` + `ClientCAs`，握手内完成 |
| **授权** | 能拨号到哪 | 目标白名单，规则只收 IP/CIDR，`Resolve` 返回 IP 形式地址 |

### 认证为什么不在应用层

客户端认证整个交给 TLS 握手，应用层一个字都不用写：

```go
cfg.ClientCAs  = loadCertPool(clientCA)
cfg.ClientAuth = tls.RequireAndVerifyClientCert
```

这两行就是全部的认证逻辑。相比在应用层自己做挑战-响应，它多一个硬保证：
**不存在"认证没过但已经能驱动服务端 dial"的窗口**——校验失败的连接在
`Accept` 之前就被 `crypto/tls` 丢掉了，压根到不了我们的代码。

代价是要养一套客户端证书：签发、分发、过期、吊销。这是 mTLS 的标准成本，
换来的是第三方系统能直接对接，以及证书里能带身份信息做审计。

### 文件

| 文件 | 职责 |
|---|---|
| `certs.go` | 自签 CA + 证书签发（示例/测试用） |
| `allow.go` | 目标白名单：解析规则、返回 IP 形式地址 |
| `main.go` | `server` / `client`，串起 TLS → 会话（认证在 TLS 握手里做完了） |
| `Makefile` | 见上级 `minibridge/Makefile`：`make certs` 用 openssl 签发证书（三版本共用） |
| `crypto_test.go` | 白名单、TLS 校验、mTLS 必须有客户端证书、端到端 |
| `frame.go` / `mux.go` | 与 v2 相同的多路复用核心（各阶段保持自包含） |

### 验收标准

- 服务端证书由陌生 CA 签发 → 客户端连不上
- 明文客户端（不走 TLS）→ 建不了会话
- 不带客户端证书 → `Handshake` 拿不到 PONG，客户端直接退出（TLS 1.3 的坑：`tls.Dial` 成功不代表认证通过）
- 带上 `-client-ca` 签发的客户端证书 → 完整走通一次转发
- 白名单外的目标 → 直接 `RST`，一个包都不往外发
- `-race` 下全绿

### v3 故意不做的事

- **自签证书**：只为示例可复现；生产必须换企业 CA 或 ACME
- **无吊销机制**：没接 CRL / OCSP，客户端证书只能等它过期
- 客户端仍不会自动重连（v4 已做）

## v4：从能跑到敢一直开着

```bash
# 命令与 v3 完全一致——心跳与重连都是默认行为，没有开关
# 服务端（机器 A，能访问目标服务）
go run ./v4-keepalive server -addr 127.0.0.1:9000 \
    -cert certs/srv.crt -key certs/srv.key -client-ca certs/ca.crt -allow 127.0.0.1:8000

# 客户端（机器 B，你的本机）
go run ./v4-keepalive client -listen :8080 -server 127.0.0.1:9000 \
    -target 127.0.0.1:8000 -ca certs/ca.crt -cert certs/cli.crt -key certs/cli.key
```

三件事：

| 事 | 怎么做 |
|---|---|
| 死了知道 | 30s 一拍 PING，单拍 10s 内没等到 PONG 算丢，连丢 3 拍判死（约 60~90s）；两端互发互回 |
| 死了会换 | 一条 TLS 连接断了就是断了：重连立即首试，失败再 1s 起步指数退避、封顶 30s；实测 kill 服务端 3 秒内自动恢复 |
| 换不动会喊人 | 连续 8 次失败非零退出，交给 systemd 拉起；首连失败仍然 Fatal（启动期错误重试无意义） |

细节：心跳和握手各守一条 PONG 通道——共用一条的话，心跳会拿握手的旧回包冒充本次探测结果，判死慢一整拍。

### 验收标准

- 静置 30 分钟仍能 `Open`
- kill 服务端后客户端自动恢复、无需手工重启
- 重连 8 次失败 → 非零退出码
- `-race` 下全绿

## v5：反向转发——把内网服务发布到公网

前面都是正向：客户端把本机流量送到服务端侧。v5 反过来：服务端监听公网端口，
连上来的流量打回客户端侧的目标。场景：机器 A 在内网跑了个服务，机器 B 在公网，
公网用户连 B 的 18000，等于访问 A 视角下的 8000。

```bash
# 服务端（公网机器 B）：与 v3/v4 完全一致
go run ./v5-reverse server -addr 127.0.0.1:9000 \
    -cert certs/srv.crt -key certs/srv.key -client-ca certs/ca.crt \
    -allow 127.0.0.1:8000

# 客户端（内网机器 A）：-reverse 反向发布，-bind 是 SSH -R 风格的端口映射
go run ./v5-reverse client -server 127.0.0.1:9000 -bind 127.0.0.1:8001:127.0.0.1:80 \
    -ca certs/ca.crt -cert certs/cli.crt -key certs/cli.key -reverse 

# 访问 18000 == 访问 A 视角下的 8000（上面两条命令本机即可复现）
curl http://127.0.0.1:18000

# 真实部署两件事：
#   1. server 的 -addr 与 -bind 的监听地址换成 0.0.0.0（-bind 0.0.0.0:18000:127.0.0.1:8000）
#   2. 服务端证书的 SAN 必须覆盖客户端实际连接的地址（默认只签了 127.0.0.1）：
#      make certs SRV_SAN=IP:<B的公网IP>
```

要点：

- **数据面零新协议**：Session 本来就是对称的（v2 埋的），服务端在 bind 上 accept
  用户连接，`Open(target)` 把流开回客户端，客户端 Accept 后按 target 拨号再桥接——
  用的就是正向转发那同一段代码
- **注册表在客户端**：反向清单存在客户端、按会话重放；服务端不持久化任何状态，
  会话死了它名下的监听自动清零，重连后整体重放，不怕两端漂移
- **客户端侧白名单**：`-allow`（默认只放行回环）——反向意味着服务端能驱动本端
  dial，这道闸门与服务端的 `-allow` 对称
- **管理面**：`-admin 127.0.0.1:9090` 开一个行协议接口，运行时增删反向转发不用重启：

```bash
echo 'LISTEN 127.0.0.1:18001 127.0.0.1:8001' | nc 127.0.0.1 9090  # 加一条
echo 'LIST' | nc 127.0.0.1 9090                                    # 列出清单
echo 'UNLISTEN 127.0.0.1:18001' | nc 127.0.0.1 9090                # 删一条
```

能下 LISTEN 就等于拿到了本机的拨号权，admin 只该绑回环。

### 文件

| 文件 | 职责 |
|---|---|
| `reverse.go` | 反向转发数据面：bind 上 accept、流开回客户端 |
| `admin.go` | 声明式注册表 + 本地管理面（LISTEN / UNLISTEN / LIST） |
| `tunnel.go` / `mux.go` | v4 的重连循环与心跳，原样继承 |

### 验收标准

- 反向 echo 全链路通（`TestReverseEcho`）；客户端白名单外的目标被拒（`TestReverseDeniedByClientAllow`）
- 会话死亡 → 它名下的监听全部关闭（`TestReverseListenersDieWithSession`），重连后清单整体重放（`TestAdminResyncAfterReconnect`）
- 重复 bind 被拒、UNLISTEN 后端口释放（`TestReverseDuplicateBindRejected` / `TestReverseUnlisten`）
- 客户端证书 IP 绑定生效（`TestClientCertIPBinding`）
- admin 绑非回环直接拒绝启动（`TestAdminAddrIsLoopback`）

### 安全边界：防什么、不防什么

这套东西防的是**路上的攻击者**（窃听、篡改、中间人、无证书接入），防不了**端点上的攻击者**。按现实程度排：

| 攻击面 | 现状 | 对策 |
|---|---|---|
| 端点机器被入侵 | 客户端/服务端沦陷即可读走私钥冒充合法端 | 机器自身安全，超出本项目范围 |
| `ca.key` 泄露 | 可签任意"合法"客户端证书 | 只留在签发机，**绝不随 certs/ 目录分发**；客户端只发 `cli.crt/key + ca.crt` |
| 证书被盗 | 无 CRL/OCSP，无法吊销 | 默认 `DAYS=90` 短周期压风险窗口 |
| admin 暴露 | LISTEN 等于本机拨号权 | 代码强制回环，非回环拒绝启动 |
| 白名单写宽 | `-allow` 放开即开放代理 | 保持最小化，规则只收 IP/CIDR |
| NAT 后多机同 IP | IP 绑定按来源 IP 判，NAT 内一台被黑整 NAT 覆盖 | 无解（证书 SAN 表达不了 CIDR），换独立证书 |
| DoS | 应用层资源放大已封顶（`-max-conns` 默认 64，满则关连接不做握手）；SYN 洪水归内核/防火墙 | 需要更强时在防火墙层做 |
| 密码学/实现缺陷 | TLS 1.3 + Go crypto/tls，无已知实际攻击 | 依赖上游更新 |

信道本身（TLS 1.3 + mTLS + 前向保密）在当前密码学下是硬的；攻击者会绕开它，去打上表。
