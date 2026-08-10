# Bidirectional change-capture contract

**Research and design report — 2026-08-10**  
**Recommendation:** adopt one common **Change Delivery Workset** contract above two
different change-observation mechanisms. Do not pretend that a warehouse-derived
delta is a source CDC transaction. In particular, the advertised
'change_capture' mode remains inbound-only; reverse delivery is an explicitly
labelled derived delta with weaker ordering and atomicity facts.

## Executive decision

There is a useful common contract, but it is not a common source-event shape.
The shared contract begins only after either side has established a bounded,
immutable set of changes and ends when a destination receipt makes progress
durable:

    observe changes
        -> freeze a Change Delivery Workset
        -> plan / preview / approval
        -> apply to a declared destination capability
        -> persist a receipt and checkpoint

There are two truthfully different producers of that workset:

| Producer | What PM knows | Boundary that may be named |
|---|---|---|
| Received source feed | PostgreSQL supplies operations, replica identity, order, commit position, and real source transactions. | One committed PostgreSQL transaction. |
| Derived warehouse delta | PM compares a complete, immutable Parquet materialization with the last receipt-backed delivery baseline using DuckDB. | One materialization-to-materialization comparison. It is not a source transaction. |

The common pieces are the stable scope, key, operation, immutable input,
receipt, replay identity, and checkpoint rules. The origin, ordering, delete
evidence, and atomicity are tagged rather than erased. That gives callers one
delivery protocol without claiming that a polling-style comparison can recreate
the source log.

The smallest honest first implementation is deliberately narrow:

1. Retain the already accepted inbound PostgreSQL 14+ streamed pgoutput-v2,
   bounded-stage, receipt-before-LSN-acknowledgement decision unchanged.
2. Add the common workset and receipt seam, but initially implement
   warehouse-derived delivery only to a PM-managed PostgreSQL target with
   keyed, idempotent state upserts.
3. Produce the workset with DuckDB over Parquet and an owned, prior-delivery
   Parquet baseline; never scan the warehouse JSONL WAL for reverse change
   detection.
4. Admit an explicit tombstone only when a dedicated publication projection
   preserves it and the managed target can store the tombstone as keyed state.
   Reject physical-absence deletes and all API destinations in this first
   slice. Do not silently retain stale records.
5. Add physical-absence reconciliation and API delivery only after each has a
   declared delete policy, receipt semantics, and replay proof.

This is a recommendation, not an implementation change. No source files in the
worktree were modified.

## Scope, method, and repository snapshot

The inspected worktree was detached at commit 'f96a47e80'. 'git status --short'
returned no changes. There is no '.codegraph' directory, so the code evidence
below was obtained with page-wise 'rg', 'nl', and read-only 'git show' commands.

The current default snapshot and the active PostgreSQL CDC lane are materially
different:

* Current default 'internal/connectors/native/postgres/cdc.go' is an
  unsupported stub. The in-flight
  'fm/cli-postgres-cdc-logical-replication-r1' branch has the requested
  'cdc_lifecycle.go' and 'cdc_decode.go', but its 'ReadCDC' intentionally
  fails before connecting to a source until streamed transaction staging exists.
  Its descriptor is 'planned', not executable (branch 'cdc.go:18-52').
* This report treats the default branch as current shipped behavior and the
  in-flight branch as design evidence for the fail-closed admission boundary.
  It does not assume that CDC is already released.
* The accepted framework report and accepted large-transaction report are
  binding inputs, not alternatives reconsidered here. The latter's PostgreSQL
  14+ v2 streaming, bounded stage, StreamAbort discard, quota failure, and
  receipt-before-acknowledgement decisions are preserved.

### Commands and local evidence read

| Read-only command or source | Finding used |
|---|---|
| 'git status --short; git rev-parse --short HEAD' | Clean detached snapshot at 'f96a47e80'. |
| 'rg -n ... internal/synccontract/mode.go internal/app/sync_modes.go' | The closed vocabulary has seven modes, and parsing is explicitly separate from native executor/conformance admission. See [mode.go](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/synccontract/mode.go:8) and [sync_modes.go](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/app/sync_modes.go:44). |
| 'rg -n ... internal/app/app.go internal/app/local_warehouse.go' | ETL enriches rows with run, sync time, delete, and cursor facts; raw records also retain a canonical primary-key field. See [app.go](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/app/app.go:1167) and [local_warehouse.go](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/app/local_warehouse.go:38). |
| 'nl -ba internal/app/local_warehouse.go' | A deduped final table folds raw records by primary key and drops deleted rows. See [local_warehouse.go](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/app/local_warehouse.go:339). |
| 'nl -ba internal/warehouse/parquet.go' and reverse Parquet test | Warehouse table reads use DuckDB 'read_parquet'; an end-to-end test proves reverse ETL reads a Parquet table. See [parquet.go](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/warehouse/parquet.go:170) and [warehouse_parquet_read_test.go](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/app/warehouse_parquet_read_test.go:124). |
| 'nl -ba internal/app/app.go' around planning, preview, and run | Current reverse ETL reads a bounded full table slice, hashes it, then invokes one writer call. It has no persisted prior-delivery baseline or derived-delta algorithm. See [app.go](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/app/app.go:1368), [app.go](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/app/app.go:1734), and [app.go](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/app/app.go:2135). |
| 'nl -ba internal/warehouse/layout.go' | Warehouse ownership is the structural triple workspace, connector, and connection ID; display name is excluded from identity. See [layout.go](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/warehouse/layout.go:124), [layout.go](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/warehouse/layout.go:282), and [layout.go](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/warehouse/layout.go:386). |
| 'git show fm/cli-postgres-cdc-logical-replication-r1:.../cdc.go' | The in-flight lane requires a bounded crash-recoverable transaction stage, drops aborts, refuses source acknowledgement on a stage limit, and demands a whole-transaction receipt before LSN acknowledgement. See branch 'cdc.go:36-52'. |
| Accepted reports | The database framework calls for a transaction-scoped CDC sink, connection-owned WAL frames, an extraction checkpoint before standby feedback, a later delivery receipt, and managed target identity. See [framework report](/Users/karthiksivadas/karthik-agent-workspace/data/cli-database-connector-framework-design-r1/report.md:1355), [framework report](/Users/karthiksivadas/karthik-agent-workspace/data/cli-database-connector-framework-design-r1/report.md:1427), and [large-transaction report](/Users/karthiksivadas/karthik-agent-workspace/data/cli-cdc-large-transaction-strategy-r1/report.md:21). |

## Ground truth that constrains the design

### The current mode vocabulary is not a bidirectional promise

The closed contract names:

* 'full_overwrite'
* 'full_append'
* 'incremental_append'
* 'incremental_upsert'
* 'incremental_dedupe'
* 'incremental_dedupe_history'
* 'change_capture'

That list is a vocabulary, not an admission grant. 'mode.go:8-10' says so
explicitly. The current app maps 'change_capture' to a change-capture source and
an upsert destination; it does not establish that a reverse-ETL warehouse
comparison is a source feed. 'sync_modes.go:123-155' also makes cursor and key
requirements directional implementation facts.

### The available warehouse facts are necessary but insufficient by themselves

Every ETL record is enriched with:

* '_polymetrics_run_id'
* '_polymetrics_synced_at'
* '_polymetrics_deleted'
* '_polymetrics_cursor' when the stream has a cursor

The local raw envelope additionally has '_polymetrics_raw_id',
'_polymetrics_primary_key', extraction/load lineage, and opaque cursor state.
That is valuable raw material for change derivation, but it is not yet an
outbound checkpoint:

* A new extraction run changes its run lineage and sync timestamp even where
  business payload is unchanged. Those volatile fields must not make a record
  appear changed.
* A deduped Parquet table omits deleted rows entirely. A consumer looking only
  at that final table cannot tell whether a missing key was deleted, filtered
  out, was never in the prior source, or fell outside a bounded read.
* Current reverse plans default to a finite read limit and preserve a source
  connection selector, then re-read the same slice for preview and write. That
  is correct approval drift protection for a bulk plan, but it is not a complete
  key reconciliation or a delivery cursor.

Therefore the outgoing change detector needs its own durable, destination-scoped
baseline and a complete candidate materialization. It must not query the WAL:
the WAL remains the appendable extraction durability log, whereas a Parquet
table and its derived delivery indexes are the reverse-delta query surface.

### Ownership is an input to both identities

'warehouse.LocationFor' constructs a directory from workspace ID, connector,
and connection ID. 'Owner.SameIdentity' intentionally ignores the display name.
'AssertOwnedTable' independently rejects a foreign or missing owner record.
The delivery contract must preserve that rule rather than keying a baseline,
receipt, or managed target on a table name or a raw credential.

The accepted framework additionally requires connection IDs, not names, to
survive reverse planning and binds the source owner triple into target identity
and receipts. See [framework report](/Users/karthiksivadas/karthik-agent-workspace/data/cli-database-connector-framework-design-r1/report.md:1593).

## Field research: how established systems draw the line

The following sources were read as architectural evidence only. They do not
authorize a dependency or implementation choice, and their names must not
appear in shipping code, comments, public docs, PR bodies, or commit messages.

| System and documentation read | Documented mechanism | Design conclusion for PM |
|---|---|---|
| [Hightouch sync overview](https://hightouch.com/docs/syncs/overview) and [Lightning sync engine](https://hightouch.com/docs/syncs/lightning-sync-engine) | It documents warehouse-side CDC/diffing by comparing all rows in a model with the last sync, stores prior data in managed warehouse tables, requires a unique primary key, and reports added, changed, and removed rows. It also documents that a full resync cannot identify removed rows because there is no prior diff. | A warehouse-derived delta needs a retained prior projection and a unique key. A current query result alone cannot prove a deletion. |
| [Snowflake CREATE STREAM](https://docs.snowflake.com/en/sql-reference/sql/create-stream) | A stream exposes insert/delete action metadata, an immutable row ID, and advances its offset at the end of a DML transaction. Multiple statements can read the same stream data inside an explicit transaction. | Warehouse-native logs can have real offsets and transactional consumption. PM's Parquet store does not currently have that feature, so PM must label its comparison as derived rather than borrow the word transaction. |
| [Databricks Change Data Feed](https://docs.databricks.com/gcp/en/tables/features/change-data-feed) | CDF supplies change type, commit version, and commit timestamp; it records changes with the table transaction log and makes a batch include an entire commit. It also states that feed history is transient under retention. | A native table CDF has stronger event and commit facts than a whole-table comparison. Retention and explicit checkpoints remain necessary even where the table supplies a log. |
| [dbt snapshots](https://docs.getdbt.com/docs/build/snapshots) and [dbt hard deletes](https://docs.getdbt.com/reference/resource-configs/hard-deletes) | Snapshots compare current mutable rows with history using a unique key and either timestamp or selected-column checks. Hard deletes are an explicit policy: ignore, invalidate, or create a delete-marked record. | State comparison is a valid mechanism, but it needs an explicit key and explicit delete policy. It cannot silently infer a historical event stream. |

The research supports the asymmetry in the task rather than removing it:
received CDC retains source-native operation and transaction facts; reverse ETL
diffing retains a prior state and derives a current-state transition. Mature
systems use both, but the contracts expose their different sources of truth.

## Recommended common contract

### 1. Contract placement and vocabulary

Keep the accepted inbound connector port:

    ChangefeedTransactionSink
      BeginChangefeedTransaction
      AppendChange
      CommitChangefeedTransaction(candidate checkpoint)
      AbortChangefeedTransaction

It is the correct lossless boundary for an inbound source transaction. The
accepted framework specifies that successful commit means both the complete
transaction data and candidate checkpoint are durable
([framework report](/Users/karthiksivadas/karthik-agent-workspace/data/cli-database-connector-framework-design-r1/report.md:1357)).

Add an adjacent higher-level 'Change Delivery Workset' contract in
'synccontract', consumed by app planning and by destination writers. It is
created from either a committed inbound transaction or a derived warehouse
comparison. It must have these required facts:

| Fact | Received workset | Derived workset |
|---|---|---|
| 'origin' | 'received_source_feed' | 'derived_warehouse_delta' |
| 'scope' | Source identity, generation, stream, and the connection-owned warehouse owner triple. | The same source owner triple plus destination fingerprint, logical stream, mapping revision, delete policy, and baseline identity. |
| 'boundary' | Source transaction ID, transaction-end LSN, source order, and ordinal. | Prior delivery baseline plus one immutable candidate Parquet materialization. |
| 'changes' | Source operation, replica key, before/after images where supplied, schema fingerprint, and ordinal. | Keyed upsert or delete intent, baseline/current canonical payload fingerprints, and a deterministic order. It must not claim a source insert/update distinction it cannot prove. |
| 'workset identity' | Source identity + generation + end LSN + transaction ordinal/range. | Delivery scope + baseline ID + candidate materialization ID + mapping/schema/delete-policy revision. |
| 'receipt requirement' | Extraction receipt before source acknowledgement; later target delivery receipt before delivery progress. | Target delivery receipt before advancing the per-destination baseline/checkpoint. |
| 'atomicity label' | 'source_transaction' for source observation; target atomicity remains separately declared. | 'materialization_comparison'; no source transaction claim. Target atomicity remains separately declared. |

Each change has a stable key, a delivery operation, and a replay identity:

* Preserve the source operation separately from delivery operation. An inbound
  source insert or update normally becomes a delivery upsert; a source delete
  becomes a delivery delete or tombstone. A derived comparison produces
  delivery upsert/delete only.
* A received event identity is source identity, generation, transaction-end
  LSN, and per-transaction ordinal. A single LSN is not unique enough for
  multiple DML messages; the accepted framework has already identified that
  requirement.
* A derived change identity is a hash of delivery scope, immutable baseline and
  candidate IDs, canonical key, delivery operation, canonical mapped business
  payload or tombstone, and mapping/schema/delete-policy revision. It excludes
  run ID and sync timestamp.
* A key-changing inbound update must retain the old replica key and new row
  state, then normalize to a delete-old/upsert-new pair inside the received
  source transaction if the sink does not support key moves. The current
  in-flight decoder discards an old tuple for an update, so it is not a
  sufficient source for this contract until that loss is repaired.
* 'truncate' is not a row delete. It remains separately gated and rejected
  unless a destination declares an approved, safe truncate semantics.

This is one contract because every caller can see what will be delivered, why it
is eligible, how it is replayed, and when it is acknowledged. It is not a leaky
false symmetry because every workset carries its origin and boundary kind.

### 2. Freeze the input before approval

The current table pathname is not a sufficient delivery input because a later
ETL run can atomically replace its Parquet file. Before preview/approval, the
derived builder must:

1. Resolve and assert the source owner identity.
2. Read a complete current publication projection from Parquet with DuckDB.
   It may not use the current reverse-plan limit as a reconciliation boundary.
3. Read the last durable baseline projection for the exact delivery scope.
4. Generate a deterministic delta with a DuckDB keyed comparison.
5. Persist an immutable candidate workset Parquet file plus manifest containing
   the input digests, key schema, mapping revision, delete policy, workset ID,
   and ordered change count.
6. Bind that manifest into the existing plan, preview, approval, and target
   receipt. A later source change creates a new workset; it does not mutate the
   approved one.

The baseline should be a compact owned Parquet projection of canonical key,
canonical mapped payload fingerprint, tombstone state, and last receipt-linked
change ID. It is not a second copy of the extraction WAL. DuckDB can compare
the two owned Parquet projections with a full keyed join.

### 3. Destination capability is explicit

A destination must declare, for the selected action:

* stable key matching and canonical key encoding;
* upsert semantics;
* supported delete behavior: tombstone, idempotent delete, or unsupported;
* receipt kind and the evidence returned on success;
* replay mechanism: managed database convergence, target receipt ledger, or
  provider-issued idempotency key plus status lookup;
* actual atomicity: target transaction, delivery chunk transaction, or
  per-record;
* whether the destination is PM-managed and can carry reserved system columns.

No generic HTTP write or generic SQL write receives this capability. API
destinations omit PM metadata from their payloads, as the accepted framework
already requires; metadata belongs in the PM receipt state or a PM-managed
database control schema.

## Direct answers to the hard questions

### 1. Deletes in both directions

| Case | Inbound ETL | Outbound reverse delivery |
|---|---|---|
| PostgreSQL row delete | The logical feed supplies a delete under source replica identity. The lossless envelope must retain the key and supplied old/before data. A relation without an admitted replica identity fails rather than guessing. | A preserved explicit tombstone can create a keyed tombstone/delete intent. The destination must have declared a safe delete or tombstone action. |
| Physical absence from current state | A source log knows this was a delete event; there is no need to infer it from a scan. | Detect only by complete keyed reconciliation: keys in the prior receipt-backed baseline but absent from the full current authoritative projection. This is a state removal, not proof of a historical source delete event. |
| Current deduped warehouse table | The raw record retains the deleted fact. | It is insufficient alone: dedupe omits deleted rows from final Parquet. Do not read the WAL to compensate. A phase-one publication projection must retain explicit tombstones, or physical absence must be rejected until full reconciliation exists. |
| Destination without safe delete semantics | Not applicable to accepting the source change; it remains durably extracted. | Default is admission failure for a workset containing that delete. Do not report success while retaining a stale record. A future explicit retain policy could be offered only as a degraded state-upsert contract and must state that removals are not propagated. |
| Truncate | Separate operation, not a generated list of row deletes. | Reject in phase one. A later destination-specific operation needs its own preview/approval and receipt semantics. |

The recommended complete design supports two mutually visible delete signals:

1. An explicit normalized '_polymetrics_deleted' tombstone in a dedicated
   publication projection, with an admitted stable key.
2. A full key anti-join against the last delivery baseline when the source
   projection is declared complete and authoritative.

The first phase uses only the first signal and only maps it to a tombstone in a
PM-managed PostgreSQL target. It neither physically deletes customer rows nor
claims to mirror physical absences. The anti-join is a later feature because it
must prove that the current Parquet input is complete, unbounded, keyed, and
scoped to the same delivery contract.

### 2. Ordering and transaction boundaries

Inbound has a genuine source boundary. PostgreSQL v2 stages in-progress chunks
privately, publishes only at 'StreamCommit', discards 'StreamAbort', preserves
transaction order, and attaches the transaction-end LSN and ordinal. The
accepted large-transaction decision owns these facts; this design does not
weaken them.

Outbound has a delivery window, not a source transaction:

* The logical boundary is one comparison from receipt-backed baseline to
  immutable candidate materialization.
* Changes receive a deterministic order by canonical key and a documented
  operation ordering. That order is for reproducible delivery and replay, not
  a claim about original source commit order.
* A managed database writer may commit one bounded delivery chunk in one real
  target transaction, but only its confirmed commit receipt can call that chunk
  atomic.
* API writers are per-record unless an individual connector proves an atomic
  provider operation. Transport batching is not atomicity.

Consequently, "per-batch atomicity" is honest only when the concrete sink
returns a receipt for exactly one target transaction covering the batch. It is
not a universal reverse-ETL guarantee. The common workset never labels
warehouse-derived changes as source transactions.

### 3. Durability boundaries and receipts

Inbound has two distinct durable milestones:

    PostgreSQL v2 chunks
        -> bounded private transaction stage
        -> StreamCommit
        -> sealed connection-owned WAL begin/event/commit frame, fsync
        -> durable extraction checkpoint for transaction-end LSN
        -> PostgreSQL standby-status acknowledgement
        -> immutable delivery workset
        -> later target delivery receipt/checkpoint

The private stage can be fsynced for crash recovery, but it is not a receipt and
never advances the source slot. The extraction receipt is the sealed complete
committed source transaction plus durable extraction checkpoint in the
connection-owned warehouse. That is the downstream durable boundary for source
acknowledgement. It reconciles the accepted reports: source acknowledgement
does not wait through human approval for a final target, but it also never
acknowledges merely in-progress staged data. The accepted framework specifies
this exact split: local WAL frame and checkpoint before standby feedback, then
an immutable delivery window and target receipt
([framework report](/Users/karthiksivadas/karthik-agent-workspace/data/cli-database-connector-framework-design-r1/report.md:1427)).

Outbound has a different receipt:

    immutable Parquet workset
        -> plan / preview / approval
        -> target apply
        -> confirmed target receipt
        -> durable local receipt record and delivery checkpoint/baseline advance

For managed PostgreSQL, the target receipt is a successful COMMIT after the
owner assertion and required local durability settings. The accepted design
requires server 'fsync=on' and transaction 'synchronous_commit=on', and says a
checkpoint may advance only after that confirmed commit
([framework report](/Users/karthiksivadas/karthik-agent-workspace/data/cli-database-connector-framework-design-r1/report.md:566)).

For an API, an ordinary request attempt or bare HTTP success is not sufficient
evidence. A connector may return a receipt only when its documented native
protocol proves the selected idempotent operation has completed, with a
provider identifier or status lookup sufficient for replay. Otherwise the
outcome is unknown, the delivery checkpoint remains unchanged, and automated
blind retry is prohibited.

### 4. Idempotency and replay after an unknown outcome

The proof comes from a restricted operation set, stable workset identity, and
receipt ordering:

1. The immutable workset makes every retry use the same canonical key,
   operation, payload, and change ID.
2. A phase-one managed database upsert is state-idempotent: retrying the same
   keyed upsert converges to the same business/tombstone state rather than
   inserting a second row. It is an at-least-once delivery mechanism, not
   exactly-once external-effect delivery.
3. The delivery checkpoint and baseline advance only after the confirmed target
   receipt is durably persisted. A crash before that point replays the exact
   workset.
4. If a database commit is unknown, replaying a keyed state upsert is safe for
   the state claim but must still surface the uncertainty. A full receipt ledger
   is needed before claiming exactly-once operation execution.
5. Append, event, and generic webhook actions do not have that convergence
   property. An unknown result can duplicate customer-visible data, so phase one
   does not admit them for derived delivery.
6. A future API writer is admitted only if it accepts the deterministic change
   ID as a provider-supported idempotency key and can reconcile a timeout from
   provider state; otherwise it stops at 'DeliveryOutcomeUnknown'.

This matches the accepted framework's explicit ruling: phase-one append is
at-least-once, unknown commits remain explicit, no checkpoint advances, and no
blind safe retry is claimed without the deferred receipt ledger
([framework report](/Users/karthiksivadas/karthik-agent-workspace/data/cli-database-connector-framework-design-r1/report.md:1604)).

### 5. Identity

Inbound identity comes from the PostgreSQL source identity and admitted replica
identity, plus generation, transaction-end LSN, and ordinal. Delete and
key-changing update handling must use that key rather than a record map's
accidental field order.

Outbound identity has two scopes:

* The source owner is the warehouse structural triple: workspace, source
  connector, and source connection ID. The display name is not an identity
  component.
* The delivery scope additionally binds logical stream/table, target connector
  and non-secret target fingerprint, credential revision, mapping/schema
  revision, key definition, delete policy, and contract version.

The stable record key is a required, canonicalized source primary-key tuple,
not '_polymetrics_raw_id', run ID, sync time, or a raw credential. The final
Parquet publication projection must carry it explicitly; a value held only in a
raw WAL envelope cannot be used by a Parquet/DuckDB-derived detector.

Persist the opaque source connection ID in a reverse plan and in each delivery
receipt. Existing reverse-plan selectors prevent ambiguous table reads, but the
accepted framework correctly requires a migration from connection names to
connection IDs for rename-safe durable state.

### 6. Truthfulness by advertised mode and direction

This table says what may be claimed after the proposed contract exists. Every
cell still requires native conformance admission; parsing one of these mode
names is never enough.

| Mode | Inbound ETL truthfulness | Outbound reverse truthfulness | Replay and phase-one disposition |
|---|---|---|---|
| 'full_overwrite' | An explicit full snapshot/rebootstrap, not a source change stream. It can establish current state after an intentional CDC discontinuity but cannot restore missing event history. | A full replacement of a PM-managed database target can be truthful only if one actual target transaction replaces the complete frozen snapshot. It is not "only what changed." Generic APIs are not admitted. | Repeating the same committed transactional replacement converges. Not part of derived-delta phase one. |
| 'full_append' | A snapshot append, not CDC. It has no general delete or source-event history claim. | Not a safe derived-delta action. Replaying an unknown outcome can duplicate records. | At-least-once only; no phase-one derived delivery. |
| 'incremental_append' | A cursor-based current-state read where the source declaration proves an ordered cursor. It does not recreate hard deletes or source transaction history. | A future insertion-only delivery could be admitted only with an immutable event source and a destination idempotency proof. A Parquet state comparison alone is insufficient for the broad claim. | No phase-one derived delivery. |
| 'incremental_upsert' | A cursor/key current-state contract, not a transaction feed. It needs the existing cursor/key admission and explicit delete handling. | The phase-one reverse mode: derive keyed business-payload changes from Parquet plus baseline and send a state upsert to a PM-managed PostgreSQL target. Explicit tombstones only under the narrow policy above. | At-least-once state convergence; no exactly-once external-effect promise. |
| 'incremental_dedupe' | Current-state incremental delivery with deterministic key/winner semantics; it remains distinct from event history. | Can later consume the same workset if its winner/order rule is bound into the workset and destination semantics are keyed state. | Design-compatible, but postpone until the workset baseline proves the same winner rule. |
| 'incremental_dedupe_history' | Truthful only when a source explicitly supplies immutable ordered history. A cursor scan over mutable state is not enough. | A before/after Parquet comparison cannot reconstruct all historical intermediate events. | Not admitted for derived delivery; future work needs a durable event archive and receipt ledger. |
| 'change_capture' | Truthful only for an admitted source CDC executor: PostgreSQL 14+ v2 stream, committed transaction boundary, deletes/keys preserved, quota fail-closed, and extraction receipt before LSN acknowledgement. The current PostgreSQL branch stays fail-closed until then. | Not an outbound selectable synonym. A derived workset may carry changes to a target, but its origin must be 'derived_warehouse_delta' and its boundary 'materialization_comparison', not 'change_capture'. | Inbound phase remains the accepted at-least-once receipt design. Outbound phase one uses 'incremental_upsert', not this mode. |

## Smallest honest phase one and the seam for later work

### Phase one: bounded state delivery, not a fake reverse CDC feed

1. **Common types and validation.** Introduce the additive workset, receipt,
   delivery-scope, boundary-kind, and destination-capability types. Preserve
   the accepted transaction-scoped inbound connector sink; adapt it into a
   received workset only after commit.
2. **Complete Parquet publication projection.** Materialize an owned Parquet
   projection with canonical key, mapped-business payload digest, schema/mapping
   revision, and normalized tombstone state. It must be a complete projection
   and must not inherit a CLI read limit. If dedupe would omit a tombstone,
   preserve the tombstone in this separate projection rather than reading the
   WAL during reverse planning.
3. **DuckDB diff and immutable workset.** Compare current projection with the
   exact per-destination prior-delivery Parquet baseline. Freeze the resulting
   workset and manifest before plan/preview/approval. Start with additions and
   changed keys plus explicit tombstones; reject physical absence.
4. **One managed PostgreSQL delivery capability.** Require an exact owner
   assertion, canonical non-null key, state upsert action, and a real target
   transaction. Store target tombstones as managed state rather than deleting
   arbitrary customer objects.
5. **Receipt ordering.** After target COMMIT, persist the target receipt, item
   receipts or committed chunk range, and next baseline/checkpoint. On any
   failure before this local persistence, retain the immutable workset for
   replay; do not mark it delivered.
6. **No API, no append, no physical removal mirroring.** Refuse those
   configurations with a named reason. That protects customer systems from
   duplicate or stale silent behavior while the idempotency and delete seams
   remain unimplemented.

### Forward seams

The first slice leaves these clean extensions rather than semantic debt:

* **Physical-absence reconciliation:** accept only an authoritative, complete,
  keyed Parquet projection; anti-join against the receipt-backed baseline; bind
  the completeness declaration and delete policy to the plan.
* **API sinks:** require a connector-specific idempotency-key mapping, provider
  receipt/status lookup, explicit delete capability, and per-record receipt
  persistence. No generic transport path.
* **Exactly-once operation claim:** add a receipt/idempotency ledger in a
  PM-managed target control schema or provider reconciliation implementation.
  Until then retain the at-least-once state-convergence wording.
* **Dedupe/history modes:** add deterministic winner rules or immutable event
  archive semantics as an explicit source/delivery capability, never as an
  inference from current Parquet state.
* **Broader database drivers:** each declares its own DDL/transactional
  capability. A driver with implicit-commit DDL cannot claim PostgreSQL's
  atomic provisioning or receipt semantics.

### Required verification for implementation

The first implementation should have failure-injection tests that assert data
and checkpoint contents, not merely exit status:

1. An unchanged business record with a new run ID/sync timestamp produces no
   derived upsert.
2. A changed mapped business field produces exactly one deterministic workset
   change; mappings/schema/key changes invalidate the old baseline.
3. A preserved explicit tombstone is delivered only to a declared tombstone
   sink; a physical absence is rejected in phase one.
4. A duplicate key, null key, bounded input, ambiguous/foreign warehouse owner,
   or mismatched destination fingerprint fails before target mutation.
5. A target failure, crash before local receipt, and unknown commit leave the
   delivery checkpoint unchanged; replay of a keyed managed upsert converges.
6. A received PostgreSQL transaction has no visible workset or LSN
   acknowledgement before StreamCommit and durable extraction receipt; abort
   leaves no workset.
7. A source and target connection sharing a human-readable table name produce
   distinct owner-scoped baseline, workset, target, and receipt paths.

## Recommendation and non-goals

Adopt the shared Change Delivery Workset contract now, with origin and boundary
as required first-class fields. It is the smallest abstraction that correctly
shares durability, identity, replay, plan/preview/approval, and receipt
machinery across both directions.

Do not call a Parquet/DuckDB comparison a source changefeed, do not change the
meaning of 'change_capture', and do not use a cursor/timestamp scan as an
automatic CDC recovery. Do not route delete semantics through a destination
that cannot prove them, and do not call a transport batch atomic merely because
PM grouped records into it.

This report leaves no captain decision open: the recommendation selects the
narrow phase-one admission above and records the later capability gates rather
than asking a human to choose among unsafe defaults.

## Sources

### Local and accepted-design evidence

* [Current mode vocabulary](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/synccontract/mode.go:8)
* [Current source/destination mode mapping](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/app/sync_modes.go:44)
* [Warehouse enrichment and raw record envelope](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/app/local_warehouse.go:38)
* [Final Parquet dedupe deletion behavior](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/app/local_warehouse.go:339)
* [DuckDB Parquet reads](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/warehouse/parquet.go:170)
* [Current reverse plan path](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/app/app.go:1368)
* [Warehouse owner identity](/Users/karthiksivadas/.treehouse/cli-83d592/1/cli/internal/warehouse/layout.go:124)
* [Accepted database connector framework](/Users/karthiksivadas/karthik-agent-workspace/data/cli-database-connector-framework-design-r1/report.md:1355)
* [Accepted large-transaction strategy](/Users/karthiksivadas/karthik-agent-workspace/data/cli-cdc-large-transaction-strategy-r1/report.md:21)

### External documentation read

* [PostgreSQL logical-replication protocol](https://www.postgresql.org/docs/current/protocol-logical-replication.html)
* [Hightouch sync overview](https://hightouch.com/docs/syncs/overview)
* [Hightouch Lightning sync engine](https://hightouch.com/docs/syncs/lightning-sync-engine)
* [Snowflake CREATE STREAM reference](https://docs.snowflake.com/en/sql-reference/sql/create-stream)
* [Databricks Change Data Feed](https://docs.databricks.com/gcp/en/tables/features/change-data-feed)
* [dbt snapshots](https://docs.getdbt.com/docs/build/snapshots)
* [dbt hard deletes](https://docs.getdbt.com/reference/resource-configs/hard-deletes)

## Completion notes

This is a Markdown research report, not a visual artifact or a UI change, so no
visual review was applicable. No production code, generated artifact, commit,
push, or pull request was created.

