# ADR 0002: Terminal graphics — braille for line art, half-blocks for sprites

**Status**: accepted (2026-07)

## Context
We want expressive visuals (the dino, the mini-game) that survive every
terminal/font, including phone terminals that render braille and other
East-Asian-ambiguous glyphs double-wide.

## Decision
- **Braille (2x4 mono dots)** for the compact header animation: a dense,
  portable mode for small line art. Sprite rows contain ONLY braille chars.
  Blanks use U+2800. This keeps each row the same width.
- **Half-block ▀ PixelCanvas (1x2 full-RGB pixels per cell)** for sprite
  scenes (the game): color carries more detail than dot density. The glyph is
  universal and width-safe. Chafa and notcurses use the same portable fallback.
- **No octants** (font support is too new) and **no sixel/kitty for now**
  (terminal-dependent). Image placement conflicts with cell-diff renderers.
  See issue #94 for a capability-detected future tier.
- Never place ambiguous-width glyphs (✓ · ❯ arrows) inside bordered boxes. Huh
  forms use a left accent bar without a right edge.
