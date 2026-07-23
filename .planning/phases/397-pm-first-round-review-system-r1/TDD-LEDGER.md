# Issue #397 PM First-Round Review System TDD Ledger

**Status:** deterministic treatment and packet route GREEN; full verification/review pending

**Lifecycle:** active `/pm-orchestrate` owner because `programming-loop` is absent from the 69-command registry

**Stable lineage:** `0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20...pm-first-round-review-system-r1`
**Correction budget:** 0/5

| ID | Risk / required behavior | RED evidence | GREEN target | State |
|---|---|---|---|---|
| A495-1 | canonical `final_parent_readiness` is accepted as human-ready | classifier returned `human_ready`, wanted `blocked_human_decision` | exact current-schema gate kind allowlist is only `parent_ready`, `correction_cap_exceeded`; unknown kind blocks | green |
| A495-2 | `SEC1` invalid disposition bypasses F/N/R parser | focused validator exited 0 with canonical F1 control plus invalid SEC1 row | parse every data row in disposition tables independent of identifier prefix | green |
| M495-1 | prose names predecessor but authoritative state allows dispatch | treatment missed historical and renamed opaque dependency cases | parsed authority/ready-state relationship rejects prose-only dependencies | green |
| M495-2 | clean direct root reaches prohibited template transitively | treatment missed two- and three-hop cross-format prohibited targets | active closure carries edge reasons and rejects prohibited/missing reachable targets | green |
| M495-3 | replacement head resets/fragments correction count | treatment missed historical fragmented map and opaque replacement/resume reset | stable exact-base/candidate-lineage key is monotonic across replacement heads and cap transition blocks | green |
| STATE-1 | unknown schema/kind or missing canonical discriminator fails open | treatment missed explicit `canonical_v9`/unknown kind | exact schema and enum validation with explicit legacy read-only handling | green |
| STATE-2 | stale review evidence remains valid after head change | treatment missed candidate/packet exact-head mismatch | packet/synthesis exact identities must equal current candidate | green |
| PACKET-1 | large/mixed review silently truncates yet returns clean | treatment missed clean status with overflow and no declared gaps | split by threshold; missing/unreviewed/overflow/truncation prevents clean synthesis | green |
| PACKET-2 | packet coverage is incomplete or duplicated without proof | treatment missed one unreviewed changed file hidden by empty unreviewed list | every changed file assigned; responses declare reviewed, closure, invariants, unreviewed, findings | green |
| PACKET-3 | unsafe reference or argument escapes repository / leaks content | treatment missed parent/absolute/option/control-character path case | reject absolute, `..`, symlink, control, option-like and malformed identities; packets contain metadata/paths only | green |
| PACKET-4 | threshold boundary produces silent over-budget packet | treatment returned `combined` for 21/25/26 files, 601/800/801 lines and 2/3 domains; failed blocked partition case | test 20/21/25/26 files, 600/601/800/801 lines, one/two/three domains and per-packet caps | green |
| CLOSURE-1 | non-Markdown reference, cycle, missing/ambiguous target escapes | treatment missed absent active target; clean cycle control stayed clean | typed edge reasons across Markdown/JSON/YAML/frontmatter/script; cycle-safe traversal; missing active target fails | green |
| LINEAGE-1 | restart or legacy migration reduces count / rewrites heads | treatment missed replacement/resume count reduction; later transition RED caught post-migration legacy write and dropped head prefix | event-sequence fixtures prove monotonic rounds, append-only heads, one-way migration, cap boundaries | green |
| OWNER-1 | packet fan-out creates multiple lifecycle/verdict owners | synthesis test requires `parent_orchestrator`; packet templates declare input-only roles | one PM synthesis; packets are bounded review inputs only | green |
| SHEPHERD-1 | Shepherd duplicates Codex review or runs before clean | synthesis output retains Shepherd `pending`; route docs require downstream clean-only validation | workflow/test requires separate downstream trajectory validation | green |
| MEASURE-1 | stronger prose is presented as review improvement | baseline/treatment observe+score replay and committed machine report | separate detection/score steps, historical + opaque cases + controls, machine metrics and limitations | green |
| MEASURE-2 | same owner tunes against nominally held-out cases after GREEN | `inputs.json` and separate `oracle.json` frozen before GREEN; hashes recorded in `corpus-manifest.json` | preserve hashes and report fixture-level blinding limitation | red fixture frozen |
| SCOPE-1 | implementation collides with PR #493-owned paths | compiler config freezes all 18 forbidden PR #493 paths; final exact diff check pending committed head | exact-base changed-path disjointness rejects forbidden paths | green implementation / final check pending |
| SCOPE-2 | implementation expands beyond active PM system | compiler rejects changed paths outside positive allowlist | exact positive allowlist rejects new CLI/subcommand, article, dependency, product, connector, or #408 paths | green |

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
- 2026-07-23 RED command: `bash scripts/tests/pm-review-system.sh` exited 1. Exact semantic groups: `final_parent_readiness` classified `human_ready`; SEC1 noncanonical disposition accepted; treatment missed 15 defect cases (five PR #495/accepted and ten opaque) and six split/blocked threshold expectations. Ten paired clean controls produced no reported false positives. The failure was not caused by a missing marker or missing script.
- 2026-07-23 corpus freeze: `inputs.json` SHA-256 `21048975abdaac5be8a6f9335c3c5e32b28ffb1d6b718f38ef591271d0aed436`; separate `oracle.json` SHA-256 `5fc03b6ed1faddc51eec011894753e9481575db5ff8f6f55fd3ef93b3194b0e3`. Detector subprocess receives no oracle argument.
- 2026-07-23 transition RED: focused test exited 1 because one-way legacy migration and append-only head-history transitions were not enforced. GREEN added explicit rejection of post-migration legacy writes and dropped/reordered/duplicate history while the monotonic clean transition remained clean.
- 2026-07-23 GREEN: `bash scripts/tests/pm-review-system.sh`, `bash scripts/tests/pm-orchestrator-contract.sh`, and `bash scripts/tests/pi-model-routing.sh` pass. `python3 -m py_compile scripts/pm-review-system.py`, shell syntax, Shellcheck, JSON parsing, YAML syntax parsing, and `git diff --check` pass.
- 2026-07-23 measured deterministic fixture result: baseline 0/15 TP, 15 escapes, recall 0.0, precision undefined; treatment 15/15 TP, 0/10 clean-control FP, recall 1.0, precision 1.0, seven of seven threshold decisions. Wall time was 0.089584 ms baseline and 0.108041 ms treatment on this run. Token/cost/review-round/prospective data are explicitly unavailable. These are deterministic preflight fixtures, not local-Codex or production claims.
