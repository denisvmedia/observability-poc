package main

import (
	"log/slog"
	"os"

	_ "github.com/denisvmedia/observability-poc/registry/clickhouse"
)

func main() {
	if os.Getenv("LOG_FORMAT") == "json" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	}

	if err := execute(); err != nil {
		//revive:disable-next-line:deep-exit
		os.Exit(1)
	}
}
