---
phase: issue-4081-warehouse-mediated-transport-demo-r1
plan: 01
type: tdd
status: green_pending_verify_review
---

# #4081 — construct demonstrable warehouse-mediated Transport

**Issue:** [#4081](https://github.com/polymetrics-ai/cli/issues/4081)
**Parent issue:** [#4015](https://github.com/polymetrics-ai/cli/issues/4015)
**Required base:** `docs/4015-connector-release-certification` at
`e7d2b2963fc1dd164f63b31fccb8a3bab8084bec` (squashed #4019, including #4077)
**Final child branch:** `feat/4081-warehouse-mediated-transport-demo`
**Final draft PR base:** `docs/4015-connector-release-certification`

## Manual GSD lifecycle

The commands were resolved through `scripts/gsd doctor`, `scripts/gsd sources`,
and `scripts/gsd prompt`. The named issue phase is outside the numeric roadmap,
and compatible isolated GSD roles are unavailable/forbidden by the repository's
single-worker contract. Execute the lifecycle inline and record it here:

1. `scripts/gsd prompt discuss-phase issue-4081-warehouse-mediated-transport-demo-r1 --auto`
   → `CONTEXT.md` and `DISCUSSION-LOG.md`.
2. `scripts/gsd prompt plan-phase issue-4081-warehouse-mediated-transport-demo-r1 --tdd --skip-research --auto`
   → this plan, `TDD-LEDGER.md`, `VERIFICATION.md`, and `CHECKPOINT-PLAN.md`.
3. With the required base guard recorded below, execute RED → GREEN → refactor
   inline, recording each committed checkpoint and exact commands in
   `TDD-LEDGER.md` and `VERIFICATION.md`, then run `verify-work` and
   `code-review`. Use gap planning only for an observed verification gap.
4. Run the child `no-mistakes axi run` only after the implementation commit;
   do not use `--yes`, do not hand-edit during an active run, and stop at five
   correction loops.

## Base admission checkpoint — passed before any code or RED test

```text
git fetch --no-tags origin refs/heads/docs/4015-connector-release-certification:refs/remotes/origin/docs/4015-connector-release-certification
git rev-parse origin/docs/4015-connector-release-certification
# e7d2b2963fc1dd164f63b31fccb8a3bab8084bec
git rev-parse e7d2b2963fc1dd164f63b31fccb8a3bab8084bec:internal/synctransport/types.go
git rev-parse aaf288d069adc1b67a09500afcca4be4a6d1bab3:internal/synctransport/types.go
# both 7b5eafd34c78bea690dcb23eee716ea840ea813e
```

PR #4019 was squash-merged, so `aaf288d...` is intentionally **not** an
ancestor of the combined head. Admission is proven by exact accepted-content
identity instead: all four Transport blobs (`types.go`, `orchestrator.go`,
`registry.go`, and `transport_test.go`) at `e7d2b296...` equal their `aaf288d...`
counterparts, and the #4077 phase evidence is present in the combined tree.
The final child branch was created directly from that just-fetched combined head.
No test or production file precedes this planning checkpoint.

## Frozen smallest design

This is deliberately one closed GitHub `issues` page → GitHub label action
walking slice. It does not create a reusable connector-to-connector framework
or declare support for another provider family.

1. `synctransport.WarehouseStage` changes from a source-record handoff to
   `Stage(context.Context, WarehouseStageRequest) (WarehouseReceipt, error)`
   plus `Reopen(context.Context, WarehouseReceipt) (WarehouseWorkset, error)`.
   `WarehouseReceipt` carries only `ID`, `Owner`, `Generation`,
   `ManifestSHA256`, `ContentSHA256`, and bounded counts. It never carries
   records or tombstones. `WarehouseWorkset` is reconstructed only by
   `Reopen`, includes the receipt it came from, and is cloned again before the
   destination sees it.
2. The concrete app stage is connection-owned: `WarehouseStageRequest` carries
   the opaque connection ID and generation, never a caller file path. It writes
   one run-generated JSONL WAL under the existing owner-scoped location,
   materializes one DuckDB-written Parquet table, atomically writes a manifest,
   fsyncs the file/directory chain, and only then returns the receipt. `Reopen`
   rereads the owner record and manifest and compares owner, generation,
   manifest hash, WAL/Parquet content hash, and bounded count before reading the
   Parquet artifact.
3. The one composition root, `newGitHubWarehouseMediatedTransport`, is called
   by `App.Open`. It installs a non-nil app stage, a read-only accepted-evidence
   verifier, and exactly two declarative GitHub registrations:
   `github_issues_source` and `github_issue_label_destination`. The GitHub
   definition declares only `issues`/`full_append` for this slice. The verifier
   accepts the two fixed evidence references from a compiled allow-list; neither
   descriptor nor executor can self-certify.
4. The source uses the existing declarative GitHub engine `Read` against one
   declared bounded `issues` page. It emits only the configured singleton issue
   into the workset and the reopened singleton is rechecked against that source
   predicate before the destination can apply. The destination consumes only
   reopened records. It derives one configured, separate target issue and label
   through the existing typed
   `add_issue_labels` plan → preview → approval → execute path; it independently
   reads back the target with the same declared GitHub reader before returning
   a durable receipt. `remove_issue_label` uses the same typed path as its
   inverse and is invoked by the demo cleanup.
5. `DestinationExecutor.ApplyDestination` returns a typed durable receipt and
   `ReadBackDestination` verifies it. The orchestrator performs
   `Stage → Reopen → ApplyDestination → ReadBackDestination → checkpoint CAS`.
   A destination/read-back error produces no CAS. A CAS error leaves the same
   receipt and workset artifact intact for deterministic replay.

The test-only faithful GitHub server is the local proof authority. A real
provider is attempted only after the local proof is green and the normal
encrypted GitHub App credential boundary resolves without inspection; otherwise
the harness emits the fixed safe-blocker code before provider I/O.

## TDD execution slices

### 1. Planning checkpoint

- Commit the phase artifacts and immutable issue/topology traces on the final
  child branch after base admission.
- Re-read the exact base's existing Transport, warehouse, GitHub engine/direct
  write, checkpoint, and certification symbols before freezing file names.
- Confirm the target connector is GitHub only and record changed-path ownership.

### 2. RED — reject dormant/unsafe construction

- Add app-level production-construction tests that open the application through
  its real constructor and demonstrate: empty registry/unavailable verifier;
  nil stage; no exact GitHub source/destination registration; and no source I/O
  before preflight failure.
- Add stage contract tests proving a destination cannot receive raw `SourcePage`
  or source-owned records; in-memory workset aliases are rejected or cannot
  satisfy the production adapter.
- Add tamper tests for owner, generation, manifest, and content identity;
  `Reopen` must fail closed before records are delivered.
- Add failure-order tests: destination failure leaves checkpoint unchanged;
  checkpoint-store failure keeps the same durable handle/receipt replayable;
  destination receipt/read-back must precede checkpoint CAS.
- Add a #4079 regression control using mutable raw JSON/string-map values at
  source, stage, reopen, and destination boundaries.
- Commit this test-only RED and ledger evidence before touching production code.

### 3. GREEN — minimal closed composition

- Add one explicit app/demo composition function that installs a non-nil durable
  stage, read-only accepted-evidence verifier, and exact GitHub source/destination
  adapters. Missing/typed-nil inputs remain rejecting defaults.
- Adapt existing connection-owned WAL, DuckDB, Parquet, and structural layout
  primitives. `Stage` must fsync/materialize/publish before minting immutable
  owner/generation/manifest/content receipt data; do not duplicate warehouse
  semantics or use caller paths.
- Implement bounded `Reopen(handle)` that reasserts ownership and identity and
  creates independent record copies from the durable artifact. Source/page
  references are explicitly discarded before destination planning/apply.
- Consume the existing declarative GitHub read and typed, approval-bound write
  paths. The adapter never exposes generic request construction, raw provider
  cursors, source connector, destination connector, or either credential to its
  counterpart.
- Require a durable typed destination receipt plus independent provider read-back
  before the existing checkpoint CAS is invoked. Preserve deterministic replay
  for a checkpoint persistence failure and never advance for a destination or
  read-back failure.

### 4. Exact-binary demo and refactor

- Add one bounded faithful-server integration test that builds a fresh `pm`,
  records tested commit, SHA-256 and byte size, starts a faithful GitHub test
  server, initializes an isolated project, and drives the accepted closed
  plan/preview/stdin-approved `pm etl run` carrier through one read → stage →
  discard → reopen → typed destination mutation → independent read-back →
  receipt/CAS → separately planned typed inverse → repeat cleanup → zero-residue
  path.
- Emit only sanitized machine-readable facts: binary digest/size, owner/workset
  and Parquet identities/hashes/counts, receipt sequencing, checkpoint outcome,
  and cleanup/read-back result. Do not serialize credential values or raw
  provider payloads.
- Use an approved real GitHub App run only if normal PM credential resolution and
  the disposable resource boundary succeed. Otherwise return the documented
  blocker before provider I/O and label the faithful-server proof local.
- Run `gofmt`, retain closed error context, and make all new handles/records
  defensive copies. Do not add a dependency, broad public command, or generic
  writer.

## CLI help/manual/website parity

The accepted carrier is a visible, closed command family rather than the
rejected one-shot demo command:

```text
pm etl transport github-issue-label plan --connection <name> [--json]
pm etl transport github-issue-label preview <plan-id> [--json]
pm etl run --connection <name> --stream issues --batch-size 1 \
  --approval-plan <plan-id> --approval-token-stdin --confirm destructive [--json]
pm etl transport github-issue-label cleanup plan --connection <name> \
  --forward-plan <forward-plan-id> [--json]
pm etl transport github-issue-label cleanup run <plan-id> --connection <name> \
  --approval-token-stdin --confirm destructive [--json]
```

Only the connection/App owns repository, base URL, source issue, target issue,
label, action, record, and credential configuration. The raw one-time token is
accepted only as one bounded stdin line, never through argv, environment, token
files, JSON, persisted state, runtime records, logs, or error text. The
transport namespace and its leaf/cleanup namespace render contextual help before
`App.Open`; ordinary `pm etl run` rejects non-empty approval material unless the
resolved persisted route is exactly this closed transport. Runtime help,
`docs/cli/etl.md`, website ETL/CLI reference pages, JSON/manual goldens, and
focused parser tests are mandatory Green evidence.

## Required skills

`github-issue-first-delivery`, `golang-how-to`, `golang-cli`,
`golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`,
`golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`,
`golang-database`, `golang-testing`, `golang-documentation`, `golang-lint`,
`gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`,
`gsd-code-review`, and `no-mistakes`.

## Verification plan

```text
go test -timeout 20m ./internal/synctransport
go test -timeout 20m ./internal/app
go test -race -timeout 20m -run '^TestGitHubIssueLabel' ./internal/app
go test -timeout 20m -run '^TestGithubPullRequestsETLSupportsAllSyncModes$' ./internal/app
go test -timeout 20m ./internal/connectors/engine
go test -timeout 20m ./internal/connectors/commandrunner
go test -timeout 20m ./internal/cli
go test -timeout 20m -run '^TestPMBinaryExecutesGitHubWarehouseTransportLifecycle$' ./internal/cli
go vet ./internal/app ./internal/cli ./internal/synctransport ./internal/connectors/engine ./internal/connectors/commandrunner
go run ./cmd/connectorgen validate internal/connectors/defs/github
go run ./cmd/connectorgen surface-sync --check
scripts/verify-gsd-workflow origin/docs/4015-connector-release-certification
make tidy-check
make lint
make docs-check
make agent-contract-check
make connectorgen-validate
make connectorgen-surface-sync
make connector-boundary
make release-workflow-check
```

`go test ./...` and `make verify` remain split under the repository per-command
timeout rule. The exact selector/path names and demo argv will be pinned only
after the base guard passes and existing accepted symbols are inspected.
