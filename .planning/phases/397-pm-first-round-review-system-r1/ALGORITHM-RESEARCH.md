# PM Impact-Graph Algorithm Research and Selection

**Phase:** `397-pm-first-round-review-system-r1`

**Research checkpoint:** captain-required before production graph implementation

**Exact benchmark candidate:** `5601be8d01f9f0044a5e7e875be5669f7d6fc280`

**Stable lineage:** `0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20...pm-first-round-review-system-r1`
**Date:** 2026-07-24

## Decision scope and definition of “best”

There is no universal best impact algorithm. “Best for this repository” means the smallest
maintainable, dependency-free design that, in this order:

1. does not silently miss an oracle-required practical file/package relationship;
2. fails closed on unresolved active/unknown relationships and explicit bounds;
3. is deterministic and explains every included edge;
4. is bound to one exact candidate head/tree;
5. avoids indiscriminately selecting the repository;
6. is fast enough for repeated local PM review;
7. does not introduce a shipped dependency or a second PM owner.

This first increment is **practical deterministic file/package impact coverage**, not symbol-level
call/data-flow coverage. CodeQL, Kythe, and SCIP show what a semantic index can provide, but adopting
one would change dependency, indexing, build-capture, and maintenance assumptions beyond this PR.

## Research method

- Used `chrome-devtools-axi --help`, then browser sessions against official documentation, official
  repository source, DOI/Crossref metadata, and OpenAlex publication records.
- Used `gh-axi` only for read-only official repository/source discovery and exact source commits.
- Used local authoritative Go 1.26.4 `go help list` and `go help buildconstraint` in addition to
  `pkg.go.dev`.
- Built a disposable `/tmp/pm-impact-algorithm-benchmark.py` prototype. It did not edit the project,
  add a dependency, call credentials, or mutate a remote. The prototype indexed the exact candidate,
  used `go list -json -deps -test`, compared algorithms on an oracle fixture and a 20K-node
  synthetic graph, and measured with `time.perf_counter_ns` plus Python `tracemalloc`.
- Browser/source access date for continuously updated documentation: 2026-07-24.

## Primary-source inventory

| ID | Primary source; author/organization; date/version | Stable URL | Mechanism used here | Limits for this PR |
|---|---|---|---|---|
| S1 | **The Bazel Query Reference**, Bazel/Google, Nightly and 9.2 docs, accessed 2026-07-24 | https://bazel.build/query/language | `rdeps` and Sky Query `allrdeps` require a defined universe; query results can conservatively over-approximate configurations. | Build-target graph, not arbitrary Markdown/script/state semantics; a bad universe can return only the seed.
| S2 | **Skyframe**, Bazel/Google, Nightly and 9.2 docs, accessed 2026-07-24 | https://bazel.build/reference/skyframe | Immutable keys/values, dependency recording, bottom-up invalidation of transitive reverse dependencies, top-down validation. | A full incremental evaluator is excessive for a one-shot review manifest.
| S3 | **Configurable Query (cquery)**, Bazel/Google, Nightly and 9.2 docs, accessed 2026-07-24 | https://bazel.build/query/cquery | Runs on configured targets and resolves `select()`/build options unlike loading-phase query. | Does not cover actions/artifacts or this repository's non-Bazel artifacts.
| S4 | **Dynamic Incremental Computation Engine (DICE)** and **Introduction to Modern DICE**, Meta/Buck2, source commit `6c690f75e48a423f01c7bdbee434e6d40afd0bda`, 2026-07-23 | https://github.com/facebook/buck2/blob/6c690f75e48a423f01c7bdbee434e6d40afd0bda/dice/README.md and https://github.com/facebook/buck2/blob/6c690f75e48a423f01c7bdbee434e6d40afd0bda/docs/insights_and_knowledge/modern_dice.md | Tracks dependencies/reverse dependencies, invalidates leaves, recomputes demanded invalid nodes, and uses value/version equality for early cut-off. | Larger state/lifecycle burden; Buck's own source notes cold-work tradeoffs.
| S5 | **Buck Query Language** and DICE incrementality docs, Meta/Buck2, same exact source commit | https://github.com/facebook/buck2/blob/6c690f75e48a423f01c7bdbee434e6d40afd0bda/docs/concepts/buck_query_language.md and https://github.com/facebook/buck2/blob/6c690f75e48a423f01c7bdbee434e6d40afd0bda/dice/dice/docs/incrementality.md | Distinguishes unconfigured/configured queries and reverse-dependency behavior over a graph universe. | Build-rule graph and Buck configuration semantics do not directly model PM files.
| S6 | **Advanced target selection**, Pants, version 2.32, accessed 2026-07-24 | https://www.pantsbuild.org/stable/docs/using-pants/advanced-target-selection | Git changed files plus direct/transitive dependents; exposes a known third-party transitive-dependency limitation. | Target metadata must already exist; not a substitute for active-reference parsing.
| S7 | **Run Only Tasks Affected by a PR** and **Explore your Workspace**, Nx, v23, accessed 2026-07-24 | https://nx.dev/ci/features/affected and https://nx.dev/features/explore-graph | Combines Git base/head changes with a project graph; project-edge provenance can identify creating files; task graph is distinct. | Project granularity is coarser than required PM file/state edges.
| S8 | **Caching**, Turborepo/Vercel, current docs, accessed 2026-07-24 | https://turborepo.com/docs/crafting-your-repository/caching | Deterministic task-input fingerprints and cached outputs. | Fingerprinting accelerates known tasks; it does not discover omitted impact edges.
| S9 | **Incremental build**, Gradle, current docs, accessed 2026-07-24 | https://docs.gradle.org/current/userguide/incremental_build.html | Declared task inputs/outputs and up-to-date checks avoid repeated work; undeclared inputs undermine correctness. | Task-level cache model; not a cross-format impact oracle.
| S10 | **About CodeQL queries** and **About data flow analysis**, GitHub CodeQL, current docs, accessed 2026-07-24 | https://codeql.github.com/docs/writing-codeql-queries/about-codeql-queries/ and https://codeql.github.com/docs/writing-codeql-queries/about-data-flow-analysis/ | Relational code databases; local/global data-flow graphs and path queries model semantic value flow. | Much stronger semantic scope and index cost than practical file/package context; does not natively model PM prose/state authority.
| S11 | **Modeling and Discovering Vulnerabilities with Code Property Graphs**, Yamaguchi, Golde, Arp, Rieck, 2014 | https://doi.org/10.1109/SP.2014.44 | Combines AST, CFG, and program-dependence properties in a queryable graph. | Symbol/control/data semantics require language front ends and are explicitly out of scope.
| S12 | **Kythe Schema Reference** and **An Overview of Kythe**, Kythe/Google, source release v0.0.76 (`26056edfc953b5d4ea0ed8e94db072caa7f7d4c7`, 2026-07-16) | https://kythe.io/docs/schema/ and https://kythe.io/docs/kythe-overview.html | Language-agnostic typed edges (`defines`, `ref`, `ref/call`, `reads`, `writes`, `generates`) and cross-reference indexing with provenance. | Requires compilation extraction/indexers; schema intentionally tolerates incomplete data, while this review gate must fail closed.
| S13 | **SCIP Code Intelligence Protocol**, scip-code/Sourcegraph, source commit `44d39fcfc95486d066a796e2cec8c7ec5d429aae`, 2026-07-21 | https://github.com/scip-code/scip/tree/44d39fcfc95486d066a796e2cec8c7ec5d429aae | Language-agnostic occurrences/symbol relationships for definitions, references, and implementations. | Protocol plus language indexers/protobuf dependencies; cross-format PM relations still need custom edges.
| S14 | **go command: List packages or modules** and `go help list`, Go Authors, local Go 1.26.4 | https://pkg.go.dev/cmd/go#hdr-List_packages_or_modules | `go list -json -deps -test` reports package/test variants and import closure; fields include active and `IgnoredGoFiles`. | Current environment/configuration only unless other variants are queried; test variants can duplicate packages.
| S15 | **Build constraints** and `go help buildconstraint`, Go Authors, local Go 1.26.4 | https://pkg.go.dev/cmd/go#hdr-Build_constraints | Authoritative `//go:build`, GOOS/GOARCH, feature tags, and filename suffix rules. | Unknown target matrices cannot be called inactive; ignored/platform files need conservative unknown/lateral treatment.
| S16 | **go/packages**, Go tools team, current module docs, accessed 2026-07-24 | https://pkg.go.dev/golang.org/x/tools/go/packages | Rich configurable package loading (`NeedImports`, `NeedDeps`, syntax/types). | External module dependency; not approved or needed for file/package edges.
| S17 | **Ekstazi: Lightweight Test Selection**, Gligoric, Eloussi, Marinov, 2015 | https://doi.org/10.1109/ICSE.2015.230 | Dynamic dependencies from tests to files; evaluated over 615 revisions/32 projects and reported 32% average end-to-end reduction. | Runtime history/instrumentation can miss never-observed paths and is unavailable before first execution.
| S18 | **STARTS: STAtic regression test selection**, Legunsen, Shi, Marinov, 2017 | https://doi.org/10.1109/ASE.2017.8115710 | Static type-dependency graph; selects tests that can reach a changed type in transitive closure. | Java/Maven/type granularity; reported selection/runtime tradeoff is not transferable to this Go/PM graph.
| S19 | **Speed up testing by using Test Impact Analysis**, Microsoft, published 2018-12-07, updated 2026-05-07 | https://learn.microsoft.com/en-us/azure/devops/pipelines/test/test-impact-analysis?view=azure-devops | Generated/custom test dependency maps, includes changed/failing/new tests, and falls back to all tests when unsupported. | Managed-code/single-machine scope; safe fallback principle is more relevant than implementation.
| S20 | **Depth-First Search and Linear Graph Algorithms**, Robert Tarjan, June 1972 | https://doi.org/10.1137/0201010 | Linear-time SCC algorithm, `O(V+E)`, for directed cycles/condensation. | SCC materialization is not necessary merely to make reachability cycle-safe.
| S21 | **Introduction to Algorithms**, Cormen, Leiserson, Rivest, Stein, 4th ed., 2022 | https://mitpress.mit.edu/9780262046305/introduction-to-algorithms/ | Multi-source BFS/DFS with visited sets gives deterministic `O(V+E)` reachability when adjacency is materialized. | Textbook traversal does not decide repository-specific edge semantics.

## Findings from production systems and research

### Universe and configuration are correctness inputs

Bazel's `rdeps`/`allrdeps` documentation is the clearest warning: reverse traversal is only as good
as its universe. `--infer_universe_scope` can accidentally create a universe containing only the
seed. Bazel also distinguishes loading-phase conservative `query` from configuration-aware
`cquery`. Therefore this system must index the declared review-relevant universe first, emit the
universe and exclusions, and represent condition certainty instead of claiming one current build is
all variants.

### Reverse adjacency is normal; blindly undirecting every edge is not

Skyframe, DICE, Pants changed dependents, Nx affected, STARTS, and Microsoft TIA all depend on
recorded reverse relationships. Kythe/SCIP demonstrate typed provenance. None support treating every
textual mention as an equally traversable edge. The edge parser and relation policy are part of the
soundness boundary.

### Static and dynamic evidence are complementary

Ekstazi's dynamic test/file dependencies are precise for observed executions; STARTS' static
transitive type graph operates without runtime history but is more conservative. This first
increment has no trusted historical runtime coverage, so static/declared practical edges are the
base. Later prospective data can augment, not replace, static edges.

### Unknown conditions must not be relabeled inactive

Bazel's configuration distinction and Go build constraints both reject binary thinking. For the
exact current GOOS/GOARCH, `go list` is authoritative. `IgnoredGoFiles`, platform siblings, unresolved
shell/Python conditions, or unqueried configurations remain `unknown`; policy includes them when
bounded and otherwise blocks with provenance. Explicit historical/deprecated examples may be
`inactive` but remain indexed as such.

## Capability decision matrix

Scores are repository-relative: **H** strong, **M** usable with limits, **L** poor. “Selected” means
selected for this first dependency-free increment.

| Capability | Alternative | Recall / precision | Determinism / provenance | Runtime / memory | Dependency / maintenance | Decision |
|---|---|---|---|---|---|---|
| Reverse references | Re-scan every edge for each frontier node | H / H with same parser | H | `O(VE)` worst case; synthetic 10,479.531 ms | Low code, poor scale | Reject.
| Reverse references | Materialized forward+reverse adjacency | H / H with same parser | H; edge retained once with parser/reason | `O(V+E)` build/traversal; synthetic 63.735 ms | Standard library only | **Select**.
| Reverse references | Kythe/SCIP external semantic index | Potentially H symbols / H | H when complete | Separate index/build capture | New protocol/indexers/deps | Defer.
| Typed relations | Undirected untyped path graph | High recall / low precision | Weak explanation | Can collapse large components | Easy but unsafe over-selection | Reject.
| Typed relations | Directed multigraph with relation/certainty/provenance and traversal policy | H practical / controllable precision | H | Linear adjacency | Moderate parser policy | **Select**.
| Typed relations | Full CPG/data-flow graph | H semantic / H | H | High language/index cost | Major new architecture | Defer.
| Certainty | Binary active/inactive regex | L under conditions | M | Cheap | Misclassifies unknown | Reject.
| Certainty | `active/inactive/unknown`, parser/build-tool specific | Conservative H | H, reason emitted | Small edge metadata cost | Moderate | **Select**.
| Go discovery | Regex/AST imports only | M; misses build/test variants | H syntax provenance | Fast | Standard library | Keep only as fallback/check.
| Go discovery | `go list -json -deps -test` plus ignored/build-tag metadata | H package/test current variant | H authoritative command output | 1,653.950 ms in full prototype | No new dependency | **Select**.
| Go discovery | `go/packages` syntax/types | H+ semantic | H | Additional loading cost | New unapproved module | Reject for this PR.
| Traversal | Root-only outgoing DFS | Low upstream recall | H | Fast | Existing gap | Reject.
| Traversal | Deterministic multi-source BFS over typed forward/reverse policy | H fixture recall | H ordered shortest provenance | `O(V+E)` | Simple | **Select**.
| Traversal | Unrestricted bidirectional closure over all parsed relations | High recall / very low precision | H | 201 nodes at depth 3 and still hit bound for one PM file | Packet explosion | Reject.
| Cycles | Recursive DFS + visited only | Correct but stack-sensitive | H | Linear | Simple | Reject recursion risk.
| Cycles | Iterative BFS + visited, explicit frontier/bound evidence | Correct reachability | H | Linear | Simple | **Select**.
| Cycles | Tarjan SCC condensation | Correct; useful repeated component queries | H | 9.056 ms prototype; extra component state | More code | Defer; not justified yet.
| Incrementality | Rebuild exact-head index each compile | H; no stale cache | H | 2,004.212 ms cold prototype | Lowest invalidation risk | **Select initially**.
| Incrementality | Exact-head/content-keyed per-file cache | H only with complete keying | H if head/config/parser/toolchain keyed | PM parse 4.068 ms; Go package 568.192 ms | Cache lifecycle/security complexity | Defer until repeated-run evidence.
| Incrementality | Skyframe/DICE-style dynamic evaluator | H | H | Excellent warm potential | Disproportionate architecture | Reject for this PR.
| Packet selection | Stop graph when packet/token limit reached | Low, silently truncated | Low | Fast | Unsafe | Reject.
| Packet selection | Complete policy traversal, then stable partition; any graph/packet bound hit blocks | H within declared universe/policy | H manifest equality | More packets | Fail-closed and simple | **Select**.
| Test impact | Dynamic observed test/file map | Precise observed / incomplete unobserved | H | Runtime collection | No history now | Defer prospective augmentation.
| Test impact | Static package/test/reverse-import graph | Conservative | H | Included in Go index | No dependency | **Select**.

## Disposable benchmark design

### Real repository prototype

The prototype indexed:

- tracked `.md`, `.json`, `.yaml`, `.yml`, `.sh`, and `.py` files under `.agents`, `.pi`,
  `scripts`, and `.planning`;
- explicit PM path references with simple active/inactive/unknown line classification;
- configured authority writer/reader/mirror relations;
- local Go package, import, test, ignored-file, and package-membership relations from
  `go list -json -deps -test ./...`.

It compared:

1. outgoing-only traversal;
2. bidirectional traversal that scans the full edge list per frontier node;
3. materialized forward/reverse adjacency with deterministic BFS;
4. directed Tarjan SCC construction as a separately timed cycle strategy.

This prototype deliberately over-approximated generic textual references. Its depth-3 bound hits
are evidence **against** unrestricted all-edge traversal, not proposed production packet counts.
Python `tracemalloc` excludes peak RSS of the Go subprocess.

### Oracle and adversarial fixture

Eight cases represented reverse leaf reference, script upstream/downstream, schema authority,
generator/generated consumer, Go importer/test, platform unknown variant, cycle, and unrelated
control. The expected union contained 21 nodes. The prototype also generated a deterministic graph
with 20,000 nominal chain nodes, cycles every 100 nodes, 4,999 high-fan-in edges, and 4,999
high-fan-out edges (30,197 edges total).

## Measured benchmark results

One run on Darwin/arm64, Go 1.26.4; timings are local evidence, not universal performance claims.

### Index and graph shape

| Metric | Result |
|---|---:|
| Review-relevant PM files indexed | 590 |
| Total practical nodes | 1,669 |
| Total edges | 5,039 |
| Active / inactive / unknown edges | 5,001 / 6 / 32 |
| Cold index wall time | 2,004.212 ms |
| Full `go list` portion | 1,653.950 ms |
| Python peak allocated memory | 51,118,867 bytes |
| Go subprocess peak RSS | unavailable |
| Directed SCC construction | 9.056 ms; 615 components; largest 117 nodes |

| Relationship | Edges |
|---|---:|
| `artifact_references` | 899 |
| `script_invokes` | 78 |
| `authority_writes` / `authority_reads` / `authority_mirror` | 6 / 5 / 8 |
| `go_imports` | 1,261 |
| `go_member_of` | 1,276 |
| `go_contains` | 899 |
| `go_test` | 601 |
| `go_build_variant` | 6 |

### Correctness against fixture oracle

| Algorithm | TP / expected | FN | FP | Recall | Precision |
|---|---:|---:|---:|---:|---:|
| Outgoing-only root/diff traversal | 14 / 21 | 7 | 0 | 0.6667 | 1.0000 |
| Repeated reverse edge scan | 21 / 21 | 0 | 0 | 1.0000 | 1.0000 |
| Materialized bidirectional adjacency | 21 / 21 | 0 | 0 | 1.0000 | 1.0000 |

The unrelated seed remained one node for all algorithms. Public fixtures are regression evidence,
not a secret benchmark.

### Traversal, incremental cost, and packet consequence

| Seed / method, depth 3 prototype | Median traversal | Impact nodes | Packets at 10 | Bound hit |
|---|---:|---:|---:|---|
| PM file / outgoing only | 2.565 ms | 57 | 6 | yes |
| PM file / repeated reverse scan | 35.298 ms | 201 | 21 | yes |
| PM file / adjacency BFS | 2.915 ms | 201 | 21 | yes |
| Go file / outgoing only | 2.280 ms | 39 | 4 | yes |
| Go file / repeated reverse scan | 38.747 ms | 213 | 22 | yes |
| Go file / adjacency BFS | 3.079 ms | 213 | 22 | yes |
| Reparse one PM file, 5 runs | 4.068 ms median (4.039 min) | — | — | — |
| Re-index one Go package, 2 runs | 568.192 ms median (563.064 min) | — | — | — |

All depth-3 traversals still had an unvisited frontier. Therefore production must not call these
sets complete. It must use a narrower, documented typed relation policy and block if its independent
node/edge/depth bounds are actually reached. Merely increasing depth until the PM graph is mostly
selected would fail the precision and packet-cost criterion.

### Large synthetic graph and identity

| Method | Result |
|---|---:|
| Materialized adjacency BFS, median 3 runs | 63.735 ms |
| Repeated reverse scan, 1 run | 10,479.531 ms |
| Relative traversal time | reverse scan about 164.4× slower |
| Repeated deterministic output | identical; SHA-256 `539a4fb6bbede9842e386f04ee6e3c820c1c86e2bc8cab2629b24d8c263f3c68` |
| Exact-head/config cache-key test | same identity stable; changed head changed key |

The synthetic graph's start node reached its connected fan component before depth 4, so no bound
was hit in that particular traversal. Permanent tests must separately force a continuing frontier
at every configured bound.

## Selected architecture

The required candidate architecture is **confirmed with three revisions**: no SCC condensation in
v1, no persistent cache in v1, and no unrestricted traversal of every textual edge.

1. **Typed directed multigraph.** Nodes are repository files and internal Go package nodes. Each
   edge has stable id, source, target, relation, parser, provenance reason/location, certainty
   (`active`, `inactive`, `unknown`), and configuration when applicable.
2. **Review universe first.** Enumerate safe tracked files in configured PM prefixes, every changed
   file, and Go files/packages surfaced by authoritative `go list`. Record excluded scopes and hit an
   index bound before reading too broadly. Seed canonical roots plus every changed file.
3. **Materialized forward and reverse adjacency.** Build the complete declared-universe index before
   traversal. Never discover reverse edges only from root-reachable outgoing files.
4. **Parser-specific edges.** Parse Markdown/frontmatter, JSON/YAML paths, shell source/exec,
   standard-library Python AST imports/run paths, `go list` package/test/current-build data, Go
   build-tag/platform unknown variants, configured authorities, generators/generated artifacts,
   mirrors, fixtures, siblings, and temporal transitions.
5. **Relation-policy multi-source BFS.** Iterative stable-order BFS carries provenance and relation
   budgets. Required references/script invocations may be transitive; authority, lateral, platform,
   and generated relations are explicit; generic descriptive links do not become unbounded control
   dependencies. Every policy stop is a declared semantic rule, not packet truncation.
6. **Three-valued certainty.** Do not traverse inactive edges. Include bounded unknown neighbors and
   expose them. An unresolved/missing unknown that cannot be conservatively represented blocks;
   unknown does not become silently active or inactive.
7. **Cycle safety by visited state.** Iterative BFS with `(node, policy-state)` visited keys gives
   deterministic cycle-safe reachability. Tarjan SCC is rejected for v1 because no repeated
   component query justifies extra component/provenance complexity. Add SCC only if later metrics
   show a need.
8. **Fail-closed bounds.** Configurable maximum indexed files/bytes, graph nodes/edges, traversal
   states/depth, impact files/edges, Go command duration, and packet files/edges/tokens. Any genuine
   frontier at a traversal bound, unresolved active edge, missing target, overflow, or truncation is
   a blocker with exact evidence.
9. **Discover before packetize.** Packet limits never stop discovery. Stable partition occurs only
   after a complete policy closure. The coverage manifest and response synthesis require exact
   impact-file and edge-id equality.
10. **No persistent cache in v1.** A roughly 2-second cold prototype is acceptable and exact-head
    rebuild eliminates stale-cache risk. Preserve an exact-head/config/parser/toolchain cache-key
    design, but do not persist until repeated production measurements justify it.
11. **Authoritative Go integration.** Use installed standard `go list -json -deps -test` with a
    scrubbed, non-networking environment and timeout. No `go/packages` dependency. Treat unqueried
    variants/ignored files conservatively and report the current GOOS/GOARCH context.
12. **One PM owner.** Impact graph selects context only. Existing PM packet synthesis, correction
    lineage, independent Shepherd, and human merge authority stay unchanged.

## Correctness and safety invariants

- Every changed file and configured canonical root is a seed or compilation blocks.
- Reverse adjacency is built from the indexed universe before traversal.
- Every impact node has at least one seed/provenance path; the unrelated control stays excluded.
- Every traversed edge has source, target, relation, direction, parser, reason/location, certainty,
  and stable id.
- Active missing targets block. Unknown targets are included with uncertainty or block when safe
  representation is impossible. Inactive edges never silently become active.
- Current Go package/test edges come from exact-tool output; ignored/platform files remain unknown.
- Cycles terminate deterministically without dropping nodes.
- Any index/graph/traversal/packet bound with a continuing frontier blocks; no partial set is clean.
- Complete impact files and edge ids are assigned to packets and echoed by responses before clean
  synthesis.
- Base/head/tree, graph config, parser version, and Go environment are recorded.
- Packetization occurs after discovery and cannot reduce coverage.
- Graph output contains metadata/paths only, never file contents or environment secret values.

## Rejected alternatives

- **Outgoing-only roots/diff:** escaped 7 of 21 required fixture nodes and is the captain-identified
  defect.
- **Repeated reverse scans:** same fixture correctness as adjacency but about 164× slower on the
  synthetic graph and `O(VE)` in the traversal shape.
- **Unrestricted bidirectional graph:** a single PM seed reached 201 nodes by depth 3 and still hit a
  frontier; this is not efficient review context.
- **SCC condensation now:** Tarjan is sound and cheap at this scale, but visited BFS already handles
  cycles and retains simpler per-edge shortest provenance. Revisit for repeated component queries.
- **Persistent DICE/Skyframe-style cache now:** cold cost is acceptable; invalidation/cache state
  adds more failure modes than measured benefit for one exact-head compilation.
- **`go/packages`, Kythe, SCIP, CodeQL/CPG:** valuable future semantic options, but add dependencies,
  build capture, indexers, or symbol semantics beyond this bounded PR.
- **Dynamic-only test impact:** no complete prior runtime map; it cannot cover unexecuted paths.
- **Packet-budget traversal:** fast but silently loses impact and violates the required contract.

## Known gaps and non-claims

- No symbol-level caller/callee, control-flow, alias, or data-flow precision.
- Python dynamic import, shell-computed paths, reflection, generated-at-runtime files, and arbitrary
  templating may remain unknown and block when relevant.
- `go list` describes the current build configuration. Configured platform/build-tag siblings are
  conservative unknown context, not proven active for the current platform.
- No persistent incremental cache; warm figures are benchmark observations only.
- Static practical impact is not prospective model-review effectiveness. Actual local-Codex token,
  cost, latency, experiment, and defect metrics remain unavailable until exact-head packet review.
- Peak Go subprocess RSS was not measured; only Python allocations were captured.

## Implementation slices and required RED tests

1. **Freeze correction corpus and RED.** Separate detector-visible inputs/oracle for upstream leaf,
   script/Python both directions, authority, generated chain, Go importer/test, platform variant,
   cycle, unknown condition, bound hit, and unrelated control. Root-only baseline must miss the
   intended cases.
2. **Index/parsers GREEN.** Build safe tracked universe, typed edge schema/certainty, standard-library
   cross-format parsers, configured relations, and authoritative Go package/test data.
3. **Traversal GREEN.** Materialize both adjacency directions; deterministic multi-source
   relation-policy BFS; explicit bound/frontier, cycle, missing, and unknown blockers.
4. **Coverage/packet GREEN.** Add impact files/edges to exact-version manifest and bounded packets;
   synthesis rejects omitted/stale/overflowed impact evidence.
5. **Measurement GREEN.** Reproduce baseline/treatment recall/precision, graph counts, cold/traversal
   time, impact size, bound hits, deterministic digest, and exact-head invalidation. Keep fixture,
   actual model, and prospective claims separate.
6. **Counterfactual lab RED/GREEN.** Implement the already-required isolated hypothesis-lab safety
   contract only after graph context is available; experiment responses reference examined impact
   edge ids.
7. **Full revalidation/review.** Prior verification is historical. Run focused/full gates, exact-head
   packet review, PM synthesis, independent Shepherd, no-mistakes, and CI on the final head.

## Dependency decision

The selected production design uses Python and Go standard tooling already required by the
repository. It adds **no shipped dependency** and does not modify `go.mod` or `go.sum`. No captain
dependency decision is needed. If future symbol-level indexing is proposed, Kythe/SCIP/CodeQL or
`go/packages` must be a separate measured, human-approved architecture decision.
