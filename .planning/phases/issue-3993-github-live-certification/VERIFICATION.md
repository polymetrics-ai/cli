# Verification checklist — Issue #3993

Status: harness slice verified inline on 2026-08-11; GitHub certification is
intentionally not ready.

- [x] Deterministic harness tests show that the supplied `Polymetrics-Cert` boundary controls both emitted cases and classification.
- [x] Deterministic harness tests show a common barrier release; write cases require an independent read-back before any PM child can start.
- [x] Write validation rejects every direct or `--config` owner/repository override before the credential-inspection child can start.
- [x] The two artifact self-checks pass after regeneration.
- [x] A real `pm` binary builds locally without secrets.
- [x] The App installation credential is used without disclosure; the revoked fine-grained token is untouched.
- [x] The whole applicable surface has a complete/failed/unavailable tally and failures are grouped by actual cause and quota bucket.
- [x] No provider mutation was dispatched; final repository read-back and temporary-run cleanup prove no created-resource residue remains.
- [x] GitHub → Parquet warehouse → DuckDB inbound flow is independently proven.
- [x] The outbound workflow refusal is attributed to #3994/#3992, with no duplicate action-path implementation.
- [x] Targeted local checks and required non-full-suite verification gates pass.
- [x] Inline/manual GSD `verify-work` is recorded in `UAT.md`; the adapter has
  no roadmap phase for issue 3993 and the dispatch prohibits role spawning.
- [x] Inline/manual standard code review is recorded in `REVIEW.md`; manual
  Shepherd-compatible evidence in `SHEPHERD-COMPAT.md` explicitly is not an
  automatic #3995 approval.

## Commands run

```text
node --test scripts/tests/github-live-cases.test.mjs scripts/tests/github-live-proof-sweep.test.mjs scripts/tests/github-live-lab.test.mjs
node scripts/github-live-proof-sweep.mjs --self-test
node scripts/github-live-lab-manifest.mjs --check
node scripts/github-live-bootstrap-probes.mjs --check
node --test --test-name-pattern='rejects a write case that overrides' scripts/tests/github-live-proof-sweep.test.mjs
go build ./cmd/pm
go run ./cmd/connectorgen surface-sync --check
make connector-runtime-preflight
make connector-canon-check
go run ./cmd/agentcontractgen check
make tidy-check
make lint
make docs-check
make smoke-no-build
make agent-contract-check
make connectorgen-validate
make connectorgen-surface-sync
make connector-boundary
make release-workflow-check
node --test --test-name-pattern='retires the legacy rate proof without current certification semantics|rejects an unsafe bootstrap manifest before inventory hashing|rejects unsafe supplied bootstrap inventories before drift validation' scripts/tests/github-live-rate-limit-proof.test.mjs scripts/tests/github-live-lab.test.mjs
node --test scripts/tests/github-live-cases.test.mjs scripts/tests/github-live-lab.test.mjs scripts/tests/github-live-proof-sweep.test.mjs scripts/tests/github-live-rate-limit-proof.test.mjs
```

The named synthetic command intentionally returned three RED failures before
the production edits. The final focused command passed all 50 tests without a
credential, provider command, or network test.

## Credentialed result and evidence handling

The App-authenticated fresh binary was exercised only against the immutable
`Polymetrics-Cert` boundary. The current canonical classifier records 1,521
commands: 182 attemptable, 1,339 blocked, and 169 / 470 direct reads. Its static
movement from the separately named historical terminal measurement is -483 / +483.
No provider command was run for that correction.

The historical full one-barrier release records 665 attempted operations: 0
proven, 665 terminal-bound failures, and 856 untestable rows. No HTTP status was
available for the failures. Rate snapshots immediately around that historical
sweep measured REST core 15,000 → 14,997 and GraphQL 5,000 → 5,000.

The App token itself was not printed or committed.  A single returned-record
control and a 16-way returned-record control establish that the App path worked;
the all-at-once 665-child result is a local harness scaling finding, not a
synthetic authentication result or a #3990 policy implementation.

The inbound flow was planned, previewed, run, and then queried through DuckDB:
1 record read, 1 record written, 1 row returned from the connection-scoped
Parquet table.  The temporary project (including local credential state and
warehouse files) was deleted after its sanitized evidence was checked.

## Verification interpretation

The 0/665 barrier result is verified as a complete, truthful measurement, not
as a provider-certification success. The isolated returned-record controls and
rate snapshots rule out treating it as a revoked credential or a #3990 quota
policy result. Follow-up is explicitly limited to #3990 admission, cleanup-safe
fixtures, and outbound #3994/#3992 foundations.

## Final correction loop

The completed correction sequence is `1/#4020 -> 2/#4022 -> 3/#4027 ->
4/#4039 -> 5/#4050`. Round 1 closed the direct/config write-scope escape
before process launch; round 3 made pre-barrier inputs fail closed; round 4
bound current artifacts to the canonical classifier; and round 5 retired the
legacy rate proof's current-certification claim and guarded bootstrap artifact
input/output. The three named round-5 RED tests were observed failing before
production edits, then the focused 50-test GREEN suite passed. This completes
**5 of 5** permitted correction loops without altering barrier concurrency,
rate admission, or the measured historical 0/665/856 result.
