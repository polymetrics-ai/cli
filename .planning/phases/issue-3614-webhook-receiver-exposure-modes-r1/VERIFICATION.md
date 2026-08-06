# VERIFICATION — issue #3614 webhook receiver exposure modes

Status: planned; no production implementation has begun.

## Required evidence

- Direct unit coverage for R1–R7 in `TDD-LEDGER.md`, including signature,
  receipt-before-ack, duplicate/out-of-order, oversize, backpressure, endpoint
  generation, and heartbeat-loss paths.
- Targeted package tests, race test for the receiver package, `go vet`,
  `go build ./cmd/pm`, `gofmt`, and `git diff --check`.
- CLI parity: `pm help <receiver-topic>`, bare namespace, command `--help`,
  `docs/cli/**`, website docs, and generated references where applicable.
- Individual gates: `tidy-check`, `lint`, `docs-check`, `smoke-no-build` only
  when it does not use credentials or a reverse write, `agent-contract-check`,
  `connectorgen-validate`, `connectorgen-surface-sync`, `connector-boundary`,
  and `release-workflow-check`.
- Tailscale evidence: installed version, observed stable node DNS, zero `pm`
  dependency delta, entitlement ports, unmodified existing mapping evidence,
  and—only if safe—loopback receiver/Funnel success on unused port 10000.
