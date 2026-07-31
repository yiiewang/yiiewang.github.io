---
description: 个人开发工作台 — 文件浏览器、Todo 看板、静态文件服务
hide:
  - navigation
  - toc
---

# Workbench

个人开发工作台，集文件浏览、任务管理、知识服务于一体。

<div class="grid cards" markdown>

- :material-folder: **文件浏览器**

    ---

    目录树导航、Tab 编辑、Markdown / JSON / 代码预览，支持跨文件链接跳转和语法高亮。

- :material-checkbox-marked-circle: **Todo 看板**

    ---

    多用户多组织任务管理，支持冲突检测、版本同步、多设备协作。

- :material-shield-check: **安全防护**

    ---

    路径穿越防护、符号链接校验、CORS 沙箱、Token 鉴权，默认严格模式。

- :material-download: **下载与分享**

    ---

    一键下载文件，Share 链接支持 UI 视图和原始内容两种模式。

</div>

---

## :material-arrow-right: 进入工作台

<div style="text-align: center; margin: 2rem 0;">
  <a href="http://9.134.196.89/" target="_blank" rel="noopener" class="md-button md-button--primary">
    打开 Workbench
  </a>
</div>

## :material-github: 源码与发布

- **GitHub 仓库**：[yiiewang/workbench](https://github.com/yiiewang/workbench)
- **最新版本**：[v1.0.0 Release](https://github.com/yiiewang/workbench/releases/tag/v1.0.0)

## :material-package-down: 本地部署

```bash
# 下载分发包
wget https://github.com/yiiewang/workbench/releases/download/v1.0.0/workbench-v1.0.0-linux-x86_64.tar.gz

# 解压运行
tar xzf workbench-v1.0.0-linux-x86_64.tar.gz
cd workbench-v1.0.0-linux-x86_64
./workbench
```

支持 Linux / macOS / Windows 全平台，详见 [Release 页面](https://github.com/yiiewang/workbench/releases/tag/v1.0.0)。
