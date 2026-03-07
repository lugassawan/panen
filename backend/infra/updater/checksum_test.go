package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/lugassawan/panen/backend/domain/shared"
)

func TestSHA256Verifier(t *testing.T) {
	dir := t.TempDir()
	content := []byte("hello checksum world")
	path := filepath.Join(dir, "test.bin")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	h := sha256.Sum256(content)
	correctHash := hex.EncodeToString(h[:])

	v := &SHA256Verifier{}

	t.Run("correct hash passes", func(t *testing.T) {
		if err := v.Verify(path, correctHash); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("wrong hash returns ErrChecksumMismatch", func(t *testing.T) {
		err := v.Verify(path, "0000000000000000000000000000000000000000000000000000000000000000")
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, shared.ErrChecksumMismatch) {
			t.Fatalf("expected ErrChecksumMismatch, got %v", err)
		}
	})

	t.Run("nonexistent file returns error", func(t *testing.T) {
		err := v.Verify(filepath.Join(dir, "nope"), correctHash)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestParseChecksumFile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		archiveName string
		wantHash    string
		wantErr     bool
	}{
		{
			name: "finds matching entry",
			content: "abc123  panen-darwin-universal.zip\n" +
				"def456  panen-linux-amd64.tar.gz\n",
			archiveName: "panen-linux-amd64.tar.gz",
			wantHash:    "def456",
		},
		{
			name:        "first entry",
			content:     "abc123  panen-darwin-universal.zip\n",
			archiveName: "panen-darwin-universal.zip",
			wantHash:    "abc123",
		},
		{
			name:        "not found",
			content:     "abc123  other.zip\n",
			archiveName: "panen-darwin-universal.zip",
			wantErr:     true,
		},
		{
			name:        "empty content",
			content:     "",
			archiveName: "anything",
			wantErr:     true,
		},
		{
			name:        "handles extra whitespace",
			content:     "  abc123  panen.zip  \n",
			archiveName: "panen.zip",
			wantHash:    "abc123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := ParseChecksumFile(
				[]byte(tc.content), tc.archiveName,
			)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if hash != tc.wantHash {
				t.Errorf("hash = %q, want %q", hash, tc.wantHash)
			}
		})
	}
}
