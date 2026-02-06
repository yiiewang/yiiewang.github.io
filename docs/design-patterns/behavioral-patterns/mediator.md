---
title: 中介者模式
date: 2023-02-25 17:18:47
category:
  - 设计模式
tag:
  - 行为模式
---

# 中介者模式：让对象不再"直接对话"

想象一个机场的塔台：飞机之间不会直接通信说"我要降落了，你让一下"，而是都和塔台沟通，由塔台来协调。塔台就是一个"中介者"，它知道所有飞机的状态，负责避免冲突。

**中介者模式** 用一个中介对象来封装一系列对象的交互。对象之间不再直接通信，而是通过中介者进行协调，从而降低耦合度。

## 为什么需要中介者？

假设你在开发一个聊天室系统，用户之间可以互相发消息：

```go
// ❌ 糟糕的写法：用户之间直接通信
type User struct {
    friends []*User
}

func (u *User) SendMessage(msg string) {
    for _, friend := range u.friends {
        friend.Receive(msg)
    }
}
```

这样写的问题：

- **紧耦合**：每个用户都持有其他用户的引用
- **关系复杂**：N 个用户就有 N×(N-1) 条关系线
- **难以扩展**：想加个"禁言"功能？每个用户都要改

中介者模式的解法： **引入一个聊天室（中介者），所有消息都发给聊天室，由它负责转发** 。

## 模式结构

![中介者模式结构](https://refactoringguru.cn/images/patterns/diagrams/mediator/structure.png)

| 角色 | 职责 | 类比 |
|------|------|------|
| **Mediator（中介者接口）** | 定义组件间通信的接口 | 机场塔台的通信协议 |
| **ConcreteMediator（具体中介者）** | 实现协调逻辑，管理所有组件 | 具体的塔台控制系统 |
| **Component（组件）** | 持有中介者引用，通过中介者与其他组件通信 | 飞机 |

## 动手实现：火车站调度系统

用火车站调度系统来演示中介者模式。多列火车共用一个站台，需要协调进站顺序。

### 第一步：定义中介者接口

=== "📄 mediator.go: 中介者接口"

    ```go
    package main

    // Mediator 中介者接口
    type Mediator interface {
        CanArrive(train Train) bool   // 询问是否可以进站
        NotifyDeparture()              // 通知已离站
    }
    ```

### 第二步：实现具体中介者（车站调度员）

=== "📄 station_manager.go: 具体中介者"

    ```go
    package main

    import "fmt"

    // StationManager 车站调度员，协调所有火车
    type StationManager struct {
        isPlatformFree bool
        waitingQueue   []Train
    }

    // NewStationManager 创建调度员
    func NewStationManager() *StationManager {
        return &StationManager{isPlatformFree: true}
    }

    // CanArrive 判断火车是否可以进站
    func (s *StationManager) CanArrive(t Train) bool {
        if s.isPlatformFree {
            s.isPlatformFree = false
            fmt.Println("🚉 站台空闲，允许进站")
            return true
        }
        // 站台占用，加入等待队列
        s.waitingQueue = append(s.waitingQueue, t)
        fmt.Println("⏳ 站台占用，请等待...")
        return false
    }

    // NotifyDeparture 火车离站后通知调度员
    func (s *StationManager) NotifyDeparture() {
        if !s.isPlatformFree {
            s.isPlatformFree = true
        }
        // 通知等待队列中的第一列火车
        if len(s.waitingQueue) > 0 {
            firstTrain := s.waitingQueue[0]
            s.waitingQueue = s.waitingQueue[1:]
            firstTrain.PermitArrival()
        }
    }
    ```

### 第三步：定义组件接口和具体组件（火车）

=== "📄 train.go: 火车接口"

    ```go
    package main

    // Train 火车接口
    type Train interface {
        Arrive()
        Depart()
        PermitArrival()
    }
    ```

=== "📄 passenger_train.go: 客运列车"

    ```go
    package main

    import "fmt"

    // PassengerTrain 客运列车
    type PassengerTrain struct {
        mediator Mediator
        name     string
    }

    func (t *PassengerTrain) Arrive() {
        if !t.mediator.CanArrive(t) {
            fmt.Printf("🚃 %s：收到，等待进站\n", t.name)
            return
        }
        fmt.Printf("🚃 %s：已进站\n", t.name)
    }

    func (t *PassengerTrain) Depart() {
        fmt.Printf("🚃 %s：正在离站...\n", t.name)
        t.mediator.NotifyDeparture()
    }

    func (t *PassengerTrain) PermitArrival() {
        fmt.Printf("🚃 %s：收到进站许可，正在进站...\n", t.name)
        fmt.Printf("🚃 %s：已进站\n", t.name)
    }
    ```

=== "📄 freight_train.go: 货运列车"

    ```go
    package main

    import "fmt"

    // FreightTrain 货运列车
    type FreightTrain struct {
        mediator Mediator
        name     string
    }

    func (t *FreightTrain) Arrive() {
        if !t.mediator.CanArrive(t) {
            fmt.Printf("🚂 %s：收到，等待进站\n", t.name)
            return
        }
        fmt.Printf("🚂 %s：已进站\n", t.name)
    }

    func (t *FreightTrain) Depart() {
        fmt.Printf("🚂 %s：正在离站...\n", t.name)
        t.mediator.NotifyDeparture()
    }

    func (t *FreightTrain) PermitArrival() {
        fmt.Printf("🚂 %s：收到进站许可，正在进站...\n", t.name)
        fmt.Printf("🚂 %s：已进站\n", t.name)
    }
    ```

### 第四步：组装并使用

=== "📄 main.go: 客户端代码"

    ```go
    package main

    import "fmt"

    func main() {
        // 创建中介者（调度员）
        stationManager := NewStationManager()

        // 创建火车，注册到调度员
        passengerTrain := &PassengerTrain{
            mediator: stationManager,
            name:     "G1234 高铁",
        }
        freightTrain := &FreightTrain{
            mediator: stationManager,
            name:     "K5678 货运",
        }

        // 模拟场景：两列火车同时请求进站
        fmt.Println("=== 场景：两列火车同时到达 ===")
        passengerTrain.Arrive()  // 第一列进站成功
        freightTrain.Arrive()    // 第二列需要等待

        fmt.Println("\n=== 高铁离站 ===")
        passengerTrain.Depart()  // 高铁离站，货运自动获得进站许可
    }
    ```

=== "📄 output.txt: 执行结果"

    ```
    === 场景：两列火车同时到达 ===
    🚉 站台空闲，允许进站
    🚃 G1234 高铁：已进站
    ⏳ 站台占用，请等待...
    🚂 K5678 货运：收到，等待进站

    === 高铁离站 ===
    🚃 G1234 高铁：正在离站...
    🚂 K5678 货运：收到进站许可，正在进站...
    🚂 K5678 货运：已进站
    ```

## 中介者模式 vs 观察者模式

这两个模式容易混淆，关键区别在于：

| 特性 | 中介者模式 | 观察者模式 |
|------|-----------|-----------|
| **通信方向** | 双向：组件 ↔ 中介者 | 单向：发布者 → 订阅者 |
| **组件关系** | 组件互不知道对方存在 | 订阅者可以知道发布者 |
| **中心化程度** | 高度中心化（所有逻辑在中介者） | 分布式（每个对象处理自己的逻辑） |
| **典型场景** | GUI 组件协调、航空调度 | 事件通知、数据订阅 |

## 什么时候该用中介者？

| 场景 | 说明 |
|------|------|
| **对象间关系复杂** | 多对多的交互关系让代码像蜘蛛网 |
| **需要集中控制** | 想在一个地方管理所有交互逻辑 |
| **组件需要复用** | 组件不应该和特定的其他组件绑定 |

**常见应用** ：

- **GUI 框架**：表单验证、组件联动
- **聊天室**：消息路由和分发
- **航空/铁路调度**：资源协调
- **MVC 架构**：Controller 就是一种中介者

## 优缺点分析

| ✅ 优点 | ❌ 缺点 |
|---------|---------|
| **降低耦合**：组件之间不直接依赖 | **中介者可能变成"上帝对象"**：承担过多职责 |
| **集中管理**：交互逻辑清晰可见 | **单点故障**：中介者挂了，整个系统瘫痪 |
| **易于复用**：组件可以在不同中介者下工作 | |

## 与其他模式的关系

| 模式组合 | 说明 |
|----------|------|
| **中介者 + 观察者** | 中介者通过观察者模式接收组件通知 |
| **中介者 vs 外观** | 外观是单向简化接口，中介者是双向协调 |

> **一句话总结**：中介者模式就像微信群——成员之间不需要互加好友，所有消息都发到群里，由群来分发。
