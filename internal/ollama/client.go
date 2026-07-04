package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Client talks to a local Ollama server.
type Client struct {
	Host   string
	Model  string
	apiKey string
	http   *http.Client
}

// New returns a Client for the given host and model.
func New(host, model, apiKey string) *Client {
	return &Client{
		Host:   strings.TrimRight(host, "/"),
		Model:  model,
		apiKey: apiKey,
		http:   &http.Client{Timeout: 5 * time.Minute},
	}
}

// Ping checks that the Ollama server is reachable.
func (c *Client) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.Host+"/api/version", nil)
	if err != nil {
		return err
	}
	c.applyAuth(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("Ollama not reachable at %s: %w", c.Host, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama returned status %d from %s", resp.StatusCode, c.Host)
	}
	return nil
}

// Chat streams a chat completion. The onToken callback receives incremental
// content as it arrives. The full accumulated content is returned when done.
// If format is non-nil it is passed as the Ollama structured-output schema.
func (c *Client) Chat(ctx context.Context, messages []Message, format json.RawMessage, onToken func(string)) (string, error) {
	payload := chatRequest{
		Model:    c.Model,
		Messages: messages,
		Stream:   true,
		Format:   format,
		Options:  map[string]any{"temperature": 0.2},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode chat request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Host+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	c.applyAuth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("Ollama not reachable at %s: %w", c.Host, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama returned status %d", resp.StatusCode)
	}

	var sb strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var chunk chatResponse
		if err := json.Unmarshal(line, &chunk); err != nil {
			return sb.String(), fmt.Errorf("decode stream chunk: %w", err)
		}
		if chunk.Error != "" {
			return sb.String(), fmt.Errorf("ollama error: %s", chunk.Error)
		}
		if chunk.Message.Content != "" {
			sb.WriteString(chunk.Message.Content)
			if onToken != nil {
				onToken(chunk.Message.Content)
			}
		}
		if chunk.Done {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return sb.String(), fmt.Errorf("read stream: %w", err)
	}
	return sb.String(), nil
}

func (c *Client) applyAuth(req *http.Request) {
	if c.apiKey == "" {
		return
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
}
