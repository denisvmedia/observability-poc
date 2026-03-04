// Package version holds build-time variables injected via -ldflags.
package version

// These variables are set at build time via:
//
//	-X github.com/denisvmedia/observability-poc/internal/version.Version=...
//	-X github.com/denisvmedia/observability-poc/internal/version.Commit=...
//	-X github.com/denisvmedia/observability-poc/internal/version.Date=...
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)
