# Phase 402 Plan — Config env migration

Issue: polymetrics-ai/cli#402
Parent: #397 / PR #438
Branch: `refactor/402-config-env-migration`
Base: `feat/cli-architecture-v2` @ `14f5e433`
Execution decision: `local_critical_path` — Pi worker already isolated in cwd/branch; no subagent tool available to worker.
Review-fix cycle: PR #448 findings accepted; no merge, no deps, no parent/shared orchestration edits.

## GSD adapter

- `scripts/gsd doctor` — pass.
- `scripts/gsd prompt plan-phase 402 --skip-research >/tmp/gsd-plan-phase-402.prompt` — pass.
- `scripts/gsd prompt programming-loop init --phase 402 --dry-run >/tmp/gsd-programming-loop-402.prompt` — blocked: `scripts/gsd: unknown GSD command: programming-loop`; manual GSD fallback active using `gsd-universal-runtime-loop.md` and issue contract.
- Review-fix 2026-07-16: `scripts/gsd doctor` pass; `scripts/gsd prompt plan-phase 402 --skip-research >/tmp/gsd-plan-phase-402-reviewfix.prompt` pass; `scripts/gsd prompt programming-loop init --phase 402 --dry-run >/tmp/gsd-programming-loop-402-reviewfix.prompt` still blocked with unknown command, manual GSD fallback remains active.

## Required skills loaded

- `gsd-core`, `caveman`.
- Go: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-concurrency`.
- Cobra/Viper due config/Cobra plumbing: `golang-spf13-cobra`, `golang-spf13-viper`.
- Docs if caveat updated: `golang-documentation`.
- Review-fix hardening: `golang-lint` for `go vet`/quality-gate disposition.

Rule anchors for handoff: CLI stdout/stderr + command tests; testing named table tests/red-first; errors wrapped/lowercase/single handling; security no secret values/user env trust boundary; safety zero/nil/defaults; context propagation/cancel; design no globals/dependency injection; structs small typed config; concurrency worker lifecycle/cancel; viper instance/test isolation/no AutomaticEnv; cobra fresh tree per test; lint/vet after significant changes.

## Scope / exclusions

Allowed production scope:

- `internal/config` minimal explicit-source API for injection.
- `internal/runtimecheck` FromConfig primary, FromEnv compatibility.
- CLI plumbing in `internal/cli` for one resolved config injection.
- Named readers: runtimecheck consumers, worker CLI/submit, schedule select/crontab, agent image CLI, RLM non-secret settings.
- Focused tests; config docs caveat parity if behavior becomes active.

Strict exclusions:

- `credentials add --from-env` raw user-named env values.
- connector certify credsfile and secret env scanners.
- provider API-key secret intake (`PM_LLM_API_KEY`, `OPENROUTER_API_KEY`) and container env forwarding.
- generic shell/HTTP/SQL write surfaces, connector bundles, unrelated namespaces.
- Runtime service startup or credentialed checks.

## Env-reader classification before edits

| Location | Current raw read | Classification | Planned action |
|---|---|---|---|
| `internal/config/config.go` `os.LookupEnv` | allowlisted typed config bindings | typed config loader | keep; add explicit-source metadata only if needed |
| `internal/runtimecheck/runtimecheck.go` | `POLYMETRICS_POSTGRES_URL`, `POLYMETRICS_DRAGONFLY_ADDR`, `POLYMETRICS_TEMPORAL_ADDR` | config-shaped runtime settings | add `FromConfig(config.Config)`; make `FromEnv` delegate through config load fallback |
| `internal/cli/cli.go` `runRuntime` | `runtimecheck.FromEnv()` | config-shaped runtime invocation | inject resolved config; call `FromConfig` |
| `internal/cli/runtime_helpers.go` | `runtimecheck.FromEnv()` | config-shaped runtime ETL path | inject resolved config from `runETL` |
| `internal/schedule/select.go` | `POLYMETRICS_TEMPORAL_ADDR` | config-shaped scheduler backend opt-in | add typed/narrow backend config; CLI injects explicit temporal addr |
| `internal/schedule/crontab.go` | `PM_CRONTAB_FILE` | config-shaped crontab redirection seam | add `CrontabBackend.File`; CLI injects `schedule.crontab_file`; preserve compatibility/certify semantics |
| `internal/cli/schedule.go` | SelectBackend/CrontabBackend | config-shaped schedule invocation | inject resolved config into install/remove |
| `internal/cli/agent_image_cli.go` | `POLYMETRICS_RLM_IMAGE`, `POLYMETRICS_PODMAN_BIN` | config-shaped RLM image/bin | inject `cfg.RLM.Image/PodmanBin` |
| `internal/cli/worker_cli.go` | `POLYMETRICS_TEMPORAL_ADDR` | config-shaped worker Temporal opt-in | inject explicit runtime temporal setting; preserve empty default golden |
| `internal/worker/submit.go` | `POLYMETRICS_PODMAN_BIN`, `POLYMETRICS_RLM_IMAGE` | config-shaped worker activity defaults | review-fix: remove ambient config reload from default activities; CLI passes typed activities from invocation config |
| `internal/cli/rlm_cli.go` | Agent config/fake/embedded env reads | config-shaped non-secret RLM settings | inject typed config for non-secret settings; keep API keys env-only |
| `internal/cli/extract_cli.go` | LLM provider/base/model env via `LLMConfigFromEnv` | config-shaped non-secret LLM settings + secret API key seam | inject typed non-secret settings; keep secret key env-only |
| `internal/worker/podman_cmd.go` | `os.LookupEnv` over `EnvPass` | provider/container env forwarding incl secrets | exclude; no change |
| `internal/cli/cli.go` credentials add | user-named `--from-env` | credential secret intake | exclude; no change |
| `internal/connectors/certify/*` | secret env, credsfile, `PM_CRONTAB_FILE` save/restore | certify secret/test seams | exclude except behavior preserved through CLI config load |
| `internal/cli/golden_transcript_test.go` | update golden env flag | test-only seam | exclude |

## Delivered implementation matrix

| Scope | Delivery |
|---|---|
| `internal/config` | Added explicit-source metadata (`ExplicitKeys`, `IsExplicit`) without global Viper state or AutomaticEnv. |
| `internal/runtimecheck` | `FromConfig(config.Config)` primary; `FromEnv` compatibility delegates through typed load. |
| CLI config injection | `Run` resolves config once, then Cobra legacy handlers inject it into runtime, ETL runtime recording, worker, schedule, agent image, RLM, extract, and flow RLM paths. |
| Schedule | Added `BackendConfig`, `SelectBackendFromConfig`, and `CrontabBackend.File`; CLI injects explicit Temporal and crontab file; legacy `SelectBackend` remains compatible. |
| Worker/RLM | Worker status/serve and RLM agent use explicit typed Temporal; review-fix threads typed image/bin into `pm worker serve` activities and removes ambient config reload from worker defaults; RLM image/bin/fake/embedded and non-secret LLM provider/base/model come from typed config; API keys/env forwarding remain env-only. |
| Perf runtime compare | Review-fix threads CLI-resolved runtime endpoints into `pm perf compare --runtime`; `runtimecheck.FromEnv` remains available but perf no longer calls it. |
| Docs | Updated embedded config help, `docs/cli/config.md`, website source, and generated website docs data. |

## Slice plan

1. Planning checkpoint ✅
   - Create phase artifacts and record GSD adapter fallback, skills, classification, verification plan.
   - Commit/push plan-only checkpoint if green enough.

2. Red tests ✅
   - Add focused failing tests for runtimecheck FromConfig/FromEnv alias, CLI config-file injection for runtime/worker/schedule/agent image/RLM, and save/restore crontab behavior.
   - Capture exact red output in `TDD-LEDGER.md`.

3. Green implementation ✅
   - Add `internal/config` explicit key metadata helper if needed for worker/RLM/schedule opt-in semantics while preserving goldens.
   - Add runtimecheck FromConfig and compatibility FromEnv.
   - Inject config through Cobra wrapper/legacy handlers; migrate named call sites.
   - Add worker typed activities submitter and RLM config helpers; keep API keys/env forwarding excluded.
   - Add CrontabBackend file field and backend config injection; preserve `PM_CRONTAB_FILE` behavior.

4. Docs parity slice ✅
   - Update `docs/cli/config.md`, `website/content/docs/cli-reference.mdx`, and generated website data only if caveat changes.
   - Runtime help strings unchanged unless explicitly needed; golden transcripts should remain byte-identical.

5. Review-fix slice ✅ (`#448`)
   - Red tests: worker serve captures typed image/podman activity config without Temporal startup; perf compare captures config-file runtime endpoints without services.
   - Green implementation: CLI passes typed `worker.NewPodmanActivities` into worker serve; worker default activities no longer reload config; CLI passes `runtimecheck.FromConfig(cfg)` into perf compare, and perf no longer falls back to `runtimecheck.FromEnv`.
   - Global check: grep CLI/runtime issue scope for remaining `pmconfig.Load(Options{})`, `runtimecheck.FromEnv()`, and config-shaped raw env reads; preserve documented secret/user env exclusions.
   - Verification: run requested focused package tests, CLI gates, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, diff checks, and go.mod/go.sum diff.
   - Review route: do not request Claude/Copilot in this cycle; parent ledger stale finding is parent-owned and handoff-only.

## Planned tests / validations

- `go test ./internal/runtimecheck/... -count=1`
- `go test ./internal/schedule/... -count=1`
- `go test ./internal/perf/... -count=1`
- `go test ./internal/worker/... -count=1`
- `go test ./internal/cli/ -run 'Golden|Config|Runtime|Perf|Worker' -count=1`
- `go test ./internal/cli/ -run Certify -count=1`
- Required full gates from issue when green.

## Parity stance

Golden transcripts remain byte-identical. Config docs caveat changed because behavior is active in migrated call sites; embedded help, `docs/cli/config.md`, website source, and generated website data are aligned. Hidden `worker --help` remains unavailable as pre-existing hidden-command behavior; visible namespace help/bare namespace checks passed for runtime, agent, rlm, and schedule.
