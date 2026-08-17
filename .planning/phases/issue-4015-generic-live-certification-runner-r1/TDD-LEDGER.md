# TDD ledger — generic live certification runner

| Slice | Red | Green | Observable proof |
| --- | --- | --- | --- |
| Connector-parameterized runner | `test ! -e scripts/certify-connector-live.mjs` proved the shared runner was absent. | The new script derives candidates/eligibility/credential defaults from the selected definition bundle and records an explicit no-definition/no-candidate non-pass where needed. | `node --check`; 36 `--definition-check` invocations passed; unchanged Freshchat invocation reported its no-candidate non-pass. |
| Accepted proof record | Before the runner existed, no script could produce a v2 observed-operations record. | Each passing operation receives a complete sanitized proof and an immediate matrix validation. | The live eligible sweep executed 122 candidates and wrote 38 accepted records; `go run ./cmd/connectorgen certification-matrix --check` passed after every record and at the end. |
| Secret safety | No script boundary guaranteed that a captured provider payload was fingerprinted before persistence. | Raw command output is held only in runner memory, scalar values are HMAC fingerprinted, and the credential is scanned for before evidence commit. | Accepted records contain fingerprint markers rather than scalar material; the runner/evidence/receipts scan found no credential-shaped content. |

No Go code is added or altered; the red/green steps constrain the new Node runner's observable behavior.
