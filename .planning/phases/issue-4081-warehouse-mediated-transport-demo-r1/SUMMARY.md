---
coverage:
  - id: D1
    description: Production construction installs fail-closed GitHub Transport and reopens a durable owner-scoped workset independently.
    verification:
      - kind: integration
        ref: internal/app/transport_construction_red_test.go
        status: pass
      - kind: unit
        ref: internal/synctransport/durable_reopen_red_test.go
        status: pass
    human_judgment: false
  - id: D2
    description: The transport accepts only a receipt/workset-bound, pre-run one-time approval and checks durable acknowledgement/read-back before CAS.
    verification:
      - kind: integration
        ref: internal/app/github_warehouse_transport_approval_test.go under -race
        status: pass
      - kind: unit
        ref: internal/app/transport_dispatch_test.go:TestRunETLRejectsDestinationApprovalOutsideClosedTransportRoute
        status: pass
    human_judgment: false
  - id: D3
    description: The closed CLI plan/preview/cleanup carrier accepts only a bounded stdin token and never exposes provider/approval internals.
    verification:
      - kind: unit
        ref: internal/cli/etl_transport_test.go
        status: pass
      - kind: unit
        ref: internal/cli/golden_transcript_test.go and TestGoldenDocsGenerateMatchesTrackedCLIManuals
        status: pass
    human_judgment: false
  - id: D4
    description: A fresh real pm binary completes the faithful GitHub source-to-cleanup lifecycle with zero residue.
    verification:
      - kind: integration
        ref: internal/cli/github_transport_binary_test.go:TestPMBinaryExecutesGitHubWarehouseTransportLifecycle
        status: pass
    human_judgment: false
---

# Summary — #4081 demonstrable warehouse-mediated Transport

## Delivered

- A single explicit App composition installs a non-nil connection-owned
  `WarehouseStage`, read-only accepted-evidence verifier, and exact GitHub
  source/destination registrations. Defaults remain fail closed.
- The durable handoff is a receipt rather than a source record slice: owner
  identity, generation, manifest/content hashes, checkpoint/tombstone state,
  and immutable Parquet reopening are verified across fresh `App.Open` calls.
- The closed GitHub destination uses the existing declarative engine read and
  typed `add_issue_labels` / `remove_issue_label` operations. Independent
  read-back and durable acknowledgement occur before checkpoint CAS.
- Added the accepted operator carrier: closed plan/preview/cleanup commands in
  `pm etl transport github-issue-label` plus only the required approval fields
  on ordinary `pm etl run`. The token is one bounded stdin line and never a
  persisted/printed CLI carrier value.
- Added a fresh-binary faithful server proof. It logs sanitized
  machine-readable evidence for the WAL/DuckDB/Parquet/manifest artifact,
  one reopened singleton, independent provider read-back, acknowledgement/CAS
  order, typed cleanup 204 and separately approved 404, replay rejection, and
  zero residue.

## Exact local binary evidence

`TestPMBinaryExecutesGitHubWarehouseTransportLifecycle` at `6220144db` built
the fresh binary with SHA-256
`146d030c27d4b45e19ebb318ea6ebe04b80482eb710e61b6ef5d7bc91f465c1f` and
size `148513250` bytes. The local faithful-server provider event sequence was:

```text
GET:source:100 → POST → GET:read-back:100 → DELETE:204 → DELETE:404
```

The normal approved GitHub App/disposable `Polymetrics-Cert` boundary was not
available in this isolated project, so live provider I/O was not attempted:
`LIVE_PROVIDER_BLOCKED_NO_APPROVED_GITHUB_APP_BOUNDARY`.

## Lifecycle status

The GSD plan, RED checkpoints, Green commits, UAT, and manual code review are
complete. No-mistakes, draft PR creation, forge CI, and automated-review
disposition remain pending; this summary does not claim them complete.

## Deferred scope

- PostgreSQL source/destination legs and remaining family compositions.
- Schedules, flows, bounded database query, CDC, cross-process auth/rate
  coordination, and restart orchestration beyond named failure injection.
- Exhaustive seven-mode certification, GitHub `full_overwrite`, broad
  certification-generator truth, and final release promotion.
- TOCTOU locking, strict hash-syntax validation, and `O_EXCL` collision
  reservation in the larger production Transport MVP.
