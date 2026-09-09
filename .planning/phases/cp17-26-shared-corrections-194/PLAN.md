# Shared corrections194 plan

## Task Delivery Header
- Issue: Refs #4344; programme https://github.com/polymetrics-ai/cli/issues/4325.
- Base branch: main (API-confirmed existing https://github.com/polymetrics-ai/cli/pull/4294); working integration fm/cli-top100-declaration-batch-r1.
- Merges into: fm/cli-top100-declaration-batch-r1; no main merge.
- Working branch: fm/cli-top100-declaration-batch-r1-cp17-26 from45c235dda353dc9e30838e860acac0aa5ebd2e5c.
- Delivery: coherent shared correction candidate with source-bound normal/race/physical evidence, then separately bound correction review; Firstmate owns connector dispatch/publication schedule.
- Task: CR-173-01 reject double-hyphen option tokens at any command position without changing flag-name grammar or real parser; WR-173-01 enforce monotonic native PostgreSQL LSN and stable checkpoint identity in the physical test oracle.
- Verification: actual canonical render/rejection and hand-parser controls, affected CLI/commandrunner/connectorgen normal/race; PostgreSQL oracle falsifiers plus existing authorized disposable physical witness, exact original runner receipts and cleanup.

## TDD sequence
1. CR-173-01: preserve original173-P01 RED, add coherent option-prefix/position/availability and healthy literal/dotted/hyphenated aliases. Capture actual failing render/parser regression before shared guard edit. Add smallest prefix check to existing guard; verify new and existing legal flags remain distinct. Atlas and authoring docs same-change ownership update.
2. WR-173-01: preserve173-P02 RED. Tests use native LSN tokens, forward/backward/unchanged/malformed, same-LSN tie/time, changed source/mechanism/protocol identity. Capture RED before oracle edit; reuse existing pglogrepl parser and checkpoint identity fields. No product CDC edits. Rerun affected physical witness after focused GREEN/race.
3. Verify relevant packages, gofmt/scoped vet/lint and GSD. Exact source-bound candidate handoff for Firstmate-authored correction review. Do not replay unchanged full corpus/API proof or claim hosted CI green.

## Owners and proof scope
Only shared commandrunner guard, dedicated generator/CLI tests, PostgreSQL test oracle/test file, related Atlas/docs and this phase. All connector authored/runtime files remain excluded, especially Notion. Existing Atlas authoring.source-lock-vnext.v1 and runtime.direct-execution.v1 identify shared owners; transport.sync-contract.v1/warehouse references supply existing storage/identity contracts. No new foundation/dependency.

| Contract | Evidence | Assertion |
|---|---|---|
| Positional token survives parser | fake provider; actual local renderer/parser | invalid aliases rejected before publication, healthy aliases preserve exact positional/source mapping; no provider I/O needed |
| Checkpoint advancement | actual helper + fake token states | native LSN strictly advances, identity preserved; backward/equal/malformed refuse |
| Physical checkpoint witness | authorized disposable local PostgreSQL | actual independent transaction LSN/rows after barrier and process restart, receipt/lease/storage cleanup; no customer database |

## Lifecycle and skills
scripts/gsd sources and prompt discuss-phase, plan-phase --tdd, execute-phase, verify-work; inline execution due single-owner routing. Existing loaded Go how-to/CLI/testing/error/security/safety/design/structs/context/concurrency/database/lint/documentation, connector lane skill, exhaustive review and evidence policy. No new child.

Help/manual/site: no existing legal command or flag changes. Authoring grammar docs and Atlas updated; existing generated corpus remains unchanged. Check help/parity only if actual command surface changes. Preserve all failed attempts and before/after input hashes.
