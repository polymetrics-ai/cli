# Issue #397 PM First-Round Review System Plan

**Phase:** `397-pm-first-round-review-system-r1`  
**Parent issue:** https://github.com/polymetrics-ai/cli/issues/397  
**Parent PR:** https://github.com/polymetrics-ai/cli/pull/438 (draft, human-only)  
**Sub-PR base:** `feat/cli-architecture-v2`  
**Branch:** `chore/pm-first-round-review-system-r1`  
**Exact starting base:** `0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20`  
**Stable candidate lineage:** `0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20...pm-first-round-review-system-r1`  
**Correction budget:** 0/5 used; changed heads retain this lineage.

## Objective

Implement the captain-authorized, audit-backed first-round PM review system as a separate measured
stacked PR. Deterministic preflight must catch the two accepted PR #495 findings and the three
original preventable misses, compile bounded exact-version review packets, synthesize one local
Codex verdict under one PM owner, keep Shepherd independent/downstream, and publish only measured
fixture/replay claims.

## Authority and source material

- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-first-round-review-audit-r1/report.md`
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-first-round-review-audit-r1/blog-source-notes.md`
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-first-round-review-audit-r1/ship-instructions.md`
- `/Users/karthiksivadas/karthik-agent-workspace/data/decisions/cli-pr495-one-time-review-waiver-and-merge-2026-07-23.md`
- `/Users/karthiksivadas/karthik-agent-workspace/data/decisions/cli-pr495-snyk-deferral-2026-07-23.md`
- PR #495 accepted source head `fc7167990c92292625493f05b495c70e2c7ce886`; squash on parent `0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20`.

Historical PR #495 findings and pending review/Shepherd records remain historical evidence. This
phase does not rewrite them as clean.

## GSD and active PM execution route

- `scripts/gsd doctor`: passed.
- `scripts/gsd list`: passed, 69 commands.
- `scripts/gsd sources plan-phase`: resolved registry, lock, and official docs.
- `scripts/gsd prompt plan-phase 397-pm-first-round-review-system-r1 --skip-research`: generated and executed through Pi tools.
- `scripts/gsd prompt programming-loop ...`: unavailable (`unknown GSD command or prompt: programming-loop`).
- Canonical owner: `/pm-orchestrate`, as required by the post-#495 adapter/workflow when the registry lacks `programming-loop`. This is not a generic manual-GSD fallback. The active owner executes PLAN → RED → GREEN → REFACTOR → VERIFY → REVIEW → INTEGRATE.
- Parent queue for this turn: this promoted review-system follow-up is the selected local critical path; #408/TUI remains out of scope; PR #493-owned paths remain collision-blocked for this branch; parent PR #438 remains draft/human-only.
- Planning-cycle execution decision: `read_only_spawned`; one read-only measurement scout completed, three read-only scouts failed to start because their isolated provider lacked authentication. No mutating worker was spawned. The single cohesive/shared PM-system write scope remains `local_critical_path` in this isolated task worktree.
- Durable setup/fetch/branch evidence: `SETUP-EVIDENCE.md`. Every later cycle appends an explicit decision to `RUN-STATE.json`.

## Required skills loaded

- `gsd-core`
- `caveman`
- `golang-how-to`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-lint`
- `no-mistakes` (delivery gate)

No CLI command/help surface changes; CLI help/manual/website parity is not applicable.

## Scope and files

### Positive path allowlist

- `scripts/pm-review-system.py`: standard-library deterministic compiler, semantic detectors,
  packet synthesis, observations, and scorer.
- `scripts/pm-terminal-classifier.sh`
- `scripts/tests/pm-review-system.sh`
- `scripts/tests/pm-orchestrator-contract.sh`
- `scripts/tests/fixtures/pm-review-system/**`
- `scripts/tests/fixtures/pm-orchestrator-review-state/**` only for focused terminal fixtures.
- `.agents/agentic-delivery/contracts/pm-review-system.json`
- `.agents/agentic-delivery/contracts/pm-review-packet-template.md`
- `.agents/agentic-delivery/contracts/pm-code-review-disposition-template.md`
- `.agents/agentic-delivery/contracts/pm-worker-handoff-template.md`
- `.agents/agentic-delivery/prompts/local-codex-review-prompt.md`
- `.agents/agentic-delivery/workflows/local-codex-review-loop.md`
- `.agents/agentic-delivery/workflows/pi-autonomous-orchestration-loop.md`
- `.agents/agentic-delivery/schemas/orchestration-state.schema.yaml`
- `.pi/agents/pm-reviewer.md`
- `.pi/prompts/pm-review-loop.md`
- `.pi/prompts/pm-auto-loop.md`
- `.pi/prompts/pm-orchestrate.md`
- `.planning/phases/397-pm-first-round-review-system-r1/**`

Any additional production path requires a plan update and write-scope check before editing.

### Forbidden PR #493-owned paths

The focused test rejects any diff from the exact base touching:

- `AGENTS.md`
- `Makefile`
- `.agents/agentic-delivery/matrices/task-skill-matrix.yaml`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/skills/cli-architecture-v2-delivery/**`
- `.planning/phases/397-cli-architecture-v2-delivery-skill/**`
- `scripts/tests/cli-architecture-v2-delivery-skill.sh`

Also excluded: another orchestration architecture, a shipped PM/CLI subcommand, issue #498/article
work, `go.mod`, `go.sum`, #408/TUI, Gong/#497, write-URL product behavior, dependencies,
credentials, connector runtime, raw generic write tools, and reverse ETL behavior.

## Architecture

1. Freeze exact base/head and verify candidate identity.
2. Validate all arguments and discovered references before reading: exact 40-hex commit identity,
   repository-relative allowlisted roots only, no absolute/`..`/control-character/option-like
   paths, no symlink escape, and subprocess argument vectors without shell evaluation. Packets carry
   paths/metadata, never file contents or environment/credential values.
3. Compile changed-file/domain manifest, active transitive required-reference closure with edge
   reasons, and authoritative state/writer/reader registry. Handle cycles. Parse Markdown links and
   required paths plus JSON/YAML/frontmatter/script path references; reject missing or ambiguous
   active targets.
4. Run semantic negative gates for current-schema terminal enums, disposition rows independent of
   finding-ID prefix, dependency/dispatch consistency, transitive prohibited targets, stable
   correction lineage, restart/one-way legacy migration, stale evidence, append-only head history,
   off-by-one correction caps, and authoritative-state disagreement.
5. Pilot packet thresholds are test hypotheses with deterministic boundaries: combine only when
   all three hold—≤20 review-relevant files, ≤600 non-generated changed lines, exactly one domain.
   If any combined limit is exceeded (21–25 files, 601–800 lines, or two domains included), split
   conservatively. More than 25 files, more than 800 lines, or more than two domains is a mandatory
   split. Each split packet allows ≤20 changed files, ≤10 closure/authority files, and a declared
   30K-token target; if safe partitioning cannot satisfy those caps, compilation returns blocked
   rather than truncating. Boundary cases are measured; thresholds are not claimed optimal.
6. Require packet responses to bind exact base/head, list reviewed/closure/invariant/unreviewed
   coverage, disclose overflow/truncation, and report unlimited findings. Missing coverage cannot
   synthesize clean. Preserve raw packet responses and the coverage manifest beside synthesis.
7. PM synthesizes exactly one local-Codex status/disposition. Any changed head invalidates packet
   responses and synthesis.
8. Shepherd independently validates the clean review transition and trajectory; it does not repeat
   code review. Human authority remains final.
9. Detection and scoring run in separate commands so the detector does not receive the oracle.
   Freeze and hash the opaque mutation inputs plus separate oracle in RED before implementing the
   treatment detector. Historical source identities, opaque cases, and clean/metamorphic controls
   produce a machine-readable measurement report. This fixture-level blinding is not a private
   model benchmark; token/cost/prospective data remain explicitly unavailable until captured.

## TDD slices and checkpoints

### Slice 0 — plan

- Create PLAN, TDD ledger, verification checklist, prompt snapshot, run state, and summary.
- Commit and normally push the plan-only checkpoint; no PR opens yet.

### Slice 1 — freeze corpus and capture complete RED baseline

Before any treatment detector or route implementation:

- Add `final_parent_readiness` canonical negative fixture; current classifier must incorrectly
  return `human_ready` before the fix.
- Add an arbitrary `SEC1` invalid-disposition fixture and a configurable test seam; current prefix
  parser must fail to reject it.
- Freeze source-faithful cases for prose-only dependency, transitive prohibited target, replacement
  lineage fragmentation, terminal enum drift, and arbitrary finding ID.
- Freeze opaque independently proposed mutation families plus paired clean/metamorphic controls and
  a separate oracle. Record content hashes and provenance before GREEN.
- Add RED cases for unknown schema/kind, stale exact-head evidence, missing/ambiguous/escaping
  reference paths, closure cycles, threshold boundaries, incomplete packet coverage,
  overflow/truncation, restart, one-way legacy migration, append-only heads, and cap off-by-one.
- Retain explicit baseline detector behavior; treatment initially reproduces current false
  negatives.
- Run each focused command and capture its intended semantic failure signature in this ledger.
- Commit and normally push the RED checkpoint; do not claim green.

### Slice 2 — semantic gates and compiler GREEN

Only after all corresponding RED evidence exists:

- Implement generic semantic detectors and exact current-schema state transitions.
- Fix terminal classifier and disposition parsing while preserving usage, malformed-JSON, legacy,
  stdout/stderr, and exit-code behavior.
- Implement safe active closure, missing-target rejection, authority registry, exact identity,
  changed-file assignment, threshold selection, and forbidden/allowlisted-path checks.
- Run focused tests; commit and normally push the green checkpoint.

### Slice 3 — packet/synthesis route

- First capture RED packet/synthesis failures from Slice 1.
- Implement packet schema/template, bounded packet generation, response coverage validation, stale
  identity rejection, overflow/truncation block, raw-response preservation, and one PM-owned
  synthesis.
- Update active PM review route and state schema. Preserve independent Shepherd ordering.
- Run focused tests; commit and normally push the green/refactor checkpoint.

### Slice 4 — measured replay and pre-frozen blinded fixtures

- Replay compact source-faithful PR #495 cases with exact source identities.
- Run the pre-frozen opaque mutation inputs and paired clean/metamorphic controls with detector
  inputs separated from oracle scoring.
- Capture baseline/treatment precision, recall, escaped defects, false positives, exact-version
  invalidation, rounds, overflow, latency, and available token/cost fields.
- Label deterministic preflight results only; do not claim local-Codex or prospective improvement.
- Commit and normally push the measurement artifact and updated GSD evidence.

### Slice 5 — verification, review, delivery

- Run focused and full gates.
- Commit every tracked implementation, measurement, and GSD evidence artifact and normally push the
  exact green head. No tracked write may follow the final exact-head review gates.
- Compile packets, run the fresh-context packet reviewers, preserve raw responses/coverage outside
  the tracked worktree (and summarize/hash them in the PR delivery record), and synthesize the one
  canonical PM local-Codex verdict at the exact committed head. Run independent Shepherd only after
  clean synthesis. Do not create an evidence commit afterward.
- Inspect `no-mistakes axi`, then run `no-mistakes axi run --intent <complete captain intent>`;
  process every synchronous gate with `no-mistakes axi respond`, inspect `branch_sync`, and continue
  until `checks-passed` or a genuine decision. Never edit while the pipeline owns the run. A
  `passed` outcome means the PR was merged or closed and is a stop/escalation because this stacked
  PR must remain open and unmerged.
- AXI's own review is delivery-pipeline input, not a second PM lifecycle owner or canonical PM
  verdict. Do not launch any parallel/manual reviewer outside the packet system.
- If AXI creates a commit, changes the exact head, or rebases onto a changed base, finish the active
  AXI run first, then rerun applicable full verification, compile fresh packets, obtain one fresh PM
  synthesis, and rerun Shepherd at the new exact identities before claiming readiness. Preserve
  those final raw responses outside the tracked worktree and make no tracked write afterward. A
  head/base change invalidates all prior packet and Shepherd evidence.
- The branch is already published through additive checkpoints. If pipeline behavior proposes a
  rebase/force/non-additive rewrite of published history, stop for Firstmate/captain instead of
  authorizing it.
- Open a Conventional Commit PR title, target `feat/cli-architecture-v2`, use `Refs #397`, and report
  full PR URL, exact source/head, risk, metrics, limitations, and 0–5 correction-round usage.
- Do not merge.

## Human and safety gates

- No secrets or credentialed connector checks.
- No dependency additions.
- No generic shell/HTTP/SQL write tools.
- Reverse ETL remains plan → preview → approval → execute.
- No parent/default-branch merge.
- No Claude or Copilot. The captain ship instructions and current canonical PM route supersede the
  legacy generic automated-review language in `AGENTS.md` for this task.
- Quality gates are not reduced.
