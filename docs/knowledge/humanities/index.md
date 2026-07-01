---
title: 人文素养
domain: humanities
summary: "读书笔记、写作、哲学思考 —— 技术之外的知识滋养与精神基建"
---

<style>
.hum-summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  margin: 2rem 0;
}
.hum-summary-card {
  background: var(--md-default-bg-color);
  border-radius: 12px;
  border: 1px solid var(--md-default-fg-color--lightest);
  padding: 1.2rem;
  text-align: center;
}
.hum-summary-card .num {
  font-size: 2rem;
  font-weight: 800;
  color: #9f7aea;
  line-height: 1.2;
}
.hum-summary-card .label {
  font-size: 0.82rem;
  color: var(--md-default-fg-color--light);
  margin-top: 0.3rem;
}

.hum-domain-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 1rem;
  margin-bottom: 1.5rem;
}
.hum-domain-group {
  background: var(--md-default-bg-color);
  border-radius: 12px;
  border: 1px solid var(--md-default-fg-color--lightest);
  overflow: hidden;
}
.hum-domain-group-header {
  padding: 0.9rem 1.2rem;
  font-weight: 700;
  font-size: 0.95rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  border-bottom: 1px solid var(--md-default-fg-color--lightest);
}
.hum-domain-group-header.violet { background: #faf5ff; color: #553c9a; }

.hum-domain-items { padding: 0.4rem 0; }
.hum-domain-item {
  display: flex;
  align-items: center;
  padding: 0.55rem 1.2rem;
  gap: 0.6rem;
  text-decoration: none;
  color: inherit;
  transition: background 0.15s;
}
.hum-domain-item:hover { background: var(--md-default-fg-color--lightest); }
.hum-domain-item-name { flex: 1; font-size: 0.88rem; font-weight: 500; }
.hum-domain-item-count {
  font-size: 0.72rem;
  padding: 0.1rem 0.5rem;
  border-radius: 8px;
  font-weight: 600;
}
.hum-domain-item-count.has-content { background: #c6f6d5; color: #22543d; }
.hum-domain-item-count.empty        { background: var(--md-default-fg-color--lightest); color: var(--md-default-fg-color--light); }

@media (max-width: 768px) {
  .hum-domain-grid { grid-template-columns: 1fr; }
  .hum-summary { grid-template-columns: repeat(2, 1fr); }
}
</style>

# 📖 人文素养

> 优秀的工程师不只会写代码。阅读让你看到不同的世界，写作帮你理清思路，思考让你知道自己为什么在做这些事。

<div class="hum-summary">
  <div class="hum-summary-card">
    <div class="num">0</div>
    <div class="label">已填充条目</div>
  </div>
  <div class="hum-summary-card">
    <div class="num">3</div>
    <div class="label">子域数量</div>
  </div>
  <div class="hum-summary-card">
    <div class="num">10+</div>
    <div class="label">可提炼博客</div>
  </div>
  <div class="hum-summary-card">
    <div class="num">🟢</div>
    <div class="label">建设状态</div>
  </div>
</div>

## 子域导航

<div class="hum-domain-grid">

<div class="hum-domain-group">
  <div class="hum-domain-group-header violet">📚 读书笔记</div>
  <div class="hum-domain-items">
    <a href="reading/" class="hum-domain-item">
      <span class="hum-domain-item-name">好书提炼 · 金句 · 个人思考</span>
      <span class="hum-domain-item-count empty">0</span>
    </a>
  </div>
</div>

<div class="hum-domain-group">
  <div class="hum-domain-group-header violet">✍️ 写作</div>
  <div class="hum-domain-items">
    <a href="writing/" class="hum-domain-item">
      <span class="hum-domain-item-name">技术写作方法论 · 习惯养成</span>
      <span class="hum-domain-item-count empty">0</span>
    </a>
  </div>
</div>

<div class="hum-domain-group">
  <div class="hum-domain-group-header violet">💭 哲学思考</div>
  <div class="hum-domain-items">
    <a href="philosophy/" class="hum-domain-item">
      <span class="hum-domain-item-name">技术伦理 · 人生意义 · 自由与责任</span>
      <span class="hum-domain-item-count empty">0</span>
    </a>
  </div>
</div>

</div>

## 🔗 相关博客

博客中有 10 篇读书笔记类文章（如《爱的艺术》《Clean Architecture》等），这些可以直接作为知识条目或扩展为更系统的书评。

---

*维护人：yiiewang · 最后更新：2026-07-01*
