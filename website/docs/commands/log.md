---
sidebar_position: 4
---

# log

View the persistent operation log for debugging and compliance.

```bash
skillshare log                  # Show recent operations
skillshare log --audit          # Show audit log
skillshare log --tail 50        # Show last 50 entries
skillshare log --clear          # Clear operations log
skillshare log -p               # Show project operations log
```

## What Gets Logged

Every mutating CLI and Web UI operation is recorded as a JSONL entry with timestamp, command, status, duration, and contextual args.

| Command | Log File |
|---------|----------|
| `install`, `uninstall`, `sync`, `push`, `pull`, `collect`, `backup`, `restore`, `update` | `operations.log` |
| `audit` | `audit.log` |

## Log Types

### Operations Log (default)

Records install, uninstall, sync, push, pull, collect, backup, restore, and update operations.

```bash
skillshare log
```

### Audit Log

Records security audit scans separately from normal operations.

```bash
skillshare log --audit
```

## Example Output

```
┌─ Operations Log (last 5) ───────────────┐

  2026-02-10 14:30  INSTALL  anthropics/skills/pdf        ok       1.2s
  2026-02-10 14:31  SYNC     some targets failed to sync  error    0.8s
  2026-02-10 14:35  SYNC                                  ok       0.3s
  2026-02-10 15:00  UPDATE   _team-skills                 ok       2.1s
  2026-02-10 15:01  PUSH     Add new skill                ok       0.5s
```

## Log Format

Entries are stored in JSONL format (one JSON object per line):

```json
{"ts":"2026-02-10T14:30:00Z","cmd":"install","args":{"source":"anthropics/skills/pdf"},"status":"ok","ms":1200}
```

| Field | Description |
|-------|-------------|
| `ts` | ISO 8601 timestamp |
| `cmd` | Command name |
| `args` | Command-specific context (source, name, target, etc.) |
| `status` | `ok`, `error`, `partial`, or `blocked` |
| `msg` | Error message (when status is not ok) |
| `ms` | Duration in milliseconds |

## Log Location

```
~/.config/skillshare/logs/operations.log    # Global operations
~/.config/skillshare/logs/audit.log         # Global audit
<project>/.skillshare/logs/operations.log   # Project operations
<project>/.skillshare/logs/audit.log        # Project audit
```

## Options

| Flag | Description |
|------|------------|
| `-a`, `--audit` | Show audit log instead of operations log |
| `-t`, `--tail <N>` | Show last N entries (default: 20) |
| `-c`, `--clear` | Clear the log file |
| `-p`, `--project` | Use project-level log |
| `-g`, `--global` | Use global log |
| `-h`, `--help` | Show help |

## Web UI

The log is also available in the web dashboard at `/log`:

```bash
skillshare ui
# Navigate to Log page
```

The Log page provides:
- **Tabs** to switch between Operations and Audit logs
- **Table view** with time, command, details, status, and duration
- **Clear** and **Refresh** controls

## Related

- [audit](/docs/commands/audit) — Security scanning (logged to audit.log)
- [status](/docs/commands/status) — Show current sync state
- [doctor](/docs/commands/doctor) — Diagnose setup issues
