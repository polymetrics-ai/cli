# PROMPTS — issue-3062 Notion parity wave03 r1

## Kickoff

Task: implement complete documented Notion connector parity for parent #3062 and children #3063-#3069, wave03 worker-only. Do not run `/no-mistakes`, push, PR, merge, live provider checks, credentials, certification, VPS, or Thaalam changes.

GSD command attempted: `scripts/gsd prompt programming-loop init --phase issue-3062-notion-parity-wave03-r1 --dry-run`.

Result: command missing; manual GSD fallback recorded in PLAN/TDD/RUN-STATE.

Downstream artifact: `.planning/phases/issue-3062-notion-parity-wave03-r1/PLAN.md`.
Verification result: required local fixture-only gates passed; final `make verify` passed.

## Completion prompt summary for handoff

Branch `fm/cli-notion-parity-wave03-r1` implements fixture-only Notion official API parity for issues #3062-#3069. Final official counts: 49 total, 45 implemented/fixture-tested, 1 blocked/planned multipart upload, 3 excluded OAuth lifecycle, 0 certified. No push/PR/no-mistakes/live-provider work performed.
