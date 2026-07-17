# TDD Ledger — Phase 410 OpenTelemetry tracing

## Skills loaded

- `gsd-core` — repo-local GSD adapter workflow.
- `golang-how-to` — Go skill routing.
- `golang-testing` — test-first, named subtests, behavior specs.
- `golang-security` — trust boundaries, secret and injection avoidance.
- `golang-safety` — nil/map/default/resource safety.
- `golang-observability` — OTel spans, context propagation, safe cardinality.
- `golang-context` — context-first APIs, bounded shutdown.
- `golang-concurrency` — shutdown/leak discipline.
- `golang-error-handling` — wrapped errors, single handling, warn-and-neutral telemetry failures.
- `golang-cli` — stdout/stderr, exit-code, config/help parity.
- `golang-documentation` — CLI docs and website parity.
- `caveman` — final worker handoff compression only.

## GSD command evidence

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 410 --skip-research >/tmp/gsd-plan-phase-410.txt
scripts/gsd prompt programming-loop init --phase 410 --dry-run >/tmp/gsd-programming-loop-410.txt
```

Result:

- `doctor`: pass.
- `plan-phase`: generated 142-line prompt.
- `programming-loop`: failed with `scripts/gsd: unknown GSD command: programming-loop`; manual GSD/TDD fallback recorded in PLAN/RUN-STATE.

## Ledger

| # | Cycle | Type | Command / evidence | Result | Notes |
|---:|---|---|---|---|---|
| 1 | plan | Planning | Create phase artifacts before production edits | Pass | `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `RUN-STATE.json`, `SUMMARY.md`, `PROMPTS.md` created. |
| 2 | red | Test | `go test ./internal/telemetry ./internal/config ./internal/cli ./internal/connectors/connsdk ./internal/app ./internal/flow -run 'Telemetry|TestLoadTelemetry' -count=1` | Fail | Disabled mode no SDK/dir, file exporter, allowlist, command/certify, ETL/flow, HTTP spans, and neutral failure tests added before production code. |
| 3 | red | Output | `internal/telemetry/telemetry_test.go:18:17: undefined: Init`; `internal/config/telemetry_config_test.go:15:9: cfg.Telemetry undefined`; CLI telemetry dir missing and warning absent; connsdk/app/flow build failed pending telemetry package | Fail | Expected red: telemetry package/config/instrumentation not implemented yet. |
| 4 | green | Implementation | Pending | Pending | Minimal tracing core/instrumentation. |
| 5 | refactor | Verification | Pending | Pending | Docs/help parity and full gates. |

## Red-test requirements

Required failing evidence before production edits:

- Disabled telemetry: no SDK constructed, no `.polymetrics/telemetry` directory.
- Enabled file exporter: JSONL spans include `pm.command` and child operation spans.
- HTTP span attributes: allowlist only, no `url.full`, `http.request.header.*`, request body, query strings, or known token.
- Exit-code neutrality: init/exporter/shutdown failures print warning to stderr and preserve command exit code/stdout envelope.

## Secret-safety probes

Use artificial non-secret test token literals only. Do not use or print real credentials.

Planned grep evidence:

```bash
rg -n "pm_test_secret_token|token=|Authorization|url.full|request.body|http.request.header|argv" "$TELEMETRY_DIR"
```

Expected: no matches for exported telemetry from secret-safety smoke.
