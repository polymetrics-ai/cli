# Phase 425 TDD Ledger

Issue: #425 — nativize version namespace.
Invocation session: `issue-425-pi-openai-codex-gpt-5.6-sol-high-20260718T095316Z`
Model: `openai-codex/gpt-5.6-sol`; thinking: `high`
Starting HEAD: `479a62f930e7c8a9a51ba0b3deb088bf3aad3ecc`

## GSD and skills

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt plan-phase 425 --skip-research --model=openai-codex/gpt-5.6-sol --thinking=high
scripts/gsd prompt programming-loop init --phase 425 --dry-run --model=openai-codex/gpt-5.6-sol --thinking=high
```

Doctor/list passed (69 commands); plan-phase generated its prompt. Programming-loop returned exact error `scripts/gsd: unknown GSD command: programming-loop` (exit 1), so the manual universal-loop fallback was used without weakening TDD. Final local prompt generation also passed for `verify-work` (7292 bytes) and `code-review` (6158 bytes); no external review was requested.

Loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`.

## Red / green / refactor log

| Step | Kind | Command | Result |
|---:|---|---|---|
| 0 | Planning | Create all six issue-local phase artifacts before production edits | Pass |
| 1 | RED | `go test ./internal/cli/ -run 'Version|CobraRouterShellBuildsFreshHiddenWrapperTree' -count=1` | Failed as expected (`0.612s`) |
| 2 | GREEN | Native leaf registration, hidden positional help alias, legacy wrapper removal, focused tests | Pass (`0.553s`) |
| 3 | Refactor | `go test ./internal/cli/... -run 'Version|CobraRouterShell|Golden' -count=1` | Pass (`7.814s`) |
| 4 | Broader CLI | `go test ./internal/cli/... -count=1` | Pass (`195.315s`) |
| 5 | Full gate | gofmt, vet, full test, build, `make verify` | Pass |
| 6 | Parity/safety | built-binary help/output/error checks, docs/website/golden/dependency/scope diffs | Pass |

## Exact RED

Captured after test-only edits and before any production-code edit:

```text
--- FAIL: TestCobraRouterShellBuildsFreshHiddenWrapperTree (0.00s)
    cobra_router_test.go:55: expectedHidden covers 21 commands, legacy commands plus native commands registers 22
--- FAIL: TestVersionCommandIsNativeCobraLeaf (0.00s)
    cobra_router_test.go:213: version command must use native Cobra flag parsing
FAIL
FAIL\tpolymetrics.ai/internal/cli\t0.612s
FAIL
```

Behavior-focused tests already passed under the legacy wrapper; the RED isolated parser ownership/registration.

## Exact focused GREEN

```text
$ go test ./internal/cli/ -run 'Version|CobraRouterShellBuildsFreshHiddenWrapperTree' -count=1
ok  \tpolymetrics.ai/internal/cli\t0.553s

$ go test ./internal/cli/... -run 'Version|CobraRouterShell|Golden' -count=1
ok  \tpolymetrics.ai/internal/cli\t7.814s

$ go test ./internal/cli/... -count=1
ok  \tpolymetrics.ai/internal/cli\t195.315s
```

Covered native registration, deterministic plain/JSON metadata, flag and positional help, JSON manual, unknown-flag usage mapping, invalid-action usage, no-file completion seam, fresh tree, and unchanged goldens.

## Full GREEN

```bash
gofmt -w cmd internal
go vet ./...
go test -timeout 20m ./...
go build ./cmd/pm
make verify
```

Results: gofmt/vet/build passed without diagnostics; full tests passed (`internal/cli 203.747s`, `internal/connectors/certify 355.702s` among the slow packages). `make verify` passed through fmt, tidy-check, vet, tests, build, docs validation, local temp-dir smoke, lint (`0 issues.`), and `connectorgen validate: 547 connector(s) checked, 0 findings`.

The required `make verify` smoke used only its existing temporary sample fixtures and followed reverse ETL plan → preview → approval → run; no external service or credentialed connector check was used.

## Parity GREEN

Built binary summary:

```text
plain_bytes=35 help_bytes=350 version=Version/dev manuals=CommandManual/version unknown_exit=2 invalid_exit=2
```

`pm help version`, `pm version --help`, `pm version -h`, and `pm version help` were byte-identical. JSON flag/positional help both returned `CommandManual/version`. Unknown flag and invalid action both returned usage exit 2 and no manual. CLI docs temp generation/diff and validation passed; website generator wrote 11 pages with no tracked delta; help, docs, website, goldens, dependencies, and connector defs remained unchanged.
