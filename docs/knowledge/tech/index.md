---
title: 技术能力
status: effective
domain: tech
summary: "编程语言、区块链、架构设计、云原生、网络协议、数据存储、测试工程 — 最核心的硬技能沉淀"
---

<style>
.tech-domain-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 1rem;
  margin-bottom: 1.5rem;
}
.tech-domain-group {
  background: var(--md-default-bg-color);
  border-radius: 12px;
  border: 1px solid var(--md-default-fg-color--lightest);
  overflow: hidden;
}
.tech-domain-group-header {
  padding: 0.9rem 1.2rem;
  font-weight: 700;
  font-size: 0.95rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  border-bottom: 1px solid var(--md-default-fg-color--lightest);
}
.tech-domain-group-header.purple { background: #f8f7ff; color: #553c9a; }
.tech-domain-group-header.orange { background: #fffaf0; color: #c05621; }
.tech-domain-group-header.green  { background: #f0fff4; color: #22543d; }
.tech-domain-group-header.blue   { background: #ebf8ff; color: #2a4365; }
.tech-domain-group-header.teal   { background: #e6fffa; color: #234e52; }

.tech-domain-items {
  padding: 0.4rem 0;
}
.tech-domain-item {
  display: flex;
  align-items: center;
  padding: 0.55rem 1.2rem;
  gap: 0.6rem;
  text-decoration: none;
  color: inherit;
  transition: background 0.15s;
}
.tech-domain-item:hover {
  background: var(--md-default-fg-color--lightest);
}
.tech-domain-item-name {
  flex: 1;
  font-size: 0.88rem;
  font-weight: 500;
}
.tech-domain-item-count {
  font-size: 0.72rem;
  padding: 0.1rem 0.5rem;
  border-radius: 8px;
  font-weight: 600;
}
.tech-domain-item-count.has-content { background: #c6f6d5; color: #22543d; }
.tech-domain-item-count.empty        { background: var(--md-default-fg-color--lightest); color: var(--md-default-fg-color--light); }
.tech-domain-item-count.linked       { background: #bee3f8; color: #2a4365; }

.tech-summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  margin: 2rem 0;
}
.tech-summary-card {
  background: var(--md-default-bg-color);
  border-radius: 12px;
  border: 1px solid var(--md-default-fg-color--lightest);
  padding: 1.2rem;
  text-align: center;
}
.tech-summary-card .num {
  font-size: 2rem;
  font-weight: 800;
  color: #667eea;
  line-height: 1.2;
}
.tech-summary-card .label {
  font-size: 0.82rem;
  color: var(--md-default-fg-color--light);
  margin-top: 0.3rem;
}

@media (max-width: 768px) {
  .tech-domain-grid { grid-template-columns: 1fr; }
  .tech-summary { grid-template-columns: repeat(2, 1fr); }
}
</style>

# 🖥️ 技术能力

> 最活跃的知识域。这里的内容来自真实项目和博客的深度提炼，不是概念罗列。

<div class="tech-summary">
  <div class="tech-summary-card">
    <div class="num">4</div>
    <div class="label">已填充条目</div>
  </div>
  <div class="tech-summary-card">
    <div class="num">128</div>
    <div class="label">可提炼博客</div>
  </div>
  <div class="tech-summary-card">
    <div class="num">23</div>
    <div class="label">设计模式 (已有)</div>
  </div>
  <div class="tech-summary-card">
    <div class="num">10</div>
    <div class="label">子域数量</div>
  </div>
</div>

## 子域导航

<div class="tech-domain-grid">

<div class="tech-domain-group">
  <div class="tech-domain-group-header purple">🔤 编程语言</div>
  <div class="tech-domain-items">
    <a href="languages/go/" class="tech-domain-item">
      <span class="tech-domain-item-name">Go 语言</span>
      <span class="tech-domain-item-count has-content">1</span>
    </a>
    <a href="languages/java/" class="tech-domain-item">
      <span class="tech-domain-item-name">Java</span>
      <span class="tech-domain-item-count empty">0</span>
    </a>
    <a href="languages/python/" class="tech-domain-item">
      <span class="tech-domain-item-name">Python</span>
      <span class="tech-domain-item-count empty">0</span>
    </a>
  </div>
</div>

<div class="tech-domain-group">
  <div class="tech-domain-group-header orange">⛓️ 区块链</div>
  <div class="tech-domain-items">
    <a href="blockchain/consensus/" class="tech-domain-item">
      <span class="tech-domain-item-name">共识算法</span>
      <span class="tech-domain-item-count has-content">1</span>
    </a>
    <a href="blockchain/smart-contract/" class="tech-domain-item">
      <span class="tech-domain-item-name">智能合约</span>
      <span class="tech-domain-item-count empty">0</span>
    </a>
    <a href="blockchain/chainmaker/" class="tech-domain-item">
      <span class="tech-domain-item-name">ChainMaker</span>
      <span class="tech-domain-item-count empty">0</span>
    </a>
    <a href="blockchain/ledger/" class="tech-domain-item">
      <span class="tech-domain-item-name">分布式账本</span>
      <span class="tech-domain-item-count empty">0</span>
    </a>
  </div>
</div>

<div class="tech-domain-group">
  <div class="tech-domain-group-header green">🏗️ 架构设计</div>
  <div class="tech-domain-items">
    <a href="architecture/system-design/" class="tech-domain-item">
      <span class="tech-domain-item-name">系统设计</span>
      <span class="tech-domain-item-count has-content">1</span>
    </a>
    <a href="architecture/patterns/" class="tech-domain-item">
      <span class="tech-domain-item-name">设计模式</span>
      <span class="tech-domain-item-count linked">23</span>
    </a>
    <a href="architecture/distributed/" class="tech-domain-item">
      <span class="tech-domain-item-name">分布式系统</span>
      <span class="tech-domain-item-count has-content">1</span>
    </a>
    <a href="architecture/microservices/" class="tech-domain-item">
      <span class="tech-domain-item-name">微服务</span>
      <span class="tech-domain-item-count empty">0</span>
    </a>
  </div>
</div>

<div class="tech-domain-group">
  <div class="tech-domain-group-header blue">☁️ 云原生</div>
  <div class="tech-domain-items">
    <a href="cloud-native/docker/" class="tech-domain-item">
      <span class="tech-domain-item-name">Docker</span>
      <span class="tech-domain-item-count empty">0</span>
    </a>
    <a href="cloud-native/kubernetes/" class="tech-domain-item">
      <span class="tech-domain-item-name">Kubernetes</span>
      <span class="tech-domain-item-count empty">0</span>
    </a>
    <a href="cloud-native/cicd/" class="tech-domain-item">
      <span class="tech-domain-item-name">CI/CD</span>
      <span class="tech-domain-item-count empty">0</span>
    </a>
  </div>
</div>

<div class="tech-domain-group">
  <div class="tech-domain-group-header teal">🌐 基础设施</div>
  <div class="tech-domain-items">
    <a href="network/" class="tech-domain-item">
      <span class="tech-domain-item-name">网络协议</span>
      <span class="tech-domain-item-count empty">0</span>
    </a>
    <a href="storage/" class="tech-domain-item">
      <span class="tech-domain-item-name">数据存储</span>
      <span class="tech-domain-item-count empty">0</span>
    </a>
    <a href="testing/" class="tech-domain-item">
      <span class="tech-domain-item-name">测试工程</span>
      <span class="tech-domain-item-count empty">0</span>
    </a>
  </div>
</div>

</div>

---

*维护人：yiiewang · 最后更新：2026-07-01*
