---
title: 首页
hide:
  - navigation
  - toc
---

<style>
.md-typeset h1 { display: none; }

/* Hero Section */
.hero-section {
  text-align: center;
  padding: 3rem 1rem 2rem;
  background: linear-gradient(135deg, rgba(var(--md-primary-fg-color--light-rgb), 0.05) 0%, transparent 100%);
  border-radius: 16px;
  margin-bottom: 2rem;
}
.hero-avatar {
  width: 100px;
  height: 100px;
  border-radius: 50%;
  border: 3px solid var(--md-primary-fg-color);
  margin-bottom: 1rem;
  box-shadow: 0 4px 20px rgba(0,0,0,0.1);
}
.hero-title {
  font-size: 2rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
  background: linear-gradient(135deg, var(--md-primary-fg-color) 0%, var(--md-accent-fg-color) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}
.hero-subtitle {
  font-size: 1rem;
  color: var(--md-default-fg-color--light);
  margin-bottom: 1.5rem;
}
.hero-quote {
  font-style: italic;
  color: var(--md-default-fg-color--light);
  font-size: 0.95rem;
  padding: 0.8rem 1.5rem;
  background: var(--md-default-fg-color--lightest);
  border-radius: 8px;
  display: inline-block;
  max-width: 400px;
}
.hero-buttons {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
  margin-top: 1.5rem;
}

/* Stats */
.stats-row {
  display: flex;
  justify-content: center;
  gap: 2rem;
  padding: 1.5rem 0;
  flex-wrap: wrap;
}
.stat-item {
  text-align: center;
  padding: 0 1rem;
}
.stat-number {
  font-size: 1.6rem;
  font-weight: 700;
  color: var(--md-primary-fg-color);
  line-height: 1.2;
}
.stat-label {
  font-size: 0.8rem;
  color: var(--md-default-fg-color--light);
}

/* Section */
.section-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 2.5rem 0 1rem;
}
.section-header h2 {
  margin: 0 !important;
  border: none !important;
}
</style>

<div class="hero-section">
  <img src="assets/avatar.jpg" alt="Cloaks" class="hero-avatar" onerror="this.src='https://avatars.githubusercontent.com/u/39525230'">
  <div class="hero-title">嗨，我是 Cloaks 👋</div>
  <div class="hero-subtitle">一个热爱技术、乐于分享的开发者</div>
  <div class="hero-quote">「为众人抱薪者，不可使其冻毙于风雪」</div>
  <div class="hero-buttons">
    <a href="blog/" class="md-button md-button--primary">开始阅读</a>
    <a href="about/" class="md-button">了解更多</a>
  </div>
</div>

<div class="stats-row">
  <div class="stat-item">
    <div class="stat-number">100+</div>
    <div class="stat-label">原创文章</div>
  </div>
  <div class="stat-item">
    <div class="stat-number">4 年</div>
    <div class="stat-label">持续写作</div>
  </div>
  <div class="stat-item">
    <div class="stat-number">23 种</div>
    <div class="stat-label">设计模式</div>
  </div>
  <div class="stat-item">
    <div class="stat-number">50+</div>
    <div class="stat-label">重构技巧</div>
  </div>
</div>

---

## :material-compass: 探索内容

这里是我的技术知识库，涵盖日常开发实践、经典设计模式、代码优化技巧等内容。

<div class="grid cards" markdown>

-   :material-post-outline:{ .lg .middle } **技术博客**

    ---

    开发日常中的思考与实践：后端架构、云原生、区块链...
    
    每一篇都是真实项目中踩过的坑、总结的经验。

    [:octicons-arrow-right-24: 浏览文章](blog/index.md)

-   :material-palette:{ .lg .middle } **设计模式**

    ---

    23 种 GoF 设计模式的系统讲解。
    
    不只是「是什么」，更关注「为什么」和「怎么用」。

    [:octicons-arrow-right-24: 开始学习](design-patterns/index.md)

-   :material-wrench:{ .lg .middle } **代码重构**

    ---

    如何把「能跑」的代码变成「好」的代码？
    
    识别代码异味，掌握重构手法，写出优雅代码。

    [:octicons-arrow-right-24: 查看指南](refactoring/index.md)

-   :material-certificate:{ .lg .middle } **软考笔记**

    ---

    系统架构师考试的备考笔记。
    
    知识点整理 + 真题解析，助你高效通关。

    [:octicons-arrow-right-24: 复习资料](ruankao/index.md)

</div>

---

## :material-star: 推荐阅读

如果你是第一次来，不妨从这些内容开始：

=== ":material-fire: 热门文章"

    | 文章 | 简介 |
    |------|------|
    | [单例模式详解](design-patterns/creational-patterns/singleton.md) | 最简单也最容易用错的设计模式 |
    | [什么是技术债务](refactoring/what-is-refactoring/technical-debt.md) | 理解代码腐化的根源 |
    | [软件架构设计](ruankao/medium%20software%20architecture.md) | 架构师必备的设计思维 |

=== ":material-map-marker-path: 学习路径"

    **入门设计模式**
    
    ```
    单例模式 → 工厂模式 → 策略模式 → 观察者模式
    ```
    
    **学习代码重构**
    
    ```
    技术债务 → 代码异味 → 重构技巧 → 实践应用
    ```

=== ":material-tools: 我的技术栈"

    | 领域 | 常用技术 |
    |------|----------|
    | 后端开发 | Golang · Java · Python |
    | 区块链 | 智能合约 · 分布式账本 |
    | 云原生 | Docker · Kubernetes |
    | 数据库 | MySQL · Redis · MongoDB |

---

<div style="text-align: center; padding: 2rem 0;">
  <p style="color: var(--md-default-fg-color--light); margin-bottom: 1rem;">
    💬 有问题或想法？欢迎在文章下方留言交流
  </p>
  <a href="https://github.com/cloakscn/cloakscn.github.io" target="_blank" style="color: var(--md-default-fg-color--light); font-size: 0.85rem;">
    :fontawesome-brands-github: 本站源码
  </a>
</div>
