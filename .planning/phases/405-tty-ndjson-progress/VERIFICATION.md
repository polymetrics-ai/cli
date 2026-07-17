# VERIFICATION — Issue 405 TTY gate and NDJSON progress

## Checklist

GSD / process:

- [x] `scripts/gsd doctor`
- [x] `scripts/gsd list`
- [x] `scripts/gsd prompt plan-phase 405 --skip-research`
- [x] `scripts/gsd prompt programming-loop init --phase 405 --dry-run` attempted; adapter gap recorded.
- [x] Manual inline GSD fallback recorded.
- [x] Required skills recorded.
- [x] Phase plan, TDD ledger, verification checklist, run state, summary, prompts created before production edits.

Focused TDD gates:

- [x] `go test ./internal/ui/... -count=1`
- [x] `go test ./internal/cli/... -run 'TestRunWithOptions|TestGlobalUI|TestProgress' -count=1`
- [x] `go test ./internal/cli/... -run TestProgressNDJSON -count=1`
- [x] `go test ./internal/cli/... -run 'Test.*Help|TestGoldenDocsGenerateMatchesTrackedCLIManuals' -count=1`
- [x] `go test ./internal/cli/... -count=1`

CLI help/docs/website parity:

- [x] Runtime help documents `--plain`, `--no-input`, `--progress ndjson`.
- [x] `pm --help`
- [x] `pm help config`
- [x] `pm etl --help`
- [x] `pm flow --help`
- [x] Bare namespace behavior checked: `pm etl`, `pm flow` exit 0 with manuals.
- [x] Invalid actions still return usage errors: `pm --root "$tmpdir" flow bogus` exited 2.
- [x] `docs/cli/**` regenerated and matched `pm docs generate`.
- [x] `website/content/docs/cli-reference.mdx` updated.
- [x] Generated help/manual fixture drift addressed via golden transcript update.
- [x] Completion/discovery metadata: not applicable to this slice beyond root help/manual flags; no completion command changes.

Required local gates:

- [x] `gofmt -w cmd internal`
- [x] `go vet ./...`
- [x] `go test ./...`
- [x] `go build ./cmd/pm`
- [x] `make verify`
- [x] `git diff --check origin/feat/cli-architecture-v2...HEAD`
- [x] `git diff -- go.mod go.sum` reviewed; dependency delta restricted to approved `golang.org/x/term`.

Optional/runtime-backed gates:

- [x] Runtime services not run; no credentialed checks requested.

## Command log

| Command | Result | Notes |
|---|---|---|
| `scripts/gsd doctor` | pass | Adapter health OK. |
| `scripts/gsd list` | pass | 69 commands. |
| `scripts/gsd prompt plan-phase 405 --skip-research` | pass | Prompt generated. |
| `scripts/gsd prompt programming-loop init --phase 405 --dry-run` | fail | `scripts/gsd: unknown GSD command: programming-loop`; manual inline loop fallback. |
| `go test ./internal/ui/... -count=1` | pass | `internal/ui 0.155s`; `internal/ui/styles 0.286s`. |
| `go test ./internal/cli/... -run 'TestRunWithOptions\|TestGlobalUI\|TestProgress' -count=1` | pass | `internal/cli 1.033s`. |
| `go test ./internal/cli/... -run TestProgressNDJSON -count=1` | pass | `internal/cli 1.045s`. |
| `go test ./internal/cli/... -run 'Test.*Help\|TestGoldenDocsGenerateMatchesTrackedCLIManuals' -count=1` | pass | `internal/cli 1.448s`. |
| `go test ./internal/cli/... -count=1` | pass | `internal/cli 169.642s`. |
| `gofmt -w cmd internal && go vet ./... && go test ./... && go build ./cmd/pm && make verify` | pass | `go test ./...` included `internal/cli 173.250s`, `internal/connectors/certify 346.045s`; `make verify` included `internal/cli 174.571s`, `internal/connectors/certify 348.926s`, `smoke ok`, `0 issues`, `connectorgen validate: 547 connector(s) checked, 0 findings`. |
| CLI parity commands | pass | `./pm --help`, `./pm help config`, `./pm etl --help`, `./pm flow --help`, `./pm etl`, `./pm flow`, initialized-root `./pm flow bogus` exit 2, docs/website grep for new flags. |
| `git diff --check origin/feat/cli-architecture-v2...HEAD` | pass | no output. |
| `git diff -- go.mod go.sum` | pass | no output after final gates; dependency delta reviewed against base separately. |
| Remote Website checks (initial PR run) | fail | Generated website data drift: `M lib/docs.generated.ts`. |
| `cd website && pnpm run gen:website-data` | pass | Regenerated `website/lib/docs.generated.ts`; no connector/generated data changes beyond docs data. |
| `cd website && pnpm install --frozen-lockfile` | blocked local | `ERR_PNPM_LOCKFILE_CONFIG_MISMATCH`; CI install succeeded, so local typecheck/unit not rerun here. |

## Gate notes

- Runtime-backed services and credentialed connector checks are not in scope and were not run.
- Parent PR #438 remains draft/human-gated; do not merge to `main`.
- `verificationPassed=true` in `RUN-STATE.json` is valid after full local gates and `make verify` passed.

## Review-fix verification checklist — PR #457 head `3702318efa5514b8fad20c99bba2e3281164bec7`

Cycle status: focused red evidence captured, implementation completed, focused green gates and full local gates passed. PR #457 body updated via GitHub API; branch push remains the only post-verification step.

Focused red/green gates:

- [x] `go test ./internal/ui/... -run 'TestDetectModeUsesADRGate|TestDetectCapabilitiesDegradeForColorAndASCII|TestPaletteANSI16DimUsesBrightBlackSGR' -count=1` — pass, `internal/ui 0.172s`, `internal/ui/styles 0.317s`.
- [x] `go test ./internal/cli/... -run 'TestInvocationEnvCapturesColorControls|TestFlow.*Sanitizes.*Human|TestGlobalUIFlagsDocumentedInHelp|TestProgressNDJSONFailureDocumentsMixedStderr' -count=1` — pass, `internal/cli 6.686s`.
- [x] `go test ./internal/cli/... -count=1` — pass, `internal/cli 169.138s`.

Docs / parity gates:

- [x] Runtime help reworded truthfully for future TTY-gated renderers; no current interactive activation claim while `cmd/pm` uses `cli.Run` plain mode.
- [x] `docs/cli/config.md`, `docs/cli/etl.md`, and `docs/cli/flow.md` regenerated/updated from embedded docs.
- [x] Website ETL docs mention `--progress ndjson`.
- [x] Website architecture flow command/prose mention `--progress ndjson`.
- [x] Website CLI reference documents future TTY-gated wording and mixed stderr diagnostics on failures.
- [x] `website/lib/docs.generated.ts` regenerated via `cd website && pnpm run gen:docs`.
- [x] Grep: `rg -n -- '--progress ndjson|Future TTY renderers|stderr may also include|invalid UI/progress flag' docs/cli website/content/docs website/lib/docs.generated.ts`.
- [x] Unintended generated `docs/connectors/**` drift from docs generation was removed; `git diff --name-only docs/connectors | wc -l` returned `0`.

Full gates after implementation:

- [x] `gofmt -w cmd internal`
- [x] `go vet ./...`
- [x] `go test ./...`
- [x] `go build ./cmd/pm`
- [x] `make verify`
- [x] `git diff --check origin/feat/cli-architecture-v2...HEAD` and `git diff --check`
- [x] PR #457 body update via GitHub API (`gh pr edit` hit a GitHub Projects classic GraphQL deprecation error; REST PATCH succeeded)
- [ ] `git push origin feat/405-tty-ndjson-progress`

Review-fix command log:

| Command | Result | Notes |
|---|---|---|
| `go test ./internal/ui/... -run 'TestDetectModeUsesADRGate\|TestDetectCapabilitiesDegradeForColorAndASCII\|TestPaletteANSI16DimUsesBrightBlackSGR' -count=1` | pass | `internal/ui 0.172s`; `internal/ui/styles 0.317s`. |
| `go test ./internal/cli/... -run 'TestInvocationEnvCapturesColorControls\|TestFlow.*Sanitizes.*Human\|TestGlobalUIFlagsDocumentedInHelp\|TestProgressNDJSONFailureDocumentsMixedStderr' -count=1` | pass | `internal/cli 6.686s`. |
| `cd website && pnpm run gen:docs` | pass | Wrote 11 docs pages to `lib/docs.generated.ts`. |
| `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli -run TestGoldenTranscripts -count=1` | pass | Updated `internal/cli/testdata/golden_transcripts.json`. |
| `go test ./internal/cli/... -count=1` | pass | `internal/cli 169.138s`. |
| `gofmt -w cmd internal && go vet ./... && go test ./... && go build ./cmd/pm && make verify` | pass | `go test ./...`: `internal/cli 170.511s`, `internal/connectors/certify 340.438s`; `make verify`: `internal/cli 171.287s`, `internal/connectors/certify 342.514s`, `smoke ok`, `0 issues`, `connectorgen validate: 547 connector(s) checked, 0 findings`. |
| `git diff --check origin/feat/cli-architecture-v2...HEAD && git diff --check` | pass | no output. |
| `git diff -- go.mod go.sum` | pass | no output. |
| `git diff --name-only docs/connectors \| wc -l` | pass | `0`. |
| `gh pr edit 457 --body-file /tmp/pr-457-body.md` | fail_then_fixed | GraphQL Projects classic deprecation error; REST PATCH fallback used. |
| `gh api --method PATCH repos/polymetrics-ai/cli/pulls/457 --input /tmp/pr-457-body.json --silent` | pass | no output. |

Safety notes:

- [x] No secrets needed.
- [x] No new dependencies added.
- [x] No credentialed connector checks run.
- [x] No reverse ETL execution run outside `make verify` smoke gate using local sample/outbox fixture.
- [x] No parent/shared ledger edits made.

## Review-fix #2 verification checklist — PR #457 head `2195a66659be9d62bf99bfc8e2506e77da81e02f`

Focused red/green gates:

- [x] Red docs validation: `rg -n "CLICOLOR_FORCE" docs/design/tui-ux-design.md` showed stale design-doc claims at lines 63 and 376.
- [x] Red help test after test expectation update: `go test ./internal/cli/... -run 'TestGlobalUIFlagsDocumentedInHelp' -count=1` failed for root/ETL/flow missing `3 validation error` wording.
- [x] Regenerated docs via existing project command: `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir "$tmp_connectors"`.
- [x] Regenerated golden/manual artifacts: `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli -run TestGoldenTranscripts -count=1`.
- [x] Focused gate: `go test ./internal/cli/... -run 'TestGolden|TestGlobalUIFlagsDocumentedInHelp|TestProgressNDJSONFailureDocumentsMixedStderr' -count=1` passed, `internal/cli 6.724s`.

Full gates if Go/help changed:

- [x] `gofmt -w cmd internal`
- [x] `go vet ./...`
- [x] `go test ./...`
- [x] `go build ./cmd/pm`
- [x] `make verify`
- [x] PR #457 body updated with review-fix #2 disposition via GitHub REST API.
- [x] `git push origin feat/405-tty-ndjson-progress`

Review-fix #2 command log:

| Command | Result | Notes |
|---|---|---|
| `go test ./internal/cli/... -run 'TestGlobalUIFlagsDocumentedInHelp' -count=1` | fail | root/ETL/flow help missing `3 validation error`. |
| `tmp_connectors=$(mktemp -d); go run ./cmd/pm docs generate --dir docs/cli --connectors-dir "$tmp_connectors"; rm -rf "$tmp_connectors"` | pass | Regenerated `docs/cli`; connector docs written to temp dir only. |
| `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli -run TestGoldenTranscripts -count=1` | pass | `internal/cli 9.869s`. |
| `go test ./internal/cli/... -run 'TestGolden|TestGlobalUIFlagsDocumentedInHelp|TestProgressNDJSONFailureDocumentsMixedStderr' -count=1` | pass | `internal/cli 6.724s`. |
| `gofmt -w cmd internal && go vet ./... && go test ./... && go build ./cmd/pm && make verify` | pass | `go test ./...`: `internal/cli 170.546s`, `internal/connectors/certify 339.739s`; `make verify`: `internal/cli 171.209s`, `internal/connectors/certify 342.470s`, `smoke ok`, `0 issues`, `connectorgen validate: 547 connector(s) checked, 0 findings`. |
| `gh api --method PATCH repos/polymetrics-ai/cli/pulls/457 --input /tmp/pr-457-body-rf2.json --silent` | pass | PR body updated; no output. |
| `git push origin feat/405-tty-ndjson-progress` | pass | pushed `2195a666..8afb0ea5`; GitHub reported existing default-branch Dependabot vulnerability notice. |

Safety notes:

- [x] No secrets needed.
- [x] No new dependencies planned.
- [x] No credentialed connector checks planned.
- [x] No reverse ETL execution planned outside standard local verification smoke gates.
