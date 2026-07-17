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
- `golang-lint` — review-fix gate quality, `go vet`, and lint discipline.
- `golang-code-style` — minimal, clear Go changes during security hardening.
- `golang-design-patterns` — explicit lifecycle and bounded shutdown patterns.
- `golang-structs-interfaces` — small test seams and config/source metadata types.
- `golang-dependency-management` — no new dependencies; OTel dependency budget unchanged.
- `caveman` — final worker handoff compression only.

Note: required stack skill `.pi/skills/go-implementation/SKILL.md` is absent in this checkout (`.pi/skills` contains only `gsd-core`); loaded repo/global Go implementation skills above instead and will record this in the handoff.

## GSD command evidence

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 410 --skip-research >/tmp/gsd-plan-phase-410.txt
scripts/gsd prompt programming-loop init --phase 410 --dry-run >/tmp/gsd-programming-loop-410.txt
scripts/gsd prompt plan-phase 410 --skip-research >/tmp/gsd-plan-phase-410-reviewfix.txt
scripts/gsd prompt programming-loop init --phase 410 --dry-run >/tmp/gsd-programming-loop-410-reviewfix.txt
```

Result:

- `doctor`: pass (rerun 2026-07-18; 69 commands).
- `plan-phase`: generated 142-line prompt (initial and review-fix rerun).
- `programming-loop`: failed with `scripts/gsd: unknown GSD command: programming-loop`; manual GSD/TDD fallback recorded in PLAN/RUN-STATE.

## Review-fix accepted findings

| Finding | Disposition | Planned red evidence |
|---|---|---|
| HIGH error `RecordError` leaks via SDK `exception.*` attrs | Accepted | exported-span tests for command/HTTP/flow failures with synthetic registered marker; assert no `exception.*`, raw message/body/query/marker; minimal capture suppresses messages |
| HIGH/MED config-sourced OTLP can redirect network telemetry | Accepted | config-file OTLP disabled/rejected by default; env opt-in accepted |
| MED file exporter path traversal/symlink | Accepted | absolute/`..`/symlink dir/file tests fail before implementation |
| MED AddEvent attrs attached to span not event | Accepted | JSON export event attr test asserts event `pm.http.attempt`/retry/status attrs exist and are allowlisted |
| MED OTLP failures bypass warning/bounded shutdown | Accepted | CLI OTLP failure test preserves exit/stdout and emits sanitized `warning: telemetry:` |
| LOW endpoint validation | Accepted | reject userinfo/query/fragment and non-http(s); no endpoint secret leak |
| LOW docs root HTTP attr wording | Accepted | help/golden/docs/website checks updated |
| LOW exporter `none/off` wording mismatch | Accepted | remove `/off` wording unless code alias intentionally added |
| LOW warnings stdout/redaction hardening | Accepted | warning tests/secret smoke assert stderr only and registered marker redacted |

## Ledger

| # | Cycle | Type | Command / evidence | Result | Notes |
|---:|---|---|---|---|---|
| 1 | plan | Planning | Create phase artifacts before production edits | Pass | `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `RUN-STATE.json`, `SUMMARY.md`, `PROMPTS.md` created. |
| 2 | red | Test | `go test ./internal/telemetry ./internal/config ./internal/cli ./internal/connectors/connsdk ./internal/app ./internal/flow -run 'Telemetry|TestLoadTelemetry' -count=1` | Fail | Disabled mode no SDK/dir, file exporter, allowlist, command/certify, ETL/flow, HTTP spans, and neutral failure tests added before production code. |
| 3 | red | Output | `internal/telemetry/telemetry_test.go:18:17: undefined: Init`; `internal/config/telemetry_config_test.go:15:9: cfg.Telemetry undefined`; CLI telemetry dir missing and warning absent; connsdk/app/flow build failed pending telemetry package | Fail | Expected red: telemetry package/config/instrumentation not implemented yet. |
| 4 | green | Implementation | `go test ./internal/telemetry ./internal/config ./internal/cli ./internal/connectors/connsdk ./internal/app ./internal/flow -run 'Telemetry|TestLoadTelemetry|Golden|Config' -count=1` | Pass | Telemetry core, config, command/certify, ETL, flow, and connsdk HTTP span tests green. |
| 5 | green | Smoke | File/off/secret smoke script with `PM_TELEMETRY=file`, `PM_CERT_SAMPLE_TOKEN=<synthetic marker>`, and `rg` forbidden-pattern scan | Pass | Off mode no dir; file mode command/certify spans; stdout JSON parse ok; no synthetic marker/headers/body/full-url/query tokens in telemetry. |
| 6 | green | Gates | `gofmt -w cmd internal`; `go vet ./...`; `go test ./...`; `go build ./cmd/pm` | Pass | Full Go tests passed; CLI package ~208s, certify package ~352s. |
| 7 | parity | Help/docs/website | `./pm --help`; `./pm help config`; `./pm etl`; `./pm flow`; `./pm connectors`; invalid `./pm connectors bogus --json` exit 2; docs generate/diff; `npm --prefix website run gen:docs` | Pass | Runtime help/docs/website parity updated for telemetry config. |
| 8 | gate | `make verify` before commit | Fail | Expected tidy-check failure because go.mod/go.sum dependency changes were uncommitted; rerun after green-slice commit. |
| 9 | gate | `make verify` after green-slice commit | Pass | fmt, tidy-check, vet, 20m tests, build, docs validate, smoke, lint, and connectorgen validate passed. |
| 10 | review-fix plan | Planning | Updated phase artifacts for PR #459 review findings before production edits | Pass | New accepted findings, planned red tests, verification checklist, and local-critical-path decision recorded. |
| 11 | review-fix red | Planned test | `go test ./internal/telemetry ./internal/config ./internal/cli ./internal/connectors/connsdk ./internal/flow -run 'Telemetry|OTLP|Endpoint|Event|RecordError|FileExporter' -count=1` | Pending | Add failing tests before production edits; record exact red output here. |

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
