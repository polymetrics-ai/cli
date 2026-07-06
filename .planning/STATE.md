# State

Current milestone: connector-architecture-v2

Current phase: wave2-fanout-http-sm

Status: pending (BLOCKED on S3 engine mini-wave: incremental-lower-bound query vars + ResolveCheck ==/in when-grammar + carried minors)

Completed: wave1-pilot (2026-07-02) — 10/10 pilots migrated at parity, 2 Fable review rounds,
gap-loop cycle 1 (engine mini-wave + 10-pilot repair), re-review GO completed_with_warnings.
See .planning/phases/wave1-pilot/SUMMARY.md + docs/migration/pilot-costs.json.

Completed: wave0-engine-harness (2026-07-02) — declarative engine (85.7% cov), 3 goldens with
parity (stripe/searxng/postgres), connectorgen, conformance v2, certify source stages, lint gates,
migration recipe + schemas, inventory (557 connectors S137/M388/L31/XL1 → ~77 Pass A bundle
agents). Reviewer GO after 1 gap-loop cycle. See .planning/phases/wave0-engine-harness/SUMMARY.md.

Model policy (user directive): all GSD loop roles use Codex `gpt-5.5` with `xhigh` reasoning
effort for this GitHub CLI parity implementation track; Go work applies cc-skills-golang skills.

Previous phase: wave1-http-api-longtail-complete-opencode (completed)

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
