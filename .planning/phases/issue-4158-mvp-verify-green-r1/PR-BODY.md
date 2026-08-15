Refs #4170

## Intent

Restore the Production MVP's fresh-binary declarative GitHub → warehouse →
GitHub flow without adding legacy inline-action authorization compatibility.

## What Changed

- Migrated #4170's acceptance fixture action to reference the already-created,
  approved reverse plan as its `job`; `action_cfg` now retains only
  `read_back_stream`.
- Preserved the externally built-binary test and all of its observable-result
  assertions.
- Added permanent pre-I/O contract tests for a missing approved job reference,
  a revoked job, and a stale job. They assert the exact `JobReferenceError`
  reason, `validation/flow_job_reference_refused`, and zero target side effects.

## Causal Evidence

### Reproduction

At `ef3c71caf`:

```bash
go test -timeout 20m ./internal/cli \
  -run '^TestFreshBinaryDeclarativeGitHubWarehouseFlowRoundTrip$' -count=2 -v
```

Both runs built a fresh 152,132,226-byte `pm` binary and exited 3 at real
`pm flow run`; the result was repeatable.

### Trigger, mask, symptom

- Trigger: #4170's action serialized legacy inline scope with no `job`.
- Mask: #4168 re-resolves every stored flow before execution and treats the
  approved reverse-plan job as the authorization boundary.
- Symptom: the binary refused the empty job before dispatch with
  `validation/flow_job_reference_refused` (exit 3), yielding no provider write
  or durable receipt.

### Divergence, counterfactual, and falsifier

The earliest divergence is `flow run` → `resolveManifestJobs` → empty action
job → typed malformed refusal. The proven path resolves a valid `rplan_…` job
and derives action scope from it. `resolveManifestJobs` landed in #4168
(`5c12fb536`), after both #4150 and #4155; those route changes are therefore
not causal.

The smallest counterfactual changed only that action: `job: planID` and
`action_cfg.read_back_stream`. The same fresh-binary test passed. This PR makes
that counterfactual the fixture, without weakening an assertion.

The shared-root theory with #4158 was falsified: the failing test never
constructs a PostgreSQL managed-target driver, database-write plan, or history
route. Its tagged live PostgreSQL check correctly skips without explicit
container opt-in, so #4158 remains separately owned.

## Test Contract

| Class | Test | Assertion |
| --- | --- | --- |
| Happy | `TestFreshBinaryDeclarativeGitHubWarehouseFlowRoundTrip` | A fresh binary emits 1 sync record and 1 action record; 2 provider comments, 1 warehouse row, checkpoint, and receipt are observable. |
| Bad | `TestFlowActionWithoutApprovedJobReferenceRefusesBeforeIO` | Malformed typed job-reference error maps to `validation/flow_job_reference_refused`; no stored flow and zero target events. |
| Edge — revoked | `TestFlowActionRevokedJobReferenceRefusesBeforeIO` | Exact unapproved reason and `AuthorizationRevokedError`; zero target events. |
| Edge — stale | `TestFlowActionStaleJobReferenceRefusesBeforeIO` | Exact missing reason; zero target events. |

## Lifecycle, Skills, and Parity

- GSD lifecycle recorded inline: `discuss-phase`, `plan-phase --tdd`,
  `execute-phase`, `verify-work`, and `code-review`. Inline fallback is
  documented because this lane forbids lifecycle-role spawning.
- Loaded Go skills: `golang-how-to`, troubleshooting, CLI, testing, testify,
  error handling, security/safety, design/interfaces, context/concurrency,
  database, and lint.
- No CLI command, flag, output, or documentation contract changed. The fixture
  follows the existing job-backed documentation in `docs/cli/flow.md`.
  `pm help flow`, `pm flow`, and `pm flow --help` all succeed.

## Verification

- `go test -timeout 20m ./internal/cli` — pass (382.569s)
- `go test -timeout 20m ./internal/flow ./internal/app` — pass
- Fresh-binary acceptance + three contract guards — pass
- `go vet ./...`, `go build ./cmd/pm` — pass
- `make tidy-check`, `lint`, `docs-check-no-build`, `smoke-no-build`,
  `agent-contract-check`, connector validation/drift/boundary/canon checks,
  and `release-workflow-check` — pass

## Safety and Follow-up

No credentials, network credentials, or approval tokens are logged or stored.
The happy path performs the test's approved reverse action only through the
fresh binary. #4158 is not changed; its managed-target route investigation is
separate.

## Review

Primary route: `claude_auto` after this non-draft PR opens against
`integration/4015-mvp-flat-r1`. No manual Claude request is appropriate before
the automatic route has a chance to run; Copilot is fallback-only. Manual
pre-PR deep review found no blocking findings. Commit/push checkpoints:
`bf3af2c2d` (plan), `840c91826` (investigation), and the fixture migration
commit in this PR.
