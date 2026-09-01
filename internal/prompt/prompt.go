package prompt

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed templates/product-generate.tmpl
var templates embed.FS

func ProductGeneration(goal string) (string, error) {
	tmpl, err := template.ParseFS(templates, "templates/product-generate.tmpl")
	if err != nil {
		return "", fmt.Errorf("parse product prompt: %w", err)
	}
	var out bytes.Buffer
	if err := tmpl.ExecuteTemplate(&out, "product-generate", goal); err != nil {
		return "", fmt.Errorf("render product prompt: %w", err)
	}
	return out.String(), nil
}
