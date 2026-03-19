# 🗺️ Roadmap

This document outlines the potential future direction for `bluefin-cli`.

## 🚨 Immediate Priorities
- **Fix Homebrew Installation**: The README notes "NOT WORKING YET Via Homebrew". Need to investigate and fix the tap formula or CI process publishing it.
- **CI/CD Pipeline**: Ensure automated builds and releases are working correctly for cross-platform (if applicable) or at least Linux/AMD64.

## 🌟 Short-term Improvements
- **Use bold-brew fo
- **Expanded Bundles**:
    - Add `devops` bundle (Terraform, Ansible, etc.).
    - Add `gaming` bundle (Lutris, Steam tools, etc.) Flatpaks are now available in Brewfiles on Linux ans Casks on macOS for Steam
- **Enhanced Bling**:
    - Add `fzf` integration (it's a common request).
    - Add `ranger` for terminal file management.
- **Validation**: ~~Add a `doctor` command to verify environment health~~ (covered by `status`).


## 🔭 Long-term Vision
- **Configurability**:
- Move hardcoded bundle URLs and definitions to a config file (YAML/JSON) or fetch them dynamically from a remote manifest.
- Allow users to define custom bundles.
- **GUI Integration**: Potentially expose some of these features via a small GUI (using Fyne or similar) for non-terminal users, or integrate with GNOME Settings if possible.
