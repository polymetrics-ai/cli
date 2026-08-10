# Verification checklist — Issue #3993

- [x] Deterministic harness tests show that the supplied `Polymetrics-Cert` boundary controls both emitted cases and classification.
- [x] Deterministic harness tests show a common barrier release; write cases require an independent read-back before any PM child can start.
- [x] The two artifact self-checks pass after regeneration.
- [x] A real `pm` binary builds locally without secrets.
- [ ] The App installation credential is used without disclosure; the revoked fine-grained token is untouched.
- [ ] The whole applicable surface has a complete/failed/unavailable tally and failures are grouped by actual cause and quota bucket.
- [ ] Every created resource has read-back, inverse cleanup, and final empty-residue proof.
- [ ] GitHub → Parquet warehouse → DuckDB inbound flow is independently proven.
- [ ] The outbound workflow refusal is attributed to #3994/#3992, with no duplicate action-path implementation.
- [x] Outbound remains attributed to #3994/#3992: `./pm help flow` confirms action steps are approval-gated, and no duplicate action path was added.
- [x] Targeted local checks and required non-full-suite verification gates pass.

## Commands run

```text
node --test scripts/tests/github-live-cases.test.mjs scripts/tests/github-live-proof-sweep.test.mjs scripts/tests/github-live-lab.test.mjs
node scripts/github-live-proof-sweep.mjs --self-test
node scripts/github-live-lab-manifest.mjs --check
node scripts/github-live-bootstrap-probes.mjs --check
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

## Live-proof blocker

No provider request was sent. The isolated worktree does not contain the
captain's credential/runbook input, and the safe ephemeral credential/proof
capture foundation (#3989) plus shared REST/GraphQL admission coordinator
(#3990) are not available on this branch. Creating a normal vault credential
would violate #3989's ownership and cannot substitute for a live run. The
missing live measurements remain unchecked rather than being represented by a
synthetic 0-of-N result.
