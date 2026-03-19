# Cloaks Blog

> 欢迎来到我的博客！这里是我分享个人见解、技术文章、生活点滴和创意的地方。

## 🚀 快速开始

### 方式一：Docker 启动（推荐）

使用 Docker 可以快速启动 MkDocs 开发服务器，无需本地安装 Python 环境：

```bash
# 启动容器（使用 -a 0.0.0.0:8000 确保可访问）
docker run -itd --name mkdocs \
  -v $(pwd):/docs \
  -p 8080:8000 \
  cloaks/mkdocs \
  serve -a 0.0.0.0:8000 --dirtyreload

# 查看日志
docker logs -f mkdocs

# 停止容器
docker stop mkdocs
docker rm mkdocs
```

**访问地址**：http://localhost:8080

**参数说明**：
- `-itd`: 交互式、分配终端、后台运行
- `-v $(pwd):/docs`: 挂载当前目录到容器的 /docs
- `-p 8080:8000`: 端口映射（宿主机 8080 → 容器 8000）
- `--dirtyreload`: 增量构建，只重新加载修改的文件，大幅提升热重载速度

### 方式二：本地启动

如果你已安装 Python 环境：

```bash
# 安装依赖
pip install -r requirements.txt

# 启动开发服务器
mkdocs serve

# 访问 http://127.0.0.1:8000
```

## 📝 项目结构

```
.
├── docs/              # 文档目录
│   ├── blog/          # 博客文章
│   ├── assets/        # 静态资源
│   └── index.md       # 首页
├── overrides/         # 主题覆盖
├── mkdocs.yml         # MkDocs 配置文件
└── README.md          # 本文件
```

## 🔧 配置说明

- **主题**：Material for MkDocs
- **语言**：中文
- **功能特性**：
  - 博客系统
  - 代码高亮
  - 数学公式（KaTeX）
  - 图表（Mermaid）
  - 图片灯箱
  - 标签系统
  - RSS 订阅
  - 社交分享

## 📖 写作规范

- 博客文章存放在 `docs/blog/posts/{year}/{month}/` 目录
- 文件名使用日期格式：`{day}.md`
- 同一天多篇文章使用 `{day}-{序号}.md`
- 参考 [blog-writing 规则](.codebuddy/rules/blog-writing.mdc)

## 🐳 Docker 优化建议

### 提升编译速度

1. **使用 --dirtyreload 参数**
   ```bash
   mkdocs serve --dirtyreload
   ```
   只重新加载修改的文件，避免全量构建

2. **添加缓存卷**
   ```bash
   docker run -itd --name mkdocs \
     -v $(pwd):/docs \
     -v mkdocs-cache:/root/.cache \
     -p 8080:8000 \
     cloaks/mkdocs \
     serve --dirtyreload
   ```

3. **限制资源使用**
   ```bash
   docker run -itd --name mkdocs \
     --cpus="2.0" \
     --memory="2g" \
     -v $(pwd):/docs \
     -p 8080:8000 \
     cloaks/mkdocs \
     serve --dirtyreload
   ```

### 生产构建

```bash
# 完整构建
docker run --rm -v $(pwd):/docs cloaks/mkdocs mkdocs build

# 构建结果在 site/ 目录
```

## 📄 许可证

Copyright &copy; 2018 - 2026 Cloaks

## 🔗 相关链接

- [MkDocs 文档](https://www.mkdocs.org/)
- [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/)
- [博客](https://cloaks.cn/blog/)
