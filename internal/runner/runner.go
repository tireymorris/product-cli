package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"product-cli/internal/config"
)

type Runner struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Runner { return &Runner{cfg: cfg} }

func (r *Runner) Run(ctx context.Context, prompt string, output io.Writer) error {
	command, args, err := r.command()
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = r.cfg.WorkDir
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Stdout = output
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("runner %s: %w", command, ctx.Err())
		}
		return fmt.Errorf("runner %s: %w", command, err)
	}
	return nil
}

func (r *Runner) command() (string, []string, error) {
	switch r.cfg.Runner {
	case "claude":
		return "claude", []string{"--print", "--dangerously-skip-permissions"}, nil
	case "cursor":
		return "cursor-agent", []string{"--print", "--trust", "--yolo"}, nil
	case "pi":
		return "pi", []string{"--print", "--no-session"}, nil
	case "opencode":
		return "opencode", []string{"run"}, nil
	case "copilot":
		return "copilot", []string{"--allow-all-tools", "--allow-all-paths", "--no-ask-user", "--output-format", "text", "--plan"}, nil
	default:
		return "", nil, fmt.Errorf("unsupported runner %q", r.cfg.Runner)
	}
}
