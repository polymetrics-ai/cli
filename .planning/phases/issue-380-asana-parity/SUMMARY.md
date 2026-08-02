# Summary — Asana connector parity (#380)

Status: implemented locally on `fm/cli-asana-parity-wave02-r1`; ready for the requested implementation commit.

Manual GSD fallback was used because `scripts/gsd` does not expose `programming-loop` in this adapter. Plan, TDD ledger, and verification checklist were created before production edits and updated with evidence.

## Implemented

- Replaced Asana API surface with a 249-row pinned official OpenAPI ledger: missing=0, extra=0, duplicate=0.
- Added 249 typed Asana operation records and 249 provider-style fixed-target CLI metadata records without raw API passthrough.
- Preserved 12 implemented read streams and 13 implemented reverse-ETL writes.
- Hardened implemented delete actions with `confirm: "destructive"`, idempotent 404 handling, redacted path fields, and documented plan -> preview -> approval -> execute safety.
- Added fixture-only `certification.json` truthfulness metadata; no live certification claim.
- Updated `docs.md`, generated Asana connector manual/skill docs, and CLI golden transcript fixture.

## Verification snapshot

Pass: `connectorgen validate`, Asana conformance, full connector conformance, engine tests, focused dynamic/golden/docs CLI tests, docs validate, `go vet ./...`, `go build ./cmd/pm`, `make connector-boundary`, `git diff --check`.

Timeboxed: broad certification-heavy `go test ./internal/cli -run 'Connector|Dynamic|Golden'` and `go test ./...` exceeded package/test timeouts while repeatedly loading all connector bundles; no Asana credentials or live provider calls were used.
