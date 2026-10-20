#!/usr/bin/env python3
"""生成默认分享卡 docs/assets/images/og-card.png（1200×630）。

v4 纯图标版：站点 favicon（docs/assets/favicon.ico，260px RGBA）放大居中，
身后一圈淡 indigo 光晕，无任何文案。

og:image / twitter:image / BlogPosting JSON-LD 的 image 都指向它。
分享到微信/Twitter/Slack 及 Google Discover 抓取时展示。

依赖（仅本机运行，不在 Zensical 构建镜像内）：
  pip install pillow

用法：
  python3 scripts/og_card.py
"""

from __future__ import annotations

from pathlib import Path

from PIL import Image, ImageDraw, ImageFilter

W, H = 1200, 630
ROOT = Path(__file__).resolve().parent.parent
OUT = ROOT / "docs/assets/images/og-card.png"
MARK = ROOT / "docs/assets/favicon.ico"

# 站点设计令牌（indigo primary / slate dark scheme，见 zensical.toml 与 site.css）
BG_TOP = (10, 14, 26)        # slate-950 压暗一档
BG_BOTTOM = (18, 25, 48)     # 深蓝渐变
INDIGO = (129, 140, 248)     # indigo-400（光晕内层）
INDIGO_DEEP = (79, 70, 229)  # indigo-600（光晕主体）

MARK_SIZE = 420  # 260px 源放大到 420，作为画面主体


def main() -> None:
    # 1. 垂直渐变底
    img = Image.new("RGB", (W, H), BG_TOP)
    grad = Image.new("RGB", (1, H))
    for y in range(H):
        t = y / H
        grad.putpixel((0, y), tuple(
            round(a + (b - a) * t) for a, b in zip(BG_TOP, BG_BOTTOM)))
    img.paste(grad.resize((W, H)))

    # 2. logo 身后一圈淡 indigo 光晕（透明层合成，不压暗四角）
    glow = Image.new("RGBA", (W, H), (0, 0, 0, 0))
    gd = ImageDraw.Draw(glow)
    gd.ellipse((230, -55, 970, 685), fill=INDIGO_DEEP + (100,))
    gd.ellipse((380, 95, 820, 535), fill=INDIGO + (40,))
    glow = glow.filter(ImageFilter.GaussianBlur(130))
    img = Image.alpha_composite(img.convert("RGBA"), glow).convert("RGB")

    # 3. favicon 头像式居中
    mark = Image.open(MARK).convert("RGBA")
    mark = mark.resize((MARK_SIZE, MARK_SIZE), Image.Resampling.LANCZOS)
    mx, my = (W - MARK_SIZE) // 2, (H - MARK_SIZE) // 2
    img.paste(mark, (mx, my), mark)

    OUT.parent.mkdir(parents=True, exist_ok=True)
    img.save(OUT, "PNG", optimize=True)
    print(f">>> {OUT.relative_to(ROOT)}（{OUT.stat().st_size // 1024} KB, {W}×{H}）")
    print(f"    favicon {mark.size[0]}px @ ({mx},{my})，无文案")


if __name__ == "__main__":
    main()
