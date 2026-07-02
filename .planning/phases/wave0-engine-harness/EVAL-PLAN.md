# EVAL-PLAN — wave0-engine-harness (quantitative exit metrics)

All metrics are hard gates for phase completion (V-21 in PLAN.md). No metric may be relaxed
without a human gate (quality-gate reduction).

## 1. Engine coverage

- **Metric**: statement coverage of `internal/connectors/engine` ≥ **85%**.
- **Measure**: `go test -cover ./internal/connectors/engine` (record exact % in VERIFICATION.md).
- Rationale: the engine executes 550+ future connectors; per design §F.5, engine unit tests own
  all generic behavior.

## 2. Golden parity

- **Metric**: **3/3** goldens pass their full parity suites (stripe, searxng, postgres — TEST-PLAN
  §5): identical records for identical fixture input, identical write request shapes (stripe),
  manifest-surface equality, config-validation parity (postgres).
- **Measure**: `go test ./internal/connectors/engine -run 'TestParity' -v` and
  `go test ./internal/connectors/native/postgres -v` — 0 failures, 0 skips.
- Documented deviations allowed ONLY if listed in `docs/migration/conventions.md` deviation ledger
  (wave0 budget: ≤2 entries; currently 1 planned — stripe create_customer `minProperties`).

## 3. connectorgen validate defect detection

- **Metric**: validate catches ≥ **8 distinct seeded defect classes** (≥10 seeded bundles), with
  findings naming file + rule; and **accepts all 3 goldens** with zero findings.
- **Seeded classes** (corpus in `cmd/connectorgen/testdata/invalid/`):
  1. missing required bundle file (metadata.json)
  2. spec.json fails meta-schema / draft-07 subset compile
  3. unresolvable interpolation key (`{{ config.nope }}`)
  4. `schema` ref to missing file
  5. `x-primary-key` field absent from stream schema
  6. `incremental.cursor_field` absent from stream schema
  7. write `path_fields ⊄ record_schema` properties
  8. api_surface violation (endpoint with both/neither covered_by+excluded, or declared stream
     missing from surface)
  9. connector name regex violation
  10. secret-shaped literal in fixtures / missing docs.md heading
- **Measure**: `go test ./cmd/connectorgen -v` (one subtest per class) +
  `go run ./cmd/connectorgen validate internal/connectors/defs` exit 0.

## 4. Conformance v2

- **Metric**: `TestConformance` green for 3/3 goldens, including dynamic replay checks
  (`pagination_terminates`, `records_match_schema`, `cursor_advances`, `write_request_shape`,
  `delete_semantics` where applicable); conformance self-test corpus: every static check has a
  failing negative case (10/10 checks exercised).
- **Measure**: `go test ./internal/connectors/conformance -v`.

## 5. Certify core

- **Metric**: source-stage run against `sample` = `passed: true` with all applicable stages 0–11
  green, 5/5 sync-mode matrix rows reported with correct `data_source` (2 live-local, 3 capture),
  resume stage green, 0 secrets found by the redaction scan, ephemeral root cleaned.
- **Measure**: `go test ./internal/connectors/certify -run TestSourceStages -v`.

## 6. Whole-tree gates

- `go build ./... && go vet ./... && go test ./...` — green.
- `golangci-lint run` — 0 issues with committed `.golangci.yml`.
- `make verify` (extended) — green; `go run ./cmd/registrygen` produces zero diff
  (coexistence invariant).
- `docs/migration/inventory.json` exists with > 500 connector entries and validates as JSON.
- TDD ledger: every B-task in PLAN.md has RED evidence recorded before its implementation commit
  (TDD-GATE check).

## 7. Prompt/agent eval notes (for the coordinator)

- Record per-task executor token/cost in `.planning/phases/wave0-engine-harness/agents/*` notes —
  wave0 numbers seed the wave1 `pilot-costs.json` methodology.
- Repair-rate signal: >1 repair retry on any single engine task, or the same `ENGINE_GAP` filed
  twice, triggers a planner review of SPEC §1.1 before Wave F dispatch.
