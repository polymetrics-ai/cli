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
- `vercel-react-best-practices` — website generated-data/docs parity sanity; no React component changes.
- `vercel-composition-patterns` — website composition awareness; no reusable component changes.
- `caveman` — final worker handoff compression only.

Note: required stack skills `.pi/skills/go-implementation/SKILL.md` and `.pi/skills/ts-website/SKILL.md` are absent in this checkout (`.pi/skills` contains only `gsd-core`); loaded repo/global Go and website skills above instead and will record this in the handoff.

## GSD command evidence

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 410 --skip-research >/tmp/gsd-plan-phase-410.txt
scripts/gsd prompt programming-loop init --phase 410 --dry-run >/tmp/gsd-programming-loop-410.txt
scripts/gsd prompt plan-phase 410 --skip-research >/tmp/gsd-plan-phase-410-reviewfix.txt
scripts/gsd prompt programming-loop init --phase 410 --dry-run >/tmp/gsd-programming-loop-410-reviewfix.txt
scripts/gsd prompt plan-phase 410 --skip-research >/tmp/gsd-plan-phase-410-reviewfix-rerun.txt
scripts/gsd prompt programming-loop init --phase 410 --dry-run >/tmp/gsd-programming-loop-410-reviewfix-rerun.txt
scripts/gsd prompt plan-phase 410 --skip-research >/tmp/gsd-plan-phase-410-final-reviewfix.txt
scripts/gsd prompt programming-loop init --phase 410 --dry-run >/tmp/gsd-programming-loop-410-final-reviewfix.txt
scripts/gsd prompt plan-phase 410 --skip-research >/tmp/gsd-plan-phase-410-final-sdk-env.txt
scripts/gsd prompt programming-loop init --phase 410 --dry-run >/tmp/gsd-programming-loop-410-final-sdk-env.txt
scripts/gsd prompt plan-phase 410 --skip-research >/tmp/gsd-plan-phase-410-final-alias.txt
scripts/gsd prompt programming-loop init --phase 410 --dry-run >/tmp/gsd-programming-loop-410-final-alias.txt
```

Result:

- `doctor`: pass (rerun 2026-07-18; 69 commands).
- `plan-phase`: generated 142-line prompt (initial and review-fix rerun).
- `programming-loop`: failed with `scripts/gsd: unknown GSD command: programming-loop`; manual GSD/TDD fallback recorded in PLAN/RUN-STATE.
- Review-fix rerun in this Pi worker: `scripts/gsd doctor` passed; `plan-phase` prompt regenerated; `programming-loop` still unavailable, so manual GSD/TDD fallback remains active.
- Final focused review-fix rerun at PR #459 head `75433cefa9a00671b06c6c3e83bcde1e4730211c`: `scripts/gsd doctor` passed; `plan-phase` prompt regenerated; `programming-loop` still unavailable, so manual GSD/TDD fallback remains active.
- Final SDK-env hardening rerun at PR #459 head `216b5e076d82302621574a9fb4fa71c8acb79204`: `scripts/gsd doctor` passed; `plan-phase` prompt regenerated; `programming-loop` still unavailable, so manual GSD/TDD fallback remains active.
- Narrow final alias/tracer-closure rerun at PR #459 head `0fc39148004d699f35239a17418cd095bdd4a1ed`: `scripts/gsd doctor` passed; `plan-phase` prompt regenerated; `programming-loop` still unavailable, so manual GSD/TDD fallback remains active.

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

## Final residual review-fix accepted findings

| Finding | Disposition | Planned red evidence |
|---|---|---|
| MED ambient OTel env bypass can consume `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`/headers and write raw stderr or route to unvalidated collector | Accepted | CLI tests capture process stderr while setting unsafe traces endpoint and unsupported headers; expect only project `warning: telemetry:` with no marker and no unvalidated collector/header use |
| LOW ADR 0004 says OTLP endpoint via config | Accepted | ADR docs updated with superseding trusted-env/flag note |
| LOW root help/goldens omit full exporter values | Accepted | Golden/root help asserts `none`, `off`, `file`, `otlp` |
| LOW config docs/website imply config endpoint works when network telemetry env-enabled | Accepted | Reword docs/website to require both exporter and endpoint from trusted env/flag; config-file endpoint alone ignored |

## Final SDK-level env hardening accepted findings

| Finding | Disposition | Planned red evidence |
|---|---|---|
| MED SDK/provider/resource OpenTelemetry env bypass can read `OTEL_RESOURCE_ATTRIBUTES`, `OTEL_SERVICE_NAME`, `OTEL_TRACES_SAMPLER(_ARG)`, span limits, BSP limits, and experimental OTel env during resource/provider construction | Accepted | File and OTLP regression tests assert synthetic `api_key` resource attrs/service name never export and warnings name env only |
| MED invalid sampler args may hit raw process stderr before project warning/redaction | Accepted | CLI test captures raw process stderr while invalid sampler env is set; expects empty process stderr and project warning only by env name |
| LOW config test isolation missing traces endpoint cleanup/list if needed | Accepted | `clearBoundEnv`/env list includes `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` |

## Narrow final alias/tracer-closure accepted finding

| Finding | Disposition | Planned red evidence |
|---|---|---|
| MED `OTEL_GO_X_SELF_OBSERVABILITY` alias omitted and `provider.Tracer("polymetrics.ai/pm")` created after SDK env restore | Accepted | Focused CLI tests for `OTEL_GO_X_OBSERVABILITY` and `OTEL_GO_X_SELF_OBSERVABILITY` assert project warnings by env name only, no raw process stderr, and no self-observability/secret marker leakage from exported telemetry. |

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
| 11 | review-fix red | Test | `go test ./internal/telemetry ./internal/config ./internal/cli ./internal/connectors/connsdk ./internal/flow -run 'Telemetry|OTLP|Endpoint|Event|RecordError|FileExporter' -count=1` | Fail | Expected red before production edits. Exact key output: `internal/telemetry/telemetry_test.go:23:75: unknown field ProjectRoot in struct literal of type Config`; `internal/connectors/connsdk/telemetry_test.go:37:105: unknown field ProjectRoot in struct literal of type "polymetrics.ai/internal/telemetry".Config`; `internal/flow/telemetry_test.go:18:105: unknown field ProjectRoot in struct literal of type "polymetrics.ai/internal/telemetry".Config`; `telemetry_cli_test.go:92: telemetry output missing "pm.error.type"` with exported `exception` event; `telemetry_cli_test.go:119: stderr missing config-sourced OTLP warning: ""`; `telemetry_cli_test.go:167: stderr missing telemetry warning: ""`; exit code 1. |
| 12 | review-fix resume | Planning | `scripts/gsd doctor`; `scripts/gsd prompt plan-phase 410 --skip-research`; `scripts/gsd prompt programming-loop init --phase 410 --dry-run` | Pass/fallback | Doctor passed; plan prompt regenerated; programming-loop command still unavailable (`scripts/gsd: unknown GSD command: programming-loop`). Continuing local-critical-path manual GSD/TDD with existing red evidence and no subagent spawn. |
| 13 | review-fix red | Test | `go test ./internal/cli ./internal/connectors/connsdk ./internal/flow -run 'TestTelemetryFailedCommandSpanDoesNotExportRawError|TestRequesterDoFailedHTTPSpanHasSafeErrorAndEventAttrs|TestEngineRunFailedStepTelemetryRedactsError' -count=1` | Fail | Added stricter stable error class/code assertions before implementation. Exact key output: CLI telemetry missing `internal_error` and exported `pm.error.type=cli.cobraLegacyError`; connsdk telemetry contained forbidden `connsdk.HTTPError`; flow telemetry contained forbidden `errors.errorString`/`fmt.wrapError`; exit code 1. |
| 14 | review-fix green | Test | `go test ./internal/cli ./internal/connectors/connsdk ./internal/flow -run 'TestTelemetryFailedCommandSpanDoesNotExportRawError|TestRequesterDoFailedHTTPSpanHasSafeErrorAndEventAttrs|TestEngineRunFailedStepTelemetryRedactsError' -count=1` | Pass | Stable error metadata now uses allowlisted `pm.error.type`/`pm.error.code`/`pm.error.status_code` without SDK `exception.*` or Go wrapper type names. |
| 15 | review-fix green | Test | `go test ./internal/telemetry ./internal/config ./internal/cli ./internal/connectors/connsdk ./internal/flow -run 'Telemetry|OTLP|Endpoint|Event|RecordError|FileExporter' -count=1` | Pass | Review-fix focused telemetry/config/CLI/connsdk/flow test set passed. |
| 16 | review-fix green | Focused gates/smoke/docs | `go test ./internal/telemetry -count=1`; `go test ./internal/connectors/connsdk -run Telemetry -count=1`; `go test ./internal/cli -run 'Telemetry|Golden|Config|Agentic' -count=1`; `go test ./internal/app -run Telemetry -count=1`; `go test ./internal/flow -run Telemetry -count=1`; `go test ./internal/config -count=1`; file/off/secret smoke; OTLP endpoint smoke; golden/docs/website generation | Pass | Focused package gates, stdout/secret smoke, help parity smoke, golden update, `docs/cli` generation, and `npm --prefix website run gen:docs` passed. |
| 17 | review-fix green | Broad gates | `go vet ./...`; `go test ./...`; `go build ./cmd/pm` | Pass | Full Go vet/test/build passed. |
| 18 | review-fix verify | `make verify` | Fail | `tidy-check` promoted existing `go.opentelemetry.io/otel/trace v1.44.0` from indirect to direct because event attrs use `trace.WithAttributes`; no new version/checksum. Rerun after accepting tidy diff required. |
| 19 | review-fix verify | `make verify` after review-fix commit | Pass | fmt, tidy-check, vet, 20m tests, build, docs validate, smoke, lint, and connectorgen validate passed. |
| 20 | final residual red | Test | `go test ./internal/cli -run 'TestTelemetryRejectsUnsafeAmbientOTLPTracesEndpointBeforeExporter|TestTelemetryNeutralizesUnsupportedAmbientOTLPHeaders' -count=1` | Fail | Expected red before production edits. Exact key output: `project stderr missing redacted OTLP endpoint warning: ""`; `project stderr missing unsupported headers warning: ""`; exit code 1. |
| 21 | final residual green | Test | `go test ./internal/cli -run 'TestTelemetryRejectsUnsafeAmbientOTLPTracesEndpointBeforeExporter|TestTelemetryNeutralizesUnsupportedAmbientOTLPHeaders' -count=1` | Pass | Unsafe traces endpoint now rejected before exporter construction; unsupported OTLP headers warned/neutralized with no raw process stderr or Authorization header. |
| 22 | final residual green | Focused tests/docs | `go test ./internal/telemetry ./internal/config ./internal/cli -run 'Telemetry|TestLoadTelemetry|Golden|Config' -count=1`; `go test ./internal/telemetry -count=1`; `go test ./internal/connectors/connsdk -run Telemetry -count=1`; `go test ./internal/app -run Telemetry -count=1`; `go test ./internal/flow -run Telemetry -count=1` | Pass | Ambient env hardening, config alias, golden transcripts, generated docs drift, and affected telemetry consumers green after docs/golden/website regeneration. |
| 23 | final residual verify | Broad gates/parity | `gofmt -w cmd internal`; `go vet ./...`; `go test ./...`; `go build ./cmd/pm`; `make verify`; runtime help/docs parity; `git diff --check`; `git diff -- go.mod go.sum` | Pass | Full final gates passed; no go.mod/go.sum diff; docs generation and website data clean. |
| 24 | sdk-env final hardening plan | Planning | Phase artifacts updated before test/production edits; `scripts/gsd doctor`; `scripts/gsd prompt plan-phase 410 --skip-research`; `scripts/gsd prompt programming-loop init --phase 410 --dry-run` | Pass/fallback | Doctor passed; plan prompt regenerated; programming-loop still unavailable (`scripts/gsd: unknown GSD command: programming-loop`). Continuing local-critical-path manual GSD/TDD. |
| 25 | sdk-env final hardening red | Test | `go test ./internal/cli ./internal/config -run 'TestTelemetrySanitizesSDKResourceEnvForFileExporter|TestTelemetrySanitizesSDKResourceEnvForOTLPExporter|TestTelemetrySanitizesInvalidSamplerEnvBeforeProvider|TestAllBoundEnvVarsIncludesOTLPTracesEndpoint' -count=1` | Fail | Expected red before production edits. Exact key output: `stderr missing SDK env warnings by name: ""`; `stderr missing resource env warning by name: ""`; `process stderr = "2026/07/18 04:46:28 parsing sampler argument: strconv.ParseFloat: parsing \"invalid-pm_invalid_sampler_marker\": invalid syntax\n", want empty`; `allBoundEnvVars missing OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`; exit code 1. |
| 26 | sdk-env final hardening green | Test | `go test ./internal/cli ./internal/config -run 'TestTelemetrySanitizesSDKResourceEnvForFileExporter|TestTelemetrySanitizesSDKResourceEnvForOTLPExporter|TestTelemetrySanitizesInvalidSamplerEnvBeforeProvider|TestAllBoundEnvVarsIncludesOTLPTracesEndpoint' -count=1`; `go test ./internal/telemetry ./internal/config ./internal/cli -run 'Telemetry|TestLoadTelemetry|Config' -count=1` | Pass | SDK resource/service/sampler env now warned by name only and unset around explicit safe resource/provider construction; file and OTLP payloads omit synthetic `api_key` markers; process stderr remains empty. |
| 27 | sdk-env final hardening gates | Focused/full Go gates | `gofmt -w cmd internal`; `go vet ./...`; `go build ./cmd/pm`; `go test ./...` | Pass | `go test ./...` first two cold-cache attempts hit harness timeouts at 20m/30m before all packages printed; package segment rerun and final full rerun passed with cached packages. |
| 28 | sdk-env final hardening verify | Full verify/docs/diff | `make verify`; temp `go run ./cmd/pm docs generate --dir "$TMP_DOCS/cli" --connectors-dir "$TMP_DOCS/connectors"` + `diff -ru docs/cli "$TMP_DOCS/cli"`; `git diff --check`; `git diff -- go.mod go.sum` | Pass | `make verify` passed: fmt, tidy-check, vet, `go test -timeout 20m ./...`, build, docs validate, smoke, lint, connectorgen validate. Docs generation diff clean; no go.mod/go.sum diff. |
| 29 | final alias/tracer-closure plan | Planning | Phase artifacts updated before tests/production edits; `scripts/gsd doctor`; `scripts/gsd prompt plan-phase 410 --skip-research`; `scripts/gsd prompt programming-loop init --phase 410 --dry-run` | Pass/fallback | Doctor passed; plan prompt regenerated; programming-loop still unavailable (`scripts/gsd: unknown GSD command: programming-loop`). Continuing local-critical-path manual GSD/TDD. |
| 30 | final alias/tracer-closure red | Test | `go test ./internal/cli ./internal/telemetry -run 'Telemetry|OTEL_GO_X' -count=1` | Fail | Expected red before production edits. First 120s run hit harness timeout; 600s rerun failed as expected: `TestTelemetryOTEL_GO_XSelfObservabilityEnvWarningIsProjectOnly`: `project stderr missing OTEL_GO_X_SELF_OBSERVABILITY warning by env name: "warning: telemetry: unsupported OpenTelemetry SDK environment variable OTEL_RESOURCE_ATTRIBUTES ignored; configure telemetry only through trusted pm env/flag\n"`; `internal/telemetry` passed. |
| 31 | final alias/tracer-closure green | Test | `gofmt -w internal/cli/telemetry_cli_test.go internal/telemetry/telemetry.go && go test ./internal/cli ./internal/telemetry -run 'Telemetry|OTEL_GO_X' -count=1` | Pass | `OTEL_GO_X_OBSERVABILITY` and `OTEL_GO_X_SELF_OBSERVABILITY` focused tests passed; warnings project-only, no raw process stderr, exported telemetry leak checks clean. |
| 32 | final alias/tracer-closure broad gate | Test | `go vet ./...`; `go test ./...` (twice) | Mixed | `go vet ./...` passed. `go test ./...` hit Go default 10m package timeout twice in pre-existing slow `internal/cli`/`internal/connectors/certify` tests while loading connector bundles (`TestGoldenTranscripts`/`TestGitHubDestructiveCommandRequiresTypedConfirmation`; certify sabotage/secret-leak stages). Reran with verify-equivalent 20m timeout. |
| 33 | final alias/tracer-closure timeout-adjusted broad gate | Test | `go test -timeout 20m ./internal/cli ./internal/connectors/certify -count=1`; `go test -timeout 20m ./...` | Pass | Slow packages passed with 20m timeout (`internal/cli` 651s, `internal/connectors/certify` 1109s for package rerun); full 20m suite passed. |
| 34 | final alias/tracer-closure verify | Full verify/diff | `go build ./cmd/pm`; `make verify`; `git diff --check`; `git diff -- go.mod go.sum` | Pass | `make verify` passed: fmt, tidy-check, vet, `go test -timeout 20m ./...`, build, docs validate, smoke, lint, connectorgen validate. No go.mod/go.sum diff. |
| 35 | final alias/tracer-closure push | PR update | PR body update via GitHub API; `git push origin feat/410-otel-tracing` | Pass | PR #459 body updated without Claude/Copilot; branch pushed to origin. |

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
