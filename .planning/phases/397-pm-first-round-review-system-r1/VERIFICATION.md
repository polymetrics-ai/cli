# Issue #397 PM First-Round Review System Verification

**Status:** correction round 1/5 full verification passed at `528dcc685`; final evidence-head rerun and fresh exact-head recurrence review pending

> Earlier checked graph/lab/compiler items record pre-review focused-fixture evidence. Round-1 review disproved several production-completeness claims; they are not final acceptance. The correction checklist below supersedes them.

## Identity and scope

- [x] Isolated disposable worktree is not the primary clone (`SETUP-EVIDENCE.md`).
- [x] Status/log/untracked/stash/diff inventory captured clean before production edits (`SETUP-EVIDENCE.md`).
- [x] Normal fetch verified remote parent contains and equals PR #495 squash `0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20` (`SETUP-EVIDENCE.md`).
- [x] Clean detached remote-base transition and branch creation recorded without reset/rebase/amend/force (`SETUP-EVIDENCE.md`).
- [x] Branch is `chore/pm-first-round-review-system-r1`.
- [x] Parent PR #438 exists and remains draft/human-only.
- [x] Diff remains disjoint from all PR #493-owned paths.
- [x] No #408/TUI, Gong/#497, write-URL, dependency, credential, connector, CLI help, or reverse-ETL behavior change.
- [x] Historical PR #495 evidence remains truthful and is not rewritten as clean.
- [x] Diff is within the exact positive path allowlist in `PLAN.md`.
- [x] No new CLI/subcommand, orchestration owner, article/#498, `go.mod`, or `go.sum` change.

## RED/GREEN requirements

- [x] RED: canonical `final_parent_readiness` incorrectly classifies `human_ready`.
- [x] RED: valid identifier `SEC1` with noncanonical disposition is not rejected by current parser.
- [x] RED: prose-only dependency mutation escapes baseline.
- [x] RED: transitive prohibited-template mutation escapes direct-file baseline.
- [x] RED: replacement-head lineage reset/fragmentation escapes shape-only baseline.
- [x] RED: unsafe absolute/parent/control/option-like references are rejected at the semantic corpus layer; real symlink/identity integration RED remains pending.
- [x] RED: cross-format graph, cycle, and missing-target semantics are exercised in the frozen corpus; real file-parser integration RED remains pending.
- [x] RED: stale packet identity, incomplete coverage, overflow, threshold boundaries, replacement/resume, and cap boundary fail semantically; one-way legacy migration/append-only integration RED remains pending.
- [x] RED: opaque corpus and separate oracle are frozen and hashed before treatment implementation.
- [x] GREEN: all five concrete cases and pre-frozen mutation cases are detected for the intended semantic reason.
- [x] GREEN: unknown schema/kind, stale evidence, cap exceeded, arbitrary IDs, and missing active targets block.
- [x] GREEN: clean/metamorphic controls do not produce findings.

## Review compiler

- [x] Active required-reference closure records source, target, and edge reason.
- [x] Missing active targets and prohibited reachable targets fail.
- [x] Authority registry records authoritative state plus writers/readers/mirrors.
- [x] Dispatch/readiness checks parse relationships rather than trust prose.
- [x] Exact base/head are verified and embedded in each packet.
- [x] Small coherent changes stay one packet only at ≤20 files, ≤600 lines, one domain.
- [x] 21–25 files, 601–800 lines, or exactly two domains split conservatively; >25 files, >800 lines, or >2 domains split mandatorily.
- [x] Any partition that cannot meet ≤20 changed, ≤10 closure/authority, and declared 30K-token packet caps blocks rather than truncates.
- [x] Every changed file is assigned; each response declares reviewed, closure, invariants, unreviewed, findings, and overflow/truncation.
- [x] Missing response/coverage, stale identity, overflow, or silent truncation cannot synthesize clean.
- [x] Findings are unlimited and synthesize to one PM-owned local-Codex disposition.
- [x] Shepherd remains independent and runs only after clean synthesis.

## Algorithm research checkpoint

- [x] All three additive captain requirements were read in full before production graph/lab edits.
- [x] `ALGORITHM-RESEARCH.md` records primary title, author/organization, date/version, stable URL/DOI, mechanism, and limits.
- [x] Every required capability compares at least two credible alternatives.
- [x] Disposable exact-head benchmark covers real graph shape, cold/traversal/incremental timing, memory availability, oracle precision/recall, synthetic fan-in/fan-out/cycles, determinism, invalidation, and packet consequences.
- [x] Candidate baseline was explicitly revised: typed directed multigraph + forward/reverse adjacency + multi-source relation-policy BFS selected; v1 SCC condensation, unrestricted traversal, persistent cache, and new dependencies rejected with measured reasons.
- [x] Selected design adds no shipped dependency.

## Bidirectional practical impact graph

- [x] Separate correction inputs/oracle are frozen and hashed before GREEN; old root/outgoing behavior failed 17 defect fixtures and real integration assertions while four clean controls stayed clean.
- [x] Union of every changed file and canonical roots forms seeds; declared universe is indexed before traversal.
- [x] Every edge records stable id, source, target, relation/direction, parser, provenance reason/location, and `active|inactive|unknown` certainty.
- [x] Markdown/frontmatter/JSON/YAML, shell, Python, authoritative Go package/test/import/current-build, authority, generator/generated, lateral, mirror, and temporal edges are covered.
- [x] Reverse leaf, script both directions, authority, generated chain, Go importer/test, platform variant, cycle, unknown conditional, and unrelated control pass.
- [x] Iterative deterministic relation-policy BFS terminates on cycles; a continuing frontier or any index/node/edge/depth/file/token/Go-command bound blocks explicitly.
- [x] Missing active edges and unresolved unknowns include safely or block; inactive edges do not become active.
- [x] Complete impact files and edge ids enter exact-version coverage and bounded packets only after discovery.
- [x] Missing/stale impact response coverage, overflow, truncation, or graph-bound evidence cannot synthesize clean.
- [x] Output distinguishes practical file/package impact from unavailable symbol-level call/data-flow coverage.

## Counterfactual hypothesis lab

- [x] Old system's absent lab support and missing response/synthesis semantics are captured as RED before production lab code.
- [x] Every experiment uses a detached disposable exact-head copy under a private temporary root; canonical candidate stays clean/read-only.
- [x] Environment/config/credentials are scrubbed; captured artifacts contain no secret sentinel.
- [x] Canonical/outside/symlink writes and cross-lab access are denied.
- [x] Network, generic shell, commit/push/PR/remote mutation, install, credential/live connector, deployment, and destructive external commands are denied.
- [x] A proven OS sandbox is required; unavailable/ambiguous/policy-only fallback blocks clean experiments.
- [x] Time/process/disk/output bounds, descendant termination, evidence bounds, and whole-lab destruction are enforced.
- [x] Candidate exact base/head/tree/status are proven unchanged before/after; drift or cleanup failure blocks synthesis.
- [x] Evidence captures hypothesis/alternative, examined impact edges, temporary diff hash/summary, argv, expected discriminator, stdout/stderr, exit, duration, observation, and safety proof.
- [x] Competing hypotheses discriminate; an inconclusive performed experiment cannot prove clean; decisive static evidence uses an explicit no-experiment reason.
- [x] Concurrent packet labs cannot inspect peers and are destroyed; an unrelated clean control performs no unnecessary experiment.
- [x] Graph/lab/packet/response/synthesis contracts are versioned; incompatible v1 migration fixtures fail closed.
- [x] Focused PM tests remain transitively present in repository `make verify`/CI routing.

## Exact-head Codex round 1 correction

- [x] 17/17 fresh Sol/xhigh packet reviews completed at `b1d869732d230575ab7c8295b15cef42cc0078ef`, zero unreviewed files, 89 unlimited raw findings.
- [x] Raw manifest/synthesis/response hashes and 22 provider attempts/5 operational failures are recorded separately from correction round usage in `REVIEW-R1-MEASUREMENT.json`.
- [x] Duplicate authentication failure was diagnosed after two attempts; sequential auth probe passed and serialized recovery succeeded. Provider retries did not consume correction rounds.
- [x] Every raw finding has an explicit accepted-with-modification or duplicate disposition in `REVIEW-R1-DISPOSITION.md`; none is declined, waived, hidden, or deferred.
- [x] Stable lineage advanced to correction round 1/5; Shepherd was correctly not run on non-clean synthesis.
- [x] R1-A exact compile/synthesis head/tree/base/ready-manifest binding passes RED→GREEN.
- [x] R1-B strict response/invariant/experiment/lab-evidence contract passes RED→GREEN.
- [x] R1-C relation/direction policy-state BFS and global frontier handling pass RED→GREEN.
- [x] R1-D coherent exact-blob slices and bounded endpoint excerpts stay ≤30K tokens without truncation.
- [x] R1-E parser/certainty/endpoint/deletion/prohibited/root/shell/fixture cases fail closed correctly.
- [x] R1-F file/node/edge limits prevent broad reads/materialization before blocking.
- [x] R1-G external-module offline Go indexing and deleted-Go base+head context pass.
- [x] R1-H outside config/home/temp reads, Git administration, and descendant process creation are denied; lab evidence is exact-bound.
- [x] R1-I explicit-null/conflicting schema aliases block while key-absent legacy remains compatible.
- [x] R1-J reusable policy and validated per-run scope are separated and manifest-bound.
- [x] R1-K current PM route siblings are reconciled without touching explicit PR #493-owned paths.
- [x] R1-L post-#495 parent authority and phase-state temporal contract are truthful and validated.
- [x] R1-M durable phase evidence is reconciled while historical logs remain intact.
- [x] R1-N remote freshness is parent-owned and packet review remains local/no-network.
- [x] R1-O governing root/source docs enter explicit bounded closure.
- [x] Re-run focused/full verification on correction head `528dcc68599af4efaaee46c9b038bbd022178411`.
- [ ] Re-run exact-head packets, recurrence comparison, synthesis, and Shepherd on the final evidence head.

## Measurement

- [x] Historical source identities are retained for PR #495 replays.
- [x] Detector execution does not receive the separate oracle.
- [x] Opaque held-out mutations and clean/metamorphic controls run.
- [x] Machine report captures recall, precision, escapes, false positives, exact invalidations, rounds, overflows, wall time, and available token/cost fields.
- [x] Deterministic fixture results are not described as model-review or prospective production results.
- [x] Corpus provenance/hash and fixture-level blinding limitation are explicit.
- [x] Unavailable token/cost/prospective evidence is explicit.
- [x] Packet artifacts contain paths/metadata only; environment-sentinel regression proves no environment-value copy.
- [x] Correction measurement separates static/semantic findings, impact-graph coverage, lab experiments/safety/cleanup, deterministic mutation results, actual packet metrics, and prospective evidence.
- [x] Baseline/treatment reports impact recall/precision, edge/file coverage, bound hits, cold/traversal time, memory availability, determinism, invalidation, impact size, and packet consequences.
- [x] Lab measurement reports attempted/denied/completed/inconclusive experiments, cleanup/identity proofs, latency, output/disk/process bounds, and available model token/cost fields.
- [x] Public fixtures are labeled regression evidence rather than a secret benchmark; prospective #493/#408/later Architecture v2 evidence remains unavailable until observed.

## Focused commands

```bash
bash scripts/tests/pm-review-system.sh
bash scripts/tests/pm-orchestrator-contract.sh
bash scripts/tests/pi-model-routing.sh
bash -n scripts/pm-terminal-classifier.sh scripts/tests/pm-review-system.sh scripts/tests/pm-orchestrator-contract.sh
# focused test additionally exercises classifier usage/malformed JSON/legacy stdout+stderr+exit compatibility,
# JSON envelope fields, non-TTY execution, unsafe paths and symlinks, closure formats/cycles,
# threshold boundaries, state transitions, and stale evidence
ruby -e 'require "psych"; Psych.parse_file(ARGV.fetch(0))' .agents/agentic-delivery/schemas/orchestration-state.schema.yaml
ruby -e 'require "psych"; Psych.parse_file(ARGV.fetch(0))' .planning/traces/cli-architecture-v2-orchestration-state.yaml
python3 -m py_compile scripts/pm-review-system.py
python3 -m json.tool .agents/agentic-delivery/contracts/pm-review-system.json >/dev/null
python3 -m json.tool .planning/phases/397-pm-first-round-review-system-r1/MEASUREMENT.json >/dev/null
```

## Full local gates

```bash
gofmt -w cmd internal
git diff --exit-code -- cmd internal
git diff --check
go vet ./...
go test -timeout 20m ./...
go build ./cmd/pm
go mod verify
go mod tidy -diff
make verify
```

The pre-correction passes at `7f1b2d8fe`/`5601be8d0` are historical and do not satisfy delivery.
After both correction GREENs on one committed exact head:

- [x] Focused graph/lab/PM gates pass at exact implementation head `0e210d1819f2642ee600aa6921873997101ba7bd`.
- [x] Shellcheck is available and passes for all changed shell scripts.
- [x] JSON/YAML parsing passes.
- [x] Go formatting produces no product diff.
- [x] `go vet ./...` passes.
- [x] `go test -timeout 20m ./...` passes (`internal/cli` observed 472.174s).
- [x] `go build ./cmd/pm` passes.
- [x] `go mod verify` passes.
- [x] `go mod tidy -diff` passes.
- [x] `make verify` passes, including safe sample reverse ETL in required plan → preview → approval → execute order.

## Review and delivery

- [x] Captain's 2026-07-24 conditional Firstmate merge authorization is recorded; this agent remains no-merge, parent PR #438 remains draft/human-only, and the deliverable is a green open stacked PR.
- [x] Exact corrected verified commit `e4ca19ce864b6a3362a2d490aec2d0b6a3717b1f` exists and coherent research/plan/RED/GREEN/refactor checkpoints were pushed additively. (`7f1b2d8fe` remains historical evidence only.)
- [ ] All tracked evidence was committed before final exact-head packet/Shepherd gates; no tracked write followed them.
- [ ] Fresh local-Codex packet system reviews exact base/head and synthesizes one result before Shepherd; raw responses live outside the tracked worktree and are hashed/summarized in delivery evidence.
- [x] Every round-1 finding has a canonical disposition; future-round findings remain pending by definition.
- [x] Stable five-round `rounds_by_range` usage is 1/5 and append-only head history is retained without lineage reset.
- [ ] Independent Shepherd exact-head trajectory validation recorded after clean Codex review.
- [ ] `no-mistakes axi` returns `checks-passed`; `passed` (merged/closed) is treated as a violation/escalation, not success for this task.
- [ ] Any AXI-created commit/base/head change invalidated prior evidence and triggered applicable full verification, fresh packet synthesis, and fresh Shepherd at final identities.
- [ ] No parallel/manual reviewer ran outside the specified PM packet system.
- [ ] Branch pushed normally without force.
- [ ] PR has Conventional Commit title, targets `feat/cli-architecture-v2`, uses `Refs #397`, and reports full URL, exact source/head, risk, metrics, limitations, and round usage.
- [ ] Published branch history remained additive; any proposed post-publication non-additive pipeline rewrite stopped for human direction.
- [ ] CI green.
- [ ] PR remains open and unmerged for captain approval.
