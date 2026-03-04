package run

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/denisvmedia/observability-poc/apiserver"
	"github.com/denisvmedia/observability-poc/registry"
)

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return def
}

type config struct {
	addr  string
	dbDSN string
}

// New returns the run sub-command.
func New() *cobra.Command {
	cfg := &config{}

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run the observability server",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runServer(cfg)
		},
	}

	cmd.Flags().StringVar(&cfg.addr, "addr", envOrDefault("OBSERVABILITY_ADDR", ":8080"), "Bind address for the server (env: OBSERVABILITY_ADDR)")
	cmd.Flags().StringVar(&cfg.dbDSN, "db-dsn", envOrDefault("OBSERVABILITY_DB_DSN", "clickhouse://localhost:9000/observability"), "Database DSN (env: OBSERVABILITY_DB_DSN)")

	return cmd
}

const (
	dbMaxRetries    = 10
	dbRetryInterval = 3 * time.Second
)

// connectAndMigrate opens a registry connection and runs schema migrations.
func connectAndMigrate(ctx context.Context, dsn string) (registry.SessionRegistry, error) {
	setFunc, ok := registry.GetRegistry(dsn)
	if !ok {
		return nil, fmt.Errorf("unsupported database scheme in DSN: %s", dsn)
	}

	reg, err := setFunc(registry.Config(dsn))
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	if m, ok := reg.(registry.Migrator); ok {
		if err = m.Migrate(ctx); err != nil {
			return nil, fmt.Errorf("migrate database: %w", err)
		}
	}

	return reg, nil
}

// connectWithRetry retries connectAndMigrate with a fixed delay until the
// database is ready or the maximum number of attempts is reached.
func connectWithRetry(ctx context.Context, dsn string) (registry.SessionRegistry, error) {
	var err error

	for attempt := 1; attempt <= dbMaxRetries; attempt++ {
		var reg registry.SessionRegistry

		reg, err = connectAndMigrate(ctx, dsn)
		if err == nil {
			return reg, nil
		}

		if attempt == dbMaxRetries {
			break
		}

		slog.Warn("Database not ready, retrying",
			"attempt", attempt,
			"max", dbMaxRetries,
			"delay", dbRetryInterval,
			"error", err,
		)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(dbRetryInterval):
		}
	}

	return nil, fmt.Errorf("database unavailable after %d attempts: %w", dbMaxRetries, err)
}

func runServer(cfg *config) error {
	slog.Info("Starting server", "addr", cfg.addr, "db-dsn", cfg.dbDSN)

	slog.Info("Connecting to database and running migrations")

	reg, err := connectWithRetry(context.Background(), cfg.dbDSN)
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}

	slog.Info("Database ready")

	srv := &http.Server{
		Addr:              cfg.addr,
		Handler:           apiserver.New(reg),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil {
			slog.Error("Server failed to start", "error", err)
			return err
		}
	case <-sigCh:
		slog.Info("Shutting down server")
	}

	return srv.Close()
}
