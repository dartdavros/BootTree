package update

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Downloader struct {
	HTTPClient *http.Client
}

func (d Downloader) Download(ctx context.Context, sourceURL string, dst string) error {
	sourceURL = strings.TrimSpace(sourceURL)
	if sourceURL == "" {
		return fmt.Errorf("download URL is required")
	}
	if dst == "" {
		return fmt.Errorf("download destination is required")
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create download directory for %q: %w", dst, err)
	}

	parsed, err := url.Parse(sourceURL)
	if err == nil && parsed.Scheme != "" {
		switch parsed.Scheme {
		case "https":
			return d.downloadHTTPS(ctx, sourceURL, dst)
		case "file":
			return copyFile(filepath.Clean(parsed.Path), dst)
		default:
			return fmt.Errorf("unsupported download scheme %q", parsed.Scheme)
		}
	}
	return copyFile(filepath.Clean(sourceURL), dst)
}

func (d Downloader) downloadHTTPS(ctx context.Context, sourceURL string, dst string) error {
	client := d.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return fmt.Errorf("create request for %q: %w", sourceURL, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download %q: %w", sourceURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %q: unexpected status %s", sourceURL, resp.Status)
	}

	tempFile, err := os.CreateTemp(filepath.Dir(dst), ".boottree-download-*")
	if err != nil {
		return fmt.Errorf("create temp download file for %q: %w", dst, err)
	}
	tempName := tempFile.Name()
	defer func() { _ = os.Remove(tempName) }()

	if _, err := io.Copy(tempFile, resp.Body); err != nil {
		tempFile.Close()
		return fmt.Errorf("write download to %q: %w", tempName, err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close temp download %q: %w", tempName, err)
	}
	if err := os.Rename(tempName, dst); err != nil {
		return fmt.Errorf("move downloaded file to %q: %w", dst, err)
	}
	return nil
}

func copyFile(source string, dst string) error {
	in, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open source file %q: %w", source, err)
	}
	defer in.Close()

	tempFile, err := os.CreateTemp(filepath.Dir(dst), ".boottree-copy-*")
	if err != nil {
		return fmt.Errorf("create temp file for %q: %w", dst, err)
	}
	tempName := tempFile.Name()
	defer func() { _ = os.Remove(tempName) }()

	if _, err := io.Copy(tempFile, in); err != nil {
		tempFile.Close()
		return fmt.Errorf("copy %q to %q: %w", source, tempName, err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close temp file %q: %w", tempName, err)
	}
	if err := os.Rename(tempName, dst); err != nil {
		return fmt.Errorf("move %q to %q: %w", tempName, dst, err)
	}
	return nil
}
