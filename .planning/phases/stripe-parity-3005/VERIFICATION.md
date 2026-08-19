# Verification checklist — Stripe connector parity (#3005)

Run credential-free only. Do not call live Stripe APIs.

## Targeted gates

- [x] Official operation inventory script: 589 official operations and zero missing ledger rows.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` (548 connectors, 0 findings).
- [x] Stripe-only temp defs-root validation (`connectorgen validate` expects a parent defs root): 1 connector, 0 findings.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/stripe' -count=1`
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` (after golden update; passed in 121.770s).
- [x] `go vet ./internal/connectors/... ./internal/cli/...`
- [x] `go build ./cmd/pm`
- [x] `make connector-boundary`
- [x] `git diff --check`

## Full local gates (time permitting before handoff)

- [ ] `gofmt -w cmd internal` (Go files changed for shared write-default materialization; formatting owned by the implementation slice)
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`

## Safety evidence

- [x] No live credentials requested, printed, summarized, or stored.
- [x] No provider writes or live certification run.
- [x] Shared engine write-default materialization was edited to preserve defaulted `base_url` behavior; no CLI Go files edited.
- [x] Any unimplemented operation remains truthfully blocked/planned or fixture-only/uncertified.

