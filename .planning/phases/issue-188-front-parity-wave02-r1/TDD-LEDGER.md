# TDD Ledger — issue #188 Front parity

## Red / baseline evidence

Baseline before production bundle edits:

```bash
go run ./cmd/connectorgen validate internal/connectors/defs/front
```

Result: failed because this tool treats the provided directory as a connector root and therefore scanned `fixtures/` and `schemas/` as sibling connectors (`missing_file` for `metadata.json`). This records that the issue checklist's connector-dir form is not a valid invocation for this repo-local validator. Targeted validation was performed by copying `front/` under a temporary defs root; full-root validation was also run after implementation.

```bash
go test ./internal/connectors/conformance -run 'TestConformance/front' -count=1
```

Result: passed for the old six-stream read-only bundle, proving the red gap is the documented parity contract (no full ledger/writes/direct/binary metadata), not a failing legacy fixture.

## Green evidence

| Cycle | Command | Result | Notes |
| --- | --- | --- | --- |
| plan | `scripts/gsd doctor` | pass | Adapter healthy before planning. |
| plan | `scripts/gsd prompt programming-loop init --phase issue-188-front-parity --dry-run` | unavailable | Adapter lacks `programming-loop`; manual Pi GSD loop fallback recorded. |
| plan | `scripts/gsd prompt quick --full ...` | pass | Prompt trace recorded in `.planning/traces/issue-188-front-parity-wave02-r1-gsd-quick.prompt.md`. |
| red | `go run ./cmd/connectorgen validate internal/connectors/defs/front` | expected fail | Validator directory semantics; see note above. |
| red | `go test ./internal/connectors/conformance -run 'TestConformance/front' -count=1` | pass | Old scope fixture passed. |
| green | targeted temp-root `go run ./cmd/connectorgen validate <tmp-with-front> --json` | pass | 1 connector checked, 0 findings, 0 warnings. |
| green | `go run ./cmd/connectorgen validate` | pass | 549 connectors checked, 0 findings. |
| green | `go test ./internal/connectors/conformance -run 'TestConformance/front' -count=1` | pass | Front conformance passed. |
| green | `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1 -timeout=8m` | pass | Exact issue checklist command passed with extended Go timeout after updating generated golden transcripts. |
| green | `go vet ./internal/connectors/... ./internal/cli/...` | pass | No output. |
| green | `go build ./cmd/pm` | pass | No output. |
| green | `make connector-boundary` | pass | Outcome clean, no findings/warnings. |
| green | `git diff --check` | pass | No whitespace errors. |
