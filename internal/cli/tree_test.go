package cli

import "testing"

func TestParseTreeFlags(t *testing.T) {
	parsed, err := parseTreeFlags(nil, treeFlagValues{Depth: 2, All: true})
	if err != nil {
		t.Fatalf("parseTreeFlags() error = %v", err)
	}
	if parsed.Depth != 2 {
		t.Fatalf("unexpected depth: %d", parsed.Depth)
	}
	if !parsed.All {
		t.Fatal("expected --all to be enabled")
	}
}

func TestParseTreeFlags_RejectsNegativeDepth(t *testing.T) {
	if _, err := parseTreeFlags(nil, treeFlagValues{Depth: -1}); err == nil {
		t.Fatal("expected parseTreeFlags() to reject negative depth")
	}
}
