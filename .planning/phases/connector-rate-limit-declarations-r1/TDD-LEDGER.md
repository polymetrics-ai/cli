# TDD ledger — provider-cited rate-limit declarations R1

## Required red / green evidence

| Slice | Red evidence | Green evidence | Status |
| --- | --- | --- | --- |
| First declaration embed | `TestProductionDefinitionsEmbedEveryRateLimitDeclaration` failed for `harvest/rate_limits.json` before the optional wildcard was added | The same test passed after adding `*/rate_limits.json` to `defs.FS` | green |
| Closed declaration validation | Invalid source/scope/shape is rejected by existing engine tests and loader rules | The retained 24 first-batch files parse and `go test ./internal/connectors/engine` passes | green |
| Batch conformance | Surface and command metadata remain unchanged | `connectorgen validate` and `surface-sync --check` pass with zero findings or changes | green |
| Population eligibility | Every declaration directory is joined to the authoritative sweep ledger | The retained 24 first-batch records join to `status: done` provider artifacts; `vercel` was removed because it is absent from the sweep | green |
| Batch 2 eligibility | Each named candidate is rejoined to the live sweep ledger before authoring | All 25 join to `status: done` records with provider artifacts and `scope_in_current_defs: true` | green |
| Batch 2 declaration conformance | Invalid source/scope/shape is refused by existing loader tests | 25 files parse; engine, commandrunner, generator, and scoped repository gates pass with no `streams.json` change | green |

## Evidence rules

- Existing generic engine tests are the executable specification; no new runtime behavior is
  introduced by this declarative rollout.
- The declaration review is fail-closed: no source citation, exact policy shape, or compatible
  non-secret scope property means `unknown`, never a numeric estimate.
- `streams.json` is inspected for accidental changes but is not edited.
