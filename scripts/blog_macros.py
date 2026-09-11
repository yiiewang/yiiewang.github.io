"""博客列表宏：为 Zensical 渲染 /blog/ 的文章卡片列表。

mkdocs-material 的 blog 插件在 MkDocs 侧会自动渲染文章列表；Zensical 0.0.x
尚未实现 blog 插件，因此用 macros 插件（mkdocs-macros 兼容实现）补齐：
`blog_list()` 扫描 docs/blog/posts/ 下的全部 Markdown，解析 frontmatter，
按时间倒序渲染为文章卡片（标题 / 日期 / 分类 / 摘要）。

mkdocs-macros-plugin 与 Zensical macros 接口一致，本模块通过
`import zensical` 检测构建环境；MkDocs 侧宏返回空串，交由 material
blog 插件渲染自己的列表，避免重复。
"""

from __future__ import annotations

import re
from pathlib import Path

import yaml

try:
    import zensical  # noqa: F401

    _IS_ZENSICAL = True
except ImportError:
    _IS_ZENSICAL = False

_POSTS_DIR = Path("blog") / "posts"
_FRONTMATTER_RE = re.compile(r"\A---\n(.*?)\n---", re.S)
_H1_RE = re.compile(r"^#\s+(.+?)\s*$", re.M)
_DESC_LIMIT = 120


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
        "categories": meta.get("categories") or [],
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
    posts.sort(key=lambda p: p["date"], reverse=True)
    return posts


def _render_card(post: dict) -> list[str]:
    """渲染单篇文章卡片（Markdown）。"""
    lines = [f"### [{post['title']}]({post['url']})"]
    bits = [f"`{post['date']}`"]
    if post["categories"]:
        bits.append(" / ".join(str(c) for c in post["categories"]))
    lines.append(" · ".join(bits))
    if post["description"]:
        lines.append("")
        lines.append(post["description"])
    return lines


# ---------------------------------------------------------------------------
# Macros
# ---------------------------------------------------------------------------


def define_env(env) -> None:  # noqa: ANN001, ANN201
    """mkdocs-macros / Zensical macros 入口。"""

    @env.macro
    def blog_list() -> str:
        """渲染全部文章卡片列表（按时间倒序）。"""
        if not _IS_ZENSICAL:
            return ""

        conf = getattr(env, "conf", {}) or {}
        docs_dir = Path(conf.get("docs_dir", "docs"))
        posts = _collect_posts(docs_dir / _POSTS_DIR, docs_dir / "blog")
        if not posts:
            return ""

        out: list[str] = []
        for post in posts:
            out.extend(_render_card(post))
            out.append("")
        return "\n".join(out)
