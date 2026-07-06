# Roadmap

## Milestone: connector-architecture-v2 (current)

Rewrite the connector architecture to be JSON-schema-declarative, unified-named (clean break from
source-/destination- slugs), with full capability on first run (every GET → ETL stream; every
POST/PUT/PATCH/DELETE → reverse-ETL write action) plus a credential-driven certification harness.
Plan of record: `~/.claude/plans/please-check-all-the-serialized-storm.md` (approved 2026-07-02);
program PRD: `docs/plans/universal-programming-loop-prd.md`.

Execution: GSD Universal Programming Loop per phase; the GitHub CLI parity implementation track uses
Codex `gpt-5.5` with `xhigh` reasoning effort for all GSD roles; role prompts apply
cc-skills-golang.

### Phase: wave0-engine-harness

Declarative engine (`internal/connectors/engine/`), bundle JSON Schemas + `cmd/connectorgen`
(validate|gen|new), conformance v2 (httptest fixture replay), `pm connectors certify` harness core,
`docs/migration/conventions.md`, `.golangci.yml`, `docs/migration/inventory.json`.

Acceptance:
- Engine unit tests green (interpolation, auth selection, pagination matrix, read/write paths, error mapping).
- 3 goldens migrated with engine-vs-legacy parity tests passing: stripe, searxng, postgres.
- `connectorgen validate` rejects seeded-invalid bundles; accepts the goldens.
- Conformance v2 passes for the 3 goldens (static + httptest fixture replay).
- Certify source stages pass against the `sample` connector end-to-end.
- `go build ./... && go test ./... && golangci-lint run` green.

### Phase: wave1-pilot

Migrate 10 pilot connectors (xkcd, vitally, bitly, calendly, sentry, chargebee, zendesk-support,
monday, github, gmail), one Sonnet backend-agent each; Fable line-by-line review of every diff;
conventions + executor prompt template patched with learnings; per-connector cost data recorded.

Acceptance: all 10 pass agent self-check + wave gate + review; conventions.md updated; cost report
written to `docs/migration/pilot-costs.json`; Pass B budget decision made with user.

### Phase: wave2-fanout-http-sm

Fan-out migration of declarative-HTTP S (<300 loc, 12/agent) and M (300–699, 7/agent) connectors
(~49 bundle agents). Wave gate: path guard, registrygen, full build/test/conformance/lint; 20%
adversarial review; 1 repair retry then quarantine.

### Phase: wave3-fanout-http-lxl

Fan-out migration of declarative-HTTP L (700–899, 4/agent) and XL (≥900, 1/agent) connectors
(~56 bundle agents). 100% review of XL output.

### Phase: wave4-fanout-nonhttp

Migration of database_go / file_go / destination_go / native_go kinds (~15 agents) to Tier-3
native layout (component split) + defs bundles.

### Phase: wave5-capability-expansion

Pass B (roster per pilot decision): per-connector `api_surface.json` from documentation_url →
implement missing streams + write actions → completeness critic ≥95% coverage or documented
exclusions → `docs/migration/coverage-report.json`.

### Phase: wave6-convergence

Registry flip to bundles; legacy deletion (slug.go, catalog_data.json, registryset, native_port.go,
native_conformance.go, manifest.go structs, alias/live-registry machinery); naming clean-break
sweep; catalog generated from manifests; docs regen; full `certify --replay` + credential-gated
live certification. HUMAN GATE before deletion.

## Completed milestone: go-cli-mvp

## Phase: go-cli-mvp

Build a working Go CLI vertical slice for local ETL and reverse ETL using the architecture in `POLYMETRICS_GO_CLI_MONOLITH_PRD_ARCHITECTURE.md`.

Acceptance:

- `poly init` creates a usable project directory.
- `poly help` and `poly man` expose detailed docs.
- Credentials can be added from environment values and stored encrypted.
- A connection can sync sample data into a local JSONL warehouse.
- A reverse ETL plan can preview warehouse data and write approved mapped records to an outbox.
- Commands support JSON output for agent callers.
- `go test ./...` and `go build ./cmd/poly` pass.
