---
name: skillshare
version: 0.6.0
description: Manages and syncs skills across AI CLI tools (Claude, Cursor, Codex) from a single source of truth. Use when asked to "sync my skills", "pull skills", "show skillshare status", "list my skills", "install a skill", "install tracked repo", "update skills", or manage skill targets.
argument-hint: "[command] [target] [--dry-run]"
---

# Skillshare CLI

Syncs skills across multiple AI CLI tools from a single source of truth.

## Quick Start

```bash
skillshare status              # See current state (always run first)
skillshare sync                # Push skills to all targets
skillshare sync --dry-run      # Preview changes before sync
skillshare pull claude         # Bring new skills from target to source
skillshare list                # Show installed skills
```

## AI Behavior Guide

| User Intent | Commands |
|-------------|----------|
| "sync my skills" | `skillshare sync` |
| "sync but show me first" | `skillshare sync --dry-run` → `skillshare sync` |
| "pull from Claude" | `skillshare pull claude` → `skillshare sync` |
| "pull all" | `skillshare pull --all` → `skillshare sync` |
| "pull from remote" | `skillshare pull --remote` |
| "push to remote" | `skillshare push` |
| "show status" | `skillshare status` |
| "what skills do I have" | `skillshare list` |
| "install X skill" | `skillshare install <source>` → `skillshare sync` |
| "install team repo" | `skillshare install <git-url> --track` → `skillshare sync` |
| "update skill" | `skillshare update <name>` → `skillshare sync` |
| "update all tracked repos" | `skillshare update --all` → `skillshare sync` |
| "remove X skill" | `skillshare uninstall <name>` → `skillshare sync` |
| "add cursor as target" | `skillshare target add cursor ~/.cursor/skills` |
| "something's broken" | `skillshare doctor` |
| "initialize skillshare" | See [Init Workflow](#init-workflow) |

## Init Workflow

**CRITICAL:** AI cannot respond to CLI prompts. Use flags for non-interactive mode.

### Init Checklist

Copy and track when initializing:
- [ ] Ask: "Do you have existing skills to copy?" → `--copy-from <name>` or `--no-copy`
- [ ] Ask: "Which CLI tools to sync?" → `--targets <list>`, `--all-targets`, or `--no-targets`
- [ ] Ask: "Initialize git?" → `--git` or `--no-git`
- [ ] Run: `skillshare init [flags]`
- [ ] Verify: `skillshare status`

### Quick Defaults

If user just says "initialize skillshare":
```bash
skillshare init --no-copy --all-targets --git
```

### Examples

```bash
skillshare init --copy-from claude --targets "claude,cursor" --git
skillshare init --no-copy --all-targets --git     # Fresh start
skillshare init --no-copy --no-targets --no-git   # Minimal
```

## Core Commands

| Command | Use Case |
|---------|----------|
| `status` | First command - see current state |
| `sync` | Push skills to all targets |
| `pull <target>` | Bring target's skills to source |
| `diff` | See differences between source and targets |
| `list` | Show installed skills and tracked repos |
| `install <source>` | Add skill from path or git repo |
| `install --track` | Install git repo as tracked repository |
| `update <name>` | Update skill or tracked repo |
| `update --all` | Update all tracked repos |
| `uninstall <name>` | Remove skill or tracked repo |
| `doctor` | Diagnose issues |

## Team Edition (v0.6.0)

Share skills with your team using tracked repositories:

```bash
# Install team skills repo
skillshare install github.com/team/skills --track

# Skills sync with _repo__ prefix
# _team-skills/frontend/ui → _team-skills__frontend__ui

# Update later
skillshare update _team-skills    # git pull
skillshare update --all           # update all tracked repos
```

**Features:**
- Tracked repos start with `_` prefix (e.g., `_team-skills`)
- Nested paths use `__` separator (e.g., `team__frontend__ui`)
- Auto-pruning removes orphaned symlinks on sync
- Name collision detection warns about duplicate SKILL.md names
- Uninstall checks for uncommitted changes

**Best Practice:** Use `{team}:{name}` format in SKILL.md to avoid name collisions:
```yaml
# In _acme-corp/frontend/ui/SKILL.md
name: acme:ui
```

## Nested Skills (Manual Organization)

Organize skills in subdirectories — skillshare auto-flattens on sync:

```
Source                    Target
───────────────────────────────────
my-skill/            →   my-skill/
work/api/            →   work__api/
personal/writing/    →   personal__writing/
```

A directory with `SKILL.md` is treated as a skill.

## Symlink Safety

**CRITICAL:** In `merge` mode, editing a skill in ANY target edits the source.

- **NEVER** use `rm -rf` on symlinked skills - this deletes the source
- Use `skillshare target remove <name>` to safely unlink targets
- Use `skillshare uninstall <name>` to safely remove skills

## Cross-Machine Workflow

```bash
# Machine A: push changes
skillshare push -m "Add new skill"

# Machine B: pull and sync
skillshare pull --remote
```

## Zero-Install Runner

If skillshare is not installed, run directly via curl:

```bash
curl -fsSL https://raw.githubusercontent.com/runkids/skillshare/main/skills/skillshare/scripts/run.sh | sh -s -- status
```

See [scripts/run.sh](scripts/run.sh) for the full runner script.

## References

- [status.md](references/status.md) - status, diff, list, doctor
- [sync.md](references/sync.md) - sync, pull, push
- [install.md](references/install.md) - install, uninstall
- [targets.md](references/targets.md) - target management
- [backup.md](references/backup.md) - backup, restore
- [init.md](references/init.md) - init flags reference
- [TROUBLESHOOTING.md](references/TROUBLESHOOTING.md) - Common issues and recovery
