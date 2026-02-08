---
title: 观察者模式
date: 2023-02-25 17:18:47
category:
  - 设计模式
tag:
  - 行为模式
---

# 观察者模式：订阅你关心的一切

你有没有订阅过某个 UP 主的更新？一旦他发布新视频，你就会收到通知。你不需要每分钟刷新一次页面，系统会"推送"给你。这就是观察者模式的精髓——**当感兴趣的事情发生时，主动通知你**。

**观察者模式** 定义了一种一对多的依赖关系：当一个对象状态改变时，所有依赖它的对象都会自动收到通知并更新。

## 为什么需要观察者模式？

假设你在开发一个电商网站，用户可以订阅商品到货通知：

```go
// ❌ 糟糕的写法：轮询检查
func checkStock() {
    for {
        if product.InStock {
            notifyUser1()
            notifyUser2()
            notifyUser3()
            // ... 每新增一个用户就要改代码
        }
        time.Sleep(time.Minute)
    }
}
```

这样写的问题：

- **浪费资源** ：大部分时间商品都没到货，但程序一直在检查
- **紧耦合** ：每新增一个订阅者，都要修改核心代码
- **不灵活** ：无法动态添加或移除订阅者

观察者模式的解法： **让商品（发布者）维护一个订阅者列表，到货时主动通知所有人**。

## 模式结构

![观察者模式结构](./structure.png)

| 角色 | 职责 | 类比 |
|------|------|------|
| **Subject（发布者）** | 维护订阅者列表，发送通知 | B站 UP 主 |
| **Observer（观察者接口）** | 定义接收通知的方法 | 观众的"接收推送"能力 |
| **ConcreteObserver（具体观察者）** | 接收通知并做出响应 | 具体的粉丝 |
| **Client（客户端）** | 创建发布者和观察者，建立订阅关系 | 用户点击"关注"按钮 |

## 动手实现：商品到货通知系统

用电商商品到货通知来演示观察者模式。

### 第一步：定义观察者接口

=== "📄 observer.go: 观察者接口"

    ```go
    package main

    // Observer 观察者接口
    type Observer interface {
        Update(itemName string)   // 接收通知
        GetID() string            // 获取观察者标识
    }
    ```

### 第二步：定义发布者接口

=== "📄 subject.go: 发布者接口"

    ```go
    package main

    // Subject 发布者接口
    type Subject interface {
        Register(observer Observer)    // 添加订阅者
        Deregister(observer Observer)  // 移除订阅者
        NotifyAll()                    // 通知所有订阅者
    }
    ```

### 第三步：实现具体观察者（顾客）

=== "📄 customer.go: 具体观察者"

    ```go
    package main

    import "fmt"

    // Customer 顾客，订阅商品到货通知
    type Customer struct {
        email string
    }

    // Update 收到通知时的处理
    func (c *Customer) Update(itemName string) {
        fmt.Printf("📧 发送邮件给 %s：您关注的「%s」已到货！\n", c.email, itemName)
    }

    // GetID 返回顾客标识
    func (c *Customer) GetID() string {
        return c.email
    }
    ```

### 第四步：实现具体发布者（商品）

=== "📄 item.go: 具体发布者"

    ```go
    package main

    import "fmt"

    // Item 商品，作为发布者
    type Item struct {
        observerList []Observer
        name         string
        inStock      bool
    }

    // NewItem 创建商品
    func NewItem(name string) *Item {
        return &Item{name: name}
    }

    // Register 添加订阅者
    func (i *Item) Register(o Observer) {
        i.observerList = append(i.observerList, o)
        fmt.Printf("✅ %s 已订阅「%s」的到货通知\n", o.GetID(), i.name)
    }

    // Deregister 移除订阅者
    func (i *Item) Deregister(o Observer) {
        for idx, observer := range i.observerList {
            if observer.GetID() == o.GetID() {
                i.observerList = append(i.observerList[:idx], i.observerList[idx+1:]...)
                fmt.Printf("❌ %s 已取消订阅「%s」\n", o.GetID(), i.name)
                return
            }
        }
    }

    // NotifyAll 通知所有订阅者
    func (i *Item) NotifyAll() {
        for _, observer := range i.observerList {
            observer.Update(i.name)
        }
    }

    // Restock 商品到货
    func (i *Item) Restock() {
        fmt.Printf("\n🎉 「%s」已到货！正在通知所有订阅者...\n", i.name)
        i.inStock = true
        i.NotifyAll()
    }
    ```

### 第五步：组装并使用

=== "📄 main.go: 客户端代码"

    ```go
    package main

    func main() {
        // 1. 创建商品（发布者）
        shirt := NewItem("Nike 限量款 T恤")

        // 2. 创建顾客（观察者）
        customer1 := &Customer{email: "zhangsan@example.com"}
        customer2 := &Customer{email: "lisi@example.com"}
        customer3 := &Customer{email: "wangwu@example.com"}

        // 3. 顾客订阅商品到货通知
        shirt.Register(customer1)
        shirt.Register(customer2)
        shirt.Register(customer3)

        // 4. 商品到货，自动通知所有订阅者
        shirt.Restock()
    }
    ```

=== "📄 output.txt: 执行结果"

    ```
    ✅ zhangsan@example.com 已订阅「Nike 限量款 T恤」的到货通知
    ✅ lisi@example.com 已订阅「Nike 限量款 T恤」的到货通知
    ✅ wangwu@example.com 已订阅「Nike 限量款 T恤」的到货通知

    🎉 「Nike 限量款 T恤」已到货！正在通知所有订阅者...
    📧 发送邮件给 zhangsan@example.com：您关注的「Nike 限量款 T恤」已到货！
    📧 发送邮件给 lisi@example.com：您关注的「Nike 限量款 T恤」已到货！
    📧 发送邮件给 wangwu@example.com：您关注的「Nike 限量款 T恤」已到货！
    ```

## 观察者模式的变体

### 推模型 vs 拉模型

- **推模型**：发布者主动把数据推给观察者（上面的例子）
- **拉模型**：发布者只通知"有变化"，观察者自己去获取数据

```go
// 拉模型：观察者自己获取数据
func (c *Customer) Update(subject Subject) {
    item := subject.(*Item)
    if item.inStock {
        fmt.Printf("我来看看 %s 是什么...\n", item.name)
    }
}
```

### 事件驱动架构

现代框架通常用"事件"来实现观察者模式：

```go
// 事件总线
eventBus.Subscribe("item.restocked", func(event Event) {
    fmt.Println("收到事件:", event.Data)
})

eventBus.Publish("item.restocked", itemData)
```

## 什么时候该用观察者模式？

| 场景 | 说明 |
|------|------|
| **一对多依赖关系** | 一个对象变化需要通知多个对象 |
| **松耦合通信** | 发布者和订阅者互不知道对方的具体类型 |
| **动态订阅** | 订阅关系可以在运行时添加或移除 |
| **事件驱动系统** | GUI 事件、消息队列、WebSocket 推送 |

**常见应用** ：

- **GUI 框架**：按钮点击事件、输入框变化
- **消息队列**：RabbitMQ、Kafka
- **前端框架**：Vue 的响应式系统、Redux 的状态订阅
- **社交平台**：关注/粉丝系统

## 优缺点分析

| ✅ 优点 | ❌ 缺点 |
|---------|---------|
| **开闭原则**：新增订阅者无需修改发布者 | **通知顺序不确定**：无法控制谁先收到 |
| **运行时建立关系**：动态添加/移除订阅 | **可能导致循环依赖**：A 通知 B，B 又通知 A |
| **松耦合**：发布者和订阅者互不依赖 | **内存泄漏风险**：忘记取消订阅会导致对象无法回收 |

## 与其他模式的关系

| 模式对比 | 说明 |
|----------|------|
| **观察者 vs 中介者** | 观察者是分布式通信，中介者是集中式通信 |
| **观察者 vs 发布-订阅** | 发布-订阅有消息代理，更解耦 |
| **观察者 + 责任链** | 事件先经过责任链过滤，再通知观察者 |

> **一句话总结**：观察者模式就像微信公众号——你关注了，就能收到推送；取消关注，就不再收到。公众号不需要知道你是谁，只管发文章。
