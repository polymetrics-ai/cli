# Plan — Issue #3978: final PostgreSQL certification and publication

## Goal

Make the published PostgreSQL capability projection agree with current, executable, definition-owned certification evidence: six warehouse-mediated polling/managed-target modes and a source-only pgoutput CDC route. `query` remains false. This is a certification/publication repair, not a new transport implementation.

## TDD slices

1. **Red — expose the mismatch.** The current matrix had twelve live mode cells but published `write=false` and evidence-free `cdc=true`. Add tests that fail on the missing capability proof and prove that a `write=true` declaration alone cannot pass certification.
2. **Green — definition-owned certification projection.** Change only certification/publication interpretation necessary to bind a declared native-database destination transport to one capability-scoped aggregate live proof, while separately requiring accepted exact-mode evidence for all declared destination modes. Never bind a `sync_mode` record into a capability cell. Bind the source-only declared CDC route to its own accepted binary/live proof. Do not change the direct `Connector.Write` operation or introduce a generic SQL/direct write API.
3. **Green — complete mode result.** Represent the six executable target `synccontract.Mode` outcomes and source-only `change_capture`; retain destination/API exclusions as concrete non-pass reasons. `incremental_dedupe_history` remains executable. PostgreSQL CDC→API is not declared, so it is N/A/deferred rather than silently skipped or passed.
4. **Live certification.** Run the current-SHA built binary against PostgreSQL through the explicit Docker socket. Capture only redacted proof-bearing records from independently asserted live runs, regenerate the PostgreSQL shard, run the global shard drift gate, and prove website documentation output deterministic with two `gen:docs` runs.
5. **Failure control.** After schema compilation accepts the scratch bundle, make a safe local runtime-only declaration invalid, observe certification failure, restore exactly, then rerun green.
6. **Verify and review.** Exercise the four existing binary warehouse-flow proofs, current database CDC proof, targeted consumers, generated docs/help/website parity, repository gates, GSD verification, and inline code review. Record all commands and divergences in the PR.

## Publication boundary

- `write=true` means the declared `postgres_polling_watermark → postgres_managed_target` warehouse-mediated destination transport has passed its exact live certification for every declared mode. It does **not** mean `Connector.Write` became a generic direct writer.
- `cdc=true` means PostgreSQL 14+ pgoutput change capture is executable only as a database source into the connection-owned warehouse, with durable staging, receipt, checkpoint, then LSN acknowledgement. It does **not** mean CDC-to-API or any destination `change_capture` route is executable.
- `query=false` stays a concrete non-support claim.

## CLI/help/docs/website parity

The command surface is unchanged. The plan will still verify `pm help connectors`, `pm connectors`, and `pm connectors certify --help`; regenerate any connector/manual/website output the existing generators identify; and state an explicit not-applicable result for new flags, namespaces, completion, or syntax.
