# Domain glossary

- **Shell experience / bling**: the curated set of modern CLI tools (eza,
  bat, fzf, zoxide, atuin, starship, ugrep) wired into a user's shell rc
  files. Toggled per shell (bash/zsh/fish/powershell) via `internal/shell`.
- **Bundle**: a curated Brewfile of packages, fetched from
  projectbluefin/common by category id (ai, cli, cncf, ide, k8s…);
  installed with Homebrew (Linux/macOS) or winget (Windows).
- **MOTD**: message-of-the-day printed on new terminals (image info + tip),
  rendered with glow when available; where uwelcome is installed, uwelcome's
  greeting instead (bluefin-cli shows and toggles it, never greets twice).
- **Sunset switching**: Windows/WSL-only automation flipping light/dark
  theme + wallpapers at sunrise/sunset from geocoded coordinates.
- **Variants**: `bluefin-cli` (standard) and `bluefin-cli-plus`
  (`-tags extra`: wallpapers, fonts, sunset). One codebase, two binaries.
- **The shell (TUI)**: the persistent bubbletea program in
  `internal/tui/app` — a stack of Screens under a shared header (gradient
  wordmark + braille dino) and footer. Screen wrappers: MenuScreen,
  FormScreen (huh v2 host), TextScreen, RunnerScreen (captured task log),
  GameScreen (PixelCanvas easter egg).
- **Legacy bridge (`RunExternal`)**: releases the terminal for a genuinely
  external interactive program. Sole user: WSL→Windows sunset delegation.
- **Delivery**: semantic-release drives goreleaser on pushes to main —
  archives, deb/rpm, brew tap, scoop bucket, AUR, winget; one-liner
  installers (install.sh/install.ps1) and checksum-verified self-update
  (`internal/update`).
