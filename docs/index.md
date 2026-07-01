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
  --card-radius: 12px;
  --card-radius-sm: 8px;
  --card-shadow: 0 2px 12px rgba(0,0,0,0.06);
  --card-shadow-hover: 0 8px 30px rgba(0,0,0,0.12);
  --accent: #667eea;
  --accent-2: #764ba2;
  --accent-orange: #ed8936;
  --accent-green: #48bb78;
  --accent-teal: #38b2ac;
  --accent-red: #e53e3e;
}

/* ===== Hero Banner ===== */
.hero-banner {
  position: relative;
  padding: 4rem 2.5rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 16px;
  color: white;
  margin-bottom: 2.5rem;
  overflow: hidden;
  box-shadow: 0 10px 40px rgba(102,126,234,0.25);
}
.hero-banner::before {
  content: '';
  position: absolute;
  top: -60%;
  right: -15%;
  width: 480px;
  height: 480px;
  background: rgba(255,255,255,0.06);
  border-radius: 50%;
  pointer-events: none;
}
.hero-banner::after {
  content: '';
  position: absolute;
  bottom: -40%;
  left: -10%;
  width: 360px;
  height: 360px;
  background: rgba(255,255,255,0.05);
  border-radius: 50%;
  pointer-events: none;
}
.hero-content {
  display: flex;
  align-items: center;
  gap: 2.5rem;
  flex-wrap: wrap;
  position: relative;
  z-index: 1;
}
.hero-avatar-wrapper {
  flex-shrink: 0;
  position: relative;
}
.hero-avatar-ring {
  width: 132px;
  height: 132px;
  border-radius: 50%;
  padding: 3px;
  background: linear-gradient(135deg, rgba(255,255,255,0.4), rgba(255,255,255,0.1));
  display: flex;
  align-items: center;
  justify-content: center;
}
.hero-avatar {
  width: 120px;
  height: 120px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid rgba(255,255,255,0.3);
}
.hero-text {
  flex: 1;
  min-width: 280px;
}
.hero-greeting {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.85rem;
  opacity: 0.85;
  margin-bottom: 0.6rem;
  letter-spacing: 1px;
  padding: 0.25rem 0.75rem;
  background: rgba(255,255,255,0.15);
  border-radius: 20px;
  backdrop-filter: blur(10px);
}
.hero-greeting-dot {
  width: 6px;
  height: 6px;
  background: #48bb78;
  border-radius: 50%;
  box-shadow: 0 0 0 4px rgba(72,187,120,0.3);
  animation: pulse 2s ease-in-out infinite;
}
@keyframes pulse {
  0%, 100% { box-shadow: 0 0 0 4px rgba(72,187,120,0.3); }
  50% { box-shadow: 0 0 0 8px rgba(72,187,120,0); }
}
.hero-name {
  font-size: 2.6rem;
  font-weight: 800;
  margin-bottom: 0.4rem;
  line-height: 1.15;
  letter-spacing: -0.5px;
}
.hero-title {
  font-size: 1.05rem;
  opacity: 0.85;
  margin-bottom: 1rem;
  font-weight: 500;
}
.hero-title-tag {
  display: inline-block;
  padding: 0.15rem 0.6rem;
  background: rgba(255,255,255,0.18);
  border-radius: 4px;
  margin-right: 0.4rem;
  font-size: 0.9rem;
}
.hero-quote {
  font-size: 0.95rem;
  font-style: italic;
  opacity: 0.8;
  padding: 0.7rem 1rem;
  border-left: 3px solid rgba(255,255,255,0.4);
  max-width: 440px;
  margin: 0.4rem 0 1.2rem;
  line-height: 1.6;
}
.hero-buttons {
  display: flex;
  gap: 0.7rem;
  flex-wrap: wrap;
}
.hero-btn {
  padding: 0.7rem 1.5rem;
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.92rem;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  transition: all 0.2s ease;
}
.hero-btn-primary {
  background: white;
  color: #553c9a;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}
.hero-btn-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  color: #553c9a;
}
.hero-btn-secondary {
  background: rgba(255,255,255,0.12);
  color: white !important;
  border: 1.5px solid rgba(255,255,255,0.3);
  backdrop-filter: blur(10px);
}
.hero-btn-secondary:hover {
  background: rgba(255,255,255,0.2);
  border-color: rgba(255,255,255,0.6);
}

/* ===== Stats Bar ===== */
.stats-bar {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 1rem;
  margin-bottom: 3rem;
}
.stat-card {
  background: var(--md-default-bg-color);
  padding: 1.4rem 1rem;
  border-radius: var(--card-radius);
  text-align: center;
  box-shadow: var(--card-shadow);
  transition: all 0.2s ease;
  border: 1px solid transparent;
}
.stat-card:hover {
  box-shadow: var(--card-shadow-hover);
  transform: translateY(-2px);
  border-color: var(--accent);
}
.stat-icon {
  font-size: 1.6rem;
  margin-bottom: 0.5rem;
  display: block;
}
.stat-number {
  font-size: 1.85rem;
  font-weight: 800;
  background: linear-gradient(135deg, var(--accent), var(--accent-2));
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  line-height: 1.2;
}
.stat-label {
  font-size: 0.82rem;
  color: var(--md-default-fg-color--light);
  margin-top: 0.3rem;
}

/* ===== Section Title ===== */
.section-title {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  margin: 3.5rem 0 1.5rem;
  font-size: 1.4rem;
  font-weight: 700;
  color: var(--md-default-fg-color);
}
.section-title-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 1.1rem;
  flex-shrink: 0;
}
.section-title-icon.purple { background: linear-gradient(135deg, #667eea, #764ba2); color: white; }
.section-title-icon.orange { background: linear-gradient(135deg, #fb923c, #f97316); color: white; }
.section-title-icon.green  { background: linear-gradient(135deg, #4ade80, #22c55e); color: white; }
.section-title-icon.teal   { background: linear-gradient(135deg, #2dd4bf, #06b6d4); color: white; }
.section-title::after {
  content: '';
  flex: 1;
  height: 1px;
  background: linear-gradient(90deg, var(--md-default-fg-color--lightest), transparent);
}

/* ===== Featured Card Grid (改造：圆角 + icon 背景 + 悬停效果) ===== */
.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1.2rem;
  margin-bottom: 2rem;
}
.feature-card {
  background: var(--md-default-bg-color);
  border-radius: var(--card-radius);
  box-shadow: var(--card-shadow);
  transition: all 0.25s ease;
  text-decoration: none;
  color: inherit;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid transparent;
}
.feature-card:hover {
  box-shadow: var(--card-shadow-hover);
  transform: translateY(-3px);
  border-color: var(--accent);
}
.feature-card-head {
  padding: 1.4rem 1.4rem 0.4rem;
  display: flex;
  align-items: center;
  gap: 0.8rem;
}
.feature-icon-wrap {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.3rem;
  flex-shrink: 0;
}
.feature-icon-wrap.purple { background: linear-gradient(135deg, #ebf4ff, #e0e7ff); }
.feature-icon-wrap.orange { background: linear-gradient(135deg, #fff7ed, #ffedd5); }
.feature-icon-wrap.green  { background: linear-gradient(135deg, #f0fdf4, #dcfce7); }
.feature-icon-wrap.teal   { background: linear-gradient(135deg, #ecfeff, #cffafe); }
.feature-icon-wrap.red    { background: linear-gradient(135deg, #fef2f2, #fecaca); }
.feature-icon-wrap.violet { background: linear-gradient(135deg, #faf5ff, #ede9fe); }
.feature-icon-wrap.blue   { background: linear-gradient(135deg, #eff6ff, #dbeafe); }
.card-body {
  padding: 0.4rem 1.4rem 1.4rem;
  flex: 1;
  display: flex;
  flex-direction: column;
}
.card-title {
  font-size: 1.15rem;
  font-weight: 700;
  margin-bottom: 0.4rem;
  color: var(--md-default-fg-color);
}
.card-desc {
  font-size: 0.88rem;
  color: var(--md-default-fg-color--light);
  line-height: 1.6;
  margin-bottom: 0.8rem;
  flex: 1;
}
.card-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 0.8rem;
  color: var(--md-default-fg-color--light);
  padding-top: 0.8rem;
  border-top: 1px solid var(--md-default-fg-color--lightest);
}
.card-tag {
  display: inline-block;
  padding: 0.15rem 0.55rem;
  background: var(--md-default-fg-color--lightest);
  border-radius: 4px;
  font-weight: 600;
  font-size: 0.72rem;
}
.card-arrow {
  color: var(--accent);
  font-size: 1.1rem;
  font-weight: 600;
  transition: transform 0.2s ease;
}
.feature-card:hover .card-arrow {
  transform: translateX(3px);
}

/* ===== Recent Articles (改造：去掉冗余的彩色方块图标，改用类型徽章) ===== */
.article-list {
  display: flex;
  flex-direction: column;
  gap: 0.8rem;
}
.article-item {
  display: flex;
  gap: 1rem;
  padding: 1.1rem 1.3rem;
  background: var(--md-default-bg-color);
  border-radius: var(--card-radius);
  box-shadow: var(--card-shadow);
  transition: all 0.2s ease;
  text-decoration: none;
  color: inherit;
  align-items: center;
  border: 1px solid transparent;
}
.article-item:hover {
  box-shadow: var(--card-shadow-hover);
  transform: translateX(3px);
  border-color: var(--accent);
}
.article-badge {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.6rem;
  flex-shrink: 0;
  position: relative;
  overflow: hidden;
}
.article-badge::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, rgba(255,255,255,0.3), transparent);
}
.article-badge.purple { background: linear-gradient(135deg, #667eea, #764ba2); }
.article-badge.blue   { background: linear-gradient(135deg, #4facfe, #00f2fe); }
.article-badge.green  { background: linear-gradient(135deg, #43e97b, #38f9d7); }
.article-badge.orange { background: linear-gradient(135deg, #fa709a, #fee140); }
.article-badge.red    { background: linear-gradient(135deg, #ee0979, #ff6a00); }
.article-content {
  flex: 1;
  min-width: 0;
}
.article-title {
  font-weight: 600;
  font-size: 1rem;
  margin-bottom: 0.25rem;
  color: var(--md-default-fg-color);
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.article-desc {
  font-size: 0.85rem;
  color: var(--md-default-fg-color--light);
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.article-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-shrink: 0;
  font-size: 0.78rem;
  color: var(--md-default-fg-color--light);
}
.article-type {
  padding: 0.15rem 0.55rem;
  border-radius: 4px;
  font-weight: 600;
  font-size: 0.72rem;
  background: var(--md-default-fg-color--lightest);
}

/* ===== Tech Stack (改造：横排 chip 风格，去掉方形卡片) ===== */
.tech-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem;
  margin-bottom: 2rem;
}
.tech-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.55rem 1rem;
  background: var(--md-default-bg-color);
  border: 1px solid var(--md-default-fg-color--lightest);
  border-radius: 100px;
  font-weight: 600;
  font-size: 0.88rem;
  transition: all 0.2s ease;
  cursor: default;
}
.tech-chip:hover {
  border-color: var(--accent);
  transform: translateY(-1px);
  box-shadow: var(--card-shadow);
}
.tech-chip-icon {
  font-size: 1.1rem;
}

/* ===== Footer ===== */
.home-footer {
  text-align: center;
  padding: 3rem 1.5rem;
  margin-top: 3rem;
  border-radius: 16px;
  background: linear-gradient(135deg, var(--md-default-bg-color), var(--md-default-fg-color--lightest));
}
.footer-quote {
  font-size: 1.1rem;
  font-style: italic;
  color: var(--md-default-fg-color);
  margin-bottom: 1.5rem;
  font-weight: 500;
}
.footer-links {
  display: flex;
  gap: 0.8rem;
  justify-content: center;
  flex-wrap: wrap;
}
.footer-link {
  padding: 0.6rem 1.2rem;
  background: var(--md-default-bg-color);
  border: 1px solid var(--md-default-fg-color--lightest);
  border-radius: 100px;
  font-size: 0.85rem;
  font-weight: 500;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}
.footer-link:hover {
  border-color: var(--accent);
  color: var(--accent);
  transform: translateY(-1px);
}

/* ===== Responsive ===== */
@media (max-width: 900px) {
  .stats-bar { grid-template-columns: repeat(3, 1fr); }
}
@media (max-width: 768px) {
  .hero-banner { padding: 2.5rem 1.5rem; }
  .hero-content {
    flex-direction: column;
    text-align: center;
    gap: 1.5rem;
  }
  .hero-greeting { margin: 0 auto 0.6rem; }
  .hero-name { font-size: 2.1rem; }
  .hero-title { font-size: 0.95rem; }
  .hero-quote { margin: 0.4rem auto 1.2rem; font-size: 0.88rem; }
  .hero-buttons { justify-content: center; }
  
  .stats-bar { grid-template-columns: repeat(2, 1fr); gap: 0.7rem; }
  .stat-card { padding: 1rem 0.6rem; }
  .stat-number { font-size: 1.5rem; }
  .stat-label { font-size: 0.75rem; }
  
  .section-title { font-size: 1.15rem; margin: 2.5rem 0 1rem; }
  .section-title-icon { width: 28px; height: 28px; font-size: 0.95rem; }
  
  .card-grid { grid-template-columns: 1fr; gap: 1rem; }
  .feature-card-head { padding: 1.1rem 1.1rem 0.3rem; }
  .card-body { padding: 0.3rem 1.1rem 1.1rem; }
  
  .article-item { padding: 0.9rem 1rem; gap: 0.8rem; }
  .article-badge { width: 44px; height: 44px; font-size: 1.3rem; border-radius: 10px; }
  .article-title { font-size: 0.92rem; }
  .article-desc { font-size: 0.78rem; }
  .article-meta { display: none; }
  
  .tech-chips { gap: 0.5rem; }
  .tech-chip { padding: 0.45rem 0.85rem; font-size: 0.82rem; }
  
  .home-footer { padding: 2rem 1rem; border-radius: 12px; }
  .footer-quote { font-size: 0.95rem; }
}

@media (max-width: 480px) {
  .hero-banner { padding: 2rem 1rem 2.5rem; border-radius: 12px; }
  .hero-avatar-ring { width: 100px; height: 100px; }
  .hero-avatar { width: 90px; height: 90px; }
  .hero-name { font-size: 1.7rem; }
  .hero-title { font-size: 0.88rem; }
  .hero-quote { font-size: 0.82rem; padding: 0.5rem 0.8rem; max-width: 100%; }
  .hero-buttons { gap: 0.5rem; }
  .hero-btn { flex: 1; justify-content: center; padding: 0.6rem 0.8rem; font-size: 0.85rem; }
}
</style>

<!-- ===== Hero Banner ===== -->
<div class="hero-banner">
  <div class="hero-content">
    <div class="hero-avatar-wrapper">
      <div class="hero-avatar-ring">
        <img src="assets/images/avatar.jpg" alt="Cloaks｜Yiiewang" class="hero-avatar" onerror="this.src='https://avatars.githubusercontent.com/u/39525230'">
      </div>
    </div>
    <div class="hero-text">
      <div class="hero-greeting">
        <span class="hero-greeting-dot"></span>
        <span>Welcome to my blog</span>
      </div>
      <div class="hero-name">Cloaks｜Yiiewang</div>
      <div class="hero-title">
        <span class="hero-title-tag">⛓️ 区块链架构师</span>
        <span class="hero-title-tag">🛠️ 技术探索者</span>
      </div>
      <div class="hero-quote">「为众人抱薪者，不可使其冻毙于风雪」</div>
      <div class="hero-buttons">
        <a href="blog/" class="hero-btn hero-btn-primary">
          <span>📖</span> 开始阅读
        </a>
        <a href="knowledge/" class="hero-btn hero-btn-secondary">
          <span>📚</span> 知识库
        </a>
        <a href="about/" class="hero-btn hero-btn-secondary">
          <span>👋</span> 了解我
        </a>
      </div>
    </div>
  </div>
</div>

<!-- ===== Stats Bar ===== -->
<div class="stats-bar">
  <div class="stat-card">
    <span class="stat-icon">📝</span>
    <div class="stat-number">128</div>
    <div class="stat-label">原创文章</div>
  </div>
  <div class="stat-card">
    <span class="stat-icon">⏰</span>
    <div class="stat-number">5+ 年</div>
    <div class="stat-label">持续写作</div>
  </div>
  <div class="stat-card">
    <span class="stat-icon">🎨</span>
    <div class="stat-number">23</div>
    <div class="stat-label">设计模式</div>
  </div>
  <div class="stat-card">
    <span class="stat-icon">📚</span>
    <div class="stat-number">7</div>
    <div class="stat-label">知识域</div>
  </div>
  <div class="stat-card">
    <span class="stat-icon">⛓️</span>
    <div class="stat-number">3f+1</div>
    <div class="stat-label">PBFT 容错</div>
  </div>
</div>

<!-- ===== Featured Content ===== -->
<div class="section-title">
  <span class="section-title-icon purple">🧭</span>
  <span>探索内容</span>
</div>

<div class="card-grid">
  <a href="blog/" class="feature-card">
    <div class="feature-card-head">
      <div class="feature-icon-wrap purple">📝</div>
      <div class="card-title">技术博客</div>
    </div>
    <div class="card-body">
      <div class="card-desc">开发日常中的思考与实践：后端架构、云原生、区块链…… 每一篇都是真实项目中踩过的坑。</div>
      <div class="card-meta">
        <span class="card-tag">128 篇文章</span>
        <span class="card-arrow">→</span>
      </div>
    </div>
  </a>

  <a href="design-patterns/" class="feature-card">
    <div class="feature-card-head">
      <div class="feature-icon-wrap orange">🧩</div>
      <div class="card-title">设计模式</div>
    </div>
    <div class="card-body">
      <div class="card-desc">23 种 GoF 设计模式的系统讲解。不只是「是什么」，更关注「为什么」和「怎么用」。</div>
      <div class="card-meta">
        <span class="card-tag">23 种模式</span>
        <span class="card-arrow">→</span>
      </div>
    </div>
  </a>

  <a href="ruankao/" class="feature-card">
    <div class="feature-card-head">
      <div class="feature-icon-wrap green">📚</div>
      <div class="card-title">软考笔记</div>
    </div>
    <div class="card-body">
      <div class="card-desc">系统架构师考试的备考笔记。知识点整理 + 真题解析，助你高效通关。</div>
      <div class="card-meta">
        <span class="card-tag">架构师备考</span>
        <span class="card-arrow">→</span>
      </div>
    </div>
  </a>
</div>

<!-- ===== Knowledge Base ===== -->
<div class="section-title">
  <span class="section-title-icon violet">📚</span>
  <span>知识库</span>
</div>

<div class="card-grid">
  <a href="knowledge/tech/" class="feature-card">
    <div class="feature-card-head">
      <div class="feature-icon-wrap blue">🖥️</div>
      <div class="card-title">技术能力</div>
    </div>
    <div class="card-body">
      <div class="card-desc">Go、区块链、架构设计、云原生…… 最核心的硬技能沉淀</div>
      <div class="card-meta">
        <span class="card-tag">10 子域</span>
        <span class="card-arrow">→</span>
      </div>
    </div>
  </a>

  <a href="knowledge/engineering/" class="feature-card">
    <div class="feature-card-head">
      <div class="feature-icon-wrap orange">🔧</div>
      <div class="card-title">工程实践</div>
    </div>
    <div class="card-body">
      <div class="card-desc">项目复盘、踩坑记录、性能优化 —— 实战中提炼的经验</div>
      <div class="card-meta">
        <span class="card-tag">4 子域</span>
        <span class="card-arrow">→</span>
      </div>
    </div>
  </a>

  <a href="knowledge/thinking/" class="feature-card">
    <div class="feature-card-head">
      <div class="feature-icon-wrap green">🧠</div>
      <div class="card-title">思维方法</div>
    </div>
    <div class="card-body">
      <div class="card-desc">学习方法论、问题拆解、决策框架 —— 元能力的系统化</div>
      <div class="card-meta">
        <span class="card-tag">正在建设</span>
        <span class="card-arrow">→</span>
      </div>
    </div>
  </a>
</div>

<!-- ===== Recent Articles ===== -->
<div class="section-title">
  <span class="section-title-icon orange">🔥</span>
  <span>热门推荐</span>
</div>

<div class="article-list">
  <a href="blog/posts/2025/03/3/" class="article-item">
    <div class="article-badge red">❤️</div>
    <div class="article-content">
      <div class="article-title">爱的艺术</div>
      <div class="article-desc">弗洛姆经典著作读书笔记 —— 爱不是找到对的人，而是培养爱的能力</div>
    </div>
    <div class="article-meta">
      <span class="article-type">读书笔记</span>
    </div>
  </a>

  <a href="blog/posts/2025/01/19/" class="article-item">
    <div class="article-badge blue">⚖️</div>
    <div class="article-content">
      <div class="article-title">负载均衡设计思路与实践指南</div>
      <div class="article-desc">轮询、权重、随机…… 分布式系统核心技术详解</div>
    </div>
    <div class="article-meta">
      <span class="article-type">系统设计</span>
    </div>
  </a>

  <a href="design-patterns/creational-patterns/singleton/" class="article-item">
    <div class="article-badge purple">🎯</div>
    <div class="article-content">
      <div class="article-title">单例模式详解</div>
      <div class="article-desc">最简单也最容易用错的设计模式</div>
    </div>
    <div class="article-meta">
      <span class="article-type">设计模式</span>
    </div>
  </a>

  <a href="ruankao/medium%20software%20architecture/" class="article-item">
    <div class="article-badge green">🏗️</div>
    <div class="article-content">
      <div class="article-title">软件架构设计</div>
      <div class="article-desc">架构师必备的设计思维与方法论</div>
    </div>
    <div class="article-meta">
      <span class="article-type">软考</span>
    </div>
  </a>
</div>

<!-- ===== Tech Stack ===== -->
<div class="section-title">
  <span class="section-title-icon teal">🛠️</span>
  <span>技术栈</span>
</div>

<div class="tech-chips">
  <span class="tech-chip"><span class="tech-chip-icon">⛓️</span>区块链</span>
  <span class="tech-chip"><span class="tech-chip-icon">🐹</span>Golang</span>
  <span class="tech-chip"><span class="tech-chip-icon">☕</span>Java</span>
  <span class="tech-chip"><span class="tech-chip-icon">🐳</span>Docker</span>
  <span class="tech-chip"><span class="tech-chip-icon">☸️</span>Kubernetes</span>
  <span class="tech-chip"><span class="tech-chip-icon">🗄️</span>MySQL</span>
  <span class="tech-chip"><span class="tech-chip-icon">🔀</span>分布式系统</span>
  <span class="tech-chip"><span class="tech-chip-icon">🏛️</span>微服务</span>
  <span class="tech-chip"><span class="tech-chip-icon">🔐</span>密码学</span>
  <span class="tech-chip"><span class="tech-chip-icon">📊</span>共识算法</span>
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
