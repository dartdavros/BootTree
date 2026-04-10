package planner

import (
	"path/filepath"
	"testing"

	"boottree/internal/core/model"
)

func TestService_Build_ForceOverwritesExistingTemplateFiles(t *testing.T) {
	snapshot := model.TreeSnapshot{
		Root:    t.TempDir(),
		Entries: []model.TreeEntry{{Path: filepath.Clean("README.md"), IsDir: false}},
	}
	preset := model.Preset{
		Name:        "software-product",
		Sections:    []model.Section{{ID: "eng", Label: "Engineering"}},
		Directories: []model.DirectoryNode{{Path: "docs"}},
		Templates:   []model.TemplateFile{{SourceTemplate: "x.tmpl", TargetPath: "README.md"}},
	}

	plan, err := Service{}.Build(snapshot, preset, model.InitOptions{Mode: model.InitModeFoldersAndTemplates, Force: true})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if len(plan.FilesToCreate) != 1 || plan.FilesToCreate[0].Reason != "overwrite existing template file" {
		t.Fatalf("expected overwrite action, got %#v", plan.FilesToCreate)
	}
	if len(plan.SkippedExisting) != 0 {
		t.Fatalf("expected no skipped files, got %#v", plan.SkippedExisting)
	}
}
