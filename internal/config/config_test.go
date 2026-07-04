// Package config tests.
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/m0rtadelo/termia/internal/safety"
)

func TestDefault(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")
	cfg := Default()
	if cfg.DefaultModel != DefaultModelName {
		t.Errorf("default_model = %q, want %q", cfg.DefaultModel, DefaultModelName)
	}
	if cfg.Shell != "/bin/zsh" {
		t.Errorf("shell = %q, want %q", cfg.Shell, "/bin/zsh")
	}
	if !cfg.Safety.Safe || !cfg.Safety.Caution || !cfg.Safety.Danger {
		t.Errorf("safety defaults = %+v, want all true", cfg.Safety)
	}
}

func TestDefaultModelsNormalizesHostFromEnv(t *testing.T) {
	t.Setenv("OLLAMA_HOST", "localhost:11434")
	models := DefaultModels()
	if len(models) != 1 {
		t.Fatalf("models len = %d, want 1", len(models))
	}
	if models[0].Host != "http://localhost:11434" {
		t.Errorf("host = %q, want normalized with scheme", models[0].Host)
	}
}

func TestLoadMissingFileSeedsDefaults(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DefaultModel != DefaultModelName {
		t.Errorf("default_model = %q, want default", cfg.DefaultModel)
	}
	if _, err := os.Stat(filepath.Join(dir, "termia", "config.json")); err != nil {
		t.Fatalf("config file not seeded: %v", err)
	}
}

func TestSaveThenLoad(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	want := Config{
		DefaultModel: "cloud-qwen",
		Shell:        "/bin/bash",
		Safety:       SafetyConfig{Safe: false, Caution: true, Danger: true},
	}
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

func TestLoadModelsMissingFileSeedsDefaults(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	models, err := LoadModels()
	if err != nil {
		t.Fatalf("LoadModels() error = %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("models len = %d, want 1", len(models))
	}
	if models[0].Name != DefaultModelName {
		t.Errorf("name = %q, want %q", models[0].Name, DefaultModelName)
	}
	if _, err := os.Stat(filepath.Join(dir, "termia", "models.json")); err != nil {
		t.Fatalf("models file not seeded: %v", err)
	}
}

func TestResolveModelByNameAndModel(t *testing.T) {
	models := []ModelSpec{{Name: "llama-local", Model: "llama3.2:3b", Host: DefaultHost}}

	byName, err := ResolveModel(models, "llama-local")
	if err != nil {
		t.Fatalf("ResolveModel(name) error = %v", err)
	}
	if byName.Model != "llama3.2:3b" {
		t.Errorf("model = %q, want llama3.2:3b", byName.Model)
	}

	byID, err := ResolveModel(models, "llama3.2:3b")
	if err != nil {
		t.Fatalf("ResolveModel(model) error = %v", err)
	}
	if byID.Name != "llama-local" {
		t.Errorf("name = %q, want llama-local", byID.Name)
	}
}

func TestSafetyConfirm(t *testing.T) {
	s := SafetyConfig{Safe: false, Caution: true, Danger: false}
	if s.Confirm(safety.Safe) {
		t.Fatalf("safe should be false")
	}
	if !s.Confirm(safety.Caution) {
		t.Fatalf("caution should be true")
	}
	if s.Confirm(safety.Danger) {
		t.Fatalf("danger should be false")
	}
}
