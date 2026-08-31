# Operation evidence

`connectorgen operation-evidence` projects one deterministic evidence row for
every operation named by a connector-owned provider source lock. The checked-in
artifact is [`internal/connectors/operation-evidence.json`](../../internal/connectors/operation-evidence.json);
the executable aggregate contract is
[`internal/connectors/operation-evidence-fixed-100.json`](../../internal/connectors/operation-evidence-fixed-100.json).

Each row records the provider trace, canonical mapping, declaration-owned
runtime target, CLI command, generated website command, fixture, matching
conformance check, and every capability classification (`etl`, `reverse_etl`,
`direct_read`, `direct_write`, `binary_download`, `binary_upload`). A missing
surface is a named gap. It is never silently omitted or represented as `N/A`.

The only non-gap absence is a source lock whose provider evidence explicitly
declares `skipped` or `dynamic`; the artifact carries that reason and evidence
verbatim. An explicit declaration-only `source_reference` remains a row: its
canonical citation and operation identity are visible, while the named
`source_contract_unavailable` gap prevents its six-lane classification from
becoming enabled merely because a similarly shaped declaration exists. The
projector reads source-lock v2 and v3 inputs but does not own or alter their
parser or schema.

Run the generator after a reviewed source-lock or connector-surface change:

```bash
go run ./cmd/connectorgen operation-evidence --write-fixed-100
```

For a reviewed bounded cohort, repeat `--connector` and supply an explicit
report path. The selector reads only the named connectors' source locks and
never replaces the repository-wide artifact or its fixed-100 reference:

```bash
go run ./cmd/connectorgen operation-evidence \
  --connector asana --connector bitbucket --output data/connector-canon/batch1-operation-evidence.json
```

The scoped report is an identity-preserving source-lock-to-mapping bridge: it
records the current canonical, runtime, command, website, proof, and named-gap
cells for every selected source identity. `mapped` is not an execution claim;
only the declaration-admission and runtime-preflight path can mark a command
implemented. A source row with no current command therefore remains visible
with its citation and gap instead of being omitted. A scoped report must not be
used to replace the global fixed-100 contract.

## Source-operation multi-lane mapping admission

`connectorgen source-operation-mapping <manifest> --check` validates a
separate authoring-only manifest before it is linked from an artifact. Each
source-lock operation has one cited source row and explicit facts for
pagination, scope/path variables, media, and event/cursor evidence. Its zero
or more lane cells use one of `implemented`, `mapped_unproven`,
`missing_foundation`, or source-evidenced `not_applicable`; a pageable source
row must explicitly include an `etl` cell.

The manifest can list multiple source locks for one connector. A supplemental
source row remains independently visible with its own source ID and citation.
When two cited source rows intentionally describe one canonical operation, the
supplemental row points `canonical_operation_id` at a self-canonical source row
with the same locked protocol, method, and path. This preserves the source-row
denominator while reporting the canonical-operation denominator without
inferring equivalence from a route alone. Artifact links name only an existing
source-operation/lane cell; they cannot create source IDs.

The checker reads source locks through the mapping-only parser. It does not
retain or fetch provider bytes, change connector definitions, alter a runtime
executor, or act as certification evidence.

## Deferred source visibility

`connectorgen deferred-visibility <cohort-manifest> --check [--json]` reads
the frozen cohort, the connector-local source-lane matrices, existing
connector-local missing-foundation ledgers, and the Foundation Atlas. It emits
one deterministic report entry for every `mapped_unproven` or
`missing_foundation` lane cell. Each entry preserves both the matrix source
identity and the exact operation identity in the cited immutable source lock,
plus method, path, citation, source fact, typed reason, and Atlas capability.

The report is discovery/preflight evidence only. It has no executable command,
operation, stream, write, transport, credential, descriptor, or executor
binding (opaque provider source facts may retain similarly named evidence
keys), and it does not call source import, source materialization, or a
provider. A source identity that differs from the lock operation key is
accepted only when the matrix explicitly declares the provider operation ID
and the locked method and path resolve uniquely; zero or multiple matches fail
closed.

For Batch R1 this distinguishes the 4,341 primary source operations from the
4,343 source-lane matrix rows (two cited supplemental rows) and their 30,401
seven-lane cells. It is not an execution-parity report: a later declaration,
public command/engine preflight, and appropriate warehouse or transport proof
are still required before a cell can become executable.

`make verify` runs `connectorgen-operation-evidence`, equivalent to:

```bash
go run ./cmd/connectorgen operation-evidence --check
```

The check rejects artifact drift and validates each source-locked expectation in
the fixed cohort. The cohort is a deterministic, capability-stratified sample
of 100 complete provider operations, so a regression in any selected operation
fails with its source ID rather than being hidden by an aggregate count.
