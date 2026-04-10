package planner

import (
	"path/filepath"
	"testing"

	"boottree/internal/core/model"
)

func TestService_Build_FoldersOnlyFiltersSections(t *testing.T) {
	preset := model.Preset{
		Name:     "software-product",
		Sections: []model.Section{{ID: "docs", Label: "Docs"}, {ID: "eng", Label: "Engineering"}},
		Directories: []model.DirectoryNode{
			{Path: "docs", Sections: []string{"docs"}, Children: []model.DirectoryNode{{Path: "notes"}}},
			{Path: "src", Sections: []string{"eng"}},
		},
		Templates: []model.TemplateFile{{SourceTemplate: "readme.tmpl", TargetPath: "README.md"}},
	}
	snapshot := model.TreeSnapshot{Entries: []model.TreeEntry{{Path: "docs", IsDir: true}}}

	plan, err := Service{}.Build(snapshot, preset, model.InitOptions{Mode: model.InitModeFoldersOnly, SelectedSections: []string{"docs"}})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	wantDir := filepath.Join("docs", "notes")
	if len(plan.DirectoriesToCreate) != 1 || plan.DirectoriesToCreate[0].Path != wantDir {
		t.Fatalf("unexpected directories to create: %#v", plan.DirectoriesToCreate)
	}
	if len(plan.FilesToCreate) != 0 {
		t.Fatalf("folders-only should not create files: %#v", plan.FilesToCreate)
	}
}

func TestService_Build_DetectsConflictsAndSkips(t *testing.T) {
	preset := model.Preset{
		Name:     "software-product",
		Sections: []model.Section{{ID: "eng", Label: "Engineering"}},
		Directories: []model.DirectoryNode{
			{Path: "docs", Sections: []string{"eng"}},
			{Path: "bin", Sections: []string{"eng"}},
		},
		Templates: []model.TemplateFile{{SourceTemplate: "arch.tmpl", TargetPath: "docs/README.md", Sections: []string{"eng"}}},
	}
	snapshot := model.TreeSnapshot{Entries: []model.TreeEntry{
		{Path: "docs", IsDir: false},
		{Path: filepath.Join("docs", "README.md"), IsDir: false},
		{Path: "bin", IsDir: true},
	}}

	plan, err := Service{}.Build(snapshot, preset, model.InitOptions{Mode: model.InitModeFoldersAndTemplates})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if len(plan.Conflicts) != 2 {
		t.Fatalf("unexpected conflicts: %#v", plan.Conflicts)
	}
	if len(plan.SkippedExisting) != 1 || plan.SkippedExisting[0].Path != "bin" {
		t.Fatalf("unexpected skipped items: %#v", plan.SkippedExisting)
	}
}

func TestService_Build_WarnsOnDuplicateTemplateTargets(t *testing.T) {
	preset := model.Preset{
		Name:     "software-product",
		Sections: []model.Section{{ID: "eng", Label: "Engineering"}},
		Directories: []model.DirectoryNode{
			{Path: "docs", Sections: []string{"eng"}},
		},
		Templates: []model.TemplateFile{
			{SourceTemplate: "a.tmpl", TargetPath: "README.md"},
			{SourceTemplate: "b.tmpl", TargetPath: "README.md"},
		},
	}

	plan, err := Service{}.Build(model.TreeSnapshot{}, preset, model.InitOptions{Mode: model.InitModeFoldersAndTemplates})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if len(plan.FilesToCreate) != 1 || plan.FilesToCreate[0].Path != filepath.Clean("README.md") {
		t.Fatalf("unexpected files to create: %#v", plan.FilesToCreate)
	}
	if len(plan.Warnings) != 1 {
		t.Fatalf("unexpected warnings: %#v", plan.Warnings)
	}
}
