package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/lugassawan/panen/backend/domain/shared"
)

// Mock implementations

type mockDownloader struct {
	downloadFn func(ctx context.Context, url, destPath string, progressFn func(int64, int64)) error
}

func (m *mockDownloader) Download(
	ctx context.Context,
	url, destPath string,
	progressFn func(int64, int64),
) error {
	if m.downloadFn != nil {
		return m.downloadFn(ctx, url, destPath, progressFn)
	}
	return nil
}

type mockVerifier struct {
	verifyFn func(filePath, expectedHash string) error
}

func (m *mockVerifier) Verify(filePath, expectedHash string) error {
	if m.verifyFn != nil {
		return m.verifyFn(filePath, expectedHash)
	}
	return nil
}

type mockExtractor struct {
	extractFn func(archivePath, destDir string) error
}

func (m *mockExtractor) Extract(archivePath, destDir string) error {
	if m.extractFn != nil {
		return m.extractFn(archivePath, destDir)
	}
	return nil
}

type mockInstaller struct {
	archiveName    string
	installPath    string
	installErr     error
	rollbackCalled bool
	cleanupCalled  bool
	installPathErr error
	mu             sync.Mutex
}

func (m *mockInstaller) ArchiveName() string {
	return m.archiveName
}

func (m *mockInstaller) InstallPath() (string, error) {
	return m.installPath, m.installPathErr
}

func (m *mockInstaller) Install(_, _ string) error {
	return m.installErr
}

func (m *mockInstaller) Rollback(_ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rollbackCalled = true
	return nil
}

func (m *mockInstaller) CleanupBackup(_ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cleanupCalled = true
	return nil
}

type selfUpdateChecker struct {
	info *ReleaseInfo
	err  error
}

func (c *selfUpdateChecker) LatestRelease(
	_ context.Context,
) (*ReleaseInfo, error) {
	return c.info, c.err
}

func TestSelfUpdateHappyPath(t *testing.T) {
	tmpDir := t.TempDir()

	archiveContent := []byte("fake archive data")
	h := sha256.Sum256(archiveContent)
	hash := hex.EncodeToString(h[:])
	checksumContent := hash + "  panen-test.zip\n"

	checker := &selfUpdateChecker{
		info: &ReleaseInfo{
			Version:    "1.0.0",
			ReleaseURL: "https://example.com",
			Assets: []ReleaseAsset{
				{
					Name:        "panen-test.zip",
					DownloadURL: "https://github.com/lugassawan/panen/releases/download/v1.0.0/panen-test.zip",
					Size:        int64(len(archiveContent)),
				},
				{
					Name:        "SHA256SUMS.txt",
					DownloadURL: "https://github.com/lugassawan/panen/releases/download/v1.0.0/SHA256SUMS.txt",
					Size:        int64(len(checksumContent)),
				},
			},
		},
	}

	downloader := &mockDownloader{
		downloadFn: func(_ context.Context, url, destPath string, _ func(int64, int64)) error {
			if filepath.Base(destPath) == "SHA256SUMS.txt" {
				return os.WriteFile(destPath, []byte(checksumContent), 0o644)
			}
			return os.WriteFile(destPath, archiveContent, 0o644)
		},
	}

	verifier := &mockVerifier{}
	extractor := &mockExtractor{}
	installer := &mockInstaller{
		archiveName: "panen-test.zip",
		installPath: filepath.Join(tmpDir, "panen.app"),
	}
	emitter := newMockEventEmitter()

	svc := NewSelfUpdateService(
		checker, downloader, verifier, extractor,
		installer, emitter, "0.9.0",
	)

	if err := svc.PerformUpdate(context.Background()); err != nil {
		t.Fatalf("PerformUpdate: %v", err)
	}

	events := emitter.eventsByName(EventUpdateProgress)
	if len(events) == 0 {
		t.Fatal("no progress events emitted")
	}

	// Verify states: should include downloading, verifying, installing, ready
	states := make(map[string]bool)
	for _, e := range events {
		if p, ok := e.data.(UpdateProgress); ok {
			states[p.State] = true
		}
	}
	for _, want := range []string{
		"downloading", "verifying", "installing", "ready",
	} {
		if !states[want] {
			t.Errorf("missing state %q in events", want)
		}
	}
}

func TestSelfUpdateConcurrentRejected(t *testing.T) {
	checker := &selfUpdateChecker{
		info: &ReleaseInfo{
			Version: "1.0.0",
			Assets: []ReleaseAsset{
				{
					Name:        "panen-test.zip",
					DownloadURL: "https://github.com/lugassawan/panen/releases/download/v1.0.0/test.zip",
				},
				{
					Name:        "SHA256SUMS.txt",
					DownloadURL: "https://github.com/lugassawan/panen/releases/download/v1.0.0/SHA256SUMS.txt",
				},
			},
		},
	}

	// Block the download so we can test concurrency
	started := make(chan struct{})
	block := make(chan struct{})
	downloader := &mockDownloader{
		downloadFn: func(ctx context.Context, _, destPath string, _ func(int64, int64)) error {
			// Write dummy file so the flow doesn't fail on missing file
			_ = os.WriteFile(destPath, []byte("data"), 0o644)
			select {
			case started <- struct{}{}:
			default:
			}
			select {
			case <-block:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	}

	svc := NewSelfUpdateService(
		checker, downloader, &mockVerifier{}, &mockExtractor{},
		&mockInstaller{
			archiveName: "panen-test.zip",
			installPath: t.TempDir(),
		},
		newMockEventEmitter(), "0.9.0",
	)

	go func() {
		_ = svc.PerformUpdate(context.Background())
	}()

	<-started

	err := svc.PerformUpdate(context.Background())
	if !errors.Is(err, shared.ErrUpdateInProgress) {
		t.Fatalf("expected ErrUpdateInProgress, got %v", err)
	}

	close(block)
}

func TestSelfUpdateChecksumMismatch(t *testing.T) {
	checker := &selfUpdateChecker{
		info: &ReleaseInfo{
			Version: "1.0.0",
			Assets: []ReleaseAsset{
				{
					Name:        "panen-test.zip",
					DownloadURL: "https://github.com/lugassawan/panen/releases/download/v1.0.0/test.zip",
				},
				{
					Name:        "SHA256SUMS.txt",
					DownloadURL: "https://github.com/lugassawan/panen/releases/download/v1.0.0/SHA256SUMS.txt",
				},
			},
		},
	}

	downloader := &mockDownloader{
		downloadFn: func(_ context.Context, _, destPath string, _ func(int64, int64)) error {
			if filepath.Base(destPath) == "SHA256SUMS.txt" {
				return os.WriteFile(
					destPath,
					[]byte("abc123  panen-test.zip\n"),
					0o644,
				)
			}
			return os.WriteFile(destPath, []byte("data"), 0o644)
		},
	}

	verifier := &mockVerifier{
		verifyFn: func(_, _ string) error {
			return shared.ErrChecksumMismatch
		},
	}

	emitter := newMockEventEmitter()
	svc := NewSelfUpdateService(
		checker, downloader, verifier, &mockExtractor{},
		&mockInstaller{
			archiveName: "panen-test.zip",
			installPath: t.TempDir(),
		},
		emitter, "0.9.0",
	)

	err := svc.PerformUpdate(context.Background())
	if !errors.Is(err, shared.ErrChecksumMismatch) {
		t.Fatalf("expected ErrChecksumMismatch, got %v", err)
	}

	// Verify error event was emitted
	events := emitter.eventsByName(EventUpdateProgress)
	var hasError bool
	for _, e := range events {
		if p, ok := e.data.(UpdateProgress); ok && p.State == "error" {
			hasError = true
		}
	}
	if !hasError {
		t.Error("expected error event")
	}
}

func TestSelfUpdateInstallFailureTriggersRollback(t *testing.T) {
	checker := &selfUpdateChecker{
		info: &ReleaseInfo{
			Version: "1.0.0",
			Assets: []ReleaseAsset{
				{
					Name:        "panen-test.zip",
					DownloadURL: "https://github.com/lugassawan/panen/releases/download/v1.0.0/test.zip",
				},
				{
					Name:        "SHA256SUMS.txt",
					DownloadURL: "https://github.com/lugassawan/panen/releases/download/v1.0.0/SHA256SUMS.txt",
				},
			},
		},
	}

	downloader := &mockDownloader{
		downloadFn: func(_ context.Context, _, destPath string, _ func(int64, int64)) error {
			if filepath.Base(destPath) == "SHA256SUMS.txt" {
				return os.WriteFile(
					destPath,
					[]byte("abc123  panen-test.zip\n"),
					0o644,
				)
			}
			return os.WriteFile(destPath, []byte("data"), 0o644)
		},
	}

	installer := &mockInstaller{
		archiveName: "panen-test.zip",
		installPath: t.TempDir(),
		installErr:  errors.New("install failed"),
	}

	svc := NewSelfUpdateService(
		checker, downloader, &mockVerifier{}, &mockExtractor{},
		installer, newMockEventEmitter(), "0.9.0",
	)

	err := svc.PerformUpdate(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	installer.mu.Lock()
	defer installer.mu.Unlock()
	if !installer.rollbackCalled {
		t.Error("expected rollback to be called")
	}
}

func TestSelfUpdateMissingArchive(t *testing.T) {
	checker := &selfUpdateChecker{
		info: &ReleaseInfo{
			Version: "1.0.0",
			Assets:  []ReleaseAsset{},
		},
	}

	svc := NewSelfUpdateService(
		checker, &mockDownloader{}, &mockVerifier{}, &mockExtractor{},
		&mockInstaller{archiveName: "panen-test.zip"},
		newMockEventEmitter(), "0.9.0",
	)

	err := svc.PerformUpdate(context.Background())
	if err == nil {
		t.Fatal("expected error for missing archive")
	}
}

func TestFindAssets(t *testing.T) {
	tests := []struct {
		name        string
		assets      []ReleaseAsset
		archiveName string
		wantErr     bool
	}{
		{
			name: "both present",
			assets: []ReleaseAsset{
				{Name: "panen-test.zip", DownloadURL: "url1"},
				{Name: "SHA256SUMS.txt", DownloadURL: "url2"},
			},
			archiveName: "panen-test.zip",
		},
		{
			name: "archive missing",
			assets: []ReleaseAsset{
				{Name: "SHA256SUMS.txt", DownloadURL: "url2"},
			},
			archiveName: "panen-test.zip",
			wantErr:     true,
		},
		{
			name: "checksum missing",
			assets: []ReleaseAsset{
				{Name: "panen-test.zip", DownloadURL: "url1"},
			},
			archiveName: "panen-test.zip",
			wantErr:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := findAssets(tc.assets, tc.archiveName)
			if tc.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
