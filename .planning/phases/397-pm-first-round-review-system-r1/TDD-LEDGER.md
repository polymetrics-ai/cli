# Issue #397 PM First-Round Review System TDD Ledger

**Status:** captain algorithm research complete; impact-graph and hypothesis-lab RED pending; prior verification historical

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
| RESEARCH-1 | graph design selected by intuition/popularity | disposable exact-head benchmark compared outgoing, reverse scan, indexed BFS, SCC, real/synthetic/adversarial cases after primary-source research | `ALGORITHM-RESEARCH.md` selects repository-fit design with measured correctness/safety/runtime/dependency tradeoffs | green research checkpoint |
| IMPACT-1 | changed leaf misses upstream active referencer | pending pre-GREEN integration RED | index before traversal; changed+canonical multi-source reverse closure includes referencer | planned |
| IMPACT-2 | changed script/package omits upstream importer or downstream consumer | pending pre-GREEN integration RED | typed forward+reverse script/Python/Go edges include both | planned |
| IMPACT-3 | authority schema omits writer/reader/mirror | pending pre-GREEN integration RED | configured authority relations enter practical impact closure and packet coverage | planned |
| IMPACT-4 | generated/source/consumer or lateral variant omitted | pending pre-GREEN integration RED | configured generator/generated, fixtures/tests, platform siblings, mirrors are typed edges | planned |
| IMPACT-5 | Go source omits same-package tests/importing package/build variant | pending pre-GREEN integration RED | authoritative `go list -json -deps -test` plus unknown ignored/platform files | planned |
| IMPACT-6 | cycle or graph bound silently truncates clean | pending pre-GREEN integration RED | iterative deterministic visited BFS; any continuing frontier/index/node/edge/file/token bound blocks | planned |
| IMPACT-7 | impact graph degenerates to whole repository | pending unrelated-control RED | declared universe + typed relation policy excludes unrelated control with provenance | planned |
| IMPACT-8 | packet/synthesis ignores impact file/edge gaps | pending pre-GREEN packet RED | complete discovery before packetization; exact impact file/edge ids echoed or synthesis blocks | planned |
| IMPACT-9 | active/inactive/unknown and exact build configuration collapse | pending conditional/build-tag RED | three-valued certainty; current Go build provenance; unresolved unknown includes or blocks | planned |
| LAB-1 | reviewer experiments mutate canonical candidate | pending pre-GREEN lab RED | exact-head disposable copy only; before/after head/tree/status proof | planned |
| LAB-2 | lab escapes path/symlink/network/process effect boundary | pending adversarial RED | proven OS sandbox plus static command/path policy; ambiguity or unavailable backend blocks | planned |
| LAB-3 | lab exposes secrets or permits Git/install/live/deploy action | pending adversarial RED | scrubbed allowlisted environment and explicit denials; captured evidence secret scan | planned |
| LAB-4 | time/process/disk/output overflow or cleanup failure can synthesize clean | pending adversarial RED | monitored hard bounds, descendant kill, whole-root destruction; failure blocks | planned |
| LAB-5 | hypothesis is confidence prose, inconclusive, or lacks alternative | pending response-contract RED | claim/alternative/discriminator/observation/support/disconfirmation and no-experiment reason validated | planned |
| LAB-6 | packet labs collide or inspect one another | pending concurrent-isolation RED | private unpredictable per-experiment roots; cross-lab access denied; roots destroyed | planned |
| CONTRACT-2 | incompatible graph/lab/packet schema is silently accepted | pending v1 migration RED | explicit v2 graph/packet/response/synthesis and v1 lab migration fixtures fail closed | planned |
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
- 2026-07-23 exact implementation head `7f1b2d8fe`: focused review-system/PM-contract/model-routing tests, Shellcheck, Python compile, JSON/YAML syntax, `gofmt` no-diff, `go vet ./...`, `go test -timeout 20m ./...`, `go build ./cmd/pm`, `go mod verify`, `go mod tidy -diff`, and `make verify` all passed. The required smoke used only sample local data and preserved reverse ETL plan → preview → approval → execute order. This verification became historical when captain added graph/lab requirements.
- 2026-07-24 research checkpoint: read the impact-graph, counterfactual-lab, and algorithm-research requirements in full. `scripts/gsd prompt plan-phase ... --research` generated the official route. Read-only scout/security calls failed from WebSocket/provider authentication, so inline GSD fallback was recorded; no production graph/lab edit had started.
- 2026-07-24 primary-source/benchmark GREEN: `ALGORITHM-RESEARCH.md` compares Bazel/Skyframe, Buck2/DICE, Pants, Nx, Turborepo, Gradle, CodeQL/CPG, Kythe, SCIP, Go tooling, Ekstazi, STARTS, Microsoft TIA, BFS/DFS, and Tarjan. At exact head `5601be8d01f9f0044a5e7e875be5669f7d6fc280`, the disposable prototype indexed 1,669 nodes/5,039 edges in 2,004.212 ms with 51,118,867 Python-allocated peak bytes; indexed bidirectional BFS found 21/21 oracle nodes with 0 FP, outgoing-only found 14/21, and a 20K-node/30,197-edge synthetic traversal took 63.735 ms indexed versus 10,479.531 ms repeated-scan. The unrestricted depth-3 graph still hit a frontier (201 PM/213 Go nodes), so selection requires typed relation policy and fail-closed bounds rather than whole-graph inclusion.
