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

- [ ] `go test ./internal/ui/... -count=1`
- [ ] `go test ./internal/cli/... -run 'TestRunWithOptions|TestGlobalUI|TestProgress' -count=1`
- [ ] `go test ./internal/cli/... -run TestProgressNDJSON -count=1`
- [ ] `go test ./internal/cli/... -run 'Test.*Help|TestGoldenDocsGenerateMatchesTrackedCLIManuals' -count=1`

CLI help/docs/website parity:

- [ ] Runtime help documents `--plain`, `--no-input`, `--progress ndjson`.
- [ ] `pm --help`
- [ ] `pm help config`
- [ ] `pm etl --help`
- [ ] `pm flow --help`
- [ ] Bare namespace behavior checked: `pm etl`, `pm flow` exit 0 with manuals.
- [ ] Invalid actions still return usage errors.
- [ ] `docs/cli/**` regenerated or manually matched to `pm docs generate`.
- [ ] `website/content/docs/cli-reference.mdx` updated.
- [ ] Generated help/manual fixture drift addressed.
- [ ] Completion/discovery metadata: not applicable to this slice beyond root help/manual flags; no completion command changes.

Required local gates:

- [ ] `gofmt -w cmd internal`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`
- [ ] `git diff --check origin/feat/cli-architecture-v2...HEAD`
- [ ] `git diff -- go.mod go.sum` reviewed; dependency delta restricted to approved `golang.org/x/term`.

Optional/runtime-backed gates:

- [ ] Runtime services not run; no credentialed checks requested.

## Command log

| Command | Result | Notes |
|---|---|---|
| `scripts/gsd doctor` | pass | Adapter health OK. |
| `scripts/gsd list` | pass | 69 commands. |
| `scripts/gsd prompt plan-phase 405 --skip-research` | pass | Prompt generated. |
| `scripts/gsd prompt programming-loop init --phase 405 --dry-run` | fail | `scripts/gsd: unknown GSD command: programming-loop`; manual inline loop fallback. |

## Gate notes

- Runtime-backed services and credentialed connector checks are not in scope.
- Parent PR #438 remains draft/human-gated; do not merge to `main`.
- `verificationPassed` in `RUN-STATE.json` remains false until full local gates pass, especially `make verify`.
