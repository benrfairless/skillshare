# Sync, Collect, Push & Pull

| Command | Direction | Description |
|---------|-----------|-------------|
| `sync` | Source → Targets | Distribute skills to all targets |
| `collect` | Targets → Source | Import skills from target(s) |
| `push` | Source → Remote | Git commit and push |
| `pull` | Remote → Source → Targets | Git pull and sync |

## sync

Distribute skills from source to all targets via symlinks.

```bash
skillshare sync                # Execute
skillshare sync --dry-run      # Preview
skillshare sync --force        # Override conflicts
```

## collect

Import skills from target(s) to source.

```bash
skillshare collect claude      # From specific target
skillshare collect --all       # From all targets
skillshare collect --dry-run   # Preview
```

## push

Git commit and push source to remote.

```bash
skillshare push                # Default message
skillshare push -m "message"   # Custom message
skillshare push --dry-run      # Preview
```

## pull

Git pull from remote and sync to all targets.

```bash
skillshare pull                # Pull + sync
skillshare pull --dry-run      # Preview
```

## Common Workflows

**Local editing:** Edit skill anywhere → `sync` (symlinks update source automatically)

**Import local changes:** `collect <target>` → `sync`

**Cross-machine sync:** Machine A: `push` → Machine B: `pull`
