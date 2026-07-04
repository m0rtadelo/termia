# Plan: TermIA — AI Terminal Assistant (Go + Ollama TUI)

Build **TermIA**, a Go terminal UI where you type natural-language requests, a **local Ollama model** suggests a shell command + explanation (streamed live), and it runs **only after you confirm**. Cross-platform (macOS + Linux). Interactive TUI is the main experience; a one-shot `termia "request"` mode is a bonus.

## Confirmed decisions
- **Form:** CLI tool / TUI in the terminal
- **Language:** Go
- **AI backend:** local model via Ollama (`http://localhost:11434`)
- **Execution:** always confirm before running

## Tech stack
- **TUI:** `charmbracelet/bubbletea` + `bubbles` (textinput, viewport, spinner) + `lipgloss` (styling)
- **Ollama:** direct REST call to `/api/chat` with `"stream": true` via `net/http` + `encoding/json` (keeps deps light vs. importing the whole `ollama` module, which is heavy and carries advisories)
- **Execution:** `os/exec` running `$SHELL -c "<cmd>"` (fallback `/bin/sh`), streaming stdout/stderr
- **Config:** JSON at `~/.config/termia/config.json` (XDG). Fields: `model`, `ollama_host`, `shell`
- **Build:** Go 1.22+, single static binary. Makefile for build/run/test/lint

## Ollama model (prerequisite)
The app does **not** auto-pull; the user pulls a model once. Model is configurable via config and the `--model` flag.

| Model | Size (Q4) | RAM | Notes |
|---|---|---|---|
| `qwen2.5-coder:1.5b` | ~1 GB | ~2–3 GB | Tiny, code-focused; for constrained hardware |
| **`llama3.2:3b`** | ~2 GB | ~4 GB | **Default** — good instruct + JSON adherence |
| `qwen2.5-coder:3b` | ~2 GB | ~4 GB | Best small option for command generation |
| `qwen2.5-coder:7b` | ~4.7 GB | ~8 GB | **Recommended** if RAM allows — best quality/size |
| `mistral:7b` | ~4.1 GB | ~8 GB | Reliable all-rounder |

Small models can break the JSON schema → use Ollama structured output (a JSON schema in `format`, not just `format:"json"`) plus a parse fallback.

## Git configuration (required)
Set **repo-local** git identity (not `--global`) so no work identity leaks into this project:
- `git config user.email "ricard.figuls@gmail.com"`
- `git config user.name "Ricard Fíguls Mateu"`

**Hard requirement:** no commit may use an `@capgemini.com` email. Verify with `git config user.email` and `git log --format='%ae'`.

## Architecture / layout
```
termia/
├── go.mod / go.sum
├── main.go                  # entry: flags, load config, one-shot vs TUI
├── Makefile
├── internal/
│   ├── config/config.go     # Load/Save, defaults, XDG path
│   ├── ollama/client.go     # Client{Host,Model}; Chat(ctx, msgs, onToken) streaming
│   ├── ollama/types.go      # chatRequest, chatResponse, Message
│   ├── prompt/prompt.go     # system prompt; parse output -> Command + Explanation + Dangerous
│   ├── executor/executor.go # Run(ctx, shell, cmd) streaming output; exit code
│   ├── safety/safety.go     # DangerLevel + regex (rm -rf, dd, mkfs, fork bomb)
│   └── tui/
│       ├── model.go         # bubbletea Model + state machine
│       ├── update.go        # key handling, async cmds
│       ├── view.go          # render history, input, confirm panel
│       └── messages.go      # tea.Msg types (token, done, execOut, err)
├── PLAN.md
└── README.md
```

## Model prompt contract
System prompt instructs the model to reply ONLY as JSON:
`{"command": "...", "explanation": "...", "dangerous": bool}`
Include OS (`runtime.GOOS`) + shell so commands fit the platform. Fallback: if JSON parse fails, treat the whole response as explanation with an empty command.

## TUI state machine
`Input → Thinking (stream tokens) → Confirm` where Confirm handles:
- `y` → `Executing` (stream output) → `Input`
- `e` → edit command in textinput → `Confirm`
- `n` → discard → `Input`

Global: `Ctrl+C` / `q` quits. Scrollable history viewport.

## Steps

### Phase 1 — Scaffold & config (foundation)
0. `git init` (if needed), then set repo-local git identity above; confirm `git config user.email` = `ricard.figuls@gmail.com` (never `@capgemini.com`).
1. `go mod init github.com/m0rtadelo/termia`; add bubbletea, bubbles, lipgloss deps.
2. `internal/config`: struct + `Load()`/`Save()` + defaults (model `llama3.2:3b`, host from `OLLAMA_HOST` or `localhost:11434`, shell from `$SHELL`).
3. `main.go`: parse flags (`--model`, `--host`, one-shot positional args), load config.

### Phase 2 — Ollama client (parallel with Phase 3)
4. `internal/ollama`: types + streaming `Chat()` that POSTs `/api/chat`, decodes the JSON-per-line stream, invokes `onToken`, returns final content + error. Add a heartbeat to detect Ollama not running → friendly error.

### Phase 3 — Prompt + safety + executor (parallel with Phase 2)
5. `internal/prompt`: build system prompt, `ParseResponse()` → `{Command, Explanation, Dangerous}`.
6. `internal/safety`: regex list → `DangerLevel` (safe/caution/danger); colors the confirm panel.
7. `internal/executor`: `Run()` with `exec.CommandContext($SHELL, "-c", cmd)`, stream combined output via channel, capture exit code.

### Phase 4 — TUI wiring (depends on 1–7)
8. `tui/model`+`messages`: define Model, states, `tea.Msg` types.
9. `tui/update`: Enter submits, stream tokens as `tokenMsg`, `y/n/e` in confirm, run executor as `tea.Cmd` emitting `execOutMsg`, errors as `errMsg`.
10. `tui/view`: lipgloss-styled layout — history viewport, spinner while thinking, command box + explanation + danger badge, live output while executing.
11. `main.go`: launch `tea.NewProgram(model)` for interactive; one-shot path prints the suggested command and asks `y/N` inline.

### Phase 5 — Polish
12. Makefile (build/run/test/lint), unit tests for `prompt.ParseResponse` and safety classification. Update README with install + usage + Ollama prerequisite.

## Verification
1. `go build ./...` and `go vet ./...` succeed.
2. `go test ./...` passes (prompt parsing, safety classification unit tests).
3. Git identity correct: `git config user.email` = `ricard.figuls@gmail.com`, name = `Ricard Fíguls Mateu`; `git log --format='%ae'` shows **no** `@capgemini.com` addresses.
4. Manual: with Ollama running + a pulled model, run `termia`, type "list files larger than 100MB" → command streams, confirm, runs.
5. Manual: one-shot `termia "show disk usage"` → prints command + `y/N`.
6. Manual: Ollama stopped → clear "Ollama not reachable" error, no crash.
7. Manual: destructive request → danger badge + required confirm.

## Scope
- **Included:** interactive TUI, one-shot mode, streaming, confirm-before-run, danger detection, config file, macOS + Linux.
- **Excluded (v1):** multi-turn memory, history persistence, multiple AI providers, auto-run of "safe" commands, Windows, piping output back into the model.

## Open considerations
1. **Module path:** assumed `github.com/m0rtadelo/termia` (confirm if different).
2. **Default model:** `llama3.2:3b` (small/fast, good JSON). Alt: `qwen2.5-coder:7b`, or `qwen2.5-coder:3b`/`1.5b` on tighter hardware. Configurable, no auto-pull.
3. **Multi-turn memory:** recommend excluding for v1 (each request independent) to keep it simple.
