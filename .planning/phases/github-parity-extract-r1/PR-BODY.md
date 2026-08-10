## Intent

Extract GitHub's documented-operation parity from the paused
`fm/cli-top50-sweep-resume2-r1` work and land it as a separate PR. This PR is not the
consolidated sweep PR and does not merge anything into `main`.

Refs #3971.

The complete GSD/no-mistakes delivery record is included under
`.planning/phases/github-parity-extract-r1/`.

## Current-ref — merged GitHub surface

The authoritative source-derived count and its provenance are in
[VERIFICATION.md](VERIFICATION.md). This PR body does not duplicate that generated-surface
measurement.

The historical execution evidence below remains tied to its original checkpoint and does not
assert current-ref execution coverage.

## What Changed

- Regenerated GitHub's full documented surface from the parity generator, including the
  `covered_by.writes` foundation needed for multiple write actions on one endpoint.
- Kept provider-specific behavior in GitHub definitions, generator, hooks, conformance tests, and
  proof scripts. The one shared production change is provider-neutral: direct-read path encoding
  now supports slash-bearing `ref` path variables with a generic safety test.
- Regenerated the operation/command proof ledgers, GitHub generated surface artifacts, golden
  transcripts, and website/catalog artifacts through their owning generators. No generated file
  was hand-merged.
- Added exhaustive source accounting and deterministic provider-double evidence, with current
  source totals kept distinct from retained historical execution totals.
- Kept the exhaustive provider-double implementation in test-only scope and verified the
  connector-boundary scanner remains clean; no GitHub-specific shared production policy was added.
- Added current-head binary dispatch evidence for every command and explicit generic routes for
  the 23 stream and 38 write-action members without dedicated connector commands.
- Added a GitHub-specific proof of the existing rate-limit admission/observation path; no second
  limiter or credential-derived coordination key was introduced.

## Historical safety classification

`repo create` is `implemented` as an approval-gated reverse-ETL write. Creating a private test
repository is a bounded, reversible state-creation operation and is no less defensible than the
already implemented issue/file writes. `repo archive`, `repo unarchive`, and `secret set` remain
approval-only writes.

`repo delete` is also reachable under its deliberate name, but remains protected by the existing
typed destructive confirmation derived from its `DELETE` method. It still requires plan, preview,
typed `--confirm destructive`, approval, and execute. `issue delete` remains blocked because the
documented capability is GraphQL mutation-only and this PR does not add a GraphQL mutation
executor. The approved live target already existed and was required to be retained, so `repo
create` was not executed in the live sweep; its classification remains implemented and
approval-gated.

## Historical evidence

These checkpoint results are preserved as recorded at base ref
`4df0b0416e46958d9acb1b02708464570c070e0f` on 2026-08-10. The credentialed live report was
measured at 2026-08-08T22:27:16.716Z; neither it nor the provider-double report certifies the
current-ref surface above.

| Evidence | Result |
| --- | --- |
| Source-derived operation/command ledgers | 1,224 endpoints; 1,179 commands; all declarations accounted for |
| Built-binary reachability | 1,179 / 1,179 exact `pm github <path>` names reachable; 1,081 implemented, 37 partial, 5 unsafe/disallowed, 21 unsupported-local, 8 planned, 27 unsupported-api |
| Deterministic provider double | 37 streams + 574 writes + 377 operations; 985 exercised, 3 concrete untestable, 0 failed |
| Generic ETL/reverse-ETL routes | 23 / 23 streams and 38 / 38 write actions exercised through the generic engine routes |
| Limiter proof | 2 same-scope requests, 1 local 60-second wait, 0 provider 429s; independent scope adds 0 waits |
| Credentialed live proof (2026-08-08) | 124 / 1,081 proven; 957 terminally `UNTESTABLE` with concrete target, permission, or safety reasons; 0 `FAILED` |

`LIVE-PROOF-REPORT.json` is historical credentialed acceptance evidence. The run was hard-pinned
to private `karthik-sivadas/pm-live-test-direct-read-20260808081515`, which was retained after
validation. It records one terminal `PROVEN`, `UNTESTABLE`, or `FAILED` result for each of the
1,081 implemented commands in that snapshot, matches its then-current surface/case/binary hashes,
and contains no subprocess output, response body, approval grant, credential, or token-derived
value. The approved private repository was retained after reversible write/readback checks.

## Generated-artifact boundary

The shared operation endpoint ledger delta was regenerated and audited to contain only the
`github` connector key. Website connector data and GitHub docs/manual artifacts were regenerated
through their owning generators. Golden transcripts were regenerated as well; their diff includes
the expected shared reverse-contract snapshots from the foundation change plus the GitHub inspect
and generated-surface snapshots. The generated GitHub write paths use the engine's
`{{ config.* }}` / `{{ record.* }}` path dialect; provider API paths remain in their single-brace
source form.

## GSD / TDD / review record

- Lifecycle resolved through `scripts/gsd sources discuss-phase`, `plan-phase --tdd`,
  `execute-phase`, `verify-work`, and `code-review`.
- The generated prompts were executed inline/manual because this runtime cannot provide the
  compatible isolated Pi roles; that fallback is recorded in `VERIFICATION.md`, the phase plan,
  and the TDD ledger.
- Red/Green evidence is in `TDD-LEDGER.md`, including source-ledger red/green, binary red/green,
  provider-double red/green, limiter proof, and the resolved credentialed live acceptance cycle;
  earlier external-blocker history remains recorded as historical evidence.
- Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-documentation`, and `golang-lint`.

## Testing

Targeted and generated checks:

```text
node --test scripts/tests/github-parity-proof.test.mjs
node --test scripts/tests/github-live-proof-sweep.test.mjs
node scripts/github-parity-proof.mjs --check
node scripts/github-live-cases.mjs
node scripts/github-live-rate-limit-proof.mjs
go run ./cmd/connectorgen surface-sync --check
go run ./cmd/connectorgen validate internal/connectors/defs
go run ./cmd/connectorgen boundary . --json
go test -timeout 20m ./cmd/connectorgen/
go test -timeout 20m ./internal/connectors/engine/
go test -timeout 20m ./internal/connectors/conformance/
go test -timeout 20m ./internal/connectors/commandrunner/
go test -timeout 20m ./internal/cli/
go vet ./...
go build ./cmd/pm
go test -timeout 20m ./internal/connectors/engine/ -run TestDirectReadAllowsSlashBearingRefPathVariables
```

The full suite and `make verify` are run as separate gates per `AGENTS.md` because this repository's
550+ connector suite exceeds per-command agent timeouts; CI remains the final full-suite gate.

CLI parity checked: `pm help github`, bare `pm github`, representative `pm github <path> --help`,
generated GitHub manual/skill/catalog artifacts, website catalog scope, and golden transcripts.

## Checkpoints and follow-up

- Branch: `fm/cli-github-parity-extract-r1`; never push to `main`.
- The paused `fm/cli-top50-sweep-resume2-r1` branch must be rebased onto the resulting `main`
  after this PR lands; its copied GitHub commits and shared generated ledger will conflict
  otherwise. This is an explicit post-merge action, not part of this PR.
- The final no-mistakes pipeline must run after the coherent proof commit and before merge. Any
  actionable automated-review finding will be dispositioned in GitHub; merge remains human-gated.
