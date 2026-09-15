# 滑动窗口缓存：环形缓冲 + 双索引 + 分片锁

面向“最近 N 个区块”场景的缓存实现：区块本体放环形缓冲区，高度索引与交易索引负责 O(1) 查询，容量满时淘汰最旧区块并同步清理两级索引。

对应博客《[从字节流到持久化：Go 后端一条线上的八件事](../../../docs/blog/posts/2026/09/15.md)》第 8 节「滑动窗口缓存：读路径上的账」。

## 文件结构

```text
rolling-window-cache/
├── cache.go           # 缓存实现：环形缓冲、两级索引、分片锁、监控指标
├── cache_test.go      # 功能测试：淘汰、索引一致性、并发、边界
├── benchmark_test.go  # 基准测试：分片竞争、混合负载、与单锁 map 对比
├── example_test.go    # 可执行示例（go test 直接校验输出）
└── demo/main.go       # 可运行 demo：写入 8 个区块观察淘汰与指标
```

## 快速开始

`go.mod` 在 `example/` 下，命令需从那里发起：

```bash
cd example

go run ./20260915/rolling-window-cache/demo      # 跑 demo

cd 20260915/rolling-window-cache
go test -race ./...                                # 功能测试 + 竞态检测
go test -run Example ./...                         # 只跑示例（带期望输出）
go test -bench='ShardContention|CompareWithMap' -benchmem ./...   # 关键基准
```

## 设计要点

- **环形缓冲**：`ringBuffer` 定长，`head`/`tail` 取模推进，淘汰是 O(1) 的覆盖写，没有内存抖动
- **两级索引**：高度 → 槽位下标（`heightIndex`，全局读锁保护）；交易ID → 高度（分片 `txIndex`，哈希到独立读写锁）
- **分片的意义**：交易ID查询只碰一个分片的锁，不阻塞其他分片；高度索引仍然是全局读锁，但读多写少场景下 `RWMutex` 足够。实测分片 1 → 4 有约 20% 收益，之后趋于平坦
- **淘汰的原子性**：区块数据、高度索引、交易索引必须一起删，否则会留下指向空槽位的脏索引
- **指标**：请求数 / 命中 / 未命中 / 插入 / 淘汰 / 平均耗时；`GetBlockByTxID` 内部委托高度查询时只记一次请求
- **指标的代价**：`updateLookupMetrics` 每次查询拿一次全局锁，实测占读路径成本的大头（444ns → 84ns 的差距），高并发下需要原子化或采样

## 复杂度

| 操作 | 复杂度 | 说明 |
|------|--------|------|
| `AddBlock` | O(T) | T 为区块内交易数（写交易索引） |
| `GetBlockByHeight` | O(1) | 哈希表 + 环形缓冲下标 |
| `GetBlockByTxID` | O(1) | 分片哈希表 + 一次高度查询 |
| `GetRecentBlocks` | O(N) | 从 head 向前取 N 个 |
| `GetBlocksInTimeWindow` | O(M) | M 为当前区块数（无时间索引） |

## 注意

- `windowSize` 只被 `GetBlocksInWindow()` 使用；按任意区间查询请用 `GetBlocksInTimeWindow(start, end)`
- 时间戳是秒级，毫秒精度的窗口查询不适用
- 交易ID哈希分布不均会导致分片热点，真实场景可用前缀打散
