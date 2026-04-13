package update

import (
	"fmt"
	"runtime"
	"strings"
)

func ResolveRelease(manifest Manifest, version string) (Release, error) {
	version = normalizeVersion(version)
	if version == "" {
		version = manifest.Latest
	}
	if version == "" {
		return Release{}, fmt.Errorf("manifest does not define a target version")
	}
	for _, release := range manifest.Releases {
		if normalizeVersion(release.Version) == version {
			return release, nil
		}
	}
	return Release{}, fmt.Errorf("manifest does not contain release %q", version)
}

func ResolveAsset(release Release, goos string, goarch string) (Asset, error) {
	for _, asset := range release.Assets {
		if strings.EqualFold(asset.OS, goos) && strings.EqualFold(asset.Arch, goarch) {
			if asset.Binary == "" {
				asset.Binary = defaultBinaryName(goos)
			}
			if asset.Archive == "" {
				asset.Archive = inferArchiveType(asset.URL, asset.Binary)
			}
			return asset, nil
		}
	}
	return Asset{}, fmt.Errorf("release %q does not contain an asset for %s/%s", release.Version, goos, goarch)
}

func defaultBinaryName(goos string) string {
	if goos == "windows" {
		return "boottree.exe"
	}
	return "boottree"
}

func inferArchiveType(sourceURL string, binaryName string) string {
	lower := strings.ToLower(sourceURL)
	switch {
	case strings.HasSuffix(lower, ".tar.gz"):
		return "tar.gz"
	case strings.HasSuffix(lower, ".zip"):
		return "zip"
	case strings.HasSuffix(lower, ".exe"):
		return "binary"
	case strings.HasSuffix(lower, "/"+strings.ToLower(binaryName)):
		return "binary"
	default:
		if runtime.GOOS == "windows" {
			return "binary"
		}
		return "tar.gz"
	}
}
