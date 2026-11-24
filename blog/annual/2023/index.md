# 2023 年度集锦

> 刷题年。25 篇里 10 篇是算法，剩下的分散在架构入门、Go 进阶和零星的 Java。

## 卷首语

2023 的骨架是二月：2 月 3 日到 2 月 20 日，十八天写了八篇算法——滑动窗口、二分查找、 双指针、BFS、DFS、递归回溯、动态规划、位运算，最后用一篇 [刷题总纲](<https://cloaks.cn/blog/posts/2023/02/15/index.md>)收口（这篇后来一直置顶在博客首页）。

那是最后一次集中刷题。之后三月转向架构：[Service Mesh](<https://cloaks.cn/blog/posts/2023/02/16-2/index.md>) → [Istio](<https://cloaks.cn/blog/posts/2023/02/16-1/index.md>) → [C10K](<https://cloaks.cn/blog/posts/2023/03/01/index.md>)， 六月补了云服务三篇。Go 则贯穿三月到四月：[反射](<https://cloaks.cn/blog/posts/2023/03/26-2/index.md>)、 [非侵入式架构](<https://cloaks.cn/blog/posts/2023/03/26-1/index.md>)、[模块发布](<https://cloaks.cn/blog/posts/2023/04/07/index.md>)、 [函数式编程](<https://cloaks.cn/blog/posts/2023/04/14/index.md>)。

## 主线一：算法十篇（二月）

按顺序读效果最好，前面是"形"，后面是"意"：

1. [滑动窗口](<https://cloaks.cn/blog/posts/2023/02/03/index.md>)
2. [二分查找](<https://cloaks.cn/blog/posts/2023/02/04/index.md>)
3. [双指针：从「背模板」到「懂不变量」](<https://cloaks.cn/blog/posts/2023/02/05/index.md>)
4. [广度优先搜索](<https://cloaks.cn/blog/posts/2023/02/13/index.md>)
5. [深度优先搜索](<https://cloaks.cn/blog/posts/2023/02/14-1/index.md>)
6. [递归与回溯](<https://cloaks.cn/blog/posts/2023/02/14-2/index.md>)
7. [动态规划](<https://cloaks.cn/blog/posts/2023/02/19/index.md>)
8. [位运算](<https://cloaks.cn/blog/posts/2023/02/20/index.md>)
9. [异或运算的应用场景](<https://cloaks.cn/blog/posts/2023/03/05/index.md>)
10. [刷题总结的正确打开方式](<https://cloaks.cn/blog/posts/2023/02/15/index.md>)（总纲，置顶）

> [!NOTE]
>
> **总纲里那句**
>
> 一道题只沉淀三个问题：**结构性质、不变量、边界**。 背模板解决不了"为什么指针敢淘汰候选"这类问题，不变量才行。

[双指针那一篇](<https://cloaks.cn/blog/posts/2023/02/05/index.md>)是整个系列里最有价值的——它把 "两端向中间""一快一慢"这种口诀，还原成了能推导的东西。

## 主线二：架构入门的两条腿

一条是服务网格：[Service Mesh 概念](<https://cloaks.cn/blog/posts/2023/02/16-2/index.md>) → [Istio 架构](<https://cloaks.cn/blog/posts/2023/02/16-1/index.md>)（Envoy 数据平面 + Pilot/Citadel/Galley 控制平面）。

另一条是并发与云：[C10K 问题](<https://cloaks.cn/blog/posts/2023/03/01/index.md>)（select / poll / epoll / kqueue）、 [SaaS、PaaS、IaaS](<https://cloaks.cn/blog/posts/2023/06/13/index.md>)、[如何设计一个应用解决方案](<https://cloaks.cn/blog/posts/2023/06/15/index.md>)、 [云服务的数据存储设计](<https://cloaks.cn/blog/posts/2023/06/21/index.md>)。

## 主线三：Go 的四篇

- [Go 反射（reflect）详解](<https://cloaks.cn/blog/posts/2023/03/26-2/index.md>) — Type / Value / Kind，顺手实现 BeanUtil
- [Go 语言非侵入式架构设计](<https://cloaks.cn/blog/posts/2023/03/26-1/index.md>) — 接口、组合、依赖注入做松耦合
- [Go 模块开发与发布指南](<https://cloaks.cn/blog/posts/2023/04/07/index.md>) — 版本控制与私有仓库配置
- [Go 函数式编程：装饰器与工厂模式](<https://cloaks.cn/blog/posts/2023/04/14/index.md>)

## 支线：Java、工控与一篇区块链

- **Java 两篇**：[日期格式化](<https://cloaks.cn/blog/posts/2023/01/03/index.md>)、[线程编程](<https://cloaks.cn/blog/posts/2023/02/27/index.md>)。这也是最后一次写 Java
- **安全**：[工控系统网络安全：纵深防御](<https://cloaks.cn/blog/posts/2023/01/26/index.md>)
- **区块链**：[去中心化应用（dApp）详解](<https://cloaks.cn/blog/posts/2023/02/28/index.md>)
- **投资**：[《有效资产管理》读书笔记](<https://cloaks.cn/blog/posts/2023/02/18/index.md>)

## 代表作五篇

1. [双指针：从「背模板」到「懂不变量」](<https://cloaks.cn/blog/posts/2023/02/05/index.md>) — 系列里唯一讲"为什么"的
2. [算法刷题问题总结（总纲）](<https://cloaks.cn/blog/posts/2023/02/15/index.md>) — 13 类题型的速查地图，长期置顶
3. [C10K 问题：单机一万并发](<https://cloaks.cn/blog/posts/2023/03/01/index.md>) — 把 I/O 多路复用的演进讲完整了
4. [Go 反射（reflect）详解](<https://cloaks.cn/blog/posts/2023/03/26-2/index.md>) — 反射这块写得比多数教程清楚
5. [Go 语言非侵入式架构设计](<https://cloaks.cn/blog/posts/2023/03/26-1/index.md>) — 组合优于继承，Go 版本

## 全年目录

### 1 月 · 2 篇

- [Java 日期格式化详解](<https://cloaks.cn/blog/posts/2023/01/03/index.md>) · 01-03 · 札记
- [工控系统网络安全：纵深防御策略](<https://cloaks.cn/blog/posts/2023/01/26/index.md>) · 01-26 · 札记

### 2 月 · 14 篇

- [滑动窗口算法详解](<https://cloaks.cn/blog/posts/2023/02/03/index.md>) · 02-03 · 算法与数据结构
- [二分查找算法详解](<https://cloaks.cn/blog/posts/2023/02/04/index.md>) · 02-04 · 算法与数据结构
- [双指针：从「背模板」到「懂不变量」](<https://cloaks.cn/blog/posts/2023/02/05/index.md>) · 02-05 · 算法与数据结构
- [广度优先搜索算法详解](<https://cloaks.cn/blog/posts/2023/02/13/index.md>) · 02-13 · 算法与数据结构
- [深度优先搜索算法详解](<https://cloaks.cn/blog/posts/2023/02/14-1/index.md>) · 02-14 · 算法与数据结构
- [递归与回溯算法详解](<https://cloaks.cn/blog/posts/2023/02/14-2/index.md>) · 02-14 · 算法与数据结构
- [pin: true](<https://cloaks.cn/blog/posts/2023/02/15/index.md>) · 02-15 · 算法与数据结构
- [Istio 服务网格架构详解](<https://cloaks.cn/blog/posts/2023/02/16-1/index.md>) · 02-16 · 系统与架构
- [Service Mesh 服务网格详解](<https://cloaks.cn/blog/posts/2023/02/16-2/index.md>) · 02-16 · 系统与架构
- [《有效资产管理》读书笔记](<https://cloaks.cn/blog/posts/2023/02/18/index.md>) · 02-18 · 投资与交易
- [动态规划算法详解](<https://cloaks.cn/blog/posts/2023/02/19/index.md>) · 02-19 · 算法与数据结构
- [位运算算法详解](<https://cloaks.cn/blog/posts/2023/02/20/index.md>) · 02-20 · 算法与数据结构
- [Java 线程编程详解](<https://cloaks.cn/blog/posts/2023/02/27/index.md>) · 02-27 · 札记
- [去中心化应用详解](<https://cloaks.cn/blog/posts/2023/02/28/index.md>) · 02-28 · 区块链

### 3 月 · 4 篇

- [C10K 问题：单机一万并发](<https://cloaks.cn/blog/posts/2023/03/01/index.md>) · 03-01 · 系统与架构
- [异或运算的应用场景](<https://cloaks.cn/blog/posts/2023/03/05/index.md>) · 03-05 · 算法与数据结构
- [Go 语言非侵入式架构设计](<https://cloaks.cn/blog/posts/2023/03/26-1/index.md>) · 03-26 · Go 语言
- [Go 反射（reflect）详解](<https://cloaks.cn/blog/posts/2023/03/26-2/index.md>) · 03-26 · Go 语言

### 4 月 · 2 篇

- [Go 模块开发与发布指南](<https://cloaks.cn/blog/posts/2023/04/07/index.md>) · 04-07 · Go 语言
- [Go 函数式编程：装饰器与工厂模式](<https://cloaks.cn/blog/posts/2023/04/14/index.md>) · 04-14 · Go 语言

### 6 月 · 3 篇

- [云服务模式详解：SaaS、PaaS、IaaS](<https://cloaks.cn/blog/posts/2023/06/13/index.md>) · 06-13 · 系统与架构
- [如何设计一个应用解决方案？](<https://cloaks.cn/blog/posts/2023/06/15/index.md>) · 06-15 · 系统与架构
- [云服务的数据存储设计](<https://cloaks.cn/blog/posts/2023/06/21/index.md>) · 06-21 · 系统与架构

## 回看：后来怎样了

1. 二月那批算法是最后一次集中刷题；2025 年只剩[堆排序](<https://cloaks.cn/blog/posts/2025/08/18/index.md>)和[时间轮](<https://cloaks.cn/blog/posts/2025/12/15/index.md>)两篇补漏，2026 年算法归零
2. 架构这条线没断，但方向换了：2023 写服务和云，2026 写[无锁队列](<https://cloaks.cn/blog/posts/2026/09/07/index.md>)和[xv6](<https://cloaks.cn/blog/posts/2026/09/17/index.md>)
3. Java 停在 2023 年 2 月，此后再没写过
4. 那篇《有效资产管理》读书笔记，是投资这条线最早的起点——它到 2026 年才长成完整的交易系统

---

*最后更新：2026-10-24*
