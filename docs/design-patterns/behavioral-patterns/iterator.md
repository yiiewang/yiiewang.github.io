---
title: 迭代器模式
date: 2023-02-25 17:18:47
category:
  - 设计模式
tag:
  - 行为模式
---

# 迭代器模式：不拆箱也能数清里面有什么

想象一下：你有一个装满玩具的大箱子，想知道里面有什么，但你不想把所有玩具都倒出来。于是你发明了一个"探测器"，可以一个个地取出玩具看看，看完再放回去。这个"探测器"就是迭代器。

**迭代器模式** 让你能在不暴露集合内部结构的情况下，遍历集合中的所有元素。无论是数组、链表、树还是图，都能用统一的方式访问。

## 为什么需要迭代器？

假设你有一个用户管理系统，需要遍历所有用户：

```go
// ❌ 糟糕的写法：直接暴露内部结构
type UserManager struct {
    users []*User  // 暴露了内部实现（切片）
}

// 客户端代码必须知道内部是切片
for i := 0; i < len(manager.users); i++ {
    fmt.Println(manager.users[i].Name)
}
```

这样写的问题：

- **暴露内部结构**：客户端知道你用的是切片，以后想换成链表就麻烦了
- **重复遍历代码**：每次遍历都要写 for 循环
- **无法并行遍历**：两个地方同时遍历会相互干扰

迭代器模式的解法： **提供一个统一的"遍历器"，隐藏集合的内部实现** 。

## 模式结构

![迭代器模式结构](images/image-1.png)

| 角色 | 职责 | 类比 |
|------|------|------|
| **Iterator（迭代器接口）** | 定义遍历元素的方法 | 遥控器的"下一个"按钮 |
| **ConcreteIterator（具体迭代器）** | 实现遍历算法，记住当前位置 | 电视频道切换器 |
| **Collection（集合接口）** | 声明创建迭代器的方法 | 电视机 |
| **ConcreteCollection（具体集合）** | 返回对应的迭代器实例 | 某品牌电视机 |

## 动手实现：用户列表遍历器

用一个用户管理系统来演示迭代器模式。

### 第一步：定义迭代器接口

=== "📄 iterator.go: 迭代器接口"

    ```go
    package main

    // Iterator 迭代器接口
    type Iterator interface {
        HasNext() bool    // 是否还有下一个元素
        Next() *User      // 获取下一个元素
    }
    ```

### 第二步：定义集合接口

=== "📄 collection.go: 集合接口"

    ```go
    package main

    // Collection 集合接口
    type Collection interface {
        CreateIterator() Iterator
    }
    ```

### 第三步：实现具体迭代器

=== "📄 user_iterator.go: 具体迭代器"

    ```go
    package main

    // UserIterator 用户集合的迭代器
    type UserIterator struct {
        index int
        users []*User
    }

    // HasNext 检查是否还有下一个用户
    func (u *UserIterator) HasNext() bool {
        return u.index < len(u.users)
    }

    // Next 获取下一个用户
    func (u *UserIterator) Next() *User {
        if u.HasNext() {
            user := u.users[u.index]
            u.index++
            return user
        }
        return nil
    }
    ```

### 第四步：实现具体集合

=== "📄 user_collection.go: 具体集合"

    ```go
    package main

    // UserCollection 用户集合
    type UserCollection struct {
        users []*User
    }

    // CreateIterator 创建迭代器
    func (u *UserCollection) CreateIterator() Iterator {
        return &UserIterator{
            users: u.users,
        }
    }

    // AddUser 添加用户
    func (u *UserCollection) AddUser(user *User) {
        u.users = append(u.users, user)
    }
    ```

=== "📄 user.go: 用户结构"

    ```go
    package main

    // User 用户信息
    type User struct {
        Name string
        Age  int
    }
    ```

### 第五步：使用迭代器遍历

=== "📄 main.go: 客户端代码"

    ```go
    package main

    import "fmt"

    func main() {
        // 创建用户集合
        collection := &UserCollection{}
        collection.AddUser(&User{Name: "张三", Age: 25})
        collection.AddUser(&User{Name: "李四", Age: 30})
        collection.AddUser(&User{Name: "王五", Age: 28})

        // 获取迭代器
        iterator := collection.CreateIterator()

        // 遍历所有用户
        fmt.Println("=== 用户列表 ===")
        for iterator.HasNext() {
            user := iterator.Next()
            fmt.Printf("姓名: %s, 年龄: %d\n", user.Name, user.Age)
        }
    }
    ```

=== "📄 output.txt: 执行结果"

    ```
    === 用户列表 ===
    姓名: 张三, 年龄: 25
    姓名: 李四, 年龄: 30
    姓名: 王五, 年龄: 28
    ```

## 迭代器的威力：支持多种遍历方式

同一个集合可以有多种迭代器：

```go
// 正向迭代器
type ForwardIterator struct { /* ... */ }

// 反向迭代器
type ReverseIterator struct { /* ... */ }

// 过滤迭代器（只遍历符合条件的元素）
type FilterIterator struct {
    inner     Iterator
    predicate func(*User) bool
}
```

## 什么时候该用迭代器？

| 场景 | 说明 |
|------|------|
| **隐藏集合内部结构** | 不想暴露底层是数组、链表还是树 |
| **统一遍历接口** | 让不同类型的集合有相同的遍历方式 |
| **支持多种遍历方式** | 正向、反向、过滤、跳跃遍历 |
| **并行遍历** | 多个迭代器可以独立遍历同一集合 |

**常见应用** ：

- **数据库游标**：逐行读取查询结果
- **文件系统遍历**：遍历目录树
- **DOM 遍历**：遍历 HTML 节点
- **Go 的 `range`**：其实就是语法糖版的迭代器

## 优缺点分析

| ✅ 优点 | ❌ 缺点 |
|---------|---------|
| **单一职责**：遍历逻辑独立于集合 | **简单集合不值得**：数组直接用 for 更简单 |
| **开闭原则**：新增集合类型无需修改客户端 | **可能效率较低**：比直接访问多一层间接调用 |
| **支持并行遍历**：每个迭代器有自己的状态 | |
| **可暂停遍历**：随时保存进度，稍后继续 | |

## 与其他模式的关系

| 模式组合 | 说明 |
|----------|------|
| **迭代器 + 组合** | 遍历树形结构 |
| **迭代器 + 工厂** | 让子集合返回不同类型的迭代器 |
| **迭代器 + 访问者** | 遍历时对每个元素执行特定操作 |
| **迭代器 + 备忘录** | 保存遍历进度，支持回退 |

> **一句话总结**：迭代器模式就像电视遥控器的"下一个频道"按钮——你不需要知道电视内部怎么存储频道，只需要按按钮就能切换。
