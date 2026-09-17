# TCP 三件套：回声、代理、扫描

三个最小可运行的 TCP 程序，合起来覆盖网络编程的三件基本功：

| 程序 | 目录 | 教什么 |
|------|------|--------|
| 回声服务器 | [echo-server](./echo-server/) | 字节流语义：短读、EOF、三种读法（raw / io.Copy / bufio） |
| TCP 代理 | [proxy](./proxy/) | 双向转发与半关闭：CloseWrite 而不是 Close |
| 端口扫描器 | [scanner](./scanner/) | worker 池 + DialTimeout，并发与超时的配合 |

对应博客《[从字节流到持久化：Go 后端一条线上的八件事](../../../docs/blog/posts/2026/09/15.md)》第 1 节「TCP：一切从字节流开始」。

## 快速开始

`go.mod` 在 `example/` 下，命令需从那里发起：

```bash
cd example

# 1) 回声服务器：三种模式任选
go run ./2026/09/15/tcp-lab/echo-server -mode=bufio -addr=:20080

# 2) 代理：把 :8080 转发到 :20080（先起 echo-server 或任意 TCP 服务）
go run ./2026/09/15/tcp-lab/proxy -listen=:8080 -target=127.0.0.1:20080

# 3) 扫描器：扫本机 1-1024
go run ./2026/09/15/tcp-lab/scanner -host=127.0.0.1 -start=1 -end=1024 -workers=256
```

没有现成的客户端时，用 bash 自带的 `/dev/tcp` 就能连：

```bash
exec 3<>/dev/tcp/127.0.0.1/20080
printf 'hello\n' >&3
head -n 1 <&3        # [conn-1] hello
```

## 三种回声写法

| 模式 | 读法 | 适用 |
|------|------|------|
| `raw` | `conn.Read` + `conn.Write` | 要自己控制缓冲区与边界；必须直面短读与 `n>0 && err==EOF` |
| `iocopy` | `io.Copy(conn, conn)` | 原样转发；`*net.TCPConn` 上走 `splice(2)`，数据不进用户态 |
| `bufio` | `bufio.Reader.ReadString('\n')` | 面向“行/消息”的协议；注意 `Flush` |

## 要点

- **TCP 是字节流**：一次 `Read` 拿到多少字节都不保证，消息边界要靠协议自己定（长度前缀、分隔符、定长）
- **`Read` 的两个返回值要一起看**：`n > 0` 和 `err == io.EOF` 可以同时出现，先处理数据再处理错误
- **代理用半关闭**：`CloseWrite()` 只发 FIN 关掉写方向，读方向保留；直接 `Close()` 会掐断另一个方向
- **goroutine 里不要 `log.Fatal`**：一条连接出错不应该带走整个进程；`log.Fatal` 只留给 `main` 里的致命错误
- **扫描器要有超时**：`net.DialTimeout` 是必须的，否则不可达的端口会把 worker 挂住；结果通道要带缓冲，避免 worker 投递时阻塞

## 注意

- 端口扫描仅用于测试你自己有权限的主机
- 三个程序都是单文件 `main`，改动与试验从各自的 `main.go` 开始
