---
title: 思维方法
domain: thinking
summary: "学习方法论、问题拆解、决策框架、系统思维 —— 元能力的系统化"
---

<style>
.thk-summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  margin: 2rem 0;
}
.thk-summary-card {
  background: var(--md-default-bg-color);
  border-radius: 12px;
  border: 1px solid var(--md-default-fg-color--lightest);
  padding: 1.2rem;
  text-align: center;
}
.thk-summary-card .num {
  font-size: 2rem;
  font-weight: 800;
  color: #48bb78;
  line-height: 1.2;
}
.thk-summary-card .label {
  font-size: 0.82rem;
  color: var(--md-default-fg-color--light);
  margin-top: 0.3rem;
}

.thk-domain-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 1rem;
  margin-bottom: 1.5rem;
}
.thk-domain-group {
  background: var(--md-default-bg-color);
  border-radius: 12px;
  border: 1px solid var(--md-default-fg-color--lightest);
  overflow: hidden;
}
.thk-domain-group-header {
  padding: 0.9rem 1.2rem;
  font-weight: 700;
  font-size: 0.95rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  border-bottom: 1px solid var(--md-default-fg-color--lightest);
}
.thk-domain-group-header.green { background: #f0fff4; color: #22543d; }

.thk-domain-items { padding: 0.4rem 0; }
.thk-domain-item {
  display: flex;
  align-items: center;
  padding: 0.55rem 1.2rem;
  gap: 0.6rem;
  text-decoration: none;
  color: inherit;
  transition: background 0.15s;
}
.thk-domain-item:hover { background: var(--md-default-fg-color--lightest); }
.thk-domain-item-name { flex: 1; font-size: 0.88rem; font-weight: 500; }
.thk-domain-item-count {
  font-size: 0.72rem;
  padding: 0.1rem 0.5rem;
  border-radius: 8px;
  font-weight: 600;
}
.thk-domain-item-count.has-content { background: #c6f6d5; color: #22543d; }
.thk-domain-item-count.empty        { background: var(--md-default-fg-color--lightest); color: var(--md-default-fg-color--light); }

@media (max-width: 768px) {
  .thk-domain-grid { grid-template-columns: 1fr; }
  .thk-summary { grid-template-columns: repeat(2, 1fr); }
}
</style>

# 🧠 思维方法

> 如果说技术能力是「术」，思维方法就是「道」。这些能力不绑定任何具体技术，但影响你在任何领域的学习、决策和创造效率。

<div class="thk-summary">
  <div class="thk-summary-card">
    <div class="num">1</div>
    <div class="label">已填充条目</div>
  </div>
  <div class="thk-summary-card">
    <div class="num">4</div>
    <div class="label">子域数量</div>
  </div>
  <div class="thk-summary-card">
    <div class="num">🟡</div>
    <div class="label">建设状态</div>
  </div>
  <div class="thk-summary-card">
    <div class="num">—</div>
    <div class="label">向内挖掘型</div>
  </div>
</div>

## 子域导航

<div class="thk-domain-grid">

<div class="thk-domain-group">
  <div class="thk-domain-group-header green">📖 学习方法论</div>
  <div class="thk-domain-items">
    <a href="learning/" class="thk-domain-item">
      <span class="thk-domain-item-name">费曼学习法 · 间隔重复 · 主题阅读</span>
      <span class="thk-domain-item-count empty">0</span>
    </a>
  </div>
</div>

<div class="thk-domain-group">
  <div class="thk-domain-group-header green">🧩 问题拆解</div>
  <div class="thk-domain-items">
    <a href="problem-solving/" class="thk-domain-item">
      <span class="thk-domain-item-name">分解 · 定位 · 逐层击破</span>
      <span class="thk-domain-item-count empty">0</span>
    </a>
  </div>
</div>

<div class="thk-domain-group">
  <div class="thk-domain-group-header green">⚖️ 决策框架</div>
  <div class="thk-domain-items">
    <a href="decision-making/" class="thk-domain-item">
      <span class="thk-domain-item-name">技术选型 · 架构决策 · 优先级排序</span>
      <span class="thk-domain-item-count empty">0</span>
    </a>
  </div>
</div>

<div class="thk-domain-group">
  <div class="thk-domain-group-header green">🌐 系统思维</div>
  <div class="thk-domain-items">
    <a href="systems-thinking/" class="thk-domain-item">
      <span class="thk-domain-item-name">整体视角 · 反馈循环 · 涌现行为</span>
      <span class="thk-domain-item-count empty">0</span>
    </a>
  </div>
</div>

</div>

## 📝 已收录条目

<div class="thk-domain-grid">

<div class="thk-domain-group">
  <div class="thk-domain-group-header green">📄 独立条目</div>
  <div class="thk-domain-items">
    <a href="effective-reporting.md" class="thk-domain-item">
      <span class="thk-domain-item-name">结构化汇报方法论</span>
      <span class="thk-domain-item-count has-content">1</span>
    </a>
  </div>
</div>

</div>

## 💡 这个域的特殊之处

思维方法类知识不容易从博客直接提炼——它更多体现在「怎么写」和「为什么写」之中。这个域的建设方式会比较不同：

- 📝 **向内挖掘**：回顾自己做过的重要决策和思考过程，试着把它们写下来
- 📚 **向外借鉴**：从经典书籍（如《系统之美》《思考，快与慢》）中提取框架
- 🔄 **迭代完善**：这些条目会随着认知升级不断重写

---

*维护人：yiiewang · 最后更新：2026-07-01*
