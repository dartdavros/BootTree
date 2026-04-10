package scanner

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"

	"boottree/internal/core/model"
	"boottree/internal/platform"
)

type Service struct {
	FS model.FileSystem
}

func (s Service) Scan(ctx context.Context, root string) (model.TreeSnapshot, error) {
	_ = ctx

	if s.FS == nil {
		return model.TreeSnapshot{}, fmt.Errorf("scan tree: file system is required")
	}

	cleanRoot := filepath.Clean(root)
	entries := make([]model.TreeEntry, 0, 32)

	err := s.FS.WalkDir(cleanRoot, func(path string, entry model.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk %q: %w", path, err)
		}
		if entry == nil {
			return nil
		}

		rel, err := filepath.Rel(cleanRoot, path)
		if err != nil {
			return fmt.Errorf("compute relative path for %q: %w", path, err)
		}
		if rel == "." {
			return nil
		}
		if platform.ShouldIgnore(rel) {
			return nil
		}

		entries = append(entries, model.TreeEntry{
			Path:  filepath.Clean(rel),
			IsDir: entry.IsDir(),
		})
		return nil
	})
	if err != nil {
		return model.TreeSnapshot{}, err
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Path == entries[j].Path {
			return entries[i].IsDir && !entries[j].IsDir
		}
		return entries[i].Path < entries[j].Path
	})

	return model.TreeSnapshot{Root: cleanRoot, Entries: entries}, nil
}
