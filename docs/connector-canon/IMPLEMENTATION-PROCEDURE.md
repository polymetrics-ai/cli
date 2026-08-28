# Implementing a Connector End to End

This is the single required procedure for a connector change. It covers an API
or database connector, including CDC, reverse delivery, flows, schedules, live
proof, and certification. It is intentionally stricter than “the bundle
validates”: a declared surface is not an executable surface.

Read [the canon index](INDEX.md) first. The database, CDC, bidirectional, and
warehouse rulings there are binding.

## Non-negotiable delivery rules

- The warehouse always mediates movement. There is no direct source →
  destination hop. API → API means **API → warehouse → API**, even if both
  ends use the same connector.
- Use a typed connector capability. Never add generic shell, generic HTTP
  write, or generic SQL write execution to fill a gap.
- Database work uses the typed database framework. PostgreSQL is its reference
  driver; do not create a second sync-mode enum, a generic SQL executor, or a
  separate connector repository.
- All reverse delivery remains `plan → preview → approval → execute`.
- A missing shared capability is a foundation issue, not a connector-local
  workaround and not an `implemented` declaration.

## 1. Start with evidence and an issue

1. Create or link one scoped issue, then follow the repository's issue-first
   GSD lifecycle. Record the source documentation version, endpoint inventory,
   risk classification, intended streams/actions, and every claimed flow.
2. Read the connector's existing bundle, generated manual, catalog entry, and
   comparable reference bundle. Inspect the runtime-visible surface with:

   ```bash
   pm connectors inspect <connector> --json
   ```

   This inspection is metadata-only; it must not read credentials.
3. Build a source-of-truth inventory. For each provider operation, record one
   primary classification: stream, direct read, reverse action, CDC/changefeed,
   typed exclusion, or not applicable. Do not create duplicate operations for
   aliases or documentation cross-links.
4. Name the intended warehouse flows explicitly. A connector that only has a
   read definition has not delivered a reverse or bidirectional flow.

## 2. FOUNDATION CHECK — perform this before claiming “done”

Run the [Foundation Atlas discovery procedure](foundations/README.md) before
filling this section. Record the selected stable Atlas IDs and the `reuse`,
`constrained_extension`, or `actual_gap` classification in the issue plan; the
Atlas README alone owns its same-change maintenance procedure.

Before an operation, flow, or certification can be declared complete, prove
that the foundations it needs **exist and execute**. “Declared in JSON”, “has a
fixture”, and “is in a plan” are not proof.

For every intended operation and flow, make this table in the issue plan and
fill it with command/test evidence:

| Need | Required proof | If absent |
| --- | --- | --- |
| Connector operation | Its executor is supported by the real command runner, accepts the derived surface, and passes runtime preflight. | Open a foundation issue; keep the operation truthfully non-implemented or typed-excluded. |
| Read/ETL materialization | A connection can write the bounded batches, WAL, and owned warehouse table that the flow relies on. | Build the storage/runtime foundation first. |
| Reverse delivery | A typed destination action can plan, preview, obtain approval, and execute with receipt/idempotency evidence. | Build the target/action foundation; do not create a direct hop. |
| Database / CDC | The typed database contract and its protocol-specific safety rules are executable. | Build that framework capability as its own issue. No generic SQL or cursor fallback. |
| Flow / schedule | The persisted flow can use warehouse tables and its approval gate still holds when the scheduler invokes it. | Build the flow/scheduler integration first. |
| Certification | The required fixture, conformance, live environment, cleanup/receipt, and accepted artifact are all available. | Leave the connector uncertified and record the missing proof. |

Run the structural proof on every connector change:

```bash
make connector-runtime-preflight
make connector-canon-check
go run ./cmd/connectorgen validate internal/connectors/defs
go run ./cmd/connectorgen surface-sync --check
```

`connector-runtime-preflight` runs
`TestEveryImplementedCommandPassesRuntimePreflight` against the real
`commandrunner.Preflight` entry point for every registered bundle. It is the
repository-wide guard against shipping a command whose metadata says
`availability: implemented` but whose runtime refuses it. It is structural
proof, not live-provider proof.

When a foundation is absent, open/link a dedicated foundation issue, state the
missing executable contract, and stop that claim at the truthful state. Do not
hide the gap in an adapter, a raw query, a hard-coded flag, or a certification
filename.

## 3. Define the connector and derive its command surface

1. Author the definition bundle under `internal/connectors/defs/<name>/` using
   [the migration conventions](../migration/conventions.md). The bundle is the
   source for metadata, authentication shape, streams, schemas, writes,
   operations, API provenance, fixtures, and connector documentation.
2. Derive command parameters from provider operation specifications using
   `connectorgen params-import`, then derive surface metadata with
   `connectorgen surface-sync`. Do not hand-author opaque provider cursors,
   generated `maps_to` fields, output policy, or a made-up API endpoint.
3. Direct reads follow the declared paginator. Callers navigate through
   `--page` or `--page-cursor`; raw opaque provider cursors are not a second
   navigation channel. Returned page metadata must describe what reached the
   wire and must not imply completeness it cannot prove.
4. Mark an operation implemented only after the Foundation Check is green. An
   unsupported, unsafe, provider-restricted, partial, or planned operation must
   say so in the declaration and docs.
5. Keep generated output synchronized rather than hand-edited:

   ```bash
   go run ./cmd/connectorgen validate internal/connectors/defs
   go run ./cmd/connectorgen surface-sync
   pm docs generate --dir docs/cli --connectors-dir docs/connectors
   ```

## 4. Deliver the four warehouse-mediated flows

All four flows use the same durable middle: bounded extraction into the
connector's owned warehouse path, then a typed consumer from that warehouse.
The end connectors may be the same, but the warehouse step is never skipped.

| Flow | Required delivery contract |
| --- | --- |
| **API → warehouse → API** | Extract API records into the warehouse, then deliver a selected warehouse workset through a typed API action. Reverse delivery needs plan, preview, explicit approval, execution, and receipt/idempotency evidence. |
| **API → warehouse → database** | Extract API records into the warehouse, then use the typed database destination contract. No generic SQL executor may substitute for that contract. |
| **Database → warehouse → API** | Read through the typed database source into the warehouse, then use the same approved typed API reverse action. |
| **Database → warehouse → database** | Read from the typed database source into the warehouse, then deliver through the typed database destination contract. Source and destination identities remain isolated. |

For PostgreSQL CDC specifically, use the accepted contract: PostgreSQL 14+,
streamed `pgoutput` v2, bounded durable transaction staging, and a durable
receipt before acknowledgement/checkpoint advance. A stage quota failure is
fail-closed. Never substitute cursor or timestamp polling for the CDC contract.

For bidirectional changefeeds, keep one delivery contract while honestly
modeling the two producers: an inbound committed source transaction and an
outbound keyed warehouse/Parquet-DuckDB delta. Tombstones are explicit; do not
infer deletion from an absent record.

## 5. Author flows and schedules without bypasses

1. Create the connection/materialization first. A flow consumes named
   warehouse tables, not a live source-to-destination shortcut.
2. Write a flow manifest with explicit inputs and outputs. Use `pm flow plan`
   and `pm flow preview` before `pm flow run`; approval-gated action steps stay
   gated in the flow.
3. Bind a schedule only to a named persisted flow after its plan/preview
   evidence is accepted. A schedule is repetition, not an authorization to
   bypass plan, preview, approval, safety confirmation, or credential scope.
4. Test resume, retry, idempotency, and failure behavior around durable
   warehouse state. Do not use raw source records as a reverse-delivery retry
   queue.

## 6. Test in layers

1. **Definition and derivation:** schema validation, API-surface checks,
   parameter import/surface sync, and focused fixtures.
2. **Runtime preflight:** run the real preflight test above. Add a focused
   regression that would fail if this connector claimed an unexecutable
   operation.
3. **Warehouse flow:** exercise the expected record counts through extraction,
   warehouse materialization, query/workset selection, and approved target
   delivery. Exit status alone is not enough for pagination/data-loss cases.
4. **CDC or database contract:** test the protocol/receipt/acknowledgement and
   quota behavior through the typed framework. For native database live tests,
   use the reusable `native/dbtest` harness; its [maintainer guide](../../internal/connectors/native/dbtest/README.md)
   owns the explicit Docker-or-Podman runtime and direct Unix endpoint contract.
5. **Docs and surface parity:** regenerate manuals/catalog where needed, check
   `pm help`, bare namespace help, command help, and website/docs validation.

Run changed-package tests and `internal/cli` separately with `-timeout 20m`;
use CI for the full repository suite. Do not mistake a timed-out full local
suite for a passing check.

## 7. Produce live proof and certify

1. Use an approved sandbox or explicitly authorized live environment. Supply
   credentials only through environment variables or stdin; never put values in
   source, issue text, fixtures, docs, logs, or report artifacts.
2. Record the provider version/environment, intended read and write scope,
   bounded test data, cleanup result, receipt/idempotency result, and the exact
   flow(s) exercised. A live failure or missing environment is an honest
   uncertified result, not a reason to relax the test.
3. Certification is all-or-nothing across the connector's applicable surface
   and the relevant warehouse flows. A `certification.json` file is harness
   input, not certification. Fixture success is useful evidence but not live
   certification.
4. Only accepted live artifacts count. The current accepted count is zero;
   [issue #3984](https://github.com/polymetrics-ai/cli/issues/3984) owns the
   truthful capability/flow matrix needed to make that status mechanically
   visible.

## Completion checklist

- [ ] Source inventory and classifications are recorded with provenance.
- [ ] Every claimed operation and flow has passed the FOUNDATION CHECK, or a
      linked foundation issue and truthful non-implemented status exists.
- [ ] All applicable warehouse-mediated flows are demonstrated; no direct hop
      exists.
- [ ] Generated surface/docs are synchronized and runtime preflight is green.
- [ ] Focused tests, flow tests, and required docs/website checks are green.
- [ ] Live proof is recorded, or the connector remains explicitly uncertified.
- [ ] Certification claim is supported by accepted live artifacts, not names or
      fixture files.
