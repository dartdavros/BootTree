package stats

import (
	"path/filepath"
	"reflect"
	"testing"

	"boottree/internal/core/model"
)

func TestService_Build(t *testing.T) {
	snapshot := model.TreeSnapshot{
		Root: "project",
		Entries: []model.TreeEntry{
			{Path: "docs", IsDir: true},
			{Path: "docs/README.md", IsDir: false},
			{Path: "src", IsDir: true},
			{Path: "src/main.go", IsDir: false},
			{Path: "src/config", IsDir: false},
			{Path: "empty", IsDir: true},
			{Path: ".env", IsDir: false},
			{Path: "certs", IsDir: true},
			{Path: "certs/server.pem", IsDir: false},
			{Path: "keys", IsDir: true},
			{Path: "keys/service.key", IsDir: false},
			{Path: "ops", IsDir: true},
			{Path: "ops/secrets.prod", IsDir: false},
			{Path: "ops/credentials.json", IsDir: false},
		},
	}

	got := Service{}.Build(snapshot)

	if got.Directories != 6 {
		t.Fatalf("Directories = %d, want 6", got.Directories)
	}
	if got.Files != 8 {
		t.Fatalf("Files = %d, want 8", got.Files)
	}
	if got.EmptyDirectories != 1 {
		t.Fatalf("EmptyDirectories = %d, want 1", got.EmptyDirectories)
	}
	if !reflect.DeepEqual(got.EmptyDirectoryPaths, []string{"empty"}) {
		t.Fatalf("EmptyDirectoryPaths = %#v, want [empty]", got.EmptyDirectoryPaths)
	}

	wantExtensions := []model.ExtensionStat{
		{Extension: "[no extension]", Count: 2},
		{Extension: ".go", Count: 1},
		{Extension: ".json", Count: 1},
		{Extension: ".key", Count: 1},
		{Extension: ".md", Count: 1},
		{Extension: ".pem", Count: 1},
		{Extension: ".prod", Count: 1},
	}
	if !reflect.DeepEqual(got.ByExtension, wantExtensions) {
		t.Fatalf("ByExtension = %#v, want %#v", got.ByExtension, wantExtensions)
	}

	wantSecrets := []string{".env", filepath.FromSlash("certs/server.pem"), filepath.FromSlash("keys/service.key"), filepath.FromSlash("ops/credentials.json"), filepath.FromSlash("ops/secrets.prod")}
	if !reflect.DeepEqual(got.SecretLikeFilePaths, wantSecrets) {
		t.Fatalf("SecretLikeFilePaths = %#v, want %#v", got.SecretLikeFilePaths, wantSecrets)
	}
}

func TestNormalizeExtension(t *testing.T) {
	if got := normalizeExtension("README"); got != "[no extension]" {
		t.Fatalf("normalizeExtension without extension = %q, want [no extension]", got)
	}
	if got := normalizeExtension("src/MAIN.GO"); got != ".go" {
		t.Fatalf("normalizeExtension should lowercase extensions, got %q", got)
	}
}
