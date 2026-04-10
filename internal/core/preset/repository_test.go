package preset

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"boottree/internal/core/model"
)

func TestLoad_ValidPreset(t *testing.T) {
	raw := []byte(`{"name":"software-product","description":"test","sections":[{"id":"engineering","label":"Engineering"}],"directories":[{"path":"docs","sections":["engineering"]}],"templates":[{"sourceTemplate":"readme.tmpl","targetPath":"README.md","sections":["engineering"]}]}`)
	preset, err := Load(raw)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if preset.Name != "software-product" {
		t.Fatalf("unexpected preset name: %s", preset.Name)
	}
}

func TestLoad_InvalidPreset_MissingName(t *testing.T) {
	raw := []byte(`{"sections":[{"id":"engineering","label":"Engineering"}],"directories":[{"path":"docs"}]}`)
	_, err := Load(raw)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), `invalid preset field "name"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoad_InvalidPreset_UnknownSectionReference(t *testing.T) {
	raw := []byte(`{"name":"software-product","sections":[{"id":"engineering","label":"Engineering"}],"directories":[{"path":"docs","sections":["missing"]}]}`)
	_, err := Load(raw)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), `unknown section "missing"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEmbeddedRepository_Get(t *testing.T) {
	repo := NewEmbeddedRepository()
	preset, err := repo.Get(context.Background(), "software-product")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if preset.Name != "software-product" {
		t.Fatalf("unexpected preset name: %s", preset.Name)
	}
	if len(preset.Sections) != 12 {
		t.Fatalf("unexpected sections count: %d", len(preset.Sections))
	}
	if len(preset.Templates) != 2 {
		t.Fatalf("unexpected templates count: %d", len(preset.Templates))
	}

	directories := flattenPaths(preset.Directories)
	if len(directories) != 49 {
		t.Fatalf("unexpected directories count: %d", len(directories))
	}

	mustContain := []string{
		filepath.Clean("00_inbox"),
		filepath.Clean("01_business/concept"),
		filepath.Clean("02_product/vision"),
		filepath.Clean("06_engineering/adrs"),
		filepath.Clean("06_engineering/dev-plans"),
		filepath.Clean("08_deploy/ci-cd"),
		filepath.Clean("10_secrets/local-only"),
		filepath.Clean("99_archive"),
	}
	for _, path := range mustContain {
		if _, ok := directories[path]; !ok {
			t.Fatalf("expected preset directory %q", path)
		}
	}

	mustNotContain := []string{
		filepath.Clean("01_business/vision"),
		filepath.Clean("01_business/roadmap"),
		filepath.Clean("06_engineering/adr"),
		filepath.Clean("06_engineering/plans"),
		filepath.Clean("07_repos/app"),
		filepath.Clean("07_repos/infra"),
	}
	for _, path := range mustNotContain {
		if _, ok := directories[path]; ok {
			t.Fatalf("did not expect preset directory %q", path)
		}
	}
}

func flattenPaths(nodes []model.DirectoryNode) map[string]struct{} {
	result := make(map[string]struct{})
	for _, node := range nodes {
		flattenNode(result, "", node)
	}
	return result
}

func flattenNode(result map[string]struct{}, parent string, node model.DirectoryNode) {
	current := filepath.Clean(filepath.Join(parent, node.Path))
	result[current] = struct{}{}
	for _, child := range node.Children {
		flattenNode(result, current, child)
	}
}
