# Prompt trace — issue 196 Gorgias parity

Kickoff task: implement complete documented Gorgias connector parity for parent issue #196 and subissues #197-#203 in the disposable worktree branch `fm/cli-gorgias-parity-wave02-r1`.

GSD prompt path attempted:

```bash
scripts/gsd doctor
scripts/gsd prompt programming-loop init --phase issue-196-gorgias-parity --dry-run
```

`programming-loop` was unavailable in `scripts/gsd list`, so this run follows `.pi/prompts/pm-gsd-loop.md` and `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` as a manual GSD fallback.

Downstream artifact: connector-local Gorgias bundle parity implemented in `internal/connectors/defs/gorgias/**`, with official operation trace in `traces/official-operations.json` and verification evidence in `VERIFICATION.md`.
Verification result: local connector gates passed except the broad `internal/cli` regex gate, which timed out in existing certification batch coverage and is recorded as not green.
