# Changelog

本文件记录本项目的所有重要变更，格式遵循 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)。

## [Unreleased]

### Added

- 构建工具 Zensical 从 `0.0.66` 升级到 `0.1.3`。版本仍只固定在 `.github/workflows/Dockerfile`，`Makefile` 的 `PINNED_VERSION` 与 `ci.yml` 的 `Resolve image tag` 步骤自动推导镜像 tag；`zensical --version` 输出格式在新版仍是纯版本号，`Verify image version` 校验不会误判。
- 启用官方 `exclude_docs`（0.0.70+）排除图源 `_*.ditaa.txt` / `_*.svgbob.txt`，7 个图源不再随站点发布，生成的 7 个 `_*.svg` 不受影响。配置校验要求字符串形式的多行 glob，写成数组会报 `'exclude_docs' must be a str`。
- 启用官方 `llmstxt` 插件（0.0.67+）：按首页 / 博客 / 知识库 / 软考·架构师 / 设计模式 / 实验室六个命名分节生成 `llms.txt` 索引、约 307 个页面 Markdown 副本，并用 `full_output = "llms-full.txt"` 输出全文；`markdown_description` 保留站点介绍。配置校验要求 `sections` 为 mapping（分节名 → 页面路径列表），`title` / `pages` 数组形式会报 `sections must be a mapping`。
- 启用官方 `rss` 插件（0.0.65+）：`match_path = "blog/posts/.*"` 限定只收录博文，`date_from_meta.as_creation = "date"` 配合 `datetime_format = "%Y-%m-%d"` 让 `pubDate` 取 frontmatter 日期并按最新 20 篇排序；产物为 `feed_rss_created.xml`、`feed_rss_updated.xml`、两个 JSON Feed 与 `rss.xsl`。未配 `date_from_meta` 时所有日期会退化为构建时间、且收录最老的 20 篇。
- 启用官方 `social` 插件（0.0.67+）按页生成社交卡。`theme.logo` 是 ICO，插件不支持该格式，直接启用会报 `RuntimeError: unsupported social background format`；需用 `[project.plugins.social.cards_layout_options] logo = "docs/assets/images/logo.png"` 覆盖，且该路径从项目根解析（不是 docs_dir）。实测 0.1.3 的原生实现无需额外 Python / 系统依赖。
- 新增根目录 `CHANGELOG.md`，记录本次升级与官方插件迁移。

### Changed

- `scripts/gen_seo.py` 的 llms 两段逻辑由官方 `llmstxt` 接管、RSS 逻辑由官方 `rss` 接管；构建前不再向 `docs/` 生成 `llms.txt` / `llms-full.txt` / `rss.xml`，改为构建时直接输出到 `site/`。
- 站内 RSS 入口统一指向 `feed_rss_created.xml`：`zensical.toml` 的 `extra.social`、`overrides/main.html` 的 RSS autodiscovery，并恢复 `docs/index.md` 首页 RSS 订阅按钮。对外旧地址 `/rss.xml` 不再提供（未加 redirect：静态 HTML 重定向对 RSS 阅读器无效）。
- 分享预览从全站共用一张 `og-card.png` 改为官方 `social` 插件按页生成 PNG；`overrides/main.html` 移除手写的 `og:image` / `og:image:width` / `og:image:height` / `twitter:image` 以免与插件输出重复，`og_image` 变量保留给 JSON-LD 的 `image` 字段（结构化数据仍引用统一固定卡）。`scripts/og_card.py` 与 `docs/assets/images/og-card.png` 保留。
- `Makefile` 移除 `seo` 目标，`dev` / `serve` / `build` / `prod` 直接依赖 `ensure-image`；`.github/workflows/ci.yml` 移除 "Generate RSS feed" 步骤；`.gitignore` 移除三个生成物规则。
- `zensical.toml` 插件现状注释、`Makefile` 图源约定说明、CI 步骤注释同步更新；`README.md` / `deploy/github-workflow.md` 的版本字面量统一到 `0.1.3`。
- 博客插件保持不接入，`scripts/blog_macros.py` 与博客首页、年度页原样保留：官方 `blog` 插件不提供 `_CATEGORY_META` 自定义分类元数据、`annual_list()` 按月年度归档与「未收录分类」硬门禁。

### Fixed

- 修复 0.1.3 将 `config.site_url` 规范化为带尾斜杠后，`overrides/main.html` 中 `{{ config.site_url }}/...` 拼接产生双斜杠的问题：RSS autodiscovery 与 JSON-LD 使用的 `og_image` 现在输出单斜杠绝对 URL；`og:image` 标签已改由官方 social 插件输出。
- 修正 README 中「Zensical 0.0.x 尚未实现 blog 插件」的过时描述：官方 blog 插件自 0.0.64 起可用，本站是主动选择不接入。

### Removed

- 删除 `scripts/gen_seo.py`（自建 SEO / RSS 生成脚本）与 `docs/rss.xml`。
- 从 git 索引移除此前被跟踪的 `docs/llms.txt` / `docs/llms-full.txt` 生成物（未提交，删除随下次提交生效）。
