# TDD ledger — Airtable official API parity

## Red/green plan

| Slice | Red evidence before production edit | Green evidence target | Final evidence |
| --- | --- | --- | --- |
| Audit ledger | Current `api_surface.json` had 30 rows, not the 103 official OpenAPI operations. | Bundle validation accepts 103 partitioned rows. | `api_surface.json` tracks all 103 audited operations: 28 streams, 44 writes, 1 direct read, and 30 blocked operations. |
| Streams/fixtures | Current bundle had 5 streams and fixtures only for those streams. | Conformance `TestConformance/airtable` passes with fixtures for every executable stream. | 28 streams have sanitized fixtures; comments use the exact per-record endpoint instead of bulk fanout. |
| Writes/fixtures | Current bundle had 12 write actions and lacked broad destructive/admin coverage. | Conformance write request-shape checks pass for every executable write action. | 44 typed write actions have fixtures; attachment upload remains blocked until base64 and decoded-size validation is enforceable. |
| Direct read CLI | Current Airtable bundle had no `operations.json`/`cli_surface.json`, so `pm airtable hyperdb get-records` was unknown. | `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` includes Airtable direct read coverage and passes. | Added HyperDB operation/CLI command and fixture server test; CLI dynamic/golden gate passed. |
| Docs/catalog | Current generated docs/catalogs reported 5 streams / 12 writes. | Generated docs/catalogs report actual post-change stream/write/operation counts. | Regenerated Airtable docs, connector catalog, website connector data, and golden transcripts. |
| Review hardening | Webhook replay expected a non-executable limit, comments ignored `record_id`, attachment upload accepted unbounded strings, and delivery artifacts retained pre-containment counts. | Replay uses `limit=50`; comments issue one exact per-record request; attachment upload stays blocked without decoded-size enforcement; artifacts report the contained partition. | Focused Airtable definition and conformance tests are the review-fix gate for the 28/44/1/30 partition. |
| CI issue-link checkpoint | Red captured: `go test ./internal/coordination/issueguard -run 'TestValidatePRAcceptsUnvalidatedCheckpointCanonicalIssueLinks|TestValidatePRRejectsAmbiguousIssueRelationship' -count=1` failed because PR #3540's generated checkpoint body contains canonical issue URLs for #3070-#3077 but no recognized relationship token. | Guard extracts non-closing issue references only from the generated canonical-link section when the checkpoint identifies a completed task, while retaining negative coverage for standalone bare URLs and vague issue wording. | Green: focused guard tests and vet pass; the exact CI-shaped invocation reports `issueguard: ok (8 linked issues)`. |

## Captured red baseline

- `internal/connectors/defs/airtable/api_surface.json`: 30 endpoint rows; 17 covered; 13 excluded; no operation rows.
- `internal/connectors/defs/airtable/streams.json`: 5 streams.
- `internal/connectors/defs/airtable/writes.json`: 12 actions.
- `cli_surface.json`: missing.
- `operations.json`: missing.
- `certification.json`: missing.
- Official OpenAPI re-audit: 103 operations with lane counts `27/69/1/1/5`.

## Verification ledger

See `VERIFICATION.md` for command results. Earlier full gates predate review hardening; the current phase records a separate focused Airtable definition and conformance gate for the final contained tree.
