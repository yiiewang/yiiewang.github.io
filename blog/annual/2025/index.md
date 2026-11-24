# 2025 年度集锦

> 一月把 Go 的细节补了一遍，三月转身去搞懂区块链，其余月份在软考、读书和市场之间来回跳。

## 卷首语

2025 的开头很密：1 月 12 日到 1 月 25 日，两周写了 15 篇 Go——切片删除、字符串与 `[]byte`、 `rune` 迭代、`nil` 规则、变量遮蔽、方法接收者、`RWMutex`、`sync.Map`、原子操作、NoCopy、 无锁编程。那半个月做的事只有一件：把"会用 Go"往"知道 Go 为什么这样"推。

三月换方向，从[拜占庭问题](<https://cloaks.cn/blog/posts/2025/03/12-2/index.md>)一路读到[零知识证明](<https://cloaks.cn/blog/posts/2025/04/11/index.md>)。 五月之后节奏明显慢下来：全年 41 篇，有一半压在 1 月和 3 月。

## 主线一：Go 的两周

集中在 1 月 12—25 日，可以分成三小簇：

- **语言细节**：[切片删除的两种方法](<https://cloaks.cn/blog/posts/2025/01/12/index.md>)、[`nil` 与变量初始化](<https://cloaks.cn/blog/posts/2025/01/17-1/index.md>)、[变量幽灵（Shadowing）](<https://cloaks.cn/blog/posts/2025/01/17/index.md>)、[字符串与 `[]byte` 的转换](<https://cloaks.cn/blog/posts/2025/01/17-3/index.md>)、[字符串与 `rune` 的迭代](<https://cloaks.cn/blog/posts/2025/01/17-4/index.md>)、[map 的内存与容量](<https://cloaks.cn/blog/posts/2025/01/17-2/index.md>)、[方法接收者的值与指针](<https://cloaks.cn/blog/posts/2025/01/18/index.md>)
- **资源与连接**：[正确关闭 `Response.Body`](<https://cloaks.cn/blog/posts/2025/01/18-1/index.md>)、[长连接与短连接](<https://cloaks.cn/blog/posts/2025/01/24/index.md>)、[负载均衡设计](<https://cloaks.cn/blog/posts/2025/01/19/index.md>)
- **并发**：[`RWMutex` 底层实现](<https://cloaks.cn/blog/posts/2025/01/24-1/index.md>)、[`sync.Map` 读写分离](<https://cloaks.cn/blog/posts/2025/01/25/index.md>)、[原子操作与 `unsafe.Pointer`](<https://cloaks.cn/blog/posts/2025/01/25-3/index.md>)、[NoCopy 机制](<https://cloaks.cn/blog/posts/2025/01/25-1/index.md>)、[结构体拷贝与锁](<https://cloaks.cn/blog/posts/2025/01/25-2/index.md>)、[无锁编程](<https://cloaks.cn/blog/posts/2025/01/25-4/index.md>)

那两周最后一篇是[无锁编程：内存屏障与无锁栈、队列](<https://cloaks.cn/blog/posts/2025/01/25-4/index.md>)。 当时只是笔记，一年后它长成了另一条长线的起点。

## 主线二：从拜占庭问题到零知识证明

三月开工，顺序是刻意排的：先[拜占庭问题](<https://cloaks.cn/blog/posts/2025/03/12-2/index.md>)搞清楚"什么叫不可靠节点"， 再读 [PBFT](<https://cloaks.cn/blog/posts/2025/03/12-1/index.md>) 的三阶段协议与视图变更，然后看 [比特币的创新设计](<https://cloaks.cn/blog/posts/2025/03/20/index.md>)怎么用博弈论和 PoW 绕开同一道题。

> [!NOTE]
>
> **这一年对区块链的判断**
>
> PBFT 是"在已知成员里达成一致"，比特币是"在陌生人的网络里达成一致"—— 同一个问题的两种代价。

四月[零知识证明](<https://cloaks.cn/blog/posts/2025/04/11/index.md>)，五月[区块 DAG](<https://cloaks.cn/blog/posts/2025/05/06/index.md>)收尾。 同期两篇工程侧：[分布式系统的核心问题](<https://cloaks.cn/blog/posts/2025/03/14/index.md>)（FLP / CAP / ACID）、 [高频 `R/W` 场景的性能设计](<https://cloaks.cn/blog/posts/2025/03/27/index.md>)（Cache-Aside 到 Write-Back）。

## 主线三：软考、读书与市场

- **软考**：[高级架构师高频题型总结](<https://cloaks.cn/blog/posts/2025/03/12/index.md>)——这一年唯一的软考产出
- **读书**：[《男人来自火星，女人来自金星》](<https://cloaks.cn/blog/posts/2025/02/3/index.md>)、[《爱的艺术》](<https://cloaks.cn/blog/posts/2025/03/3/index.md>)、[组件设计原则](<https://cloaks.cn/blog/posts/2025/03/31/index.md>)（Clean Architecture 的 REP/CCP/CRP）、[君子不器](<https://cloaks.cn/blog/posts/2025/06/03/index.md>)
- **市场**：[什么是通货膨胀](<https://cloaks.cn/blog/posts/2025/05/05-1/index.md>)、[农产品价格和通胀的关系](<https://cloaks.cn/blog/posts/2025/05/05-2/index.md>)、年底一篇[股票上涨的本质](<https://cloaks.cn/blog/posts/2025/12/04/index.md>)

## 支线：算法、工具与两篇 COM

- **算法四篇**：[回文十进制数](<https://cloaks.cn/blog/posts/2025/01/15/index.md>)、[组合问题](<https://cloaks.cn/blog/posts/2025/03/8/index.md>)、[堆排序](<https://cloaks.cn/blog/posts/2025/08/18/index.md>)、[时间轮算法](<https://cloaks.cn/blog/posts/2025/12/15/index.md>)
- **工具**：[用 Shell 管理 OpenVPN 守护进程](<https://cloaks.cn/blog/posts/2025/01/2/index.md>)、[bufio.Scanner: token too long](<https://cloaks.cn/blog/posts/2025/07/24/index.md>)
- **札记**：[深入了解 COM](<https://cloaks.cn/blog/posts/2025/02/21/index.md>)、[用 C++ 实现 COM 组件](<https://cloaks.cn/blog/posts/2025/02/21-1/index.md>)。这两篇是全年最孤立的——前后都没有同类

## 代表作五篇

1. [无锁编程：内存屏障与无锁栈、队列的实现](<https://cloaks.cn/blog/posts/2025/01/25-4/index.md>) — 今年最硬的一篇，也是次年那条十五篇长线的种子
2. [深入理解 Go 的原子操作与 `unsafe.Pointer`](<https://cloaks.cn/blog/posts/2025/01/25-3/index.md>) — 讲清了"为什么原子操作要配 unsafe.Pointer"
3. [PBFT 共识算法](<https://cloaks.cn/blog/posts/2025/03/12-1/index.md>) — 三阶段协议讲得比多数中文资料清楚
4. [零知识证明：不泄露秘密的证明艺术](<https://cloaks.cn/blog/posts/2025/04/11/index.md>) — 从直觉到原理，全程不丢人
5. [时间轮算法](<https://cloaks.cn/blog/posts/2025/12/15/index.md>) — 年底最好的一篇，滑动时间窗口从 O(n) 压到 O(1)

## 全年目录

### 1 月 · 18 篇

- [使用 Shell 脚本管理 OpenVPN 守护进程](<https://cloaks.cn/blog/posts/2025/01/2/index.md>) · 01-02 · DevOps 与工具
- [Go 切片中删除元素的两种方法对比](<https://cloaks.cn/blog/posts/2025/01/12/index.md>) · 01-12 · Go 语言
- [回文十进制数](<https://cloaks.cn/blog/posts/2025/01/15/index.md>) · 01-15 · 算法与数据结构
- [Golang 中的意外变量幽灵（Accidental Variable Shadowing）](<https://cloaks.cn/blog/posts/2025/01/17/index.md>) · 01-17 · Go 语言
- [Go 语言中的 `nil` 和变量初始化规则](<https://cloaks.cn/blog/posts/2025/01/17-1/index.md>) · 01-17 · Go 语言
- [Go 语言中 `map` 的内存分配与容量管理](<https://cloaks.cn/blog/posts/2025/01/17-2/index.md>) · 01-17 · Go 语言
- [Go 语言中字符串与 `[]byte` 之间的转换与性能优化](<https://cloaks.cn/blog/posts/2025/01/17-3/index.md>) · 01-17 · Go 语言
- [Go 语言中字符串与 `rune` 的迭代与非 UTF-8 字符的处理](<https://cloaks.cn/blog/posts/2025/01/17-4/index.md>) · 01-17 · Go 语言
- [Go 方法接收者的值类型与指针类型详解](<https://cloaks.cn/blog/posts/2025/01/18/index.md>) · 01-18 · Go 语言
- [正确关闭 HTTP Response.Body 的重要性](<https://cloaks.cn/blog/posts/2025/01/18-1/index.md>) · 01-18 · Go 语言
- [负载均衡设计思路与实践指南](<https://cloaks.cn/blog/posts/2025/01/19/index.md>) · 01-19 · 系统与架构
- [长连接与短链接的学习](<https://cloaks.cn/blog/posts/2025/01/24/index.md>) · 01-24 · Go 语言
- [读写锁（`sync.RWMutex`）的底层实现原理](<https://cloaks.cn/blog/posts/2025/01/24-1/index.md>) · 01-24 · Go 语言
- [读写分离 `sync.Map`](<https://cloaks.cn/blog/posts/2025/01/25/index.md>) · 01-25 · Go 语言
- [理解 Go 语言中的 NoCopy 机制](<https://cloaks.cn/blog/posts/2025/01/25-1/index.md>) · 01-25 · Go 语言
- [理解 Go 语言中结构体拷贝和锁的行为](<https://cloaks.cn/blog/posts/2025/01/25-2/index.md>) · 01-25 · Go 语言
- [深入理解 Go 的原子操作与 `unsafe.Pointer`](<https://cloaks.cn/blog/posts/2025/01/25-3/index.md>) · 01-25 · Go 语言
- [无锁编程：内存屏障与无锁栈、队列的实现](<https://cloaks.cn/blog/posts/2025/01/25-4/index.md>) · 01-25 · Go 语言

### 2 月 · 4 篇

- [男人来自火星，女人来自金星](<https://cloaks.cn/blog/posts/2025/02/3/index.md>) · 02-03 · 读书与思考
- [当时只道是寻常](<https://cloaks.cn/blog/posts/2025/02/17/index.md>) · 02-17 · 读书与思考
- [深入了解 COM（Component Object Model）](<https://cloaks.cn/blog/posts/2025/02/21/index.md>) · 02-21 · 札记
- [使用 C++ 实现 COM 组件的示例](<https://cloaks.cn/blog/posts/2025/02/21-1/index.md>) · 02-21 · 札记

### 3 月 · 10 篇

- [爱的艺术](<https://cloaks.cn/blog/posts/2025/03/3/index.md>) · 03-03 · 读书与思考
- [关于组合问题的算法设计](<https://cloaks.cn/blog/posts/2025/03/8/index.md>) · 03-08 · 算法与数据结构
- [高级架构师高频题型总结](<https://cloaks.cn/blog/posts/2025/03/12/index.md>) · 03-12 · 软考
- [PBFT（Practical Byzantine Fault Tolerance）共识算法](<https://cloaks.cn/blog/posts/2025/03/12-1/index.md>) · 03-12 · 区块链
- [拜占庭问题解析](<https://cloaks.cn/blog/posts/2025/03/12-2/index.md>) · 03-12 · 区块链
- [分布式系统的核心问题](<https://cloaks.cn/blog/posts/2025/03/14/index.md>) · 03-14 · 系统与架构
- [比特币的创新设计](<https://cloaks.cn/blog/posts/2025/03/20/index.md>) · 03-20 · 区块链
- [2025-03-23 有感](<https://cloaks.cn/blog/posts/2025/03/23/index.md>) · 03-23 · 读书与思考
- [高频 `R/W` 场景下的性能设计方案](<https://cloaks.cn/blog/posts/2025/03/27/index.md>) · 03-27 · 系统与架构
- [组件设计原则](<https://cloaks.cn/blog/posts/2025/03/31/index.md>) · 03-31 · 读书与思考

### 4 月 · 1 篇

- [零知识证明：不泄露秘密的证明艺术](<https://cloaks.cn/blog/posts/2025/04/11/index.md>) · 04-11 · 区块链

### 5 月 · 3 篇

- [什么是通货膨胀？](<https://cloaks.cn/blog/posts/2025/05/05-1/index.md>) · 05-05 · 投资与交易
- [农产品价格和通胀的关系](<https://cloaks.cn/blog/posts/2025/05/05-2/index.md>) · 05-05 · 投资与交易
- [区块 DAG（Directed Acyclic Graph，有向无环图）](<https://cloaks.cn/blog/posts/2025/05/06/index.md>) · 05-06 · 区块链

### 6 月 · 1 篇

- [君子不器](<https://cloaks.cn/blog/posts/2025/06/03/index.md>) · 06-03 · 读书与思考

### 7 月 · 1 篇

- [解决：bufio.Scanner: token too long 错误](<https://cloaks.cn/blog/posts/2025/07/24/index.md>) · 07-24 · Go 语言

### 8 月 · 1 篇

- [堆排序](<https://cloaks.cn/blog/posts/2025/08/18/index.md>) · 08-18 · 算法与数据结构

### 12 月 · 2 篇

- [股票上涨的本质](<https://cloaks.cn/blog/posts/2025/12/04/index.md>) · 12-03 · 投资与交易
- [时间轮算法：高效实现滑动时间窗口统计](<https://cloaks.cn/blog/posts/2025/12/15/index.md>) · 12-15 · 算法与数据结构

## 回看：2026 给出的答案

1. 一月那批无锁与内存屏障的笔记，在 2026 年九月长成了[十五篇的因果链](<https://cloaks.cn/blog/posts/2026/09/10/index.md>)，一路追到六百万 TPS
2. 三月的区块链入门，2026 年变成了长安链压测报告、跨链追踪器和[拆库记](<https://cloaks.cn/blog/posts/2026/10/20/index.md>)
3. 年底那篇《股票上涨的本质》，2026 年接上了量化交易与全天候复合系统的四季实测
4. 软考停在三月——2026 年一篇软考都没写

---

*最后更新：2026-10-24*
