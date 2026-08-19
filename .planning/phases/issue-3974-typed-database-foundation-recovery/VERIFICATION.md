# Verification checklist — Issue #3974 typed database foundation recovery

## Inline verify-work result

The manual inline `verify-work` fallback is green after one bounded correction round. The
only failed gate was the source-derived certification matrix check. Its attribution, #4026
correction-child creation, RED output, regeneration, and GREEN recheck are retained in
`traces/capability-matrix-red.txt`. It updated only generated source-location metadata and
did not change connector capabilities or source behavior.

## Required focused checks

- [x] `go test -timeout 20m -count=1 ./internal/connectors/database`
- [x] `go test -timeout 20m -count=1 ./internal/warehouse`
- [x] `go test -timeout 20m -count=1 ./internal/synccontract`
- [x] `go test -timeout 20m -count=1 ./internal/connectors/engine -run '^TestBundleLoadPostgresDatabaseDefinitionWithoutCapabilityPromotion$'`
- [x] `go test -timeout 20m -count=1 ./internal/connectors/native/postgres -run '^TestPostgresDatabaseDriverReferenceSeam$'`
- [x] `./internal/connectors/bundleregistry`, `./internal/connectors/native/nativeset`,
  `./internal/connectors/engine`, `./internal/connectors/native/postgres`, and `./internal/app`
- [x] `go test -timeout 20m -count=1 ./internal/cli`

All focused suites passed. The application suite was not repeated after its initial green run;
the CLI suite passed in 197.366s.

## Required static/repository checks

- [x] `gofmt -d` changed Go paths and `git diff --check` (no output)
- [x] `go vet ./...`
- [x] `go build ./cmd/pm`
- [x] `make tidy-check`, `make lint`, `make docs-check-no-build`, `make smoke-no-build`
- [x] `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`
- [x] `make github-parity-artifacts-check`, `make connectorgen-certification-matrix`,
  `make connector-runtime-preflight`, `make connector-canon-check`
- [x] `make connector-boundary`, `make release-workflow-check`

`connectorgen-certification-matrix` was initially RED because #3974 shifted generated
`bundle.go` source locations. It is GREEN after #4026's documented regeneration:
`connectors=556`, `capability_complete=0`, `certified=0`.

## Truth boundaries

- [x] `database.json` admits no modes (`admitted_modes=[]`).
- [x] PostgreSQL metadata has `write:false`, `query:false`, and `cdc:false`; binary inspection
  reports `write=false`, `query=false`, `changefeed_status=planned`, and its CDC catalog has
  `count=0` with no PostgreSQL entry.
- [x] Diff/source review finds no generic SQL/query/write surface, target executor, canonical mode
  execution, or CDC v2 introduction.
- [x] #3995 status is recorded as a bounded `RETRY`-compatible verdict in
  `SHEPHERD-COMPATIBILITY.json`; #3995 is not in child/parent ancestry and no automatic approval is claimed.

## Delivery checks

- [x] Inline `verify-work` result recorded; #4026 is the sole gap-repair child and correction
  round 1/5.
- [x] Deep inline code review is dispositioned in `REVIEW.md` (PASS; no actionable source finding).
- [x] no-mistakes run `01KZPXF5VNSD89NNBWJT74BD00` completed without `--yes`:
  intent, rebase, review, test, document, and lint passed with zero findings. The canonical child
  remote stages (`push`, `pr`, `ci`) were skipped because they cannot set the required non-default
  parent base; those stages are owned manually below.
- [ ] PR targets `feat/3972-postgres-parity`, uses both Refs lines, and preserves/links #4014 audit history.
