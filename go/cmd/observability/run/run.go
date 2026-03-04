package run

import (
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

func runServer(cfg *config) error {
	slog.Info("Starting server", "addr", cfg.addr, "db-dsn", cfg.dbDSN)

	setFunc, ok := registry.GetRegistry(cfg.dbDSN)
	if !ok {
		return fmt.Errorf("run: unsupported database scheme in DSN: %s", cfg.dbDSN)
	}

	reg, err := setFunc(registry.Config(cfg.dbDSN))
	if err != nil {
		return fmt.Errorf("run: connect to database: %w", err)
	}

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
