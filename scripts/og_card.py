#!/usr/bin/env python3
"""生成默认分享卡 docs/assets/images/og-card.png（1200×630）。

v3 头像版：站点 favicon（docs/assets/favicon.ico，260px RGBA）居中放大，
身后一圈淡 indigo 光晕，底部只留 cloaks.cn 一行字。

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
MARK = ROOT / "docs/assets/favicon.ico"
FONT = "/usr/share/fonts/truetype/NotoSansSC.ttf"

# 站点设计令牌（indigo primary / slate dark scheme，见 zensical.toml 与 site.css）
BG_TOP = (10, 14, 26)        # slate-950 压暗一档
BG_BOTTOM = (18, 25, 48)     # 深蓝渐变
INDIGO = (129, 140, 248)     # indigo-400（光晕内层）
INDIGO_DEEP = (79, 70, 229)  # indigo-600（光晕主体）
GRAY = (148, 163, 184)       # slate-400（域名）

MARK_SIZE = 320  # 260px 源放大到 320，LANCZOS 轻度上采样


def font(size: int, weight: str = "Regular") -> ImageFont.FreeTypeFont:
    f = ImageFont.truetype(FONT, size)
    try:  # NotoSansSC.ttf 是可变字体
        f.set_variation_by_name(weight)
    except Exception:
        pass
    return f


def spaced_text(draw: ImageDraw.ImageDraw, xy: tuple[int, int], text: str,
                f: ImageFont.FreeTypeFont, fill, tracking: int = 8) -> None:
    """带字距的文本（左上角定位）。"""
    x, y = xy
    for ch in text:
        draw.text((x, y), ch, font=f, fill=fill)
        x += draw.textlength(ch, font=f) + tracking


def spaced_width(draw: ImageDraw.ImageDraw, text: str,
                 f: ImageFont.FreeTypeFont, tracking: int = 8) -> float:
    return (sum(draw.textlength(ch, font=f) for ch in text)
            + tracking * (len(text) - 1))


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
    gd.ellipse((250, -75, 950, 625), fill=INDIGO_DEEP + (100,))
    gd.ellipse((430, 105, 770, 445), fill=INDIGO + (40,))
    glow = glow.filter(ImageFilter.GaussianBlur(130))
    img = Image.alpha_composite(img.convert("RGBA"), glow).convert("RGB")

    # 3. favicon 头像式居中
    mark = Image.open(MARK).convert("RGBA")
    mark = mark.resize((MARK_SIZE, MARK_SIZE), Image.Resampling.LANCZOS)
    mx, my = (W - MARK_SIZE) // 2, 115
    img.paste(mark, (mx, my), mark)

    draw = ImageDraw.Draw(img)

    # 4. 域名（唯一文案）
    f_dom = font(34, "Medium")
    dom = "cloaks.cn"
    spaced_text(draw, ((W - spaced_width(draw, dom, f_dom, 12)) / 2, 478),
                dom, f_dom, GRAY, 12)

    OUT.parent.mkdir(parents=True, exist_ok=True)
    img.save(OUT, "PNG", optimize=True)
    print(f">>> {OUT.relative_to(ROOT)}（{OUT.stat().st_size // 1024} KB, {W}×{H}）")
    print(f"    favicon {mark.size[0]}px @ ({mx},{my})，文案仅 cloaks.cn")


if __name__ == "__main__":
    main()
