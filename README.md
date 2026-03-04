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
  ghcr.io/denisvmedia/observability-poc:latest
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

## Storage design

### Why ClickHouse

The workload is append-only writes and analytical reads: every dashboard query scans a large number of rows and applies conditional aggregations across several columns. That matches what column-oriented databases are built for — they only read the columns a query references and compress repeated values efficiently.

Alternatives considered:

| Option | Why not |
|---|---|
| **PostgreSQL** | Row-oriented storage reads entire rows for every aggregate scan. Works fine at small scale but gets expensive fast as row count grows. |
| **TimescaleDB** | Postgres extension with time-series optimisations. Closer fit than plain Postgres, but still row-oriented and lacks native conditional aggregate combinators. |
| **SQLite** | Fine for local use, not suitable for a multi-container deployment or concurrent writes from an ingestion service. |

ClickHouse was chosen because it is column-oriented, built for analytical aggregation, and its native `sumIf`/`countIf`/`avgIf` combinators let you compute all KPIs in a single table scan.

### Table engine — MergeTree

`MergeTree` is ClickHouse's general-purpose storage engine. Data is written in parts and merged in the background, keeping rows physically sorted by the `ORDER BY` key.

```sql
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (app_version, timestamp)
```

The `ORDER BY (app_version, timestamp)` choice means rows for the same version are stored together on disk. A query filtered to a specific `app_version` reads a contiguous range rather than scattering across the whole table.

`PARTITION BY toYYYYMM(timestamp)` puts each calendar month in a separate directory. Queries with a date range skip irrelevant month partitions entirely (partition pruning). Old data can be dropped with `ALTER TABLE … DROP PARTITION` — an instantaneous metadata operation, not a row-level delete.

**Alternatives within ClickHouse:**

| Engine | When to use instead |
|---|---|
| `ReplacingMergeTree` | Deduplicates rows with the same primary key during background merges. Useful if the source data can produce duplicates (re-uploads, retries). Reads may temporarily see duplicates until a merge runs, so queries need `FINAL` or an explicit `GROUP BY` dedup. See the [Known limitations](#known-limitations) section. |
| `SummingMergeTree` | Pre-aggregates numeric columns during merges. Useful when you only ever need totals, never individual rows. |
| `AggregatingMergeTree` | Stores partial aggregate states that ClickHouse merges incrementally. The most powerful option for pre-computed rollups, but requires using `AggregateFunction` column types and `*Merge` / `*State` combinators in queries. |

For this PoC the base `MergeTree` is the right choice: we need the raw session rows to remain queryable (to support any future ad-hoc metric), and deduplication is not yet required.

### ClickHouse SQL decisions

**Conditional aggregates instead of subqueries.** ClickHouse provides `-If` suffix combinators on every aggregate function: `sumIf(x, cond)`, `countIf(cond)`, `avgIf(x, cond)`. This keeps the entire KPI computation as a single `SELECT` with a single pass over the data, rather than a set of correlated subqueries or `CASE WHEN` expressions inside a `SUM`.

**Division-by-zero guard.** When a version has no rows matching a condition (e.g. no sessions with `attempts = 1`), `sumIf / countIf` would produce `inf` or `nan`. Every division is wrapped in an explicit guard:

```sql
if(countIf(attempts = 1) > 0,
   sumIf(vsf, attempts = 1) / countIf(attempts = 1),
   0) AS vsf_rate
```

**`LowCardinality(String)` for categorical columns.** `app_version`, `player_version`, and `player_name` have a small number of distinct values. `LowCardinality` encodes them as a dictionary of integers internally, reducing storage size and allowing ClickHouse to operate on integer codes during scans rather than comparing full strings. Not applied to `uuid` because it is high-cardinality by definition.

**`count()` returns `UInt64`.** ClickHouse's aggregate `count()` always returns an unsigned 64-bit integer. The corresponding Go struct field must be `uint64`, not `int64` — the driver rejects the conversion at runtime. This is a common gotcha when moving from Postgres where counts come back as signed integers.

## Known limitations

**No deduplication on re-upload.** Uploading the same XLSX file twice inserts duplicate rows — there is no uniqueness constraint in the current schema. For a PoC this is acceptable because uploads are manual and intentional.

If deduplication becomes necessary, two paths:

- **`ReplacingMergeTree`** — change the table engine. ClickHouse will deduplicate rows sharing the same `ORDER BY` key during background merges. Cheap to enable, but merges are asynchronous, so a query immediately after an insert may still see duplicates. Adding `FINAL` to the query forces synchronous dedup at read time with a performance cost.

- **Application-level check** — before inserting a batch, query for existing `(uuid, timestamp)` pairs and filter them from the batch. Gives an immediate guarantee with no schema change, but adds a round-trip per batch and couples the ingestion service to the data model more tightly.

## Development

```
make test       # Go + Vitest
make lint       # nolintguard → qtlint → golangci-lint + eslint + stylelint
make build      # frontend + backend with embedded UI
make help       # full list of targets
```

