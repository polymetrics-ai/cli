# Twenty CRM recovery and delivery context

## Task Delivery Header

- Issue: Refs #277 — twenty (Twenty CRM): connector all-ops CLI parity (parent)
- Base branch: fm/cli-reverse-etl-destination-r1
- Merges into: fm/cli-reverse-etl-destination-r1 → main
- Delivery: PR #4298 is retargeted to `fm/cli-reverse-etl-destination-r1`, contains the dependency merge and a complete connector-owned seven-surface declaration, has its repository gates green, and retains or re-runs real-data read/write/delete proof from a disposable self-hosted Twenty instance.
- Working branch: fm/cli-twenty-crm-gtm-parity-r1
- Task: Recover the previous Twenty bundle, reconcile it with current declarative connector foundations, and prove the real CLI can read, page, write, read back, and delete bounded disposable records without contacting the captain's Twenty workspace.
- Verification: targeted Twenty tests and conformance; connectorgen validation, surface sync and certification sweep checks; connector-boundary, generated docs/goldens, lint, build, and verification gates; built-binary live proof using a self-hosted Twenty instance and stdin-only credentials.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every researched Twenty operation is declared or has a fixed-vocabulary exclusion | live | The checked bundle ledger reports 168 classified endpoints and every command/action binding resolves. |
| Read, get, and paging work through the built CLI | live | A disposable instance returns records created for this proof; list/get results and page transition counts are recorded without credentials. |
| Create, update, delete, and round-trip work through the built CLI | live | A CLI-created record is read with matching fields, updated, then absent after the typed destructive flow. |
| Current generated and runtime contracts accept the bundle | live | The real generator, runtime preflight, conformance, and boundary commands pass. |

## Recovery assessment (before import)

Source recovered from `origin/fm/cli-twenty-parity-wave02-r1` commit
`11eb2a74d5e812c94f8bb2c10a3e0eb86f21f618`; it has no merge base with current
`main`. The commit contains 1,898 historical tree changes rather than an
isolatable connector commit. Its cherry-pick was aborted without a commit after
it created unrelated conflicts. Recovery therefore extracts only the verified
`internal/connectors/defs/twenty/**` subtree from that commit; no unrelated
historical file is imported.

### Survives

- Connector-local footprint only: 28 schemas, 28 streams, 168 API-surface rows,
  168 CLI rows, 112 typed `writes.json` actions (56 create, 28 update, 28 delete),
  29 stream fixture pages, and 112 write fixtures.
- The API ledger accounts for the researched REST CRUD inventory: 56 GET, 56 POST,
  28 PATCH, and 28 DELETE endpoints.
- Deletes are genuine `kind: "delete"` actions with path fields; they are not omitted.
- `spec.json` has a `base_url` override and an `x-secret` bearer key, which is
  necessary for a disposable self-hosted instance.

### Stale or incorrect under current main

- `api_surface.json` is v1 and covers get-by-id endpoints through a stream instead
  of an operation-backed direct-read command. It has no v2 provider-artifact
  provenance.
- The CLI surface leaves all 28 object `get` commands `planned` and all 28
  batch commands `partial`; ETL lists and scalar create/update/delete commands
  already resolve. Thus 56 declared API operations are not runtime-executable,
  which cannot meet all-ops working CLI parity.
- No `operations.json` binds the direct-read commands to current operation-backed
  endpoint contracts, so generated `maps_to`, output policies, and runtime
  preflight cannot prove those get operations.
- The delete actions retain legacy `confirm: "destructive"`; current typed
  confirmation syntax should be adopted while preserving the same safety gate.
- No `certification.json`, certification sweep artifact, or allowlisted
  certification matrix is present. Current `certificationConnectorAllowlist`
  contains only github, postgres, and zoom; changing it is foundation code and
  is outside this connector-only lane.
- The task brief's `data/PARITY-BAR.md` and Zoom declaration-plan path do not
  exist on current main. Current fixed-vocabulary enforcement instead comes from
  the loader/schema and `docs/migration/conventions.md`.

### Missing proof

- No current-main validation, runtime-preflight, generated-surface, or conformance
  evidence exists for the recovered tree.
- No live self-hosted Twenty proof, cleanup receipt, or secret-safe command record exists.

## Scope and foundation boundary

Only `internal/connectors/defs/twenty/**` plus this issue's planning evidence may
change. The single curated `internal/connectors/icon_data.json` fallback row for
Twenty is explicitly authorized: it is authored provenance (`source` and
`review_status` are `polymetrics`) and uses the documented sample fallback rather
than mislabelling another provider's asset. The other 556 curated rows remain
byte-identical; no upstream registry regeneration is authorized. A required
shared-engine, generator, allowlist, schema, or website change is a foundation
decision, not a connector workaround. The required GSD prompts were resolved
with `scripts/gsd prompt`; this execution uses the documented inline/manual
fallback because compatible isolated Pi workers are unavailable.

## Reconciliation — 2026-08-20

The historical recovery survives as a coherent six-surface REST bundle: it has
168 documented REST operations, 168 command rows, 28 ETL streams, 28 direct
reads, and 112 typed actions (56 create, 28 update, and 28 delete). Its source
locks and fixtures remain connector-owned, and its prior self-hosted proof is
preserved as historical evidence only. It is stale against the temporary
typed-destination foundation in PR #4304: no action has the new
connector-owned destination declaration required to make eligible actions
reverse-ETL destinations. The required next slice is therefore to merge
`origin/fm/cli-reverse-etl-destination-r1` without rewriting history, retarget
PR #4298 to that exact branch, and bind every eligible action through the
installed destination schema. Documented file transfer is modeled only as a
binary capability, never as a REST writer or reverse-ETL action.

This reconciliation stays scoped to the Twenty definition bundle, its derived
artifacts, and delivery evidence. An absent schema or executor capability is a
keyed `needs-decision`, not a foundation edit or an omitted operation.
