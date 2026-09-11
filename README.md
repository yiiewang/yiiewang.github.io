# Cloaks｜Yiiewang Blog

> 欢迎来到我的博客！这里是我分享个人见解、技术文章、生活点滴和创意的地方。

站点使用 **Zensical** 构建（Material for MkDocs 团队的新一代静态站点生成器），
配置沿用 `mkdocs.yml`（Zensical 原生读取 MkDocs 配置格式）。

## 🚀 快速开始

### 方式一：Docker 启动（推荐）

```bash
# 构建镜像（首次或 Dockerfile 改动后执行）
make docker-build

# 启动开发服务器
make docker-serve
```

**访问地址**：http://localhost:8000

### 方式二：本地启动

```bash
# 安装 Zensical（版本需与 Dockerfile / CI 保持一致）
pip install "zensical==0.0.60"

# 开发服务器（增量构建 + 热重载）
make serve

# 生产构建（产物在 site/）
make build
```

## 📝 项目结构

```
.
├── docs/                  # 文档目录
│   ├── blog/              # 博客文章
│   │   ├── index.md       # 博客索引（调用 blog_list 宏）
│   │   └── posts/         # 文章源文件
│   ├── assets/            # 静态资源
│   └── index.md           # 首页
├── overrides/             # 主题模板覆盖
├── scripts/
│   └── blog_macros.py     # 博客列表宏（Zensical macros 插件加载）
├── mkdocs.yml             # 站点配置（唯一配置源）
├── Makefile               # 构建入口
└── .github/workflows/     # CI 与 Dockerfile
```

## 🔧 配置说明

- **构建工具**：Zensical 0.0.60
- **语言**：中文
- **功能特性**：
  - 博客系统（文章列表由 `scripts/blog_macros.py` 的 `blog_list()` 宏渲染）
  - 代码高亮 / 数学公式（KaTeX）/ 图表（Mermaid）
  - 图片灯箱（glightbox）
  - 标签系统
  - 全站搜索

### 博客列表宏

Zensical 0.0.x 尚未实现 Material 的 blog 插件，`docs/blog/index.md` 通过
`{{ blog_list() }}` 调用 `scripts/blog_macros.py` 中的宏，扫描
`docs/blog/posts/` 下的全部文章并按时间倒序渲染卡片。

宏通过 `plugins: macros` 加载，配置项：

- `module_name: scripts/blog_macros` —— 指向宏模块（默认值是 `main`，
  即要求项目根目录有 `main.py`，这里改到 `scripts/` 下）
- `render_by_default: false` —— 白名单模式，只有声明了 `render_macros: true`
  的页面才会走 Jinja 渲染，避免代码块中的 `{{ }}` 被误渲染

## 📖 写作规范

- 博客文章存放在 `docs/blog/posts/{year}/{month}/` 目录
- 文件名使用日期格式：`{day}.md`
- 同一天多篇文章使用 `{day}-{序号}.md`
- 参考 [blog-writing 规则](.codebuddy/rules/blog-writing.mdc)

## 🚢 部署

推送到 `master` / `main` 分支后，GitHub Actions 会自动：

1. 安装 Zensical 并执行 `zensical build -f mkdocs.yml --clean`
2. 将 `site/` 发布到 `gh-pages` 分支（GitHub Pages 源）

配置文件：`.github/workflows/ci.yml`

> 版本一致性：`zensical==0.0.60` 同时固定在 `.github/workflows/Dockerfile`
> 与 `.github/workflows/ci.yml`，升级时需同步修改。

## 🐳 Docker 优化建议

### 提升构建速度

Zensical 自带差分构建缓存（`.cache/`），重复构建只需数秒：

```bash
# 保留缓存卷，跨容器复用构建缓存
docker run --rm -v $(pwd):/docs -v zensical-cache:/docs/.cache \
  cloaks/zensical:0.0.60 build -f mkdocs.yml
```

### 限制资源使用

```bash
docker run --rm \
  --cpus="2.0" \
  --memory="2g" \
  -v $(pwd):/docs \
  -p 8000:8000 \
  cloaks/zensical:0.0.60 \
  serve -f mkdocs.yml -a 0.0.0.0:8000
```

## 📄 许可证

Copyright &copy; 2018 - 2026 Cloaks

## 🔗 相关链接

- [Zensical 官网](https://zensical.org/)
- [Zensical 路线图](https://zensical.org/about/roadmap/)
- [博客](https://cloaks.cn/blog/)
