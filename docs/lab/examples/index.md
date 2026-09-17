---
title: 示例代码
comments: true
description: 博客全部可运行示例代码的目录：example/ 路径与文章同构（YYYY/MM/DD-序号），每个示例标注对应文章、源码位置与运行命令，go build / vet / race 全绿。
---

# 示例代码

博客里出现过的代码，可运行版本全部收在仓库 [example/](https://github.com/yiiewang/yiiewang.github.io/tree/master/example) 目录——**路径与文章一一同构**：`example/YYYY/MM/DD-序号/` 对应 `docs/blog/posts/YYYY/MM/DD-序号.md`，文件按文章小节命名，看完文章就能按图索骥。

## 目录

| 示例 | 内容 | 对应文章 | 运行 |
|---|---|---|---|
| `2026/09/07-1` | MPSC 环形队列：per-slot 状态机、恰好一次投递校验、与 channel 对比压测 | [07-1](../../blog/posts/2026/09/07-1.md) | `go run ./2026/09/07-1/per-slot` |
| `2026/09/08-7` | 最小 Actor 模型：Spawn 返回 mailbox，多生产者无锁累加，永远输出 1000 | [08-7](../../blog/posts/2026/09/08-7.md) | `go run ./2026/09/08-7` |
| `2026/09/08-8` | LMAX / Disruptor 全套 14 文件：缓存行填充、单/多生产者、依赖图、journal / 快照 / 恢复、输出 Topic | [08-8](../../blog/posts/2026/09/08-8.md) | `go run ./2026/09/08-8` |
| `2026/09/09-1` | 三种等待策略成本实测：stall / ladder 两套实验，多核对照 | [09-1](../../blog/posts/2026/09/09-1.md) | `go run ./2026/09/09-1/park-or-spin -exp=stall -procs=1` |
| `2026/09/15` | Go 网络九连练：TCP / RPC / gRPC / gin / gorm / 存储 / 滑动窗口缓存 / 微服务（9 个模块） | [15](../../blog/posts/2026/09/15.md) | 各模块 README，部分自带 `go.mod` |

运行命令都在 `example/` 根目录下执行（`cd example && go run ./…`）。

## 结构约定

- **同构映射**：示例路径 = 文章路径，新文章配新示例目录，不看文档也知道代码属于哪篇
- **文件即小节**：`sequence.go`、`single.go`、`blp.go`……文件名对应文章代码块的 `title` 标注
- **一个根模块**：除 15 号的 6 个多模块示例（grpc-basics、grpc-middleware、user-service 等自带 `go.mod`）外，全部挂在根 `example/go.mod` 下，`go build ./...` 一把过
- **质量线**：收录的示例全部过 `go build` / `go vet`，能跑的过实际运行验证（08-8 的三个 demo 输出数字与文章逐一对账）

## 独立工具

不挂文章的小工具收在 `example/script/`——目前一件：[setup-zsh.sh](../tools/zsh-setup/)，一行 curl 装好 zsh 环境，文档见[工具页](../tools/zsh-setup/)。

---

*最后更新：2026-09-17*
