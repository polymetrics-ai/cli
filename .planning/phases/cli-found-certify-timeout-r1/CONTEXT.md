# Context — certify-harness cost foundation (#3795)

## GSD discussion record

`scripts/gsd doctor` passed and the prompt traces were generated with:

```text
scripts/gsd prompt discuss-phase 3795 --auto
scripts/gsd prompt plan-phase 3795 --tdd --skip-research
```

The adapter's normal interactive runtime is not available in this single-worker
session, and the canonical delivery contract forbids role spawning. The
documented inline/manual fallback is therefore used. `--auto` is valid here:
the parent issue and all five children lock the remaining design decisions.

## Locked decisions

- #3795 is one parent implementation PR, with #3798 → #3801 → #3805 → #3806
  → #3807 executed in that dependency order.
- The production `Harness.Run` continues to invoke real in-process `cli.Run`.
  The only new seam may be test-only or strictly behavior-preserving.
- Count every real `cli.Run` through the harness seam. A scripted driver is
  counted separately and must assert command arguments, envelope kind, exit
  code, stage outcome, leaks, and idempotency so it cannot mask protocol drift.
- Keep an explicit exhaustive real certification proof that covers source/read
  full sweep, flow, schedule, and destructive plan → preview → approval →
  execute. Keep exactly one real CLI router proof for
  `pm connectors certify sample --json`.
- Render/output/persistence/batch tests use complete synthetic
  `certify.Report` and `certify.BatchReport` fixtures. Fixtures are complete,
  synthetic, non-secret, and unredacted.
- The Verify timing target consumes real `go test -count=1 -json` output for
  `./internal/connectors/certify` and `./internal/cli`, prints source events,
  and fails clearly for malformed streams or target test failures.
- Invocation count is the deterministic primary cost guard. A wall-clock
  budget is a secondary guard and is derived from a fresh final-topology timing
  measurement after the single real sample/outbox lifecycle proof is retained.
  It must never be guessed, cached, or used to raise the global Go-test
  timeout.

## Scope fences

- No connector definitions, provider artifacts, credentials, provider calls,
  report schema, runtime capability, browser/vault/mechanism behavior, new
  dependency, output suppression, or redaction path changes.
- No retry of non-idempotent write work; the approval replay remains a negative
  case.
- The only planned production-adjacent paths are the smallest
  behavior-preserving harness seam and the narrow Make/Verify wiring required
  by #3806/#3807. A shared command schema, connector engine, or commandrunner
  change is a stop-and-split condition.
- CLI help/manual/website parity is intentionally not applicable: no command,
  flag, help text, or user-visible report schema changes. The retained route
  proof tests existing wiring only.
