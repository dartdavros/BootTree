package update

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"boottree/internal/buildinfo"
)

type Service struct {
	ManifestClient ManifestClient
	Downloader     Downloader
	Verifier       ChecksumVerifier
	Extractor      Extractor
}

func NewService() Service {
	return Service{
		ManifestClient: ManifestClient{},
		Downloader:     Downloader{},
		Verifier:       ChecksumVerifier{},
		Extractor:      Extractor{},
	}
}

func (s Service) BuildPlan(ctx context.Context, options Options) (Plan, error) {
	manifestURL := strings.TrimSpace(options.ManifestURL)
	if manifestURL == "" {
		manifestURL = strings.TrimSpace(buildinfo.UpdateManifestURL)
	}
	if manifestURL == "" {
		return Plan{}, fmt.Errorf("update manifest URL is not configured; provide --manifest-url or inject boottree/internal/buildinfo.UpdateManifestURL at build time")
	}

	installPath, currentExecutable, err := resolveInstallPath(options.InstallPath)
	if err != nil {
		return Plan{}, err
	}
	manifest, err := s.ManifestClient.Fetch(ctx, manifestURL)
	if err != nil {
		return Plan{}, err
	}

	channel := strings.TrimSpace(options.Channel)
	if channel == "" {
		channel = strings.TrimSpace(manifest.Channel)
	}
	if channel == "" {
		channel = "stable"
	}

	release, err := ResolveRelease(manifest, options.Version)
	if err != nil {
		return Plan{}, err
	}
	asset, err := ResolveAsset(release, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return Plan{}, err
	}

	currentVersion := buildinfo.Version
	plan := Plan{
		CurrentVersion:       currentVersion,
		TargetVersion:        normalizeVersion(release.Version),
		GOOS:                 runtime.GOOS,
		GOARCH:               runtime.GOARCH,
		Channel:              channel,
		ManifestURL:          manifestURL,
		InstallPath:          installPath,
		Asset:                asset,
		TempDir:              filepath.Join(os.TempDir(), fmt.Sprintf("boottree-update-%s-%s", normalizeVersion(release.Version), runtime.GOOS)),
		TempArchivePath:      filepath.Join(os.TempDir(), fmt.Sprintf("boottree-update-%s-%s%s", normalizeVersion(release.Version), runtime.GOOS, archiveExtension(asset.Archive))),
		ExtractedBinaryPath:  filepath.Join(os.TempDir(), fmt.Sprintf("boottree-update-%s-%s", normalizeVersion(release.Version), asset.Binary)),
		BackupPath:           installPath + ".bak",
		NeedsElevation:       !canWriteTarget(installPath),
		RequiresDeferredSwap: runtime.GOOS == "windows" && sameFilePath(currentExecutable, installPath),
	}

	if compare, err := compareCurrentToTarget(currentVersion, plan.TargetVersion); err == nil && compare >= 0 {
		plan.IsNoop = true
	}
	if normalizeVersion(options.Version) != "" {
		plan.IsNoop = compareExactVersion(currentVersion, plan.TargetVersion)
	}
	return plan, nil
}

func (s Service) Apply(ctx context.Context, plan Plan) (Result, error) {
	if plan.IsNoop {
		return Result{InstalledVersion: plan.CurrentVersion, PreviousVersion: plan.CurrentVersion, InstallPath: plan.InstallPath}, nil
	}
	if err := os.RemoveAll(plan.TempDir); err != nil {
		return Result{}, fmt.Errorf("cleanup temp directory %q: %w", plan.TempDir, err)
	}
	if err := os.MkdirAll(plan.TempDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("create temp directory %q: %w", plan.TempDir, err)
	}
	cleanupTempDir := true
	defer func() {
		if cleanupTempDir {
			_ = os.RemoveAll(plan.TempDir)
		}
	}()

	archivePath := filepath.Join(plan.TempDir, filepath.Base(plan.TempArchivePath))
	if err := s.Downloader.Download(ctx, plan.Asset.URL, archivePath); err != nil {
		return Result{}, err
	}
	if err := s.Verifier.VerifyFile(archivePath, plan.Asset.SHA256); err != nil {
		return Result{}, err
	}
	extractedPath, err := s.Extractor.ExtractBinary(archivePath, plan.Asset.Binary, plan.Asset.Archive, plan.TempDir)
	if err != nil {
		return Result{}, err
	}

	result, err := installBinary(ctx, plan, extractedPath)
	if err != nil {
		return Result{}, err
	}
	cleanupTempDir = shouldCleanupTempDir(result)
	return result, nil
}

func shouldCleanupTempDir(result Result) bool {
	return !result.Deferred
}

func resolveInstallPath(override string) (string, string, error) {
	currentExecutable, err := os.Executable()
	if err != nil {
		return "", "", fmt.Errorf("resolve current executable path: %w", err)
	}
	currentExecutable = filepath.Clean(currentExecutable)
	if strings.TrimSpace(override) != "" {
		return filepath.Clean(override), currentExecutable, nil
	}
	return currentExecutable, currentExecutable, nil
}

func compareCurrentToTarget(currentVersion string, targetVersion string) (int, error) {
	return compareVersions(normalizeVersion(currentVersion), normalizeVersion(targetVersion))
}

func compareExactVersion(currentVersion string, targetVersion string) bool {
	return normalizeVersion(currentVersion) == normalizeVersion(targetVersion)
}

func archiveExtension(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "zip":
		return ".zip"
	case "tar.gz", "tgz":
		return ".tar.gz"
	case "binary", "exe", "":
		return filepath.Ext(defaultBinaryName(runtime.GOOS))
	default:
		return ".bin"
	}
}

func canWriteTarget(targetPath string) bool {
	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false
	}
	probe, err := os.CreateTemp(dir, ".boottree-write-check-*")
	if err != nil {
		return false
	}
	name := probe.Name()
	_ = probe.Close()
	_ = os.Remove(name)
	return true
}

func sameFilePath(left string, right string) bool {
	left = filepath.Clean(left)
	right = filepath.Clean(right)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}
