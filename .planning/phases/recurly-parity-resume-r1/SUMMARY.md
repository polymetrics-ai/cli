# Recurly parity-resume r1 summary

Status: pre-citation implementation is complete; the requested no-mistakes validation is in progress,
followed by final citation-convention integration and a focused gate rerun.

Paused only for the shared request-contract convention: its source lane remains active and has not
landed on `origin/main`. The branch intentionally does not claim completion until its 538 raw
field citations are represented in that canonical, enforced format.

Current executed evidence:

- Recurly's v2021-02-25 provider OpenAPI reports 197 method/path operations. The bundle has exact
  coverage for all 197: 93 streams, 96 typed writes, five JSON direct reads/previews, and three
  bounded binary downloads. There are zero planned or blocked operations.
- The raw provider matrix contains 2,951 provider request inputs. The local-field reconciliation
  covers 538 exposed Recurly request fields with zero unmatched fields. These artifacts deliberately
  await the shared citation metadata convention rather than inventing a competing shape.
- Focused Recurly conformance, commandrunner's all-implemented-command preflight, Recurly binary
  execution fixtures, surface sync/check, focused validation, vet, `pm` build/help, root CLI golden
  generation, and website data generation have passed.
- `pm connectors inspect recurly --json` confirms 93 registered streams, 96 registered typed write
  actions, and read/write capability support. The complete `go test ./internal/cli/... -count=1`
  passed after the root-help transcript regeneration.
- A live provider audit followed the requested `developers.recurly.com` reference redirect and
  verified the pinned OAS SHA-256 `b98a3f85d0a1190c2c8e11f57fa5ec13b841665e658596dcb5d7f3ddce70baca`.
  It contains exactly 197 operations: 97 GET reads and 100 POST/PUT/DELETE writes. The bundle has
  all and only those 197 method/path pairs.
- Recovery initially replaced five current-main stream contracts with generic schemas and fixtures.
  The restored contracts retain their primary/cursor/required metadata, typed fields, query defaults,
  computed projections, and fixture response bodies. A new connector-local regression test locks
  this preservation in place; `base_url` also retains its prior `uri` format.
