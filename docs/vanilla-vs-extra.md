# 🍦 Vanilla vs. ✨ Extra Features

This document outlines the split between standard core features and automated enhancements in `bluefin-cli`.

## 📦 Binary Versions

You can build two versions of the CLI:
1.  **`bluefin-cli` (Standard)**: The standard tool for management.
2.  **`bluefin-cli-plus` (Plus)**: The enhanced version with deep automation.

---

## 🍦 Standard Features
These features are included in **both** binaries. They focus on providing tools and resources while leaving final configuration to the user.

| Feature | Command | Description |
|---------|---------|-------------|
| **App Bundles** | `install` | Install tool bundles (CLI, AI, CNCF, etc.) via Homebrew. |
| **Shell Config** | `shell` | Enable modern shell enhancements, aliases, and tool integrations. |
| **Init System** | `init` | Standard shell initialization logic. |
| **System Status** | `status` | Overview of enabled tools and shell status. |
| **Health Check** | `doctor` | Verify Homebrew health and configuration integrity. |
| **MOTD** | `motd` | Manage system info and helpful tips. |
| **Download Fonts** | `fonts` | Download recommended dev fonts (JetBrains Mono, Fira Code, etc.). |
| **Download Wallpapers** | `wallpapers` | Download monthly wallpaper packs from `ublue-os/tap`. |
| **Starship Themes** | `starship` | Quick switching between curated Starship prompt themes. |

---

## ✨ Plus Enhancements
These features are **only** included in the `bluefin-cli-plus` binary (compiled with `-tags extra`). They provide deep system integration and automation.

| Feature | Command | Description |
|---------|---------|-------------|
| **Sunset Switching** | `sunset` | **[Windows/WSL]** Automatically switches system themes and wallpapers based on solar time. |
| **Theme Automation** | `wallpapers` | Adds the ability to automatically generate and apply Windows themes after downloading wallpapers. |
| **Font Automation** | `fonts` | Adds automated terminal configuration for newly installed fonts (Coming soon). |

---

## 🏗️ Technical Implementation

### Build Tags
We use Go **build tags** to gate "Extra" logic. 
- Files marked with `//go:build extra` are excluded from the standard standard build.
- This ensures the standard binary remains lean and focused on core tasks.

### Dynamic UI
The interactive `bluefin-cli menu` automatically detects which features are compiled in and adjusts its options accordingly, hiding headers and commands that aren't available in the current binary.
