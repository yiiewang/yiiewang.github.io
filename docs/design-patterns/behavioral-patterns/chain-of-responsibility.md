---
title: 责任链模式
date: 2023-02-25 17:18:47
category:
  - 设计模式
tag:
  - 行为模式
---

# 责任链模式：让请求找到合适的"接盘侠"

想象一下这个场景：你去医院看病，需要依次经过 **挂号→看诊→取药→缴费** 四个环节。每个环节都会问一句："这事儿我能办吗？"——能办就处理，不能办就交给下一个。这就是责任链模式的核心思想。

**责任链模式** 允许你将请求沿着处理者链传递。每个处理者收到请求后，要么自己处理，要么传给下一个。就像击鼓传花，直到有人"接盘"为止。

## 为什么需要责任链？

假设你正在开发一个在线订单系统，订单需要经过多重验证：

```
用户身份验证 → 权限检查 → 数据校验 → 缓存检查 → 实际处理
```

如果把这些逻辑写在一起，代码会变成一团乱麻：

```go
// ❌ 糟糕的写法：所有校验逻辑堆在一起
func processOrder(order Order) error {
    // 身份验证
    if !validateUser(order.UserID) {
        return errors.New("用户未登录")
    }
    // 权限检查
    if !checkPermission(order.UserID) {
        return errors.New("无权限")
    }
    // 数据校验
    if !validateData(order) {
        return errors.New("数据无效")
    }
    // ... 还有更多检查
    return nil
}
```

这样写的问题很明显：

- **难以扩展**：想加个新检查？得改这个臃肿的函数
- **难以复用**：另一个地方只需要身份验证？抱歉，得复制代码
- **难以测试**：想单独测试某个检查？做不到

责任链模式的解法是： **把每个检查封装成独立的"处理者"，然后像串珠子一样串起来** 。

## 模式结构

![责任链模式结构](https://refactoringguru.cn/images/patterns/diagrams/chain-of-responsibility/structure.png)

责任链由四个核心角色组成：

| 角色 | 职责 | 类比 |
|------|------|------|
| **Handler（处理者接口）** | 定义处理请求的标准接口 | 医院各科室的统一服务规范 |
| **BaseHandler（基础处理者）** | 存储下一个处理者的引用，实现链式传递 | 导诊台，知道下一站在哪 |
| **ConcreteHandler（具体处理者）** | 实现具体的处理逻辑 | 挂号处、诊室、药房、收费处 |
| **Client（客户端）** | 组装责任链，发起请求 | 患者，从第一个窗口开始办事 |

## 动手实现：医院就诊系统

用一个完整的医院就诊系统来演示责任链模式。患者就诊需要依次经过： **前台挂号 → 医生诊断 → 药房取药 → 收银结账** 。

### 第一步：定义处理者接口

首先定义一个统一的接口，所有科室都要实现这个接口：

=== "📄 department.go: 处理者接口"

    ```go
    package main

    // Department 定义了所有科室的统一接口
    type Department interface {
        // Execute 处理患者请求
        Execute(patient *Patient)
        // SetNext 设置下一个处理者
        SetNext(department Department)
    }
    ```

### 第二步：定义患者结构

患者是在责任链中流转的"请求对象"：

=== "📄 patient.go: 患者结构"

    ```go
    package main

    // Patient 代表一个患者（请求对象）
    type Patient struct {
        Name              string // 患者姓名
        RegistrationDone  bool   // 是否已挂号
        DoctorCheckUpDone bool   // 是否已看诊
        MedicineDone      bool   // 是否已取药
        PaymentDone       bool   // 是否已缴费
    }
    ```

### 第三步：实现各个科室（具体处理者）

每个科室只关心自己的职责，处理完后交给下一个：

=== "📄 reception.go: 前台挂号"

    ```go
    package main

    import "fmt"

    // Reception 前台，负责患者挂号
    type Reception struct {
        next Department
    }

    func (r *Reception) Execute(p *Patient) {
        if p.RegistrationDone {
            fmt.Println("✓ 患者已挂号，跳过此步骤")
        } else {
            fmt.Println("→ 前台：正在为患者办理挂号...")
            p.RegistrationDone = true
        }
        // 传递给下一个处理者
        r.next.Execute(p)
    }

    func (r *Reception) SetNext(next Department) {
        r.next = next
    }
    ```

=== "📄 doctor.go: 医生诊断"

    ```go
    package main

    import "fmt"

    // Doctor 医生，负责诊断患者
    type Doctor struct {
        next Department
    }

    func (d *Doctor) Execute(p *Patient) {
        if p.DoctorCheckUpDone {
            fmt.Println("✓ 患者已完成诊断，跳过此步骤")
        } else {
            fmt.Println("→ 医生：正在为患者进行诊断...")
            p.DoctorCheckUpDone = true
        }
        d.next.Execute(p)
    }

    func (d *Doctor) SetNext(next Department) {
        d.next = next
    }
    ```

=== "📄 medical.go: 药房取药"

    ```go
    package main

    import "fmt"

    // Medical 药房，负责发放药品
    type Medical struct {
        next Department
    }

    func (m *Medical) Execute(p *Patient) {
        if p.MedicineDone {
            fmt.Println("✓ 患者已取药，跳过此步骤")
        } else {
            fmt.Println("→ 药房：正在为患者配药...")
            p.MedicineDone = true
        }
        m.next.Execute(p)
    }

    func (m *Medical) SetNext(next Department) {
        m.next = next
    }
    ```

=== "📄 cashier.go: 收银结账"

    ```go
    package main

    import "fmt"

    // Cashier 收银台，负责收费（链条末端）
    type Cashier struct {
        next Department
    }

    func (c *Cashier) Execute(p *Patient) {
        if p.PaymentDone {
            fmt.Println("✓ 患者已缴费，就诊完成！")
        } else {
            fmt.Println("→ 收银台：正在为患者办理缴费...")
            p.PaymentDone = true
            fmt.Println("✓ 就诊流程全部完成！")
        }
        // 链条末端，不再传递
    }

    func (c *Cashier) SetNext(next Department) {
        c.next = next
    }
    ```

### 第四步：组装责任链并运行

=== "📄 main.go: 客户端代码"

    ```go
    package main

    func main() {
        // 1. 创建各个处理者
        cashier := &Cashier{}      // 链条末端
        medical := &Medical{}
        doctor := &Doctor{}
        reception := &Reception{}  // 链条起点

        // 2. 组装责任链：前台 → 医生 → 药房 → 收银
        reception.SetNext(doctor)
        doctor.SetNext(medical)
        medical.SetNext(cashier)

        // 3. 创建患者并开始就诊
        patient := &Patient{Name: "张三"}
        fmt.Println("=== 患者", patient.Name, "开始就诊 ===")
        reception.Execute(patient)
    }
    ```

=== "📄 output.txt: 执行结果"

    ```
    === 患者 张三 开始就诊 ===
    → 前台：正在为患者办理挂号...
    → 医生：正在为患者进行诊断...
    → 药房：正在为患者配药...
    → 收银台：正在为患者办理缴费...
    ✓ 就诊流程全部完成！
    ```

## 什么时候该用责任链？

责任链模式适用于以下场景：

| 场景 | 说明 |
|------|------|
| **请求处理顺序不固定** | 需要动态调整处理流程时 |
| **多个对象都可能处理请求** | 但具体由谁处理在运行时才能确定 |
| **需要解耦发送者和接收者** | 发送者不需要知道谁会处理请求 |

**常见应用场景** ：

- **Web 中间件**：身份验证 → 日志记录 → 限流 → 实际处理
- **审批流程**：员工 → 主管 → 经理 → 总监
- **异常处理**：按优先级尝试不同的处理策略
- **过滤器链**：数据清洗 → 格式转换 → 校验

## 优缺点分析

| ✅ 优点 | ❌ 缺点 |
|---------|---------|
| **灵活扩展**：新增处理者无需修改现有代码 | **请求可能无人处理**：到达链尾仍未被处理 |
| **职责单一**：每个处理者只关心自己的事 | **调试困难**：请求流转路径不直观 |
| **动态组合**：可在运行时调整链的结构 | **性能开销**：链过长时会影响性能 |

## 与其他模式的关系

责任链经常和其他模式配合使用：

- **责任链 + 组合模式**：在树形结构中逐层传递请求
- **责任链 + 命令模式**：将请求封装成命令对象，沿链传递
- **责任链 vs 装饰器**：装饰器是"增强"，责任链是"选择处理"

> **一句话总结**：责任链模式就像接力赛——每个选手跑完自己的路程，把接力棒交给下一个人，直到冲过终点线。
