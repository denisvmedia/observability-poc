module github.com/denisvmedia/observability-poc

require github.com/denisvmedia/observability-poc/frontend v0.0.0

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
)

replace github.com/denisvmedia/observability-poc/frontend => ../frontend

go 1.26.0
