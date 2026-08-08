# Summary — Zoom Customer Managed Keys Hybrid parity, R1

## Outcome

Zoom's provider-owned **Customer Managed Keys Hybrid** category is complete under
[#3950](https://github.com/polymetrics-ai/cli/issues/3950). The category's one documented
operation is executable through the bundled CLI surface:

- `customer-managed-keys-hybrid archival-key decrypt`
- `POST /api/v2/kms/cse/archival/datakey/decrypt`

The live source artifact was fetched before the RED checkpoint from
`https://developers.zoom.us/docs/api/customer-managed-keys-hybrid.md` at
`2026-08-08T10:54:00Z` (HTTP `200`, `1,890` bytes). Its customer-hosted auth reference was
also audited at `2026-08-08T10:55:19Z` (HTTP `200`, `24,583` bytes). No secret material was
recorded.

Zoom moves from `22` to `23` executable operations and from `1,820` to `1,819` locally blocked
operations. Its executable mix is `17` direct reads, `1` direct write, and `2` reverse-ETL writes.
The Zoom surface contains `0` `unsafe_or_disallowed` rows.

## Delivered work

- Added a closed-schema, approval-gated, typed-confirmation direct write with exactly the two
  documented JSON body inputs; it has no invented paging flags.
- Added redacted output behavior for successful and error responses containing key material.
- Added a customer-hosted operation transport that selects its distinct base URL and bearer
  credential, excludes ordinary Zoom OAuth, and clears inherited secret-bearing headers.
- Added reusable direct-write endpoint-ledger runtime coverage. It unblocks safely declared
  mutation operations in other connector bundles without reclassifying any unrelated connector.
- Regenerated the Zoom command/manual/website projections and retained only Zoom-scoped generated
  output after restoring unrelated aggregate-index drift.

## Evidence and handoff

- RED commits: `5a0172053`, `4927731d9`, `016f7e558`, and `5c9518918`.
- GREEN/foundation commits: `0987a58bc`, `833a2d9d4`, `410eb1bb7`, and `dfa221bcd`.
- Connector/docs commits: `cec675503` and `f86e6a480`.
- A freshly built `pm` binary passed Zoom/base/group/command help and the complete
  plan → preview → approval → confirmation → execute flow against an isolated loopback fixture.
  Assertions confirmed the declared request wiring and redacted every synthetic sensitive value.
- `surface-sync --check`, full connector validation, scoped surface reconciliation, endpoint-ledger
  diff scope, `go vet`, lint, scoped Go tests, docs validation, website typecheck, and CLI golden
  transcript checks are green. The base-to-head global endpoint-ledger object delta resolves only
  to `zoom`.

The next worker should select the next unclaimed Zoom provider-owned category from parent
[#3915](https://github.com/polymetrics-ai/cli/issues/3915), begin with a fresh live-artifact audit,
and create a separate RED checkpoint before authoring its bundle declarations.
