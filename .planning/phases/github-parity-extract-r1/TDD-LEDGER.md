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

---

## Cycle 5 — review round: withheld ancestor subtree, authoritative redact source

**Red 5a — a withheld field that is an ancestor of flag targets had no way back.**
`ReconstituteWithheldFields` resolved a withheld target only by exact `MapsTo` match.
`internal/connectors/defs/recurly/writes.json` action `create_invoice_retry` declares
`redact_fields: ["account.billing_infos", "po_number"]`, while the `implemented` command
`invoices retries create` maps four flags at `record.account.billing_infos.0.*`. Withholding
correctly removed the whole `account.billing_infos` subtree, but `byTarget["account.billing_infos"]`
missed, so preview and run reported "re-supply account.billing_infos" — a path no flag can supply.
Confirmed by hand against the shipped bundle: this was the only ancestor-shaped instance across all
551 bundles, and it failed closed (no dispatch, no leak) but the command worked on `main`.
**Green** — `internal/connectors/commandrunner/withheld_subtree_test.go`
`TestReconstituteWithheldFieldsRebuildsSubtreeFromDescendantFlags` (fixture, plus a negative
control that the missing list names flags rather than the bare ancestor) and
`TestReconstituteWithheldFieldsRebuildsRecurlyBillingInfos` (the shipped recurly bundle, so the fix
cannot regress behind a drifting fixture). Both assert the rebuilt subtree is byte-identical to the
subtree `recordOverrides` built at plan time, which is what makes the plan hash still match.
`reconstituteWithheldSubtree` applies descendant flags in the same sorted target order
`recordOverrides` uses, because `setDottedValue` rejects a sparse array index and so depends on it.

`TestReconstituteWithheldFieldsSkipsUnsuppliedOptionalDescendant` guards the boundary that keeps
this fix from reintroducing cycle 4's bug: an **optional** descendant with no supplied value is
skipped, exactly as `recordOverrides` skipped it at plan time, while a **required** descendant is
always demanded because the plan could not exist without it.

Direct binary evidence, in a throwaway project with a fixture credential:

```
pm recurly invoices retries create --credential recurly-local --account-code acct_fixture \
  --account-billing-infos-0-gateway-code stripe \
  --account-billing-infos-0-payment-gateway-references-0-reference-type stripe_confirmation_token \
  --account-billing-infos-0-transactions-0-attempted-collection-date 2026-08-08T00:00:00Z \
  --account-billing-infos-0-transactions-0-gateway-error-code card_declined \
  --currency USD --due-at 2026-08-08T00:00:00Z --external-recovery-eligible true \
  --line-items-0-unit-amount 10 --preview
-> Created connector command plan rplan_... / resolved request: POST .../invoices/recovery

state.json: redact_fields ["account.billing_infos","po_number"]
            withheld_fields ["account.billing_infos"]
            connector_command_record has no billing_infos subtree
```

`po_number` is declared but was never supplied, so it is correctly absent from `withheld_fields` —
cycle 4's rule holding under cycle 5's shape.

**Red 5b — the withholding docs named "redact_fields" without naming which declaration site binds.**
Withholding resolves from a write action's `redact_fields` or a `direct_write` operation's
`sensitive_policy.redact_fields`. `CommandSurfaceCommand.RedactFields` is the deliberately excluded
third site (`internal/connectors/command_surface.go`), and 38 implemented/partial write commands
declare a list their write action does not — `amazon-sqs permission add` declares six redact fields
against an `add_permission` action that declares none, so nothing is withheld there even though
`pm connectors inspect` shows the list. The new help and website text therefore read as a guarantee
for commands that get none.
**Green** — `reverseHelp` DESCRIPTION and SECURITY now name the two binding sites and state
explicitly that a command-level list is not a withholding guarantee; the website section gained a
warning callout and the security-model row the same qualification. Docs only, no behaviour change.
`docs/cli/reverse.md`, `internal/cli/testdata/golden_transcripts.json` (3 reverse entries) and
`website/lib/docs.generated.ts` regenerated.

No test was weakened, skipped, or deleted. Cycle 5 added three tests.

## Cycle 6 — review round: the captain's per-command confirmation classification

The captain superseded the "retain repo delete as unsafe_or_disallowed" criterion: keep complete
documented-operation parity, build no new gate, and put every named destructive command on the
**existing** typed confirmation path. Each of the eight named commands was inspected before any
edit, because the order was explicit that a correct declaration must not be duplicated.

| command | inspected state | action taken |
| --- | --- | --- |
| `repo delete` | DELETE + `confirm: destructive` | none — already correct |
| `cache delete` | DELETE, destructive by method | none — already correct |
| `secret delete` | DELETE, destructive by method | none — already correct |
| `repo archive` | PATCH, no challenge, hook-executed | `confirm: destructive` + moved the pin to the record |
| `repo unarchive` | PATCH, no challenge, hook-executed | `confirm: destructive` + moved the pin to the record |
| `issue delete` | blocked; GraphQL-only `deleteIssue` | none — no REST endpoint to confirm |
| `issue transfer` | blocked; GraphQL-only `transferIssue` | none — no REST endpoint to confirm |
| `pr revert` | blocked; web-UI workflow, no endpoint | none — no REST endpoint to confirm |

`repo create` and `secret set` stay approval-only creates, as ordered.

**Red 6a — `repo delete` had no test that it cannot reach execution.**
`TestGitHubRestoredCommandsAreExecutable` asserted the classification and the resolver, not the
runtime. Written first and failing against no gate assertions at all:
`internal/app/github_repo_delete_gate_test.go`
`TestGitHubRepoDeleteRequiresConfirmationAndSingleUseGrant` walks plan → run-without-grant →
preview → run-without-confirmation → confirmed run → replay, asserting the DELETE call count at
every stage, because a gate that rejects after dispatching has already deleted the repository.
**Green** — the plan mints no token (preview first), an unconfirmed run is rejected with zero
calls, the confirmed run dispatches exactly once, and the replay is refused with the count still
at one.

**Red 6b — a scripted, non-interactive caller had no test that it cannot obtain a grant.**
`internal/cli/github_repo_delete_cli_test.go`
`TestGitHubRepoDeleteScriptedRunCannotObtainAGrant` drives the whole path through `pm --json`,
the documented agent surface. **Green** — `--json` plan output carries
`confirmation_challenge: destructive` and an empty `approval_token`, and `--json` preview emits
no token either; the token exists only in human-readable preview output, which is what stops an
agent from approving its own repository deletion. Single-use replay is refused there too.

**Red 6c — declaring `confirm: destructive` on `repo archive` broke it instead of gating it.**
The first attempt added the declaration alone. Preview then failed with
`engine: destructive write action "archive_repo" uses a hook without an exact prepared-request
preview`: `prepareDeclarativeWrite` refuses a destructive action whose hook overrides execution,
because the approved preview would not be the dispatched request. The narrow fix produced a
command that cannot run.
**Green** — the pin moved from the request to the record. New optional
`engine.WriteRecordHook`, applied by both `prepareDeclarativeWrite` and `executeApprovedWrite`,
so the approved body is the dispatched body by construction; `archive_repo`/`unarchive_repo`
declare `body_fields: ["archived"]` so the pinned field stays the only field sent.
`internal/connectors/engine/write_record_hook_test.go` pins both sides — the refusal for a
hook-executed destructive action, and the exact `{"archived":true}` prepared body for a
record-mapped one — plus no-op and no-mutation guarantees for the 550 connectors that do not
implement it. `internal/app/github_repo_delete_gate_test.go`
`TestGitHubRepoArchiveIsConfirmedAndSendsPinnedField` asserts end to end that the dispatched
body is exactly `archived=true`/`archived=false`.

**Red 6d — the approval-only classification was asserted in one direction only.**
`TestGitHubRestoredCommandsAreExecutable` checked the challenge when it expected one and skipped
the check otherwise, so `repo create` or `secret set` could gain a typed challenge silently.
**Green** — the else branch now fails on a non-empty challenge for an approval-only create.
`TestGitHubApprovalOnlyCreatesCarryNoTypedChallenge` asserts the same through a real plan.

**Red 6e — the three commands with no REST endpoint were unpinned.**
`TestGitHubHeldCommandsStayBlocked` covered `auth token` and `api` only, and asserted
availability alone. **Green** — `issue delete`, `issue transfer` and `pr revert` are in the held
list, and the test now also asserts each binds no write action and no operation and is refused by
the real `Preflight`. That is a stronger gate than a typed confirmation: the command reaches no
executor at all. `AGENTS.md` forbids inventing an `api_surface` endpoint to make one look
implemented, and GitHub documents none for any of the three.

Also in this round, from the same review: the dead equality branch in
`reconstituteWithheldSubtree`'s comparator was dropped for `sort.Strings` (the slice is built
from map keys, so its entries are unique by construction and the branch was unreachable), and a
dropped clause in the `pm help reverse` withholding paragraph was repaired in `internal/cli/docs.go`,
`docs/cli/reverse.md` and the regenerated `internal/cli/testdata/golden_transcripts.json`.

No test was weakened, skipped, or deleted. The two archive hook tests kept their assertions and
their reasoning and moved to the seam the pin moved to. Cycle 6 added nine tests.
