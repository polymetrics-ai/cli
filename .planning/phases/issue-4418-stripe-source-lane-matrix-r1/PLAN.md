# Plan — #4418 Stripe source-to-seven-lane matrix

## Boundaries

- Only add `internal/connectors/defs/stripe/sources/stripe-source-lane-matrix.json` and `internal/connectors/defs/stripe/source_lane_matrix_test.go`, plus this planning evidence.
- Do not change `cmd/`, `internal/engine`, shared source import/admission validation, generated artifacts, runtime/executor code, credentials, or network behavior.
- Do not derive ETL, reverse ETL, sync transport, or binary execution from an HTTP method. They remain source-cited mapping candidates only.

## Steps

1. Add a local failing contract test before the matrix exists. It will decode the existing source lock and artifacts, require every source ID exactly once with an exact source location, require all seven cells, and verify typed source reasons.
2. Materialize the matrix mechanically from the pinned Stripe source lock. Store the fact record necessary for classification: path/query scope, request/response media, paging evidence, and lack-of-event/cursor documentation. Attach an existing-artifact backlink to every source row without allowing artifacts to create IDs.
3. Make the test pass by checking the exact count reconciliation: 589 operations, 4,123 cells, 263 GET/direct-read candidates, 326 mutation/direct-write and reverse-ETL candidates, 128 source-paging ETL/sync candidates, one PDF download candidate, and one multipart upload candidate.
4. Add adversarial test mutations for a hidden source row, an invalid artifact backlink, an invalid paging disposition, an invalid mutation reverse-ETL disposition, and an incorrect source count.
5. Run `gofmt`, focused Go tests, JSON parsing, source/map checks, and a broad-suite baseline. Record unrelated baseline failures without repairing them.

## Matrix rules

| Lane | Mapping disposition rule |
| --- | --- |
| `direct_read` | Every source-documented GET response is `mapped_unproven`; mutation rows are source-evidenced `not_applicable`. |
| `direct_write` | POST and DELETE rows are `mapped_unproven`; GET rows are source-evidenced `not_applicable`. |
| `binary_download` | Only the cited PDF response is `mapped_unproven`; every other row is source-evidenced `not_applicable`. |
| `binary_upload` | Only the cited multipart request is `mapped_unproven`; every other row is source-evidenced `not_applicable`. |
| `etl` | The exact 128 documented paging candidates are `mapped_unproven`; all other rows, including list-shaped rows without continuation, are source-evidenced `not_applicable`. |
| `reverse_etl` | POST and DELETE rows are `mapped_unproven`; GET rows are source-evidenced `not_applicable`. |
| `sync_transport` | The same 128 documented paging candidates are `mapped_unproven`; all other rows are source-evidenced `not_applicable`. |

`mapped_unproven` is a source mapping state, never an `implemented` runtime claim. This plan deliberately has zero `implemented` and zero `missing_foundation` cells.

## Artifact backlinks

The matrix records a link to every existing `api_surface.json` endpoint (`METHOD path`) for every source row. Existing `streams.json` records link the five named streams to their exact direct-read/ETL/sync candidates, and existing `writes.json` records link the three named actions to their exact direct-write/reverse-ETL candidates. The validator rejects links whose source ID, lane, or artifact record does not exist.
