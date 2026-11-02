package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Defaults   Target   `yaml:"defaults"`
	MySQL      []Target `yaml:"mysql"`
	PostgreSQL []Target `yaml:"postgresql"`
	Redis      []Target `yaml:"redis"`
	MongoDB    []Target `yaml:"mongodb"`
	Kafka      []Target `yaml:"kafka"`
}

type Target struct {
	Name     string        `yaml:"name"`
	Host     string        `yaml:"host"`
	Port     int           `yaml:"port"`
	Username string        `yaml:"username"`
	Password string        `yaml:"password"`
	Database string        `yaml:"database"`
	Timeout  time.Duration `yaml:"timeout"`

	Brokers    []string          `yaml:"brokers"`
	Addresses  []string          `yaml:"addresses"`
	TLS        bool              `yaml:"tls"`
	Attributes map[string]string `yaml:"attributes"`
	SASL       SASLConfig        `yaml:"sasl"`
}

type SASLConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Mechanism string `yaml:"mechanism"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
}

func Load(path string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return cfg, fmt.Errorf("config file not found: %w", err)
		}
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse YAML %s: %w", path, err)
	}
	cfg.applyDefaults()
	if err := cfg.validate(); err != nil {
		return cfg, fmt.Errorf("validate config: %w", err)
	}
	return cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Defaults.Timeout <= 0 {
		c.Defaults.Timeout = 5 * time.Second
	}
	c.applySectionDefaults(c.MySQL, "mysql")
	c.applySectionDefaults(c.PostgreSQL, "postgresql")
	c.applySectionDefaults(c.Redis, "redis")
	c.applySectionDefaults(c.MongoDB, "mongodb")
	c.applySectionDefaults(c.Kafka, "kafka")
}

func (c *Config) applySectionDefaults(targets []Target, kind string) {
	for index := range targets {
		target := &targets[index]
		if target.Name == "" {
			target.Name = fmt.Sprintf("%s-%d", kind, index+1)
		}
		if target.Host == "" && len(c.Defaults.Host) > 0 {
			target.Host = c.Defaults.Host
		}
		if target.Port == 0 && c.Defaults.Port != 0 {
			target.Port = c.Defaults.Port
		}
		if target.Username == "" {
			target.Username = c.Defaults.Username
		}
		if target.Password == "" {
			target.Password = c.Defaults.Password
		}
		if target.Database == "" {
			target.Database = c.Defaults.Database
		}
		if target.Timeout <= 0 {
			target.Timeout = c.Defaults.Timeout
		}
	}
}

func (c *Config) validate() error {
	for kind, targets := range map[string][]Target{
		"mysql":      c.MySQL,
		"postgresql": c.PostgreSQL,
		"redis":      c.Redis,
		"mongodb":    c.MongoDB,
		"kafka":      c.Kafka,
	} {
		for index, target := range targets {
			if err := target.validate(kind); err != nil {
				return fmt.Errorf("%s[%d]: %w", kind, index, err)
			}
		}
	}
	return nil
}

func (t Target) validate(kind string) error {
	if t.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	if kind == "kafka" {
		if len(t.Brokers) == 0 {
			return fmt.Errorf("brokers is required")
		}
		return nil
	}
	if kind == "mongodb" && len(t.Addresses) > 0 {
		for index, address := range t.Addresses {
			if address == "" {
				return fmt.Errorf("addresses[%d] is required", index)
			}
		}
		return nil
	}
	if t.Host == "" {
		return fmt.Errorf("host is required")
	}
	if t.Port <= 0 || t.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	return nil
}
