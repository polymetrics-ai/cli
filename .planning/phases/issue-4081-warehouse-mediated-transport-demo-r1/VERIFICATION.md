# #4081 — verification checklist

## Phase and topology

- [x] Live parent/sub-issue/PR reconciliation completed through `gh-axi`.
- [x] No existing open or closed issue/PR owns the production Transport
  construction walking slice; #4081 is the single direct #4015 child.
- [x] #4015 visibly links #4081 and records its prerequisite/base guard.
- [x] GSD adapter doctor, all five command sources, and canonical contract check passed.
- [x] Required planning skills are loaded and recorded.
- [x] Required combined branch is authoritative head
  `e7d2b2963fc1dd164f63b31fccb8a3bab8084bec`; the squashed #4019 content is
  proven by exact #4077 Transport blob identity with `aaf288d...`, not ancestry.
- [x] Final `feat/4081-warehouse-mediated-transport-demo` branch exists from
  that remote combined head only.

## Before completion

- [x] Planning artifacts committed on final child branch before production edits
  (`1adc808e4`).
- [x] RED commit precedes every implementation commit and fails for the expected
  construction/stage/reopen/order failures.
- [x] GREEN tests prove independent durable reopen, exact closed GitHub legs,
  receipt-before-CAS, safe replay, and #4079 isolation.
- [ ] Fresh exact **committed-head** `pm` binary digest and size are recorded; bounded local
  faithful-server run asserts returned record counts—not only exit code.
- [x] Stage/Parquet/DuckDB identity, independent provider read-back, typed
  inverse cleanup, second cleanup, and zero residue are demonstrated.
- [ ] Live provider result is recorded only if the approved PM credential boundary
  resolves normally; otherwise a precise safe blocker is recorded.
- [ ] Targeted tests, split local gates, GSD verify-work/review, no-mistakes,
  exact-head CI, and automated-review disposition are terminal green/allowed.
- [ ] Draft PR is opened only to `docs/4015-connector-release-certification`,
  with `Refs #4081`, `Refs #4015`, `Refs #3862`, and `Refs #4077`; never merged.

## Resolved base-admission evidence

```text
remote combined head: e7d2b2963fc1dd164f63b31fccb8a3bab8084bec
accepted #4077 source head: aaf288d069adc1b67a09500afcca4be4a6d1bab3
types.go blob (combined/source): 7b5eafd34c78bea690dcb23eee716ea840ea813e
orchestrator.go blob (combined/source): 442ea03826f6385da3d6a04fd65af3936dc35188
registry.go blob (combined/source): c9ac2eeb6045546390c04f30769601028f8e5571
transport_test.go blob (combined/source): b191b8e2905afa397e9c184f18ee26952dda00ac
combined tree contains .planning/phases/issue-4077-transport-record-isolation-r1/
```

The non-ancestry result is expected for the #4019 squash merge. Content identity
is the required proof. No correction-loop budget is consumed.

## Current Green verification record

The carrier is tested through separate subprocess invocations of a freshly
built `pm`, not a test-only direct apply shortcut. The faithful server receives
one declared bounded GitHub issues page (the connector's declared default wire
page size is `100`; the transport emits/stages exactly one configured record),
then typed label POST, independent engine GET read-back, typed DELETE `204`,
and separately planned typed DELETE `404`. It rejects a replayed cleanup grant
before another DELETE and finishes with no label residue.

The exact local checks that have passed before the Green commit are listed in
`TDD-LEDGER.md`. They include full `internal/app`, full `internal/cli`, full
`internal/synctransport`, engine, commandrunner, the B2 approval suite under
race, the five-mode GitHub compatibility regression, focused binary lifecycle,
docs/manual goldens, vet, GitHub bundle validation, surface-sync, and
`git diff --check`.

### Fresh exact-binary local proof

```text
go test -count=1 -timeout 20m \
  -run '^TestPMBinaryExecutesGitHubWarehouseTransportLifecycle$' -v ./internal/cli

binary_sha256: 146d030c27d4b45e19ebb318ea6ebe04b80482eb710e61b6ef5d7bc91f465c1f
binary_size_bytes: 148513250
artifact_kinds: wal_jsonl, duckdb_parquet, manifest
source_records: 1
reopened_records: 1
provider_events: GET:source:100, POST, GET:read-back:100, DELETE:204, DELETE:404
independent_read_back: true
checkpoint_after_acknowledged: true
cleanup_statuses: 204, 404
replay_rejected: true
zero_residue: true
result: PASS
```

The server is local and faithful to the declared public GitHub contract; it is
not called a live-provider certification. The only on-wire mutation is the
typed label add/remove against a run-owned fixture, and all raw approval tokens
are asserted absent from argv, environment, captured command output, and every
persisted project artifact.

## Remaining honest gates

- Rebuild and record a fresh binary SHA-256/size at the implementation commit,
  then repeat `TestPMBinaryExecutesGitHubWarehouseTransportLifecycle`.
- Run inline/manual `verify-work` and deep code review, record every finding and
  its evidence, then run the no-mistakes pipeline without `--yes`.
- No approved GitHub App credential/disposable `Polymetrics-Cert` boundary has
  been supplied to this isolated local harness. Do not inspect credentials or
  issue provider I/O; record `LIVE_PROVIDER_BLOCKED_NO_APPROVED_GITHUB_APP_BOUNDARY`
  after local proof is complete.
- Open only a draft child PR to
  `docs/4015-connector-release-certification`, wait for terminal allowed/green
  checks and automated-review disposition, and never merge it or touch #4016.
