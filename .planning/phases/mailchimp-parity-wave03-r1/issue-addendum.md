## Captain-policy addendum — Mailchimp parity wave03-r1 local status

Branch/worktree: `fm/cli-mailchimp-parity-wave03-r1` (isolated Treehouse checkout). No push/PR/merge performed by this worker.

Post-change operation counts, measured from the committed bundle (not copied from the audit baseline):

- Official Mailchimp Marketing API operations represented: **298**.
- Dispositioned as executable: **295** total, every one reachable as its own `pm mailchimp <command>`:
  - **79** ETL stream rows.
  - **68** typed direct-read/search rows.
  - **148** approval-gated reverse-ETL write rows.
- Blocked rows: **3** (`GET /`, `GET /ping`, and generic `POST /batches`), each carrying a reason and an
  official `source_url`.
- Excluded/N/A rows: **0**.
- Undispositioned rows: **0**.
- Live-certified rows: **0** (no credentialed/live Mailchimp calls were made).

Baseline for comparison was 4 implemented / 291 blocked-or-planned / 3 excluded; the 3 previously
counted as excluded are now carried as blocked with citations, which is why excluded is 0.

Fixture and executability coverage:

- ETL streams: **79/79** have replay fixtures under `fixtures/streams/<stream>/` (no orphans).
- Reverse-ETL writes: **148/148** have request-shape fixtures under `fixtures/writes/<action>.json`
  (no orphans).
- Direct reads: **68/68** are covered by `connectorgen validate` plus the repo-wide runtime guard
  `TestEveryImplementedCommandPassesRuntimePreflight`. The conformance harness has no operation
  fixture slot and no connector in the repo ships one, so no new fixture category was invented.

Destructive-action contract (audited across all 148 actions):

- **57** destructive actions gated with `confirm: "destructive"`; all 57 declare required fields.
- **148/148** carry closed typed schemas (`additionalProperties: false`); **0** use a banned
  `oneOf`/`anyOf` schema root.
- **35/35** DELETE actions declare provider idempotency (`delete.idempotent: true`,
  `missing_ok_status: [404]`).
- **55** actions declare field redaction — every action whose schema carries PII. The remaining
  destructive actions take opaque IDs only, so they have nothing to redact.

Safety notes:

- No secrets requested, printed, stored, or summarized.
- No live provider calls or provider writes.
- No generic raw HTTP/batch write surface; `POST /batches` remains blocked by policy.
- Reverse ETL remains plan -> preview -> explicit approval -> execute, with destructive confirmation on risky actions.

Local gates completed (re-run after rebase onto `origin/main` @ `d8082031e`, which includes #3890-#3895):

- `go run ./cmd/connectorgen validate internal/connectors/defs/mailchimp` — 0 findings
- `go test ./internal/connectors/conformance -run 'TestConformance/mailchimp' -count=1` — ok
- `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` — ok (196s)
- `go test ./internal/connectors/... -count=1` — ok, 129 packages
- `go test ./internal/connectors/commandrunner -run 'TestEveryImplementedCommandPassesRuntimePreflight'` — ok
- `go build ./cmd/pm`, `go vet ./...`, `gofmt -l cmd internal` — clean
- `make connector-boundary` — `outcome: clean`, `findings: []`
- `make verify`
- `git diff --check` — clean

Rebase note: this round landed on a main that now enforces #3890 direct-read runtime preflight. All
68 direct reads initially failed that guard because the derived runtime endpoint ledger still carried
`"mailchimp": []`. Fixed by running `go run ./cmd/connectorgen surface-sync`, the generator that owns
`internal/connectors/defs/operation_endpoint_ledger.json`; the resulting diff is a single hunk adding
mailchimp's 68 entries and touching no other connector. No shared runtime behavior was edited.

Final commit remains local for handoff. No push, PR, merge, or `/no-mistakes` run was performed by
this worker, and this addendum claims no live certification.
