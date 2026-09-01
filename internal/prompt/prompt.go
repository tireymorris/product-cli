package prompt

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed templates/product-generate.tmpl
var templates embed.FS

type ProductData struct {
	Goal         string
	ProductFile  string
	BranchPrefix string
	Answers      []Answer
}

type Answer struct {
	Question string
	Answer   string
}

func ProductGeneration(goal, productFile, branchPrefix string, answers []Answer) (string, error) {
	tmpl, err := template.ParseFS(templates, "templates/product-generate.tmpl")
	if err != nil {
		return "", fmt.Errorf("parse product prompt: %w", err)
	}
	var out bytes.Buffer
	if err := tmpl.ExecuteTemplate(&out, "product-generate", ProductData{
		Goal: goal, ProductFile: productFile, BranchPrefix: branchPrefix, Answers: answers,
	}); err != nil {
		return "", fmt.Errorf("render product prompt: %w", err)
	}
	return out.String(), nil
}
