package args

import (
	"fmt"
	"strings"
	"time"
)

type Options struct {
	Prompt   string
	Help     bool
	Headless bool
	Runner   string
	Timeout  time.Duration
}

func Parse(argv []string) (*Options, error) {
	o := &Options{}
	var prompt []string
	for i := 0; i < len(argv); i++ {
		switch arg := argv[i]; arg {
		case "--help", "-h":
			o.Help = true
		case "--headless":
			o.Headless = true
		case "--runner":
			if i+1 >= len(argv) {
				return nil, fmt.Errorf("--runner requires a value")
			}
			i++
			o.Runner = argv[i]
		case "--timeout":
			if i+1 >= len(argv) {
				return nil, fmt.Errorf("--timeout requires a value")
			}
			i++
			timeout, err := time.ParseDuration(argv[i])
			if err != nil {
				return nil, fmt.Errorf("invalid --timeout: %w", err)
			}
			o.Timeout = timeout
		default:
			if strings.HasPrefix(arg, "-") {
				return nil, fmt.Errorf("unknown flag %q", arg)
			}
			prompt = append(prompt, arg)
		}
	}
	o.Prompt = strings.Join(prompt, " ")
	if o.Headless && o.Prompt == "" {
		return nil, fmt.Errorf("--headless requires a product goal")
	}
	return o, nil
}

func HelpText() string {
	return `Product CLI - Product outcome planner

Usage:
  product-cli "describe the product outcome"
  product-cli --headless "describe the product outcome"

Options:
  --headless        Do not prompt for a missing goal
  --runner NAME     AI runner: claude, cursor, pi, opencode, or copilot
  --timeout DURATION  Stop a runner after a Go duration, such as 30m
  --help, -h        Show this help

Environment:
  PRODUCT_RUNNER          AI runner (default: claude)
  PRODUCT_RUNNER_TIMEOUT  Runner timeout

Product CLI prints an outcome-focused Markdown plan to stdout. It does not
write a product file or implement code.
`
}
