package stats

import (
	"path/filepath"
	"sort"
	"strings"

	"boottree/internal/core/model"
)

type Service struct{}

func (Service) Build(snapshot model.TreeSnapshot) model.ProjectStats {
	directories := make([]string, 0)
	files := make([]string, 0)
	extCounts := map[string]int{}
	secretLike := make([]string, 0)
	childCounts := map[string]int{}

	for _, entry := range snapshot.Entries {
		path := filepath.Clean(entry.Path)
		if entry.IsDir {
			directories = append(directories, path)
		} else {
			files = append(files, path)
			extCounts[normalizeExtension(path)]++
			if isSecretLike(path) {
				secretLike = append(secretLike, path)
			}
		}

		parent := filepath.Dir(path)
		if parent != "." && parent != string(filepath.Separator) {
			childCounts[parent]++
		}
	}

	emptyDirectories := make([]string, 0)
	for _, dir := range directories {
		if childCounts[dir] == 0 {
			emptyDirectories = append(emptyDirectories, dir)
		}
	}

	sort.Strings(emptyDirectories)
	sort.Strings(secretLike)

	byExtension := make([]model.ExtensionStat, 0, len(extCounts))
	for extension, count := range extCounts {
		byExtension = append(byExtension, model.ExtensionStat{Extension: extension, Count: count})
	}
	sort.Slice(byExtension, func(i, j int) bool {
		if byExtension[i].Count == byExtension[j].Count {
			return byExtension[i].Extension < byExtension[j].Extension
		}
		return byExtension[i].Count > byExtension[j].Count
	})

	return model.ProjectStats{
		Files:               len(files),
		Directories:         len(directories),
		EmptyDirectories:    len(emptyDirectories),
		EmptyDirectoryPaths: emptyDirectories,
		ByExtension:         byExtension,
		SecretLikeFilePaths: secretLike,
	}
}

func normalizeExtension(path string) string {
	name := strings.ToLower(filepath.Base(path))
	if strings.HasPrefix(name, ".") && strings.Count(name, ".") == 1 {
		return "[no extension]"
	}

	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return "[no extension]"
	}
	return ext
}

func isSecretLike(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	return name == ".env" ||
		strings.HasPrefix(name, ".env.") ||
		strings.HasSuffix(name, ".key") ||
		strings.HasSuffix(name, ".pem") ||
		strings.HasPrefix(name, "secrets.") ||
		strings.HasPrefix(name, "credentials.")
}
