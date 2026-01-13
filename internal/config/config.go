package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type CheckType string

const (
	CheckTypeHTTP CheckType = "http"
	CheckTypeTCP  CheckType = "tcp"
	CheckTypeDNS  CheckType = "dns"
)

type Target struct {
	Name           string            `yaml:"name"`
	URL            string            `yaml:"url"`
	Type           CheckType         `yaml:"type"`
	Timeout        time.Duration     `yaml:"timeout"`
	ExpectedStatus int               `yaml:"expected_status,omitempty"`
	Headers        map[string]string `yaml:"headers,omitempty"`
	Interval       time.Duration     `yaml:"interval,omitempty"`
}

type Config struct {
	Targets     []Target      `yaml:"targets"`
	Concurrency int           `yaml:"concurrency"`
	Timeout     time.Duration `yaml:"timeout"`
}

func DefaultConfig() *Config {
	return &Config{
		Concurrency: 10,
		Timeout:     5 * time.Second,
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	for i := range cfg.Targets {
		if cfg.Targets[i].Timeout == 0 {
			cfg.Targets[i].Timeout = cfg.Timeout
		}
		if cfg.Targets[i].Type == "" {
			cfg.Targets[i].Type = CheckTypeHTTP
		}
		if cfg.Targets[i].ExpectedStatus == 0 && cfg.Targets[i].Type == CheckTypeHTTP {
			cfg.Targets[i].ExpectedStatus = 200
		}
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if len(c.Targets) == 0 {
		return fmt.Errorf("no targets defined")
	}
	for i, t := range c.Targets {
		if t.URL == "" {
			return fmt.Errorf("target %d: url is required", i)
		}
		if t.Name == "" {
			c.Targets[i].Name = t.URL
		}
	}
	return nil
}
