# Issue #397 PM First-Round Review System Plan

**Phase:** `397-pm-first-round-review-system-r1`

**Parent issue:** https://github.com/polymetrics-ai/cli/issues/397

**Parent PR:** https://github.com/polymetrics-ai/cli/pull/438 (draft, human-only)

**Sub-PR base:** `feat/cli-architecture-v2`

**Branch:** `chore/pm-first-round-review-system-r1`

**Exact starting base:** `0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20`

**Stable candidate lineage:** `0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20...pm-first-round-review-system-r1`
**Correction budget:** 2/5 used. Round 1 is the synthesized exact-head findings verdict at `b1d869732d230575ab7c8295b15cef42cc0078ef`; round 2 is the synthesized recurrence verdict at `92ce5e6a19cb7562aead8b224e6ba8dcc0857d34`. Changed heads retain this lineage. Packet/provider/OAuth/WebSocket/context-window attempts are tracked separately and do not consume rounds.

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
- `/Users/karthiksivadas/karthik-agent-workspace/data/decisions/cli-pm-review-f1-correction-authorization-2026-07-25.md`
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-pm-review-claude-fresh-assessment-r1/report.md` (advisory only; not PM review coverage)
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-first-round-review-audit-r1/f1-correction-and-self-review-instructions.md`
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
- Cycle log (condensed; full detail in `RUN-STATE.json` and git history): planning was `read_only_spawned` then `local_critical_path` under `/pm-orchestrate` (no `programming-loop` in the registry); round-2 route/docs/state retry ran locally. Setup/fetch/branch evidence is in `SETUP-EVIDENCE.md`; each cycle appends a decision to `RUN-STATE.json`. The 2026-07-24 research cycle selected the graph design in `ALGORITHM-RESEARCH.md`. The implementation-recovery cycle closed five lab-fixture issues (aggregate `changed_paths`, a UID-derived `__CF_USER_TEXT_ENCODING` under unchanged kill-on-denial, the `internal/lib` fixture path, a numeric loopback service replacing `python -m http.server`, and `local_only`). Exact-commit RED `1e640f9a4` exposed the 77/64 packet block; bounded efficiency reached 63 packets ≤30,000 bytes.
- 2026-07-25 captain-authorized advisory correction: the captain halted the partial prospective R3 at `c28126bb3` (no synthesis/Shepherd). Advisory F1–F6 are authorized without consuming a correction round (lineage stays 2/5); the manual GSD PLAN → RED → GREEN → REFACTOR → VERIFY → REVIEW path is active. Permanent RED tests precede every production correction; the broad refactor is deferred; hard 30,000-byte/64-packet limits, quality gates, and no-merge authority are unchanged. Baseline at `c28126bb3`: ready, 63 packets, max 29,994 bytes, 1.018025x duplication, headroom 1, 14.24s.

## Required skills loaded

- `gsd-core`
- `caveman`
- `golang-how-to`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-lint`
- `golang-cli`
- `golang-documentation`
- `golang-context`
- `golang-concurrency`
- `golang-database`
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
- Round-1 systemic parity corrections only: `.agents/agentic-delivery/contracts/parent-issue-roadmap-template.md`, `.agents/agentic-delivery/contracts/issue-prompt-template.md`, `.agents/agentic-delivery/agents/implementation/issue-first-implementation-agent.agent.yaml`, `.agents/connector-migration/{rollout-checklist.md,validation-gates.md,ownership-rules.md,templates/connector-rollout-prompt.md}`, `.planning/traces/cli-architecture-v2-pi-prompts.md`, `.planning/phases/397-cli-architecture-v2-orchestration/{RUN-STATE.json,SUMMARY.md}`, and `.planning/traces/cli-architecture-v2-orchestration-state.yaml`.
- Round-2 route/documentation/state parity only: `.agents/agentic-delivery/workflows/{parent-issue-orchestration-loop.md,gsd-universal-runtime-loop.md,pi-active-orchestration-loop.md}`, `.pi/{README.md,prompts/pm-connector-loop.md}`, `.agents/connector-migration/{ownership-rules.md,rollout-checklist.md,validation-gates.md,templates/connector-rollout-prompt.md,agents/implementation/passb-expander.agent.yaml}`, `.planning/traces/{cli-architecture-v2-pi-prompts.md,cli-architecture-v2-orchestration-state.yaml}`, and `.planning/phases/397-cli-architecture-v2-orchestration/{PLAN.md,RUN-STATE.json,SUMMARY.md,VERIFICATION.md}`.

These additions are limited to conclusive current-route/authority findings and remain disjoint from the explicit PR #493-owned list below. Shepherd workflow/prompt, driver scripts, trace driver, PM review compiler/lab implementation, their tests, and the GitHub connector operation ledger are owned by other round-2 workstreams and are not edited in this slice. Any other production path requires a plan update and write-scope check before editing.

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

## Architecture (condensed; full detail in `ALGORITHM-RESEARCH.md` and the compiler source)

Freeze and verify exact base/head/tree identity; validate every argument/reference (40-hex commits,
allowlisted repo-relative roots, no absolute/`..`/control/option/symlink paths, no shell eval;
packets carry paths/metadata only, never contents or secrets). Build the evidence-selected typed
directed multigraph with forward/reverse adjacency before a deterministic cycle-safe multi-source
relation-policy BFS, seeded by canonical roots plus every changed file, with three-valued certainty
and fail-closed index/graph/traversal/impact/Go/packet bounds (frontier/unresolved/missing/overflow
blocks; no symbol-level claim; no v1 cache/SCC). Run semantic negative gates (terminal enums,
prefix-independent disposition rows, dependency/dispatch, transitive prohibited targets, stable
lineage, one-way migration, stale evidence, append-only heads, cap off-by-one, authoritative-state
disagreement). Discover fully before packetizing; partition changed/closure/authority/impact
coverage within bounds or block. Versioned packet responses bind exact base/head/tree and echo
reviewed/closure/authority/impact/edge/invariant/hypothesis/unreviewed coverage with
overflow/truncation disclosure and unlimited findings; missing coverage cannot synthesize clean.
Require observable expert behaviors (impact model first, four-way tracing, history/sibling checks,
falsifiable hypotheses, disconfirming evidence, smallest experiment, honest limits). Experiments run
only via the fail-closed `scripts/pm-review-lab.py` disposable exact-head lab. Version all contracts;
`make verify` reaches the focused PM tests via `scripts/tests/pi-model-routing.sh` without touching
the PR #493-owned `Makefile`. The PM synthesizes exactly one disposition; any head change
invalidates it; independent Shepherd runs downstream of clean synthesis; detection and scoring stay
separate and fixture evidence is never a private model benchmark.

## TDD slices and checkpoints

### Slice 0 — plan

- Create PLAN, TDD ledger, verification checklist, prompt snapshot, run state, and summary.
- Commit and normally push the plan-only checkpoint; no PR opens yet.

### Slice 1 — freeze corpus and capture complete RED baseline (condensed)

Before any treatment, froze detector-visible fixtures and a separate oracle with recorded hashes:
`final_parent_readiness`, invalid `SEC1` disposition, prose-only dependency, transitive prohibited
target, replacement-lineage fragmentation, terminal enum drift, arbitrary finding id, opaque mutation
families with paired clean/metamorphic controls, and RED for unknown schema/kind, stale evidence,
unsafe reference paths, closure cycles, threshold boundaries, incomplete/overflow packet coverage,
restart, one-way migration, append-only heads, and cap off-by-one. Baseline detector reproduced the
current false negatives; RED committed before GREEN.

### Slices 2–7 — gates, packets, measurement, research, impact graph, and lab (condensed)

Each slice landed RED before GREEN and pushed additive checkpoints. Slice 2 implemented generic
semantic detectors, current-schema transitions, the terminal classifier/disposition parser (usage,
malformed-JSON, legacy, stdout/stderr, exit-code preserved), safe active closure, missing-target
rejection, authority registry, exact identity, changed-file assignment, and forbidden/allowlisted
checks. Slice 3 added the packet schema/template, bounded generation, response coverage validation,
stale-identity/overflow blocks, raw-response preservation, one PM synthesis, and the state schema
with Shepherd downstream. Slice 4 replayed PR #495 cases and the pre-frozen opaque mutations with
detector inputs separated from oracle scoring, reporting deterministic-only precision/recall/escape
metrics. Slice 5 was the captain algorithm-research checkpoint recorded in `ALGORITHM-RESEARCH.md`.
Slice 6 implemented the typed index, authoritative Go discovery, three-valued certainty,
relation-policy BFS, complete impact manifest/packet assignment, and migration/synthesis checks
against a separate frozen corpus. Slice 7 implemented the fail-closed versioned `pm-review-lab.py`
runner and packet/synthesis contract, requiring a proven OS sandbox (policy-only cannot authorize
clean).

### Slice 7.5 — exact-head round 1 systemic correction (condensed)

Head `b1d869732` compiled 17 packets; 17/17 fresh contexts, 89 findings, synthesis fail-closed to
`findings_correction_required` (14 invariant blockers). `REVIEW-R1-DISPOSITION.md` maps them to 15
systemic groups; `REVIEW-R1-MEASUREMENT.json` records hashes/latency and 22 provider attempts/5
operational failures separate from the 1/5 budget. RED fixtures were frozen for exact identity,
strict response/invariant/lab shape, relation-state BFS, exact-blob packet bounds,
parser/certainty/endpoint handling, pre-index bounds, offline/deleted Go impact, default-deny lab
containment, explicit-null terminal schema, per-run scope, and route/authority/reviewer parity
before each root cause was fixed once and recompiled. Repeated groups trigger diagnosis, not local
patches; at 5/5 automatic correction stops; checks/lineage are never weakened.

### Slice 7.6 — exact-head recurrence round 2 systemic correction (condensed)

Head `92ce5e6a1` compiled 41 packets; 41/41 contexts, synthesis `findings_correction_required` with
141 findings, zero blockers. `REVIEW-R2-DISPOSITION.md` maps every finding to R1-A–R1-O or six new
groups; `REVIEW-R2-MEASUREMENT.json` records 44 attempts and one context-window rejection. Round 2
replaced mechanisms whose R1 tests validated their own labels with independent-invariant RED
(same-head tampering, exact response/lab binding, immutable exact-commit compile, one-token-per-byte
rendered bounds, format-aware parsing with base+head provenance, phase/mirror validation, Go/graph
limits, default-deny lab, bounded Shepherd, trace redaction, route parity, corpus binding). The
bounded route/docs/state retry (R1-I/K/L, R2-N2/N5) reached GREEN with v4 compile → render →
synthesize → clean head → Shepherd ordering, exact disposition values, truthful vendored
`.pi/extensions/pi-sub-agent` provenance, a canonical current-state overlay with legacy history
labeled read-only, and explicit PM-worker vs coordinator-fanout connector delivery profiles.

The correction may expand the positive scope to the non-forbidden Shepherd/trace/runtime and parent-plan surfaces named in the disposition. It must not touch any PR #493-owned path. GitHub connector pending-review write coverage remains a disclosed `needs_human` product/auth decision; the PM impact graph must not pull that unrelated pre-existing ledger into this candidate through false ownership/provenance edges.

GREEN requires versioned exact machine contracts, deterministic fail-closed behavior without tracebacks, independently authenticated coverage, safe bounded prompts, and semantic parity fixtures. Shepherd still runs only after a future exact-head clean synthesis. The five-round lineage never resets.

### Slice 7.7 — captain-authorized advisory correction and review-system self-test

The advisory route is not R3 coverage and does not consume a correction round; usage stays 2/5.
Before production edits, add permanent RED regressions for: accepted `combined` domain rules silently
losing changed files; byte-for-byte changed-diff assignment for every file; missing actionability
criterion; directives to unavailable reviewer contracts; undisclosed per-impact-file provided/total
bytes; duplicate occurrence ids on replay; and missing/misstated packet headroom. Then apply only the
smallest safe F1/F2/F3/F4/F5/F6 corrections. Packing domains are data-driven and a closing production
postcondition checks every changed file and its complete exact diff bytes. Impact disclosure records
provided and total bytes without claiming full-file review. The 30,000-byte and 64-packet caps are
immutable.

Measure before/after compile wall time, peak memory, packet count, max rendered bytes, duplication,
and first-class packet headroom. Take no performance change unless repeated measurement shows a
safe win with identical guarantees. Prove deterministic semantic authentication across three
independent exact-head compiles. After one clean commit and exact clean-tree full verification,
compile a new trust-bound manifest and run the corrected system over itself, including an explicit
reintroduced-F1 mutation that must block with `packet_coverage`. Preserve one PM synthesis even when
it returns findings; do not silently iterate. Shepherd remains clean-synthesis-only.

Final outcome: the required honest self-contained F2/F3 prompt plus a legitimately larger impact
graph initially raised the accumulated candidate to 64 packets. CI (Linux) produced a slightly
larger impact graph than local (macOS), tipping it to 65 (> the 64 hard cap) and blocking the
`verify` job at zero headroom. Condensing superseded process narration in `TDD-LEDGER.md`,
`VERIFICATION.md`, `PLAN.md`, and `ALGORITHM-RESEARCH.md` (frozen R1/R2 dispositions and the other
workstream's orchestration-state files left untouched) restored the candidate to **63 packets,
headroom 1, ready**, every rendered prompt <= 30,000 bytes, 0 overflow, within the unchanged
64-packet hard cap. No limit, bound, gate, quality check, or disclosure was weakened; SHAs, lineage,
F1-F6 records, and disposition pointers are preserved.

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
