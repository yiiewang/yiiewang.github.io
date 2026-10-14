#!/usr/bin/env python3
"""生成默认分享卡 docs/assets/images/og-card.png（1200×630）。

og:image / twitter:image / BlogPosting JSON-LD 的 image 都指向它。
分享到微信/Twitter/Slack 及 Google Discover 抓取时展示。

依赖（仅本机运行，不在 Zensical 构建镜像内）：
  pip install pillow
  系统字体：Noto Sans SC（/usr/share/fonts/truetype/NotoSansSC.ttf）

用法：
  python3 scripts/og_card.py
"""

from __future__ import annotations

from pathlib import Path

from PIL import Image, ImageDraw, ImageFilter, ImageFont

W, H = 1200, 630
ROOT = Path(__file__).resolve().parent.parent
OUT = ROOT / "docs/assets/images/og-card.png"
FONT = "/usr/share/fonts/truetype/NotoSansSC.ttf"

# 站点设计令牌（indigo primary / slate dark scheme，见 zensical.toml 与 site.css）
BG_TOP = (13, 17, 33)        # slate-950
BG_BOTTOM = (22, 30, 58)     # 深蓝渐变
INDIGO_LIGHT = (129, 140, 248)   # indigo-400（eyebrow / 强调）
WHITE = (241, 245, 249)
GRAY = (203, 213, 225)       # slate-300（tagline）
GRAY_DIM = (148, 163, 184)   # slate-400（域名）


def font(size: int, weight: str = "Regular") -> ImageFont.FreeTypeFont:
    f = ImageFont.truetype(FONT, size)
    try:  # NotoSansSC.ttf 是可变字体
        f.set_variation_by_name(weight)
    except Exception:
        pass
    return f


def spaced_text(draw: ImageDraw.ImageDraw, xy: tuple[int, int], text: str,
                f: ImageFont.FreeTypeFont, fill, tracking: int = 8):
    """带字距的文本（用于 eyebrow 小标）。"""
    x, y = xy
    for ch in text:
        draw.text((x, y), ch, font=f, fill=fill)
        x += draw.textlength(ch, font=f) + tracking


def main() -> None:
    img = Image.new("RGB", (W, H), BG_TOP)

    # 垂直渐变底
    grad = Image.new("RGB", (1, H))
    for y in range(H):
        t = y / H
        grad.putpixel((0, y), tuple(
            round(a + (b - a) * t) for a, b in zip(BG_TOP, BG_BOTTOM)))
    img.paste(grad.resize((W, H)))

    # 右上角 indigo 光晕（模糊椭圆）
    glow = Image.new("RGB", (W, H), (0, 0, 0))
    gd = ImageDraw.Draw(glow)
    gd.ellipse((W - 520, -260, W + 160, 320), fill=(49, 46, 129))
    gd.ellipse((-200, H - 200, 220, H + 240), fill=(30, 41, 59))
    glow = glow.filter(ImageFilter.GaussianBlur(120))
    img = Image.blend(img, Image.blend(img, glow, 0.55), 0.9)

    draw = ImageDraw.Draw(img)

    # 左侧强调竖条
    draw.rectangle((80, 150, 88, 246), fill=INDIGO_LIGHT)

    # 文本层
    spaced_text(draw, (116, 150), "CLOAKS · TECH BLOG", font(26), INDIGO_LIGHT, 10)
    draw.text((116, 192), "Cloaks｜Yiiewang", font=font(76, "Black"), fill=WHITE)
    draw.text((116, 312), "把复杂系统拆到第一性原理，再重新装配。",
              font=font(38), fill=GRAY)

    # 底部：域名 + 栏目
    draw.line((116, 430, 460, 430), fill=(51, 65, 85), width=2)
    draw.text((116, 470), "博客 · 知识库 · 学习笔记", font=font(30), fill=GRAY_DIM)
    spaced_text(draw, (116, 540), "CLOAKS.CN", font(30, "Bold"), (226, 232, 240), 6)

    OUT.parent.mkdir(parents=True, exist_ok=True)
    img.save(OUT, "PNG", optimize=True)
    print(f">>> {OUT.relative_to(ROOT)}（{OUT.stat().st_size // 1024} KB, {W}×{H}）")


if __name__ == "__main__":
    main()
