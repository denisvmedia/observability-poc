# Interview Q&A — observability-poc

Questions an interviewer might ask about the design, implementation, and tradeoffs.

---

## Architecture & Design

**Q: Why ClickHouse instead of PostgreSQL or TimescaleDB?**

The data is append-only, read-heavy, and analytical — exactly the workload ClickHouse is built for. Each dashboard query scans tens of thousands of rows and applies conditional aggregations across multiple columns. ClickHouse stores data column-by-column, so it only reads the columns it needs and compresses them efficiently. A regular row store like Postgres would read full rows for each aggregate. TimescaleDB is a closer competitor but ClickHouse's native `sumIf`, `countIf`, and `avgIf` combinators let you express the entire KPI computation as a single pass over the data without subqueries.

---

**Q: What does `MergeTree()` mean and why did you pick it?**

`MergeTree` is ClickHouse's base storage engine. It keeps data sorted by the `ORDER BY` key and merges parts in the background. The order `(app_version, timestamp)` makes queries that filter by `app_version` fast because matching rows are physically adjacent. `PARTITION BY toYYYYMM(timestamp)` means each calendar month is stored in a separate directory — old months can be dropped cheaply and queries restricted to a date range skip irrelevant partitions entirely. For this workload the base engine is enough. `ReplacingMergeTree` would add eventual deduplication by primary key but at the cost of more complex merge semantics.

---

**Q: Describe the driver registration pattern in `registry/registry.go`.**

It mirrors the standard library's `database/sql` pattern. The `registry` package defines a `SessionRegistry` interface and a global map of `scheme → factory function`. Each backend (e.g. `clickhouse`) calls `registry.Register("clickhouse", ...)` from its `init()` function. The binary imports the driver package with a blank import (`_ "…/registry/clickhouse"`) purely for its side effect. This means you can add or remove backends without changing any core code, and in tests you register a lightweight in-memory driver instead.

---

**Q: What is the `Migrator` interface and why is it separate from `SessionRegistry`?**

Not every backend needs schema migration. An in-memory backend used in tests doesn't have a schema at all. By defining `Migrator` as an optional interface, the startup code uses a type assertion — `if m, ok := reg.(registry.Migrator); ok` — and only calls `Migrate` if the backend supports it. This keeps the core interface minimal and avoids forcing test doubles to implement a no-op method.

---

**Q: Why did you use Go build tags for frontend embedding instead of always including it?**

Running the server during development without the frontend lets you iterate on the backend without rebuilding the full Vue bundle. `go build -tags with_frontend` compiles the `frontend.go` file that contains the `//go:embed dist` directive; without the tag, `apiserver_without_frontend.go` provides a stub handler that returns a 404 or a redirect. The two files are mutually exclusive via build constraints. This also means `go test ./...` works without running `npm run build` first.

---

## ClickHouse & SQL

**Q: Your KPI query divides `sumIf` by `countIf`. What happens if no rows match the condition?**

Division by zero — ClickHouse returns `inf` or `nan` instead of an error. Go's `float64` can hold those values and they would propagate silently into the API response and the UI. The fix is to wrap each division in `if(countIf(cond) > 0, sum/count, 0)`. For `avgIf` the same guard applies because ClickHouse's behaviour on an empty set is version-dependent.

---

**Q: Why does `SessionCount` use `uint64` instead of `int64`?**

ClickHouse's `count()` returns a `UInt64` — an unsigned 64-bit integer. The Go driver refuses to scan a `UInt64` column into a `*int64` pointer at runtime with an explicit error. Using `uint64` in the Go struct aligns the types exactly and avoids the conversion panic. A session count can never be negative anyway, so `uint64` is semantically correct too.

---

**Q: What is `LowCardinality(String)` and when should you use it?**

`LowCardinality` is a dictionary encoding that maps repeated string values to small integers internally. It's efficient when a column has fewer than ~10,000 distinct values — `app_version`, `player_version`, and `player_name` all qualify. For a truly high-cardinality column like `uuid` it would hurt performance because the dictionary itself becomes large. The compression ratio and scan speed improve significantly for low-cardinality data because ClickHouse can operate on the integer dictionary codes rather than the full strings.

---

## Go Specifics

**Q: How does context propagation work across the stack?**

The HTTP handler receives a `context.Context` from the request (`r.Context()`). It passes that context to every downstream call: `reg.GetKPIs(ctx, ...)`, `ingestion.Ingest(ctx, ...)`, `reg.InsertBatch(ctx, ...)`. The ClickHouse driver uses the context to cancel in-flight queries if the client disconnects. In the ingestion loop, `flushBatch` checks `ctx.Err()` before each batch insert so a cancelled upload doesn't leave a partial write. The goroutine running the HTTP server is not given a context because `http.Server.Close()` handles its shutdown.

---

**Q: Why did you implement retry logic in the application rather than relying on `depends_on: condition: service_healthy` in docker-compose?**

`service_healthy` fires when ClickHouse's HTTP endpoint at port 8123 responds to `/ping`. But the application connects using the native binary protocol on port 9000, which initialises on a slightly different timeline inside ClickHouse. By the time the health check passes, port 9000 may still be in its startup sequence. Application-level retry with a backoff loop is more robust and works regardless of the orchestrator — bare Docker, Kubernetes, or running locally without Compose at all.

---

**Q: Explain the float epsilon comparison in `dimensionWinner`.**

Two `float64` values that are mathematically equal can differ by a tiny amount after floating-point arithmetic — especially after ClickHouse does conditional aggregation and the result crosses the wire as a binary float. Comparing with `==` could produce a spurious winner when the values should be a tie. Using `math.Abs(a-b) < 1e-9` (one billionth) treats values within that tolerance as equal. The threshold is chosen to be far above floating-point noise (~1e-15) but far below any meaningful difference between two real KPI measurements.

---

**Q: Walk me through what happens when a user uploads an XLSX file.**

1. The multipart form is parsed and the file handle passed to `ingestion.Ingest`.
2. `excelize.OpenReader` loads the file into memory (excelize does not stream from an `io.Reader`).
3. The header row is lowercased and mapped to column indices, so column order doesn't matter.
4. Each data row is parsed: timestamp (multiple layouts tried in order), floats (comma→dot normalisation), uint8 values.
5. Valid rows are accumulated in a 1,000-row buffer; malformed rows increment `RowsSkipped` and add an error message (capped at 50).
6. Each full buffer is sent to `InsertBatch` which uses a prepared batch statement — one round-trip per 1,000 rows.
7. A summary `{rows_inserted, rows_skipped, errors}` is returned as JSON.

---

**Q: What is the deduplication problem and how would you fix it in production?**

There is none: uploading the same file twice silently doubles every row. Two realistic fixes:

- **`ReplacingMergeTree`** — change the table engine so ClickHouse deduplicates rows with the same primary key during background merges. Reads may still see duplicates temporarily, so queries need a `FINAL` modifier or a `GROUP BY` dedup step. Works well for eventual consistency.
- **Application-level check** — before inserting a batch, query for existing `(uuid, timestamp)` pairs and filter them out. More code, but gives an immediate guarantee.

For a PoC neither is critical; the important thing is to be aware of it.

---

## Frontend & Integration

**Q: How is the Vue frontend embedded in the Go binary?**

The frontend is a separate Go module (`frontend/go.mod`) with a single file `frontend.go` that has `//go:embed dist` and exposes a function returning the embedded `fs.FS`. The main module imports it as a local replace directive. At build time, `npm run build` outputs files into `frontend/dist/`, then `go build -tags with_frontend` embeds that directory. The `http.FileServer(http.FS(fsys))` serves assets; requests for unknown paths fall back to `index.html` for Vue Router's client-side routing.

---

**Q: Why did the original SPA fallback use `httptest.NewRecorder`, and what's wrong with it?**

The idea was: let the file server try to serve the path; if it returns 404, serve `index.html` instead. Using `httptest.NewRecorder` as a probe works, but it buffers the entire response in memory and processes every request twice — once into the recorder and once into the real writer. The clean alternative is `fs.Stat`: if the file exists in the embedded FS, delegate to the file server; otherwise return `index.html` directly. One stat call, no buffering, no proxy overhead.

---

**Q: What was wrong with building JSON error responses by string concatenation?**

```go
http.Error(w, `{"error":"ingestion failed: `+err.Error()+`"}`, ...)
```

If `err.Error()` contains a double-quote, backslash, or newline, the resulting string is invalid JSON and any JSON parser will reject it. A `json.NewEncoder` call serialises the string correctly, escaping all special characters. There is also a secondary issue: `http.Error` sets `Content-Type: text/plain`, overriding the router-level `application/json` header. A helper that sets the header explicitly and encodes through the encoder fixes both issues.

---

## Production Readiness

**Q: What would you change before putting this in production?**

- **Deduplication** — `ReplacingMergeTree` or a UUID-based insert guard.
- **Auth** — the upload endpoint accepts files from anyone. At minimum HTTP Basic Auth or a static API key.
- **File size and format validation** — currently limited to 32 MB by `maxUploadBytes` but no MIME type check before passing to excelize.
- **Observability** — structured logging exists but no metrics (Prometheus counters for ingested rows, query latency, error rates). Tracing would help diagnose slow ClickHouse queries.
- **Rate limiting** — the ingest endpoint does unbounded CPU work per request.
- **Pagination for `/api/v1/versions`** — fine with tens of versions, problematic with thousands.
- **Graceful shutdown** — `srv.Close()` drops active connections immediately; `srv.Shutdown(ctx)` with a timeout would drain them.

---

**Q: The Dockerfile uses a multi-stage build. Walk through the stages.**

1. **`frontend-builder`** (Node Alpine) — installs npm dependencies with cache mounts and runs `npm run build`, producing `frontend/dist/`.
2. **`go-base`** (Go Alpine) — downloads Go module dependencies with cache mounts and copies source. Both the main module and the `frontend` module (with its `dist/`) land here.
3. **`backend-builder`** (extends `go-base`) — compiles the binary with `-tags with_frontend`, embedding the dist directory. `ldflags` inject `version`, `commit`, and `date` for the `version` subcommand.
4. **`production`** (Alpine) — copies only the compiled binary. No compiler, no Node, no source code. `ENTRYPOINT ["./observability"]` + `CMD ["run"]` means `docker run image` starts the server and `docker run image version` prints build info.

