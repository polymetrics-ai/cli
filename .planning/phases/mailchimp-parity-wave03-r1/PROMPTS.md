# Mailchimp parity wave03-r1 prompts

## Kickoff

Task from firstmate: implement complete documented Mailchimp connector parity for parent #3078 and subissues #3079-#3085, stop after a clean locally tested commit, do not invoke `/no-mistakes`, push, open/update PR, merge, or use live provider calls/credentials.

## GSD command trace

- `scripts/gsd doctor` passed.
- `scripts/gsd prompt programming-loop init --phase mailchimp-parity-wave03-r1 --dry-run` failed because the repo-local command registry does not expose `programming-loop`.
- Manual GSD universal programming-loop fallback is active and recorded in `PLAN.md`/`RUN-STATE.json`.
- `scripts/gsd prompt execute-phase mailchimp-parity-wave03-r1 --dry-run` was inspected as the closest repo-local official prompt trace.

## Downstream artifact

- `PLAN.md`
- `TDD-LEDGER.md`
- `VERIFICATION.md`
- `RUN-STATE.json`
- `SUMMARY.md`

## Verification result

Pending.
