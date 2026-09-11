# Zensical 站点管理
#
# 全部通过 Docker 执行，宿主机无需安装 Python / Zensical。
#
# - mkdocs.yml 唯一配置源（Zensical 原生读取 mkdocs.yml，无需派生配置）
# - 博客文章列表由项目根目录 main.py 的 macros 宏渲染（plugins: macros）
# - 镜像版本唯一定义在 .github/workflows/Dockerfile，CI 与本地共用同一镜像
# - 部署由 GitHub Actions 完成（.github/workflows/ci.yml）
#
# 目标：
#   make dev      开发服务器（增量构建 + 热重载）
#   make build    生产构建（clean，产物在 site/）
#   make preview  静态预览已构建的 site/（验证生产产物）
#   make clean    清理构建产物与缓存

DOCKER_IMAGE := cloaks/zensical:0.0.60
DOCKERFILE   := .github/workflows/Dockerfile
DOCS_MOUNT   := -v $(CURDIR):/docs
# 以宿主用户身份运行容器，避免 site/ .cache/ 被 root 拥有（否则 make clean 删不掉）
DOCKER_USER  := --user "$(shell id -u):$(shell id -g)"

DEV_PORT     := 8000
PREVIEW_PORT := 8001

.DEFAULT_GOAL := help

# ========== 镜像 ==========

# 镜像不存在时自动构建（存在则跳过，仅做一次 inspect）
.PHONY: ensure-image
ensure-image:
	@docker image inspect $(DOCKER_IMAGE) >/dev/null 2>&1 || { \
		echo ">>> 镜像 $(DOCKER_IMAGE) 不存在，开始构建..."; \
		docker build -t $(DOCKER_IMAGE) -f $(DOCKERFILE) .; \
	}

# 强制重建镜像（Dockerfile 改动后执行）
docker-build:
	docker build -t $(DOCKER_IMAGE) -f $(DOCKERFILE) .

# ========== 开发 / 构建 / 预览 ==========

# 开发服务器：增量构建 + 热重载
dev serve: ensure-image
	@echo ">>> 开发服务器 http://localhost:$(DEV_PORT)"
	docker run --rm $(DOCKER_USER) $(DOCS_MOUNT) -p $(DEV_PORT):8000 $(DOCKER_IMAGE) \
		serve -f mkdocs.yml -a 0.0.0.0:8000

# 生产构建：clean 清缓存，产物在 site/
build prod: ensure-image
	@echo ">>> 生产构建中..."
	docker run --rm $(DOCS_MOUNT) $(DOCKER_IMAGE) build -f mkdocs.yml --clean
	@echo ">>> 完成，产物在 site/"

# 静态预览已构建的 site/（验证生产产物，不重新构建）
preview: ensure-image
	@test -d site || { echo "!!! site/ 不存在，请先执行 make build"; exit 1; }
	@echo ">>> 预览已构建站点 http://localhost:$(PREVIEW_PORT)"
	docker run --rm $(DOCKER_USER) --entrypoint python3 \
		-v $(CURDIR)/site:/site:ro -p $(PREVIEW_PORT):8000 $(DOCKER_IMAGE) \
		-m http.server 8000 --directory /site

# ========== 通用 ==========

# 清理构建产物与缓存
clean:
	rm -rf site .cache
	find . -name '__pycache__' -type d -prune -exec rm -rf {} + 2>/dev/null || true

help:
	@echo "Zensical 站点管理（全部通过 Docker 执行）"
	@echo ""
	@echo "  make dev       开发服务器（热重载，http://localhost:$(DEV_PORT)）"
	@echo "  make build     生产构建（clean，产物在 site/）"
	@echo "  make preview   静态预览 site/（http://localhost:$(PREVIEW_PORT)）"
	@echo ""
	@echo "  make docker-build   强制重建镜像 $(DOCKER_IMAGE)"
	@echo "  make clean          清理 site/ .cache/ __pycache__/"
	@echo ""
	@echo "镜像：$(DOCKER_IMAGE)（定义于 $(DOCKERFILE)）"
	@echo "部署：推送到 master/main 后由 GitHub Actions 自动发布到 gh-pages"

.PHONY: ensure-image docker-build dev serve build prod preview clean help
