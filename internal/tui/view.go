package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/m0rtadelo/termia/internal/prompt"
	"github.com/m0rtadelo/termia/internal/safety"
)

var (
	userStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	assistant    = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	outStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	cmdStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("48")).Bold(true)
	badgeSafe    = lipgloss.NewStyle().Foreground(lipgloss.Color("235")).Background(lipgloss.Color("42")).Padding(0, 1)
	badgeCaution = lipgloss.NewStyle().Foreground(lipgloss.Color("235")).Background(lipgloss.Color("214")).Padding(0, 1)
	badgeDanger  = lipgloss.NewStyle().Foreground(lipgloss.Color("231")).Background(lipgloss.Color("196")).Padding(0, 1)
	footerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

// View implements tea.Model.
func (m Model) View() string {
	if !m.ready {
		return "Starting TermIA…"
	}

	var footer string
	switch m.state {
	case stateThinking:
		footer = m.spinner.View() + " thinking…"
	case stateExecuting:
		footer = m.spinner.View() + " running…"
	case stateConfirm:
		footer = footerStyle.Render("[y] run   [e] edit   [n] discard")
	case stateEditing:
		footer = m.input.View() + "\n" + footerStyle.Render("editing — enter to accept, esc to cancel")
	default:
		footer = m.input.View() + "\n" + footerStyle.Render("enter to ask · ctrl+c to quit")
	}

	// While streaming, show the live partial output under the transcript.
	body := m.viewport.View()
	if m.state == stateThinking && m.stream.Len() > 0 {
		body += "\n" + dimStyle.Render("…"+m.stream.String())
	}

	return body + "\n" + footer
}

func badgeFor(level safety.Level) string {
	switch level {
	case safety.Danger:
		return badgeDanger.Render(" DANGER ")
	case safety.Caution:
		return badgeCaution.Render(" CAUTION ")
	default:
		return badgeSafe.Render(" SAFE ")
	}
}

func userBlock(s string) string {
	return userStyle.Render("❯ "+s) + "\n"
}

func assistantBlock(s string) string {
	return assistant.Render(s) + "\n"
}

func suggestionBlock(s prompt.Suggestion) string {
	level := safety.Classify(s.Command)
	if s.Dangerous && level < safety.Danger {
		level = safety.Caution
	}
	var b strings.Builder
	b.WriteString(badgeFor(level) + " " + cmdStyle.Render(s.Command) + "\n")
	if s.Explanation != "" {
		b.WriteString(assistant.Render(s.Explanation) + "\n")
	}
	return b.String()
}

func runningBlock(cmd string) string {
	return dimStyle.Render("$ "+cmd) + "\n"
}

func outputLine(line string) string {
	return outStyle.Render(line)
}

func dimBlock(s string) string {
	return dimStyle.Render(s) + "\n"
}

func errorBlock(err error) string {
	return errStyle.Render("error: "+err.Error()) + "\n"
}
