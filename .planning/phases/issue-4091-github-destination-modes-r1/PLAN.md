---
phase: issue-4091-github-destination-modes-r1
plan: "01"
type: tdd
status: complete
base: integration/4015-mvp-flat-r1
requirements:
  - ISSUE-4091
required_skills:
  - golang-how-to
  - golang-design-patterns
  - golang-structs-interfaces
  - golang-error-handling
  - golang-security
  - golang-safety
  - golang-testing
  - gsd-discuss-phase
  - gsd-plan-phase
  - gsd-execute-phase
  - gsd-verify-work
  - gsd-code-review
files_modified:
  - internal/app/authorization.go
  - internal/app/issue_label_transport_approval.go
  - internal/app/issue_label_warehouse_transport.go
  - internal/app/transport_dispatch.go
  - internal/app/github_warehouse_transport_approval_test.go
  - internal/connectors/definition.go
  - internal/connectors/defs/github/writes.json
  - internal/connectors/engine/schema/writes.schema.json
  - internal/cli/testdata/golden_transcripts.json
---

# TDD plan: GitHub destination modes

## Acceptance evidence

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Safe additive remains the default | fake | A deterministic recording GitHub provider is required because safe test fixtures cannot perform a real GitHub write. It records the additive request and read-back labels. |
| Set-replace requires explicit per-connection consent | fake | The provider recorder must prove the security property as zero sends before an untrusted external write; it also proves an enabled scope sends exactly the intended replacement. |
| Keyed mode requires explicit per-connection consent | fake | The provider recorder safely proves zero writes while the destination configuration is disabled and proves keyed label targeting/read-back when authorized. |
| Approval token is single use | fake | The project-state test uses an in-process provider recorder so the replay has an observable zero-send result instead of risking a real duplicate mutation. |
| Later identical scope can run unattended | fake | A persisted authorization and recording provider demonstrate a second run without a token reaches the expected provider mutation and read-back. |
| Scope drift/revocation/expiry fails closed | fake | The provider recorder asserts zero sends after every mismatch; these are preflight contracts unsuitable for credentialed live tests. |
| Definition owns path/action/mode facts | live | Bundle loading and generated-surface checks read the actual GitHub definition; a missing/hand-authored fact fails validation or surface synchronization. |
| PR base is exact | live | API read-back must print `integration/4015-mvp-flat-r1`. |

## Slices and checkpoints

1. **Plan checkpoint** — record the issue decisions, the #4132 foundation, required skills, TDD red/green tests, scope guard, and delivery base. Commit and push planning evidence before production edits.
2. **Red** — add GitHub destination-mode tests that execute a non-additive request against a recorder with opt-in absent/disabled and changed scope. Assert zero sends; verify they fail before implementation.
3. **Green** — add definition-owned destination-mode declarations and connection configuration. Reuse #4132 scope identity in the GitHub reverse path; mint standing authorization on its one-time proceed, resolve it per later run, and fail before provider sends when absent, disabled, revoked, expired, or changed.
4. **Proof/refactor** — prove additive, set-replace, and keyed outcomes by reading labels back from the recorder. Run targeted and specified package tests, format, vet, static/generator checks, review, then open the explicit-base PR. **Complete:** the recorder observes exact set replacement plus zero PUTs for replay, revoked authorization, scope drift, and disabled per-connection consent.

## Safety boundaries

- The new write capability stays GitHub-specific and closed to declared issue-label operations; no generic HTTP/SQL/shell write surface is introduced.
- Tokens, credentials, raw labels, and approval values are not added to durable evidence or logs.
- Every rejection is before the provider request; error-only assertions are insufficient.
- Keep all changed paths within the GitHub hook/definition and the already-landed app authorization integration seam. A finding in unrelated code is a `needs-decision` follow-up, not a branch expansion.
