# Exact-Head Local Codex Review Round 2 Disposition

**Exact base:** `0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20`  
**Exact reviewed head:** `92ce5e6a19cb7562aead8b224e6ba8dcc0857d34`  
**Exact tree:** `0a3c3eda637210d240dd13cfd99b475d032c36c0`

**Stable lineage:** `0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20...pm-first-round-review-system-r1`  
**Correction budget:** round 2 begins; `2/5` is consumed by this synthesized findings verdict and its systemic correction. Provider/OAuth/WebSocket/context-window retries remain operational attempts and do not consume correction rounds.

## Synthesis and recurrence result

- Fresh-context runtime: `openai-codex/gpt-5.6-sol:xhigh` through repo-local Pi.
- 41/41 bounded packet reviews completed with zero response blockers; synthesis is `findings_correction_required` with 141 raw findings (78 high, 60 medium, 3 low).
- Manifest SHA-256: `6ae018048856f917a7745602477b92fbee82edc6a7203e5226fc14d71022b85e`.
- Synthesis SHA-256: `cce8c3dfe95700d5f001d5b3cdf4eaa3f24dde407f9cf06356a7c50e3c50107a`.
- Three recorded first-attempt provider failures (one OAuth, two WebSocket) succeeded on attempt two. One separately observed context-window rejection is recorded as an operational attempt; it is not correction round 2. A second matching context failure would stop retries and require rendered-prompt accounting diagnosis.
- Recurrence is systemic: R1 packet budgeting tested its own average-style estimate; graph parsing remained line-local; synthesis authenticated identities but not manifest semantics; lab tests accepted generic blocked outcomes; phase validation checked markers rather than full state. Round 2 replaces those mechanisms and adds independent adversarial fixtures.
- Shepherd remains pending and prohibited until a future exact-head synthesis is clean.

## Root-cause groups

- **R1-A** (14): recurrence: immutable compile, exact blob, and authenticated manifest/synthesis binding.
- **R1-B** (13): recurrence: exact response/hypothesis/lab-evidence machine contract.
- **R1-C** (1): recurrence: traversal-direction and policy-state verification.
- **R1-D** (30): recurrence: rendered-prompt packet accounting, atomic slices, and inspectable provenance.
- **R1-E** (27): recurrence: format-aware typed graph semantics and base/head indexing.
- **R1-F** (2): recurrence: pre-materialization graph resource bounds.
- **R1-G** (2): recurrence: authoritative offline Go impact availability and reachability.
- **R1-H** (8): recurrence: lab read/process/resource/state/network containment.
- **R1-I** (1): recurrence: exact current-schema machine vocabulary.
- **R1-K** (6): recurrence: active PM route parity and activation gates.
- **R1-L** (10): recurrence: phase/authority/mirror validation and migration.
- **R1-M** (2): recurrence: durable evidence truthfulness.
- **R1-O** (4): recurrence: safe relative/docs/root reference closure.
- **R2-N1** (1): new: measurement case-set and corpus binding.
- **R2-N2** (4): new: runtime adapter/tool/documentation parity.
- **R2-N3** (8): new: exact-identity Shepherd and bounded driver authority.
- **R2-N4** (5): new: trace containment, identity, non-overwrite, and redaction.
- **R2-N5** (2): new: connector worker/ownership lifecycle truthfulness.
- **R2-N6** (1): new needs-human: GitHub connector operation-ledger behavior.

R2-N6 is disclosed but not implemented here: changing GitHub connector pending-review write coverage is an auth/write-surface product decision outside this branch and remains human-gated. The parser/ownership correction removes that unrelated pre-existing ledger from this candidate's practical impact packets; it is not hidden or declared fixed.

PR #493-owned paths remain forbidden. R2-F018 is corrected through a fail-closed activation/integration prerequisite in allowed route surfaces; the forbidden source itself is not edited.

## Per-finding disposition

The first finding in each root group is `accepted_with_modification`; subsequent manifestations are `duplicate` under the same systemic correction. `duplicate` means one root-cause treatment, not dismissal. Every recurrence gets a regression fixture. No finding is declined or hidden.

| ID | packet | severity | category | path:line | root group | disposition | response SHA-256 |
|---|---|---|---|---|---|---|---|
| R2-F001 | `architecture_reference-01` | medium | `machine_contract` | `.agents/agentic-delivery/schemas/orchestration-state.schema.yaml:127` | R1-B | `accepted_with_modification` | `0b75a9cc5360f7385122045bf758c6361d61f2def380af0d736d856396d46b5d` |
| R2-F002 | `architecture_reference-01` | high | `exact_version_binding` | `.agents/agentic-delivery/schemas/pm-review-system-phase-state.schema.json:35-36` | R1-L | `accepted_with_modification` | `0b75a9cc5360f7385122045bf758c6361d61f2def380af0d736d856396d46b5d` |
| R2-F003 | `architecture_reference-02` | medium | `workflow_contract` | `.pi/prompts/pm-review-loop.md:24-30` | R1-A | `accepted_with_modification` | `4b1fd06032a19295831fb8a6e4e9000df6f32e73e67dd1e98ce1a098580c3fc1` |
| R2-F004 | `authority_workflow_state-01` | high | `machine_contract` | `.agents/agentic-delivery/schemas/pm-review-system-phase-state.schema.json:33-36` | R1-L | `duplicate` | `5cf8c92c729fc09d3cd2397508609f1b727b95ca7a06a76d7df98a33291f93fa` |
| R2-F005 | `authority_workflow_state-01` | medium | `schema_migration` | `.planning/traces/cli-architecture-v2-orchestration-state.yaml:1, 13, 22` | R1-L | `duplicate` | `5cf8c92c729fc09d3cd2397508609f1b727b95ca7a06a76d7df98a33291f93fa` |
| R2-F006 | `authority_workflow_state-01` | low | `authoritative_state_consistency` | `.planning/traces/cli-architecture-v2-orchestration-state.yaml:831` | R1-L | `duplicate` | `5cf8c92c729fc09d3cd2397508609f1b727b95ca7a06a76d7df98a33291f93fa` |
| R2-F007 | `authority_workflow_state-01` | low | `evidence_truthfulness` | `.planning/phases/397-pm-first-round-review-system-r1/TDD-LEDGER.md:3, 97` | R1-M | `accepted_with_modification` | `5cf8c92c729fc09d3cd2397508609f1b727b95ca7a06a76d7df98a33291f93fa` |
| R2-F008 | `implementation_test-01` | high | `packet_bounds` | `scripts/pm-review-system.py:1563-1565` | R1-D | `accepted_with_modification` | `a4b1f93c33447b0c6456ca926824b608216f1eb9c8729d3509b44c82bb3dc594` |
| R2-F009 | `implementation_test-01` | high | `fail_closed_synthesis` | `scripts/pm-review-system.py:2199-2234` | R1-A | `duplicate` | `a4b1f93c33447b0c6456ca926824b608216f1eb9c8729d3509b44c82bb3dc594` |
| R2-F010 | `implementation_test-01` | high | `path_safety` | `scripts/pm-review-system.py:2252-2258` | R1-E | `accepted_with_modification` | `a4b1f93c33447b0c6456ca926824b608216f1eb9c8729d3509b44c82bb3dc594` |
| R2-F011 | `implementation_test-01` | high | `lab_safety` | `scripts/pm-review-lab.py:149-171` | R1-H | `accepted_with_modification` | `a4b1f93c33447b0c6456ca926824b608216f1eb9c8729d3509b44c82bb3dc594` |
| R2-F012 | `implementation_test-01` | high | `lab_evidence_binding` | `scripts/pm-review-system.py:2113-2143` | R1-B | `duplicate` | `a4b1f93c33447b0c6456ca926824b608216f1eb9c8729d3509b44c82bb3dc594` |
| R2-F013 | `implementation_test-01` | medium | `resource_bounds` | `scripts/pm-review-lab.py:369-416` | R1-H | `duplicate` | `a4b1f93c33447b0c6456ca926824b608216f1eb9c8729d3509b44c82bb3dc594` |
| R2-F014 | `implementation_test-01` | medium | `measurement_integrity` | `scripts/pm-review-system.py:1837-1877` | R2-N1 | `accepted_with_modification` | `a4b1f93c33447b0c6456ca926824b608216f1eb9c8729d3509b44c82bb3dc594` |
| R2-F015 | `implementation_test-01` | medium | `machine_contract` | `scripts/pm-review-system.py:2257-2262, 2331` | R1-B | `duplicate` | `a4b1f93c33447b0c6456ca926824b608216f1eb9c8729d3509b44c82bb3dc594` |
| R2-F016 | `implementation_test-01` | medium | `go_impact_compatibility` | `scripts/pm-review-system.py:882-884` | R1-G | `accepted_with_modification` | `a4b1f93c33447b0c6456ca926824b608216f1eb9c8729d3509b44c82bb3dc594` |
| R2-F017 | `implementation_test-01` | medium | `lab_functionality` | `scripts/pm-review-lab.py:161, 221-230` | R1-H | `duplicate` | `a4b1f93c33447b0c6456ca926824b608216f1eb9c8729d3509b44c82bb3dc594` |
| R2-F018 | `architecture_reference-03` | medium | `workflow_contract_divergence` | `.agents/agentic-delivery/references/required-skills-routing.md:112-116` | R1-K | `accepted_with_modification` | `b94a8dcad44c132b50355a84a5a6079d74bbf11548ac27b522a61c0a290e6297` |
| R2-F019 | `architecture_reference-03` | medium | `tool_scope_contract` | `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md:115-120` | R2-N2 | `accepted_with_modification` | `b94a8dcad44c132b50355a84a5a6079d74bbf11548ac27b522a61c0a290e6297` |
| R2-F020 | `architecture_reference-03` | medium | `runtime_dependency_contract` | `.agents/agentic-delivery/workflows/pi-active-orchestration-loop.md:6-8` | R2-N2 | `duplicate` | `b94a8dcad44c132b50355a84a5a6079d74bbf11548ac27b522a61c0a290e6297` |
| R2-F021 | `architecture_reference-04` | high | `exact_identity_safety` | `.agents/agentic-delivery/workflows/shepherd-validator.md:51-61,83-97` | R2-N3 | `accepted_with_modification` | `3a72d650985931c36a421b581a823c9a1a927d3de9cbfc5e6dcc1849cd22d893` |
| R2-F022 | `architecture_reference-04` | high | `human_gate_safety` | `.agents/agentic-delivery/workflows/shepherd-validator.md:41-42,55-58,107` | R2-N3 | `duplicate` | `3a72d650985931c36a421b581a823c9a1a927d3de9cbfc5e6dcc1849cd22d893` |
| R2-F023 | `architecture_reference-05` | high | `workflow_gate` | `scripts/pi-auto-loop.sh:109-129` | R2-N3 | `duplicate` | `9c5cbf00a5f4981366c068a2489943f3b4fa46e8f2804b2e175ab9fad1839121` |
| R2-F024 | `architecture_reference-05` | medium | `path_safety` | `scripts/loop-trace.sh:145-175` | R2-N4 | `accepted_with_modification` | `9c5cbf00a5f4981366c068a2489943f3b4fa46e8f2804b2e175ab9fad1839121` |
| R2-F025 | `architecture_reference-05` | medium | `evidence_integrity` | `scripts/loop-trace.sh:172-175` | R2-N4 | `duplicate` | `9c5cbf00a5f4981366c068a2489943f3b4fa46e8f2804b2e175ab9fad1839121` |
| R2-F026 | `architecture_reference-05` | medium | `secret_handling` | `scripts/loop-trace.sh:124-138` | R2-N4 | `duplicate` | `9c5cbf00a5f4981366c068a2489943f3b4fa46e8f2804b2e175ab9fad1839121` |
| R2-F027 | `architecture_reference-05` | medium | `evidence_isolation` | `scripts/loop-trace.sh:26-70` | R2-N4 | `duplicate` | `9c5cbf00a5f4981366c068a2489943f3b4fa46e8f2804b2e175ab9fad1839121` |
| R2-F028 | `authority_workflow_state-02` | high | `stale_evidence` | `.agents/agentic-delivery/workflows/shepherd-validator.md:83-99` | R2-N3 | `duplicate` | `ae93da8f4a7ccbab353ea0ddefbfe94dae93a0366eafcaa5779323d0f4dab1ac` |
| R2-F029 | `authority_workflow_state-02` | medium | `exact_current_schema_enums` | `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md:111-112` | R1-I | `accepted_with_modification` | `ae93da8f4a7ccbab353ea0ddefbfe94dae93a0366eafcaa5779323d0f4dab1ac` |
| R2-F030 | `authority_workflow_state-02` | medium | `authoritative_state_consistency` | `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md:127-138` | R2-N3 | `duplicate` | `ae93da8f4a7ccbab353ea0ddefbfe94dae93a0366eafcaa5779323d0f4dab1ac` |
| R2-F031 | `authority_workflow_state-02` | medium | `revert_safety` | `.agents/agentic-delivery/workflows/shepherd-validator.md:55-73` | R2-N3 | `duplicate` | `ae93da8f4a7ccbab353ea0ddefbfe94dae93a0366eafcaa5779323d0f4dab1ac` |
| R2-F032 | `impact_graph-01` | medium | `impact_graph_relation_typing` | `scripts/pm-review-system.py:810` | R1-E | `duplicate` | `f114d56e9d0da1357bb3e289343608d4ca84db46e93c95ef52b5976648e1b628` |
| R2-F033 | `impact_graph-02` | medium | `typed_edge_provenance` | `scripts/pm-review-system.py:699-719, 806-817` | R1-E | `duplicate` | `6cb5e8076cd8d835f67635d88e22b2c51599e42a6e60fb6fddf78973ae5b4a1b` |
| R2-F034 | `impact_graph-02` | medium | `workflow_contract` | `.agents/agentic-delivery/agents/implementation/issue-first-implementation-agent.agent.yaml:37-44, 99-116` | R1-K | `duplicate` | `6cb5e8076cd8d835f67635d88e22b2c51599e42a6e60fb6fddf78973ae5b4a1b` |
| R2-F035 | `impact_graph-03` | high | `impact_graph_relation_typing` | `scripts/pm-review-system.py:696-718, 797-810` | R1-E | `duplicate` | `decbc30cfdf75207809147b9be5b20044468c472d490ee5f569081a143fc916f` |
| R2-F036 | `impact_graph-03` | high | `review_behavior_contract` | `.agents/agentic-delivery/contracts/pm-review-packet-template.md:84, 159` | R1-B | `duplicate` | `decbc30cfdf75207809147b9be5b20044468c472d490ee5f569081a143fc916f` |
| R2-F037 | `impact_graph-03` | high | `lab_evidence_binding` | `scripts/pm-review-system.py:2103-2143` | R1-B | `duplicate` | `decbc30cfdf75207809147b9be5b20044468c472d490ee5f569081a143fc916f` |
| R2-F038 | `impact_graph-04` | high | `typed_edge_semantics` | `scripts/pm-review-system.py:685-719, 731-744, 781-817` | R1-E | `duplicate` | `8773e93933f506aff2db02209239612b613e4ebf257223c6365c8b3a20cfd04a` |
| R2-F039 | `impact_graph-04` | high | `changed_impact_coverage` | `scripts/pm-review-system.py:49-53, 1050-1062, 1110-1117, 1209-1233, 1300-1305` | R1-E | `duplicate` | `8773e93933f506aff2db02209239612b613e4ebf257223c6365c8b3a20cfd04a` |
| R2-F040 | `impact_graph-04` | high | `packet_bounds` | `scripts/pm-review-system.py:1547-1576, 1630-1635, 1744-1769` | R1-D | `duplicate` | `8773e93933f506aff2db02209239612b613e4ebf257223c6365c8b3a20cfd04a` |
| R2-F041 | `impact_graph-05` | high | `impact_relation_typing` | `scripts/pm-review-system.py:696-719` | R1-E | `duplicate` | `769cce9b9dca803884e242cf66c3de5fa6122244ac0f8d2bbf58517e46b33ff3` |
| R2-F042 | `impact_graph-05` | high | `impact_reference_recall` | `scripts/pm-review-system.py:51-54, 506-526, 806-819` | R1-O | `accepted_with_modification` | `769cce9b9dca803884e242cf66c3de5fa6122244ac0f8d2bbf58517e46b33ff3` |
| R2-F043 | `impact_graph-05` | medium | `impact_certainty` | `scripts/pm-review-system.py:685-693` | R1-E | `duplicate` | `769cce9b9dca803884e242cf66c3de5fa6122244ac0f8d2bbf58517e46b33ff3` |
| R2-F044 | `impact_graph-05` | high | `packet_bounds` | `scripts/pm-review-system.py:1547-1576` | R1-D | `duplicate` | `769cce9b9dca803884e242cf66c3de5fa6122244ac0f8d2bbf58517e46b33ff3` |
| R2-F045 | `impact_graph-05` | high | `exact_blob_provenance` | `scripts/pm-review-system.py:1458-1508` | R1-A | `duplicate` | `769cce9b9dca803884e242cf66c3de5fa6122244ac0f8d2bbf58517e46b33ff3` |
| R2-F046 | `impact_graph-05` | medium | `resource_bounds` | `scripts/pm-review-system.py:850-956, 1234-1241` | R1-F | `accepted_with_modification` | `769cce9b9dca803884e242cf66c3de5fa6122244ac0f8d2bbf58517e46b33ff3` |
| R2-F047 | `impact_graph-05` | high | `synthesis_manifest_integrity` | `scripts/pm-review-system.py:2197-2234, 2237-2329` | R1-A | `duplicate` | `769cce9b9dca803884e242cf66c3de5fa6122244ac0f8d2bbf58517e46b33ff3` |
| R2-F048 | `impact_graph-05` | high | `lab_evidence_binding` | `scripts/pm-review-system.py:2108-2145` | R1-B | `duplicate` | `769cce9b9dca803884e242cf66c3de5fa6122244ac0f8d2bbf58517e46b33ff3` |
| R2-F049 | `impact_graph-06` | high | `impact_relation_misclassification` | `.agents/agentic-delivery/workflows/codex-active-orchestration-loop.md:9-25` | R1-E | `duplicate` | `2e325cbc691b5dc58197c916bad5278fc2a888b99c4ea60908f4f10b1ceb393f` |
| R2-F050 | `impact_graph-06` | high | `phase_state_validation` | `.agents/agentic-delivery/schemas/pm-review-system-phase-state.schema.json:6-39` | R1-L | `duplicate` | `2e325cbc691b5dc58197c916bad5278fc2a888b99c4ea60908f4f10b1ceb393f` |
| R2-F051 | `impact_graph-06` | high | `packet_bound_underestimation` | `scripts/pm-review-system.py:1547-1576` | R1-D | `duplicate` | `2e325cbc691b5dc58197c916bad5278fc2a888b99c4ea60908f4f10b1ceb393f` |
| R2-F052 | `impact_graph-06` | medium | `slice_revision_ambiguity` | `scripts/pm-review-system.py:1458-1513` | R1-A | `duplicate` | `2e325cbc691b5dc58197c916bad5278fc2a888b99c4ea60908f4f10b1ceb393f` |
| R2-F053 | `impact_graph-06` | medium | `workflow_contract_parity` | `.agents/agentic-delivery/workflows/codex-active-orchestration-loop.md:48-63` | R2-N2 | `duplicate` | `2e325cbc691b5dc58197c916bad5278fc2a888b99c4ea60908f4f10b1ceb393f` |
| R2-F054 | `impact_graph-07` | high | `packet_overflow_fail_open` | `scripts/pm-review-system.py:1488-1513, 1547-1576, 1630-1635` | R1-D | `duplicate` | `f0a61ef8f1123618cddd61afce993126323ef44f3898fb037ccda8ed424f37fb` |
| R2-F055 | `impact_graph-08` | high | `packet_budget_underestimation` | `scripts/pm-review-system.py:1547-1564, 1632-1635` | R1-D | `duplicate` | `d5ec051388072e716f2709ea444d7598bfc206ed0883e415477920022ab00b7e` |
| R2-F056 | `impact_graph-08` | medium | `impact_slice_bound_bypass` | `scripts/pm-review-system.py:1488-1503, 1789-1790` | R1-D | `duplicate` | `d5ec051388072e716f2709ea444d7598bfc206ed0883e415477920022ab00b7e` |
| R2-F057 | `impact_graph-08` | medium | `shepherd_gate_bypass` | `scripts/pi-auto-loop.sh:121-134` | R2-N3 | `duplicate` | `d5ec051388072e716f2709ea444d7598bfc206ed0883e415477920022ab00b7e` |
| R2-F058 | `impact_graph-09` | high | `impact_graph_completeness` | `scripts/pm-review-system.py:1196` | R1-G | `duplicate` | `3396528d539eeb427c421c10c8d52fc6a3a755b462c9d0b73e2f01f18d866150` |
| R2-F059 | `impact_graph-09` | high | `typed_edge_classification` | `scripts/pm-review-system.py:811` | R1-E | `duplicate` | `3396528d539eeb427c421c10c8d52fc6a3a755b462c9d0b73e2f01f18d866150` |
| R2-F060 | `impact_graph-09` | medium | `workflow_documentation_accuracy` | `.agents/connector-migration/ownership-rules.md:20` | R2-N5 | `accepted_with_modification` | `3396528d539eeb427c421c10c8d52fc6a3a755b462c9d0b73e2f01f18d866150` |
| R2-F061 | `impact_graph-10` | high | `impact_coverage` | `scripts/pm-review-system.py:1115-1117` | R1-E | `duplicate` | `266122c9aa814a7731e6d9e50556f6cf86690f64525293fba1de8bf79bade151` |
| R2-F062 | `impact_graph-10` | high | `packet_bounds` | `scripts/pm-review-system.py:1548-1552, 1563-1574, 1632-1633` | R1-D | `duplicate` | `266122c9aa814a7731e6d9e50556f6cf86690f64525293fba1de8bf79bade151` |
| R2-F063 | `impact_graph-10` | medium | `exact_version_binding` | `scripts/pm-review-system.py:1472-1482` | R1-A | `duplicate` | `266122c9aa814a7731e6d9e50556f6cf86690f64525293fba1de8bf79bade151` |
| R2-F064 | `impact_graph-10` | low | `documentation_accuracy` | `.pi/README.md:154-156` | R2-N2 | `duplicate` | `266122c9aa814a7731e6d9e50556f6cf86690f64525293fba1de8bf79bade151` |
| R2-F065 | `impact_graph-11` | high | `impact_graph_completeness` | `scripts/pm-review-system.py:51-53` | R1-O | `duplicate` | `75607b8bb4a6f83233a4a956317888588f3bbe819443996e656c1b63f88c60ae` |
| R2-F066 | `impact_graph-11` | high | `typed_edge_relation` | `scripts/pm-review-system.py:696-719` | R1-E | `duplicate` | `75607b8bb4a6f83233a4a956317888588f3bbe819443996e656c1b63f88c60ae` |
| R2-F067 | `impact_graph-11` | high | `packet_context_bounds` | `scripts/pm-review-system.py:1547-1552, 1631-1633` | R1-D | `duplicate` | `75607b8bb4a6f83233a4a956317888588f3bbe819443996e656c1b63f88c60ae` |
| R2-F068 | `impact_graph-12` | high | `packet_bounds` | `scripts/pm-review-system.py:1547-1576, 1632-1636` | R1-D | `duplicate` | `2d72be8d535db82b055989c62d3671e900af5270d194c6256e5fb481410ccef4` |
| R2-F069 | `impact_graph-12` | high | `typed_edge_provenance` | `scripts/pm-review-system.py:662-680, 1165-1183, 1626-1627, 1729-1739, 2267-2280` | R1-D | `duplicate` | `2d72be8d535db82b055989c62d3671e900af5270d194c6256e5fb481410ccef4` |
| R2-F070 | `impact_graph-13` | high | `impact_graph_recall` | `scripts/pm-review-system.py:51-53, 476-478` | R1-O | `duplicate` | `986f41462b6e18465d78397deab546b307cc1e0648b149decad38e121ee42d70` |
| R2-F071 | `impact_graph-13` | medium | `typed_edge_semantics` | `.pi/prompts/pm-connector-loop.md:14-15, 29-32` | R1-E | `duplicate` | `986f41462b6e18465d78397deab546b307cc1e0648b149decad38e121ee42d70` |
| R2-F072 | `impact_graph-13` | high | `packet_bounding` | `scripts/pm-review-system.py:1547-1576, 1631-1635` | R1-D | `duplicate` | `986f41462b6e18465d78397deab546b307cc1e0648b149decad38e121ee42d70` |
| R2-F073 | `impact_graph-13` | high | `workflow_contract` | `.agents/connector-migration/agents/implementation/passb-expander.agent.yaml:46-51, 61, 94` | R2-N5 | `duplicate` | `986f41462b6e18465d78397deab546b307cc1e0648b149decad38e121ee42d70` |
| R2-F074 | `impact_graph-14` | high | `impact_graph_correctness` | `scripts/pm-review-system.py:1035, 1115-1139, 1209-1232` | R1-E | `duplicate` | `5cf91f3b98d174989b575d12f0dd346e06b4bca2d90e2d1c0c79e6985e8a9b4d` |
| R2-F075 | `impact_graph-14` | high | `packet_bounding` | `scripts/pm-review-system.py:1548-1576, 1604, 1633, 1683, 1777-1789` | R1-D | `duplicate` | `5cf91f3b98d174989b575d12f0dd346e06b4bca2d90e2d1c0c79e6985e8a9b4d` |
| R2-F076 | `impact_graph-15` | high | `impact_graph_correctness` | `scripts/pm-review-system.py:696-719` | R1-E | `duplicate` | `ac0d012267a7fa7cdeee826bf59aff58e853b9cc19a6cfd731fea9ee8d60971e` |
| R2-F077 | `impact_graph-15` | high | `packet_bounds` | `scripts/pm-review-system.py:1547-1576` | R1-D | `duplicate` | `ac0d012267a7fa7cdeee826bf59aff58e853b9cc19a6cfd731fea9ee8d60971e` |
| R2-F078 | `impact_graph-15` | medium | `packet_bounds` | `scripts/pm-review-system.py:1484-1512` | R1-D | `duplicate` | `ac0d012267a7fa7cdeee826bf59aff58e853b9cc19a6cfd731fea9ee8d60971e` |
| R2-F079 | `impact_graph-16` | medium | `packet_bound_safety` | `scripts/pm-review-system.py:1547-1576` | R1-D | `duplicate` | `5d38d2f64704ab5584feca7a67f29e9d9107dc017f8bb2b56f2a5d0c480d5fe0` |
| R2-F080 | `impact_graph-17` | high | `packet_bounds` | `scripts/pm-review-system.py:1547-1575, 1630-1635` | R1-D | `duplicate` | `41070fda56b012066361e62dc4d470ee8e1b474508b6243d77c20948615401d2` |
| R2-F081 | `impact_graph-17` | high | `lab_evidence_binding` | `scripts/pm-review-system.py:2108-2145` | R1-B | `duplicate` | `41070fda56b012066361e62dc4d470ee8e1b474508b6243d77c20948615401d2` |
| R2-F082 | `impact_graph-18` | high | `packet_context_bounds` | `scripts/pm-review-system.py:1547-1577` | R1-D | `duplicate` | `a48e877e7a1483c5d0f3e8646a48266e1662366cda59f8af6331460f70981e4c` |
| R2-F083 | `impact_graph-18` | medium | `impact_graph_certainty` | `.planning/phases/397-pm-first-round-review-system-r1/REVIEW-R1-DISPOSITION.md:75-142` | R1-E | `duplicate` | `a48e877e7a1483c5d0f3e8646a48266e1662366cda59f8af6331460f70981e4c` |
| R2-F084 | `impact_graph-18` | high | `packet_slice_bound` | `scripts/pm-review-system.py:1484-1515` | R1-D | `duplicate` | `a48e877e7a1483c5d0f3e8646a48266e1662366cda59f8af6331460f70981e4c` |
| R2-F085 | `impact_graph-19` | high | `lab_information_disclosure` | `scripts/pm-review-lab.py:152-172` | R1-H | `duplicate` | `06f219f9d8c373fa9dd022446a5dc1c2efefaee41837d6799a3f7b2eb307bd12` |
| R2-F086 | `impact_graph-19` | high | `lab_resource_bounds` | `scripts/pm-review-lab.py:354-355, 369-419` | R1-H | `duplicate` | `06f219f9d8c373fa9dd022446a5dc1c2efefaee41837d6799a3f7b2eb307bd12` |
| R2-F087 | `impact_graph-19` | high | `lab_evidence_integrity` | `scripts/pm-review-lab.py:531-551` | R1-H | `duplicate` | `06f219f9d8c373fa9dd022446a5dc1c2efefaee41837d6799a3f7b2eb307bd12` |
| R2-F088 | `impact_graph-19` | high | `lab_discriminator_integrity` | `scripts/pm-review-lab.py:159-161, 221-229, 543-547` | R1-H | `duplicate` | `06f219f9d8c373fa9dd022446a5dc1c2efefaee41837d6799a3f7b2eb307bd12` |
| R2-F089 | `impact_graph-19` | high | `synthesis_lab_binding` | `scripts/pm-review-system.py:2127-2143` | R1-B | `duplicate` | `06f219f9d8c373fa9dd022446a5dc1c2efefaee41837d6799a3f7b2eb307bd12` |
| R2-F090 | `impact_graph-19` | medium | `synthesis_fail_closed` | `scripts/pm-review-system.py:2237-2261, 2331` | R1-B | `duplicate` | `06f219f9d8c373fa9dd022446a5dc1c2efefaee41837d6799a3f7b2eb307bd12` |
| R2-F091 | `impact_graph-19` | medium | `response_shape_validation` | `scripts/pm-review-system.py:2152-2195` | R1-B | `duplicate` | `06f219f9d8c373fa9dd022446a5dc1c2efefaee41837d6799a3f7b2eb307bd12` |
| R2-F092 | `impact_graph-20` | high | `packet_bounding` | `scripts/pm-review-system.py:1548-1576, 1632-1635` | R1-D | `duplicate` | `91f35af971620041a3b34605a915af1d9f883f25e32198d45939c7993755dccb` |
| R2-F093 | `impact_graph-20` | medium | `typed_edge_certainty` | `scripts/pm-review-system.py:731-784` | R1-E | `duplicate` | `91f35af971620041a3b34605a915af1d9f883f25e32198d45939c7993755dccb` |
| R2-F094 | `impact_graph-20` | high | `exact_version_binding` | `scripts/pm-review-system.py:1911-1967` | R1-A | `duplicate` | `91f35af971620041a3b34605a915af1d9f883f25e32198d45939c7993755dccb` |
| R2-F095 | `impact_graph-20` | medium | `graph_resource_bounds` | `scripts/pm-review-system.py:1064-1067, 1197-1244` | R1-F | `duplicate` | `91f35af971620041a3b34605a915af1d9f883f25e32198d45939c7993755dccb` |
| R2-F096 | `impact_graph-21` | high | `fail_closed_synthesis` | `scripts/pm-review-system.py:2199-2234, 2252-2280` | R1-A | `duplicate` | `e5d421f967c155859486aab347c8dc94651119f35f65ca907c369986b06058a6` |
| R2-F097 | `impact_graph-21` | high | `packet_bounds` | `scripts/pm-review-system.py:1547-1572, 1630-1633` | R1-D | `duplicate` | `e5d421f967c155859486aab347c8dc94651119f35f65ca907c369986b06058a6` |
| R2-F098 | `impact_graph-21` | medium | `impact_parser` | `scripts/pm-review-system.py:772-828` | R1-E | `duplicate` | `e5d421f967c155859486aab347c8dc94651119f35f65ca907c369986b06058a6` |
| R2-F099 | `impact_graph-21` | medium | `authoritative_state_consistency` | `scripts/pm-review-system.py:638-658` | R1-L | `duplicate` | `e5d421f967c155859486aab347c8dc94651119f35f65ca907c369986b06058a6` |
| R2-F100 | `impact_graph-22` | medium | `impact_graph_certainty` | `scripts/pm-review-system.py:691` | R1-E | `duplicate` | `c696cef55ca9bd7be679aa1dc11f79748b3d7519aa56e41c0cff591d85d86303` |
| R2-F101 | `impact_graph-23` | high | `impact_graph_completeness` | `scripts/pm-review-system.py:51-53` | R1-O | `duplicate` | `3f957767a9843ba13622b3a7549dc984850fd99d8ac4d623bc2db025d9057b00` |
| R2-F102 | `impact_graph-23` | high | `impact_edge_typing` | `scripts/pm-review-system.py:696-719` | R1-E | `duplicate` | `3f957767a9843ba13622b3a7549dc984850fd99d8ac4d623bc2db025d9057b00` |
| R2-F103 | `impact_graph-23` | medium | `workflow_command_correctness` | `.planning/traces/cli-architecture-v2-pi-prompts.md:21, 63, 103` | R1-K | `duplicate` | `3f957767a9843ba13622b3a7549dc984850fd99d8ac4d623bc2db025d9057b00` |
| R2-F104 | `impact_graph-23` | medium | `review_route_consistency` | `.planning/traces/cli-architecture-v2-pi-prompts.md:158-159` | R1-K | `duplicate` | `3f957767a9843ba13622b3a7549dc984850fd99d8ac4d623bc2db025d9057b00` |
| R2-F105 | `impact_graph-24` | medium | `impact_relation_classification` | `scripts/pm-review-system.py:696-719` | R1-E | `duplicate` | `efc815b2707ad61f0199ea6160de9791209a3216043a6b3e16887e71d0b5b50a` |
| R2-F106 | `impact_graph-24` | high | `packet_overflow` | `scripts/pm-review-system.py:1547-1576` | R1-D | `duplicate` | `efc815b2707ad61f0199ea6160de9791209a3216043a6b3e16887e71d0b5b50a` |
| R2-F107 | `impact_graph-24` | high | `synthesis_coverage` | `scripts/pm-review-system.py:2199-2234` | R1-A | `duplicate` | `efc815b2707ad61f0199ea6160de9791209a3216043a6b3e16887e71d0b5b50a` |
| R2-F108 | `impact_graph-24` | high | `lab_evidence_binding` | `scripts/pm-review-system.py:2127-2143` | R1-B | `duplicate` | `efc815b2707ad61f0199ea6160de9791209a3216043a6b3e16887e71d0b5b50a` |
| R2-F109 | `impact_graph-24` | medium | `hypothesis_evidence` | `scripts/pm-review-system.py:2077-2079` | R1-B | `duplicate` | `efc815b2707ad61f0199ea6160de9791209a3216043a6b3e16887e71d0b5b50a` |
| R2-F110 | `impact_graph-24` | medium | `exact_blob_provenance` | `scripts/pm-review-system.py:1472-1509` | R1-A | `duplicate` | `efc815b2707ad61f0199ea6160de9791209a3216043a6b3e16887e71d0b5b50a` |
| R2-F111 | `impact_graph-25` | high | `impact_graph_undercoverage` | `scripts/pm-review-system.py:746-754` | R1-E | `duplicate` | `805fdc3734d10604cfefb31704944c4d8bbdcf6e7cffc525f66a13664c294500` |
| R2-F112 | `impact_graph-26` | high | `packet_bound_underestimation` | `scripts/pm-review-system.py:1547-1572, 1629-1637` | R1-D | `duplicate` | `53c2772de27ebd899746c35faeaeeb22cc3ae4aa245ccf73ee6b87e78fc752b1` |
| R2-F113 | `impact_graph-26` | high | `synthesis_coverage_integrity` | `scripts/pm-review-system.py:2199-2234, 2250-2317` | R1-A | `duplicate` | `53c2772de27ebd899746c35faeaeeb22cc3ae4aa245ccf73ee6b87e78fc752b1` |
| R2-F114 | `impact_graph-26` | medium | `typed_certainty_validation` | `scripts/pm-review-system.py:654-676, 1174-1193, 1329` | R1-E | `duplicate` | `53c2772de27ebd899746c35faeaeeb22cc3ae4aa245ccf73ee6b87e78fc752b1` |
| R2-F115 | `impact_graph-26` | medium | `impact_detector_endpoint_validation` | `scripts/pm-review-system.py:218-260` | R1-E | `duplicate` | `53c2772de27ebd899746c35faeaeeb22cc3ae4aa245ccf73ee6b87e78fc752b1` |
| R2-F116 | `impact_graph-27` | high | `workflow_gate` | `.planning/phases/397-cli-architecture-v2-orchestration/PLAN.md:15, 40-57, 72-76` | R1-K | `duplicate` | `3271d0ee5374a2c94d84a78cb6355298d707003785236a6b580c5418ff3fab9c` |
| R2-F117 | `impact_graph-27` | high | `verification_gate` | `.planning/phases/397-cli-architecture-v2-orchestration/VERIFICATION.md:145, 167-173` | R1-K | `duplicate` | `3271d0ee5374a2c94d84a78cb6355298d707003785236a6b580c5418ff3fab9c` |
| R2-F118 | `impact_graph-27` | medium | `evidence_truthfulness` | `.planning/phases/397-cli-architecture-v2-orchestration/VERIFICATION.md:134` | R1-L | `duplicate` | `3271d0ee5374a2c94d84a78cb6355298d707003785236a6b580c5418ff3fab9c` |
| R2-F119 | `impact_graph-28` | medium | `evidence_truthfulness` | `.planning/phases/397-pm-first-round-review-system-r1/TDD-LEDGER.md:3, 61, 97` | R1-M | `duplicate` | `cac65bf5ff7063a674546872b3ef64cffc7a78dd75c9f958e3981bdef0595588` |
| R2-F120 | `impact_graph-29` | medium | `machine_contract_one_way_migration` | `.planning/traces/cli-architecture-v2-orchestration-state.yaml:1-45, 408-819` | R1-L | `duplicate` | `585fd2f1866015eefbe9a0ed47ffc2744851b73591cbaf3deb6a5740c732a0a6` |
| R2-F121 | `impact_graph-29` | medium | `authoritative_state_consistency` | `.planning/traces/cli-architecture-v2-orchestration-state.yaml:54-59, 479-488, 539-548, 831` | R1-L | `duplicate` | `585fd2f1866015eefbe9a0ed47ffc2744851b73591cbaf3deb6a5740c732a0a6` |
| R2-F122 | `impact_graph-29` | high | `fail_closed_authority_inventory` | `scripts/pm-review-system.py:590-638` | R1-L | `duplicate` | `585fd2f1866015eefbe9a0ed47ffc2744851b73591cbaf3deb6a5740c732a0a6` |
| R2-F123 | `impact_graph-30` | high | `packet_bounds` | `scripts/pm-review-system.py:1548-1576` | R1-D | `duplicate` | `53b04aa6edc80cb992378a59dd723bf1d9044b30f9cfd4d280b9d851fb4a53fa` |
| R2-F124 | `impact_graph-30` | medium | `packet_slice_bounds` | `scripts/pm-review-system.py:1489-1512` | R1-D | `duplicate` | `53b04aa6edc80cb992378a59dd723bf1d9044b30f9cfd4d280b9d851fb4a53fa` |
| R2-F125 | `impact_graph-31` | high | `exact_version_binding` | `scripts/pm-review-system.py:2199-2234, 2240-2303` | R1-A | `duplicate` | `923424e4c941188cf56feefc7314bd56f925a9b1c375ba47b6c2143a05a0b7d9` |
| R2-F126 | `impact_graph-31` | high | `impact_graph_correctness` | `scripts/pm-review-system.py:1034-1061, 1110-1117, 1198-1234` | R1-E | `duplicate` | `923424e4c941188cf56feefc7314bd56f925a9b1c375ba47b6c2143a05a0b7d9` |
| R2-F127 | `impact_graph-31` | high | `packet_bounds` | `scripts/pm-review-system.py:1545-1577, 1629-1637, 1741-1797` | R1-D | `duplicate` | `923424e4c941188cf56feefc7314bd56f925a9b1c375ba47b6c2143a05a0b7d9` |
| R2-F128 | `impact_graph-31` | medium | `packet_slice_bound` | `scripts/pm-review-system.py:1484-1507` | R1-D | `duplicate` | `923424e4c941188cf56feefc7314bd56f925a9b1c375ba47b6c2143a05a0b7d9` |
| R2-F129 | `impact_graph-31` | high | `exact_blob_binding` | `scripts/pm-review-system.py:1470-1482, 2167-2176` | R1-A | `duplicate` | `923424e4c941188cf56feefc7314bd56f925a9b1c375ba47b6c2143a05a0b7d9` |
| R2-F130 | `impact_graph-31` | medium | `secret_handling` | `scripts/loop-trace.sh:120-139, 157-172, 208-212, 258-289` | R2-N4 | `duplicate` | `923424e4c941188cf56feefc7314bd56f925a9b1c375ba47b6c2143a05a0b7d9` |
| R2-F131 | `impact_graph-31` | medium | `workflow_liveness` | `scripts/pi-shepherd-loop.sh:345-352, 376-400` | R2-N3 | `duplicate` | `923424e4c941188cf56feefc7314bd56f925a9b1c375ba47b6c2143a05a0b7d9` |
| R2-F132 | `impact_graph-31` | medium | `connector_operation_ledger` | `internal/connectors/defs/github/api_surface.json:3713-3767` | R2-N6 | `needs_human` | `923424e4c941188cf56feefc7314bd56f925a9b1c375ba47b6c2143a05a0b7d9` |
| R2-F133 | `impact_graph-32` | high | `packet_bounds` | `scripts/pm-review-system.py:1547-1575, 1774-1790` | R1-D | `duplicate` | `3f89926e07b0c2313c7fa0d5c0786112906b7ef02831ed7c0674d10b3af6fe94` |
| R2-F134 | `impact_graph-32` | high | `packet_bounds` | `scripts/pm-review-system.py:1563-1565, 1634-1684` | R1-D | `duplicate` | `3f89926e07b0c2313c7fa0d5c0786112906b7ef02831ed7c0674d10b3af6fe94` |
| R2-F135 | `impact_graph-32` | medium | `impact_graph_correctness` | `scripts/pm-review-system.py:829-838` | R1-E | `duplicate` | `3f89926e07b0c2313c7fa0d5c0786112906b7ef02831ed7c0674d10b3af6fe94` |
| R2-F136 | `impact_graph-33` | high | `synthesis_integrity` | `scripts/tests/pm-review-system.sh:955-1140` | R1-A | `duplicate` | `468adeb2df536ba262270dc350da8670ab2b4f79a19fbd5eb761ccc4d217c0b5` |
| R2-F137 | `impact_graph-33` | high | `packet_budget_integrity` | `scripts/tests/pm-review-system.sh:636-642, 909-920` | R1-D | `duplicate` | `468adeb2df536ba262270dc350da8670ab2b4f79a19fbd5eb761ccc4d217c0b5` |
| R2-F138 | `impact_graph-33` | high | `lab_evidence_binding` | `scripts/tests/pm-review-system.sh:1158-1172` | R1-B | `duplicate` | `468adeb2df536ba262270dc350da8670ab2b4f79a19fbd5eb761ccc4d217c0b5` |
| R2-F139 | `impact_graph-33` | medium | `impact_edge_provenance` | `scripts/tests/pm-review-system.sh:521-528, 623-642` | R1-E | `duplicate` | `468adeb2df536ba262270dc350da8670ab2b4f79a19fbd5eb761ccc4d217c0b5` |
| R2-F140 | `impact_graph-33` | medium | `impact_graph_test_coverage` | `scripts/tests/pm-review-system.sh:623-642` | R1-C | `accepted_with_modification` | `468adeb2df536ba262270dc350da8670ab2b4f79a19fbd5eb761ccc4d217c0b5` |
| R2-F141 | `impact_graph-33` | medium | `lab_network_test` | `scripts/tests/pm-review-system.sh:477-479, 797-810` | R1-H | `duplicate` | `468adeb2df536ba262270dc350da8670ab2b4f79a19fbd5eb761ccc4d217c0b5` |

## Gate state

- `local_codex.status`: `findings_correction_required`
- `shepherd.status`: `pending`
- `correction_rounds`: `2/5`
- `human_merge_authority`: Firstmate-only after all named gates; this agent must not merge.
