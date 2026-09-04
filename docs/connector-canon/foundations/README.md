# Connector Foundation Atlas

The Foundation Atlas is a positive, authoring-only inventory of shared runtime
capabilities and closed connector extension seams. It answers whether a real
encoder, executor, warehouse contract, or composition path already exists.

- [`catalog.json`](catalog.json) is the current inventory.
- [`catalog.schema.json`](catalog.schema.json) is its closed schema.
- The [source.lock vNext architecture](../SOURCE-LOCK-VNEXT.md) owns connector
  authoring and the runtime boundary.

The Atlas is never loaded by the CLI. It does not contain provider operations,
does not grant command availability, and cannot suppress execution. Provider
facts belong only in the connector's schema-4 source lock; runtime capability
comes only from rendered execution JSON plus an actual registered path.

## Mandatory discovery

Before proposing shared runtime code:

1. State the exact required lane, protocol, request/response shape, mode,
   persistence, acknowledgement, and safety contract.
2. Search `catalog.json` by lane, layer, protocol, or stable ID.
3. Inspect candidate owners, symbols, constraints, selection mechanism, and
   proof tests in the current tree.
4. Classify the need as `reuse`, `constrained_extension`, or `actual_gap`.
5. Record the chosen ID and classification in the issue plan.

`reuse` selects an existing contract. `constrained_extension` widens the named
owner without creating a parallel route. `actual_gap` means no current owner or
declared seam can satisfy the contract; report the exact mismatch and obtain
captain approval before implementing a genuine new shared foundation.

A connector-specific provider behavior selects a closed connector-owned
adapter through an existing definition reference. Shared runtime must not
branch on connector name. Do not create a generic HTTP, SQL, webhook, or binary
escape hatch to fill a connector-local gap.

## Maintenance

Update the matching entry in the same change when a shared contract, selection,
supported shape/mode, constraint, owner, symbol, proof test, status, or
replacement changes. Add an entry only for a real shared owner or closed
connector-specific adapter, not for another consumer.

`authoring.source-lock-vnext.v1` additionally records its bounded
`publication_guarantees` separately from general authoring guarantees. Every
mapping names exactly one behavior-granular claim, one registered positive
proof, and one distinct refusal proof; the Atlas validator rejects a compound,
omitted, duplicate, undeclared, or unregistered mapping. This is a
source-lock publication contract only, not a whole-catalog proof migration.

Keep stable IDs while ownership moves. Retired IDs remain only in a current
entry's `supersedes` list; their old commands, files, and procedures do not
remain in the tree.

Validate with:

```bash
jq empty docs/connector-canon/foundations/catalog.schema.json docs/connector-canon/foundations/catalog.json
jq -e '([.foundations[].id] | length) == ([.foundations[].id] | unique | length)' docs/connector-canon/foundations/catalog.json
git diff --check
```
