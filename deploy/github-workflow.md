# GitHub Actions CI/CD 部署指南

本文档介绍如何使用 GitHub Actions 自动构建并部署 MkDocs 博客到 GitHub Pages。

## 📋 目录

- [工作流程概述](#工作流程概述)
- [前置条件](#前置条件)
- [配置步骤](#配置步骤)
- [工作流配置文件](#工作流配置文件)
- [关于 cloaks/mkdocs 镜像](#关于-cloaksmkdocs-镜像)
- [触发部署](#触发部署)
- [常见问题](#常见问题)

---

## 工作流程概述

```
推送代码到 master/main 分支
        ↓
GitHub Actions 自动触发
        ↓
拉取 cloaks/mkdocs 镜像
        ↓
构建 MkDocs 静态站点
        ↓
部署到 gh-pages 分支
        ↓
GitHub Pages 自动更新
```

---

## 前置条件

1. **GitHub 仓库** - 已创建并包含 MkDocs 项目
2. **GitHub Pages 已启用** - 仓库 Settings → Pages → Source 设置为 `gh-pages` 分支
3. **MkDocs 配置** - 项目根目录包含 `mkdocs.yml` 配置文件
4. **Docker 镜像** - `cloaks/mkdocs:latest` 已推送到 Docker Hub

---

## 配置步骤

### 1. 创建工作流目录

```bash
mkdir -p .github/workflows
```

### 2. 创建工作流配置文件

```bash
cat > .github/workflows/ci.yml << 'EOF'
name: ci
on:
  push:
    branches:
      - master
      - main
permissions:
  contents: write
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Configure Git Credentials
        run: |
          git config user.name github-actions[bot]
          git config user.email 41898282+github-actions[bot]@users.noreply.github.com

      - name: Cache built site
        uses: actions/cache@v4
        with:
          key: mkdocs-material-${{ env.cache_id }}
          path: site
          restore-keys: |
            mkdocs-material-site-

      - name: Build and Deploy using MkDocs
        run: |
          docker run --rm \
            -v $(pwd):/docs \
            cloaks/mkdocs:latest \
            gh-deploy --force
EOF
```

### 3. 启用 GitHub Pages

1. 进入仓库 **Settings** → **Pages**
2. **Source** 选择 `Deploy from a branch`
3. **Branch** 选择 `gh-pages` / `/ (root)`
4. 点击 **Save**

### 4. 配置仓库权限

确保 GitHub Actions 有写入权限：

1. 进入仓库 **Settings** → **Actions** → **General**
2. 滚动到 **Workflow permissions**
3. 选择 **Read and write permissions**
4. 点击 **Save**

---

## 工作流配置文件

### 完整配置 `.github/workflows/ci.yml`

```yaml
name: ci
on:
  push:
    branches:
      - master
      - main
permissions:
  contents: write
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Configure Git Credentials
        run: |
          git config user.name github-actions[bot]
          git config user.email 41898282+github-actions[bot]@users.noreply.github.com

      - name: Cache built site
        uses: actions/cache@v4
        with:
          key: mkdocs-material-${{ env.cache_id }}
          path: site
          restore-keys: |
            mkdocs-material-site-

      - name: Build and Deploy using MkDocs
        run: |
          docker run --rm \
            -v $(pwd):/docs \
            cloaks/mkdocs:latest \
            gh-deploy --force
```

### 配置说明

| 配置项 | 说明 |
|--------|------|
| `on.push.branches` | 触发分支，推送到 master 或 main 时触发 |
| `permissions.contents` | 需要写入权限以推送到 gh-pages 分支 |
| `actions/checkout@v4` | 检出代码 |
| `actions/cache@v4` | 缓存构建产物，加速后续构建 |
| `cloaks/mkdocs:latest` | 自定义 MkDocs Docker 镜像（详见下文） |
| `gh-deploy --force` | 强制部署到 gh-pages 分支 |

---

## 关于 `cloaks/mkdocs` 镜像

### 为什么使用自定义镜像？

本项目的 `mkdocs.yml` 配置了多个插件和扩展，如果在 CI 中每次都安装依赖，会导致：

1. **构建时间长** - 每次都要 `pip install` 大量依赖，耗时 2-5 分钟
2. **依赖安装不稳定** - 网络问题可能导致安装失败
3. **版本不一致** - 不同时间构建可能安装不同版本的依赖

使用预构建的 Docker 镜像可以：

- ✅ **构建时间从 3-5 分钟缩短到 30 秒**
- ✅ **依赖版本固定，构建结果一致**
- ✅ **不受 PyPI 网络波动影响**

### 镜像包含的内容

根据项目 `mkdocs.yml` 配置，镜像预装了以下依赖：

```dockerfile
FROM squidfunk/mkdocs-material:latest

# 项目使用的 MkDocs 插件
RUN pip install --no-cache-dir \
    mkdocs-glightbox \          # 图片灯箱效果
    mkdocs-rss-plugin \         # RSS 订阅生成
    mkdocs-minify-plugin \      # HTML 压缩
    pymdown-extensions          # Markdown 扩展（代码高亮、Mermaid 等）

WORKDIR /docs
ENTRYPOINT ["mkdocs"]
```

### 镜像对应的项目插件

| 插件 | 用途 | 配置位置 |
|------|------|----------|
| `mkdocs-material` | Material 主题（基础镜像已包含） | `theme.name: material` |
| `mkdocs-glightbox` | 图片点击放大灯箱效果 | `plugins.glightbox` |
| `mkdocs-rss-plugin` | 生成 RSS 订阅源 | `plugins.rss` |
| `mkdocs-minify-plugin` | 压缩 HTML 输出 | `plugins.minify` |
| `pymdown-extensions` | 代码高亮、Mermaid 图表、数学公式等 | `markdown_extensions.pymdownx.*` |

### 如果不使用这个镜像

#### 方案一：每次安装依赖（慢，不推荐）

```yaml
- name: Setup Python
  uses: actions/setup-python@v5
  with:
    python-version: '3.x'

- name: Install dependencies
  run: |
    pip install mkdocs-material
    pip install mkdocs-glightbox
    pip install mkdocs-rss-plugin
    pip install mkdocs-minify-plugin

- name: Build and Deploy
  run: mkdocs gh-deploy --force
```

**问题**：
- 每次构建需要安装依赖，耗时 2-5 分钟
- 依赖版本可能变化，导致构建不一致
- 网络不稳定时可能安装失败

#### 方案二：使用官方镜像（功能缺失）

```yaml
- name: Build and Deploy
  run: |
    docker run --rm -v $(pwd):/docs squidfunk/mkdocs-material gh-deploy --force
```

**问题**：
- 官方镜像只包含 `mkdocs-material`，缺少其他插件
- 构建会报错：`Plugin 'glightbox' not found`

#### 方案三：使用缓存（复杂）

```yaml
- name: Cache pip
  uses: actions/cache@v4
  with:
    path: ~/.cache/pip
    key: pip-${{ hashFiles('requirements.txt') }}

- name: Install dependencies
  run: pip install -r requirements.txt
```

**问题**：
- 需要维护 `requirements.txt`
- 缓存命中率不稳定
- 配置复杂

### 如何构建这个镜像

如果需要更新镜像或添加新插件：

```bash
# 创建 Dockerfile
cat > Dockerfile << 'EOF'
FROM squidfunk/mkdocs-material:latest

RUN pip install --no-cache-dir \
    mkdocs-glightbox \
    mkdocs-rss-plugin \
    mkdocs-minify-plugin

WORKDIR /docs
ENTRYPOINT ["mkdocs"]
EOF

# 构建镜像
docker build -t cloaks/mkdocs:latest .

# 推送到 Docker Hub
docker login
docker push cloaks/mkdocs:latest
```

### 镜像版本管理建议

```bash
# 使用日期标签
docker tag cloaks/mkdocs:latest cloaks/mkdocs:2026-02-06
docker push cloaks/mkdocs:2026-02-06

# 在 CI 中使用固定版本（更稳定）
docker run --rm -v $(pwd):/docs cloaks/mkdocs:2026-02-06 gh-deploy --force
```

---

## 触发部署

### 自动触发

推送代码到 `master` 或 `main` 分支即可自动触发：

```bash
git add .
git commit -m "Update content"
git push origin main
```

### 手动触发

如需支持手动触发，修改工作流配置：

```yaml
on:
  push:
    branches:
      - master
      - main
  workflow_dispatch:  # 添加此行支持手动触发
```

然后在 GitHub 仓库 **Actions** 页面点击 **Run workflow** 即可。

### 查看部署状态

1. 进入仓库 **Actions** 页面
2. 查看最新的工作流运行状态
3. 点击进入查看详细日志

---

## 常见问题

### 1. 部署失败：权限不足

**错误信息**：
```
remote: Permission to xxx/xxx.git denied to github-actions[bot].
```

**解决方案**：
1. 进入 Settings → Actions → General
2. Workflow permissions 选择 **Read and write permissions**

### 2. 部署失败：gh-pages 分支不存在

**解决方案**：
首次部署会自动创建 `gh-pages` 分支，如果失败可手动创建：

```bash
git checkout --orphan gh-pages
git reset --hard
git commit --allow-empty -m "Initial gh-pages"
git push origin gh-pages
git checkout main
```

### 3. 页面 404

**可能原因**：
- GitHub Pages 未启用
- Source 分支设置错误

**解决方案**：
检查 Settings → Pages 配置是否正确。

### 4. Docker 镜像拉取失败

**错误信息**：
```
Unable to find image 'cloaks/mkdocs:latest' locally
```

**解决方案**：
确保 Docker Hub 上存在该镜像，或使用官方镜像：

```yaml
- name: Build and Deploy using MkDocs
  run: |
    pip install mkdocs-material
    mkdocs gh-deploy --force
```

### 5. 缓存未生效

**解决方案**：
检查 cache key 是否正确，可以添加环境变量：

```yaml
env:
  cache_id: ${{ github.run_id }}
```

---

## 相关链接

- [GitHub Actions 文档](https://docs.github.com/en/actions)
- [GitHub Pages 文档](https://docs.github.com/en/pages)
- [MkDocs 文档](https://www.mkdocs.org/)
- [MkDocs Material 主题](https://squidfunk.github.io/mkdocs-material/)
