"""
从 mkdocs.yml 派生开发配置，禁用指定插件块以加速 mkdocs serve。

禁用: rss/social/privacy/optimize/tags/minify（保留 blog+search，~30s）
原理: 纯文本处理（不解析 YAML），按缩进识别 plugins 段内的插件块并整块注释，
保留 !!python/object/apply 等 tag 和所有注释。脚本依赖在项目根目录执行（Makefile 保证）。
"""
import re

TARGETS = {"rss", "social", "privacy", "optimize", "tags", "minify"}

with open("mkdocs.yml") as f:
    lines = f.readlines()

out = []
in_plugins = False
skip = False
for line in lines:
    # 进入 plugins 段
    if re.match(r"^plugins:\s*$", line):
        in_plugins = True
        skip = False
        out.append(line)
        continue
    # 离开 plugins 段（遇到顶级非空行）
    if in_plugins and re.match(r"^\S", line):
        in_plugins = False
        skip = False
        out.append(line)
        continue
    if in_plugins:
        # 识别插件头: "  - xxx" 或 "  - xxx:"（不要求冒号，兼容 - glightbox）
        m = re.match(r"^  - (\w+)", line)
        if m:
            if m.group(1) in TARGETS:
                skip = True
                out.append("  # " + line[2:])
                continue
            else:
                skip = False
                out.append(line)
                continue
        # skip 期间的子项整行注释
        if skip:
            out.append("  # " + line[2:])
            continue
    out.append(line)

with open("mkdocs.dev.yml", "w") as f:
    f.writelines(out)

print(f">> generated mkdocs.dev.yml (disabled: {sorted(TARGETS)})")
