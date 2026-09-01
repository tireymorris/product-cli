package version

import "testing"

func TestInfo(t *testing.T) {
	oldVersion, oldCommit, oldRef := Version, Commit, Ref
	t.Cleanup(func() { Version, Commit, Ref = oldVersion, oldCommit, oldRef })

	Version = "v1.2.3"
	Commit = "abc123"
	Ref = "main"

	if got, want := Info(), "product-cli version=v1.2.3 commit=abc123 ref=main"; got != want {
		t.Fatalf("Info() = %q, want %q", got, want)
	}
}
