# Mailchimp parity wave03-r1 summary

Implemented fixture-backed Mailchimp Marketing API parity for issues #3078-#3085 on branch `fm/cli-mailchimp-parity-wave03-r1`.

## Delivered

- Replaced the legacy 9-row Mailchimp surface with a complete 298-operation ledger generated from the official Mailchimp Marketing Swagger root and all provider-owned path refs.
- Added 79 ETL streams with JSON schemas and sanitized stream fixtures.
- Added 68 typed direct-read/search operations.
- Added 148 named reverse-ETL actions with closed record schemas, request-shape fixtures, risk text, redaction hints, idempotent DELETE metadata where applicable, and destructive confirmation on risky lifecycle/delete/send actions.
- Kept 3 operations blocked/local-workflow with exact policy evidence: `GET /`, `GET /ping`, and generic `POST /batches`.
- Updated Mailchimp docs, generated connector catalog rows, website connector data/catalog surfaces, and CLI golden transcripts.
- Added generic validation/runtime scalability fixes needed by the expanded bundle: single-bundle `connectorgen validate` support and cached embedded bundle parsing in `bundleregistry.New`.

## Final counts

- Official operations represented: 298.
- Implemented fixture-backed rows: 295.
  - Streams: 79.
  - Direct reads: 68.
  - Reverse-ETL write actions: 148.
- Blocked/local-workflow rows: 3.
- Excluded/N/A rows: 0.
- Live certified rows: 0 (no credentialed/live provider calls were made).

## Safety

- No secrets requested, stored, summarized, or printed.
- No live Mailchimp provider calls.
- No push, PR, merge, `/no-mistakes`, VPS, Thaalam, or shared daemon lifecycle actions.
- No generic raw HTTP/batch write command; `POST /batches` remains blocked by policy.
- Reverse ETL remains plan -> preview -> explicit approval -> execute, with destructive confirmation where declared.

## Verification

See `VERIFICATION.md` and `traces/` for gate logs. Final passing gates include connectorgen validate, Mailchimp conformance, CLI/golden regex gate, `go build ./cmd/pm`, `make connector-boundary`, and `make verify`.
