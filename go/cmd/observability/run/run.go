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
)

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

	cmd.Flags().StringVar(&cfg.addr, "addr", ":8080", "Bind address for the server")
	cmd.Flags().StringVar(&cfg.dbDSN, "db-dsn", "memory://", "Database DSN")

	return cmd
}

func runServer(cfg *config) error {
	slog.Info("Starting server", "addr", cfg.addr, "db-dsn", cfg.dbDSN)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthzHandler)

	srv := &http.Server{
		Addr:              cfg.addr,
		Handler:           mux,
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

func healthzHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, "ok")
}
