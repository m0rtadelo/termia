package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/m0rtadelo/termia/internal/executor"
	"github.com/m0rtadelo/termia/internal/safety"
)

const footerHeight = 3

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layout()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tokenMsg:
		m.stream.WriteString(string(msg))
		return m, waitForMsg(m.tokenCh)

	case thinkingDoneMsg:
		m.stream.Reset()
		if msg.err != nil {
			m.appendHistory(errorBlock(msg.err))
			m.state = stateInput
			m.input.Focus()
			return m, nil
		}
		m.suggestion = msg.suggestion
		if m.suggestion.Command == "" {
			m.appendHistory(assistantBlock(m.suggestion.Explanation))
			m.state = stateInput
			m.input.Focus()
			return m, nil
		}
		m.appendHistory(suggestionBlock(m.suggestion))
		if !m.cfg.Safety.Confirm(effectiveLevel(m.suggestion)) {
			m.state = stateExecuting
			m.appendHistory(runningBlock(m.suggestion.Command))
			m.execCh = executor.Run(context.Background(), m.cfg.Shell, m.suggestion.Command)
			return m, tea.Batch(m.spinner.Tick, waitForExec(m.execCh))
		}
		m.state = stateConfirm
		return m, nil

	case execEventMsg:
		if msg.Line != "" {
			m.appendHistory(outputLine(msg.Line))
		}
		if msg.Done {
			m.appendHistory(execResultBlock(msg.ExitCode, msg.Err))
			m.execCh = nil
			m.state = stateInput
			m.input.Focus()
			return m, nil
		}
		return m, waitForExec(m.execCh)
	}

	// Spinner animation while thinking or executing.
	if m.state == stateThinking || m.state == stateExecuting {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Text input updates while typing or editing.
	if m.state == stateInput || m.state == stateEditing {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.ready {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	}

	switch m.state {
	case stateInput:
		switch msg.Type {
		case tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			request := strings.TrimSpace(m.input.Value())
			if request == "" {
				return m, nil
			}
			if request == "exit" || request == "quit" {
				return m, tea.Quit
			}
			m.appendHistory(userBlock(request))
			m.input.Reset()
			m.input.Blur()
			m.state = stateThinking
			return m, m.startThinking(request)
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd

	case stateConfirm:
		key := strings.ToLower(msg.String())
		// SAFE commands default to Yes: Enter runs them.
		if msg.Type == tea.KeyEnter {
			if effectiveLevel(m.suggestion) == safety.Safe {
				key = "y"
			} else {
				return m, nil
			}
		}
		switch key {
		case "y":
			m.state = stateExecuting
			m.appendHistory(runningBlock(m.suggestion.Command))
			m.execCh = executor.Run(context.Background(), m.cfg.Shell, m.suggestion.Command)
			return m, tea.Batch(m.spinner.Tick, waitForExec(m.execCh))
		case "n", "esc":
			m.appendHistory(dimBlock("✗ discarded"))
			m.state = stateInput
			m.input.Focus()
			return m, nil
		case "e":
			m.state = stateEditing
			m.input.SetValue(m.suggestion.Command)
			m.input.Focus()
			m.input.CursorEnd()
			return m, textinput.Blink
		}
		return m, nil

	case stateEditing:
		switch msg.Type {
		case tea.KeyEnter:
			edited := strings.TrimSpace(m.input.Value())
			m.input.Reset()
			m.input.Blur()
			if edited == "" {
				m.appendHistory(dimBlock("✗ discarded"))
				m.state = stateInput
				m.input.Focus()
				return m, nil
			}
			m.suggestion.Command = edited
			m.suggestion.Dangerous = safety.Classify(edited) == safety.Danger
			m.appendHistory(suggestionBlock(m.suggestion))
			if !m.cfg.Safety.Confirm(effectiveLevel(m.suggestion)) {
				m.state = stateExecuting
				m.appendHistory(runningBlock(m.suggestion.Command))
				m.execCh = executor.Run(context.Background(), m.cfg.Shell, m.suggestion.Command)
				return m, tea.Batch(m.spinner.Tick, waitForExec(m.execCh))
			}
			m.state = stateConfirm
			return m, nil
		case tea.KeyEsc:
			m.input.Reset()
			m.input.Blur()
			m.state = stateConfirm
			return m, nil
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

// layout sizes the viewport based on the window dimensions.
func (m *Model) layout() {
	if !m.ready {
		m.viewport.Width = m.width
		m.viewport.Height = max(1, m.height-footerHeight)
		m.viewport.SetContent(m.history.String())
		m.ready = true
	} else {
		m.viewport.Width = m.width
		m.viewport.Height = max(1, m.height-footerHeight)
	}
	m.input.Width = max(10, m.width-4)
}

func execResultBlock(code int, err error) string {
	if err != nil {
		return errorBlock(err)
	}
	if code == 0 {
		return dimBlock("✓ exit 0")
	}
	return dimBlock(fmt.Sprintf("✗ exit %d", code))
}
