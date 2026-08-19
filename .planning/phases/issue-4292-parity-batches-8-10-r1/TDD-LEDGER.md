# Issue #4292 — TDD ledger

## Red

- Batch 8: the pre-generation file-presence assertion failed for all thirty
  required artifacts (three source artifacts for each of brex, zoho-books,
  testrail, amplitude, posthog, metabase, dbt, looker, mode, and dremio).
- Batch 8: the first integrity run failed on duplicated Zoho Books source-lock
  operation IDs. The existing inventory has three repeated method/path pairs,
  so a method/path-only synthetic identity would silently collapse documented
  entries. The generator now makes the authoritative `api_surface.json`
  endpoint index part of every source ID.
- Remaining batches: the same assertion must fail if a source ID is omitted,
  duplicated, placed in more than one primary class, if a typed write action is
  classified as `reverse_etl`, or if a direct-write row omits the locked
  `generic-typed-destination-executor` reverse-ETL eligibility gap.

## Green

- Batch 8: `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/generate-parity-maps.mjs 8`
  generated all ten artifact triplets. TestRail is an allowed source skip:
  Chrome reached the provider's Cloudflare verification page rather than the
  published reference, so its lock records `no-public-api-description` and
  `retrieval_method: browser`, with no invented pin or operation rows.
- Batch 8: `node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/verify-parity-maps.mjs 8`
  passed for all ten connectors after the source-identity correction.
- Batch 8: all ten `connectorgen validate ... --json` commands passed with
  zero findings/warnings; `surface-sync internal/connectors/defs --check`
  passed with 552 scanned and zero corrections.
- Batch 9: the pre-generation file-presence assertion failed for all thirty
  required artifacts, then generation and
  `verify-parity-maps.mjs 9` passed for all ten connectors.
- Batch 9: all ten `connectorgen validate ... --json` commands passed with
  zero findings/warnings; `surface-sync internal/connectors/defs --check`
  again passed with 552 scanned and zero corrections.
- Pending: repeat for batch 10.

## Refactor / review

- Batch 8: integrity assertions reviewed the source lock, crosswalk, and
  ledger together; all direct writes retain only the reverse-ETL eligibility
  foundation gap and no row is primarily classified `reverse_etl`.
- Batch 9: the same assertions passed, including DELETE coverage and the
  direct-write/reverse-ETL separation for all ten connector inventories.
- Pending: inspect all final JSON ledgers for false engine-gap claims,
  invented contracts, source location drift, delete coverage, and transport
  correction compliance.

## Fixed constraints

- Reverse ETL is an eligibility attribute, not an endpoint parity class. It is
  a foundation gap in this task, not declaration-pending. Evidence:
  `internal/app/issue_label_warehouse_transport.go:85-95` at `acb85dc03`.
  No `transport_binding` action may be created; existing typed write actions
  remain enabled `direct_write` operations.
- ETL is only declarable where the definition-owned source contract is actually
  satisfied.
