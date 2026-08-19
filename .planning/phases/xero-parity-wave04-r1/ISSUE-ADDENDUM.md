<!-- xero-parity-wave04-r1-captain-policy-addendum -->
## Xero parity wave04-r1 local verification addendum

Branch-local implementation completed fixture-backed Accounting API parity without live Xero calls or credentialed provider checks.

Truthful re-audit counts from `xero_accounting.yaml` (`info.version=16.1.0`, fetched from the official XeroAPI/Xero-OpenAPI source):

- 235 official Accounting API operations (`GET=126`, `PUT=53`, `POST=46`, `DELETE=10`).
- 78 ETL/read operations.
- 11 bounded direct/report read operations.
- 87 reverse-ETL write operations.
- 59 binary/file operations (attachment paths plus the four official `/pdf` file-returning GET operations). Of these, 11 attachment metadata list reads are stream-covered and 48 binary/PDF download or attachment upload transfers remain blocked on the shared binary/file executor.
- 0 CDC/changefeed operations in the Accounting API source.

Connector-local disposition:

- `api_surface.json`: 235 exact official operation rows, `operation_ledger_version: 1`, no stale `/api.xro/2.0` prefixes, no legacy `excluded` rows.
- `streams.json`/fixtures: 100 stream fixture directories.
- `writes.json`/fixtures: 87 typed write actions and 87 sanitized write fixtures, including the two missing destructive BankTransfers status-delete POST actions.
- `operations.json`: 70 bounded report/binary/file metadata operations.
- `cli_surface.json`: implemented provider-style report direct reads; binary/file transfer commands remain planned/blocked.
- `certification.json`: fixture-only metadata; no live certification claim.

Safety policy confirmed locally: no raw SQL/query, arbitrary GraphQL, generic HTTP method/path/body, shell, file, or passthrough escape hatches were added. Reverse ETL remains plan → preview → explicit approval → execute, destructive actions are typed and `confirm: "destructive"`, and file upload/download execution remains blocked until the shared runtime can enforce approved provider paths, payload digests, redaction, and bounded streaming.

Local gates passed before commit:

- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go test ./cmd/connectorgen -run 'TestXero' -count=1`
- `go test ./internal/connectors/conformance -run 'TestConformance/xero' -count=1`
- `go test ./internal/connectors/engine -run 'TestXeroReportOperationsDirectRead' -count=1`
- `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1 -timeout=20m`
- `go build ./cmd/pm`
- built CLI help/docs spot checks and `./pm docs validate --connectors-dir docs/connectors`
- `make connector-boundary`
- `make verify`
- `git diff --check`
