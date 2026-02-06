---
title: 适配器模式
date: 2023-02-25 17:18:47
category:
  - 设计模式
tag:
  - 结构型模式
---

# 🔌 适配器模式：让"方头"插进"圆孔"

> 适配器模式是一种结构型设计模式，它能使接口不兼容的对象能够相互合作——就像你用转换插头让国内电器在国外使用一样。

## 从生活场景说起

想象这样一个场景：你从美国带了一台 MacBook 回国，想给它充电。问题来了——美国用的是两扁头插座，而国内是三孔插座。

你不可能为了充电去改造家里的电路，也不会把 MacBook 的充电器拆了重做。最简单的方案是什么？**买一个转换插头**！

这个转换插头就是**适配器**：

- 🔌 一头能插进国内的三孔插座
- 🔌 另一头能接受美式两扁头
- 🔌 中间完成电压转换（如果需要的话）

在软件开发中，适配器模式做的就是类似的事情。

## 为什么需要适配器？

### 😩 没有适配器时的尴尬

假设你的系统一直使用 Lightning 接口的设备，现在来了一个只有 USB 接口的 Windows 电脑：

```go
// 你的客户端代码只认识 Lightning 接口
type Computer interface {
    InsertIntoLightningPort()
}

// Mac 电脑有 Lightning 接口，没问题
type Mac struct{}

func (m *Mac) InsertIntoLightningPort() {
    fmt.Println("Lightning 接口已连接到 Mac")
}

// Windows 电脑只有 USB 接口
type Windows struct{}

func (w *Windows) insertIntoUSBPort() {
    fmt.Println("USB 接口已连接到 Windows")
}

// 问题来了：Windows 没有实现 Computer 接口！
// client.Connect(windowsPC) // 编译错误！
```

怎么办？改 Windows 的代码？改客户端的代码？都不现实。

### 😊 用适配器优雅解决

```go
// 创建一个适配器，让 Windows 也能"假装"有 Lightning 接口
type WindowsAdapter struct {
    windowMachine *Windows
}

func (w *WindowsAdapter) InsertIntoLightningPort() {
    fmt.Println("适配器：将 Lightning 信号转换为 USB 信号...")
    w.windowMachine.insertIntoUSBPort()
}

// 现在可以无缝使用了
client.Connect(windowsAdapter) // ✅ 完美工作！
```

## 模式结构

```
┌─────────────────────────────────────────────────────────────────┐
│                         客户端代码                               │
│                            │                                    │
│                            ▼                                    │
│                   ┌─────────────────┐                           │
│                   │  目标接口        │◀─────── 客户端只认识这个  │
│                   │  (Lightning)    │                           │
│                   └────────┬────────┘                           │
│                            │                                    │
│            ┌───────────────┼───────────────┐                    │
│            ▼               ▼               ▼                    │
│     ┌───────────┐   ┌───────────┐   ┌───────────┐               │
│     │   Mac     │   │ 适配器    │   │ (其他实现)│               │
│     │ (原生支持)│   │           │   │           │               │
│     └───────────┘   └─────┬─────┘   └───────────┘               │
│                           │                                     │
│                           ▼                                     │
│                    ┌───────────┐                                │
│                    │ Windows   │◀────── 被适配的对象            │
│                    │ (USB接口) │                                │
│                    └───────────┘                                │
└─────────────────────────────────────────────────────────────────┘
```

### 角色说明

| 角色 | 职责 | 示例 |
|------|------|------|
| **目标接口** (Target) | 客户端期望的接口 | `Computer` 接口 |
| **被适配者** (Adaptee) | 需要被适配的现有类 | `Windows` 类 |
| **适配器** (Adapter) | 包装被适配者，实现目标接口 | `WindowsAdapter` |
| **客户端** (Client) | 通过目标接口使用对象 | 使用 `Computer` 接口的代码 |

## 两种适配器类型

### 对象适配器（推荐）

通过**组合**的方式，适配器持有被适配者的引用：

```go
type WindowsAdapter struct {
    windowMachine *Windows  // 组合：持有被适配者
}

func (w *WindowsAdapter) InsertIntoLightningPort() {
    w.windowMachine.insertIntoUSBPort()  // 委托给被适配者
}
```

### 类适配器

通过**多重继承**同时继承目标接口和被适配者（Go 不支持，但 C++ 支持）：

```cpp
// C++ 示例
class WindowsAdapter : public Computer, private Windows {
public:
    void InsertIntoLightningPort() override {
        insertIntoUSBPort();  // 直接调用继承的方法
    }
};
```

## 完整实现

=== "computer.go: 目标接口"

    ```go
    package main

    // Computer 定义了客户端期望的接口
    type Computer interface {
        InsertIntoLightningPort()
    }
    ```

=== "mac.go: 原生支持的服务"

    ```go
    package main

    import "fmt"

    // Mac 原生支持 Lightning 接口
    type Mac struct{}

    func (m *Mac) InsertIntoLightningPort() {
        fmt.Println("✅ Lightning 接口已连接到 Mac")
    }
    ```

=== "windows.go: 需要被适配的服务"

    ```go
    package main

    import "fmt"

    // Windows 只支持 USB 接口
    type Windows struct{}

    func (w *Windows) insertIntoUSBPort() {
        fmt.Println("✅ USB 接口已连接到 Windows")
    }
    ```

=== "windowsAdapter.go: 适配器"

    ```go
    package main

    import "fmt"

    // WindowsAdapter 适配器：让 Windows 也能接受 Lightning
    type WindowsAdapter struct {
        windowMachine *Windows
    }

    func (w *WindowsAdapter) InsertIntoLightningPort() {
        fmt.Println("🔄 适配器：将 Lightning 信号转换为 USB 信号")
        w.windowMachine.insertIntoUSBPort()
    }
    ```

=== "client.go: 客户端代码"

    ```go
    package main

    import "fmt"

    // Client 只知道如何使用 Computer 接口
    type Client struct{}

    func (c *Client) InsertLightningConnectorIntoComputer(com Computer) {
        fmt.Println("📱 客户端：插入 Lightning 连接器...")
        com.InsertIntoLightningPort()
    }
    ```

=== "main.go: 使用示例"

    ```go
    package main

    import "fmt"

    func main() {
        client := &Client{}

        fmt.Println("=== 连接 Mac（原生支持）===")
        mac := &Mac{}
        client.InsertLightningConnectorIntoComputer(mac)

        fmt.Println("\n=== 连接 Windows（通过适配器）===")
        windowsMachine := &Windows{}
        windowsMachineAdapter := &WindowsAdapter{
            windowMachine: windowsMachine,
        }
        client.InsertLightningConnectorIntoComputer(windowsMachineAdapter)
    }
    ```

=== "output.txt: 执行结果"

    ```
    === 连接 Mac（原生支持）===
    📱 客户端：插入 Lightning 连接器...
    ✅ Lightning 接口已连接到 Mac

    === 连接 Windows（通过适配器）===
    📱 客户端：插入 Lightning 连接器...
    🔄 适配器：将 Lightning 信号转换为 USB 信号
    ✅ USB 接口已连接到 Windows
    ```

## 什么时候使用适配器？

| 场景 | 说明 |
|------|------|
| 🔧 **集成第三方库** | 第三方库接口与你的系统不兼容 |
| 🏚️ **对接遗留系统** | 老系统接口无法修改，但需要与新系统协作 |
| 🔀 **统一多个类的接口** | 多个功能相似但接口不同的类需要统一使用 |
| 📦 **封装复杂 API** | 简化复杂 API 的调用方式 |

## 优缺点分析

| ✅ 优点 | ❌ 缺点 |
|---------|---------|
| **单一职责**：接口转换逻辑独立于业务逻辑 | **复杂度增加**：需要新增适配器类 |
| **开闭原则**：新增适配器不影响现有代码 | **间接层**：调用链变长，可能影响性能 |
| **灵活性**：可以适配多个不同的类 | **维护成本**：接口变化时需要同步更新适配器 |

## 与其他模式的关系

| 模式 | 关系 |
|------|------|
| **桥接模式** | 桥接在开发前期设计，用于分离抽象和实现；适配器在后期使用，解决接口不兼容问题 |
| **装饰模式** | 装饰增强功能但保持接口不变；适配器改变接口但不增强功能 |
| **代理模式** | 代理提供相同接口；适配器提供不同接口 |
| **外观模式** | 外观为整个子系统提供新接口；适配器通常只包装单个对象 |

## 实际应用场景

1. **数据库驱动**：database/sql 包通过 Driver 接口适配不同数据库
2. **日志库**：将不同日志库（logrus、zap）适配到统一接口
3. **支付网关**：将微信、支付宝等不同支付 API 适配到统一接口
4. **文件系统**：将本地文件、S3、OSS 等适配到统一的文件操作接口

---

!!! tip "一句话总结"
    **适配器模式就像翻译官**：它不改变任何一方的"语言"（接口），只是在中间做翻译转换，让原本无法沟通的双方能够协作。
