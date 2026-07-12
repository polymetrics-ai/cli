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
   synthetic, and event-bound; event sequence numbers are contiguous and increasing. Events carry
   neutral typed facts with resource, owner, before, and after values rather than conclusion labels.
5. Fixtures contain no raw command, prompt, session-record path, credential material, or
   secret-shaped value. Files ending in `.jsonl` are rejected before reading.
6. Replay derives a violation and reason codes from fact relationships and ordering, then compares
   them to the fixture expectation; it never trusts the expected code as its detection result.
7. Historical observation is separate from policy: observed decision, observed outcome, and the
   correctness of each may equal the required result. Correct HALT/RETRY behavior is preserved.
8. JSON output uses stable field names, stable violation/exit classes, and deterministic ordering.
9. Output-bearing incident IDs and observation fields are closed vocabularies. An incident ID must
   map to the violation derived from facts; ambiguous or cross-resource fact composites fail.
10. All code uses the Go standard library and local shell utilities already required by the repo.

## Incident matrix

| Fixture | Semantic trigger | Stable violation code | Required decision / exit class |
| --- | --- | --- | --- |
| `dead_worker.json` | unproven worker completion was accepted with PROCEED | `WORKER_COMPLETION_UNPROVEN` | `retry` / `retry_required` |
| `false_green.json` | missing required artifact, PROCEED, then repo-gate failure | `VALIDATION_FALSE_GREEN` | `retry` / `retry_required` |
| `fabricated_authority.json` | agent clears a human-only gate | `AUTHORITY_FABRICATED` | `halt` / `halt_required` |
| `halt_worker_survival.json` | registered worker remains alive after HALT | `HALT_REVOCATION_MISSING` | `halt` / `halt_required` |
| `mega_turn.json` | a turn exceeds budget and becomes detached from supervision | `TURN_SUPERVISION_EXCEEDED` | `halt` / `halt_required` |
| `dual_writer.json` | overlapping actors own one worktree | `WORKTREE_DUAL_WRITER` | `halt` / `halt_required` |
| `merge_before_ratification.json` | merge precedes ratification | `MERGE_BEFORE_RATIFICATION` | `halt` / `halt_required` |
| `merge_stale_attestation.json` | same-head merge state moves OPEN→MERGED before stale PROCEED | `MERGE_ATTESTATION_STALE` | `halt` / `halt_required` |
| `merge_agent_authority.json` | an agent completes a human-only merge | `MERGE_AUTHORITY_DENIED` | `halt` / `halt_required` |
| `stale_verify_head.json` | verified head moves and Shepherd correctly RETRYs | `VERIFY_HEAD_STALE` | `retry` / `retry_required` |
| `dirty_worktree.json` | worktree becomes dirty and Shepherd correctly RETRYs | `WORKTREE_DIRTY` | `retry` / `retry_required` |
| `interim_human_wait.json` | blocked active human wait/human gate is projected done | `HUMAN_WAIT_PROJECTED_FINAL` | `wait` / `human_wait_required` |
| `terminal_projection_mismatch.json` | canonical running/non-terminal state conflicts with stale human-gate projection and PROCEED | `TERMINAL_PROJECTION_MISMATCH` | `halt` / `halt_required` |

Truth corrections are grounded in the read-only run post-mortem and driver trace: dead-worker
fail-open (`ANALYSIS-CODEX.md:98`), same-head merge-state race (`:109,126,159`), non-durable turn-23
HALT (`:169,173`), turn-26 ledger divergence (`:189-199`), and valid final human-ready projection
(`:326`). Those source artifacts remain in the completed run worktree and are not copied here.

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

The driver harness discovers candidates independently from both filename and semantic signals,
uses an empty environment and isolated PATH, makes state/prompt inputs unwritable/unreadable, and
proves enable-like environment/flag canaries cannot open the fuse.

## Explicit exclusions

- Controller leases, durable state, stage tickets, validator transactions, GitHub outbox, merge
  attestation, event-derived UI, and canary execution belong to dependent issues #326-#338.
- No connector, Twenty, prompt, validator, authentication, ruleset, provider, or live GitHub
  behavior is changed.
