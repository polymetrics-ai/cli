# TDD ledger: Gong release-0.3.0 live parity reconciliation

## Red → green slices

| Slice | Red evidence | Green evidence | Refactor boundary |
| --- | --- | --- | --- |
| Current source inventory | Fresh official OpenAPI inventory had not been compared to the Batch 2/3 source lock. | Credential-free fetch proves all 69 semantic operation rows match the refreshed strict lock exactly. | Lock the current exact artifact and compare a sorted semantic fingerprint so serialization-only changes cannot erase operations. |
| Foundation reconciliation | Preserved branch predates current main, typed destinations, and Batch 2/3 source disposition evidence. | Merge ancestry proves retained branch history plus the exact published foundation heads. | Resolve connector-owned declarations, not provider-named engine conditions. |
| Direct-read exact endpoint binding | Reproduce any Gong command that preflights with an implemented operation but no matching `api_surface` row. | Real `commandrunner.Preflight`, `surface-reconcile`, and a built CLI preflight sweep accept each declared direct read up to missing credentials. | Let `surface-sync` derive operation-owned metadata; do not hand-author it. |
| Typed write and reverse-ETL declarations | The historical CLI marks 24 exact write actions partial; the three multipart actions were recorded as an F4 foundation gap. | All 27 named actions are implemented, their CLI field shapes pass runtime preflight, and focused Gong multipart conformance passes through the shared approval-digest path. | No generic writer, raw body, arbitrary endpoint, or Gong-specific shared branch. |
| Provider output preservation | Gong command metadata described ordinary provider response fields as redacted. | A focused Gong surface test fails on direct-read redaction language or read-field redaction declarations; current descriptions preserve ordinary response fields and mask only configured credential values. | Keep credential masking at the generic output boundary; do not create field-name heuristics. |
| Six-surface enabled parity | A source-locked operation can be structurally declared yet be absent from CLI/App dispatch, generated docs, or a supported ETL, reverse-ETL, direct, or binary path. | Generated inventory-to-surface evidence and built-CLI/App checks classify each of ETL, reverse ETL, direct read, direct write, binary download, and binary upload as proven or exact-source `not_applicable`. | Safety, scope, tier, and destructive metadata can add confirmation; they cannot disable a provider-defined operation. |
| Certification evidence | No credential reference means live stages cannot assert persisted provider state. | Credential-free gates are green and the remaining external block is explicit and secret-free. | Do not substitute browser authentication or fixtures for live certification. |

## Recorded red evidence

- Source-lock import and Batch 2/3 declarations are absent from the preserved branch; current
  `origin/main` also does not contain the Batch 2/3 source-lock files.
- The historical branch's phase records 67 operations. The current official OpenAPI has 69,
  confirming that historical completion evidence is insufficient for this release certification.
- Direct-read runtime coverage must be re-proven after reconciliation because prior audits found
  declaration rows that validated structurally but lacked exact executable `api_surface` bindings.
- The first output-preservation assertion did not compile until the focused test model was extended
  with CLI risk metadata; it then failed on `calls get` claiming its fields were redacted. This
  established a declaration-level red case without provider I/O.
- `node .../verify-parity-maps.mjs` failed after the shared multipart foundation was merged because
  its generated ledger still carried the now-obsolete Gong F4 rows. Removing that stale special
  case and regenerating the ledger made the 19-connector/5,127-operation check pass with zero
  Gong gap rows.
- `go run ./cmd/connectorgen source-import gong --check` now reaches the strict lock validator but
  fails before fetch because the official fixed URL contains `?version=`. This is recorded as the
  remaining provider-neutral source-import URL-policy dependency, not a Gong-specific fallback.

## Green evidence recorded during execution

- `go test -timeout 20m ./cmd/connectorgen -run '^TestGong(FullSurfaceCommandAndOperationCoverage|MetadataEnablesWriteCapability)$' -count=1` passed after the reverse-ETL and output-preservation declarations were corrected.
- `go test -timeout 20m ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1` passed for every implemented bundle command.
- `go test -timeout 20m ./internal/connectors/conformance -run 'TestConformance/gong' -count=1` passed, including the three declaration-owned multipart actions.
- In one freshly initialized project with no configured credential, the built binary classified all 30 direct-read commands, all 27 reverse-ETL write commands, and all 12 ETL stream commands as `missing --credential`. Each command was invoked without provider credentials; zero classified as unknown, partial, or unbound.
- `node .planning/phases/issue-4289-parity-map-batches-2-3-r1/traces/verify-parity-maps.mjs` passed: 19 connectors / 5,127 documented operations, with zero remaining Gong foundation-gap rows.

## Green evidence to record during execution

- exact inventory diff result, generated source-map result, and source/disposition arithmetic;
- focused Gong test names/results and direct-read built-binary classifications;
- generator, docs, boundary, and static gate results;
- an explicit live-certification result or the one non-secret credential-reference blocker.
- six-surface inventory, CLI/help/manual/website reachability, output-preservation, and App-path
  classifications for every supported provider operation; any `not_applicable` status cites the
  exact source-audit row(s), never a safety or tier label.
