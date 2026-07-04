// Package config loads and persists TermIA's user configuration.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/m0rtadelo/termia/internal/safety"
)

const (
	// DefaultModel is the default Ollama model ID used in seeded models.json.
	DefaultModel = "llama3.2:3b"
	// DefaultModelName is the default model entry name in config.json.
	DefaultModelName = "llama-local"
	// DefaultHost is the default Ollama server address.
	DefaultHost = "http://localhost:11434"
)

// Config holds the user-configurable settings for TermIA.
type Config struct {
	DefaultModel string       `json:"default_model"`
	Shell        string       `json:"shell"`
	Safety       SafetyConfig `json:"safety"`
	ContextTurns int          `json:"context_turns"`
	SystemPrompt string       `json:"system_prompt,omitempty"`
}

// SafetyConfig controls whether confirmation prompts are shown for each level.
type SafetyConfig struct {
	Safe    bool `json:"safe"`
	Caution bool `json:"caution"`
	Danger  bool `json:"danger"`
}

// ModelSpec defines one model target in models.json.
type ModelSpec struct {
	Name      string `json:"name"`
	Model     string `json:"model"`
	Host      string `json:"host"`
	APIKeyEnv string `json:"api_key_env,omitempty"`
}

// Default returns a Config populated with sensible defaults derived from the
// environment.
func Default() Config {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	return Config{
		DefaultModel: DefaultModelName,
		Shell:        shell,
		Safety:       defaultSafety(),
		ContextTurns: 3,
	}
}

func defaultSafety() SafetyConfig {
	return SafetyConfig{Safe: true, Caution: true, Danger: true}
}

// Confirm reports whether TermIA should ask for confirmation at the given risk
// level.
func (s SafetyConfig) Confirm(level safety.Level) bool {
	switch level {
	case safety.Danger:
		return s.Danger
	case safety.Caution:
		return s.Caution
	default:
		return s.Safe
	}
}

// DefaultModels returns a seeded list that works out of the box with local
// Ollama.
func DefaultModels() []ModelSpec {
	host := DefaultHost
	if h := os.Getenv("OLLAMA_HOST"); h != "" {
		host = NormalizeHost(h)
	}
	return []ModelSpec{{
		Name:  DefaultModelName,
		Model: DefaultModel,
		Host:  host,
	}}
}

func configDir() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home dir: %w", err)
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "termia"), nil
}

// Path returns the location of the config file, honoring XDG_CONFIG_HOME.
func Path() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// ModelsPath returns the location of the models list file.
func ModelsPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "models.json"), nil
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
			if err := Save(cfg); err != nil {
				return cfg, err
			}
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config %s: %w", path, err)
	}
	var raw struct {
		DefaultModel string        `json:"default_model"`
		Shell        string        `json:"shell"`
		Safety       *SafetyConfig `json:"safety"`
		ContextTurns *int          `json:"context_turns"`
		SystemPrompt string        `json:"system_prompt"`

		// Backward-compatible fields from older config versions.
		Model      string `json:"model"`
		OllamaHost string `json:"ollama_host"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return cfg, fmt.Errorf("parse config %s: %w", path, err)
	}
	// Backfill any empty fields with defaults.
	def := Default()
	if raw.DefaultModel != "" {
		cfg.DefaultModel = raw.DefaultModel
	} else if raw.Model != "" {
		cfg.DefaultModel = raw.Model
	} else {
		cfg.DefaultModel = def.DefaultModel
	}
	if raw.Shell == "" {
		cfg.Shell = def.Shell
	} else {
		cfg.Shell = raw.Shell
	}
	// Legacy host field no longer drives runtime host selection, but retain the
	// normalization path while parsing older config files.
	if raw.OllamaHost != "" {
		_ = NormalizeHost(raw.OllamaHost)
	}
	if raw.Safety == nil {
		cfg.Safety = def.Safety
	} else {
		cfg.Safety = *raw.Safety
	}
	if raw.ContextTurns == nil {
		cfg.ContextTurns = def.ContextTurns
	} else {
		cfg.ContextTurns = *raw.ContextTurns
	}
	cfg.SystemPrompt = raw.SystemPrompt
	return cfg, nil
}

// LoadModels reads models.json and validates/normalizes entries.
func LoadModels() ([]ModelSpec, error) {
	path, err := ModelsPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			models := DefaultModels()
			if err := SaveModels(models); err != nil {
				return nil, err
			}
			return models, nil
		}
		return nil, fmt.Errorf("read models %s: %w", path, err)
	}

	var models []ModelSpec
	if err := json.Unmarshal(data, &models); err != nil {
		return nil, fmt.Errorf("parse models %s: %w", path, err)
	}
	if len(models) == 0 {
		return nil, fmt.Errorf("models file %s has no entries", path)
	}

	for i := range models {
		if models[i].Model == "" {
			return nil, fmt.Errorf("models[%d] missing model", i)
		}
		if models[i].Name == "" {
			models[i].Name = models[i].Model
		}
		if models[i].Host == "" {
			models[i].Host = DefaultHost
		}
		models[i].Host = NormalizeHost(models[i].Host)
	}
	return models, nil
}

// ResolveModel finds a model entry by name first, then by model ID.
func ResolveModel(models []ModelSpec, selected string) (ModelSpec, error) {
	for _, m := range models {
		if m.Name == selected {
			return m, nil
		}
	}
	for _, m := range models {
		if m.Model == selected {
			return m, nil
		}
	}
	return ModelSpec{}, fmt.Errorf("unknown model %q", selected)
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

// SaveModels writes models.json, creating parent directories as needed.
func SaveModels(models []ModelSpec) error {
	path, err := ModelsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := json.MarshalIndent(models, "", "  ")
	if err != nil {
		return fmt.Errorf("encode models: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write models %s: %w", path, err)
	}
	return nil
}

// NormalizeHost ensures the host has a scheme so it can be used as a base URL.
func NormalizeHost(h string) string {
	if len(h) >= 7 && h[:7] == "http://" {
		return h
	}
	if len(h) >= 8 && h[:8] == "https://" {
		return h
	}
	return "http://" + h
}
