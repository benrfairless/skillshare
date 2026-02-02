# Troubleshooting

## Quick Fixes

| Problem | Solution |
|---------|----------|
| "config not found" | `skillshare init` |
| Target shows differences | `skillshare sync` |
| Lost source files | `cd ~/.config/skillshare/skills && git checkout -- .` |
| Skill not appearing | `skillshare sync` after install |
| Git push fails | Check remote: `git -C ~/.config/skillshare/skills remote -v` |

## Diagnostic Commands

```bash
skillshare doctor          # Check environment
skillshare status          # Overview
skillshare diff            # Show differences
ls -la ~/.claude/skills    # Check symlinks
```

## Recovery

```bash
skillshare backup          # Safety backup first
skillshare sync --dry-run  # Preview changes
skillshare sync            # Apply fix
```

## Git Recovery

```bash
cd ~/.config/skillshare/skills
git status                 # Check state
git checkout -- <skill>/   # Restore specific skill
git checkout -- .          # Restore all skills
```

## AI Assistant Notes

### Symlink Safety

- **merge mode** (default): Per-skill symlinks. Edit anywhere = edit source.
- **symlink mode**: Entire directory symlinked.

**Safe commands:** `skillshare uninstall`, `skillshare target remove`

**DANGEROUS:** `rm -rf` on symlinked skills deletes source!

### Non-Interactive Usage

AI cannot respond to CLI prompts. Always use flags:

```bash
# Good (non-interactive)
skillshare init --copy-from claude --all-targets --git
skillshare uninstall my-skill --force

# Bad (requires user input)
skillshare init
skillshare uninstall my-skill
```

### When to Use --dry-run

- First-time operations
- Before `sync`, `collect --all`, `restore`
- Before `install` from unknown sources
