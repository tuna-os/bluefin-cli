# ADR 0002: Terminal graphics — braille for line art, half-blocks for sprites

**Status**: accepted (2026-07)

## Context
We want expressive visuals (the dino, the mini-game) that survive every
terminal/font, including phone terminals that render braille and other
East-Asian-ambiguous glyphs double-wide.

## Decision
- **Braille (2x4 mono dots)** for the compact header animation: densest
  portable mode for small line art. Sprite rows contain ONLY braille chars
  (blanks are U+2800). Thus a width quirk moves each row by the same amount.
- **Half-block ▀ PixelCanvas (1x2 full-RGB pixels per cell)** for sprite
  scenes (the game). Color carries more than dot density. All terminals show
  ▀, at a safe width. chafa/notcurses use the same portable fallback.
- **No octants** (font support too new). **No sixel/kitty for now**: they
  depend on the terminal. Also, an image placement conflicts with a renderer
  that compares cells. Issue #94 has the future tier, which detects this
  capability.
- Never put glyphs of ambiguous width (✓ · ❯ arrows) inside a box with a
  border. For this reason, huh forms use a style with a bar on the left and
  no right edge.
