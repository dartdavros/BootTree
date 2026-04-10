package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestNewRootCommand_HelpFlag(t *testing.T) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)

	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	os.Args = []string{"boottree", "--help"}

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	text := out.String()
	if !strings.Contains(text, "Available Commands:") {
		t.Fatalf("expected help output, got %q", text)
	}
	if !strings.Contains(text, "version") {
		t.Fatalf("expected help to include version command, got %q", text)
	}
}

func TestNewRootCommand_VersionFlag(t *testing.T) {
	cmd := NewRootCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)

	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	os.Args = []string{"boottree", "--version"}

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	text := out.String()
	if !strings.Contains(text, "boottree ") {
		t.Fatalf("expected version output, got %q", text)
	}
}
