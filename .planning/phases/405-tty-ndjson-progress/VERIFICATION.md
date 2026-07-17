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

## Gate notes

- Runtime-backed services and credentialed connector checks are not in scope and were not run.
- Parent PR #438 remains draft/human-gated; do not merge to `main`.
- `verificationPassed=true` in `RUN-STATE.json` is valid after full local gates and `make verify` passed.
