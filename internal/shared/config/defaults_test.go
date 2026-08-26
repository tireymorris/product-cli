package config

import (
	"os"
	"testing"
)

func TestDefaultConfigBranchPrefix(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.BranchPrefix != DefaultBranchPrefix {
		t.Fatalf("BranchPrefix = %q, want %q", cfg.BranchPrefix, DefaultBranchPrefix)
	}
}

func TestDefaultConfigTestCommandEmpty(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.TestCommand != "" {
		t.Fatalf("TestCommand = %q, want empty default", cfg.TestCommand)
	}
}

func TestLoadEnvBranchPrefix(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	os.Clearenv()
	os.Setenv("RALPH_BRANCH_PREFIX", "feat")
	defer os.Unsetenv("RALPH_BRANCH_PREFIX")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.BranchPrefix != "feat" {
		t.Fatalf("BranchPrefix = %q, want feat", cfg.BranchPrefix)
	}
}

func TestValidateAllowsEmptyTestCommand(t *testing.T) {
	cfg := DefaultConfig()
	cfg.TestCommand = ""
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil for empty test command", err)
	}
}
