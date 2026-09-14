# GitHub Actions CI/CD 部署指南

本文档介绍本站如何用 GitHub Actions 自动构建并发布到 GitHub Pages。

## 📋 目录

- [工作流程概述](#工作流程概述)
- [前置条件](#前置条件)
- [工作流配置](#工作流配置)
- [关键设计](#关键设计)
- [镜像的构建与发布](#镜像的构建与发布)
- [触发与排查](#触发与排查)
- [常见问题](#常见问题)

---

## 工作流程概述

```
推送到 master / main 分支
        ↓
GitHub Actions 触发（.github/workflows/ci.yml）
        ↓
拉取预构建镜像 cloaks/zensical:0.0.60
        ↓
校验镜像内版本与 Dockerfile 固定版本一致
        ↓
容器内执行 zensical build -f zensical.toml --clean
        ↓
site/ 发布到 gh-pages 分支（peaceiris/actions-gh-pages）
        ↓
GitHub Pages 自动更新
```

---

## 前置条件

1. **GitHub Pages 已启用** —— 仓库 Settings → Pages → Source 设为 `gh-pages` 分支
2. **Actions 有写权限** —— Settings → Actions → General → Workflow permissions 选 **Read and write permissions**
3. **Docker Hub 上存在镜像** —— `cloaks/zensical:0.0.60`（CI 只拉取，不构建）

---

## 工作流配置

实际文件：`.github/workflows/ci.yml`

```yaml
name: ci
on:
  push:
    branches:
      - master
      - main
  workflow_dispatch:

permissions:
  contents: write

concurrency:
  group: pages-deploy
  cancel-in-progress: true

jobs:
  deploy:
    runs-on: ubuntu-latest
    env:
      DOCKER_IMAGE: cloaks/zensical:0.0.60
    steps:
      - uses: actions/checkout@v4

      # 镜像由本地 `make docker-build` 构建后推送到 Docker Hub，CI 只拉取不构建
      - name: Pull Zensical image
        run: docker pull "$DOCKER_IMAGE"

      # 防漂移：镜像内的 zensical 版本必须与 Dockerfile 固定版本一致
      - name: Verify image version matches Dockerfile
        run: |
          pinned=$(sed -n 's/.*"zensical==\([0-9][0-9.]*\)".*/\1/p' .github/workflows/Dockerfile | head -1)
          actual=$(docker run --rm --entrypoint zensical "$DOCKER_IMAGE" --version | tr -d '[:space:]')
          echo "Dockerfile 固定版本: $pinned / 镜像内版本: $actual"
          if [ "$pinned" != "$actual" ]; then
            echo "::error::镜像版本 ($actual) 与 Dockerfile ($pinned) 不一致。请执行 make docker-build 后推送镜像。"
            exit 1
          fi

      # --user 让容器以 runner 用户身份运行：否则 site/ 归 root 所有，
      # 下一步的 touch 会因权限不足失败
      - name: Build site
        run: |
          docker run --rm --user "$(id -u):$(id -g)" -v "$PWD":/docs "$DOCKER_IMAGE" build -f zensical.toml --clean
          touch site/.nojekyll

      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v4
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./site
          publish_branch: gh-pages
          user_name: github-actions[bot]
          user_email: 41898282+github-actions[bot]@users.noreply.github.com
          commit_message: "deploy: ${{ github.event.head_commit.message }}"
```

### 配置说明

| 配置项 | 说明 |
|--------|------|
| `on.push.branches` | 推送到 `master` 或 `main` 时触发 |
| `on.workflow_dispatch` | 支持在 Actions 页面手动触发 |
| `permissions.contents` | 需要写权限以推送到 `gh-pages` |
| `concurrency.group` | 同一时间只允许一个部署任务，新的推送会取消进行中的旧任务 |
| `actions/checkout@v4` | 检出代码 |
| `peaceiris/actions-gh-pages@v4` | 把 `site/` 发布到 `gh-pages` 分支 |

---

## 关键设计

### 为什么用预构建镜像

Zensical 自带 Markdown 扩展、主题与模板，不需要安装任何第三方插件包。把版本固定在镜像里，换来三件事：

- **构建快** —— CI 里省掉一次 `pip install`
- **不受 PyPI 网络波动影响**
- **结果可复现** —— 镜像 tag 固定，不会因为上游发版导致产物变化

### 版本防漂移

`zensical` 的版本同时出现在两处：

- `.github/workflows/Dockerfile` 的 `pip install "zensical==0.0.60"`
- `.github/workflows/ci.yml` 的 `DOCKER_IMAGE: cloaks/zensical:0.0.60`（镜像 tag）

CI 中的 `Verify image version matches Dockerfile` 步骤会实际进容器执行 `zensical --version`，与 Dockerfile 里写的版本比对。这样能拦住一种静默失败：**改了 Dockerfile 但忘了重新构建并推送镜像**——否则 CI 会用旧镜像构建出与仓库不一致的结果。

### 为什么发布用 peaceiris 而不是 `gh-deploy`

Zensical 没有 `gh-deploy` 子命令（CLI 只提供 `build` / `serve` / `new`），所以发布交给 `peaceiris/actions-gh-pages`：它只负责把 `site/` 目录推到 `gh-pages` 分支，与构建工具解耦。

### `.nojekyll`

GitHub Pages 默认用 Jekyll 处理分支内容，会忽略下划线开头的文件/目录（例如 `_static/`）。构建后 `touch site/.nojekyll` 关闭 Jekyll 处理。

---

## 镜像的构建与发布

镜像**不在 CI 里构建**，由本地构建后推到 Docker Hub。Dockerfile 位于 `.github/workflows/Dockerfile`。

```bash
# 构建（Makefile 会自动比对 Dockerfile 哈希，变更时重建）
make docker-build

# 推送
docker push cloaks/zensical:0.0.60
docker push cloaks/zensical:latest
```

升级 Zensical 版本的完整流程：

1. 修改 `.github/workflows/Dockerfile` 中的 `zensical==x.y.z`
2. 同步修改 `.github/workflows/ci.yml` 的 `DOCKER_IMAGE` tag 与 Makefile 的 `DOCKER_IMAGE`
3. 修改 `README.md` 中提到的版本号
4. `make docker-build && docker push cloaks/zensical:x.y.z`
5. 本地 `make build` 验证产物无误后推送

`make check-upstream` 可以查询 PyPI 上是否有更新的 Zensical 版本。

---

## 触发与排查

### 自动触发

```bash
git add .
git commit -m "Update content"
git push origin main
```

### 手动触发

仓库 **Actions** 页面 → 选择 `ci` → **Run workflow**。

### 查看状态

Actions 页面 → 选中某次运行 → 展开各步骤日志。构建步骤会打印 Zensical 的构建输出与耗时。

---

## 常见问题

### 1. 部署失败：权限不足

```
remote: Permission to xxx/xxx.git denied to github-actions[bot].
```

Settings → Actions → General → Workflow permissions 选 **Read and write permissions**。

### 2. 部署失败：`touch: Permission denied`

说明容器以 root 身份写入了 `site/`，runner 用户无权再修改。检查构建步骤是否带了 `--user "$(id -u):$(id -g)"`。

### 3. 构建失败：镜像版本与 Dockerfile 不一致

```
::error::镜像版本 (x.y.z) 与 Dockerfile (a.b.c) 不一致
```

改了 Dockerfile 但没重建镜像。执行 `make docker-build` 后 `docker push`，或把版本号改回一致。

### 4. 拉取镜像失败

```
Error response from daemon: manifest for cloaks/zensical:0.0.60 not found
```

确认镜像已推送到 Docker Hub，且 tag 与 `ci.yml` 中的 `DOCKER_IMAGE` 完全一致。

### 5. 页面 404

- GitHub Pages 未启用，或 Source 分支不是 `gh-pages`
- `gh-pages` 分支为空（首次部署失败）

检查 Settings → Pages 与 `gh-pages` 分支的内容。

---

## 相关链接

- [GitHub Actions 文档](https://docs.github.com/en/actions)
- [GitHub Pages 文档](https://docs.github.com/en/pages)
- [peaceiris/actions-gh-pages](https://github.com/peaceiris/actions-gh-pages)
- [Zensical 文档](https://zensical.org/docs/)
