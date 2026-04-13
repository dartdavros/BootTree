package update

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func stageBinaryForReplace(sourcePath, targetPath string) (string, os.FileMode, error) {
	source, err := os.Open(sourcePath)
	if err != nil {
		return "", 0, fmt.Errorf("open extracted binary %q: %w", sourcePath, err)
	}
	defer source.Close()
	info, err := source.Stat()
	if err != nil {
		return "", 0, fmt.Errorf("stat extracted binary %q: %w", sourcePath, err)
	}
	mode := info.Mode().Perm()
	if mode == 0 {
		mode = 0o755
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return "", 0, fmt.Errorf("create target directory %q: %w", filepath.Dir(targetPath), err)
	}
	tempFile, err := os.CreateTemp(filepath.Dir(targetPath), ".boottree-update-*")
	if err != nil {
		return "", 0, fmt.Errorf("create staged binary in %q: %w", filepath.Dir(targetPath), err)
	}
	tempPath := tempFile.Name()
	if _, err := io.Copy(tempFile, source); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
		return "", 0, fmt.Errorf("copy extracted binary to %q: %w", tempPath, err)
	}
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return "", 0, fmt.Errorf("close staged binary %q: %w", tempPath, err)
	}
	if err := os.Chmod(tempPath, mode); err != nil {
		_ = os.Remove(tempPath)
		return "", 0, fmt.Errorf("set executable permissions on %q: %w", tempPath, err)
	}
	return tempPath, mode, nil
}
