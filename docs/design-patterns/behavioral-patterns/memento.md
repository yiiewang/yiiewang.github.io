---
title: 备忘录模式
date: 2023-02-25 17:18:47
category:
  - 设计模式
tag:
  - 行为模式
---

# 备忘录模式：给程序一个"后悔药"

玩游戏时你一定用过"存档"功能：打 Boss 前存个档，挂了就读档重来。备忘录模式就是程序世界的"存档系统"——在不破坏对象封装的前提下，保存和恢复对象的内部状态。

**备忘录模式**允许你捕获对象的内部状态，并在需要时恢复到之前的状态，就像时光机一样。

## 为什么需要备忘录？

假设你在开发一个文本编辑器，需要支持"撤销"功能：

```go
// ❌ 糟糕的写法：直接暴露内部状态
type Editor struct {
    Content string  // 公开内部状态，破坏封装
}

// 外部代码直接操作
backup := editor.Content
editor.Content = "新内容"
// 撤销
editor.Content = backup
```

这样写的问题：

- **破坏封装**：内部状态完全暴露给外部
- **状态管理混乱**：谁来保存这些备份？
- **难以扩展**：如果对象有多个字段，每个都要备份

备忘录模式的解法：**让对象自己创建和恢复快照，外部只管保存，不关心快照内容**。

## 模式结构

![备忘录模式结构](https://refactoringguru.cn/images/patterns/diagrams/memento/structure1.png)

| 角色 | 职责 | 类比 |
|------|------|------|
| **Originator（原发器）** | 创建自身状态的快照，也能从快照恢复 | 游戏角色 |
| **Memento（备忘录）** | 存储原发器的内部状态（不可变） | 存档文件 |
| **Caretaker（管理者）** | 保存备忘录，但不能查看或修改内容 | 存档管理器 |

## 动手实现：文本编辑器的撤销功能

用一个简单的文本编辑器来演示备忘录模式。

### 第一步：定义备忘录（存档）

=== "📄 memento.go: 备忘录"

    ```go
    package main

    // Memento 备忘录，存储编辑器的状态快照
    // 注意：字段不导出，外部无法直接访问
    type Memento struct {
        state string
    }

    // GetState 获取保存的状态（仅供原发器使用）
    func (m *Memento) GetState() string {
        return m.state
    }
    ```

### 第二步：定义原发器（编辑器）

=== "📄 editor.go: 原发器"

    ```go
    package main

    import "fmt"

    // Editor 文本编辑器，可以创建和恢复快照
    type Editor struct {
        content string
    }

    // SetContent 设置内容
    func (e *Editor) SetContent(content string) {
        e.content = content
        fmt.Printf("📝 编辑器内容变更为：「%s」\n", content)
    }

    // GetContent 获取当前内容
    func (e *Editor) GetContent() string {
        return e.content
    }

    // Save 创建当前状态的快照
    func (e *Editor) Save() *Memento {
        fmt.Printf("💾 保存当前状态：「%s」\n", e.content)
        return &Memento{state: e.content}
    }

    // Restore 从快照恢复状态
    func (e *Editor) Restore(m *Memento) {
        e.content = m.GetState()
        fmt.Printf("⏪ 恢复到状态：「%s」\n", e.content)
    }
    ```

### 第三步：定义管理者（历史记录）

=== "📄 history.go: 管理者"

    ```go
    package main

    import "fmt"

    // History 历史记录管理器
    type History struct {
        mementos []*Memento
    }

    // Push 保存一个备忘录
    func (h *History) Push(m *Memento) {
        h.mementos = append(h.mementos, m)
    }

    // Pop 取出最近的备忘录
    func (h *History) Pop() *Memento {
        if len(h.mementos) == 0 {
            fmt.Println("⚠️ 没有可恢复的状态")
            return nil
        }
        // 取出最后一个
        last := h.mementos[len(h.mementos)-1]
        h.mementos = h.mementos[:len(h.mementos)-1]
        return last
    }
    ```

### 第四步：组装并使用

=== "📄 main.go: 客户端代码"

    ```go
    package main

    import "fmt"

    func main() {
        editor := &Editor{}
        history := &History{}

        // 第一次编辑
        editor.SetContent("Hello")
        history.Push(editor.Save())

        // 第二次编辑
        editor.SetContent("Hello, World")
        history.Push(editor.Save())

        // 第三次编辑
        editor.SetContent("Hello, World!")

        fmt.Println("\n=== 开始撤销 ===")

        // 撤销到第二次编辑的状态
        editor.Restore(history.Pop())

        // 撤销到第一次编辑的状态
        editor.Restore(history.Pop())

        fmt.Printf("\n✅ 最终内容：「%s」\n", editor.GetContent())
    }
    ```

=== "📄 output.txt: 执行结果"

    ```
    📝 编辑器内容变更为：「Hello」
    💾 保存当前状态：「Hello」
    📝 编辑器内容变更为：「Hello, World」
    💾 保存当前状态：「Hello, World」
    📝 编辑器内容变更为：「Hello, World!」

    === 开始撤销 ===
    ⏪ 恢复到状态：「Hello, World」
    ⏪ 恢复到状态：「Hello」

    ✅ 最终内容：「Hello」
    ```

## 三种实现方式

备忘录模式有三种常见的实现方式：

### 1. 嵌套类实现（经典）

```go
// 备忘录作为原发器的内部类
type Editor struct {
    type snapshot struct {
        content string
    }
}
```

**优点**：备忘录只能被原发器访问，封装性最好

### 2. 中间接口实现

```go
// 对外只暴露空接口
type Memento interface{}

// 内部使用具体类型
type editorMemento struct {
    content string
}
```

**优点**：管理者完全无法访问备忘录内容

### 3. 宽接口实现（本文示例）

```go
type Memento struct {
    state string
}

func (m *Memento) GetState() string
```

**优点**：简单直接，适合大多数场景

## 什么时候该用备忘录？

| 场景 | 说明 |
|------|------|
| **需要撤销/重做** | 文本编辑器、绘图软件、IDE |
| **需要事务回滚** | 数据库事务、游戏存档 |
| **需要状态快照** | 调试时保存程序状态 |
| **需要保护封装** | 不想暴露对象内部实现 |

**常见应用**：

- **文本编辑器**：Ctrl+Z 撤销
- **浏览器**：后退按钮
- **游戏**：存档/读档
- **数据库**：事务回滚
- **版本控制**：Git commit

## 优缺点分析

| ✅ 优点 | ❌ 缺点 |
|---------|---------|
| **保护封装**：外部无法访问对象内部状态 | **内存开销**：频繁创建快照会占用大量内存 |
| **简化原发器**：状态管理逻辑由管理者负责 | **管理者生命周期**：需要跟踪原发器生命周期，及时清理过期快照 |
| **支持撤销/重做**：可以任意回到历史状态 | **动态语言限制**：JS/Python 无法真正保护备忘录内容 |

## 与其他模式的关系

| 模式组合 | 说明 |
|----------|------|
| **备忘录 + 命令** | 命令执行前用备忘录保存状态，支持撤销 |
| **备忘录 + 迭代器** | 保存迭代进度，支持暂停和继续 |
| **备忘录 vs 原型** | 原型是复制整个对象，备忘录只保存部分状态 |

> **一句话总结**：备忘录模式就像手机的"照片"功能——拍下此刻的状态，以后随时可以回看，但照片本身不能被修改。
