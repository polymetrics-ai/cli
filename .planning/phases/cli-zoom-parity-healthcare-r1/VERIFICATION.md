# Verification Checklist — Zoom Healthcare parity, R1

## Lifecycle

- [x] Required skills and connector/CLI/GSD guidance loaded.
- [x] `scripts/gsd doctor`, command provenance, all five generated prompts, and `agentcontractgen check` completed.
- [x] Official GSD phase lookup recorded `phase_found: false`; inline/manual fallback recorded in `PLAN.md`.
- [x] Live Healthcare artifact re-fetched before test or production work; URL, retrieval time, HTTP status, and byte count recorded.
- [x] RED state captured before production JSON/doc edits: `go test -count=1 ./internal/connectors/defs/zoom/...` failed with the expected 9→12, 6→8, 0→1, and unknown-command assertions (verbatim output in `TDD-LEDGER.md`).
- [ ] RED state committed and pushed before production JSON/doc edits.
- [ ] GREEN implementation committed and pushed.
- [ ] Inline verify-work evidence recorded.
- [ ] Inline code-review findings/dispositions recorded.

## Source parity

- [x] Live document has exactly GET/GET/PATCH Healthcare operations; derived ledger delta is zero.
- [x] `from`, `to`, `page_size`, and `next_page_token` are response fields, not surfaced request flags.
- [ ] `api_surface.json` has exactly 12 covered operations and no removed/non-Zoom row.
- [ ] Zero `unsafe_or_disallowed` dispositions.

## Targeted checks

- [ ] `go test -count=1 ./internal/connectors/defs/zoom/...`
- [ ] `go run ./cmd/connectorgen surface-sync --check`
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs/zoom`
- [ ] `go run ./cmd/connectorgen validate`
- [ ] `go test -timeout 20m ./internal/connectors/conformance/...`
- [ ] `go test -timeout 20m ./internal/connectors/commandrunner/...`
- [ ] `go test -timeout 20m ./internal/cli/...`
- [ ] `go vet ./...`
- [ ] `go build ./cmd/pm`

## Binary and docs parity

- [ ] Built binary: `pm help zoom`, `pm zoom`, `pm zoom healthcare`.
- [ ] Built binary: direct-read command `--help` for list/get and reverse-ETL command `--help` for update.
- [ ] Synthetic-token live reachability: both reads return a Zoom `401`, not `unknown command`.
- [ ] Update route enters plan/preview semantics only; no live PATCH is issued.
- [ ] Zoom docs/manual/catalog/golden delta is scoped and `pm docs validate --connectors-dir docs/connectors` passes.
