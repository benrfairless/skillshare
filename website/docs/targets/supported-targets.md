---
sidebar_position: 2
---

# Supported Targets

Complete list of AI CLIs that skillshare supports out of the box.

## Overview

Skillshare supports **45+ AI CLI tools**. When you run `skillshare init`, it automatically detects and configures any installed tools.

---

## Built-in Targets

These are auto-detected during `skillshare init`:

```text
┌────────────────────────────────────────────────────────────────────────────┐
│                              SUPPORTED TARGETS                             │
│                                                                            │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────────┐ ┌─────────┐           │
│  │   AdaL  │ │  Agents │ │   Amp   │ │ Antigravity │ │ Augment │           │
│  └─────────┘ └─────────┘ └─────────┘ └─────────────┘ └─────────┘           │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌───────────┐ ┌─────────┐             │
│  │   Bob   │ │  Claude │ │  Cline  │ │ CodeBuddy │ │  COMATE │             │
│  └─────────┘ └─────────┘ └─────────┘ └───────────┘ └─────────┘             │
│  ┌─────────┐ ┌──────────┐ ┌──────────┐ ┌─────────┐ ┌─────────┐             │
│  │  Codex  │ │ Cmd Code │ │ Continue │ │ Copilot │ │  Cortex │             │
│  └─────────┘ └──────────┘ └──────────┘ └─────────┘ └─────────┘             │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐   │
│  │  Crush  │ │  Cursor │ │  Droid  │ │  Gemini │ │  Goose  │ │  iFlow  │   │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘   │
│  ┌─────────┐ ┌──────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐              │
│  │  Junie  │ │ Kilocode │ │   Kimi  │ │   Kiro  │ │   Kode  │              │
│  └─────────┘ └──────────┘ └─────────┘ └─────────┘ └─────────┘              │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌────────────┐            │
│  │  Letta  │ │  MCPJam │ │   Mux   │ │ Neovate │ │ OpenClaude │            │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘ └────────────┘            │
│  ┌──────────┐ ┌──────────┐ ┌───────────┐ ┌─────────┐ ┌─────────┐           │
│  │ OpenClaw │ │ OpenCode │ │ OpenHands │ │    Pi   │ │  Pochi  │           │
│  └──────────┘ └──────────┘ └───────────┘ └─────────┘ └─────────┘           │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐   │
│  │  Qoder  │ │   Qwen  │ │   Roo   │ │   Trae  │ │ Trae CN │ │   Vibe  │   │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘   │
│  ┌──────────┐ ┌──────────────┐ ┌─────────────┐ ┌──────────┐                │
│  │ Windsurf │ │ Xcode Claude │ │ Xcode Codex │ │ Zencoder │                │
│  └──────────┘ └──────────────┘ └─────────────┘ └──────────┘                │
└────────────────────────────────────────────────────────────────────────────┘
```

---

## Target Paths

| Target | Global Path | Project Path |
|--------|-------------|--------------|
| adal | `~/.adal/skills` | `.adal/skills` |
| agents | `~/.config/agents/skills` | `.agents/skills` |
| amp | `~/.config/agents/skills` | `.agents/skills` |
| antigravity | `~/.gemini/antigravity/global_skills` | `.agent/skills` |
| augment | `~/.augment/rules` | `.augment/rules` |
| bob | `~/.bob/skills` | `.bob/skills` |
| claude | `~/.claude/skills` | `.claude/skills` |
| cline | `~/.cline/skills` | `.cline/skills` |
| codebuddy | `~/.codebuddy/skills` | `.codebuddy/skills` |
| comate | `~/.comate/skills` | `.comate/skills` |
| codex | `~/.codex/skills` | `.agents/skills` |
| commandcode | `~/.commandcode/skills` | `.commandcode/skills` |
| continue | `~/.continue/skills` | `.continue/skills` |
| cortex | `~/.snowflake/cortex/skills` | `.cortex/skills` |
| copilot | `~/.copilot/skills` | `.agents/skills` |
| crush | `~/.config/crush/skills` | `.crush/skills` |
| cursor | `~/.cursor/skills` | `.cursor/skills` |
| droid | `~/.agents/skills` | `.agents/skills` |
| gemini | `~/.gemini/skills` | `.agents/skills` |
| goose | `~/.config/goose/skills` | `.agents/skills` |
| iflow | `~/.iflow/skills` | `.iflow/skills` |
| junie | `~/.junie/skills` | `.junie/skills` |
| kilocode | `~/.kilocode/skills` | `.kilocode/skills` |
| kimi | `~/.kimi/skills` | `.agents/skills` |
| kiro | `~/.kiro/skills` | `.kiro/skills` |
| kode | `~/.kode/skills` | `.kode/skills` |
| letta | `~/.letta/skills` | `.skills` |
| mcpjam | `~/.mcpjam/skills` | `.mcpjam/skills` |
| mux | `~/.mux/skills` | `.mux/skills` |
| neovate | `~/.neovate/skills` | `.neovate/skills` |
| openclaude | `~/.openclaude/skills` | — |
| openclaw | `~/.openclaw/skills` | `skills` |
| opencode | `~/.config/opencode/skills` | `.agents/skills` |
| openhands | `~/.openhands/skills` | `.openhands/skills` |
| pi | `~/.pi/agent/skills` | `.pi/skills` |
| pochi | `~/.pochi/skills` | `.pochi/skills` |
| qoder | `~/.qoder/skills` | `.qoder/skills` |
| qwen | `~/.qwen/skills` | `.qwen/skills` |
| replit | — | `.agents/skills` |
| roo | `~/.roo/skills` | `.roo/skills` |
| trae | `~/.trae/skills` | `.trae/skills` |
| trae-cn | `~/.trae-cn/skills` | `.trae/skills` |
| vibe | `~/.vibe/skills` | `.vibe/skills` |
| windsurf | `~/.codeium/windsurf/skills` | `.windsurf/skills` |
| xcode-claude | `~/Library/Developer/Xcode/CodingAssistant/ClaudeAgentConfig/skills` | `.claude/skills` |
| xcode-codex | `~/Library/Developer/Xcode/CodingAssistant/codex/skills` | `.codex/skills` |
| zencoder | `~/.zencoder/skills` | `.zencoder/skills` |

---

## Check Target Path

For any target, run:

```bash
skillshare target <name>
```

---

## Custom Targets

Don't see your AI CLI? Add it manually:

```bash
skillshare target add myapp ~/.myapp/skills
```

See [Adding Custom Targets](./adding-custom-targets) for details.

---

## Related

- [Adding Custom Targets](./adding-custom-targets) — Add unsupported tools
- [Configuration](./configuration) — Config file reference
- [Commands: target](/docs/commands/target) — Target command
