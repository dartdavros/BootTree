package fs

import (
	iofs "io/fs"
	"os"
	"path/filepath"

	"boottree/internal/core/model"
)

type OSFileSystem struct{}

func (OSFileSystem) Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (OSFileSystem) MkdirAll(path string) error               { return os.MkdirAll(path, 0o755) }
func (OSFileSystem) WriteFile(path string, data []byte) error { return os.WriteFile(path, data, 0o644) }
func (OSFileSystem) ReadFile(path string) ([]byte, error)     { return os.ReadFile(path) }

func (OSFileSystem) ReadDir(path string) ([]model.DirEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	result := make([]model.DirEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, dirEntry{entry})
	}
	return result, nil
}

func (OSFileSystem) WalkDir(root string, fn model.WalkDirFunc) error {
	return filepath.WalkDir(root, func(path string, d iofs.DirEntry, err error) error {
		if d == nil {
			return fn(path, nil, err)
		}
		return fn(path, dirEntry{d}, err)
	})
}

type dirEntry struct{ iofs.DirEntry }
