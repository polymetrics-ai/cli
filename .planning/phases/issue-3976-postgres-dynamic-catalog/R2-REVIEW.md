# R2 code review — resumable PostgreSQL source reads

Date: 2026-08-16

## Manual inline fallback

The historical issue phase is already executed and the canonical delivery
contract permits one active inline worker only. The generated `verify-work`
and `code-review` prompts were resolved through `scripts/gsd sources`; this
record is the required manual fallback, not a lifecycle exemption.

## Scope reviewed

- Base: `integration/4015-mvp-flat-r1` (`ef3c71caf` at dispatch).
- Head reviewed: `77bf5e49f`.
- One connector: native PostgreSQL source transport; generated docs and
  website catalog are its parity output.

## Findings and disposition

No actionable findings.

- Production reach is explicit: `App.runTransportETL` carries a stream's
  cursor into `synctransport`, the PostgreSQL definition factory selects
  `PollingTransportSource`, and the compiled `pm` test reaches the shared
  `engine.PollingSourceExecutor` and resumes its persisted tuple.
- The old `postgres_bounded_snapshot` source factory and private source page
  loop are removed. The retained snapshot plan is used only by the sealed
  logical-replication bootstrap handover, which is outside this polling lane.
- Missing/unknown/nullable stream cursors and incompatible checkpoints fail
  closed. The invalid protocol checkpoint is rejected before runtime config;
  the stale schema checkpoint is rejected before a source page; neither emits
  a restarted read.
- The shared zero-row candidate preserves the prior tuple (or reports an
  unobserved position on the first empty scan), so resumption neither invents
  nor advances a cursor.
- SQL remains catalog-bound and parameterized. There is no generic SQL,
  HTTP-write, shell, credential, or new dependency surface.

## Verification reviewed

- Focused PostgreSQL, polling-engine, App transport, synctransport, and
  compiled-binary tests passed; PostgreSQL race passed.
- `go vet` on changed packages, `go build ./cmd/pm`, docs validation,
  `tidy-check`, lint, connector validation/surface sync/boundary,
  agent-contract, release-workflow, and smoke gates passed.
- `databaseintegration` PostgreSQL and CLI packages compiled with `-run '^$'`.
  The live database run is pending: this task did not start or restart the
  shared container runtime.

## Automated review route

The direct PR targets the non-default integration branch. Its primary route is
`claude_auto` on PR creation; until GitHub posts a review record, status is
`pending`. Parent-PR/human review remains the allowed fallback for the stacked
range. No Copilot request is made without a Claude blocker.
