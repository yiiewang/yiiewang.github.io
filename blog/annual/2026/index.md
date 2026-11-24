# 2026 年度集锦

> 一年的文章太多，目录会被时间流冲散。这一页是重新串起来的尝试：**不按日期列，按因果列。**

> [!NOTE]
>
> **统计口径**
>
> 本篇写于 2026-10-24，覆盖 1—10 月；年末补齐 11—12 月后再更新一次。 全年目录由 `annual_list(2026)` 宏实时生成，这一年再发新文章会自动跟上， 不需要回来手改链接。

## 卷首语：今年其实只有一件事

九月 review 一套布谷鸟过滤器分片的落盘方案，对方问了句「这套设计叫什么名字」——我卡住了。 它没有名字，因为它是五个经典思想的排列组合。为了把这个窟窿补上，我把这句问话一路追下去： 无锁队列、内存屏障、双缓冲、LMAX 与 Disruptor，最后停在"六百万 TPS 的单线程"上， 回头写了[十五篇的因果链](<https://cloaks.cn/blog/posts/2026/09/10/index.md>)。

这一条线吃掉了九月整整一个月，也顺手把别的东西带动起来：为了搞清楚队列去拆 LevelDB， 为了搞清楚"写了不等于看得见"去翻 Go 内存模型，月底又开了 xv6 的新坑。 剩下的交易、长安链、AI 三条线各有各的问题来源，但写法是同一套：**先承认自己说不清， 再一路追到说得清为止。**

## 主线一：从一个落盘方案到六百万 TPS

今年最重的一坨，21 篇系统与架构里大半在九月。

起点是[一个没有名字的持久化设计](<https://cloaks.cn/blog/posts/2026/09/04-1/index.md>)：五个经典思想（位打包、 状态上移、双缓冲、writeback、重放）的排列组合。写完我发现自己讲不清其中最基础的两行代码， 于是有了[位打包两行代码的完全解剖](<https://cloaks.cn/blog/posts/2026/09/04/index.md>)——`(N+7)/8` 和 `1<<(7-i%8)` 到底在干什么。

从这里分岔出两条路：

- **往并发走**：[单写多读为什么敢不用锁](<https://cloaks.cn/blog/posts/2026/09/07/index.md>) → [两个生产者之后](<https://cloaks.cn/blog/posts/2026/09/07-1/index.md>)的三条路 → [从自旋到休眠](<https://cloaks.cn/blog/posts/2026/09/08-2/index.md>)的成本谱系 → [一把锁罩住一切](<https://cloaks.cn/blog/posts/2026/09/08-1/index.md>)（channel 为什么不追求无锁）→ [实测：channel 到底慢多少](<https://cloaks.cn/blog/posts/2026/09/08-6/index.md>)
- **往内存走**：[写了不等于看得见](<https://cloaks.cn/blog/posts/2026/09/08/index.md>)（内存屏障与 Go 内存模型）→ [切换必须原子](<https://cloaks.cn/blog/posts/2026/09/08-5/index.md>)（双缓冲的命根子）→ [dirty 是一张欠条](<https://cloaks.cn/blog/posts/2026/09/08-4/index.md>)（writeback 的调度哲学）

两条路在[六百万 TPS 的单线程](<https://cloaks.cn/blog/posts/2026/09/08-8/index.md>)汇合：LMAX 架构与 Disruptor 精读。 读完之后回头看，[归并写者](<https://cloaks.cn/blog/posts/2026/09/08-7/index.md>)把竞争消灭在架构层， [重放与影分身](<https://cloaks.cn/blog/posts/2026/09/08-3/index.md>)解决崩溃恢复， [「满：等消费者腾地方」](<https://cloaks.cn/blog/posts/2026/09/09-1/index.md>)则把一句注释实测了一遍。

年底又接了两根尾巴：[不用堆的延迟队列](<https://cloaks.cn/blog/posts/2026/09/29-3/index.md>)和 [拆库记：LevelDB 分片演进](<https://cloaks.cn/blog/posts/2026/10/20/index.md>)——后者是"每一步都是为了解决什么问题"的写法样板。

> [!TIP]
>
> **这条线的读法**
>
> 别按目录顺序读。先读[十五篇的因果链](<https://cloaks.cn/blog/posts/2026/09/10/index.md>)拿地图， 再挑你卡住的那一篇钻进去。

## 主线二：长安链，从压测到跨链

三月开工：先立[运维与压测手册](<https://cloaks.cn/blog/posts/2026/03/7/index.md>)，然后把跨分片比例当唯一变量， 连做三份报告——[1%](<https://cloaks.cn/blog/posts/2026/03/9/index.md>)、[5%](<https://cloaks.cn/blog/posts/2026/03/10/index.md>)、[20%](<https://cloaks.cn/blog/posts/2026/03/20/index.md>)。

> [!NOTE]
>
> **结论一句话**
>
> 跨分片比例是吞吐的转折量，不是平滑衰减——到某个点之后，性能是塌的，不是降的。

五月补上[模块测试实战手册](<https://cloaks.cn/blog/posts/2026/05/27/index.md>)，把"怎么测"也规范化。 九月之后重心转到跨链与查询：[一个交集三种写法](<https://cloaks.cn/blog/posts/2026/09/29/index.md>)、 [跨链交易追踪器的三次手术与两处隐患](<https://cloaks.cn/blog/posts/2026/09/29-2/index.md>)、 [一个跨链事件的旅程](<https://cloaks.cn/blog/posts/2026/10/12/index.md>)（3 个子事件，恰好 1 次 Commit）、 [SDK-Go 从跑通到用对](<https://cloaks.cn/blog/posts/2026/10/09-1/index.md>)、 [事件驱动入门](<https://cloaks.cn/blog/posts/2026/10/09-2/index.md>)。

两处隐患至今没拆干净，这是明年的活。

## 主线三：交易系统，从设计手记到四季实测

二月读《量化交易》，顺手做了[货币中性](<https://cloaks.cn/blog/posts/2026/02/8/index.md>)和 [GLD/GDX 配对交易](<https://cloaks.cn/blog/posts/2026/02/9/index.md>)两个策略，三月用 CTA 的观点写了 [系统性抛压](<https://cloaks.cn/blog/posts/2026/03/10-1/index.md>)。

四月出[全天候复合交易系统的设计手记](<https://cloaks.cn/blog/posts/2026/04/30/index.md>)：三层架构 + 三档止盈。 十月用两篇实测把它收官——[信号层：多信号确认到底有没有边际价值](<https://cloaks.cn/blog/posts/2026/10/21/index.md>)、 [仓位层：三档止盈、网格与切换](<https://cloaks.cn/blog/posts/2026/10/21-1/index.md>)。

四季实测最大的收获是**否定**：多加信号不涨收益，网格在单边市里必死， 单标的再怎么调参也扛不住系统性风险。这些结论单独看都很朴素，但它们是跑出来的，不是想出来的。

同期还写了货殖列传系列：[《史记·货殖列传》精讲](<https://cloaks.cn/blog/posts/2026/10/01-1/index.md>)、 [「活钱」作战手册](<https://cloaks.cn/blog/posts/2026/10/02/index.md>)、[范蠡的周期模型](<https://cloaks.cn/blog/posts/2026/10/02-1/index.md>) 及其[数据验证](<https://cloaks.cn/blog/posts/2026/10/02-2/index.md>)。两千年前的文字，用现代数据跑一遍居然能闭环。

## 支线：Go、AI、读书与 xv6

- **Go（15 篇）**：从[TCP 网络编程](<https://cloaks.cn/blog/posts/2026/02/10/index.md>)、[gRPC Stream 三模式](<https://cloaks.cn/blog/posts/2026/02/11/index.md>)、[滑动窗口缓存](<https://cloaks.cn/blog/posts/2026/02/12/index.md>)，到[幂等事故复盘](<https://cloaks.cn/blog/posts/2026/02/28/index.md>)、[优雅退出](<https://cloaks.cn/blog/posts/2026/03/23/index.md>)、[SDK 死锁排查](<https://cloaks.cn/blog/posts/2026/06/03/index.md>)、[锁的 15 个反模式](<https://cloaks.cn/blog/posts/2026/06/04/index.md>)、[atomic.Value 装箱](<https://cloaks.cn/blog/posts/2026/09/29-1/index.md>)。主线是"把工程里踩过的坑写成清单"
- **AI（7 篇）**：从[用 CodeBuddy 部署项目](<https://cloaks.cn/blog/posts/2026/02/6/index.md>)、[HALO](<https://cloaks.cn/blog/posts/2026/02/25/index.md>)、[多智能体协同](<https://cloaks.cn/blog/posts/2026/03/19/index.md>)、[Prompt 工具箱](<https://cloaks.cn/blog/posts/2026/04/15/index.md>)、[LangChain 路线图](<https://cloaks.cn/blog/posts/2026/05/07/index.md>)，到七月那句[瓶颈往往不在模型](<https://cloaks.cn/blog/posts/2026/07/31/index.md>)。态度从兴奋转为务实
- **读书（5 篇）**：[《国富论》第一章](<https://cloaks.cn/blog/posts/2026/10/05/index.md>)（盈利地图）/[二三四章](<https://cloaks.cn/blog/posts/2026/10/05-1/index.md>)（传动系统）/[第五章](<https://cloaks.cn/blog/posts/2026/10/05-2/index.md>)（会漂移的秤）、[毛选《我们的经济政策》](<https://cloaks.cn/blog/posts/2026/07/04/index.md>)、[《爱的艺术》](<https://cloaks.cn/blog/posts/2026/10/19/index.md>)。外加一篇[县域钢厂的商业观察](<https://cloaks.cn/blog/posts/2026/10/05-3/index.md>)
- **xv6（9 月启动，进行中）**：[总索引](<https://cloaks.cn/blog/posts/2026/09/17/index.md>)→[00 出发](<https://cloaks.cn/blog/posts/2026/09/17-1/index.md>)→[01 操作系统接口](<https://cloaks.cn/blog/posts/2026/09/18/index.md>)→[02 操作系统组织](<https://cloaks.cn/blog/posts/2026/09/18-1/index.md>)。才读到 Ch.2

## 代表作五篇

1. [从一个落盘方案到六百万 TPS：十五篇的因果链](<https://cloaks.cn/blog/posts/2026/09/10/index.md>) — 今年的骨架。如果你只打算读一篇，读这篇，它会告诉你其余十五篇各自站在哪一格
2. [单写多读：无锁环形队列为什么敢不用锁](<https://cloaks.cn/blog/posts/2026/09/07/index.md>) — "敢"字背后是三件事：单写者假设、序列号单调、内存屏障。讲透了才敢不用锁
3. [六百万 TPS 的单线程：LMAX 架构与 Disruptor 精读](<https://cloaks.cn/blog/posts/2026/09/08-8/index.md>) — 把"单线程比多线程快"这个反直觉结论拆到能复现
4. [三档止盈、网格与切换——仓位层的实测手记](<https://cloaks.cn/blog/posts/2026/10/21-1/index.md>) — 一整年交易实验的终局，结论是"调参救不了结构问题"
5. [一次重复回调引发的库存事故](<https://cloaks.cn/blog/posts/2026/02/28/index.md>) — 唯一一篇从真实事故出发的。幂等不是概念题，是"多扣了一件货"的题

## 全年目录

年内按时间正序，方便顺着看这一年是怎么长出来的。

### 2 月 · 9 篇

- [使用 CodeBuddy 快速部署项目到腾讯云 Lighthouse](<https://cloaks.cn/blog/posts/2026/02/6/index.md>) · 02-06 · DevOps 与工具
- [《量化交易：如何建立自己的算法交易事业》读书笔记](<https://cloaks.cn/blog/posts/2026/02/7/index.md>) · 02-07 · 投资与交易
- [货币中性策略：不看大盘也能赚钱的交易方法](<https://cloaks.cn/blog/posts/2026/02/8/index.md>) · 02-08 · 投资与交易
- [配对交易实战：GLD 与 GDX 的均值回归策略](<https://cloaks.cn/blog/posts/2026/02/9/index.md>) · 02-09 · 投资与交易
- [Go TCP 网络编程实战：Echo、Proxy、Scanner](<https://cloaks.cn/blog/posts/2026/02/10/index.md>) · 02-10 · Go 语言
- [gRPC Stream 三种模式详解：服务端流、客户端流、双向流](<https://cloaks.cn/blog/posts/2026/02/11/index.md>) · 02-11 · Go 语言
- [高性能滑动窗口缓存设计：环形缓冲区 + 分片锁](<https://cloaks.cn/blog/posts/2026/02/12/index.md>) · 02-12 · Go 语言
- [HALO：AI 时代的投资新范式](<https://cloaks.cn/blog/posts/2026/02/25/index.md>) · 02-25 · 投资与交易
- [一次重复回调引发的库存事故：Go 服务幂等设计实战](<https://cloaks.cn/blog/posts/2026/02/28/index.md>) · 02-28 · Go 语言

### 3 月 · 7 篇

- [长安链运维与压测手册](<https://cloaks.cn/blog/posts/2026/03/7/index.md>) · 03-07 · 区块链
- [风暴尚未过去：CTA 全面转空，系统性抛压或创年内之最](<https://cloaks.cn/blog/posts/2026/03/10-1/index.md>) · 03-10 · 投资与交易
- [黄仁勋长文解读：AI 是"五层蛋糕"基础设施](<https://cloaks.cn/blog/posts/2026/03/11/index.md>) · 03-11 · AI 实践
- [从失败到智能创造：人机协作突破全记录](<https://cloaks.cn/blog/posts/2026/03/14/index.md>) · 03-14 · AI 实践
- [多智能体协同：从理论到工程的完整实战指南](<https://cloaks.cn/blog/posts/2026/03/19/index.md>) · 03-19 · AI 实践
- [当 AI 开始"圆桌讨论"：一次代码决策的新体验](<https://cloaks.cn/blog/posts/2026/03/19-2/index.md>) · 03-19 · AI 实践
- [Go 优雅退出完全指南：signal.NotifyContext 实战](<https://cloaks.cn/blog/posts/2026/03/23/index.md>) · 03-23 · Go 语言

### 4 月 · 2 篇

- [AI Prompt 工具箱：不同场景下的实用提示词](<https://cloaks.cn/blog/posts/2026/04/15/index.md>) · 04-15 · AI 实践
- [当趋势策略遇上震荡市：一套全天候复合交易系统的设计手记](<https://cloaks.cn/blog/posts/2026/04/30/index.md>) · 04-30 · 投资与交易

### 5 月 · 3 篇

- [LangChain & LangGraph 学习路线图与水平定位](<https://cloaks.cn/blog/posts/2026/05/07/index.md>) · 05-07 · AI 实践
- [环形缓冲区（Ring Buffer）：从原理到区块链交易去重的实战应用](<https://cloaks.cn/blog/posts/2026/05/26/index.md>) · 05-26 · 区块链
- [长安链模块测试 · 实战手册（Module Test Playbook）](<https://cloaks.cn/blog/posts/2026/05/27/index.md>) · 05-27 · 区块链

### 6 月 · 4 篇

- [Go 并发资源关闭类缺陷排查实战 —— 从一个 SDK 死锁案例说起](<https://cloaks.cn/blog/posts/2026/06/03/index.md>) · 06-03 · Go 语言
- [Go 并发编程：锁与同步的 15 个反模式及正确做法](<https://cloaks.cn/blog/posts/2026/06/04/index.md>) · 06-04 · Go 语言
- [集成测试流程设计方案](<https://cloaks.cn/blog/posts/2026/06/15/index.md>) · 06-15 · 札记
- [如何做一场有效的汇报？](<https://cloaks.cn/blog/posts/2026/06/30/index.md>) · 06-30 · 札记

### 7 月 · 2 篇

- [读毛选《我们的经济政策》—— 战略与战术的辩证法](<https://cloaks.cn/blog/posts/2026/07/04/index.md>) · 07-04 · 读书与思考
- [让 AI 落地，瓶颈往往不在模型](<https://cloaks.cn/blog/posts/2026/07/31/index.md>) · 07-31 · AI 实践

### 9 月 · 29 篇

- [用反向 SSH 隧道穿透内网：一次完整的踩坑实践](<https://cloaks.cn/blog/posts/2026/09/02/index.md>) · 09-02 · DevOps 与工具
- [ps -ef | grep ssh 查不到的端口转发](<https://cloaks.cn/blog/posts/2026/09/02-1/index.md>) · 09-02 · DevOps 与工具
- [位打包两行代码的完全解剖：(N+7)/8 与 1\<\<(7-i%8)](<https://cloaks.cn/blog/posts/2026/09/04/index.md>) · 09-04 · Go 语言
- [一个没有名字的持久化设计：五个经典思想的排列组合](<https://cloaks.cn/blog/posts/2026/09/04-1/index.md>) · 09-04 · 系统与架构
- [状态上移，原子下沉：包装优于修改的设计取舍](<https://cloaks.cn/blog/posts/2026/09/06/index.md>) · 09-06 · 系统与架构
- [单写多读：无锁环形队列为什么敢不用锁](<https://cloaks.cn/blog/posts/2026/09/07/index.md>) · 09-07 · 系统与架构
- [两个生产者之后：给无锁环形队列加同步的三条路](<https://cloaks.cn/blog/posts/2026/09/07-1/index.md>) · 09-07 · 系统与架构
- [写了不等于看得见：内存屏障与 Go 内存模型](<https://cloaks.cn/blog/posts/2026/09/08/index.md>) · 09-08 · 系统与架构
- [一把锁罩住一切：Go channel 为什么不追求无锁](<https://cloaks.cn/blog/posts/2026/09/08-1/index.md>) · 09-08 · 系统与架构
- [从自旋到休眠：「等」的成本谱系](<https://cloaks.cn/blog/posts/2026/09/08-2/index.md>) · 09-08 · 系统与架构
- [重放与影分身：崩溃恢复的两条路线](<https://cloaks.cn/blog/posts/2026/09/08-3/index.md>) · 09-08 · 系统与架构
- [dirty 是一张欠条：writeback 的调度哲学](<https://cloaks.cn/blog/posts/2026/09/08-4/index.md>) · 09-08 · 系统与架构
- [切换必须原子：双缓冲家族的命根子](<https://cloaks.cn/blog/posts/2026/09/08-5/index.md>) · 09-08 · 系统与架构
- [实测：channel 到底比无锁队列慢多少](<https://cloaks.cn/blog/posts/2026/09/08-6/index.md>) · 09-08 · 系统与架构
- [归并写者：把竞争消灭在架构层](<https://cloaks.cn/blog/posts/2026/09/08-7/index.md>) · 09-08 · 系统与架构
- [六百万 TPS 的单线程：LMAX 架构与 Disruptor 精读](<https://cloaks.cn/blog/posts/2026/09/08-8/index.md>) · 09-08 · 系统与架构
- [一行 ProxyJump，VS Code 直达内网机器](<https://cloaks.cn/blog/posts/2026/09/09/index.md>) · 09-09 · DevOps 与工具
- [「满：等消费者腾地方」——把这句注释实测一遍](<https://cloaks.cn/blog/posts/2026/09/09-1/index.md>) · 09-09 · 系统与架构
- [从一个落盘方案到六百万 TPS：十五篇的因果链](<https://cloaks.cn/blog/posts/2026/09/10/index.md>) · 09-10 · 系统与架构
- [从字节流到持久化：Go 后端一条线上的八件事](<https://cloaks.cn/blog/posts/2026/09/15/index.md>) · 09-15 · Go 语言
- [xv6 学习历程：路线图与总索引](<https://cloaks.cn/blog/posts/2026/09/17/index.md>) · 09-17 · 系统与架构
- [xv6 学习历程 00 · 出发：为什么读 xv6，怎么读](<https://cloaks.cn/blog/posts/2026/09/17-1/index.md>) · 09-17 · 系统与架构
- [xv6 学习历程 01 · Ch.1 操作系统接口：进程、fd 与管道](<https://cloaks.cn/blog/posts/2026/09/18/index.md>) · 09-18 · 系统与架构
- [xv6 学习历程 02 · Ch.2 操作系统组织：宏内核与第一次 ecall](<https://cloaks.cn/blog/posts/2026/09/18-1/index.md>) · 09-18 · 系统与架构
- [一个交集，三种写法：分片缺失交易查询的三版演进](<https://cloaks.cn/blog/posts/2026/09/29/index.md>) · 09-29 · Go 语言
- [atomic.Value 存接口为什么会 panic：装箱惯用法拆解](<https://cloaks.cn/blog/posts/2026/09/29-1/index.md>) · 09-29 · Go 语言
- [从全量遍历到 O(1)：跨链交易追踪器的三次手术与两处隐患](<https://cloaks.cn/blog/posts/2026/09/29-2/index.md>) · 09-29 · Go 语言
- [不用堆的延迟队列：一条决策链，和链上炸出的 panic](<https://cloaks.cn/blog/posts/2026/09/29-3/index.md>) · 09-29 · Go 语言
- [队列的内存演进史：从链表到 \[\]\[\]string 的五个台阶](<https://cloaks.cn/blog/posts/2026/09/30/index.md>) · 09-30 · Go 语言

### 10 月 · 23 篇

- [锁外回调、队列锁、分片锁：一句追问引出的死锁预防链](<https://cloaks.cn/blog/posts/2026/10/01/index.md>) · 10-01 · Go 语言
- [《史记·货殖列传》精讲：两千年前把市场规律说透的一篇文章](<https://cloaks.cn/blog/posts/2026/10/01-1/index.md>) · 10-01 · 投资与交易
- [「活钱」作战手册：把货殖列传翻译给现代人看](<https://cloaks.cn/blog/posts/2026/10/02/index.md>) · 10-02 · 投资与交易
- [三句话，一个闭环：范蠡经商智慧的周期模型](<https://cloaks.cn/blog/posts/2026/10/02-1/index.md>) · 10-02 · 投资与交易
- [三句话，三块仪表盘：用现代数据验证范蠡的周期模型](<https://cloaks.cn/blog/posts/2026/10/02-2/index.md>) · 10-02 · 投资与交易
- [把《国富论》第一章读成一张盈利地图](<https://cloaks.cn/blog/posts/2026/10/05/index.md>) · 10-05 · 读书与思考
- [把《国富论》二三四章读成一套传动系统](<https://cloaks.cn/blog/posts/2026/10/05-1/index.md>) · 10-05 · 读书与思考
- [把《国富论》第五章读成一把会漂移的秤](<https://cloaks.cn/blog/posts/2026/10/05-2/index.md>) · 10-05 · 读书与思考
- [当一家钢厂开始建城：县域支柱民企的创业、管理与未来](<https://cloaks.cn/blog/posts/2026/10/05-3/index.md>) · 10-05 · 商业观察
- [长安链 SDK-Go：从跑通到用对](<https://cloaks.cn/blog/posts/2026/10/09-1/index.md>) · 10-09 · 区块链
- [事件驱动入门：从“A 调 B”到手写一个事件总线](<https://cloaks.cn/blog/posts/2026/10/09-2/index.md>) · 10-09 · 系统与架构
- [一个跨链事件的旅程：3 个子事件，恰好 1 次 Commit](<https://cloaks.cn/blog/posts/2026/10/12/index.md>) · 10-12 · 区块链
- [把《爱的艺术》第二章第一节读成一张孤独突围路线图](<https://cloaks.cn/blog/posts/2026/10/19/index.md>) · 10-19 · 读书与思考
- [拆库记：LevelDB 分片演进，每一步都是为了解决什么问题](<https://cloaks.cn/blog/posts/2026/10/20/index.md>) · 10-20 · 系统与架构
- [确认越多越好吗？——信号层的实测手记](<https://cloaks.cn/blog/posts/2026/10/21/index.md>) · 10-21 · 投资与交易
- [三档止盈、网格与切换——仓位层的实测手记](<https://cloaks.cn/blog/posts/2026/10/21-1/index.md>) · 10-21 · 投资与交易
- [均线的物理课：EMA 的重心、速度与加速度](<https://cloaks.cn/blog/posts/2026/10/24/index.md>) · 10-24 · 投资与交易
- [信号有没有用？14 个标的、7450 个事件的对照实验](<https://cloaks.cn/blog/posts/2026/10/24-1/index.md>) · 10-24 · 投资与交易
- [复合交易系统操作手册：一份能照着执行的清单](<https://cloaks.cn/blog/posts/2026/10/24-2/index.md>) · 10-24 · 投资与交易
- [从 ssh -L 到自研多路复用：手写一个隧道协议](<https://cloaks.cn/blog/posts/2026/10/27/index.md>) · 10-27 · Go 语言
- [一条 TCP 上跑 100 条流：手写帧协议与 stream 生命周期](<https://cloaks.cn/blog/posts/2026/10/27-1/index.md>) · 10-27 · Go 语言
- [把 1000 条 TCP 压成 1 条：一次设计推导的完整记录](<https://cloaks.cn/blog/posts/2026/10/28/index.md>) · 10-28 · 系统与架构
- [让隧道敢监听公网：三个问题、标准方案与我们的取舍](<https://cloaks.cn/blog/posts/2026/10/28-1/index.md>) · 10-28 · 系统与架构

### 11 月 · 3 篇

- [从"能跑"到"敢一直开着"：心跳、重连与握手自检](<https://cloaks.cn/blog/posts/2026/11/02/index.md>) · 11-02 · Go 语言
- [隧道掉头：反向转发、DoS 封顶与单张证书吊销](<https://cloaks.cn/blog/posts/2026/11/10/index.md>) · 11-10 · Go 语言
- [十一月的黄昏](<https://cloaks.cn/blog/posts/2026/11/14/index.md>) · 11-14 · 札记

## 留给 2027 的问题

1. xv6 才读到 Ch.2，中断、页表、文件系统三章还没动——今年立的 flag，明年得兑现
2. 跨链追踪器的两处隐患：分片缺失查询的边界与 Commit 的幂等，都还悬着
3. 交易系统跑完了回测与四季实测，但没上过真钱。纸面和实盘之间还差一层滑点与心态
4. "六百万 TPS"是别人的架构，不是我的。明年得有一个自己压出来的数字

---

*最后更新：2026-10-24*
