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
- `apply_strategies`: exactly one `{mode, strategy, action}` entry for every
  declared mode. The action must be in `eligible_actions`; the strategy is one
  of the closed transport strategies, never a caller-selected operation.
- `source_bindings` when the destination only accepts particular sources. Each
  binding identifies an exact source executor, a positive stream allowlist,
  and a closed record mapping (`config_match` or `input_fields`).

A destination cannot declare `change_capture`: capture is source-only into the
connection warehouse. A declaration with a destination `change_capture` mode,
a missing strategy, an unlisted action, a duplicate mode, a malformed source
binding, or an incompatible executor family fails before source I/O.

## `declarative_api/declarative_typed_destination`

`declarative_api/declarative_typed_destination` is the reusable destination
adapter for a declarative API connector. It is intentionally narrower than a
generic HTTP writer:

- Declare only action names that already exist in that connector's
  `writes.json`; each declared mode selects one of those exact names through
  `apply_strategies`.
- The selected action must be an ordinary, schema-backed `writes.json` action
  with a concrete method and path. It must not carry `transport_binding`:
  that field selects a different, specialized closed adapter such as the
  issue-label destination.
- Supply at least one `source_bindings` entry, all with `input_fields`. Its
  `input` is an action record field and its `field` is a source record field.
  The adapter copies only those declared values and validates every result
  against the selected action's `record_schema` before it constructs a
  provider request.
- Declare `acknowledgement: "durable_warehouse"`,
  `delivery.idempotency: "keyed"`, and
  `delivery.deletes: "not_available"`. This adapter does not accept
  tombstones, so a connector must not claim delete propagation it cannot
  perform.
- Record conformance evidence from the declaring connector. Composition
  gathers all exact evidence references for this executor and registers one
  shared factory; evidence from a different connector is not interchangeable.

At plan time the adapter verifies the declared mode, action and source
binding, then requires the existing reverse-ETL plan, preview, approval and
per-unit authorization. At apply time it accepts only a reopened
connection-owned workset, validates the mapped records before I/O, and calls
the declaration-selected action. A successful action response for every
record, plus the keyed replay guarantee, supplies the durable acknowledgement
used by the checkpoint contract. Its read-back verifies the acknowledgement,
the declared route, and the reopened workset before checkpoint commit; an
adapter that needs provider-state read-back must use a separately named
closed executor.

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
4. Give every mode one closed apply strategy and verify the action's typed
   binding and acknowledgement policy.
5. Add happy, refusal-before-I/O, and recovery/cleanup tests, then run
   `connectorgen validate` and `surface-sync --check`.
