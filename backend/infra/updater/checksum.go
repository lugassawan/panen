package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lugassawan/panen/backend/domain/shared"
)

// SHA256Verifier verifies file checksums using SHA-256.
type SHA256Verifier struct{}

// Verify computes the SHA-256 hash of filePath and compares it to expectedHash.
// Returns shared.ErrChecksumMismatch if they differ.
func (v *SHA256Verifier) Verify(filePath, expectedHash string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file for checksum: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("compute checksum: %w", err)
	}

	actual := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(actual, expectedHash) {
		return fmt.Errorf(
			"%w: expected %s, got %s",
			shared.ErrChecksumMismatch, expectedHash, actual,
		)
	}
	return nil
}

// ParseChecksumFile parses a SHA256SUMS.txt-style file and returns
// the hash for the given archiveName.
// Format: "<hex-hash>  <filename>\n" (two spaces between hash and name).
func ParseChecksumFile(content []byte, archiveName string) (string, error) {
	text := string(content)
	for line := range strings.SplitSeq(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: hash  filename (two spaces)
		hash, name, ok := strings.Cut(line, "  ")
		if !ok {
			continue
		}
		if strings.TrimSpace(name) == archiveName {
			return strings.TrimSpace(hash), nil
		}
	}
	return "", fmt.Errorf("checksum for %q not found", archiveName)
}
