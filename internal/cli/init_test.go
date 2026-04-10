package cli

import (
	"bufio"
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	"boottree/internal/app"
	"boottree/internal/core/model"
)

func TestParseInitArgs_NonInteractiveFlags(t *testing.T) {
	parsed, err := parseInitArgs([]string{"--preset=software-product", "--mode=folders-only", "--include=01_business,06_engineering", "--dry-run", "--yes", "--force"})
	if err != nil {
		t.Fatalf("parseInitArgs returned error: %v", err)
	}
	if parsed.Interactive {
		t.Fatalf("expected non-interactive mode")
	}
	if !reflect.DeepEqual(parsed.Include, []string{"01_business", "06_engineering"}) {
		t.Fatalf("unexpected sections: %#v", parsed.Include)
	}
}

func TestCompleteInitOptions_InteractiveUsesPromptSelections(t *testing.T) {
	bootstrap := app.NewBootstrap()
	input := bufio.NewReader(strings.NewReader("\n2\n1,6,12\n"))
	var out bytes.Buffer

	options, presetName, err := completeInitOptions(context.Background(), bootstrap.Presets, input, &out, parsedInitArgs{Interactive: true})
	if err != nil {
		t.Fatalf("completeInitOptions returned error: %v", err)
	}

	if presetName != "software-product" {
		t.Fatalf("unexpected preset: %q", presetName)
	}
	if options.Mode != model.InitModeFoldersOnly {
		t.Fatalf("unexpected mode: %q", options.Mode)
	}
	wantSections := []string{"00_inbox", "05_docs", "99_archive"}
	if !reflect.DeepEqual(options.SelectedSections, wantSections) {
		t.Fatalf("unexpected selected sections: %#v", options.SelectedSections)
	}
	text := out.String()
	if !strings.Contains(text, "Select preset:") {
		t.Fatalf("expected preset prompt, got %q", text)
	}
	if !strings.Contains(text, "Select initialization mode:") {
		t.Fatalf("expected mode prompt, got %q", text)
	}
	if !strings.Contains(text, "Sections [all]:") {
		t.Fatalf("expected sections prompt, got %q", text)
	}
}

func TestConfirmApply_RepromptsUntilAnswerIsValid(t *testing.T) {
	input := bufio.NewReader(strings.NewReader("maybe\ny\n"))
	var out bytes.Buffer

	confirmed, err := confirmApply(input, &out)
	if err != nil {
		t.Fatalf("confirmApply returned error: %v", err)
	}
	if !confirmed {
		t.Fatal("expected confirmation to succeed")
	}
	if !strings.Contains(out.String(), "Enter y or n.") {
		t.Fatalf("expected validation message, got %q", out.String())
	}
}

func TestParseNumericSelection_DeduplicatesAndPreservesOrder(t *testing.T) {
	indexes, err := parseNumericSelection("3, 1, 3, 2", 4)
	if err != nil {
		t.Fatalf("parseNumericSelection returned error: %v", err)
	}
	want := []int{2, 0, 1}
	if !reflect.DeepEqual(indexes, want) {
		t.Fatalf("unexpected indexes: %#v", indexes)
	}
}

func TestValidateSections_RejectsUnknownSection(t *testing.T) {
	err := validateSections([]string{"unknown"}, []model.Section{{ID: "known", Label: "Known"}})
	if err == nil {
		t.Fatal("expected validateSections to reject unknown section")
	}
}
