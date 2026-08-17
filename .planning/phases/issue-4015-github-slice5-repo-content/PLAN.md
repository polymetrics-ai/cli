# Issue #4015 — GitHub slice 5 repo-content live certification

## Delivery contract

- Issue: `Refs #4015`
- Branch: `fm/cli-cert-slice5-repo-content`
- Base: `origin/integration/4015-mvp-flat-r1`
- Pull request: #4225
- Slice source: `data/cert-slices/slice-5-repo-content.json` from the certification workspace
- Scope: 174 GitHub repo-content commands, with live reads executed one at a time against disposable `Polymetrics-Cert` resources
- Branch contribution: 25 schema-v2 evidence records over the integration base
- Exclusions: mutations remained paused by captain direction; shared certification matrix, sweep, and candidate artifacts were not regenerated in this lane

## Method

For each eligible read command:

1. Route the measured credential type and retry once with the alternate credential when a 403 or 404 could be credential-dependent.
2. Resolve real provider object identifiers from collection endpoints before considering a fixture.
3. Execute the real `pm github ...` command against GitHub.
4. Assert a produced value, using the declared assertion or an `agent_derived` assertion with a plausible wrong-answer negative control.
5. Classify the command without treating exit status or `err == nil` as success.
6. Write and validate schema-v2 evidence immediately for accepted passes.
7. For product defects, compare the connector result with a raw `api.github.com` control.
8. Prove fixture absence through provider collection read-backs.

## Inline manual-GSD fallback

This certification lane was dispatched as one canonical worker with role spawning prohibited. The repository GSD adapter was available and its command sources were resolved, but compatible isolated lifecycle agents could not be spawned under that delivery contract. The lifecycle therefore ran inline in the lane, and this phase records that work retrospectively after CI exposed the missing branch-local artifact.

- `scripts/gsd doctor` — PASS
- `scripts/gsd sources discuss-phase` — resolved
- `scripts/gsd sources plan-phase` — resolved
- `scripts/gsd sources execute-phase` — resolved
- `scripts/gsd sources verify-work` — resolved
- `scripts/gsd sources code-review` — resolved
- `discuss-phase`: the supplied 174-command slice, reads-only captain decision, credential safety, provider controls, and no shared generated artifacts defined the boundary.
- `plan-phase --tdd`: execute each read, prove a produced value and its negative control, persist accepted schema-v2 evidence, validate immediately, and retain exact non-pass diagnostics.
- `execute-phase`: all read commands were attempted with one bounded retry per obstacle; accepted passes produced the 25 branch-owned evidence records.
- `verify-work`: branch records, assertions, connector validation, scoped Go tests, secret scans, and provider cleanup read-backs were checked.
- `code-review`: the diff was reviewed for inherited-record overcounting, secret exposure, false pass claims, generated-artifact drift, and leftover provider fixtures.

Required skills used: `golang-how-to`, `golang-cli`, `golang-testing`, and `golang-security`.

## TDD evidence

- Red: slice commands without a branch-owned, validator-accepted schema-v2 record were not certified. A plausible wrong produced value had to be rejected by the declared or agent-derived assertion.
- Green: 25 branch-owned records have `status: passed`, contain observed HTTP exchanges and produced-value assertions, and pass `go run ./cmd/connectorgen certification-matrix --check`.
- Red: `bash scripts/verify-gsd-workflow origin/integration/4015-mvp-flat-r1` exited 1 with `cmd/internal changed, but no GSD planning evidence changed` at commit `86ff218aa`.
- Green: adding this phase's `PLAN.md`, `TDD-LEDGER.md`, and `VERIFICATION.md` makes the same check accept the branch-local GSD/TDD record.
- Refactor: not applicable; the lane changes evidence and planning records only, not connector behavior.

## Threat model and safety constraints

- Secrets: credentials remain Keychain/environment inputs at point of use and never enter evidence, planning artifacts, logs, PR text, or command arguments.
- Provider scope: only disposable identities and resources inside `Polymetrics-Cert`; no third-party repository, real person, purchase, or public organization-visible fixture.
- Mutations: paused by captain direction. No approval-token-protected write was executed in the final read pass.
- Cleanup: direct provider collection reads, rather than connector delete exit status, are the acceptance proof for absence.

## Completion criteria

- The branch contributes exactly 25 evidence records over the integration base.
- All 25 records report `status: passed` and the certification validator accepts the branch.
- Non-pass outcomes and seven product defects remain documented without inflating the certified count.
- Cleanup read-backs return zero slice-prefixed labels, environments, deployments, and codespaces.
- The GSD workflow-evidence check passes against `origin/integration/4015-mvp-flat-r1`.
