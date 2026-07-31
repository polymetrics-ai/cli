# TDD Ledger: Twenty CRM connector parity wave02 r1

## Strategy

This slice restores a definition bundle and connector-local docs/tests. The red gate is validation against the absent connector; the green gates are connector validation, conformance, docs validation, help/inspect smoke, generated catalog parity, and broader Go verification without live provider calls.

## Red / validation-before-production-edit evidence

- [x] Isolation confirmed before branch creation: `pwd -P` and `git rev-parse --show-toplevel` both returned `/Users/karthiksivadas/.treehouse/cli-83d592/25/cli`.
- [x] Baseline state: `internal/connectors/defs/twenty` absent on this `main`-based branch.
- [x] Run targeted missing-connector validation before restoring bundle:
  - `go run ./cmd/connectorgen validate internal/connectors/defs/twenty --json` exited 1 with missing-dir/read-root error before bundle creation.
  - `go test ./internal/connectors/conformance -run 'TestConformance/twenty' -count=1` exited 0 because no matching Twenty subtest existed before bundle creation; this records absence rather than green coverage.

## Green implementation evidence

- [x] `internal/connectors/defs/twenty/**` restored/created with 168 API-surface rows, 28 streams, 112 write actions, 28 schemas, 29 stream fixtures, 112 write fixtures, CLI metadata, and `docs.md`.
- [x] DELETE/destructive actions have `kind: "delete"`, `confirm: "destructive"`, bounded `record_schema`, id-path templating, idempotent 404 semantics in fixtures, and docs.
- [x] Reverse-ETL create/update/batch actions have typed record schemas and fixtures; batch actions use bounded array fixtures compatible with the current shared engine.
- [x] Direct-read get-by-id commands are represented as planned/partial in CLI metadata; no unsupported generic direct-read execution is claimed.
- [x] `docs/connectors/twenty/MANUAL.md` and `docs/connectors/twenty/SKILL.md` exist and validate.
- [x] `internal/connectors/conformance/twenty_test.go` runs targeted conformance for Twenty and exact body/no-body write fixture checks.
- [x] Generated connector catalogs, website connector data, CLI golden transcripts, and connector-count tests were updated for the new Twenty catalog entry.

## Verification evidence

| Gate | Result | Notes |
| --- | --- | --- |
| `go run ./cmd/connectorgen validate internal/connectors/defs --json` | pass | `connectors_checked=550`, 0 findings, 0 warnings. Current validator expects the defs root; direct `internal/connectors/defs/twenty` targeting treats nested fixture/schema dirs as connector dirs. |
| `go test ./internal/connectors/conformance -run 'TestTwenty\|TestConformance/twenty' -count=1` | pass | Fixture-backed, credential-free. |
| `go run ./cmd/pm docs validate --dir docs/cli --connectors-dir docs/connectors --website-dir website/content/docs` | pass | Connector docs validated. |
| `go run ./cmd/pm help docs` | pass | Help read before/around docs generation. |
| `go run ./cmd/pm help twenty` | pass | Renders Twenty CRM manual. |
| `go run ./cmd/pm connectors` | pass | Bare namespace renders contextual help. |
| `go run ./cmd/pm connectors inspect twenty --json` | pass | No credentials read; manifest reports 28 streams and 112 write actions. |
| `cd website && pnpm run gen:website-data` | pass | Wrote 550 website connectors and included slug `twenty`. |
| `gofmt -w cmd internal` | pass | Run directly and again inside `make verify`. |
| `go vet ./...` | pass | No output. |
| `go test ./...` | attempted | Default package timeout hit slow `internal/cli` after 10m; recorded as broad-suite timeout. |
| `go test -timeout 20m ./...` | pass | Covered by `make verify`'s `test` target. |
| `go build ./cmd/pm` | pass | No output. |
| `git diff --check` | pass | No whitespace findings. |
| `make verify` | pass | Full local gate passed, including fmt, tidy-check, vet, `go test -timeout 20m ./...`, build, docs-check, smoke, lint, connectorgen validate, connector boundary, and release workflow check. |
| `/Users/karthiksivadas/karthik-agent-workspace/bin/fm-ensure-agents-md.sh .` | attempted | Exited 1 because both `AGENTS.md` and `CLAUDE.md` are real files; recorded as existing repo instruction-file conflict outside this connector slice. |

## Safety assertions

- No real Twenty credentials, API keys, bearer tokens, account IDs, emails, or provider records are stored in fixtures or docs.
- No live provider requests are part of the test plan or verification run.
- No destructive/admin external action is executed.
- Shared runtime/foundation production files and other connector bundles were not edited; shared test/golden/catalog count artifacts were updated only for the new Twenty connector entry.
