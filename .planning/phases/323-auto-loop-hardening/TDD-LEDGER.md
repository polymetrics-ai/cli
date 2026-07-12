# TDD Ledger: Autonomous Delivery Control-Plane Hardening

## Parent scaffold — 2026-07-12

- Task type: parent orchestration and planning; no production behavior changed.
- Red-test exemption: parent scaffold only. Every child issue except documentation-only work must
  record red evidence before production edits.
- Baseline: `origin/main` at `cab8f3df`.
- GSD doctor:
  `PATH="$HOME/.nvm/versions/node/v24.13.1/bin:$PATH" scripts/gsd doctor` -> pass.
- Adapter capability check:
  `scripts/gsd prompt programming-loop init --phase 323-auto-loop-hardening --dry-run` -> expected
  failure because `programming-loop` is not in the current adapter registry.
- Fallback preflight:
  installed programming-loop helper `run --phase 323-auto-loop-hardening --mode auto --subagents true --dry-run`
  -> pass; agent mode selected.
- Parent planning prompt:
  `scripts/gsd prompt plan-phase 323-auto-loop-hardening --skip-research` -> rendered; the active
  orchestrator executed its issue-first/TDD planning requirements in `PLAN.md`.
- First orchestration decision: `local_critical_path` to create the parent branch, durable plan,
  draft parent PR, and complete child issue contracts before mutating workers start.
- Parent issue checkpoint: draft PR #324 and fifteen native sub-issues #325-#339 created. The
  architecture pass split phases 1, 2, 4, 6, and 7 before coding. This remains a planning-only
  checkpoint; Phase 0 red evidence is not yet claimed.

## Required red/green matrix

| Phase | First red evidence | Minimum green evidence |
|---|---|---|
| 0 | Each sanitized incident fixture demonstrates the current fail-open behavior | Replay suite detects every incident; irreversible action switch is off by default |
| 1 | Second controller and orphan-child fixtures can both mutate | Singleton lease/takeover and process-group revocation tests pass under race detector |
| 2 | Concurrent/restarted transitions overwrite or double-apply | CAS winner, immutable IDs, durable RETRY/HALT, authorized resume tests pass |
| 3 | File-only scope omits a transitive gate/artifact | Closure compiler rejects incomplete scope; ticket is single-use |
| 4 | Stale/overlapping worker can edit or push | Current scoped lease only; quiescence and dirty/untracked audit pass |
| 5 | Stale/moved PROCEED can checkpoint | Nonce-bound transaction returns typed moved-evidence failure; score floors hold |
| 6 | Duplicate delivery or stale attestation repeats/misapplies an effect | Outbox is idempotent and exact-head attestation succeeds once |
| 7 | Trace collisions, untyped handoffs, or embedded child history pass | Schema validation, redaction, unique IDs, and bounded references pass |
| 8 | Model is invoked for unchanged waits or low-risk checks | Watchers poll deterministically without model; disagreement escalates |
| 9 | At least one historical/fault scenario bypasses the composed controls | Full replay/fault/shadow/canary suite passes with merge disabled |

## Existing Shepherd alignment — 2026-07-12

- RED: after merging `feat/pi-shepherd-loop`, `bash scripts/tests/auto-loop-control.sh` reported four
  inventory/guard failures for `scripts/pi-shepherd-loop.sh`.
- RED (expanded): after adding direct launcher and model assertions, the same harness reported 23
  failures: the Shepherd performed `dirname`/`mkdir` before denial, all six run/resume/help/override
  cases were unguarded, and the validator still targeted GPT-5.5. Go tests also rejected the missing
  third canonical entrypoint.
- GREEN: added first-action denial, synchronized shell/Go inventories, exercised all three wrappers,
  and changed only the validator default to `openai-codex/gpt-5.6-sol --thinking high`.
- GREEN commands:
  `go test ./internal/agentloop/... ./cmd/loopctl/... -count=1` and
  `bash -n scripts/auto-loop-safety.sh scripts/pi-shepherd-loop.sh scripts/tests/auto-loop-control.sh && bash scripts/tests/auto-loop-control.sh`.
- Compatibility evidence: Pi 0.80.3 reports no GPT-5.6 match; a read-only Pi 0.80.6 invocation lists
  exact `openai-codex/gpt-5.6-sol` support. The launcher now checks this before mutable work once the
  migration fuse is opened.
