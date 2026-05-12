---
title: 首页
hide:
  - navigation
  - toc
---

<style>
.md-typeset h1 { display: none; }

/* ===== 全局变量 ===== */
:root {
  --card-radius: 10px;
  --card-shadow: 0 2px 12px rgba(0,0,0,0.06);
  --card-shadow-hover: 0 6px 24px rgba(0,0,0,0.1);
}

/* ===== 滚动公告栏 ===== */
.announce-bar {
  background: transparent;
  padding: 0.4rem 0;
  overflow: hidden;
  margin-bottom: 1rem;
}
.announce-track {
  display: flex;
  gap: 4rem;
  animation: marquee 20s linear infinite;
  white-space: nowrap;
}
.announce-item {
  color: var(--md-default-fg-color--light);
  font-size: 0.85rem;
  flex-shrink: 0;
}
@keyframes marquee {
  0% { transform: translateX(0); }
  100% { transform: translateX(-50%); }
}
.announce-bar:hover .announce-track {
  animation-play-state: paused;
}

/* ===== Hero Banner ===== */
.hero-banner-wrapper {
  margin-bottom: 2.5rem;
}
.hero-banner {
  padding: 3rem 2rem;
  background: linear-gradient(135deg, #4a5568 0%, #2d3748 100%);
  border-radius: var(--card-radius);
  color: white;
}
.hero-content {
  display: flex;
  align-items: center;
  gap: 2.5rem;
  flex-wrap: wrap;
}
.hero-avatar-wrapper {
  flex-shrink: 0;
}
.hero-avatar {
  width: 120px;
  height: 120px;
  border-radius: 50%;
  border: 3px solid rgba(255,255,255,0.3);
  object-fit: cover;
}
.hero-text {
  flex: 1;
  min-width: 280px;
}
.hero-greeting {
  font-size: 0.9rem;
  opacity: 0.7;
  margin-bottom: 0.25rem;
  letter-spacing: 2px;
  text-transform: uppercase;
}
.hero-name {
  font-size: 2.4rem;
  font-weight: 700;
  margin-bottom: 0.25rem;
  line-height: 1.2;
}
.hero-title {
  font-size: 1.1rem;
  opacity: 0.75;
  margin-bottom: 0.8rem;
}
.hero-quote {
  font-size: 0.9rem;
  font-style: italic;
  opacity: 0.7;
  padding: 0.6rem 1rem;
  border-left: 3px solid rgba(255,255,255,0.35);
  max-width: 400px;
}
.hero-buttons {
  display: flex;
  gap: 0.8rem;
  margin-top: 1.2rem;
  flex-wrap: wrap;
}
.hero-btn {
  padding: 0.65rem 1.5rem;
  border-radius: 6px;
  font-weight: 600;
  font-size: 0.9rem;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  transition: background 0.2s ease, opacity 0.2s ease;
}
.hero-btn-primary {
  background: rgba(255,255,255,0.95);
  color: #2d3748;
}
.hero-btn-primary:hover {
  opacity: 0.9;
}
.hero-btn-secondary {
  background: transparent;
  color: #ffffff !important;
  border: 1.5px solid rgba(255,255,255,0.4);
}
.hero-btn-secondary:hover {
  border-color: rgba(255,255,255,0.7);
}

/* ===== Stats Bar ===== */

/* ===== Stats Bar ===== */
.stats-bar {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 1rem;
  margin-bottom: 3rem;
}
.stat-card {
  background: var(--md-default-bg-color);
  padding: 1.5rem;
  border-radius: var(--card-radius);
  text-align: center;
  box-shadow: var(--card-shadow);
  transition: box-shadow 0.2s ease;
}
.stat-card:hover {
  box-shadow: var(--card-shadow-hover);
}
.stat-icon {
  font-size: 1.5rem;
  margin-bottom: 0.5rem;
}
.stat-number {
  font-size: 2rem;
  font-weight: 800;
  color: #4a5568;
}
.stat-label {
  font-size: 0.85rem;
  color: var(--md-default-fg-color--light);
  margin-top: 0.3rem;
}

/* ===== Section Title ===== */
.section-title {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  margin: 3rem 0 1.5rem;
  font-size: 1.5rem;
  font-weight: 700;
}
.section-title::after {
  content: '';
  flex: 1;
  height: 2px;
  background: linear-gradient(90deg, var(--md-primary-fg-color) 0%, transparent 100%);
  opacity: 0.3;
}

/* ===== Card Grid ===== */
.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1.5rem;
  margin-bottom: 2rem;
}
.feature-card {
  background: var(--md-default-bg-color);
  border-radius: var(--card-radius);
  box-shadow: var(--card-shadow);
  transition: box-shadow 0.2s ease;
  text-decoration: none;
  color: inherit;
  display: block;
  border-left: 4px solid #667eea;
}
.feature-card:nth-child(2) { border-left-color: #ed8936; }
.feature-card:nth-child(3) { border-left-color: #38b2ac; }
.feature-card:hover {
  box-shadow: var(--card-shadow-hover);
}
.card-body {
  padding: 1.5rem;
}
.card-icon {
  font-size: 1.8rem;
  margin-bottom: 0.75rem;
}
.card-title {
  font-size: 1.25rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
  color: var(--md-default-fg-color);
}
.card-desc {
  font-size: 0.9rem;
  color: var(--md-default-fg-color--light);
  line-height: 1.6;
  margin-bottom: 1rem;
}
.card-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 0.8rem;
  color: var(--md-default-fg-color--light);
}
.card-arrow {
  color: var(--md-default-fg-color--lighter);
  font-size: 1.2rem;
}

/* ===== Article List ===== */
.article-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
.article-item {
  display: flex;
  gap: 1rem;
  padding: 1.2rem;
  background: var(--md-default-bg-color);
  border-radius: var(--card-radius);
  box-shadow: var(--card-shadow);
  transition: box-shadow 0.2s ease;
  text-decoration: none;
  color: inherit;
  align-items: center;
}
.article-item:hover {
  box-shadow: var(--card-shadow-hover);
}
.article-icon {
  width: 50px;
  height: 50px;
  border-radius: var(--card-radius);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  flex-shrink: 0;
}
.article-icon-purple { background: #667eea; color: white; }
.article-icon-blue { background: #4facfe; color: white; }
.article-icon-green { background: #38b2ac; color: white; }
.article-icon-orange { background: #ed8936; color: white; }
.article-icon-red { background: #e53e3e; color: white; }
.article-content {
  flex: 1;
}
.article-title {
  font-weight: 600;
  margin-bottom: 0.3rem;
  color: var(--md-default-fg-color);
}
.article-desc {
  font-size: 0.85rem;
  color: var(--md-default-fg-color--light);
}

/* ===== Tech Stack ===== */
.tech-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 1rem;
}
.tech-item {
  background: var(--md-default-bg-color);
  padding: 1.2rem;
  border-radius: var(--card-radius);
  text-align: center;
  box-shadow: var(--card-shadow);
  transition: box-shadow 0.2s ease;
}
.tech-item:hover {
  box-shadow: var(--card-shadow-hover);
}
.tech-icon {
  font-size: 2rem;
  margin-bottom: 0.5rem;
}
.tech-name {
  font-weight: 600;
  font-size: 0.9rem;
}

/* ===== Footer ===== */
.home-footer {
  text-align: center;
  padding: 2.5rem 1rem;
  margin-top: 2rem;
  border-radius: var(--card-radius);
}
.footer-quote {
  font-size: 1.1rem;
  font-style: italic;
  color: var(--md-default-fg-color);
  margin-bottom: 1.5rem;
}
.footer-links {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
}
.footer-link {
  padding: 0.6rem 1.2rem;
  background: var(--md-default-bg-color);
  border-radius: 6px;
  font-size: 0.85rem;
  box-shadow: var(--card-shadow);
  transition: box-shadow 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}
.footer-link:hover {
  box-shadow: var(--card-shadow-hover);
}

/* ===== Responsive ===== */
@media (max-width: 768px) {
  /* Hero */
  .hero-banner {
    padding: 2.5rem 1.5rem;
  }
  .hero-content {
    flex-direction: column;
    text-align: center;
    gap: 1.5rem;
  }
  .hero-avatar {
    width: 100px;
    height: 100px;
  }
  .hero-greeting {
    font-size: 0.85rem;
  }
  .hero-name {
    font-size: 2rem;
  }
  .hero-title {
    font-size: 1rem;
  }
  .hero-quote {
    margin: 0 auto;
    font-size: 0.85rem;
    padding: 0.6rem 1rem;
    max-width: 90%;
  }
  .hero-buttons {
    justify-content: center;
    flex-wrap: wrap;
    width: 100%;
    max-width: 300px;
    margin: 1.2rem auto 0;
  }
  .hero-btn {
    padding: 0.7rem 1.4rem;
    font-size: 0.9rem;
  }
  
  /* Stats */
  .stats-bar {
    grid-template-columns: repeat(2, 1fr);
    gap: 0.8rem;
  }
  .stat-card {
    padding: 1rem;
  }
  .stat-number {
    font-size: 1.5rem;
  }
  .stat-label {
    font-size: 0.75rem;
  }
  
  /* Section Title */
  .section-title {
    font-size: 1.2rem;
    margin: 2rem 0 1rem;
  }
  
  /* Card Grid */
  .card-grid {
    grid-template-columns: 1fr;
    gap: 1rem;
  }
  .card-body {
    padding: 1.2rem;
  }
  .card-title {
    font-size: 1.1rem;
  }
  .card-desc {
    font-size: 0.85rem;
  }
  
  /* Article List */
  .article-item {
    padding: 1rem;
  }
  .article-icon {
    width: 42px;
    height: 42px;
    font-size: 1.2rem;
  }
  .article-title {
    font-size: 0.95rem;
  }
  .article-desc {
    font-size: 0.8rem;
  }
  
  /* Tech Grid */
  .tech-grid {
    grid-template-columns: repeat(3, 1fr);
    gap: 0.8rem;
  }
  .tech-item {
    padding: 0.8rem;
  }
  .tech-icon {
    font-size: 1.5rem;
  }
  .tech-name {
    font-size: 0.75rem;
  }
  
  /* Footer */
  .home-footer {
    padding: 2rem 1rem;
  }
  .footer-quote {
    font-size: 0.95rem;
  }
  .footer-links {
    gap: 0.6rem;
  }
  .footer-link {
    padding: 0.5rem 1rem;
    font-size: 0.8rem;
  }
}

/* ===== Small Phone ===== */
@media (max-width: 480px) {
  .hero-banner {
    padding: 2rem 1rem 2.5rem;
    border-radius: 12px;
  }
  .hero-avatar {
    width: 80px;
    height: 80px;
    border-width: 3px;
  }
  .hero-greeting {
    font-size: 0.75rem;
    letter-spacing: 1px;
  }
  .hero-name {
    font-size: 1.6rem;
  }
  .hero-title {
    font-size: 0.9rem;
  }
  .hero-quote {
    font-size: 0.8rem;
    padding: 0.5rem 0.8rem;
    max-width: 100%;
  }
  .hero-buttons {
    flex-direction: row;
    gap: 0.5rem;
    margin-top: 1.2rem;
    width: 100%;
  }
  .hero-btn {
    flex: 1;
    padding: 0.6rem 0.8rem;
    font-size: 0.85rem;
    justify-content: center;
  }
  .hero-btn-secondary {
    border-width: 1.5px;
  }
  .stats-bar {
    grid-template-columns: repeat(2, 1fr);
  }
  .tech-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>

<!-- 顶部滚动公告栏 -->
<div class="announce-bar">
  <div class="announce-track">
    <span class="announce-item">🎉 博客首页全新改版，欢迎体验！</span>
    <span class="announce-item">📝 持续更新中，记录技术成长之路</span>
  </div>
</div>

<!-- ===== Hero Banner ===== -->
<div class="hero-banner-wrapper">
  <div class="hero-banner">
    <div class="hero-content">
      <div class="hero-avatar-wrapper">
        <img src="assets/images/avatar.jpg" alt="Cloaks｜Yiiewang" class="hero-avatar" onerror="this.src='https://avatars.githubusercontent.com/u/39525230'">
      </div>
      <div class="hero-text">
        <div class="hero-greeting">Welcome to my blog</div>
        <div class="hero-name">Cloaks｜Yiiewang</div>
        <div class="hero-title">区块链架构师 · 技术探索者</div>
        <div class="hero-quote">「为众人抱薪者，不可使其冻毙于风雪」</div>
        <div class="hero-buttons">
          <a href="blog/" class="hero-btn hero-btn-primary">
            <span>📖</span> 开始阅读
          </a>
          <a href="about/" class="hero-btn hero-btn-secondary">
            <span>👋</span> 了解我
          </a>
        </div>
      </div>
    </div>
  </div>
</div>

<!-- ===== Stats Bar ===== -->
<div class="stats-bar">
  <div class="stat-card">
    <div class="stat-icon">📝</div>
    <div class="stat-number">100+</div>
    <div class="stat-label">原创文章</div>
  </div>
  <div class="stat-card">
    <div class="stat-icon">⏰</div>
    <div class="stat-number">5+ 年</div>
    <div class="stat-label">持续写作</div>
  </div>
  <div class="stat-card">
    <div class="stat-icon">🎨</div>
    <div class="stat-number">23 种</div>
    <div class="stat-label">设计模式</div>
  </div>
  <div class="stat-card">
    <div class="stat-icon">🔗</div>
    <div class="stat-number">区块链</div>
    <div class="stat-label">专注领域</div>
  </div>
</div>

<!-- ===== Featured Content ===== -->
<div class="section-title">🧭 探索内容</div>

<div class="card-grid">
  <a href="blog/" class="feature-card">
    <div class="card-body">
      <div class="card-icon">📝</div>
      <div class="card-title">技术博客</div>
      <div class="card-desc">开发日常中的思考与实践：后端架构、云原生、区块链... 每一篇都是真实项目中踩过的坑。</div>
      <div class="card-meta">
        <span>100+ 篇文章</span>
        <span class="card-arrow">→</span>
      </div>
    </div>
  </a>
  
  <a href="design-patterns/" class="feature-card">
    <div class="card-body">
      <div class="card-icon">🧩</div>
      <div class="card-title">设计模式</div>
      <div class="card-desc">23 种 GoF 设计模式的系统讲解。不只是「是什么」，更关注「为什么」和「怎么用」。</div>
      <div class="card-meta">
        <span>23 种模式</span>
        <span class="card-arrow">→</span>
      </div>
    </div>
  </a>
  
  <a href="ruankao/" class="feature-card">
    <div class="card-body">
      <div class="card-icon">📚</div>
      <div class="card-title">软考笔记</div>
      <div class="card-desc">系统架构师考试的备考笔记。知识点整理 + 真题解析，助你高效通关。</div>
      <div class="card-meta">
        <span>架构师备考</span>
        <span class="card-arrow">→</span>
      </div>
    </div>
  </a>
</div>

<!-- ===== Recent Articles ===== -->
<div class="section-title">🔥 热门推荐</div>

<div class="article-list">
  <a href="blog/posts/2025/03/3/" class="article-item">
    <div class="article-icon article-icon-red">❤️</div>
    <div class="article-content">
      <div class="article-title">爱的艺术</div>
      <div class="article-desc">弗洛姆经典著作读书笔记 —— 爱不是找到对的人，而是培养爱的能力</div>
    </div>
  </a>
  
  <a href="blog/posts/2025/01/19/" class="article-item">
    <div class="article-icon article-icon-blue">⚖️</div>
    <div class="article-content">
      <div class="article-title">负载均衡设计思路与实践指南</div>
      <div class="article-desc">轮询、权重、随机... 分布式系统核心技术详解</div>
    </div>
  </a>
  
  <a href="design-patterns/creational-patterns/singleton/" class="article-item">
    <div class="article-icon article-icon-purple">🎯</div>
    <div class="article-content">
      <div class="article-title">单例模式详解</div>
      <div class="article-desc">最简单也最容易用错的设计模式</div>
    </div>
  </a>
  
  <a href="ruankao/medium%20software%20architecture/" class="article-item">
    <div class="article-icon article-icon-green">🏗️</div>
    <div class="article-content">
      <div class="article-title">软件架构设计</div>
      <div class="article-desc">架构师必备的设计思维与方法论</div>
    </div>
  </a>
</div>

<!-- ===== Tech Stack ===== -->
<div class="section-title">🛠️ 技术栈</div>

<div class="tech-grid">
  <div class="tech-item">
    <div class="tech-icon">🔗</div>
    <div class="tech-name">区块链</div>
  </div>
  <div class="tech-item">
    <div class="tech-icon">🐹</div>
    <div class="tech-name">Golang</div>
  </div>
  <div class="tech-item">
    <div class="tech-icon">☕</div>
    <div class="tech-name">Java</div>
  </div>
  <div class="tech-item">
    <div class="tech-icon">🐳</div>
    <div class="tech-name">Docker</div>
  </div>
  <div class="tech-item">
    <div class="tech-icon">☸️</div>
    <div class="tech-name">Kubernetes</div>
  </div>
  <div class="tech-item">
    <div class="tech-icon">🗄️</div>
    <div class="tech-name">MySQL</div>
  </div>
</div>

<!-- ===== Footer ===== -->
<div class="home-footer">
  <div class="footer-quote">「写下来的东西，比光想要清晰得多」</div>
  <div class="footer-links">
    <a href="https://github.com/yiiewang" target="_blank" class="footer-link">
      <span>🐙</span> GitHub
    </a>
    <a href="https://github.com/yiiewang/yiiewang.github.io" target="_blank" class="footer-link">
      <span>📦</span> 本站源码
    </a>
    <a href="feed_rss_created.xml" target="_blank" class="footer-link">
      <span>📡</span> RSS 订阅
    </a>
  </div>
</div>