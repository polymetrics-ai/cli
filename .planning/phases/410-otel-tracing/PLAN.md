# Phase 410 — Opt-in OpenTelemetry tracing

Issue: #410 `feat(obs): add opt-in OpenTelemetry tracing`  
Parent: #397 / parent PR #438 (`feat/cli-architecture-v2`, draft, human-gated)  
Branch: `feat/410-otel-tracing`  
Worker dir: `/Users/karthiksivadas/Development/polymetrics-cli-agents/wt-410-otel-tracing`

## GSD adapter and execution mode

- `scripts/gsd doctor` passed on 2026-07-18.
- `scripts/gsd prompt plan-phase 410 --skip-research` generated a 142-line prompt in `/tmp/gsd-plan-phase-410.txt`.
- `scripts/gsd prompt programming-loop init --phase 410 --dry-run` failed with `scripts/gsd: unknown GSD command: programming-loop`; manual GSD/TDD fallback is active for the programming-loop step only.
- Universal loop decision for planning cycle: `local_critical_path` — this Pi worker has no subagent tool and owns one isolated issue worktree.

## Required reading completed

- `AGENTS.md`.
- Issue #410 body and acceptance criteria.
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`.
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`.
- `.agents/agentic-delivery/references/required-skills-routing.md`.
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`.
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.
- `.agents/agentic-delivery/references/runtime-rlm-website-integration.md`.
- `docs/plans/cli-architecture-v2-improvement-plan.md` Pillar C / Stage 12.
- `docs/prompts/cli-architecture-v2-gsd-execution-prompt.md` Stage 12.
- `docs/adr/0004-opentelemetry-observability.md`; ADR 0002/0003 checked for dependency and sibling-layer constraints.
- Existing phase artifacts did not exist before this worker run; this directory is created for #410.

## Required skills loaded

- `gsd-core`.
- `golang-how-to`.
- `golang-testing` (rules 1, 3, 5, 9: named/table tests, independent, behavior not implementation, fast unit tests).
- `golang-security` (trust-boundary questions #1-#3; no bodies/headers/query/argv/secret values; path and env inputs treated as untrusted).
- `golang-safety` (rules 2, 4, 6, 10: safe assertions, map initialization, defensive copies, useful defaults).
- `golang-observability` (rules 7-12: OTel setup, meaningful spans, context propagation, no high-cardinality unsafe attrs).
- `golang-context` (rules 1-8: propagate caller context, no mid-path Background, bounded shutdown context).
- `golang-concurrency` (rules 1, 7: shutdown/ctx awareness; no goroutine leaks in exporters/shutdown).
- `golang-error-handling` (rules 2, 7, 10, 14: wrapped errors, single handling, structured warnings on stderr, exit-code neutrality).
- `golang-cli` (stdout/stderr discipline, exit code stability, flag/config/help parity).
- `golang-documentation` (CLI docs/help/website parity; concise safety docs).

## Dependency constraints

Approved by ADR 0004 Stage 12 only:

- `go.opentelemetry.io/otel@v1.44.0`
- `go.opentelemetry.io/otel/sdk@v1.44.0`
- `go.opentelemetry.io/otel/exporters/stdout/stdouttrace@v1.44.0`
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp@v1.44.0`

Review-fix note: event attributes require `trace.WithAttributes`, so `go mod tidy` promotes existing `go.opentelemetry.io/otel/trace@v1.44.0` from indirect to direct with no version/checksum change. This is treated as an existing OTel trace API submodule promotion, not a new module/version.

No `otelhttp`, no grpc exporter promotion, no metrics SDK, no otel log bridge, no Temporal OTel contrib in this issue. Any version/module deviation is a human gate.

## Scope / write plan

Allowed production scope:

- New `internal/telemetry/**` package.
- Narrow config/env/help/docs additions for telemetry enablement.
- Command seam instrumentation in `internal/cli/**`.
- Operation instrumentation in `internal/app/**`, `internal/flow/**`, `internal/connectors/certify/**` or certify CLI wrapper as needed.
- HTTP chokepoint instrumentation in `internal/connectors/connsdk/http.go`.
- Tests under touched packages plus docs/website parity updates.

Excluded:

- `.planning/traces/cli-architecture-v2-orchestration-state.yaml` and other shared parent orchestration artifacts.
- Later phases (#415 metrics, #419 log bridge), TUI work, connector defs, broad generated rewrites unrelated to docs parity.
- Credentialed connector checks and runtime service startup.

## Slice plan

### Slice 0 — red tests + dependency gate

1. Add failing tests before production code:
   - `internal/telemetry`: disabled mode no SDK/dir; file exporter smoke; attribute allowlist and secret absence; init/exporter/shutdown warn-and-neutral behavior via test hooks.
   - `internal/connectors/connsdk`: HTTP span records only scheme/host/path/status/attempt attrs and omits query/header/body/token.
   - `internal/cli` or integration-level tests: command span wraps expected commands and keeps stdout envelope-only; telemetry init/shutdown failures do not change exit codes.
   - `internal/app` / `internal/flow` / certify wrapper: ETL, flow step, and certify spans when enabled.
2. Run focused tests and record exact red output in `TDD-LEDGER.md` before production edits.
3. Add only ADR-approved OTel modules/version lines.

### Slice 1 — `internal/telemetry` core

- Implement `Config`, `Init(ctx,cfg,warn)`, disabled short-circuit with no SDK and no directory creation, file exporter under `<root>/.polymetrics/telemetry/<run-id-or-timestamp>.jsonl`, OTLP http/protobuf endpoint, `OTEL_SDK_DISABLED` override, bounded shutdown (3s in CLI, injectable in tests).
- Provide context helpers: `StartSpan`, `SpanFromContext` fallback/nop behavior, `SetAttributes` through an allowlist, and HTTP attribute builders that drop query/userinfo/fragment/body/header/argv.
- Warning callback writes sanitized warnings to stderr; never returns errors that alter command exit.

### Slice 2 — command and operation instrumentation

- Wire config/env keys into `internal/config`: exporter mode default `none`/`off`, endpoint, directory, capture mode.
- In `RunWithOptions`, initialize telemetry after config/logging setup; start `pm.command`; defer bounded shutdown warning.
- Start `pm.etl.run` around app ETL run; `pm.flow.run` / `pm.flow.step` in flow engine; `pm.certify.connector` / `pm.certify.batch` in certify paths; keep action approval semantics untouched.

### Slice 3 — connector HTTP instrumentation

- Instrument `connsdk.Requester.do` as one `pm.connector.http` logical-request span with per-attempt events.
- Attributes allowlist only: `pm.http.scheme`, `pm.http.host`, `pm.http.path`, `pm.http.method`, status/attempt counts/retry booleans; no request/response bodies, headers, query strings, full URLs, argv, or credential values.

### Slice 4 — CLI parity, docs, smoke gates

- Update embedded help (`rootHelp`/`configHelp`), generated `docs/cli/config.md` (and root docs if generator changes), website `website/content/docs/cli-reference.mdx`, and generated website data if needed.
- Verify `pm help config`, `pm --help`, relevant bare namespaces unchanged, docs generate diff, website docs generator.
- Run file exporter smoke, off-mode check, secret grep, envelope-only stdout, `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify` when feasible.

## Acceptance criteria mapping

| AC | Plan evidence |
|---|---|
| Disabled mode constructs no SDK and creates no telemetry directory | `internal/telemetry` disabled test + CLI off-mode smoke |
| Expected command, ETL/flow/certify, connector HTTP spans emitted when enabled | file exporter smoke and focused package tests |
| HTTP attributes only allowlisted scheme/host/path metadata | connsdk span attr test + exported-span walker |
| Exporter/init/shutdown failures warn without changing exit codes | CLI tests with failing exporter/shutdown hooks |
| Secret and attribute allowlist tests pass | test token absence grep + allowlist walker |

## Review-fix plan — PR #459 findings at `df9447f0fefcf2e52c8f5a0fece318d80e4ce9d8`

Execution decision: `local_critical_path` — this Pi worker owns the isolated issue worktree, has no subagent tool, and review coverage is handled by coordinator sidecars/human fallback. Do not request Copilot/Claude again.

Accepted findings and fix slices:

1. Error telemetry safety (HIGH): replace direct SDK `RecordError` use with allowlisted, registry-redacted, stable error metadata. In `capture=minimal`, suppress message-like error attrs. Add exported-span tests for failed command, HTTP, and flow spans using synthetic registered marker and response-body-like errors.
2. Config-sourced OTLP guard (HIGH/MED): prevent project config from silently enabling/redirecting OTLP. Allow network exporter/endpoint only through explicit trusted env/CLI opt-in; reject config-sourced OTLP by default with sanitized warning and disabled telemetry. Keep default/file local behavior safe.
3. File exporter path safety (MED): constrain telemetry file output under project root, reject absolute or escaping directories, reject symlinked dirs/files, use restrictive permissions and no-follow/exclusive create where feasible.
4. Event attributes (MED): attach filtered attrs to events with `trace.WithAttributes(...)` rather than span attrs; assert retry/attempt event attrs are present and allowlisted.
5. OTLP failure neutrality (MED): route init/export/shutdown failures through sanitized `telemetryWarning`, preserve exit code/stdout JSON, and bound shutdown.
6. OTLP endpoint validation (LOW): allow http/https only; reject userinfo, query, and fragment without leaking endpoint secrets.
7. Help/docs parity (LOW): update root/config help and generated docs/website data so HTTP attrs mention method/status/attempt/retry and still exclude query/headers/bodies/argv.
8. Exporter wording (LOW): keep supported values consistent as `none,file,otlp`; remove undocumented `off` wording unless implemented intentionally.
9. Warning redaction/stdout discipline (LOW): ensure Redis/exporter/log/telemetry warnings are redacted and never corrupt stdout JSON.

Planned red tests before production edits:

- `internal/telemetry`: exported JSON spans for `RecordError` contain only `pm.error.*` allowlisted metadata, not SDK `exception.*`, synthetic registered markers, response-body-like text, query strings, or raw messages in minimal capture; AddEvent retains event attrs.
- `internal/config`/`internal/cli`: config-file `telemetry.exporter: otlp` and endpoint are rejected/disabled with sanitized stderr warning; env opt-in to OTLP is accepted; endpoint validation rejects userinfo/query/fragment and non-http schemes.
- `internal/telemetry`: file exporter rejects absolute/escaping paths and symlinked dirs/files under the project telemetry root.
- `internal/connectors/connsdk`: failed HTTP request spans/events preserve allowlisted status/attempt/retry attrs and omit response-body-like error details.
- `internal/cli`: OTLP exporter failure preserves command exit code/stdout JSON and emits only `warning: telemetry:` redacted stderr.
- CLI help/docs golden: root help/config docs support values and HTTP attr wording are consistent.

## Final focused review-fix plan — PR #459 residuals at `75433cefa9a00671b06c6c3e83bcde1e4730211c`

Execution decision: `local_critical_path` — same isolated issue worktree/branch, no subagent tool, no Claude/Copilot request per user instruction.

Accepted residuals and slices:

1. Ambient OTel env bypass (MED): own OTLP env before exporter construction. Support only trusted exporter opt-in plus validated endpoint/default, add `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` as a validated endpoint alias, warn without values for unsafe endpoint/unsupported header env, neutralize unsupported `OTEL_EXPORTER_OTLP_*` env while calling `otlptracehttp.New`, and keep warnings on project stderr only.
2. ADR docs (LOW): add superseding note to ADR 0004 that config-file OTLP endpoint language is replaced by trusted env/flag endpoint behavior.
3. Root help/goldens (LOW): ensure root help lists all exporter values `none`, `off`, `file`, and `otlp`.
4. Config docs/website wording (LOW): say both OTLP exporter and endpoint must come from trusted env/flag; config-file endpoint alone is ignored.

Planned red tests before production edits:

- `internal/cli`: unsafe `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` with userinfo/query is rejected with redacted `warning: telemetry:` on project stderr, does not call the collector, and emits nothing/raw-secret to process stderr.
- `internal/cli`: unsupported `OTEL_EXPORTER_OTLP_HEADERS` is ignored with redacted project warning; no raw process stderr and no Authorization header reaches the collector.
- Docs/help focused checks fail until root/config help, ADR, docs/CLI, website source, generated website data, and goldens match trusted-env wording.

## Final security hardening plan — PR #459 residual at head `216b5e076d82302621574a9fb4fa71c8acb79204`

Execution decision: `local_critical_path` — same isolated issue worktree/branch, no subagent tool, no Claude/Copilot request per user instruction.

Accepted residuals and slices:

1. SDK/provider/resource env bypass (MED): sanitize or temporarily unset unsupported SDK-level OpenTelemetry environment while constructing the explicit resource and `sdktrace.NewTracerProvider`; warn by environment variable name only. Cover `OTEL_RESOURCE_ATTRIBUTES`, `OTEL_SERVICE_NAME`, `OTEL_TRACES_SAMPLER`, `OTEL_TRACES_SAMPLER_ARG`, span/attribute/link/event limit env, BSP env, and experimental Go OTel env that can alter SDK behavior or log parse errors.
2. Safe resource: pass an explicit project-controlled resource while SDK env is unset so ambient resource attributes cannot export secrets.
3. Regression tests: prove `OTEL_RESOURCE_ATTRIBUTES=api_key=...` does not appear in file exporter output or OTLP payloads; prove invalid sampler args do not reach raw process stderr.
4. Config test isolation: include `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` in config env cleanup/list if not already covered.

Planned red tests before production edits:

- `internal/cli`: file exporter with ambient `OTEL_RESOURCE_ATTRIBUTES=api_key=<synthetic>`/`OTEL_SERVICE_NAME=<synthetic>` exports spans without those markers, warns only by env name, and emits no raw process stderr.
- `internal/cli`: OTLP exporter with ambient `OTEL_RESOURCE_ATTRIBUTES=api_key=<synthetic>` does not send that marker to the collector payload.
- `internal/cli`: invalid `OTEL_TRACES_SAMPLER=traceidratio` + `OTEL_TRACES_SAMPLER_ARG=<synthetic-invalid>` emits only project `warning: telemetry:` by env name and nothing to raw process stderr.
- `internal/config`: env cleanup includes `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`.

## Narrow final alias/tracer-closure plan — PR #459 residual at head `0fc39148004d699f35239a17418cd095bdd4a1ed`

Execution decision: `local_critical_path` — same isolated issue worktree/branch, no subagent tool, no Claude/Copilot request per user instruction.

Accepted residual and slice:

1. OTel Go self-observability alias gap: include both `OTEL_GO_X_OBSERVABILITY` and legacy/alias `OTEL_GO_X_SELF_OBSERVABILITY` in unsupported SDK env handling; warnings must name only the env var, never values.
2. Tracer creation env window: keep SDK env sanitized through `provider.Tracer("polymetrics.ai/pm")` creation, not only through `sdktrace.NewTracerProvider`.
3. Focused tests: for both env names, prove warnings are project `warning: telemetry:` messages by env name only, raw process stderr stays empty, file/OTLP exports omit synthetic self-observability markers, and no self-observability/secret terms leak.

Planned red tests before production edits:

- `internal/cli`: file exporter with `OTEL_GO_X_OBSERVABILITY=<synthetic>` warns by name only, writes no raw process stderr, and exported telemetry omits the env value, `self_observability`, and SDK self-observability metric names.
- `internal/cli`: file exporter with `OTEL_GO_X_SELF_OBSERVABILITY=<synthetic>` has the same warning/stderr/leak constraints and fails until the alias is added to unsupported SDK env handling.
- Focused gate: `go test ./internal/cli ./internal/telemetry -run 'Telemetry|OTEL_GO_X' -count=1`.

## Commit/push checkpoints

1. Final alias/tracer-closure planning artifact checkpoint.
2. Final alias/tracer-closure red-test checkpoint with exact failing output recorded.
3. Final alias/tracer-closure implementation checkpoint.
4. Focused/full verification + PR body update checkpoint.
5. SDK env final hardening planning artifact checkpoint.
2. SDK env red-test checkpoint with exact failing output recorded.
3. SDK env implementation checkpoint.
4. Focused/full verification + PR body update checkpoint.
5. Final residual planning artifact checkpoint.
2. Final residual red-test checkpoint with exact failing output recorded.
3. Ambient OTel env hardening implementation checkpoint.
4. Docs/help/golden/website checkpoint.
5. Full verification + PR body update checkpoint.
6. Review-fix planning artifacts checkpoint.
2. Review-fix red-test checkpoint with exact failing output recorded.
3. Green security/config/path/event/OTLP implementation checkpoint.
4. Docs parity/generated artifact checkpoint if generated diffs are required.
5. Full verification + PR body update checkpoint.

## PR plan

Open stacked PR to base `feat/cli-architecture-v2`, title `feat(obs): add opt-in OpenTelemetry tracing`, body includes `Refs #410` and `Refs #397`, GSD/TDD evidence, dependency evidence, gates, skill list, parity checklist, safety notes, and review route status.
