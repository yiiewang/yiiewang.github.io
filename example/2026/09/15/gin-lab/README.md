# gin 速查：路由、分组、参数绑定与校验

五个最小示例，覆盖 gin 日常开发用得最多的五件事。

对应博客《[从字节流到持久化：Go 后端一条线上的八件事](../../../docs/blog/posts/2026/09/15.md)》第 5 节「gin：路由、参数、校验」。

| 示例 | 目录 | 端口 | 讲什么 |
|------|------|------|--------|
| 最小路由 | [gin_01](./gin_01/) | 9021 | `gin.Default()` + `ctx.JSON` |
| 路由分组 | [gin_02](./gin_02/) | 9022 | `router.Group` 与处理器函数 |
| 路径参数 | [gin_03](./gin_03/) | 9023 | `ShouldBindUri` + `binding:"required"` |
| query 与表单 | [gin_04](./gin_04/) | 9024 | `DefaultQuery` / `PostForm` |
| 参数校验 | [gin_05](./gin_05/) | 9025 | `binding` tag：必填、长度、邮箱、数值区间 |

## 快速开始

本目录自带 `go.mod`（与 `example/` 主模块隔离），命令在该目录下发起：

```bash
cd example/2026/09/15/gin-lab

go run ./gin_01        # 其他示例同理：gin_02 ... gin_05

# 验证
curl -s localhost:9021/hello
curl -s localhost:9023/good/42/on
curl -s localhost:9024/welcome?firstName=cloaks
curl -s -X POST localhost:9025/loginJSON \
  -H 'Content-Type: application/json' -d '{"user":"cloaks","password":"123"}'
```

## 要点

- **`gin.Default()` = Engine + Logger + Recovery**：生产环境想要更干净的日志可以 `gin.New()` 自己挂中间件
- **分组只是前缀**：`router.Group("/good")` 下挂的中间件只作用于该组
- **`ShouldBind` 自动选解析器**：按 `Content-Type` 决定走 JSON 还是表单；URI 参数要用 `ShouldBindUri`
- **校验规则写在 tag 里**：`required,min=3,max=10,gte=1,lte=130,email`，失败时 `err.Error()` 能直接告诉调用方哪个字段不合法

## 注意

- 校验失败应返回 `400`，别把 `err.Error()` 当成功响应返回
- 示例端口各不相同，可以同时启动对照
- gin 的依赖较多（validator、sse、go-json 等），首次 `go run` 需要联网拉依赖
