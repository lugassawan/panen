package updater

import (
	"context"
	"fmt"
	"os"

	"github.com/lugassawan/panen/backend/infra/github"
)

// Downloader wraps a GitHub client to download release assets to disk.
type Downloader struct {
	client *github.Client
}

// NewDownloader creates a Downloader backed by the given GitHub client.
func NewDownloader(client *github.Client) *Downloader {
	return &Downloader{client: client}
}

// Download fetches the asset at url and writes it to destPath.
// Partial files are removed on error.
func (d *Downloader) Download(
	ctx context.Context,
	url, destPath string,
	progressFn func(downloaded, total int64),
) error {
	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	if err := d.client.DownloadAsset(ctx, url, f, progressFn); err != nil {
		_ = f.Close()
		_ = os.Remove(destPath)
		return err
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(destPath)
		return fmt.Errorf("sync file: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(destPath)
		return fmt.Errorf("close file: %w", err)
	}
	return nil
}
