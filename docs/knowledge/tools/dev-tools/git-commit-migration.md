---
title: Git 提交迁移实战：diff / cherry-pick / merge / checkout
created: 2026-09-03
updated: 2026-09-03
domain: tools
tags:
  - Git
  - 命令行
  - 版本控制
  - 分支管理
difficulty: beginner
summary: "先看清 commit/分支差异，再按场景选择 cherry-pick、merge、checkout 三条迁移路径，并掌握冲突处理流程。"
source:
  - type: experience
    title: "分支开发中把某个 commit 的变动迁移到当前分支"
---

# Git 提交迁移实战：diff / cherry-pick / merge / checkout

> 比较差异用 diff / log，迁移变动按场景三选一：cherry-pick 搬单点、merge 合整线、checkout 取单文件。

## 背景

多分支并行开发时，经常需要把「别的分支 / 某个历史 commit」上的改动搬到当前分支。典型场景包括：hotfix 要同步回开发分支、长周期分支要捡回某个修复、或者只想参考另一个 commit 里某个文件的实现。

迁移前先回答两个问题：

- **要搬多少？** 一个提交、一条分支的全部差异，还是只是一个文件。
- **从哪搬到哪？** 先用 diff / log 把差异范围看清楚，再决定用哪条命令。

## 核心内容

### 第一步：看清差异

假设要比较的目标 commit 记为 `<commit>`（换成分支名同理）：

```bash
# 看目标 commit 与当前 HEAD 之间的代码差异
git diff <commit>..HEAD

# 只看改了哪些文件、增删多少行
git diff --stat <commit>..HEAD

# 看当前分支比该 commit 多出了哪些提交
git log --oneline <commit>..HEAD
```

如果目标是分支名，直接 `git diff <branch>..HEAD` 即可，语法完全一致。

### 场景速查表

| 场景 | 推荐命令 | 产出 |
|------|---------|------|
| 只要某个 commit 的改动 | `git cherry-pick <hash>` | 在当前分支生成一个新 commit |
| 要整条分支相对当前分支的全部差异 | `git merge <branch>` | 合并提交（或 fast-forward） |
| 只想拿某文件在该 commit 的版本 | `git checkout <commit> -- path/to/file` | 工作区文件被替换（未提交） |

### 路径一：cherry-pick —— 嫁接单个提交

```bash
# 搬一个 commit
git cherry-pick <commit-hash>

# 搬多个连续 commit（A 之后到 B，含 B）
git cherry-pick <A>^..<B>

# 搬多个不连续 commit
git cherry-pick A B C
```

cherry-pick 只搬指定 commit 的改动，并在当前分支生成新 commit，是最常用的「单点迁移」方式。

### 路径二：merge —— 合入整条分支

```bash
# 把目标分支合入当前分支
git merge <branch>

# 强制生成合并提交，保留分支拓扑
git merge --no-ff <branch>
```

merge 适合「整条线都要合进来」的场景，会连同分支上的全部提交一起并入。

### 路径三：checkout —— 只取某个文件

```bash
# 把该 commit 中某个文件的内容取到工作区
git checkout <commit> -- path/to/file.go
```

注意：这种方式拿到的变更停留在工作区，需要自己再 commit。

### 冲突处理（cherry-pick 场景）

1. `git status` 查看冲突文件
2. 手动解决冲突
3. `git add .` 标记已解决
4. `git cherry-pick --continue` 继续执行
5. 想整体放弃则 `git cherry-pick --abort`

## 注意事项

- ⚠️ 迁移前先 `git status` 确认工作区干净，避免把未提交改动混进迁移
- ⚠️ `<A>^..<B>` 语法表示「A 的下一个提交到 B」，注意边界是否包含 A 本身
- ⚠️ `checkout <commit> -- file` 的结果未提交，容易遗漏 commit 步骤
- ✅ 冲突是长周期分支间的常态，逐文件手动解决即可
- ✅ 选型口诀：cherry-pick 搬单点，merge 合整线，checkout 取单文件

## 延伸阅读

- [Pro Git（中文版）—— 分支与合并](https://git-scm.com/book/zh/v2)
- [git-cherry-pick 官方文档](https://git-scm.com/docs/git-cherry-pick)

---

*维护人：yiiewang · 最后更新：2026-09-03*
