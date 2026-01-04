# lah - A beautiful "ls" replacement

[![Status](https://img.shields.io/badge/status-beta-green.svg)](https://github.com/ipanardian/lah/releases)
[![Go](https://img.shields.io/badge/go-v1.25.x-blue.svg)](https://gitter.im/ipanardian/lah)
[![GitHub license](https://img.shields.io/badge/license-MIT-red.svg)](https://github.com/ipanardian/lah/blob/main/LICENSE)

A modern, colorful replacement for the Unix `ls -lah` command with box-drawn tables and git integration.

## Features

- **Beautiful Table Display** - Clean, box-drawn tables with colored borders for excellent readability
- **Smart Directory Listing** - Directories appear first by default, keeping your listing organized
- **Flexible Sorting Options** - Sort by name (default) or by modification time with newest first
- **Time-Aware Colors** - File ages are color-coded from recent (green) to old (gray), making it easy to spot fresh files
- **Human-Readable Sizes** - File sizes displayed in KB, MB, GB, or TB automatically
- **Always Show Hidden Files** - No need for extra flags, dotfiles are always visible
- **Git Integration** - See git status inline with added/changed lines count when in a repository
- **Color-Coded Permissions** - Read, write, and execute permissions are colorized for quick scanning
- **Terminal Responsive** - Automatically adjusts to fit your terminal width

## Installation

### Quick Install (Recommended)

**macOS and Linux:**
```bash
curl -sSL https://raw.githubusercontent.com/ipanardian/lah/main/install.sh | bash
```

Or with wget:
```bash
wget -qO- https://raw.githubusercontent.com/ipanardian/lah/main/install.sh | bash
```

### Manual Binary Download

Download the latest release from GitHub:
- [Linux (x86_64)](https://github.com/ipanardian/lah/releases/latest/download/lah-linux-amd64)
- [macOS (Intel)](https://github.com/ipanardian/lah/releases/latest/download/lah-darwin-amd64)
- [macOS (Apple Silicon)](https://github.com/ipanardian/lah/releases/latest/download/lah-darwin-arm64)

Then make it executable and move to your PATH:
```bash
chmod +x lah-*
sudo mv lah-* /usr/local/bin/lah
```

### Option 2: Build from Source

```bash
git clone https://github.com/ipanardian/lah.git
cd lah
make install
```

This will install `lah` to `~/bin`. Make sure `~/bin` is in your PATH.

### Option 3: System-wide Installation

```bash
git clone https://github.com/ipanardian/lah.git
cd lah
sudo make install-system
```

This installs `lah` to `/usr/local/bin` for all users.

## Usage

```bash
# Basic listing (always shows hidden files)
lah

# Sort by modified (newest first), reverse for oldest first
lah -t
lah -tr

# Show git status/counts inline
lah -g
lah -tg

# Combine flags
lah -tg -r
```

## Flags

- `-t, --sort-modified` — sort by modified time (newest first)
- `-r, --reverse` — reverse sort order
- `-g, --git` — show git status inline (+added/-deleted, (clean) when unchanged)

## License

MIT