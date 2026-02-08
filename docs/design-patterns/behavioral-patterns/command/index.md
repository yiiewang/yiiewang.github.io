---
title: 命令模式
date: 2023-02-25 17:18:47
category:
  - 设计模式
tag:
  - 行为模式
---

# 命令模式：把操作变成"快递包裹"

想象你在餐厅点餐：你不会直接冲进厨房告诉厨师"给我做个汉堡"，而是把订单写在纸上，服务员会在合适的时机把它交给厨房。订单就像一个"命令包裹"——封装了你的请求，可以排队、可以取消、甚至可以追溯谁点了什么。

**命令模式** 将请求封装成独立的对象，让你能够参数化、排队、记录日志，甚至支持撤销操作。

## 为什么需要命令模式？

假设你在开发一个智能家居系统，需要支持多种设备的控制：

```go
// ❌ 糟糕的写法：直接调用设备方法
func controlDevice(device string, action string) {
    switch device {
    case "tv":
        if action == "on" {
            tv.TurnOn()
        } else {
            tv.TurnOff()
        }
    case "light":
        if action == "on" {
            light.TurnOn()
        }
        // ... 无穷无尽的 switch-case
    }
}
```

这样写的问题：

- **控制逻辑和设备紧耦合** ：新增设备要改核心代码
- **无法撤销** ：关掉电视后想再打开？重写逻辑
- **无法排队** ：想定时执行一系列操作？做不到

命令模式的解法： **把每个操作封装成独立的命令对象，让调用者只管"发命令"，不管具体怎么执行**。

## 模式结构

![命令模式结构](./structure.png)

| 角色 | 职责 | 类比 |
|------|------|------|
| **Invoker（调用者）** | 发起请求，但不知道谁来执行 | 遥控器按钮 |
| **Command（命令接口）** | 定义执行命令的统一接口 | 所有按钮的"按下"动作 |
| **ConcreteCommand（具体命令）** | 封装具体操作和接收者 | "开机"按钮、"关机"按钮 |
| **Receiver（接收者）** | 实际执行业务逻辑的对象 | 电视机本身 |
| **Client（客户端）** | 创建命令并配置调用者 | 组装遥控器的人 |

## 动手实现：电视遥控器

用一个电视遥控器的例子来演示命令模式。遥控器可以控制多种设备的开关。

### 第一步：定义接收者（设备）

=== "📄 device.go: 设备接口"

    ```go
    package main

    // Device 定义所有可控制设备的接口
    type Device interface {
        On()
        Off()
    }
    ```

=== "📄 tv.go: 具体设备 - 电视"

    ```go
    package main

    import "fmt"

    // TV 电视机，实际执行开关操作
    type TV struct {
        isRunning bool
    }

    func (t *TV) On() {
        t.isRunning = true
        fmt.Println("📺 电视已打开")
    }

    func (t *TV) Off() {
        t.isRunning = false
        fmt.Println("📺 电视已关闭")
    }
    ```

### 第二步：定义命令接口和具体命令

=== "📄 command.go: 命令接口"

    ```go
    package main

    // Command 命令接口，所有命令都实现这个接口
    type Command interface {
        Execute()
    }
    ```

=== "📄 on_command.go: 开机命令"

    ```go
    package main

    // OnCommand 开机命令
    type OnCommand struct {
        device Device
    }

    func (c *OnCommand) Execute() {
        c.device.On()
    }
    ```

=== "📄 off_command.go: 关机命令"

    ```go
    package main

    // OffCommand 关机命令
    type OffCommand struct {
        device Device
    }

    func (c *OffCommand) Execute() {
        c.device.Off()
    }
    ```

### 第三步：定义调用者（遥控器按钮）

=== "📄 button.go: 调用者"

    ```go
    package main

    // Button 遥控器按钮，持有一个命令对象
    type Button struct {
        command Command
    }

    // Press 按下按钮，执行关联的命令
    func (b *Button) Press() {
        b.command.Execute()
    }
    ```

### 第四步：组装并使用

=== "📄 main.go: 客户端代码"

    ```go
    package main

    func main() {
        // 1. 创建接收者（电视机）
        tv := &TV{}

        // 2. 创建命令对象，绑定到接收者
        onCommand := &OnCommand{device: tv}
        offCommand := &OffCommand{device: tv}

        // 3. 创建调用者（按钮），绑定命令
        onButton := &Button{command: onCommand}
        offButton := &Button{command: offCommand}

        // 4. 用户操作：按下按钮
        fmt.Println("用户按下「开机」按钮...")
        onButton.Press()

        fmt.Println("\n用户按下「关机」按钮...")
        offButton.Press()
    }
    ```

=== "📄 output.txt: 执行结果"

    ```
    用户按下「开机」按钮...
    📺 电视已打开

    用户按下「关机」按钮...
    📺 电视已关闭
    ```

## 进阶用法：命令队列与撤销

命令模式的威力远不止于此，它还能实现：

### 命令队列

```go
// CommandQueue 命令队列，支持批量执行
type CommandQueue struct {
    commands []Command
}

func (q *CommandQueue) Add(cmd Command) {
    q.commands = append(q.commands, cmd)
}

func (q *CommandQueue) ExecuteAll() {
    for _, cmd := range q.commands {
        cmd.Execute()
    }
}
```

### 撤销功能

```go
// UndoableCommand 可撤销的命令接口
type UndoableCommand interface {
    Command
    Undo()
}

// OnCommand 实现撤销
func (c *OnCommand) Undo() {
    c.device.Off() // 撤销开机 = 关机
}
```

## 什么时候该用命令模式？

| 场景 | 说明 |
|------|------|
| **需要参数化操作** | 把操作当作参数传递给方法 |
| **需要排队或延迟执行** | 任务队列、定时任务、宏命令 |
| **需要撤销/重做** | 文本编辑器、绘图软件、游戏存档 |
| **需要记录操作日志** | 事务系统、审计日志 |

**常见应用** ：

- **GUI 按钮和菜单项**：同一个命令可以绑定到按钮、快捷键、菜单
- **事务系统**：数据库事务的回滚机制
- **宏录制**：Office 宏、游戏宏
- **任务调度**：Cron Job、消息队列

## 优缺点分析

| ✅ 优点 | ❌ 缺点 |
|---------|---------|
| **解耦调用者和接收者**：遥控器不需要知道电视的内部实现 | **类数量增加**：每种操作都要一个命令类 |
| **支持撤销/重做**：命令对象可以保存状态 | **简单场景过度设计**：直接调用更简单 |
| **支持组合命令**：宏命令可以包含多个子命令 | |
| **支持延迟执行**：命令可以放入队列稍后执行 | |

## 与其他模式的关系

| 模式组合 | 说明 |
|----------|------|
| **命令 + 备忘录** | 实现撤销功能：命令执行前用备忘录保存状态 |
| **命令 + 责任链** | 命令在处理者链中传递，直到被执行 |
| **命令 vs 策略** | 策略是"怎么做"，命令是"做什么" |

> **一句话总结**：命令模式就像快递系统——你把包裹（请求）交给快递员（调用者），快递员不关心里面是什么，只负责送到收件人（接收者）手里。
