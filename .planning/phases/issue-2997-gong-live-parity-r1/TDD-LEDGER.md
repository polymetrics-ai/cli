# TDD ledger: Gong release-0.3.0 live parity reconciliation

## Red → green slices

| Slice | Red evidence | Green evidence | Refactor boundary |
| --- | --- | --- | --- |
| Current source inventory | Fresh official OpenAPI inventory had not been compared to the Batch 2/3 source lock. | Credential-free fetch proves all 69 semantic operation rows match the lock exactly. | Keep the captured raw source lock immutable because the provider serializes equivalent JSON differently per request. |
| Foundation reconciliation | Preserved branch predates current main, typed destinations, and Batch 2/3 source disposition evidence. | Merge ancestry proves retained branch history plus the exact published foundation heads. | Resolve connector-owned declarations, not provider-named engine conditions. |
| Direct-read exact endpoint binding | Reproduce any Gong command that preflights with an implemented operation but no matching `api_surface` row. | Real `commandrunner.Preflight`, `surface-reconcile`, and a built CLI preflight sweep accept each declared direct read up to missing credentials. | Let `surface-sync` derive operation-owned metadata; do not hand-author it. |
| Typed write and reverse-ETL declarations | Connector-local transport/certification tests initially fail for absent or inconsistent source/destination declarations. | Merged shared foundation admits only named Gong actions, required field bindings, keyed acknowledgements, and approval-gated plans. | No generic writer, raw body, arbitrary endpoint, or Gong-specific shared branch. |
| Six-surface enabled parity | A source-locked operation can be structurally declared yet be absent from CLI/App dispatch, generated docs, or a supported ETL, reverse-ETL, direct, or binary path. | Generated inventory-to-surface evidence and built-CLI/App checks classify each of ETL, reverse ETL, direct read, direct write, binary download, and binary upload as proven or exact-source `not_applicable`. | Safety, scope, tier, and destructive metadata can add confirmation; they cannot disable a provider-defined operation. |
| Certification evidence | No credential reference means live stages cannot assert persisted provider state. | Credential-free gates are green and the remaining external block is explicit and secret-free. | Do not substitute browser authentication or fixtures for live certification. |

## Recorded red evidence

- Source-lock import and Batch 2/3 declarations are absent from the preserved branch; current
  `origin/main` also does not contain the Batch 2/3 source-lock files.
- The historical branch's phase records 67 operations. The current official OpenAPI has 69,
  confirming that historical completion evidence is insufficient for this release certification.
- Direct-read runtime coverage must be re-proven after reconciliation because prior audits found
  declaration rows that validated structurally but lacked exact executable `api_surface` bindings.
- Gong conformance reproduced unresolved meeting placeholders and a fixture schema violation;
  declaration-relative meeting and multipart paths plus the schema fixture were corrected and
  locked by the focused Gong definition test. Multipart fixture replay remains outside this slice:
  the generic fixture approval helper cannot bind the payload digests required before a multipart
  request. Firstmate accepted that as `cli-closed-operation-runtime-r1` F2/F4 work.

## Green evidence to record during execution

- exact inventory diff result, generated source-map result, and source/disposition arithmetic;
- focused Gong test names/results and direct-read built-binary classifications;
- generator, docs, boundary, and static gate results;
- an explicit live-certification result or the one non-secret credential-reference blocker.
- six-surface inventory, CLI/help/manual/website reachability, output-preservation, and App-path
  classifications for every supported provider operation; any `not_applicable` status cites the
  exact source-audit row(s), never a safety or tier label.
