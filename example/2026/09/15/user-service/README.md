# 用户微服务：gin 网关 + gRPC 服务 + gorm 存储

三层串起来的最小用户服务：HTTP 请求 → gin 网关 → gRPC 调用 → gorm/MySQL。前几节的单层 lab 在这里拼成一条完整链路。

对应博客《[从字节流到持久化：Go 后端一条线上的八件事](../../../docs/blog/posts/2026/09/15.md)》第 3/4/5/6 节的组合应用。

| 服务 | 目录 | 讲什么 |
|------|------|--------|
| user_srv | [srv/](./srv/) | gRPC 服务端：用户 CRUD，pbkdf2-sha512 密码加密 |
| user_web | [api/](./api/) | gin 网关：viper 配置、zap 日志、validator 中文校验、gRPC 错误映射 HTTP |

```
curl → api/:6789 (gin) --gRPC--> srv/:8080 (gRPC) --gorm--> MySQL :3306
```

## 快速开始

本目录自带 `go.mod`（gin/gorm/grpc/viper/zap 都是重依赖，与 `example/` 主模块隔离）。需要先起一个 MySQL：

```bash
docker run -d --name mysql-lab -p 3306:3306 -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=db_test mysql:8
```

```bash
cd example/2026/09/15/user-service

# 终端 A：gRPC 用户服务（默认 0.0.0.0:8080）
go run ./srv

# 终端 B：gin 网关（默认 :6789，从模块根目录运行，配置文件路径是 api/config-*.yaml）
go run ./api

# 注册（HTTP 表单 → CreateUser RPC，服务端做密码加密）
curl -X POST localhost:6789/v1/user/register -d 'nick_name=cloaks&mobile=13800000000&password=admin123'
# {"id":1}

# 登录（查用户拿密文 → CheckPasswd 校验）
curl -X POST localhost:6789/v1/user/pwd_login -d 'mobile=13800000000&password=admin123'
# {"msg":"登录成功"}

# 列表（分页）
curl 'localhost:6789/v1/user/list?pageNum=1&pageSize=10'
```

MySQL 不在本地默认地址时，给 `go run ./srv` 加环境变量：

```bash
MYSQL_DSN='root:root@tcp(127.0.0.1:3306)/db_test?charset=utf8mb4&parseTime=True&loc=Local' go run ./srv
```

## 要点

- **proto 只有一份**：两端共用 `proto/`，不像练习仓库那样各复制一份；`user.pb.go` + `user_grpc.pb.go` 是新版 `protoc-gen-go` + `protoc-gen-go-grpc` 分离式生成的，生成命令见 `user.proto`
- **连接只建一次**：网关在启动时 `initialize.SrvConn()` 建好 gRPC 连接（`api/global.UserSrvConn`），HTTP/2 多路复用；每个请求都 `grpc.Dial` 是常见反模式
- **错误映射**：`HandleGrpcErrorToHttp` 把 gRPC `codes.NotFound/InvalidArgument/AlreadyExists/Internal` 翻译成 404/400/400/500
- **登录两段式**：`GetUserByMobile` 拿密文 → `CheckPasswd` 验证，校验逻辑全部留在服务端，网关不碰密码加密
- **校验错误本地化**：`initialize.Trans("zh")` 注册 validator 中文翻译，`removeTopStruct` 把 `PasswordLoginForm.Mobile` 裁成 `mobile`
- **viper 热更新**：`MS_DEBUG=1` 走 `config-debug.yaml`，改配置文件会触发 `OnConfigChange` 重新加载

## 注意

- ⚠️ **列表接口不要原样返回 gRPC 响应**：`UserInfoResponse` 带密码哈希字段，必须转成 `response.UserResponse` 再出去——这是把练习代码搬过来时修掉的一个真实泄漏
- `grpc.Dial(..., insecure.NewCredentials())` 明文 HTTP/2，仅限本地实验，生产必须换 TLS
- DSN 从 `MYSQL_DSN` 环境变量读，代码里的默认值只是本地容器配置
- 端口 8080 / 6789 与其他 lab（9001-9003）不冲突，可以同时跑
