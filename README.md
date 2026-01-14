# skillshare

Share skills across AI CLI tools (Claude Code, Codex CLI, Cursor, Gemini CLI, OpenCode).

## The Problem

Each AI CLI tool has its own skills directory:

```
~/.claude/skills/
~/.codex/skills/
~/.cursor/skills/
~/.gemini/antigravity/skills/
~/.config/opencode/skills/
```

Keeping them in sync manually is tedious.

## The Solution

`skillshare` maintains a single source directory and symlinks skills to all CLI tools:

```
~/.config/skillshare/
    ├── config.yaml
    └── skills/           <- Your shared skills live here
        ├── my-skill/
        └── another-skill/

~/.claude/skills/
    ├── my-skill     -> ~/.config/skillshare/skills/my-skill (symlink)
    ├── another-skill -> ~/.config/skillshare/skills/another-skill (symlink)
    └── local-only/   <- Local skills are preserved
```

## Installation

### macOS

```bash
# Apple Silicon (M1/M2/M3/M4)
curl -sL https://github.com/runkids/skillshare/releases/latest/download/skillshare_0.1.0_darwin_arm64.tar.gz | tar xz
sudo mv skillshare /usr/local/bin/

# Intel
curl -sL https://github.com/runkids/skillshare/releases/latest/download/skillshare_0.1.0_darwin_amd64.tar.gz | tar xz
sudo mv skillshare /usr/local/bin/
```

### Linux

```bash
# x86_64
curl -sL https://github.com/runkids/skillshare/releases/latest/download/skillshare_0.1.0_linux_amd64.tar.gz | tar xz
sudo mv skillshare /usr/local/bin/

# ARM64
curl -sL https://github.com/runkids/skillshare/releases/latest/download/skillshare_0.1.0_linux_arm64.tar.gz | tar xz
sudo mv skillshare /usr/local/bin/
```

### Windows

Download from [Releases](https://github.com/runkids/skillshare/releases) and add to PATH.

### Homebrew (macOS/Linux)

```bash
brew install runkids/tap/skillshare
```

### Verify Installation

```bash
skillshare version
```

### Uninstall

```bash
# Homebrew
brew uninstall skillshare

# Manual (curl install)
sudo rm /usr/local/bin/skillshare

# Config and data (optional)
rm -rf ~/.config/skillshare
```

## Quick Start

```bash
# 1. Initialize with default source directory (~/.skills)
skillshare init

# 2. Check detected targets
skillshare status

# 3. Sync (migrate existing skills + create symlinks)
skillshare sync
```

## Usage

### Initialize

```bash
# Use default source (~/.skills)
skillshare init

# Or specify custom source
skillshare init --source ~/my-skills
```

This will:
- Create the source directory
- Detect installed CLI tools
- Create config at `~/.config/skillshare/config.yaml`

### Sync

```bash
# Preview what will happen
skillshare sync --dry-run

# Actually sync
skillshare sync
```

On first sync, existing skills are **migrated** to the source directory, then symlinks are created.

### Status

```bash
skillshare status
```

Shows:
- Source directory and skill count
- Each target's status (linked, has files, not exist, conflict)

### Manage Targets

```bash
# List targets
skillshare target list

# Add custom target
skillshare target add myapp ~/.myapp/skills

# Unlink target (restore skills and remove from config)
skillshare target remove myapp

# Unlink all targets
skillshare target remove --all
```

## Configuration

Config file: `~/.config/skillshare/config.yaml`

```yaml
source: ~/.config/skillshare/skills
mode: merge   # default mode for all targets
targets:
  claude:
    path: ~/.claude/skills
    mode: symlink   # override: use full directory symlink
  codex:
    path: ~/.codex/skills
  cursor:
    path: ~/.cursor/skills
  gemini:
    path: ~/.gemini/antigravity/skills
  opencode:
    path: ~/.config/opencode/skills
ignore:
  - "**/.DS_Store"
  - "**/.git/**"
```

### Sync Modes

| Mode | Behavior |
|------|----------|
| `merge` (default) | Each skill is symlinked individually. Local skills in target are preserved. |
| `symlink` | Entire directory becomes a symlink to source. All targets share the same skills. |

Use `symlink` mode when you want all targets to share exactly the same skills.

## How It Works

1. **init**: Detects CLI tools, optionally copy from existing skills
2. **sync**:
   - Backup targets with local skills before sync
   - Create symlinks (merge mode: per-skill, symlink mode: whole directory)
3. **status**: Check symlink health
4. **target remove**: Backup, unlink symlinks, restore skills

## Backups

Automatic backups are created before `sync` and `target remove` operations.

Location: `~/.config/skillshare/backups/<timestamp>/<target>/`

## License

MIT
