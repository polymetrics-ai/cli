# Issue #397 PM First-Round Review System Plan

**Phase:** `397-pm-first-round-review-system-r1`

**Parent issue:** https://github.com/polymetrics-ai/cli/issues/397

**Parent PR:** https://github.com/polymetrics-ai/cli/pull/438 (draft, human-only)

**Sub-PR base:** `feat/cli-architecture-v2`

**Branch:** `chore/pm-first-round-review-system-r1`

**Exact starting base:** `0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20`

**Stable candidate lineage:** `0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20...pm-first-round-review-system-r1`
**Correction budget:** 1/5 used. Round 1 is the synthesized exact-head findings verdict at `b1d869732d230575ab7c8295b15cef42cc0078ef` followed by this systemic correction. Changed heads retain this lineage. Packet/provider retries are tracked separately and do not consume rounds.

## Objective

Implement the captain-authorized, audit-backed first-round PM review system as a separate measured
stacked PR. Deterministic preflight must catch the two accepted PR #495 findings and the three
original preventable misses; build an evidence-selected, bounded bidirectional practical
file/package impact graph from every changed file plus canonical roots; provide fail-closed
per-packet disposable hypothesis labs without mutating the exact candidate; compile bounded
exact-version review packets; synthesize one local Codex verdict under one PM owner; keep Shepherd
independent/downstream; and publish only measured fixture/replay claims.

## Authority and source material

- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-first-round-review-audit-r1/report.md`
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-first-round-review-audit-r1/blog-source-notes.md`
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-first-round-review-audit-r1/ship-instructions.md`
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-first-round-review-audit-r1/impact-graph-correction.md`
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-first-round-review-audit-r1/counterfactual-review-lab-requirement.md`
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-first-round-review-audit-r1/impact-graph-algorithm-research-requirement.md`
- `.planning/phases/397-pm-first-round-review-system-r1/ALGORITHM-RESEARCH.md`
- `/Users/karthiksivadas/karthik-agent-workspace/data/decisions/cli-pr495-one-time-review-waiver-and-merge-2026-07-23.md`
- `/Users/karthiksivadas/karthik-agent-workspace/data/decisions/cli-pr495-snyk-deferral-2026-07-23.md`
- `/Users/karthiksivadas/karthik-agent-workspace/data/decisions/cli-review-system-conditional-merge-authorization-2026-07-24.md`
- `/Users/karthiksivadas/karthik-agent-workspace/data/decisions/cli-pm-review-loop-monitoring-2026-07-24.md`
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
- 2026-07-24 correction/research cycle: `scripts/gsd prompt plan-phase 397-pm-first-round-review-system-r1 --research` generated the official prompt. Read-only scout/security architecture calls failed to return usable results (WebSocket/provider authentication), so the active PM performed the documented inline GSD fallback. Primary-source research and a disposable exact-head benchmark were completed before graph design; selection is in `ALGORITHM-RESEARCH.md`. Prior verification is historical.

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
  typed bidirectional practical impact graph, packet synthesis, observations, and scorer.
- `scripts/pm-review-lab.py`: standard-library fail-closed exact-head disposable hypothesis-lab runner.
- `scripts/pm-terminal-classifier.sh`
- `scripts/tests/pm-review-system.sh`
- `scripts/tests/pm-orchestrator-contract.sh`
- `scripts/tests/pi-model-routing.sh`
- `scripts/tests/fixtures/pm-review-system/**`
- `scripts/tests/fixtures/pm-review-lab/**`
- `scripts/tests/fixtures/pm-orchestrator-review-state/**` only for focused terminal fixtures.
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
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
- Round-1 systemic parity corrections only: `.agents/agentic-delivery/contracts/parent-issue-roadmap-template.md`, `.agents/agentic-delivery/contracts/issue-prompt-template.md`, `.agents/agentic-delivery/agents/implementation/issue-first-implementation-agent.agent.yaml`, `.agents/connector-migration/{rollout-checklist.md,validation-gates.md,ownership-rules.md}`, `.planning/traces/cli-architecture-v2-pi-prompts.md`, `.planning/phases/397-cli-architecture-v2-orchestration/{RUN-STATE.json,SUMMARY.md}`, and `.planning/traces/cli-architecture-v2-orchestration-state.yaml`.

These additions are limited to conclusive round-1 current-route/authority findings and remain disjoint from the explicit PR #493-owned list below. Any other production path requires a plan update and write-scope check before editing.

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
3. Apply the evidence-selected design in `ALGORITHM-RESEARCH.md`: enumerate the declared tracked
   universe, then materialize a typed directed multigraph with stable forward and reverse adjacency
   before traversal. Seed canonical roots plus every changed file. Parse Markdown/frontmatter,
   JSON/YAML, shell source/exec, Python AST imports/run paths, configured authority/generator/mirror/
   temporal/lateral edges, and authoritative `go list -json -deps -test` package/test/import data.
   Every edge records source, target, relation, direction, parser, reason/location, and
   `active|inactive|unknown` certainty.
4. Compute deterministic cycle-safe multi-source BFS under an explicit relation policy. Include
   bounded unknown context or block when unresolved; do not traverse inactive edges. Configure
   index-byte/file, graph-node/edge, traversal-state/depth, impact-file/edge, Go-command, and packet
   bounds. A continuing frontier, unresolved active/unknown edge, missing target, overflow, or
   truncation blocks. No full symbol-level call/data-flow claim is made. Do not add persistent cache
   or SCC condensation in v1; the measured ~2-second cold rebuild and iterative visited traversal
   were selected over cache invalidation complexity.
5. Run semantic negative gates for current-schema terminal enums, disposition rows independent of
   finding-ID prefix, dependency/dispatch consistency, transitive prohibited targets, stable
   correction lineage, restart/one-way legacy migration, stale evidence, append-only head history,
   off-by-one correction caps, and authoritative-state disagreement.
6. Pilot packet thresholds remain test hypotheses. Discovery completes before packetization.
   Partition changed, closure, authority, impact-file, and impact-edge coverage stably; each packet
   remains within configured file/edge/token limits. If complete policy impact cannot fit, block
   rather than select a top-K subset or truncate.
7. Require versioned packet responses to bind exact base/head/tree; list reviewed changed, closure,
   authority, impact-file, impact-edge, invariant, hypothesis, and unreviewed coverage; disclose
   graph/packet bounds and overflow/truncation; and report unlimited findings. Missing coverage
   cannot synthesize clean.
8. Require observable expert-review behaviors: impact model before line judgment, upstream/
   downstream/lateral/temporal tracing, history/sibling comparison where relevant, explicit
   falsifiable hypotheses and strongest alternatives, disconfirming evidence, smallest useful
   experiment, and honest limitations. Do not claim equivalence to an ideal human reviewer.
9. Run any experiment only with `scripts/pm-review-lab.py`: an exact-head detached disposable copy
   under a private temporary root, scrubbed environment/config, no candidate writes, no generic
   shell, no network/Git mutation/push/install/credential/live/deploy command, proven OS sandbox or
   blocked, time/process/disk/output bounds, bounded evidence, descendant kill, whole-lab
   destruction, and before/after candidate head/tree/status proof. Setup ambiguity, denial, timeout,
   cleanup failure, identity drift, missing evidence, or inconclusive performed experiment blocks
   clean synthesis. Static decisive evidence must state why no experiment was needed.
10. Version impact, packet/response, synthesis, and lab contracts. Explicit migration fixtures reject
    incompatible v1 inputs. The existing `make verify` path already reaches the focused PM tests via
    `scripts/tests/pi-model-routing.sh`; preserve that durable CI route without touching PR
    #493-owned `Makefile`.
11. PM synthesizes exactly one local-Codex status/disposition. Any changed head invalidates packet
    responses, experiments, and synthesis.
12. Shepherd independently validates the clean trajectory after synthesis; it does not edit code or
    repeat code review. Human authority remains final.
13. Detection and scoring remain separate so the detector does not receive the oracle. Freeze a new
    correction corpus/oracle before GREEN and report semantic findings, practical impact coverage,
    lab safety/cleanup, deterministic mutation results, actual packet metrics when available, and
    prospective evidence separately. Fixture evidence is not a private model benchmark.

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

### Slice 5 — captain algorithm research checkpoint

- Pause production graph/lab edits; read all three additive captain requirements.
- Compare primary-source build/query/index/test-impact/graph alternatives capability by capability.
- Benchmark outgoing-only, reverse scan, forward/reverse adjacency BFS, SCC construction, real PM
  and Go seeds, exact-head invalidation, adversarial fixtures, and a large synthetic graph.
- Record selected/rejected algorithms, measured complexity, invariants, gaps, implementation REDs,
  and no-dependency decision in `ALGORITHM-RESEARCH.md`; update plan/TDD/checklist before production.

### Slice 6 — captain impact-graph correction RED/GREEN

Before impact implementation, freeze a separate correction corpus/oracle and add integration REDs
for: an upstream-only template referencer; script/Python downstream+upstream chains; an authority's
writers/readers/mirrors; generator/generated/consumer relations; Go same-package tests and importing
packages; a platform sibling/build-tag unknown; cycles; explicit graph-bound blocking; exact-head
invalidation; and an unrelated control. The old root-only compiler must fail these intended
assertions. Commit/push RED before GREEN.

Then implement the selected typed index, authoritative Go discovery, three-valued certainty,
relation-policy BFS, complete impact manifest, impact packet assignment, migration contract, and
synthesis checks. Measure baseline/treatment impact recall, clean precision, graph coverage, bound
hits, cold/traversal time, memory availability, determinism, and packet consequences. Commit/push
GREEN additively.

### Slice 7 — captain counterfactual-lab correction RED/GREEN

Before lab implementation, add RED tests proving the prior system cannot safely support labs. Cover
a known-defect temporary change, canonical/outside/symlink denial, network/Git mutation/install/
credential/live/deploy denial, scrubbed secrets, timeout/process/output/disk bounds, cleanup and
identity proof, cleanup/identity failure blockers, competing hypotheses, inconclusive blocking,
concurrent isolation, no-experiment clean control, and v1 migration failure. Commit/push RED.

Then implement the fail-closed versioned lab runner and packet response/synthesis contract. An OS
sandbox must prove network/write containment or the experiment stays blocked; a policy-only fallback
cannot authorize clean. Commit/push GREEN additively.

### Slice 7.5 — exact-head local Codex round 1 systemic correction

The exact head `b1d869732d230575ab7c8295b15cef42cc0078ef` compiled 17 packets. All 17 fresh Sol/xhigh contexts completed with no unreviewed files and disclosed 89 raw findings. The machine synthesis returned `blocked` with 14 invariant blockers because failed invariants were misclassified as omitted; that fail-closed result is retained, and the PM-owned disposition is `findings_correction_required`, never clean. `REVIEW-R1-DISPOSITION.md` maps every finding into 15 systemic groups and `REVIEW-R1-MEASUREMENT.json` records hashes, latency, and 22 provider attempts/5 operational failures separately from the 1/5 correction budget.

Before production edits, add RED fixtures for: exact compile/synthesis identity and ready-manifest binding; strict response/status/invariant/lab-evidence shape; relation-state BFS; coherent exact-blob packet bounds; parser/certainty/endpoint/deletion/prohibited-target handling; pre-index resource bounds; offline external-module and deleted-Go impact; default-deny lab read/Git/process containment; explicit-null terminal schema; reusable per-run scope; route parity; authority state/schema parity; remote-no-network reviewer identity; and explicit root/source document references. Then fix each root cause once, rerun focused gates, and compile a fresh exact head. A repeated group in a later round triggers diagnosis rather than another local patch. At 5/5, automatic correction stops; checks are never weakened and lineage never resets.

### Slice 8 — verification, review, delivery

- Treat all verification before the captain corrections as historical. Run focused and full gates
  only after both correction GREENs on one exact committed head.
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
- Do not merge in this execution. The 2026-07-24 decision conditionally authorizes Firstmate—not
  this agent—to use the guarded parent-branch merge path only after implementation, measurement,
  exact-head packet review, independent Shepherd, no-mistakes, and CI are all green. Report the green
  open PR and landed-commit verification remains Firstmate-owned if that authorization is exercised.

## Human and safety gates

- No secrets or credentialed connector checks.
- No dependency additions.
- No generic shell/HTTP/SQL write tools.
- Reverse ETL remains plan → preview → approval → execute.
- No merge by this agent. Firstmate has conditional authority for this stacked PR into the parent
  branch only after every gate; parent PR #438 into `main` remains draft/human-only.
- No Claude or Copilot. The captain ship instructions and current canonical PM route supersede the
  legacy generic automated-review language in `AGENTS.md` for this task.
- Quality gates are not reduced.
