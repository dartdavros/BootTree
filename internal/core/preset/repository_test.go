package preset

import (
	"context"
	"strings"
	"testing"
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
	if len(preset.Templates) != 2 {
		t.Fatalf("unexpected templates count: %d", len(preset.Templates))
	}
}
