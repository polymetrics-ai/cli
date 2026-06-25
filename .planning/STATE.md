# State

Current phase: wave1-mega-batch-110

Status: completed

Latest verification (2026-06-26):

- `make verify` exit 0 (independently re-verified); `go test ./...` → 127 packages ok.
- 118 per-system connectors registered (118 dirs == 118 registry imports); no quarantine.
- No new dependencies (go-duckdb + pgx only); 110-connector run added zero deps.

Connector count: **118 native Go connectors** (from 1 at session start).
- Templates: github (custom), stripe (declarative-HTTP).
- Parallel batch 1: slack, hubspot(+write), notion, jira, sendgrid, postgres (DB; CDC stubbed).
- Mega-batch: 110 HTTP connectors (airtable, amazon-ads, gitlab, zendesk-*, intercom, klaviyo,
  mailchimp, square, xero, trello … + alpha long tail), all on connsdk, TDD, fixture mode.

Factory: `cmd/registrygen` derives `registryset/registry_gen.go` from connector dirs (collision-free
parallel authoring); Workflow tool fans out one universal-loop agent per connector.

Known gaps / follow-ups (non-blocking):
- Manifest streams=0 for the 110 (implement Catalog not Manifest) — fix manifest builder to fall back
  to Catalog streams. Cosmetic (docs only).
- Catalog-enablement divergence: registry 118 live vs catalog enabled=2 — needs a deliberate flip
  pass + conformance-count update.
- Postgres logical-replication CDC needs pglogrepl (dependency human-gate). go1.24 floor accepted.
- No CI yet; local gates: `make verify`, `make verify-duckdb`.

Next: remaining ~390 HTTP long-tail via more Workflow runs; gated DB-CDC/cloud/file batches;
catalog-enablement + manifest-streams polish; reverse-ETL writes per-API.
