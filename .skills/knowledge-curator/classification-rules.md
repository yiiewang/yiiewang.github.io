# 自动归类规则

> AI 拿到内容后，按此规则判定 domain + subdomain。

## 归类流程

```
输入内容（博客/笔记/对话）
       │
       ▼
  ┌─────────────┐
  │ 关键词扫描  │ ──→ 命中 domain-map.md 的关键词
  └─────────────┘
       │
       ▼
  ┌──────────────┐
  │ 内容类型识别 │ ──→ 代码/故事/方法论/工具介绍/...
  └──────────────┘
       │
       ▼
  ┌──────────────┐
  │ domain 判定  │ ──→ 按优先级：tech > engineering > ...
  └──────────────┘
       │
       ▼
  ┌──────────────┐
  │ subdomain 判定 │ ──→ 域内子域匹配
  └──────────────┘
       │
       ▼
  ┌──────────────┐
  │ 冲突处理     │ ──→ 跨域内容拆分 / 不确定反问
  └──────────────┘
       │
       ▼
   输出 domain + subdomain
```

## 一、关键词 → Domain 速查表

| Domain | 强信号关键词 |
|--------|-------------|
| **tech** | 代码块、API、协议、算法、数据结构、架构、框架、库、SDK、中间件、并发、锁、事务、索引、SQL、Linux、命令、版本号 |
| **engineering** | 项目、复盘、Postmortem、踩坑、调试、Bug、故障、监控、告警、上线、压测、优化、定位、复现 |
| **thinking** | 方法论、模型、思维、决策、框架、原则、心得、反思、总结、复盘（注意：项目复盘归 engineering，个人反思归 thinking）|
| **domain** | 行业、监管、政策、市场、投资、股票、基金、宏观经济、加密货币、DeFi、监管沙盒 |
| **humanities** | 读书、书评、哲学、伦理、写作、表达、文学、艺术、思考人生 |
| **tools** | 工具、IDE、Vim、VSCode、命令行、Shell、效率、GTD、番茄、自动化脚本、CI、快捷键 |
| **career** | 面试、简历、职业、管理、Lead、团队、汇报、向上管理、晋升、跳槽 |

## 二、内容类型 → Domain 微调

| 内容类型 | 默认 Domain | 备注 |
|---------|------------|------|
| 算法/协议原理讲解 | tech | 强 |
| 代码片段/示例 | tech | 强 |
| 真实事故排查过程 | engineering | 关键看"过程"而非"原理" |
| 方法论/思维框架 | thinking | 强 |
| 行业分析/趋势 | domain | 强 |
| 工具使用技巧 | tools | 强 |
| 面试题解 | career | 强 |
| 读书笔记 | humanities | 强 |
| 个人反思/随笔 | thinking | 注意区别于工程复盘 |
| 性能优化实战 | engineering | 看是否聚焦"项目过程" |

## 三、域内 Subdomain 判定

确定 domain 后，再按该域的子域关键词做二级匹配。

### tech 子域关键词
- `tech/languages/` —— 出现具体语言名（Go/Java/Python/...）
- `tech/blockchain/` —— 共识/智能合约/账本/链
- `tech/architecture/` —— 架构/系统设计/分布式/微服务/设计模式
- `tech/cloud-native/` —— Docker/K8s/CI/CD
- `tech/network/` —— TCP/HTTP/gRPC/网络
- `tech/storage/` —— DB/缓存/存储/事务
- `tech/testing/` —— 测试/Mock/压测

### engineering 子域关键词
- `engineering/retrospectives/` —— "复盘" + 项目名
- `engineering/pitfalls/` —— "踩坑"/"调试"/"问题排查"
- `engineering/performance/` —— "性能"/"调优"/"瓶颈"
- `engineering/code-review/` —— "Code Review"/"CR"

### thinking 子域关键词
- `thinking/learning/` —— "学习方法"/"费曼"
- `thinking/problem-solving/` —— "问题拆解"
- `thinking/decision-making/` —— "决策"/"选型"
- `thinking/systems-thinking/` —— "系统思维"

### domain / humanities / tools / career
- 参见 `domain-map.md` 的关键词列表

## 四、冲突处理

### 场景 1：内容跨多个域

**默认行为**：拆分。

例：博客《Go 性能优化中 pprof 工具使用》
- pprof 工具本身 → `tools/dev-tools/`
- Go 性能优化思路 → `tech/languages/go/` 或 `engineering/performance/`
- 拆为 2 条知识，AI 提示用户"已拆分为..."

### 场景 2：内容主体不明

**默认行为**：询问用户。

例：标题是"我优化了一个慢查询" —— 既可能是 engineering 复盘（讲排查过程），也可能是 tech 存储（讲 SQL 优化）。AI 反问：
> 这篇主要是讲 **排查过程**（engineering/performance）还是 **SQL 优化原理**（tech/storage）？

### 场景 3：关键词命中多个 subdomain

**默认行为**：按命中度排序 + 询问用户。

例：内容同时命中 `tech/languages/go/` 和 `tech/architecture/distributed/`
- AI 给出 2 个候选 + 简短理由
- 用户选择

### 场景 4：完全找不到合适子域

**默认行为**：反问 + 建议新建。

AI：
> 没找到现成子域匹配「XXX」。建议新建子域 `tech/xxx/`，原因：[理由]
> 确认新建吗？

## 五、判定置信度

对每次判定，AI 内部要给出**置信度**（仅思考用，不展示给用户）：
- **高**：3+ 强信号关键词命中 + 内容类型匹配
- **中**：1-2 关键词命中，需要进一步阅读
- **低**：无明显关键词，必须反问

低置信度时**强制反问用户**。

## 六、反问模板

```
我打算把这段内容归到 [domain]/[subdomain]/，
理由：[1-2 句关键词命中说明]。

如果不对，可选：
- [domain2]/[subdomain2]/：[理由]
- [domain3]/[subdomain3]/：[理由]
- 其他子域：[用户自填]
```

避免无脑反问 —— 至少给出 1 个最佳猜测。
