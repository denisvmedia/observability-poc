package registry

import (
	"context"
	"net/url"

	"github.com/denisvmedia/observability-poc/models"
)

// SessionRegistry is the interface all storage backends must implement.
type SessionRegistry interface {
	// InsertBatch persists a batch of playback sessions.
	InsertBatch(ctx context.Context, sessions []models.PlaybackSession) error

	// ListVersions returns the distinct app_version values present in storage.
	ListVersions(ctx context.Context) ([]string, error)

	// GetKPIs returns aggregated KPIs for the requested versions.
	// Versions not present in storage are still returned with zero values.
	GetKPIs(ctx context.Context, versions []string) ([]models.VersionKPIs, error)
}

// Config is the raw DSN string used to connect to the database.
type Config string

// Parse parses the DSN into a *url.URL.
func (c Config) Parse() (*url.URL, error) {
	return url.Parse(string(c))
}

// SetFunc is the factory function signature every driver must satisfy.
type SetFunc func(Config) (SessionRegistry, error)

var drivers = make(map[string]SetFunc) //nolint:gochecknoglobals // package-level driver registry, populated via init()

// Register registers a driver factory under the given URL scheme.
// It panics if the scheme is already registered.
// It is intended to be called from package init functions.
func Register(scheme string, f SetFunc) {
	if _, ok := drivers[scheme]; ok {
		panic("registry: duplicate scheme: " + scheme)
	}

	drivers[scheme] = f
}

// GetRegistry parses the DSN and returns the matching factory function.
func GetRegistry(dsn string) (SetFunc, bool) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, false
	}

	f, ok := drivers[u.Scheme]

	return f, ok
}
