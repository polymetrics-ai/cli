# Eval Plan

## Acceptance Checks

- All five sync modes pass app-level tests.
- Failed overwrite and failed dedupe overwrite preserve previous final output.
- Incremental checkpoint state advances only after success.
- Generated docs and skills describe sync-mode requirements.

## Manual Review

Review final implementation for:

- accidental secret output
- unbounded memory warnings in dedupe path
- mode behavior accidentally implemented inside connectors

