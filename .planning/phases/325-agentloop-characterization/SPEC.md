# Specification: Phase 0 Incident Replay and Migration Fuse

Issue: #325  
Parent: #323  
Parent PR: #324  
Branch: `fix/325-agentloop-characterization`

## Outcome

Provide a dependency-free oracle for thirteen synthetic incident classes and close every tracked
autonomous-loop run/resume path before it can persist state, render prompts, launch a model or
child process, or reach git/GitHub. Phase 0 intentionally has no production enable mechanism.

## Invariants

1. Safety is closed in source. No argument, environment variable, prompt, state file, or resume
   artifact can enable run/resume.
2. Help, read-only safety status, tracked-entrypoint enumeration, and sanitized replay remain
   available while safety is closed.
3. A driver guard executes before `.planning/auto-loop` creation, prompt reads/writes, logging,
   process launch, model invocation, git, or GitHub activity.
4. Fixtures are strict JSON: unknown fields and trailing values fail; every identity is complete,
   synthetic, and event-bound; event sequence numbers are contiguous and increasing.
5. Fixtures contain no raw command, prompt, session-record path, credential material, or
   secret-shaped value. Files ending in `.jsonl` are rejected before reading.
6. Replay derives a violation from event semantics, then compares it to the fixture expectation;
   it never trusts the expected code as its detection result.
7. JSON output uses stable field names, stable violation/exit classes, and deterministic ordering.
8. All code uses the Go standard library and local shell utilities already required by the repo.

## Incident matrix

| Fixture | Semantic trigger | Stable violation code | Required decision / exit class |
| --- | --- | --- | --- |
| `dead_worker.json` | proceed after worker becomes unreachable | `WORKER_LIVENESS_LOST` | `retry` / `retry_required` |
| `false_green.json` | proceed with missing or stale validation evidence | `VALIDATION_FALSE_GREEN` | `retry` / `retry_required` |
| `fabricated_authority.json` | agent clears a human-only gate | `AUTHORITY_FABRICATED` | `halt` / `halt_required` |
| `halt_worker_survival.json` | registered worker remains alive after HALT | `HALT_REVOCATION_MISSING` | `halt` / `halt_required` |
| `mega_turn.json` | a turn crosses the explicit supervision cap | `TURN_CAP_EXCEEDED` | `halt` / `halt_required` |
| `dual_writer.json` | overlapping actors own one worktree | `WORKTREE_DUAL_WRITER` | `halt` / `halt_required` |
| `merge_before_ratification.json` | merge precedes ratification | `MERGE_BEFORE_RATIFICATION` | `halt` / `halt_required` |
| `merge_stale_attestation.json` | head moves after attestation | `MERGE_ATTESTATION_STALE` | `halt` / `halt_required` |
| `merge_agent_authority.json` | an agent completes a human-only merge | `MERGE_AUTHORITY_DENIED` | `halt` / `halt_required` |
| `stale_verify_head.json` | proceed after verified head moves | `VERIFY_HEAD_STALE` | `retry` / `retry_required` |
| `dirty_worktree.json` | transition proceeds from dirty scope | `WORKTREE_DIRTY` | `retry` / `retry_required` |
| `interim_human_wait.json` | waiting-for-human is projected as success | `HUMAN_WAIT_PROJECTED_FINAL` | `wait` / `human_wait_required` |
| `terminal_projection_mismatch.json` | durable and displayed terminal disagree | `TERMINAL_PROJECTION_MISMATCH` | `halt` / `halt_required` |

## CLI contract

- `loopctl help`, `loopctl --help`, or bare `loopctl`: print command help and exit 0.
- `loopctl safety status [--json]`: print the closed safety state and exit 0.
- `loopctl safety entrypoints [--json]`: print the two tracked driver paths and exit 0.
- `loopctl safety guard-driver <path> [--json]`: print a typed denial and exit 78.
- `loopctl replay <fixture.json> [--json]`: validate and replay one fixture; matching expected
  denial is a successful oracle run (exit 0), malformed/mismatched input exits 65.
- Unknown commands or invalid arguments exit 64. Diagnostics go to stderr; JSON/status output goes
  to stdout except driver guard denials, which go to stderr so stdout cannot be mistaken for model
  output.

`loopctl` is an internal operator tool, not a `pm` command. The issue's write scope excludes
`docs/cli/**` and `website/**`; CLI parity is therefore limited to runtime help and command tests,
with that exemption recorded in verification and the PR body.

## Shell contract

`scripts/auto-loop-safety.sh` is both sourceable and directly executable:

- `status [--json]` exits 0;
- `entrypoints [--json]` exits 0;
- `guard-driver <tracked-path> [--json]` exits 78 with `AUTO_LOOP_DISABLED_PHASE_0`;
- an untracked path fails closed with `AUTO_LOOP_ENTRYPOINT_UNTRACKED` and exits 64;
- there is no enable, open, run, or resume command.

## Explicit exclusions

- Controller leases, durable state, stage tickets, validator transactions, GitHub outbox, merge
  attestation, event-derived UI, and canary execution belong to dependent issues #326-#338.
- No connector, Twenty, prompt, validator, authentication, ruleset, provider, or live GitHub
  behavior is changed.
