# 手写 RPC：五级台阶

用标准库从零搭一条 RPC 的最短路径，五个台阶，每级只解决一个问题：

| 台阶 | 目录 | 解决的问题 |
|------|------|-----------|
| 01 | [01-http-rpc](./01-http-rpc/) | HTTP + JSON 手动挡：把一次请求包装成函数调用 |
| 02 | [02-netrpc](./02-netrpc/) | 交给 `net/rpc`：方法签名约定 + gob 编解码 |
| 03 | [03-jsonrpc](./03-jsonrpc/) | 换编解码器：gob → JSON，传输层不变 |
| 04 | [04-http-jsonrpc](./04-http-jsonrpc/) | 换传输层：裸 TCP → HTTP，服务实现不变 |
| 05 | [05-stub](./05-stub/) | 把“注册”和“调用”各包一层 stub，调用方拿到本地对象 |

对应博客《[从字节流到持久化：Go 后端一条线上的八件事](../../../docs/blog/posts/2026/09/15.md)》第 2 节「手写 RPC：五级台阶看清协议三要素」。

## 快速开始

`go.mod` 在 `example/` 下，命令需从那里发起。每个台阶都是「先起 server，再跑 client」：

```bash
cd example

# 01: HTTP + JSON
go run ./2026/09/15/rpc-series/01-http-rpc/server   # 终端 A
go run ./2026/09/15/rpc-series/01-http-rpc/client   # 终端 B → 3

# 02: net/rpc（gob 编码，裸 TCP）
go run ./2026/09/15/rpc-series/02-netrpc/server
go run ./2026/09/15/rpc-series/02-netrpc/client     # hello,cloaks

# 03: net/rpc + JSON 编码
go run ./2026/09/15/rpc-series/03-jsonrpc/server
go run ./2026/09/15/rpc-series/03-jsonrpc/client

# 04: JSON-RPC over HTTP（可以直接用 curl 调）
go run ./2026/09/15/rpc-series/04-http-jsonrpc/server
curl -s -X POST localhost:8080/jsonrpc \
  -d '{"method":"HelloService.Hello","params":["cloaks"],"id":0}'
# {"id":0,"result":"hello,cloaks","error":null}

# 05: client_stub / server_stub
go run ./2026/09/15/rpc-series/05-stub/server
go run ./2026/09/15/rpc-series/05-stub/client
```

> 五个台阶都复用 `:8080`，逐个跑即可；不要同时启动两个 server。

## 各台阶要点

- **01**：`callID` 用 URL path、参数走 query、回包是 JSON —— 这就是一份手写的“协议”。客户端把协议细节藏进 `Add()`，调用方看不出背后有网络
- **02**：`net/rpc` 规定方法必须满足 `func (t *T) Method(req T1, resp *T2) error`；默认 gob 编码，gob 是 Go 私有的二进制格式
- **03**：只换 `ServeCodec`/`NewClientWithCodec` 里的编解码器，服务实现一行不动；代价是 JSON 比 gob 大且慢
- **04**：`rpc.ServeRequest` 需要一个 `io.ReadWriteCloser`——请求体当读端、`ResponseWriter` 当写端，拼起来就能把一次 HTTP 请求喂给它
- **05**：`server_stub` 包注册、`client_stub` 包调用，服务名收敛到 `handler.HelloServiceName` 一个常量；调用方最终看到的是 `client.Hello(...)`，与本地方法调用无异

## 注意

- 01/04 的客户端用标准库 `net/http` 重写过（原稿用了第三方 HTTP 客户端），保证整组示例零外部依赖
- 04 的请求结构体字段必须导出，否则 `json.Marshal` 只会得到 `{}`
- 所有 server 都用 `log.Fatalf` 处理致命错误：`ListenAndServe` 返回非 nil 时进程就该退出
