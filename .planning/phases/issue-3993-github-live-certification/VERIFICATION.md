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
```

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

One #3993-local correction was required during final verification ([#4020](https://github.com/polymetrics-ai/cli/issues/4020)): a forged
write case could previously supply target scope through `--config` rather than
the literal `--owner`/`--repo` flags. The new regression first demonstrated
that the old validation reached a fake PM credential inspection; the green
guard now rejects direct and config owner/repository aliases before a child
process starts. This consumes **1 of 5** permitted correction loops and does
not alter barrier concurrency, rate admission, or the measured 0/665 result.
