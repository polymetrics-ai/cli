# Context — issue #3714 parent readiness

Parent issue: #3714
Parent PR: #3723 (`refactor/3714-canonical-delivery-flow` → `main`)

## Fixed decisions

- The existing parent branch and PR are the only delivery surface. No replacement branch or PR is
  created.
- The parent must incorporate current `origin/main` before it is made ready for review.
- The Codex, Claude, and Pi harness projections are all required. The canonical contract is their
  source of truth; generated projections are regenerated with `agentcontractgen sync`, never
  hand-edited.
- The destructive-write confirmation gate already on `main` is preserved exactly. This phase may
  resolve a textual or generated-file conflict, but may not relax, bypass, or special-case that
  gate.
- Captain review and the final merge remain human-gated. This phase can make #3723 ready, never
  merge it.

## Scope

Integrate the three current `main` commits beneath #3723, resolve only integration conflicts,
regenerate the canonical harness projections if needed, run focused parent validation, push the
existing parent branch, wait for its checks, and mark the existing PR ready when green.

## Non-goals

- New worker roles, a new parent PR, or a merge to `main`.
- Changes to the destructive-write confirmation semantics.
- Hand edits to `.claude/agents/**`, `.codex/agents/**`, or `.pi/agents/**` generated projections.
- Credentialed connector checks, reverse-ETL execution, dependencies, or CLI-surface changes.
