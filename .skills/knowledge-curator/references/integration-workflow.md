# 索引页更新与子域集成

> AI 写入新知识条目后，**需要**更新相关的索引页以保持站点一致。本文件定义更新规则。

## 一、需要更新的索引页层级

新条目 `docs/knowledge/tech/languages/go/go-concurrency.md` 写入后，需要更新：

| 层级 | 文件 | 更新内容 |
|------|------|---------|
| 子域 | `docs/knowledge/tech/languages/go/index.md` | 追加条目链接 + 数量 badge |
| 域 | `docs/knowledge/tech/index.md` | 更新子域的数量 badge |
| 知识库首页 | `docs/knowledge/index.md` | 更新"知识条目"统计数字 |

## 二、子域 `index.md` 更新规则

### 文件结构（标准）

```markdown
---
title: {子域名}
domain: {domain}
summary: "{子域一句话描述}"
---

# {子域名}

> {子域简介}

## 📖 知识条目

| 条目 | 难度 | 简介 |
|------|------|------|
| [{条目标题}]({文件名}.md) | {难度} | {一句话简介} |

## 🔗 相关资源

- [返回上级](../index.md)
- [知识库首页](/knowledge/)

---

*维护人：yiiewang · 最后更新：{YYYY-MM-DD}*
```

### 新增条目时

1. 在"知识条目"表格**追加一行**（按时间倒序，最新在上）
2. 评估并填写 `难度` 字段（beginner/intermediate/advanced/expert）
3. 一句话简介（不超过 30 字）
4. **更新底部"最后更新"日期**

### 模板代码

```markdown
## 📖 知识条目

| 条目 | 难度 | 简介 |
|------|------|------|
| [{新条目标题}]({新文件名}.md) | {难度} | {简介} |
| [{已有条目标题}]({已有文件名}.md) | {难度} | {简介} |
```

## 三、域 `index.md` 更新规则

域索引页有 **数量 badge** 标记每个子域的内容数。

### 更新 badge

新条目落入 `tech/languages/go/` 后，`tech/index.md` 中：
- Go 语言子域的 badge 从 `1` → `2`
- 如有"编程语言"分组的统计，也要更新

### 文件位置
- `docs/knowledge/{domain}/index.md`

### 参考现有结构
每个域 index.md 都有按子域分组的卡片，详见 `tech/index.md` 范例。

## 四、知识库首页更新

`docs/knowledge/index.md` 顶部 Hero 区有统计数字：

```html
<div class="kb-hero-stat">
  <div class="kb-hero-stat-num">4</div>   ← 知识条目总数
  <div class="kb-hero-stat-label">知识条目</div>
</div>
```

新条目写入后，**更新这个数字**（+1）。

## 五、特殊情况

### 1. 新建子域（子域 index.md 不存在）

创建子域目录后，**新建 index.md**：

1. 按标准结构创建（见上）
2. 在域 index.md 中**追加子域链接**
3. 在知识库首页（如有相应展示位）追加

### 2. 已有同名文件

**不直接覆盖**。询问用户：
```
docs/knowledge/tech/languages/go/{filename}.md 已存在。

请选择：
- 覆盖（用新内容替换）
- 追加（保留旧内容，在末尾追加新章节）
- 重命名（新文件名，如 {filename}-v2.md）
- 跳过（不写入）
```

### 3. 跨域条目（一条内容应入多个子域）

- **默认不重复归档**。在主条目里加 `related` 字段指向其他相关条目即可
- 例外：如果两个子域的视角完全不同（如 tech 讲原理、engineering 讲踩坑），**可以**拆为两条独立知识条目

## 六、自动化建议

每次写入新条目时，AI 跑一遍：

```python
# 伪代码
def on_knowledge_created(filepath):
    # 1. 解析 path
    domain, subdomain = parse_path(filepath)
    
    # 2. 更新子域 index.md
    update_subdomain_index(domain, subdomain, filepath)
    
    # 3. 更新域 index.md 数量
    update_domain_index(domain, subdomain)
    
    # 4. 更新知识库首页统计
    update_kb_homepage_stats()
    
    # 5. 报告
    print(f"已更新 {len(updated_files)} 个索引文件")
```

## 七、原子化原则

- 每次只更新**必要的**索引页
- 不要触发整站重建
- 不要修改无关文件
- 修改前用 Read 读一次（避免覆盖他人改动）
