## Intent

Foundation slice G1/G2/G6 from the Firstmate launch brief (no GitHub issue was
supplied): make the generated parity projection trustworthy, preserve delete as
an independently selectable direct-write mutation, and make accepted-evidence
publication safe under concurrent readers.

Base: `integration/4015-mvp-flat-r1` → `main`.

## What Changed

- Added one generated classifier for the eight operation kinds and five parity
  classes. Sweep rows now always carry `operation_kind` and `op_class`; declared
  write actions also carry `write_action_kind`.
- Regenerated GitHub's 1,571-row sweep. GitHub delete actions now project
  `rest_write/direct_write` with `write_action_kind=delete`; executable reads
  project `rest_read/direct_read`.
- Added capability and managed-transport classifier coverage: Zoom/GitLab
  `capability:read`, PostgreSQL's declared managed destination despite generic
  `write:false`, and the equivalent MySQL managed-destination contract.
- Replaced direct final-file proof writes with prevalidated, staged, fsynced,
  hard-link no-replace publication. Report, transport, and change-capture
  imports prepare their complete batch before any final evidence name appears.
- Changed the live runner to draft → Go import → connector-scoped matrix
  generation → connector-scoped check. It never removes accepted evidence when
  generation/checking reports drift.
- Documented the projection, `(connector × op_class)` work identity, mutable
  resource keys, and atomic fan-in ordering; added the short AGENTS pointer.

## Red / Green / Refactor Evidence

- Red: new G1/G2 tests initially failed because neither classifier nor sweep
  projection fields existed. New G6 tests initially failed because no prepared
  atomic publisher existed; a scoped-check test failed because the old check
  read another connector's shard.
- Green: focused `cmd/connectorgen` tests cover valid CLI intent/operation/
  write/stream inputs, binary, CDC, changefeed, managed transport, N/A,
  mismatch and delete cases; they pass along with byte-drift checks.
- Green: batch-prevalidation, no-replace, and a concurrent matrix-reader test
  pass. The runner-order regression rejects the old publish/check/delete path.
- Refactor: all evidence paths share the same staged publisher and only final
  fan-in owns global status.

## Live GitHub Proof

Read-only fixture command (the Keychain reference was expanded only into the
child process environment; no credential value is present in this PR):

```sh
PM_CERT_GITHUB_TOKEN="$(security find-generic-password -s pm-cert-classic -w)" \
  node scripts/certify-connector-live.mjs github \
  --pm .tmp/pm-live-certification \
  --credential-env PM_CERT_GITHUB_TOKEN --credential-field token \
  --credential-config owner=Polymetrics-Cert \
  --credential-config repo=pm-cert-3993-20260810-wz0fru \
  --credential-config rate_limit_account=polymetrics-ai-certification \
  --limit 1
```

Output: `github certification: executed=1 certified=1 no_object=0
wrong_credential=0 entitlement=0 product_defect=0`.

The `repo read-file` request returned HTTP 200 and published
`internal/connectors/certifications/evidence/github-direct_read_sweep_repo_read_file-github-1787095680987-19f5568235a6.json`.
The same run performed scoped GitHub matrix generation followed by:

```sh
go run ./cmd/connectorgen certification-matrix --connector github --check
```

Output: `certification shards are current: connectors=1 capability_complete=0
certified=0`. The accepted record remained present after the check. The
concurrent-reader regression uses the real temporary filesystem and confirms a
matrix reader sees either no record or strictly valid JSON.

## Testing

```text
go test -timeout 20m ./cmd/connectorgen                         PASS
go test -timeout 20m ./internal/cli                             PASS
go list ./... grouped into timeout-bounded go test invocations  PASS
go vet ./...                                                    PASS
go build ./cmd/pm                                               PASS
gofmt -w cmd internal                                           PASS
git diff --check                                                PASS
make lint                                                       PASS
make docs-check                                                 PASS
make smoke-no-build                                             PASS
make connector-boundary                                         PASS
go run ./cmd/agentcontractgen check                             PASS
go run ./cmd/connectorgen validate internal/connectors/defs     PASS
go run ./cmd/connectorgen surface-sync --check                  PASS
go run ./cmd/connectorgen certification-sweep --connector github --check    PASS
go run ./cmd/connectorgen certification-candidates --connector github --check PASS
go run ./cmd/connectorgen certification-matrix --check          PASS
node --test scripts/tests/github-*.test.mjs                     PASS
node scripts/gen-github-graphql-parity.mjs --check              PASS
node scripts/github-combined-operation-ledger.mjs --check       PASS
scripts/tests/{connector-canon,pinned-build-dependencies,homebrew-release-notify,release-target-parity}.sh PASS
```

## GSD / Skills / Delivery

- Lifecycle: `scripts/gsd doctor`, `scripts/gsd sources` for discuss/plan/
  execute/verify/review, the corresponding generated prompts, and
  `go run ./cmd/agentcontractgen check`. The prompts were executed inline under
  the documented manual fallback because this task's contract prohibited
  compatible isolated role spawning.
- Skills: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`,
  `golang-concurrency`, `golang-database`, and `golang-documentation`.
- Delivery mode: direct PR; no no-mistakes pipeline was run, per the captain
  instruction.

## Scope Notes and Follow-up

- No G3–G5, G7–G14 work, allowlist edit, or global `--all` refresh is included.
- At base `d842c815a`, MySQL declares `write:false` but no
  `sync_transport.json`/managed-destination descriptor. The classifier test
  proves the required handling once that exact declaration exists; adding an
  unproven MySQL transport is intentionally deferred rather than fabricated.

## Automated Review

Pending PR-open automatic Claude review. Route: `claude_auto` if GitHub runs it
for this stacked, non-draft PR; otherwise record parent-PR fallback or the
review blocker before integration. No Copilot request is made unless Claude is
unavailable.
