# Cloaks 个人站 Skills 库

> 给 AI Agent 用的**写作与知识管理工具集**。每个 skill 自包含：触发后只需读 `SKILL.md` 即可上手，细节在同目录的辅助文件里。

## Skills 索引

| Skill | 触发场景 | 主要文件 |
|-------|---------|---------|
| [blog-writer](./blog-writer/SKILL.md) | "写一篇关于 X 的博客" / "把这段经验写成文章" | `SKILL.md` + 写作风格指南 + 3 个范例 |
| [knowledge-curator](./knowledge-curator/SKILL.md) | "把这段内容沉淀到知识库" / "在知识库里搜 X" | `SKILL.md` + 7 域映射表 + 分类规则 + 检索指南 |

## 使用方法

### 对 AI Agent

1. 用户说"写博客" → 加载 `blog-writer/SKILL.md`
2. 用户说"沉淀到知识库"或"知识库搜 X" → 加载 `knowledge-curator/SKILL.md`
3. 严格按照 skill 里的工作流执行；如遇歧义先询问用户

### 对人类

直接读 `SKILL.md` 了解每个 skill 的设计意图和边界。
辅助文件是 AI 工作时的参考手册，不需要通读。

## 设计原则

- **自包含**：每个 skill 是一个目录，AI 只需读 `SKILL.md` 即可启动
- **真实导向**：范例和规则都从仓库**真实的博客和知识库**提炼，而非凭空写
- **可演进**：随博客和知识库增长，skill 内的规则和范例可以持续更新
- **不绑定 CI**：`.skills/` 不参与站点构建（`mkdocs.yml` 的 `docs_dir` 只指向 `docs/`）

## 维护建议

- 每写 5-10 篇博客，更新一次 `blog-writer/writing-style-guide.md`
- 每沉淀 10 条知识，更新一次 `knowledge-curator/domain-map.md`
- examples 优先选**你认为写得最好的**博客，不要追求全
