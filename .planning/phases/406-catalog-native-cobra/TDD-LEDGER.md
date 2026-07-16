# Phase 406 TDD Ledger

Issue: #406 — nativize catalog namespace.

## Skills loaded

`gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`, `golang-safety`.

Rule anchors:

- `golang-how-to`: CLI task routes to `golang-spf13-cobra` + `golang-cli`; tests route to `golang-testing`; args/I/O route to `golang-security` + `golang-safety`.
- `golang-cli`: exit codes, stdout/stderr discipline, CLI unit testing, common mistakes around direct stdout and noisy usage.
- `golang-testing`: #1 named table tests, #3 independent tests, #5 observable behavior not implementation-only.
- `golang-error-handling`: #1 check returned errors, #2 wrap/add context where propagating, #7 log-or-return not both, #9 no panic for expected errors.
- `golang-documentation`: concise CLI docs, no invented behavior, preserve safety wording.
- `golang-spf13-cobra`: best practices #1 RunE, #3 Args validators, #4 Out/Err writers, #5 fresh command tree; flags reference StringArray vs StringSlice and NoOptDefVal use.
- `golang-security`: trust-boundary questions #1-#3, no secrets, command args untrusted.
- `golang-safety`: #2 safe assertions and #10 useful zero/default values.

## GSD command evidence

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 406 --skip-research >/tmp/gsd-plan-phase-406.prompt
scripts/gsd prompt programming-loop init --phase 406 --dry-run >/tmp/gsd-programming-loop-406.prompt
```

Result:

- `doctor`: pass.
- `plan-phase`: prompt written to `/tmp/gsd-plan-phase-406.prompt`.
- `programming-loop`: blocked by adapter registry (`scripts/gsd: unknown GSD command: programming-loop`); manual GSD fallback active.

## Red / green / refactor log

| Step | Kind | Command / test | Result | Notes |
|---:|---|---|---|---|
| 0 | Planning | Create PLAN/TDD-LEDGER/VERIFICATION/SUMMARY/RUN-STATE/PROMPTS | Green | Pre-production artifact checkpoint; no production code touched. |
| 1 | Red | `go test ./internal/cli/ -run 'Catalog|CobraRouterShell' -count=1` | Fail | Native subtree test fails on legacy `DisableFlagParsing`; invalid action fails before usage classification by opening `.polymetrics`. |
| 2 | Red | `go test ./internal/cli/ -run 'TestCatalogConnectionFlagFormsPreserveLegacySemantics' -count=1` | Fail | Native StringArray + NoOptDefVal needs a catalog-specific normalization shim to preserve legacy `--connection value` and repeated last-wins. |
| 3 | Green | `go test ./internal/cli/ -run 'Catalog|CobraRouterShell|Golden' -count=1` | Pass | Native catalog parser green; golden transcripts unchanged. |
| 4 | Refactor | `gofmt -w cmd internal`; `go test ./internal/cli/ -run Certify -count=1` | Pass | Gofmt clean; certify re-entrancy preserved. |
| 5 | Gate | `go vet ./...`; `go test ./...`; `go build ./cmd/pm`; `make verify`; diff checks | Pass | Full local gates green; no go.mod/go.sum diff. |

## Planned red tests

- `TestCatalogCommandIsNativeCobraSubtree`: current wrapper should fail because `catalog.DisableFlagParsing` is true, no `refresh`/`show` subcommands exist, and no native `--connection` flag metadata exists.
- `TestCatalogInvalidActionIsUsageBeforeConnectionValidation`: current legacy handler should fail because `pm catalog bogus --json` reports missing `--connection` as runtime error instead of usage exit 2.

## Exact red outputs

```bash
go test ./internal/cli/ -run 'Catalog|CobraRouterShell' -count=1
```

```text
--- FAIL: TestCatalogCommandIsNativeCobraSubtree (0.00s)
    cobra_router_test.go:89: catalog command must use native Cobra flag parsing
--- FAIL: TestCatalogInvalidActionIsUsageBeforeConnectionValidation (0.00s)
    catalog_cli_test.go:73: Run(catalog bogus --json) code = 1, want 2; stdout={
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
FAIL	polymetrics.ai/internal/cli	8.894s
FAIL
```

```bash
go test ./internal/cli/ -run 'TestCatalogConnectionFlagFormsPreserveLegacySemantics' -count=1
```

```text
--- FAIL: TestCatalogConnectionFlagFormsPreserveLegacySemantics (2.57s)
    --- FAIL: TestCatalogConnectionFlagFormsPreserveLegacySemantics/space_form (0.51s)
        catalog_cli_test.go:116: Run([catalog show --connection space-form --root /var/folders/tk/bmp_tx0976s4rkh1phvrpjlw0000gn/T/TestCatalogConnectionFlagFormsPreserveLegacySemantics3946397549/001 --json]) missing "connection \"space-form\" not found"; stdout={
              "api_version": "polymetrics.ai/v1",
              "error": {
                "category": "internal",
                "code": "internal_error",
                "message": "connection \"true\" not found"
              },
              "kind": "Error"
            }
             stderr=error: connection "true" not found
    --- FAIL: TestCatalogConnectionFlagFormsPreserveLegacySemantics/repeated_last_wins (0.51s)
        catalog_cli_test.go:116: Run([catalog show --connection first --connection second --root /var/folders/tk/bmp_tx0976s4rkh1phvrpjlw0000gn/T/TestCatalogConnectionFlagFormsPreserveLegacySemantics3946397549/001 --json]) missing "connection \"second\" not found"; stdout={
              "api_version": "polymetrics.ai/v1",
              "error": {
                "category": "internal",
                "code": "internal_error",
                "message": "connection \"true\" not found"
              },
              "kind": "Error"
            }
             stderr=error: connection "true" not found
    --- FAIL: TestCatalogConnectionFlagFormsPreserveLegacySemantics/unknown_action_flag_tolerated (0.58s)
        catalog_cli_test.go:116: Run([catalog show --connection known --unknown value --root /var/folders/tk/bmp_tx0976s4rkh1phvrpjlw0000gn/T/TestCatalogConnectionFlagFormsPreserveLegacySemantics3946397549/001 --json]) missing "connection \"known\" not found"; stdout={
              "api_version": "polymetrics.ai/v1",
              "error": {
                "category": "internal",
                "code": "internal_error",
                "message": "connection \"true\" not found"
              },
              "kind": "Error"
            }
             stderr=error: connection "true" not found
FAIL
FAIL	polymetrics.ai/internal/cli	3.104s
FAIL
```

## Exact green outputs

```bash
go test ./internal/cli/ -run 'Catalog|CobraRouterShell|Golden' -count=1
```

```text
ok  	polymetrics.ai/internal/cli	16.830s
```

```bash
go test ./internal/cli/ -run Certify -count=1
```

```text
ok  	polymetrics.ai/internal/cli	90.214s
```

```bash
go vet ./...
go build ./cmd/pm
```

```text
# no output; both commands exited 0
```

```bash
go test ./...
```

```text
# pass; full package output emitted in terminal run, with slow packages including:
ok  	polymetrics.ai/internal/cli	162.073s
ok  	polymetrics.ai/internal/connectors/certify	337.523s
```

```bash
make verify
```

```text
# pass; completed gofmt, tidy-check, vet, go test -timeout 20m ./..., go build ./cmd/pm,
# docs validate, local smoke flow, golangci-lint, and connectorgen validate.
# Terminal tail:
0 issues.
go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 547 connector(s) checked, 0 findings
```
