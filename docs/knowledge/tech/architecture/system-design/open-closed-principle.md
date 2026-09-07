---
title: 开闭原则（OCP）：状态上移、原子下沉
created: 2026-09-06
updated: 2026-09-06
domain: tech
tags:
  - Go
  - 设计原则
  - 开闭原则
  - 代码重构
  - 原子性
difficulty: intermediate
summary: "把「决定做什么」上移、把「如何安全做」下沉，用包装与协调代替内嵌修改，实现对扩展开放、对修改关闭。"
source:
  - type: blog
    url: /blog/posts/2026/09/06/
    title: "状态上移，原子下沉：包装优于修改的设计取舍"
---

# 开闭原则（OCP）：状态上移、原子下沉

> 把「决定做什么」上移、把「如何安全做」下沉，用包装与协调代替给稳定代码加状态。

## 背景

改动一段已经在跑的代码是有成本的，成本来自三处：

- **回归风险**：已有调用方依赖当前行为。比如磁盘上的文件名格式，改掉后旧数据就读不回来。
- **状态耦合**：往稳定模块塞新状态（如「哪个 gen 是活跃代」），会把简单的并发模型搞复杂。
- **测试失效**：已有测试覆盖的是旧契约，行为一变，测试就从守护者变成摆设。

所以更稳的路径是：老代码保持原子职责不动，新需求放进新单元承接——包装一层、组合一下、在上层协调。这就是开闭原则（Open-Closed Principle，OCP）：对扩展开放，对修改关闭。

## 核心内容

### 执行与决策分离

关键是把「决定做什么」和「如何安全地做」拆成两层。以下是一个快照文件命名的纯函数：

```go title="birdsnest/snapshot.go"
// filterFileName 计算指定槽位与代的文件名。
// gen 仅允许取 0 或 1：gen=0 → bird_filter_NNNNN；gen=1 → bird_filter_NNNNN.1。
func filterFileName(path string, index uint16, gen int) string {
	name := snapFilePrefix + fmt.Sprintf("%05d", index)
	if gen == 1 {
		name += replicaSuffix
	}
	return filepath.Join(path, name)
}
```

职责切分：

| 层 | 职责 | 不负责 |
|---|---|---|
| 上层（分片协调层） | 持有 index 位图，**决策** 哪一代活跃、何时切换 | 不碰文件 I/O 细节 |
| `filterSnapshot` / `filterFileName` | 按给定的 `(槽位, 代)` **原子执行** 读写（tmp + Sync + rename） | 不感知「哪一代活跃」 |

`filterFileName` 是纯函数：无状态、无副作用，输入决定输出。每个函数的契约小到可以单独证明正确。

### 为什么「状态上移、原子下沉」扩展友好

**扩展点在上层**：从双代变三代、加校验和、加 WAL 回滚，改的都是协调层策略，底层 `filterFileName` / `writeFile` 一行不用动——「按参数写一个文件」这个能力是策略无关的。

**崩溃一致性边界清晰**：底层只保证单次写的原子性（rename 前先 fsync，崩溃后只会留下旧文件或完整新文件）；跨文件的一致性（双代切换时新旧文件 + 位图一致）由上层协议保证。两层故障模型互不污染。

**可测试性**：纯函数和无状态结构测试时无需模拟协调逻辑；协调层测试又可以把底层当确定性黑盒。若把位图塞进 `filterSnapshot`，两层逻辑纠缠，两类测试都难写。

## 示例

一次实际应用：把 `WalSnapshot`（让单个 wal 文件既管存储、又隐含「只保留最后一条」的策略）重构为 `filterSnapshot` + 上层位图，正是这条原则的落地。

## 注意事项

- ⚠️ 不是所有「稳定代码」都不能动：契约本身需要演进时该改就改，这条原则针对的是「为承接新需求而污染旧职责」。
- ⚠️ 过度包装会变成抽象地狱：每加一层协调者都有成本，只有当新需求确实改变策略时才值得包装。
- ✅ 判断标准：新增的能力是否策略无关？是，则下沉为原子操作；否，则上移到协调层。

## 延伸阅读

- [状态上移，原子下沉：包装优于修改的设计取舍（博客原文）](../../../../blog/posts/2026/09/06.md)
- [Open–closed principle（Wikipedia）](https://en.wikipedia.org/wiki/Open%E2%80%93closed_principle)

---

*维护人：yiiewang · 最后更新：2026-09-06*
