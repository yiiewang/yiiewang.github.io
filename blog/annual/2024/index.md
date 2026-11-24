# 2024 年度集锦

> 上半年在修工具，下半年在写论文、看市场、造轮子。全年 26 篇，但下半年每篇都比上半年重。

## 卷首语

2024 有一条很清楚的分界线，就在九月。

九月之前写的是[Code Server 部署](<https://cloaks.cn/blog/posts/2024/02/27-1/index.md>)、[Shell 循环](<https://cloaks.cn/blog/posts/2024/06/27/index.md>)、 [GitLab CI/CD](<https://cloaks.cn/blog/posts/2024/07/09/index.md>)、[数位板配置](<https://cloaks.cn/blog/posts/2024/08/11/index.md>)—— 全是"把一个东西装好、跑通"的活，写下来主要为了下次不用再查。

九月之后，A 股那波大涨把市场的注意力拽了过去，[爱在深秋](<https://cloaks.cn/blog/posts/2024/10/02-1/index.md>)是转折点。 紧接着是四篇软考论文，然后是十一、十二月那段最过瘾的：用 Go 造 [BitTorrent 客户端](<https://cloaks.cn/blog/posts/2024/11/13/index.md>)、[不到 100 行的容器](<https://cloaks.cn/blog/posts/2024/11/23/index.md>)、 [从零开始的数据库](<https://cloaks.cn/blog/posts/2024/11/25/index.md>)。

## 主线一：软考架构师论文四篇

十月到十一月，四篇论文一路写下来，题材都来自真实项目：

1. [论软件架构的选择](<https://cloaks.cn/blog/posts/2024/10/02-2/index.md>) — 移动信息化开放平台的三层分层架构
2. [论软件系统架构风格](<https://cloaks.cn/blog/posts/2024/10/04/index.md>) — 同上项目，讲设计与实施
3. [论软件的可靠性设计](<https://cloaks.cn/blog/posts/2024/11/07/index.md>) — 三维轻量化异构项目：容错、N 版本程序设计、冗余
4. [论面向对象设计](<https://cloaks.cn/blog/posts/2024/11/09-1/index.md>) — 同项目，RUP 统一过程模型

> [!TIP]
>
> **这四篇的读法**
>
> 套路是固定的：项目背景 → 理论要点 → 实践落地 → 运行效果。 想快速过一遍软考论文，看第三篇（可靠性设计）最划算。

## 主线二：动手派的两个月

十一月开始，连续四篇"照着做一遍"：

- [用 Go 实现 BitTorrent 客户端](<https://cloaks.cn/blog/posts/2024/11/13/index.md>) — 从 tracker 协议到 peer 握手
- [用不到 100 行 Go 构建自己的容器](<https://cloaks.cn/blog/posts/2024/11/23/index.md>) — namespace + cgroup 的最小实现
- [从头开始构建你自己的数据库](<https://cloaks.cn/blog/posts/2024/11/25/index.md>) — 从 B+ 树往上搭
- [B-Tree 数据结构](<https://cloaks.cn/blog/posts/2024/12/3/index.md>) 与 [B-Tree 与崩溃恢复](<https://cloaks.cn/blog/posts/2024/12/4/index.md>) — 节点结构、分裂合并、写时复制

这条线的共同点是：**不读源码，自己写一遍**。 十二月还补了两篇原理：[版本管理](<https://cloaks.cn/blog/posts/2024/12/10/index.md>)、[GC 垃圾回收算法](<https://cloaks.cn/blog/posts/2024/12/12/index.md>)。

## 主线三：九月底那波行情

- [爱在深秋](<https://cloaks.cn/blog/posts/2024/10/02-1/index.md>) — 上证单日 8.06%、深成指 10.67%，现场记录
- [什么是美林时钟？](<https://cloaks.cn/blog/posts/2024/10/13/index.md>) — 衰退/复苏/过热/滞胀四阶段与资产配置
- [本轮牛市的风险与机会](<https://cloaks.cn/blog/posts/2024/10/16/index.md>) — 三波上涨节奏、资金入场逻辑、行业轮动
- [经济杂谈](<https://cloaks.cn/blog/posts/2024/10/18/index.md>) — 房地产、养老金统筹、央行国债操作的合集

## 支线：工程杂活与数据方法论

- **Go 两篇**：[我对 Context 包的理解](<https://cloaks.cn/blog/posts/2024/04/03/index.md>)、[条件编译实践](<https://cloaks.cn/blog/posts/2024/05/27/index.md>)
- **数据两篇**：[大数据平台层级结构](<https://cloaks.cn/blog/posts/2024/04/12-2/index.md>)（ODS / 数仓 / 应用层）、[指标体系搭建方法论](<https://cloaks.cn/blog/posts/2024/04/12-1/index.md>)（OSM / AARRR / 人货场）
- **工具**：[Navicat 试用期重置](<https://cloaks.cn/blog/posts/2024/06/24/index.md>)
- **管理与格式**：[项目管理必备的 8 个核心文档](<https://cloaks.cn/blog/posts/2024/08/05/index.md>)、[XML 读书笔记](<https://cloaks.cn/blog/posts/2024/07/20/index.md>)

## 代表作五篇

1. [从头开始构建你自己的数据库](<https://cloaks.cn/blog/posts/2024/11/25/index.md>) — 年度最佳。从 B+ 树一路搭到能跑，读完对"数据库"三个字有实感
2. [用不到 100 行 Go 语言构建自己的容器](<https://cloaks.cn/blog/posts/2024/11/23/index.md>) — 把容器的神秘感拆掉了
3. [我对 Context 包的理解](<https://cloaks.cn/blog/posts/2024/04/03/index.md>) — 全年引用率最高的 Go 篇
4. [本轮牛市的风险与机会](<https://cloaks.cn/blog/posts/2024/10/16/index.md>) — 有观点、有节奏判断，不是事后复述
5. [论软件的可靠性设计](<https://cloaks.cn/blog/posts/2024/11/07/index.md>) — 四篇论文里最有干货的一篇

## 全年目录

### 2 月 · 1 篇

- [Code Server 部署指南](<https://cloaks.cn/blog/posts/2024/02/27-1/index.md>) · 02-27 · DevOps 与工具

### 4 月 · 3 篇

- [我对 Context 包的理解](<https://cloaks.cn/blog/posts/2024/04/03/index.md>) · 04-03 · Go 语言
- [指标体系搭建方法论](<https://cloaks.cn/blog/posts/2024/04/12-1/index.md>) · 04-12 · 札记
- [大数据平台层级结构](<https://cloaks.cn/blog/posts/2024/04/12-2/index.md>) · 04-12 · 札记

### 5 月 · 1 篇

- [Go 条件编译实践](<https://cloaks.cn/blog/posts/2024/05/27/index.md>) · 05-27 · Go 语言

### 6 月 · 2 篇

- [Navicat for Linux 试用期重置](<https://cloaks.cn/blog/posts/2024/06/24/index.md>) · 06-24 · DevOps 与工具
- [Shell 脚本编程【循环】](<https://cloaks.cn/blog/posts/2024/06/27/index.md>) · 06-27 · DevOps 与工具

### 7 月 · 2 篇

- [GitLab CI/CD 配置手册](<https://cloaks.cn/blog/posts/2024/07/09/index.md>) · 07-09 · DevOps 与工具
- [XML 读书笔记](<https://cloaks.cn/blog/posts/2024/07/20/index.md>) · 07-20 · 札记

### 8 月 · 2 篇

- [项目管理必备的 8 个核心文档](<https://cloaks.cn/blog/posts/2024/08/05/index.md>) · 08-05 · 札记
- [Gaoman 数位板 M6 Linux 配置](<https://cloaks.cn/blog/posts/2024/08/11/index.md>) · 08-11 · DevOps 与工具

### 10 月 · 6 篇

- [爱在深秋](<https://cloaks.cn/blog/posts/2024/10/02-1/index.md>) · 10-02 · 投资与交易
- [论软件架构的选择](<https://cloaks.cn/blog/posts/2024/10/02-2/index.md>) · 10-02 · 软考
- [论软件系统架构风格](<https://cloaks.cn/blog/posts/2024/10/04/index.md>) · 10-04 · 软考
- [什么是美林时钟？](<https://cloaks.cn/blog/posts/2024/10/13/index.md>) · 10-13 · 投资与交易
- [本轮牛市的风险与机会](<https://cloaks.cn/blog/posts/2024/10/16/index.md>) · 10-16 · 投资与交易
- [经济杂谈](<https://cloaks.cn/blog/posts/2024/10/18/index.md>) · 10-18 · 投资与交易

### 11 月 · 5 篇

- [论软件的可靠性设计](<https://cloaks.cn/blog/posts/2024/11/07/index.md>) · 11-07 · 软考
- [论面向对象设计](<https://cloaks.cn/blog/posts/2024/11/09-1/index.md>) · 11-09 · 软考
- [使用 Go 实现 BitTorrent 客户端](<https://cloaks.cn/blog/posts/2024/11/13/index.md>) · 11-13 · 读书与思考
- [用不到 100 行 Go 语言构建自己的容器](<https://cloaks.cn/blog/posts/2024/11/23/index.md>) · 11-23 · 读书与思考
- [从头开始构建你自己的数据库](<https://cloaks.cn/blog/posts/2024/11/25/index.md>) · 11-25 · 读书与思考

### 12 月 · 4 篇

- [B-Tree 数据结构](<https://cloaks.cn/blog/posts/2024/12/3/index.md>) · 12-03 · 读书与思考
- [B-Tree 与崩溃恢复](<https://cloaks.cn/blog/posts/2024/12/4/index.md>) · 12-04 · 读书与思考
- [版本管理：追溯数据的生长链](<https://cloaks.cn/blog/posts/2024/12/10/index.md>) · 12-10 · Go 语言
- [GC 垃圾回收算法](<https://cloaks.cn/blog/posts/2024/12/12/index.md>) · 12-12 · Go 语言

## 回看：后来怎样了

1. 十一、十二月"自己写一遍"的习惯，2026 年变成了 [xv6 学习历程](<https://cloaks.cn/blog/posts/2026/09/17/index.md>)和[拆库记：LevelDB 分片演进](<https://cloaks.cn/blog/posts/2026/10/20/index.md>)
2. 四篇软考论文是这年最系统的输出，但 2025 年只补了一篇题型总结，2026 年彻底停了
3. 十月的市场随笔还是"观点驱动"，到 2026 年才变成可回测的策略（[设计手记](<https://cloaks.cn/blog/posts/2026/04/30/index.md>) → [实测收官](<https://cloaks.cn/blog/posts/2026/10/21-1/index.md>)）
4. 那条"造轮子"的线没断，只是越造越深

---

*最后更新：2026-10-24*
