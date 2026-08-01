# Verification checklist — Lever Hiring connector parity (#3223)

Credential-free only. No live Lever API calls were made.

## Targeted gates

- [x] Official operation inventory script: current Lever documentation rows match `api_surface.json` with zero missing, extra, or duplicate rows.
  - Result: `official 117 local 117 missing 0 extra 0 dupes 0`.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
  - Result: `550 connector(s) checked, 0 findings`.
  - Note: `connectorgen validate` expects the connector definitions root, not an individual connector directory.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/lever-hiring' -count=1`
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`
- [x] `go vet ./internal/connectors/... ./internal/cli/...`
- [x] `go build ./cmd/pm`
- [x] CLI/help docs smoke
  - `go run ./cmd/pm lever-hiring`
  - `go run ./cmd/pm lever-hiring opportunities`
  - `go run ./cmd/pm lever-hiring direct`
  - `go run ./cmd/pm lever-hiring delete-feedback`
  - The current CLI exposes dynamic connector help via bare dynamic namespaces/groups; `pm help lever-hiring` is not a registered static help topic.
- [x] `go run ./cmd/pm docs validate --connectors-dir docs/connectors`
- [x] `make connector-boundary`
  - Result: `outcome=clean`, `findings=0`, `warnings=0`.
- [x] `git diff --check`

## Full local gates

- [x] `gofmt -w cmd internal` / `make fmt`
  - Covered by `make verify`; no Go production files were intentionally edited in this slice.
- [x] `go vet ./...`
  - Covered by `make verify`.
- [x] `go test ./...`
  - Covered by `make verify`.
- [x] `go build ./cmd/pm`
  - Run directly and covered by `make verify`.
- [x] `make verify`
  - Result: passed (`make_verify_exit=0`).

## Safety evidence

- [x] No secrets requested, printed, summarized, stored, or committed.
- [x] No live provider calls or provider writes.
- [x] No new dependencies.
- [x] No shared runtime/engine/CLI Go edits.
- [x] Unimplemented operations remain truthfully blocked/planned or fixture-only/uncertified with official-source/shared-runtime evidence.
- [x] `/Users/karthiksivadas/karthik-agent-workspace/bin/fm-ensure-agents-md.sh .` was run as required; it reported the pre-existing repo state `conflict: both AGENTS.md and CLAUDE.md are real files` and made no tracked changes.

## Final operation totals

- Official rows: 117 total (`GET=56`, `POST=26`, `PUT=14`, `DELETE=11`, `WEBHOOK=10`).
- Implemented dispositions: 25 fixture-backed streams, 23 bounded direct reads, 14 typed reverse-ETL write actions.
- Blocked/planned dispositions: 55 rows.
