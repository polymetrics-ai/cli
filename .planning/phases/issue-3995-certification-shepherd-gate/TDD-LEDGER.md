# TDD LEDGER — issue #3995 shared connector-certification Shepherd gate

| ID | Enforcement | RED evidence | GREEN evidence | Refactor/verification |
| --- | --- | --- | --- | --- |
| R1 | A GitHub `integrate_sub_pr` gate must expose the exact missing live-evidence cell and an all-green generated fixture must proceed. | Pending: add the executable evaluator test before any production evaluator source; it will fail because current `agentcontract` has no certification-gate evaluator/verdict API. | Pending. | Pending: verdict ordering and identifiers are deterministic. |
| R2 | Every applicable capability, workflow, sync-mode primitive, and flow pair needs declared, implemented, fixture-tested, live-tested, and validated live evidence. | Pending: one isolated fixture defect per criterion, plus implemented/reachable/file-only cases. | Pending. | Pending: IDs preserve the failing cell/evidence coordinate. |
| R3 | Unknown/missing schemas, unknown fields, invalid evidence, and omitted gate fields fail closed rather than silently passing. | Pending: strict decoder and projection mutation tests. | Pending. | Pending: no schema field is inferred from a provider or adapter-local default. |
| R4 | Claude, Codex, Pi, and OpenCode receive equivalent generated gate input/verdict instructions, and drift/missing registered projections fail. | Pending: require OpenCode and mutate/delete a generated target. | Pending. | Pending: shared renderer is the only gate-instruction source. |
| R5 | The current zero-certified GitHub baseline rejects at a transition while generic contract validation stays structural and read-only. | Pending: current-artifact integration test. | Pending. | Pending: evaluator creates no evidence and cannot invoke provider actions. |

No production evaluator, generator, canonical-source, or generated-projection edit is permitted
before the recorded R1 RED run.
