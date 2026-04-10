package render

import (
	"strings"
	"testing"

	"boottree/internal/core/model"
)

func TestRenderStats(t *testing.T) {
	stats := model.ProjectStats{
		Directories:         3,
		Files:               5,
		EmptyDirectories:    1,
		EmptyDirectoryPaths: []string{"empty"},
		ByExtension: []model.ExtensionStat{{Extension: ".go", Count: 2}, {Extension: "[no extension]", Count: 1}},
		SecretLikeFilePaths: []string{".env"},
	}

	output := RenderStats(stats)
	for _, expected := range []string{
		"Project stats",
		"Directories: 3",
		"Files: 5",
		"Empty directories: 1",
		"By extension",
		".go: 2",
		"Empty directories",
		"empty",
		"Secret-like files",
		".env",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("output does not contain %q:\n%s", expected, output)
		}
	}
}
