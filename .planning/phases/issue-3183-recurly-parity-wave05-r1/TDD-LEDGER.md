# TDD ledger — Recurly parity wave05 r1

| Slice | Red / failing expectation | Green evidence | Status |
| --- | --- | --- | --- |
| Official inventory | Existing Recurly bundle covered only 5 streams and excluded writes/direct/binary in legacy `api_surface.json`. | Generated operation ledger partitions all 197 official v2021-02-25 operations exactly once. | green |
| ETL streams | Existing conformance exercised only 5 stream fixtures. | 93 stream fixtures generated; `go test ./internal/connectors/conformance -run 'TestConformance/recurly' -count=1` passed. | green |
| Typed writes | Existing metadata had `write=false` and no `writes.json`. | 96 write actions validate and `write_request_shape:<action>` fixtures pass; destructive operations use destructive confirmation. | green |
| Direct/provider query | Existing Recurly had no direct-read operations or provider-style CLI metadata. | 5 fixed-path JSON preview/direct read commands validate with bounded `rest_read` operation specs and leaf-only CLI body flags. | green |
| Binary/file ledger | Existing Recurly had no binary operation metadata. | 3 official binary/export endpoints recorded as bounded binary metadata and explicitly blocked in `api_surface.json` pending shared runtime support. | green |
| Generated docs/catalog | Existing generated docs described only 5 read streams and read-only status. | Recurly connector docs and website/catalog generated surfaces updated; docs validate passed. | green |
| Captain addendum | Issues #3183-#3190 lacked the destructive-operation addendum. | Each issue has exactly one `recurly-parity-wave05-r1-captain-policy-addendum` marker with actual counts. | green |
| Verification | Required local gates had not run. | Focused/full connectorgen validation, Recurly conformance, focused CLI dynamic/golden/docs tests, build, docs validate, connector-boundary, and diff-check passed; `make verify` timeout documented as unrelated host contention. | green-with-note |

Manual GSD fallback applies because the repo-local adapter lacks a `programming-loop` command in this checkout.
