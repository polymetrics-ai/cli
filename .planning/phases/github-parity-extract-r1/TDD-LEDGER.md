# TDD-LEDGER — github-parity-extract-r1

Two red/green cycles. The first re-observes the sweep branch's parity red against
current `main` rather than inheriting it; the second is the captain's unblock order.

---

## Cycle 1 — github's documented REST surface is incomplete on `main`

**Red** — observed after cherry-picking only the surface test (`6511a8a45`), before any
bundle change:

```
$ go test ./cmd/connectorgen/ -run 'GitHub|Github'
--- FAIL: TestGitHubDocumentedRESTSurfaceIsComplete (0.00s)
    github_documented_surface_test.go:168: REST endpoints = 505, want 1220 documented operations
    github_documented_surface_test.go:177: restByMethod = map[DELETE:72 GET:259 PATCH:36 POST:91 PUT:47],
                                           want map[DELETE:187 GET:636 PATCH:70 POST:193 PUT:134]
    github_documented_surface_test.go:191: expected "GET /orgs/{org}" — the shipped bundle enumerated only /repos/{owner}/{repo}/…
    github_documented_surface_test.go:191: expected "GET /user" — …
    github_documented_surface_test.go:191: expected "POST /markdown" — …
    github_documented_surface_test.go:139: POST /app/installations/{installation_id}/access_tokens: blocked row must carry a 'Named dependency:' marker
    github_documented_surface_test.go:157: 4 synthetic path row(s) are not documented endpoints
FAIL
```

This is the important line of this ledger: the red was **re-derived here**, not copied.
505 was the true state of `main` at `08cc41c87`.

**Green** — after `8d00a55a9` (GET surface), `997d7391b` (covered_by.writes) and
`5ea17aa41` (parity):

```
$ go test ./cmd/connectorgen/
ok  	polymetrics.ai/cmd/connectorgen	12.321s
```

---

## Cycle 2 — the captain's restored commands (`CAPTAIN-ORDER-unblock-commands.md`)

**Red** — `f45640a27` adds `internal/connectors/commandrunner/github_unblocked_commands_test.go`
with no implementation behind it:

```
$ go test ./internal/connectors/commandrunner/ -run TestGitHubRestoredCommandsAreExecutable
--- FAIL: TestGitHubRestoredCommandsAreExecutable (1.55s)
    --- FAIL: .../repo_create     github "repo create" availability = "unsafe_or_disallowed", want implemented
    --- FAIL: .../repo_delete     github "repo delete" availability = "unsafe_or_disallowed", want implemented
    --- FAIL: .../repo_archive    github "repo archive" availability = "unsafe_or_disallowed", want implemented
    --- FAIL: .../repo_unarchive  github "repo unarchive" availability = "unsafe_or_disallowed", want implemented
    --- FAIL: .../cache_delete    github "cache delete" availability = "unsafe_or_disallowed", want implemented
    --- FAIL: .../secret_set      github "secret set" availability = "unsafe_or_disallowed", want implemented
    --- FAIL: .../secret_delete   github "secret delete" availability = "unsafe_or_disallowed", want implemented
FAIL
```

7/7 subtests red. `TestGitHubHeldCommandsStayBlocked` was green from the start — it asserts
`auth token` and `api` stay blocked, which is the state the captain asked to preserve.

**Red (intermediate, found by the validator, not by me)** — wiring `secret set` to
`actions_secrets_secret_name3` surfaced a defect in the already-`implemented` twin:

```
$ go run ./cmd/connectorgen validate internal/connectors/defs
github: cli_surface.json: [cli_surface_safety] implemented reverse ETL command 88 ("secret set")
  flag --encrypted-value maps outside write "actions_secrets_secret_name3" schema:
  record field "encrypted_value" is not declared
github: cli_surface.json: [cli_surface_safety] … --key-id … "key_id" is not declared
connectorgen validate: 551 connector(s) checked, 2 finding(s)
```

and, after declaring them:

```
github: cli_surface.json: [cli_surface_missing_mapping] implemented reverse ETL command 191
  ("secret set-2") for write "actions_secrets_secret_name3" lacks flag mappings for required
  record fields: encrypted_value, key_id
```

The second finding is the real one: `secret set-2` was already shipping `implemented` while
its write action's `record_schema` carried only the path parameter, so it could only ever
PUT an empty body. Both commands now declare the fields the API requires.

**Green** — `93de56c5a`:

```
$ go test ./internal/connectors/commandrunner/ -run 'TestGitHubRestoredCommandsAreExecutable|TestGitHubHeldCommandsStayBlocked|TestEveryImplementedCommandPassesRuntimePreflight'
ok  	polymetrics.ai/internal/connectors/commandrunner	10.778s

$ go test ./internal/connectors/hooks/github/ -run 'ArchiveRepo|UnarchiveRepo' -v
--- PASS: TestExecuteWrite_ArchiveRepoPinsArchivedTrue (0.00s)
--- PASS: TestExecuteWrite_UnarchiveRepoPinsArchivedFalse (0.00s)

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 551 connector(s) checked, 0 findings
```

No test was weakened, skipped, or deleted in either cycle. Three tests were added.

---

## Cycle 3 — the shared-code redaction defect (found in review, closed here)

Not a github cycle. Review of this branch found that restoring `secret set` made
`--encrypted-value` reachable while the `redact_fields` mitigation declared for it was inert.
Closed in three red/green steps on the captain's ruling.

**Red 3a — the `--json` half.** `RedactReversePlanRecords(plan.Sample, plan.RedactFields)` was a
no-op because the connector-command plan constructor never populated `RedactFields`, so
`sample[0].encrypted_value` was emitted verbatim.
**Green** — `3f78743ad`; `internal/cli/reverse_plan_redaction_test.go` asserts the sample field is
`"redacted"`, that the sentinel is absent from the emitted JSON, and that a plan declaring no
redact fields round-trips unredacted.

**Red 3b — the at-rest half.** `ConnectorCommandRecord` still carried the raw value into
`state.json`, permanently, because reverse plans are append-only. The first fix's test asserted
only on the `sample` sub-slice, which is exactly how this survived.
**Green** — `148b763b1`; `internal/app/connector_command_withholding_test.go` reads the **raw
bytes** of `.polymetrics/state/state.json` and asserts the sentinel is absent after plan, after
preview and after execute, for both a sealed-secret field and a required bearer-token fixture,
while non-sensitive fields the plan needs survive.

**Red 3c — the operation-backed path.** The withholding key resolved `writeCommand.Write` against
`ManifestOf(connector).WriteActions`, but `buildOperationDirectWriteCommand` sets `Write` to an
operation ID. The namespaces are disjoint in 550 of 551 bundles, so the lookup returned nil and
withholding silently did nothing. asana was worse: 11 names exist in both namespaces.
**Green** — `133d7174a`; `internal/app/operation_command_withholding_test.go` proves an
operation-backed plan withholds its `sensitive_policy.redact_fields`, and
`TestOperationBackedPlanIgnoresSameNamedWriteAction` proves the asana collision resolves to the
operation, not the write action — the test that fails if anyone reintroduces a fallback.

Two existing tests changed, both contract adaptations rather than weakenings, and both recorded
here rather than quietly updated:

- `defs/asana/reverse_etl_execute_test.go` now passes `WithheldFlags: flagsFromRecord(...)` at
  execute, mirroring the real operator flow under the new re-supply contract. It still asserts the
  write executes.
- `defs/zendesk-support/reverse_etl_execute_test.go` passes `nil` for `PreviewReversePlan`'s new
  parameter. Signature only.

No test was weakened, skipped, or deleted. Cycle 3 added three test files.

---

## Note on the run that produced cycles 2–3

The `no-mistakes` run `01KZEY2NZJ88R819PY5J3XK5JD` completed its review step with 0 findings and
then **failed at the test step for an external reason** — the agent's monthly spend limit, not a
code failure. Its five commits were preserved unpublished in the local gate and recovered with
`no-mistakes axi sync --recover` before pushing. The gates below were therefore re-run locally on
the recovered head.

---

## Cycle 4 — review round: withheld-field re-supply, plural write coverage, CLI parity

**Red 4a — a declared field that was never supplied became an unsatisfiable precondition.**
`reconstituteConnectorCommandRecord` built its pending set from `plan.RedactFields` and treated
"absent from the record" as "withheld". Those are different states: `withholdRecordFields` deletes
nothing when the operator never supplied an optional declared-sensitive field, yet the field was
still demanded back, and no later value could recover the plan because its hash was computed
without the field. A sweep of the committed bundles found 71 implemented/partial reverse-ETL
commands whose declared redact field has no mapping flag at all (for example `mailchimp actions
post-batch-webhooks` -> `url`, `amazon-sqs queue create` -> `attributes`), permanently un-previewable
and un-executable, plus 36 more that break whenever an operator omits an optional declared field.
GitHub itself was unaffected: `actions_secrets_secret_name3` has `encrypted_value` in
`record_schema.required`, so the field is always present and always genuinely withheld.
**Green** — `internal/app/connector_command_withholding_test.go`
`TestDeclaredButUnsuppliedFieldIsNotOwedBack` covers both shapes (declared field behind an
omitted optional flag, and declared field with no flag on the command at all) and asserts plan,
preview and run all succeed with no re-supply. `withholdRecordFields` now returns the fields it
actually deleted, `ReversePlan.WithheldFields` persists that set, and reconstitution iterates it.
Pre-existing plans carry no `withheld_fields` and therefore ask for nothing.

**Red 4b — withhold and reconstitute disagreed on the field-path prefix.**
`withholdRecordFields` stripped only `record.`, while `commandrunner.ReconstituteWithheldFields`
strips `body.` for an operation-backed command. A `body.`-prefixed `sensitive_policy.redact_fields`
entry would have left the secret in `state.json` while still forcing a re-supply. Latent today (no
bundle uses that spelling, and the schema does not constrain it), so this is a same-cycle
structural fix rather than a separate failing test: `connectorCommandRedactFields` now normalizes
every declared path against `connectorCommandRecordPrefix(operation)`, which is the same mode
dispatch the runner half uses, so the two halves cannot drift apart by construction.

**Red 4c — `pm connectors certify github --full` failed on this branch.**
`certify`'s surface inventory was never migrated to `SurfaceCoverage.WriteTargets()` when the
plural `covered_by.writes` was introduced in this program. Three GitHub endpoints are covered only
through the plural form (`PATCH /repos/{owner}/{repo}`, `.../issues/{issue_number}`,
`.../pulls/{pull_number}`), none carries an `operation` block, so the stage fell to its default
branch and failed with "api_surface endpoint N is neither covered nor blocked with typed reason".
**Green** — `internal/connectors/certify/stages_surface_inventory_internal_test.go`
`TestSurfaceInventoryCountsPluralWriteCoverage` pins a `writes`-only endpoint as covered with all
its targets counted; `hasSurfaceCoverage` and `addSurfaceCoverageCounts` now route through
`WriteTargets()`, the same helper `connectorgen` and `conformance` already use.

**Red 4d — two certify counts were stale and red on this branch.**
`internal/connectors/certify/` was not in the gate list of the previous verify-work run, so the
sweep's growth of GitHub's declared surface went unnoticed there.
`TestSurfaceInventoryForGitHubAccountsForAllReviewedEndpoints` expected 509/440/69 and
`TestGithubWriteActionInventoryAccountsForAllDeclaredActions` expected 231.
**Green** — both updated to the surface this branch actually ships: 1224 endpoints, 1126 covered,
98 blocked (the deliberately-blocked 98 foundations), `covered_by[write]` 555 and 555 declared
write actions — the two agree, which is the coherence check. Assertions were tightened, not
relaxed: `covered_by[direct_read]` (368) is now asserted too.

**Red 4e — the re-supply step changed the CLI contract with no help, docs or website record.**
`pm reverse preview` and `pm reverse run` now accept and require the connector command's own flags
whenever a plan withheld a declared field, and nothing said so.
**Green** — `reverseHelp` in `internal/cli/docs.go` documents the re-supply flags in USAGE, FLAGS,
DESCRIPTION, both COMMANDS entries and SECURITY; `docs/cli/reverse.md` regenerated with
`pm docs generate`; `internal/cli/testdata/golden_transcripts.json` regenerated (3 reverse
entries); `website/content/docs/reverse-etl.mdx` gained a re-supply section and a security-model
row, with `website/lib/docs.generated.ts` regenerated.

No test was weakened, skipped, or deleted. Cycle 4 added two tests and corrected two stale counts.
