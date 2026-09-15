# gRPC 基础：一元调用、三种流、标准类型

三个可运行的最小示例，覆盖 gRPC 的调用形态与 proto3 的常用类型。

对应博客《[从字节流到持久化：Go 后端一条线上的八件事](../../../docs/blog/posts/2026/09/15.md)》第 3 节「gRPC：把协议换成契约」。

| 示例 | 目录 | 讲什么 |
|------|------|--------|
| 一元调用 | [01-unary](./01-unary/) | `rpc SayHello(Req) returns (Resp)`，最小闭环 |
| 三种流 | [02-stream](./02-stream/) | 服务端流 / 客户端流 / 双向流 |
| 标准类型 | [03-timestamp](./03-timestamp/) | 导入 `google/protobuf/timestamp.proto` + enum |

## 快速开始

本目录自带 `go.mod`（与 `example/` 主模块隔离，避免把 gRPC 依赖带进零依赖的示例模块），命令在该目录下发起：

```bash
cd example/20260915/grpc-basics

# 01 一元调用
go run ./01-unary/server        # 终端 A：监听 127.0.0.1:9001
go run ./01-unary/client        # 终端 B：message:"Hello cloaks"

# 02 三种流（-mode=get|put|all）
go run ./02-stream/server
go run ./02-stream/client -mode=get   # 服务端持续推
go run ./02-stream/client -mode=put   # 客户端持续发
go run ./02-stream/client -mode=all   # 双向

# 03 标准类型
go run ./03-timestamp/server
go run ./03-timestamp/client
```

## 要点

- **生成物已提交**：`*.pb.go` 都在仓库里，不装 protoc 也能直接跑；生成命令写在各自的 `.proto` 注释里
- **端口互不冲突**：9001 / 9002 / 9003，可以同时启动
- **流的方向**：`GetStream` 服务端流、`PutStream` 客户端流、`AllStream` 双向流，服务端三个方法的签名各不相同
- **标准类型**：`timestamppb.New(time.Now())` 把 Go 的 `time.Time` 转成 proto 的 `google.protobuf.Timestamp`；enum 在 Go 侧是常量（`v1.Gender_FEMALE`）

## 注意

- 生成物是旧版 `protoc-gen-go`（v1.26）+ grpc 插件一次性生成的（grpc 桩代码与消息代码在同一个 `.pb.go` 里），与新版 `protoc-gen-go` + `protoc-gen-go-grpc` 分离式生成不同
- 示例统一用 `grpc.WithInsecure()`（明文 HTTP/2），生产环境必须换 TLS 凭据
- gRPC 依赖较多（grpc / protobuf / genproto / x/net 等），首次 `go run` 需要联网拉依赖
