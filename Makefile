# MkDocs 站点管理
# - mkdocs.yml     生产完整配置（所有插件启用）
# - mkdocs.dev.yml 开发配置（禁 rss/social/privacy/optimize/tags/minify，~30s，保留博客+搜索）
# dev 配置由 scripts/gen_dev_config.py 从 mkdocs.yml 派生，mkdocs.yml 是唯一真源
#
# 两种执行环境：
# - 宿主机直接跑（make serve / make build）：快，依赖宿主已装 mkdocs + 插件
# - Docker 容器跑（make docker-*）：环境隔离可复现，依赖镜像 cloaks/mkdocs:9.7.6

# Docker 镜像配置
DOCKER_IMAGE := cloaks/mkdocs:9.7.6
DOCKERFILE   := .github/workflows/Dockerfile
DOCKER_RUN   := docker run --rm -v $(CURDIR):/docs -p 8000:8000 $(DOCKER_IMAGE)

.DEFAULT_GOAL := help

# ========== 宿主机直接执行 ==========

# 开发：保留博客+搜索，禁其余重插件（~30s）
serve: mkdocs.dev.yml
	mkdocs serve -f mkdocs.dev.yml --dev-addr=0.0.0.0:8000 --dirty

dev: serve

# 生产构建：完整配置 + clean
build:
	mkdocs build -f mkdocs.yml --clean

prod: build

# 派生开发配置（mkdocs.yml 改动后自动重新生成）
mkdocs.dev.yml: mkdocs.yml scripts/gen_dev_config.py
	@python3 scripts/gen_dev_config.py

# 仅生成开发配置（不启动）
dev-config: mkdocs.dev.yml

# ========== Docker 容器执行 ==========

# 构建镜像（首次或 Dockerfile 改动后执行）
docker-build:
	docker build -t $(DOCKER_IMAGE) -f $(DOCKERFILE) .

# 容器开发：dev 配置 + dirty + 热重载（~30s）
docker-serve: mkdocs.dev.yml
	$(DOCKER_RUN) serve -f mkdocs.dev.yml --dirty --dev-addr=0.0.0.0:8000

# 容器生产构建：完整配置 + clean
docker-prod:
	$(DOCKER_RUN) build -f mkdocs.yml --clean

# 容器部署：gh-deploy 推送到 GitHub Pages（需容器内配置 git 凭证）
docker-deploy:
	$(DOCKER_RUN) gh-deploy --force

# ========== 通用 ==========

# 清理构建产物与派生配置
clean:
	rm -rf site .cache mkdocs.dev.yml

help:
	@echo "宿主机执行（需本地装 mkdocs）："
	@echo "  make serve            开发启动（dirty，~30s，博客+搜索可用）"
	@echo "  make build            生产构建（完整配置，clean）"
	@echo "  make dev-config       仅生成 mkdocs.dev.yml"
	@echo ""
	@echo "Docker 执行（需先 make docker-build 构建镜像）："
	@echo "  make docker-build     构建 $(DOCKER_IMAGE) 镜像"
	@echo "  make docker-serve     容器开发（dirty，~30s，热重载）"
	@echo "  make docker-prod      容器生产构建"
	@echo "  make docker-deploy    容器部署到 GitHub Pages（需 git 凭证）"
	@echo ""
	@echo "  make clean            清理 site/ .cache/ mkdocs.dev.yml"

.PHONY: serve dev build prod dev-config docker-build docker-serve docker-prod docker-deploy clean help
