package tui

import (
	"github.com/m0rtadelo/termia/internal/executor"
	"github.com/m0rtadelo/termia/internal/prompt"
)

// tokenMsg carries an incremental chunk of streamed model output.
type tokenMsg string

// thinkingDoneMsg is emitted when the model has finished responding.
type thinkingDoneMsg struct {
	suggestion prompt.Suggestion
	err        error
}

// execEventMsg wraps a single line/result from a running command.
type execEventMsg executor.Event

// errMsg reports a fatal error to display.
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }
