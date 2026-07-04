package tui

import (
	"context"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/m0rtadelo/termia/internal/config"
	"github.com/m0rtadelo/termia/internal/executor"
	"github.com/m0rtadelo/termia/internal/ollama"
	"github.com/m0rtadelo/termia/internal/prompt"
)

type state int

const (
	stateInput state = iota
	stateThinking
	stateConfirm
	stateEditing
	stateExecuting
)

// Model is the root Bubble Tea model for TermIA.
type Model struct {
	cfg    config.Config
	client *ollama.Client

	input    textinput.Model
	viewport viewport.Model
	spinner  spinner.Model

	state   state
	history *strings.Builder // rendered transcript
	stream  *strings.Builder // in-progress model output

	suggestion prompt.Suggestion
	execCh     <-chan executor.Event
	tokenCh    chan tea.Msg

	turns      []ollama.Message // user+assistant pairs from previous requests
	pendingReq string           // user request for the current in-flight chat

	ready  bool
	width  int
	height int
	err    error
}

// New builds the initial model.
func New(cfg config.Config, client *ollama.Client) Model {
	ti := textinput.New()
	ti.Placeholder = "Describe what you want to do…"
	ti.Prompt = "❯ "
	ti.Focus()
	ti.CharLimit = 0

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return Model{
		cfg:     cfg,
		client:  client,
		input:   ti,
		spinner: sp,
		state:   stateInput,
		history: &strings.Builder{},
		stream:  &strings.Builder{},
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// appendHistory writes a block to the transcript and refreshes the viewport.
func (m *Model) appendHistory(block string) {
	m.history.WriteString(block)
	if !strings.HasSuffix(block, "\n") {
		m.history.WriteString("\n")
	}
	if m.ready {
		m.viewport.SetContent(m.history.String())
		m.viewport.GotoBottom()
	}
}

// startThinking launches a streamed chat request and returns commands that feed
// tokens back into the update loop.
func (m *Model) startThinking(request string) tea.Cmd {
	tokens := make(chan tea.Msg, 64)
	m.tokenCh = tokens
	m.pendingReq = request

	go func() {
		cwd, _ := os.Getwd()
		sysPrompt := prompt.System(m.cfg.Shell, cwd, m.cfg.SystemPrompt)

		messages := make([]ollama.Message, 0, 1+len(m.turns)+1)
		messages = append(messages, ollama.Message{Role: "system", Content: sysPrompt})
		if m.cfg.ContextTurns > 0 {
			maxMsgs := m.cfg.ContextTurns * 2
			history := m.turns
			if len(history) > maxMsgs {
				history = history[len(history)-maxMsgs:]
			}
			messages = append(messages, history...)
		}
		messages = append(messages, ollama.Message{Role: "user", Content: request})

		raw, err := m.client.Chat(context.Background(), messages, prompt.ResponseSchema, func(t string) {
			tokens <- tokenMsg(t)
		})
		if err != nil {
			tokens <- thinkingDoneMsg{err: err}
		} else {
			tokens <- thinkingDoneMsg{suggestion: prompt.Parse(raw), rawResponse: raw}
		}
		close(tokens)
	}()

	return tea.Batch(m.spinner.Tick, waitForMsg(tokens))
}

// waitForMsg returns a command that blocks until the next message arrives on ch.
func waitForMsg(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg { return <-ch }
}

// waitForExec returns a command that reads the next executor event.
func waitForExec(ch <-chan executor.Event) tea.Cmd {
	return func() tea.Msg { return execEventMsg(<-ch) }
}
