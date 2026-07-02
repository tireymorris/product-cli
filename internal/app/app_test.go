package app

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"ralph/internal/args"
	"ralph/internal/shared/config"
	"ralph/internal/version"
)

func TestRunVersion(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	codeCh := make(chan int, 1)
	go func() {
		codeCh <- Run([]string{"version"})
		w.Close()
	}()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	if code := <-codeCh; code != 0 {
		t.Errorf("Run(version) = %d, want 0", code)
	}
	want := version.Info() + "\n"
	if got := buf.String(); got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

func TestRunClean(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	prdPath := filepath.Join(tmpDir, "prd.json")
	if err := os.WriteFile(prdPath, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("removes prd.json", func(t *testing.T) {
		if code := Run([]string{"clean"}); code != 0 {
			t.Fatalf("Run(clean) = %d, want 0", code)
		}
		if _, err := os.Stat(prdPath); !os.IsNotExist(err) {
			t.Fatalf("prd.json still exists after clean: %v", err)
		}
	})

	t.Run("skips ValidateResume with --resume", func(t *testing.T) {
		if code := Run([]string{"clean", "--resume"}); code != 0 {
			t.Fatalf("Run(clean --resume) = %d, want 0 (ValidateResume must not run)", code)
		}
	})
}

func TestApplyRuntimeOptionsSetsAutoApprove(t *testing.T) {
	cfg := config.DefaultConfig()
	opts := &args.Options{AutoApprove: true}

	applyRuntimeOptions(cfg, opts)

	if !cfg.AutoApprove {
		t.Error("AutoApprove should be copied from parsed options")
	}
}

func TestApplyRuntimeOptionsSetsDryRun(t *testing.T) {
	cfg := config.DefaultConfig()
	opts := &args.Options{DryRun: true}

	applyRuntimeOptions(cfg, opts)

	if !cfg.DryRun {
		t.Error("DryRun should be copied from parsed options")
	}
}

func TestApplyRuntimeOptionsSetsProductMode(t *testing.T) {
	cfg := config.DefaultConfig()
	opts := &args.Options{Product: true}

	applyRuntimeOptions(cfg, opts)

	if !cfg.DryRun {
		t.Error("DryRun should be enabled for product mode")
	}
	if cfg.PRDFile != "product.json" {
		t.Fatalf("PRDFile = %q, want %q", cfg.PRDFile, "product.json")
	}
}

func TestValidateResumeProductDocumentEnablesDryRun(t *testing.T) {
	dir := t.TempDir()
	productPath := filepath.Join(dir, "product.json")
	productData := `{"project_name":"Product","stories":[{"id":"1","title":"S1","description":"d","slices":[{"id":"slice-1","behavior":"users can sign in"}],"priority":1}]}`
	if err := os.WriteFile(productPath, []byte(productData), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = dir
	cfg.PRDFile = "prd.json"

	stderr := captureStderr(t, func() {
		if err := ValidateResume(cfg, true); err != nil {
			t.Fatalf("ValidateResume() error = %v", err)
		}
	})
	if cfg.PRDFile != "product.json" {
		t.Fatalf("PRDFile = %q, want product.json", cfg.PRDFile)
	}
	if !cfg.DryRun {
		t.Fatal("DryRun should be enabled when resuming product.json")
	}
	if !strings.Contains(stderr, "prd.json not found") {
		t.Fatalf("stderr = %q, want fallback warning", stderr)
	}
}

func TestValidateResume_warnsWhenSiblingDocumentExists(t *testing.T) {
	dir := t.TempDir()
	prdData := `{"project_name":"PRD","stories":[{"id":"1","title":"S1","description":"d","slices":[{"id":"slice-1","behavior":"a","red_hint":"add failing test"}],"priority":1}]}`
	productData := `{"project_name":"Product","stories":[{"id":"1","title":"S1","description":"d","slices":[{"id":"slice-1","behavior":"users can sign in"}],"priority":1}]}`
	if err := os.WriteFile(filepath.Join(dir, "prd.json"), []byte(prdData), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "product.json"), []byte(productData), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.WorkDir = dir
	cfg.PRDFile = "prd.json"

	stderr := captureStderr(t, func() {
		if err := ValidateResume(cfg, true); err != nil {
			t.Fatalf("ValidateResume() error = %v", err)
		}
	})
	if !strings.Contains(stderr, "product.json also exists and will be ignored") {
		t.Fatalf("stderr = %q, want sibling ignored warning", stderr)
	}
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()
	fn()
	w.Close()
	os.Stderr = oldStderr
	return <-done
}

func TestRunHeadlessProductWithYolo(t *testing.T) {
	t.Setenv("RALPH_RUNNER", "mock")
	t.Setenv("RALPH_YOLO", "1")

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	initGitRepoForAppTest(t, tmpDir)
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	commitFileForAppTest(t, tmpDir, "main.go", "init")

	if code := Run([]string{"--headless", "--product", "define widget"}); code != 0 {
		t.Fatalf("Run() = %d, want 0", code)
	}
	raw, err := os.ReadFile(filepath.Join(tmpDir, "product.json"))
	if err != nil {
		t.Fatalf("product.json missing: %v", err)
	}
	if strings.Contains(string(raw), "red_hint") {
		t.Fatalf("product.json should not contain implementation fields:\n%s", raw)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "prd.json")); err == nil {
		t.Fatal("prd.json should not be created in product mode")
	}
}

func initGitRepoForAppTest(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "t@example.com"},
		{"config", "user.name", "t"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

func commitFileForAppTest(t *testing.T, dir, file, message string) {
	t.Helper()
	for _, args := range [][]string{
		{"add", file},
		{"commit", "-m", message},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

func TestRunBareNoTTY(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() {
		os.Stdin = oldStdin
		r.Close()
		w.Close()
	}()
	w.Close()

	if code := Run([]string{}); code != 1 {
		t.Errorf("Run() with no args and non-TTY stdin = %d, want 1", code)
	}
}
