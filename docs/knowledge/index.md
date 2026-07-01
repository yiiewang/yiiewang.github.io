---
title: 知识库
hide:
  - navigation
  - toc
---

<style>
/* ===== 知识库专用变量 ===== */
:root {
  --kb-purple: #667eea;
  --kb-orange: #ed8936;
  --kb-green: #48bb78;
  --kb-blue: #4299e1;
  --kb-violet: #9f7aea;
  --kb-teal: #38b2ac;
  --kb-red: #f56565;
}

/* ===== Hero ===== */
.kb-hero {
  padding: 3rem 2.5rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 16px;
  color: white;
  margin-bottom: 2.5rem;
  position: relative;
  overflow: hidden;
}
.kb-hero::before {
  content: '';
  position: absolute;
  top: -50%;
  right: -20%;
  width: 400px;
  height: 400px;
  background: rgba(255,255,255,0.05);
  border-radius: 50%;
}
.kb-hero::after {
  content: '';
  position: absolute;
  bottom: -30%;
  left: -10%;
  width: 300px;
  height: 300px;
  background: rgba(255,255,255,0.04);
  border-radius: 50%;
}
.kb-hero h1 {
  color: white;
  font-size: 2.2rem;
  margin-bottom: 0.6rem;
  font-weight: 800;
  position: relative;
  z-index: 1;
}
.kb-hero-desc {
  color: rgba(255,255,255,0.85);
  font-size: 1.05rem;
  max-width: 600px;
  margin: 0 auto 2rem;
  line-height: 1.8;
  position: relative;
  z-index: 1;
}
.kb-hero-stats {
  display: flex;
  gap: 2.5rem;
  justify-content: center;
  flex-wrap: wrap;
  position: relative;
  z-index: 1;
}
.kb-hero-stat {
  text-align: center;
  min-width: 80px;
}
.kb-hero-stat-num {
  font-size: 2.4rem;
  font-weight: 800;
  line-height: 1.2;
}
.kb-hero-stat-label {
  font-size: 0.8rem;
  opacity: 0.75;
  margin-top: 0.2rem;
  font-weight: 500;
}

/* ===== Section Header ===== */
.kb-section-header {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  margin: 3rem 0 1.5rem;
}
.kb-section-header span {
  font-size: 1.3rem;
  font-weight: 700;
  color: var(--md-default-fg-color);
}
.kb-section-header::after {
  content: '';
  flex: 1;
  height: 1px;
  background: var(--md-default-fg-color--lightest);
}

/* ===== Domain Grid ===== */
.kb-domain-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1rem;
  margin-bottom: 2.5rem;
}
.kb-domain-card {
  background: var(--md-default-bg-color);
  border-radius: 12px;
  padding: 1.6rem;
  text-decoration: none;
  color: inherit;
  display: flex;
  gap: 1rem;
  align-items: flex-start;
  border: 1px solid var(--md-default-fg-color--lightest);
  transition: border-color 0.2s, box-shadow 0.2s, transform 0.15s;
}
.kb-domain-card:hover {
  border-color: transparent;
  box-shadow: 0 8px 30px rgba(0,0,0,0.08);
  transform: translateY(-2px);
}
.kb-domain-icon-wrap {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  flex-shrink: 0;
}
.kb-domain-icon-wrap.purple { background: #ebf4ff; }
.kb-domain-icon-wrap.orange { background: #fffaf0; }
.kb-domain-icon-wrap.green  { background: #f0fff4; }
.kb-domain-icon-wrap.blue   { background: #ebf8ff; }
.kb-domain-icon-wrap.violet { background: #faf5ff; }
.kb-domain-icon-wrap.teal   { background: #e6fffa; }
.kb-domain-icon-wrap.red    { background: #fff5f5; }

.kb-domain-body { flex: 1; min-width: 0; }
.kb-domain-name {
  font-size: 1.05rem;
  font-weight: 700;
  margin-bottom: 0.3rem;
}
.kb-domain-desc {
  font-size: 0.85rem;
  color: var(--md-default-fg-color--light);
  line-height: 1.55;
  margin-bottom: 0.5rem;
}
.kb-domain-tag {
  display: inline-block;
  font-size: 0.7rem;
  padding: 0.15rem 0.55rem;
  border-radius: 6px;
  font-weight: 600;
  background: var(--md-default-fg-color--lightest);
  color: var(--md-default-fg-color--light);
}
.kb-domain-tag.active { background: #c6f6d5; color: #22543d; }

/* ===== Usage Cards ===== */
.kb-usage-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 1rem;
  margin-bottom: 2.5rem;
}
.kb-usage-card {
  background: var(--md-default-bg-color);
  border-radius: 12px;
  padding: 1.4rem;
  border: 1px solid var(--md-default-fg-color--lightest);
}
.kb-usage-card-icon {
  font-size: 1.6rem;
  margin-bottom: 0.6rem;
}
.kb-usage-card h4 {
  margin: 0 0 0.4rem;
  font-size: 0.95rem;
  font-weight: 700;
}
.kb-usage-card p {
  margin: 0;
  font-size: 0.82rem;
  color: var(--md-default-fg-color--light);
  line-height: 1.6;
}

/* ===== Status ===== */
.kb-status-row {
  display: flex;
  gap: 2rem;
  flex-wrap: wrap;
  margin-bottom: 2.5rem;
}
.kb-status-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.85rem;
}
.kb-status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}
.kb-status-dot.effective { background: #48bb78; }
.kb-status-dot.draft     { background: #ecc94b; }
.kb-status-dot.outdated  { background: #fc8181; }

/* ===== Responsive ===== */
@media (max-width: 768px) {
  .kb-hero { padding: 2rem 1.5rem; }
  .kb-hero h1 { font-size: 1.6rem; }
  .kb-hero-stats { gap: 1.5rem; }
  .kb-hero-stat-num { font-size: 1.8rem; }
  .kb-domain-grid { grid-template-columns: 1fr; }
  .kb-usage-grid { grid-template-columns: 1fr; }
}
</style>

<div class="kb-hero">
  <h1>📚 知识库</h1>
  <p class="kb-hero-desc">我全部能力的结构化沉淀。不只是一堆笔记，而是思维、经验和认知的系统建模。</p>
  <div class="kb-hero-stats">
    <div class="kb-hero-stat">
      <div class="kb-hero-stat-num">7</div>
      <div class="kb-hero-stat-label">能力域</div>
    </div>
    <div class="kb-hero-stat">
      <div class="kb-hero-stat-num">30+</div>
      <div class="kb-hero-stat-label">知识子域</div>
    </div>
    <div class="kb-hero-stat">
      <div class="kb-hero-stat-num">4</div>
      <div class="kb-hero-stat-label">知识条目</div>
    </div>
  </div>
</div>

<div class="kb-section-header"><span>🧭 七大能力域</span></div>

<div class="kb-domain-grid">

  <a href="tech/" class="kb-domain-card">
    <div class="kb-domain-icon-wrap purple">🖥️</div>
    <div class="kb-domain-body">
      <div class="kb-domain-name">技术能力</div>
      <div class="kb-domain-desc">Go · 区块链 · 架构设计 · 云原生 · 网络 · 存储 · 测试</div>
      <span class="kb-domain-tag active">4 条知识</span>
    </div>
  </a>

  <a href="engineering/" class="kb-domain-card">
    <div class="kb-domain-icon-wrap orange">🔧</div>
    <div class="kb-domain-body">
      <div class="kb-domain-name">工程实践</div>
      <div class="kb-domain-desc">项目复盘 · 踩坑记录 · 性能优化 · 代码审查</div>
      <span class="kb-domain-tag">待填充</span>
    </div>
  </a>

  <a href="thinking/" class="kb-domain-card">
    <div class="kb-domain-icon-wrap green">🧠</div>
    <div class="kb-domain-body">
      <div class="kb-domain-name">思维方法</div>
      <div class="kb-domain-desc">学习方法论 · 问题拆解 · 决策框架 · 系统思维</div>
      <span class="kb-domain-tag">待填充</span>
    </div>
  </a>

  <a href="domain/" class="kb-domain-card">
    <div class="kb-domain-icon-wrap blue">🌐</div>
    <div class="kb-domain-body">
      <div class="kb-domain-name">领域知识</div>
      <div class="kb-domain-desc">区块链行业 · 金融投资 · 经济学</div>
      <span class="kb-domain-tag">待填充</span>
    </div>
  </a>

  <a href="humanities/" class="kb-domain-card">
    <div class="kb-domain-icon-wrap violet">📖</div>
    <div class="kb-domain-body">
      <div class="kb-domain-name">人文素养</div>
      <div class="kb-domain-desc">读书笔记 · 写作 · 哲学思考</div>
      <span class="kb-domain-tag">待填充</span>
    </div>
  </a>

  <a href="tools/" class="kb-domain-card">
    <div class="kb-domain-icon-wrap teal">⚡</div>
    <div class="kb-domain-body">
      <div class="kb-domain-name">工具与效率</div>
      <div class="kb-domain-desc">开发工具 · 效率方法论 · 自动化脚本</div>
      <span class="kb-domain-tag">待填充</span>
    </div>
  </a>

  <a href="career/" class="kb-domain-card">
    <div class="kb-domain-icon-wrap red">🚀</div>
    <div class="kb-domain-body">
      <div class="kb-domain-name">职业发展</div>
      <div class="kb-domain-desc">面试准备 · 软技能 · 技术管理</div>
      <span class="kb-domain-tag">待填充</span>
    </div>
  </a>

</div>

<div class="kb-section-header"><span>📋 知识状态</span></div>

<div class="kb-status-row">
  <div class="kb-status-pill">
    <span class="kb-status-dot effective"></span> <strong>有效</strong> 当前可用，直接参考
  </div>
  <div class="kb-status-pill">
    <span class="kb-status-dot draft"></span> <strong>草稿</strong> 内容不完整，持续完善
  </div>
  <div class="kb-status-pill">
    <span class="kb-status-dot outdated"></span> <strong>待更新</strong> 部分过时，需要修订
  </div>
</div>

<div class="kb-section-header"><span>🔍 使用方式</span></div>

<div class="kb-usage-grid">
  <div class="kb-usage-card">
    <div class="kb-usage-card-icon">🔎</div>
    <h4>搜索即入口</h4>
    <p>顶部搜索框支持<strong>全文检索</strong>，同时搜索博客和知识库，输入关键词直达答案。</p>
  </div>
  <div class="kb-usage-card">
    <div class="kb-usage-card-icon">🏷️</div>
    <h4>标签即线索</h4>
    <p>每篇知识文档都有 <code>tags</code>，点一个标签就能串联所有相关内容。</p>
  </div>
  <div class="kb-usage-card">
    <div class="kb-usage-card-icon">📎</div>
    <h4>来源可追溯</h4>
    <p>每条知识的 <code>source</code> 标注了出处——博客、书籍或项目经验，有据可查。</p>
  </div>
  <div class="kb-usage-card">
    <div class="kb-usage-card-icon">🔄</div>
    <h4>持续迭代</h4>
    <p>不是一次性上传的文档库，而是不断更新的能力图谱。</p>
  </div>
</div>

<div style="text-align: center; padding: 2rem 0; color: var(--md-default-fg-color--light);">
  <p style="font-weight: 600; font-size: 1.05rem;">知识库不是笔记堆砌，而是思维的结构化</p>
  <p style="font-size: 0.8rem; margin-top: 0.5rem;">维护人：yiiewang · 内容有误或过时欢迎留言</p>
</div>
