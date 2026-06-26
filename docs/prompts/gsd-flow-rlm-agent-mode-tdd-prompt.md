# GSD Programming-Loop Prompt — Flow Engine, RLM, Reverse-ETL Actions, Scheduling & Agent Mode (Strict TDD)

Date: 2026-06-27
Loop: `$gsd-programming-loop` (init → profile → run --phase N → verify --phase N), one phase at a time, strict TDD.
Status: canonical prompt. Full rendered per-phase prompts go under `.planning/phases/<phase>/PROMPTS.md`.

---

## 0. Mission

Turn `pm` into the most token-efficient agentic ETL / reverse-ETL tool — an alternative to per-tool MCP servers — by adding a **declarative multi-step Flow engine**, a **Reasoning/Learning Module (RLM)** analysis stage, **approval-gated reverse-ETL actions** (email, LinkedIn, any connector with `Write`), **scheduling**, and a token-lean **Agent Mode**. Build it on the primitives that already exist (`internal/app` ETL + reverse, `internal/connectors`, `internal/vault`, `internal/ledger`, `internal/state`, `internal/runtime`, `internal/runtimecheck`).

Worked end-to-end target: *collect contact/email data from multiple sources → RLM scores likely customers → RLM drafts personalized emails → approval-gated action sends via an email connector and messages via LinkedIn → receipts + metrics recorded.*

## 1. Non-negotiable invariants (do not violate in any phase)

These are existing project laws (`AGENTS.md`, `CONTEXT.md`). Treat as hard gates:

1. **Dependency-free is the default.** Core flow/RLM logic uses the Go standard library only. Do **not** add `cobra`, `viper`, or any new third-party module to the default path. The `pm` CLI keeps its existing hand-rolled dispatcher.
2. **Runtime-backed mode is opt-in and additive.** The only allowed heavy deps are the ones already in `go.mod`: `pgx` (Postgres), `go-redis` (DragonflyDB), `temporal-sdk`. Selected via existing env vars `POLYMETRICS_POSTGRES_URL`, `POLYMETRICS_DRAGONFLY_ADDR`, `POLYMETRICS_TEMPORAL_ADDR` and gated by `internal/runtimecheck`. Every feature must work dependency-free first, then optionally accelerate with these.
3. **Secrets:** never request, print, log, or store secret *values*. Expose secret *field names* only. Add creds from env/stdin.
4. **Side effects flow through `plan → preview → approval-token → execute`.** No generic shell, generic HTTP write, or generic SQL write primitive — ever. RLM/LLM output is DATA, never an execution path.
5. **JSON envelopes for agents; stderr for humans.** `--json` is machine-contract; keep it stable and compact.
6. **Strict TDD.** Behavior-adding code requires a failing test first (red evidence in `TDD-LEDGER.md`) before implementation. No weakening tests or quality gates to pass.
7. **Verification gate per phase:** `export GOTOOLCHAIN=auto; gofmt -w cmd internal; go vet ./...; go test ./...; go build ./cmd/pm; make verify` — must end green.

## 2. Apply these installed Go skills (samber/cc-skills-golang)

Reference the matching skill before writing code in each area. Installed at `~/.claude/skills/cc-skills-golang/skills/`.

- **`golang-testing` + `golang-stretchr-testify`** — table-driven tests, `httptest`, fixtures FIRST. (Match the repo's existing testify usage; do not introduce a different assertion lib.)
- **`golang-cli`** — subcommand ergonomics, `--json`/exit codes. (Patterns only — keep the existing stdlib dispatcher; do NOT adopt cobra/viper.)
- **`golang-error-handling`** — wrapped sentinel errors (`errors.Is/As`), no panics across package boundaries.
- **`golang-context`** — every long op takes `ctx`, honors cancellation/timeouts (mirrors `runtimecheck`).
- **`golang-concurrency`** — bounded worker pools for batched action sends; no unbounded goroutines.
- **`golang-database`** — pooled `pgxpool`, parameterized queries, transactions for the runtime-backed ledger/state.
- **`golang-observability`** — structured logging + the sync/error/dropped-record metrics (must-have feature set).
- **`golang-security` + `golang-safety`** — SSRF-validate any base_url override; treat all args as untrusted; no secret leakage; nil-safe maps.
- **`golang-design-patterns`** — functional options for engine/step constructors; strategy interface for RLM backends.
- **`golang-naming` + `golang-code-style` + `golang-lint` + `golang-project-layout`** — match existing `internal/connectors/*` conventions.
- **`golang-performance` + `golang-benchmark`** — benchmark token/throughput claims for Agent Mode (Phase 5) against MCP-style output.

## 3. Architecture (build on what exists)

```
internal/schedule  (NEW)  pm schedule   → materializes cron onto launchd/systemd/Temporal-cron; invokes `pm flow run`
internal/flow      (NEW)  pm flow       → typed DAG: steps {sync, query, rlm, action}; derives deps from in/out tables
internal/rlm       (NEW)  pm rlm        → analysis/scoring/drafting → materializes warehouse table; deterministic|model|fixture
internal/agentmode (NEW)  pm agent ...  → token-lean envelopes, --fields projection, generated skills, optional `pm mcp serve`
   ↑ all reuse ↓
internal/app (etl run, query, reverse plan/preview/run), internal/connectors (Write), internal/vault,
internal/ledger (receipts/audit), internal/state (cursors/locks), internal/runtime (+runtimecheck) for optional accel
```

Flow step kinds map 1:1 to existing verbs: `sync`=`pm etl run`, `query`=warehouse SQL materialization, `action`=`pm reverse`. Only `rlm` is genuinely new logic. A flow run generalizes the reverse-ETL gate: read-only steps execute for real; `action` steps dry-run in `plan`; `preview` surfaces every action's record count + sample payloads + redacted recipients; one approval token (or `--per-action`) gates `run`.

## 4. Must-have features to implement (from 2025/26 ETL/reverse-ETL research)

Bake these into the relevant phase, each test-first:

- **Idempotent writes** keyed by deterministic record id; re-run never duplicates. (action)
- **Identity mapping**: warehouse PK ↔ external system id, persisted in state. (action)
- **Dedupe/merge rules** on email/domain/external-id before load. (query/rlm)
- **Rate-limit handling**: 429-aware, exponential backoff + jitter, batching, optional Dragonfly-backed queue. (action, reuse `connsdk` retry)
- **Dead-letter queue + bounded retries**; failed records quarantined with reason, never silently dropped. (action)
- **Schema-drift detection**: compare live catalog vs stored snapshot; on breaking change, **pause the step and refuse partial pushes**, emit alert. (sync/action)
- **Incremental sync + clean backfill + cursor state** (extend existing `StreamState`). (sync)
- **Observability metrics**: sync latency, records read/written/failed/dropped, error rate, schema-mismatch count — in the run envelope + ledger. (flow)
- **Audit log / receipts / lineage**: every action writes a receipt (what was sent where, redacted) to the ledger. (flow/action)

## 5. Phase plan (run one phase per loop iteration; human gate between phases)

### Phase 0 — Flow engine skeleton (`sync` + `query` steps)
- `internal/flow`: manifest parse/validate (YAML authoring → normalized JSON), DAG build from `in`/`out` tables, topological execute, checkpoint to ledger, lease via `internal/runtime` to prevent overlap.
- `pm flow plan|preview|run|status|list`. `plan` runs read-only steps for real, dry-runs nothing yet (no actions).
- TDD: manifest validation errors, cycle detection, dependency ordering, checkpoint/resume, lease contention.
- Gate: full chain `sync→query` runs against existing app primitives, `make verify` green.

### Phase 1 — Reverse-ETL `action` step + the safety must-haves
- `action` step = thin adapter over `internal/app` reverse plan/preview/run. Lift the single-step approval to a flow-level token (and `--per-action`).
- Implement: idempotent writes, identity mapping (state), dedupe/merge, rate-limit/backoff, DLQ+retries, receipts/audit, schema-drift pause.
- TDD: `httptest` destination — assert no duplicate on re-run, 429→backoff→success, DLQ on permanent failure, approval required, schema-drift halts before any write.
- **Human gate:** this is the first code that *sends*. Require explicit approval before enabling real network writes.

### Phase 2 — RLM engine (deterministic backend first)
- `internal/rlm`: `Analyzer` strategy interface; backends `deterministic` (SQL/weighted-feature scoring, fully offline, reproducible), `fixture` (canned table, credential-free), and a stubbed `model` seam.
- `pm rlm run --spec ... --in <table> --out <table> --mode deterministic`. Output is a materialized warehouse table (e.g. `lead_scores`).
- Wire `rlm` as a flow step kind. End-to-end `likely-customers` flow runs fully offline in fixture/deterministic mode.
- TDD: scoring determinism (same input→same output), feature/weight spec parsing, materialization schema, fixture parity.

### Phase 3 — Scheduling (`pm schedule`)
- `internal/schedule`: schedule manifest binds cron → flow. `pm schedule create|list|install|remove`. Dependency-free: `install` emits launchd plist / systemd-user timer / crontab line whose payload is `pm flow run <name> --json`. No resident daemon.
- Runtime-backed: register a Temporal cron workflow instead (overlap-prevention, retries, durability) via existing runtime adapter.
- TDD: cron parse/validate, unit-file rendering golden tests, runtime-mode selection.

### Phase 4 — RLM `model` backend (opt-in, gated)
- Implement the `model` backend: classification + email/subject **drafting** via Claude. Output cached to warehouse (`lead_outreach`), treated as side-effect-free read behind a network+credential gate; logged to ledger. Model never sees credentials or a send primitive.
- TDD: fixture-replayed model responses (no live calls in CI), cache hit determinism, redaction, "draft never sends" invariant.
- **Human gate:** dependency/credential/network behavior change.

### Phase 5 — Agent Mode + token efficiency ("replace MCP")
- `internal/agentmode`: compact stable envelopes (counts+samples, not dumps), `--fields` projection, NDJSON streaming, deterministic ids; extend `pm skills generate` so agents load on-demand skill docs instead of fat tool schemas. Upgrade `pm agent plan` → `pm agent flow --request "..."` that emits a **flow manifest** (then goes through the same plan/preview/approve gates — agent proposes, gate disposes).
- Optional thin `pm mcp serve` exposing ~6 high-level verbs (catalog, etl.run, flow.plan/preview/run, rlm.run, reverse) — NOT one-tool-per-connector.
- TDD + **benchmark** (`golang-benchmark`): measure tokens/bytes of Agent-Mode output vs an MCP-style fat-schema baseline; assert the reduction. Round-trip: agent-authored manifest → preview → approval → run.

## 6. Loop mechanics (how to actually run this)

Preconditions: clean git tree; `.planning/PROJECT.md` present (run `$gsd-new-project` if missing). Then per phase:

```bash
$gsd-programming-loop init
$gsd-programming-loop profile                 # writes docs/architecture/repo-profile.json
$gsd-programming-loop run --phase 0           # strict TDD; red evidence in TDD-LEDGER.md before code
$gsd-programming-loop verify --phase 0        # gofmt/vet/test/build/make verify
# commit phase trace; STOP at human gate; then --phase 1 ... through --phase 5
```

Per-phase artifacts: `.planning/phases/<phase>/{PROMPTS.md,TDD-LEDGER.md,SUMMARY.md,VERIFICATION.md,agents/,traces/}`.

Stop for human approval before: any dependency addition, the first network-write enablement (Phase 1), the model/credential change (Phase 4), and any quality-gate reduction. Do not batch all five phases past a failing/red gate — advance only on green.

## 7. Definition of done

- `pm flow`, `pm rlm`, `pm schedule`, `pm agent`/`pm mcp serve` implemented, all dependency-free by default, optionally runtime-accelerated.
- `likely-customers` example flow runs end-to-end offline (fixture/deterministic) and, with creds + approval, sends real personalized email + LinkedIn messages with idempotency, rate-limit safety, DLQ, schema-drift pause, receipts, and metrics.
- Every must-have feature (§4) has a passing test. Agent-Mode token reduction is benchmarked. `make verify` green at every phase boundary.
