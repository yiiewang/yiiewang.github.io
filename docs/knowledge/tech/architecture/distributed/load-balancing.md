---
title: "负载均衡设计策略与实践"
status: effective
created: 2026-07-01
updated: 2026-07-01
domain: tech
tags:
  - Go
  - 负载均衡
  - 分布式系统
  - 架构设计
difficulty: intermediate
summary: "四种常见负载均衡策略的对比与 Go 实现：轮询、随机、权重选择、洗牌算法，以及健康检查与重试机制。"
source:
  - type: blog
    url: /blog/posts/2025/01/19/
---

# 负载均衡设计策略与实践

> 负载均衡将请求均匀分布到多个服务器节点上，提升系统的性能和可靠性。策略选择取决于业务对**均匀性、性能差异容忍度、故障感知**的需求。

## 四种核心策略对比

| 策略 | 原理 | 优点 | 缺点 | 适用场景 |
|------|------|------|------|---------|
| **轮询（Round Robin）** | 按固定顺序依次选择节点 | 简单均匀 | 无法感知节点性能差异 | 同构节点、简单场景 |
| **随机（Random）** | 随机挑选节点 | 无状态维护 | 分布可能不均 | 快速原型 |
| **权重选择（Weighted）** | 按权重比例分配流量 | 适配异构性能 | 权重需人工维护 | 异构节点、灰度发布 |
| **洗牌（Shuffle）** | 打乱列表后按序选择 | 均匀 + 可复用结果 | 节点多时性能受影响 | 需要均匀又不想随机 |

## Go 实现示例

### 轮询

```go
var nodes = []string{"node1", "node2", "node3"}
var index = 0

func nextNode() string {
    node := nodes[index]
    index = (index + 1) % len(nodes)
    return node
}
```

### 权重选择

```go
type Node struct {
    Address string
    Weight  int
}

func weightedRandomNode(nodes []Node) string {
    totalWeight := 0
    for _, n := range nodes {
        totalWeight += n.Weight
    }
    randValue := rand.Intn(totalWeight)
    for _, n := range nodes {
        if randValue < n.Weight {
            return n.Address
        }
        randValue -= n.Weight
    }
    return ""
}
```

### 洗牌算法

```go
func shuffleNodes(nodes []string) []string {
    shuffled := append([]string(nil), nodes...)
    rand.Shuffle(len(shuffled), func(i, j int) {
        shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
    })
    return shuffled
}
```

## 健康检查与重试

生产环境不可或缺的两个机制：

- **健康检查**：定期检测节点可用性，不可用节点从候选列表移除
- **重试机制**：当前节点失败后选择其他节点重试

```go
func getHealthyNode(nodes []string) string {
    for _, node := range nodes {
        if isHealthy(node) { return node }
    }
    return "" // 无可用节点
}
```

## 策略选择决策树

```
需要高性能？  ──是──→  权重选择 / 洗牌算法
                     还要故障感知？ ──→ 加上健康检查+重试
                     
简单够用？    ──是──→  轮询 / 随机
```

## 延伸阅读

- [负载均衡设计思路与实践指南（博客原文）](/blog/posts/2025/01/19/)
- [Google SRE: Load Balancing](https://sre.google/sre-book/load-balancing-datacenter/)

---

*维护人：yiiewang · 最后更新：2026-07-01*
