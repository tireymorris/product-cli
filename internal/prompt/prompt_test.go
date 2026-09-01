package prompt

import (
	"strings"
	"testing"
)

func TestProductGenerationPromptRequestsFormattedOutcomeText(t *testing.T) {
	text, err := ProductGeneration("share itineraries")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"share itineraries", "Markdown", "user-facing outcomes", "Do not describe implementation details"} {
		if !strings.Contains(text, want) {
			t.Errorf("prompt missing %q", want)
		}
	}
	for _, unwanted := range []string{"prd.json", "product.json", `"mode": "product"`} {
		if strings.Contains(text, unwanted) {
			t.Errorf("prompt should not mention %q", unwanted)
		}
	}
}
