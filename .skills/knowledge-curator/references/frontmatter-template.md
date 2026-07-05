---
title: {知识标题}
created: {YYYY-MM-DD}              # 必填：创建日期
updated: {YYYY-MM-DD}              # 必填：最后更新日期
domain: tech                        # 必填：tech / engineering / thinking / domain / humanities / tools / career
tags:                               # 3-6 个
  - {标签1}
  - {标签2}
difficulty: intermediate            # beginner / intermediate / advanced / expert
summary: "{一句话总结，不超过 80 字}"
source:                             # 可选，关联的来源
  - type: blog                      # blog / book / project / course / experience
    url: /blog/posts/2025/01/19/
    title: "负载均衡设计思路与实践指南"
related:                            # 可选，关联的其他知识
  - /knowledge/tech/languages/go/
---

<!-- 重要：知识库内容的编写风格必须基于 blog-writer 技能中的 Markdown 规范，特别是列表序号需要换行的要求 -->

# {知识标题}

> {一句话总结，同 summary}

## 背景

{为什么需要这个知识？在什么场景下会用到？2-3 段。}

## 核心内容

{知识的主体内容，按逻辑组织。}

### {子主题 1}

{内容}

### {子主题 2}

{内容}

## 示例

{具体的代码示例、操作步骤或场景案例。可选。}

```go
// 代码示例
```

## 注意事项

{容易踩的坑、常见误区、边界条件。可选。}

- ⚠️ {需要注意的事项}
- ✅ {推荐做法}

## 延伸阅读

- [{相关文档/文章标题}]({链接})
- 博客：[{文章标题}]({链接})

---

*维护人：yiiewang · 最后更新：{YYYY-MM-DD}*
```

## 字段填写说明

### 必填字段
- `title` —— 文件标题，也是导航显示
- `created` / `updated` —— 首次创建与最后更新
- `domain` —— 7 选 1
- `tags` —— 3-6 个关键词
- `difficulty` —— 4 选 1，AI 评估目标读者水平
- `summary` —— 80 字内的核心要点

### 可选字段
- `source` —— 关联的博客/书/项目/课程/经验
- `related` —— 站内相关知识（用相对路径）

### `domain` 严格取值
```
tech | engineering | thinking | domain | humanities | tools | career
```

### `difficulty` 严格取值
```
beginner | intermediate | advanced | expert
```

### `tags` 原则
- 3-6 个，覆盖：技术栈 / 核心概念 / 应用场景
- 用英文或简短中文：`Go` / `分布式锁` / `consensus` / `Raft`
- 同一概念在不同文件中标签一致

### `source` 写法
```yaml
source:
  - type: blog           # 类型
    url: /blog/posts/2025/01/19/   # 站内用相对路径
    title: "负载均衡设计思路与实践指南"
  - type: book
    title: "数据密集型应用系统设计"
    author: "Martin Kleppmann"
```

### `related` 写法
```yaml
related:
  - /knowledge/tech/languages/go/go-concurrency/
  - /knowledge/tech/architecture/distributed/load-balancing/
```

### ⚠️ 不要写 `status` 字段
Material 主题会把 `status` 渲染成导航中的空圆圈（之前修复过）。如需在内容页展示状态，用内容里的色块/标签，**不要**写在 frontmatter。

### 📝 Markdown 编写规范
知识库内容必须遵循 blog-writer 技能的 Markdown 编写规范：
- 列表序号后必须换行，避免渲染到同一行
- 代码块必须带语言标签
- 段落控制在 2-4 句，段间空一行
- 技术名词保留英文原文
- **Material for MkDocs 增强语法**：可使用警告框、选项卡、数学公式、图表等丰富样式
