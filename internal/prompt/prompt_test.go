package prompt

import "testing"

func TestParseCleanJSON(t *testing.T) {
	raw := `{"command":"ls -la","explanation":"list files","dangerous":false}`
	got := Parse(raw)
	if got.Command != "ls -la" {
		t.Errorf("command = %q, want %q", got.Command, "ls -la")
	}
	if got.Explanation != "list files" {
		t.Errorf("explanation = %q, want %q", got.Explanation, "list files")
	}
	if got.Dangerous {
		t.Error("dangerous = true, want false")
	}
}

func TestParseJSONInCodeFence(t *testing.T) {
	raw := "```json\n{\"command\": \"rm -rf /tmp/x\", \"explanation\": \"remove\", \"dangerous\": true}\n```"
	got := Parse(raw)
	if got.Command != "rm -rf /tmp/x" {
		t.Errorf("command = %q, want %q", got.Command, "rm -rf /tmp/x")
	}
	if !got.Dangerous {
		t.Error("dangerous = false, want true")
	}
}

func TestParseTrimsCommandWhitespace(t *testing.T) {
	raw := `{"command":"  du -sh *  ","explanation":"disk usage","dangerous":false}`
	got := Parse(raw)
	if got.Command != "du -sh *" {
		t.Errorf("command = %q, want %q", got.Command, "du -sh *")
	}
}

func TestParseFallbackToExplanation(t *testing.T) {
	raw := "I cannot help with that."
	got := Parse(raw)
	if got.Command != "" {
		t.Errorf("command = %q, want empty", got.Command)
	}
	if got.Explanation != raw {
		t.Errorf("explanation = %q, want %q", got.Explanation, raw)
	}
}
