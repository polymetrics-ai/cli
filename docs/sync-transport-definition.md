# Declaring a truthful sync transport

`sync_transport.json` is optional runtime execution metadata. It is one file in
the deterministic execution bundle rendered from `source.lock.json`; it is not
an authoring input and the runtime never reads the source lock. See
[`connector-canon/SOURCE-LOCK-VNEXT.md`](connector-canon/SOURCE-LOCK-VNEXT.md)
for the complete source-lock-to-execution pipeline.

The file declares only transport behavior that an existing, registered
executor can perform. It cannot turn a provider operation, SQL statement,
shell command, or arbitrary HTTP request into a transport. Runtime preflight
rejects a malformed definition, an unknown or incompatible executor, an
unsupported mode, an ambiguous binding, or an acknowledgement contract the
selected executor cannot satisfy.

Place the rendered file at
`internal/connectors/defs/<connector>/sync_transport.json` with
`schema_version: 1`. Its structural schema is
`internal/connectors/engine/schema/sync_transport.schema.json`.

## Source transport

A `source_transport` declares:

- `executor`: an exact closed `{family, id}` reference compatible with the
  connector integration type;
- `eligible_streams`: a non-empty allowlist, or `"*"` only when that executor
  can enumerate dynamic streams;
- `modes`: supported closed `synccontract` modes;
- `delivery`: explicit idempotency, ordering, and delete behavior; and
- optional `ordered_pipeline: true` only when the named executor implements
  the bounded ordered-pipeline contract.

For `declarative_api/declarative_stream_source`, list every executable stream
explicitly. That executor does not implement dynamic enumeration and rejects
`"*"`.

## Destination transport

A `destination_transport` declares:

- an exact `executor`, supported `modes`, and delivery behavior;
- a non-empty `eligible_actions` allowlist;
- `acknowledgement: "durable_warehouse"` for an executable route;
- one `apply_strategies` entry per supported mode/action pair; and
- `source_bindings` when the destination accepts only bounded source
  executor/stream combinations.

Each source binding chooses one destination action, a bounded batch, and
either an ordinary record mapping or a tombstone mapping. It cannot carry
both. A destination cannot declare `change_capture`; capture is source-only
into the connection warehouse.

The runtime rejects duplicate or absent mode/action strategies, unknown
actions, ambiguous source bindings, incompatible executor families, invalid
batch bounds, and missing durable acknowledgement before source or provider
I/O.

## Declarative typed destination

`declarative_api/declarative_typed_destination` is the closed reusable
destination for schema-backed API writes. It accepts only actions already
rendered in `writes.json` and listed in `eligible_actions`.

- Every selected action must have a concrete method, path, and closed record
  schema. It cannot carry another specialized `transport_binding`.
- Each action needs its own `source_bindings` entry. `input_fields.input` must
  exactly match a top-level action schema property and `input_fields.field`
  must exactly match a source record field. The runtime does not normalize
  either name.
- Every required action property needs exactly one mapping. Empty, duplicate,
  unknown, cross-action, and incomplete mappings fail before I/O.
- The executor has no run-scoped full-overwrite protocol, so it rejects
  `full_overwrite` during preflight.
- Declare `acknowledgement: "durable_warehouse"` and keyed idempotency. Each
  binding declares a bounded per-record batch and read-back policy.
- Tombstones require a distinct delete action, a tombstone mapping using the
  key or before image, and a tombstone read-back policy. Ordinary mappings do
  not silently serve tombstones.

The executor maps only the declared fields, validates each result against the
selected action schema, and calls the already rendered write route. A source
executor/stream may feed several actions, but each action has an independent,
unambiguous binding.

## Persisted selection, approval, and results

`destination_action` is stored on the connection stream. A create or update
validates the selected action, source allowlist, mappings, mode/strategy,
acknowledgement, and executor before catalog, source, warehouse, or provider
I/O.

The command lifecycle remains:

```text
pm etl transport declarative-typed-destination plan --connection <name> --stream <stream>
pm etl transport declarative-typed-destination preview <plan-id>
pm etl run --connection <name> --stream <stream> --batch-size <n> \
  --approval-plan <plan-id> --approval-token-stdin --confirm destructive
```

Those commands cannot override connector, action, route, verb, body, mapping,
batch contract, or executor. They resolve the saved connection and stream,
reopen its owned warehouse workset, enforce approval and authorization, and
invoke the selected action.

Successful applies persist `run.destination_results`, including record
accounting and bounded provider response facts. Runtime-generated plans,
diagnostics, and errors remain secret-taint-safe. When provider effect,
read-back, and checkpoint are durable but local receipt finalization fails,
the run enters `delivered_reconciliation_required`; recovery repairs only the
named local stage and never replays provider I/O.

## Author and review checklist

1. Author the `sync_transport` lane in `source.lock.json`; do not hand-author
   the execution file.
2. Name an existing registered executor and only modes, streams, actions,
   mappings, batches, and acknowledgement behavior it actually supports.
3. Render with `connectorgen lock-render`, then verify byte-for-byte stability
   with `connectorgen lock-render --check`.
4. Test happy execution, malformed/ambiguous definitions, incompatible modes,
   missing executors, refusal before I/O, approval/auth boundaries, durable
   acknowledgement, and recovery.
5. Run `connectorgen validate` and the relevant sync/transport/engine/CLI
   suites. A missing genuine executor is a foundation gap: preserve the source
   mapping and report the exact gap instead of inventing a provider-specific
   runtime route.
