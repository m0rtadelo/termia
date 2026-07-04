# TermIA

A terminal with IA that helps you by suggesting commands and providing responses to your terminal requests. This app lets you use natural language to execute complex terminal commands.

TermIA runs your natural-language request through an **Ollama model** (local or cloud), streams back a suggested shell command with a short explanation, and executes it after confirmation unless that safety level is configured to auto-run.

Works on **macOS** and **Linux**.

> Status: implemented (active development).

## Features

- Natural-language → shell command suggestions, streamed live
- Interactive TUI plus a one-shot mode: `termia "show disk usage"`
- Per-level confirmation controls (`safe`, `caution`, `danger`) so you can require prompts only where you want
- Danger detection (e.g. `rm -rf`, `dd`, `mkfs`, fork bombs) highlighted before you confirm
- Ollama local and Ollama-hosted endpoints (optional API key from environment)

## Requirements

- **Go** 1.22+ (to build)
- **[Ollama](https://ollama.com)** endpoint available (local or cloud)

### Install Ollama and pull a model

```bash
# macOS
brew install ollama
# Linux
curl -fsSL https://ollama.com/install.sh | sh

# Pull a model (default local setup), then make sure the server is running
ollama pull llama3.2:3b
ollama serve   # skip if Ollama already runs as a service
```

### Recommended models

The model list is configurable in `~/.config/termia/models.json`; TermIA never auto-pulls.

| Model | Size (Q4) | RAM | Notes |
| --- | --- | --- | --- |
| `qwen2.5-coder:1.5b` | ~1 GB | ~2–3 GB | Tiny, code-focused; constrained hardware |
| **`llama3.2:3b`** | ~2 GB | ~4 GB | **Default** — good instruct + JSON output |
| `qwen2.5-coder:3b` | ~2 GB | ~4 GB | Best small option for command generation |
| `qwen2.5-coder:7b` | ~4.7 GB | ~8 GB | **Recommended** if RAM allows |
| `mistral:7b` | ~4.1 GB | ~8 GB | Reliable all-rounder |

## Usage

```bash
# Interactive TUI
termia

# One-shot
termia "list files larger than 100MB"

# Use a model entry from models.json
termia --model llama-local

# Override the selected entry host
termia --model llama-local --host http://localhost:11434
```

## Configuration

On first run, TermIA seeds both files under `~/.config/termia/`:
- `config.json`
- `models.json`

`config.json`:

```json
{
  "default_model": "llama-local",
  "shell": "/bin/zsh",
  "safety": {
    "safe": true,
    "caution": true,
    "danger": true
  },
  "context_turns": 3,
  "system_prompt": ""
}
```

`models.json`:

```json
[
  {
    "name": "llama-local",
    "model": "llama3.2:3b",
    "host": "http://localhost:11434"
  },
  {
    "name": "cloud-qwen",
    "model": "qwen2.5-coder:7b",
    "host": "https://ollama.com",
    "api_key_env": "OLLAMA_API_KEY"
  }
]
```

Notes:
- `--model` selects the model entry by `name`.
- `api_key_env` is optional. If set, TermIA reads the API key from that environment variable and sends it as `Authorization: Bearer ...`.
- Safety toggles control whether that level asks for confirmation.
  - `true`: ask for confirmation.
  - `false`: auto-run commands at that level.
- `context_turns` sets how many previous request/response pairs are included in each prompt so the model has conversation context. Default is `3`. Set to `0` for stateless (fresh each request).
- `system_prompt` lets you replace the built-in system prompt entirely. Leave empty (or omit) to use the default. When set, TermIA sends your string verbatim as the system message; it is your responsibility to instruct the model to return the expected JSON shape.

### Example: Custom system prompt

You can override the system prompt to customize behavior. For example, to encourage more defensive commands:

```json
{
  "system_prompt": "You are TermIA, a cautious assistant that converts natural-language requests into shell commands.\n\nEnvironment:\n- Operating system: {OS}\n- Shell: {SHELL}\n- Working directory: {CWD}\n\nRules:\n- Reply ONLY with JSON: {\"command\": string, \"explanation\": string, \"dangerous\": boolean}.\n- Prefer commands that are safe and non-destructive.\n- If unsure, err on the side of caution and set \"dangerous\" to true.\n- Never generate commands that could cause data loss without explicit confirmation.\n- Include comments in multiline commands for clarity.\n- Never include commentary outside the JSON object."
}
```

Note: `{OS}`, `{SHELL}`, and `{CWD}` are filled in at request time with your actual OS, shell, and current working directory.

## Development

```bash
go build ./...
go vet ./...
go test ./...
```

### Git identity (contributors)

This repository uses a **repo-local** git identity. Configure it before committing:

```bash
git config user.email "ricard.figuls@gmail.com"
git config user.name "Ricard Fíguls Mateu"
```

Commits must **not** use an `@capgemini.com` email. Verify with:

```bash
git config user.email
git log --format='%ae'
```