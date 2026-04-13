package cli

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"strings"
	"testing"

	"boottree/internal/app"
	"boottree/internal/core/model"

	"github.com/spf13/cobra"
)

type stubInitPrompter struct {
	preset    string
	mode      model.InitMode
	sections  []string
	confirmed bool
}

func (s stubInitPrompter) SelectPreset(_ io.Reader, _ io.Writer, _ io.Writer, _ []model.Preset) (string, error) {
	return s.preset, nil
}

func (s stubInitPrompter) SelectMode(_ io.Reader, _ io.Writer, _ io.Writer) (model.InitMode, error) {
	return s.mode, nil
}

func (s stubInitPrompter) SelectSections(_ io.Reader, _ io.Writer, _ io.Writer, _ []model.Section) ([]string, error) {
	return s.sections, nil
}

func (s stubInitPrompter) ConfirmApply(_ io.Reader, _ io.Writer, _ io.Writer) (bool, error) {
	return s.confirmed, nil
}

func TestReadInitFlags_NonInteractiveWhenAnyFlagChanged(t *testing.T) {
	cmd := &cobra.Command{Use: "init"}
	flags := &initCommandFlags{}
	bindInitFlags(cmd, flags)

	args := []string{"--preset=software-product", "--mode=folders-only", "--include=business,engineering", "--dry-run", "--yes", "--force"}
	if err := cmd.ParseFlags(args); err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	parsed := readInitFlags(cmd, flags)
	if parsed.Interactive {
		t.Fatal("expected non-interactive mode")
	}
	if !reflect.DeepEqual(parsed.Include, []string{"business", "engineering"}) {
		t.Fatalf("unexpected sections: %#v", parsed.Include)
	}
}

func TestCompleteInitOptions_InteractiveUsesPromptSelections(t *testing.T) {
	bootstrap := app.NewBootstrap()
	var out bytes.Buffer
	prompter := stubInitPrompter{
		preset:   "software-product",
		mode:     model.InitModeFoldersOnly,
		sections: []string{"inbox", "docs", "archive"},
	}

	options, presetName, err := completeInitOptions(context.Background(), bootstrap.Presets, strings.NewReader(""), &out, &out, prompter, parsedInitArgs{Interactive: true})
	if err != nil {
		t.Fatalf("completeInitOptions returned error: %v", err)
	}

	if presetName != "software-product" {
		t.Fatalf("unexpected preset: %q", presetName)
	}
	if options.Mode != model.InitModeFoldersOnly {
		t.Fatalf("unexpected mode: %q", options.Mode)
	}
	wantSections := []string{"inbox", "docs", "archive"}
	if !reflect.DeepEqual(options.SelectedSections, wantSections) {
		t.Fatalf("unexpected selected sections: %#v", options.SelectedSections)
	}
}

func TestCompactStrings_DeduplicatesAndPreservesOrder(t *testing.T) {
	got := compactStrings([]string{"marketing", "business", "marketing", "", " product "})
	want := []string{"marketing", "business", "product"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("compactStrings() = %#v, want %#v", got, want)
	}
}

func TestValidateSections_RejectsUnknownSection(t *testing.T) {
	err := validateSections([]string{"unknown"}, []model.Section{{ID: "known", Label: "Known"}})
	if err == nil {
		t.Fatal("expected validateSections to reject unknown section")
	}
}
