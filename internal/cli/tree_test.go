package cli

import (
	"strings"
	"testing"
)

func TestTreeCommand_RejectsNegativeDepth(t *testing.T) {
	cmd := newTreeCommand()
	cmd.SetArgs([]string{"--depth=-1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected Execute() to reject negative depth")
	}
	if !strings.Contains(err.Error(), "invalid depth") {
		t.Fatalf("unexpected error: %v", err)
	}
}
