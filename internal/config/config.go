package config

import (
	"fmt"
	"os"
	"time"
)

const DefaultRunner = "claude"

var supportedRunners = map[string]bool{
	"claude":   true,
	"cursor":   true,
	"pi":       true,
	"opencode": true,
	"copilot":  true,
}

type Config struct {
	WorkDir string
	Runner  string
	Timeout time.Duration
}

func Load() (*Config, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}
	cfg := &Config{WorkDir: workDir, Runner: DefaultRunner}
	if runner := os.Getenv("PRODUCT_RUNNER"); runner != "" {
		cfg.Runner = runner
	}
	if raw := os.Getenv("PRODUCT_RUNNER_TIMEOUT"); raw != "" {
		cfg.Timeout, err = time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("PRODUCT_RUNNER_TIMEOUT must be a Go duration: %w", err)
		}
	}
	return cfg, cfg.Validate()
}

func (c *Config) Validate() error {
	if !supportedRunners[c.Runner] {
		return fmt.Errorf("unknown runner %q (supported: claude, cursor, pi, opencode, copilot)", c.Runner)
	}
	return nil
}
