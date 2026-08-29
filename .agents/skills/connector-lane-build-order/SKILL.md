---
name: connector-lane-build-order
description: >-
  Build or review source-locked connector operations, direct commands, binary
  transfers, ETL, reverse ETL, and saved transports without confusing provider
  evidence with runtime policy. Use whenever connector lanes, source locks,
  projected JSON definitions, certification, or usable CLI parity are in scope.
user-invocable: false
metadata:
  internal: true
---

# Connector lane build order

Use this procedure before designing, declaring, implementing, or reviewing a
connector lane. It owns the construction order, the source-authoring versus
runtime-loading boundary, the direct-command versus saved-connection boundary,
and the evidence required to call a connector operation usable. Provenance, an
`implemented` label, and a passing declaration-only gate do not prove execution.

## 1. Lock the provider source first

Find the provider artifact that actually defines the connector surface before
modeling anything downstream. Prefer a provider-published OpenAPI or Swagger
document, GraphQL schema, discovery document, or equivalent machine-readable
specification.

When no machine-readable artifact exists, retain and hash-pin the provider's
rendered reference pages through the `rendered_reference` import path, including
the documented request and response shapes. If the provider pages are
insufficient or unavailable, Context7 may supply citable documentation evidence,
but it is not a retained provider artifact. Mark an operation unmappable only
after none of these source kinds yields a citable contract, and record what was
tried.

Validate a fetched candidate before retaining it. An error page, login wall,
redirect target, or documentation index is not a specification. Retain the exact
verified provider bytes and digest. Preserve a copy of
`internal/connectors/defs/<connector>/sources/` before changing it because local
source material may be untracked and Git may not restore it. Treat upstream
drift as an explicit re-pin with old and new bytes and digests; replacement is
not preservation.

The source lock is the connector's provider-fact authority. It must contain or
cite enough contract detail to account for each downstream mapping, plus an
explicit gap wherever the provider does not publish a required fact. Do not put
system execution policy into provider provenance.

## 2. Separate source authoring from runtime loading

Map operations to the retained machine-readable operation node when one exists,
or to retained rendered-reference sections when it does not. Every declared
operation, field, schema constraint, media type, and response shape needs a
source-lock citation. Documentation-derived mapping never permits inferring a
contract the documentation does not state.

For a named connector cohort, freeze and commit one machine-checkable
source-lock denominator before mapping. Reconcile every source identity to one
of `implemented`, `blocked_with_named_foundation`, or
`unsupported_with_provider_evidence`, and retain its citation, projected lane
cell, command/help reachability, and named gap. Certification, credentials,
and importer limitations are overlays; they cannot erase a source identity or
decide cohort membership. Consult the Foundation Atlas before recording a
missing foundation. Provider-only legacy source fragments may be retained as
non-executable evidence, but they never become a generic transport or a reason
to infer an undocumented provider contract.

The mandatory authoring path is source lock → mapping/projection → connector
definition artifacts → lane-specific execution witness. Do not skip from source
documentation to certification, and do not use certification output as an
authoring input.

The runtime loads projected connector JSON, not the source lock. The source lock
nevertheless must account for every provider fact used by `spec.json`, schemas,
streams, writes, operations, API and CLI surfaces, and saved-connection transport
definitions. Keep connector execution policy and conformance overlays separate:

- provider facts and source-backed mappings originate in the immutable lock;
- projected runtime JSON turns those facts into executable declarations;
- system-owned limits, approvals, warehouse mediation, retry policy, and lane
  admission live in explicit execution-policy fields or overlays;
- certification verifies one exact source-lock digest, projected-definition
  digest, executor, scope, resource type, and mode.

Certification cannot configure runtime, supply a missing provider fact, or make
a declaration true merely by repeating it. It is proof-only: missing live
certification cannot block source mapping or an independently executable lane.
Raw bytes, a URL, or a digest prove provenance, not execution semantics.

Keep the three durable data roles distinct:

- JSON source locks retain immutable provider facts and citations;
- JSONL WAL segments and manifests retain immutable staged worksets, receipt
  identity, continuation, and candidate checkpoints for recovery;
- DuckDB over Parquet owns warehouse materialization and queryable connector
  state.

A provider key identifies a source resource, and an opaque provider cursor or
sync token identifies source progress. A local receipt ID and binding identify
one staged warehouse apply. None may be substituted for another.

## Consult the Foundation Atlas before claiming a gap

Before creating or updating a connector's `missing-foundation.json`, consult
the CLI-owned `docs/connector-canon/foundations/` Atlas when present. Inspect
the matching owner symbols and proof tests, then classify each claimed gap as
reuse, a constrained extension, or genuinely absent. Record the exact gap or
contract mismatch and matching Atlas ID; never duplicate provider facts or call
a capability missing merely because an importer or certification label exists.
The Atlas is authoring-only: it cannot block source mapping or certification.
Implementing a genuinely new runtime foundation still requires the captain's
approval.
After changing a real foundation, maintain its Atlas entry in the same change
under the Atlas README's procedure. Keep provider-specific execution behind a
closed connector-definition reference; shared owners must not branch on
connector name.

For a possible shared runtime gap, first add only an Atlas `investigating`
entry with the owner symbols, closed contract mismatch, affected source rows,
and proof-test plan. Re-check existing foundations and declared extension seams
before asking the captain for approval. While it is investigating, retain every
affected operation as a typed source-cited gap and continue materializing
unaffected existing-runtime lanes. Do not call the candidate planned or
available, add a connector-specific hook, or implement an engine/CLI path until
the captain explicitly approves that named foundation.

## 3. Declare only source-backed operations

Declare an operation, field, media type, and schema constraint only when the
locked source states it. Do not invent operations, fields, media types, stable
keys, or `additionalProperties: false` to make a connector look complete. Keep
an unsupported or unavailable capability explicit, with its source or runtime
blocker, instead of filling the lane with a plausible declaration.

## 4. Build bounded provider commands first

Classify a provider-facing command by what one invocation actually does:

- `direct_read`: one bounded provider read that returns provider output;
- `direct_write`: one exact provider mutation for one input record;
- `binary_download`: the binary form of a bounded direct read;
- `binary_upload`: the binary form of an exact direct write.

A stream-backed command is still a direct read only when the command constrains
the executor to the promised bounded provider request. A record `Limit` does not
bound provider calls when pagination or fanout remains enabled; set and test the
request/page/fanout bound explicitly. One interactive request budget must cover
discovery, every fanout child, pagination, retries, and redirects in aggregate;
saved ETL retains its separate transport lifecycle and established zero-value
budget behavior. If a command exhausts a stream, owns a checkpoint, or may issue
an unbounded number of provider reads, it is not a one-request direct command
merely because its result is printed interactively.

Direct commands do not own warehouse checkpoints, replication state, saved
sync lifecycles, schedules, or flows. Do not label them ETL or reverse ETL just
because the underlying executor can stream records.

## 5. Project only closed source-backed request shapes

Treat request and response direction separately. Exclude OpenAPI `readOnly`
properties recursively from projected request schemas, including properties
introduced by resolved `allOf` arms, remove excluded names from `required`, and
reject attempts to send them before provider I/O. A bounded named-object
fallback must not reintroduce those fields.

Encode array query parameters from the locked style and explode facts. For
OpenAPI form parameters with `explode: false`, serialize with the documented
delimiter (normally comma), omit an absent value, and do not double-encode it.
Never accept a language-native slice string such as `[a b]` as wire evidence or
hard-code a provider name into shared encoding.

Structured `record.*` JSON flags are admitted only for action-backed
`direct_write` commands and only at fields present in the action's closed
request schema. Operation-backed direct writes remain distinct; this rule does
not create a generic JSON-body or HTTP path.

An arbitrary-MIME binary upload requires an explicit source-backed closed media
policy for that exact file part. It remains path-confined, byte-bounded,
digest-bound, previewed, and approval-gated. An absent allow-list is not evidence
that the provider accepts every media type.

A documented provider batch endpoint is executable only as a closed
declared-action adapter. Its definition allow-lists existing named typed actions
and methods; the engine derives every relative path and body from those actions.
Reject caller-authored HTTP methods, paths, raw bodies, query-bearing
subrequests, nested batches, and undeclared actions before provider I/O.

## 6. Compose ETL through the local warehouse

In the current runtime, the local warehouse boundary means DuckDB. An ETL
connection is a saved provider-to-DuckDB pipeline that invokes a connector read
mapping and durably materializes its output. A reverse-ETL plan reads DuckDB
warehouse rows, maps one row to one bounded provider write action, previews the
exact requests, consumes approval, and dispatches through that action.

Connector-to-connector sync composes two saved sides as source -> DuckDB ->
destination. It must never create a hidden API-to-API shortcut. Schedules and
flows orchestrate saved connections at the system layer; they are not
connector-owned CLI intents.

Build in this order:

1. source-locked bounded provider operation;
2. projected direct read or direct write declaration and executable proof;
3. saved warehouse-mediated connection that references the operation;
4. scheduler, flow, or managed transport composition.

Preserve the warehouse boundary even when an interactive command offers a
convenient one-record route.

Preserve connector JSON shape across the DuckDB/Parquet boundary. Nested JSON
strings remain strings even when they look like dates or timestamps, scalar
types round-trip unchanged, and a valid batch made only of `{}` rows retains
its cardinality. Empty-object reconstruction requires file-bound metadata and
the expected physical schema; a sentinel column name alone is not authority.

## 7. Separate full refresh from provider-backed incremental ETL

A full refresh must exhaust one declared source scope before reporting success.
Full-refresh overwrite replaces destination state only after the complete read;
full-refresh append appends another complete snapshot. Page hashes, row hashes,
timestamps, ordinals, and content comparisons do not create an incremental
cursor.

Incremental execution requires documented provider cursor or event-token
semantics and an executor that consumes that exact position, follows documented
continuation such as `has_more`, and persists the next position only after the
warehouse durably acknowledges the token window. Implement documented bootstrap,
expiration, and rebootstrap behavior exactly, including any required new full
snapshot.

An event-source JSON contract may bind an exact registered executor and
conformance reference to immutable source-lock citations. Keep provider facts
(scope, resource types, request parameters, bootstrap/reset status, response
pointers, actions, stable identity, hydration, snapshot, authentication, and
explicitly undocumented ordering) separate from runtime policy. The contract
must be closed data, never an arbitrary lifecycle, handler, code path, retry
algorithm, page cap, coalescing function, checkpoint-commit hook, or
connector-name switch. The registered connector-specific executor owns those
behaviors. Treat any future common event lifecycle as a separately reviewed
foundation; do not generalize one connector's evidence contract by imitation.

Before admitting an event-token mode for a stream, retain evidence that the
subscription scope and emitted resource types cover that stream. A hydration
endpoint alone does not prove event coverage. Hydrate only supported event
actions and resource types.

When the provider distinguishes `deleted` and `removed`, only `deleted` may
produce a resource tombstone, and only when the stable key and deletion scope
are known. `removed` is a scope or relationship change, not a global deletion.

When the provider guarantees token windows but no total event order, a complete
window may be coalesced by stable key and hydrated to final current state. This
proves only order-independent current-state application. It must reject partial
windows and must not be admitted for ordered history or change capture. Ordered
history requires documented positions or versions sufficient to reproduce every
transition, including deletes.

If a requested mode lacks one of these prerequisites, return a deterministic
mode-not-executable error before credential use, provider I/O, or warehouse
mutation.

Stage one complete opaque event-token window as an immutable JSONL receipt and
persist `PendingTransportApply` after every stage, under the same valid lease,
before any next provider read. The pending record retains the exact receipt,
candidate checkpoint, typed continuation where the source contract supports
one, and a versioned canonical binding with integrity SHA-256. Recovery reopens
and finishes that receipt before source I/O, records the durable acknowledgement,
then atomically promotes the checkpoint and clears pending state. It must never
reread A and expand a previously staged A→B window into A→C.

The durable receipt ledger treats the same receipt ID and same binding as an
exact replay: return the original acknowledgement without appending again. The
same ID with a different target, mode, manifest, checkpoint, or integrity
binding fails before mutation. A payload or row-content hash is not provider
idempotency, a provider cursor, a source key, or permission to deduplicate equal
content across token windows.

Use the existing JSONL/WAL plus manifest substrate, not a second storage engine.
If a workset requires multiple segments for atomic recovery, seal the smallest
group manifest over those immutable receipts and necessary continuation. Use a
generic typed continuation only where the source contract defines resumable
positions; otherwise stage the complete opaque provider window atomically.
Fence both durable stage publication and pending binding under the same lease.
Prove crash-after-stage, apply-before-checkpoint replay, A→B recovery with B→C
already available, stale-lease takeover, receipt-binding collision, and
multi-segment continuation behavior with observable state and request counts.

## 8. Use documented keys, cursors, and idempotency only

A provider primary key must be a documented stable identifier or documented
composite identity. A content/version hash is not a primary key or cursor merely
because it changes. Use concurrency tokens, cursors, and watermarks only for the
role the provider documents; otherwise choose an explicit full-refresh strategy
or retain the missing incremental capability.

Never invent an idempotency header, provider replay guarantee, or retry-safe
claim. The engine's internal write idempotency key is local correlation, not
provider-side idempotency.

Absence of provider idempotency does not by itself block the first approved
write. Such an action may remain usable when the runtime demonstrably issues one
request per row, disables automatic retry and implicit replay, consumes approval
before provider I/O, captures the bounded provider response, and reports partial
failure. It is then explicitly **single-attempt**, not retry-safe. An ambiguous
result must not be automatically replayed; recovery requires operator review or
a source-backed reconciliation/read-back strategy.

Use `batchable: false` only when one warehouse row cannot deterministically map
to one bounded request or the bulk runtime cannot execute that action. Do not
use it merely because provider idempotency, a read-back endpoint, or scheduled
automation is absent. Provider-documented idempotent delete and missing-ok
semantics may support their declared retry policy, but must be cited.

## 9. Add managed transport only after saved execution works

Do not confuse generic saved reverse-ETL execution with an optional managed
destination transport. Prove the plan -> preview -> approval -> run route for
each action first, against an isolated warehouse and capture provider where
possible.

A source transport must bind executor references and stream allowlists to exact
modes, scope/resource evidence, continuation, checkpoints, ordering, delete
semantics, and durable acknowledgement. A managed destination transport must
name exact eligible actions and mode-to-apply bindings, enforce one row to one
bounded request, and state acknowledgement, partial-failure, ordering,
idempotency, retry, and ambiguous-result behavior.

An action can be available for single-attempt saved reverse ETL while still
lacking unattended retry/reconciliation semantics required by a stronger
managed transport. Record that distinction rather than advertising an empty
transport or suppressing the already executable action. Any unsupported managed
mode must fail deterministically before provider I/O.

## 10. Prove usable surface at the credential boundary

A command counts as usable only when the built binary or actual App route reaches
the credential boundary. Declaration reconciliation is necessary but not enough.
For direct commands, exercise command resolution, record construction, preview,
approval, and the real engine request. For saved ETL/reverse ETL, use an isolated
local warehouse and prove plan, preview, approval, execution, checkpoint, and
request-count behavior appropriate to the lane.

Reconcile every locked operation into exactly one accounted state: implemented
lane, planned mapping with a named blocker, or unsupported with source evidence.
Classify saved targets and interactive commands independently: one source
operation may truthfully support ETL and direct read, or reverse ETL and direct
write, without becoming two provider operations. Count aliases separately from
provider operations so a CLI alias never inflates source coverage. Record exact
commands, source-lock identifiers, projected actions/streams,
credential-boundary results, and remaining gaps in delivery evidence.

Use explicit timeouts for changed suites and leave memory-heavy aggregate tests
to CI. Judge merge readiness by usable surface and truthful remaining gaps, not
by source provenance or generated-file counts alone.
