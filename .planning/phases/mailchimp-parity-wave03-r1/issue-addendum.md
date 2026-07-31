## Captain-policy addendum — Mailchimp parity wave03-r1 local status

Branch/worktree: `fm/cli-mailchimp-parity-wave03-r1` (isolated Treehouse checkout). No push/PR/merge performed by this worker.

Post-change operation counts from the fixture-backed Mailchimp definition:

- Official Mailchimp Marketing API operations represented: **298**.
- Fixture-backed implemented rows: **295** total:
  - **79** ETL stream rows.
  - **68** typed direct-read/search rows.
  - **148** approval-gated reverse-ETL write rows.
- Blocked/local-workflow rows: **3** (`GET /`, `GET /ping`, and generic `POST /batches`).
- Excluded/N/A rows: **0**.
- Live-certified rows: **0** (no credentialed/live Mailchimp calls were made).

Safety notes:

- No secrets requested, printed, stored, or summarized.
- No live provider calls or provider writes.
- No generic raw HTTP/batch write surface; `POST /batches` remains blocked by policy.
- Reverse ETL remains plan -> preview -> explicit approval -> execute, with destructive confirmation on risky actions.

Local gates completed:

- `go run ./cmd/connectorgen validate internal/connectors/defs/mailchimp`
- `go test ./internal/connectors/conformance -run 'TestConformance/mailchimp' -count=1`
- `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`
- `go build ./cmd/pm`
- `make connector-boundary`
- `make verify`

Final commit will remain local for handoff; this addendum does not claim live certification.
