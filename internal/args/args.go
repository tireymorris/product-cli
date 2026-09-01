package args

import (
	"fmt"
	"strings"
	"time"
)

type Options struct {
	Prompt   string
	Resume   bool
	Status   bool
	Clean    bool
	Help     bool
	Headless bool
	Yolo     bool
	Verbose  bool
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
		case "--resume":
			o.Resume = true
		case "--status", "status":
			o.Status = true
		case "--clean", "clean":
			o.Clean = true
		case "--headless":
			o.Headless = true
		case "--yolo":
			o.Yolo = true
		case "--verbose", "-v":
			o.Verbose = true
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
	if o.Resume && o.Prompt != "" {
		return nil, fmt.Errorf("--resume cannot be combined with a prompt")
	}
	if o.Status || o.Clean || o.Help {
		return o, nil
	}
	if !o.Resume && o.Prompt == "" && o.Headless {
		return nil, fmt.Errorf("--headless requires a prompt or --resume")
	}
	return o, nil
}

func HelpText() string {
	return `Product CLI - Product outcome planner

Usage:
  product-cli "describe the product outcome"
  product-cli --headless "describe the product outcome"
  product-cli --resume
  product-cli status
  product-cli clean

Options:
  --resume          Load the existing product document
  --headless        Do not prompt for a missing goal
  --yolo            Skip clarification questions
  --runner NAME     AI runner: claude, cursor, pi, opencode, or copilot
  --timeout DURATION  Stop a runner after a Go duration, such as 30m
  --verbose, -v     Show verbose runner output
  --help, -h        Show this help

Environment:
  PRODUCT_RUNNER          AI runner (default: claude)
  PRODUCT_RUNNER_TIMEOUT  Runner timeout
  PRODUCT_YOLO             Set to 1 to skip clarification

The product document is written to prd.json. Product CLI plans outcomes only;
it does not implement code.
`
}
