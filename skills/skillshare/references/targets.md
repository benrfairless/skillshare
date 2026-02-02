# Target Management

Manage AI CLI tool targets (Claude, Cursor, Windsurf, etc.).

## Commands

```bash
skillshare target list                        # List all targets
skillshare target claude                      # Show target info
skillshare target add myapp ~/.myapp/skills   # Add custom target
skillshare target remove myapp                # Remove target (safe)
```

## Sync Modes

```bash
skillshare target claude --mode merge         # Per-skill symlinks (default)
skillshare target claude --mode symlink       # Entire dir symlinked
```

| Mode | Description | Local Skills |
|------|-------------|--------------|
| `merge` | Individual symlinks per skill | Preserved |
| `symlink` | Single symlink for entire dir | Not possible |

## Safety

**Always use** `target remove` to unlink targets.

**NEVER** `rm -rf` on symlinked targets — this deletes the source!
