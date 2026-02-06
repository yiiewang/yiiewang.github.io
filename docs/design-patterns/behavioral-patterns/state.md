---
title: 状态模式
date: 2023-02-25 17:18:47
category:
  - 设计模式
tag:
  - 行为模式
---

# 状态模式：让对象"变身"

你有没有注意过手机的行为会随着"状态"变化？静音模式下来电不响铃，飞行模式下无法上网。同一部手机，不同状态下表现完全不同。这就是状态模式的核心思想。

**状态模式**让一个对象在内部状态改变时，改变它的行为。看起来就像对象换了一个类一样。

## 为什么需要状态模式？

假设你在开发一个自动售货机，它有多种状态：

```go
// ❌ 糟糕的写法：用 if-else 处理所有状态
func (m *VendingMachine) InsertMoney() {
    if m.state == "NO_ITEM" {
        fmt.Println("没货，退钱")
    } else if m.state == "HAS_ITEM" {
        fmt.Println("请先选择商品")
    } else if m.state == "ITEM_SELECTED" {
        fmt.Println("收到钱，出货")
        m.state = "HAS_MONEY"
    } else if m.state == "HAS_MONEY" {
        fmt.Println("已收过钱了")
    }
    // 每个方法都要这样写一遍...
}
```

这样写的问题：

- **代码臃肿**：每个方法都有大量 if-else
- **难以维护**：新增状态要改所有方法
- **状态转换不清晰**：很难看出完整的状态流转图

状态模式的解法：**把每种状态封装成独立的类，让状态类自己决定如何响应操作**。

## 模式结构

![状态模式结构](https://refactoringguru.cn/images/patterns/diagrams/state/structure-zh.png?id=9d132abe67abef895172aad954f1daaf)

| 角色 | 职责 | 类比 |
|------|------|------|
| **Context（上下文）** | 持有当前状态对象，把请求委托给状态处理 | 售货机本体 |
| **State（状态接口）** | 定义所有状态共有的行为 | 售货机的操作面板 |
| **ConcreteState（具体状态）** | 实现特定状态下的行为 | 各种状态：有货、无货、已投币... |

## 动手实现：自动售货机

用自动售货机来演示状态模式。售货机有四种状态：**有货 → 已选商品 → 已投币 → 出货**。

### 第一步：定义状态接口

=== "📄 state.go: 状态接口"

    ```go
    package main

    // State 状态接口，定义售货机的所有操作
    type State interface {
        SelectItem() error   // 选择商品
        InsertMoney() error  // 投入钱币
        Dispense() error     // 出货
        AddItem() error      // 补货
    }
    ```

### 第二步：实现各种具体状态

=== "📄 has_item_state.go: 有货状态"

    ```go
    package main

    import "fmt"

    // HasItemState 有货状态
    type HasItemState struct {
        machine *VendingMachine
    }

    func (s *HasItemState) SelectItem() error {
        fmt.Println("🛒 商品已选择")
        s.machine.SetState(s.machine.itemSelected)
        return nil
    }

    func (s *HasItemState) InsertMoney() error {
        return fmt.Errorf("❌ 请先选择商品")
    }

    func (s *HasItemState) Dispense() error {
        return fmt.Errorf("❌ 请先选择商品并付款")
    }

    func (s *HasItemState) AddItem() error {
        fmt.Println("📦 补货成功")
        s.machine.itemCount++
        return nil
    }
    ```

=== "📄 item_selected_state.go: 已选商品状态"

    ```go
    package main

    import "fmt"

    // ItemSelectedState 已选商品状态
    type ItemSelectedState struct {
        machine *VendingMachine
    }

    func (s *ItemSelectedState) SelectItem() error {
        return fmt.Errorf("❌ 已选择商品，请投币")
    }

    func (s *ItemSelectedState) InsertMoney() error {
        fmt.Println("💰 收到付款")
        s.machine.SetState(s.machine.hasMoney)
        return nil
    }

    func (s *ItemSelectedState) Dispense() error {
        return fmt.Errorf("❌ 请先付款")
    }

    func (s *ItemSelectedState) AddItem() error {
        return fmt.Errorf("❌ 交易进行中，无法补货")
    }
    ```

=== "📄 has_money_state.go: 已投币状态"

    ```go
    package main

    import "fmt"

    // HasMoneyState 已投币状态
    type HasMoneyState struct {
        machine *VendingMachine
    }

    func (s *HasMoneyState) SelectItem() error {
        return fmt.Errorf("❌ 已付款，正在出货")
    }

    func (s *HasMoneyState) InsertMoney() error {
        return fmt.Errorf("❌ 已付过款")
    }

    func (s *HasMoneyState) Dispense() error {
        fmt.Println("🎁 出货中...")
        s.machine.itemCount--
        if s.machine.itemCount == 0 {
            fmt.Println("⚠️ 商品已售罄")
            s.machine.SetState(s.machine.noItem)
        } else {
            s.machine.SetState(s.machine.hasItem)
        }
        return nil
    }

    func (s *HasMoneyState) AddItem() error {
        return fmt.Errorf("❌ 交易进行中，无法补货")
    }
    ```

=== "📄 no_item_state.go: 无货状态"

    ```go
    package main

    import "fmt"

    // NoItemState 无货状态
    type NoItemState struct {
        machine *VendingMachine
    }

    func (s *NoItemState) SelectItem() error {
        return fmt.Errorf("❌ 商品已售罄")
    }

    func (s *NoItemState) InsertMoney() error {
        return fmt.Errorf("❌ 商品已售罄，无法付款")
    }

    func (s *NoItemState) Dispense() error {
        return fmt.Errorf("❌ 无商品可出")
    }

    func (s *NoItemState) AddItem() error {
        fmt.Println("📦 补货成功")
        s.machine.itemCount++
        s.machine.SetState(s.machine.hasItem)
        return nil
    }
    ```

### 第三步：实现上下文（售货机）

=== "📄 vending_machine.go: 上下文"

    ```go
    package main

    import "fmt"

    // VendingMachine 自动售货机
    type VendingMachine struct {
        hasItem      State
        itemSelected State
        hasMoney     State
        noItem       State

        currentState State
        itemCount    int
    }

    // NewVendingMachine 创建售货机
    func NewVendingMachine(itemCount int) *VendingMachine {
        m := &VendingMachine{itemCount: itemCount}
        
        // 初始化所有状态
        m.hasItem = &HasItemState{machine: m}
        m.itemSelected = &ItemSelectedState{machine: m}
        m.hasMoney = &HasMoneyState{machine: m}
        m.noItem = &NoItemState{machine: m}

        // 设置初始状态
        if itemCount > 0 {
            m.currentState = m.hasItem
        } else {
            m.currentState = m.noItem
        }
        return m
    }

    func (m *VendingMachine) SetState(s State) {
        m.currentState = s
    }

    func (m *VendingMachine) SelectItem() error {
        return m.currentState.SelectItem()
    }

    func (m *VendingMachine) InsertMoney() error {
        return m.currentState.InsertMoney()
    }

    func (m *VendingMachine) Dispense() error {
        return m.currentState.Dispense()
    }

    func (m *VendingMachine) AddItem() error {
        return m.currentState.AddItem()
    }

    func (m *VendingMachine) PrintStatus() {
        fmt.Printf("📊 库存: %d 件\n", m.itemCount)
    }
    ```

### 第四步：使用售货机

=== "📄 main.go: 客户端代码"

    ```go
    package main

    import "fmt"

    func main() {
        // 创建售货机，初始库存 1 件
        machine := NewVendingMachine(1)
        machine.PrintStatus()

        fmt.Println("\n=== 正常购买流程 ===")
        machine.SelectItem()    // 选择商品
        machine.InsertMoney()   // 投币
        machine.Dispense()      // 出货
        machine.PrintStatus()

        fmt.Println("\n=== 售罄后尝试购买 ===")
        err := machine.SelectItem()
        if err != nil {
            fmt.Println(err)
        }

        fmt.Println("\n=== 补货后继续购买 ===")
        machine.AddItem()
        machine.SelectItem()
        machine.InsertMoney()
        machine.Dispense()
    }
    ```

=== "📄 output.txt: 执行结果"

    ```
    📊 库存: 1 件

    === 正常购买流程 ===
    🛒 商品已选择
    💰 收到付款
    🎁 出货中...
    ⚠️ 商品已售罄
    📊 库存: 0 件

    === 售罄后尝试购买 ===
    ❌ 商品已售罄

    === 补货后继续购买 ===
    📦 补货成功
    🛒 商品已选择
    💰 收到付款
    🎁 出货中...
    ```

## 状态模式 vs 策略模式

这两个模式结构相似，但意图不同：

| 特性 | 状态模式 | 策略模式 |
|------|---------|---------|
| **切换时机** | 自动切换（内部触发） | 手动切换（外部指定） |
| **状态之间** | 知道彼此存在，可以互相切换 | 互不知道，完全独立 |
| **典型场景** | 有限状态机、工作流 | 算法选择、行为参数化 |

## 什么时候该用状态模式？

| 场景 | 说明 |
|------|------|
| **对象行为依赖状态** | 不同状态下同一操作有不同表现 |
| **状态数量多** | 大量条件分支判断当前状态 |
| **状态转换逻辑复杂** | 需要清晰的状态转换图 |

**常见应用**：

- **订单系统**：待支付 → 已支付 → 已发货 → 已收货
- **游戏角色**：正常 → 受伤 → 死亡
- **TCP 连接**：监听 → 已建立 → 关闭
- **文档审批**：草稿 → 审核中 → 已批准/已拒绝

## 优缺点分析

| ✅ 优点 | ❌ 缺点 |
|---------|---------|
| **消除条件分支**：每个状态类职责单一 | **类数量增加**：每个状态一个类 |
| **开闭原则**：新增状态无需修改现有代码 | **状态少时过度设计**：简单情况用 if-else 更直接 |
| **状态转换清晰**：每个状态知道下一个状态是谁 | |

## 与其他模式的关系

| 模式组合 | 说明 |
|----------|------|
| **状态 vs 策略** | 策略是"怎么做"，状态是"现在是什么情况" |
| **状态 + 单例** | 状态对象通常可以做成单例 |
| **状态 + 享元** | 共享状态对象以节省内存 |

> **一句话总结**：状态模式就像变形金刚——同一个机器人，在不同模式下有完全不同的形态和能力。
