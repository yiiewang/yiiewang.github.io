#!/usr/bin/env bash
# setup-zsh.sh —— 一键安装 zsh + oh-my-zsh + 常用插件
#
# 安装内容：
#   zsh、git、fzf            包管理器安装（fzf 失败不影响其余组件）
#   oh-my-zsh               --unattended 模式；已有 .zshrc 会被备份为 .zshrc.pre-oh-my-zsh
#   zsh-autosuggestions     历史灰色建议
#   zsh-syntax-highlighting 语法高亮（.zshrc 托管块里放在最后加载）
#   chsh                    默认 shell 切为 zsh（失败仅告警，容器环境常见）
#
# 用法（推送到 master 后）：
#   curl -fsSL https://raw.githubusercontent.com/yiiewang/yiiewang.github.io/master/example/script/setup-zsh.sh | bash
#   或下载后执行：bash setup-zsh.sh
#
# 特性：
#   幂等——重复执行安全，已装组件自动跳过
#   非交互——适合 curl | bash 与容器环境
#   环境变量 ZSH_SETUP_NO_CHSH=1 可跳过默认 shell 切换（测试 / CI 用）
set -euo pipefail

OMZ_INSTALL_URL="https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh"
PLUGIN_AUTOSUGGEST="https://github.com/zsh-users/zsh-autosuggestions"
PLUGIN_HIGHLIGHT="https://github.com/zsh-users/zsh-syntax-highlighting"
ZSH_CUSTOM_DIR="$HOME/.oh-my-zsh/custom"

info() { printf '\033[1;32m[+]\033[0m %s\n' "$*"; }
skip() { printf '\033[1;33m[=]\033[0m %s\n' "$*"; }
warn() { printf '\033[1;31m[!]\033[0m %s\n' "$*"; }

# ---------- 包管理器探测 ----------
PKG_SUDO=""
[ "$(id -u)" -ne 0 ] && command -v sudo >/dev/null 2>&1 && PKG_SUDO="sudo"

pkg_install() {
    if command -v apt-get >/dev/null 2>&1; then
        $PKG_SUDO apt-get update -qq && $PKG_SUDO apt-get install -y -qq "$@"
    elif command -v dnf >/dev/null 2>&1; then
        $PKG_SUDO dnf install -y -q "$@"
    elif command -v yum >/dev/null 2>&1; then
        $PKG_SUDO yum install -y -q "$@"
    elif command -v apk >/dev/null 2>&1; then
        $PKG_SUDO apk add -q "$@"
    elif command -v pacman >/dev/null 2>&1; then
        $PKG_SUDO pacman -Sy --noconfirm --needed "$@"
    elif command -v zypper >/dev/null 2>&1; then
        $PKG_SUDO zypper --non-interactive install -y "$@"
    elif command -v brew >/dev/null 2>&1; then
        brew install "$@"
    else
        warn "未识别的包管理器，跳过：$*"
        return 1
    fi
}

# ---------- 1. zsh ----------
if command -v zsh >/dev/null 2>&1; then
    skip "zsh 已安装：$(zsh --version | head -1)"
else
    info "安装 zsh …"
    pkg_install zsh || { warn "zsh 安装失败，中止"; exit 1; }
    info "zsh 安装完成：$(zsh --version | head -1)"
fi

# ---------- 2. git / fzf ----------
command -v git >/dev/null 2>&1 || pkg_install git || warn "git 不可用，插件克隆将跳过"
if command -v fzf >/dev/null 2>&1; then
    skip "fzf 已安装"
else
    pkg_install fzf || warn "fzf 安装失败（不影响其余组件）"
fi

# ---------- 3. oh-my-zsh ----------
if [ -d "$HOME/.oh-my-zsh" ]; then
    skip "oh-my-zsh 已安装"
else
    info "安装 oh-my-zsh …"
    omz_script="$(curl -fsSL "$OMZ_INSTALL_URL")" || { warn "下载 install.sh 失败（网络？），中止"; exit 1; }
    # 显式钉死安装目标：用户可能带着已导出的 ZSH 变量（OMZ 的 .zshrc 会 export），
    # 不钉死的话安装器会用旧值而不是 $HOME/.oh-my-zsh
    ZSH="$HOME/.oh-my-zsh" sh -c "$omz_script" "" --unattended || { warn "oh-my-zsh 安装失败，中止"; exit 1; }
    info "oh-my-zsh 安装完成"
fi

# ---------- 4. 插件 ----------
clone_or_skip() {
    local url="$1" dir="$2" name="$3"
    if [ -d "$dir" ]; then
        skip "$name 已安装"
        return 0
    fi
    command -v git >/dev/null 2>&1 || { warn "git 不可用，跳过 $name"; return 0; }
    info "安装 $name …"
    git clone --depth 1 -q "$url" "$dir" && info "$name 安装完成" || warn "$name 克隆失败"
}

clone_or_skip "$PLUGIN_AUTOSUGGEST" "$ZSH_CUSTOM_DIR/plugins/zsh-autosuggestions" "zsh-autosuggestions"
clone_or_skip "$PLUGIN_HIGHLIGHT" "$ZSH_CUSTOM_DIR/plugins/zsh-syntax-highlighting" "zsh-syntax-highlighting"

# ---------- 5. .zshrc 托管块 ----------
ZSHRC="$HOME/.zshrc"
touch "$ZSHRC"
if grep -q "zsh-setup managed" "$ZSHRC"; then
    skip ".zshrc 托管块已存在"
else
    cat >> "$ZSHRC" <<'EOF'

# >>> zsh-setup managed >>>
# fzf 快捷键与补全（哪个存在用哪个）
[ -f "$HOME/.fzf.zsh" ] && source "$HOME/.fzf.zsh"
for _f in \
    /usr/share/doc/fzf/examples/key-bindings.zsh \
    /usr/share/doc/fzf/examples/completion.zsh \
    /usr/share/fzf/key-bindings.zsh \
    /usr/share/fzf/completion.zsh \
    /opt/homebrew/opt/fzf/shell/key-bindings.zsh \
    /opt/homebrew/opt/fzf/shell/completion.zsh \
    /usr/local/opt/fzf/shell/key-bindings.zsh \
    /usr/local/opt/fzf/shell/completion.zsh; do
    [ -f "$_f" ] && source "$_f"
done
unset _f
# 历史灰色建议
[ -f "$HOME/.oh-my-zsh/custom/plugins/zsh-autosuggestions/zsh-autosuggestions.zsh" ] && \
    source "$HOME/.oh-my-zsh/custom/plugins/zsh-autosuggestions/zsh-autosuggestions.zsh"
# 语法高亮：必须最后加载
[ -f "$HOME/.oh-my-zsh/custom/plugins/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh" ] && \
    source "$HOME/.oh-my-zsh/custom/plugins/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh"
# <<< zsh-setup managed <<<
EOF
    info ".zshrc 已追加插件配置（managed block）"
fi

# ---------- 6. 默认 shell ----------
if [ "${ZSH_SETUP_NO_CHSH:-0}" = "1" ]; then
    skip "跳过默认 shell 切换（ZSH_SETUP_NO_CHSH=1）"
elif [ "${SHELL:-}" = "$(command -v zsh)" ]; then
    skip "默认 shell 已是 zsh"
else
    ZSH_BIN="$(command -v zsh)"
    USER_NAME="${USER:-$(id -un)}"
    if chsh -s "$ZSH_BIN" "$USER_NAME" 2>/dev/null; then
        info "默认 shell 已切换为 zsh（重新登录生效）"
    else
        warn "chsh 失败（容器/无密码环境常见），手动执行：$PKG_SUDO chsh -s $ZSH_BIN $USER_NAME"
    fi
fi

echo
info "全部完成。执行 exec zsh 立即体验，或重新登录。"
