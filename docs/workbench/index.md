---
date: 2026-08-25
authors:
  - cloaks
categories:
  - 札记
tags:
  - Go
  - SQLite
  - 前端
  - 工程化
  - 开源
comments: true
description: 从零到一做一个个人开发工作台，把文件浏览、Todo 看板、用户管理揉进一个 Go + SQLite 的二进制里。聊聊这半年踩过的坑和想明白的事。
hide:
  - navigation
  - toc
---

<style>
  /* Workbench 封面：复用 home.css 组件，仅微调 hero 布局（无头像，居中） */
  .wb-hero .hero-content { justify-content: center; text-align: center; }
  .wb-hero .hero-text { min-width: 0; }
  .wb-hero .hero-title { justify-content: center; }
  .wb-hero .hero-quote { margin: 0 auto 1.3rem; max-width: 560px; }
  .wb-hero .hero-buttons { justify-content: center; }
  .wb-hero .hero-name { font-size: 2.2rem; }
</style>

<!-- ===== Hero ===== -->
<div class="hero-banner wb-hero">
  <div class="hero-content">
    <div class="hero-text">
      <div class="hero-greeting">
        <span class="hero-greeting-dot"></span>
        <span>Personal Dev Workbench</span>
      </div>
      <div class="hero-name">Workbench</div>
      <div class="hero-title">
        <span class="hero-title-tag">🛠️ 个人开发工作台</span>
        <span class="hero-title-tag">🐹 Go + SQLite</span>
        <span class="hero-title-tag">⚡ 单二进制</span>
      </div>
      <div class="hero-quote">把个人数据全攥在自己手里，想加什么功能就加什么功能。不是又一个"演示项目"，而是打开浏览器就能干活的东西。</div>
      <div class="hero-buttons">
        <a href="https://github.com/yiiewang/workbench" target="_blank" class="hero-btn hero-btn-primary">🐙 GitHub</a>
        <a href="http://159.75.112.77/" target="_blank" class="hero-btn hero-btn-secondary">🚀 在线体验</a>
      </div>
    </div>
  </div>
</div>

半年前我给自己定了个小目标：做一个能真正天天用的个人开发工作台。现在它叫 **Workbench**，跑在 Go + SQLite 上，前端是 Vite + Vue3 + TypeScript。文件浏览、Todo 看板、用户与组织管理、分享下载，全在一个二进制里。

<!-- more -->

## 为什么不用现成的

市面上不是没有类似的东西。但用别人的工具，总有几个绕不开的坎：

- **数据不在自己手里**。笔记、任务、文件散落在各家云端，想导出、想迁移，麻烦。
- **改不动**。想要个"按组织隔离文件"的功能，得等官方排期，或者自己 fork 一坨。
- **太重**。为了看个文件、记个任务，要起一整套服务，杀鸡用牛刀。

所以干脆自己写。**自己写的好处是：每一个功能都是"我需要的"，而不是"别人觉得我需要的"。**

## 技术选型：Go + SQLite 就够了

选型上我几乎没纠结。

**后端 Go** ——单二进制、无依赖、部署就是 `scp` 一个文件过去。**存储 SQLite** ——一个 `workbench.db` 文件搞定所有数据，WAL 模式，备份就是复制文件。**前端 Vite + Vue3 + TS** ——组件化开发，`//go:embed` 把构建产物直接嵌进二进制，前端后端一个包。

这套组合最大的好处是 **心智负担极低**。没有微服务、没有消息队列、没有容器编排，一个进程跑完所有事。对个人工具来说，简单就是最大的可靠。

## 现在它能做什么

<div class="card-grid">

  <div class="feature-card">
    <div class="feature-card-head">
      <div class="feature-icon-wrap blue">📁</div>
      <div class="card-title">文件浏览器</div>
    </div>
    <div class="card-body">
      <div class="card-desc">目录树、Tab 编辑、Markdown / JSON / 代码预览，跨文件链接跳转。</div>
    </div>
  </div>

  <div class="feature-card">
    <div class="feature-card-head">
      <div class="feature-icon-wrap orange">✅</div>
      <div class="card-title">Todo 看板</div>
    </div>
    <div class="card-body">
      <div class="card-desc">多用户多组织任务管理，冲突检测、版本同步、多设备协作。</div>
    </div>
  </div>

  <div class="feature-card">
    <div class="feature-card-head">
      <div class="feature-icon-wrap green">👥</div>
      <div class="card-title">用户与组织管理</div>
    </div>
    <div class="card-body">
      <div class="card-desc">角色体系、组织切换、功能开关，每个用户在每个组织有独立权限。</div>
    </div>
  </div>

  <div class="feature-card">
    <div class="feature-card-head">
      <div class="feature-icon-wrap violet">🔗</div>
      <div class="card-title">下载与分享</div>
    </div>
    <div class="card-body">
      <div class="card-desc">带访问次数、时间范围、密码的分享链接，文件夹递归包含子目录。</div>
    </div>
  </div>

  <div class="feature-card">
    <div class="feature-card-head">
      <div class="feature-icon-wrap teal">📊</div>
      <div class="card-title">访问统计</div>
    </div>
    <div class="card-body">
      <div class="card-desc">按访问者与页面统计，持久化到 SQLite，数据完全自持。</div>
    </div>
  </div>

</div>

## 技术栈

<div class="tech-chips">
  <span class="tech-chip"><span class="tech-chip-icon">🐹</span>Go</span>
  <span class="tech-chip"><span class="tech-chip-icon">🗄️</span>SQLite</span>
  <span class="tech-chip"><span class="tech-chip-icon">⚡</span>Vite</span>
  <span class="tech-chip"><span class="tech-chip-icon">🟢</span>Vue3</span>
  <span class="tech-chip"><span class="tech-chip-icon">🔷</span>TypeScript</span>
  <span class="tech-chip"><span class="tech-chip-icon">🧩</span>Element Plus</span>
  <span class="tech-chip"><span class="tech-chip-icon">📦</span>golang-migrate</span>
  <span class="tech-chip"><span class="tech-chip-icon">🔐</span>HMAC-SHA256</span>
  <span class="tech-chip"><span class="tech-chip-icon">📝</span>slog</span>
</div>

## 从"能用"到"好用"：几个关键决策

### 账号体系：从字符串主键到整数自增

最早我图省事，用业务名（比如 `yiiewang`）当主键。后来发现这是个坑——改名、重名、跨组织引用，全都会出问题。最后老老实实迁到整数自增 `id`，业务名只当展示字段。

**教训：主键就该是"无意义"的，有意义的字段永远不该当主键。**

### 权限模型：用户、组织、功能三层

一开始只有"登录/未登录"两态，后来加了组织，再后来加了角色（admin / org_admin / user）和功能开关。现在每个用户在每个组织有独立的功能开关，文件按用户物理隔离。

这套模型不是一次设计出来的，是 **被需求推着长出来的**。每次加一层，都对应一个真实的使用场景。

### 安全：默认严格

路径穿越、符号链接逃逸、CORS 沙箱、Token 鉴权……这些不是"锦上添花"，是 **默认就该做对** 的事。个人工具暴露在公网上，一个洞就是整个数据目录。

## 踩过的坑

- **SQLite 迁移**：`ALTER TABLE ADD COLUMN` 不能带外键，旧库升级时索引依赖新列会直接崩。最后上了 golang-migrate 版本化迁移，每个 schema 变更加一个脚本。
- **前端状态**：KeepAlive + onMounted 的组合下，任何依赖 onMounted 的"模式切换重置"逻辑都会因组件复用而不执行，必须在路由 watch 里显式收敛。
- **构建链路**：`//go:embed` 要求 `frontend/dist` 存在，fresh clone 必须先 `make all` 才能编译，否则直接报错。

## 写在最后

写这个工作台，最大的收获不是代码，是 **想清楚了一个问题：工具应该长在人的工作流里，而不是让人去适应工具**。

它现在还在持续迭代。如果你也想要一个"数据在自己手里、想改就改"的个人工具，欢迎来 [GitHub](https://github.com/yiiewang/workbench) 看看，或者直接体验[在线实例](http://159.75.112.77/)。

---

*最后更新：2026-08-25*