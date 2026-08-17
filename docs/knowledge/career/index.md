---
title: 职业发展
domain: career
summary: "面试准备、软技能、技术管理 —— 职业道路上的导航地图和工具箱"
---

<style>
.car-summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  margin: 2rem 0;
}
.car-summary-card {
  background: var(--md-default-bg-color);
  border-radius: 12px;
  border: 1px solid var(--md-default-fg-color--lightest);
  padding: 1.2rem;
  text-align: center;
}
.car-summary-card .num {
  font-size: 2rem;
  font-weight: 800;
  color: #f56565;
  line-height: 1.2;
}
.car-summary-card .label {
  font-size: 0.82rem;
  color: var(--md-default-fg-color--light);
  margin-top: 0.3rem;
}

.car-domain-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 1rem;
  margin-bottom: 1.5rem;
}
.car-domain-group {
  background: var(--md-default-bg-color);
  border-radius: 12px;
  border: 1px solid var(--md-default-fg-color--lightest);
  overflow: hidden;
}
.car-domain-group-header {
  padding: 0.9rem 1.2rem;
  font-weight: 700;
  font-size: 0.95rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  border-bottom: 1px solid var(--md-default-fg-color--lightest);
}
.car-domain-group-header.red { background: #fff5f5; color: #c53030; }

.car-domain-items { padding: 0.4rem 0; }
.car-domain-item {
  display: flex;
  align-items: center;
  padding: 0.55rem 1.2rem;
  gap: 0.6rem;
  text-decoration: none;
  color: inherit;
  transition: background 0.15s;
}
.car-domain-item:hover { background: var(--md-default-fg-color--lightest); }
.car-domain-item-name { flex: 1; font-size: 0.88rem; font-weight: 500; }
.car-domain-item-count {
  font-size: 0.72rem;
  padding: 0.1rem 0.5rem;
  border-radius: 8px;
  font-weight: 600;
}
.car-domain-item-count.has-content { background: #c6f6d5; color: #22543d; }
.car-domain-item-count.empty        { background: var(--md-default-fg-color--lightest); color: var(--md-default-fg-color--light); }

@media (max-width: 768px) {
  .car-domain-grid { grid-template-columns: 1fr; }
  .car-summary { grid-template-columns: repeat(2, 1fr); }
}
</style>

# 🚀 职业发展

> 技术决定你能走多快，职业素养决定你能走多远。

<div class="car-summary">
  <div class="car-summary-card">
    <div class="num">1</div>
    <div class="label">已填充条目</div>
  </div>
  <div class="car-summary-card">
    <div class="num">3</div>
    <div class="label">子域数量</div>
  </div>
  <div class="car-summary-card">
    <div class="num">🟡</div>
    <div class="label">建设状态</div>
  </div>
  <div class="car-summary-card">
    <div class="num">—</div>
    <div class="label">成长导向型</div>
  </div>
</div>

## 子域导航

<div class="car-domain-grid">

<div class="car-domain-group">
  <div class="car-domain-group-header red">🎯 面试准备</div>
  <div class="car-domain-items">
    <a href="interview/" class="car-domain-item">
      <span class="car-domain-item-name">常见面试题 · 系统设计面试 · 行为面试</span>
      <span class="car-domain-item-count empty">0</span>
    </a>
  </div>
</div>

<div class="car-domain-group">
  <div class="car-domain-group-header red">🗣️ 软技能</div>
  <div class="car-domain-items">
    <a href="soft-skills/" class="car-domain-item">
      <span class="car-domain-item-name">沟通 · 协作 · 向上管理 · 技术演讲</span>
      <span class="car-domain-item-count has-content">1</span>
    </a>
  </div>
</div>

<div class="car-domain-group">
  <div class="car-domain-group-header red">👥 技术管理</div>
  <div class="car-domain-items">
    <a href="tech-leadership/" class="car-domain-item">
      <span class="car-domain-item-name">Tech Lead · 团队建设 · 技术规划</span>
      <span class="car-domain-item-count empty">0</span>
    </a>
  </div>
</div>

</div>

## 💡 为什么建这个域

软考·架构师的备考过程中积累了大量职业发展相关内容。此外，从执行者到架构师再到 Tech Lead 的成长路径中，这些「非技术能力」的重要性越来越凸显。

---

*维护人：yiiewang · 最后更新：2026-08-04*
