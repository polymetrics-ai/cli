# Verification Checklist — Zoom Healthcare parity, R1

## Lifecycle

- [x] Required skills and connector/CLI/GSD guidance loaded.
- [x] `scripts/gsd doctor`, command provenance, all five generated prompts, and `agentcontractgen check` completed.
- [x] Official GSD phase lookup recorded `phase_found: false`; inline/manual fallback recorded in `PLAN.md`.
- [x] Live Healthcare artifact re-fetched before test or production work; URL, retrieval time, HTTP status, and byte count recorded.
- [x] RED state captured before production JSON/doc edits: `go test -count=1 ./internal/connectors/defs/zoom/...` failed with the expected 9→12, 6→8, 0→1, and unknown-command assertions (verbatim output in `TDD-LEDGER.md`).
- [x] RED state committed and pushed before production JSON/doc edits (`5c68143e4`).
- [x] GREEN implementation committed and pushed (`1d260747c`).
- [x] Inline verify-work evidence recorded under the documented manual-GSD fallback.
- [x] Inline code-review findings/dispositions recorded in `REVIEW.md`.

## Source parity

- [x] Live document has exactly GET/GET/PATCH Healthcare operations; derived ledger delta is zero.
- [x] `from`, `to`, `page_size`, and `next_page_token` are response fields, not surfaced request flags.
- [x] `api_surface.json` has exactly 12 covered operations and no removed/non-Zoom row.
- [x] Zero `unsafe_or_disallowed` dispositions (`rg` returned zero Zoom rows).

## Targeted checks

- [x] `go test -count=1 ./internal/connectors/defs/zoom/...`
- [x] `go run ./cmd/connectorgen surface-sync --check`
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/zoom`
- [x] `go run ./cmd/connectorgen validate`
- [x] `go test -timeout 20m ./internal/connectors/conformance/...`
- [x] `go test -timeout 20m ./internal/connectors/commandrunner/...`
- [x] `go test -count=1 -timeout 20m ./internal/cli/...`
- [x] `go vet ./...`
- [x] `go build -o <temporary>/pm ./cmd/pm`
- [x] `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, and `make release-workflow-check`

## Binary and docs parity

- [x] Built binary: `pm help zoom`, `pm zoom`, `pm zoom healthcare`.
- [x] Built binary: direct-read command `--help` for list/get and reverse-ETL command `--help` for update.
- [x] Synthetic-token live reachability: both reads return a Zoom `401`, not `unknown command`.
- [x] Update route enters plan/preview semantics only; no live PATCH is issued.
- [x] Zoom docs/manual/catalog/golden delta is scoped and `pm docs validate --connectors-dir docs/connectors` passes.
