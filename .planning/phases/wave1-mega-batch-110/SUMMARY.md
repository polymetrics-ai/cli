# SUMMARY — Wave 1 mega-batch: 110 HTTP connectors (Workflow factory)

Status: **completed** (GO). 110/110 green, 0 quarantined, `make verify` green (independently verified).

## What ran
Workflow `wf_03ad10f6-628`: **110 universal-loop agents in parallel** (~7 concurrency waves), one per
connector, each research-driven (read its catalog entry + WebFetch API docs) on connsdk with strict
TDD + mandatory fixture mode, writing ONLY its own `internal/connectors/<name>/` dir. Then a
convergence agent ran registrygen + docs-gen + `make verify`, quarantining failures.
**111 agents, ~9.48M subagent tokens, ~50 min wall-clock.**

## Result
- **110 connectors built green, 0 quarantined.** Total per-system connectors: **118** (8 prior + 110).
- Spans GA → beta → alpha: airtable, amazon-ads, amazon-seller-partner, amplitude, bing-ads, gitlab,
  google-analytics, intercom, klaviyo, mailchimp, monday, sentry, twilio, zendesk-*, square, xero,
  trello … through the alpha long tail (100ms, 7shifts, activecampaign, …).
- Real connectors: each defines 3–50 streams with field mappers, auth (Bearer / API-key / Basic /
  custom OAuth2 refresh-token e.g. amazon-ads), pagination, and fixture mode. Verified airtable
  Catalog → [bases, tables, records]; gitlab/zendesk/amazon-ads/100ms 30–50 stream entries.

## Convergence self-correction (notable)
The agent found a real bug in `cmd/registrygen`: it required a dir's package clause to equal the dir
name, so 33 connectors with hyphenated/digit-leading dirs (amazon-ads→pkg `amazonads`, 100ms→`onehms`,
zendesk-support, google-search-console, …) had sanitized package identifiers and were silently
dropped from the registry. It fixed the GENERATOR (key the blank import off the import PATH, not the
package identifier) — recovering all 33. (cmd/registrygen/main.go updated accordingly.)

## Independent verification (orchestrator)
- 118 connector dirs == 118 registry blank imports; no `_quarantine/`.
- `go test ./...` → 127 packages ok. `make verify` exit 0.
- **No new dependencies** — go.mod still only go-duckdb + pgx (the 110 agents respected the no-deps rule).
- `pm connectors list` → 124 registered (118 per-system + built-ins + 2 aliases); diverse inspects resolve.

## Known gaps / follow-ups (non-blocking)
- **Manifest streams=0**: the 110 implement `Catalog` (functional) but not optional `Manifest()` — so
  generated docs don't list streams. Fix once centrally: have the manifest builder fall back to
  `Catalog()` streams when `Manifest()` is absent. (stripe/github implement Manifest.)
- **Catalog enablement divergence**: registry has 118 live connectors but catalog_data.json still
  marks them planned_native_port (enabled=2). A deliberate catalog-enablement pass should flip them
  and update the conformance-count assertions.
- Most are read-only; reverse-ETL writes to be added per-API where mutations are sensible.

## Significance
The parallel factory scales: 110 connectors in one run, green, zero shared-file collisions, no deps.
The same Workflow handles the remaining ~390 HTTP long-tail in further runs; DB-CDC/cloud/file
families run as gated batches.
