// Command termia is an AI-assisted terminal: describe a task in natural language
// and it suggests a shell command via a local Ollama model, running it only
// after you confirm.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/m0rtadelo/termia/internal/config"
	"github.com/m0rtadelo/termia/internal/executor"
	"github.com/m0rtadelo/termia/internal/ollama"
	"github.com/m0rtadelo/termia/internal/prompt"
	"github.com/m0rtadelo/termia/internal/safety"
	"github.com/m0rtadelo/termia/internal/tui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "warning:", err)
	}

	model := flag.String("model", cfg.Model, "Ollama model to use")
	host := flag.String("host", cfg.OllamaHost, "Ollama host URL")
	flag.Parse()

	cfg.Model = *model
	cfg.OllamaHost = *host

	client := ollama.New(cfg.OllamaHost, cfg.Model)
	if err := client.Ping(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "Is Ollama running? Try: ollama serve")
		os.Exit(1)
	}

	// One-shot mode: `termia "do something"`.
	if args := flag.Args(); len(args) > 0 {
		os.Exit(runOnce(cfg, client, strings.Join(args, " ")))
	}

	// Interactive TUI mode.
	p := tea.NewProgram(tui.New(cfg, client), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// runOnce handles the non-interactive path: suggest a command, ask y/N, run it.
func runOnce(cfg config.Config, client *ollama.Client, request string) int {
	ctx := context.Background()
	messages := []ollama.Message{
		{Role: "system", Content: prompt.System(cfg.Shell)},
		{Role: "user", Content: request},
	}

	raw, err := client.Chat(ctx, messages, prompt.ResponseSchema, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	s := prompt.Parse(raw)
	if s.Command == "" {
		fmt.Println(s.Explanation)
		return 0
	}

	level := safety.Classify(s.Command)
	if s.Dangerous && level < safety.Danger {
		level = safety.Caution
	}

	fmt.Printf("[%s] %s\n", level, s.Command)
	if s.Explanation != "" {
		fmt.Println(s.Explanation)
	}
	fmt.Print("Run this command? [y/N] ")

	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(answer)) != "y" {
		fmt.Println("discarded")
		return 0
	}

	exit := 0
	for ev := range executor.Run(ctx, cfg.Shell, s.Command) {
		if ev.Line != "" {
			fmt.Println(ev.Line)
		}
		if ev.Done {
			if ev.Err != nil {
				fmt.Fprintln(os.Stderr, "error:", ev.Err)
				exit = 1
			} else {
				exit = ev.ExitCode
			}
		}
	}
	return exit
}
