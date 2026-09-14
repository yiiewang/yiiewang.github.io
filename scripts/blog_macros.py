"""博客列表宏：为 Zensical 渲染 /blog/ 的文章卡片列表。

Zensical 0.0.x 尚未实现 blog 插件，因此用 macros 插件补齐文章列表：
扫描 docs/blog/posts/ 下的全部 Markdown，解析 frontmatter，按时间倒序
渲染文章卡片（标题 / 日期 / 分类 / 摘要）。

宏接口（通过 `plugins.macros` 加载，见 zensical.toml，入口 define_env）：
- blog_index()               博客首页主体：分类导航（含篇数）+ 按分类分组的
                             文章列表，全部由文章 frontmatter 动态生成。
                             已知分类按 _CATEGORY_META 的顺序与文案展示；
                             未收录的新分类自动附加到末尾（默认图标/锚点）。
- blog_list(category="X")   指定分类的文章卡片，按时间倒序；category 为空
                             时渲染全部文章（卡片附带分类行）。
"""

from __future__ import annotations

import re
from pathlib import Path

import yaml

_POSTS_DIR = Path("blog") / "posts"
_FRONTMATTER_RE = re.compile(r"\A---\n(.*?)\n---", re.S)
_H1_RE = re.compile(r"^#\s+(.+?)\s*$", re.M)
_DESC_LIMIT = 120
_LATEST_COUNT = 10  # 首页「最新博客」置顶区块展示的篇数

# 已知分类的展示顺序、锚点、图标与一句描述（数据驱动；未收录分类自动 fallback）
_CATEGORY_META: list[dict] = [
    {
        "name": "Go 语言",
        "anchor": "go",
        "icon": "language-go",
        "desc": "语言机制、并发原语与工程实践——从 `nil` 语义到无锁编程。",
    },
    {
        "name": "系统与架构",
        "anchor": "systems",
        "icon": "server",
        "desc": "存储、并发与内存模型——从一个落盘方案到六百万 TPS 的完整因果链。",
    },
    {
        "name": "区块链",
        "anchor": "blockchain",
        "icon": "link-variant",
        "desc": "共识算法、密码学与长安链（ChainMaker）实战——含压测报告与测试手册。",
    },
    {
        "name": "算法与数据结构",
        "anchor": "algorithms",
        "icon": "calculator",
        "desc": "刷题时代的沉淀：滑动窗口、动态规划、位运算……",
    },
    {
        "name": "DevOps 与工具",
        "anchor": "devops",
        "icon": "console",
        "desc": "Linux、Shell、CI/CD 与远程开发环境——把工具用顺手。",
    },
    {
        "name": "AI 实践",
        "anchor": "ai",
        "icon": "robot",
        "desc": "多智能体协同、AI 辅助工程与 LangChain——人机协作的一线记录。",
    },
    {
        "name": "投资与交易",
        "anchor": "investing",
        "icon": "chart-line",
        "desc": "趋势跟踪、量化策略与宏观经济——把交易也当工程做。",
    },
    {
        "name": "读书与思考",
        "anchor": "reading",
        "icon": "book-open-variant",
        "desc": "技术书的动手笔记，与非技术书的随想。",
    },
    {
        "name": "软考",
        "anchor": "ruankao",
        "icon": "school",
        "desc": "系统架构师（高级）备考：论文范文与高频题型。",
    },
    {
        "name": "札记",
        "anchor": "notes",
        "icon": "notebook",
        "desc": "无处安放的杂记：Java、COM、项目管理、汇报方法……",
    },
]

# 未收录分类的锚点生成（与 pymdownx 的 unicode slugify 规则一致，仅作 fallback）
try:  # pragma: no cover - 构建环境必装有 pymdownx
    from pymdownx.slugs import slugify as _pmd_slugify

    _pmd = _pmd_slugify(case="lower")

    def _slugify(value: str) -> str:
        return _pmd(value, "-") or "cat"

except ImportError:  # pragma: no cover
    def _slugify(value: str) -> str:
        slug = re.sub(r"[^a-z0-9]+", "-", value.lower()).strip("-")
        return slug or "cat"


# ---------------------------------------------------------------------------
# Internal helpers
# ---------------------------------------------------------------------------


def _parse_post(path: Path, blog_root: Path) -> dict | None:
    """解析单篇博文，返回元数据；草稿返回 None。"""
    text = path.read_text(encoding="utf-8")
    meta: dict = {}
    if m := _FRONTMATTER_RE.match(text):
        try:
            meta = yaml.safe_load(m.group(1)) or {}
        except yaml.YAMLError:
            meta = {}
    if not isinstance(meta, dict) or meta.get("draft"):
        return None

    title = meta.get("title")
    if not title:
        h1 = _H1_RE.search(text)
        title = h1.group(1).strip() if h1 else path.stem

    date = meta.get("date")
    date = str(date)[:10] if date else ""

    rel = path.relative_to(blog_root)
    url = rel.with_suffix("").as_posix() + "/"

    desc = " ".join(str(meta.get("description") or "").split())
    if len(desc) > _DESC_LIMIT:
        desc = desc[: _DESC_LIMIT - 1].rstrip() + "…"

    return {
        "title": str(title),
        "date": date,
        "url": url,
        "categories": [str(c) for c in (meta.get("categories") or [])],
        "tags": [str(t) for t in (meta.get("tags") or [])],
        "pinned": bool(meta.get("pin")),  # frontmatter pin: true → 首页置顶
        "description": desc,
    }


def _collect_posts(posts_root: Path, blog_root: Path) -> list[dict]:
    if not posts_root.is_dir():
        return []
    posts = []
    for path in posts_root.rglob("*.md"):
        post = _parse_post(path, blog_root)
        if post:
            posts.append(post)
    # 时间倒序；同日按 URL 倒序，保证构建结果稳定
    posts.sort(key=lambda p: (p["date"], p["url"]), reverse=True)
    return posts


def _render_card(post: dict, show_categories: bool = True) -> list[str]:
    """渲染单篇文章卡片（Markdown）。

    meta 行：`日期` · 分类 · #tag #tag——分类仅跨分类区块（置顶/最新）显示，
    分类分节内以 tags 代替（同节文章分类相同，tags 更有区分度）。
    tag 含空格时以连字符连接，保证 # 形式不断裂。
    """
    lines = [f"### [{post['title']}]({post['url']})"]
    bits = [f"`{post['date']}`"]
    if show_categories and post["categories"]:
        bits.append(" / ".join(post["categories"]))
    if post["tags"]:
        bits.append(" ".join(f"#{t.replace(' ', '-')}" for t in post["tags"]))
    lines.append(" · ".join(bits))
    if post["description"]:
        lines.append("")
        lines.append(post["description"])
    return lines


# ---------------------------------------------------------------------------
# Macros
# ---------------------------------------------------------------------------


def define_env(env) -> None:  # noqa: ANN001, ANN201
    """Zensical macros 入口。"""

    def _load() -> list[dict]:
        conf = getattr(env, "conf", {}) or {}
        docs_dir = Path(conf.get("docs_dir", "docs"))
        return _collect_posts(docs_dir / _POSTS_DIR, docs_dir / "blog")

    @env.macro
    def blog_list(category: str | None = None) -> str:
        """渲染文章卡片列表（按时间倒序）。

        category 为空时渲染全部文章（卡片带分类行）；
        指定分类时只渲染该分类，卡片不再重复显示分类。
        """
        posts = _load()
        if category is not None:
            posts = [p for p in posts if category in p["categories"]]
        if not posts:
            return ""

        out: list[str] = []
        for post in posts:
            out.extend(_render_card(post, show_categories=category is None))
            out.append("")
        return "\n".join(out)

    @env.macro
    def blog_index() -> str:
        """渲染博客首页主体：分类导航 + 按分类分组的文章列表。

        分组与篇数完全由文章 frontmatter 的 categories 动态聚合：
        - 已知分类按 _CATEGORY_META 的顺序/图标/描述展示
        - 未收录的新分类自动附加到末尾（默认图标，锚点由分类名 slug 化）
        """
        posts = _load()
        by_cat: dict[str, list[dict]] = {}
        for p in posts:
            for c in p["categories"]:
                by_cat.setdefault(c, []).append(p)
        if not by_cat:
            return ""

        meta_by_name = {m["name"]: m for m in _CATEGORY_META}

        def _meta(name: str) -> dict:
            if name in meta_by_name:
                return meta_by_name[name]
            return {
                "name": name,
                "anchor": _slugify(name),
                "icon": "bookmark-outline",
                "desc": "",
            }

        # 排序：已知分类按预定义顺序，未知分类按篇数降序附加
        ordered = [m["name"] for m in _CATEGORY_META if m["name"] in by_cat]
        unknown = sorted(
            (c for c in by_cat if c not in meta_by_name),
            key=lambda c: (-len(by_cat[c]), c),
        )
        if unknown:
            # 硬门禁：中止构建，本地与 CI 行为一致。
            # SystemExit 继承 BaseException，可穿透 jinja2 / zensical 的
            # except Exception；在 serve 的 watch 线程里只终止本次重建，
            # 不会杀掉 dev server。
            # 注：zensical 的 Rust 边界会将其包装为
            # "RuntimeError: Python error: SystemExit: <消息>" 并附堆栈，
            # 错误消息位于第一行——已接受该输出形态。
            lines = ["博客存在未收录分类，构建中止："]
            for c in unknown:
                lines.append(f"  「{c}」（{len(by_cat[c])} 篇）：")
                for post in by_cat[c]:
                    lines.append(f"    - {post['title']}（{post['url']}）")
            lines.append(
                "处理方式：在 scripts/blog_macros.py 的 _CATEGORY_META 中补充"
                "对应条目，或修正文章 frontmatter 的分类名。"
            )
            raise SystemExit("\n".join(lines))
        ordered += unknown

        out: list[str] = []
        # 置顶博客（pin: true）与最新博客（时间逆序前 N 篇，排除已置顶的避免重复）
        pinned = [p for p in posts if p["pinned"]]
        latest = [p for p in posts if not p["pinned"]][:_LATEST_COUNT]
        # 分类导航（「置顶」「最新」入口置前）
        nav = []
        if pinned:
            nav.append(f"[**置顶** ({len(pinned)})](#pinned)")
        if latest:
            nav.append(f"[**最新** ({len(latest)})](#latest)")
        nav += [
            f"[**{n}** ({len(by_cat[n])})](#{_meta(n)['anchor']})" for n in ordered
        ]
        out.append(" · ".join(nav))
        # 置顶博客（posts 已按时间逆序，置顶文章保持该顺序）
        if pinned:
            out.append("")
            out.append("## :material-pin: 置顶博客 { #pinned }")
            out.append("")
            out.append("精选文章，值得优先一读。")
            for post in pinned:
                out.append("")
                out.extend(_render_card(post, show_categories=True))
        # 最新博客
        if latest:
            out.append("")
            out.append("## :material-clock-outline: 最新博客 { #latest }")
            out.append("")
            out.append(f"最近 {_LATEST_COUNT} 篇，按时间逆序。")
            for post in latest:
                out.append("")
                out.extend(_render_card(post, show_categories=True))
        # 各分类分节
        for name in ordered:
            m = _meta(name)
            out.append("")
            out.append(f"## :material-{m['icon']}: {name} {{ #{m['anchor']} }}")
            if m["desc"]:
                out.append("")
                out.append(m["desc"])
            for post in by_cat[name]:
                out.append("")
                out.extend(_render_card(post, show_categories=False))
        return "\n".join(out) + "\n"
