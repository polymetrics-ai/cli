# TDD Ledger — Issue #204 Crisp CLI parity parent

## GSD mode

- Adapter health: `scripts/gsd doctor`, `scripts/gsd verify-pi`, `scripts/gsd list --json` passed on 2026-07-10.
- Plan command: `scripts/gsd prompt plan-phase 204 --skip-research`.
- Programming-loop command: `scripts/gsd prompt programming-loop init --phase issue-204-crisp-cli-parity --dry-run` was unavailable (`unknown GSD command: programming-loop`). Manual GSD/TDD fallback is active and recorded in this ledger.

## Red / green / refactor ledger

| Slice | Issue | Red evidence | Green evidence | Refactor evidence | Status |
|---|---:|---|---|---|---|
| Parent orchestration seed | #204 | Not behavior-changing; no red test required | Planning artifacts created; adapter health passed | N/A | in progress |
| Crisp metadata/ledger scaffold | #205 | Run `go run ./cmd/connectorgen validate internal/connectors/defs/crisp` before bundle files exist; expect missing path/load failure | Same command passes after adding non-executable scaffold; then validate all defs | JSON formatting normalization only | planned |
| Stream runner | #207 | Failing conformance fixture/test for first Crisp stream before stream definitions | Conformance + connectorgen validation pass for stream slice | Remove over-broad schemas/fixtures | planned |
| Direct reads | #209 | Failing direct-read validation/test for selected bounded command | Direct-read command validates and rejects unsafe path/body shapes | Tighten output policy docs | planned |
| Sensitive/admin policy | #211 | Failing write schema/safety validation for one destructive/admin action | Write action requires plan/preview/approval/typed confirmation | Redaction and risk text cleanup | planned |
| Help/docs parity | #206 | Failing grep/help/docs check for Crisp command surfacing | Runtime help/docs/website/generated artifacts agree | Copyediting only | planned |

## TDD rules for future slices

- Behavior changes need a failing test, fixture, or validation artifact before production edits.
- Metadata-only rows may use validation-as-test where no Go behavior changes.
- Do not weaken existing tests or conformance checks.
- Do not mark `verificationPassed=true` in `RUN-STATE.json` until the full requested gate actually exits 0.
