// Package clickhouse provides a ClickHouse-backed implementation of registry.SessionRegistry.
package clickhouse

import (
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

var _ registry.SessionRegistry = (*Registry)(nil)

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
