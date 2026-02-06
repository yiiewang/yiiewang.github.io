---
title: 桥接模式
date: 2023-02-25 17:18:47
category:
  - 设计模式
tag:
  - 结构型模式
---

# 🌉 桥接模式：让"形状"和"颜色"各自演化

> 桥接模式是一种结构型设计模式，它可将一个大类或一系列紧密相关的类拆分为抽象和实现两个独立的层次结构，从而能在开发时分别使用。

## 从生活场景说起

想象你在开一家电脑专卖店，同时销售 **Mac** 和 **Windows** 电脑。现在你需要为每台电脑配打印机，有 **爱普生** 和 **惠普** 两个品牌可选。

按照传统思路，你可能会这样设计产品组合：

- Mac + 爱普生打印机
- Mac + 惠普打印机  
- Windows + 爱普生打印机
- Windows + 惠普打印机

这是 **2 × 2 = 4** 种组合。如果再加一种电脑（Linux），再加一种打印机（佳能），就变成 **3 × 3 = 9** 种组合！

```
组合爆炸！😱
电脑种类 × 打印机种类 = 产品组合数
```

### 桥接模式的解法

与其创建 N × M 种组合类，不如把问题拆成两个独立的维度：

- **电脑维度**：Mac、Windows、Linux...
- **打印机维度**：爱普生、惠普、佳能...

两个维度通过一座"桥"连接：**电脑持有打印机的引用**。

```
┌─────────────┐          ┌─────────────┐
│   电脑层     │   桥     │   打印机层   │
│  (抽象)     │◀────────▶│  (实现)     │
├─────────────┤          ├─────────────┤
│ Mac         │          │ Epson       │
│ Windows     │          │ HP          │
│ Linux       │          │ Canon       │
└─────────────┘          └─────────────┘

组合方式：电脑.SetPrinter(打印机)
```

这样，新增电脑或打印机都只需要增加一个类，而不是 N 个组合类。

## 为什么需要桥接？

### 😩 没有桥接时的类爆炸

```go
// 错误做法：为每种组合创建一个类
type MacWithEpson struct{}
type MacWithHP struct{}
type MacWithCanon struct{}
type WindowsWithEpson struct{}
type WindowsWithHP struct{}
type WindowsWithCanon struct{}
// ... 组合爆炸！

// 新增一种打印机？所有电脑类型都要加一遍
// 新增一种电脑？所有打印机品牌都要配一遍
```

### 😊 桥接模式的优雅解法

```go
// 电脑和打印机分开定义
type Computer interface {
    Print()
    SetPrinter(Printer)
}

type Printer interface {
    PrintFile()
}

// Mac 持有任意打印机的引用
type Mac struct {
    printer Printer  // 这就是"桥"
}

func (m *Mac) Print() {
    fmt.Println("Mac 发起打印请求")
    m.printer.PrintFile()  // 委托给打印机
}

// 运行时灵活组合
mac := &Mac{}
mac.SetPrinter(&Epson{})   // Mac + 爱普生
mac.SetPrinter(&HP{})      // Mac + 惠普
```

## 模式结构

```
┌────────────────────────────────────────────────────────────────┐
│                                                                │
│  抽象层 (Abstraction)              实现层 (Implementation)      │
│                                                                │
│  ┌──────────────┐                 ┌──────────────┐            │
│  │  Computer    │    "桥"        │   Printer    │            │
│  │  (抽象接口)   │◀──────────────▶│  (实现接口)   │            │
│  └──────┬───────┘                 └──────┬───────┘            │
│         │                                │                    │
│    ┌────┴────┐                     ┌─────┴─────┐              │
│    ▼         ▼                     ▼           ▼              │
│ ┌─────┐  ┌─────────┐          ┌───────┐   ┌───────┐          │
│ │ Mac │  │ Windows │          │ Epson │   │  HP   │          │
│ └─────┘  └─────────┘          └───────┘   └───────┘          │
│                                                                │
│  精确抽象 (Refined Abstraction)   具体实现 (Concrete Impl)      │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

### 角色说明

| 角色 | 职责 | 示例 |
|------|------|------|
| **抽象部分** (Abstraction) | 定义高层控制逻辑 | `Computer` 接口 |
| **精确抽象** (Refined Abstraction) | 抽象部分的具体实现 | `Mac`、`Windows` |
| **实现部分** (Implementation) | 定义底层实现接口 | `Printer` 接口 |
| **具体实现** (Concrete Implementation) | 实现部分的具体实现 | `Epson`、`HP` |

## 完整实现

=== "computer.go: 抽象层接口"

    ```go
    package main

    // Computer 定义电脑的抽象接口
    type Computer interface {
        Print()
        SetPrinter(Printer)
    }
    ```

=== "printer.go: 实现层接口"

    ```go
    package main

    // Printer 定义打印机的实现接口
    type Printer interface {
        PrintFile()
    }
    ```

=== "mac.go: 精确抽象"

    ```go
    package main

    import "fmt"

    // Mac 是 Computer 的具体实现
    type Mac struct {
        printer Printer  // 持有实现层的引用（桥）
    }

    func (m *Mac) Print() {
        fmt.Println("🍎 Mac 发起打印请求")
        m.printer.PrintFile()
    }

    func (m *Mac) SetPrinter(p Printer) {
        m.printer = p
    }
    ```

=== "windows.go: 精确抽象"

    ```go
    package main

    import "fmt"

    // Windows 是 Computer 的另一个具体实现
    type Windows struct {
        printer Printer
    }

    func (w *Windows) Print() {
        fmt.Println("🪟 Windows 发起打印请求")
        w.printer.PrintFile()
    }

    func (w *Windows) SetPrinter(p Printer) {
        w.printer = p
    }
    ```

=== "epson.go: 具体实现"

    ```go
    package main

    import "fmt"

    // Epson 爱普生打印机
    type Epson struct{}

    func (p *Epson) PrintFile() {
        fmt.Println("   └─ 📄 使用 EPSON 打印机打印文件")
    }
    ```

=== "hp.go: 具体实现"

    ```go
    package main

    import "fmt"

    // HP 惠普打印机
    type Hp struct{}

    func (p *Hp) PrintFile() {
        fmt.Println("   └─ 📄 使用 HP 打印机打印文件")
    }
    ```

=== "main.go: 客户端代码"

    ```go
    package main

    import "fmt"

    func main() {
        hpPrinter := &Hp{}
        epsonPrinter := &Epson{}

        // Mac 电脑
        macComputer := &Mac{}
        
        fmt.Println("=== Mac + HP ===")
        macComputer.SetPrinter(hpPrinter)
        macComputer.Print()

        fmt.Println("\n=== Mac + Epson ===")
        macComputer.SetPrinter(epsonPrinter)
        macComputer.Print()

        // Windows 电脑
        winComputer := &Windows{}

        fmt.Println("\n=== Windows + HP ===")
        winComputer.SetPrinter(hpPrinter)
        winComputer.Print()

        fmt.Println("\n=== Windows + Epson ===")
        winComputer.SetPrinter(epsonPrinter)
        winComputer.Print()
    }
    ```

=== "output.txt: 执行结果"

    ```
    === Mac + HP ===
    🍎 Mac 发起打印请求
       └─ 📄 使用 HP 打印机打印文件

    === Mac + Epson ===
    🍎 Mac 发起打印请求
       └─ 📄 使用 EPSON 打印机打印文件

    === Windows + HP ===
    🪟 Windows 发起打印请求
       └─ 📄 使用 HP 打印机打印文件

    === Windows + Epson ===
    🪟 Windows 发起打印请求
       └─ 📄 使用 EPSON 打印机打印文件
    ```

## 什么时候使用桥接？

| 场景 | 说明 |
|------|------|
| 🎨 **多维度变化** | 类有多个独立变化的维度（如形状×颜色、平台×功能） |
| 📦 **避免类爆炸** | 用继承会导致子类数量爆炸式增长 |
| 🔧 **运行时切换实现** | 需要在运行时动态切换具体实现 |
| 🏗️ **跨平台开发** | GUI 框架需要支持多个操作系统 |

## 优缺点分析

| ✅ 优点 | ❌ 缺点 |
|---------|---------|
| **平台无关**：可以创建与平台无关的类和程序 | **复杂度增加**：对于高内聚的类，可能过度设计 |
| **开闭原则**：新增抽象和实现互不影响 | **理解成本**：初学者可能难以理解抽象与实现的分离 |
| **单一职责**：抽象处理高层逻辑，实现处理平台细节 | |
| **灵活组合**：运行时动态组合任意抽象和实现 | |

## 与其他模式的关系

| 模式 | 关系 |
|------|------|
| **适配器模式** | 桥接在**开发前期**设计，分离抽象和实现；适配器在**后期**使用，解决已有接口不兼容 |
| **策略模式** | 两者结构相似，但意图不同。策略用于切换算法，桥接用于分离抽象层次 |
| **抽象工厂** | 可以结合使用：抽象工厂创建桥接中的具体实现对象 |
| **生成器模式** | 可以结合使用：主管类负责抽象，生成器负责实现 |

## 实际应用场景

### 跨平台 GUI 框架

```go
// 窗口抽象（与平台无关）
type Window interface {
    Draw()
    SetRenderer(Renderer)
}

// 渲染器实现（与平台相关）
type Renderer interface {
    RenderCircle()
    RenderRectangle()
}

// 具体窗口类型
type DialogWindow struct { renderer Renderer }
type MainWindow struct { renderer Renderer }

// 具体平台渲染器
type WindowsRenderer struct{}
type MacRenderer struct{}
type LinuxRenderer struct{}
```

### 数据库驱动

```go
// 数据访问抽象
type Repository interface {
    Save(entity interface{})
    Find(id string) interface{}
}

// 数据库驱动实现
type DBDriver interface {
    Connect()
    Execute(sql string)
}

// MySQL、PostgreSQL、SQLite 各自实现 DBDriver
```

---

!!! tip "一句话总结"
    **桥接模式就是把"是什么"和"怎么做"分开**：抽象定义"做什么"，实现定义"怎么做"，两者通过组合（而非继承）连接，各自独立演化。
