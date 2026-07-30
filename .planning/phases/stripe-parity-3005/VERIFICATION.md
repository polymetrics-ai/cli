# Verification checklist — Stripe connector parity (#3005)

Run credential-free only. Do not call live Stripe APIs.

## Targeted gates

- [ ] Official operation inventory script: 589 official operations and zero missing ledger rows.
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs/stripe`
- [ ] `go test ./internal/connectors/conformance -run 'TestConformance/stripe' -count=1`
- [ ] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` (if connector command/help metadata changed)
- [ ] `go vet ./internal/connectors/... ./internal/cli/...`
- [ ] `go build ./cmd/pm`
- [ ] `make connector-boundary`
- [ ] `git diff --check`

## Full local gates (time permitting before handoff)

- [ ] `gofmt -w cmd internal` (only if Go files changed; expected not applicable)
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`

## Safety evidence

- [ ] No live credentials requested, printed, summarized, or stored.
- [ ] No provider writes or live certification run.
- [ ] No shared runtime files edited.
- [ ] Any unimplemented operation remains truthfully blocked/planned or fixture-only/uncertified.

