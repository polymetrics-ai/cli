# Connector delivery canon

**Status:** current and binding as of 2026-09-01.

This directory defines how a provider fact becomes a PM CLI connector. Current
runtime capability is established by rendered execution JSON and executable
proof, never by an evidence filename or external review state.

## Read in this order

1. [source.lock vNext architecture](SOURCE-LOCK-VNEXT.md) — the only authoring
   pipeline, schema-4 contract, canonical descriptor, seven lanes,
   deterministic renderer, runtime boundary, errors, and proof requirements.
2. [Implementation procedure](IMPLEMENTATION-PROCEDURE.md) — the required
   connector-author and migration workflow.
3. [Connector terminology](connector-terminology.md) — the shared vocabulary
   for execution operations, streams, actions, lanes, and warehouse flows.
4. [Remote reproducibility](REMOTE-REPRODUCIBILITY.md) — clean-clone checks and
   the boundary around separately authorized live-provider tests.
5. [Foundation Atlas](foundations/README.md) — shared executor inventory and the
   approval rule for a genuine missing runtime foundation.

The mechanical execution formats remain documented in
[migration conventions](../migration/conventions.md), the
[operation kernel](../architecture/connector-operation-kernel.md), and the
[sync transport definition](../sync-transport-definition.md). Where an older
document conflicts with the source-lock vNext architecture, vNext controls.

## Binding architecture

- `source.lock.json` is immutable authoring/evidence input only.
- The sole transformation is source lock → canonical per-operation descriptor
  with shared schemas → deterministic execution JSON.
- Runtime and runtime validation read execution JSON only.
- Authoring publication is connector-local and descriptor-confined: verified
  connector and generation descriptors remain no-follow for the complete
  operation; a complete deterministic generation is staged under a durable
  typed ownership marker, semantically revalidated by its recomputed content
  address and exact closed tree, then recorded in a durable prepared journal
  before the same-directory final-generation rename and atomic `CURRENT`
  selection. Recovery uses that one journal/stage state machine; pruning starts
  only after integrity proves publisher ownership and readers release their
  leases. Contended lock acquisition rechecks cancellation after successful
  nonblocking acquisition before mutation.
  This checkpoint does not materialize the checked-in corpus or add a runtime
  `CURRENT` reader.
- Every connector explicitly declares direct read, direct write, binary
  download, binary upload, ETL, reverse ETL, and sync transport as
  `implemented` or `unsupported`.
- Existing commandrunner, protocol encoders, credential/approval boundary,
  DuckDB/warehouse, and sync transport machinery execute those artifacts.
- Direct operations remain interactive. ETL and reverse ETL remain saved,
  warehouse-mediated pipelines. Sync composition always crosses the warehouse.
- Only malformed/missing execution artifacts, ambiguous binding, missing real
  executor/encoder, invalid bounded route/schema/invocation/auth/approval, or
  incompatible sync executor/mode may block runtime.
- A genuine shared runtime capability gap is reported before implementation;
  it is never filled with a provider-specific shortcut.

## Required authoring commands

```bash
go run ./cmd/connectorgen lock-render <connector>
go run ./cmd/connectorgen lock-render <connector> --check
go run ./cmd/connectorgen validate internal/connectors/defs
go build ./cmd/pm
```

There is no command that reconstructs a lock from execution JSON. Migrations
author the lock directly from provider facts, then review the first deterministic
render as a closed generation. `lock-render --check` verifies an already
published generation without writing; an intentionally unmaterialized reference
corpus is covered by the in-memory renderer parity test rather than a flat-file
fallback.

## Proof standard

For every changed connector, show deterministic render equality, execution-only
runtime inventory, provider-I/O-free discovery through the credential or
approval boundary, malformed execution-artifact rejection, correct seven-lane
surface, real encoder/executor reachability, DuckDB proofs for saved lanes, and
focused plus affected broader green tests. External provider access is optional
and requires separate authority; it is not runtime admission state.
