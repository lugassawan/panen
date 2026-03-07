package updater

// PlatformInstaller provides platform-specific installation logic
// for the self-update flow.
type PlatformInstaller interface {
	// ArchiveName returns the platform-specific archive filename
	// (e.g. "panen-darwin-universal.zip").
	ArchiveName() string

	// InstallPath returns the path to the currently installed application.
	InstallPath() (string, error)

	// Install replaces the current installation at installPath with
	// the contents of extractedDir. Creates a .backup internally.
	Install(extractedDir, installPath string) error

	// Rollback restores the .backup created by Install.
	Rollback(installPath string) error

	// CleanupBackup removes the .backup from a previous successful update.
	CleanupBackup(installPath string) error
}
