# 知识库域映射表

> 7 大域 / 30+ 子域完整结构。每个子域附带**关键词索引**，AI 分类时按关键词命中度排序。

## 顶层结构

```
docs/knowledge/
├── tech/            # 技术能力
├── engineering/     # 工程实践
├── thinking/        # 思维方法
├── domain/          # 领域知识
├── humanities/      # 人文素养
├── tools/           # 工具与效率
└── career/          # 职业发展
```

## 一、tech（技术能力）— 7 个子域

| 子域路径 | 中文名 | 关键词 |
|---------|--------|--------|
| `tech/languages/` | 编程语言 | Go, Java, Python, Rust, TypeScript, 语法, 类型系统 |
| `tech/blockchain/` | 区块链 | 共识算法, 智能合约, ChainMaker, 分布式账本, 区块链架构 |
| `tech/architecture/` | 架构设计 | 系统设计, 设计模式, 分布式系统, 微服务, 架构 |
| `tech/cloud-native/` | 云原生 | Docker, Kubernetes, CI/CD, 容器, Service Mesh |
| `tech/network/` | 网络协议 | TCP, HTTP, gRPC, RPC, 网络模型, 七层协议 |
| `tech/storage/` | 数据存储 | 关系型数据库, NoSQL, 缓存, 索引, 事务, 一致性 |
| `tech/testing/` | 测试工程 | 单元测试, 集成测试, Mock, 性能测试, TDD |

### 已有内容参考
- `tech/languages/go/go-concurrency.md` — Go 并发 15 反模式
- `tech/blockchain/consensus/pbft.md` — PBFT 三阶段协议
- `tech/architecture/distributed/load-balancing.md` — 负载均衡
- `tech/architecture/system-design/idempotency.md` — 幂等设计

## 二、engineering（工程实践）— 4 个子域

| 子域路径 | 中文名 | 关键词 |
|---------|--------|--------|
| `engineering/retrospectives/` | 项目复盘 | 项目总结, Postmortem, 复盘, 总结 |
| `engineering/pitfalls/` | 踩坑记录 | 踩坑, 调试, Bug, 问题排查, 故障 |
| `engineering/performance/` | 性能优化 | 性能, 调优, profiling, 瓶颈, 优化 |
| `engineering/code-review/` | 代码审查 | Code Review, CR, 审查清单, 最佳实践 |

## 三、thinking（思维方法）— 4 个子域

| 子域路径 | 中文名 | 关键词 |
|---------|--------|--------|
| `thinking/learning/` | 学习方法论 | 学习方法, 费曼学习法, 间隔重复, 主题阅读 |
| `thinking/problem-solving/` | 问题拆解 | 拆解, 定位, 分析, 复杂问题 |
| `thinking/decision-making/` | 决策框架 | 决策, 选型, 优先级, 权衡 |
| `thinking/systems-thinking/` | 系统思维 | 系统, 反馈, 涌现, 整体观 |

### 已存在
- `thinking/effective-reporting.md` — 高效汇报方法论（独立条目）

## 四、domain（领域知识）— 3 个子域

| 子域路径 | 中文名 | 关键词 |
|---------|--------|--------|
| `domain/blockchain-industry/` | 区块链行业 | 行业生态, 政策, 监管, 趋势, 应用场景 |
| `domain/finance/` | 金融与投资 | 投资, 资产配置, 股票, 基金, 市场分析 |
| `domain/economics/` | 经济学 | 宏观经济, 微观经济, 货币, 通胀, 政策 |

## 五、humanities（人文素养）— 3 个子域

| 子域路径 | 中文名 | 关键词 |
|---------|--------|--------|
| `humanities/reading/` | 读书笔记 | 书评, 读书, 笔记, 经典 |
| `humanities/writing/` | 写作 | 写作方法, 表达, 排版, 习惯 |
| `humanities/philosophy/` | 哲学思考 | 哲学, 思辨, 伦理, 自由, 责任 |

## 六、tools（工具与效率）— 3 个子域

| 子域路径 | 中文名 | 关键词 |
|---------|--------|--------|
| `tools/dev-tools/` | 开发工具 | IDE, Vim, VSCode, 命令行, Git, 调试 |
| `tools/productivity/` | 效率方法论 | GTD, 番茄工作法, 深度工作, 知识管理 |
| `tools/automation/` | 自动化脚本 | Shell, Python 脚本, CI 自动化, 批处理 |

## 七、career（职业发展）— 3 个子域

| 子域路径 | 中文名 | 关键词 |
|---------|--------|--------|
| `career/interview/` | 面试准备 | 面试题, 系统设计面试, 行为面试, 算法面试 |
| `career/soft-skills/` | 软技能 | 沟通, 协作, 向上管理, 演讲, 汇报 |
| `career/tech-leadership/` | 技术管理 | Tech Lead, 团队建设, 技术规划, 架构师 |

---

## 分类优先级（冲突时使用）

```
tech > engineering > thinking > domain > humanities > tools > career
```

### 判定逻辑
- 内容**主要讲技术原理/代码/算法** → tech
- 内容**讲项目经验/事故复盘/排查过程** → engineering
- 内容**讲方法论/思维框架** → thinking
- 内容**讲行业洞察/经济金融** → domain
- 内容**讲读书/哲学/写作** → humanities
- 内容**讲工具使用/效率技巧** → tools
- 内容**讲面试/管理/职业** → career

### 跨域内容处理
- 一篇内容跨多个域 → **拆分为多条知识**，每条聚焦一个主题
- 例：Go 性能调优（讲 pprof 工具）→ tech/tools 各一条

---

## 子域创建规则

**允许 AI 自由新建子域**（用户已确认）。

### 命名规范
- 全小写英文 + 连字符：`kafka-basics` / `system-trading` / `design-thinking`
- 避免缩写（除非是约定俗成的：ci/cd、tdd）
- 不超过 3 个单词

### 新建流程
1. 创建目录 `docs/knowledge/{domain}/{subdomain}/`
2. 创建 `index.md`（按统一风格，见 `integration-workflow.md`）
3. 通知用户：已新建子域 + 原因
4. 后续内容直接落入

### 不建议新建的情况
- 内容可归入已有子域（即使不太完美）
- 子域名与已有概念重叠
- 只是临时分类、未来可能废弃
