# =============================================================================
# cloaks/zensical Docker Image
#
# Base:    python:3.12-slim
# Purpose: Build & serve yiiewang.github.io (cloaks.cn) with Zensical
#
# Zensical 自带 Markdown / pymdown-extensions / 主题与模板，
# 无需再安装 mkdocs、mkdocs-material 及各类插件。
# 博客文章列表由项目根目录 main.py 的 macros 宏渲染（plugins: macros）。
#
# Usage:
#   docker build -t cloaks/zensical:0.0.60 -f .github/workflows/Dockerfile .
#   docker run --rm -v $(pwd):/docs cloaks/zensical:0.0.60 build -f mkdocs.yml --clean
#   docker run --rm -v $(pwd):/docs -p 8000:8000 cloaks/zensical:0.0.60 \
#     serve -f mkdocs.yml -a 0.0.0.0:8000
#
# Push:
#   docker push cloaks/zensical:0.0.60
#   docker push cloaks/zensical:latest
# =============================================================================

FROM python:3.12-slim

LABEL maintainer="cloaks@qq.com"
LABEL description="Zensical static site generator for yiiewang.github.io (cloaks.cn)"

# 固定版本：Zensical 仍处于 0.0.x alpha，避免上游变更导致构建漂移
RUN pip install --no-cache-dir "zensical==0.0.60"

WORKDIR /docs
ENTRYPOINT ["zensical"]

# 默认命令：完整构建（clean 清缓存，产物在 site/）
CMD ["build", "-f", "mkdocs.yml", "--clean"]
