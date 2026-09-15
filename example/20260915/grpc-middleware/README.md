# gRPC 中间件：元数据、拦截器、鉴权、错误与超时

四个示例覆盖 gRPC 的“横切面”能力：请求带什么、在哪拦截、怎么拒绝、出错怎么表达。

对应博客《[从字节流到持久化：Go 后端一条线上的八件事](../../../docs/blog/posts/2026/09/15.md)》第 4 节「gRPC 的横切面：元数据、拦截器、鉴权、超时」。

| 示例 | 目录 | 讲什么 |
|------|------|--------|
| 元数据 | [01-metadata](./01-metadata/) | `metadata.NewOutgoingContext` 传参、服务端读 incoming metadata |
| 拦截器 | [02-interceptor](./02-interceptor/) | 客户端计时拦截器 + 服务端日志拦截器 |
| 鉴权 | [03-auth](./03-auth/) | `PerRPCCredentials` 携带 APPID/APPKEY，服务端拦截器校验 |
| 错误与超时 | [04-error-timeout](./04-error-timeout/) | `status.Error` 带 code，`context.WithTimeout` 控制超时 |

## 快速开始

本目录自带 `go.mod`（与 `example/` 主模块隔离），命令在该目录下发起：

```bash
cd example/20260915/grpc-middleware

# 01 元数据（:9011）
go run ./01-metadata/server
go run ./01-metadata/client

# 02 拦截器（:9012）
go run ./02-interceptor/server
go run ./02-interceptor/client        # 客户端会打印本次调用耗时

# 03 鉴权（:9013）—— 凭据通过环境变量注入，默认 demo-appid/demo-appkey
go run ./03-auth/server
go run ./03-auth/client                                  # 通过
DEMO_APPKEY=wrong go run ./03-auth/client                # Unauthenticated

# 04 错误与超时（:9014）
go run ./04-error-timeout/server
go run ./04-error-timeout/client -name=gopher   # ok
go run ./04-error-timeout/client -name=slow     # DeadlineExceeded（服务端睡 5s，客户端 3s 超时）
go run ./04-error-timeout/client -name=         # InvalidArgument
```

## 要点

- **metadata 的 key 会被转成小写**：客户端写 `"APPID"`，服务端要用 `md.Get("appid")` 读
- **拦截器是每个请求的必经之路**：鉴权、日志、耗时统计放这里，业务实现不用关心
- **`PerRPCCredentials` 的两个方法**：`GetRequestMetadata` 提供凭据，`RequireTransportSecurity` 声明是否要求 TLS（本地示例返回 false）
- **错误要带 code**：`status.Error(codes.InvalidArgument, "...")`，客户端用 `status.FromError` 还原 code 与 message
- **超时用 context**：`context.WithTimeout` + `defer cancel()`；`cancel` 漏掉会造成 context 泄漏

## 注意

- 演示凭据（`demo-appid` / `demo-appkey`）是占位符，可用 `DEMO_APPID` / `DEMO_APPKEY` 覆盖；真实项目从配置中心读取，不要写进代码
- 服务端拦截器里**不要 panic**：handler 的 error 原样返回即可，panic 会拖垮整个进程
- `04-error-timeout` 服务端在慢请求里 `select ctx.Done()`，客户端放弃后立即返回，不做无用功
