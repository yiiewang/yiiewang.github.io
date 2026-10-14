#!/usr/bin/env bash
# =============================================================================
# 同步本地化的第三方前端资源（KaTeX / Mermaid）→ docs/assets/vendor/
#
# 背景：unpkg / jsdelivr 在大陆访问不稳定，拖累 LCP 与爬虫渲染（Core Web
# Vitals 是排名因素），因此改为提交本地副本——同 hero.bundle.js 哲学：
# CI 构建不依赖 CDN 网络。
#
# 用法（需联网）：
#   scripts/vendor_sync.sh
#
# 升级：改下面的 KATEX_VERSION / MERMAID_VERSION 后重新执行，并确认
#   zensical.toml 的 extra_javascript / extra_css 引用路径不变。
#
# KaTeX 字体只保留 woff2（现代浏览器与 AI 渲染器全覆盖，体积约为
# woff2+woff+ttf 全量的 1/3），并同步从 CSS 剥掉其余 src 条目。
# =============================================================================
set -euo pipefail

KATEX_VERSION="0.19.0"
MERMAID_VERSION="11.3.0"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
KATEX_DIR="$ROOT/docs/assets/vendor/katex"
MERMAID_DIR="$ROOT/docs/assets/vendor/mermaid"

KATEX_BASE="https://unpkg.com/katex@${KATEX_VERSION}/dist"
MERMAID_BASE="https://cdn.jsdelivr.net/npm/mermaid@${MERMAID_VERSION}/dist"

fetch() { # fetch <url> <dest>
  echo "  -> $(basename "$2")"
  curl -fsSL --retry 3 --retry-delay 2 -o "$2" "$1"
}

echo ">>> KaTeX ${KATEX_VERSION}"
mkdir -p "$KATEX_DIR/fonts"
fetch "$KATEX_BASE/katex.min.css"              "$KATEX_DIR/katex.min.css"
fetch "$KATEX_BASE/katex.min.js"               "$KATEX_DIR/katex.min.js"
fetch "$KATEX_BASE/contrib/auto-render.min.js" "$KATEX_DIR/auto-render.min.js"

# 剥掉 woff / ttf 的 src 条目，使 CSS 引用与本地 woff2 文件一一对应
sed -E 's/,url\(fonts\/[^)]+\.(woff|ttf)\) format\("(woff|truetype)"\)//g' \
  "$KATEX_DIR/katex.min.css" > "$KATEX_DIR/katex.min.css.tmp"
mv "$KATEX_DIR/katex.min.css.tmp" "$KATEX_DIR/katex.min.css"

grep -oE 'fonts/[A-Za-z0-9._-]+\.woff2' "$KATEX_DIR/katex.min.css" | sort -u \
  | while read -r font; do
      fetch "$KATEX_BASE/$font" "$KATEX_DIR/$font"
    done

echo ">>> Mermaid ${MERMAID_VERSION}"
mkdir -p "$MERMAID_DIR"
fetch "$MERMAID_BASE/mermaid.min.js" "$MERMAID_DIR/mermaid.min.js"

echo ">>> 完成：$(du -sh "$ROOT/docs/assets/vendor" | cut -f1) → docs/assets/vendor/"
