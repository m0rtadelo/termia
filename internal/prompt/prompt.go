// Package prompt builds the system prompt sent to the model and parses its
// structured response into a command suggestion.
package prompt

import (
	"encoding/json"
	"fmt"
	"regexp"
	"runtime"
	"strings"
)

// Suggestion is the parsed result of a model response.
type Suggestion struct {
	Command     string `json:"command"`
	Explanation string `json:"explanation"`
	Dangerous   bool   `json:"dangerous"`
}

// ResponseSchema is the JSON schema passed to Ollama's structured-output
// `format` parameter to keep small models on-contract.
var ResponseSchema = json.RawMessage(`{
  "type": "object",
  "properties": {
    "command": {"type": "string"},
    "explanation": {"type": "string"},
    "dangerous": {"type": "boolean"}
  },
  "required": ["command", "explanation", "dangerous"]
}`)

// System returns the system prompt tailored to the current OS, shell, and
// working directory. If custom is non-empty it is returned as-is, giving the
// user full control over the prompt via config.json.
func System(shell, cwd, custom string) string {
	if custom != "" {
		return custom
	}
	return fmt.Sprintf(`You are TermIA, an assistant that converts a user's natural-language request into a single shell command.

Environment:
- Operating system: %s
- Shell: %s
- Working directory: %s

Rules:
- Reply ONLY with a JSON object: {"command": string, "explanation": string, "dangerous": boolean}.
- "command" must be a single shell command line valid for the shell and OS above. Do not wrap it in markdown or backticks.
- "explanation" is a short (one or two sentences) description of what the command does.
- "dangerous" is true if the command deletes data, overwrites files, changes permissions broadly, or is otherwise hard to undo.
- If the request cannot be satisfied with a command, set "command" to an empty string and explain why in "explanation".
- Never include commentary outside the JSON object.`, runtime.GOOS, shell, cwd)
}

// jsonObjectRe matches the first {...} block in a string as a fallback for
// models that wrap JSON in prose or code fences.
var jsonObjectRe = regexp.MustCompile(`(?s)\{.*\}`)

// Parse converts a raw model response into a Suggestion. If the response is not
// valid JSON, the whole text is treated as an explanation with no command.
func Parse(raw string) Suggestion {
	raw = strings.TrimSpace(raw)

	var s Suggestion
	if err := json.Unmarshal([]byte(raw), &s); err == nil {
		s.Command = strings.TrimSpace(s.Command)
		return s
	}

	// Fallback: extract the first JSON object embedded in the text.
	if match := jsonObjectRe.FindString(raw); match != "" {
		if err := json.Unmarshal([]byte(match), &s); err == nil {
			s.Command = strings.TrimSpace(s.Command)
			return s
		}
	}

	// Give up on structure; surface the raw text as the explanation.
	return Suggestion{Explanation: raw}
}
