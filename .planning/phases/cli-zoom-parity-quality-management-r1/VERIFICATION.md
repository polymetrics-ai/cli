# Verification Checklist — Zoom Quality Management parity, R1

## Lifecycle

- [x] Required skills and connector/CLI/GSD guidance are loaded for the parent run and recorded in `PLAN.md`.
- [x] GSD command provenance is resolved with `scripts/gsd sources`.
- [x] Live artifact URL, retrieval time, HTTP status, byte count, and exact six-operation audit are recorded before RED.
- [x] RED test failure is captured, committed, and pushed before production declarations (`91b7526a5`).
- [x] GREEN implementation is committed and pushed (`b90dcff04`).
- [x] Inline verify-work and code-review evidence are recorded in `REVIEW.md` under the documented manual-GSD fallback.

## Source parity

- [x] All six live Quality Management operations match the derived ledger (delta `0`).
- [x] No `unsafe_or_disallowed` disposition is permitted.
- [x] Five GETs and one POST have exactly one covered declaration each.
- [x] List reads have no source-invented paging/date input flags.
- [x] POST carries every documented typed input and a closed nested request schema.

## Runtime/docs checks

- [x] Focused Zoom, conformance, commandrunner, CLI, vet, and build checks pass.
- [x] Every route is reachable through the built binary; safe reads reach Zoom as provider `401`, and POST remains preview-only.
- [x] `surface-sync --check`, Zoom/full validation, docs validation, and scoped generated-file review pass.

## Green evidence

- `go test -count=1 ./internal/connectors/defs/zoom/...` passed after the six declarations.
- `go test -count=1 -v -run '^TestConformance/zoom$' ./internal/connectors/conformance` passed;
  the full `go test -count=1 -timeout 20m ./internal/connectors/conformance/...` package also
  passed (`18.479s`).
- `go test -count=1 -timeout 20m ./internal/connectors/commandrunner/...` passed (`6.977s`).
- `go test -count=1 -timeout 20m ./internal/cli/...` passed (`560.881s`) after the approved golden
  fixture regeneration. The normal golden-only check also passed (`26.419s`).
- `go vet ./...`, `go build ./cmd/pm`, `make tidy-check`, `make lint`, `make docs-check`, `make
  smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make
  connectorgen-surface-sync`, `make connector-boundary`, and `make release-workflow-check` passed.
- A temporary built `pm` exited successfully for `pm help zoom`, bare `pm zoom`, bare `pm zoom
  quality-management`, and all six route `--help` calls. Five read invocations safely reached Zoom
  as provider `401` with no `unknown command`; the POST route was `--preview --json` only.
- `npm --prefix website run gen:catalog` regenerated the website connector bundle/catalog. A
  structural comparison against `HEAD` confirms `zoom` is the only changed connector; its generated
  capability is now `write=true`, its two actions and 18 commands are present, and the generated
  capability count increases from 236 to 237.
- `git diff --check` passed. Generated golden review found exactly the nine root-help variants;
  generated CLI docs and website catalog review retained only Zoom and catalog deltas.
