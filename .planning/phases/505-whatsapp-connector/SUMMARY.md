# Phase 505 Summary

Status: implementation complete on `fm/cli-whatsapp-connector-r1`; full local verification passed;
ready for the no-mistakes PR against `main` (human-gated merge).

## Delivered

- New connectors-as-data bundle `internal/connectors/defs/whatsapp/`: `metadata.json`, `spec.json`,
  `streams.json` (4 streams), `writes.json` (28 typed reverse-ETL actions), `operations.json`
  (8 typed ops: 4 POST analytics read-queries, media upload/download, whatsmeow web sync/media),
  `api_surface.json` (52 endpoints — every stream/write covered; blocked ledger rows for webhooks,
  a QR-list duplicate, and 8 whatsmeow web ops as `local_workflow`), `cli_surface.json` (50 commands
  across both modes), `schemas/` (4), `fixtures/` (check + stream pages), `docs.md`.
- WhatsApp Web (whatsmeow) mode modeled as documented, config-scoped ops (no new dependency, no
  native adapter) following the `github` local-workflow precedent.
- Count touchpoints bumped for the new connector: `internal/connectors/bundleregistry/registry_test.go`
  (547→548), `internal/cli/catalog_cli_test.go` (`"count": 552`), `internal/cli/docs.go` +
  `docs/cli/connectors.md` (552/548), `internal/cli/testdata/golden_transcripts.json` (regenerated),
  `docs/connectors/{whatsapp/,catalog/all-connectors.{json,md},README.md}` (scoped whatsapp entry),
  and the four `website/**` generated connector-catalog files (regenerated to 548).

## Scope guardrail

`pm docs generate` and the website generators exhibit pre-existing drift across ~450 other connector
docs; that churn was deliberately excluded. Only the whatsapp entry + catalog count were spliced in,
keeping the PR scoped to issue #505.

## Verification

See `VERIFICATION.md`. Key gates: `connectorgen validate` 0 findings; conformance PASS for whatsapp;
`go build`/`vet`/`gofmt` clean; affected test packages green; `make docs-check` green.
Manual-GSD fallback used (Pi-only adapter unavailable under Claude Code) — recorded in `PLAN.md`.

## Sibling-review audit (Bahmni's 4 patterns)

Audited feat/whatsapp-connector against the 4 bug patterns Bahmni's review surfaced:

- (a) PHI-redaction — **APPLIED**. WhatsApp PHI/credential record fields (recipient `to`, message
  content keys, media path, `pin`, `code`, `prefilled_message`) were not matched by the reverse-ETL
  plan redactor. Fix: renamed the media-path field `file` -> `media_file` (matches an existing
  marker) and extended `commandrunner.isSensitiveRecordField` with an exact-match set for
  provider-dictated short/ambiguous keys (to/pin/code/file + message-content keys) plus substring
  markers (message/recipient/phone/msisdn/patient). Locked by `TestRedactRecordRedactsMessagingPHIFields`;
  `title`/`type`/`status`/`country_code` stay visible. certify (all connectors) still green.
- (b) offset_limit on a root-array/non-paginated endpoint — **N/A**. WhatsApp uses Graph `cursor`
  pagination (stops on empty `paging.cursors.after`); `waba` is `pagination: none`. No `offset_limit`
  and no root-array streams.
- (c) auth check passing with bad creds — **N/A**. Check hits the authenticated
  `/{WABA_ID}/phone_numbers` (401 on bad/missing token), not a public/status endpoint.
- (d) group/command hyphen-vs-underscore mismatch — **N/A**. Group ids exactly match command-path
  first tokens (all hyphenated consistently; stream/write names stay snake_case as required).

## Human gates

Parent PR merge to `main`; live Cloud API / whatsmeow credentials; live sends / template
submission; media payloads beyond fixtures; healthcare consent / template pre-approval.
