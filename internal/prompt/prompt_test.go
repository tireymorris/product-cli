package prompt

import (
	"strings"
	"testing"
)

func TestProductGenerationPromptStaysOutcomeFocused(t *testing.T) {
	text, err := ProductGeneration("share itineraries", "prd.json", "feature", []Answer{{Question: "Who shares?", Answer: "Travelers"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"share itineraries", "prd.json", `"mode": "product"`, "Travelers", "Do not describe implementation details", "then STOP"} {
		if !strings.Contains(text, want) {
			t.Errorf("prompt missing %q", want)
		}
	}
}
