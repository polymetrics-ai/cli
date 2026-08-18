# Verification — PostgreSQL CDC Restart Recovery

## Acceptance checklist

- [x] Current failure reproduced live before production changes, with the mechanism hypothesis confirmed or refuted.
- [x] Focused restart regression observed red before green.
- [x] Accepted checkpoint resumes at the exact durable logical-replication position.
- [x] Source/system/timeline/slot/publication/relation/schema drift still fails closed with typed rebootstrap outcomes.
- [x] Receipt, checkpoint, and acknowledgement ordering remains durable under injected failure.
- [x] Independent target query records exact row counts before interruption, at interruption, and after restart.
- [x] The post-interruption key appears exactly once; no row is lost or duplicated.
- [x] PostgreSQL CDC capability evidence matches the observed truth.
- [x] Focused tests, package tests, `internal/cli` where touched, vet, build, generated/connector/doc gates, and diff hygiene pass.
- [x] CLI help/manual/website parity is confirmed not applicable, or completed if the user-facing contract changes.
- [ ] PR targets `integration/4015-mvp-flat-r1`, carries `Refs #4015`, and identifies release 0.2.1 in its title.

## Executed checks

1. Focused red/green and package results are recorded in `traces/focused-red.txt` and `traces/focused-green.txt`.
2. Live failure and repaired process-restart runs are recorded in `traces/live-red.txt` and `traces/live-green.txt`.
3. Exact CDC target counts: before interruption `1`, at interruption `1`, after restart `2`; resumed key multiplicity `1`. The independent control target stayed at `1,001` rows.
4. The immutable capability record remains `passed`: its explicit bounded-stage/receipt/readback/acknowledgement facts passed again with the fixed binary. It does not claim generic exactly-once delivery; PostgreSQL CDC remains declared at-least-once.
5. `go test -timeout 20m ./internal/connectors/native/postgres/...` — pass in 1.678s.
6. `go test -timeout 20m ./internal/cli` — pass in 558.084s.
7. `go vet ./...` — pass.
8. `go build ./cmd/pm` — pass.
9. `scripts/verify-gsd-workflow` — pass.
10. `make tidy-check lint docs-check smoke-no-build agent-contract-check connectorgen-validate connectorgen-surface-sync connector-runtime-preflight connector-canon-check connector-boundary release-workflow-check` — pass. Notable results: zero lint issues; 552 connector bundles validated with zero findings; zero surface drift; runtime preflight, connector canon, boundary, docs, smoke, agent-contract, tidy, and release-target parity gates passed.
11. `go run ./cmd/connectorgen certification-matrix --check` — pass; certification shards current.
12. `git diff --check origin/integration/4015-mvp-flat-r1...HEAD` — pass.
13. `/Users/karthiksivadas/karthik-agent-workspace/bin/fm-ensure-agents-md.sh .` — pass; `AGENTS.md` already satisfied the project-memory contract and remained unchanged.

The repository's aggregate `go test -timeout 20m ./...` and `make verify` are delegated to CI as required by `AGENTS.md`; the changed packages, `internal/cli`, and every named non-full-suite verification gate were run locally.

CLI help/manual/website parity is not applicable: the change adds no command, flag, output contract, connector surface, or help topic. The binary's existing typed error disappears only for a valid internal restart checkpoint; invalid checkpoint families still fail closed.
