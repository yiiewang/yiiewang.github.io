# 知识库检索指南

> AI 在工作流 B（检索）时，遵循本指南的查询语法与最佳实践。

## 一、检索语法

### 关键词组合

支持以下方式（可组合）：

| 方式 | 语法 | 示例 |
|------|------|------|
| 中文关键词 | 直接给中文词 | "分布式锁" |
| 英文关键词 | 直接给英文 | "consensus" |
| 多关键词 | 空格分隔 | "Go 并发 channel" |
| 域限定 | `domain:tech` | "domain:tech 分布式" |
| 子域限定 | `in:tech/architecture` | "in:tech/architecture 负载均衡" |
| 难度限定 | `difficulty:advanced` | "difficulty:advanced 一致性" |
| 标签 | `tag:Go` | "tag:Go" |

## 二、检索流程

### 浅层（默认）

```
1. 解析查询 → 关键词列表
2. Grep 扫 docs/knowledge/ 下所有 .md 的 frontmatter
   - 命中 title / tags / summary
3. 对命中文件，正文只读前 300 字
4. 按 domain 优先级 + updated 时间倒序返回
```

### 深度（用户追问时）

```
1. 浅层结果 → 用户说"详细说" / "没找到" / "再找找"
2. 用 Read 工具读命中文件全文
3. 跨文件交叉引用（related 字段）
4. 综合答案 + 引用每条来源
```

## 三、结果展示格式

### 浅层结果

```
找到 N 条相关知识：

[tech/architecture/distributed/load-balancing]
  标题: 负载均衡策略与实现
  domain: tech/architecture/distributed
  难度: intermediate
  标签: Go, 分布式系统, 负载均衡
  摘要: 4 种负载均衡策略对比 + Go 实现
  updated: 2025-01-19
  路径: /knowledge/tech/architecture/distributed/load-balancing/

[thinking/effective-reporting]
  标题: 高效汇报方法论
  domain: thinking
  难度: beginner
  ...
```

### 深度回答

```
# 关于「分布式锁」的综合回答

## 核心要点（来自 N 条知识）

[综合分析...]

## 引用来源

1. /knowledge/tech/architecture/distributed/load-balancing
   > 关键引用片段...
2. /knowledge/tech/languages/go/go-concurrency
   > 关键引用片段...
```

## 四、检索策略

### 中文为主

- 用户用中文问 → 优先匹配中文关键词
- 同时尝试英文同义词（Go / golang、分布式 / distributed）
- 反之亦然

### 同义词扩展

| 中文 | 英文同义词 |
|------|-----------|
| 分布式锁 | distributed lock |
| 共识算法 | consensus algorithm |
| 幂等 | idempotency / idempotent |
| 负载均衡 | load balancing |
| 高可用 | HA / high availability |
| 事务 | transaction |
| 索引 | index |
| 缓存 | cache |
| 容器化 | containerization |
| 微服务 | microservice |

### 反向检索

当用户问"我有没有写过 X" / "X 这块我有总结吗"：
- 直接全库扫 `tags`、`title`
- 也扫 `summary` 字段
- 找不到时明确说"知识库未收录"

## 五、无结果处理

### 步骤 1：扩展检索

- 拆词、换同义词
- 用更宽泛的关键词
- 尝试关联域

### 步骤 2：明确告知

```
未在知识库中找到「XXX」相关内容。
可能原因：
- 该主题尚未沉淀到知识库
- 用词不同（已尝试同义词：...）
- 内容在博客中而非知识库（要不要我搜博客？）

是否要我把这段内容沉淀为新知识？
```

### 步骤 3：建议沉淀

如果用户给了具体内容 → 直接转工作流 A。
如果用户只是问"有没有" → 等用户确认要不要新建。

## 六、性能注意

- **不要**全量读所有知识库文件（数量增长后会很慢）
- **优先**用 Grep 扫 frontmatter（不读全文）
- **只有**用户确认要详细内容时才 Read
- 维护一个**轻量级索引文件**（可选）记录每条知识的关键词

## 七、检索结果质量

### 优秀结果
- 命中 ≥ 3 条
- 跨域覆盖（不只是 tech）
- 有 difficulty 梯度
- 关联清晰（related 字段串联）

### 改进结果
- 只命中 1-2 条 → 提示用户"知识库覆盖较少，要不要补全？"
- 命中过多（> 10）→ 进一步限定（domain / difficulty / tag）
- 命中但内容不相关 → 检查是否同义词误判

## 八、典型查询模式

| 用户说 | 检索策略 |
|--------|---------|
| "我有没有写过 X" | 全库扫 title / tags / summary |
| "X 怎么做的" | 浅层检索 + top 3 结果 |
| "详细说说 X" | 浅层 → 深度 |
| "X 相关的还有哪些" | 浅层 → 顺着 related 字段扩展 |
| "对比 X 和 Y" | 分别检索 X 和 Y，综合对比 |
| "最新的知识" | 按 updated 时间倒序展示 |
| "XX 域都有啥" | 列子域 + 各 1-2 个代表条目 |
