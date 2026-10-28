---
date: 2026-10-28
authors:
  - cloaks
categories:
  - Go 语言
tags:
  - Go
  - TCP
  - 隧道
  - 多路复用
  - TLS
  - 网络协议
comments: true
description: minibridge 是把 ssh -L 一路写到底的学习项目——v1 一条转发一条连接，v2 一条 TCP 跑 N 条流，v3 加 TLS、可插拔认证与目标白名单。三阶段源码、帧格式、验收标准，以及每阶段"故意不做的事"都在这里。
hide:
  - navigation
---

<!-- ===== Top Announce Bar ===== -->
<div class="hero-topbar">
  <div class="announce-bar">
    <div class="announce-track">
      <span class="announce-item">✦ minibridge · 从 ssh -L 到自研多路复用</span>
      <span class="announce-item">🔌 v1 转发 · 一条连接一条流</span>
      <span class="announce-item">🧩 v2 帧协议 · 一条 TCP 跑 100 条流</span>
      <span class="announce-item">🔐 v3 TLS 1.3 + sshkey/mTLS + 目标白名单</span>
      <span class="announce-item">🧪 三阶段 42 个测试函数 · -race 全绿</span>
      <span class="announce-item">✦ minibridge · 从 ssh -L 到自研多路复用</span>
      <span class="announce-item">🔌 v1 转发 · 一条连接一条流</span>
      <span class="announce-item">🧩 v2 帧协议 · 一条 TCP 跑 100 条流</span>
      <span class="announce-item">🔐 v3 TLS 1.3 + sshkey/mTLS + 目标白名单</span>
      <span class="announce-item">🧪 三阶段 42 个测试函数 · -race 全绿</span>
    </div>
  </div>
</div>

<!-- ===== Hero ===== -->
<div class="hero hero--center">
  <div class="hero__inner">
    <div class="hero__text">
      <div class="hero__eyebrow">
        <span class="hero__eyebrow-dot"></span>
        <span>Mini Tunnel Bridge</span>
      </div>
      <div class="hero__name">minibridge</div>
      <div class="hero__tags">
        <span class="hero__tag">🔌 手写隧道协议</span>
        <span class="hero__tag">🐹 Go 标准库 + x/crypto</span>
        <span class="hero__tag">🧩 单连接多路复用</span>
        <span class="hero__tag">🔐 TLS 1.3 · 白名单</span>
      </div>
      <div class="hero__quote">从 <code>ssh -L</code> 出发，一步步造出 VSCode Remote-SSH 那种"一条连接多路复用"的桥接能力。每个阶段一个可运行目录，一层层往上加，每层都能看见。</div>
      <div class="hero__actions">
        <a href="https://github.com/yiiewang/yiiewang.github.io/tree/master/example/2026/10/27" target="_blank" class="btn btn--primary">📦 源码目录</a>
        <a href="../../blog/posts/2026/10/27.md" class="btn btn--secondary">📖 v1 文章</a>
        <a href="../../blog/posts/2026/10/27-1.md" class="btn btn--secondary">📖 v2 文章</a>
      </div>
    </div>
  </div>
</div>

## 它是什么

一个**把端口转发写到底**的学习项目。目标不是替代 SSH，而是把"SSH 当加密网线、两端各挂一个桥接端点"这件事拆开，一层一层自己实现，直到能在一 TCP 上跑成百上千条流、还能上锁。

三阶段都在同一个 Go module 下，阶段之间代码可复用，命令形态也保持一致：

```bash
cd example/2026/10/27
go test -count=1 -race ./...   # 跑全部阶段的测试
```

## 三阶段

<div class="card-grid">

  <a href="https://github.com/yiiewang/yiiewang.github.io/tree/master/example/2026/10/27/v1-forward" target="_blank" class="card card--link card--pad">
    <div class="card__head">
      <div class="icon-box blue">🔌</div>
      <div class="card__title">v1 · 最简转发</div>
    </div>
    <div class="card__body">
      <div class="card__desc">一条转发独占一条 TCP：握手告诉服务端"去连谁"，<code>io.Copy</code> 双向桥接，收尾用半关闭而不是直接 Close。最笨，也最好懂。</div>
      <span class="badge badge--success">✅ 完成</span>
    </div>
  </a>

  <a href="https://github.com/yiiewang/yiiewang.github.io/tree/master/example/2026/10/27/v2-mux" target="_blank" class="card card--link card--pad">
    <div class="card__head">
      <div class="icon-box violet">🧩</div>
      <div class="card__title">v2 · 帧协议多路复用</div>
    </div>
    <div class="card__body">
      <div class="card__desc">自定义 9 字节帧头，<code>Session</code> 管一条 TCP、<code>Stream</code> 实现 <code>net.Conn</code>，100 条流并发跑同一条连接且不串包。</div>
      <span class="badge badge--success">✅ 完成</span>
    </div>
  </a>

  <a href="https://github.com/yiiewang/yiiewang.github.io/tree/master/example/2026/10/27/v3-crypto" target="_blank" class="card card--link card--pad">
    <div class="card__head">
      <div class="icon-box orange">🔐</div>
      <div class="card__title">v3 · 加密与认证</div>
    </div>
    <div class="card__body">
      <div class="card__desc">TLS 1.3 打底，认证可插拔（sshkey 挑战-响应 / mTLS 客户端证书），目标白名单治 SSRF。顺序是铁律：先认证，才接受开流。</div>
      <span class="badge badge--success">✅ 完成</span>
    </div>
  </a>

</div>

## 路线图

| 阶段 | 目录 | 状态 | 学到什么 | 文章 |
|---|---|---|---|---|
| v1 最简转发 | `v1-forward` | ✅ 完成 | 监听/拨号、握手协议、`io.Copy` 双向桥接、半关闭 | [27](../../blog/posts/2026/10/27.md) |
| v2 帧协议多路复用 | `v2-mux` | ✅ 完成 | 帧格式、stream 生命周期、单连接跑多条流 | [27-1](../../blog/posts/2026/10/27-1.md) |
| v3 加密与认证 | `v3-crypto` | ✅ 完成 | TLS 校验证书、挑战-响应、目标白名单 | 代码先行 |
| v4 反向转发与控制面 | `v4-control` | ⏳ 计划 | 反向通道语义、运行时注册/注销端口 | — |
| v5 生产化 | `v5-ops` | ⏳ 计划 | 心跳保活、断线重连、端口探测、优雅退出 | — |

v4 不需要改结构：v2 已经把 `Open` / `Accept` 做成对称的，服务端反向 Open 直接接上即可。

---

## v1：一条转发一条连接

```bash
# 机器 A（能访问目标服务的那台）
go run ./v1-forward server -addr 127.0.0.1:19000

# 机器 B（你的本机）
go run ./v1-forward client -listen 127.0.0.1:18000 \
    -server 127.0.0.1:19000 -target 127.0.0.1:8000

curl http://127.0.0.1:18000/   # 访问本机 18000 == 访问机器 A 视角下的 8000
```

| 文件 | 职责 |
|---|---|
| `main.go` | `server` / `client` 两个子命令，只有 flag 解析和 accept 循环 |
| `tunnel.go` | 握手编解码 + `bridge` 双向桥接，全部可测逻辑都在这 |
| `tunnel_test.go` | 握手单元测试 + 端到端转发测试 |

**握手格式**：请求 `[MBRD:4][ver:1][len:2][target:N]`，响应 `[MBRD:4][ver:1][status:1]`。4 字节 magic 只为一个理由——**认出对端不是自己人**。没有它时，客户端会把 HTTP 响应的首字节 `'H' = 72` 当成状态位，日志里只留下要推理半天的 `status=72`。

**验收标准**

- 端到端回显：客户端写什么，经隧道原样回来
- 目标不可达：服务端回 `statusDialFailed` 并关闭，本地读到 EOF 而不是挂死
- 误连识别：连到 HTTP 服务时报错点名 `looks like HTTP` 并立刻关闭
- 半关闭穿透：客户端 `CloseWrite` 后仍能收到目标端补发的响应
- 并发：20 条连接同时转发互不干扰
- `-race` 下全绿

**v1 故意不做的事**：不加密不鉴权（隧道裸奔）、不限制目标地址（SSRF 面）、连接数与转发数 1:1。前两条归 v3，最后一条归 v2。

---

## v2：一条连接跑 N 条流

命令与 v1 完全一致，变的只是搬运层——所有转发复用同一条隧道连接。

**帧格式**

```text
[ len:4 ][ type:1 ][ streamID:4 ][ payload:N ]
└──────── 9 字节固定头 ────────┘
```

| type | 谁发 | payload | 含义 |
|---|---|---|---|
| `SYN` | 发起方 | 目标地址 | 请求开流 |
| `SYNACK` | 接受方 | 空 | 同意 |
| `RST` | 接受方 | 空 | 拒绝（目标不可达） |
| `DATA` | 双方 | 业务字节 | 数据（单帧上限 64KB） |
| `FIN` | 双方 | 空 | 半关闭 |
| `PING` / `PONG` | 双方 | 时间戳 | 保活（v5 启用） |

**三个对象**：`Session` 持有唯一一条 TCP，一个读 goroutine 按 streamID 分发；`Stream` 是一条逻辑流，实现 `net.Conn`——这样 v1 那套 `io.Copy` 桥接代码可以原样复用；`bridge` 把本地连接和 `Stream` 接起来。

**四个必须踩住的坑**

1. **streamID 分奇偶**：客户端奇数、服务端偶数，双方同时 Open 也不撞号（HTTP/2 同规则）
2. **背压靠阻塞**：接收队列（容量 16）满了就阻塞投递，压力沿读循环回传到 TCP 窗口——所以投递时**绝不能持锁**
3. **流的清理**：流结束必须从 `streams` 摘掉；会话断开时所有流一起收尾，否则上层连接集体挂死
4. **按能力断言而非按类型**：`bridge` 用 `interface{ CloseWrite() error }` 收尾，同一段代码既能服务 TCP 连接，也能服务 `Stream`

**验收标准**：100 条并发流跑单条连接不串包；目标不可达 → `RST` → `Open` 立刻返回错误且不影响其他流；半关闭穿透；会话断开 → 所有流一起 EOF。

**v2 故意不做的事**：仍无加密与鉴权、仍无目标白名单、客户端只建一条会话断了不重连、接收队列固定条数没有窗口通告。前两条 v3，后两条 v5。

---

## v3：给隧道上锁

```bash
# 服务端：证书 + 白名单（只允许连 127.0.0.1:8000）
go run ./v3-crypto server -addr :9000 -cert srv.crt -key srv.key \
    -allow 127.0.0.1:8000

# 客户端：校验证书的 CA + 身份证明
go run ./v3-crypto client -listen :8080 -server host:9000 \
    -target 127.0.0.1:8000 -ca ca.crt -identity ~/.ssh/id_ed25519
```

### 顺序是铁律

```text
TLS 握手 → 认证 → 才接受 SYN → 才 dial 目标
```

认证没过就绝不能让对端驱动我们 dial——否则等于给陌生人留了一个 SSRF 入口。所以 `Authenticator` 的位置是：**TLS 之后，第一个 SYN 之前**。

### 认证可插拔

| 方式 | flag | 怎么证明身份 |
|---|---|---|
| `sshkey`（默认） | `-identity ~/.ssh/id_ed25519` | 服务端发 32 字节随机 nonce，客户端用私钥签名，服务端拿 `authorized_keys` 里的公钥验签 |
| `mtls` | `-cert` `-key` + 服务端 `-client-ca` | 客户端证书校验在 TLS 握手内完成，应用层两个方法都是空实现 |

`sshkey` 用的不是 SSH 协议，只是 SSH 的**密钥格式与签名算法**；挑战-响应本身与通道用什么加密无关。mTLS 那条"空实现"能成立的前提是服务端设了 `ClientAuth: RequireAndVerifyClientCert` 并配好 `ClientCAs`——少了这两行，它就等于"谁都能连"。

### 白名单：治 SSRF，也治 DNS rebinding

规则只接受 **IP 或 CIDR**，不接受域名——域名会在"校验"和"真正拨号"之间留下二次解析的缝。目标写域名时先解析，命中的那个 IP 会被**返回并直接拿去拨号**，校验和拨号用的是同一个地址。

```text
127.0.0.1:8000     精确
127.0.0.1:*        该 IP 的任意端口
10.0.0.0/8:3306    整个网段 + 指定端口
```

### 证书从哪来

`certs.go` 造自签 CA 与叶子证书，只供示例和测试使用。生产环境的正确做法是企业 CA 或 Let's Encrypt；自签的差别只在"谁来证明这张证书可信"，协议本身完全一样。注意 `ExtKeyUsage`：服务端要 `ServerAuth`、客户端要 `ClientAuth`，少了这一项握手会以 `incompatible key usage` 失败。

**验收标准**：`-ca` 不匹配 → 握手失败；明文客户端连 TLS 服务端 → 失败；mTLS 模式无客户端证书 → 拒绝；端到端转发成功；白名单外的目标 → `RST`，一个包都不往外发。

**v3 故意不做的事**：流的 deadline 暂不支持（`SetDeadline` 直接返回错误，不静默忽略）、客户端断线不重连、没有心跳——都留给 v5。

---

## 测试

三个阶段的测试就是各阶段的不变量清单，`-race` 下全绿：

```bash
cd example/2026/10/27 && go test -count=1 -race ./...
```

| 阶段 | 关键用例 |
|---|---|
| v1 | 握手往返与跨包读取、超长目标拒绝、版本/ magic 校验、`looks like HTTP` 识别、端到端回显、不可达目标、半关闭穿透、20 并发 |
| v2 | 帧往返与超大帧拒绝、流回显、`RST` 后会话存活、100 条流并发不串包、半关闭、会话关闭带动所有流 |
| v3 | 白名单解析与域名规则拒绝、sshkey 往返与未知密钥拒绝、不可信 CA 拒绝、明文客户端拒绝、mTLS 强制客户端证书、加密隧道端到端转发、白名单外目标拒绝 |

## 技术栈

<div class="chips">
  <span class="chip"><span class="chip__icon">🐹</span>Go 1.24</span>
  <span class="chip"><span class="chip__icon">🌐</span>net</span>
  <span class="chip"><span class="chip__icon">🔐</span>crypto/tls 1.3</span>
  <span class="chip"><span class="chip__icon">🔑</span>x/crypto/ssh</span>
  <span class="chip"><span class="chip__icon">🧾</span>crypto/x509</span>
  <span class="chip"><span class="chip__icon">🧪</span>testing · -race</span>
  <span class="chip"><span class="chip__icon">🧵</span>goroutine + channel</span>
</div>

<div class="page-footer">
  <p class="page-footer__title">每一层都能看见，才敢说自己懂了</p>
  <p class="page-footer__sub">源码：<a href="https://github.com/yiiewang/yiiewang.github.io/tree/master/example/2026/10/27" target="_blank">example/2026/10/27</a> · 维护人：yiiewang</p>
</div>

---

*最后更新：2026-10-28（v1 / v2 / v3 已完成）*
