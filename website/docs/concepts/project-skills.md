---
sidebar_position: 6
---

# Project Skills

Run skillshare at the project level — skills scoped to a single repository, shared via git.

## Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                    PROJECT MODE                                  │
│                                                                  │
│    ┌────────────────────────────────────────────────┐            │
│    │           .skillshare/skills/                  │            │
│    │   (project source — committed to git)          │            │
│    │                                                │            │
│    │   my-skill/   remote-skill/                    │            │
│    └────────────────────────┬───────────────────────┘            │
│                  sync ↓     │                                    │
│          ┌──────────────────┼──────────────────┐                 │
│          ▼                  ▼                  ▼                 │
│    ┌──────────┐       ┌──────────┐       ┌──────────┐            │
│    │  .claude │       │  .cursor │       │  custom  │            │
│    │  /skills │       │  /skills │       │  /skills │            │
│    └──────────┘       └──────────┘       └──────────┘            │
│                         TARGETS                                  │
│           (merge or symlink mode, per-target)                    │
└──────────────────────────────────────────────────────────────────┘
```

---

## Global vs Project

| | Global Mode | Project Mode |
|---|---|---|
| **Source** | `~/.config/skillshare/skills/` | `.skillshare/skills/` (project root) |
| **Config** | `~/.config/skillshare/config.yaml` | `.skillshare/config.yaml` |
| **Targets** | System-wide AI CLI directories | Per-project directories |
| **Sync mode** | Merge or symlink (per-target) | Merge or symlink (per-target, default merge) |
| **Tracked repos** | Supported (`--track`) | Supported (`--track -p`) |
| **Git integration** | Optional (`push`/`pull`) | Skills committed directly to project repo |
| **Scope** | All projects on machine | Single repository |

---

## `.skillshare/` Directory Structure

```
<project-root>/
├── .skillshare/
│   ├── config.yaml              # Targets + remote skills list
│   ├── .gitignore               # Ignores cloned remote/tracked skill dirs
│   └── skills/
│       ├── my-local-skill/      # Created manually or via `skillshare new`
│       │   └── SKILL.md
│       ├── remote-skill/        # Installed via `skillshare install -p`
│       │   ├── SKILL.md
│       │   └── .skillshare-meta.json
│       └── _team-skills/        # Installed via `skillshare install --track -p`
│           ├── .git/            # Git history preserved
│           ├── frontend/ui/
│           └── backend/api/
├── .claude/
│   └── skills/
│       ├── my-local-skill → ../../.skillshare/skills/my-local-skill
│       ├── remote-skill → ../../.skillshare/skills/remote-skill
│       ├── _team-skills__frontend__ui → ../../.skillshare/skills/_team-skills/frontend/ui
│       └── _team-skills__backend__api → ../../.skillshare/skills/_team-skills/backend/api
└── .cursor/
    └── skills/
        └── (same symlink structure as .claude/skills/)
```

---

## Config Format

`.skillshare/config.yaml`:

```yaml
targets:
  - claude-code                    # Known target (uses default path)
  - cursor                         # Known target
  - name: custom-ide               # Custom target with explicit path
    path: ./tools/ide/skills
    mode: symlink                  # Optional: "merge" (default) or "symlink"

skills:                            # Remote skills (installed via install -p)
  - name: pdf-skill
    source: anthropic/skills/pdf
  - name: review
    source: github.com/team/tools
```

**Targets** support two formats:
- **Short**: Just the target name (e.g., `claude-code`). Uses known default path, merge mode.
- **Long**: Object with `name`, optional `path`, and optional `mode` (`merge` or `symlink`). Supports relative paths (resolved from project root) and `~` expansion.

**Skills** list tracks remote installations only. Local skills don't need entries here.

---

## Auto-Detection

Skillshare automatically enters project mode when `.skillshare/config.yaml` exists in the current directory:

```bash
cd my-project/           # Has .skillshare/config.yaml
skillshare sync          # → Project mode (auto-detected)
skillshare status        # → Project mode (auto-detected)
```

No flags needed. To force a specific mode:

```bash
skillshare sync -p       # Force project mode
skillshare sync -g       # Force global mode
```

---

## Mode Restrictions

Project mode has some intentional limitations:

| Feature | Supported? | Notes |
|---------|-----------|-------|
| Merge sync mode | ✓ | Default, per-skill symlinks |
| Symlink sync mode | ✓ | Per-target via `skillshare target <name> --mode symlink -p` |
| `--track` repos | ✓ | Cloned to `.skillshare/skills/_repo/`, added to `.gitignore` |
| `--discover` | ✓ | Detect and add new targets to existing project config |
| `push` / `pull` | ✗ | Use git directly on the project repo |
| `collect` | ✓ | Collect local skills from project targets to `.skillshare/skills/` |
| `backup` / `restore` | ✗ | Not needed (project targets are reproducible) |

---

## Related

- [Project Setup Guide](/docs/guides/project-setup) — Step-by-step setup tutorial
- [Project Workflow](/docs/workflows/project-workflow) — Daily operations
- [Source & Targets](/docs/concepts/source-and-targets) — How source and targets work
- [Sync Modes](/docs/concepts/sync-modes) — Merge vs symlink
