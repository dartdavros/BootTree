package platform

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	ApplicationName = "BootTree"
	CommandName     = "boottree"
)

type InstallState struct {
	CommandName          string
	CurrentExecutable    string
	SuggestedInstallDir  string
	AvailableInPath      bool
	SupportsPathMutation bool
}

type InstallResult struct {
	CommandName              string
	InstalledExecutable      string
	InstallDir               string
	PathUpdated              bool
	ManualPathUpdateRequired bool
	ShellRestartRecommended  bool
}

type SelfInstaller struct{}

func (SelfInstaller) Detect() (InstallState, error) {
	currentExecutable, err := currentExecutablePath()
	if err != nil {
		return InstallState{}, err
	}

	installDir, err := userInstallDir()
	if err != nil {
		return InstallState{}, err
	}

	return InstallState{
		CommandName:          commandBinaryName(),
		CurrentExecutable:    currentExecutable,
		SuggestedInstallDir:  installDir,
		AvailableInPath:      commandAvailableInPath(commandBinaryName()),
		SupportsPathMutation: runtime.GOOS == "windows",
	}, nil
}

func (SelfInstaller) InstallForCurrentUser() (InstallResult, error) {
	currentExecutable, err := currentExecutablePath()
	if err != nil {
		return InstallResult{}, err
	}

	installDir, err := userInstallDir()
	if err != nil {
		return InstallResult{}, err
	}

	targetExecutable := filepath.Join(installDir, commandBinaryName())
	if err := copyExecutable(currentExecutable, targetExecutable); err != nil {
		return InstallResult{}, err
	}

	pathUpdated, err := ensureUserPathEntry(installDir)
	if err != nil {
		return InstallResult{}, err
	}

	return InstallResult{
		CommandName:              commandBinaryName(),
		InstalledExecutable:      targetExecutable,
		InstallDir:               installDir,
		PathUpdated:              pathUpdated,
		ManualPathUpdateRequired: runtime.GOOS != "windows",
		ShellRestartRecommended:  true,
	}, nil
}

func commandBinaryName() string {
	if runtime.GOOS == "windows" {
		return CommandName + ".exe"
	}
	return CommandName
}

func currentExecutablePath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve current executable: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		if os.IsNotExist(err) {
			return filepath.Clean(path), nil
		}
		return "", fmt.Errorf("resolve executable symlinks: %w", err)
	}
	return filepath.Clean(resolved), nil
}

func commandAvailableInPath(binaryName string) bool {
	_, err := exec.LookPath(binaryName)
	return err == nil
}

func userInstallDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}

	switch runtime.GOOS {
	case "windows":
		base := strings.TrimSpace(os.Getenv("LOCALAPPDATA"))
		if base == "" {
			base = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(base, "Programs", ApplicationName, "bin"), nil
	default:
		return filepath.Join(home, ".local", "bin"), nil
	}
}

func copyExecutable(source, target string) error {
	source = filepath.Clean(source)
	target = filepath.Clean(target)

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create install directory %q: %w", filepath.Dir(target), err)
	}

	if sameFilePath(source, target) {
		return nil
	}

	in, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open source executable %q: %w", source, err)
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return fmt.Errorf("stat source executable %q: %w", source, err)
	}

	tempFile, err := os.CreateTemp(filepath.Dir(target), ".boottree-install-*")
	if err != nil {
		return fmt.Errorf("create temp executable in %q: %w", filepath.Dir(target), err)
	}
	tempName := tempFile.Name()
	defer func() {
		_ = os.Remove(tempName)
	}()

	if _, err := io.Copy(tempFile, in); err != nil {
		tempFile.Close()
		return fmt.Errorf("copy executable to temp file %q: %w", tempName, err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close temp executable %q: %w", tempName, err)
	}

	if runtime.GOOS != "windows" {
		mode := info.Mode().Perm()
		if mode == 0 {
			mode = 0o755
		}
		if err := os.Chmod(tempName, mode); err != nil {
			return fmt.Errorf("set executable permissions on %q: %w", tempName, err)
		}
	}

	if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove previous executable %q: %w", target, err)
	}
	if err := os.Rename(tempName, target); err != nil {
		return fmt.Errorf("install executable to %q: %w", target, err)
	}
	return nil
}

func sameFilePath(a, b string) bool {
	left := filepath.Clean(a)
	right := filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}

func pathListContainsDir(pathValue, dir string, caseInsensitive bool) bool {
	want := normalizePathEntry(strings.TrimSpace(dir), caseInsensitive)
	if want == "" {
		return false
	}

	for _, entry := range splitPathList(pathValue, caseInsensitive) {
		candidate := normalizePathEntry(strings.TrimSpace(entry), caseInsensitive)
		if candidate == "" {
			continue
		}
		if caseInsensitive {
			if strings.EqualFold(candidate, want) {
				return true
			}
			continue
		}
		if candidate == want {
			return true
		}
	}
	return false
}

func splitPathList(pathValue string, caseInsensitive bool) []string {
	if caseInsensitive {
		return strings.Split(pathValue, ";")
	}
	return filepath.SplitList(pathValue)
}

func normalizePathEntry(value string, caseInsensitive bool) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	if caseInsensitive {
		normalized := strings.ReplaceAll(value, "\\", "/")
		normalized = path.Clean(normalized)
		if normalized == "." {
			return ""
		}
		return normalized
	}

	normalized := filepath.Clean(value)
	if normalized == "." {
		return ""
	}
	return normalized
}
