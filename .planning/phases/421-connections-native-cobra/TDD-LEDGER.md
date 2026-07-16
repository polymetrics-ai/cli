# Phase 421 TDD Ledger

Issue: #421 — nativize connections namespace.

## Skills loaded

`gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`, `golang-safety`.

Repo skill gap: `.pi/skills/go-implementation/SKILL.md` was required by worker instructions but is absent in this checkout (`ENOENT`); global Go skills above are loaded and used.

Rule anchors:

- `golang-how-to`: CLI command tree routes to `golang-spf13-cobra` + `golang-cli`; tests route to `golang-testing`; args/I/O route to `golang-security` + `golang-safety`.
- `golang-cli`: preserve exit codes, stdout/stderr discipline, CLI unit tests, and no noisy usage walls.
- `golang-testing`: #1 named table tests, #3 independent tests, #5 observable behavior/public contract over implementation-only details.
- `golang-error-handling`: #1 check returned errors, #2 wrap/add context when propagating, #7 log-or-return not both, #9 no panic for expected errors.
- `golang-documentation`: concise CLI docs, no invented behavior, preserve safety wording; application CLI help is primary documentation.
- `golang-spf13-cobra`: best practices #1 RunE, #3 Args validators, #4 Out/Err writers, #5 fresh command tree; flags guidance for `StringArray`, `NoOptDefVal`, and unknown-flag compatibility.
- `golang-security`: trust-boundary questions #1-#3; no secrets; command args are untrusted.
- `golang-safety`: #2 safe assertions and #10 useful zero/default values.

## GSD command evidence

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 421 --skip-research >/tmp/gsd-plan-phase-421.prompt
scripts/gsd prompt programming-loop init --phase 421 --dry-run >/tmp/gsd-programming-loop-421.prompt
```

Result:

- `doctor`: pass.
- `plan-phase`: prompt written to `/tmp/gsd-plan-phase-421.prompt`.
- `programming-loop`: blocked by adapter registry (`scripts/gsd: unknown GSD command: programming-loop`); manual GSD fallback active using `.pi/prompts/pm-gsd-loop.md` + universal runtime loop.

## Red / green / refactor log

| Step | Kind | Command / test | Result | Notes |
|---:|---|---|---|---|
| 0 | Planning | Create PLAN/TDD-LEDGER/VERIFICATION/SUMMARY/RUN-STATE/PROMPTS | Green | Pre-production artifact checkpoint; no production code touched. |
| 1 | Red | `go test ./internal/cli/ -run 'Connections|CobraRouterShell' -count=1` | Fail | Native-subtree tests fail because `connections` remains legacy; invalid action opens project before usage classification. |
| 2 | Green | `go test ./internal/cli/... -run 'Connections|CobraRouterShell|Golden' -count=1` | Pass | Native connections parser green; golden transcripts unchanged. |
| 3 | Refactor | `gofmt -w internal/cli/cobra_router.go internal/cli/cli.go internal/cli/cobra_router_test.go internal/cli/connections_cli_test.go`; `go test ./internal/cli/ -run Certify -count=1`; `go vet ./...`; `go build ./cmd/pm` | Pass | Gofmt clean; certify re-entrancy, vet, and build preserved. |
| 4 | Gate | `go vet ./...`; `go test ./...`; `go build ./cmd/pm`; `make verify`; diff checks | Pending | Full local gates. |

## Planned red tests

- `TestConnectionsCommandIsNativeCobraSubtree`: current wrapper should fail because `connections.DisableFlagParsing` is true, no `create`/`list` subcommands exist, native flags are missing, and no completion seam exists.
- `TestConnectionsCreateFlagFormsPreserveLegacySemantics`: current wrapper/native metadata path should fail until pflag declarations and normalization exist; behavior cases cover space/equals forms, repeated singleton last-wins, repeated `--primary-key` accumulation, bare bool values, unknown flags, extra args, and late globals.
- `TestConnectionsInvalidActionIsUsage`: invalid actions must remain usage exit 2 without app/domain side effects.

## Exact red outputs

```bash
go test ./internal/cli/ -run 'Connections|CobraRouterShell' -count=1
```

```text
--- FAIL: TestCobraRouterShellBuildsFreshHiddenWrapperTree (0.00s)
    cobra_router_test.go:55: expectedHidden covers 21 commands, legacy commands plus native commands registers 22
--- FAIL: TestConnectionsCommandIsNativeCobraSubtree (0.00s)
    cobra_router_test.go:135: connections command must use native Cobra flag parsing
--- FAIL: TestConnectionsInvalidActionIsUsageBeforeProjectOpen (0.00s)
    connections_cli_test.go:130: Run(connections bogus --json) code = 1, want 2; stdout={
          "api_version": "polymetrics.ai/v1",
          "error": {
            "category": "internal",
            "code": "internal_error",
            "message": "open project at .polymetrics: stat .polymetrics: no such file or directory"
          },
          "kind": "Error"
        }
         stderr=error: open project at .polymetrics: stat .polymetrics: no such file or directory
FAIL
FAIL	polymetrics.ai/internal/cli	5.562s
FAIL
```

## Exact green outputs

```bash
gofmt -w internal/cli/cobra_router.go internal/cli/cli.go internal/cli/cobra_router_test.go internal/cli/connections_cli_test.go
go test ./internal/cli/ -run 'Connections|CobraRouterShell' -count=1
```

```text
ok  	polymetrics.ai/internal/cli	6.329s
```

```bash
go test ./internal/cli/... -run 'Connections|CobraRouterShell|Golden' -count=1
```

```text
ok  	polymetrics.ai/internal/cli	12.034s
```

```bash
go test ./internal/cli/ -run Certify -count=1
```

```text
ok  	polymetrics.ai/internal/cli	95.754s
```

```bash
go vet ./...
go build ./cmd/pm
```

```text
# no output; both commands exited 0
```
