# Connector Foundation Atlas

The Foundation Atlas is the CLI's positive, authoring-only inventory of shared
connector foundations.
It answers whether a capability already exists, where its contract is owned,
how definitions select it, what it supports, and which tests prove it.
It does not inventory provider operations or replace connector-specific source
locks, definitions, or truthful gap records.

The files in this directory have separate roles:

- [`catalog.json`](catalog.json) records the current foundation inventory.
- [`catalog.schema.json`](catalog.schema.json) is the closed catalog shape.
- [`BATCH-R1-DEMAND-REGISTER.md`](BATCH-R1-DEMAND-REGISTER.md) reconciles the
  frozen Batch R1 deferred evidence with existing and planned Atlas entries.
  It is issue-scoped authoring evidence, not a runtime input or execution
  declaration.
- This README is the sole owner of the discovery and maintenance procedure.

## Authority boundaries

| Concern | Authority |
| --- | --- |
| Provider operations, fields, media types, identifiers, cursors, and documented semantics | The connector's retained source lock and its citations |
| Discovery of reusable CLI foundations and their extension seams | This Atlas |
| Runtime executability of a declared command or transport | The real runtime preflight and executor path |
| Conformance and live-provider claims | The conformance and certification systems |

The Atlas is not loaded by the CLI.
It does not participate in source mapping, declaration admission, runtime
preflight, transport admission, or certification admission, and it cannot block
any of them.
An absent, incomplete, or stale Atlas entry is a documentation defect, not a
reason to suppress source-backed mapping.

## Generic and connector-specific ownership

Provider facts, source-backed mapping, and connector-specific executor
selection belong under `internal/connectors/defs/<connector>/`.
Reusable engine, transport, warehouse, reverse-delivery, flow, and
certification foundations must not branch on connector name.
When provider behavior requires runtime code, keep it in a connector-owned
extension and select it through an exact, closed definition reference rather
than an arbitrary hook.

Catalog reusable owners as `generic` or `extension_seam` and a closed
connector-owned adapter as `connector_specific_reference`.
The latter records how a connector definition selects an implementation; it
does not turn provider policy into shared runtime policy or make the Atlas a
runtime registry.

## Mandatory discovery procedure

Use this procedure before proposing shared foundation code and before creating
or updating a connector's `missing-foundation.json`.

1. State the required contract precisely: lane, protocol or source form,
   request and response shape, sync mode, persistence or acknowledgement
   guarantee, and any provider-specific fact that must remain source-cited.
2. Search [`catalog.json`](catalog.json) by layer, lane, supported contract, or
   stable Atlas ID.
3. Inspect every candidate's owner files and symbols, selection mechanism,
   constraints, non-goals, and exact proof tests in the current tree.
4. Classify the requirement as `reuse`, `constrained_extension`, or
   `actual_gap`.
5. Record the selected Atlas ID and classification in the issue plan or other
   delivery evidence before implementation starts.

The classifications mean:

- `reuse`: the shared contract already covers the requirement.
  Select it through its documented definition field, interface, command, or
  composition path and do not create another foundation.
- `constrained_extension`: an existing declared seam is the correct owner, but
  its shared contract must be widened without bypassing its constraints.
  Extend that owner and its proof rather than creating a parallel engine.
- `actual_gap`: no listed owner or extension seam can satisfy the required
  contract.
  Record the exact mismatch, add or update a `planned` Atlas entry with a stable
  ID, and obtain the captain's approval before implementing a genuinely new
  shared runtime foundation.

For `missing-foundation.json`, record the exact unsupported contract or
mismatch and its Atlas ID.
Do not copy provider operations, schemas, or citations into the Atlas or the gap
record, and do not call a capability missing merely because an importer,
projection, preflight, or certification label mentions a foundation.
Those labels must resolve to a real owner contract or remain a precise gap.

## Same-change maintenance rule

Update the matching Atlas entry in the same change whenever a shared
foundation's contract, selection mechanism, supported protocol, source form,
request or response shape, sync mode, constraint, non-goal, owner file or
symbol, proof test, status, or replacement relationship changes.
Add a new entry only for a new shared contract owner or a connector-specific
runtime adapter selected by a closed definition reference, not for another
connector that merely consumes an existing foundation.
Keep stable IDs unchanged while ownership moves; when a contract is replaced,
retire its entry and link the replacement through `supersedes` rather than
reusing the old ID.

A connector-specific source-lock refresh, mapping, definition, or new consumer
does not require a catalog change when the shared foundation contract and the
representative consumer examples remain accurate.
Update a consumer example only when the catalog would otherwise point at a
stale or misleading integration.

## Reading the catalog

Find foundations for one lane:

```bash
jq -r '.foundations[] | select(.supported_contracts.lanes | index("direct_write")) | [.id, .status, .kind] | @tsv' docs/connector-canon/foundations/catalog.json
```

Inspect one stable ID:

```bash
jq --arg id 'runtime.direct-execution.v1' '.foundations[] | select(.id == $id)' docs/connector-canon/foundations/catalog.json
```

An empty `known_gap_refs` array means no shared catalog-level gap is recorded
for that foundation.
It is not a claim that every connector mapping or provider contract is
complete.

## Validation

The Atlas deliberately has no runtime loader, generated view, script, or CI
admission gate.
Review catalog changes against [`catalog.schema.json`](catalog.schema.json),
verify referenced files, symbols, and tests in the current tree, and run:

```bash
jq empty docs/connector-canon/foundations/catalog.schema.json docs/connector-canon/foundations/catalog.json
jq -e '([.foundations[].id] | length) == ([.foundations[].id] | unique | length)' docs/connector-canon/foundations/catalog.json
git diff --check
```

Runtime preflight and certification remain required where their own procedures
apply; Atlas validation does not substitute for either one.
