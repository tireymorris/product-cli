package args

import "testing"

func TestParseVersionFlag(t *testing.T) {
	for _, argv := range [][]string{{"--version"}, {"version"}} {
		opts, err := Parse(argv)
		if err != nil {
			t.Fatalf("Parse(%v) error = %v", argv, err)
		}
		if !opts.Version {
			t.Errorf("Parse(%v).Version = false, want true", argv)
		}
	}
}
