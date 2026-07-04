// Package config loads and persists TermIA's user configuration.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// DefaultModel is the Ollama model used when none is configured.
	DefaultModel = "llama3.2:3b"
	// DefaultHost is the default Ollama server address.
	DefaultHost = "http://localhost:11434"
)

// Config holds the user-configurable settings for TermIA.
type Config struct {
	Model      string `json:"model"`
	OllamaHost string `json:"ollama_host"`
	Shell      string `json:"shell"`
}

// Default returns a Config populated with sensible defaults derived from the
// environment.
func Default() Config {
	host := DefaultHost
	if h := os.Getenv("OLLAMA_HOST"); h != "" {
		host = normalizeHost(h)
	}
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	return Config{
		Model:      DefaultModel,
		OllamaHost: host,
		Shell:      shell,
	}
}

// Path returns the location of the config file, honoring XDG_CONFIG_HOME.
func Path() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home dir: %w", err)
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "termia", "config.json"), nil
}

// Load reads the config file, falling back to defaults for any missing values.
// A missing file is not an error: defaults are returned.
func Load() (Config, error) {
	cfg := Default()
	path, err := Path()
	if err != nil {
		return cfg, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config %s: %w", path, err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config %s: %w", path, err)
	}
	// Backfill any empty fields with defaults.
	def := Default()
	if cfg.Model == "" {
		cfg.Model = def.Model
	}
	if cfg.OllamaHost == "" {
		cfg.OllamaHost = def.OllamaHost
	} else {
		cfg.OllamaHost = normalizeHost(cfg.OllamaHost)
	}
	if cfg.Shell == "" {
		cfg.Shell = def.Shell
	}
	return cfg, nil
}

// Save writes the config to disk, creating parent directories as needed.
func Save(cfg Config) error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}
	return nil
}

// normalizeHost ensures the host has a scheme so it can be used as a base URL.
func normalizeHost(h string) string {
	if len(h) >= 7 && h[:7] == "http://" {
		return h
	}
	if len(h) >= 8 && h[:8] == "https://" {
		return h
	}
	return "http://" + h
}
