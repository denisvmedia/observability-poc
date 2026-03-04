// Package clickhouse provides a ClickHouse-backed implementation of registry.SessionRegistry.
package clickhouse

import (
	"context"
	"fmt"

	chdriver "github.com/ClickHouse/clickhouse-go/v2"

	"github.com/denisvmedia/observability-poc/registry"
)

func init() {
	registry.Register("clickhouse", func(cfg registry.Config) (registry.SessionRegistry, error) {
		return newRegistry(cfg)
	})
}

// Registry is the ClickHouse-backed SessionRegistry.
type Registry struct {
	conn chdriver.Conn
}

var (
	_ registry.SessionRegistry = (*Registry)(nil)
	_ registry.Migrator        = (*Registry)(nil)
)

// Migrate ensures the required schema exists in ClickHouse.
// It is idempotent and safe to call on every application startup.
func (r *Registry) Migrate(ctx context.Context) error {
	err := r.conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS playback_sessions
		(
		    timestamp      DateTime,
		    uuid           String,
		    app_version    LowCardinality(String),
		    player_version LowCardinality(String),
		    player_name    LowCardinality(String),
		    attempts       UInt8,
		    plays          UInt8,
		    ended_plays    UInt8,
		    vsf            Float64,
		    vpf            Float64,
		    cirr           Float64,
		    vst            Float64
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMM(timestamp)
		ORDER BY (app_version, timestamp)
	`)
	if err != nil {
		return fmt.Errorf("clickhouse: migrate: %w", err)
	}

	return nil
}

func newRegistry(cfg registry.Config) (*Registry, error) {
	opts, err := chdriver.ParseDSN(string(cfg))
	if err != nil {
		return nil, fmt.Errorf("clickhouse: parse DSN: %w", err)
	}

	conn, err := chdriver.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("clickhouse: open connection: %w", err)
	}

	return &Registry{conn: conn}, nil
}
