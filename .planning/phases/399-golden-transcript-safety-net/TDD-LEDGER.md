# TDD Ledger — Issue 399 Golden Transcript Safety Net

## Classification

Test-harness and docs-parity safety-net work for existing CLI behavior. No production dispatcher changes are allowed.

## Loaded skills

- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-documentation`
- `golang-security`
- `gsd-core`

Skill routing source: `.agents/agentic-delivery/references/required-skills-routing.md` sections **Always-on Go skill routing**, **CLI and command behavior**, and **Documentation for Go behavior**.

## Initial red / absent evidence to capture before harness edits

| Evidence | Command / Source | Result |
|---|---|---|
| Golden transcript suite absent | `go test ./internal/cli/ -run Golden -count=1` | `ok  	polymetrics.ai/internal/cli	0.525s [no tests to run]` (`go_test_status=0`) |
| No existing Golden test symbol | `rg -n "Golden|golden transcript|docs generation diff|docs-generate-diff" internal/cli` | no matches (`rg_status=1`) |
| GSD programming-loop shell command missing | `scripts/gsd prompt programming-loop init --phase 399 --dry-run` | `scripts/gsd: unknown GSD command: programming-loop` |

## Planned red/green/refactor evidence

| Slice | Red evidence | Green evidence | Refactor evidence |
|---|---|---|---|
| Golden transcripts | Absent `Golden` suite before edit; first run generated fixture with update mode | `go test ./internal/cli/ -run Golden -count=1` passed with 89 cases pinning exit code/stdout/stderr | `gofmt -w cmd internal`; no production dispatcher code changed |
| Docs generate diff | New docs-diff test initially failed because tracked `docs/cli/connectors.md` had a stale `GITHUB CERTIFICATION` block absent from `pm docs generate` output | Removed the stale generated-doc drift block; `go test ./internal/cli/ -run Golden -count=1` passed | Full internal/cli focused test remains green |
| Parity spot checks | N/A: this issue does not introduce a new CLI surface; docs-diff test exposed stale generated markdown drift | `pm help docs`, bare `pm connectors`, `pm docs --help`, and docs/website grep passed | Full gates pass |

## Green evidence targets

| Target | Verification |
|---|---|
| Targeted golden suite | `go test ./internal/cli/ -run Golden` exits 0 |
| Formatting | `gofmt -w cmd internal` produces no unintended diff |
| Static analysis | `go vet ./...` exits 0 |
| Full tests | `go test ./...` exits 0 |
| CLI binary builds | `go build ./cmd/pm` exits 0 |
| Full repo verify | `make verify` exits 0 |
| No dependency drift | `git diff -- go.mod` has no output |
| CLI parity | Runtime help/docs spot checks recorded in `VERIFICATION.md`; website docs marked not applicable if unchanged |

## Evidence log

- 2026-07-16: GSD adapter healthy via `scripts/gsd doctor`.
- 2026-07-16: `scripts/gsd prompt plan-phase 399 --skip-research` generated `/tmp/gsd-plan-phase-399.prompt`.
- 2026-07-16: `scripts/gsd prompt programming-loop init --phase 399 --dry-run` failed with `scripts/gsd: unknown GSD command: programming-loop`; manual fallback is `.pi/prompts/pm-gsd-loop.md`.
- 2026-07-16 red/absent evidence before harness edits:
  ```text
  ok  	polymetrics.ai/internal/cli	0.525s [no tests to run]
  go_test_status=0
  rg_status=1
  ```
- 2026-07-16 docs-diff red after adding the test: `TestGoldenDocsGenerateMatchesTrackedCLIManuals` failed on `docs/cli/connectors.md`; generated docs lacked the stale `GITHUB CERTIFICATION` block present in the tracked file.
- 2026-07-16 green targeted evidence after fixture/test/docs-drift fix:
  ```text
  gofmt -w cmd internal && go test ./internal/cli/ -run Golden -count=1
  ok  	polymetrics.ai/internal/cli	6.316s
  ```
- 2026-07-16 full green evidence:
  - `go vet ./...`: pass.
  - `go test ./...`: pass.
  - `go build ./cmd/pm`: pass.
  - `make verify`: pass; completed docs-check, smoke, lint, and `connectorgen validate: 547 connector(s) checked, 0 findings`.
  - `git diff -- go.mod`: empty.
  - CLI parity spot checks: `pm help docs`, bare `pm connectors`, `pm docs --help`, docs/website grep all pass.

## Notes

- Do not store secrets or credential values in transcripts.
- Use temp directories for docs generation and project roots.
- Prefer deterministic commands that avoid wall-clock, network, credential, or local-machine-specific output.
- If a transcript is intentionally broad, document why it is representative rather than exhaustive.
