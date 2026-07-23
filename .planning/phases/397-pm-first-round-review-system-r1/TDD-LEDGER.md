# Issue #397 PM First-Round Review System TDD Ledger

**Status:** plan-check correction applied; RED not yet captured  
**Lifecycle:** active `/pm-orchestrate` owner because `programming-loop` is absent from the 69-command registry  
**Stable lineage:** `0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20...pm-first-round-review-system-r1`  
**Correction budget:** 0/5

| ID | Risk / required behavior | RED evidence | GREEN target | State |
|---|---|---|---|---|
| A495-1 | canonical `final_parent_readiness` is accepted as human-ready | pending | exact current-schema gate kind allowlist is only `parent_ready`, `correction_cap_exceeded`; unknown kind blocks | planned |
| A495-2 | `SEC1` invalid disposition bypasses F/N/R parser | pending | parse every data row in disposition tables independent of identifier prefix | planned |
| M495-1 | prose names predecessor but authoritative state allows dispatch | pending | parsed authority/ready-state relationship rejects prose-only dependencies | planned |
| M495-2 | clean direct root reaches prohibited template transitively | pending | active closure carries edge reasons and rejects prohibited/missing reachable targets | planned |
| M495-3 | replacement head resets/fragments correction count | pending | stable exact-base/candidate-lineage key is monotonic across replacement heads and cap transition blocks | planned |
| STATE-1 | unknown schema/kind or missing canonical discriminator fails open | pending | exact schema and enum validation with explicit legacy read-only handling | planned |
| STATE-2 | stale review evidence remains valid after head change | pending | packet/synthesis exact identities must equal current candidate | planned |
| PACKET-1 | large/mixed review silently truncates yet returns clean | pending | split by threshold; missing/unreviewed/overflow/truncation prevents clean synthesis | planned |
| PACKET-2 | packet coverage is incomplete or duplicated without proof | pending | every changed file assigned; responses declare reviewed, closure, invariants, unreviewed, findings | planned |
| PACKET-3 | unsafe reference or argument escapes repository / leaks content | pending | reject absolute, `..`, symlink, control, option-like and malformed identities; packets contain metadata/paths only | planned |
| PACKET-4 | threshold boundary produces silent over-budget packet | pending | test 20/21/25/26 files, 600/601/800/801 lines, one/two/three domains and per-packet caps | planned |
| CLOSURE-1 | non-Markdown reference, cycle, missing/ambiguous target escapes | pending | typed edge reasons across Markdown/JSON/YAML/frontmatter/script; cycle-safe traversal; missing active target fails | planned |
| LINEAGE-1 | restart or legacy migration reduces count / rewrites heads | pending | event-sequence fixtures prove monotonic rounds, append-only heads, one-way migration, cap boundaries | planned |
| OWNER-1 | packet fan-out creates multiple lifecycle/verdict owners | pending | one PM synthesis; packets are bounded review inputs only | planned |
| SHEPHERD-1 | Shepherd duplicates Codex review or runs before clean | pending | workflow/test requires separate downstream trajectory validation | planned |
| MEASURE-1 | stronger prose is presented as review improvement | pending | separate detection/score steps, historical + opaque cases + controls, machine metrics and limitations | planned |
| MEASURE-2 | same owner tunes against nominally held-out cases after GREEN | pending | freeze/hash opaque corpus and separate oracle before treatment implementation; report fixture-level blinding limitation | planned |
| SCOPE-1 | implementation collides with PR #493-owned paths | pending | exact-base changed-path disjointness rejects forbidden paths | planned |
| SCOPE-2 | implementation expands beyond active PM system | pending | exact positive allowlist rejects new CLI/subcommand, article, dependency, product, connector, or #408 paths | planned |

## Required measurement fields

- first-round true positives, false positives, false negatives;
- recall, precision, escaped-defect rate, false-positive rate;
- exact-version invalidations;
- review rounds and context overflows;
- wall-clock latency;
- input/output tokens and cost when available, otherwise explicit `unavailable` with reason;
- historical, held-out fixture, and prospective observation scopes separated.

## Evidence log

- 2026-07-23: post-#495 base verified at `0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20`; worktree clean; no stash or scratch change carried.
- 2026-07-23: `scripts/gsd doctor` and `scripts/gsd list` passed. `scripts/gsd prompt programming-loop ...` failed because the command is absent. `scripts/gsd prompt plan-phase 397-pm-first-round-review-system-r1 --skip-research` generated and was executed as the planning route. The active `/pm-orchestrate` owner—not a generic manual fallback—owns the lifecycle.
- 2026-07-23: loaded `gsd-core`, `caveman`, `golang-how-to`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-lint`, and `no-mistakes`.
- 2026-07-23: one read-only measurement scout completed. Three parallel read-only scouts did not start because their isolated provider lacked authentication; no result is claimed from them and no retry loop was started. Single cohesive implementation remains local critical path.
- 2026-07-23: read-only plan checker returned BLOCKED. Plan corrected before RED: active PM ownership, complete test-first ordering, path/identity security, preimplementation corpus freeze, durable setup evidence, positive path allowlist, numeric thresholds, exact no-mistakes drive protocol, stable range budget, and final reporting requirements were added.
