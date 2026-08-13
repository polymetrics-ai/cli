# #4077 — TDD ledger

## Baseline

- **Exact parent head:** `c67f40a5ff67a131950f3123e70527027dca8493`
- **Temporary pre-change reproduction:** failed as expected; see `PROBLEM.md`.
- **Disconfirming control:** existing `[]byte` and `map[string]any` cloning remained independent.

## Slice 1 — record value isolation

| Stage | Required evidence | Status |
|---|---|---|
| Red | Direct and nested `json.RawMessage` / `map[string]string` mutations affect source data; source → stage → destination paths show the same escape. | Pending committed RED |
| Green | Explicit clones leave source storage unchanged at every boundary. | Pending |
| Refactor | Preserve existing `[]byte`, `map[string]any`, `[]any`, and `[]connectors.Record` coverage and keep code closed. | Pending |

## Slice 2 — unknown mutable values

| Stage | Required evidence | Status |
|---|---|---|
| Red | Unknown mutable map/slice can currently pass the default clone case by alias. | Pending committed RED |
| Green | Source-page and stage-workset copies return contextual errors before stage/apply respectively. | Pending |
| Refactor | No reflection-based cloning, no replacement/drop/panic, no widening of provider types. | Pending |

## Regression preservation

- [ ] `[]byte`
- [ ] `map[string]any`
- [ ] `[]any`
- [ ] `[]connectors.Record`
- [ ] checkpoints and acknowledgements
- [ ] per-stream CAS
- [ ] seven canonical sync modes
- [ ] normal and race package tests
