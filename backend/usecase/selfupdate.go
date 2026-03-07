package usecase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lugassawan/panen/backend/domain/shared"
	"github.com/lugassawan/panen/backend/infra/platform"
	"github.com/lugassawan/panen/backend/infra/updater"
)

const (
	// EventUpdateProgress is the Wails event name for update progress.
	EventUpdateProgress = "update:progress"

	checksumFileName = "SHA256SUMS.txt"
)

// UpdateProgress is the payload emitted with EventUpdateProgress.
type UpdateProgress struct {
	State           string `json:"state"`
	DownloadedBytes int64  `json:"downloadedBytes"`
	TotalBytes      int64  `json:"totalBytes"`
	Version         string `json:"version"`
	Error           string `json:"error,omitempty"`
}

// AssetDownloader downloads a release asset.
type AssetDownloader interface {
	Download(
		ctx context.Context,
		url, destPath string,
		progressFn func(downloaded, total int64),
	) error
}

// ChecksumVerifier verifies file integrity.
type ChecksumVerifier interface {
	Verify(filePath, expectedHash string) error
}

// ArchiveExtractor extracts archive files.
type ArchiveExtractor interface {
	Extract(archivePath, destDir string) error
}

// PlatformInstaller handles platform-specific installation.
type PlatformInstaller interface {
	ArchiveName() string
	InstallPath() (string, error)
	Install(extractedDir, installPath string) error
	Rollback(installPath string) error
	CleanupBackup(installPath string) error
}

// SelfUpdateService orchestrates the full self-update flow.
type SelfUpdateService struct {
	checker    ReleaseChecker
	downloader AssetDownloader
	verifier   ChecksumVerifier
	extractor  ArchiveExtractor
	installer  PlatformInstaller
	emitter    EventEmitter
	currentVer string
	mu         sync.Mutex
	cancelFn   context.CancelFunc
	inProgress atomic.Bool
}

// NewSelfUpdateService creates a new SelfUpdateService.
func NewSelfUpdateService(
	checker ReleaseChecker,
	downloader AssetDownloader,
	verifier ChecksumVerifier,
	extractor ArchiveExtractor,
	installer PlatformInstaller,
	emitter EventEmitter,
	currentVer string,
) *SelfUpdateService {
	return &SelfUpdateService{
		checker:    checker,
		downloader: downloader,
		verifier:   verifier,
		extractor:  extractor,
		installer:  installer,
		emitter:    emitter,
		currentVer: currentVer,
	}
}

// PerformUpdate executes the full download-verify-install flow.
func (s *SelfUpdateService) PerformUpdate(ctx context.Context) error {
	if !s.inProgress.CompareAndSwap(false, true) {
		return shared.ErrUpdateInProgress
	}
	defer s.inProgress.Store(false)

	ctx, cancel := context.WithCancel(ctx)
	s.mu.Lock()
	s.cancelFn = cancel
	s.mu.Unlock()
	defer func() {
		cancel()
		s.mu.Lock()
		s.cancelFn = nil
		s.mu.Unlock()
	}()

	// Fetch latest release info
	info, err := s.checker.LatestRelease(ctx)
	if err != nil {
		s.emitError(info, fmt.Errorf("check release: %w", err))
		return err
	}

	version := info.Version
	archiveName := s.installer.ArchiveName()

	// Find archive and checksum assets
	archiveURL, checksumURL, err := findAssets(
		info.Assets, archiveName,
	)
	if err != nil {
		s.emitError(info, err)
		return err
	}

	// Create temp directory
	tempDir, err := s.createTempDir()
	if err != nil {
		s.emitError(info, err)
		return err
	}
	defer os.RemoveAll(tempDir)

	// Download checksum file
	s.emitProgress("downloading", 0, 0, version)
	checksumPath := filepath.Join(tempDir, checksumFileName)
	if err := s.downloader.Download(
		ctx, checksumURL, checksumPath, nil,
	); err != nil {
		s.emitError(info, fmt.Errorf("download checksum: %w", err))
		return err
	}

	// Download archive with progress
	archivePath := filepath.Join(tempDir, archiveName)
	if err := s.downloader.Download(
		ctx, archiveURL, archivePath,
		func(downloaded, total int64) {
			s.emitProgress("downloading", downloaded, total, version)
		},
	); err != nil {
		s.emitError(info, fmt.Errorf("download archive: %w", err))
		return err
	}

	// Parse and verify checksum
	s.emitProgress("verifying", 0, 0, version)
	checksumData, err := os.ReadFile(checksumPath)
	if err != nil {
		s.emitError(info, fmt.Errorf("read checksum file: %w", err))
		return err
	}

	expectedHash, err := updater.ParseChecksumFile(checksumData, archiveName)
	if err != nil {
		s.emitError(info, err)
		return err
	}

	if err := s.verifier.Verify(archivePath, expectedHash); err != nil {
		s.emitError(info, err)
		return err
	}

	// Extract archive
	s.emitProgress("installing", 0, 0, version)
	extractDir := filepath.Join(tempDir, "extracted")
	if err := os.MkdirAll(extractDir, 0o750); err != nil {
		s.emitError(info, fmt.Errorf("create extract dir: %w", err))
		return err
	}

	if err := s.extractor.Extract(archivePath, extractDir); err != nil {
		s.emitError(info, fmt.Errorf("extract archive: %w", err))
		return err
	}

	// Install
	installPath, err := s.installer.InstallPath()
	if err != nil {
		s.emitError(info, fmt.Errorf("resolve install path: %w", err))
		return err
	}

	if err := s.installer.Install(extractDir, installPath); err != nil {
		s.emitError(info, fmt.Errorf("install: %w", err))
		_ = s.installer.Rollback(installPath)
		return err
	}

	s.emitProgress("ready", 0, 0, version)
	return nil
}

// Cancel aborts an in-progress update.
func (s *SelfUpdateService) Cancel() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancelFn != nil {
		s.cancelFn()
	}
}

// CleanupPreviousUpdate removes backup files and stale temp dirs
// from a previous successful update. Safe to call on startup.
// Errors from non-critical operations (path resolution, dir reads) are
// silently ignored since cleanup is best-effort.
func (s *SelfUpdateService) CleanupPreviousUpdate() {
	installPath, err := s.installer.InstallPath()
	if err != nil {
		return
	}
	_ = s.installer.CleanupBackup(installPath)

	dataDir, err := platform.DataDir()
	if err != nil {
		return
	}
	updatesDir := filepath.Join(dataDir, "updates")
	entries, err := os.ReadDir(updatesDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			_ = os.RemoveAll(filepath.Join(updatesDir, e.Name()))
		}
	}
}

func (s *SelfUpdateService) createTempDir() (string, error) {
	dataDir, err := platform.DataDir()
	if err != nil {
		return "", fmt.Errorf("resolve data dir: %w", err)
	}
	updatesDir := filepath.Join(dataDir, "updates")
	if err := os.MkdirAll(updatesDir, 0o750); err != nil {
		return "", fmt.Errorf("create updates dir: %w", err)
	}
	prefix := time.Now().Format("20060102-150405-")
	tempDir, err := os.MkdirTemp(updatesDir, prefix)
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}
	return tempDir, nil
}

func (s *SelfUpdateService) emitProgress(
	state string,
	downloaded, total int64,
	version string,
) {
	s.emitter.Emit(EventUpdateProgress, UpdateProgress{
		State:           state,
		DownloadedBytes: downloaded,
		TotalBytes:      total,
		Version:         version,
	})
}

func (s *SelfUpdateService) emitError(
	info *ReleaseInfo,
	err error,
) {
	version := ""
	if info != nil {
		version = info.Version
	}
	s.emitter.Emit(EventUpdateProgress, UpdateProgress{
		State:   "error",
		Version: version,
		Error:   err.Error(),
	})
}

func findAssets(
	assets []ReleaseAsset,
	archiveName string,
) (archiveURL, checksumURL string, err error) {
	for _, a := range assets {
		switch a.Name {
		case archiveName:
			archiveURL = a.DownloadURL
		case checksumFileName:
			checksumURL = a.DownloadURL
		}
	}
	if archiveURL == "" {
		return "", "", fmt.Errorf("archive %q not found in release", archiveName)
	}
	if checksumURL == "" {
		return "", "", fmt.Errorf(
			"%s not found in release", checksumFileName,
		)
	}
	return archiveURL, checksumURL, nil
}
