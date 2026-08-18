# UAT — Issue #3973 mapping contract completion

All four deliverables in `SUMMARY.md` have passing automated unit coverage and
need no visual or judgment-dependent UAT:

- D1: sealed mapping and pre-session approval refusal — passed.
- D2: lossless round trip and unrepresentable value refusal — passed.
- D3: absence retains the seeded fake target row; explicit tombstone removes it — passed.
- D4: named receipt is available only after durable ledger persistence — passed.

No live PostgreSQL UAT is applicable because this PR intentionally contains no
native driver, DDL, SQL, credentials, or connection behavior. #3982 owns the
real managed-table driver proof.
