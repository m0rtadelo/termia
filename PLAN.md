# Plan: TermIA — Current State and Next Steps

TermIA is implemented as a Go TUI + CLI assistant that converts natural language into shell commands with Ollama, classifies command risk, and executes with confirmation controls.

## Current status
- Interactive TUI and one-shot mode are implemented.
- Ollama streaming client is implemented via `/api/chat`.
- Config and model catalog are implemented under `~/.config/termia/`.
- Per-level safety prompt toggles are implemented in both one-shot and TUI paths.

## Implemented behavior

### Configuration files
- `~/.config/termia/config.json`
	- `default_model`: model entry name from `models.json`
	- `shell`: shell path used by executor
	- `safety`: `{ "safe": bool, "caution": bool, "danger": bool }`
- `~/.config/termia/models.json`
	- Array of model entries: `{ "name", "model", "host", "api_key_env"? }`

Both files are seeded on first run when missing.

### Model selection and host
- `--model` selects by entry `name` in `models.json` (with model-id fallback for compatibility).
- `--host` overrides the selected model host.
- Hosts are normalized to include scheme.

### Ollama local/cloud support
- TermIA remains Ollama-only.
- Cloud-style usage is supported by setting a remote host and optional `api_key_env`.
- When `api_key_env` is set and present in environment, requests include `Authorization: Bearer <token>`.

### Safety prompt policy
- Risk levels: safe, caution, danger.
- If a level toggle is `true`, TermIA prompts for confirmation.
- If a level toggle is `false`, TermIA auto-runs commands at that level.
- This policy is applied consistently in both one-shot and TUI flows.

## Architecture summary
- `main.go`: loads config/models, resolves selected model, builds Ollama client, runs one-shot or TUI.
- `internal/config`: config/models paths, load/save, seeding, model resolution.
- `internal/ollama`: streaming client + optional bearer auth.
- `internal/prompt`: prompt contract and parser for structured model output.
- `internal/safety`: command classification.
- `internal/executor`: shell command execution + streamed output events.
- `internal/tui`: state machine, streaming UX, confirm/edit/discard flow.

## Verification snapshot
- `go test ./...` passes.
- `go build ./...` passes.

## Next steps roadmap
1. Add explicit integration tests for `models.json` + `config.json` first-run seeding behavior.
2. Improve startup error guidance for cloud auth misconfiguration (missing env var when `api_key_env` is set).
3. Add a small `termia doctor` command to validate config files, reachable host, and model entry.
4. Add optional command history persistence for TUI sessions.

## Scope notes
- Included now: TUI, one-shot, streaming, risk classification, per-level safety toggles, models catalog, Ollama local/cloud-style auth.
- Out of scope currently: non-Ollama providers, Windows support, multi-turn memory.
