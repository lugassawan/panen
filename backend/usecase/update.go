package usecase

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// ReleaseInfo holds version information about a release.
type ReleaseInfo struct {
	Version    string
	ReleaseURL string
}

// ReleaseChecker fetches the latest release information.
type ReleaseChecker interface {
	LatestRelease(ctx context.Context) (*ReleaseInfo, error)
}

// UpdateResult holds the result of an update check.
type UpdateResult struct {
	Available  bool
	CurrentVer string
	LatestVer  string
	ReleaseURL string
}

// UpdateService checks for application updates.
type UpdateService struct {
	checker    ReleaseChecker
	currentVer string
}

// NewUpdateService creates a new UpdateService.
func NewUpdateService(checker ReleaseChecker, currentVer string) *UpdateService {
	return &UpdateService{checker: checker, currentVer: currentVer}
}

// Check fetches the latest release and compares it with the current version.
func (s *UpdateService) Check(ctx context.Context) (*UpdateResult, error) {
	info, err := s.checker.LatestRelease(ctx)
	if err != nil {
		return nil, fmt.Errorf("check for updates: %w", err)
	}

	available := compareVersions(s.currentVer, info.Version) < 0
	return &UpdateResult{
		Available:  available,
		CurrentVer: s.currentVer,
		LatestVer:  info.Version,
		ReleaseURL: info.ReleaseURL,
	}, nil
}

// CurrentVersion returns the embedded application version.
func (s *UpdateService) CurrentVersion() string {
	return s.currentVer
}

// compareVersions compares two semver strings.
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
// "dev" is always considered older than any real version.
func compareVersions(a, b string) int {
	if a == b {
		return 0
	}
	if a == "dev" {
		return -1
	}
	if b == "dev" {
		return 1
	}

	aParts := parseSemver(a)
	bParts := parseSemver(b)

	for i := range 3 {
		if aParts[i] < bParts[i] {
			return -1
		}
		if aParts[i] > bParts[i] {
			return 1
		}
	}
	return 0
}

// parseSemver extracts [major, minor, patch] from a version string.
func parseSemver(v string) [3]int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.SplitN(v, ".", 3)
	var result [3]int
	for i := range min(len(parts), 3) {
		n, _ := strconv.Atoi(parts[i])
		result[i] = n
	}
	return result
}
