# Agent Trace — Cycle 9 Seam Mapper

## Rendered Prompt Or Prompt Reference

See `../agents/cycle9-seam-map.md`. The explorer was asked to map the two complete stable-head
reviews onto exact issue-owned seams while remaining strictly read-only.

## Files Inspected

- `/tmp/475-REVIEW-CYCLE8-1.md`
- `/tmp/475-REVIEW-CYCLE8-2.md`
- `.pi/extensions/shepherd/agent-session-runtime.ts`
- `.pi/extensions/shepherd/tool-policy.ts`
- `.pi/extensions/shepherd/role-prompts.ts`
- `.pi/extensions/shepherd/agent-session-runtime.test.ts`
- `.pi/extensions/shepherd/tool-policy.test.ts`
- Pi 0.80.6 public declaration files for `ToolDefinition`, `AgentToolResult`, `TSchema`, and
  `CreateAgentSessionOptions`

## Actions Taken

- Mapped thirteen deduplicated defect families to named functions and test seams.
- Produced a compact behavior-level RED matrix spanning all fourteen planned invariants, including
  retained Cycle 8 behavior.
- Identified generic hostile-Proxy enumeration and transitive TypeBox imports as invalid designs.
- Confirmed the explorer made no files, edits, commits, network calls, or verification runs.

## Commands Run

Read-only source/declaration searches and inspections only. No formatter, test, type-check, build,
runtime, GitHub, push, or service command was run by the explorer.

## Findings

1. Capture the SDK creation result once into a normalized ownership record; never reread its
   `session` accessor during validation, prompt, cancellation, or cleanup.
2. Keep expected tool names in a private frozen oracle and give Pi distinct arrays.
3. Snapshot schemas and tool results before await boundaries with explicit bounds and one-read
   semantics.
4. Distinguish `fulfilled`, `rejected`, and `pending`; settled setup rejection is retryable, while
   unresolved or malformed ownership stays fail-closed.
5. Bound unsubscribe and dispose independently so a hung unsubscribe cannot suppress dispose.
6. Capture listener operations at acquisition and retain a native fallback detach.
7. Wrap every public failure, including literal `undefined`, with an own `cause`; aggregate primary
   and cleanup failures deterministically.
8. Parse only known terminal event fields into bounded immutable DTOs. A generic `Reflect.ownKeys`
   walk cannot pre-bound a hostile Proxy because proxy key materialization is atomic.
9. Extend the shared redactor, credential-path classifier, sensitive capability-name classifier,
   and terminal-control validator at their common policy boundary.
10. Replace the local custom-tool shim and hidden double cast with Pi 0.80.6 public tool/result
    types and required `details`.

## Handoff Summary

Use a one-way capture pipeline: raw request -> immutable admitted DTO -> private policy snapshot ->
directly typed Pi tool definitions -> discriminated creation settlement -> one captured owned-
session port -> fixed-shape terminal DTO -> handoff. No caller- or SDK-owned object may survive an
await, and validation must not reread its source.

Pi 0.80.6 does not re-export TypeBox's `Type`, and `typebox` exists only as a nested transitive
installation in this environment. Do not import through that brittle package-internal path and do
not add a dependency. Pi's validator accepts compatible plain JSON schemas; type those definitions
directly as exported `ToolDefinition<TSchema>`, return exported `AgentToolResult` shapes with own
`details`, and prove the boundary through the pinned offline adapter exercise.

## Verification Evidence

The explorer began on clean frozen head `0cdcda7e`. It later observed the shared worktree clean at
PLAN head `b175cc4a`; it explicitly reported no intervening edit or commit of its own. The seam map
is therefore analysis of `0cdcda7e`, not an independent review of the PLAN commit.

## Unresolved Risks

- A generic graph walker cannot safely enumerate a hostile Proxy under a pre-allocation budget;
  closed known-field readers or explicit Proxy rejection are required.
- Lexical root path checks cannot prove physical symlink containment; the host workspace boundary
  and #479 stable physical identity remain parent-owned.
- Full Shepherd tests may still reproduce the known managed-sandbox `/bin/ps` `spawn EPERM`; that
  must be classified separately rather than treated as a Cycle 9 behavior regression.
