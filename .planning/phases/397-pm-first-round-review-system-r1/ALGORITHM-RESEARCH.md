# PM Impact-Graph Algorithm Research and Selection

> Research narration condensed; the selected architecture, invariants, rejected alternatives, gaps,
> implementation slices, and dependency decision below are authoritative.

## Decision scope, method, and sources

Goal: choose a repository-fit practical file/package impact algorithm that is correct (no escaped
required references), safe (fail-closed bounds), deterministic, dependency-free, and fast enough for
exact-head compilation. Method: primary-source review of build/query/test-impact/graph systems plus
a disposable exact-head prototype benchmarked against a frozen oracle and adversarial fixtures.

Primary sources reviewed: Bazel/Skyframe, Buck2/DICE, Pants, Nx, Turborepo, Gradle build graphs;
CodeQL/CPG, Kythe, SCIP code graphs; Go tooling (`go list`, `go/packages`); test-impact systems
(Ekstazi, STARTS, Microsoft TIA); and graph algorithms (multi-source BFS/DFS, Tarjan SCC). Key
findings: the review universe and configuration are correctness inputs; reverse adjacency is normal
but blindly undirecting every edge is not; static and dynamic evidence are complementary; and
unknown conditions must never be relabeled inactive.

## Measured benchmark (prototype head `5601be8d0`)

- Real prototype: indexed 1,669 nodes / 5,039 edges in 2,004.212 ms, 51,118,867 Python-allocated
  peak bytes.
- Correctness vs oracle: indexed bidirectional BFS found **21/21** required nodes with 0 false
  positives; outgoing-only found **14/21** (the captain-identified escape); repeated reverse scans
  matched adjacency correctness but were ~164× slower on the synthetic graph.
- Traversal: a 20K-node / 30,197-edge synthetic traversal took **63.735 ms** indexed vs
  **10,479.531 ms** repeated-scan.
- Unrestricted depth-3 from one PM seed reached 201 PM / 213 Go nodes and still hit a frontier —
  proof that a typed relation policy plus fail-closed bounds is required over whole-graph inclusion.

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

- **Outgoing-only roots/diff:** escaped 7 of 21 required fixture nodes; the captain-identified defect.
- **Repeated reverse scans:** same fixture correctness as adjacency but ~164× slower and `O(VE)`.
- **Unrestricted bidirectional graph:** one PM seed reached 201 nodes by depth 3 and still hit a
  frontier; not efficient review context.
- **SCC condensation now:** sound and cheap, but visited BFS already handles cycles; revisit for
  repeated component queries.
- **Persistent DICE/Skyframe-style cache now:** cold cost is acceptable; cache invalidation adds more
  failure modes than measured benefit for one exact-head compilation.
- **`go/packages`, Kythe, SCIP, CodeQL/CPG:** valuable future semantic options, but add dependencies,
  build capture, indexers, or symbol semantics beyond this bounded PR.
- **Dynamic-only test impact:** no complete prior runtime map; cannot cover unexecuted paths.
- **Packet-budget traversal:** fast but silently loses impact and violates the required contract.

## Known gaps and non-claims

- No symbol-level caller/callee, control-flow, alias, or data-flow precision.
- Python dynamic import, shell-computed paths, reflection, generated-at-runtime files, and arbitrary
  templating may remain unknown and block when relevant.
- `go list` describes the current build configuration; configured platform/build-tag siblings are
  conservative unknown context, not proven active.
- No persistent incremental cache; warm figures are benchmark observations only.
- Static practical impact is not prospective model-review effectiveness; actual local-Codex token,
  cost, latency, experiment, and defect metrics remain unavailable until exact-head packet review.
- Peak Go subprocess RSS was not measured; only Python allocations were captured.

## Implementation slices and required RED tests

1. **Freeze correction corpus and RED.** Separate detector-visible inputs/oracle for upstream leaf,
   script/Python both directions, authority, generated chain, Go importer/test, platform variant,
   cycle, unknown condition, bound hit, and unrelated control. Root-only baseline must miss them.
2. **Index/parsers GREEN.** Safe tracked universe, typed edge schema/certainty, standard-library
   cross-format parsers, configured relations, authoritative Go package/test data.
3. **Traversal GREEN.** Both adjacency directions; deterministic multi-source relation-policy BFS;
   explicit bound/frontier, cycle, missing, and unknown blockers.
4. **Coverage/packet GREEN.** Impact files/edges in the exact-version manifest and bounded packets;
   synthesis rejects omitted/stale/overflowed impact evidence.
5. **Measurement GREEN.** Baseline/treatment recall/precision, graph counts, cold/traversal time,
   impact size, bound hits, deterministic digest, exact-head invalidation; fixture/actual/prospective
   claims separate.
6. **Counterfactual lab RED/GREEN.** Isolated hypothesis-lab safety contract after graph context is
   available; experiment responses reference examined impact edge ids.
7. **Full revalidation/review.** Focused/full gates, exact-head packet review, PM synthesis,
   independent Shepherd, no-mistakes, and CI on the final head.

## Dependency decision

The selected design uses Python and Go standard tooling already required by the repository. It adds
**no shipped dependency** and does not modify `go.mod`/`go.sum`. Future symbol-level indexing
(Kythe/SCIP/CodeQL or `go/packages`) would be a separate measured, human-approved decision.
