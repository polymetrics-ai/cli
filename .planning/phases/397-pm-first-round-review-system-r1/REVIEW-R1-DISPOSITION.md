# Exact-Head Local Codex Review Round 1 Disposition

**Exact base:** `0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20`

**Exact reviewed head:** `b1d869732d230575ab7c8295b15cef42cc0078ef`

**Exact tree:** `ec828909478685de6dcce8e095efa8ef255e4334`

**Stable lineage:** `0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20...pm-first-round-review-system-r1`
**Correction budget:** round 1 begins; `1/5` consumed only when this findings verdict is followed by correction and fresh exact-head review. Provider retries are separate.

## Synthesis and evidence

- Fresh-context runtime: `openai-codex/gpt-5.6-sol:xhigh` through repo-local Pi.
- Packet result: 17/17 completed; every packet returned findings; 89 raw findings; no unreviewed files.
- Raw compiler manifest SHA-256: `1f12f1c429f6fb9048917a7a0874212618b4236a20c385a12174dee0ce884763`.
- Raw synthesis SHA-256: `42626c4555d9769105b8d07ece6c831f44e4cbbdadcfa8a9b50842e9e4f8bd5c`.
- Machine synthesis returned `blocked` with 14 invariant blockers because it treated honestly failed assigned invariants as missing coverage. That behavior is itself accepted finding R1-B. The PM-owned lifecycle conclusion is `findings_correction_required`; it does not claim clean evidence.
- Raw responses and logs remain Git-ignored under `pm-review-evidence.tmp/b1d869732/`; per-packet hashes and provider attempts are committed in `REVIEW-R1-MEASUREMENT.json`.
- Independent Shepherd remains pending and is prohibited until a future exact-head synthesis is clean.

## Systemic correction groups

### R1-A
Bind compile/synthesis to a ready finding-free manifest, exact clean HEAD/tree, exact merge-base, and complete coverage; remove unsafe non-head worktree reads.

### R1-B
Implement strict dependency-free response/lab-evidence shape, exact set and status/content validation; accept explicit failed invariants as findings rather than missing coverage.

### R1-C
Carry independent relation/direction budgets in traversal state and detect global continuing frontiers; add mixed-path and convergent-policy fixtures.

### R1-D
Packet coherent file/edge neighborhoods and estimate exact-tree bytes conservatively, including edge endpoints and prompt metadata; block unsplittable context.

### R1-E
Make indexing/parser errors fail closed; preserve three-valued certainty from explicit context; validate endpoints/deletions/prohibited targets and cover shell, fixtures, root files, and unsafe paths.

### R1-F
Bound index resources before reads/materialization and enforce edge/node limits incrementally.

### R1-G
Run authoritative Go indexing offline with a proven pre-populated read-only module cache and base+head package context for deletions.

### R1-H
Harden labs with default-deny reads, immutable Git administration, full descendant containment/kill proof, and synthesis-bound lab evidence.

### R1-I
Reject explicit null/conflicting terminal schema aliases and preserve legacy only when schema keys are absent.

### R1-J
Separate reusable canonical policy from this branch delivery scope and bind a validated per-run scope to the manifest.

### R1-K
Reconcile canonical PM route siblings away from active legacy hosted-bot instructions without editing PR #493-owned paths.

### R1-L
Refresh parent/current authority state after PR #495 and replace the false temporal schema-instance claim with a validated phase-state contract/mapping.

### R1-M
Reconcile durable phase verification/TDD evidence at the next exact head and retain immutable historical entries.

### R1-N
Move remote freshness to the PM precondition; packet reviewers verify local exact identities under a no-network boundary.

### R1-O
Add root-file/source-document references such as AGENTS.md through explicit safe allowlists rather than broad docs indexing.

## Per-finding disposition

Every raw finding is disclosed. The first finding assigned to each systemic group is `accepted_with_modification`; later findings in the same group are `duplicate` and are covered by the same root-cause correction and regression suite. No finding is declined, hidden, waived, or deferred.

| ID | packet | severity | category | path:line | systemic group | disposition | response SHA-256 |
|---|---|---|---|---|---|---|---|
| R1-F001 | `architecture_reference-01` | high | `workflow_gate` | `.agents/agentic-delivery/contracts/parent-issue-roadmap-template.md:20, 30-32, 48-60` | R1-K | `accepted_with_modification` | `54b7fdd45615c69be2a82c3f8f27f4caa5f8cfabe01f768faf59143543d4a095` |
| R1-F002 | `architecture_reference-01` | medium | `safety_contract` | `.agents/agentic-delivery/workflows/local-codex-review-loop.md:25, 47-51` | R1-N | `accepted_with_modification` | `54b7fdd45615c69be2a82c3f8f27f4caa5f8cfabe01f768faf59143543d4a095` |
| R1-F003 | `architecture_reference-01` | medium | `machine_contract` | `.agents/agentic-delivery/contracts/pm-review-packet-template.md:45-57, 168-177` | R1-B | `accepted_with_modification` | `54b7fdd45615c69be2a82c3f8f27f4caa5f8cfabe01f768faf59143543d4a095` |
| R1-F004 | `architecture_reference-01` | high | `fail_closed_synthesis` | `scripts/pm-review-system.py:1564-1575` | R1-B | `duplicate` | `54b7fdd45615c69be2a82c3f8f27f4caa5f8cfabe01f768faf59143543d4a095` |
| R1-F005 | `authority_workflow_state-01` | medium | `authoritative_state_consistency` | `.planning/phases/397-cli-architecture-v2-orchestration/RUN-STATE.json:6-9, 687-692, 849-862` | R1-L | `accepted_with_modification` | `4471f32b2b838ecb0426236b5f121ea5c05d3158e6d799f0808159c35457a62d` |
| R1-F006 | `authority_workflow_state-01` | medium | `machine_contract` | `.planning/phases/397-pm-first-round-review-system-r1/RUN-STATE.json:2-15, 26-47` | R1-L | `duplicate` | `4471f32b2b838ecb0426236b5f121ea5c05d3158e6d799f0808159c35457a62d` |
| R1-F007 | `authority_workflow_state-01` | low | `evidence_truthfulness` | `.planning/phases/397-pm-first-round-review-system-r1/TDD-LEDGER.md:3, 28, 46` | R1-M | `accepted_with_modification` | `4471f32b2b838ecb0426236b5f121ea5c05d3158e6d799f0808159c35457a62d` |
| R1-F008 | `implementation_test-01` | high | `exact_version_binding` | `scripts/pm-review-system.py:1511` | R1-B | `duplicate` | `39c1a9177f308b45e9a05295d30dc1b2503c1a96087a82475391da7397ccbf95` |
| R1-F009 | `implementation_test-01` | high | `exact_version_binding` | `scripts/pm-review-system.py:1373` | R1-A | `accepted_with_modification` | `39c1a9177f308b45e9a05295d30dc1b2503c1a96087a82475391da7397ccbf95` |
| R1-F010 | `implementation_test-01` | high | `synthesis_contract` | `scripts/pm-review-system.py:1543` | R1-B | `duplicate` | `39c1a9177f308b45e9a05295d30dc1b2503c1a96087a82475391da7397ccbf95` |
| R1-F011 | `implementation_test-01` | high | `lab_safety` | `scripts/pm-review-system.py:1483` | R1-B | `duplicate` | `39c1a9177f308b45e9a05295d30dc1b2503c1a96087a82475391da7397ccbf95` |
| R1-F012 | `implementation_test-01` | high | `lab_safety` | `scripts/pm-review-lab.py:253` | R1-H | `accepted_with_modification` | `39c1a9177f308b45e9a05295d30dc1b2503c1a96087a82475391da7397ccbf95` |
| R1-F013 | `implementation_test-01` | high | `lab_safety` | `scripts/pm-review-lab.py:304` | R1-H | `duplicate` | `39c1a9177f308b45e9a05295d30dc1b2503c1a96087a82475391da7397ccbf95` |
| R1-F014 | `implementation_test-01` | high | `semantic_negative_fixtures` | `scripts/pm-terminal-classifier.sh:25` | R1-I | `accepted_with_modification` | `39c1a9177f308b45e9a05295d30dc1b2503c1a96087a82475391da7397ccbf95` |
| R1-F015 | `implementation_test-01` | medium | `impact_graph_correctness` | `scripts/pm-review-system.py:999` | R1-C | `accepted_with_modification` | `39c1a9177f308b45e9a05295d30dc1b2503c1a96087a82475391da7397ccbf95` |
| R1-F016 | `architecture_reference-02` | high | `correctness_exact_version_binding` | `scripts/pm-review-system.py:1512` | R1-B | `duplicate` | `1cb28f3c8cc1fa58c12c8dc29f7993481eeff089c947b471bd43599b533df899` |
| R1-F017 | `architecture_reference-02` | medium | `coverage_active_reference_closure` | `scripts/pm-review-system.py:48` | R1-O | `accepted_with_modification` | `1cb28f3c8cc1fa58c12c8dc29f7993481eeff089c947b471bd43599b533df899` |
| R1-F018 | `architecture_reference-02` | medium | `machine_contract_invariant_disposition` | `scripts/pm-review-system.py:1554` | R1-B | `duplicate` | `1cb28f3c8cc1fa58c12c8dc29f7993481eeff089c947b471bd43599b533df899` |
| R1-F019 | `architecture_reference-03` | high | `packet_bounds` | `scripts/pm-review-system.py:1160-1165, 1219-1224` | R1-D | `accepted_with_modification` | `ae10faca2d265767f033e59d5e648ec2f4ae0760517007fe6d95db0f4b7b578a` |
| R1-F020 | `architecture_reference-03` | medium | `active_reference_closure` | `.planning/traces/cli-architecture-v2-pi-prompts.md:8-10, 75-76` | R1-K | `duplicate` | `ae10faca2d265767f033e59d5e648ec2f4ae0760517007fe6d95db0f4b7b578a` |
| R1-F021 | `architecture_reference-03` | medium | `reference_parser` | `scripts/pm-review-system.py:480-522` | R1-E | `accepted_with_modification` | `ae10faca2d265767f033e59d5e648ec2f4ae0760517007fe6d95db0f4b7b578a` |
| R1-F022 | `impact_graph-01` | high | `impact_graph_correctness` | `scripts/pm-review-system.py:951-959, 1033-1035` | R1-E | `duplicate` | `e3c7df89a90a937af296437b38b1fe5ff63f5d91495bbbc69f09ac430225752b` |
| R1-F023 | `impact_graph-01` | high | `packet_bounds` | `scripts/pm-review-system.py:1160-1165, 1232-1242` | R1-D | `duplicate` | `e3c7df89a90a937af296437b38b1fe5ff63f5d91495bbbc69f09ac430225752b` |
| R1-F024 | `impact_graph-01` | high | `graph_bounds` | `scripts/pm-review-system.py:1025-1032` | R1-E | `duplicate` | `e3c7df89a90a937af296437b38b1fe5ff63f5d91495bbbc69f09ac430225752b` |
| R1-F025 | `impact_graph-01` | high | `stale_evidence` | `scripts/pm-review-system.py:1509-1524` | R1-B | `duplicate` | `e3c7df89a90a937af296437b38b1fe5ff63f5d91495bbbc69f09ac430225752b` |
| R1-F026 | `impact_graph-01` | medium | `machine_contract` | `scripts/pm-review-system.py:1564-1573` | R1-B | `duplicate` | `e3c7df89a90a937af296437b38b1fe5ff63f5d91495bbbc69f09ac430225752b` |
| R1-F027 | `impact_graph-01` | high | `workflow_contract` | `.agents/agentic-delivery/agents/implementation/issue-first-implementation-agent.agent.yaml:42-44, 88-94, 103-108, 120` | R1-K | `duplicate` | `e3c7df89a90a937af296437b38b1fe5ff63f5d91495bbbc69f09ac430225752b` |
| R1-F028 | `impact_graph-01` | medium | `exact_version_binding` | `scripts/pm-review-system.py:1097-1101, 1368-1380` | R1-A | `duplicate` | `e3c7df89a90a937af296437b38b1fe5ff63f5d91495bbbc69f09ac430225752b` |
| R1-F029 | `impact_graph-02` | high | `exact_version_binding` | `scripts/pm-review-system.py:1511-1575` | R1-B | `duplicate` | `762112e135fc07bd94784bd6bf8a5077b8c367152fd7e36d255e6277541cff76` |
| R1-F030 | `impact_graph-02` | high | `impact_graph_correctness` | `scripts/pm-review-system.py:881-904, 1033-1040` | R1-E | `duplicate` | `762112e135fc07bd94784bd6bf8a5077b8c367152fd7e36d255e6277541cff76` |
| R1-F031 | `impact_graph-02` | high | `workflow_regression` | `.agents/agentic-delivery/contracts/pm-review-system.json:29-44, 233-268` | R1-J | `accepted_with_modification` | `762112e135fc07bd94784bd6bf8a5077b8c367152fd7e36d255e6277541cff76` |
| R1-F032 | `impact_graph-02` | high | `packet_bounding` | `scripts/pm-review-system.py:1160-1165, 1232-1243` | R1-D | `duplicate` | `762112e135fc07bd94784bd6bf8a5077b8c367152fd7e36d255e6277541cff76` |
| R1-F033 | `impact_graph-02` | medium | `exact_version_binding` | `scripts/pm-review-system.py:867-904, 926-929, 1373-1379, 1623` | R1-A | `duplicate` | `762112e135fc07bd94784bd6bf8a5077b8c367152fd7e36d255e6277541cff76` |
| R1-F034 | `impact_graph-02` | medium | `machine_contract` | `scripts/pm-review-system.py:1524-1569` | R1-B | `duplicate` | `762112e135fc07bd94784bd6bf8a5077b8c367152fd7e36d255e6277541cff76` |
| R1-F035 | `impact_graph-03` | high | `correctness` | `scripts/pm-review-system.py:999` | R1-C | `duplicate` | `f4621eea8a1e124ec020c053d14f3dd8b809717f63bf8a16db3f5239886bd34d` |
| R1-F036 | `impact_graph-03` | high | `exact_version_binding` | `scripts/pm-review-system.py:1510` | R1-B | `duplicate` | `f4621eea8a1e124ec020c053d14f3dd8b809717f63bf8a16db3f5239886bd34d` |
| R1-F037 | `impact_graph-03` | high | `fail_closed_synthesis` | `scripts/pm-review-system.py:1512` | R1-B | `duplicate` | `f4621eea8a1e124ec020c053d14f3dd8b809717f63bf8a16db3f5239886bd34d` |
| R1-F038 | `impact_graph-03` | high | `exact_version_binding` | `scripts/pm-review-system.py:1373` | R1-A | `duplicate` | `f4621eea8a1e124ec020c053d14f3dd8b809717f63bf8a16db3f5239886bd34d` |
| R1-F039 | `impact_graph-03` | high | `packet_boundedness` | `scripts/pm-review-system.py:1160` | R1-D | `duplicate` | `f4621eea8a1e124ec020c053d14f3dd8b809717f63bf8a16db3f5239886bd34d` |
| R1-F040 | `impact_graph-03` | high | `graph_fail_closed` | `scripts/pm-review-system.py:736` | R1-E | `duplicate` | `f4621eea8a1e124ec020c053d14f3dd8b809717f63bf8a16db3f5239886bd34d` |
| R1-F041 | `impact_graph-03` | medium | `typed_edge_provenance` | `scripts/pm-review-system.py:701` | R1-E | `duplicate` | `f4621eea8a1e124ec020c053d14f3dd8b809717f63bf8a16db3f5239886bd34d` |
| R1-F042 | `impact_graph-04` | high | `impact_graph_correctness` | `scripts/pm-review-system.py:999` | R1-C | `duplicate` | `0ad752df40a1b56853fdb62ccb5186388dd63684427f79f00dcf8e9a84194303` |
| R1-F043 | `impact_graph-04` | high | `packet_coverage` | `scripts/pm-review-system.py:1232` | R1-D | `duplicate` | `0ad752df40a1b56853fdb62ccb5186388dd63684427f79f00dcf8e9a84194303` |
| R1-F044 | `impact_graph-04` | high | `exact_version_binding` | `scripts/pm-review-system.py:1512` | R1-B | `duplicate` | `0ad752df40a1b56853fdb62ccb5186388dd63684427f79f00dcf8e9a84194303` |
| R1-F045 | `impact_graph-04` | medium | `typed_edge_certainty` | `scripts/pm-review-system.py:737` | R1-E | `duplicate` | `0ad752df40a1b56853fdb62ccb5186388dd63684427f79f00dcf8e9a84194303` |
| R1-F046 | `impact_graph-05` | high | `impact_graph_recall` | `scripts/pm-review-system.py:999-1029` | R1-C | `duplicate` | `4b0a330ec5b57d2cbfa753750e4fe585c70a7245fc7d67831006904e6306a095` |
| R1-F047 | `impact_graph-05` | high | `impact_certainty` | `scripts/pm-review-system.py:736-742` | R1-E | `duplicate` | `4b0a330ec5b57d2cbfa753750e4fe585c70a7245fc7d67831006904e6306a095` |
| R1-F048 | `impact_graph-05` | high | `exact_version_binding` | `scripts/pm-review-system.py:1373-1378, 1623` | R1-A | `duplicate` | `4b0a330ec5b57d2cbfa753750e4fe585c70a7245fc7d67831006904e6306a095` |
| R1-F049 | `impact_graph-05` | high | `go_impact_availability` | `scripts/pm-review-system.py:773-789, 926-933` | R1-G | `accepted_with_modification` | `4b0a330ec5b57d2cbfa753750e4fe585c70a7245fc7d67831006904e6306a095` |
| R1-F050 | `impact_graph-05` | medium | `workflow_contract_divergence` | `.agents/agentic-delivery/contracts/parent-issue-roadmap-template.md:20, 32, 57-60` | R1-K | `duplicate` | `4b0a330ec5b57d2cbfa753750e4fe585c70a7245fc7d67831006904e6306a095` |
| R1-F051 | `impact_graph-05` | low | `impact_graph_precision` | `scripts/pm-review-system.py:653-674` | R1-E | `duplicate` | `4b0a330ec5b57d2cbfa753750e4fe585c70a7245fc7d67831006904e6306a095` |
| R1-F052 | `impact_graph-06` | high | `exact_version_binding` | `scripts/pm-review-system.py:1509-1574` | R1-B | `duplicate` | `3d2aa35794845a27eede4e63196a776dd7d886cdf187641634fdd419802dc708` |
| R1-F053 | `impact_graph-06` | high | `impact_graph_fail_open` | `scripts/pm-review-system.py:691-695` | R1-E | `duplicate` | `3d2aa35794845a27eede4e63196a776dd7d886cdf187641634fdd419802dc708` |
| R1-F054 | `impact_graph-06` | high | `packet_coverage_bounds` | `scripts/pm-review-system.py:1160-1164,1232-1243,1543-1553` | R1-D | `duplicate` | `3d2aa35794845a27eede4e63196a776dd7d886cdf187641634fdd419802dc708` |
| R1-F055 | `impact_graph-06` | medium | `temporal_edge_provenance` | `.agents/agentic-delivery/contracts/pm-review-system.json:82-88` | R1-L | `duplicate` | `3d2aa35794845a27eede4e63196a776dd7d886cdf187641634fdd419802dc708` |
| R1-F056 | `impact_graph-06` | medium | `resource_bounds` | `scripts/pm-review-system.py:882-910,949-953` | R1-F | `accepted_with_modification` | `3d2aa35794845a27eede4e63196a776dd7d886cdf187641634fdd419802dc708` |
| R1-F057 | `impact_graph-07` | high | `packet_bounds` | `scripts/pm-review-system.py:1160-1167, 1232-1243` | R1-D | `duplicate` | `565d83b118fba35bd6f493b9acd73bdb6311b376ff5593476473216cd50b70be` |
| R1-F058 | `impact_graph-07` | high | `exact_version_synthesis` | `scripts/pm-review-system.py:1509-1524, 1570-1592` | R1-B | `duplicate` | `565d83b118fba35bd6f493b9acd73bdb6311b376ff5593476473216cd50b70be` |
| R1-F059 | `impact_graph-07` | high | `exact_version_compile` | `scripts/pm-review-system.py:890-904, 1372-1392` | R1-A | `duplicate` | `565d83b118fba35bd6f493b9acd73bdb6311b376ff5593476473216cd50b70be` |
| R1-F060 | `impact_graph-07` | medium | `response_contract` | `scripts/pm-review-system.py:1564-1575` | R1-B | `duplicate` | `565d83b118fba35bd6f493b9acd73bdb6311b376ff5593476473216cd50b70be` |
| R1-F061 | `impact_graph-07` | high | `go_impact_coverage` | `scripts/pm-review-system.py:881, 890-892, 927-933` | R1-G | `duplicate` | `565d83b118fba35bd6f493b9acd73bdb6311b376ff5593476473216cd50b70be` |
| R1-F062 | `impact_graph-07` | high | `edge_certainty` | `scripts/pm-review-system.py:736-742` | R1-E | `duplicate` | `565d83b118fba35bd6f493b9acd73bdb6311b376ff5593476473216cd50b70be` |
| R1-F063 | `impact_graph-08` | high | `go_impact_index_unusable` | `scripts/pm-review-system.py:770-793, 927-932` | R1-G | `duplicate` | `a4aa3bf5a2e551309185eb9a4cd41c7338e75753d7564973c147f0e052520471` |
| R1-F064 | `impact_graph-08` | high | `packet_bound_underestimated` | `scripts/pm-review-system.py:1160-1167` | R1-D | `duplicate` | `a4aa3bf5a2e551309185eb9a4cd41c7338e75753d7564973c147f0e052520471` |
| R1-F065 | `impact_graph-08` | high | `mixed_relation_paths_omitted` | `scripts/pm-review-system.py:999-1029` | R1-C | `duplicate` | `a4aa3bf5a2e551309185eb9a4cd41c7338e75753d7564973c147f0e052520471` |
| R1-F066 | `impact_graph-08` | medium | `certainty_semantics_incorrect` | `scripts/pm-review-system.py:646-648` | R1-E | `duplicate` | `a4aa3bf5a2e551309185eb9a4cd41c7338e75753d7564973c147f0e052520471` |
| R1-F067 | `impact_graph-08` | medium | `packet_coverage_not_exact` | `scripts/pm-review-system.py:1542-1553` | R1-B | `duplicate` | `a4aa3bf5a2e551309185eb9a4cd41c7338e75753d7564973c147f0e052520471` |
| R1-F068 | `impact_graph-08` | medium | `exact_head_verification_evidence_gap` | `.planning/phases/397-pm-first-round-review-system-r1/RUN-STATE.json:29-37` | R1-M | `duplicate` | `a4aa3bf5a2e551309185eb9a4cd41c7338e75753d7564973c147f0e052520471` |
| R1-F069 | `impact_graph-09` | high | `exact_version_binding` | `scripts/pm-review-system.py:1512-1575` | R1-B | `duplicate` | `a96df4970ffb9509b7a946cfb19ddc4d02efe96916e092e2a85d5726eb5947b7` |
| R1-F070 | `impact_graph-09` | high | `exact_version_binding` | `scripts/pm-review-system.py:1097-1101, 1357-1382, 1623` | R1-A | `duplicate` | `a96df4970ffb9509b7a946cfb19ddc4d02efe96916e092e2a85d5726eb5947b7` |
| R1-F071 | `impact_graph-09` | high | `impact_graph_completeness` | `scripts/pm-review-system.py:999-1040` | R1-C | `duplicate` | `a96df4970ffb9509b7a946cfb19ddc4d02efe96916e092e2a85d5726eb5947b7` |
| R1-F072 | `impact_graph-09` | high | `path_and_reference_safety` | `scripts/pm-review-system.py:691-695, 716-719, 736-742` | R1-E | `duplicate` | `a96df4970ffb9509b7a946cfb19ddc4d02efe96916e092e2a85d5726eb5947b7` |
| R1-F073 | `impact_graph-09` | medium | `packet_coverage` | `scripts/pm-review-system.py:1543-1553` | R1-B | `duplicate` | `a96df4970ffb9509b7a946cfb19ddc4d02efe96916e092e2a85d5726eb5947b7` |
| R1-F074 | `impact_graph-09` | medium | `graph_bounds` | `scripts/pm-review-system.py:876-905, 955-961` | R1-F | `duplicate` | `a96df4970ffb9509b7a946cfb19ddc4d02efe96916e092e2a85d5726eb5947b7` |
| R1-F075 | `impact_graph-10` | high | `exact_version_synthesis` | `scripts/pm-review-system.py:1509-1589` | R1-B | `duplicate` | `be1c35ae7960bf270cc3181212f4439f5fd52a85623efb8de37c610a0f6f5602` |
| R1-F076 | `impact_graph-10` | high | `exact_version_compilation` | `scripts/pm-review-system.py:1372-1379, 1623` | R1-A | `duplicate` | `be1c35ae7960bf270cc3181212f4439f5fd52a85623efb8de37c610a0f6f5602` |
| R1-F077 | `impact_graph-10` | high | `prohibited_active_reference` | `scripts/pm-review-system.py:541-556, 974-1045` | R1-E | `duplicate` | `be1c35ae7960bf270cc3181212f4439f5fd52a85623efb8de37c610a0f6f5602` |
| R1-F078 | `impact_graph-10` | medium | `impact_graph_fail_open` | `scripts/pm-review-system.py:691-695, 736-742` | R1-E | `duplicate` | `be1c35ae7960bf270cc3181212f4439f5fd52a85623efb8de37c610a0f6f5602` |
| R1-F079 | `impact_graph-10` | high | `hypothesis_lab_information_disclosure` | `scripts/pm-review-lab.py:144-163` | R1-H | `duplicate` | `be1c35ae7960bf270cc3181212f4439f5fd52a85623efb8de37c610a0f6f5602` |
| R1-F080 | `impact_graph-11` | high | `exact_version_binding` | `scripts/pm-review-system.py:1511-1542` | R1-B | `duplicate` | `60f09c715be23da4b988dce310d681ecbd1741b1c9dc25ef0d50623eb2822892` |
| R1-F081 | `impact_graph-11` | high | `exact_version_binding` | `scripts/pm-review-system.py:859-904, 1372-1379, 1623` | R1-A | `duplicate` | `60f09c715be23da4b988dce310d681ecbd1741b1c9dc25ef0d50623eb2822892` |
| R1-F082 | `impact_graph-11` | high | `impact_graph_contract` | `scripts/pm-review-system.py:881, 890-892, 1034-1036` | R1-E | `duplicate` | `60f09c715be23da4b988dce310d681ecbd1741b1c9dc25ef0d50623eb2822892` |
| R1-F083 | `impact_graph-11` | high | `typed_edge_certainty` | `scripts/pm-review-system.py:736-742` | R1-E | `duplicate` | `60f09c715be23da4b988dce310d681ecbd1741b1c9dc25ef0d50623eb2822892` |
| R1-F084 | `impact_graph-11` | medium | `packet_bounds` | `scripts/pm-review-system.py:1160-1163, 1232-1242` | R1-D | `duplicate` | `60f09c715be23da4b988dce310d681ecbd1741b1c9dc25ef0d50623eb2822892` |
| R1-F085 | `impact_graph-12` | high | `exact_version_binding` | `scripts/pm-review-system.py:1373-1379, 1623` | R1-A | `duplicate` | `d2dc0fead7cd8f6e8da6cbbd9ad46c67c4e5bc3e6e565bfb736bf39ee90ccc8b` |
| R1-F086 | `impact_graph-12` | medium | `impact_coverage` | `scripts/pm-review-system.py:679-683` | R1-E | `duplicate` | `d2dc0fead7cd8f6e8da6cbbd9ad46c67c4e5bc3e6e565bfb736bf39ee90ccc8b` |
| R1-F087 | `impact_graph-12` | medium | `resource_bound` | `scripts/pm-review-system.py:882-900` | R1-F | `duplicate` | `d2dc0fead7cd8f6e8da6cbbd9ad46c67c4e5bc3e6e565bfb736bf39ee90ccc8b` |
| R1-F088 | `impact_graph-12` | medium | `certainty_classification` | `scripts/pm-review-system.py:644-650` | R1-E | `duplicate` | `d2dc0fead7cd8f6e8da6cbbd9ad46c67c4e5bc3e6e565bfb736bf39ee90ccc8b` |
| R1-F089 | `impact_graph-12` | medium | `synthesis_contract` | `scripts/pm-review-system.py:1564-1574` | R1-B | `duplicate` | `d2dc0fead7cd8f6e8da6cbbd9ad46c67c4e5bc3e6e565bfb736bf39ee90ccc8b` |

## Gate state

- `local_codex.status`: `findings_correction_required` (raw machine synthesis blocked by accepted R1-B defect; not clean)
- `local_codex.disposition_artifact`: this file
- `shepherd.status`: `pending`
- `shepherd.verdict`: absent
- `human.gate`: none for these corrections; PR #493-owned paths remain forbidden; no merge by this agent.

## Recurrence monitoring

Round 1 is the recurrence baseline. Before any later correction, compare new findings by systemic group. A repeated accepted defect triggers root-cause diagnosis rather than another local patch. At 5/5, automatic correction stops with unresolved root-cause evidence. Provider startup/auth/WebSocket attempts follow the separate bounded retry record in `REVIEW-R1-MEASUREMENT.json`.
