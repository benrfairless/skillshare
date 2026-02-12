# Community Hub — Spec

> Date: 2026-02-12
> Status: Draft

## Goal

Build a public, community-maintained skill catalog that anyone can search via `skillshare search --hub`.

## Architecture

```
Contributors                    CI                        Consumers
───────────                    ──                        ─────────
PR: edit hub.json  ──→  GitHub Actions              CLI:
                        └─ validate hub.json          skillshare search --hub <raw-url>
                                                     Website:
                                                      /hub page fetches JSON, renders cards
```

## Hub JSON Schema

All hub.json files (community, enterprise, private) share the same schema:

```json
{
  "schemaVersion": 1,
  "skills": [
    {
      "name": "commit-helper",
      "description": "Git commit best practices for conventional commits",
      "source": "alice/commit-helper",
      "tags": ["git", "workflow"]
    },
    {
      "name": "react-patterns",
      "description": "React performance patterns and best practices",
      "source": "bob/react-patterns",
      "tags": ["react", "frontend", "performance"]
    }
  ]
}
```

### Skill Entry Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique skill name (lowercase, hyphens) |
| `description` | Yes | One-line description of what the skill does |
| `source` | Yes | Installable source (`owner/repo` or full URL) |
| `tags` | No | Classification tags for filtering/grouping |

### Tags Convention

- Lowercase, single words or hyphenated (e.g. `code-review`, `testing`)
- Keep to 1-3 tags per skill
- Common categories: `git`, `workflow`, `testing`, `security`, `frontend`, `backend`, `devops`, `docs`, `ai`
- Maintainers may request tag changes during PR review

## Phase 1: Community Repo

### Repo Structure

```
skillshare-community/
├── skillshare-hub.json          ← hand-curated, contributors edit via PR
├── CONTRIBUTING.md
└── .github/
    └── workflows/
        └── validate-pr.yml
```

### Contribution Flow

1. Fork repo
2. Add a new entry to `skillshare-hub.json`
3. Open PR
4. CI validates: JSON schema, no duplicate names, tags are valid
5. Maintainer reviews & merges

### CI: validate-pr.yml

```yaml
name: Validate PR
on:
  pull_request:
    paths: ['skillshare-hub.json']

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Validate hub.json
        run: |
          # Check valid JSON
          jq empty skillshare-hub.json || exit 1

          # Check required fields
          jq -e '.skills[] | select(.name == "" or .source == "")' skillshare-hub.json && {
            echo "ERROR: All skills must have name and source"
            exit 1
          } || true

          # Check no duplicate names
          dupes=$(jq -r '[.skills[].name] | group_by(.) | map(select(length > 1)) | flatten | .[]' skillshare-hub.json)
          if [ -n "$dupes" ]; then
            echo "Duplicate skill names: $dupes"
            exit 1
          fi
```

### Consumer Usage

```bash
# Search the community hub
skillshare search --hub https://raw.githubusercontent.com/<org>/skillshare-community/main/skillshare-hub.json

# Browse all
skillshare search --hub https://raw.githubusercontent.com/<org>/skillshare-community/main/skillshare-hub.json --json
```

## Phase 2: Website Hub Page

Add a `/hub` page to the Docusaurus site that renders the community catalog.

### Implementation

`website/src/pages/hub.tsx`:

- Client-side fetch from raw GitHub URL
- Render skill cards (name, description, source, tags)
- Search/filter input
- Tag-based filtering (click tag to filter)
- One-click copy of `skillshare install <source>` command
- Link to "Add your skill" → CONTRIBUTING.md

### Design Constraints

- No backend — pure static site + client-side fetch
- No accounts, no ratings, no analytics
- Fallback: show "failed to load" with direct URL for CLI usage

## Phase 3: Multiple Hubs (Future)

If the ecosystem grows, support multiple community hubs:

- A curated list of "known hubs" on the website
- Each org maintains their own `skillshare-hub.json`
- Website aggregates multiple sources on the `/hub` page

Not in scope for Phase 1-2.

## Seed Content Strategy

Before announcing, populate with 5-10 high-quality skills:

| Skill | Tags | Description |
|-------|------|-------------|
| commit-helper | git, workflow | Git commit best practices |
| code-review | workflow, quality | Code review guidelines |
| react-patterns | react, frontend | React performance patterns |
| python-style | python, quality | Python coding standards |
| deploy-checklist | devops, security | Deployment safety checklist |

Quality > quantity. Each seed skill should be genuinely useful, not placeholder.

## Hub Types

All use the same `skillshare-hub.json` schema. Difference is how they are produced:

| Type | How produced | Example |
|------|-------------|---------|
| Community | Hand-curated via PR | `skillshare-community` repo |
| Enterprise | Hand-curated or `skillshare hub index` | Internal company catalog |
| Private | `skillshare hub index` auto-generated | Local skill directory |

## Launch Checklist

- [ ] Create `skillshare-community` repo
- [ ] Add CI workflow (validate-pr)
- [ ] Add 5+ seed skills to hub.json
- [ ] Write CONTRIBUTING.md with clear instructions
- [ ] Add `/hub` page to website
- [ ] Add "Community Hub" link to website navbar
- [ ] Write announcement blog post
- [ ] Add badge to README: `Search community skills`
