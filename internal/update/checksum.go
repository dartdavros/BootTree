package update

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

type ChecksumVerifier struct{}

func (ChecksumVerifier) VerifyFile(path string, expectedSHA256 string) error {
	expectedSHA256 = strings.ToLower(strings.TrimSpace(expectedSHA256))
	if expectedSHA256 == "" {
		return fmt.Errorf("expected SHA-256 checksum is required")
	}
	if _, err := hex.DecodeString(expectedSHA256); err != nil {
		return fmt.Errorf("invalid SHA-256 checksum %q: %w", expectedSHA256, err)
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file %q for checksum verification: %w", path, err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("hash file %q: %w", path, err)
	}
	actual := hex.EncodeToString(hash.Sum(nil))
	if actual != expectedSHA256 {
		return fmt.Errorf("checksum mismatch for %q: expected %s, got %s", path, expectedSHA256, actual)
	}
	return nil
}
