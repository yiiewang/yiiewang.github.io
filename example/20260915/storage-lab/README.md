# 存储三件套：WAL、LevelDB、MySQL 二进制列

三个存储组件的动手示例，都用测试驱动，`go test` 就能跑。

对应博客《[从字节流到持久化：Go 后端一条线上的八件事](../../../docs/blog/posts/2026/09/15.md)》第 7 节「存储三件套：WAL、LevelDB、MySQL」。

| 组件 | 目录 | 讲什么 |
|------|------|--------|
| WAL | [wal](./wal/) | `tidwall/wal` 的写、读、截断、重开 |
| LevelDB | [leveldb](./leveldb/) | `goleveldb` 基本读写删 + 逐 key 比对两个库 |
| MySQL | [mysql](./mysql/) | `database/sql` 读写二进制列、SQL 脚本切分 |

## 快速开始

本目录自带 `go.mod`（与 `example/` 主模块隔离）：

```bash
cd example/20260915/storage-lab

go test ./...                 # 全部（MySQL 不可达时自动 skip）

go test -v ./wal/             # 单独跑
go test -v ./leveldb/
go test -v ./mysql/           # 需要 MySQL，见下
```

MySQL 用例默认连本地 3306，连不上会 skip：

```bash
docker run -d --name demo-mysql \
  -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=db_test \
  -p 3306:3306 mysql:8

DEMO_MYSQL_DSN='root:root@tcp(127.0.0.1:3306)/db_test?parseTime=true' go test -v ./mysql/
```

## 要点

- **WAL 的 index 是绝对位置**：从 1 开始严格递增，`TruncateFront/TruncateBack` 按 index 丢弃两端；关掉再打开数据仍在
- **LevelDB 的 `Get` 查不到返回 `ErrNotFound`**：不是 nil 错误，`errors.Is` 判断最稳
- **逐 key 比对**：迭代器遍历 + value 的 md5，用来校验数据复制/迁移后的完整性
- **二进制列要看 hex**：`[]byte` 直接打印是乱码，hex + ASCII 一起看才能确认内容
- **SQL 脚本切分要带 `(?m)`**：`;(?:\s*)$` 少了多行标志，只会在整个脚本末尾切一刀

## 注意

- 测试数据目录统一用 `t.TempDir()`，跑完自动清理，不会往仓库里写数据文件
- 连接串从 `DEMO_MYSQL_DSN` 读取，示例默认值只适用于本地容器
- 依赖需要联网拉取（`tidwall/wal`、`goleveldb`、`go-sql-driver/mysql`、`testify`）
