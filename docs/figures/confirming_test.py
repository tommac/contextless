"""Draw the confirming answer test (README, "Confirming test, the other 120
questions") as one image, 1600 x 900 pixels.

    python3 docs/figures/confirming_test.py            # writes confirming-test.png beside this file

The numbers are copied from the README table. Nothing is computed here.
Needs matplotlib.
"""
import pathlib
import sys

import matplotlib

matplotlib.use("Agg")
import matplotlib.pyplot as plt
from matplotlib.path import Path
from matplotlib.patches import PathPatch
from matplotlib.transforms import IdentityTransform

# arm, share of needed passages stated, mean input tokens, questions, highlighted
ROWS = [
    ("Only the exact needed paragraphs (control)", 0.982, 1214, 30, False),
    ("ContextlessAI packet with its map of notes", 0.882, 58517, 119, True),
    ("The whole conversation", 0.810, 247854, 30, False),
    ("Note lines, then the paragraphs asked for", 0.762, 31415, 120, False),
    ("ContextlessAI packet without the map", 0.706, 28984, 119, False),
    ("Search results of the same size", 0.672, 34768, 120, False),
    ("The question only (control)", 0.000, 130, 30, False),
]

SURFACE = "#fcfcfb"
INK = "#0b0b0b"
INK2 = "#52514e"
MUTED = "#898781"
GRID = "#e1e0d9"
BASELINE = "#c3c2b7"
ACCENT = "#2a78d6"
QUIET = "#c3c2b7"
FONT = "Helvetica Neue"

W, H, DPI = 1600, 900, 200


def bar(ax, fig, y, value, color, thick_px):
    """A bar from 0 to value, square at the baseline, rounded at the data end."""
    (x0, yc) = ax.transData.transform((0, y))
    (x1, _) = ax.transData.transform((value, y))
    if x1 - x0 < 1:
        return
    h = thick_px
    r = min(8, (x1 - x0) / 2, h / 2)
    top, bot = yc + h / 2, yc - h / 2
    k = 0.5523 * r
    verts = [
        (x0, bot), (x1 - r, bot),
        (x1 - r + k, bot), (x1, bot + r - k), (x1, bot + r),
        (x1, top - r),
        (x1, top - r + k), (x1 - r + k, top), (x1 - r, top),
        (x0, top), (x0, bot),
    ]
    codes = [
        Path.MOVETO, Path.LINETO,
        Path.CURVE4, Path.CURVE4, Path.CURVE4,
        Path.LINETO,
        Path.CURVE4, Path.CURVE4, Path.CURVE4,
        Path.LINETO, Path.CLOSEPOLY,
    ]
    fig.add_artist(PathPatch(Path(verts, codes), facecolor=color, edgecolor="none",
                             transform=IdentityTransform(), zorder=3))


def main(out):
    plt.rcParams["font.family"] = FONT
    fig = plt.figure(figsize=(W / DPI, H / DPI), dpi=DPI, facecolor=SURFACE)

    # text block above the plot
    fig.text(0.045, 0.925, "Answering questions about a 250,000-token conversation",
             fontsize=13.5, fontweight="bold", color=INK, ha="left", va="center")
    fig.text(0.045, 0.868,
             "Share of the needed passages the model stated, by what it was given. "
             "Confirming test, 120 held-out questions.",
             fontsize=7.6, color=INK2, ha="left", va="center")

    ax = fig.add_axes([0.345, 0.215, 0.475, 0.565], facecolor=SURFACE)
    n = len(ROWS)
    ax.set_xlim(0, 1.0)
    ax.set_ylim(n - 0.5, -0.5)
    for s in ax.spines.values():
        s.set_visible(False)
    ax.set_yticks([])
    ax.set_xticks([0, 0.25, 0.5, 0.75, 1.0])
    ax.set_xticklabels(["0", "0.25", "0.50", "0.75", "1.00"], fontsize=6.6, color=MUTED)
    ax.tick_params(axis="x", length=0, pad=5)
    for x in (0.25, 0.5, 0.75):
        ax.axvline(x, color=GRID, linewidth=0.6, zorder=1)
    ax.axvline(0, color=BASELINE, linewidth=0.8, zorder=2)

    fig.canvas.draw()
    for i, (label, value, tokens, q, hot) in enumerate(ROWS):
        bar(ax, fig, i, value, ACCENT if hot else QUIET, thick_px=34)
        # arm name, left of the baseline
        ax.text(-0.025, i, label, ha="right", va="center", fontsize=7.6,
                color=INK if hot else INK2, fontweight="bold" if hot else "normal")
        # value at the tip
        ax.text(value + 0.014, i, f"{value:.3f}", ha="left", va="center", fontsize=7.6,
                color=INK, fontweight="bold" if hot else "normal")
        # tokens and questions, as plain columns on the right
        ax.text(1.215, i, f"{tokens:,}", ha="right", va="center", fontsize=7.6,
                color=INK if hot else INK2, fontweight="bold" if hot else "normal",
                clip_on=False)
        ax.text(1.335, i, f"{q}", ha="right", va="center", fontsize=7.6,
                color=INK if hot else INK2, clip_on=False)
    ax.text(1.215, -0.92, "tokens sent", ha="right", va="center", fontsize=6.6, color=MUTED, clip_on=False)
    ax.text(1.335, -0.92, "questions", ha="right", va="center", fontsize=6.6, color=MUTED, clip_on=False)
    ax.text(0, -0.92, "share of needed passages stated", ha="left", va="center",
            fontsize=6.6, color=MUTED, clip_on=False)

    fig.text(0.045, 0.108,
             "Two private conversations, both by the author. Answers by Claude Opus 5.5, "
             "judged blind by two models; no person graded an answer.",
             fontsize=6.3, color=MUTED, ha="left", va="center")
    fig.text(0.045, 0.072,
             "The packet with the map is about twice the size of the packet without it. "
             "The test is not yet reproducible by others: the conversations are private.",
             fontsize=6.3, color=MUTED, ha="left", va="center")
    fig.text(0.045, 0.036, "github.com/tommac/contextless",
             fontsize=6.6, color=INK2, ha="left", va="center")

    fig.savefig(out, dpi=DPI, facecolor=SURFACE)
    print("wrote", out)


if __name__ == "__main__":
    here = pathlib.Path(__file__).resolve().parent
    main(sys.argv[1] if len(sys.argv) > 1 else str(here / "confirming-test.png"))
