---
sidebar_position: 3
---

# audit

Scan installed skills for security threats and malicious patterns.

```bash
skillshare audit              # Scan all skills
skillshare audit <name>       # Scan a specific skill
skillshare audit -p           # Scan project skills
```

## What It Detects

### CRITICAL (blocks installation)

| Pattern | Description |
|---------|------------|
| `prompt-injection` | "Ignore previous instructions", "SYSTEM:", "You are now", etc. |
| `data-exfiltration` | `curl`/`wget` commands sending environment variables externally |
| `credential-access` | Reading `~/.ssh/`, `.env`, `~/.aws/credentials` |
| `hidden-unicode` | Zero-width characters that hide content from human review |

### HIGH (strong warning)

| Pattern | Description |
|---------|------------|
| `destructive-commands` | `rm -rf /`, `chmod 777`, `sudo`, `dd if=`, `mkfs` |
| `obfuscation` | Base64 decode pipes, long base64-encoded strings |

### MEDIUM (informational)

| Pattern | Description |
|---------|------------|
| `suspicious-fetch` | URLs used in command context (`curl`, `wget`, `fetch`) |
| `system-writes` | Commands writing to `/usr`, `/etc`, `/var` |

## Example Output

```
┌─ skillshare audit ───────────────────┐
│  Scanning 12 skills for threats      │
└──────────────────────────────────────┘

[1/12] ✓ react-best-practices         0.1s
[2/12] ✓ typescript-patterns           0.1s
[3/12] ✗ suspicious-skill              0.2s
       ├─ CRITICAL: Prompt injection (SKILL.md:15)
       │  "Ignore all previous instructions and..."
       └─ HIGH: Destructive command (SKILL.md:42)
          "rm -rf / # clean up"
[4/12] ! frontend-utils                0.1s
       └─ MEDIUM: URL in command context (SKILL.md:3)

┌─ Summary ────────────────────────────┐
│  Scanned:  12 skills                 │
│  Passed:   10                        │
│  Warning:  1 (1 medium)             │
│  Failed:   1 (1 critical, 1 high)   │
└──────────────────────────────────────┘
```

## Install-time Scanning

Skills are automatically scanned during installation. If **CRITICAL** threats are detected, the installation is blocked:

```bash
skillshare install /path/to/evil-skill
# Error: security audit failed: critical threats detected in skill

skillshare install /path/to/evil-skill --force
# Installs with warnings (use with caution)
```

HIGH and MEDIUM findings are shown as warnings but don't block installation.

## Web UI

The audit feature is also available in the web dashboard at `/audit`:

```bash
skillshare ui
# Navigate to Audit page → Click "Run Audit"
```

![Security Audit page in web dashboard](/img/web-audit-demo.png)

The Dashboard page includes a Security Audit section with a quick-scan summary.

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | All skills passed (or only MEDIUM/HIGH findings) |
| `1` | One or more CRITICAL findings detected |

## Scanned Files

The audit scans text-based files in skill directories:

- `.md`, `.txt`, `.yaml`, `.yml`, `.json`, `.toml`
- `.sh`, `.bash`, `.zsh`, `.fish`
- `.py`, `.js`, `.ts`, `.rb`, `.go`, `.rs`
- Files without extensions (e.g., `Makefile`, `Dockerfile`)

Binary files (images, `.wasm`, etc.) and hidden directories (`.git`) are skipped.

## Options

| Flag | Description |
|------|------------|
| `-p`, `--project` | Scan project-level skills |
| `-g`, `--global` | Scan global skills |
| `-h`, `--help` | Show help |

## Related

- [install](/docs/commands/install) — Install skills (with automatic scanning)
- [doctor](/docs/commands/doctor) — Diagnose setup issues
- [list](/docs/commands/list) — List installed skills
