---
title: 访问者模式
date: 2023-02-25 17:18:47
category:
  - 设计模式
tag:
  - 行为模式
---

# 访问者模式：给对象"外挂"新能力

想象你是一个博物馆的管理员，馆里有各种展品：画作、雕塑、古董。有一天，专家团来了：艺术评论家想给画作写评语，历史学家想研究古董年代。你不可能为每种需求都改造展品，但可以让专家们"访问"展品，各取所需。

**访问者模式** 让你能在不修改现有类的情况下，给这些类添加新的操作。就像给对象"装外挂"一样。

## 为什么需要访问者模式？

假设你维护一个图形库，有多种形状：

```go
// ❌ 糟糕的写法：每次新增功能都要改所有形状类
type Shape interface {
    GetArea() float64        // 原有功能
    GetPerimeter() float64   // 新增功能 1
    ToSVG() string           // 新增功能 2
    ToJSON() string          // 新增功能 3
    // 每加一个功能，所有实现类都要改...
}
```

这样写的问题：

- **违反开闭原则**：每次新增功能都要修改所有形状类
- **职责不单一**：形状类承担了太多职责
- **协作困难**：如果你不是库的维护者，无法添加新功能

访问者模式的解法： **把新功能封装成独立的"访问者"类，让它去访问各种形状** 。

## 模式结构

![访问者模式结构](https://refactoringguru.cn/images/patterns/diagrams/visitor/structure-zh.png)

| 角色 | 职责 | 类比 |
|------|------|------|
| **Visitor（访问者接口）** | 声明访问各种元素的方法 | 专家的"考察能力" |
| **ConcreteVisitor（具体访问者）** | 实现对各种元素的具体操作 | 艺术评论家、历史学家 |
| **Element（元素接口）** | 声明接受访问者的方法 | 展品的"可访问性" |
| **ConcreteElement（具体元素）** | 实现接受方法，调用访问者的对应方法 | 画作、雕塑、古董 |

## 动手实现：形状计算器

用形状库来演示访问者模式。我们有圆形、正方形、矩形，需要支持计算面积和导出 JSON。

### 第一步：定义访问者接口

=== "📄 visitor.go: 访问者接口"

    ```go
    package main

    // Visitor 访问者接口，为每种形状定义访问方法
    type Visitor interface {
        VisitCircle(c *Circle)
        VisitSquare(s *Square)
        VisitRectangle(r *Rectangle)
    }
    ```

!!! note "为什么不用一个 Visit(Shape) 方法？"
    Go 语言不支持方法重载，所以需要为每种形状单独定义方法。这样访问者可以针对不同形状执行不同逻辑。

### 第二步：定义元素接口和具体元素

=== "📄 shape.go: 元素接口"

    ```go
    package main

    // Shape 形状接口，所有形状都可以被访问
    type Shape interface {
        Accept(visitor Visitor)
        GetType() string
    }
    ```

=== "📄 circle.go: 圆形"

    ```go
    package main

    // Circle 圆形
    type Circle struct {
        Radius float64
    }

    func (c *Circle) Accept(v Visitor) {
        v.VisitCircle(c)  // 关键：调用访问者的对应方法
    }

    func (c *Circle) GetType() string {
        return "Circle"
    }
    ```

=== "📄 square.go: 正方形"

    ```go
    package main

    // Square 正方形
    type Square struct {
        Side float64
    }

    func (s *Square) Accept(v Visitor) {
        v.VisitSquare(s)
    }

    func (s *Square) GetType() string {
        return "Square"
    }
    ```

=== "📄 rectangle.go: 矩形"

    ```go
    package main

    // Rectangle 矩形
    type Rectangle struct {
        Width  float64
        Height float64
    }

    func (r *Rectangle) Accept(v Visitor) {
        v.VisitRectangle(r)
    }

    func (r *Rectangle) GetType() string {
        return "Rectangle"
    }
    ```

### 第三步：实现具体访问者

=== "📄 area_calculator.go: 面积计算器"

    ```go
    package main

    import (
        "fmt"
        "math"
    )

    // AreaCalculator 计算面积的访问者
    type AreaCalculator struct{}

    func (a *AreaCalculator) VisitCircle(c *Circle) {
        area := math.Pi * c.Radius * c.Radius
        fmt.Printf("⭕ 圆形面积: %.2f\n", area)
    }

    func (a *AreaCalculator) VisitSquare(s *Square) {
        area := s.Side * s.Side
        fmt.Printf("⬜ 正方形面积: %.2f\n", area)
    }

    func (a *AreaCalculator) VisitRectangle(r *Rectangle) {
        area := r.Width * r.Height
        fmt.Printf("📐 矩形面积: %.2f\n", area)
    }
    ```

=== "📄 json_exporter.go: JSON 导出器"

    ```go
    package main

    import "fmt"

    // JSONExporter 导出 JSON 的访问者
    type JSONExporter struct{}

    func (j *JSONExporter) VisitCircle(c *Circle) {
        fmt.Printf(`{"type": "circle", "radius": %.2f}%s`, c.Radius, "\n")
    }

    func (j *JSONExporter) VisitSquare(s *Square) {
        fmt.Printf(`{"type": "square", "side": %.2f}%s`, s.Side, "\n")
    }

    func (j *JSONExporter) VisitRectangle(r *Rectangle) {
        fmt.Printf(`{"type": "rectangle", "width": %.2f, "height": %.2f}%s`, 
            r.Width, r.Height, "\n")
    }
    ```

### 第四步：使用访问者

=== "📄 main.go: 客户端代码"

    ```go
    package main

    import "fmt"

    func main() {
        // 创建形状
        shapes := []Shape{
            &Circle{Radius: 5},
            &Square{Side: 4},
            &Rectangle{Width: 3, Height: 6},
        }

        // 使用面积计算器访问所有形状
        fmt.Println("=== 计算面积 ===")
        areaCalc := &AreaCalculator{}
        for _, shape := range shapes {
            shape.Accept(areaCalc)
        }

        // 使用 JSON 导出器访问所有形状
        fmt.Println("\n=== 导出 JSON ===")
        jsonExp := &JSONExporter{}
        for _, shape := range shapes {
            shape.Accept(jsonExp)
        }
    }
    ```

=== "📄 output.txt: 执行结果"

    ```
    === 计算面积 ===
    ⭕ 圆形面积: 78.54
    ⬜ 正方形面积: 16.00
    📐 矩形面积: 18.00

    === 导出 JSON ===
    {"type": "circle", "radius": 5.00}
    {"type": "square", "side": 4.00}
    {"type": "rectangle", "width": 3.00, "height": 6.00}
    ```

## 双重分派：访问者模式的核心技巧

访问者模式利用了"双重分派"技术：

1. **第一次分派**：调用 `shape.Accept(visitor)` 时，根据 shape 的具体类型分派
2. **第二次分派**：在 Accept 中调用 `visitor.VisitXxx(this)` 时，根据 visitor 的具体类型分派

```go
// 第一次分派：确定是哪种 Shape
circle.Accept(visitor)

// Accept 内部 - 第二次分派：确定是哪种 Visitor
func (c *Circle) Accept(v Visitor) {
    v.VisitCircle(c)  // v 可能是 AreaCalculator 或 JSONExporter
}
```

## 什么时候该用访问者模式？

| 场景 | 说明 |
|------|------|
| **需要对复杂对象结构执行操作** | 如遍历组合模式的树结构 |
| **需要分离算法和数据结构** | 数据结构稳定，但经常需要新增操作 |
| **不同类型对象需要不同处理** | 且不想在每个类中都实现这些逻辑 |

**常见应用** ：

- **编译器**：对 AST（抽象语法树）执行类型检查、代码生成
- **XML/JSON 解析**：对不同节点类型执行不同操作
- **报表系统**：对不同数据对象生成不同格式的报表
- **序列化/反序列化**：把对象转换成不同格式

## 优缺点分析

| ✅ 优点 | ❌ 缺点 |
|---------|---------|
| **开闭原则**：新增访问者无需修改元素类 | **新增元素困难**：每加一种元素，所有访问者都要改 |
| **单一职责**：相关操作集中在访问者中 | **可能破坏封装**：访问者需要访问元素的内部数据 |
| **收集信息**：遍历时可以积累状态 | **元素结构需稳定**：频繁新增元素类型则不适用 |

## 与其他模式的关系

| 模式组合 | 说明 |
|----------|------|
| **访问者 + 组合** | 遍历树形结构，对每个节点执行操作 |
| **访问者 + 迭代器** | 迭代器遍历集合，访问者处理元素 |
| **访问者 vs 命令** | 访问者是命令的加强版，支持对不同类型对象的操作 |

> **一句话总结**：访问者模式就像"上门服务"——你不用改造房子，只需开门迎客，让不同的服务人员（访问者）进来干不同的活。
