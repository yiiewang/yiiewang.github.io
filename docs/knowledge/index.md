---
title: 知识库
hide:
  - navigation
  - toc
---

<style>
.kb-hero {
  padding: 2.5rem 2rem;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
  border-radius: 12px;
  color: white;
  margin-bottom: 2.5rem;
  text-align: center;
}
.kb-hero h1 {
  color: white;
  font-size: 2rem;
  margin-bottom: 0.5rem;
  font-weight: 700;
}
.kb-hero p {
  color: rgba(255,255,255,0.75);
  font-size: 1.05rem;
  max-width: 600px;
  margin: 0 auto 1.5rem;
  line-height: 1.7;
}
.kb-hero .kb-stats {
  display: flex;
  gap: 2rem;
  justify-content: center;
  flex-wrap: wrap;
}
.kb-hero .kb-stat {
  text-align: center;
}
.kb-stat-num {
  font-size: 2rem;
  font-weight: 800;
  color: #e2b04a;
}
.kb-stat-label {
  font-size: 0.8rem;
  opacity: 0.7;
  margin-top: 0.2rem;
}

.domain-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1.2rem;
  margin-bottom: 2.5rem;
}
.domain-card {
  background: var(--md-default-bg-color);
  border-radius: 10px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.06);
  padding: 1.5rem;
  transition: box-shadow 0.2s, transform 0.2s;
  text-decoration: none;
  color: inherit;
  display: block;
  border-left: 4px solid #667eea;
}
.domain-card:hover {
  box-shadow: 0 6px 24px rgba(0,0,0,0.1);
  transform: translateY(-2px);
}
.domain-card:nth-child(2) { border-left-color: #ed8936; }
.domain-card:nth-child(3) { border-left-color: #48bb78; }
.domain-card:nth-child(4) { border-left-color: #4299e1; }
.domain-card:nth-child(5) { border-left-color: #9f7aea; }
.domain-card:nth-child(6) { border-left-color: #38b2ac; }
.domain-card:nth-child(7) { border-left-color: #f56565; }

.domain-icon {
  font-size: 1.8rem;
  margin-bottom: 0.6rem;
}
.domain-title {
  font-size: 1.15rem;
  font-weight: 700;
  margin-bottom: 0.4rem;
}
.domain-desc {
  font-size: 0.85rem;
  color: var(--md-default-fg-color--light);
  line-height: 1.6;
  margin-bottom: 0.8rem;
}
.domain-meta {
  font-size: 0.75rem;
  color: var(--md-default-fg-color--lighter);
}

.status-badge {
  display: inline-block;
  padding: 0.15rem 0.6rem;
  border-radius: 10px;
  font-size: 0.7rem;
  font-weight: 600;
}
.status-effective { background: #c6f6d5; color: #22543d; }
.status-draft { background: #fefcbf; color: #744210; }
.status-outdated { background: #fed7d7; color: #742a2a; }

.status-list {
  list-style: none;
  padding: 0;
  margin: 0 0 2.5rem;
}
.status-item {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  padding: 0.6rem 0;
  border-bottom: 1px dashed var(--md-default-fg-color--lightest);
}
.status-item:last-child { border-bottom: none; }
.status-desc {
  color: var(--md-default-fg-color--light);
  font-size: 0.9rem;
}

.usage-section {
  background: var(--md-default-bg-color);
  border-radius: 10px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.06);
  padding: 1.5rem 2rem;
  margin-bottom: 2.5rem;
}
.usage-section h3 {
  margin-top: 0;
}

@media (max-width: 768px) {
  .kb-hero { padding: 2rem 1rem; }
  .kb-hero h1 { font-size: 1.6rem; }
  .domain-grid { grid-template-columns: 1fr; }
}
</style>

<div class="kb-hero">
  <h1>📚 知识库</h1>
  <p>这里是我全部能力的结构化沉淀 —— 不只是技术栈，更是思维方式、工程实践和领域认知的系统建模。</p>
  <div class="kb-stats">
    <div class="kb-stat">
      <div class="kb-stat-num">7</div>
      <div class="kb-stat-label">能力域</div>
    </div>
    <div class="kb-stat">
      <div class="kb-stat-num">30+</div>
      <div class="kb-stat-label">知识子域</div>
    </div>
    <div class="kb-stat">
      <div class="kb-stat-num">4</div>
      <div class="kb-stat-label">知识条目</div>
    </div>
  </div>
</div>

## 🧭 七大能力域

<div class="domain-grid">

  <a href="tech/" class="domain-card">
    <div class="domain-icon">🖥️</div>
    <div class="domain-title">技术能力</div>
    <div class="domain-desc">编程语言、区块链、架构设计、云原生、网络协议、数据存储、测试工程</div>
    <div class="domain-meta">4 条知识 · 128 篇博客可提炼</div>
  </a>

  <a href="engineering/" class="domain-card">
    <div class="domain-icon">🔧</div>
    <div class="domain-title">工程实践</div>
    <div class="domain-desc">项目复盘、踩坑记录、性能优化、代码审查 —— 从真实项目中提取的经验</div>
    <div class="domain-meta">经验驱动 · 持续积累</div>
  </a>

  <a href="thinking/" class="domain-card">
    <div class="domain-icon">🧠</div>
    <div class="domain-title">思维方法</div>
    <div class="domain-desc">学习方法论、问题拆解、决策框架、系统思维 —— 如何思考比思考什么更重要</div>
    <div class="domain-meta">元能力 · 正在建设中</div>
  </a>

  <a href="domain/" class="domain-card">
    <div class="domain-icon">🌐</div>
    <div class="domain-title">领域知识</div>
    <div class="domain-desc">区块链行业洞察、金融与投资、经济学 —— 对所处领域的深度理解</div>
    <div class="domain-meta">行业纵深 · 博客可提炼</div>
  </a>

  <a href="humanities/" class="domain-card">
    <div class="domain-icon">📖</div>
    <div class="domain-title">人文素养</div>
    <div class="domain-desc">读书笔记、写作、哲学思考 —— 技术之外的知识滋养</div>
    <div class="domain-meta">精神基建 · 持续更新</div>
  </a>

  <a href="tools/" class="domain-card">
    <div class="domain-icon">⚡</div>
    <div class="domain-title">工具与效率</div>
    <div class="domain-desc">开发工具、效率方法论、自动化脚本 —— 磨刀不误砍柴工</div>
    <div class="domain-meta">效率杠杆 · 正在建设中</div>
  </a>

  <a href="career/" class="domain-card">
    <div class="domain-icon">🚀</div>
    <div class="domain-title">职业发展</div>
    <div class="domain-desc">面试准备、软技能、技术管理 —— 职业路上的导航地图</div>
    <div class="domain-meta">长期建设 · 正在规划</div>
  </a>

</div>

## 📋 知识状态说明

每篇知识文档都会标注一个状态，帮你判断信息的可信度和时效性：

<ul class="status-list">
  <li class="status-item">
    <span class="status-badge status-effective">有效</span>
    <span class="status-desc">当前有效，可直接参考使用</span>
  </li>
  <li class="status-item">
    <span class="status-badge status-draft">草稿</span>
    <span class="status-desc">内容尚不完整，欢迎贡献</span>
  </li>
  <li class="status-item">
    <span class="status-badge status-outdated">待更新</span>
    <span class="status-desc">部分内容已过时，需要补充</span>
  </li>
</ul>

## 🔍 如何使用

<div class="usage-section" markdown="1">

### 搜索是入口
顶部搜索框支持**全文检索** —— 输入任何关键词，可以同时搜索博客和知识库的内容。

### 标签是纽带
每篇知识文档都有 `tags`，点击标签可以找到该主题下的所有相关内容（包括博客文章）。

### 来源可追溯
知识条目的 `source` 字段标注了知识来源 —— 可能是某篇博客、某本书或某个项目经验。

### 持续迭代
这不是一次性上传的文档库，而是持续更新的知识体系。每篇条目都在不断丰富和完善。

</div>

---

<div style="text-align: center; padding: 2rem 0; color: var(--md-default-fg-color--light);">
  <p><strong>「知识库不是笔记堆砌，而是思维的结构化」</strong></p>
  <p style="font-size: 0.85rem;">维护人：yiiewang · 如果你发现内容有误或过时，欢迎在对应页面留言</p>
</div>
