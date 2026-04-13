package update

import "testing"

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name  string
		left  string
		right string
		want  int
	}{
		{name: "equal", left: "0.2.0", right: "0.2.0", want: 0},
		{name: "upgrade", left: "0.2.0", right: "0.3.0", want: -1},
		{name: "downgrade", left: "1.0.0", right: "0.9.0", want: 1},
		{name: "strip v", left: "v1.2.3", right: "1.2.4", want: -1},
		{name: "prerelease before stable", left: "1.2.3-rc1", right: "1.2.3", want: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := compareVersions(tt.left, tt.right)
			if err != nil {
				t.Fatalf("compareVersions() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("compareVersions(%q, %q) = %d, want %d", tt.left, tt.right, got, tt.want)
			}
		})
	}
}

func TestNormalizeManifest(t *testing.T) {
	manifest := Manifest{
		Channel: " stable ",
		Latest:  " v0.3.0 ",
		Releases: []Release{{
			Version: " v0.3.0 ",
			Assets: []Asset{{
				OS:      " windows ",
				Arch:    " amd64 ",
				URL:     " https://example.test/boottree.exe ",
				SHA256:  " ABC123 ",
				Archive: " BINARY ",
				Binary:  " boottree.exe ",
			}},
		}},
	}

	normalizeManifest(&manifest)

	if manifest.Channel != "stable" {
		t.Fatalf("manifest.Channel = %q, want %q", manifest.Channel, "stable")
	}
	if manifest.Latest != "0.3.0" {
		t.Fatalf("manifest.Latest = %q, want %q", manifest.Latest, "0.3.0")
	}
	if manifest.Releases[0].Version != "0.3.0" {
		t.Fatalf("release.Version = %q, want %q", manifest.Releases[0].Version, "0.3.0")
	}
	asset := manifest.Releases[0].Assets[0]
	if asset.OS != "windows" || asset.Arch != "amd64" {
		t.Fatalf("asset target = %s/%s, want windows/amd64", asset.OS, asset.Arch)
	}
	if asset.URL != "https://example.test/boottree.exe" {
		t.Fatalf("asset.URL = %q", asset.URL)
	}
	if asset.SHA256 != "abc123" {
		t.Fatalf("asset.SHA256 = %q, want %q", asset.SHA256, "abc123")
	}
	if asset.Archive != "binary" {
		t.Fatalf("asset.Archive = %q, want %q", asset.Archive, "binary")
	}
}

func TestResolveReleaseAndAsset(t *testing.T) {
	manifest := Manifest{Channel: "stable", Latest: "0.3.0", Releases: []Release{{Version: "0.3.0", Assets: []Asset{{OS: "windows", Arch: "amd64", URL: "https://example.test/boottree.exe", SHA256: "abc", Archive: "binary", Binary: "boottree.exe"}}}}}
	normalizeManifest(&manifest)

	release, err := ResolveRelease(manifest, "")
	if err != nil {
		t.Fatalf("ResolveRelease() error = %v", err)
	}
	if release.Version != "0.3.0" {
		t.Fatalf("release.Version = %q, want %q", release.Version, "0.3.0")
	}
	asset, err := ResolveAsset(release, "windows", "amd64")
	if err != nil {
		t.Fatalf("ResolveAsset() error = %v", err)
	}
	if asset.Binary != "boottree.exe" {
		t.Fatalf("asset.Binary = %q, want %q", asset.Binary, "boottree.exe")
	}
}

func TestCompareCurrentToTarget(t *testing.T) {
	compare, err := compareCurrentToTarget("0.3.0", "0.3.0")
	if err != nil {
		t.Fatalf("compareCurrentToTarget() error = %v", err)
	}
	if compare != 0 {
		t.Fatalf("compareCurrentToTarget() = %d, want 0", compare)
	}

	if !compareExactVersion("0.4.0", "0.4.0") {
		t.Fatal("expected exact version comparison to match")
	}
	if compareExactVersion("0.4.0", "0.3.0") {
		t.Fatal("expected exact version comparison to differ")
	}
}
