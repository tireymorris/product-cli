package app

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"product-cli/internal/args"
	"product-cli/internal/config"
	"product-cli/internal/product"
	"product-cli/internal/prompt"
	"product-cli/internal/runner"
)

const branchPrefix = "feature"

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
	cfg.Yolo = cfg.Yolo || opts.Yolo
	cfg.Headless = opts.Headless
	cfg.Verbose = opts.Verbose
	if err := cfg.Validate(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}

	switch {
	case opts.Status:
		return status(cfg)
	case opts.Clean:
		return clean(cfg)
	case opts.Resume:
		return resume(cfg)
	default:
		goal := strings.TrimSpace(opts.Prompt)
		if goal == "" {
			if cfg.Headless {
				fmt.Fprintln(os.Stderr, "Error: --headless requires a product goal")
				return 1
			}
			fmt.Print("Product goal: ")
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
		return generate(cfg, goal)
	}
}

func generate(cfg *config.Config, goal string) int {
	r := runner.New(cfg)
	answers := []prompt.Answer{}
	if !cfg.Yolo {
		questionsPath := filepath.Join(cfg.StateDir(), "questions.json")
		_ = os.Remove(questionsPath)
		clarify := fmt.Sprintf("You are a product planning assistant. Review this goal: %s\nIf clarification is needed, write a JSON array of question strings to %s. If no clarification is needed, do not create the file. Do not implement anything.", goal, questionsPath)
		if err := runRunner(cfg, r, clarify); err != nil {
			fmt.Fprintln(os.Stderr, "Warning: clarification failed; continuing without it:", err)
		}
		questions := readQuestions(questionsPath)
		for _, question := range questions {
			fmt.Printf("%s\nAnswer: ", question)
			answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil && err != io.EOF {
				fmt.Fprintln(os.Stderr, "Error reading clarification:", err)
				return 1
			}
			answers = append(answers, prompt.Answer{Question: question, Answer: strings.TrimSpace(answer)})
		}
		_ = os.Remove(questionsPath)
	}

	generation, err := prompt.ProductGeneration(goal, cfg.ProductFile, branchPrefix, answers)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error preparing product prompt:", err)
		return 1
	}
	if err := runRunner(cfg, r, generation); err != nil {
		fmt.Fprintln(os.Stderr, "Error generating product document:", err)
		return 1
	}
	doc, err := product.Load(cfg.ProductPath())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading generated product document:", err)
		return 1
	}
	if err := product.Save(cfg.ProductPath(), doc); err != nil {
		fmt.Fprintln(os.Stderr, "Error saving generated product document:", err)
		return 1
	}
	fmt.Printf("Product document saved to %s (%d stories)\n", cfg.ProductFile, len(doc.Stories))
	return 0
}

func runRunner(cfg *config.Config, r *runner.Runner, text string) error {
	ctx := context.Background()
	var cancel context.CancelFunc
	if cfg.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}
	return r.Run(ctx, text, os.Stdout)
}

func readQuestions(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var questions []string
	if json.Unmarshal(data, &questions) != nil {
		return nil
	}
	return questions
}

func resume(cfg *config.Config) int {
	migrated, err := product.MigrateLegacy(cfg.WorkDir, cfg.ProductPath())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error migrating legacy product:", err)
		return 1
	}
	if migrated {
		fmt.Printf("Migrated %s to %s\n", product.LegacyFilename, cfg.ProductFile)
	}
	doc, err := product.Load(cfg.ProductPath())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error loading product document:", err)
		return 1
	}
	printDocument(doc)
	return 0
}

func status(cfg *config.Config) int {
	doc, err := product.Load(cfg.ProductPath())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	completed, total := doc.Progress()
	fmt.Printf("%s: %d/%d outcomes marked complete\n", doc.ProjectName, completed, total)
	return 0
}

func clean(cfg *config.Config) int {
	for _, path := range []string{cfg.ProductPath(), cfg.ProductPath() + ".lock", filepath.Join(cfg.WorkDir, product.LegacyFilename), filepath.Join(cfg.WorkDir, product.LegacyFilename+".lock")} {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "Error removing", path+":", err)
			return 1
		}
	}
	if err := os.RemoveAll(cfg.StateDir()); err != nil {
		fmt.Fprintln(os.Stderr, "Error removing state:", err)
		return 1
	}
	fmt.Println("Product CLI state removed.")
	return 0
}

func printDocument(doc *product.Document) {
	fmt.Printf("%s\n", doc.ProjectName)
	for _, story := range doc.Stories {
		mark := " "
		if story.Passes {
			mark = "✓"
		}
		fmt.Printf("[%s] %s\n", mark, story.Title)
		for _, slice := range story.Slices {
			fmt.Printf("  - %s\n", slice.Behavior)
		}
	}
}
