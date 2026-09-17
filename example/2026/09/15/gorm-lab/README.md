# gorm 上手：连接、模型与 CRUD

两个示例：先验证连接与自定义 logger，再跑一遍 AutoMigrate + CRUD。

对应博客《[从字节流到持久化：Go 后端一条线上的八件事](../../../docs/blog/posts/2026/09/15.md)》第 6 节「gorm：四颗雷」。

| 示例 | 目录 | 讲什么 |
|------|------|--------|
| 连接与 logger | [gorm_01](./gorm_01/) | DSN、`gorm.Config`、自定义 logger、`sql.DB` 健康检查 |
| 模型与 CRUD | [gorm_02](./gorm_02/) | 结构体 tag、`TableName`、AutoMigrate、Create/First/Update/Delete |

## 快速开始

先准备一个 MySQL（示例默认连本地 3306）：

```bash
docker run -d --name demo-mysql \
  -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=db_test \
  -p 3306:3306 mysql:8
```

然后运行（本目录自带 `go.mod`，与 `example/` 主模块隔离）：

```bash
cd example/2026/09/15/gorm-lab

go run ./gorm_01        # 连接 + ping
go run ./gorm_02        # AutoMigrate + 一轮 CRUD

# 连接串可通过环境变量覆盖
DEMO_MYSQL_DSN='user:pass@tcp(127.0.0.1:3306)/db_test?charset=utf8mb4&parseTime=True&loc=Local' \
  go run ./gorm_02
```

## 要点

- **字段必须导出**：`balance int` 这样的小写字段 gorm 会静默忽略，表里根本不会建这一列（示例已改为 `Balance`）
- **列名与表名**：默认列名是字段名的蛇形，表名是结构体名的蛇形复数；用 `gorm:"column:user_name"` 和 `TableName()` 精确控制
- **`parseTime=True`**：不写这个参数，`time.Time` 字段读出来会是 `[]byte`
- **AutoMigrate 只加不减**：它只做“补齐缺失的表/列/索引”，不会删列，适合开发期，不适合当生产迁移工具
- **`First` 查不到返回 `gorm.ErrRecordNotFound`**：用 `errors.Is` 判断，别拿它当严重错误

## 注意

- 示例 DSN 是本地容器的演示配置（`root:root`），不要照搬到线上；真实连接串从环境变量/配置中心读取
- `gorm_02` 每次运行都会清理并重建演示数据，方便反复执行
