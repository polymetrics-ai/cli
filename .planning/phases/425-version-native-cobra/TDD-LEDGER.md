# Phase 425 TDD Ledger

Issue: #425 — nativize version namespace.

Invocation session: `issue-425-pi-openai-codex-gpt-5.6-sol-high-20260718T095316Z`  
Model: `openai-codex/gpt-5.6-sol`  
Thinking: `high`  
Starting HEAD: `479a62f930e7c8a9a51ba0b3deb088bf3aad3ecc`

## GSD command evidence

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt plan-phase 425 --skip-research --model=openai-codex/gpt-5.6-sol --thinking=high
scripts/gsd prompt programming-loop init --phase 425 --dry-run --model=openai-codex/gpt-5.6-sol --thinking=high
```

Results: doctor/list passed (69 commands); plan-phase prompt generated; programming-loop failed with exact stderr `scripts/gsd: unknown GSD command: programming-loop` and exit 1. Manual universal-loop fallback is active.

## Skills loaded

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`.

## Red / green / refactor log

| Step | Kind | Command | Result | Evidence |
|---:|---|---|---|---|
| 0 | Planning | Create all six issue-local phase artifacts | Pass | Completed before production edits. |
| 1 | RED | `go test ./internal/cli/ -run 'Version|CobraRouterShellBuildsFreshHiddenWrapperTree' -count=1` | Fail as expected | Legacy/native registration count mismatched and `version` retained `DisableFlagParsing`. |
| 2 | GREEN | Native Cobra registration + minimal handler adaptation | Pending | Not started. |
| 3 | Refactor | Focused version/router/golden and full internal CLI | Pending | Not started. |
| 4 | Full gate | gofmt/vet/tests/build/`make verify` + parity/safety checks | Pending | Not started. |
| 5 | Delivery | coherent commits and branch push, no PR | Pending | Not started. |

## Planned RED coverage

- Native Cobra registration and removal from `cobraLegacyCommands`.
- Bare deterministic version output.
- Flag help (`--help`, `-h`) and positional `help` compatibility.
- JSON version output and JSON manual output.
- Unknown flag rejection remains usage exit 2.
- Invalid action remains usage exit 2 and does not render `CommandManual`.

## Exact RED output

Captured after test-only edits and before any production-code edit:

```bash
gofmt -w internal/cli/cobra_router_test.go internal/cli/version_cli_test.go
go test ./internal/cli/ -run 'Version|CobraRouterShellBuildsFreshHiddenWrapperTree' -count=1
```

```text
--- FAIL: TestCobraRouterShellBuildsFreshHiddenWrapperTree (0.00s)
    cobra_router_test.go:55: expectedHidden covers 21 commands, legacy commands plus native commands registers 22
--- FAIL: TestVersionCommandIsNativeCobraLeaf (0.00s)
    cobra_router_test.go:213: version command must use native Cobra flag parsing
FAIL
FAIL\tpolymetrics.ai/internal/cli\t0.612s
FAIL
```

The behavior-focused tests passed under the legacy wrapper; the RED is specifically the intended parser-ownership/registration gap.
