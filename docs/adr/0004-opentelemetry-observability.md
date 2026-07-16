# ADR 0004 — Observability: slog foundation first, OpenTelemetry opt-in

- Status: Accepted (2026-07-16)
- Deciders: user (approved plan)
- Context docs: `docs/plans/cli-architecture-v2-improvement-plan.md` (Pillar C),
  `POLYMETRICS_GO_CLI_MONOLITH_PRD_ARCHITECTURE.md` §15 (observability),
  `docs/adr/0003-interactive-tui-layer.md` (events bus — sibling layer)

## Context

The repo has zero observability: no logging framework in any non-test code (the Temporal
client is handed a `noopLogger`, `internal/worker/submit.go:99-106`), no metrics, no
tracing. PRD §15 specifies per-run JSONL logs under `.polymetrics/logs/` (the directory is
created by `pm init` and never written to), a metrics list, redaction middleware, and an
audit ledger — only the ledger exists. A failed ETL run leaves a one-line ledger row and
nothing else. Constraints: stdout carries exactly one JSON envelope per invocation (logs
must go to stderr or files); secrets must never appear anywhere (`api_key_query` auth puts
credentials in URL query strings — `internal/connectors/engine/auth.go:93-98`); the
product tenet is dependency-free default behavior; most `pm` commands are short-lived
processes (flush-on-exit latency matters); the per-record ETL loop is hot
(`internal/app/app.go:511-543`). otel-go status (May 2026 train, v1.44.0/v0.66.0/v0.20.0):
traces and metrics stable, logs beta; stdout exporters accept `WithWriter`; a slog bridge
(`otelslog` v0.19.0) and a Temporal contrib package
(`go.temporal.io/sdk/contrib/opentelemetry` v0.7.0) exist. The pinned Temporal SDK already
ships `log.NewStructuredLogger(*slog.Logger)`.

## Decision

1. **`log/slog` is THE logging API, adopted first, stdlib-only, default ON.** A
   `RedactingHandler` (always outermost) fans out to a per-run JSONL file
   (`.polymetrics/logs/<run-id>.jsonl`, routed by a run ID carried in `context`) and to
   stderr at warn+. Redaction is defense-in-depth: key-based (connector `SecretFields` +
   a fixed sensitive-key set), value-based (a registry of secret values fed from the
   single vault chokepoint `vault.Get`, `internal/vault/vault.go:80`), and the existing
   `safety.RedactErrorText`/`SanitizeTerminal`. Temporal's `noopLogger` is replaced with
   `tlog.NewStructuredLogger` over the same redacting logger (zero new dependencies).
2. **OpenTelemetry arrives after logging, opt-in, in signal order traces → metrics →
   (optional) log bridge.** `internal/telemetry` owns `Init(ctx, cfg)` / bounded
   `Shutdown` (3s, warn-and-continue — telemetry never alters exit codes). Exporter modes:
   `none` (default — **no SDK constructed**, zero flush latency), `file` (JSONL under
   `.polymetrics/telemetry/` via stdout exporters with `WithWriter` — the offline
   inspection mode), `otlp` (http/protobuf; endpoint via config or standard
   `OTEL_EXPORTER_OTLP_ENDPOINT`; `OTEL_SDK_DISABLED` always wins).
3. **Span map**: `pm.command` (root, in the dispatch seam — portable to cobra
   `PersistentPreRunE`) → `pm.etl.run` / `pm.flow.run` / `pm.flow.step` /
   `pm.certify.batch` / `pm.certify.connector` / `pm.rlm.submit` → `pm.connector.http`.
   HTTP is instrumented **directly in the single chokepoint `connsdk.Requester.do`**
   (`internal/connectors/connsdk/http.go:216`), one span per logical request with
   per-attempt events and retry/rate-limit attributes.
4. **Secret safety is structural**: an attribute-key allowlist (enforced by a test that
   walks exported spans), URL attributes reduced to scheme+host+path (query strings always
   dropped — stricter than semconv, required by `api_key_query`), never bodies/headers/raw
   argv, the value registry scrubbing error strings, and a `capture: minimal` mode that
   strips attributes entirely. The red test: run the smoke flow with a known token and
   grep logs + telemetry for absence, with a test hook proving the grep can fail.
5. **Metrics implement the PRD §15.2 list** (`pm.records.*`, `pm.batches.*`, `pm.api.*`,
   `pm.bytes.*`, latency/rate-limit/stage-duration histograms) under one hard rule: **no
   per-record instrument calls** — hot loops keep local counters and flush per batch, with
   a benchmark guard. Temporal gains the contrib tracing interceptor + metrics handler,
   gated on enabled.
6. **Layering**: events bus (real-time progress, ADR-0003), run ledger (durable audit),
   and slog/OTel (diagnostics) are **siblings** instrumenting the same call sites,
   correlated by `run_id` — none derives from another.
7. **The otel log bridge (beta) is last and droppable**: `otelslog` appended as one more
   fan-out branch inside the redacting handler, only when `exporter=otlp`; versions
   pinned; slog remains the permanent app-facing API.

## Alternatives considered

- **No telemetry / logs only**: rejected — PRD §15 explicitly requires metrics and
  diagnostics; but the phasing ensures logging alone (zero deps) stands if the go.mod gate
  ever rejects the rest.
- **Prometheus exporter**: rejected — a pull endpoint suits daemons, not short-lived CLI
  processes; OTLP push + file export cover both `pm` and `pm worker serve`.
- **otelhttp for connector HTTP**: rejected — it records `url.full` (semconv scrubs only
  userinfo, not query strings), which leaks `api_key_query` credentials; `Requester.do`
  also owns retries, so one logical-request span with attempt events is more useful than
  N transport spans.
- **grpc OTLP as default**: rejected — http/protobuf avoids promoting grpc to a direct
  dependency (it stays indirect via Temporal); the grpc exporter remains a config option.
- **zerolog/zap instead of slog**: rejected — stdlib slog is dependency-free, has the
  otel bridge, and the Temporal SDK ships a slog adapter.
- **Events derived from spans (single instrumentation)**: rejected — couples the UI to
  telemetry configuration (ADR-0003 records the sibling decision).

## Consequences

- (+) Every run leaves a redacted, structured JSONL trail (PRD §15.1) with zero new
  dependencies and no behavior change to stdout.
- (+) Opt-in traces answer "why is this connector slow" down to per-attempt HTTP events;
  metrics reconcile with envelope counts by construction (tested equality).
- (+) Disabled-path cost is provably zero: no SDK constructed, no flush on exit.
- (−) Phases B–D add ~10 otel modules (genuinely new transitives: otel/*, go-logr,
  backoff/v5, proto/otlp; grpc/protobuf already present via Temporal; est. +2–4 MB
  binary) → accepted through the go.mod human gate this ADR records; the "dependency-free"
  tenet is interpreted as *default runtime behavior* (no egress, no collector, telemetry
  off) rather than compile-time purity.
- (−) The log signal is beta → confined to one file, pinned, optional, droppable; slog is
  the stable API surface.
- (−) Three sinks (events/ledger/telemetry) at the same call sites add per-step
  boilerplate → accepted as explicit; a thin facade may wrap them later if it grows
  unwieldy (not part of this plan).
