---
title: 装饰模式
date: 2023-02-25 17:18:47
category:
  - 设计模式
tag:
  - 结构型模式
---

# 🎁 装饰模式：给对象"穿衣服"

> 装饰模式是一种结构型设计模式，允许你通过将对象放入包含行为的特殊封装对象中来为原对象绑定新的行为——就像给人穿衣服一样，可以层层叠加。

## 从生活场景说起

想象你去披萨店点餐：

1. **基础披萨**：15 元的蔬菜披萨
2. **加芝士**：+10 元
3. **加番茄**：+7 元
4. **加培根**：+12 元

你的最终订单是：**蔬菜披萨 + 芝士 + 番茄 = 32 元**

关键问题：如果用继承来实现所有组合，需要多少个类？

- 蔬菜披萨
- 蔬菜披萨+芝士
- 蔬菜披萨+番茄
- 蔬菜披萨+芝士+番茄
- 蔬菜披萨+芝士+培根
- ... **组合爆炸！** 😱

装饰模式的解法： **把配料做成可以层层包装的"装饰器"** 。

```
┌──────────────────────────────────┐
│        番茄装饰器 (+7元)          │
│  ┌──────────────────────────┐   │
│  │    芝士装饰器 (+10元)      │   │
│  │  ┌────────────────────┐  │   │
│  │  │   蔬菜披萨 (15元)   │  │   │
│  │  └────────────────────┘  │   │
│  └──────────────────────────┘   │
└──────────────────────────────────┘

总价 = 15 + 10 + 7 = 32 元
```

## 为什么需要装饰模式？

### 😩 用继承的问题

```go
// 继承方式：类爆炸
type VeggiePizza struct{}
type VeggiePizzaWithCheese struct{}
type VeggiePizzaWithTomato struct{}
type VeggiePizzaWithCheeseAndTomato struct{}
type VeggiePizzaWithCheeseAndBacon struct{}
// ... 每种组合都要一个类！

// 新增一种配料？所有组合类都要翻倍！
```

### 😊 装饰模式的优雅解法

```go
// 基础披萨
pizza := &VeggieMania{}                    // 15 元

// 动态添加配料（装饰）
pizzaWithCheese := &CheeseTopping{pizza}   // +10 元
pizzaWithAll := &TomatoTopping{pizzaWithCheese}  // +7 元

fmt.Println(pizzaWithAll.getPrice())       // 32 元
```

## 模式结构

```
┌────────────────────────────────────────────────────────────────┐
│                                                                │
│                    ┌──────────────┐                            │
│                    │  Component   │ ◀─── 定义统一接口          │
│                    │  (部件接口)   │                           │
│                    └──────┬───────┘                            │
│                           │                                    │
│              ┌────────────┴────────────┐                       │
│              ▼                         ▼                       │
│    ┌──────────────────┐      ┌──────────────────┐             │
│    │ConcreteComponent │      │    Decorator     │             │
│    │  (具体部件)       │      │   (装饰基类)     │             │
│    │   🍕 披萨        │      │    🧀 配料      │             │
│    └──────────────────┘      └────────┬─────────┘             │
│                                       │                        │
│                              ┌────────┴────────┐              │
│                              ▼                 ▼              │
│                    ┌──────────────┐   ┌──────────────┐        │
│                    │CheeseTopping │   │TomatoTopping │        │
│                    │   芝士装饰    │   │   番茄装饰   │        │
│                    └──────────────┘   └──────────────┘        │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

### 角色说明

| 角色 | 职责 | 示例 |
|------|------|------|
| **部件** (Component) | 定义被装饰对象的接口 | `IPizza` 接口 |
| **具体部件** (Concrete Component) | 被装饰的原始对象 | `VeggieMania` 披萨 |
| **装饰** (Decorator) | 持有部件引用，实现相同接口 | 配料装饰器 |
| **具体装饰** (Concrete Decorator) | 添加具体的装饰行为 | `CheeseTopping`、`TomatoTopping` |

## 完整实现

=== "pizza.go: 部件接口"

    ```go
    package main

    // IPizza 披萨接口
    type IPizza interface {
        getPrice() int
        getDescription() string
    }
    ```

=== "veggieMania.go: 具体部件"

    ```go
    package main

    // VeggieMania 蔬菜披萨（基础产品）
    type VeggieMania struct{}

    func (p *VeggieMania) getPrice() int {
        return 15
    }

    func (p *VeggieMania) getDescription() string {
        return "🍕 蔬菜披萨"
    }
    ```

=== "cheeseTopping.go: 具体装饰"

    ```go
    package main

    // CheeseTopping 芝士配料装饰器
    type CheeseTopping struct {
        pizza IPizza  // 持有被装饰对象的引用
    }

    func (c *CheeseTopping) getPrice() int {
        return c.pizza.getPrice() + 10  // 在原价基础上加价
    }

    func (c *CheeseTopping) getDescription() string {
        return c.pizza.getDescription() + " + 🧀 芝士"
    }
    ```

=== "tomatoTopping.go: 具体装饰"

    ```go
    package main

    // TomatoTopping 番茄配料装饰器
    type TomatoTopping struct {
        pizza IPizza
    }

    func (c *TomatoTopping) getPrice() int {
        return c.pizza.getPrice() + 7
    }

    func (c *TomatoTopping) getDescription() string {
        return c.pizza.getDescription() + " + 🍅 番茄"
    }
    ```

=== "baconTopping.go: 具体装饰"

    ```go
    package main

    // BaconTopping 培根配料装饰器
    type BaconTopping struct {
        pizza IPizza
    }

    func (c *BaconTopping) getPrice() int {
        return c.pizza.getPrice() + 12
    }

    func (c *BaconTopping) getDescription() string {
        return c.pizza.getDescription() + " + 🥓 培根"
    }
    ```

=== "main.go: 客户端代码"

    ```go
    package main

    import "fmt"

    func main() {
        // 基础披萨
        fmt.Println("=== 基础披萨 ===")
        pizza := &VeggieMania{}
        fmt.Printf("%s = %d 元\n", pizza.getDescription(), pizza.getPrice())

        // 加芝士
        fmt.Println("\n=== 加芝士 ===")
        pizzaWithCheese := &CheeseTopping{pizza: pizza}
        fmt.Printf("%s = %d 元\n", 
            pizzaWithCheese.getDescription(), 
            pizzaWithCheese.getPrice())

        // 再加番茄
        fmt.Println("\n=== 再加番茄 ===")
        pizzaWithCheeseAndTomato := &TomatoTopping{pizza: pizzaWithCheese}
        fmt.Printf("%s = %d 元\n", 
            pizzaWithCheeseAndTomato.getDescription(), 
            pizzaWithCheeseAndTomato.getPrice())

        // 豪华套餐：芝士 + 番茄 + 培根
        fmt.Println("\n=== 豪华套餐 ===")
        deluxePizza := &BaconTopping{
            pizza: &TomatoTopping{
                pizza: &CheeseTopping{
                    pizza: &VeggieMania{},
                },
            },
        }
        fmt.Printf("%s = %d 元\n", 
            deluxePizza.getDescription(), 
            deluxePizza.getPrice())
    }
    ```

=== "output.txt: 执行结果"

    ```
    === 基础披萨 ===
    🍕 蔬菜披萨 = 15 元

    === 加芝士 ===
    🍕 蔬菜披萨 + 🧀 芝士 = 25 元

    === 再加番茄 ===
    🍕 蔬菜披萨 + 🧀 芝士 + 🍅 番茄 = 32 元

    === 豪华套餐 ===
    🍕 蔬菜披萨 + 🧀 芝士 + 🍅 番茄 + 🥓 培根 = 44 元
    ```

## 装饰模式 vs 继承

| 特性 | 装饰模式 | 继承 |
|------|----------|------|
| **组合方式** | 运行时动态组合 | 编译时静态确定 |
| **扩展性** | 新增装饰器即可 | 需要修改类层次 |
| **组合数量** | N 个装饰器可任意组合 | 需要 2^N 个子类 |
| **灵活性** | 可以添加/移除装饰 | 继承关系固定 |

## 什么时候使用装饰模式？

| 场景 | 说明 |
|------|------|
| 🎨 **动态扩展** | 需要在运行时为对象添加功能 |
| 🔒 **无法继承** | 类使用了 `final` 关键字，无法被继承 |
| 📦 **功能组合** | 有多种可选功能需要自由组合 |
| 🧱 **单一职责** | 希望将复杂功能拆分成独立的装饰器 |

## 实际应用场景

### 1. I/O 流（Java 经典示例）

```java
// Java 的 I/O 流就是装饰模式的典型应用
InputStream input = new FileInputStream("file.txt");
input = new BufferedInputStream(input);      // 添加缓冲功能
input = new DataInputStream(input);          // 添加数据类型读取功能
```

### 2. HTTP 中间件

```go
// 给 HTTP Handler 添加功能
handler := http.HandlerFunc(myHandler)
handler = loggingMiddleware(handler)    // 日志装饰
handler = authMiddleware(handler)       // 认证装饰
handler = rateLimitMiddleware(handler)  // 限流装饰
```

### 3. 数据加密/压缩

```go
type DataSource interface {
    WriteData(data string)
    ReadData() string
}

type FileDataSource struct{ filename string }
type EncryptionDecorator struct{ source DataSource }  // 加密装饰
type CompressionDecorator struct{ source DataSource } // 压缩装饰
```

## 优缺点分析

| ✅ 优点 | ❌ 缺点 |
|---------|---------|
| **无需继承即可扩展** | **装饰顺序可能影响结果** |
| **运行时添加/移除功能** | **多层装饰难以调试** |
| **单一职责，每个装饰器职责明确** | **初始化代码可能很"丑"** |
| **可以组合多种装饰** | **删除中间某个装饰比较麻烦** |

## 与其他模式的关系

| 模式 | 关系 |
|------|------|
| **适配器** | 适配器改变接口，装饰器保持接口不变 |
| **代理** | 代理控制访问，装饰器增强功能 |
| **组合** | 组合是树形结构，装饰只有单个子组件 |
| **责任链** | 结构相似，但责任链可以中断，装饰器不能中断 |
| **策略** | 策略改变对象内核，装饰器改变对象外壳 |

!!! tip "2009 综合知识 60"
    装饰（Decorator）模式可以在不修改对象外观和功能的情况下添加或删除对象功能。在以下情况中应该使用装饰模式：
    
    * 想要在单个对象中动态并且透明地添加责任
    * 想要在以后可能要修改的对象中添加责任
    * 当无法通过静态子类化实现扩展时

---

!!! tip "一句话总结"
    **装饰模式就是"穿衣服"**：你可以给一个人穿 T 恤、外套、围巾，每件衣服都是一个装饰器，可以随意穿脱，人还是那个人，但功能（保暖、美观）增强了。
