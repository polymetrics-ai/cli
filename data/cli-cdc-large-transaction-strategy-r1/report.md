# PostgreSQL CDC: large-transaction strategy

**Research report — 2026-08-10**  
**Decision for `cli-postgres-cdc-logical-replication-r1`:** make PostgreSQL 14+
`pgoutput` transaction streaming the baseline, pair it with a **bounded durable
per-transaction staging journal (Option A) and a hard fail-closed quota (Option
B)**, and do **not** add cursor/timestamp reconciliation (Option C) as an
automatic CDC recovery path.

> A staging journal is not downstream durability. It may make an uncommitted
> source transaction restartable locally, but it must never cause a PostgreSQL
> LSN acknowledgement. The acknowledgement boundary remains the durable
> downstream receipt/checkpoint for the *complete committed source
> transaction*.

This is a measurement/research result, not a production change. It is an
internal report, so the open-source systems named below are permitted here;
they must not be copied into shipping code, comments, docs, PR bodies, or
commit messages.

## Executive answer

PostgreSQL 14 already has the protocol mechanism that removes the specific
"all row changes arrive at commit" receiver burst: logical-decoding streaming.
With `pgoutput` protocol version 2 and `streaming=on`, PostgreSQL streams a
large **in-progress** transaction once decoding crosses
`logical_decoding_work_mem`; the protocol explicitly supplies stream start,
stop, commit, and abort messages. It is not a licence to publish or
acknowledge those changes early. A receiver must retain them behind a
transaction boundary until `StreamCommit`, and discard them on `StreamAbort`.
[PostgreSQL's streaming-decoding documentation](https://www.postgresql.org/docs/current/logicaldecoding-streaming.html)
and its [logical-replication protocol](https://www.postgresql.org/docs/current/protocol-logical-replication.html)
are unambiguous on that distinction.

The phase-one decision is therefore:

1. Admit `change_capture` only on PostgreSQL 14+ for this path; start
   replication with `proto_version '2'` and `streaming 'on'`.
2. Stream each source transaction into a bounded, crash-recoverable **local
   transaction stage**. It exists to bound process memory and to survive a
   receiver restart, not to establish source progress.
3. On `StreamCommit`, hand the whole staged transaction to a downstream
   transaction/receipt port. Send status that advances the source position
   only after that port has made the complete transaction durably accepted and
   recorded its receipt/checkpoint. On `StreamAbort`, discard the stage
   without publishing its records.
4. Give the stage a hard byte/record quota. Exceeding it is the named,
   fail-closed `TransactionStageLimitExceeded` outcome: no source progress is
   acknowledged, no cursor fallback is attempted, and the run clearly enters
   retry-or-rebootstrap handling.
5. Ship slot-health observability and an explicit connector-owned-slot
   teardown/rebootstrap procedure with the feature. A quota failure prevents
   our OOM; it does **not** make an inactive slot harmless.

Option C is rejected for phase one and must never silently stand in for
`change_capture`. A timestamp/cursor scan cannot see hard deletes in the
skipped interval, cannot reproduce intermediate changes or transaction
boundaries, and has no generic, trustworthy cutover point to a replication
slot. The field's successful incremental-snapshot algorithms keep consuming
the log and use watermarks/collision handling; they are not "stop CDC, scan a
timestamp, and resume" fallbacks.

## The source-side fact that dominates the decision

A logical replication slot retains WAL on the source until the consumer's
confirmed progress permits recycling. Slots persist across crashes and retain
WAL/catalog state; PostgreSQL explicitly warns that an unused slot can make
the source retain enough resources to cause a shutdown. See
[Logical Decoding — Replication Slots](https://www.postgresql.org/docs/current/logicaldecoding-explanation.html).

This has two consequences that are easy to blur together:

- PostgreSQL 14 streaming reduces **decoder/receiver burst memory**. The
  server can spill decoded changes after `logical_decoding_work_mem`, and it
  can send streamed chunks before source commit. It does **not** make an
  uncommitted transaction eligible for WAL recycling.
- Any consumer that has not durably completed the transaction must leave its
  acknowledged position at the preceding durable commit. During a very long
  source transaction, that necessarily holds WAL from before the transaction.
  A client cannot safely engineer that fact away; it can only avoid adding an
  avoidable receiver-side backlog or make the source's retention risk visible
  and bounded where the platform supports it.

AWS documents the same operational behaviour for PostgreSQL: a large
transaction does not advance its replication progress until it completes, and
the server spills decoded transaction data to disk after the configured memory
threshold. [AWS DMS PostgreSQL latency guidance](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Troubleshooting_Latency_Source_PostgreSQL.html)
and [its PostgreSQL source documentation](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Source.PostgreSQL.html)
recommend breaking up long/large source work rather than treating it as an
invisible consumer problem.

## What PostgreSQL already solves — and what it does not

### Already supplied by PostgreSQL 14+

| PostgreSQL facility | Evidence | What the CDC connector should use |
| --- | --- | --- |
| In-progress logical transaction streaming | `streaming=on` with `pgoutput` protocol v2 sends `StreamStart`, `StreamStop`, then `StreamCommit` or `StreamAbort`; v2 is available from PostgreSQL 14. [Protocol reference](https://www.postgresql.org/docs/current/protocol-logical-replication.html); [PostgreSQL 14 release notes](https://www.postgresql.org/docs/14/release-14.html). | Request v2 and implement its state machine. Do not invent a second WAL/paging protocol. |
| Bounded server decode memory / server spill | `logical_decoding_work_mem` is the upper memory threshold before logical decoding spills transaction data to disk. [Resource configuration](https://www.postgresql.org/docs/current/runtime-config-resource.html). | Preflight/observe the effective value. It bounds PostgreSQL's decode memory; it does not by itself bound PM's staging or establish downstream durability. |
| Ordered transaction stream | The streaming API preserves commit order while it can transmit chunks of a large transaction before its commit. [Streaming-decoding documentation](https://www.postgresql.org/docs/current/logicaldecoding-streaming.html). | Preserve source transaction identity and only make data visible at source commit. |
| Prepared-transaction support is separate | Protocol v3 adds two-phase support (PostgreSQL 15); v4 covers parallel streamed transactions (PostgreSQL 16). [Protocol reference](https://www.postgresql.org/docs/current/protocol-logical-replication.html). | Do not enable two-phase merely as a large-transaction workaround. Add it only if prepared transactions become an explicit supported source feature. |

The pinned dependency on the parked implementation branch already contains
the protocol message types and `ParseV2` support:
[`pglogrepl` `messageV2.go` at the pinned commit](https://github.com/jackc/pglogrepl/blob/e37c41485510/messageV2.go).
That avoids a new parser dependency; it does **not** avoid the work of giving
PM a committed-transaction staging/delivery boundary.

### Not solved by PostgreSQL

PostgreSQL cannot know when PM's destination has made a transaction durable.
Nor can it safely turn an in-progress stream into committed public change
events. `StreamAbort` is the proof: any records exposed before its matching
`StreamCommit` can be records that never committed at the source. PostgreSQL
also cannot make an unbounded source transaction occupy finite disk forever.
The operator still needs a source-side retention policy, monitoring, and a
documented response for a stalled connector.

## Field evidence: what mature systems actually document

The table separates documented fact from inference. An empty cell is not
evidence that a product lacks a feature; it means the public material read did
not disclose the relevant transaction/acknowledgement detail.

| System | Direct evidence read | What it establishes for this decision |
| --- | --- | --- |
| **PostgreSQL** | Logical decoding normally calls output callbacks at commit; PG14's streaming callbacks expose large in-progress transactions after `logical_decoding_work_mem`. [Streaming decoding](https://www.postgresql.org/docs/current/logicaldecoding-streaming.html). Slots persist and retain WAL across consumer failure. [Slots documentation](https://www.postgresql.org/docs/current/logicaldecoding-explanation.html). | The core burst is a protocol/configuration problem first. V2 streaming is the correct baseline, but it cannot advance a slot past an uncommitted or not-yet-durable transaction. |
| **Debezium and the Netflix DBLog algorithm** | Debezium's incremental snapshot reads primary-key chunks while **continuous log capture stays active**, using low/high watermarks and collision handling. [Debezium PostgreSQL connector](https://debezium.io/documentation/reference/stable/connectors/postgresql.html); [Debezium's DBLog-derived incremental-snapshot explanation](https://debezium.io/blog/2021/10/07/incremental-snapshots/); [DBLog paper](https://arxiv.org/abs/2010.12597). | This is not Option C. It is a second, coordinated capture path that retains log continuity during a snapshot. It needs a key, watermarks, and collision storage; it does not license skipping a transaction and trusting a timestamp scan. |
| **Airbyte** | Its PostgreSQL CDC tutorial says it uses Debezium/`pgoutput`, requires primary keys for its CDC incremental path, and uses a delete marker (`_ab_cdc_deleted_at`) to convey deletions. [Tutorial](https://airbyte.com/tutorials/incremental-change-data-capture-cdc-replication). | An architectural lead, not a PM dependency: a log-based CDC path and explicit delete representation are used instead of a generic timestamp rescue. The material read does not document an independent oversized-transaction cursor fallback. |
| **Fivetran PostgreSQL connector** | Its logical-replication path is recommended for large databases and tracks deletes; its separate query-based method uses `xmin` for inserted/updated rows and can compare a complete `ctid` key set to find deletes, with documented blind spots for fast changes/delete-reinsert cycles. [Connector documentation](https://fivetran.com/docs/connectors/databases/postgresql). Its timeout guidance names very large transactions/too many slot records as causes and retries later rather than claiming a transparent reconciliation. [Timeout guidance](https://fivetran.com/docs/connectors/databases/postgresql/troubleshooting/logical-replication-connector-timed-out). | The field closes hard-delete gaps with an explicit full identity/key reconciliation, not a cursor alone—and documents remaining limitations. That is a different, current-state sync contract, not faithful event capture. |
| **Fivetran HVR / former HVR** | `TxSplitLimit` can split a source transaction into smaller target transactions to control resources. [Integrate action reference](https://fivetran.com/docs/hvr6/action-reference/integrate). | Some systems deliberately trade target transaction atomicity for resource control. That is evidence of an explicit semantic choice, not a way to preserve PM `change_capture` semantics invisibly. |
| **Singer-family taps** | Singer integration guidance defines incremental replication around a declared replication key and says it does not capture deletes; the Singer hub recommends log-based capture, soft deletes, or a versioned/full snapshot pattern when hard deletes matter. [Meltano integration guide](https://docs.meltano.com/guide/integration); [Singer documentation](https://hub.meltano.com/singer/docs/). | A cursor is not a general CDC recovery mechanism. The ecosystem requires per-table configuration and a declared delete solution. |
| **AWS DMS** | DMS documents logical-decoding disk spill and that a source transaction does not let replication progress move until completion; it advises source-side mitigation for long/large work. [Latency guidance](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Troubleshooting_Latency_Source_PostgreSQL.html); [PostgreSQL source notes](https://docs.aws.amazon.com/dms/latest/userguide/CHAP_Source.PostgreSQL.html). | Mature managed CDC does not hide the source transaction behind an ordinary cursor. It relies on PostgreSQL spill plus source operational controls. |
| **Qlik / Attunity Replicate** | Qlik documents keeping transaction data in memory until source and target commit, then offloading large/aged transaction data to disk under configurable limits. It also documents PostgreSQL WAL heartbeats to help a slot advance during idle periods. [Apply-change tuning](https://help.qlik.com/en-US/enterprise-manager/May2026/Content/Global_Common/Content/SharedEMReplicate/Customize%20Tasks/tasks_applychangtunestab.htm); [PostgreSQL source options](https://help.qlik.com/en-US/replicate/May2026/Content/Replicate/Main/PostgreSQL/advanced_prop_postgresql_source.htm). | This is the closest documented model for A: disk staging bounds RAM and the receiver keeps consuming. It does **not** establish that PM may acknowledge PostgreSQL when only its own spool is durable. |
| **Striim** | Its public PostgreSQL CDC announcement claims support for very large transactions. [GA announcement](https://www.striim.com/press/striim-announces-general-availability-of-version-4-0-of-its-streaming-platform/). I did not find an authoritative public PostgreSQL design describing its spool, LSN acknowledgement, or slot recovery boundary. | Treat it as a lead for further vendor inquiry, not evidence for a PM correctness algorithm. No durability or slot-safety conclusion is drawn from this marketing-level claim. |
| **Google Datastream** | Google warns that a large transaction/high volume can stall other tables sharing a slot, recommends isolating high-churn tables, and warns that a paused/deleted stream which leaves its slot can retain WAL indefinitely. It also discusses source retention controls and their managed-service limitations. [PostgreSQL source guide](https://docs.cloud.google.com/datastream/docs/sources-postgresql); [WAL retention guide](https://docs.cloud.google.com/datastream/docs/work-with-postgresql-database-wal-log-files). | The production response is slot isolation, monitoring, source limits, and lifecycle control—not an opaque client-side cursor fallback. |

### What the field evidence says about Option C

No source found supplied a credible percentage for “the fraction of real
tables with a usable cursor.” It would be a false finding to invent one. The
systems that support cursor replication make it a **per-stream declared
property**: a stable identity key plus a replication key whose value/tie-breaker
is valid for every mutation being captured. PM's current configuration check
also only verifies that a cursor field was supplied; it cannot prove the
column is updated for every change or that its ordering is safe.

For a table to be eligible for any future current-state reconciliation, it
would need all of the following evidence, not merely a timestamp-shaped
column:

- a stable primary/unique key;
- a monotonic, tie-broken cursor that changes for every insert and update in
  the intended scope;
- a source snapshot/log barrier tied to the cutover, not a wall-clock guess;
- a declared hard-delete contract: source soft deletes, a complete key-set
  anti-join/diff, or continued logical-event capture; and
- an explicit final-state, not event-history, product contract.

Hard deletes are decisive. A cursor query cannot return a row that no longer
exists. The Fivetran query path's full identity comparison and the Singer
guidance's soft-delete/log/full-snapshot alternatives are the documented
answers; both make the extra work and the semantic limitation visible.

## Option scorecard — replication-slot safety first

These are qualitative operational assessments from the cited mechanics, not
measured vendor reliability scores. “Slot position” means the last source
position that PM may honestly acknowledge as downstream-durable.

| Candidate | Slot during an oversized transaction | Failure path | Recovery path | Truthfulness and slot-first verdict |
| --- | --- | --- | --- | --- |
| **A. Bounded durable spool/journal** | With PG14 streaming, PM can keep reading and stage chunks without growing RAM. The acknowledged slot position remains at the prior durable commit until the source `StreamCommit` is fully accepted downstream. WAL for the in-flight transaction remains retained—as it must. | If downstream stops or the finite journal quota/disk fills, PM must stop before an unsafe write and must not advance the source position. The slot then retains WAL. A finite spool does **not** handle an unbounded transaction; it converts OOM into a named capacity failure. | Restart/replay from the old durable LSN. A journal can speed crash recovery but is never proof that the destination accepted the transaction. If recovery cannot succeed, explicit teardown plus rebootstrap frees the source only by intentionally giving up CDC continuity. | **Best normal-running option when coupled to V2, but only if it never ACKs on spool durability.** It minimizes receiver backlog while preserving the source's real durability boundary. It still needs quota, slot monitoring, and an operator-visible failure path. |
| **B. Explicit fail-closed transaction limit** | Before the limit, the slot behaves normally. On the first transaction over the cap, source progress remains at the previous durable commit and WAL for the poison transaction is retained. | No PM OOM and no silent loss. But a deterministic cap causes every retry to encounter the same transaction and the slot remains pinned. This is a source-disk incident until something changes. | Increase/remediate capacity and replay from the old LSN, **or** deliberately tear down the PM-owned slot and rebootstrap. The latter creates a known history gap and cannot be reported as seamless `change_capture` resume. | **Necessary containment, poor primary strategy.** Use as A's hard quota/error rather than as the only answer. It is safer than early acknowledgement, but without explicit lifecycle controls it is exactly the stalled-slot failure the task warns about. |
| **C. Stop/skip to cursor/timestamp reconciliation, then resume CDC** | If PM stops consuming while it scans, the old slot cannot advance and retains WAL. If it keeps consuming but cannot durably account for the skipped transaction, it still cannot honestly advance. | Dropping/advancing the slot to relieve retention makes an unreconstructable gap. A timestamp scan cannot see hard deletes, and it loses intermediate events/transaction structure. A scan without a slot-bound snapshot/watermark has a race at cutover. | Keeping the old slot requires an additional durable log/collision path while snapshotting—i.e. an incremental-snapshot architecture, not C as proposed. Dropping it requires explicit full rebootstrap and never restores event history. | **Reject as automatic CDC recovery.** It has the worst slot story and silently changes semantics. It could be a future, explicitly named current-state reconciliation feature for eligible tables only. |

### The acknowledgement boundary for Option A

The ordering below is the non-negotiable durability rule. The local journal is
deliberately outside the acknowledgement boundary.

```text
PostgreSQL StreamStart / DML chunks
        │
        ▼
bounded local transaction stage (fsync makes this local stage recoverable)
        │                         │
        │                         └─ never advances source slot
        ▼
PostgreSQL StreamCommit
        │
        ▼
downstream applies the complete committed source transaction
        │
        ▼
durable downstream receipt + PM checkpoint for TransactionEndLSN
        │
        ▼
send PostgreSQL standby-status acknowledgement for that LSN
```

If the process crashes before the receipt, it must replay from the earlier
durable LSN. If the destination might have committed before the receipt was
persisted, the replay is an at-least-once delivery outcome, handled by the
same receipt/deduplication seam adopted for append delivery—not by asserting
that a spool had already made the transaction delivered.

This follows the local fleet learning exactly: PostgreSQL once acknowledged
before the downstream write was durable, advanced the slot, and made data
permanently unrecoverable. The spool must not recreate that defect under a
new name.

## Direct answers to the five requested questions

### 1. Phase-one recommendation

**Recommend A + B, enabled by PostgreSQL's native v2 streaming, and reject C.**

Option A is the practical way to consume an in-progress large transaction
without unbounded PM memory; PostgreSQL v2 is what makes it possible to get
the chunks before source commit. Option B is still essential as the finite
journal's hard, named failure mode. Treat PostgreSQL 14 as the minimum version
for `change_capture` rather than silently falling back to v1 for a source
whose large-transaction behaviour PM cannot safely bound.

The phase-one guarantee is not “any source transaction is harmless.” It is:
“PM will preserve committed source-event ordering within its published
at-least-once delivery contract, will not acknowledge an LSN before the
downstream durable receipt, and will fail explicitly before exhausting its
transaction-stage capacity.”

If the CDC implementation cannot introduce the committed-transaction delivery
port described below, it should remain non-executable. A smaller change that
feeds V2 chunks into the present event callback would expose source-aborted
records and is not an acceptable shortcut.

### 2. Exact replication-slot behaviour in each failure and recovery path

**Common rule.** The slot's acknowledged/confirmed progress cannot move
through the current source transaction until the complete transaction has
committed and its downstream receipt/checkpoint is durable. PostgreSQL may
retain WAL longer than the just-acknowledged point because of its own restart
and active-transaction constraints; an acknowledgement makes WAL eligible for
recycling, not synchronously deleted.

**A — normal path.** V2 lets PM continue reading chunks into the journal, so
the receiver need not wait for one final in-memory burst. The slot remains
behind the transaction while it is in progress. After `StreamCommit` and the
downstream durable receipt, PM sends source progress for the transaction end
LSN; the slot can then advance and old WAL can eventually recycle.

**A — journal/target failure.** No source acknowledgement is sent for the
transaction. The position stays at the prior durable commit and the slot
retains WAL. A restart retries/replays from that position. If the customer
elects to destroy the connector-owned slot, PostgreSQL can reclaim the retained
WAL, but PM must mark the stream `rebootstrap_required`; it has deliberately
given up the skipped CDC history.

**B — cap failure.** The same non-acknowledgement is correct. It prevents
silent data loss and PM OOM but leaves a deterministic “poison transaction”
at the front of the slot. Recovery is either increase/fix the capacity and
replay, or the explicit destructive teardown/rebootstrap path. Retrying with
the same cap does not solve retention.

**C — scan fallback.** Pausing CDC retains WAL. Continuing to consume without
durably representing the omitted transaction cannot advance the slot. Dropping
the slot lets PostgreSQL recycle WAL but makes the gap permanent; creating a
new slot after a scan does not reconstruct it. A correct concurrent snapshot
would need to keep log continuity and manage watermark/collision state, which
is a separate incremental-snapshot feature rather than cursor fallback.

**Operational guard.** Expose/alert on slot name, `confirmed_flush_lsn`,
`restart_lsn`, retained WAL, state age, journal usage, and the last downstream
receipt. Where a source supports `max_slot_wal_keep_size`, make its configured
value visible as a source-enforced last resort; an invalidated slot must route
to rebootstrap. Do not claim this protects every managed PostgreSQL offering:
Google documents services where the relevant source control is unavailable.
The existing connector-owned `TeardownCDC` lifecycle operation is a useful
explicit release primitive, but it is not a substitute for monitoring or a
safe acknowledgement boundary.

### 3. Advertised sync-mode truthfulness

Under the recommended phase one, **there is no successful automatic fallback**
from CDC to cursor reconciliation. Therefore no existing mode's meaning is
silently degraded: a stage-limit failure is an error/rebootstrap requirement,
not a run that claims success with a different method.

| Mode in `internal/synccontract/mode.go` | With v2 + A+B phase one | If a future C-style current-state feature is added |
| --- | --- | --- |
| `change_capture` | Truthful only when streamed records stay private until source commit and the end LSN is acknowledged after the downstream receipt. On quota failure, stop and report failure; do not skip/reconcile. | **Never use C as an automatic fallback.** It would lose intermediate changes, transaction boundaries, and possibly deletes. |
| `full_overwrite` | Unchanged. It is an honest *explicit rebootstrap* route after CDC continuity is intentionally abandoned. | A full snapshot can establish present state but cannot restore the historical changefeed gap. |
| `full_append`, `incremental_append`, `incremental_dedupe_history` | Unchanged and not a remedy for CDC stage failure. They preserve appended/history semantics that a final-state scan cannot recreate. | Not compatible with a lossy current-state reconciliation. |
| `incremental_upsert`, `incremental_dedupe` | Unchanged regular incremental modes; the CDC implementation must not substitute them automatically. | Potentially compatible only as a separately declared **final-state** contract, after table-by-table proof of stable key, complete cursor, slot-bound cutover, and a hard-delete repair mechanism. This is not a universal table capability. |

The local mapping confirms why this distinction matters: `change_capture`
maps to a change-capture source and upsert destination, while the incremental
modes require a cursor and/or primary key. Destination upsert does not turn a
lost change history into a valid capture history.

### 4. Protocol work PM should not reimplement

Do not build a home-grown transaction-paging/cursor protocol or try to make
PostgreSQL emit a second version of its WAL stream. PostgreSQL 14+ already
does server-side decoded-memory spill and protocol-level in-progress stream
framing. PM should request it, parse it using the already pinned `pglogrepl`
v2 support, and preserve its `StreamCommit`/`StreamAbort` semantics. Two-phase
v3 is not a replacement for v2 streaming.

PM still owns the application boundary: durable staging, downstream receipt,
checkpoint, retry, quota, and slot-health lifecycle. Those are product
semantics and must not be inferred from a successfully received replication
message.

### 5. Smallest honest phase one, and the forward seam

The smallest implementation that actually fixes the burst—rather than merely
documenting it—is:

1. **Preflight:** require PostgreSQL 14+, logical WAL/publication/slot
   prerequisites, and record the effective `logical_decoding_work_mem` plus
   source slot retention posture.
2. **Protocol:** request pgoutput v2 with `streaming=on`; dispatch v2 messages
   by source transaction ID, including abort. Do not route a streamed DML
   message directly to the existing public event callback.
3. **Stage:** introduce a bounded, durable `LogicalTransactionStage` behind a
   narrow port: open/append/abort/seal/recover. Its identity is the source
   transaction and its capacity accounting is explicit. It has no method that
   advances a source checkpoint.
4. **Delivery:** introduce a `CommittedLogicalTransactionSink` (or equivalent)
   that accepts only a sealed committed transaction and returns a durable
   downstream receipt. Persist the checkpoint from that receipt, then send
   source status for `TransactionEndLSN`.
5. **Containment:** make stage exhaustion a typed error; preserve the old
   checkpoint; provide retry, status/alerting, and explicit
   teardown/rebootstrap instructions. Test a large streamed commit, a streamed
   abort, crash/replay, quota exhaustion, and the invariant that source status
   cannot advance before the receipt.

This is consistent with the captain's append-delivery ruling: publish the
honest at-least-once guarantee now and introduce a receipt seam now, rather
than claiming exactly-once or hidden recovery. A later, better state
reconciliation/incremental-snapshot implementation can plug in behind an
explicit recovery strategy and the same committed-transaction receipt port.
It must be a separately admitted capability/contract; it must not change what
the already-published `change_capture` mode means.

## Repository evidence and implementation implications

### Commands and local evidence read

The worktree was not CodeGraph-indexed, so code references below were obtained
with page-wise `rg`, `nl`, and `git show` reads. No source code was changed.

| Command / source read | Result used in this report |
| --- | --- |
| `nl -ba /Users/karthiksivadas/karthik-agent-workspace/data/learnings.md | sed -n '1,70p'` | [`data/learnings.md:6-38`](/Users/karthiksivadas/karthik-agent-workspace/data/learnings.md:6) records the prior PostgreSQL defect: the slot was acknowledged before the downstream write was durable, making data unrecoverable. |
| `nl -ba .../DECISION-append-delivery-guarantee.md | sed -n '1,220p'` | [`DECISION-append-delivery-guarantee.md:4-13`](/Users/karthiksivadas/karthik-agent-workspace/data/cli-database-connector-framework-design-r1/DECISION-append-delivery-guarantee.md:4) says phase one must state at-least-once honestly and create a receipt seam for later improvement. |
| `nl -ba internal/synccontract/mode.go` and `nl -ba internal/app/sync_modes.go` | [`mode.go:14`](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/synccontract/mode.go:14) lists the closed mode vocabulary; [`sync_modes.go:64`](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/app/sync_modes.go:64) maps `change_capture` to source capture/upsert and [`sync_modes.go:123`](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/app/sync_modes.go:123) documents cursor/key requirements. |
| `nl -ba internal/connectors/native/postgres/cdc.go` | The default branch is still a documented unsupported CDC stub at [`cdc.go:10`](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/connectors/native/postgres/cdc.go:10); this report does not assume a released PostgreSQL CDC implementation. |
| `git show origin/fm/cli-postgres-cdc-logical-replication-r1:internal/connectors/native/postgres/cdc.go | nl -ba` | The parked branch starts `proto_version '1'` at lines 92-102, uses scalar `Begin`/`Commit` state at 145-225, emits DML before its parsed commit at 212-220, and advances source status only after `CommitDurableChangefeedCheckpoint` at 199-206. V1's commit-only delivery makes its current ordering viable; v2 streamed aborts require the new stage/transaction port before it can be enabled. |
| `git show origin/fm/cli-postgres-cdc-logical-replication-r1:internal/connectors/connectors.go | nl -ba` | Lines 421-427 state the essential contract already: a durable changefeed checkpoint follows caller acceptance of the emitted source transaction and unacknowledged candidates are not resumable. The current `emit func(CDCEvent)` interface lacks a source transaction/abort boundary. |
| `git show origin/fm/cli-postgres-cdc-logical-replication-r1:internal/connectors/native/postgres/cdc_lifecycle.go | nl -ba` | Lines 260-294 provide explicit teardown of only an inactive derived slot. This supports intentional rebootstrap cleanup but does not automatically protect a failed/stalled slot. |
| `git show origin/fm/cli-postgres-cdc-logical-replication-r1:go.mod` | The parked branch pins `github.com/jackc/pglogrepl v0.0.0-20260401131349-e37c41485510`, whose source contains v2 streaming message support. |

### Specific changes the parked lane needs before admission

- Change the source protocol/checkpoint version deliberately from
  `pgoutput-v1` to a v2-aware version; a checkpoint compatibility decision must
  be explicit, never silently treated as the same decoder semantics.
- Replace the scalar `inTransaction` plus direct callback path with a
  per-source-transaction stage that handles V2 start/stop/commit/abort and
  restart cleanup.
- Keep `DurableCheckpointCommitter` as the final downstream-boundary port, but
  invoke it only after the sealed transaction has been durably accepted. The
  local stage is not an alternate committer.
- Keep the existing derived-slot teardown guard, add slot/journal health
  reporting, and make `rebootstrap_required` visible for retention-gap,
  invalidated-slot, and stage-limit outcomes.
- Add integration coverage with a low server `logical_decoding_work_mem` that
  proves records are streamed before source commit but are not delivered or
  acknowledged before the source commits and the downstream receipt succeeds.

## Limits of the finding

- This report did not run a live PostgreSQL workload. Protocol claims are from
  PostgreSQL's documentation/release notes and the pinned parser source; the
  implementation plan requires an integration test against PostgreSQL 14+
  before admission.
- Vendor documentation demonstrates patterns and operational constraints; it
  is not proof of a vendor's hidden acknowledgement implementation. In
  particular, no correctness inference is taken from Striim's public
  large-transaction marketing claim.
- There is no researched, defensible global percentage of tables with a safe
  incremental cursor. The actionable fact is that the field treats it as
  table-specific configuration plus delete semantics, so PM must do the same
  if it ever offers a separate reconciliation feature.

## Final recommendation for the parked work

Unpark the CDC design only with this contract: **PostgreSQL 14+ streamed
`pgoutput` v2, a bounded per-transaction stage, downstream-receipt-before-LSN
acknowledgement, a hard stage-limit error, and source-slot observability/
explicit rebootstrap lifecycle.** Do not build cursor/timestamp reconciliation
into `change_capture`. If that transaction boundary cannot be implemented in
this phase, leave `change_capture` non-executable rather than shipping the
v1-at-commit burst or a semantic downgrade.

