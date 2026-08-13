<!-- Prepared only. Do not submit until fresh no-mistakes succeeds and the #3754 parent route is authoritative. -->

## Intent

Refs #4072

Refs #3754

Deliver the preserved #3754 child that routes the GitHub App
installation-token POST through engine-owned declared-route rate admission.
Missing or lost `require_shared` coordination must refuse before transport; a
grant must retain exactly one Decide/send/Finish lifecycle. This child excludes
UDS availability, GraphQL policy changes, provider calls, credentials, CLI
surface changes, generic transport changes, parent completion, and merge.

## What Changed

- Builds the engine rate resolver before custom GitHub App authentication.
- Gives the GitHub hook a narrow declared-route JSON capability rather than a
  coordinator, runtime, arbitrary URL, or generic HTTP writer.
- Routes the actual escaped installation-token path through declared `POST
  /app/installations/{installation_id}/access_tokens` admission.
- Retains secret-blind local tests for zero-send missing/lost coordination,
  grant lifecycle, route separation, privacy, and existing GitHub behavior.

## Red / Green / GSD

- RED: `9a44c9163` (causal premature send) and `3f20bf7ba` (expanded
  no/lost/grant/privacy matrix).
- GREEN: `3f83bf3af`.
- Focused verification/review: `72c573bca`.
- Fresh correction ledger: 1/5; the only consumed correction is the resolved
  configured-linter RED. The inherited generated-artifact synchronization does
  not consume correction 2/5.
- The named issue phase is outside the numeric roadmap and the canonical
  single-worker contract forbids role spawning. GSD discuss/plan/execute,
  gap planning/execution, verify-work, and deep review are therefore recorded
  as inline fallbacks in `.planning/phases/issue-4072-github-app-auth-admission-r1/`.

## Testing

Focused local tests, bounded race coverage, and scoped vet are recorded as
passing in the phase. The report-defined broad local matrix and fresh
no-mistakes delivery record are intentionally **pending** until the prepared
handoff is released; do not publish this body before those fields are updated.

No credentials were inspected and no provider call or mutation was made.

## Generated artifact traceability

- RED: at clean `31be96f…`,
  `go run ./cmd/connectorgen certification-matrix --check` failed because the
  tracked capability matrix had inherited drift; recovered base `da8a8ff…`
  fails identically.
- GREEN: the canonical generator and its `--check` passed. The resulting
  matrix SHA-256 is
  `e63b906cb640b8fb4fc8fd46c1076b77b7dbced7889919d60527f9b4335d520a`.
  Its diff is six `discovery_source` line updates only; the stripped semantic
  SHA-256 is unchanged at
  `bc5d14758c26755d83a9dc4dcbb715da31d95f67de38e352bc652b752c0819bc`.
- #4026's #4034 delivery is generator precedent only. No PostgreSQL or #4034
  ancestry is imported, and no generator/source/semantic change is included.

## Pipeline

The eventual local no-mistakes run uses `--skip=push,pr,ci` and no `--yes`.
Record its accepted run ID, final child head, and exact broad-validation results
here before any draft PR is created. External CI and automated-review evidence
must be exact-head evidence after the draft PR exists.

## Stack and follow-up

This draft child must target `feat/3754-shared-rate-coordinator`, only after
that parent contains `da8a8ff…` and has its one authoritative draft PR to
`docs/4015-connector-release-certification`. The UDS availability child stays
deferred until this child has been human/captain-integrated into #3754.
