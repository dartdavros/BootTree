package update

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestManifestClientFetch_HTTPS(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/manifest.json" {
			http.NotFound(w, r)
			return
		}
		_, _ = fmt.Fprint(w, `{
			"channel": "stable",
			"latest": "v0.4.0",
			"releases": [{
				"version": "v0.4.0",
				"assets": [{
					"os": "linux",
					"arch": "amd64",
					"url": "https://example.test/boottree_0.4.0_linux_amd64.tar.gz",
					"sha256": "abc123",
					"archive": "tar.gz",
					"binary": "boottree"
				}]
			}]
		}`)
	}))
	defer server.Close()

	client := ManifestClient{HTTPClient: server.Client()}
	manifest, err := client.Fetch(context.Background(), server.URL+"/manifest.json")
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if manifest.Channel != "stable" {
		t.Fatalf("manifest.Channel = %q, want %q", manifest.Channel, "stable")
	}
	if manifest.Latest != "0.4.0" {
		t.Fatalf("manifest.Latest = %q, want %q", manifest.Latest, "0.4.0")
	}
}

func TestManifestClientFetch_RejectsNonHTTPS(t *testing.T) {
	cases := []string{
		"C:/temp/manifest.json",
		`C:\temp\manifest.json`,
		"./dist/manifest.json",
		"file:///tmp/manifest.json",
		"http://example.test/manifest.json",
	}

	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			_, err := (ManifestClient{}).Fetch(context.Background(), raw)
			if err == nil {
				t.Fatalf("Fetch() error = nil, want non-nil")
			}
		})
	}
}

func TestNormalizeManifest_LegacyAssetsFallback(t *testing.T) {
	manifest := Manifest{
		Latest: "v0.6.0",
		Assets: []Asset{{
			OS:      "linux",
			Arch:    "amd64",
			URL:     "https://example.test/boottree_0.6.0_linux_amd64.tar.gz",
			SHA256:  "ABC789",
			Archive: "TAR.GZ",
			Binary:  "boottree",
		}},
	}

	normalizeManifest(&manifest)

	if len(manifest.Releases) != 1 {
		t.Fatalf("len(manifest.Releases) = %d, want 1", len(manifest.Releases))
	}
	if manifest.Releases[0].Version != "0.6.0" {
		t.Fatalf("release.Version = %q, want %q", manifest.Releases[0].Version, "0.6.0")
	}
	if manifest.Releases[0].Assets[0].SHA256 != "abc789" {
		t.Fatalf("release asset SHA256 = %q, want %q", manifest.Releases[0].Assets[0].SHA256, "abc789")
	}
}
