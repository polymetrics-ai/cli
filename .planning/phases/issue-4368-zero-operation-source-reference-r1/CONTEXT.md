# Context — issue #4368 zero-operation source-reference foundation

## Task Delivery Header

- Issue: Closes #4368 — fix(connectorgen): retain zero-operation source-reference coverage. Refs #4292 — map parity batches 8, 9 and 10.
- Base branch: main (`origin/main` frozen at `cf29d302c13f7fcd340d31ad6dc27872880ccf42` before edits).
- Merges into: main.
- Delivery: A direct pull request against `main`, with the current-main base and exact branch head recorded and verified through the GitHub API after opening.
- Working branch: `fm/cli-zero-operation-source-reference-foundation-r1`.
- Task: Admit only an explicit, integrity-checked rendered coverage document with an intentionally empty operation inventory; retain its source evidence and permit the five named provider inventories to project visible source-cited deferred rows without inventing a provider contract or runnable action.
- Verification: Targeted RED/GREEN source-import, source-artifact, source-projection, operation-evidence, declaration-admission, surface-sync and commandrunner checks; exact five-cohort 720 reconciliation; JSON duplicate/invariant checks; `gofmt`, `go vet`, build/help/manual checks where applicable, and repository generator gates.

## Decisions recorded by inline discuss-phase

- The issue fixes a shared foundation, not a connector implementation lane; the five named cohorts are the bounded generated evidence consumers.
- `rendered_reference` retains its existing non-empty semantics. An empty operation array becomes valid only through a new closed, explicit coverage-only discriminator and continues to require the ordinary artifact, published-source, byte/digest, and retained-manifest proof.
- The marker is provenance-only: it must not create a descriptor operation, stream, write, action, transport, credential lookup, or provider call.
- Every source operation in the five cohorts must remain in exactly one source-accounting disposition. The requested deferred outcome uses one named `missing_foundation` per affected source row and preserves citation, provider ID, identity, and lane.
- This work does not absorb #4364, #4366, or #4367. It records those gaps only when the real registry surface requires them.

## Workflow and skills

- GSD adapter: `scripts/gsd doctor`, `scripts/gsd sources {discuss-phase,plan-phase,execute-phase,verify-work,code-review}`, `go run ./cmd/agentcontractgen check`, and generated prompts were run before planning. This runner has no compatible Pi isolated-worker runtime and the canonical single-worker contract forbids role spawning, so discuss/plan/execute/verify/review are recorded inline.
- Loaded skills: `golang-how-to`, `golang-testing`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-cli`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-lint`, `go-engineering`, and `github-issue-first-delivery`.
- The task-named `connector-lane-build-order` and `firstmate-exhaustive-review` references are absent from this isolated worktree (`find .` returned no matching path). This is a documented lookup fallback, not permission to weaken the repository contracts.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Explicit zero-operation rendered coverage is admitted | live | A strict source-lock fixture with the marker, retained artifact and manifest imports successfully; removing any element makes the same importer fail. |
| Accidental/malformed/mixed-invalid coverage fails closed | live | Table-driven strict fixtures assert a nonzero error naming the document/source location before any descriptor projection. |
| Existing source forms remain byte-stable | live | Existing rendered-reference and OpenAPI v1-v3 focused test corpus remains green with unchanged fixtures. |
| Five cohorts reconcile to 720 deferred source rows | live | A bounded exact-count test/generator invocation checks per-connector counts, unique source IDs/JSON keys, source provenance, lane and one foundation disposition. |
| Deferred command boundary is safe | live | Real registry/commandrunner preflight returns structured `missing_foundation` before a credential lookup or transport; pre-existing runnable rows still stop at missing credential. |
