package platform

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPathListContainsDir(t *testing.T) {
	dir := filepath.Join("home", "user", ".local", "bin")
	pathValue := strings.Join([]string{"/usr/bin", dir, "/bin"}, string(filepath.ListSeparator))
	if !pathListContainsDir(pathValue, dir, false) {
		t.Fatalf("expected %q to contain %q", pathValue, dir)
	}
}

func TestPathListContainsDir_CaseInsensitive(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("case-insensitive check is already covered by runtime-specific behavior")
	}

	pathValue := `C:\Users\Example\AppData\Local\Programs\BootTree\bin`
	dir := `c:\users\example\appdata\local\programs\boottree\bin`
	if !pathListContainsDir(pathValue, dir, true) {
		t.Fatalf("expected case-insensitive PATH match for %q and %q", pathValue, dir)
	}
}

func TestSameFilePath(t *testing.T) {
	left := filepath.Join("tmp", "boottree")
	right := filepath.Join("tmp", ".", "boottree")
	if !sameFilePath(left, right) {
		t.Fatalf("expected %q and %q to be treated as the same path", left, right)
	}
}

func TestPathListContainsDir_WindowsListOnNonWindowsRunner(t *testing.T) {
	pathValue := strings.Join([]string{
		`C:\Windows\System32`,
		`C:\Users\Example\AppData\Local\Programs\BootTree\bin`,
		`C:\Tools`,
	}, ";")
	dir := `c:\users\example\appdata\local\programs\boottree\bin`
	if !pathListContainsDir(pathValue, dir, true) {
		t.Fatalf("expected case-insensitive Windows PATH match for %q and %q", pathValue, dir)
	}
}
