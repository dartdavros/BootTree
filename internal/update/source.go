package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type ManifestClient struct {
	HTTPClient *http.Client
}

func (c ManifestClient) Fetch(ctx context.Context, manifestURL string) (Manifest, error) {
	payload, err := readSource(ctx, manifestURL, c.httpClient())
	if err != nil {
		return Manifest{}, err
	}

	var manifest Manifest
	if err := json.Unmarshal(payload, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("parse manifest %q: %w", manifestURL, err)
	}

	normalizeManifest(&manifest)
	if len(manifest.Releases) == 0 {
		return Manifest{}, fmt.Errorf("manifest %q does not contain any releases", manifestURL)
	}
	return manifest, nil
}

func (c ManifestClient) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

func normalizeManifest(manifest *Manifest) {
	if manifest == nil {
		return
	}
	manifest.Channel = strings.TrimSpace(manifest.Channel)
	manifest.Latest = normalizeVersion(manifest.Latest)

	for i := range manifest.Releases {
		release := &manifest.Releases[i]
		release.Version = normalizeVersion(release.Version)
		for j := range release.Assets {
			normalizeAsset(&release.Assets[j])
		}
	}

	if len(manifest.Releases) == 0 && len(manifest.Assets) > 0 {
		release := Release{Version: manifest.Latest, PublishedAt: manifest.PublishedAt, Assets: append([]Asset(nil), manifest.Assets...)}
		for j := range release.Assets {
			normalizeAsset(&release.Assets[j])
		}
		manifest.Releases = []Release{release}
	}

	if manifest.Latest == "" && len(manifest.Releases) > 0 {
		manifest.Latest = manifest.Releases[0].Version
	}
}

func normalizeAsset(asset *Asset) {
	if asset == nil {
		return
	}
	asset.OS = strings.TrimSpace(asset.OS)
	asset.Arch = strings.TrimSpace(asset.Arch)
	asset.URL = strings.TrimSpace(asset.URL)
	asset.SHA256 = strings.ToLower(strings.TrimSpace(asset.SHA256))
	asset.Archive = strings.ToLower(strings.TrimSpace(asset.Archive))
	asset.Binary = strings.TrimSpace(asset.Binary)
}

func readSource(ctx context.Context, raw string, client *http.Client) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("manifest URL is required")
	}

	parsed, err := url.Parse(raw)
	if err == nil && parsed.Scheme != "" {
		switch parsed.Scheme {
		case "https":
			return readHTTPS(ctx, raw, client)
		case "file":
			return os.ReadFile(filepath.Clean(parsed.Path))
		default:
			return nil, fmt.Errorf("unsupported manifest scheme %q", parsed.Scheme)
		}
	}
	return os.ReadFile(filepath.Clean(raw))
}

func readHTTPS(ctx context.Context, raw string, client *http.Client) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return nil, fmt.Errorf("create request for %q: %w", raw, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download %q: %w", raw, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %q: unexpected status %s", raw, resp.Status)
	}
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body for %q: %w", raw, err)
	}
	return payload, nil
}
