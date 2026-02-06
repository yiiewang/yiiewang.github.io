---
title: 组合模式
date: 2023-02-25 17:18:47
category:
  - 设计模式
tag:
  - 结构型模式
---

# 🌲 组合模式：让"文件"和"文件夹"一视同仁

> 组合模式是一种结构型设计模式，你可以使用它将对象组合成树状结构，并且能像使用独立对象一样使用它们。

## 从生活场景说起

打开你电脑的文件管理器，你会看到一个熟悉的树形结构：

```
📁 项目文件夹
├── 📄 README.md
├── 📄 main.go
└── 📁 src
    ├── 📄 handler.go
    └── 📁 utils
        ├── 📄 logger.go
        └── 📄 helper.go
```

现在，假设你要搜索关键词 "TODO"：

- 对于**文件**：直接搜索文件内容
- 对于**文件夹**：递归搜索所有子文件和子文件夹

关键问题来了：**你希望用同一个 `search()` 方法来处理文件和文件夹吗？**

当然希望！这就是组合模式要解决的问题。

## 为什么需要组合模式？

### 😩 没有组合模式时的麻烦

```go
// 每次操作都要判断类型
func searchInPath(path string, keyword string) {
    info, _ := os.Stat(path)
    
    if info.IsDir() {
        // 是文件夹：递归处理
        files, _ := ioutil.ReadDir(path)
        for _, f := range files {
            searchInPath(filepath.Join(path, f.Name()), keyword)
        }
    } else {
        // 是文件：直接搜索
        content, _ := ioutil.ReadFile(path)
        if strings.Contains(string(content), keyword) {
            fmt.Println("Found in:", path)
        }
    }
}
```

问题：

- 客户端代码需要知道"文件"和"文件夹"的区别
- 新增节点类型（如符号链接）需要修改客户端代码
- 代码充斥着 `if-else` 类型判断

### 😊 组合模式的优雅解法

```go
// 统一接口
type Component interface {
    search(keyword string)
}

// 文件实现
type File struct { name string }
func (f *File) search(keyword string) {
    fmt.Printf("搜索文件 %s 中的 '%s'\n", f.name, keyword)
}

// 文件夹实现
type Folder struct { 
    name string
    children []Component  // 可以包含文件或子文件夹
}
func (f *Folder) search(keyword string) {
    fmt.Printf("搜索文件夹 %s\n", f.name)
    for _, child := range f.children {
        child.search(keyword)  // 统一调用，不用判断类型
    }
}

// 客户端代码
root.search("TODO")  // 不管 root 是文件还是文件夹，调用方式一样
```

## 模式结构

```
┌────────────────────────────────────────────────────────────────┐
│                                                                │
│                    ┌──────────────┐                            │
│                    │  Component   │ ◀─── 统一接口              │
│                    │  (组件接口)   │                           │
│                    └──────┬───────┘                            │
│                           │                                    │
│              ┌────────────┴────────────┐                       │
│              ▼                         ▼                       │
│       ┌───────────┐            ┌──────────────┐               │
│       │   Leaf    │            │  Composite   │               │
│       │  (叶节点)  │            │   (容器)     │               │
│       │   📄      │            │    📁        │               │
│       └───────────┘            └──────┬───────┘               │
│                                       │                        │
│                                       ▼                        │
│                              ┌────────────────┐               │
│                              │ []Component    │               │
│                              │ (子组件列表)    │               │
│                              └────────────────┘               │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

### 角色说明

| 角色 | 职责 | 示例 |
|------|------|------|
| **组件** (Component) | 声明统一接口 | `Component` 接口 |
| **叶节点** (Leaf) | 基本元素，不包含子元素 | `File` 类 |
| **容器** (Composite) | 包含子元素的复合对象 | `Folder` 类 |
| **客户端** (Client) | 通过统一接口操作所有对象 | 调用 `search()` 的代码 |

## 完整实现

=== "component.go: 组件接口"

    ```go
    package main

    // Component 定义文件系统组件的统一接口
    type Component interface {
        search(keyword string)
        getName() string
    }
    ```

=== "file.go: 叶节点"

    ```go
    package main

    import "fmt"

    // File 叶节点：文件
    type File struct {
        name string
    }

    func (f *File) search(keyword string) {
        fmt.Printf("📄 在文件 [%s] 中搜索关键词 '%s'\n", f.name, keyword)
    }

    func (f *File) getName() string {
        return f.name
    }
    ```

=== "folder.go: 容器"

    ```go
    package main

    import "fmt"

    // Folder 容器：文件夹
    type Folder struct {
        components []Component
        name       string
    }

    func (f *Folder) search(keyword string) {
        fmt.Printf("📁 递归搜索文件夹 [%s] 中的 '%s'\n", f.name, keyword)
        for _, component := range f.components {
            component.search(keyword)  // 统一调用
        }
    }

    func (f *Folder) getName() string {
        return f.name
    }

    // add 添加子组件（可以是文件或文件夹）
    func (f *Folder) add(c Component) {
        f.components = append(f.components, c)
    }

    // remove 移除子组件
    func (f *Folder) remove(c Component) {
        for i, child := range f.components {
            if child.getName() == c.getName() {
                f.components = append(f.components[:i], f.components[i+1:]...)
                return
            }
        }
    }
    ```

=== "main.go: 客户端代码"

    ```go
    package main

    import "fmt"

    func main() {
        // 构建文件系统树
        //
        // 📁 project
        // ├── 📄 README.md
        // ├── 📄 main.go
        // └── 📁 src
        //     ├── 📄 handler.go
        //     └── 📁 utils
        //         └── 📄 logger.go

        // 创建文件
        file1 := &File{name: "README.md"}
        file2 := &File{name: "main.go"}
        file3 := &File{name: "handler.go"}
        file4 := &File{name: "logger.go"}

        // 创建 utils 文件夹
        utils := &Folder{name: "utils"}
        utils.add(file4)

        // 创建 src 文件夹
        src := &Folder{name: "src"}
        src.add(file3)
        src.add(utils)

        // 创建项目根目录
        project := &Folder{name: "project"}
        project.add(file1)
        project.add(file2)
        project.add(src)

        // 执行搜索 - 客户端不需要知道内部结构
        fmt.Println("=== 搜索整个项目 ===")
        project.search("TODO")

        fmt.Println("\n=== 只搜索 src 目录 ===")
        src.search("TODO")

        fmt.Println("\n=== 只搜索单个文件 ===")
        file1.search("TODO")
    }
    ```

=== "output.txt: 执行结果"

    ```
    === 搜索整个项目 ===
    📁 递归搜索文件夹 [project] 中的 'TODO'
    📄 在文件 [README.md] 中搜索关键词 'TODO'
    📄 在文件 [main.go] 中搜索关键词 'TODO'
    📁 递归搜索文件夹 [src] 中的 'TODO'
    📄 在文件 [handler.go] 中搜索关键词 'TODO'
    📁 递归搜索文件夹 [utils] 中的 'TODO'
    📄 在文件 [logger.go] 中搜索关键词 'TODO'

    === 只搜索 src 目录 ===
    📁 递归搜索文件夹 [src] 中的 'TODO'
    📄 在文件 [handler.go] 中搜索关键词 'TODO'
    📁 递归搜索文件夹 [utils] 中的 'TODO'
    📄 在文件 [logger.go] 中搜索关键词 'TODO'

    === 只搜索单个文件 ===
    📄 在文件 [README.md] 中搜索关键词 'TODO'
    ```

## 什么时候使用组合模式？

| 场景 | 说明 |
|------|------|
| 🌲 **树形结构** | 数据天然具有层级关系（文件系统、组织架构、菜单） |
| 🔄 **递归处理** | 需要对整个树结构执行相同的操作 |
| 📦 **整体-部分** | 希望客户端能一致地处理简单元素和复合元素 |
| 🎯 **统一接口** | 希望忽略对象的具体类型，统一处理 |

## 实际应用场景

### 1. 组织架构

```go
type Employee interface {
    getSalary() int
    getName() string
}

type Developer struct { name string; salary int }
type Manager struct { 
    name string
    salary int
    team []Employee  // 可以管理开发者或其他经理
}

func (m *Manager) getSalary() int {
    total := m.salary
    for _, e := range m.team {
        total += e.getSalary()  // 递归计算团队总薪资
    }
    return total
}
```

### 2. GUI 组件

```go
type Widget interface {
    render()
}

type Button struct { label string }
type Panel struct { 
    children []Widget  // 可以包含按钮、文本框、或其他面板
}

func (p *Panel) render() {
    fmt.Println("渲染面板")
    for _, child := range p.children {
        child.render()
    }
}
```

### 3. 商品分类

```go
type Category interface {
    getPrice() float64
}

type Product struct { price float64 }
type Bundle struct { 
    items []Category  // 套餐可以包含单品或其他套餐
    discount float64
}
```

## 优缺点分析

| ✅ 优点 | ❌ 缺点 |
|---------|---------|
| **简化客户端**：统一接口，无需判断对象类型 | **接口泛化**：可能需要在接口中声明所有子类的方法 |
| **开闭原则**：新增节点类型不影响现有代码 | **类型安全**：难以限制容器中的组件类型 |
| **灵活组合**：可以构建任意复杂的树结构 | **设计难度**：需要找到合适的通用接口 |

## 与其他模式的关系

| 模式 | 关系 |
|------|------|
| **迭代器** | 可以使用迭代器遍历组合树 |
| **访问者** | 可以使用访问者对整个组合树执行操作 |
| **装饰器** | 两者都使用递归组合，但装饰器只有一个子组件 |
| **享元** | 可以用享元实现组合树的共享叶节点 |
| **责任链** | 叶组件可以沿父组件链传递请求 |
| **生成器** | 可以用生成器递归构建复杂组合树 |

!!! info "2011 综合知识 33,34"
    组合（Composite）模式又称为整体-部分（Part whole）模式，属于对象的结构模式。在组合模式中，通过组合多个对象形成树形结构以表示整体部分的结构层次。组合模式对单个对象（即叶子对象）和组合对象（即容器对象）的使用具有一致性。
    
    * 类 Component 为组合中的对象声明接口；
    * 类 Leaf 在组合中表示叶结点对象，叶结点没有子结点；
    * 类 Composite 定义有子部件的那些部件的行为，存储子部件；
    * 类 Client 通过 Component 接口操纵组合部件的对象。

---

!!! tip "一句话总结"
    **组合模式就是"套娃"**：容器里可以放单品，也可以放其他容器，但对外提供统一的接口，客户端不用关心里面装的是什么。
