---
date: 2026-09-17
authors:
  - codebuddy
tags:
  - zsh
  - oh-my-zsh
  - Shell 脚本
  - 开发环境
comments: true
description: 一行 curl 装好 zsh + oh-my-zsh + 自动建议/语法高亮/fzf：脚本怎么做到非交互、幂等、可降级，附三个使用场景与 ZSH 变量劫持踩坑实录。
---

# setup-zsh.sh：一行 curl 装好 zsh

新容器、新 VPS、重装的系统——每次都是同一套动作：`apt install zsh`、装 oh-my-zsh、clone 两个插件、改 `.zshrc`、`chsh`。手敲五遍之后，它变成了仓库里一个不到 200 行的脚本和一条命令：

```bash
curl -fsSL https://raw.githubusercontent.com/yiiewang/yiiewang.github.io/master/example/script/setup-zsh.sh | bash
```

完整源码在 [example/script/setup-zsh.sh](https://github.com/yiiewang/yiiewang.github.io/blob/master/example/script/setup-zsh.sh)。这个页面是它的文档：装什么、怎么用、为什么这么设计，以及一个真实的踩坑记录。

## 安装内容

| 组件 | 作用 | 方式 |
|---|---|---|
| zsh / git / fzf | 基础件 | 包管理器，fzf 失败不阻断 |
| oh-my-zsh | 框架 | 官方 install.sh `--unattended` |
| zsh-autosuggestions | 灰色历史建议 | clone 到 custom/plugins |
| zsh-syntax-highlighting | 命令语法高亮 | clone 到 custom/plugins |
| `.zshrc` 托管块 | 插件加载配置 | 追加，marker 识别 |
| chsh | 默认 shell 切换 | best-effort，失败只告警 |

## 快速开始

=== "全新机器 / 容器"

    ```bash
    curl -fsSL https://raw.githubusercontent.com/yiiewang/yiiewang.github.io/master/example/script/setup-zsh.sh | bash
    ```

    实测输出节选（容器环境，zsh 预装、fzf 源不可用正好演示降级）：

    ```text
    [=] zsh 已安装：zsh 5.5.1
    [!] fzf 安装失败（不影响其余组件）
    [+] 安装 oh-my-zsh …
    [+] zsh-autosuggestions 安装完成
    [+] zsh-syntax-highlighting 安装完成
    [+] .zshrc 已追加插件配置（managed block）
    [+] 默认 shell 已切换为 zsh（重新登录生效）
    ```

=== "已装过 oh-my-zsh"

    还是同一条命令。脚本认得现场：OMZ 跳过、缺的插件补装、托管块只在缺失时追加——**一条命令同时是安装器和体检器**，这就是幂等的红利：

    ```text
    [=] zsh 已安装：zsh 5.5.1
    [=] oh-my-zsh 已安装
    [=] zsh-autosuggestions 已安装
    [=] .zshrc 托管块已存在
    ```

=== "CI / 验证模式"

    不想动真机就隔离 HOME 跑，`ZSH_SETUP_NO_CHSH=1` 跳过 shell 切换：

    ```bash
    mkdir -p /tmp/zsh-test
    HOME=/tmp/zsh-test ZSH_SETUP_NO_CHSH=1 bash setup-zsh.sh
    HOME=/tmp/zsh-test zsh -ic \
      'print $+functions[_zsh_autosuggest_start] ${ZSH_AUTOSUGGEST_HIGHLIGHT_STYLE:-missing}'
    # 期望输出：1 fg=8
    ```

    最后一行是安装是否生效的判据：`1` 表示插件入口函数已注册，`fg=8` 是默认建议色。比"打开终端看看"可靠得多，适合塞进 CI 断言。

## 设计要点

### 敢被管道执行的三个前提

`curl | bash` 的脚本没有人工盯着，设计约束和交互式脚本完全不同：

1. **非交互**。oh-my-zsh 官方安装器默认要问问题、顺手 `chsh` 还 `exec zsh`——管道里全是死路，`--unattended` 一个参数关干净。`chsh` 在容器里经常没密码可用，失败就告警并打印手动命令，绝不卡死
2. **幂等**。每个组件先查存在再动手，`.zshrc` 用 marker 识别已追加的块。机器不是一次性的，脚本的输出要能当体检报告读——第二次跑应该全是 `[=] 已安装`
3. **降级不中断**。fzf 装不上、git 缺失，只影响对应功能：`pkg_install fzf || warn` 之后继续走。只有核心组件（zsh、oh-my-zsh）失败才中止

后两条合起来是所有环境配置脚本的通用纪律：**把机器当状态机，把脚本当收敛函数**——无论当前装到什么程度，跑一遍都收敛到完整状态。

### 包管理器探测

七个发行版一个函数：

```bash
pkg_install() {
    if command -v apt-get >/dev/null 2>&1; then
        $PKG_SUDO apt-get update -qq && $PKG_SUDO apt-get install -y -qq "$@"
    elif command -v dnf >/dev/null 2>&1; then
        $PKG_SUDO dnf install -y -q "$@"
    # …yum / apk / pacman / zypper / brew 依次探测
    fi
}
```

`PKG_SUDO` 在非 root 时自动加 `sudo`，macOS 走 brew 分支。

### .zshrc 托管块

不改用户已有配置，只追加一个自识别的块：

```zsh
# >>> zsh-setup managed >>>
[ -f "$HOME/.oh-my-zsh/custom/plugins/zsh-autosuggestions/zsh-autosuggestions.zsh" ] && \
    source "$HOME/.oh-my-zsh/custom/plugins/zsh-autosuggestions/zsh-autosuggestions.zsh"
# 语法高亮：必须最后加载
[ -f "$HOME/.oh-my-zsh/custom/plugins/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh" ] && \
    source "$HOME/.oh-my-zsh/custom/plugins/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh"
# <<< zsh-setup managed <<<
```

两个细节：

- **直接 source，不进 `plugins=()`**。`plugins=()` 依赖 oh-my-zsh 模板没被改坏，且加载顺序不可控；托管块自己掌握顺序——syntax-highlighting **必须最后加载** 这条硬规矩才守得住
- **`[ -f ] &&` 守卫**。插件 clone 失败的机器上 `.zshrc` 依然能正常启动，坏一行不坏整个 shell

## 踩坑记录：被劫持的 ZSH 变量

沙箱测试时撞上的真 bug。HOME 明明钉在 `/tmp/zsh-sandbox`，oh-my-zsh 却拒绝安装：

```text
The $ZSH folder already exists (/root/.oh-my-zsh).
```

沙箱里根本没有这个目录，它从哪来的？答案是环境变量 `ZSH`。oh-my-zsh 的模板 `.zshrc` 末尾有一行 `export ZSH="$HOME/.oh-my-zsh"`——也就是说 **任何装过 OMZ 的机器，shell 里都飘着这个变量**；而官方 install.sh 会尊重已存在的 `ZSH` 作为安装目标。沙箱改了 HOME 没改 ZSH，安装器就顺着旧值摸去了真机目录。

修复一行，调用安装器时显式钉死目标：

```bash
ZSH="$HOME/.oh-my-zsh" sh -c "$omz_script" "" --unattended
```

教训比修复值钱：**写给别人管道执行的脚本，不能假设环境干净**。所有会影响第三方工具行为的环境变量——这里是 `ZSH`，换个场景可能是 `PREFIX`、`DESTDIR`、`SHELL`——要么显式设置，要么显式清空。用户 shell 里飘着什么，你控制不了，但可以不让它进你的脚本。

## 扩展：加自己的插件

!!! tip

    两步：`clone_or_skip` 一行，托管块按加载顺序加一行 source。以 `zsh-history-substring-search` 为例：

    ```bash
    # 脚本顶部加变量
    PLUGIN_HSS="https://github.com/zsh-users/zsh-history-substring-search"
    # 第 4 节之后加一行
    clone_or_skip "$PLUGIN_HSS" \
        "$ZSH_CUSTOM_DIR/plugins/zsh-history-substring-search" "zsh-history-substring-search"
    ```

    托管块里紧挨语法高亮之后追加：

    ```zsh
    [ -f "$HOME/.oh-my-zsh/custom/plugins/zsh-history-substring-search/zsh-history-substring-search.zsh" ] && \
        source "$HOME/.oh-my-zsh/custom/plugins/zsh-history-substring-search/zsh-history-substring-search.zsh"
    ```

    老规矩：高亮类之后加载的插件，永远放在托管块末尾。

## 小结

这脚本没有一行聪明代码，全是笨功夫：查存在再动手、失败就降级、变量钉死、marker 识别。但"敢 `curl | bash`"的信任恰恰来自笨功夫——非交互是给管道的礼貌，幂等是给重复执行的礼貌，钉死环境变量是给陌生机器的礼貌。

国内直连 `raw.githubusercontent.com` 不稳的话，先 `wget` 下来再跑，或者套一层代理前缀——脚本本体不挑下载方式。

---

*最后更新：2026-09-17*
