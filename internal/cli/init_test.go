package cli

import (
	"reflect"
	"testing"

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

func TestValidateSections_RejectsUnknownSection(t *testing.T) {
	err := validateSections([]string{"unknown"}, []model.Section{{ID: "known", Label: "Known"}})
	if err == nil {
		t.Fatal("expected validateSections to reject unknown section")
	}
}
