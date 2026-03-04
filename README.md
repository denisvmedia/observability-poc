# observability-poc

A small quality-of-experience dashboard for comparing playback session metrics across app versions. Ingests XLSX exports, stores them in ClickHouse, and shows a side-by-side KPI comparison in a Vue 3 frontend.

![Dashboard](.github/screenshots/dashboard.png)

## Stack

- Go + Chi (API)
- ClickHouse (storage)
- Vue 3 + PrimeVue + Pinia (frontend, embedded in the binary)

## Running

**Pre-built image from ghcr.io:**

```
docker run -p 8080:8080 \
  -e OBSERVABILITY_DB_DSN="clickhouse://user:password@host:9000/dbname" \
  ghcr.io/denisvmedia/observability-poc:latest run
```

Or with Docker Compose using the pre-built image, set `image: ghcr.io/denisvmedia/observability-poc:latest` instead of the `build` block in `docker-compose.yaml`.

**Build from source:**

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

## KPIs and scoring

Six metrics are computed per version from the raw session rows:

| Metric | Formula | Direction |
|---|---|---|
| VSF Rate | `sum(vsf where attempts=1) / count(attempts=1)` | lower is better |
| VPF Rate | `sum(vpf where plays=1) / count(plays=1)` | lower is better |
| CIRR | `sum(cirr where plays=1) / count(plays=1)` | lower is better |
| Avg VST | `avg(vst where attempts=1)` | lower is better |
| Play Rate | `sum(plays where attempts=1) / count(attempts=1)` | higher is better |
| Completion Rate | `sum(ended_plays where plays=1) / count(plays=1)` | higher is better |

**Winner selection:** each metric is scored independently. The version that wins more dimensions (out of 6) is declared the overall winner. Ties are possible.

**Alerts** fire when a version crosses these thresholds:

- VSF Rate > 5%
- VPF Rate > 5%
- CIRR > 10%
- Avg VST > 5s
- Session count < 100 (low statistical confidence)
- Session count = 0 (no data)

## Troubleshooting

**macOS: "cannot be opened because the developer cannot be verified"**

Binaries built locally are unsigned, so Gatekeeper blocks them. To remove the quarantine flag:

```
xattr -d com.apple.quarantine ./bin/observability
```

If that doesn't help (e.g. the attribute isn't there but it still won't run):

```
sudo spctl --master-disable
./bin/observability run
sudo spctl --master-enable
```

Or right-click the binary in Finder → Open → Open anyway.

More details: https://donatstudios.com/mac-terminal-run-unsigned-binaries

## Development

```
make test       # Go + Vitest
make lint       # nolintguard → qtlint → golangci-lint + eslint + stylelint
make build      # frontend + backend with embedded UI
make help       # full list of targets
```

