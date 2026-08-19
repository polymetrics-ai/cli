# Agent Harness & Pi CLI Improvement Plan

Status: **PLAN ONLY — no implementation yet.** Created from a codebase sweep of `.pi/`,
`.agents/agentic-delivery/`, `.codex/`, `.opencode/`, `.planning/`, plus web research on
autonomous-agent best practices. This document records findings and proposed changes; it does
not edit any runtime/policy files. Templates and existing contracts are preserved — changes
extend, not rewrite.

---

## 1. Research basis (web search)

Sources reviewed (full text captured to `/tmp` during research):

1. **Anthropic — "Building effective agents"** (Dec 2024, `anthropic.com/engineering/building-effective-agents`).
2. **OpenAI — "A practical guide to building agents"** (PDF, `cdn.openai.com/.../a-practical-guide-to-building-agents.pdf`).
3. **OpenAI — "Harness engineering"** index (`openai.com/index/harness-engineering/`).
4. **LangChain — "The Anatomy of an Agent Harness"** (Mar 2026, `langchain.com/blog/the-anatomy-of-an-agent-harness`).
5. DuckDuckGo result sets for "autonomous coding agents implementation harness" and
   "AI coding agent harness sandbox worktree best practices".

### Distilled best practices (the yardstick)

**A. Agent = Model + Harness.** Everything that is not the model is the harness: system
prompts, tools/skills/MCPs + their descriptions, bundled infrastructure (filesystem, sandbox,
browser), orchestration logic (subagent spawning, handoffs, model routing), and
hooks/middleware for deterministic execution (compaction, continuation, lint checks).
(LangChain)

**B. Agent = Model + Tools + Instructions + Guardrails.** (OpenAI) Tools fall into three
types: *Data* (retrieve context), *Action* (interact/take actions), *Orchestration* (agents
as tools). Standardized, many-to-many, well-documented, reusable tool definitions.

**C. Start simple; add complexity only when it demonstrably improves outcomes.** (Anthropic
core principle #1.) Prefer the simplest solution; workflows (predefined paths) for
predictability, agents (dynamic, model-directed) for flexibility.

**D. Three Anthropic core principles for agents:**
1. Maintain simplicity in design.
2. **Prioritize transparency — explicitly show the agent's planning steps.**
3. **Carefully craft the agent-computer interface (ACI)** through thorough tool documentation
   and testing. Invest as much in ACI as in HCI. Poka-yoke tools (make mistakes hard, e.g.
   absolute filepaths). Give the model tokens to "think." Keep tool formats close to natural
   text.

**E. Orchestration.** (OpenAI) Maximize a single agent first (add tools incrementally); split
into multi-agent only on *complex logic* (many if-then-else branches) or *tool overload*
(overlapping tools). Two multi-agent patterns: **Manager** (agents as tools, central control)
and **Decentralized** (peer handoffs). Every orchestration needs a **"run" loop with explicit
exit conditions** (final-output tool, no tool call, error, max turns). Prefer **prompt
templates with policy variables** over many individual prompts.

**F. Model selection.** (OpenAI) Prototype with the most capable model per task to establish
a baseline, then swap in smaller models for cost/latency where evals show acceptable results.
Not every task needs the smartest model (e.g., routing/classification → smaller model).

**G. Agents must get "ground truth" from the environment each step** (tool results, code
execution) to assess progress; pause for human feedback at checkpoints/blockers; include
**stopping conditions** (max iterations). Extensive testing in **sandboxed environments** +
guardrails. (Anthropic)

**H. Guardrails.** (OpenAI) Focus on data privacy + content safety first; layer additional
guardrails based on real-world edge cases; optimize for both security and UX. Incremental,
not upfront-perfect.

**I. Coding agents are a strong fit** because output is verifiable via tests, agents can
iterate using test results as feedback, the problem space is well-defined, and quality is
objectively measurable — but **human review remains crucial**. (Anthropic)

**J. Workflows seen in production** (Anthropic): prompt chaining, routing, parallelization
(sectioning/voting), orchestrator-workers, evaluator-optimizer. The project's
parent-issue-orchestrator → worker handoff → review-disposition chain is an
**orchestrator-workers + evaluator-optimizer** composition — a recognized good pattern.

---

## 2. Current-state assessment (what is already good)

Mapped against the yardstick above, the project is genuinely strong on policy and
guardrails:

- **Single source of truth policy.** `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
  is the shared contract; no runtime forks its rules. (Matches best practice C — simplicity,
  one policy.)
- **Orchestrator-workers + evaluator-optimizer composition.** Parent orchestrator →
  worker handoff → review-disposition is a recognized Anthropic pattern (J).
- **Explicit, mandatory spawn-decision vocabulary.** `spawned | read_only_spawned |
  local_critical_path | not_spawned_*` with eight blocker reasons, consistent across the
  universal loop, parent loop, Codex/Pi adapters, contracts, schema, and agent YAML.
  (Matches G — transparency of planning steps; D2.)
- **Absolute human gates.** Parent-PR merge, auth-scope changes, secrets, new dependencies,
  destructive actions, quality-gate reductions, generic write tools — all hard-stopped
  consistently. (Matches H.)
- **Path/tool permissions per role.** 12 agent YAMLs with `allowed_paths`,
  `conditional_denied_paths`, `denied_tools`; read-only reviewer locked to
  `read,grep,find,ls`. (Matches B, D3 — ACI scoping.)
- **Observable/auditable artifact set.** PLAN / TDD-LEDGER / VERIFICATION / RUN-STATE /
  AGENT-ORCHESTRATION / PROMPTS / traces. (Matches D2 — transparency.)
- **Prompt templates with argument placeholders + `argument-hint`** (`.pi/prompts/*.md`).
  (Matches E — templates over many prompts.)
- **Read-only subagents for recon/review** with isolated context. (Matches A, E —
  delegation with isolated context.)

---

## 3. Gap analysis (failures vs best practices)

Organized by theme. Each gap cites the offending file(s) and the best-practice letter it
violates.

### 3.1 Pi is not a first-class runtime (central gap) — A, E

- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` "Runtime Adapters"
  section lists **Claude Code, Codex, OpenCode** but **no `### Pi`** section. Pi appears only
  generically in the "Treat spawn generically" line.
- `.agents/agentic-delivery/agents/coordination/parent-issue-orchestrator.agent.yaml`
  `adapters:` block has `claude_code`, `codex`, `opencode`, `github_actions` — **no `pi:`**.
- `AGENTS.md` and `docs/prompts/universal-programming-loop-prompts.md` do not mention Pi as a
  runtime, yet `.pi/`, `.pi/prompts/*`, and `pi-active-orchestration-loop.md` exist.
- **Effect:** The Pi runtime is a second-class citizen. A Pi operator cannot discover the
  required load-list, spawn semantics, or isolation rules from the source-of-truth files; the
  `.pi/` adapter is effectively undocumented policy.

### 3.2 No mutating Pi worker agent — A, B, E

- `.pi/agents/` contains only **read-only** `pm-scout` and `pm-reviewer`. The orchestrator
  prompt (`pm-orchestrate.md`) instructs delegation to `.pi/agents/`, but there is **no
  `pm-gsd-worker` / `pm-implement` editor agent**.
- **Effect:** Code-editing work cannot be delegated to an isolated Pi subagent; it must run
  in the main session via `/pm-gsd-loop`. This defeats the orchestrator-workers delegation
  model and forces `mode: inline` (see 3.5). Violates A (orchestration logic: subagent
  spawning) and E (split work across agents for isolation).

### 3.3 No worktree isolation despite contracts requiring it — G, A

- `.planning/config.json` has `"use_worktrees": false`.
- Every mutating-worker contract/adapter requires separate worktrees or `cwd`, else
  `not_spawned_isolation_missing`.
- **Effect:** Config contradicts policy. The phase recorded `mode: inline` /
  `local_critical_path` precisely because isolation+subagent were unavailable. Mutating work
  ran in the main checkout with no isolation, which the contracts forbid for spawned workers.

### 3.4 No model routing / per-role model scoping for cost — F

- `.pi/settings.json` default is `gpt-5.5:xhigh`; `.planning/config.json` maps **every** GSD
  role to `gpt-5.5:xhigh`. Read-only `pm-scout` inherits `xhigh`.
- **Effect:** Best practice F (swap smaller models for cheap tasks) is not applied. Scouting,
  routing, and classification tasks pay for the most capable model unnecessarily. No eval
  baseline exists to justify the uniform xhigh choice.

### 3.5 GSD loop run-state honesty issues (correctness) — D2, G

Investigation of whether the GSD programming loop is used correctly found structural problems
in `.planning/phases/github-projects-discussions/`:

- **Verification gate integrity.** `RUN-STATE.json` reports `verificationPassed: true` and
  gate `make_verify` status, while the same file admits **full `make verify` timed out** and
  "should be re-run in CI before merge." Marking verification *passed* on focused/partial
  gates (gofmt/vet/engine tests/connectorgen) while the full suite timed out is a gate-
  integrity defect. (Violates G — ground-truth environment feedback; D2 — transparency.)
- **Stale `PROMPTS.md`.** Two kickoff snapshots still show `Downstream artifact: pending` and
  `Verification result: pending`, while `RUN-STATE.json` is at
  `review_fix_local_verified`. The loop did not update prompt-snapshot outcomes as it
  progressed. (Violates D2.)
- **Spawn-decision contradiction.** `PLAN.md` claims read-only sidecars were `spawned`, but
  `AGENT-ORCHESTRATION.json` records `mode: inline` with reason "Subagent tools unavailable;
  using sequential inline role passes." Inline role passes are **not** spawns; the
  spawn-decision vocabulary was applied loosely. (Violates D2, G.)
- **Missing per-cycle spawn decisions.** The universal loop requires **one explicit
  execution decision per cycle**. `RUN-STATE.json` records only **one**
  `orchestrationDecisions` entry (`local_critical_path`) for the entire lifecycle, despite a
  multi-phase lifecycle (plan → tdd → execute → verify → gap-loop). Per-cycle decisions were
  not captured. (Violates the loop's own contract.)

### 3.6 ACI / tool-description gaps — D3

- Read-only Pi agents request `grep`/`find`/`ls`, but pi's default active tool set is only
  `read,bash,edit,write`. The README documents the explicit
  `--tools read,bash,edit,write,grep,find,ls,subagent --approve` launch, but the agents
  themselves don't state the required parent allowlist — a poka-yoke gap (D3): an agent can
  be launched without its required tools and fail opaquely.
- Agent YAML `workflow:` steps are terse verbs (`load_gsd_programming_loop`,
  `implement_minimal_slice`) with no embedded prompts; execution depends on referenced
  Markdown. This is acceptable for policy scaffolds but risks divergence between the YAML and
  the Markdown it references (already observed: the universal loop claims
  `.opencode/agents/gsd-worker.md` exists, but it does not — only the command file does).

### 3.7 Cross-reference drift in adapters — C, D2

- `parent-issue-orchestration-loop.md` step 3 load-list omits
  `automated-review-routing-loop.md` and `claude-review-loop.md` (covered only indirectly
  via the parent-orchestrator contract).
- `codex-active-orchestration-loop.md` load-list omits `stacked-parent-subissue-workflow.md`
  and the review loops.
- OpenCode has no runtime-specific adapter file (relies on the universal loop's OpenCode
  section); Codex has one (`codex-active-orchestration-loop.md`); Pi has one
  (`pi-active-orchestration-loop.md`) but it is unregistered (3.1). Asymmetry invites drift.

### 3.8 No deterministic hooks/middleware surfaced — A

- Pi supports compaction (`reserveTokens`, `keepRecentTokens`) and retry, configured in
  `.pi/settings.json`. But there are **no deterministic hooks** for the loop's required
  artifacts: e.g., a hook that refuses to mark `verificationPassed: true` unless `make verify`
  exited 0, or a hook that auto-stamps `PROMPTS.md` outcomes. The harness relies entirely on
  the model to update state honestly, which is exactly where 3.5 failed.

---

## 4. Proposed improvements (prioritized, phased)

All changes **preserve templates** — the universal loop, contracts, and agent YAMLs are
extended, not rewritten. No implementation in this plan; this is the change set to execute
later under the GSD loop.

### Phase 1 — Register Pi as a first-class runtime (fixes 3.1, 3.7)

1. Add a `### Pi` subsection to `gsd-universal-runtime-loop.md` "Runtime Adapters" covering:
   settings (`settings.json`), prompts (`.pi/prompts/*.md` with `$@`/`$1` + `argument-hint`),
   agents (`.pi/agents/*.md`), subagent mechanism (`subagent` tool with per-task `cwd`,
   concurrency 8 total / 4 concurrent), skills (`.agents/skills/<name>/SKILL.md`), required
   launch (`--tools read,bash,edit,write,grep,find,ls,subagent`), and isolation rules
   (per-worker `cwd`/worktree, else `not_spawned_isolation_missing`).
2. Add a `pi:` entry to `parent-issue-orchestrator.agent.yaml` `adapters:` mirroring the
   Codex/OpenCode notes (thin `.pi/` wrappers over the YAML; `subagent` for workers;
   worktree/`cwd` required for mutating workers).
3. Add Pi to the runtime map in `docs/prompts/universal-programming-loop-prompts.md` and to
   the `AGENTS.md` "shared runtime policy" sentence (currently "Codex, Claude, OpenCode, and
   future agents" → add Pi explicitly).
4. Tighten cross-references (3.7): add `automated-review-routing-loop.md` and
   `claude-review-loop.md` to `parent-issue-orchestration-loop.md` step 3; add
   `stacked-parent-subissue-workflow.md` + review loops to
   `codex-active-orchestration-loop.md` step 3.

### Phase 2 — Add a mutating Pi worker agent + isolation (fixes 3.2, 3.3)

1. Create `.pi/agents/pm-gsd-worker.md`: a mutating worker agent scoped to one issue / one
   branch / one `cwd`, tools `read,bash,edit,write,grep,find,ls` (no `subagent` — no
   recursive delegation), thinking `high`, model `gpt-5.5:high` (cheaper than orchestrator),
   following `issue-agent-contract.md` + the universal loop. Guardrails: never merge parent
   PR, never edit shared parent artifacts unless granted, stop on human gates.
2. Wire `pm-orchestrate.md` to dispatch `pm-gsd-worker` via `subagent` with per-task `cwd`
   for independent ready sub-issues; keep `local_critical_path` for coupled slices.
3. Reconcile worktree policy: either (a) flip `.planning/config.json` `use_worktrees: true`
   and document the worktree convention for Pi, or (b) keep `false` but mandate per-worker
   `cwd` isolation in the Pi adapter and record `not_spawned_isolation_missing` when it
   cannot be provided. Pick one and make config + contracts agree.

### Phase 3 — Model routing for cost efficiency (fixes 3.4)

1. Add per-agent `model` overrides in `.pi/agents/*.md` frontmatter and
   `.planning/config.json` `model_overrides`: `pm-scout` → `gpt-5.4-mini:high` (recon),
   `pm-gsd-worker` → `gpt-5.5:high` (implementation), keep `pm-reviewer` and orchestrator on
   `gpt-5.5:xhigh`.
2. Document the model-selection rationale (best practice F) in
   `docs/prompts/universal-programming-loop-prompts.md`: "prototype on xhigh, route cheap
   models to recon/classification once evals pass."
3. Record an eval baseline note in the phase `VERIFICATION.md` so future downgrades are
   evidence-based, not arbitrary.

### Phase 4 — GSD loop honesty & deterministic hooks (fixes 3.5, 3.8)

1. **Gate-integrity rule.** Add to `gsd-universal-runtime-loop.md` and
   `VERIFICATION.md` template: `verificationPassed` may be `true` **only** when the full
   `make verify` (or the declared equivalent) exits 0. A timeout or partial run must record
   `verificationPassed: false` / `blocked` with the failing gate named, even if focused
   gates pass. Update `RUN-STATE.json` schema docs accordingly.
2. **Per-cycle spawn-decision capture.** Add a step to the universal loop requiring a
   `orchestrationDecisions` entry **per lifecycle cycle**, not one for the whole phase;
   update `RUN-STATE.json` and the `orchestration-state.schema.yaml` to validate this.
3. **PROMPTS.md outcome stamping.** Add a loop step: when a cycle completes, update the
   originating `PROMPTS.md` snapshot's `Downstream artifact` and `Verification result`
   fields (no stale `pending`). Add a deterministic check (script or hook) that fails the
   gap-loop if any snapshot is stale while RUN-STATE is terminal.
4. **Spawn-decision honesty.** Clarify in the universal loop that inline role passes are
   `local_critical_path` or `not_spawned_runtime_capability_missing`, **never** `spawned`.
   Update the `github-projects-discussions` phase artifacts to correct the record
   (PLAN.md claim vs AGENT-ORCHESTRATION.json mode).
5. **Deterministic hooks (best-effort, Pi-supported).** Explore a small `scripts/gsd-check`
   helper invoked via `bash` at gate points: validate RUN-STATE vs actual `make verify` exit,
   flag stale PROMPTS.md, assert per-cycle decisions exist. Keep it advisory first, enforce
   later.

### Phase 5 — ACI poka-yoke & doc cleanup (fixes 3.6)

1. Add a "Required parent tool allowlist" note to each `.pi/agents/*.md` and to the worker
   agent (Phase 2) so an agent launched without its tools fails fast with a clear message.
2. Fix the universal-loop claim that `.opencode/agents/gsd-worker.md` exists (only the
   command file does) — either create the agent file or correct the reference.
3. Add `argument-hint` examples to all `.pi/prompts/*.md` (issue numbers / PR URLs / branch
   names) so the model parses structured args instead of inferring from free text.
4. Reconcile Codex/OpenCode worker asymmetry: decide whether Codex needs a
   `gsd-loop-worker.toml` (currently absent) or document that Codex workers are spawned as
   default agents with the contract pasted in.

### Phase 6 — Harness efficiency tuning (best-practice-aligned, low risk)

1. Confirm compaction settings (`reserveTokens: 16384`, `keepRecentTokens: 20000`) are
   appropriate for xhigh reasoning loops; document the rationale in `.pi/README.md`.
2. Use the `caveman` skill consistently for cross-agent handoffs (already required for
   coordinators) — verify all orchestrator/worker prompts reference it for status/handoff
   fields only (never for code/commands/safety text).
3. Add a `pi` non-interactive invocation example to `.pi/README.md` for CI/parent-PR review
   coverage (using `agentScope both` + `confirmProjectAgents` guidance).

---

## 5. GSD programming-loop correctness — verdict

**Partially correct, with honesty defects.** The loop is instantiated with the right
artifacts and follows the right lifecycle order (plan → TDD gate → execute → verify →
gap-loop → summary), and TDD was genuinely applied for new slices
(`TestBundleLoadRejectsGraphQLVariableDefaultTypeMismatch`,
`TestReadGraphQLBodyOmitsExplicitlyEmptyQueryVariable`). However:

- The **verification gate is dishonestly marked passed** despite `make verify` timing out
  (3.5).
- **Spawn decisions are mislabeled** (inline passes recorded as spawned) and
  **under-captured** (one decision for a multi-cycle lifecycle) (3.5).
- **PROMPTS.md is stale** (3.5).
- **Config contradicts policy** on worktree isolation (3.3).

These are exactly the transparency/ground-truth failures best practices D2 and G warn
against. Phase 4 addresses them.

---

## 6. Non-goals / out of scope

- Rewriting the universal loop, contracts, or agent YAMLs wholesale (templates preserved).
- Changing the connector-architecture-v2 program or legacy connector Go.
- Adding generic shell/HTTP/SQL write tools (forbidden by AGENTS.md).
- Merging PR #74 or parent PR #49 (human-gated).
- Inventing run evidence for trace files (traces remain DRAFT until cross-checked against
  real run logs).

---

## 7. Suggested execution order

Phase 1 → 2 → 4 → 3 → 5 → 6. Phases 1 and 2 unblock real Pi orchestration; Phase 4 restores
gate honesty before further automated work trusts RUN-STATE; Phase 3/5/6 are efficiency and
polish. Each phase should be delivered as one issue/PR under the GSD loop with TDD where
behavior changes (Phase 4 hooks, Phase 2 worker) are involved.
