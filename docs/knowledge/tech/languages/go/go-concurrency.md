---
title: "Go 并发编程：锁与同步全景"
created: 2026-07-01
updated: 2026-07-01
domain: tech
tags:
  - Go
  - 并发编程
  - 锁
  - Channel
  - goroutine
  - 反模式
difficulty: advanced
summary: "系统梳理 Go 并发编程中锁与同步的核心准则、15 个常见反模式及正确做法、死锁公式和自检清单。结合 ChainMaker 项目真实案例。"
source:
  - type: blog
    url: /blog/posts/2026/06/04/
---

# Go 并发编程：锁与同步全景

> 锁的核心准则：**临界区越短越好，里面只碰内存、不碰世界**。Channel 侧：**退出用 close 广播、关闭要幂等、谁生产谁关闭**。Goroutine 侧：**每个都要有退出信号，ctx 要级联**。

## 背景

并发 bug 是所有 bug 中最难复现、最难定位的一类。它们往往在 code review 时看不出来、单测跑不出来，直到生产环境高并发场景才爆发。

Go 虽然以「并发简单」著称，但 goroutine、channel、sync 包的正确组合仍然需要清晰的规则。本文从 ChainMaker 项目的真实并发问题中提炼出普适性的准则和反模式。

## 核心内容

### 一、临界区准则

| 反模式 | 正确做法 |
|--------|---------|
| 锁内做 I/O、网络调用、`time.Sleep`、channel 收发 | 锁只保护**共享内存的读写**，阻塞操作移到锁外 |
| 持锁时调用外部回调/接口方法 | 复制数据，**释放锁后再回调** |
| 一把大锁保护所有字段 | 拆分锁（按字段分组）、读写锁 `sync.RWMutex`、`sync/atomic` |

### 二、加锁顺序

| 反模式 | 正确做法 |
|--------|---------|
| AB-BA 死锁：两个 goroutine 以不同顺序加多把锁 | **全局统一加锁顺序** |
| 嵌套加锁：持锁方法 A 调同锁方法 B | 拆出「不加锁的内部版」`xxxLocked()` |

### 三、Channel 操作

| 反模式 | 正确做法 |
|--------|---------|
| `stopC <- struct{}{}` 通知多个消费者（只送达 1 个） | `close(stopC)` — 广播语义 |
| 重复 `close(ch)` → panic | `sync.Once` 或 `mu + 标志位` 保证幂等 |
| 向已关闭 channel 发送 → panic | 关闭权归生产者 |
| 消费者 close 生产者的 channel | **channel 由发送方关闭** |

### 四、Goroutine 生命周期

| 反模式 | 正确做法 |
|--------|---------|
| `go func(){ for { <-ch } }()` 无退出路径 → 泄漏 | 每个后台 goroutine 必须有 `ctx.Done()` / `close(stopC)` |
| 发停止信号就 `return`，不等 goroutine 退出 | `sync.WaitGroup.Wait()` |
| 内部 `context.Background()` 新建 → 取消无法级联 | 从传入 `ctx` 派生 `WithCancel(ctx)` |

### 五、同步原语误用

| 反模式 | 正确做法 |
|--------|---------|
| 双重检查锁定（DCL）不用 atomic → 数据竞争 | 用 `sync.Once`（Go 单例标准答案） |
| `WaitGroup.Add` 放进 goroutine 内部 → 竞争 | `Add` 在**启动 goroutine 之前** |
| 值传递含 `sync.Mutex` 的 struct → 锁被复制 | 用指针接收者/指针传递 |

## 死锁公式

> **持锁 L → 做阻塞操作 B → B 依赖外部资源 → 外部消失 → B 永不返回 → L 永不释放 → 全员卡死**

项目中的两步改造拆了这条链的两环：

1. `close(stopC)` 广播替代发送 → 阻塞操作不再依赖消费者存活
2. 清理移出锁外 → 持锁期间不再有阻塞操作

```
改造前：Stop() { mu.Lock(); pool.Close(); close(stopC); mu.Unlock() }
                       ↑ 阻塞操作在锁内，可能永不返回

改造后：Stop() { mu.Lock(); close(stopC); mu.Unlock(); pool.Close() }
                                                  ↑ 阻塞操作在锁外
```

## 速查表

| 场景 | 准则 |
|------|------|
| 锁内能做什么 | 只碰内存，不碰世界 |
| 多把锁怎么加 | 全局统一顺序 |
| 退出信号怎么发 | `close(ch)` 广播，不发送 |
| 关闭操作怎么保护 | `sync.Once` 或 `mu + 标志位` |
| channel 谁关 | 发送方关 |
| goroutine 怎么停 | `ctx.Done()` 或 `stopC` 退出分支 |
| goroutine 怎么等 | `sync.WaitGroup`，Add 在 go 之前 |
| context 怎么传 | 从外部派生，不用 `Background()` |
| 单例怎么初始化 | `sync.Once` |
| 带锁 struct 怎么传 | 指针 |

## 自检清单

- [ ] `mu.Lock()` 到 `mu.Unlock()` 之间，有没有 I/O / 网络 / channel 操作？
- [ ] 有没有持锁调用外部接口方法？
- [ ] 退出信号是 `close(ch)` 还是 `ch <- signal`？
- [ ] `close(ch)` 有没有幂等保护？
- [ ] 每个 `go func()` 是否都有退出分支？
- [ ] `context` 是从外部派生还是 `Background()`？
- [ ] `WaitGroup.Add()` 是否在 `go` 语句之前？
- [ ] 有没有值传递含 `sync.Mutex` 的结构体？

## 延伸阅读

- [Go 并发编程反模式与最佳实践（博客原文）](/blog/posts/2026/06/04/)
- [Go 并发模式（Go Blog）](https://go.dev/blog/pipelines)
- 《Go 语言高级编程》第 4 章

---

*维护人：yiiewang · 最后更新：2026-07-01*
