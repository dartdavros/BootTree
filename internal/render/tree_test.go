package render

import (
	"path/filepath"
	"strings"
	"testing"

	"boottree/internal/core/model"
)

func TestRenderTree_RendersStableHierarchy(t *testing.T) {
	snapshot := model.TreeSnapshot{
		Root: "/workspace/demo",
		Entries: []model.TreeEntry{
			{Path: filepath.Join("src"), IsDir: true},
			{Path: filepath.Join("src", "internal"), IsDir: true},
			{Path: filepath.Join("src", "internal", "service.go"), IsDir: false},
			{Path: filepath.Join("README.md"), IsDir: false},
			{Path: filepath.Join("cmd"), IsDir: true},
			{Path: filepath.Join("cmd", "boottree"), IsDir: true},
			{Path: filepath.Join("cmd", "boottree", "main.go"), IsDir: false},
		},
	}

	got := RenderTree(snapshot, TreeRenderOptions{})

	for _, want := range []string{
		"demo/",
		"├── cmd/",
		"│   └── boottree/",
		"│       └── main.go",
		"├── src/",
		"│   └── internal/",
		"│       └── service.go",
		"└── README.md",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("RenderTree() missing %q in output:\n%s", want, got)
		}
	}
}

func TestRenderTree_AppliesDepthLimit(t *testing.T) {
	snapshot := model.TreeSnapshot{
		Root: "/workspace/demo",
		Entries: []model.TreeEntry{
			{Path: filepath.Join("src"), IsDir: true},
			{Path: filepath.Join("src", "internal"), IsDir: true},
			{Path: filepath.Join("src", "internal", "service.go"), IsDir: false},
		},
	}

	got := RenderTree(snapshot, TreeRenderOptions{MaxDepth: 1})

	if !strings.Contains(got, "└── src/") {
		t.Fatalf("RenderTree() should render first level directory:\n%s", got)
	}
	if strings.Contains(got, "internal/") || strings.Contains(got, "service.go") {
		t.Fatalf("RenderTree() should trim deeper entries when depth=1:\n%s", got)
	}
}
