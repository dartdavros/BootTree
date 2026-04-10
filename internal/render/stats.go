package render

import (
	"fmt"
	"strings"

	"boottree/internal/core/model"
)

func RenderStats(stats model.ProjectStats) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Project stats\n")
	fmt.Fprintf(&b, "  Directories: %d\n", stats.Directories)
	fmt.Fprintf(&b, "  Files: %d\n", stats.Files)
	fmt.Fprintf(&b, "  Empty directories: %d\n", stats.EmptyDirectories)
	fmt.Fprintf(&b, "  Secret-like files: %d\n", len(stats.SecretLikeFilePaths))

	if len(stats.ByExtension) > 0 {
		b.WriteString("\nBy extension\n")
		for _, item := range stats.ByExtension {
			fmt.Fprintf(&b, "  - %s: %d\n", item.Extension, item.Count)
		}
	}

	if len(stats.EmptyDirectoryPaths) > 0 {
		b.WriteString("\nEmpty directories\n")
		for _, path := range stats.EmptyDirectoryPaths {
			fmt.Fprintf(&b, "  - %s\n", path)
		}
	}

	if len(stats.SecretLikeFilePaths) > 0 {
		b.WriteString("\nSecret-like files\n")
		for _, path := range stats.SecretLikeFilePaths {
			fmt.Fprintf(&b, "  - %s\n", path)
		}
	}

	return strings.TrimRight(b.String(), "\n")
}
