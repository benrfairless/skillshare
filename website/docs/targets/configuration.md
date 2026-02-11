---
sidebar_position: 4
---

# Configuration

Configuration file reference for skillshare.

## Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    SKILLSHARE FILES                             │
│                                                                 │
│  ~/.config/skillshare/                                          │
│  ├── config.yaml          ← Configuration file                  │
│  ├── skills/              ← Source directory (your skills)      │
│  │   ├── my-skill/                                              │
│  │   ├── another/                                               │
│  │   └── _team-repo/      ← Tracked repository                  │
│  └── backups/             ← Automatic backups                   │
│      └── 2026-01-20.../                                         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Config File

**Location:** `~/.config/skillshare/config.yaml`

### Full Example

```yaml
# Source directory (where you edit skills)
source: ~/.config/skillshare/skills

# Default sync mode for new targets
mode: merge

# Targets (AI CLI skill directories)
targets:
  claude:
    path: ~/.claude/skills
    # mode: merge (inherits from default)

  codex:
    path: ~/.codex/skills
    mode: symlink  # Override default mode

  cursor:
    path: ~/.cursor/skills

  # Custom target
  myapp:
    path: ~/apps/myapp/skills

# Files to ignore during sync
ignore:
  - "**/.DS_Store"
  - "**/.git/**"
  - "**/node_modules/**"
  - "**/*.log"
```

---

## Fields

### `source`

Path to your skills directory (single source of truth).

```yaml
source: ~/.config/skillshare/skills
```

**Default:** `~/.config/skillshare/skills`

### `mode`

Default sync mode for all targets.

```yaml
mode: merge
```

| Value | Behavior |
|-------|----------|
| `merge` | Each skill symlinked individually. Local skills preserved. **(default)** |
| `symlink` | Entire target directory is one symlink. |

### `targets`

AI CLI skill directories to sync to.

```yaml
targets:
  <name>:
    path: <path>
    mode: <mode>  # optional, overrides default
```

**Example:**
```yaml
targets:
  claude:
    path: ~/.claude/skills

  codex:
    path: ~/.codex/skills
    mode: symlink

  custom:
    path: ~/my-app/skills
```

### `ignore`

Glob patterns for files to skip during sync.

```yaml
ignore:
  - "**/.DS_Store"
  - "**/.git/**"
  - "**/node_modules/**"
```

**Default patterns:**
- `**/.DS_Store`
- `**/.git/**`

### `audit`

Security audit configuration.

```yaml
audit:
  block_threshold: CRITICAL
```

| Field | Values | Default | Description |
|-------|--------|---------|-------------|
| `block_threshold` | `CRITICAL`, `HIGH`, `MEDIUM`, `LOW`, `INFO` | `CRITICAL` | Minimum severity to block `skillshare install` |

- `block_threshold` only controls when install is **blocked** — scanning always runs
- Use `--skip-audit` to bypass scanning for a single install
- Use `--force` to override a block (findings are still shown)

---

## Project Config

**Location:** `.skillshare/config.yaml` (in project root)

Project config uses a different format from global config.

```yaml
# Targets — string or object form
targets:
  - claude-code                    # String: known target with defaults
  - cursor
  - name: custom-ide               # Object: custom path and mode
    path: ./tools/ide/skills
    mode: symlink

# Remote skills — auto-managed by install/uninstall
skills:
  - name: pdf
    source: anthropic/skills/pdf
  - name: _team-skills
    source: github.com/team/skills
    tracked: true                  # Cloned with git history

# Audit — same as global
audit:
  block_threshold: HIGH
```

### `targets` (project)

Supports two YAML forms:

| Form | Example | When to use |
|------|---------|-------------|
| **String** | `- claude-code` | Known target, default path and merge mode |
| **Object** | `- name: x, path: ..., mode: ...` | Custom path or symlink mode |

### `skills` (project only)

Tracks remotely-installed skills. Auto-managed by `skillshare install -p` and `skillshare uninstall -p`.

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Skill directory name |
| `source` | Yes | GitHub URL or local path |
| `tracked` | No | `true` if installed with `--track` (default: `false`) |

:::tip Portable Manifest
`config.yaml` is a portable skill manifest. Anyone who clones the repo can run `skillshare install -p && skillshare sync` to reproduce the same setup.
:::

---

## Managing Config

### View current config

```bash
skillshare status
# Shows source, targets, modes
```

### Edit config directly

```bash
# Open in editor
$EDITOR ~/.config/skillshare/config.yaml

# Then sync to apply changes
skillshare sync
```

### Reset config

```bash
rm ~/.config/skillshare/config.yaml
skillshare init
```

---

## Custom Audit Rules

**Location:**

| Mode | Path |
|------|------|
| Global | `~/.config/skillshare/audit-rules.yaml` |
| Project | `.skillshare/audit-rules.yaml` |

Rules are merged in order: **built-in → global → project**. You can add new rules, disable built-in rules, or override severity.

```yaml
rules:
  # Add a custom rule
  - id: flag-todo
    severity: MEDIUM
    pattern: todo-comment
    message: "TODO comment found"
    regex: '(?i)\bTODO\b'

  # Disable a built-in rule
  - id: system-writes-0
    enabled: false
```

| Field | Required | Description |
|-------|----------|-------------|
| `id` | Yes | Unique rule identifier |
| `severity` | Yes | `CRITICAL`, `HIGH`, `MEDIUM`, `LOW`, `INFO` |
| `pattern` | Yes | Pattern category name |
| `message` | Yes | Human-readable finding description |
| `regex` | Yes | Regular expression to match |
| `exclude` | No | Suppress match when line also matches this regex |
| `enabled` | No | Set `false` to disable a built-in rule |

To scaffold a starter file:

```bash
skillshare audit --init-rules       # Global
skillshare audit --init-rules -p    # Project
```

See [audit command](/docs/commands/audit) for full details.

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `SKILLSHARE_CONFIG` | Override config file path |
| `GITHUB_TOKEN` | For API rate limit issues |

**Example:**
```bash
SKILLSHARE_CONFIG=~/custom-config.yaml skillshare status
```

---

## Skill Metadata

When you install a skill, skillshare creates a `.skillshare-meta.json` file:

```json
{
  "source": "anthropics/skills/skills/pdf",
  "type": "github",
  "installed_at": "2026-01-20T15:30:00Z",
  "repo_url": "https://github.com/anthropics/skills.git",
  "subdir": "skills/pdf",
  "version": "abc1234"
}
```

| Field | Description |
|-------|-------------|
| `source` | Original install source input |
| `type` | Source type (`github`, `local`, etc.) |
| `installed_at` | Installation timestamp |
| `repo_url` | Git clone URL (git sources only) |
| `subdir` | Subdirectory path (monorepo sources only) |
| `version` | Git commit hash at install time |

This is used by `skillshare update` and `skillshare check` to know where to fetch updates from.

**Don't edit this file manually.**

---

## Platform Differences

### macOS / Linux

```yaml
source: ~/.config/skillshare/skills
targets:
  claude:
    path: ~/.claude/skills
```

Uses symlinks.

### Windows

```yaml
source: %USERPROFILE%\.config\skillshare\skills
targets:
  claude:
    path: %USERPROFILE%\.claude\skills
```

Uses NTFS junctions (no admin required).

---

## Related

- [Source & Targets](/docs/concepts/source-and-targets) — Core concepts
- [Sync Modes](/docs/concepts/sync-modes) — Merge vs symlink
- [Environment Variables](/docs/reference/environment-variables) — All variables
