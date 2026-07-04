# TermIA

A terminal with IA that helps you by suggesting commands and providing responses to your terminal requests. This app lets you use natural language to execute complex terminal commands.

TermIA runs your natural-language request through a **local AI model via [Ollama](https://ollama.com)**, streams back a suggested shell command with a short explanation, and executes it **only after you confirm**.

Works on **macOS** and **Linux**.

> Status: planned / not yet implemented. See [PLAN.md](PLAN.md) for the full design and build plan.

## Features (planned)

- Natural-language → shell command suggestions, streamed live
- Interactive TUI plus a one-shot mode: `termia "show disk usage"`
- **Always confirm before running** — review, edit, or reject each command
- Danger detection (e.g. `rm -rf`, `dd`, `mkfs`, fork bombs) highlighted before you confirm
- 100% local inference via Ollama — no cloud API keys required

## Requirements

- **Go** 1.22+ (to build)
- **[Ollama](https://ollama.com)** installed and running locally

### Install Ollama and pull a model

```bash
# macOS
brew install ollama
# Linux
curl -fsSL https://ollama.com/install.sh | sh

# Pull a model (default), then make sure the server is running
ollama pull llama3.2:3b
ollama serve   # skip if Ollama already runs as a service
```

### Recommended models

The model is configurable (`--model` flag or config file); TermIA never auto-pulls.

| Model | Size (Q4) | RAM | Notes |
| --- | --- | --- | --- |
| `qwen2.5-coder:1.5b` | ~1 GB | ~2–3 GB | Tiny, code-focused; constrained hardware |
| **`llama3.2:3b`** | ~2 GB | ~4 GB | **Default** — good instruct + JSON output |
| `qwen2.5-coder:3b` | ~2 GB | ~4 GB | Best small option for command generation |
| `qwen2.5-coder:7b` | ~4.7 GB | ~8 GB | **Recommended** if RAM allows |
| `mistral:7b` | ~4.1 GB | ~8 GB | Reliable all-rounder |

## Usage (planned)

```bash
# Interactive TUI
termia

# One-shot
termia "list files larger than 100MB"

# Pick a model / host
termia --model qwen2.5-coder:7b --host http://localhost:11434
```

## Configuration

JSON at `~/.config/termia/config.json`:

```json
{
  "model": "llama3.2:3b",
  "ollama_host": "http://localhost:11434",
  "shell": "/bin/zsh"
}
```

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