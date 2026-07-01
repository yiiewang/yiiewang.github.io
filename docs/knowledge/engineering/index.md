---
title: 工程实践
status: effective
domain: engineering
summary: "项目复盘、踩坑记录、性能优化、代码审查 —— 从真实项目中提取的可复用经验"
---

<style>
.eng-summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  margin: 2rem 0;
}
.eng-summary-card {
  background: var(--md-default-bg-color);
  border-radius: 12px;
  border: 1px solid var(--md-default-fg-color--lightest);
  padding: 1.2rem;
  text-align: center;
}
.eng-summary-card .num {
  font-size: 2rem;
  font-weight: 800;
  color: #ed8936;
  line-height: 1.2;
}
.eng-summary-card .label {
  font-size: 0.82rem;
  color: var(--md-default-fg-color--light);
  margin-top: 0.3rem;
}

.eng-domain-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 1rem;
  margin-bottom: 1.5rem;
}
.eng-domain-group {
  background: var(--md-default-bg-color);
  border-radius: 12px;
  border: 1px solid var(--md-default-fg-color--lightest);
  overflow: hidden;
}
.eng-domain-group-header {
  padding: 0.9rem 1.2rem;
  font-weight: 700;
  font-size: 0.95rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  border-bottom: 1px solid var(--md-default-fg-color--lightest);
}
.eng-domain-group-header.orange { background: #fffaf0; color: #c05621; }

.eng-domain-items { padding: 0.4rem 0; }
.eng-domain-item {
  display: flex;
  align-items: center;
  padding: 0.55rem 1.2rem;
  gap: 0.6rem;
  text-decoration: none;
  color: inherit;
  transition: background 0.15s;
}
.eng-domain-item:hover { background: var(--md-default-fg-color--lightest); }
.eng-domain-item-name { flex: 1; font-size: 0.88rem; font-weight: 500; }
.eng-domain-item-count {
  font-size: 0.72rem;
  padding: 0.1rem 0.5rem;
  border-radius: 8px;
  font-weight: 600;
}
.eng-domain-item-count.has-content { background: #c6f6d5; color: #22543d; }
.eng-domain-item-count.empty        { background: var(--md-default-fg-color--lightest); color: var(--md-default-fg-color--light); }

@media (max-width: 768px) {
  .eng-domain-grid { grid-template-columns: 1fr; }
  .eng-summary { grid-template-columns: repeat(2, 1fr); }
}
</style>

# 🔧 工程实践

> 技术能力告诉你「怎么做」，工程实践告诉你「怎么做得更好」。这里的每条知识都来自真实项目的血与泪。

<div class="eng-summary">
  <div class="eng-summary-card">
    <div class="num">0</div>
    <div class="label">已填充条目</div>
  </div>
  <div class="eng-summary-card">
    <div class="num">4</div>
    <div class="label">子域数量</div>
  </div>
  <div class="eng-summary-card">
    <div class="num">30+</div>
    <div class="label">可提炼博客</div>
  </div>
  <div class="eng-summary-card">
    <div class="num">🟡</div>
    <div class="label">建设状态</div>
  </div>
</div>

## 子域导航

<div class="eng-domain-grid">

<div class="eng-domain-group">
  <div class="eng-domain-group-header orange">📋 项目复盘</div>
  <div class="eng-domain-items">
    <a href="retrospectives/" class="eng-domain-item">
      <span class="eng-domain-item-name">项目回顾与总结</span>
      <span class="eng-domain-item-count empty">0</span>
    </a>
  </div>
</div>

<div class="eng-domain-group">
  <div class="eng-domain-group-header orange">🐛 踩坑记录</div>
  <div class="eng-domain-items">
    <a href="pitfalls/" class="eng-domain-item">
      <span class="eng-domain-item-name">调试与故障排查</span>
      <span class="eng-domain-item-count empty">0</span>
    </a>
  </div>
</div>

<div class="eng-domain-group">
  <div class="eng-domain-group-header orange">⚡ 性能优化</div>
  <div class="eng-domain-items">
    <a href="performance/" class="eng-domain-item">
      <span class="eng-domain-item-name">瓶颈定位与优化方法论</span>
      <span class="eng-domain-item-count empty">0</span>
    </a>
  </div>
</div>

<div class="eng-domain-group">
  <div class="eng-domain-group-header orange">✅ 代码审查</div>
  <div class="eng-domain-items">
    <a href="code-review/" class="eng-domain-item">
      <span class="eng-domain-item-name">审查清单与最佳实践</span>
      <span class="eng-domain-item-count empty">0</span>
    </a>
  </div>
</div>

</div>

## 💡 这个域的价值

工程实践是最容易被忽略的知识类型 —— 因为它不像技术文章那样「有产出感」。但恰恰是这些经验，决定了一个工程师的下限。这个域的内容主要来源于：

- 博客中标记为「技术实践」「踩坑」的文章
- 项目中的 Post-mortem 文档
- 日常开发中的随手记录

---

*维护人：yiiewang · 最后更新：2026-07-01*
