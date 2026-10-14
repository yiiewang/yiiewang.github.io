#!/usr/bin/env python3
"""生成 SEO / AI 检索文件（构建前运行，产物进 docs/ 随站点发布）。

产物（构建时生成，不入库，见 .gitignore——避免手写内容随站点演进腐化）：
- docs/llms.txt        llmstxt.org 规范：站点概览 + 栏目 + 最新文章 + Optional
- docs/llms-full.txt   全站页面完整索引：全部博文（按分类）+ 导航内全部页面
- docs/rss.xml         RSS 2.0 订阅源（最新 20 篇），搜索引擎 / AI 助手发现新内容的主通道

数据源（单一事实源，不重复维护）：
- scripts/blog_macros.py  文章 frontmatter 解析、分类元数据（与博客卡片同一套逻辑）
- zensical.toml           站点信息与导航树（tomllib，Python 3.11+）

用法（容器内执行，宿主机无需 Python，见 Makefile 的 seo 目标）：
  docker run --rm -v "$PWD":/docs --entrypoint python3 cloaks/zensical:<ver> scripts/gen_seo.py
"""

from __future__ import annotations

import email.utils
import re
import sys
import tomllib
import urllib.parse
import xml.sax.saxutils as sx
from datetime import datetime, timezone
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(ROOT / "scripts"))

from blog_macros import _CATEGORY_META, _FRONTMATTER_RE, _collect_posts  # noqa: E402

RSS_ITEM_COUNT = 20   # rss.xml 收录的最新文章数
LLMS_LATEST = 10      # llms.txt「最新文章」展示数

# llms.txt 的栏目速览（目录名是结构性路径，比 nav 显示名稳定；URL 动态拼接）
_SECTIONS = [
    ("博客", "blog/", "全部技术文章，按分类组织"),
    ("知识库", "knowledge/", "技术、工程、思维、领域、人文、工具、职业七大能力域"),
    ("软考·架构师", "ruankao/", "系统架构师（高级）备考笔记与论文范文"),
    ("设计模式", "design-patterns/", "23 种设计模式详解"),
    ("实验室", "lab/", "项目示例、电子书与实用工具"),
]

# llms-full 页面索引的顶级分组顺序（首页的 about 并入「其他」）
_TOP_ORDER = ["知识库", "学习笔记", "实验室"]


def _abspath(site: str, path: str) -> str:
    """拼接绝对 URL 并对路径做百分号编码（空格等）。"""
    return site + "/" + urllib.parse.quote(path.lstrip("/"), safe="/:@")


def _nav_pages(items, prefix: tuple[str, ...] = ()):
    """展平 nav 树 → [(路径标签元组, 文档相对路径)]。

    - {"标签" = "a/b.md"}            → ((标签,), "a/b.md")
    - {"标签" = [子树]}              → 递归，前缀追加标签
    - "a/b.md"（无显示名，如 ruankao/index.md） → 标签由文档 frontmatter title 兜底
    """
    out = []
    for item in items:
        if isinstance(item, str):
            out.append((prefix, item))
        elif isinstance(item, dict):
            (label, val), = item.items()
            if isinstance(val, str):
                out.append((prefix + (label,), val))
            else:
                out.extend(_nav_pages(val, prefix + (label,)))
    return out


def _doc_meta(rel: str) -> dict:
    """读取 docs/<rel> 的 frontmatter（容忍缺失/解析失败）。"""
    path = ROOT / "docs" / rel
    try:
        text = path.read_text(encoding="utf-8")
    except OSError:
        return {}
    if m := _FRONTMATTER_RE.match(text):
        try:
            meta = yaml.safe_load(m.group(1)) or {}
            return meta if isinstance(meta, dict) else {}
        except yaml.YAMLError:
            return {}
    return {}


def _doc_title(rel: str) -> str:
    """无 nav 显示名条目的标题兜底：frontmatter title → H1 → 文件名。"""
    meta = _doc_meta(rel)
    if meta.get("title"):
        return str(meta["title"])
    try:
        text = (ROOT / "docs" / rel).read_text(encoding="utf-8")
    except OSError:
        return Path(rel).stem
    if m := re.search(r"^#\s+(.+?)\s*$", text, re.M):
        return m.group(1)
    return Path(rel).stem


def _doc_desc(rel: str) -> str:
    """页面简介：description / summary 字段，压成单行。"""
    meta = _doc_meta(rel)
    raw = meta.get("description") or meta.get("summary") or ""
    return " ".join(str(raw).split())


def _post_url(site: str, post: dict) -> str:
    """blog_macros 的 post['url'] 相对 blog/ 目录，需补前缀。"""
    return _abspath(site, f"blog/{post['url']}")


def _ordered_categories(posts: list[dict]) -> list[str]:
    """分类排序：已知分类按 _CATEGORY_META 顺序，未知按篇数降序追加。"""
    known = [m["name"] for m in _CATEGORY_META]
    by_cat: dict[str, int] = {}
    for p in posts:
        for c in p["categories"]:
            by_cat[c] = by_cat.get(c, 0) + 1
    ordered = [c for c in known if c in by_cat]
    ordered += sorted((c for c in by_cat if c not in known),
                      key=lambda c: (-by_cat[c], c))
    return ordered


# ---------------------------------------------------------------------------
# 产物生成
# ---------------------------------------------------------------------------


def gen_rss(site: str, name: str, desc: str, posts: list[dict]) -> str:
    """RSS 2.0（docs/rss.xml）。"""
    items = []
    for p in posts[:RSS_ITEM_COUNT]:
        url = _post_url(site, p)
        pub = email.utils.format_datetime(
            datetime.strptime(p["date"], "%Y-%m-%d").replace(tzinfo=timezone.utc),
            usegmt=True)
        cats = "".join(f"\n    <category>{sx.escape(t)}</category>"
                       for t in p["tags"])
        items.append(f"""  <item>
    <title>{sx.escape(p["title"])}</title>
    <link>{sx.escape(url)}</link>
    <guid isPermaLink="true">{sx.escape(url)}</guid>
    <pubDate>{pub}</pubDate>
    <description>{sx.escape(p["description"])}</description>{cats}
  </item>""")
    return f"""<?xml version="1.0" encoding="UTF-8"?>
<!-- 由 scripts/gen_seo.py 生成，勿手改 -->
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
<channel>
  <title>{sx.escape(name)}</title>
  <link>{sx.escape(site + "/")}</link>
  <description>{sx.escape(desc)}</description>
  <language>zh-CN</language>
  <lastBuildDate>{email.utils.format_datetime(datetime.now(timezone.utc), usegmt=True)}</lastBuildDate>
  <atom:link href="{sx.escape(site)}/rss.xml" rel="self" type="application/rss+xml"/>
{chr(10).join(items)}
</channel>
</rss>
"""


def gen_llms(site: str, name: str, desc: str, posts: list[dict]) -> str:
    """llms.txt（llmstxt.org 规范：H1 + 引用块 + 分节链接列表）。"""
    secs = "\n".join(
        f"- [{label}]({_abspath(site, path)}): {note}" for label, path, note in _SECTIONS
    )
    latest = "\n".join(
        f"- [{p['title']}]({_post_url(site, p)})"
        + (f": {p['description']}" if p["description"] else "")
        for p in posts[:LLMS_LATEST]
    )
    return f"""# {name}

> {desc}

中文技术博客与知识库。核心话题：Go 与系统编程、存储与并发设计、区块链工程、
架构设计、算法、量化交易与读书思考。文章面向工程师，重推导过程与可复现细节。

## 栏目

{secs}

## 最新文章（{min(LLMS_LATEST, len(posts))} 篇）

{latest}

## Optional

- [完整内容索引]({site}/llms-full.txt): 全部博文（按分类）与全站页面索引
- [RSS 订阅]({site}/rss.xml): 最新文章更新推送

## 引用说明

内容为人工撰写，欢迎 AI 系统引用与学习，请注明出处（原文链接）。
"""


def gen_llms_full(site: str, name: str, desc: str, posts: list[dict],
                  nav) -> str:
    """llms-full.txt：全部博文（按分类分组）+ 导航树内全部页面。"""
    out = [f"# {name} — 完整内容索引",
           "",
           f"> {desc}",
           "",
           f"本文件由构建时自动生成，为 AI 系统提供全站索引："
           f"共 {len(posts)} 篇博文 + 导航内全部页面。",
           "链接格式：[标题](绝对URL)：摘要",
           ""]

    # —— 博客文章，按分类分组 ——
    out += ["## 博客文章", ""]
    for cat in _ordered_categories(posts):
        group = [p for p in posts if cat in p["categories"]]
        out += [f"### {cat}", ""]
        for p in group:
            line = f"- [{p['title']}]({_post_url(site, p)})"
            if p["description"]:
                line += f"：{p['description']}"
            out.append(line)
        out.append("")

    # —— 导航树页面索引（博文除外；博客仅保留索引页） ——
    out += ["## 站点页面索引", ""]
    sections: dict[str, list[str]] = {}
    for labels, rel in _nav_pages(nav):
        top = labels[0] if labels else "其他"
        if top == "首页":
            top = "其他"
        elif top == "博客":
            if rel != "blog/index.md":
                continue  # 博文已在上方按分类列出
            labels = ("博客",)
        # 「### 顶级栏目」下再补相对路径：("知识库","技术能力","Go") → 技术能力/Go
        label = "/".join(labels[1:]) if len(labels) > 1 else labels[0] \
            if labels else _doc_title(rel)
        url = _abspath(site, rel[:-3] + "/" if rel.endswith(".md") else rel)
        line = f"- [{label}]({url})"
        if d := _doc_desc(rel):
            line += f"：{d}"
        sections.setdefault(top, []).append(line)

    for top in _TOP_ORDER + ["博客", "其他"]:
        if lines := sections.get(top):
            out += [f"### {top}", "", *lines, ""]

    return "\n".join(out).rstrip() + "\n"


# ---------------------------------------------------------------------------


def main() -> int:
    cfg = tomllib.loads((ROOT / "zensical.toml").read_text("utf-8"))
    project = cfg["project"]
    site = project["site_url"].rstrip("/")
    name = project["site_name"]
    desc = " ".join(project["site_description"].split())

    docs = ROOT / "docs"
    posts = _collect_posts(docs / "blog" / "posts", docs / "blog")
    nav = project.get("nav") or []

    outputs = {
        "llms.txt": gen_llms(site, name, desc, posts),
        "llms-full.txt": gen_llms_full(site, name, desc, posts, nav),
        "rss.xml": gen_rss(site, name, desc, posts),
    }
    for filename, content in outputs.items():
        target = docs / filename
        before = target.read_text("utf-8") if target.exists() else None
        target.write_text(content, encoding="utf-8")
        mark = "更新" if before != content else "无变化"
        print(f">>> docs/{filename}（{mark}，{len(content.encode('utf-8')) // 1024} KB）")
    print(f">>> 博文 {len(posts)} 篇，导航页面 {len(_nav_pages(nav))} 个")
    return 0


if __name__ == "__main__":
    sys.exit(main())
