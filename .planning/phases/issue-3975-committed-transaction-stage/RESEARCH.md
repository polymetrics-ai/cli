# Research — Issue #3975: committed-transaction staging and durable receipts

## Manual inline research record

The normal `plan-phase` runtime cannot resolve issue #3975 as a numbered
roadmap phase and the canonical delivery contract forbids planner/researcher
role spawning. The active worker performed this bounded research inline before
planning; no dependency, credential, database, or production mutation was
introduced.

## Findings that constrain the implementation

1. `data/cli-cdc-large-transaction-strategy-r1/report.md` selects a bounded,
   crash-recoverable per-transaction journal plus a hard fail-closed quota. It
   explicitly rejects source acknowledgement based solely on a local spool.
2. `data/cli-cdc-bidirectional-changefeed-design-r1/report.md` separates a
   private stage from the durable extraction receipt. It requires StreamAbort
   discard, whole-transaction materialization at commit, and receipt-before-
   source acknowledgement.
3. `data/cli-database-connector-framework-design-r1/report.md` and
   `/Users/karthiksivadas/karthik-agent-workspace/data/learnings.md` require
   every acknowledgement to be derived from the durable fact it asserts;
   attempted write, stage persistence, and generic success are insufficient.
4. `internal/synccontract/commit.go` already owns the unforgeable downstream
   acknowledgement and checkpoint ordering. The new stage must adapt a
   persisted receipt to that contract rather than creating a second checkpoint
   or acknowledgement schema.
5. `internal/synccontract/state.go`, `tombstone.go`, and `recovery.go` already
   own opaque source tokens, checkpoint envelopes, tombstones/history, and
   typed recovery outcomes. The stage stores opaque chunk bytes and must not
   parse, normalize, invent, or flatten these values.
6. `internal/connectors/database/resources.go` establishes the nearest
   package-local pattern: finite validated limits and context-aware boundaries.
   `registry.go` confirms that #3974's descriptor/admission truth must remain
   closed; the stage is not an admission or capability promotion.
7. `internal/connectors/native/postgres/cdc.go` is intentionally fail-closed
   until this stage exists. It is an integration consumer only and remains
   untouched by this issue.

## Chosen implementation direction

- Use a transaction-specific private filesystem stage with atomic, fsynced
  state transitions and a root-local recovery scan.
- Stream chunks through a fixed-size buffer into framed immutable files; keep
  only bounded counters/digests in memory.
- Seal a transaction before delivery. Publish chunks in source order through a
  narrow injected receipt port only at commit.
- Persist and validate an immutable receipt before returning any object capable
  of creating a `synccontract.DownstreamAcknowledgement`.
- On restart, delete incomplete/orphan temporary stages, retain sealed but
  receipt-less work for explicit resume, and remove sealed residue only after a
  valid durable receipt exists.
- Use fault-injecting filesystem and receipt fakes to prove each write/sync/
  rename/receipt/cancellation crash boundary. No external package is needed.

## Risks and mitigations

| Risk | Mitigation |
|---|---|
| Premature acknowledgement loses source data | Receipt-derived eligibility has an unexported validity marker and is returned only after receipt file + parent directory durability. |
| Transaction ID path traversal | Derive stage paths from a validated hash; retain provider ID only as serialized opaque metadata. |
| Unbounded memory/disk | Fixed streaming buffer plus byte/record/time limits and recovered-root accounting. |
| Crash leaves misleading state | Temporary writes are removed/recovered; sealed no-receipt work remains non-eligible; receipt presence is validated before cleanup. |
| A future CDC wire-up promotes a capability early | Keep the PostgreSQL source untouched and include a scope scan in final verification. |

## Package and dependency decision

The implementation remains in `internal/connectors/database/` as the explicit
source-agnostic #3975 foundation. It adds no external module, does not open a
database, does not expose generic SQL/HTTP/shell writes, and does not require
Podman or a credentialed integration environment.
