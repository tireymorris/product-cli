package runner

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"product-cli/internal/config"
)

type Runner struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Runner { return &Runner{cfg: cfg} }

func (r *Runner) Run(ctx context.Context, prompt string, output io.Writer) error {
	command, args, err := r.command(prompt)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = r.cfg.WorkDir
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Stdout = output
	cmd.Stderr = output
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("runner %s: %w", command, ctx.Err())
		}
		return fmt.Errorf("runner %s: %w", command, err)
	}
	return nil
}

func (r *Runner) command(prompt string) (string, []string, error) {
	switch r.cfg.Runner {
	case "claude":
		return "claude", []string{"--print", "--verbose", "--output-format", "stream-json", "--dangerously-skip-permissions"}, nil
	case "cursor":
		return "cursor-agent", []string{"--print", "--output-format", "stream-json", "--trust", "--yolo"}, nil
	case "pi":
		return "pi", []string{"--print", "--mode", "json", "--no-session"}, nil
	case "opencode":
		return "opencode", []string{"run", "--print-logs"}, nil
	case "copilot":
		return "copilot", []string{"--allow-all-tools", "--allow-all-paths", "--no-ask-user", "--output-format", "json", "--plan"}, nil
	case "mock":
		return "sh", []string{"-c", "printf '%s\\n' mock runner; cat >/dev/null"}, nil
	default:
		return "", nil, fmt.Errorf("unsupported runner %q", r.cfg.Runner)
	}
}
