---
title: 原型模式
date: 2023-02-25 17:18:47
category:
  - 设计模式
tag:
  - 创建型模式
---

# 原型模式：复制比新建更快

你有没有用过"复制粘贴"？选中一个文件，Ctrl+C → Ctrl+V，瞬间就有了一个副本。原型模式就是程序世界的"复制粘贴"——**通过克隆现有对象来创建新对象，而不是从零开始构造**。

**原型模式**允许你复制已有对象，而无需让代码依赖它们所属的类。

## 为什么需要原型模式？

假设你需要复制一个复杂对象：

```go
// ❌ 糟糕的写法：手动复制每个字段
func copyFile(original *File) *File {
    copy := &File{
        Name:    original.Name,
        Size:    original.Size,
        Content: original.Content,
        // 私有字段怎么办？
        // 嵌套对象怎么办？
        // 如果不知道具体类型怎么办？
    }
    return copy
}
```

这样写的问题：

- **私有字段无法访问**：外部代码看不到私有成员
- **嵌套对象难处理**：深拷贝还是浅拷贝？
- **依赖具体类**：必须知道对象的确切类型

原型模式的解法：**让对象自己负责克隆自己，因为它最清楚自己有哪些字段**。

## 模式结构

![原型模式结构](https://refactoringguru.cn/images/patterns/diagrams/prototype/structure.png?id=088102c5e9785ff45debbbce86f4df81)

| 角色 | 职责 | 类比 |
|------|------|------|
| **Prototype（原型接口）** | 声明克隆方法 | "可复制"能力 |
| **ConcretePrototype（具体原型）** | 实现克隆逻辑 | 文件、文件夹 |
| **Client（客户端）** | 调用克隆方法获取副本 | 使用复制功能的人 |

## 动手实现：文件系统复制

用文件系统的复制功能来演示原型模式。文件和文件夹都可以被克隆。

### 第一步：定义原型接口

=== "📄 inode.go: 原型接口"

    ```go
    package main

    // Inode 文件系统节点接口（原型）
    type Inode interface {
        Print(indent string)
        Clone() Inode
    }
    ```

### 第二步：实现具体原型

=== "📄 file.go: 文件（具体原型）"

    ```go
    package main

    import "fmt"

    // File 文件节点
    type File struct {
        Name string
    }

    func (f *File) Print(indent string) {
        fmt.Println(indent + "📄 " + f.Name)
    }

    // Clone 克隆文件
    func (f *File) Clone() Inode {
        return &File{Name: f.Name + "_copy"}
    }
    ```

=== "📄 folder.go: 文件夹（具体原型）"

    ```go
    package main

    import "fmt"

    // Folder 文件夹节点
    type Folder struct {
        Name     string
        Children []Inode
    }

    func (f *Folder) Print(indent string) {
        fmt.Println(indent + "📁 " + f.Name)
        for _, child := range f.Children {
            child.Print(indent + "  ")
        }
    }

    // Clone 深度克隆文件夹及其所有子节点
    func (f *Folder) Clone() Inode {
        cloneFolder := &Folder{Name: f.Name + "_copy"}
        
        // 递归克隆所有子节点
        for _, child := range f.Children {
            cloneFolder.Children = append(cloneFolder.Children, child.Clone())
        }
        
        return cloneFolder
    }
    ```

### 第三步：使用原型模式

=== "📄 main.go: 客户端代码"

    ```go
    package main

    import "fmt"

    func main() {
        // 创建原始文件结构
        file1 := &File{Name: "报告.docx"}
        file2 := &File{Name: "数据.xlsx"}
        file3 := &File{Name: "图片.png"}

        folder1 := &Folder{
            Name:     "文档",
            Children: []Inode{file1},
        }

        folder2 := &Folder{
            Name:     "项目",
            Children: []Inode{folder1, file2, file3},
        }

        fmt.Println("=== 原始文件夹结构 ===")
        folder2.Print("")

        // 克隆整个文件夹
        clonedFolder := folder2.Clone()

        fmt.Println("\n=== 克隆后的文件夹结构 ===")
        clonedFolder.Print("")
    }
    ```

=== "📄 output.txt: 执行结果"

    ```
    === 原始文件夹结构 ===
    📁 项目
      📁 文档
        📄 报告.docx
      📄 数据.xlsx
      📄 图片.png

    === 克隆后的文件夹结构 ===
    📁 项目_copy
      📁 文档_copy
        📄 报告.docx_copy
      📄 数据.xlsx_copy
      📄 图片.png_copy
    ```

## 深拷贝 vs 浅拷贝

这是原型模式中最重要的概念：

| 类型 | 说明 | 后果 |
|------|------|------|
| **浅拷贝** | 只复制对象本身，嵌套对象共享引用 | 修改副本会影响原对象 |
| **深拷贝** | 递归复制所有嵌套对象 | 完全独立，但开销更大 |

```go
// 浅拷贝示例
func (f *Folder) ShallowClone() *Folder {
    return &Folder{
        Name:     f.Name,
        Children: f.Children,  // 共享同一个切片！
    }
}

// 深拷贝示例（推荐）
func (f *Folder) DeepClone() *Folder {
    clone := &Folder{Name: f.Name}
    for _, child := range f.Children {
        clone.Children = append(clone.Children, child.Clone())
    }
    return clone
}
```

## 什么时候该用原型模式？

| 场景 | 说明 |
|------|------|
| **对象创建成本高** | 如需要读取配置、连接数据库初始化 |
| **需要保存对象状态** | 如游戏存档、文档历史版本 |
| **不想依赖具体类** | 只知道接口，不知道具体类型 |
| **需要复制复杂对象** | 如包含嵌套结构的组合对象 |

**常见应用**：

- **图形编辑器**：复制图形对象
- **游戏开发**：克隆敌人、道具
- **缓存系统**：缓存对象模板，需要时克隆
- **配置对象**：复制默认配置后修改

## 优缺点分析

| ✅ 优点 | ❌ 缺点 |
|---------|---------|
| **不依赖具体类**：通过接口复制对象 | **循环引用麻烦**：A 引用 B，B 引用 A |
| **避免重复初始化**：克隆比新建快 | **深拷贝复杂**：嵌套层级深时实现复杂 |
| **动态配置对象**：克隆后再修改 | |

## 与其他模式的关系

| 模式组合 | 说明 |
|----------|------|
| **原型 + 命令** | 保存命令执行前的对象状态 |
| **原型 vs 备忘录** | 备忘录保存状态，原型复制整个对象 |
| **原型 + 工厂** | 工厂内部可以用原型来创建对象 |

> **一句话总结**：原型模式就像细胞分裂——从一个细胞复制出完全相同的新细胞，比重新长一个细胞快多了。
