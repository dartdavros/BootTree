package cobra

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestExecute_RootHelpFlag(t *testing.T) {
	cmd := &Command{Use: "app", Short: "Example app"}
	var out bytes.Buffer
	cmd.SetOut(&out)

	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	os.Args = []string{"app", "--help"}

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Fatalf("expected help output, got %q", out.String())
	}
}

func TestExecute_RootVersionFlag(t *testing.T) {
	root := &Command{Use: "app", Short: "Example app"}
	version := &Command{Use: "version", Run: func(cmd *Command, args []string) {
		fmt.Fprintln(cmd.OutOrStdout(), "app dev")
	}}
	root.AddCommand(version)
	var out bytes.Buffer
	root.SetOut(&out)

	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	os.Args = []string{"app", "--version"}

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if strings.TrimSpace(out.String()) != "app dev" {
		t.Fatalf("unexpected version output: %q", out.String())
	}
}

func TestHelpText_IncludesExamples(t *testing.T) {
	cmd := &Command{Use: "app", Short: "Example app", Example: "  app run"}
	text := cmd.helpText()
	if !strings.Contains(text, "Examples:") {
		t.Fatalf("expected examples section, got %q", text)
	}
}
