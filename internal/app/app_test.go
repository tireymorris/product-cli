package app

import "testing"

func TestFormatOutputTrimsMarkdownCodeFence(t *testing.T) {
	got := formatOutput("```markdown\n# Product plan\n\n## Why\nIt matters.\n```")
	want := "# Product plan\n\n## Why\nIt matters."
	if got != want {
		t.Fatalf("formatOutput() = %q, want %q", got, want)
	}
}

func TestFormatOutputPreservesMarkdown(t *testing.T) {
	want := "# Product plan\n\n## Outcomes\nUsers succeed."
	if got := formatOutput("  " + want + "  "); got != want {
		t.Fatalf("formatOutput() = %q, want %q", got, want)
	}
}
