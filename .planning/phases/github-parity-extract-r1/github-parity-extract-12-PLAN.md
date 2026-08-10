# Plan — target rebind and authenticated PM-read finding

## Goal

Safely rebind the live GitHub lab to the captain-approved private repository
`karthik-sivadas/pm-live-test-direct-read-20260808081515`, with immutable owner
ID `6113982` and immutable repository ID `1327549621`, without making any
historical target eligible for another provider request. Preserve the two
authenticated `pm github repos list-for-authenticated-user` failures as a
first-class connector finding, capture its complete safe `Error` envelope, and
determine whether its failure is shared direct-read machinery or specific to
this GitHub operation.

## Fixed boundary and evidence decisions

- The current `allowed_targets` array contains exactly one run-owned repository:
  the captain-approved owner slug/ID and repository slug/ID above. Every normal
  PM fixture read or write continues to pass both slug and immutable-ID checks
  before a PM process starts.
- The prior lab target and its append-only cleanup events remain historical
  evidence only. Historical entries may be validated against their exact
  archived run/target identity, but no executor may authorize them for a
  current PM request. This preserves custody without adding a second mutation
  target.
- The authenticated repository-list command is a defect, not setup friction.
  Its exact PM invocation, complete safe `Error` envelope, and a provider HTTP
  status when present are captured in a GitHub planning finding. Tokens,
  approval grants, response payloads, stderr, and credential secrets are never
  recorded.
- No provider mutation is authorized by this plan until the re-bound boundary,
  historical-ledger guard, and relevant local tests are green. The diagnostic
  read uses only `pm github`, is bounded, and is allowed by the captain's
  explicit credentialed-live-proof authority.

## TDD slices

### Red 32a — rebinding must not revive a retired provider target

Add a lab-boundary regression with one current target and one archived
historical run. It must prove that the historical append-only cleanup ledger
still validates, while `authorizeLabTarget`, `runPMScopedRead`, and
`runPMPlannedWrite` refuse that retired target before any PM process can start.
The existing one-run validator is expected to fail because it requires every
entry to have the current run ID and current allowlist identity.

### Green 32a — archived evidence is validation-only

Extend only the GitHub live-lab harness schema/validator with an explicit
historical-run target archive. Validation selects the archive solely for old
ledger rows; execution authorization continues to use `allowed_targets` only.
Rebind the committed current boundary to the captain's exact immutable ID and
preserve the prior ledger unchanged.

### Red 32b — an authenticated PM error must be capturable without becoming a
secret or a vague failure

Add a unit regression for a full JSON `kind: Error` envelope. The helper must
retain every safe envelope field, reject token-shaped/sensitive fields, preserve
the exact PM command identity, and report a provider status only when the
envelope actually carries one. It must not classify a missing status as a
successful or credential-specific response.

### Green 32b — bounded, sanitized finding evidence

Add a narrowly scoped lab-harness helper used by the live diagnostic runner to
validate the error envelope and write a sanitized finding record. Reproduce the
same current-head `pm github repos list-for-authenticated-user` invocation once
from a fresh 0700 lab root; retain the complete safe envelope and status/null,
then trace dispatch, credential resolution, direct-read execution, and error
serialization. Compare it with a local loopback/direct-read control to decide
whether the defect is shared infrastructure or this operation's declaration.

## Verification

Before a credentialed PM diagnostic read:

```bash
node --test scripts/tests/github-live-lab.test.mjs
node scripts/github-live-lab.mjs --check-boundary \
  --boundary .planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-BOUNDARY.json
go build ./cmd/pm
git diff --check
```

After implementation and the one bounded diagnostic read:

```bash
node --test scripts/tests/github-live-lab.test.mjs
go test -timeout 20m ./internal/cli ./internal/connectors/commandrunner -count=1
go vet ./internal/cli ./internal/connectors/commandrunner
go build ./cmd/pm
node scripts/github-live-lab.mjs --check-boundary \
  --boundary .planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-BOUNDARY.json
git diff --check
```

CLI help/manual/website parity is intentionally not applicable unless diagnosis
changes a user-visible PM command, flag, help topic, or output contract. The
lab-only finding serializer must not change provider command help.

## GSD and skills record

`scripts/gsd doctor`, all required command-source resolutions, and the generated
`discuss-phase` / `plan-phase --tdd` prompts were run and read inline. The
single-worker manual fallback is required because GSD role delegation is not
authorized for this lane. Skills loaded: `golang-how-to`, `golang-cli`,
`golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, `golang-structs-interfaces`, `golang-troubleshooting`,
and `javascript-testing-patterns`.
