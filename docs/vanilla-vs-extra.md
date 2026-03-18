# 🍦 Vanilla vs. ✨ Extra Features

This document outlines the split between standard core features and automated enhancements in `bluefin-cli`.

## 🍦 Standard (Vanilla) Features
These are the core tools for managing your development environment and shell experience. They focus on manual control and system status.

| Feature | Command | Description |
|---------|---------|-------------|
| **App Bundles** | `install` | Install curated sets of tools (CLI, AI, CNCF, etc.) via Homebrew. |
| **Shell Config** | `shell` | Enable modern shell enhancements, aliases, and tool integrations. |
| **Init System** | `init` | Standard shell initialization logic for various shells. |
| **System Status** | `status` | Overview of enabled tools, shell status, and MOTD configuration. |
| **MOTD** | `motd` | Manage the Message of the Day, system information, and helpful tips. |
| **Diagnostics** | `countme` | Fedora-compatible anonymous usage reporting. |

## ✨ Extra Enhancements
These features provide automation, aesthetic improvements, and platform-specific polish. They go beyond simple management to provide an "opinionated" automated experience.

| Feature | Command | Description |
|---------|---------|-------------|
| **Sunset Switching** | `sunset` | **[Windows/WSL]** Automatically switches Windows/Native themes and wallpapers based on solar time (Sunrise/Sunset). |
| **Automated Fonts** | `fonts` | One-click installation of recommended development fonts (Fira Code, JetBrains Mono, etc.). |
| **Wallpaper Themes** | `wallpapers` | Discovery and installation of monthly wallpaper packs from `ublue-os/tap`. |
| **Starship Themes** | `starship` | Quick switching between curated Starship prompt themes. |

---

## 🏗️ Technical Implementation of the Split

### Command Grouping
The CLI help menu now visually separates these features using **Cobra Command Groups**. Running `bluefin-cli --help` will show distinct sections for each category.

### Menu Organization
The interactive `bluefin-cli menu` follows this same hierarchy, providing visual headers to guide users between core management and automated enhancements.

### Platform Availability
While "Vanilla" features are generally cross-platform, some "Extra" features (like Sunset switching) are optimized for specific environments (Windows/WSL) where deep desktop integration is required.
