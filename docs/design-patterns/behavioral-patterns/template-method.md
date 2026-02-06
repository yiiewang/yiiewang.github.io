---
title: 模版方法模式
date: 2023-5-22 16:40:17
category:
  - 设计模式
tag:
  - 行为模式
---

# 模板方法模式：定好骨架，填充血肉

做菜有固定流程：备菜 → 热锅 → 烹饪 → 装盘。不管做什么菜，这个"骨架"不变，变的只是具体怎么备菜、怎么烹饪。模板方法模式就是这个思路—— **在父类中定义算法骨架，让子类填充具体步骤** 。

**模板方法模式** 在超类中定义一个算法的框架，允许子类在不修改结构的情况下重写特定步骤。

## 为什么需要模板方法？

假设你在开发一个发送验证码的功能，支持短信和邮件两种方式：

```go
// ❌ 糟糕的写法：两套几乎一样的代码
func sendSmsOTP() {
    otp := generateOTP()      // 相同
    saveToCache(otp)          // 相同
    msg := "短信验证码: " + otp  // 不同
    sendSms(msg)              // 不同
}

func sendEmailOTP() {
    otp := generateOTP()      // 相同
    saveToCache(otp)          // 相同
    msg := "邮件验证码: " + otp  // 不同
    sendEmail(msg)            // 不同
}
```

这样写的问题：

- **代码重复**：相同的步骤写了两遍
- **修改困难**：想改生成逻辑？两处都要改
- **容易出错**：复制粘贴时可能遗漏细节

模板方法的解法： **把相同的步骤抽到父类，不同的步骤让子类实现** 。

## 模式结构

![模板方法结构](https://refactoringguru.cn/images/patterns/diagrams/template-method/structure.png)

| 角色 | 职责 | 类比 |
|------|------|------|
| **AbstractClass（抽象类）** | 定义算法骨架，调用各个步骤 | 菜谱模板 |
| **ConcreteClass（具体类）** | 实现具体步骤 | 红烧肉/糖醋排骨 |

## 动手实现：OTP 验证码发送系统

用验证码发送系统来演示模板方法模式。

### 第一步：定义模板接口和基础结构

=== "📄 otp.go: 模板方法"

    ```go
    package main

    // OTPSender 验证码发送器接口
    type OTPSender interface {
        GenerateOTP(length int) string  // 生成验证码
        SaveToCache(otp string)          // 保存到缓存
        GetMessage(otp string) string    // 构造消息内容
        SendNotification(msg string) error // 发送通知
    }

    // OTP 模板，定义算法骨架
    type OTP struct {
        sender OTPSender
    }

    // Send 模板方法：定义验证码发送的完整流程
    func (o *OTP) Send(length int) error {
        // 1. 生成验证码
        otp := o.sender.GenerateOTP(length)
        
        // 2. 保存到缓存
        o.sender.SaveToCache(otp)
        
        // 3. 构造消息
        message := o.sender.GetMessage(otp)
        
        // 4. 发送通知
        return o.sender.SendNotification(message)
    }
    ```

### 第二步：实现具体发送方式

=== "📄 sms_otp.go: 短信验证码"

    ```go
    package main

    import (
        "fmt"
        "math/rand"
    )

    // SmsOTP 短信验证码
    type SmsOTP struct{}

    func (s *SmsOTP) GenerateOTP(length int) string {
        otp := fmt.Sprintf("%04d", rand.Intn(10000))
        fmt.Printf("📱 [短信] 生成验证码: %s\n", otp)
        return otp
    }

    func (s *SmsOTP) SaveToCache(otp string) {
        fmt.Printf("💾 [短信] 保存验证码到缓存: %s\n", otp)
    }

    func (s *SmsOTP) GetMessage(otp string) string {
        return fmt.Sprintf("【XX公司】您的验证码是 %s，5分钟内有效。", otp)
    }

    func (s *SmsOTP) SendNotification(msg string) error {
        fmt.Printf("📤 [短信] 发送内容: %s\n", msg)
        return nil
    }
    ```

=== "📄 email_otp.go: 邮件验证码"

    ```go
    package main

    import (
        "fmt"
        "math/rand"
    )

    // EmailOTP 邮件验证码
    type EmailOTP struct{}

    func (e *EmailOTP) GenerateOTP(length int) string {
        otp := fmt.Sprintf("%04d", rand.Intn(10000))
        fmt.Printf("📧 [邮件] 生成验证码: %s\n", otp)
        return otp
    }

    func (e *EmailOTP) SaveToCache(otp string) {
        fmt.Printf("💾 [邮件] 保存验证码到缓存: %s\n", otp)
    }

    func (e *EmailOTP) GetMessage(otp string) string {
        return fmt.Sprintf("<h1>您的验证码</h1><p>验证码: <strong>%s</strong></p>", otp)
    }

    func (e *EmailOTP) SendNotification(msg string) error {
        fmt.Printf("📤 [邮件] 发送内容: %s\n", msg)
        return nil
    }
    ```

### 第三步：使用模板

=== "📄 main.go: 客户端代码"

    ```go
    package main

    import "fmt"

    func main() {
        fmt.Println("=== 发送短信验证码 ===")
        smsOTP := &SmsOTP{}
        otpSms := OTP{sender: smsOTP}
        otpSms.Send(4)

        fmt.Println("\n=== 发送邮件验证码 ===")
        emailOTP := &EmailOTP{}
        otpEmail := OTP{sender: emailOTP}
        otpEmail.Send(4)
    }
    ```

=== "📄 output.txt: 执行结果"

    ```
    === 发送短信验证码 ===
    📱 [短信] 生成验证码: 1234
    💾 [短信] 保存验证码到缓存: 1234
    📤 [短信] 发送内容: 【XX公司】您的验证码是 1234，5分钟内有效。

    === 发送邮件验证码 ===
    📧 [邮件] 生成验证码: 5678
    💾 [邮件] 保存验证码到缓存: 5678
    📤 [邮件] 发送内容: <h1>您的验证码</h1><p>验证码: <strong>5678</strong></p>
    ```

## 钩子方法：可选的扩展点

模板方法可以提供"钩子"，让子类选择性地扩展某些步骤：

```go
type OTPSender interface {
    // ... 必须实现的方法

    // 钩子方法：子类可以选择重写
    BeforeSend() bool  // 发送前检查，返回 false 取消发送
    AfterSend()        // 发送后回调
}

func (o *OTP) Send(length int) error {
    // 调用钩子
    if !o.sender.BeforeSend() {
        return fmt.Errorf("发送被取消")
    }
    
    // ... 正常流程
    
    o.sender.AfterSend()  // 完成后回调
    return nil
}
```

## 什么时候该用模板方法？

| 场景 | 说明 |
|------|------|
| **算法骨架固定** | 整体流程不变，只有某些步骤不同 |
| **消除重复代码** | 多个类有相似的代码，可以提取公共部分 |
| **控制扩展点** | 只允许子类修改特定步骤，防止破坏整体流程 |

**常见应用** ：

- **测试框架**：setUp → test → tearDown
- **游戏 AI**：感知 → 决策 → 行动
- **数据处理**：读取 → 转换 → 保存
- **Web 请求处理**：认证 → 授权 → 执行 → 响应

## 优缺点分析

| ✅ 优点 | ❌ 缺点 |
|---------|---------|
| **消除重复**：公共代码只写一次 | **框架限制**：子类必须遵循父类定义的骨架 |
| **易于扩展**：只需实现变化的部分 | **里氏替换**：子类可能通过重写破坏父类行为 |
| **反向控制**：父类调用子类方法（好莱坞原则） | **步骤越多越复杂**：维护成本增加 |

## 模板方法 vs 策略模式

| 特性 | 模板方法 | 策略模式 |
|------|---------|---------|
| **实现方式** | 继承（is-a） | 组合（has-a） |
| **灵活性** | 编译时确定 | 运行时可切换 |
| **控制粒度** | 控制整个算法骨架 | 只控制某一个算法 |
| **类层次** | 需要继承体系 | 扁平化，无需继承 |

> **一句话总结**：模板方法就像填空题——题目（骨架）是固定的，你只需要填写答案（具体步骤）。
