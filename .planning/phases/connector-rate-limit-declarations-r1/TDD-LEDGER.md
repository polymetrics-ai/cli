# TDD ledger — provider-cited rate-limit declarations R1

## Required red / green evidence

| Slice | Red evidence | Green evidence | Status |
| --- | --- | --- | --- |
| First declaration embed | `TestProductionDefinitionsEmbedEveryRateLimitDeclaration` failed for `harvest/rate_limits.json` before the optional wildcard was added | The same test passed after adding `*/rate_limits.json` to `defs.FS` | green |
| Closed declaration validation | Invalid source/scope/shape is rejected by existing engine tests and loader rules | All 25 authored files parse and `go test ./internal/connectors/engine` passes | green |
| Batch conformance | Surface and command metadata remain unchanged | `connectorgen validate` and `surface-sync --check` pass with zero findings or changes | green |

## Evidence rules

- Existing generic engine tests are the executable specification; no new runtime behavior is
  introduced by this declarative rollout.
- The declaration review is fail-closed: no source citation, exact policy shape, or compatible
  non-secret scope property means `unknown`, never a numeric estimate.
- `streams.json` is inspected for accidental changes but is not edited.
