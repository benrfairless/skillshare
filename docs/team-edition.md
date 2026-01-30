# Team Edition

Share skills across your entire team. Clone once, update with one command, sync everywhere.

## Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    TEAM EDITION WORKFLOW                        │
│                                                                 │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │              GitHub: team/shared-skills                 │   │
│   │   frontend/ui/   backend/api/   devops/deploy/          │   │
│   └─────────────────────────────────────────────────────────┘   │
│                              │                                  │
│              skillshare install --track                         │
│                              │                                  │
│                              ▼                                  │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │  Alice's Machine          Bob's Machine                 │   │
│   │  _team-skills/            _team-skills/                 │   │
│   │  ├── frontend/ui/         ├── frontend/ui/              │   │
│   │  ├── backend/api/         ├── backend/api/              │   │
│   │  └── devops/deploy/       └── devops/deploy/            │   │
│   └─────────────────────────────────────────────────────────┘   │
│                              │                                  │
│              skillshare update _team-skills                     │
│                              │                                  │
│                              ▼                                  │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │  Everyone gets updates instantly                        │   │
│   └─────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Quick Start

### For Team Members

```bash
# 1. Install team skills repo
skillshare install github.com/your-team/skills --track

# 2. Sync to all your AI CLIs
skillshare sync

# Done! You now have all team skills.
```

### For Team Leads

```bash
# Create a skills repo on GitHub
# Add your team's skills to it
# Share the install command with your team
```

---

## Why Team Edition?

| Without Team Edition | With Team Edition |
|---------------------|-------------------|
| "Hey, grab the latest deploy skill from Slack" | `skillshare update --all` |
| Copy-paste skills between machines | One command installs everything |
| "Which version of the skill do you have?" | Everyone syncs from same source |
| Skills scattered across docs/repos | One curated repo for the team |

---

## Features

### Tracked Repositories

Install a git repo that stays connected to its source:

```bash
skillshare install github.com/team/skills --track
```

**What happens:**
```
┌─────────────────────────────────────────────────────────────────┐
│ skillshare install github.com/team/skills --track               │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 1. git clone github.com/team/skills                             │
│    → ~/.config/skillshare/skills/_team-skills/                  │
│    (note: _ prefix indicates tracked repo)                      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. .git directory preserved (unlike regular install)            │
│    This allows updates via git pull                             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. skillshare sync                                              │
│    Skills distributed to all targets                            │
└─────────────────────────────────────────────────────────────────┘
```

### Update with One Command

```bash
skillshare update _team-skills    # Update specific repo
skillshare update --all           # Update all tracked repos + skills
skillshare sync                   # Sync changes to targets
```

**What happens:**
```
┌─────────────────────────────────────────────────────────────────┐
│ skillshare update _team-skills                                  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ cd ~/.config/skillshare/skills/_team-skills                     │
│ git pull origin main                                            │
│                                                                 │
│ → New skills added                                              │
│ → Modified skills updated                                       │
│ → Deleted skills removed                                        │
└─────────────────────────────────────────────────────────────────┘
```

---

### Nested Skills

Organize skills in folders. Skillshare flattens them for AI CLIs:

```
┌─────────────────────────────────────────────────────────────────┐
│           SOURCE                      TARGET                    │
│  (your organization)            (what AI CLI sees)              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  _team-skills/                                                  │
│  ├── frontend/                                                  │
│  │   ├── react/          ───►   _team-skills__frontend__react/  │
│  │   └── vue/            ───►   _team-skills__frontend__vue/    │
│  ├── backend/                                                   │
│  │   └── api/            ───►   _team-skills__backend__api/     │
│  └── devops/                                                    │
│      └── deploy/         ───►   _team-skills__devops__deploy/   │
│                                                                 │
│  personal/                                                      │
│  └── notes/              ───►   personal__notes/                │
│                                                                 │
│  my-skill/               ───►   my-skill/                       │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

Legend:
  • _ prefix = tracked repository
  • __ (double underscore) = path separator
  • Nested paths flattened for CLI compatibility
```

**Create nested skills:**
```bash
mkdir -p ~/.config/skillshare/skills/work/frontend/react
cat > ~/.config/skillshare/skills/work/frontend/react/SKILL.md << 'EOF'
---
name: react-helper
description: React development assistance
---
# React Helper
EOF

skillshare sync
# Creates: work__frontend__react/ in all targets
```

---

### Auto-Pruning

Delete a skill from source, and it's automatically removed from targets:

```bash
rm -rf ~/.config/skillshare/skills/_team/old-skill
skillshare sync
```

**Output:**
```
✓ claude: merged (5 linked, 0 local, 0 updated, 1 pruned)
                                              ^^^^^^^^
                                              auto-removed
```

**Safety rules:**
- Only removes symlinks pointing to source
- Only touches directories with `_` prefix or `__` in name
- Local-only skills are never touched

---

### Collision Detection

When multiple skills have the same `name` field, sync warns you:

```
┌─────────────────────────────────────────────────────────────────┐
│ skillshare sync                                                 │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ Name conflicts detected                                         │
│ ─────────────────────────────────────────                       │
│ ! Skill name 'deploy' is defined in multiple places:            │
│ →   - _team-a/devops/deploy                                     │
│ →   - _team-b/infra/deploy                                      │
│ → CLI tools may not distinguish between them.                   │
│ → Suggestion: Rename one in SKILL.md                            │
└─────────────────────────────────────────────────────────────────┘
```

**Best practice — namespace your skills:**
```yaml
# In _acme-corp/frontend/ui/SKILL.md
name: acme:ui

# In _other-team/frontend/ui/SKILL.md
name: other:ui
```

---

## Commands Reference

| Command | Description |
|---------|-------------|
| `install <url> --track` | Clone repo as tracked repository |
| `update _repo-name` | Git pull tracked repo |
| `update --all` | Update all tracked repos + skills with metadata |
| `list` | Show skills and tracked repos |
| `status` | Show repo status (up-to-date/has changes) |
| `uninstall _repo-name` | Remove tracked repo (checks uncommitted) |

### Examples

```bash
# Install team repo
skillshare install github.com/team/skills --track
skillshare install github.com/team/skills --track --name custom-name

# Check status
skillshare list
skillshare status

# Update
skillshare update _team-skills
skillshare update --all
skillshare sync

# Uninstall
skillshare uninstall _team-skills       # Checks for uncommitted changes
skillshare uninstall _team-skills -f    # Force
```

---

## Team Setup Guide

### Step 1: Create Team Skills Repo

```bash
# Create a new repo on GitHub: your-org/team-skills

# Clone locally
git clone git@github.com:your-org/team-skills.git
cd team-skills

# Create skill structure
mkdir -p frontend/react backend/api devops/deploy

# Add skills (each needs SKILL.md)
cat > frontend/react/SKILL.md << 'EOF'
---
name: org:react
description: React development standards
---
# React Helper
[Your skill content]
EOF

# Push
git add . && git commit -m "Initial skills" && git push
```

### Step 2: Share with Team

Send your team this command:
```bash
skillshare install github.com/your-org/team-skills --track && skillshare sync
```

### Step 3: Ongoing Updates

When you update the repo:
```bash
# You: push changes to GitHub
git add . && git commit -m "Update react skill" && git push

# Team: pull updates
skillshare update --all && skillshare sync
```

---

## Related

- [install.md](install.md) — Install commands
- [sync.md](sync.md) — Sync operations
- [cross-machine.md](cross-machine.md) — Personal cross-machine sync
