# =============================================================================
# cloaks/zensical Docker Image
#
# Base:    python:3.12-slim
# Purpose: Build & serve yiiewang.github.io (cloaks.cn) with Zensical
#
# Zensical 自带 Markdown 扩展 / 主题与模板，无需再安装任何第三方插件包。
# 博客文章列表由 scripts/blog_macros.py 的 macros 宏渲染（plugins.macros）。
#
# Usage:
#   docker build -t cloaks/zensical:0.0.63 -f .github/workflows/Dockerfile .
#   docker run --rm -v $(pwd):/docs cloaks/zensical:0.0.63 build -f zensical.toml --clean
#   docker run --rm -v $(pwd):/docs -p 8000:8000 cloaks/zensical:0.0.63 \
#     serve -f zensical.toml -a 0.0.0.0:8000
#
# Push:
#   docker push cloaks/zensical:0.0.63
#   docker push cloaks/zensical:latest
# =============================================================================

FROM python:3.12-slim

LABEL maintainer="cloaks@qq.com"
LABEL description="Zensical static site generator for yiiewang.github.io (cloaks.cn)"

# 固定版本：Zensical 仍处于 0.0.x alpha，避免上游变更导致构建漂移
RUN pip install --no-cache-dir "zensical==0.0.63"

WORKDIR /docs
ENTRYPOINT ["zensical"]

# 默认命令：完整构建（clean 清缓存，产物在 site/）
CMD ["build", "-f", "zensical.toml", "--clean"]
