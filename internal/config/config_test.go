package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	content := `
defaults:
  username: root
  password: secret
  timeout: 2s
mysql:
  - name: primary
    host: 127.0.0.1
    port: 3306
  - name: secondary
    host: 127.0.0.2
    port: 3307
    username: root2
kafka:
  - name: events
    brokers: [127.0.0.1:9092]
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.MySQL) != 2 || len(cfg.Kafka) != 1 {
		t.Fatalf("unexpected target counts: mysql=%d kafka=%d", len(cfg.MySQL), len(cfg.Kafka))
	}
	if cfg.MySQL[0].Username != "root" || cfg.MySQL[0].Password != "secret" || cfg.MySQL[0].Timeout != 2*time.Second {
		t.Fatalf("defaults not applied: %+v", cfg.MySQL[0])
	}
	if cfg.MySQL[1].Username != "root2" || cfg.MySQL[1].Password != "secret" {
		t.Fatalf("target values not preserved: %+v", cfg.MySQL[1])
	}
	if cfg.Kafka[0].Brokers[0] != "127.0.0.1:9092" {
		t.Fatalf("unexpected kafka brokers: %+v", cfg.Kafka[0])
	}
}

func TestLoadRequiresTargetFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("mysql:\n  - name: broken\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "missing.yaml"))
	if err == nil {
		t.Fatal("expected missing-file error")
	}
}
