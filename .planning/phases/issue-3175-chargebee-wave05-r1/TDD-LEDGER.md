# TDD LEDGER — issue-3175 Chargebee parity wave05 r1

## Red checks planned before production edits

1. `go test ./cmd/connectorgen -run 'TestChargebeeAPISurfaceOperationLedgerMetrics' -count=1`
   - Expected initial state: fail because current Chargebee surface is the old 428-row exclusion ledger, not the 655-row official operation ledger.
2. `go run ./cmd/connectorgen validate internal/connectors/defs/chargebee`
   - Expected initial state: may pass or reveal old-surface limitations; use as focused validation after generated changes.
3. `go test ./internal/connectors/conformance -run 'TestConformance/chargebee' -count=1`
   - Expected initial state: baseline fixture/conformance for old bundle; after expansion it must pass with every executable stream/write fixture.

## Green/refactor evidence to fill

- [x] Red count test captured: `go test ./cmd/connectorgen -run 'TestChargebeeAPISurfaceOperationLedgerMetrics' -count=1` failed with `operation_ledger_version = 0, want 1` on the pre-change Chargebee surface.
- [x] Generated/curated definitions compile and `connectorgen validate` passes: focused temp-root validation returned `1 connector(s) checked, 0 findings`.
- [x] Chargebee conformance passes: `go test ./internal/connectors/conformance -run 'TestConformance/chargebee' -count=1` passed.
- [x] Focused CLI dynamic/golden tests pass: targeted dynamic connector/help/golden subset passed (`TestDynamicConnectorHelpAndBareNamespace`, `TestConnectorInspectJSONIncludesManifest`, `TestRootHelpListsDynamicConnectorCommands`, `TestGoldenTranscripts`). Broad regex `Connector|Dynamic|Golden` was too broad and timed out after 300s due unrelated certification tests, so the focused subset was used.
- [x] Full requested local gates pass or blocker recorded truthfully: focused gates plus vet/build/docs/smoke/lint/connectorgen/connector-boundary/diff-check passed; `make verify` is recorded as blocked by the repository `go test -timeout 20m ./...` timeout in `internal/connectors/certify`.

## Notes

Manual GSD programming-loop fallback is recorded in `PLAN.md` because `scripts/gsd prompt programming-loop ...` is absent from the healthy repo-local adapter registry.
