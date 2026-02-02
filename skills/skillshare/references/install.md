# Install, Update, Uninstall & New

## install

Install skills from local path or git repository.

### Source Formats

```bash
# GitHub shorthand
user/repo                     # Browse repo for skills
user/repo/path/to/skill       # Direct path

# Full URLs
github.com/user/repo          # Discovers skills in repo
github.com/user/repo/path     # Direct subdirectory
https://github.com/...        # HTTPS URL
git@github.com:...            # SSH URL

# Local
~/path/to/skill               # Local directory
```

### Examples

```bash
skillshare install anthropics/skills              # Browse official skills
skillshare install anthropics/skills/skills/pdf   # Direct install
skillshare install ~/Downloads/my-skill           # Local
skillshare install github.com/team/repo --track   # Team repo
```

### Flags

| Flag | Description |
|------|-------------|
| `--name <n>` | Override skill name |
| `--force, -f` | Overwrite existing |
| `--update, -u` | Update if exists |
| `--track, -t` | Track for updates (preserves .git) |
| `--dry-run, -n` | Preview |

**Tracked repos:** Prefixed with `_`, nested with `__` (e.g., `_team__frontend__ui`).

**After install:** `skillshare sync`

## update

Update installed skills or tracked repositories.

- **Tracked repos (`_repo-name`):** Runs `git pull`
- **Regular skills:** Reinstalls from stored source metadata

```bash
skillshare update my-skill       # Update from stored source
skillshare update _team-skills   # Git pull tracked repo
skillshare update team-skills    # _ prefix is optional
skillshare update --all          # All tracked repos + skills
skillshare update --all -n       # Preview updates
skillshare update _repo --force  # Discard local changes
```

**Safety:** Tracked repos with uncommitted changes are skipped. Use `--force` to override.

**After update:** `skillshare sync`

## uninstall

Remove a skill from source.

```bash
skillshare uninstall my-skill          # With confirmation
skillshare uninstall my-skill --force  # Skip confirmation
```

**After uninstall:** `skillshare sync`

## new

Create a new skill template.

```bash
skillshare new <name>           # Create SKILL.md template
skillshare new <name> --dry-run # Preview
```

**After create:** Edit SKILL.md → `skillshare sync`
