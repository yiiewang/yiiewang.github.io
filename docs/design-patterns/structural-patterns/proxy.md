---
title: 代理模式
date: 2023-02-25 17:18:47
category:
  - 设计模式
tag:
  - 结构型模式
---

# 🛡️ 代理模式：对象的"替身演员"

> 代理模式是一种结构型设计模式，让你能够提供对象的替代品或占位符。代理控制着对原对象的访问，并允许在请求前后进行一些处理。

## 从生活场景说起

想象 Nginx 服务器和你的应用服务器的关系：

- 🌐 用户请求 → Nginx（代理）→ 应用服务器
- Nginx 可以做什么？
  - 🔒 访问控制（限流、认证）
  - 📦 缓存响应
  - 📊 记录日志

用户以为在和应用服务器对话，实际上是 Nginx 在"代理"处理。

## 代理的类型

| 类型 | 说明 | 示例 |
|------|------|------|
| **虚拟代理** | 延迟初始化重量级对象 | 图片懒加载 |
| **保护代理** | 访问控制 | 权限检查 |
| **远程代理** | 代理远程服务 | RPC 调用 |
| **缓存代理** | 缓存请求结果 | Nginx 缓存 |
| **日志代理** | 记录请求历史 | 审计日志 |

## 完整实现

=== "server.go: 服务接口"

    ```go
    package main

    type server interface {
        handleRequest(string, string) (int, string)
    }
    ```

=== "application.go: 真实服务"

    ```go
    package main

    type Application struct{}

    func (a *Application) handleRequest(url, method string) (int, string) {
        if url == "/app/status" && method == "GET" {
            return 200, "Ok"
        }
        if url == "/create/user" && method == "POST" {
            return 201, "User Created"
        }
        return 404, "Not Found"
    }
    ```

=== "nginx.go: 代理"

    ```go
    package main

    type Nginx struct {
        application       *Application
        maxAllowedRequest int
        rateLimiter       map[string]int
    }

    func newNginxServer() *Nginx {
        return &Nginx{
            application:       &Application{},
            maxAllowedRequest: 2,
            rateLimiter:       make(map[string]int),
        }
    }

    func (n *Nginx) handleRequest(url, method string) (int, string) {
        // 前置处理：限流检查
        if !n.checkRateLimiting(url) {
            return 403, "Not Allowed"
        }
        // 委托给真实服务
        return n.application.handleRequest(url, method)
    }

    func (n *Nginx) checkRateLimiting(url string) bool {
        n.rateLimiter[url]++
        return n.rateLimiter[url] <= n.maxAllowedRequest
    }
    ```

=== "main.go: 客户端代码"

    ```go
    package main

    import "fmt"

    func main() {
        nginx := newNginxServer()
        
        // 前两次请求成功
        code, body := nginx.handleRequest("/app/status", "GET")
        fmt.Printf("Code: %d, Body: %s\n", code, body)
        
        code, body = nginx.handleRequest("/app/status", "GET")
        fmt.Printf("Code: %d, Body: %s\n", code, body)
        
        // 第三次被限流
        code, body = nginx.handleRequest("/app/status", "GET")
        fmt.Printf("Code: %d, Body: %s\n", code, body)
    }
    ```

=== "output.txt: 执行结果"

    ```
    Code: 200, Body: Ok
    Code: 200, Body: Ok
    Code: 403, Body: Not Allowed
    ```

## 优缺点

| ✅ 优点 | ❌ 缺点 |
|---------|---------|
| 控制服务对象访问 | 代码复杂度增加 |
| 管理对象生命周期 | 响应可能延迟 |
| 开闭原则：可添加新代理 | |

## 与其他模式的关系

| 模式 | 区别 |
|------|------|
| **适配器** | 适配器改变接口，代理保持相同接口 |
| **装饰器** | 装饰器增强功能，代理控制访问 |
| **外观** | 外观定义新接口，代理与服务接口相同 |

---

!!! tip "一句话总结"
    **代理模式就是"门卫"**：你要进大楼（访问对象），先经过门卫（代理）检查，门卫可以放行、拒绝，或者记录你的访问。
