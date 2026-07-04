// Package executor runs shell commands and streams their output.
package executor

import (
	"bufio"
	"context"
	"io"
	"os/exec"
)

// Event is a single line of output or the terminal result of a run.
type Event struct {
	Line     string // a line of combined stdout/stderr (without trailing newline)
	Done     bool   // true for the final event
	ExitCode int    // valid when Done is true
	Err      error  // non-nil if the command failed to start or run
}

// Run executes command via `shell -c command`, streaming combined output on the
// returned channel. The channel is closed after a final Event with Done=true.
func Run(ctx context.Context, shell, command string) <-chan Event {
	ch := make(chan Event)

	go func() {
		defer close(ch)

		cmd := exec.CommandContext(ctx, shell, "-c", command)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			ch <- Event{Done: true, ExitCode: -1, Err: err}
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			ch <- Event{Done: true, ExitCode: -1, Err: err}
			return
		}

		if err := cmd.Start(); err != nil {
			ch <- Event{Done: true, ExitCode: -1, Err: err}
			return
		}

		done := make(chan struct{}, 2)
		stream := func(r io.Reader) {
			scanner := bufio.NewScanner(r)
			scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
			for scanner.Scan() {
				ch <- Event{Line: scanner.Text()}
			}
			done <- struct{}{}
		}
		go stream(stdout)
		go stream(stderr)
		<-done
		<-done

		err = cmd.Wait()
		exit := 0
		if err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				exit = ee.ExitCode()
				err = nil
			} else {
				exit = -1
			}
		}
		ch <- Event{Done: true, ExitCode: exit, Err: err}
	}()

	return ch
}
