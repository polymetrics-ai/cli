# Code Review — Issue 4166 Live Write Certification

**Method:** manual standard-depth fallback. `scripts/gsd prompt code-review
4166` resolved the official workflow, but this execution environment cannot
provide its required isolated reviewer role. The project contract permits the
documented inline/manual fallback; no reviewer role was spawned.

## Reviewed scope

- `internal/connectors/certify`: report aggregation, full-parity stage
  eligibility, durable lifecycle ledger, and repository wave production path.
- `internal/cli`: option parsing, help/manual output, external-proof refusal,
  and golden transcript.
- User docs, generated docs data, and Issue 4166 GSD/TDD artifacts.

## Checks and disposition

- **Safety:** the wave rejects every owner/repository other than
  `Polymetrics-Cert/pm-cert-3993-20260810-wz0fru` before provider setup.
  Mutations use the normal seed → reverse plan → preview → approval → run
  route; no raw provider client or ambient GitHub CLI authentication is added.
- **Truthfulness:** `pass` requires each scenario's production mutation,
  independent read-back, and verified cleanup. The item-read refusal produces
  `blocked`; interrupted-resource reconciliation produces
  `recovered_unverified`; both prevent a pass/full-parity claim.
- **Recovery:** resource cleanup requires an exact ownership tag, and topics
  restore only a ledger-captured baseline. Unknown topic state becomes a leak
  instead of an overwrite.
- **CLI/docs parity:** `--full-parity` forces full/write; `--write-only` is
  bounded and cannot claim parity. Runtime help, CLI manual, website reference,
  generated docs, help tests, and golden transcript agree.
- **Finding:** none. The review also removed the superseded combined issue/
  label scenario that lint identified as unused.

## Evidence

`go test -timeout 20m ./internal/connectors/certify`,
`go test -timeout 20m ./internal/cli`, `make lint`, the generated-doc stability
check, and the repository validation gates recorded in `VERIFICATION.md` pass.

## Connector-boundary correction review

**Method:** manual standard-depth fallback after `scripts/gsd prompt
code-review 4166`; this execution environment cannot provide the required
isolated reviewer role, and the project contract permits the documented inline
fallback.

- **Definition boundary:** the shared certification runner now consumes typed
  `write_inventory` and `write_wave` declarations. The fixture, action set,
  bindings, pairings, tags, inventory prerequisites, and known item-read
  blocker remain in the connector definition.
- **Fail closed:** malformed profiles, missing fixture configuration, unknown
  scenario selections, undeclared actions, and action bindings outside the
  declared wave stop the stage before a provider write.
- **Regression coverage:** the focused engine/certify tests load the real
  definition, prove the action inventory/wave relation, blocked outcome, and
  fixture guard. `connector-boundary` and byte-stable matrix regeneration pass.
- **Finding:** none.

## Certification timing correction review

**Method:** manual standard-depth fallback after `scripts/gsd prompt
code-review 4166`; the environment has no compatible isolated reviewer role.

- **Performance:** inventory classification receives the one already-loaded
  profile, eliminating one complete bundle load per declared action without
  weakening malformed-definition failures.
- **Non-vacuity:** lightweight test discovery counts every `write_wave`
  declaration and fails if none exist or a declaration cannot produce the
  production `certificationWriteWaveFor` result.
- **Verification:** `go test -timeout 20m ./internal/connectors/certify`, the
  full CLI package, the repository gate set, and unchanged-budget `make
  certify-timing` pass.
- **Finding:** none.
