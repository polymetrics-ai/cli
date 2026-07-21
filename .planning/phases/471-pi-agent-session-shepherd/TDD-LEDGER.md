# TDD Ledger

Phase: `471-pi-agent-session-shepherd`

## Pre-production gate

- Status: planning complete; RED tests not yet written.
- Production edits: none.
- Required next action: add focused tests and capture their expected failures before creating any
  production module under `.pi/extensions/shepherd/`.

## Planned RED slices

| Slice | Test command | Expected pre-implementation failure |
| --- | --- | --- |
| arguments | `node --test .pi/extensions/shepherd/arguments.test.ts` | production parser module is absent |
| domain | `node --test .pi/extensions/shepherd/domain.test.ts` | domain/rating module is absent |
| state/controller | `node --test .pi/extensions/shepherd/state-store.test.ts .pi/extensions/shepherd/controller.test.ts` | store/controller modules are absent |
| SDK adapter | `node --test .pi/extensions/shepherd/sdk-runner.test.ts` | adapter module is absent |

GREEN and refactor evidence will be appended immediately after each observed transition.
