# State

Current phase: wave1-http-api-longtail-complete-opencode

Status: completed

Latest verification (2026-06-26):

- `make verify` exit 0; `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, docs validation, and smoke all passed.
- 556 per-system connector dirs registered (556 dirs == 556 registry imports); no quarantine.
- Connector docs generation/validation is manual-only; per-connector `docs/connectors/**/SKILL.md` generation is no longer required for verify.
- No new dependencies added.

Connector count: **556 native Go connectors** (from 1 at session start).
- Templates: github (custom), stripe (declarative-HTTP).
- Parallel batch 1: slack, hubspot(+write), notion, jira, sendgrid, postgres (DB; CDC stubbed).
- Mega-batch: 110 HTTP connectors (airtable, amazon-ads, gitlab, zendesk-*, intercom, klaviyo,
  mailchimp, square, xero, trello … + alpha long tail), all on connsdk, TDD, fixture mode.
- Pending batch repair + OpenCode batch: 50 inherited connector dirs repaired/converged, then 100 more
  API connectors built by 10 parallel OpenCode subagents with red-first package tests.
- Final long-tail completion: remaining 140 HTTP/API source connectors implemented, including the final
  18 test-only packages (`tavus` through `zoho-bigin`) repaired locally with read-only fixture-capable code.

Factory: `cmd/registrygen` derives `registryset/registry_gen.go` from connector dirs (collision-free
parallel authoring); Workflow tool fans out one universal-loop agent per connector.

Known gaps / follow-ups (non-blocking):
- Catalog-enablement divergence: registry 556 live vs catalog enabled=2 — needs a deliberate flip
  pass + conformance-count update.
- Postgres logical-replication CDC needs pglogrepl (dependency human-gate). go1.24 floor accepted.
- No CI yet; local gates: `make verify`, `make verify-duckdb`.
- All current HTTP/API source catalog entries have native connector dirs.

Next: gated DB-CDC/cloud/file batches; catalog-enablement + manifest-streams polish; reverse-ETL writes per-API.
