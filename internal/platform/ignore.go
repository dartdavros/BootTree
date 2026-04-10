package platform

import (
	"path/filepath"
	"strings"
)

var DefaultIgnoredPaths = []string{
	".git",
	".idea",
	".vscode",
	"node_modules",
	"bin",
	"obj",
	"dist",
	"build",
	".DS_Store",
}

func ShouldIgnore(relPath string) bool {
	cleaned := filepath.Clean(relPath)
	if cleaned == "." || cleaned == string(filepath.Separator) || cleaned == "" {
		return false
	}

	for _, segment := range strings.Split(cleaned, string(filepath.Separator)) {
		if segment == "" || segment == "." {
			continue
		}
		for _, ignored := range DefaultIgnoredPaths {
			if segment == ignored {
				return true
			}
		}
	}

	return false
}
