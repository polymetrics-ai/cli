# Phase 428 TDD Ledger

Issue: #428 — nativize agent namespace.
Invocation session: `issue-428-pi-openai-codex-gpt-5.6-sol-high-20260718T124925Z`
Model: `openai-codex/gpt-5.6-sol`; thinking: `high`
Starting HEAD: `235233f7cfde4a24612be6b0f95fb37a412d388a`

## GSD and skills

Doctor/list passed; `scripts/gsd prompt plan-phase 428 --skip-research` generated 10668 bytes and is being executed inline. `scripts/gsd prompt programming-loop init --phase 428 --dry-run` failed because the command is absent, so the manual universal-runtime-loop fallback is active.

Loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-documentation`, `golang-spf13-cobra`.

## High-finding correction ledger

Session: `issue-428-review-fix-pi-openai-codex-gpt-5.6-sol-high-20260718T132841Z`; model `openai-codex/gpt-5.6-sol`; thinking `high`; exact start `746b2a98b01ba1e119974e31569fc8deb06cd897`.

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| C0 | Planning | Update PLAN/TDD-LEDGER/VERIFICATION/RUN-STATE with accepted High, Sol/high identity, exact start, bounded slice, and gates | Complete before test/production edits |
| C1 | RED | Add fake-runtime action-boundary table for both levels and build/pull/ensure; `go test ./internal/cli -run '^TestAgentLeadingInvalidActionTokensCannotReachImageRuntime$' -count=1` | Failed as expected before production edits (`0.587s`) |
| C2 | GREEN | Agent-scoped pre-Cobra exact-action boundary plus image-parent positional rejection before runtime lookup | Pass; 30/30 fake-runtime cases fail closed with zero lookups/files/runs |
| C3 | Refactor | Focused/race agent tests and base differential for invalid heads plus preserved valid/help/literal routes | Pass; focused `4.446s`, race `1.679s`, repeated boundary test `0.582s`, differential 35/35 exact |
| C4 | Verify | gofmt, vet, build, diff/scope/dependency checks; full CLI only if needed | Pass; full CLI `234.335s`; vet/build/diff/scope/dependencies clean |

Correction RED cases cross `build`, `pull`, and `ensure` with these leading forms at both `agent` and `agent image`: `--unknown=x`, bare `--unknown`, short `-x`, assigned help-like `--help=false`, and literal `--`. Every case must return usage and leave fake runtime lookups, file checks, and runs empty. Exact agent help and exact actions remain valid; unknown/help/literal tokens after an exact action retain the established compatibility contract.

Exact correction RED was captured before production edits. All agent-level assigned unknown/help-like cases returned success and reached later actions; all image-level assigned unknown cases returned success and reached later actions; image-level assigned help-like cases performed a runtime lookup; image-level literal-boundary cases returned success and reached later actions. Bare/short cases already failed closed. The focused command exited 1 with package failure in `0.587s`.

GREEN inserts a pre-Cobra separator at the first non-exact agent/image action head and returns immediately so the existing legacy-tail normalizer cannot remove it. The image parent preserves the legacy-specific unknown-subcommand usage text while rejecting positional invalid heads before `LookPath`. Exact actions/help bypass the boundary; valid action tails retain existing unknown/help/literal normalization. No generic command or runtime surface was added.

## Planned red / green / refactor log

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| 0 | Planning | Create all six phase artifacts before production edits | Complete |
| 1 | RED | `go test ./internal/cli/ -run 'Agent|CobraRouterShellBuildsFreshHiddenWrapperTree' -count=1` after test-only edits | Failed as expected (build gate) |
| 2 | GREEN | Native tree, typed request, agent-only compatibility, injected image runtime, validation | Pass (`4.408s`; expanded invalid-help rerun `4.480s`) |
| 3 | Refactor | Focused router/golden/full CLI and legacy differential | Pass |
| 4 | Full gate | gofmt, vet, full tests, build, `make verify` | Pass |
| 5 | Parity/safety | built binary help/plan only; docs/website/generated/runtime dependency-free/scope guards | Pass |

## Planned RED coverage

- `agent` is native and absent from legacy wrappers; native `plan`, `image`, `image build|pull|ensure`, and hidden positional `help` nodes exist.
- `plan --request` is `stringArray`, bare value is `true`, unknown flags remain tolerated, no-file completion seam exists, repeated values remain last-wins.
- Plan spaced/assigned/bare/repeated/unknown/extra-positional forms preserve deterministic exact text and JSON output.
- Bare/text/JSON/long/short/positional help routes preserve the canonical manual; invalid agent/image actions remain usage errors.
- Global root/json/plain/no-input assigned booleans and configured JSON override behavior remain stable.
- Every image action runs only against an injected fake: build command/context/Containerfile path, pull, ensure-present, ensure-pull, exact text/JSON kinds/status, error propagation, and deterministic invocation ordering.
- Request/control, project/build path, Podman binary, and image reference validation reject unsafe input before external execution.
- Legacy action-tail `--help`/`-h` and continuation after literal `--` remain compatible without Phase 19 help churn.

## Exact RED

Captured after all focused test-only edits and before any production-code edit:

```text
# polymetrics.ai/internal/cli [polymetrics.ai/internal/cli.test]
internal/cli/agent_cli_test.go:276:11: undefined: runAgentImageAction
internal/cli/agent_cli_test.go:307:11: undefined: newRootCmdWithAgentImageRuntime
internal/cli/agent_cli_test.go:340:11: undefined: runAgentImageAction
internal/cli/agent_cli_test.go:359:9: undefined: runAgentImageAction
internal/cli/agent_cli_test.go:365:8: undefined: runAgentImageAction
FAIL\tpolymetrics.ai/internal/cli [build failed]
FAIL
```

The missing injected-runtime/action seams are intentional RED. The same test-only checkpoint also specifies native tree ownership, plan/help/global/compatibility behavior, unsafe-input rejection, deterministic output, and every image action without invoking Podman or Docker.

## Focused GREEN

```text
$ go test ./internal/cli/ -run 'Agent|CobraRouterShellBuildsFreshHiddenWrapperTree' -count=1
ok  \tpolymetrics.ai/internal/cli\t4.408s

$ go test ./internal/cli/ -run 'Agent|CobraRouterShellBuildsFreshHiddenWrapperTree' -count=1
ok  \tpolymetrics.ai/internal/cli\t4.480s
```

The second run includes expanded invalid image-action trailing-help coverage. Native `agent`, `plan`, `image`, `build`, `pull`, `ensure`, and hidden positional `help` nodes now own routing. `plan --request` is typed and the agent-only `parseFlags` call is removed. Image operations use an injected context-aware runtime seam; all success/error/ensure branches run against fakes. Agent-scoped normalization preserves legacy trailing help and literal separator behavior. Request, build root, Podman binary, and image references are validated before any runtime lookup or execution.

## Refactor, full GREEN, and parity

| Command / gate | Result |
|---|---|
| Final focused agent/router | Pass (`4.386s`) |
| Focused agent/router/golden | Pass (`11.408s`) |
| Standalone golden | Pass (`5.816s`; final rerun `6.054s`) |
| Full `internal/cli/...` | Pass (`235.686s`) |
| Runtime dependency-free config/RLM/worker packages | Pass; final config `0.411s`, RLM `0.553s`, router `0.864s`, worker `1.204s` |
| Focused race gate | Pass (`1.751s`) |
| Legacy differential | 25/25 exact exit/stdout/stderr matches: 20 help/plan/global/compatibility cases plus 5 missing/invalid image-action help cases |
| Built binary | Help routes byte-identical (`450` bytes); deterministic plan; invalid action exit `2`; invalid assigned boolean and unsafe request exit `3` |
| Docs/manual parity | Temp CLI docs byte diff clean; temp and tracked connector docs validation pass |
| Website parity | `npm --prefix website run gen:docs` wrote 11 pages; tracked diff clean |
| Formatting/static/build | `gofmt -w cmd internal`, `go vet ./...` (`4.162s`), `go build ./cmd/pm` (`3.829s`) pass |
| Full repository tests | Pass (`real 345.240s`; CLI `238.990s`, certify `341.079s`) |
| `make verify` | Pass (`real 25.853s`, cached tests; smoke OK; lint `0 issues`; 547 connectors / 0 findings) |
| Scope/dependency guards | No go.mod/go.sum, connector-def, docs, website, golden, or unrelated namespace delta; `git diff --check` pass |

All image-action success/error/ensure branches were exercised only through injected fakes and temporary roots. No Podman/Docker command, image build, image pull, publish, Temporal, PostgreSQL, or Dragonfly service was invoked. Invalid image-action differential checks performed executable lookup only and never executed the configured binary. The required `make verify` local sample smoke followed reverse ETL plan → preview → approval → run without external writes.
