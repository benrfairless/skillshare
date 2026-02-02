---
name: skillshare
version: 0.8.2
description: |
  Syncs skills across AI CLI tools (Claude, Cursor, Windsurf, etc.) from a single source of truth.
  Use when: "sync skills", "install skill", "search skills", "list skills", "show skill status",
  "backup skills", "restore skills", "update skills", "new skill", "collect skills",
  "push/pull skills", "add/remove target", "find a skill for X", "is there a skill that can...",
  "how do I do X with skills", "skillshare init", "skillshare upgrade", "skill not syncing",
  "diagnose skillshare", "doctor", or any skill/target management across AI tools.
argument-hint: "[command] [target] [--dry-run]"
---

# Skillshare CLI

```
Source: ~/.config/skillshare/skills  ← Single source of truth
         ↓ sync (symlinks)
Targets: ~/.claude/skills, ~/.cursor/skills, ...
```

## Quick Start

```bash
skillshare status              # Check state
skillshare sync --dry-run      # Preview
skillshare sync                # Execute
```

## Commands

| Category | Commands |
|----------|----------|
| **Inspect** | `status`, `diff`, `list`, `doctor` |
| **Sync** | `sync`, `collect`, `push`, `pull` |
| **Skills** | `new`, `install`, `uninstall`, `update`, `search` |
| **Targets** | `target add/remove/list`, `backup`, `restore` |
| **Upgrade** | `upgrade [--cli\|--skill]` |

**Workflow:** Most commands require `sync` afterward to distribute changes.

## AI Usage Notes

### Non-Interactive Mode

AI cannot respond to CLI prompts. Always use flags:

```bash
# Init - check existing skills first
ls ~/.claude/skills ~/.cursor/skills 2>/dev/null | head -10

# Then run with appropriate flags
skillshare init --copy-from claude --all-targets --git  # If skills exist
skillshare init --no-copy --all-targets --git           # Fresh start

# Add new agents later
skillshare init --discover --select "windsurf,kilocode"
```

### Safety

**NEVER** `rm -rf` symlinked skills — deletes source. Always use:
- `skillshare uninstall <name>` to remove skills
- `skillshare target remove <name>` to unlink targets

### Finding Skills

When users ask "how do I do X" or "find a skill for...":

```bash
skillshare search <query>           # Interactive install
skillshare search <query> --list    # List only
skillshare search <query> --json    # JSON output
```

**Query examples:** `react performance`, `pr review`, `commit`, `changelog`

**No results?** Try different keywords, or offer to help directly.

## References

| Topic | File |
|-------|------|
| Init flags | [init.md](references/init.md) |
| Sync/collect/push/pull | [sync.md](references/sync.md) |
| Install/update/new | [install.md](references/install.md) |
| Status/diff/list/search | [status.md](references/status.md) |
| Target management | [targets.md](references/targets.md) |
| Backup/restore | [backup.md](references/backup.md) |
| Troubleshooting | [TROUBLESHOOTING.md](references/TROUBLESHOOTING.md) |
