package cli

import "testing"

func TestParseTreeArgs(t *testing.T) {
	parsed, err := parseTreeArgs([]string{"--depth=2", "--all"})
	if err != nil {
		t.Fatalf("parseTreeArgs() error = %v", err)
	}
	if parsed.Depth != 2 {
		t.Fatalf("unexpected depth: %d", parsed.Depth)
	}
	if !parsed.All {
		t.Fatal("expected --all to be enabled")
	}
}

func TestParseTreeArgs_RejectsNegativeDepth(t *testing.T) {
	if _, err := parseTreeArgs([]string{"--depth=-1"}); err == nil {
		t.Fatal("expected parseTreeArgs() to reject negative depth")
	}
}
