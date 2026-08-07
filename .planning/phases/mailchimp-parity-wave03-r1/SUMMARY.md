# Mailchimp parity wave03-r1 summary

Implemented fixture-backed Mailchimp Marketing API parity for issues #3078-#3085 on branch `fm/cli-mailchimp-parity-wave03-r1`.

## Delivered

- Replaced the legacy 9-row Mailchimp surface with a complete 298-operation ledger generated from the official Mailchimp Marketing Swagger root and all provider-owned path refs.
- Added 79 ETL streams with JSON schemas and sanitized stream fixtures.
- Added 68 typed direct-read/search operations.
- Added 148 named reverse-ETL actions with closed record schemas, request-shape fixtures, risk text, redaction hints, idempotent DELETE metadata where applicable, and destructive confirmation on risky lifecycle/delete/send actions.
- Kept 3 operations blocked/local-workflow with exact policy evidence: `GET /`, `GET /ping`, and generic `POST /batches`.
- Updated Mailchimp docs, generated connector catalog rows, website connector data/catalog surfaces, and CLI golden transcripts.
- Synced `internal/connectors/defs/operation_endpoint_ledger.json` via `connectorgen surface-sync`, which
  is what makes the 68 direct reads pass the #3890 runtime preflight guard. Single hunk, mailchimp key only.

This branch carries **no shared runtime changes**. Two earlier shared edits were deliberately dropped:
single-bundle `connectorgen validate` support and `sync.Once` bundle caching in `bundleregistry.New`.
Main has since grown its own equivalent of the former; the latter was a pure performance optimization
that does not belong in a connector-scoped change and is not present here. Repeated bundle parsing is
therefore uncached, which is why the `internal/cli` gate takes ~200s rather than ~20s.

## Final counts

Measured from the committed bundle after the rebase, not copied from the audit baseline.

- Official operations represented: 298.
- Executable rows: 295, each reachable as its own `pm mailchimp <command>`.
  - Streams: 79 (79/79 fixture-covered).
  - Direct reads: 68 (covered by validate + runtime preflight; the conformance harness has no
    operation fixture slot, and no connector in the repo ships one).
  - Reverse-ETL write actions: 148 (148/148 fixture-covered).
- Blocked rows: 3, each with a reason and official `source_url`.
- Excluded/N/A rows: 0. Undispositioned rows: 0.
- Live certified rows: 0 (no credentialed/live provider calls were made).

The baseline was 4 implemented / 291 blocked-or-planned / 3 excluded. The 3 formerly counted as
excluded are now carried as blocked with citations, so excluded is 0 rather than 3.

## Safety

- No secrets requested, stored, summarized, or printed.
- No live Mailchimp provider calls.
- No push, PR, merge, `/no-mistakes`, VPS, Thaalam, or shared daemon lifecycle actions.
- No generic raw HTTP/batch write command; `POST /batches` remains blocked by policy.
- Reverse ETL remains plan -> preview -> explicit approval -> execute, with destructive confirmation where declared.

## Verification

See `VERIFICATION.md` for the post-rebase gate results. `traces/` holds the original pre-rebase run and
the reusable audit evidence (`mailchimp-official-audit.json`, `mailchimp-official-audit.md`); its gate
logs describe the older base and were superseded by the results recorded in `VERIFICATION.md`.

Passing gates: connectorgen validate, Mailchimp conformance, the CLI connector/dynamic/golden gate, the
full `./internal/connectors/...` suite, the `TestEveryImplementedCommandPassesRuntimePreflight` guard,
`go build ./cmd/pm`, `go vet ./...`, `make connector-boundary`, `make verify`, and `git diff --check`.
