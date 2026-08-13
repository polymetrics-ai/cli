# #4077 — TDD ledger

## Baseline

- **Exact parent head:** `c67f40a5ff67a131950f3123e70527027dca8493`
- **Temporary pre-change reproduction:** failed as expected; see `PROBLEM.md`.
- **Disconfirming control:** existing `[]byte` and `map[string]any` cloning remained independent.

## Slice 1 — record value isolation

| Stage | Required evidence | Status |
|---|---|---|
| Red | Direct and nested `json.RawMessage` / `map[string]string` mutations affect source data; source → stage → destination paths show the same escape. | Failed as expected; recorded by `test(4077-01)` |
| Green | Explicit clones leave source storage unchanged at every boundary. | Passed focused normal test |
| Refactor | Preserve existing `[]byte`, `map[string]any`, `[]any`, and `[]connectors.Record` coverage and keep code closed. | Passed focused, package, race, and all-mode checks |

## Slice 2 — unknown mutable values

| Stage | Required evidence | Status |
|---|---|---|
| Red | Unknown mutable map/slice can currently pass the default clone case by alias. | Failed as expected; source record reached the stage and stage workset reached destination with `map[string]int` |
| Green | Source-page and stage-workset copies return contextual errors before stage/apply respectively. | Passed focused normal test |
| Refactor | No reflection-based cloning, no replacement/drop/panic, no widening of provider types. | Passed review and static checks |

## RED execution evidence

Command:

```bash
go test -count=1 -run 'TestCloneRecordCopiesRawMessageAndStringMapValuesAtEveryNestingLevel|TestOrchestratorProtectsRawMessageAndStringMapFromStageAndDestinationMutation|TestOrchestratorRejectsUnsupportedMutableValuesBeforeBoundaryCrossing' ./internal/synctransport
```

Observed failures before production changes:

- Direct `json.RawMessage` mutation changed source storage to `[`-prefixed JSON.
- A warehouse stage changed the source-owned raw message and a destination changed the source-owned `map[string]string` owner to `destination`.
- Both an unsupported source `map[string]int` and one injected by the warehouse stage completed with `Run() error = nil`.

The pre-existing `[]byte` / `map[string]any` controls remain the disconfirming evidence: those have dedicated clone cases and did not exhibit the aliasing escape.

## GREEN execution evidence

The focused command above passed after the minimal implementation:

- `json.RawMessage` is copied as its own named byte slice before the `[]byte` case.
- `map[string]string` receives an allocated key/value copy at every supported nesting level.
- The closed clone contract returns a wrapped `errUnsupportedTransportRecordValue` for unrecognized values; the source page now fails before `Stage`, and a stage-created unsupported workset fails before `ApplyDestination`.
- No reflection, provider adapter, mode, credential, registry, polling, or public CLI change was introduced.

## Regression preservation

- [x] `[]byte`
- [x] `map[string]any`
- [x] `[]any`
- [x] `[]connectors.Record`
- [x] checkpoints and acknowledgements
- [x] per-stream CAS
- [x] seven canonical sync modes
- [x] normal and race package tests
