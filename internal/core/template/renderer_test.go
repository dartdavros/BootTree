package template

import (
	"context"
	"strings"
	"testing"
	"time"

	"boottree/internal/core/model"
)

func TestRenderer_Render(t *testing.T) {
	r := Renderer{}
	result, err := r.Render("# {{ .ProjectName }} {{ .Year }}", model.TemplateData{ProjectName: "BootTree", Year: 2026, CreatedAt: time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if result != "# BootTree 2026" {
		t.Fatalf("unexpected result: %q", result)
	}
}

func TestRenderer_Render_MissingKey(t *testing.T) {
	r := Renderer{}
	_, err := r.Render("{{ .Variables.owner }}", model.TemplateData{})
	if err == nil {
		t.Fatal("expected render error")
	}
}

func TestEmbeddedRepository_Get(t *testing.T) {
	repo := NewEmbeddedRepository()
	content, err := repo.Get(context.Background(), "software-product/docs/README.md.tmpl")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !strings.Contains(content, "ProjectName") {
		t.Fatalf("unexpected template content: %q", content)
	}
}
