# Init Command

Initialize skillshare configuration.

**Source:** Always `~/.config/skillshare/skills` (use `--source` only if user explicitly requests).

## Flags

| Flag | Description |
|------|-------------|
| `--copy-from <name\|path>` | Import skills from target/path |
| `--no-copy` | Start with empty source |
| `--targets "claude,cursor"` | Specific targets |
| `--all-targets` | All detected targets |
| `--no-targets` | Skip target setup |
| `--git` | Initialize git repo |
| `--no-git` | Skip git init |
| `--discover` | Discover new AI tools (interactive) |
| `--discover --select "a,b"` | Non-interactive discovery |
| `--source <path>` | Custom source path |
| `--remote <url>` | Set git remote |
| `--dry-run` | Preview changes |

## AI Usage (Non-Interactive)

```bash
# Step 1: Check for existing skills
ls ~/.claude/skills ~/.cursor/skills 2>/dev/null | head -10

# Step 2a: Fresh start
skillshare init --no-copy --all-targets --git

# Step 2b: Import existing skills
skillshare init --copy-from claude --all-targets --git

# Step 3: Verify
skillshare status
```

## Adding New Targets Later

```bash
skillshare init --discover --select "windsurf,kilocode"
```
