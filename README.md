# gobrew

A terminal UI for managing Homebrew packages, built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

![Go](https://img.shields.io/badge/Go-1.24-00ADD8?style=flat&logo=go)
![Platform](https://img.shields.io/badge/platform-macOS-lightgrey?style=flat)
![License](https://img.shields.io/badge/license-MIT-blue?style=flat)

## Features

- View all top-level installed packages (`brew leaves`) with installed version, latest version, disk size, and status
- Outdated packages are highlighted with a warning badge
- Upgrade a single package or all outdated packages at once
- Uninstall packages
- Search Homebrew formulae and install directly from the TUI
- Run `brew doctor` with colorized, scrollable output
- Stat bar showing total package count and outdated count
- Adaptive color scheme — works on both dark and light terminals
- All destructive actions require `y/n` confirmation

## Screenshots

### Package List
```
  gobrew                                               homebrew manager
────────────────────────────────────────────────────────────────────────
 #   NAME             INSTALLED     LATEST        SIZE     STATUS
────────────────────────────────────────────────────────────────────────
 1   git              2.43.0        2.53.0        28M      ↑ outdated
 2   gh               2.52.0        2.87.2        54M      ↑ outdated
 3   node             24.1.0        25.6.1        112M     ↑ outdated
 4   postgresql@17    17.5          17.5          312M     ✓ ok
 5   redis            8.0.2         8.0.2         8M       ✓ ok
────────────────────────────────────────────────────────────────────────
 5 packages  ·  3 outdated
```

## Requirements

- macOS
- [Homebrew](https://brew.sh) installed
- Go 1.24+

## Installation

```bash
git clone https://github.com/akifzdemir/go-brew.git
cd go-brew
go install .
```

The `gobrew` binary will be placed in `$GOPATH/bin`. Make sure it is in your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

Then run from anywhere:

```bash
gobrew
```

## Key Bindings

### Package List

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `PgUp` / `PgDn` | Page up / down |
| `g` / `G` | Jump to top / bottom |
| `Enter` | View package detail |
| `u` | Upgrade selected (outdated only) |
| `U` | Upgrade all outdated |
| `x` | Uninstall selected |
| `/` | Search & install |
| `d` | Run brew doctor |
| `r` | Refresh package list |
| `q` / `Ctrl+C` | Quit |

### Detail View

| Key | Action |
|-----|--------|
| `↑↓` / `j/k` | Scroll |
| `u` | Upgrade package |
| `x` | Uninstall package |
| `Esc` | Back to list |

### Search View

| Key | Action |
|-----|--------|
| `Enter` | Search / install selected |
| `Tab` | Toggle input focus |
| `↑↓` | Navigate results |
| `Esc` | Back to list |

> All destructive actions (`u`, `U`, `x`) prompt for `y/n` confirmation before executing.

## Tech Stack

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) — UI components (spinner, text input)
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — styling and layout
