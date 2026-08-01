# SUMMARY — issue #3247 Marketo parity

Status: implemented and locally verified on branch `fm/cli-marketo-parity-wave05-r1`.

## Delivered

- Expanded Marketo from a 3-stream, read-only partial bundle to the full official AdobeDocs Swagger operation ledger.
- Current ledger: 327 official operations = 117 fixture-backed ETL/changefeed streams + 28 bounded redacted direct reads + 158 typed reverse-ETL writes + 24 blocked/not-applicable rows.
- CLI metadata now exposes 303 Marketo commands: 117 ETL, 28 direct reads, 158 reverse ETL.
- Marketo metadata enables `read` and `write` while keeping generic `query=false`.
- Added `writes.json`, `operations.json`, `cli_surface.json`, `certification.json`, schemas, fixtures, docs, and count/safety tests.
- Blocked unsafe/not-applicable operations instead of exposing escape hatches: 10 binary/InputStream downloads, 2 identity token issuance operations, 11 write-query-selector operations, and 1 dynamic custom-field body operation.

## Safety posture

- No live Marketo calls and no credentials used.
- No raw API/query/write command, arbitrary request body, shell, file, or passthrough escape hatch.
- Reverse ETL remains plan → preview → approval → execute.
- Destructive writes require typed selectors and `confirm: destructive`.
- Operations that need structured write query parameters are blocked until shared runtime support exists.

## Verification

See `VERIFICATION.md`. Final `make verify` passed after rerunning one unrelated flaky certify timing test that failed on the first attempt and passed in isolation.
