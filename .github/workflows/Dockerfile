# =============================================================================
# cloaks/mkdocs Docker Image
# 
# Base:     squidfunk/mkdocs-material 9.7.6
# Purpose:  Build & deploy yiiewang.github.io (cloaks.cn)
# Usage:
#   docker build -t cloaks/mkdocs:9.7.6 .
#   docker run --rm -v $(pwd):/docs cloaks/mkdocs:9.7.6 build
#   docker run --rm -v $(pwd):/docs cloaks/mkdocs:9.7.6 gh-deploy --force
#   docker run --rm -v $(pwd):/docs -p 8000:8000 cloaks/mkdocs:9.7.6 serve --dev-addr=0.0.0.0:8000
#
# Push:
#   docker push cloaks/mkdocs:9.7.6
#   docker push cloaks/mkdocs:latest
# =============================================================================

FROM squidfunk/mkdocs-material:9.7.6

LABEL maintainer="cloaks@qq.com"
LABEL description="MkDocs Material with plugins for yiiewang.github.io"

# ---------------------------------------------------------------------------
# System dependencies for social cards & optimize plugins
#   Base image is Alpine → use apk
# ---------------------------------------------------------------------------
RUN apk add --no-cache \
    cairo \
    cairo-dev \
    libjpeg-turbo-dev \
    libpng-dev \
    libwebp-dev

# ---------------------------------------------------------------------------
# Pip plugins and extensions
#   glightbox  -> 图片灯箱放大效果 (plugins.glightbox)
#   rss        -> RSS 订阅源生成 (plugins.rss)
#   minify     -> HTML/CSS/JS 压缩 (plugins.minify)
#   privacy    -> 外部资源本地化 (plugins.privacy)       -- needs aiohttp, jsbeautifier
#   optimize   -> 图片压缩 + srcset (plugins.optimize)   -- needs pillow, cairosvg (already installed)
#   pymdownx   -> 代码高亮 / Mermaid / Emoji 等 Markdown 扩展 (markdown_extensions)
# ---------------------------------------------------------------------------
RUN pip install --no-cache-dir \
    "mkdocs-material[imaging]>=9.7" \
    "mkdocs-glightbox>=0.4" \
    "mkdocs-rss-plugin>=1.0" \
    "mkdocs-minify-plugin>=0.8" \
    "pymdown-extensions>=10" \
    "aiohttp>=3" \
    "jsbeautifier>=1"

WORKDIR /docs
ENTRYPOINT ["mkdocs"]

# ---------------------------------------------------------------------------
# Default command: development server with hot reload
# --dev-addr=0.0.0.0:8000 绑定容器对外地址，配合 -p 8000:8000 宿主可访问
# ---------------------------------------------------------------------------
CMD ["serve", "--dev-addr=0.0.0.0:8000"]