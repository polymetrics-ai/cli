# TDD LEDGER — issue #3584 HubSpot / Bitbucket forward remediation

## Required skills recorded

Loaded skills:

- `gsd-core`
- `no-mistakes`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-documentation`
- `golang-lint`

Rule hooks for this slice:

- `gsd-core` implementation workflow steps 1-9.
- `golang-testing` rules 1, 3, 5.
- `golang-error-handling` rules 1, 2, 7.
- `golang-security` security thinking model 1-3 and no-secret common mistake.
- `golang-safety` rules 2, 4, 6, 7.
- `golang-design-patterns` rules 5, 9, 20, 21.
- `golang-structs-interfaces` small consumer-owned interface / accept-interface-return-struct guidance.
- `golang-documentation` writing principles: concise, intent, no invented context, preserve obligations.
- `golang-lint` suppression rules 1-4 and workflow rules 1-3.

## GSD evidence

```bash
scripts/gsd doctor
```

Result: pass.

```bash
scripts/gsd prompt programming-loop init --phase connector-guardrail-remediation-r1 --dry-run
```

Result: failed as expected for this repo-local adapter state; exact output:

```text
scripts/gsd: unknown GSD command: programming-loop
```

Fallback: manual GSD universal runtime loop with `scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run` trace.

```bash
scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run
```

Result: prompt rendered successfully.

## Red / green log

| Time | Slice | Command | Result | Notes |
| --- | --- | --- | --- | --- |
| 2026-08-02 | A ledger completeness | `go test ./.planning/phases/connector-guardrail-remediation-r1/workers/issue-3584` | red | Failed before ledger existed: `read remediation-ledger.json: open remediation-ledger.json: no such file or directory`; exit 1. |
| 2026-08-02 | A ledger completeness | `go test ./.planning/phases/connector-guardrail-remediation-r1/workers/issue-3584` | green | Passed after adding `remediation-ledger.json` with all #3529/#3531 shared path dispositions. |
| 2026-08-02 | B preservation/correction | `go test ./internal/cli ./internal/app ./internal/connectors/bundleregistry` | infra red then green | First run failed from local disk exhaustion (`no space left on device`), not code. After `go clean -cache -testcache`, rerun passed. Existing Bitbucket reverse ETL and bundle registry tests serve as preservation evidence; no forward code correction needed. |
| 2026-08-02 | B guard/foundation packages | `go test ./internal/connectors/boundary ./cmd/connectorgen` | green | Passed; boundary/connectorgen remain unchanged in this slice. |
| 2026-08-02 | C broad verification | `gofmt -w cmd internal`; `go vet ./...`; `go build ./cmd/pm`; `make verify` | green | All broader local gates passed. |

## Acceptance coverage checklist

- [x] PR #3529 URL and merge SHA recorded.
- [x] PR #3531 URL and merge SHA recorded.
- [x] Every #3529 shared path listed in audit report has path classification and disposition.
- [x] Every #3531 shared path listed in audit report has path classification and disposition.
- [x] Each shared path records connector-lane verdict (`fail_shared_runtime`, `fail_foundation_test_outside_target`, `fail_shared_app_runtime`, `fail_shared_manifest_scope`, or `defer_engine_scope_to_issue_3585`) instead of pass.
- [x] Each preserved foundation cites existing/focused evidence.
- [x] Each harmful/unrelated correction, if any, uses forward commit only and has red test first; no harmful forward code correction found in this slice.
- [x] Bitbucket engine-path code corrections, if needed, are deferred to #3585 instead of edited here.
- [x] No history rewrite, no force push, no blanket revert.
