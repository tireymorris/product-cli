package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const DefaultRunner = "claude"
const DefaultProductFile = "prd.json"

var supportedRunners = map[string]bool{
	"claude":   true,
	"cursor":   true,
	"pi":       true,
	"opencode": true,
	"copilot":  true,
	"mock":     true,
}

type Config struct {
	WorkDir     string
	ProductFile string
	Runner      string
	Timeout     time.Duration
	Yolo        bool
	Headless    bool
	Verbose     bool
}

func Load() (*Config, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	cfg := &Config{
		WorkDir:     workDir,
		ProductFile: DefaultProductFile,
		Runner:      DefaultRunner,
	}
	if runner := os.Getenv("PRODUCT_RUNNER"); runner != "" {
		cfg.Runner = runner
	} else if runner := os.Getenv("RALPH_RUNNER"); runner != "" {
		cfg.Runner = runner
	}
	if os.Getenv("PRODUCT_YOLO") == "1" || os.Getenv("RALPH_YOLO") == "1" {
		cfg.Yolo = true
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
		return fmt.Errorf("unknown runner %q (supported: claude, cursor, pi, opencode, copilot, mock)", c.Runner)
	}
	if c.ProductFile == "" || filepath.Base(c.ProductFile) != c.ProductFile || strings.Contains(c.ProductFile, "..") {
		return fmt.Errorf("product file must be a simple filename, got %q", c.ProductFile)
	}
	return nil
}

func (c *Config) ProductPath() string { return filepath.Join(c.WorkDir, c.ProductFile) }
func (c *Config) StateDir() string    { return filepath.Join(c.WorkDir, ".product-cli") }
