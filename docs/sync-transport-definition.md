# Declaring a truthful sync transport

`sync_transport.json` is an optional, versioned claim that a connector can
participate in the closed warehouse-mediated sync path. It is not a way to
turn an arbitrary provider endpoint, SQL statement, or HTTP request into a
transport. The engine loads it strictly; runtime preflight additionally
requires a registered exact executor and accepted conformance evidence before
any source or destination I/O.

Place the file only at `internal/connectors/defs/<connector>/sync_transport.json`.
Use `schema_version: 1`. The complete structural schema is
`internal/connectors/engine/schema/sync_transport.schema.json`; run `go run
./cmd/connectorgen validate` after editing.

## Source declaration

A `source_transport` must supply all of the following:

- `executor`: an exact closed `{family, id}` reference compatible with the
  connector's `metadata.json.integration_type`.
- `eligible_streams`: a non-empty positive allowlist. `"*"` is permitted only
  as the sole entry where the source adapter can truthfully enumerate dynamic
  streams.
- `modes`: the supported closed `synccontract` modes.
- `delivery`: explicit `idempotency`, `ordering`, and `deletes` guarantees.
- `conformance`: its own `{suite, run_id}` evidence reference.

The evidence reference belongs to this definition. Do not copy an evidence
constant from another connector. Production composition collects the evidence
values selected by every definition using the same executor and admits only
those exact values; it never selects an adapter by connector name.

`ordered_pipeline` is optional and may be true only when the source's exact
executor supports the bounded ordered pipeline contract.

For the current `declarative_api/declarative_stream_source` adapter, use a
concrete allowlist matching every executable stream in `streams.json`; that
adapter does not implement dynamic-stream enumeration, so it refuses `"*"`.

## Destination declaration

A `destination_transport` must supply the same `executor`, `modes`,
`delivery`, and definition-owned `conformance` evidence, plus:

- `eligible_actions`: the non-empty allowlist of action names owned by that
  connector definition.
- `acknowledgement`: `durable_warehouse` for an executable route. `none` may
  be inspected but fails runtime preflight.
- `apply_strategies`: one or more `{mode, strategy, action}` entries for every
  declared mode. The `(mode, action)` pair must be unique and the action must
  be in `eligible_actions`; the strategy is one of the closed transport
  strategies, never a caller-selected operation. Multiple actions may share a
  mode only when each is independently selectable through a persisted stream
  `destination_action`; leaving that field empty is rejected when a mode has
  more than one declared action.
- `source_bindings` when the destination only accepts particular sources. Each
  binding identifies an exact source executor, a positive stream allowlist,
  one exact destination `action`, a bounded `batch`, and either a closed
  ordinary record mapping (`config_match` or `input_fields`) or an explicit
  tombstone mapping. A binding cannot carry both mapping forms.

A destination cannot declare `change_capture`: capture is source-only into the
connection warehouse. A declaration with a destination `change_capture` mode,
a missing strategy, an unlisted action, a duplicate `(mode, action)` pair, a
malformed source binding, or an incompatible executor family fails before
source I/O.

## `declarative_api/declarative_typed_destination`

`declarative_api/declarative_typed_destination` is the reusable destination
adapter for a declarative API connector. It is intentionally narrower than a
generic HTTP writer:

- Declare only action names that already exist in that connector's
  `writes.json`; every declared record-driven action needs an explicit
  reverse-ETL eligibility disposition. An eligible action appears in
  `eligible_actions` and an exact `apply_strategies` entry; an action excluded
  for semantics (for example, it is not record-driven or cannot acknowledge a
  workset) stays directly CLI-reachable only through its own declared command,
  never by silently disappearing behind a safety label.
- The selected action must be an ordinary, schema-backed `writes.json` action
  with a concrete method and path. It must not carry `transport_binding`:
  that field selects a different, specialized closed adapter such as the
  issue-label destination.
- Supply one action-owned ordinary `source_bindings` entry for every selected
  apply action, all with `input_fields`. Its
  `input` is an action record field and its `field` is a source record field.
  The input name is matched byte-for-byte against the selected action's
  top-level `record_schema.properties`: both `target_id`, a provider's
  `targetId`, and any other provider-owned property spelling are valid only
  when that exact property exists. The runtime never trims, rewrites, or
  normalizes property or action names. Every required top-level selected-action
  property needs exactly one mapping. Empty, duplicate, unknown, cross-action,
  and undeclared names fail before source or provider I/O. A provider property
  that the declaration model cannot represent is a foundation gap, never a
  silently renamed field. A source executor/stream may serve several actions,
  but each action needs its own binding: a mapping never falls back from one
  action to another.
  The adapter copies only those declared values and validates every result
  against the selected action's `record_schema` before it constructs a
  provider request.
- The declarative typed destination has no run-scoped full-overwrite protocol.
  A `full_overwrite` destination declaration is therefore rejected during
  preflight before any source, stage, or provider I/O.
- Declare `acknowledgement: "durable_warehouse"` and
  `delivery.idempotency: "keyed"`. Provider idempotency binds the durable
  workset occurrence, selected action, and record identity: reopening one
  uncommitted workset reuses its key, while otherwise-equal records in another
  workset use a different key. Each action binding declares
  `batch: {"disposition":"per_record","max_records":N}` and an
  action-owned read-back policy. Before provider I/O, the runtime validates
  both ordinary and tombstone policies and clamps a caller's `--batch-size` to
  every selected acknowledgement/read-back and encoded private-receipt bound;
  callers may request fewer source rows but can never enlarge the declared
  acknowledgement unit. The adapter sends one declaration-owned request per
  record and retains the full ordered provider receipt for every 2xx/4xx/5xx
  attempt. The canonical private receipt limit is 8 KiB, including escaped
  provider output and the composite durable acknowledgement envelope.
- `delivery.deletes: "not_available"` requires every strategy to omit a
  tombstone action. `delivery.deletes: "tombstone"` requires every selected
  strategy to name a distinct `tombstone_action`, whose binding has
  `tombstone_mapping` with `image: "key"` or `"before"`, and a
  `tombstone_read_back` policy. The named action must be a `writes.json`
  `kind: "delete"` action with its own stable idempotency header. Tombstones
  never use the ordinary record mapping and cannot be sent to create/update
  actions. Every ordinary action and paired tombstone delete is listed and
  digest-bound in plan and preview before approval. A declared missing-ok
  tombstone outcome is an independently proven absent identity, not a replayed
  mutation; provider read-back proves that absence before the source checkpoint
  advances.
- Record conformance evidence from the declaring connector. Composition
  gathers all exact evidence references for this executor and registers one
  shared factory; evidence from a different connector is not interchangeable.

### Persisted action selection and application dispatch

`destination_action` lives on the saved connection stream, not in
`sync_transport.json` and not on `pm etl run`. It is the stable,
definition-owned identity that selects one eligible action when a connector
declares more than one typed destination action for a mode. A connection create
or update validates the exact descriptor, source allowlist, `input_fields`
mapping, mode/strategy, action, acknowledgement, and conformance evidence
before catalog, source, warehouse, or provider I/O.

The application/CLI lifecycle is deliberately generic only at the dispatch
layer:

```text
pm etl transport declarative-typed-destination plan --connection <name> --stream <stream>
pm etl transport declarative-typed-destination preview <plan-id>
pm etl run --connection <name> --stream <stream> --batch-size <n> \
  --approval-plan <plan-id> --approval-token-stdin --confirm destructive
```

Those commands carry no connector, action, route, verb, body, mapping, batch,
or evidence override. They resolve the saved stream, seal its selected action,
ordinary/tombstone mappings, batch disposition, read-back policy and
definition binding, stage/reopen the connection-owned workset, issue fresh
typed evidence for each acknowledged unit, and invoke the already named
`writes.json` action. Certification can therefore prove dispatch, plan/preview
binding, approval, workset ownership, durable acknowledgement, and read-back
without granting a tool that can execute an arbitrary provider operation.

### Persisted result output and credential boundary

After a provider-successful `declarative_typed_destination` apply, the
persisted App run and `pm etl run --json` / `pm etl status --json` expose
`run.destination_results`. Each element is the complete result of the already
selected named `writes.json` action: record accounting and every successful
provider response's status, headers, and body. Ordinary response fields are
never omitted because they are rare, destructive, paid-tier-specific,
unfamiliar, or outside a summary list. JSON bodies retain their original
structure; text and binary bodies retain an explicit encoding.

Concrete configured credential material is masked wherever it occurs in the
public/persisted provider result; provider-owned field names and ordinary
values are otherwise preserved. System-generated plans, logs, request
diagnostics, and synthetic errors remain secret-taint-safe. If a later local
receipt, acknowledgement, composition, or output step fails before checkpoint,
the failed uncheckpointed run still retains the ordered sanitized provider
evidence. No route, request body, action selector, or credential is accepted
from a runtime caller to influence this output. The regular persisted
reverse-ETL run uses the same `destination_result` contract, and an
operation-direct-write run preserves the same provider response facts. Its
declared `output_policy` may select a parsing form, but it does not suppress a
successful ordinary provider response.

When the provider effect, read-back, and checkpoint are durable but bounded
receipt retirement or a declaration-owned approval marker cannot be finalized,
the App persists terminal run status `delivered_reconciliation_required` with
`delivery_reconciliation`. The record retains its exact checkpoint and
`destination_results`; the nonzero CLI result still presents that exact
`ETLRun`. A repeated saved connection/stream invocation repairs only the named
local stage or plan marker before endpoint resolution. It never replays a
source read, destination action, or arbitrary route. Missing, corrupt, or
conflicting reconciliation evidence remains a typed terminal refusal.

At plan time the adapter verifies the declared mode, persisted selected action,
and source binding, then requires the existing reverse-ETL plan, preview,
approval and per-unit authorization. At apply time it accepts only a reopened
connection-owned workset, validates the mapped records before I/O, and calls
the declaration-selected action. A successful action response for every
record, plus the keyed replay guarantee, supplies the durable acknowledgement
used by the checkpoint contract. Its read-back verifies the acknowledgement,
the declared route, and the reopened workset before checkpoint commit; an
adapter that needs provider-state read-back must use a separately named closed
executor.

## Typed action contract

A destination action is a connector-owned, named action, not a generic HTTP
or SQL writer. For the action to be executable, its `writes.json` entry must
be present in `eligible_actions` and satisfy the exact closed adapter's
contract. The `source_bindings` mapping supplies only the upstream values; it
cannot add action fields, an operation name, an endpoint, or a body template.

The adapter must validate the declaration-selected strategy and source binding
at planning time, operate on a reopened owned warehouse receipt, perform the
approved named action, return a durable acknowledgement, and perform the
read-back promised by its exact adapter before the checkpoint advances.
Reverse-ETL approval remains plan, preview, approval, then execution; a
declaration never bypasses that lifecycle.

## Minimal review checklist

1. Confirm the executor has a production `DefinitionFactory` for the exact
   family and ID; an unknown executor is a fail-closed declaration, not a
   request to add a generic writer.
2. Record new conformance evidence for this connector and put its own suite
   and run ID in the declaration.
3. List only streams, actions, source bindings, delivery facts, and modes that
   the adapter can execute and read back.
4. Give every eligible action/mode pair one closed apply strategy. Where a
   mode has multiple actions, prove each is independently selected by a
   persisted `destination_action` and that cross-connector selection fails
   before I/O.
5. Add happy, refusal-before-I/O, and recovery/cleanup tests, then run
   `connectorgen validate` and `surface-sync --check`.
