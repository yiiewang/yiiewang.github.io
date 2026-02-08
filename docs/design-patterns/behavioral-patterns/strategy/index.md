---
title: 策略模式
date: 2023-02-25 17:18:47
category:
  - 设计模式
tag:
  - 行为模式
---

# 策略模式：换个"大脑"就换种玩法

导航软件问你：走高速还是走国道？打车还是坐地铁？同样是从 A 到 B，不同的"策略"给出完全不同的路线。策略模式就是这个意思—— **定义一系列算法，让它们可以互相替换** 。

**策略模式** 将算法封装成独立的类，使算法可以独立于使用它的客户端变化。

## 为什么需要策略模式？

假设你在开发一个缓存系统，需要支持多种淘汰策略：

```go
// ❌ 糟糕的写法：用 switch 处理所有策略
func (c *Cache) Evict() {
    switch c.strategy {
    case "LRU":
        // 最近最少使用
        c.evictLRU()
    case "LFU":
        // 最不常用
        c.evictLFU()
    case "FIFO":
        // 先进先出
        c.evictFIFO()
    // 每加一种策略就要改这里...
    }
}
```

这样写的问题：

- **违反开闭原则**：每新增策略都要修改核心代码
- **代码膨胀**：所有算法堆在一起，越来越臃肿
- **难以测试**：无法单独测试某个算法

策略模式的解法： **把每种算法封装成独立的类，用接口统一调用** 。

## 模式结构

![策略模式结构](./structure.png)

| 角色 | 职责 | 类比 |
|------|------|------|
| **Context（上下文）** | 持有策略对象，把工作委托给策略 | 导航 App |
| **Strategy（策略接口）** | 定义算法的统一接口 | 路线规划能力 |
| **ConcreteStrategy（具体策略）** | 实现具体的算法 | 高速优先/省钱优先/步行... |

## 动手实现：缓存淘汰策略

用缓存系统的淘汰策略来演示策略模式。

### 第一步：定义策略接口

=== "📄 eviction_strategy.go: 策略接口"

    ```go
    package main

    // EvictionStrategy 缓存淘汰策略接口
    type EvictionStrategy interface {
        Evict(cache *Cache)
    }
    ```

### 第二步：实现具体策略

=== "📄 lru_strategy.go: LRU 策略"

    ```go
    package main

    import "fmt"

    // LRU 最近最少使用策略
    type LRU struct{}

    func (l *LRU) Evict(cache *Cache) {
        fmt.Println("🔄 执行 LRU 策略：淘汰最久未访问的数据")
        // 实际实现中会根据访问时间淘汰
    }
    ```

=== "📄 lfu_strategy.go: LFU 策略"

    ```go
    package main

    import "fmt"

    // LFU 最不常用策略
    type LFU struct{}

    func (l *LFU) Evict(cache *Cache) {
        fmt.Println("📊 执行 LFU 策略：淘汰访问次数最少的数据")
        // 实际实现中会根据访问频率淘汰
    }
    ```

=== "📄 fifo_strategy.go: FIFO 策略"

    ```go
    package main

    import "fmt"

    // FIFO 先进先出策略
    type FIFO struct{}

    func (f *FIFO) Evict(cache *Cache) {
        fmt.Println("➡️ 执行 FIFO 策略：淘汰最早进入的数据")
        // 实际实现中会按照插入顺序淘汰
    }
    ```

### 第三步：实现上下文（缓存）

=== "📄 cache.go: 上下文"

    ```go
    package main

    import "fmt"

    // Cache 缓存，使用可替换的淘汰策略
    type Cache struct {
        storage   map[string]string
        strategy  EvictionStrategy
        capacity  int
        maxSize   int
    }

    // NewCache 创建缓存
    func NewCache(maxSize int, strategy EvictionStrategy) *Cache {
        return &Cache{
            storage:  make(map[string]string),
            strategy: strategy,
            maxSize:  maxSize,
        }
    }

    // SetStrategy 运行时切换策略
    func (c *Cache) SetStrategy(strategy EvictionStrategy) {
        c.strategy = strategy
        fmt.Println("🔧 策略已切换")
    }

    // Add 添加数据
    func (c *Cache) Add(key, value string) {
        if c.capacity >= c.maxSize {
            fmt.Printf("⚠️ 缓存已满（%d/%d），触发淘汰...\n", c.capacity, c.maxSize)
            c.strategy.Evict(c)
        }
        c.storage[key] = value
        c.capacity++
        fmt.Printf("✅ 已添加：%s = %s\n", key, value)
    }

    // Get 获取数据
    func (c *Cache) Get(key string) (string, bool) {
        val, ok := c.storage[key]
        return val, ok
    }
    ```

### 第四步：使用缓存

=== "📄 main.go: 客户端代码"

    ```go
    package main

    import "fmt"

    func main() {
        // 创建缓存，初始使用 LFU 策略
        cache := NewCache(2, &LFU{})

        fmt.Println("=== 使用 LFU 策略 ===")
        cache.Add("a", "1")
        cache.Add("b", "2")
        cache.Add("c", "3")  // 触发淘汰

        fmt.Println("\n=== 切换到 LRU 策略 ===")
        cache.SetStrategy(&LRU{})
        cache.Add("d", "4")  // 使用新策略

        fmt.Println("\n=== 切换到 FIFO 策略 ===")
        cache.SetStrategy(&FIFO{})
        cache.Add("e", "5")
    }
    ```

=== "📄 output.txt: 执行结果"

    ```
    === 使用 LFU 策略 ===
    ✅ 已添加：a = 1
    ✅ 已添加：b = 2
    ⚠️ 缓存已满（2/2），触发淘汰...
    📊 执行 LFU 策略：淘汰访问次数最少的数据
    ✅ 已添加：c = 3

    === 切换到 LRU 策略 ===
    🔧 策略已切换
    ⚠️ 缓存已满（3/2），触发淘汰...
    🔄 执行 LRU 策略：淘汰最久未访问的数据
    ✅ 已添加：d = 4

    === 切换到 FIFO 策略 ===
    🔧 策略已切换
    ⚠️ 缓存已满（4/2），触发淘汰...
    ➡️ 执行 FIFO 策略：淘汰最早进入的数据
    ✅ 已添加：e = 5
    ```

## 策略模式的威力：运行时切换

策略模式最大的优势是可以在 **运行时** 切换算法，而不需要修改任何代码：

```go
// 根据用户配置选择策略
func getStrategy(config string) EvictionStrategy {
    switch config {
    case "lru":
        return &LRU{}
    case "lfu":
        return &LFU{}
    default:
        return &FIFO{}
    }
}

// 也可以用工厂模式创建策略
```

## 什么时候该用策略模式？

| 场景 | 说明 |
|------|------|
| **多种算法可选** | 排序、压缩、加密等有多种实现方式 |
| **需要运行时切换** | 根据用户偏好或系统状态选择算法 |
| **消除条件分支** | 避免大量 if-else 或 switch |
| **算法需要独立变化** | 算法的修改不应影响使用它的代码 |

**常见应用** ：

- **支付系统**：微信支付/支付宝/银行卡
- **排序算法**：快排/归并/堆排序
- **压缩工具**：ZIP/RAR/7Z
- **表单验证**：不同字段的验证规则
- **日志系统**：输出到文件/控制台/远程服务器

## 优缺点分析

| ✅ 优点 | ❌ 缺点 |
|---------|---------|
| **运行时切换**：可以动态改变对象行为 | **客户端需了解策略**：必须知道有哪些策略可选 |
| **开闭原则**：新增策略无需修改现有代码 | **策略数量多时**：类的数量会增加 |
| **消除条件分支**：代码更清晰 | **简单场景过度设计**：只有两三种选择时用 if-else 更简单 |
| **算法可复用**：同一策略可用于多个上下文 | |

## 策略模式 vs 其他模式

| 模式对比 | 说明 |
|----------|------|
| **策略 vs 状态** | 策略是外部指定，状态是内部自动切换 |
| **策略 vs 命令** | 策略是"怎么做"，命令是"做什么" |
| **策略 vs 模板方法** | 模板方法用继承，策略用组合 |

## 函数式替代方案

在支持函数式编程的语言中，可以用函数代替策略类：

```go
// 用函数类型代替接口
type EvictFunc func(cache *Cache)

// 定义策略函数
func lruEvict(cache *Cache) {
    fmt.Println("LRU 淘汰")
}

func lfuEvict(cache *Cache) {
    fmt.Println("LFU 淘汰")
}

// 使用
cache.evictFunc = lruEvict
```

> **一句话总结**：策略模式就像手机换 SIM 卡——同一部手机，换张卡就能用不同的号码和套餐，但打电话的方式没变。
