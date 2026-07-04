// Package config tests.
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	t.Setenv("OLLAMA_HOST", "")
	t.Setenv("SHELL", "/bin/zsh")
	cfg := Default()
	if cfg.Model != DefaultModel {
		t.Errorf("model = %q, want %q", cfg.Model, DefaultModel)
	}
	if cfg.OllamaHost != DefaultHost {
		t.Errorf("host = %q, want %q", cfg.OllamaHost, DefaultHost)
	}
	if cfg.Shell != "/bin/zsh" {
		t.Errorf("shell = %q, want %q", cfg.Shell, "/bin/zsh")
	}
}

func TestDefaultNormalizesHostFromEnv(t *testing.T) {
	t.Setenv("OLLAMA_HOST", "localhost:11434")
	cfg := Default()
	if cfg.OllamaHost != "http://localhost:11434" {
		t.Errorf("host = %q, want normalized with scheme", cfg.OllamaHost)
	}
}

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("OLLAMA_HOST", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Model != DefaultModel {
		t.Errorf("model = %q, want default", cfg.Model)
	}
}

func TestSaveThenLoad(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	want := Config{Model: "qwen2.5-coder:7b", OllamaHost: "http://localhost:1234", Shell: "/bin/bash"}
	if err := Save(want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "termia", "config.json")); err != nil {
		t.Fatalf("config file not written: %v", err)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got != want {
		t.Errorf("loaded = %+v, want %+v", got, want)
	}
}
