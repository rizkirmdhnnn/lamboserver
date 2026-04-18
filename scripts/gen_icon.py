"""
Generate LamboServer app icon — geometric Lamborghini wireframe on dark background.

Design:
- 1024x1024 canvas
- Dark gray rounded-rect background (#2D2D2D) — macOS squircle style
- Darker inner circle (#1A1A1A)
- Simplified geometric Lamborghini side-profile outline in orange (#FF7A00)
  - Low-poly wedge shape: angular hood, flat roofline, pronounced wheel arches
  - 10-12 main polygon segments for the car body
  - Thick orange lines (14px) visible at small sizes
- "</>" code tag on windshield area in orange
"""

from PIL import Image, ImageDraw, ImageFont
import math
import os

SIZE = 1024
OUT = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "build", "appicon.png")

# Colors
BG_OUTER = "#2D2D2D"
BG_INNER = "#1A1A1A"
ORANGE = "#FF7A00"

img = Image.new("RGBA", (SIZE, SIZE), (0, 0, 0, 0))
draw = ImageDraw.Draw(img)

# ── 1. Outer rounded-rect (macOS squircle approximation) ──────────────────────
# Draw filled rounded rectangle for the app icon background
corner_r = 220  # ~22% of 1024, close to macOS squircle ratio
draw.rounded_rectangle([0, 0, SIZE - 1, SIZE - 1], radius=corner_r, fill=BG_OUTER)

# ── 2. Inner darker circle ────────────────────────────────────────────────────
cx, cy = SIZE // 2, SIZE // 2
circle_r = 440
draw.ellipse(
    [cx - circle_r, cy - circle_r, cx + circle_r, cy + circle_r],
    fill=BG_INNER
)

# ── 3. Geometric Lamborghini side profile (wireframe, orange lines) ───────────
# The car sits centered, slightly below mid, going left to right.
# All coords in 1024x1024 space.
#
# Design rationale:
#   - Lambo wedge: very low front, sharply raked windshield, flat roof, fast drop at rear
#   - Wheel arches as semicircular cutouts
#   - Keep segment count to ~10 for the main silhouette
#
# Car bounding box: roughly x: 100..924, y: 480..760
# Key x positions:
#   x_front_tip = 110   (nose of car, very low)
#   x_hood_rise = 260   (hood peak)
#   x_windshield_base = 390  (base of windshield / dashboard line)
#   x_windshield_top = 440   (top of windshield)
#   x_roof_end = 640   (rear end of roof)
#   x_rear_top = 670   (top rear corner, fast drop)
#   x_rear_base = 760  (rear lower corner before spoiler)
#   x_tail = 900       (tail end, low)
#
# y positions:
#   y_roof = 460        (roofline height)
#   y_hood_peak = 530   (hood peak)
#   y_body_bottom = 700 (bottom of body, above wheels)
#   y_front_low = 640   (front nose height)

LW = 14  # line width

# --- Car body outline (main polygon, clockwise from front nose) ---
car_body = [
    (115, 650),   # P0: front nose tip (low)
    (175, 530),   # P1: front hood rise
    (320, 490),   # P2: hood peak
    (400, 470),   # P3: windshield base (dashboard line)
    (445, 460),   # P4: top of windshield
    (640, 460),   # P5: roofline end
    (680, 490),   # P6: rear roof corner (slight drop)
    (750, 590),   # P7: rear pillar base
    (820, 660),   # P8: rear body
    (900, 665),   # P9: tail end
    (910, 700),   # P10: tail low
    (115, 700),   # P11: front bottom
]

# Draw body outline as line segments (not filled polygon)
for i in range(len(car_body) - 1):
    draw.line([car_body[i], car_body[i + 1]], fill=ORANGE, width=LW)
# Close front vertical
draw.line([car_body[-1], car_body[0]], fill=ORANGE, width=LW)

# --- Underbody line (connects front low to rear low cleanly) ---
# Already included in P11 -> P0 above

# --- Wheel arches (two semicircular cutouts from the bottom line) ---
# Front wheel: center at ~x=270, radius ~60
# Rear wheel: center at ~x=730, radius ~65
wheel_y = 700   # sits on underbody line

front_wheel_cx = 265
front_wheel_r = 65
rear_wheel_cx = 735
rear_wheel_r = 70

# Draw wheel arch (upper semicircle, going INTO the car from below)
def draw_wheel_arch(draw, cx, r, y_base, color, lw):
    """Draw an upward-facing semicircle (wheel arch) clipped into the underbody."""
    # We draw a semicircle arc above y_base
    bbox = [cx - r, y_base - r, cx + r, y_base + r]
    draw.arc(bbox, start=180, end=360, fill=color, width=lw)

draw_wheel_arch(draw, front_wheel_cx, front_wheel_r, wheel_y, ORANGE, LW)
draw_wheel_arch(draw, rear_wheel_cx, rear_wheel_r, wheel_y, ORANGE, LW)

# --- Wheels (full circles) ---
wheel_bottom_y = 760
front_wheel_bottom_cx = front_wheel_cx
rear_wheel_bottom_cx = rear_wheel_cx
wh_r = front_wheel_r

draw.ellipse(
    [front_wheel_bottom_cx - wh_r, wheel_bottom_y - wh_r,
     front_wheel_bottom_cx + wh_r, wheel_bottom_y + wh_r],
    outline=ORANGE, width=LW
)
draw.ellipse(
    [rear_wheel_bottom_cx - rear_wheel_r, wheel_bottom_y - rear_wheel_r,
     rear_wheel_bottom_cx + rear_wheel_r, wheel_bottom_y + rear_wheel_r],
    outline=ORANGE, width=LW
)

# --- Windshield line (inner diagonal line for windshield panel) ---
# From windshield base to roof start
draw.line([(400, 470), (445, 460)], fill=ORANGE, width=LW)

# ── 4. </> code tag on windshield area ───────────────────────────────────────
# Windshield area: roughly x=400-640, y=460-550 (interior of windshield panel)
# Place text centered at windshield panel mid
tag_cx = 500
tag_cy = 500

# Try to load a monospace/system font, fall back to default
font_size = 64
font = None
font_paths = [
    "/System/Library/Fonts/Menlo.ttc",
    "/System/Library/Fonts/Monaco.dfont",
    "/System/Library/Fonts/SFNSMono.ttf",
    "/Library/Fonts/Courier New.ttf",
]
for fp in font_paths:
    if os.path.exists(fp):
        try:
            font = ImageFont.truetype(fp, font_size)
            break
        except Exception:
            pass

tag_text = "</>"

if font:
    bbox = draw.textbbox((0, 0), tag_text, font=font)
    tw = bbox[2] - bbox[0]
    th = bbox[3] - bbox[1]
    draw.text((tag_cx - tw // 2, tag_cy - th // 2), tag_text, fill=ORANGE, font=font)
else:
    # Fallback: draw a simple </> manually with lines
    # Use small line segments to approximate the symbol
    s = 28  # half-size
    tx, ty = tag_cx, tag_cy
    # "<"
    draw.line([(tx - 60, ty), (tx - 40, ty - s)], fill=ORANGE, width=8)
    draw.line([(tx - 60, ty), (tx - 40, ty + s)], fill=ORANGE, width=8)
    # "/"
    draw.line([(tx - 20, ty + s), (tx + 10, ty - s)], fill=ORANGE, width=8)
    # ">"
    draw.line([(tx + 30, ty - s), (tx + 50, ty)], fill=ORANGE, width=8)
    draw.line([(tx + 50, ty), (tx + 30, ty + s)], fill=ORANGE, width=8)

# ── 5. Save ───────────────────────────────────────────────────────────────────
os.makedirs(os.path.dirname(OUT), exist_ok=True)
img.save(OUT, "PNG")
print(f"Saved: {OUT}")
print(f"Size: {img.size}")
