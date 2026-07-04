package tui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/m0rtadelo/termia/internal/config"
	"github.com/m0rtadelo/termia/internal/ollama"
)

// TestUpdateDoesNotPanicOnBuilderCopy guards against the regression where the
// Model embedded strings.Builder by value; Bubble Tea copies the Model on every
// Update, which panics once the builder has been written to.
func TestUpdateDoesNotPanicOnBuilderCopy(t *testing.T) {
	m := New(config.Default(), ollama.New("http://localhost:1", "test"))

	var model tea.Model = m
	// Establish size (marks the model ready and sets the viewport).
	model, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	// Write to the history builder.
	model, _ = model.Update(thinkingDoneMsg{err: errors.New("boom")})
	// Force another copy of the now-non-zero builder. This panicked before the fix.
	model, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 25})

	// A trailing render must also not panic.
	_ = model.View()
}
