package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Extractor struct{}

func (Extractor) ExtractBinary(archivePath string, binaryName string, archiveType string, dstDir string) (string, error) {
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return "", fmt.Errorf("create extraction directory %q: %w", dstDir, err)
	}
	binaryName = filepath.Base(strings.TrimSpace(binaryName))
	archiveType = strings.ToLower(strings.TrimSpace(archiveType))
	if binaryName == "" {
		return "", fmt.Errorf("binary name is required")
	}

	switch archiveType {
	case "binary", "exe", "":
		return extractSingleBinary(archivePath, filepath.Join(dstDir, binaryName))
	case "zip":
		return extractFromZip(archivePath, binaryName, dstDir)
	case "tar.gz", "tgz":
		return extractFromTarGz(archivePath, binaryName, dstDir)
	default:
		return "", fmt.Errorf("unsupported archive type %q", archiveType)
	}
}

func extractSingleBinary(source string, dst string) (string, error) {
	if err := copyFile(source, dst); err != nil {
		return "", err
	}
	return dst, nil
}

func extractFromZip(archivePath string, binaryName string, dstDir string) (string, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("open zip archive %q: %w", archivePath, err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		if file.FileInfo().IsDir() || filepath.Base(file.Name) != binaryName {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return "", fmt.Errorf("open %q in %q: %w", file.Name, archivePath, err)
		}
		defer rc.Close()
		target := filepath.Join(dstDir, binaryName)
		if err := writeExtractedFile(target, rc, file.Mode()); err != nil {
			return "", err
		}
		return target, nil
	}
	return "", fmt.Errorf("binary %q not found in archive %q", binaryName, archivePath)
}

func extractFromTarGz(archivePath string, binaryName string, dstDir string) (string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return "", fmt.Errorf("open archive %q: %w", archivePath, err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("open gzip reader for %q: %w", archivePath, err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("read tar entry from %q: %w", archivePath, err)
		}
		if header.FileInfo().IsDir() || filepath.Base(header.Name) != binaryName {
			continue
		}
		target := filepath.Join(dstDir, binaryName)
		if err := writeExtractedFile(target, tarReader, header.FileInfo().Mode()); err != nil {
			return "", err
		}
		return target, nil
	}
	return "", fmt.Errorf("binary %q not found in archive %q", binaryName, archivePath)
}

func writeExtractedFile(target string, source io.Reader, mode os.FileMode) error {
	tempFile, err := os.CreateTemp(filepath.Dir(target), ".boottree-extract-*")
	if err != nil {
		return fmt.Errorf("create temp extracted file for %q: %w", target, err)
	}
	tempName := tempFile.Name()
	defer func() { _ = os.Remove(tempName) }()

	if _, err := io.Copy(tempFile, source); err != nil {
		tempFile.Close()
		return fmt.Errorf("write extracted binary to %q: %w", tempName, err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close extracted binary %q: %w", tempName, err)
	}
	if mode == 0 {
		mode = 0o755
	}
	if err := os.Chmod(tempName, mode); err != nil {
		return fmt.Errorf("chmod extracted binary %q: %w", tempName, err)
	}
	if err := os.Rename(tempName, target); err != nil {
		return fmt.Errorf("move extracted binary to %q: %w", target, err)
	}
	return nil
}
