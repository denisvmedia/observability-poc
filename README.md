# observability-poc

A small quality-of-experience dashboard for comparing playback session metrics across app versions. Ingests XLSX exports, stores them in ClickHouse, and shows a side-by-side KPI comparison in a Vue 3 frontend.

## Stack

- Go + Chi (API)
- ClickHouse (storage)
- Vue 3 + PrimeVue + Pinia (frontend, embedded in the binary)

## Running

The easiest way is Docker Compose:

```
docker compose up --build
```

Open http://localhost:8080, upload an XLSX file, then go to /dashboard to compare versions.

For local development (requires a running ClickHouse):

```
make run-clickhouse   # starts CH in Docker, then runs the binary
```

## XLSX format

The file must have a header row with these columns (order doesn't matter, names are case-insensitive):

```
timestamp, uuid, app_version, player_version, player_name,
attempts, plays, ended_plays, vsf, vpf, cirr, vst
```

Timestamps can include timezone offset and microseconds (`2026-02-22 19:04:30.015208-05:00`). Floats can use either `.` or `,` as decimal separator.

## Development

```
make test       # Go + Vitest
make lint       # nolintguard → qtlint → golangci-lint + eslint + stylelint
make build      # frontend + backend with embedded UI
make help       # full list of targets
```

