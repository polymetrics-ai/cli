# Issue 4166 Discussion Log

## Workflow

- Generated command: `scripts/gsd prompt discuss-phase issue-4166-validation-proof-gaps-r1`
- Execution mode: inline/manual fallback; repository rules prohibit spawning planner, reviewer, verifier, or orchestrator roles.
- Inputs: issue #4166 in full, the captain's binding Gap 3 flow addition posted to the issue, repository delivery contracts, and current implementation call sites.

## Decisions

1. Treat this as one validation-only PR for one primary issue. No production feature or provider defect is repaired here.
2. Use negative controls for Gaps 1 and 2. Each proof must first demonstrate that the deliberately broken or unregistered condition produces a failed certification terminal result naming the affected action/transport; the intact control then passes.
3. Count an operation as exercised only when certification reaches its real declarative validation/preparation path. Merely inventorying an action as `blocked` or `untested` does not count.
4. Keep unsafe or unpaired writes non-mutating. Their negative-control proof may validate and dry-run their declarative contract but must not send a provider request.
5. Drive Gap 3 through a flow manifest whose sync step runs the real GitHub-to-warehouse job and whose action step consumes the warehouse output and invokes the real GitHub destination action.
6. Prefer a label create/read/delete lifecycle in the disposable repository: it is definition-owned, curated in `certification.json`, provider-readable, and fully reversible. If the existing flow mapping cannot form that loop without product changes, report the exact defect instead of changing product behavior.
7. Provider setup and teardown may use bounded test harness calls, but the accepted data round trip itself must cross the declarative connector and flow production call chains in the freshly built binary.
8. Assert durable warehouse state from a new process/reopen after extraction, not from in-memory records held by the test harness.
9. Assert refusal-path zero side effects by reading the provider before and after each refusal. Typed error matching alone is insufficient.
10. Record every GitHub-specific dispatch or adapter on the observed call chain, especially the current issue-label transport selection in `internal/app`, and classify whether it prevents a definition-only proof.

## Open Execution Dependency

The required live credential environment variables are absent in this worktree session. This does not block the deterministic harness and negative-control work. It does block claiming the Gap 3 live acceptance proof; if they remain absent at delivery, the issue comment and PR will state that Gap 3 is still open and name the exact unrun test and command.
