"""
Generate LamboServer DMG installer background — 660x400 PNG + @2x Retina variant.

Design:
- 660x400 canvas (1x) and 1320x800 canvas (2x, rendered natively at scale, not resized)
- Light/neutral palette — subtle vertical gradient from #FFFFFF to #F5F5F7
  (safe under Light and Dark Finder per D-11)
- Source logo (build/appicon.png, 1024x1024 — D-09) placed on the LEFT half of the canvas,
  resized with Image.LANCZOS to 180x180 (at 1x) so it reads as brand artwork in the
  background. create-dmg overlays the actual .app icon on top of this region at (165, 220)
  per D-15 — the logo here is decorative underlay.
- Subtle gray chevron/arrow pointing RIGHT (from ~x=270 to ~x=400 at 1x) towards the
  Applications drop-link slot create-dmg places at (495, 220) per D-15.
- Caption "Drag to Applications" (D-SPECIFIC canonical copy) centered above the arrow
  at (335, 170) at 1x, rendered in a system sans-serif font with graceful fallback.

Palette (hex, per D-11):
- BG_LIGHT = "#F5F5F7" (off-white bottom of gradient)
- ACCENT   = "#FF7A00" (brand orange — matches gen_icon.py)
- TEXT     = "#1D1D1F" (near-black for caption)
- ARROW    = "#8E8E93" (mid-gray chevron — "subtle" per D-11)

Outputs (per D-07, D-12):
- build/darwin/dmg-background.png       (660x400)
- build/darwin/dmg-background@2x.png    (1320x800)

Run: python3 scripts/gen-dmg-background.py
Pillow is a dev-only dependency per D-08; the committed PNGs are what build-dmg.sh reads.
"""

from PIL import Image, ImageDraw, ImageFont
import os

# Paths (resolved from script location — matches scripts/gen_icon.py idiom)
REPO_ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SRC_LOGO  = os.path.join(REPO_ROOT, "build", "appicon.png")
OUT_1X    = os.path.join(REPO_ROOT, "build", "darwin", "dmg-background.png")
OUT_2X    = os.path.join(REPO_ROOT, "build", "darwin", "dmg-background@2x.png")

# Colors
BG_LIGHT = "#F5F5F7"
ACCENT   = "#FF7A00"
TEXT     = "#1D1D1F"
ARROW    = "#8E8E93"


def render(scale: int) -> Image.Image:
    """Render the DMG background at the given integer scale (1 for @1x, 2 for @2x)."""
    W, H = 660 * scale, 400 * scale
    img = Image.new("RGBA", (W, H), BG_LIGHT)
    draw = ImageDraw.Draw(img)

    # ── 1. Background gradient (subtle vertical, #FFFFFF → BG_LIGHT) ─────────────
    # Fill horizontal 1px strips; interpolate linearly between the top color
    # (pure white) and BG_LIGHT at the bottom. Deterministic — no random state.
    top_rgb = (0xFF, 0xFF, 0xFF)
    bot_rgb = (0xF5, 0xF5, 0xF7)
    for y in range(H):
        t = y / max(H - 1, 1)
        r = int(round(top_rgb[0] + (bot_rgb[0] - top_rgb[0]) * t))
        g = int(round(top_rgb[1] + (bot_rgb[1] - top_rgb[1]) * t))
        b = int(round(top_rgb[2] + (bot_rgb[2] - top_rgb[2]) * t))
        draw.line([(0, y), (W - 1, y)], fill=(r, g, b, 255))

    # ── 2. Load and place source logo on left half (D-09) ────────────────────────
    # Read the canonical 1024x1024 app icon from build/appicon.png and downscale
    # with LANCZOS for clean resampling. Place centered around (165, 220) at 1x —
    # same column as the create-dmg LamboServer.app icon slot (D-15), so the
    # background logo sits beneath / blends with the real app icon overlay.
    logo = Image.open(SRC_LOGO).convert("RGBA")
    logo_px = 180 * scale
    logo = logo.resize((logo_px, logo_px), Image.LANCZOS)
    logo_x = (75 * scale)
    logo_y = (110 * scale)
    img.paste(logo, (logo_x, logo_y), logo)

    # ── 3. Arrow pointing from center-left to Applications slot (D-15) ───────────
    # Simple stroked right-pointing chevron: horizontal shaft + two head strokes.
    # At 1x: shaft (270,220)→(400,220), upper head (400,220)→(375,200),
    # lower head (400,220)→(375,240). All coords scaled uniformly.
    lw = 8 * scale
    shaft_start = (270 * scale, 220 * scale)
    shaft_end   = (400 * scale, 220 * scale)
    head_upper  = (375 * scale, 200 * scale)
    head_lower  = (375 * scale, 240 * scale)
    draw.line([shaft_start, shaft_end], fill=ARROW, width=lw)
    draw.line([shaft_end, head_upper], fill=ARROW, width=lw)
    draw.line([shaft_end, head_lower], fill=ARROW, width=lw)

    # ── 4. "Drag to Applications" caption above arrow ────────────────────────────
    # Canonical copy per CONTEXT §Specific Ideas. Centered around (335, 170) at 1x —
    # above the shaft midpoint. Font loading uses the system sans-serif fallback
    # list; if no font resolves, the caption is skipped (graceful — do not crash).
    caption = "Drag to Applications"
    caption_size = 28 * scale
    font = None
    font_paths = [
        "/System/Library/Fonts/SFNS.ttf",
        "/System/Library/Fonts/Helvetica.ttc",
        "/System/Library/Fonts/HelveticaNeue.ttc",
        "/Library/Fonts/Arial.ttf",
    ]
    for fp in font_paths:
        if os.path.exists(fp):
            try:
                font = ImageFont.truetype(fp, caption_size)
                break
            except Exception:
                pass

    if font:
        bbox = draw.textbbox((0, 0), caption, font=font)
        tw = bbox[2] - bbox[0]
        th = bbox[3] - bbox[1]
        cx = 335 * scale
        cy = 170 * scale
        draw.text((cx - tw // 2, cy - th // 2), caption, fill=TEXT, font=font)

    return img


# ── 5. Save both variants ────────────────────────────────────────────────────
for scale, out in ((1, OUT_1X), (2, OUT_2X)):
    img = render(scale)
    os.makedirs(os.path.dirname(out), exist_ok=True)
    img.save(out, "PNG")
    print(f"Saved: {out} ({img.size})")
