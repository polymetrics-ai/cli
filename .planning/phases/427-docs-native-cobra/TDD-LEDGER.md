# Phase 427 TDD Ledger

Issue: #427 — nativize docs namespace.
Invocation session: `issue-427-pi-openai-codex-gpt-5.6-sol-high-20260718T112639Z`
Model: `openai-codex/gpt-5.6-sol`; thinking: `high`
Starting HEAD: `ab847da28ebf78e5732ac1edcde8e39f92dc2656`

## GSD and skills

Doctor/list passed; `scripts/gsd prompt plan-phase 427 --skip-research` generated the official prompt and it was executed inline. `scripts/gsd prompt programming-loop init --phase 427 --dry-run` failed with `scripts/gsd: unknown GSD command: programming-loop`, so the recorded manual universal-runtime-loop fallback is active.

Loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`.

## Planned red / green / refactor log

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| 0 | Planning | Create all six phase artifacts before production edits | Complete |
| 1 | RED | `go test ./internal/cli/ -run 'Docs|CobraRouterShellBuildsFreshHiddenWrapperTree' -count=1` | Failed as expected (`11.332s`) |
| 2 | GREEN | Native namespace/actions/typed flags and docs-only legacy parser removal | Pass (`11.462s`) |
| 3 | Refactor | Focused docs/router/golden + full CLI tests | Pass (`18.453s`; golden `5.470s`; full CLI `227.224s`) |
| 4 | Full gate | gofmt, vet, full tests, build, `make verify` | Pass |
| 5 | Parity/safety | built binary, docs/website/generated/golden/dependency/scope checks | Pass |

## Planned RED coverage

- `docs` is native and absent from legacy wrappers; `generate` and `validate` are native actions with declared `stringArray` output-directory flags, `NoOptDefVal=true`, unknown-flag whitelists, and no-file completion fallback.
- `generate` preserves `--dir` and `--connectors-dir` spaced/assigned/repeated-last-wins/bare/comma forms, exact output text, exact CLI-manual bytes, and expected connector artifacts.
- `validate` preserves `--connectors-dir`, fallback `--dir`, default semantics where safely applicable, unknown flags, extra positionals, and exact output text.
- Bare/text/JSON/flag/short/positional help all resolve to the canonical docs manual.
- Invalid action remains usage exit 2; missing/empty generate dir and malformed assigned globals retain existing categories; errors are not masked by help.
- Global/config forms cover `--root value`, `--root=value`, late placement, assigned `--json=true|false`, `--plain=false`, and `--no-input=true`.
- Filesystem tests write only under safe temporary roots, assert all generated relative paths remain local, compare CLI docs byte-for-byte to the canonical map/checked-in tree, and validate connector output without secrets.

## Exact RED

Captured after test-only edits and before any production-code edit:

```text
--- FAIL: TestCobraRouterShellBuildsFreshHiddenWrapperTree (0.00s)
    cobra_router_test.go:55: expectedHidden covers 21 commands, legacy commands plus native commands registers 22
--- FAIL: TestDocsCommandIsNativeCobraSubtree (0.00s)
    cobra_router_test.go:213: docs command must use native Cobra flag parsing
FAIL
FAIL\tpolymetrics.ai/internal/cli\t11.332s
FAIL
```

All observable docs behavior, byte-parity, action/flag, error, global/config, and safe temp-root filesystem tests passed through the legacy wrapper. RED isolates the required parser ownership and registration change. Production files remained untouched at this checkpoint.

## Focused GREEN

```text
$ go test ./internal/cli/ -run 'Docs|CobraRouterShellBuildsFreshHiddenWrapperTree' -count=1
ok  \tpolymetrics.ai/internal/cli\t11.462s

$ go test ./internal/cli/... -run 'Docs|CobraRouterShell|Golden' -count=1
ok  \tpolymetrics.ai/internal/cli\t18.453s

$ go test ./internal/cli/ -run '^TestGoldenTranscripts$' -count=1
ok  \tpolymetrics.ai/internal/cli\t5.470s

$ go test ./internal/cli/... -count=1
ok  \tpolymetrics.ai/internal/cli\t227.224s
```

Implementation: `docs`, `docs generate`, and `docs validate` are native Cobra nodes; `--dir` and `--connectors-dir` are `StringArray` flags with legacy bare-flag behavior and spaced-value normalization; unknown action flags remain whitelisted; positional help is a hidden compatibility node. `runDocs` accepts typed action flags directly, so the docs-only `parseFlags` call and legacy wrapper are removed. Focused byte/output/help/global-config/filesystem tests and unchanged goldens pass.

## Full GREEN and parity

| Command / gate | Result |
|---|---|
| Final focused docs/router tests after default-path hardening | Pass (`13.710s`) |
| Final docs/router/golden subset | Pass (`20.158s`) |
| `go test ./internal/cli/... -count=1` | Pass (`227.224s`) |
| `go test ./internal/cli/ -run '^TestGoldenTranscripts$' -count=1` | Pass (`5.470s`) |
| `gofmt -w cmd internal`; `go vet ./...`; `go build -o /tmp/pm-427 ./cmd/pm` | Pass, no diagnostics |
| `go test -timeout 20m ./...` | Pass (`real 347.167s`; CLI `229.851s`, certify `342.890s`) |
| final `make verify` | Pass (`real 249.846s`; CLI `229.679s`, smoke OK, lint `0 issues`, 547 connectors / 0 findings) |
| built binary help/generation/validation/error/config/global matrix | Pass: `help_bytes=818`, exits invalid=`2`, missing=`1`, plain override=`2` |
| temp docs generation + `diff -ru docs/cli`; temp and tracked docs validation | Pass, no diff |
| `npm --prefix website run gen:docs`; website diff | Pass, 11 pages, no diff |
| start-HEAD dependency/docs/website/golden/connector-def and `git diff --check` guards | Pass |

`make verify` used only its existing local temporary sample smoke. Its reverse ETL step followed plan → preview → approval → run; no external service or credentialed connector check was used. Focused generation wrote only beneath temporary roots, compared every generated CLI manual byte-for-byte, asserted generated relative paths stayed local, and validated connector catalog/manual/icon artifacts without reading credentials.

## Bounded review correction ledger

Correction session: `issue-427-review-correction-pi-openai-codex-gpt-5.6-sol-high-20260718T121208Z`; model `openai-codex/gpt-5.6-sol`; thinking `high`; exact start `ea93b4bb7a7eb09236ad829d5ad6055b0c00c30d`.

GSD evidence: `scripts/gsd doctor` and `scripts/gsd list` passed. `scripts/gsd prompt programming-loop "Issue #427 bounded correction: preserve legacy docs trailing help and literal separator parsing"` failed because the healthy adapter still has no `programming-loop` command, so the previously documented manual universal-runtime-loop fallback remains active. Required skills reloaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-documentation`, and `golang-spf13-cobra`.

| Step | Kind | Planned command / evidence | Status |
|---:|---|---|---|
| C0 | Planning | Update correction plan, ledger, checklist, and run state before production edits | Complete |
| C1 | RED | `go test ./internal/cli/ -run 'TestDocsActions(PreserveLegacyTrailingHelpBehavior|ContinueParsingAfterLiteralSeparator)$' -count=1` | Failed as expected (`0.570s`) |
| C2 | GREEN | Docs-only pre-Cobra normalization seam; focused correction test `7.487s`, post-comment rerun `7.596s` | Pass |
| C3 | Refactor/parity | Docs/router/golden `26.541s`; 12-case legacy-base differential matrix | Pass, 0 differences |
| C4 | Gates | docs generation/validation/website parity, gofmt, vet, build, full CLI/repository, `make verify` | Pass |

No production file was edited before C1.

Exact correction RED: all 10 generate/validate/bogus `--help`/`-h` subtests failed because Cobra returned the docs `CommandManual` with exit 0 instead of running the legacy action/error path. Literal-separator generation failed with exit 1 and `error: missing --dir`, proving pflag stopped before the supplied typed flags. These failures match both accepted medium findings and occurred after test-only edits.

Correction GREEN: `normalizeNativeStringArrayArgs` now invokes a docs-only action-tail seam. It first preserves typed spaced-value normalization for generate/validate, then removes only literal `--`, action-tail `--help`/assigned help, and `-h` tokens that legacy `parseFlags` recorded but docs handlers ignored. Namespace `docs --help`/`-h` and positional `docs help` bypass the seam; no other namespace is touched. Native command nodes and typed pflags remain intact.

## Correction verification evidence

| Gate | Result |
|---|---|
| Focused correction tests | Pass (`7.487s`; post-comment rerun `7.596s`) |
| Focused all docs/router | Pass (`20.032s`) |
| Focused docs/router/golden | Pass (`26.541s`) |
| Full `internal/cli/...` | Pass (`238.822s`) |
| Legacy differential | 12/12 exact exit/stdout/stderr matches against `ab847da28ebf78e5732ac1edcde8e39f92dc2656`; generate/validate/bogus, long/short trailing help, missing/supplied flags, literal separator |
| Temp docs parity | Generated CLI tree byte-diff clean; generated and tracked connector validation pass; namespace help routes byte-identical (`818` bytes) |
| Website parity | `npm --prefix website run gen:docs` wrote 11 pages; tracked diff clean |
| Formatting/static/build | `gofmt -w cmd internal`, `go vet ./...`, `go build ./cmd/pm`, `git diff --check` pass |
| Full repository tests | Pass; CLI `238.885s`, certify `340.747s` |
| `make verify` | Pass; docs validation, local smoke, lint `0 issues`, 547 connectors / 0 findings |
| Scope guards | No `go.mod`/`go.sum`, connector-def, checked-in docs, website, or golden delta |

The existing `make verify` local sample smoke followed reverse ETL plan → preview → approval → run. No service, credentialed connector check, secret, or external write was used.
