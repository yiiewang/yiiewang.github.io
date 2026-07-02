---
name: knowledge-curator
description: >
  知识库维护技能，用于内容归档与检索。
  写入触发：(1) 将内容沉淀到知识库, (2) 整理博客到知识库, (3) 归档会议结论,
  (4) 新建知识条目。检索触发：(1) 在知识库搜索某主题, (2) 查找已有知识,
  (3) 浏览某领域内容。覆盖 7 大知识域（tech/engineering/thinking/domain/humanities/tools/career）。
---

# Skill: knowledge-curator

> **目标**：让 AI 拿到**任意内容**（博客、对话、笔记、文件），自动**分类、提炼、归档**到正确子域，并能**检索**整个知识库。

## 触发场景

### 写入类
- "把这段内容沉淀到知识库"
- "整理这篇博客到知识库"
- "X 这块知识我之前没记录过，帮我建一条"
- "把这次会议结论归档"

### 检索类
- "在知识库里搜 X"
- "我之前写过分布式锁的知识在哪"
- "看看知识库里关于 Go 并发有哪些"
- "我有没有总结过 PBFT？"

---

## 工作流 A：写入（输入 → 归档）

### 1. 加载规则文件
- `references/domain-map.md` —— 七大域 / 30+ 子域的完整映射
- `references/classification-rules.md` —— 自动归类规则
- `references/frontmatter-template.md` —— 标准 frontmatter

### 2. 判定 domain + subdomain

按 `references/classification-rules.md` 流程判定：
1. **关键词扫描**：标题/正文里的技术词
2. **内容类型识别**：代码块？故事？方法论？
3. **域名优先级**：冲突时按 tech > engineering > thinking > domain > humanities > tools > career
4. **不确定时反问用户**

### 3. 检查子域是否存在
- 子域目录已存在 → 走"追加新条目"流程
- 子域目录不存在 → **询问用户**是否新建子目录（已确认允许自由新建）

### 4. 检查文件冲突
- 同子域下已有同名 `.md` 文件 → 询问用户：覆盖 / 追加 / 改名 / 跳过

### 5. 填充 frontmatter
按 `references/frontmatter-template.md` 标准填写：
```yaml
title: ...
created: YYYY-MM-DD    # 今天
updated: YYYY-MM-DD    # 今天
domain: ...
tags: [...]
difficulty: beginner|intermediate|advanced|expert   # AI 评估
summary: "..."         # 一句话，不超过 80 字
source:                # 可选，关联的博客/书/项目
  - type: blog
    url: /blog/posts/2025/01/19/
    title: "负载均衡设计思路与实践指南"
related:               # 可选
  - /knowledge/tech/languages/go/
```

### 6. 提炼正文（关键！）

**不是直接复制来源内容**，而是要：
- 去掉时间线叙事（"今天我遇到了..."）
- 去掉情感表达（"非常有意思..."）
- 去掉铺垫场景
- **保留**：核心定义、原理、代码、表格、列表
- **重组**：按"背景 → 核心 → 示例 → 注意"的**知识结构**重排
- 控制在 1500-3000 字

### 7. 写入文件
- 路径：`docs/knowledge/{domain}/{subdomain}/{filename}.md`
- 文件名：小写英文 + 连字符（`load-balancing.md` / `pbft.md`）
- 同主题多文件加后缀：`pbft.md` / `raft.md` / `pow.md`

### 8. 更新索引页（可选但推荐）
- 若新条目所属子域有 `index.md` → **追加一行链接 + 更新数量 badge**
- 若该子域原本无 `index.md` → **新建**（按统一风格）

### 9. 完成报告
- 新文件路径
- 所属域 / 子域
- 关联的来源（博客 URL 等）
- 反向链接：知识库首页统计数字（可选）

---

## 工作流 B：检索（提问 → 答案）

### 1. 浅层检索（默认）

读取 `references/retrieval-guide.md` 后执行：

1. **解析查询**：提取关键词（中英文都要）
2. **frontmatter 索引**：用 Grep 扫 `docs/knowledge/` 下所有 `title:`、`tags:`、`summary:`
3. **正文前 300 字**：命中 frontmatter 后再看正文首段
4. **结果排序**：按 domain 优先级 + 时间倒序

返回格式：
```
找到 N 条相关知识：

[tech/architecture/distributed/load-balancing.md]
  标题: 负载均衡策略与实现
  标签: Go, 分布式系统, 负载均衡
  摘要: 4 种负载均衡策略对比 + Go 实现
  路径: /knowledge/tech/architecture/distributed/load-balancing/

[thinking/effective-reporting.md]
  ...
```

### 2. 深度检索（用户追问时）

当用户要求"详细说"或浅层结果不足时：
- **读全文**：用 Read 工具读命中的 `.md` 文件
- **语义匹配**：理解上下文，给出综合答案
- **跨域串联**：如果涉及多个子域，列出相关知识形成知识图

### 3. 无结果处理
- 明确告知"知识库未收录"
- 建议是否要**当场沉淀**（转工作流 A）

---

## 关键决策点

### 子域创建
- **允许 AI 自由新建子域**（用户已确认）
- 命名规则：小写英文 + 连字符（`kafka-basics` / `system-trading`）
- 新建子域时**同步新建 `index.md`**（按统一风格）

### 与博客的边界
- **博客**：时间线叙事 + 个人体验 + 详细过程
- **知识库**：结构化沉淀 + 抽象原理 + 可直接复用

### 状态管理
- `status` 字段**不要**写在 frontmatter（Material 主题会渲染成空圆圈）
- 状态信息**只在内容页**用色块/标签展示
- 如果需要快速区分"是否已审核"等元信息 → 用 `difficulty` 或新增自定义字段

### 知识库不写的内容
- ❌ 时间相关的"今天"（已发生的事）
- ❌ 个人情感表达
- ❌ 与特定博客的强引用（用 `source` 字段关联即可）
- ❌ 过度铺垫的"引子"

---

## 完成标志

### 写入工作流
- [ ] 文件已创建在 `docs/knowledge/...`
- [ ] frontmatter 完整
- [ ] 子域目录已创建（含 `index.md`）
- [ ] 索引页已更新（如适用）
- [ ] 关联的 source 字段已填（如有来源）
- [ ] 向用户报告新文件路径 + domain + summary

### 检索工作流
- [ ] 关键词已正确解析
- [ ] 已浅层扫 frontmatter
- [ ] 必要时做了深度读全文
- [ ] 结果已按 domain 优先级 + 时间排序
- [ ] 无结果时明确告知 + 建议沉淀

---

## 辅助文件清单

| 文件 | 用途 |
|------|------|
| `references/domain-map.md` | 7 大域 / 30+ 子域完整映射表 |
| `references/frontmatter-template.md` | frontmatter 标准格式 |
| `references/classification-rules.md` | 自动归类规则 + 冲突处理 |
| `references/integration-workflow.md` | 索引页自动更新规则 |
| `references/retrieval-guide.md` | 检索语法 + 最佳实践 |
