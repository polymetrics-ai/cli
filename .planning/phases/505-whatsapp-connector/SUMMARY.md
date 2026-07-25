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

## Human gates

Parent PR merge to `main`; live Cloud API / whatsmeow credentials; live sends / template
submission; media payloads beyond fixtures; healthcare consent / template pre-approval.
