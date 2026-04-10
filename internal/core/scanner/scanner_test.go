package scanner

import (
	"context"
	"path/filepath"
	"testing"

	fsimpl "boottree/internal/fs"
)

func TestService_Scan_IgnoresDefaultPaths(t *testing.T) {
	tmp := t.TempDir()
	fs := fsimpl.OSFileSystem{}

	if err := fs.MkdirAll(tmp + "/src"); err != nil {
		t.Fatalf("MkdirAll(src) error = %v", err)
	}
	if err := fs.MkdirAll(tmp + "/node_modules/pkg"); err != nil {
		t.Fatalf("MkdirAll(node_modules) error = %v", err)
	}
	if err := fs.WriteFile(tmp+"/src/main.go", []byte("package main")); err != nil {
		t.Fatalf("WriteFile(main.go) error = %v", err)
	}
	if err := fs.WriteFile(tmp+"/.DS_Store", []byte("ignored")); err != nil {
		t.Fatalf("WriteFile(.DS_Store) error = %v", err)
	}
	if err := fs.WriteFile(tmp+"/node_modules/pkg/index.js", []byte("ignored")); err != nil {
		t.Fatalf("WriteFile(index.js) error = %v", err)
	}

	snapshot, err := Service{FS: fs}.Scan(context.Background(), tmp)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	got := map[string]bool{}
	for _, entry := range snapshot.Entries {
		got[entry.Path] = entry.IsDir
	}

	if !got["src"] {
		t.Fatalf("expected src directory in snapshot: %#v", got)
	}
	if got["node_modules"] {
		t.Fatalf("node_modules should be ignored: %#v", got)
	}
	if got[".DS_Store"] {
		t.Fatalf(".DS_Store should be ignored: %#v", got)
	}
	wantFile := filepath.Join("src", "main.go")
	if _, ok := got[wantFile]; !ok {
		t.Fatalf("expected %s in snapshot: %#v", wantFile, got)
	}
}
