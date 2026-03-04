package backend

// version is set at build time via -ldflags "-X github.com/lugassawan/panen/backend.version=X.Y.Z".
var version = "dev"

// Version returns the application version string.
func Version() string {
	return version
}
