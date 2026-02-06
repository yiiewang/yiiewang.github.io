---
title: 外观模式
date: 2023-02-25 17:18:47
category:
  - 设计模式
tag:
  - 结构型模式
---

# 🏠 外观模式：复杂系统的"一键操作"

> 外观模式是一种结构型设计模式，能为程序库、框架或其他复杂类提供一个简单的接口——就像酒店的前台，帮你处理所有复杂的入住流程。

## 从生活场景说起

想象你要用信用卡在网上买一份披萨。表面上你只是点了"支付"按钮，但背后发生了一大堆事情：

1. 验证账户是否存在
2. 检查安全码是否正确
3. 检查余额是否充足
4. 扣除相应金额
5. 记录交易日志
6. 发送通知短信

如果让你直接调用这 6 个子系统，代码可能是这样的：

```go
// 😩 没有外观模式时
account.verify(accountID)
security.checkCode(code)
balance.check(amount)
balance.deduct(amount)
ledger.record(accountID, "debit", amount)
notification.send(accountID, "支付成功")
```

每次支付都要写这么多代码？而且还要记住正确的调用顺序？太痛苦了！

### 外观模式的解法

```go
// 😊 有外观模式后
wallet := NewWalletFacade(accountID, code)
wallet.Pay(amount)  // 一行代码搞定！
```

外观（Facade）就像一个**智能管家**：

- 你只需要告诉它"我要付款"
- 它会帮你协调所有子系统
- 你不需要知道背后的复杂流程

## 模式结构

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│      客户端                    外观                              │
│        │                       │                                │
│        │    简单接口           │                                │
│        └──────────────────────▶│                                │
│                                │                                │
│                      ┌─────────┴─────────┐                      │
│                      │   WalletFacade    │                      │
│                      │  ┌─────────────┐  │                      │
│                      │  │ Pay()       │  │                      │
│                      │  │ Recharge()  │  │                      │
│                      │  └─────────────┘  │                      │
│                      └─────────┬─────────┘                      │
│                                │                                │
│         ┌──────────┬───────────┼───────────┬──────────┐         │
│         ▼          ▼           ▼           ▼          ▼         │
│    ┌─────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌──────────┐   │
│    │ Account │ │Security│ │ Wallet │ │ Ledger │ │Notification│   │
│    │  账户    │ │ 安全码 │ │  余额  │ │  账本  │ │   通知    │   │
│    └─────────┘ └────────┘ └────────┘ └────────┘ └──────────┘   │
│                                                                 │
│                        复杂子系统                                │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 角色说明

| 角色 | 职责 | 示例 |
|------|------|------|
| **外观** (Facade) | 提供简化接口，协调子系统 | `WalletFacade` |
| **子系统** (Subsystem) | 实现具体功能的类 | `Account`, `Wallet`, `Ledger` 等 |
| **客户端** (Client) | 通过外观使用子系统 | 调用 `Pay()` 的代码 |

## 完整实现

=== "account.go: 账户子系统"

    ```go
    package main

    import "fmt"

    // Account 账户验证子系统
    type Account struct {
        name string
    }

    func newAccount(accountName string) *Account {
        return &Account{name: accountName}
    }

    func (a *Account) checkAccount(accountName string) error {
        if a.name != accountName {
            return fmt.Errorf("❌ 账户名不正确")
        }
        fmt.Println("✅ 账户验证通过")
        return nil
    }
    ```

=== "securityCode.go: 安全码子系统"

    ```go
    package main

    import "fmt"

    // SecurityCode 安全码验证子系统
    type SecurityCode struct {
        code int
    }

    func newSecurityCode(code int) *SecurityCode {
        return &SecurityCode{code: code}
    }

    func (s *SecurityCode) checkCode(incomingCode int) error {
        if s.code != incomingCode {
            return fmt.Errorf("❌ 安全码不正确")
        }
        fmt.Println("✅ 安全码验证通过")
        return nil
    }
    ```

=== "wallet.go: 钱包子系统"

    ```go
    package main

    import "fmt"

    // Wallet 钱包余额子系统
    type Wallet struct {
        balance int
    }

    func newWallet() *Wallet {
        return &Wallet{balance: 0}
    }

    func (w *Wallet) creditBalance(amount int) {
        w.balance += amount
        fmt.Printf("✅ 充值成功，当前余额: %d 元\n", w.balance)
    }

    func (w *Wallet) debitBalance(amount int) error {
        if w.balance < amount {
            return fmt.Errorf("❌ 余额不足，当前余额: %d 元", w.balance)
        }
        w.balance -= amount
        fmt.Printf("✅ 扣款成功，剩余余额: %d 元\n", w.balance)
        return nil
    }
    ```

=== "ledger.go: 账本子系统"

    ```go
    package main

    import "fmt"

    // Ledger 交易记录子系统
    type Ledger struct{}

    func (l *Ledger) makeEntry(accountID, txnType string, amount int) {
        fmt.Printf("📝 记录交易: 账户=%s, 类型=%s, 金额=%d\n", 
            accountID, txnType, amount)
    }
    ```

=== "notification.go: 通知子系统"

    ```go
    package main

    import "fmt"

    // Notification 通知子系统
    type Notification struct{}

    func (n *Notification) sendWalletCreditNotification() {
        fmt.Println("📱 发送通知: 钱包充值成功")
    }

    func (n *Notification) sendWalletDebitNotification() {
        fmt.Println("📱 发送通知: 钱包支付成功")
    }
    ```

=== "walletFacade.go: 外观"

    ```go
    package main

    import "fmt"

    // WalletFacade 钱包外观 - 统一入口
    type WalletFacade struct {
        account      *Account
        wallet       *Wallet
        securityCode *SecurityCode
        notification *Notification
        ledger       *Ledger
    }

    // NewWalletFacade 创建钱包外观
    func NewWalletFacade(accountID string, code int) *WalletFacade {
        fmt.Println("🏦 创建钱包账户...")
        facade := &WalletFacade{
            account:      newAccount(accountID),
            securityCode: newSecurityCode(code),
            wallet:       newWallet(),
            notification: &Notification{},
            ledger:       &Ledger{},
        }
        fmt.Println("✅ 账户创建成功\n")
        return facade
    }

    // Recharge 充值 - 简化接口
    func (w *WalletFacade) Recharge(accountID string, code int, amount int) error {
        fmt.Println("💰 开始充值流程...")
        
        // 1. 验证账户
        if err := w.account.checkAccount(accountID); err != nil {
            return err
        }
        // 2. 验证安全码
        if err := w.securityCode.checkCode(code); err != nil {
            return err
        }
        // 3. 增加余额
        w.wallet.creditBalance(amount)
        // 4. 记录交易
        w.ledger.makeEntry(accountID, "credit", amount)
        // 5. 发送通知
        w.notification.sendWalletCreditNotification()
        
        return nil
    }

    // Pay 支付 - 简化接口
    func (w *WalletFacade) Pay(accountID string, code int, amount int) error {
        fmt.Println("💳 开始支付流程...")
        
        // 1. 验证账户
        if err := w.account.checkAccount(accountID); err != nil {
            return err
        }
        // 2. 验证安全码
        if err := w.securityCode.checkCode(code); err != nil {
            return err
        }
        // 3. 扣除余额
        if err := w.wallet.debitBalance(amount); err != nil {
            return err
        }
        // 4. 记录交易
        w.ledger.makeEntry(accountID, "debit", amount)
        // 5. 发送通知
        w.notification.sendWalletDebitNotification()
        
        return nil
    }
    ```

=== "main.go: 客户端代码"

    ```go
    package main

    import (
        "fmt"
        "log"
    )

    func main() {
        // 创建钱包（外观对象）
        wallet := NewWalletFacade("user123", 1234)

        // 充值 - 只需一行代码
        fmt.Println("=== 充值 100 元 ===")
        if err := wallet.Recharge("user123", 1234, 100); err != nil {
            log.Fatalf("充值失败: %s\n", err.Error())
        }

        // 支付 - 只需一行代码
        fmt.Println("\n=== 支付 30 元 ===")
        if err := wallet.Pay("user123", 1234, 30); err != nil {
            log.Fatalf("支付失败: %s\n", err.Error())
        }

        // 再次支付
        fmt.Println("\n=== 支付 50 元 ===")
        if err := wallet.Pay("user123", 1234, 50); err != nil {
            log.Fatalf("支付失败: %s\n", err.Error())
        }
    }
    ```

=== "output.txt: 执行结果"

    ```
    🏦 创建钱包账户...
    ✅ 账户创建成功

    === 充值 100 元 ===
    💰 开始充值流程...
    ✅ 账户验证通过
    ✅ 安全码验证通过
    ✅ 充值成功，当前余额: 100 元
    📝 记录交易: 账户=user123, 类型=credit, 金额=100
    📱 发送通知: 钱包充值成功

    === 支付 30 元 ===
    💳 开始支付流程...
    ✅ 账户验证通过
    ✅ 安全码验证通过
    ✅ 扣款成功，剩余余额: 70 元
    📝 记录交易: 账户=user123, 类型=debit, 金额=30
    📱 发送通知: 钱包支付成功

    === 支付 50 元 ===
    💳 开始支付流程...
    ✅ 账户验证通过
    ✅ 安全码验证通过
    ✅ 扣款成功，剩余余额: 20 元
    📝 记录交易: 账户=user123, 类型=debit, 金额=50
    📱 发送通知: 钱包支付成功
    ```

## 什么时候使用外观模式？

| 场景 | 说明 |
|------|------|
| 🏗️ **简化复杂 API** | 第三方库或框架 API 过于复杂 |
| 📚 **封装子系统** | 需要将多个子系统的调用封装成简单接口 |
| 🔒 **解耦客户端** | 不希望客户端直接依赖复杂的子系统 |
| 🏢 **分层架构** | 为每个子系统层次创建入口点 |

## 实际应用场景

### 1. 数据库操作

```go
// 没有外观
db.Connect()
tx := db.BeginTransaction()
stmt := db.Prepare(sql)
result := stmt.Execute()
tx.Commit()
db.Close()

// 有外观
dbFacade.Execute(sql)  // 一行搞定
```

### 2. 视频转换

```go
// 复杂的视频转换流程
converter := NewVideoConverterFacade()
converter.Convert("video.mp4", "video.avi")  // 内部处理解码、编码、格式转换
```

### 3. 订单处理

```go
// 创建订单涉及多个系统
orderFacade := NewOrderFacade()
orderFacade.PlaceOrder(cart, payment, address)
// 内部：库存检查、价格计算、支付处理、物流安排、通知发送
```

## 优缺点分析

| ✅ 优点 | ❌ 缺点 |
|---------|---------|
| **简化接口**：客户端代码更简洁 | **可能成为上帝对象**：外观可能承担过多职责 |
| **解耦**：客户端与子系统解耦 | **额外的间接层**：可能影响性能 |
| **易于使用**：降低学习成本 | **可能限制灵活性**：高级用户可能需要直接访问子系统 |

## 与其他模式的关系

| 模式 | 关系 |
|------|------|
| **适配器** | 适配器改变单个对象的接口；外观为整个子系统提供新接口 |
| **抽象工厂** | 抽象工厂可以替代外观来隐藏子系统对象的创建 |
| **中介者** | 两者都组织协作，但中介者的组件知道中介者，外观的子系统不知道外观 |
| **单例** | 外观通常可以做成单例 |
| **代理** | 代理与服务接口相同；外观定义新接口 |

---

!!! tip "一句话总结"
    **外观模式就是"一站式服务"**：就像酒店前台帮你处理入住、订餐、叫车，你不需要分别联系客房部、餐厅、出租车公司——一个电话搞定所有事情。
