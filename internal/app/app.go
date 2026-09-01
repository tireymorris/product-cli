package app

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"product-cli/internal/args"
	"product-cli/internal/config"
	"product-cli/internal/prompt"
	"product-cli/internal/runner"
	"product-cli/internal/version"
)

func Run(argv []string) int {
	opts, err := args.Parse(argv)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		fmt.Print(args.HelpText())
		return 1
	}
	if opts.Help {
		fmt.Print(args.HelpText())
		return 0
	}
	if opts.Version {
		fmt.Println(version.Info())
		return 0
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading configuration:", err)
		return 1
	}
	if opts.Runner != "" {
		cfg.Runner = opts.Runner
	}
	if opts.Timeout > 0 {
		cfg.Timeout = opts.Timeout
	}
	if err := cfg.Validate(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}

	goal := strings.TrimSpace(opts.Prompt)
	if goal == "" {
		if opts.Headless {
			fmt.Fprintln(os.Stderr, "Error: --headless requires a product goal")
			return 1
		}
		fmt.Fprint(os.Stderr, "Product goal: ")
		goal, err = bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil && err != io.EOF {
			fmt.Fprintln(os.Stderr, "Error reading product goal:", err)
			return 1
		}
		goal = strings.TrimSpace(goal)
	}
	if goal == "" {
		fmt.Fprintln(os.Stderr, "Error: product goal cannot be empty")
		return 1
	}

	return generate(cfg, goal, os.Stdout)
}

func generate(cfg *config.Config, goal string, stdout io.Writer) int {
	text, err := prompt.ProductGeneration(goal)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error preparing product prompt:", err)
		return 1
	}

	var output bytes.Buffer
	ctx := context.Background()
	var cancel context.CancelFunc
	if cfg.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}
	if err := runner.New(cfg).Run(ctx, text, &output); err != nil {
		fmt.Fprintln(os.Stderr, "Error generating product plan:", err)
		return 1
	}

	formatted := formatOutput(output.String())
	if formatted == "" {
		fmt.Fprintln(os.Stderr, "Error generating product plan: runner returned no plan")
		return 1
	}
	_, _ = fmt.Fprintln(stdout, formatted)
	return 0
}

func formatOutput(raw string) string {
	text := strings.TrimSpace(raw)
	if strings.HasPrefix(text, "```") && strings.HasSuffix(text, "```") {
		lines := strings.Split(text, "\n")
		if len(lines) >= 2 {
			text = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}
	return strings.TrimSpace(text)
}
