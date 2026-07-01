---
title: 工具与效率
domain: tools
summary: "开发工具、效率方法论、自动化脚本 —— 把自己从重复劳动中解放出来"
---

<style>
.tol-summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  margin: 2rem 0;
}
.tol-summary-card {
  background: var(--md-default-bg-color);
  border-radius: 12px;
  border: 1px solid var(--md-default-fg-color--lightest);
  padding: 1.2rem;
  text-align: center;
}
.tol-summary-card .num {
  font-size: 2rem;
  font-weight: 800;
  color: #38b2ac;
  line-height: 1.2;
}
.tol-summary-card .label {
  font-size: 0.82rem;
  color: var(--md-default-fg-color--light);
  margin-top: 0.3rem;
}

.tol-domain-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 1rem;
  margin-bottom: 1.5rem;
}
.tol-domain-group {
  background: var(--md-default-bg-color);
  border-radius: 12px;
  border: 1px solid var(--md-default-fg-color--lightest);
  overflow: hidden;
}
.tol-domain-group-header {
  padding: 0.9rem 1.2rem;
  font-weight: 700;
  font-size: 0.95rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  border-bottom: 1px solid var(--md-default-fg-color--lightest);
}
.tol-domain-group-header.teal { background: #e6fffa; color: #234e52; }

.tol-domain-items { padding: 0.4rem 0; }
.tol-domain-item {
  display: flex;
  align-items: center;
  padding: 0.55rem 1.2rem;
  gap: 0.6rem;
  text-decoration: none;
  color: inherit;
  transition: background 0.15s;
}
.tol-domain-item:hover { background: var(--md-default-fg-color--lightest); }
.tol-domain-item-name { flex: 1; font-size: 0.88rem; font-weight: 500; }
.tol-domain-item-count {
  font-size: 0.72rem;
  padding: 0.1rem 0.5rem;
  border-radius: 8px;
  font-weight: 600;
}
.tol-domain-item-count.has-content { background: #c6f6d5; color: #22543d; }
.tol-domain-item-count.empty        { background: var(--md-default-fg-color--lightest); color: var(--md-default-fg-color--light); }

@media (max-width: 768px) {
  .tol-domain-grid { grid-template-columns: 1fr; }
  .tol-summary { grid-template-columns: repeat(2, 1fr); }
}
</style>

# ⚡ 工具与效率

> 善用工具的人，一天有 48 小时。

<div class="tol-summary">
  <div class="tol-summary-card">
    <div class="num">0</div>
    <div class="label">已填充条目</div>
  </div>
  <div class="tol-summary-card">
    <div class="num">3</div>
    <div class="label">子域数量</div>
  </div>
  <div class="tol-summary-card">
    <div class="num">🟡</div>
    <div class="label">建设状态</div>
  </div>
  <div class="tol-summary-card">
    <div class="num">—</div>
    <div class="label">实用导向型</div>
  </div>
</div>

## 子域导航

<div class="tol-domain-grid">

<div class="tol-domain-group">
  <div class="tol-domain-group-header teal">🛠️ 开发工具</div>
  <div class="tol-domain-items">
    <a href="dev-tools/" class="tol-domain-item">
      <span class="tol-domain-item-name">IDE · 调试技巧 · 命令行 · Git 工作流</span>
      <span class="tol-domain-item-count empty">0</span>
    </a>
  </div>
</div>

<div class="tol-domain-group">
  <div class="tol-domain-group-header teal">⏱️ 效率方法论</div>
  <div class="tol-domain-items">
    <a href="productivity/" class="tol-domain-item">
      <span class="tol-domain-item-name">GTD · 番茄工作法 · 深度工作</span>
      <span class="tol-domain-item-count empty">0</span>
    </a>
  </div>
</div>

<div class="tol-domain-group">
  <div class="tol-domain-group-header teal">🤖 自动化脚本</div>
  <div class="tol-domain-items">
    <a href="automation/" class="tol-domain-item">
      <span class="tol-domain-item-name">Shell · CI 自动化 · 日常任务</span>
      <span class="tol-domain-item-count empty">0</span>
    </a>
  </div>
</div>

</div>

---

*维护人：yiiewang · 最后更新：2026-07-01*
