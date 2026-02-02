# Status & Inspection Commands

## status

Overview of source, targets, and sync state.

```bash
skillshare status
```

## diff

Show differences between source and targets.

```bash
skillshare diff                # All targets
skillshare diff claude         # Specific target
```

## list

List installed skills.

```bash
skillshare list                # Basic list
skillshare list --verbose      # With source info
```

## search

Search GitHub for skills (repos containing SKILL.md).

```bash
skillshare search <query>           # Interactive (select to install)
skillshare search <query> --list    # List only
skillshare search <query> --json    # JSON output
skillshare search <query> -n 10     # Limit results (default: 20)
```

**Requires:** GitHub auth (`gh` CLI or `GITHUB_TOKEN` env var).

**Query examples:**
- `react performance` - Performance optimization
- `pr review` - Code review skills
- `commit` - Git commit helpers
- `changelog` - Changelog generation

## doctor

Diagnose configuration and environment issues.

```bash
skillshare doctor
```

## upgrade

Upgrade CLI binary and/or built-in skillshare skill.

```bash
skillshare upgrade              # Both CLI + skill
skillshare upgrade --cli        # CLI only
skillshare upgrade --skill      # Skill only
skillshare upgrade --force      # Skip confirmation
skillshare upgrade --dry-run    # Preview
```

**After upgrading skill:** `skillshare sync`
