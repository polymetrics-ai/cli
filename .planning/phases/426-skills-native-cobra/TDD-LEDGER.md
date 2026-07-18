# Phase 426 TDD Ledger

Issue: #426 — nativize skills namespace.
Invocation session: `issue-426-pi-openai-codex-gpt-5.6-sol-high-20260718T104457Z`
Model: `openai-codex/gpt-5.6-sol`; thinking: `high`
Starting HEAD: `54bfcbab5a997c81676b286fe78de00a109f5fba`

## GSD and skills

Doctor/list passed; `scripts/gsd prompt plan-phase 426 --skip-research` generated the official prompt. `scripts/gsd prompt programming-loop init --phase 426 --dry-run` failed with `scripts/gsd: unknown GSD command: programming-loop`, so the recorded manual universal-runtime-loop fallback is active.

Loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-documentation`, `golang-spf13-cobra`, `golang-security`.

## Planned red / green / refactor log

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| 0 | Planning | Create all six phase artifacts before production edits | Complete |
| 1 | RED | `go test ./internal/cli/ -run 'Skills|CobraRouterShellBuildsFreshHiddenWrapperTree' -count=1` | Failed as expected (`29.549s`) |
| 2 | GREEN | Native namespace/action/typed flags and legacy parser removal | Pass (`29.454s`) |
| 3 | Refactor | Focused skills/router/golden + full CLI tests | Pass (`37.019s`; full CLI `223.229s`) |
| 4 | Full gate | gofmt, vet, full tests, build, `make verify` | Pass |
| 5 | Parity/safety | built binary, docs/website/generated/golden/dependency/scope checks | Pass |

## Planned RED coverage

- `skills` is native and absent from legacy wrappers; `generate` is a native action with a native `stringArray` `--dir`, `NoOptDefVal=true`, unknown-flag whitelist, and no-file completion fallback.
- `generate` preserves `--dir value`, `--dir=value`, repeated-last-wins, bare `--dir=true`, comma-preserving values, unknown flags, ignored extra positionals, plain output, JSON `SkillGeneration`, and missing/empty-dir validation.
- Bare/text/JSON/flag/short/positional help all resolve to the canonical skills manual.
- Invalid action remains usage exit 2 and does not render help; malformed global boolean remains validation exit 3.
- Global/config forms cover `--root value`, `--root=value`, late placement, `--json`, `--json=true`, and `--json=false` overriding config/env JSON.
- Existing generation contract writes only expected metadata skill files under test temp directories and contains no secret values.

## Exact RED

Captured after test-only edits and before any production-code edit:

```text
--- FAIL: TestCobraRouterShellBuildsFreshHiddenWrapperTree (0.00s)
    cobra_router_test.go:55: expectedHidden covers 21 commands, legacy commands plus native commands registers 22
--- FAIL: TestSkillsCommandIsNativeCobraSubtree (0.00s)
    cobra_router_test.go:213: skills command must use native Cobra flag parsing
FAIL
FAIL\tpolymetrics.ai/internal/cli\t29.549s
FAIL
```

All observable skills behavior tests passed through the legacy wrapper. RED isolates the required parser ownership and registration change. Production files remained untouched at this checkpoint.

## Focused GREEN

```text
$ go test ./internal/cli/ -run 'Skills|CobraRouterShellBuildsFreshHiddenWrapperTree' -count=1
ok  \tpolymetrics.ai/internal/cli\t29.454s

$ go test ./internal/cli/... -run 'Skills|CobraRouterShell|Golden' -count=1
ok  \tpolymetrics.ai/internal/cli\t37.019s
```

Implementation: `skills` and `skills generate` are native Cobra nodes; `--dir` is a `StringArray` with legacy bare-flag behavior and spaced-value normalization; unknown action flags remain whitelisted; positional help is a hidden compatibility node. `runSkills` accepts the typed last directory directly, so the skills-only `parseFlags` call and legacy wrapper are removed. All focused output/help/global-config/filesystem tests and unchanged goldens pass.

## Full GREEN and parity

| Command / gate | Result |
|---|---|
| `go test ./internal/cli/... -count=1` | Pass (`223.229s`) |
| `go test ./internal/cli/ -run '^TestGoldenTranscripts$' -count=1` | Pass (`5.902s`) |
| `gofmt -w cmd internal`; `go vet ./...`; `go build -o /tmp/pm-426 ./cmd/pm` | Pass, no diagnostics |
| `go test -timeout 20m ./...` | Pass (`real 351.94s`; CLI `227.351s`, certify `347.262s`) |
| `make verify` | Pass (`real 27.70s`; lint `0 issues`; 547 connectors, 0 findings) |
| built binary help/generation/error/config/global matrix | Pass: `help_bytes=716`, 12 generated skills, exits invalid=2/missing=3/unknown=2 |
| temp docs generation + `diff -ru docs/cli`; docs validation | Pass, no diff |
| `npm --prefix website run gen:docs`; website diff | Pass, 11 pages, no diff |
| start-HEAD dependency/docs/website/golden/connector-def and `git diff --check` guards | Pass |

`make verify` used only its existing local temporary sample smoke. Its reverse ETL step followed plan → preview → approval → run; no external service or credentialed connector check was used. The generated skills checks wrote only beneath temporary directories and confirmed canonical metadata files remained local and contained no token-shaped values.
