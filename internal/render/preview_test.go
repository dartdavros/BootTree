package render

import (
	"strings"
	"testing"

	"boottree/internal/core/model"
)

func TestRenderExecutionPlan(t *testing.T) {
	plan := model.ExecutionPlan{
		DirectoriesToCreate: []model.PlanAction{{Path: "docs", Reason: "missing directory from preset"}},
		FilesToCreate:       []model.PlanAction{{Path: "README.md", Reason: "missing template target"}},
		SkippedExisting:     []model.PlanAction{{Path: "src", Reason: "directory already exists"}},
		Conflicts:           []model.PlanAction{{Path: "docs", Reason: "target directory path already occupied by file"}},
		Warnings:            []string{"template target \"README.md\" is duplicated in preset"},
	}

	output := RenderExecutionPlan(plan)
	for _, expected := range []string{
		"Plan summary",
		"Directories to create: 1",
		"Files to create: 1",
		"docs (missing directory from preset)",
		"README.md (missing template target)",
		"Warnings",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("output does not contain %q:\n%s", expected, output)
		}
	}
}
