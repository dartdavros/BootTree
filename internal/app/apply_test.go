package app

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"boottree/internal/core/model"
	coretemplate "boottree/internal/core/template"
	fsinfra "boottree/internal/fs"
)

func TestInitApplier_Apply_CreatesDirectoriesAndFiles(t *testing.T) {
	root := t.TempDir()
	applier := InitApplier{
		FS:        fsinfra.OSFileSystem{},
		Templates: templateRepoStub{templates: map[string]string{"preset/README.md.tmpl": "# {{.ProjectName}}\n{{.Variables.Note}}"}},
		Renderer:  coretemplate.Renderer{},
	}
	preset := model.Preset{
		Name:     "test",
		Sections: []model.Section{{ID: "eng", Label: "Engineering"}},
		Templates: []model.TemplateFile{{
			SourceTemplate: "preset/README.md.tmpl",
			TargetPath:     filepath.Join("docs", "README.md"),
			Variables:      map[string]string{"Note": "hello"},
		}},
	}
	plan := model.ExecutionPlan{
		DirectoriesToCreate: []model.PlanAction{{Path: filepath.Clean("docs"), Reason: "missing directory from preset"}},
		FilesToCreate:       []model.PlanAction{{Path: filepath.Join("docs", "README.md"), Reason: "missing template target"}},
	}

	if err := applier.Apply(context.Background(), root, preset, model.InitOptions{Mode: model.InitModeFoldersAndTemplates}, plan); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	content, err := fsinfra.OSFileSystem{}.ReadFile(filepath.Join(root, "docs", "README.md"))
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, filepath.Base(root)) {
		t.Fatalf("expected project name in rendered content, got %q", text)
	}
	if !strings.Contains(text, "hello") {
		t.Fatalf("expected custom variable in rendered content, got %q", text)
	}
}

type templateRepoStub struct{ templates map[string]string }

func (s templateRepoStub) Get(ctx context.Context, path string) (string, error) {
	_ = ctx
	return s.templates[path], nil
}
