---
title: 多分片求交集：缺失交易查询的三种实现
created: 2026-09-29
updated: 2026-09-29
domain: tech
tags:
  - Go
  - 数据结构
  - 集合运算
  - 分片
difficulty: intermediate
summary: "分片系统查'所有分片都缺失'本质是求交集；删除法、重建法、计数法三版对照，计数法以 count==分片数 一次判定胜出。"
source:
  - type: blog
    url: /blog/posts/2026/09/29/
    title: "一个交集，三种写法：分片缺失交易查询的三版演进"
---

# 多分片求交集：缺失交易查询的三种实现

> 分片系统查"所有分片都缺失的交易"，本质是求交集；三种实现里，计数法以 `count == 分片数` 一次判定胜出。

## 背景

分片存储的常见查询模式：一批交易 ID 广播到每个分片，每个分片只能回答"本分片缺失哪些"；业务需要的是"**所有分片都没有**"的交易——典型于数据拉取、恢复、对账场景。

数学模型一句话：**所有分片缺失表的交集**：

```text
missing = shard1.missing ∩ shard2.missing ∩ ... ∩ shardN.missing
```

前提：每个分片的缺失表必然是入参 `txIds` 的子集（查询的就是这批 ID），因此交集结果无需再与 `txIds` 求交。

## 核心内容

### 标准例子

入参 `txIds = [tx1, tx2, tx3]`，三个分片返回：

| 分片 | 本分片命中 | 本分片缺失 |
| :--- | :--- | :--- |
| shard1 | tx1 | tx2, tx3 |
| shard2 | tx2 | tx1, tx3 |
| shard3 | 无 | tx1, tx2, tx3 |

正确答案：`missing = [tx3]`（只有它一个分片都没命中）。

### 删除法：从候选集里划掉"被找到的"

初始假设全缺失，逐分片遍历，本分片**找到**的（不在缺失表里）从候选集删除。

```go title="v1-delete.go"
missingSet := make(map[string]struct{}, len(txIds))
for _, id := range txIds {
    missingSet[id] = struct{}{}
}

for _, shardMissing := range shardMissings {
    // 索引表：切片判断成员关系是 O(n)，先转 map 换 O(1)
    shardMissingSet := make(map[string]struct{}, len(shardMissing))
    for _, id := range shardMissing {
        shardMissingSet[id] = struct{}{}
    }
    for id := range missingSet {
        if _, ok := shardMissingSet[id]; !ok {
            delete(missingSet, id) // 本分片找到了 → 划掉
        }
    }
}
```

要点：`shardMissingSet` 是**性能配件**，不参与业务语义——它存在的唯一原因是把切片查询压到 O(1)。

### 重建法：每轮保留"两边都缺"的

每轮新建集合，只把 `候选集 ∩ 本分片缺失` 的元素搬入，再整体替换。

```go title="v2-rebuild.go"
missingSet := make(map[string]struct{}, len(txIds))
for _, id := range txIds {
    missingSet[id] = struct{}{}
}

for _, shardMissing := range shardMissings {
    nextMissingSet := make(map[string]struct{})
    for _, id := range shardMissing {
        if _, ok := missingSet[id]; ok { // 之前的分片也没找到它
            nextMissingSet[id] = struct{}{}
        }
    }
    missingSet = nextMissingSet // 用"两边都缺"替换
}
```

要点：实质是标准交集运算，且不需要索引表；但"新建 + 替换"动作（`missingSet = nextMissingSet`）意图不直观，可读性差在循环末尾的这次整体替换。

### 计数法：被所有分片说过缺失才算缺

不做集合运算，只记每个 id 被几个分片"说过缺失"，`count == shardCnt` 一次判定。

```go title="v3-count.go"
shardCnt := len(shardMissings)

missingCount := make(map[string]int)
for _, shardMissing := range shardMissings {
    for _, id := range shardMissing {
        missingCount[id]++
    }
}

missing := make([]string, 0, len(txIds))
seen := make(map[string]struct{}, len(txIds))
for _, id := range txIds {
    if _, dup := seen[id]; dup {
        continue
    }
    seen[id] = struct{}{}
    if missingCount[id] == shardCnt {
        missing = append(missing, id)
    }
}
```

### 三版对照

| 方案 | 每轮做什么 | 想表达的意思 | 额外代价 |
| :--- | :--- | :--- | :--- |
| 删除法 | 删掉"本分片找到的" | 找到了就划掉 | 一张索引表 |
| 重建法 | 交集重建替换 | 两边都没找到才留下 | 每轮建新集合 + 替换 |
| 计数法 | 只记缺失次数 | 全员说缺才算缺 | 无集合运算 |

计数法胜出的理由：

1. 不维护集合状态、没有中途替换，循环内只有计数一个动作
2. `count == shardCnt` 自带语义："所有分片都没有它"
3. 最终过滤遍历入参，输出顺序即入参顺序，去重顺手解决
4. map 只当 counter，职责单一
5. 效率最优：全程 1 个 map，零 per-shard 分配，GC 压力最小

### 效率对比与真实瓶颈

设每批查询 n 个 txIds（去重后 m 个）、分片数 N、第 i 个分片返回缺失数为 K_i：

| 方案 | 每轮时间 | 每轮分配 | 全过程分配 |
| :--- | :--- | :--- | :--- |
| 删除法 | O(K_i + \|候选集\|) | 1 张索引表 | 1 + N 个 map |
| 重建法 | O(K_i) | 1 个新 map，旧的变垃圾 | 1 + N 个 map |
| 计数法 | O(K_i) | **0** | **1 个 map** |

三版时间同为 O(ΣK_i)，但计数法零 per-shard 分配，GC 压力最小；代价仅 map 的 value 从 `struct{}` 变 `int`（每 entry 多 8 字节）。

集合运算的耗时是微秒级，真正决定函数耗时的是分片 RPC（毫秒级）：

1. 分片数 N：客户端串行遍历时 N 次 RPC 串行叠加，N 大时并发化（`sync.WaitGroup` + 结果聚合）才是数量级优化，需保持"任一分片失败即整体报错"的语义
2. txIds 规模 m：一次 RPC 传全部 txIds 受 gRPC 消息大小限制，m 大需分批

常规场景（单次查询、m 不大）计数法即最优解；N 或 m 明显变大才需要结构性优化。

## 示例

计数法执行过程（`shardCnt = 3`）：

| 轮次 | missingCount |
| :--- | :--- |
| shard1 | tx2:1, tx3:1 |
| shard2 | tx1:1, tx2:1, tx3:2 |
| shard3 | tx1:2, tx2:2, tx3:3 |
| 判定 | `count == 3` 的只有 tx3 |

三版跑同一例子，结果一致，均为 `[tx3]`。

## 注意事项

- ⚠️ 计数法前提：每个分片缺失表内部无重复 id，否则 count 虚增导致漏判（分片侧保证去重，或查询前先去重）
- ⚠️ 删除法 / 重建法的结果从 map 收集，输出乱序，需另行排序
- ✅ "N 个条件全满足"型判定优先考虑计数：每个条件满足一次 `+1`，`count == N` 一次收口
- ✅ 遍历入参生成结果，顺序即入参顺序，顺带去重
- ✅ 时间同为 O(ΣK_i)，计数法零 per-shard 分配，GC 压力最小
- ⚠️ 真正耗时在分片 RPC 串行往返（毫秒级）：N 大考虑并发化（保持失败语义），m 大考虑分批

## 延伸阅读

- 博客：[一个交集，三种写法：分片缺失交易查询的三版演进](/blog/posts/2026/09/29/)
- 相关条目：[Go 并发编程：锁与同步全景](go-concurrency.md)

---

*维护人：yiiewang · 最后更新：2026-09-29*
