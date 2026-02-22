package configx

import (
	"os"
	"path/filepath"
	"testing"
)

type sampleConfig struct {
	Mode  string `json:"mode"`
	Count string `json:"count"`
}

func TestLoadPrecedence(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "config.json")
	if err := os.WriteFile(file, []byte(`{"mode":"file","count":"10"}`), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	got, err := Load(Options{
		Defaults: map[string]any{
			"mode":  "default",
			"count": "1",
		},
		FilePath: file,
		Env: map[string]string{
			"mode": "env",
		},
		Flags: map[string]string{
			"count": "99",
		},
	})
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if got["mode"] != "env" {
		t.Fatalf("mode precedence mismatch: %v", got["mode"])
	}
	if got["count"] != "99" {
		t.Fatalf("count precedence mismatch: %v", got["count"])
	}
}

func TestDecode(t *testing.T) {
	raw := map[string]any{
		"mode":  "safe",
		"count": "7",
	}
	cfg, err := Decode[sampleConfig](raw)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if cfg.Mode != "safe" || cfg.Count != "7" {
		t.Fatalf("decoded config mismatch: %+v", cfg)
	}
}

func TestNormalizeEnv(t *testing.T) {
	got := NormalizeEnv("APP_", []string{
		"APP_MODE=prod",
		"APP_LOG_LEVEL=debug",
		"OTHER=x",
	})
	if got["mode"] != "prod" {
		t.Fatalf("mode mismatch: %q", got["mode"])
	}
	if got["log_level"] != "debug" {
		t.Fatalf("log_level mismatch: %q", got["log_level"])
	}
	if _, ok := got["other"]; ok {
		t.Fatal("unexpected non-prefixed variable")
	}
}
