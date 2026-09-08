# Connector Foundation Atlas

The Foundation Atlas is a positive, authoring-only inventory of shared runtime
capabilities and closed connector extension seams. It answers whether a real
encoder, executor, warehouse contract, or composition path already exists.

- [`catalog.json`](catalog.json) is the current inventory.
- [`catalog.schema.json`](catalog.schema.json) is its closed schema.
- The [source.lock vNext architecture](../SOURCE-LOCK-VNEXT.md) owns connector
  authoring and the runtime boundary.

The `pm` runtime never loads the Atlas; authoring tools may read it. It does not contain provider operations,
does not grant command availability, and cannot suppress execution. Provider
facts are retained in immutable provider evidence and the connector's schema-4
source lock. The [retained source/lane report](../SOURCE-LANE-MANIFEST.md) accounts
for source-only archives without changing execution admission; runtime capability
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

## Source demand register and scoped assertion evidence

`connectorgen source-demands` reads the sole Atlas together with retained source
facts and sparse [cell assessments](../../../data/connector-canon/batch1-foundation-assessments.json).
It generates an authoring report under the closed
[demand register schema](demand-register.schema.json). See the
[generation and check instructions](../SOURCE-LANE-MANIFEST.md#foundation-demand-register).

The independent universe is every retained source key crossed with seven lanes.
Authored assessments and their exact unassessed complement remain separate.
Current source-owned obligations are reconciled with the independently pinned
[historical baseline](../../../data/connector-canon/batch1-foundation-obligations.json);
neither an omitted assessment nor a rewritten count removes an obligation.
Each requirement retains its source citations, examined Atlas contract occurrence,
owner, source/configuration evidence, missing proof and existing decision refs.

[Foundation assertion records](proofs.json) follow their own
[closed schema](proofs.schema.json). The authoring reader checks actual registered
owner/test declarations, selected run/pass events, original capture/output/input
pins and a source-owned reviewed assertion binding. It reads receipts as data;
it never executes their commands or adopts their historical working directories.
A top-level mode-vocabulary assertion does not establish an executor/mode
intersection. A selected refusal subtest does not prove its unselected siblings.

Missing optional proof documents, selected records or evidence files retain the
unresolved demand and its source identity, citations and next owner. The report
separates `requested_proof_ids` from observed records and records typed
`proof_issues`. Present historical pins whose bytes have changed produce
`proof_stale`; missing evidence produces `proof_unavailable`. Actual input pins
describe observed bytes, including stale files. An absent file has no invented
input pin or assertion. Resolved shared/local claims still require current proof.
Malformed evidence, unsafe paths, capacity violations, cancellation and changes
during observation refuse the complete report. Required current source and Atlas
joins remain strict, including when an optional proof refers to the same file.

The register keeps those narrow proof observations separate from resolved reuse
requirements. Atlas examples retain their literal files, occurrence hashes and
owners. Missing files or absent exact source/selector joins stay
`example_unresolved`; they do not fabricate provider identities or foundation gaps.
Adopter relations derive from actual admitted requirement/proof/binding tuples.

A current proof must fit the exact mechanism selected by its source-owned review,
not merely share an Atlas owner file. `mechanism_fits` and adopter relations name
the proof, exact binding, source citations, and pinned declaration selectors.
Structured REST body and static API-key header fits remain separate assertions.

`source_admission` preserves the actual source report's validation and complete
diagnostics. Normally its kind is `current_valid`. A narrowly verified whole
missing canonical CLI file may produce `canonical_intended_missing_cli` while
its current execution report remains invalid. The authoring reader verifies the
partial and complete generations, every supporting artifact, the full lost
command surface, exact source-reference diagnostics, and independently required
local work. Deriving this witness adds one observed replay of the actual source
producer and bounded file reads; ordinary materialized reports keep their existing
path. The reader checks file identities, required absence, and the execution
namespace before returning a register. A source report containing only intended
command references can remain valid while the CLI file is absent. Requesting a
canonical-intended fit still requires the same witness and completed local
accounting during assessment. Arbitrary invalid source reports remain refused.

Such fits say `canonical_intended`, and `intended_cli` pins staged bytes rather
than pretending the file exists. Shared reuse still requires the distinct,
verified local-configuration companion; a merely unresolved local row accounts
for work but authorizes no reuse or adopter. The original `source-lanes` report
and check retain their actual validation and exit status, including exit 1 for
invalid reports. A successful `source-demands` check validates the authoring
report, not executable connector completeness.

These authoring records never change lane applicability, accepted execution
targets, lane proof refs or runtime availability. Receiver scope remains with
`cli-batch1-vercel-inbound-sync-decision-r1`; exposure is a separate conditional
decision owned by
`cli-plan-webhook-receiver-foundation-r1-decision-webhook-exposure-product-boundary`.
Neither reference approves a receiver, hosted relay or live endpoint.

## Compiled source visibility

`connectorgen gen` projects retained source identities, seven-lane observations
and scoped citations into immutable metadata alongside the existing manifest
index. The compiled capability table validates diagnostic IDs; it does not
register executors. Source inspection decodes only the selected connector's
bounded payload and refuses malformed metadata before execution construction.
Ordinary commands retain their separate execution identity and resolver path,
even when source diagnostic metadata is unavailable or invalid. See the
[connector CLI manual](../../cli/connectors.md) for the exact tuple selector.

## Maintenance

Update the matching entry in the same change when a shared contract, selection,
supported shape/mode, constraint, owner, symbol, proof test, status, or
replacement changes. Add an entry only for a real shared owner or closed
connector-specific adapter, not for another consumer.

`authoring.source-lock-vnext.v1` additionally records its bounded
`publication_guarantees` separately from general authoring guarantees. Every
mapping names exactly one behavior-granular claim, one registered physical or
durable positive proof, and one distinct refusal proof. Do not map an
in-memory comparator to a filesystem publication guarantee. The Atlas
validator rejects a compound, omitted, duplicate, undeclared, or unregistered
mapping. This is a source-lock publication contract only, not a whole-catalog
proof migration.

Keep stable IDs while ownership moves. Retired IDs remain only in a current
entry's `supersedes` list; their old commands, files, and procedures do not
remain in the tree.

Validate with:

```bash
jq empty docs/connector-canon/foundations/catalog.schema.json docs/connector-canon/foundations/catalog.json
jq -e '([.foundations[].id] | length) == ([.foundations[].id] | unique | length)' docs/connector-canon/foundations/catalog.json
git diff --check
```
