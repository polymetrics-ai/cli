# TDD-LEDGER — github-parity-extract-r1

Two red/green cycles. The first re-observes the sweep branch's parity red against
current `main` rather than inheriting it; the second is the captain's unblock order.

---

## Current-ref — merged GitHub surface

The authoritative source-derived count and its provenance are in
[VERIFICATION.md](VERIFICATION.md). This red/green ledger does not duplicate that
generated-surface measurement.

The red/green checkpoints below are historical records preserved at base ref
`4df0b0416e46958d9acb1b02708464570c070e0f` on 2026-08-10. Their then-current transcripts,
including Cycle 16 and Cycle 33, are not rewritten with these current-ref totals.

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

## Cycle 7 — review round: the documented surface had to say what the runtime now enforces

Cycle 6 put `repo archive`/`repo unarchive` behind the closed typed confirmation but stopped at the
runtime. The declaration never reached `cli_surface.json`'s `notes`, so `pm connectors inspect
github --json`, `MANUAL.md` and `SKILL.md` still described both commands as plain reverse-ETL
writes. An agent reading the documented surface would build a plan and then fail at run with a
confirmation error the docs never mentioned.

**Red 7a — two destructive commands documented no confirmation.** `repo delete`, `cache delete`,
`secret delete` and five others carry a `notes` marker; `repo archive`/`repo unarchive` had
`notes: null`. **Green** — both entries carry the marker, and all three generated manuals now
render it.

**Red 7b — the marker itself named a flag that does not exist.** The convention string was
`destructive; requires --allow-destructive + typed confirmation`, present on five commands at base
and extended to eight by this branch. Verified against the built binary in a throwaway project:

```
pm github repo delete --credential gh-test
-> Created connector command plan rplan_... / Preview required before an approval token is issued.

pm github repo delete --credential gh-test --allow-destructive
-> error: unknown flag --allow-destructive for command "repo delete"
```

Adding it to two more commands would have documented a second failure of exactly the kind 7a
exists to remove, so the narrow fix was rejected for the class fix. **Green** — all ten github
destructive commands read `destructive; requires typed confirmation --confirm destructive`, which
is the flag `runConnectorWriteCommandFromPlan` parses. No test pinned the old string.

**Red 7c — the derived artifacts were stale against cycle 6's `writes.json`.** The catalog records
`confirm` per write action and had 181 for github where the bundle implied 183, and
`body_fields: ["archived"]` makes `writeActionOptionalFields` return `["archived"]`, which
`guide.go` renders as an `optional fields:` line — the manuals rendered 9 where a regenerated copy
renders 11. `pm docs validate` checks headings, entry counts and icon metadata, not a byte-diff
against the renderer, so nothing downstream would have caught it.
**Green** — regenerated with the generators, not hand-merged:

```
./pm docs generate --dir docs/cli       -> docs/connectors/github/{MANUAL,SKILL}.md
                                           docs/connectors/catalog/all-connectors.json
./pm skills generate --dir docs/skills  -> docs/skills/pm-github/SKILL.md
npm --prefix website run gen:catalog    -> website/data/connectors.generated.json
                                           website/lib/connectors.catalog.data.generated.json
POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli/ -run GoldenTranscript
```

Catalog confirm count 181 -> 183; manual `optional fields:` lines 9 -> 11. The golden transcripts
did not move, because `connectors inspect --json` does not carry command `notes`.

The generators again wanted to rewrite all 551 connectors' manuals, four `pm-*` skills, and the
catalog's `gorgias` and `warehouse` entries. That is the pre-existing drift on `main` that commit
216feae1e already isolated and left for its own PR, so it is reverted here again. Both website
catalogs were diffed per connector after regeneration: the only changed entry is GitHub.

No test was weakened, skipped, or deleted. Cycle 7 is declarations and generated output only; no
runtime behaviour changed.

## Cycle 8 — review round: the confirmation marker had to stop depending on prose

The captain's question for this round: does a generated user-facing help path already make the
typed confirmation discoverable at point of use? Two paths were measured against the built binary
in a throwaway project, using an annotated command and an unannotated one that reaches the same
gate.

**Runtime plan output — already exhaustive, no change needed.** The plan step prints the
requirement for every destructive command, annotated or not, because it reads the resolved write
action rather than the bundle's notes:

```
pm github repo delete --credential gh-help
-> Created connector command plan rplan_... for repo delete
   Preview required before an approval token is issued.
   Confirmation required: --confirm destructive

pm github release delete --credential gh-help --release-id 42     # no bundle note
-> Created connector command plan rplan_... for release delete
   Preview required before an approval token is issued.
   Confirmation required: --confirm destructive

pm github issue create --credential gh-help --title t             # not destructive
-> Created connector command plan rplan_...
   Approval token: f0d38b2a...
```

So the failure the finding describes — build a plan, then fail at run with an unexplained
confirmation error — does not occur: the plan itself names the flag.

**Red 8a — `--help` was the path that did not.** `pm github repo delete --help` rendered
`NOTES: destructive; requires typed confirmation --confirm destructive`, but
`pm github release delete --help` rendered no confirmation at all, because NOTES is prose an
author writes per command and only ten of github's 173 destructive commands carry it. After cycle
7 gave ten commands the marker, its absence on the other 163 actively reads as "no confirmation
needed" — a reader comparing the two help pages would conclude `release delete` does not need one.

**Green** — the help states the requirement from the bound executor instead of from prose. New
`commandrunner.ConfirmationChallengeForCommand` resolves it through the same declarations
`buildWriteCommand` and `buildOperationDirectWriteCommand` read, and
`renderConnectorCommandDetail` renders a `CONFIRMATION` section from it. This is the boundary
`writeConnectorDownloadFlags` already establishes for download flags: help and runtime read one
source, so they cannot disagree. It covers every connector's destructive commands, not just
github's, and it adds no bundle notes.

```
pm github release delete --help
-> APPROVAL      Reverse ETL writes require plan, preview, approval, execute.
   CONFIRMATION  execution requires the typed confirmation --confirm destructive
```

`internal/cli/connector_confirmation_help_test.go` pins both directions: `repo delete` (annotated),
`release delete` and `repo deploy-key delete` (not annotated) must all state it, and the test
asserts NOTES presence separately so a regression to reading notes fails rather than passes;
`issue create`, `repo create`, `secret set` and `issue list` must not claim a confirmation they
never demand.

Per the captain's decision the ten scoped bundle notes are unchanged and no note was added to the
other 163 commands. `internal/connectors/guide.go` was deliberately left alone: adding the line
there would mark all 183 actions in every one of the 551 committed connector manuals, which is
both the bulk change the captain declined and unrelated-connector churn this branch must not make.

No test was weakened, skipped, or deleted. Cycle 8 added two tests.

## Cycle 9 — review round: one source for the typed confirmation

Cycle 8 made the confirmation derivable at point of use. Ten commands then stated the same fact
twice on one help page, three lines apart: a derived `CONFIRMATION` section and a hand-authored
`NOTES` string. The captain's decision: the derived help is the single source; delete the notes,
but only after proving the derivation carries every fact each note carried.

**The proof, before deleting anything.** All ten notes are one byte-identical string,
`destructive; requires typed confirmation --confirm destructive`, carrying two facts: the
classification (`destructive`) and the requirement (`--confirm destructive`). The derived line
states both — the closed challenge kind it names *is* the classification — and it was confirmed
present on all ten against the built binary before the removal:

```
repo delete            | CONFIRMATION: execution requires the typed confirmation --confirm destructive
repo archive           | CONFIRMATION: execution requires the typed confirmation --confirm destructive
repo unarchive         | CONFIRMATION: execution requires the typed confirmation --confirm destructive
cache delete           | CONFIRMATION: execution requires the typed confirmation --confirm destructive
secret delete          | CONFIRMATION: execution requires the typed confirmation --confirm destructive
repo delete-2          | CONFIRMATION: execution requires the typed confirmation --confirm destructive
actions logs delete    | CONFIRMATION: execution requires the typed confirmation --confirm destructive
branches rename create | CONFIRMATION: execution requires the typed confirmation --confirm destructive
deployments delete     | CONFIRMATION: execution requires the typed confirmation --confirm destructive
transfer create        | CONFIRMATION: execution requires the typed confirmation --confirm destructive
```

No note carried anything the derivation lacks, so nothing had to be extended first and no patch
was retained. The ten `notes` fields are removed; the 107 github commands whose notes state
something genuinely per-command keep them, including `issue delete`'s explanation of why it is
not exposed as a connector write.

**Red 9a — nothing stopped the copy coming back.** The removal is only durable if a future edit
cannot reintroduce a note that restates the gate. **Green** —
`TestGitHubNotesDoNotRestateTheTypedConfirmation` sweeps every github command, not just the ten,
and fails on any `notes` mentioning `--confirm`; it also asserts the resolver still answers for at
least one command, so the test cannot pass by the derivation silently breaking.

**Red 9b — the help test asserted the notes were present.** Cycle 8's test pinned
`NOTES` presence for `repo delete` and `repo archive`, which is exactly the duplication being
removed. **Green** — `TestConnectorCommandHelpStatesTheTypedConfirmation` now asserts the opposite
for six destructive commands, two of which used to carry a note: the derived section must state
the requirement and no note may restate it. The negative case is unchanged — `issue create`,
`repo create`, `secret set` and `issue list` must claim no confirmation.

Derived artifacts regenerated from the declarations: `docs/connectors/github/{MANUAL,SKILL}.md`
and `docs/skills/pm-github/SKILL.md` each drop the ten `; notes: destructive; requires typed
confirmation --confirm destructive` suffixes, and `website/data/connectors.generated.json` plus
`website/lib/connectors.catalog.data.generated.json` drop the ten `notes` values.
`docs/connectors/catalog/all-connectors.json` is unchanged and correctly so: it carries write
actions, not `cli_surface` commands, so it never held these strings — its `confirm: destructive`
on all 183 destructive actions is untouched. Both website catalogs were diffed per connector after
regeneration: GitHub is the only changed entry. The generators again wanted the pre-existing
`main` drift across the other 550 connectors, `gorgias` and `warehouse`; reverted as before.

No test was weakened, skipped, or deleted. Cycle 9 added one test and inverted one assertion that
pinned the duplication itself.

## Cycle 10 — the declared contract had to match the enforced one

The review round asked for four corrections, all of them the same shape: a place where what the
surface *says* about a write's gate does not match what the runtime *does* at that gate. A gate
described as absent when it is present is the dangerous direction; a gate described as present
when it is absent is the other one, and it is not harmless — it teaches an operator a step that
does nothing and hides which commands the real step is protecting.

**Red 10a — a safe preview announced a confirmation nothing would ask for.**
`engine.PreviewPreparedWrite` stamped `Confirmation: destructive` on *every* prepared write's
approval target, while `GateDestructiveExecution` consults evidence only when
`DestructiveTarget.RequiresApproval()` is true. So `pm github repo create --preview --json`
published `approval_target.confirmation = "destructive"` for an approval-only create, and
`--confirm destructive` on that plan is rejected as invalid — the preview described a gate that
does not exist.

New `TestPreparedWritePreviewDeclaresConfirmationOnlyWhereTheGateDemandsIt` asserts the preview
and the gate agree, over four targets. Red on the two safe ones:

```
--- FAIL: .../approval-only_create
    preview approval target confirmation = "destructive", want ""; the gate does not demand one
--- FAIL: .../update
    preview approval target confirmation = "destructive", want ""; the gate does not demand one
```

**Green** — the declaration is keyed on `prepared.Target.RequiresApproval()`, the same predicate
the gate uses, so the two cannot describe different contracts. DELETE and a declared
`confirm: destructive` on a non-delete method both still carry it, which is what
`validateApprovalTarget` requires before `IssueWriteGrant` mints any evidence.

**Red 10b — one caller minted a grant for writes nothing was holding back.** The narrowing broke
`internal/connectors/conformance`, and the failure was precisely diagnostic: for
`zendesk-sunshine`, `write_request_shape:delete_object_record` and `:delete_object_type` passed
while `:create_object_record`, `:upsert_object_record_by_external_id` and
`:create_relationship_record` failed with `write approval target requires destructive
confirmation`. `approvedFixtureWriteRequest` requested a grant for every action with a hard-coded
destructive confirmation. **Green** — it returns the request unchanged when the preview declares
no confirmation, because a safe action reaches the same dispatch seam with no evidence; asking
the authority to authorize a write nothing is stopping was always the wrong request. No
conformance check was relaxed: the destructive ones still mint, verify and consume a real grant.

**Red 10c — the approval metadata claimed every write requires a preview.** All 525 github
`reverse_etl` commands carried one blanket sentence, `Reverse ETL writes require plan, preview,
approval, execute.` It is true of the 176 that resolve a typed confirmation and false of the other
349: `PlanConnectorCommand` mints the approval token at plan time for a command with no
confirmation and no bound operation, and `RunReverseETL`'s `planRequiresPersistedPreview` is false
for exactly that command.

New `TestGitHubApprovalTextMatchesTheEnforcedWriteContract` classifies each command through the
same `ConfirmationChallengeForCommand` resolver the help uses and requires the matching sentence.
Red on exactly the 349 safe commands:

```
github "issue create" approval = "Reverse ETL writes require plan, preview, approval, execute.",
  want "Reverse ETL writes require plan, approval, execute; preview is optional."
... 349 such failures, 0 on the 176 destructive commands
```

**Green** — the 349 safe commands now declare `Reverse ETL writes require plan, approval, execute;
preview is optional.` The 176 destructive ones are unchanged, because the original sentence is
accurate for them. `scripts/gen-github-parity.py` derives the sentence from `command_approval`
rather than emitting one literal, so a regeneration cannot restore the blanket claim.

Proven against the built binary, as an A/B on the two classes:

```
$ pm github repo create ... --name demo-repo
  Created connector command plan rplan_37e5... for repo create
  Approval token: 0fe836e9...                        <- issued at plan; no preview step
$ pm github repo create ... --plan rplan_37e5... --approve 0fe836e9...
  Reverse ETL run rrun_2f2b... completed: succeeded=1 failed=0
  mock GitHub received: POST /user/repos BODY={"name":"demo-repo"}

$ pm github repo delete ...
  Preview required before an approval token is issued.
  Confirmation required: --confirm destructive
$ pm github repo delete ... --plan rplan_ad67... --approve <any>
  error: reverse plan "rplan_ad67..." must be previewed before approval
  requests dispatched: 0
```

The destructive gate itself is unchanged end to end — preview issues the token, approval without
`--confirm` is refused with 0 requests dispatched, and `--confirm destructive` dispatches exactly
one DELETE.

**Red 10d — the marker naming a flag that does not exist could still come back.** Cycle 7 removed
`destructive; requires --allow-destructive + typed confirmation` from the bundle, but
`scripts/gen-github-parity.py` — the generator that put it there — still emitted it for every
`destructive_action`, so the next regeneration would document the nonexistent flag again on every
newly covered endpoint. **Green** — the generator emits no `notes` marker at all; the confirmation
is derived once by `ConfirmationChallengeForCommand` and rendered by the help/manual/skill
CONFIRMATION field, which `TestGitHubNotesDoNotRestateTheTypedConfirmation` already forbids any
bundle from duplicating.

**Not red, and recorded as such: `repo delete-2` already carried `repo delete`'s contract.** Both
bind write action `repo`, so both resolve `destructive` through the same resolver and both render
the same CONFIRMATION block. The gap was that nothing pinned it, which is how the asymmetry this
whole phase exists to fix arises in the first place.
`TestGitHubGeneratedTwinsShareTheirAliasWriteContract` discovers every write action reachable
under two command names from the bundle rather than from a list, and requires matching
confirmation, availability and approval; it then pins the `repo delete` / `repo delete-2` pair by
name. It passed on first run — reported as verified, not as a fix. `risk` is deliberately excluded:
github states it as prose on the aliases and as a level on the twins, and two names for one
endpoint may honestly describe different callers' exposure.

Derived artifacts regenerated from the declarations: `docs/connectors/github/{MANUAL,SKILL}.md`,
`docs/skills/pm-github/SKILL.md`, `website/data/connectors.generated.json` and
`website/lib/connectors.catalog.data.generated.json`. Each carries exactly 349 changed lines and
nothing else. The generators again wanted the pre-existing `main` drift across the other 1031 doc
files; reverted as before.

No test was weakened, skipped, or deleted. Cycle 10 added three tests.

## Cycle 11 — `repo archive`/`repo unarchive` were gated against their own classification

Cycle 10 closed every place the *declared* contract disagreed with the *enforced* one except the
one it created a decision about rather than resolved. Cycle 6 gave `repo archive` and
`repo unarchive` `confirm: destructive`, reading "destructive" as a mutation class. The captain's
decision uses it to name the gate, and lists both commands with `repo create` and `secret set` as
non-destructive. Cycle 10 flagged the disagreement and left the gate on, on the reasoning that
removing a working gate on an ambiguous reading is worse than leaving one that is too strict. The
captain has now resolved the ambiguity in the other direction, so this cycle removes it.

The direction matters and is worth stating plainly: this **removes a gate**. What makes that
correct is not that the gate was useless but that it was never authorized. `pm github repo archive`
demanded a typed confirmation no decision granted it, and every command that carries the typed
confirmation without being on the destructive list makes the marker mean less on the ones that are.

**Red 11a — the runtime resolved a typed challenge on an approval-only write.**
`TestGitHubRestoredCommandsAreExecutable` already classifies each restored command and asserts
against `connectors.ConfirmationForWriteAction`, the same resolver the runtime uses; the two rows
were flipped to non-destructive:

```
--- FAIL: TestGitHubRestoredCommandsAreExecutable/repo_archive
    github "repo archive" reaches PATCH /repos/{{ config.owner }}/{{ config.repo }} with
    confirmation "destructive"; an approval-only write must not gain a typed challenge
--- FAIL: TestGitHubRestoredCommandsAreExecutable/repo_unarchive
    (identical)
```

**Red 11b — the whole approval path was pinned to the destructive shape.**
`internal/app/github_repo_delete_gate_test.go` asserted plan-with-no-token, preview-first and
`--confirm`-or-refusal for both commands. Rewritten as
`TestGitHubRepoArchiveIsApprovalOnlyAndSendsPinnedField`, which asserts the safe contract
end to end — plan mints the token, the preview publishes an empty confirmation, an unapproved run
is still refused with zero requests dispatched, and the plan-time token executes with no preview
and no `--confirm`:

```
--- FAIL: TestGitHubRepoArchiveIsApprovalOnlyAndSendsPinnedField/archive
    repo archive ConfirmationChallenge = "destructive", want none for an approval-only write
--- FAIL: .../unarchive   (identical)
```

**Red 11c — the published help demanded the flag.** `repo archive`/`repo unarchive` moved from
`TestConnectorCommandHelpStatesTheTypedConfirmation` to
`TestConnectorCommandHelpOmitsConfirmationWhereNoneIsDemanded`. Red, from `pm github repo
unarchive --help` as shipped:

```
APPROVAL
  Reverse ETL writes require plan, preview, approval, execute.

CONFIRMATION
  execution requires the typed confirmation --confirm destructive
```

`repo delete-2` took the vacated slot in the states-it test, so the generated twin of the one
command that *is* destructive is now covered there by name as well as by
`TestGitHubGeneratedTwinsShareTheirAliasWriteContract`.

**Green — two declarations, and nothing else.** `confirm: destructive` was deleted from
`archive_repo` and `unarchive_repo` in `internal/connectors/defs/github/writes.json`. Everything
downstream is derived and followed: the approval sentence became the safe one on both commands
(github's reverse-ETL split moves 176/349 → **174 destructive / 351 safe**, total unchanged at
525), the help drops the CONFIRMATION section, and `PreviewPreparedWrite` publishes an empty
confirmation kind because cycle 10 keyed that on `RequiresApproval()` rather than on a constant.

**What did not change, deliberately.** `repo delete` keeps its full contract — no token at plan,
preview required, typed `--confirm destructive`, single-use request-bound grant — and its tests are
untouched. `issue delete`, `issue transfer` and `pr revert` stay blocked: GitHub documents no REST
endpoint for any of them and there is no GraphQL mutation executor, so they remain unreachable
rather than confirmable. The `WriteRecordHook` that pins `archived` stays exactly as cycle 6 built
it: it was introduced *because* the actions were destructive, but its real value is that preview
and execution build one body from one record, which is the reason a `repo unarchive` cannot
silently forward an empty body. Only the stale rationale in its comments was corrected.

Derived artifacts regenerated: `docs/connectors/github/{MANUAL,SKILL}.md`,
`docs/skills/pm-github/SKILL.md`, `docs/connectors/catalog/all-connectors.json`,
`website/data/connectors.generated.json`, `website/lib/connectors.catalog.data.generated.json` and
`internal/cli/testdata/golden_transcripts.json` — two changed lines each and nothing else.
`TestGoldenTranscripts/connectors_inspect_github_json` caught the last of those on its own before
the regeneration, which is the artifact gate working as intended. The generators again wanted the
pre-existing `main` drift across the other 1031 doc files; reverted as in cycles 7 and 10.

No test was weakened, skipped, or deleted. Cycle 11 rewrote three assertions to the classification
the decision states and moved one help case between the two converse tests.

---

## Cycle 12 — GitHub rate-limit declaration and live-proof harness

**Red 12a — GitHub shipped no provider-cited active policy.** The new GitHub-focused engine test
loads the production embedded bundle, not a hand-made fixture, and asks the existing
`Runtime.RequesterFor` resolver for every intended auth scope. Before any declaration, spec, or
embed change it failed exactly because GitHub has no `rate_limits.json`:

```
$ go test -timeout 20m ./internal/connectors/engine/ -run TestGitHubDeclaredRateLimits -count=1
--- FAIL: TestGitHubDeclaredRateLimits (0.05s)
    github_rate_limits_test.go:20: GitHub has no rate_limits.json declaration
FAIL
```

The test requires declared policies for authenticated-user, GitHub App installation, GitHub
Actions token, and unauthenticated traffic, provider source/date, a non-secret scope for each,
the documented primary hourly capacity, and both admission/observation hooks. It intentionally
does not permit a raw token scope, an `unknown` placeholder, or a no-op attachment.

**Green 12a — declaration through the existing requester.** `rate_limits.json` now contains four
GitHub-documentation-cited policies, all retrieved `2026-08-08`:

| policy | selector | opaque non-secret subject | primary budget |
| --- | --- | --- | --- |
| `authenticated-user` | `token`, `oauth` | `rate_limit_account` | 5,000 requests/hour |
| `app-installation` | `github_app` | `installation_id` | 5,000 requests/hour minimum |
| `actions-token` | `github_token` | `rate_limit_repository` | 1,000 requests/hour per repository |
| `unauthenticated` | `public`, `none`, `anonymous`, `unauthenticated` | `rate_limit_ip` | 60 requests/hour per originating IP |

Each also carries the documented 900-point/minute REST secondary ceiling as a conservative
five-points-per-request sliding budget. GitHub leaves some REST point costs unpublished, so that
strict client accounting can only stop before the provider's published ceiling; it never claims an
unpublished cost is known. The normal response parser still tightens the policy from
`x-ratelimit-*` and `retry-after` without retaining raw headers.

No limiter, request path, registry, or credential-derived key was added. The first production
declaration only adds `*/rate_limits.json` to `defs.FS`; `Runtime.RequesterFor` and its existing
`Admit`/`Observe` wiring remain the sole mechanism.

```
$ go test -timeout 20m ./internal/connectors/engine/ -run 'TestGitHubDeclaredRateLimits|TestProductionDefinitionsEmbedEveryRateLimitDeclaration' -count=1
ok   polymetrics.ai/internal/connectors/engine  0.794s

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 551 connector(s) checked, 0 findings

$ go run ./cmd/connectorgen surface-sync --check
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)
```

---

## Cycle 29 — fixed-contract GraphQL operation runtime

**Red 29a — a GraphQL root has no typed direct-operation contract.**
The new loopback-only engine tests name the missing contract rather than using a generic GraphQL
escape hatch: a fixed query/mutation document, connector-relative `/graphql` endpoint, positive
response cap, closed variable schema, declared cursor map, partial-data/error metadata, rate-limit
metadata, and the existing single-use destructive approval gate. They also require unrecognised
variables, raw query overrides, and numeric page navigation to be rejected before a request. Before
the runtime additions, the test is intentionally a compile-time red: `GraphQLOperationSpec` cannot
describe a typed executable operation and neither result type can state partial GraphQL data.

```
$ go test -timeout 20m ./internal/connectors/engine -run 'TestOperationDirect(ReadExecutesFixedGraphQLQueryAndPreservesPartialData|ReadRejectsUntypedGraphQLInputsBeforeNetwork|WriteUsesSharedApprovalForFixedGraphQLMutation|WriteFailsClosedOnGraphQLErrors)' -count=1
# polymetrics.ai/internal/connectors/engine [polymetrics.ai/internal/connectors/engine.test]
internal/connectors/engine/graphql_operation_test.go:50:5: unknown field Path in struct literal of type GraphQLOperationSpec
internal/connectors/engine/graphql_operation_test.go:51:5: unknown field MaxBytes in struct literal of type GraphQLOperationSpec
internal/connectors/engine/graphql_operation_test.go:52:5: unknown field VariablesSchema in struct literal of type GraphQLOperationSpec
internal/connectors/engine/graphql_operation_test.go:62:18: undefined: GraphQLOperationPaginationSpec
internal/connectors/engine/graphql_operation_test.go:139:12: result.GraphQL undefined (type connectors.DirectReadResult has no field or method GraphQL)
internal/connectors/engine/graphql_operation_test.go:259:12: result.GraphQL undefined (type connectors.OperationDirectWriteResult has no field or method GraphQL)
FAIL	polymetrics.ai/internal/connectors/engine [build failed]
```

The green implementation must remain provider-neutral and local-test-only: no caller may supply a
document, selection, endpoint, or raw GraphQL transport; queries may retain bounded `data` with
bounded redacted errors; mutations must fail closed on any `errors[]` after the exact preview,
approval, and destructive-confirmation gate.

**Green 29 — fixed documents become bounded, typed operations rather than a raw transport.**
`graphql_operation.go` admits an executable GraphQL declaration only when it has a fixed named
query/mutation document, a rooted canonical connector-relative path, a positive response cap, and
a closed recursively bounded root variables schema whose properties are referenced by that fixed
document. The caller cannot provide a document, selection, endpoint, raw body, query override,
or a numeric page. Cursor navigation enters only through `--page-cursor` and a declared connection
map; the response reports bounded partial-data/error/rate-limit metadata and applies the established
redaction policy to `data`.

`graphql_mutation` now prepares one literal `POST` payload through the existing prepared-write
digest, rate limiter, no-retry transport policy, approval evidence, and typed destructive
confirmation. A `data: null`, malformed GraphQL envelope, or any GraphQL `errors[]` makes the
approved mutation fail rather than report completion. The command-runner received the matching
no-network direct-write preflight seam: its one API surface method/path/output-policy tuple must
match the operation's declaration exactly. This is a physical `POST /graphql` binding, not an
invented `GRAPHQL` verb or a generic endpoint selector.

The runtime read-endpoint ledger now carries a GraphQL operation ID in addition to `POST /graphql`,
so sharing the one transport endpoint cannot authorize a different fixed query. Legacy
`variables_path` metadata remains schema-compatible but is deliberately ignored by the executable
path. `TestPreflightOperationDirectReadValidatesDeclaredContract` changed its old unsupported-kind
fixture from `graphql_query` to `graphql_mutation`: queries are now intentionally supported by the
direct-read executor, while mutations remain write-only. The test retains the same unsupported-read
boundary; no test was removed, skipped, or weakened.

```text
$ go test -timeout 20m ./internal/connectors/engine -run 'Test(OperationDirect(ReadExecutesFixedGraphQLQueryAndPreservesPartialData|ReadRejectsUntypedGraphQLInputsBeforeNetwork|GraphQLRejectsUnboundVariableSchemaBeforeNetwork|WriteUsesSharedApprovalForFixedGraphQLMutation|WriteFailsClosedOnGraphQLErrors|WriteFailsClosedOnMissingGraphQLData)|PreflightOperationDirectWriteRequiresExactFixedGraphQLBinding)' -count=1
ok   polymetrics.ai/internal/connectors/engine

$ go test -timeout 20m ./internal/connectors/commandrunner -run 'TestPreflightOperationDirectWrite(RejectsMismatchedOperationPolicy|RequiresRuntimeBinding)' -count=1
ok   polymetrics.ai/internal/connectors/commandrunner

$ go test -timeout 20m ./internal/app -run 'TestGitHubDeployKeyDeleteDoesNotMaskNotFoundForAVisibleBoundFixture|TestGitHubLabelDeleteKeepsNotFoundVisibleForTheSameScopedWritePath' -count=1
ok   polymetrics.ai/internal/app

$ node --test scripts/tests/github-live-lab.test.mjs
23 passed
$ node --test scripts/tests/github-live-proof-sweep.test.mjs
7 passed
$ node scripts/github-live-lab-manifest.mjs --check
github live lab manifest: rows=957 personal_repo=427 sandbox_org_free=291 github_app_or_marketplace=33 unavailable_entitlement=206
$ node scripts/github-live-lab.mjs --check-boundary --boundary .planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-BOUNDARY.json
github live lab boundary: ok allowed_targets=1
```

**Red 29c — the full local app suite exposed a rate-limit fixture drift.** The first full run
failed only in `TestGithubPullRequestsETLSupportsAllSyncModes`: its loopback GitHub credential
selects the declared `authenticated-user` policy but omitted the required non-secret
`rate_limit_account` subject key. Each of the five pre-existing sync-mode assertions stopped before
its ETL request with `rate-limit policy "authenticated-user" requires non-secret config
"rate_limit_account" for its declared scope`.

**Green 29c — retain the policy and complete the fixture contract.** The test credential now
provides the fixed synthetic cohort `rate_limit_account: fixture-account`; no secret, provider URL,
or test assertion changed. The failing five-mode test and the full app package pass, alongside the
complete local GraphQL and CLI gates:

```text
$ go test -timeout 20m ./internal/app -run '^TestGithubPullRequestsETLSupportsAllSyncModes$' -count=1
ok   polymetrics.ai/internal/app
$ go test -timeout 20m ./internal/app -count=1
ok   polymetrics.ai/internal/app  223.438s
$ go test -timeout 20m ./internal/connectors/engine -count=1
ok   polymetrics.ai/internal/connectors/engine  4.988s
$ go test -timeout 20m ./internal/connectors/commandrunner -count=1
ok   polymetrics.ai/internal/connectors/commandrunner  17.952s
$ go vet ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/app ./cmd/connectorgen
$ go build ./cmd/pm
$ go test -timeout 20m ./cmd/connectorgen -count=1
ok   polymetrics.ai/cmd/connectorgen
$ go test -timeout 20m ./internal/cli -count=1
ok   polymetrics.ai/internal/cli  563.781s
```

No PM, GitHub, browser, or provider invocation occurred in Cycle 29. The deploy-key regression,
PM-only target boundary, immutable read-back, and cleanup-accounting tests above are the gate that
must stay green before a later explicit live cohort resumes.

---

## Cycle 18 — PM-only 957-case live-lab boundary

**Red 18a — the captain-required lab manifest and fail-closed fixture boundary did not exist.**
The new deterministic test imports the GitHub-only lab module and manifest generator, then requires
all 957 preserved pre-skipped cases to be represented exactly once in one of four factual cohorts.
It also requires PM-only commands, immutable slug-and-ID matching, explicit production/worktree
denial, pre-dispatch rejection, append-only/idempotent cleanup, redaction, and one terminal result
per command. No credential or provider request is involved in this red test.

```
$ node --test scripts/tests/github-live-lab.test.mjs
Error [ERR_MODULE_NOT_FOUND]: Cannot find module
'.../scripts/github-live-lab.mjs' imported from
'.../scripts/tests/github-live-lab.test.mjs'
✖ scripts/tests/github-live-lab.test.mjs
```

The green implementation must stay in GitHub-specific scripts/evidence, must not introduce a raw
GitHub escape hatch, and must validate the boundary before credential inspection or a PM subprocess.

**Red 18b — a fresh private-lab bootstrap had no account-identity constraint.**
The retained repository is recorded as retained after validation but has no preserved evidence that
this proof program created it through PM. The follow-up test therefore defines the only permitted
bootstrap exception: one exact authenticated-user immutable ID may invoke only `pm github repo
create` for one exact lab-prefixed private, auto-initialized repository. Before the implementation,
the test could not import `authorizeBootstrapRepoCreate` or the pre-dispatch executor:

```
$ node --test scripts/tests/github-live-lab.test.mjs
SyntaxError: The requested module '../github-live-lab.mjs' does not provide an export named
'authorizeBootstrapRepoCreate'
```

The green path must reject a different command, user ID, name, or public repository before PM starts.

**Red 18c — no reusable PM plan/preview/approval/execute runner existed for the bootstrap.**
The deterministic lifecycle test supplies a fake PM process and expects exactly three invocations:
plan with the named credential, preview with the created plan, then JSON execution with the transient
approval grant. It asserts the returned summary has only command/status/count fields and cannot
contain the process-only grant. Before implementation the import failed:

```
$ node --test scripts/tests/github-live-lab.test.mjs
SyntaxError: The requested module '../github-live-lab.mjs' does not provide an export named
'runPMPlannedWrite'
```

The real runner must keep stdout/stderr, plan IDs, approval grants, and confirmation challenges in
memory, redact all failure output, and execute only after the bootstrap boundary accepts the exact
private repository request.

---

## Cycle 19 — credential-pinned repo-view control and PM-only bootstrap discovery

**Red 19a — the malformed `repo view` target override needed a durable regression boundary.**
The preserved 124-case runner's proven `repo view` invocation is exactly `pm github repo view
--credential <profile> --root <isolated-project> --json`; it has no `--owner`, `--repo`,
`--config`, or `--connection` override. During this recovery, the malformed owner/repo form exited
nonzero and the output sanitizer classified it as `flag_parse_rejected`; the failed invocation did
not produce a provider assertion and no raw output was retained. Reproducing the preserved control
returned `ConnectorCommandRead` with one record, but its stored credential scope was a different
repository, so it cannot bind the new fixture. The regression record must preserve both facts
without persisting the old repository, a credential name, or response data.

**Red 19b — no ID-safe PM-only bootstrap discovery existed.**
The captain selected the implemented `pm github repos list-for-authenticated-user` command as the
independent discovery read. The new test requires an exact private result whose `name`, owner login,
and owner immutable ID all match the bootstrap principal, then returns only the repository immutable
ID and the known slugs. It rejects public, wrong-user, duplicate, and `--config`-overridden results
before PM starts. Before implementation the test failed at the missing exports:

```
$ node --test scripts/tests/github-live-lab.test.mjs
SyntaxError: The requested module '../github-live-lab.mjs' does not provide an export named
'authorizeBootstrapRepoDiscovery'
```

The green path must run the authenticated-repository list with the existing credential-pinned PM
argument shape, filter the response only in memory, bind one exact immutable ID, and remove the
one-time bootstrap exception before any normal cohort fixture action.

**Green 18/19 — source-derived 957-row boundary, control regression, and first PM-only cohort result.**
`scripts/github-live-lab.mjs` now validates the default-deny boundary before a PM subprocess,
permits one exact private bootstrap create, and then permits one exact PM
`repos list-for-authenticated-user` discovery read only until it binds the generated private target
by both slug and immutable ID. `repo view` remains the historical credential-pinned control; the
owner/repo-flag failure is retained in `GITHUB-LIVE-LAB-DIVERGENCES.json` without raw output,
credential material, or a provider response.

The deterministic manifest remains exactly 957 rows in four exclusive cohorts. The first historical
pre-skip, `repo create`, has a new incremental `proven` terminal record only after its PM
plan/preview/approval/execute lifecycle completed and the independent PM authenticated-repository
listing found exactly one private generated slug under authenticated user immutable ID `6113982`.
The boundary now contains that one exact run-owned repository ID, its one-time bootstrap exception
is removed, and the append-only cleanup ledger carries `created` then `read_back` events for it.

```
$ node --test scripts/tests/github-live-lab.test.mjs
11 passed

$ node scripts/github-live-lab-manifest.mjs --check
github live lab manifest: rows=957 personal_repo=427 sandbox_org_free=291 github_app_or_marketplace=33 unavailable_entitlement=206

$ node scripts/github-live-lab.mjs --check-boundary --boundary .planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-BOUNDARY.json
github live lab boundary: ok allowed_targets=1

$ node --input-type=module  # validate append-only ledger and incremental terminal record
github_live_lab_increment=ok fixtures=1 terminal_records=1

$ node --test scripts/tests/github-live-proof-sweep.test.mjs
7 passed
```

No test was weakened, skipped, or deleted. The PM-only bootstrap write and discovery response were
held in process memory; evidence records only the terminal assertion, generated target identity, and
sanitized control divergence.

---

## Cycle 20 — ID-bound normal PM reads and writes

**Red 20 — a normal write could still reach PM after a boundary was supplied.**
`runPMPlannedWrite` originally guarded only the special bootstrap request. The new deterministic
test supplies the protected `polymetrics-ai/cli` identity, then a mismatched `--config repo`, and
asserts the injected PM runner has not started. Before the guard, the test observed `started=true`:

```
$ node --test scripts/tests/github-live-lab.test.mjs
✖ normal planned writes bind the exact repository target before any PM process starts
AssertionError: true !== false
```

**Red 20b — the independent read path could similarly bypass ID binding.**
The next test imports `runPMScopedRead`, supplies the protected target, and requires pre-dispatch
denial. Before implementation the import failed with no `runPMScopedRead` export.

**Green 20 — one target contract for every normal PM subprocess.**
Both normal writes and reads now call `authorizeLabTarget` before credential-name validation or a
subprocess. Repository operations require exactly one `--config owner` and `--config repo` matching
the allowed immutable target; duplicate or mismatched config and explicit owner/repo selectors are
rejected. These are existing PM connector config overlays, held only in the invocation/plan, not a
rewrite of the stored credential scope. Bootstrap remains the only exception and cannot be combined
with a normal target.

## Cycle 21 — issue-create neutralization rather than unsafe deletion

**Red 21a — the generated `issue create` recipe chose the wrong cleanup resource.**
The generic cleanup picker fell back to `repo delete` when it could not use `issue delete`, even
though issue deletion is `unsafe_or_disallowed`. The new manifest test requires every row to name a
cleanup strategy and specifically requires `issue create` to use `issue close` with
`neutralize_and_retain`; before the generator change `cleanup_strategy` was absent.

**Green 21 — the first reusable personal-repository fixture family is live and neutralized.**
The generator now prefers an implemented same-namespace delete and otherwise requires explicit
retention; `issue create` is explicitly paired with PM `issue close` and retained-state evidence.
After an ID-bound PM baseline read returned zero open issues and PM's `rate-limit get` preflight
succeeded, the lab ran `issue create` through plan/preview/approval/execute. An independent PM
issue-list read found exactly one generated issue; PM `issue close` then ran through the same
lifecycle, and an independent PM read confirmed it closed. The append-only ledger records
`created`, `read_back`, `neutralized`, and `retained`; the retention reason is the recorded
`unsafe_or_disallowed` issue-delete decision. `GITHUB-LIVE-LAB-REPORT.json` now has two `proven`
historical pre-skips: `repo create` and `issue create`.

```
$ node --test scripts/tests/github-live-lab.test.mjs
13 passed

$ node scripts/github-live-lab-manifest.mjs --check
github live lab manifest: rows=957 personal_repo=427 sandbox_org_free=291 github_app_or_marketplace=33 unavailable_entitlement=206

$ node scripts/github-live-lab.mjs --check-boundary --boundary .planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-BOUNDARY.json
github live lab boundary: ok allowed_targets=1

$ node --input-type=module  # validates ledger + incremental report
github_live_lab_increment=ok fixtures=2 terminal_records=2

$ node --test scripts/tests/github-live-proof-sweep.test.mjs
7 passed
```

## Cycle 22 — PM-only organization and App/Marketplace bootstrap probes

**Red 22 — no closed account-level PM probe or source-derived external bootstrap map existed.**
Before any organization/App provider request, the plan was extended with a dedicated bootstrap
sub-slice. The new deterministic test requires a fixed allowlist of account-level PM direct reads,
rejects `apps create-from-manifest` and repository selectors before its injected runner starts, and
requires a source-derived inventory for the organization and App/Marketplace cohorts. The initial
red run failed because neither the inventory module nor the account-probe exports existed:

```
$ node --test scripts/tests/github-live-lab.test.mjs
Error [ERR_MODULE_NOT_FOUND]: Cannot find module
'.../scripts/github-live-bootstrap-probes.mjs' imported from
'.../scripts/tests/github-live-lab.test.mjs'
```

The green implementation must never turn this into a generic account command executor. It may run
only the two named PM reads with the existing user credential, must keep response bodies and
credential material process-local, and must prove an organization delete cannot begin without a
run-owned immutable organization plus cleanup provenance. The inventory must count the current
manifest's 291 organization and 33 App/Marketplace cases from preserved source artifacts and
record an exact PM-surface or GitHub entitlement divergence rather than calling a UI, `gh`, or raw
API.

**Red 22b — the live report did not yet account for the two safe account-probe outcomes.**
After the fixed PM reads completed, the report test was strengthened to require a terminal
credential blocker for `apps get-authenticated`, a terminal proof for the successful Marketplace
user read, and response-body-free source/guard evidence for the organization and App bootstrap
gaps. It failed against the prior two-record report:

```
$ node --test scripts/tests/github-live-lab.test.mjs
✖ records personal-repository cohort results only after immutable target binding, read-back, and neutralization
AssertionError: Expected values to be strictly deep-equal:
  actual:   { terminal_records: 2 }
  expected: { terminal_records: 4 }
```

The green record must add only sanitized command/outcome/status facts. It must not serialize the
credential profile, response body, stdout/stderr, provider URLs, or a fabricated organization/App
fixture.

**Red 22c — the existing divergence ledger did not name the new external-boundary facts.**
The follow-up test requires four compact records: the missing PM organization-create command, the
missing PM App-manifest code issuer, the observed App-authentication 401, and the reachable
Marketplace-user read that still cannot bootstrap a fixture. Before those records were added, the
first lookup was absent:

```
$ node --test scripts/tests/github-live-lab.test.mjs
✖ records exact PM-surface and GitHub credential divergences without a provider fallback or retained account data
AssertionError: actual undefined; expected org-bootstrap-create-command-absent
```

The green ledger must preserve only the exact command, phase, classified outcome/status, and next
PM-only prerequisite. No raw help/provider output or account data belongs in the divergence ledger.

**Green 22 — external bootstrap boundary is proved before personal-cohort expansion resumes.**
`scripts/github-live-bootstrap-probes.mjs` now derives a checked-in, response-body-free inventory
from the current GitHub CLI/API surfaces and the 957-row manifest. It accounts for exactly 291
`sandbox_org_free` rows and 33 `github_app_or_marketplace` rows. The inventory proves that current
PM exposes implemented `orgs delete` for `DELETE /orgs/{org}`, but no registered command for either
`POST /user/orgs` or `POST /organizations`; deletion was therefore never planned or dispatched.
It also proves `apps create-from-manifest` consumes a required `--code` at
`POST /app-manifests/{code}/conversions`, while no PM command issues that conversion code. Both are
PM command-surface blockers, not a claim about whether GitHub would permit the action through a
forbidden fallback.

The lab runner now admits only two targetless account probes: PM
`apps get-authenticated` and PM `apps list-subscriptions-for-authenticated-user`. It forbids every
repository selector, `--config`, connection override, unlisted command, and write before the
subprocess starts. The user-credential App probe returned sanitized HTTP 401, producing one
terminal `credential_blocker` for the historical App-authentication row. The Marketplace-user read
returned sanitized HTTP 200, producing one terminal `proven` record for its historical direct-read
row; it did not create or inspect a listing/App/plan/installation fixture. The report now has four
terminal records total (three proven, one credential blocker), while the organization/App bootstrap
gaps remain explicitly non-dispatched.

```
$ node --test scripts/tests/github-live-lab.test.mjs
16 passed

$ node scripts/github-live-bootstrap-probes.mjs --check
github live bootstrap probes: organization_cases=291 app_marketplace_cases=33

$ node scripts/github-live-lab-manifest.mjs --check
github live lab manifest: rows=957 personal_repo=427 sandbox_org_free=291 github_app_or_marketplace=33 unavailable_entitlement=206

$ node scripts/github-live-lab.mjs --check-boundary --boundary .planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-BOUNDARY.json
github live lab boundary: ok allowed_targets=1

$ node --input-type=module  # validates boundary, append-only cleanup, terminal records, and source inventory
{"boundary_targets":1,"fixtures":2,"terminal_records":4,"organization_cases":291,"app_marketplace_cases":33}

$ node --test scripts/tests/github-live-proof-sweep.test.mjs
7 passed

$ git diff --check
```

## Cycle 23 — reversible personal-repository labels

**Red 23a — a label fixture had no closed PM read-back resolver.**
After the external-bootstrap audit completed, the next planned resource family is
`label create` → `label edit` → typed-confirmed `label delete` in the already
ID-bound private lab repository. The deterministic test requires a PM `label list`
envelope with exactly one generated label name and immutable label ID; it rejects a
wrong command, an absent label, a duplicate label, or an attempted create when the
name already exists. Before adding the resolver, the test failed at import time:

```
$ node --test scripts/tests/github-live-lab.test.mjs
SyntaxError: The requested module '../github-live-lab.mjs' does not provide an export named
'assertBoundLabLabelAbsent'
```

The green resolver must consume a PM response only in memory and return only the generated label
name plus immutable ID. The live slice must use the existing exact owner/repository config guard,
must not create a label if the baseline has one, and must record final absence after cleanup.

**Red 23b — the incremental report did not yet contain a complete label lifecycle.**
The report/ledger test was extended before any label mutation to require the three historical label
rows, a created/read-back/edit-read-back/cleanup-completed ledger sequence, and final terminal
accounting of seven records (six proven and one credential blocker). It failed against the prior
four-record report:

```
$ node --test scripts/tests/github-live-lab.test.mjs
✖ records personal-repository cohort results only after immutable target binding, read-back, and neutralization
AssertionError: actual { terminal_records: 4 }; expected { terminal_records: 7 }
```

The green report must be written only after the baseline/list/create/edit/delete/list sequence has
completed through PM. It may record the immutable generated label ID and sanitized lifecycle facts,
but never a credential profile or provider response body.

**Red 23c — label edit had existence proof but not a returned-data assertion.**
The resolver test was strengthened to require the PM list record's canonical label color after an
edit, while continuing to return only the generated name/immutable ID. Before the narrow assertion
existed, the import failed:

```
$ node --test scripts/tests/github-live-lab.test.mjs
SyntaxError: The requested module '../github-live-lab.mjs' does not provide an export named
'assertBoundLabLabelProperties'
```

The green assertion must compare only caller-provided non-secret expected properties and must never
return the provider record or serialize it to evidence. This makes the live edit proof a returned
data assertion rather than an exit-status claim.

**Green 23 — one fully reversible label family is live, asserted, and removed.**
The new label resolver accepts only a PM `ConnectorCommandRead` for `label list`, matches the
generated name in memory, and returns only that known name plus immutable provider ID. It refuses
an existing baseline, malformed/wrong-command input, zero/duplicate matches, or a non-matching
canonical color. The edit assertion therefore verifies returned PM data instead of inferring
success from process exit.

After a PM baseline established that the generated label name was absent in the immutable-bound
private lab repository, the lab ran `label create`, `label edit`, and typed-confirmed `label delete`
through PM plan → preview → approval → execute. PM list read-back found one immutable label ID after
create, then the same ID with the edited canonical color, and finally no generated label after
delete. The append-only ledger records `created`, two `read_back` events, and `cleanup_completed`;
no label/provider response body or credential profile was written. The three historical label rows
are terminally proven, taking the incremental report to seven records (six proven, one credential
blocker).

```
$ node --test scripts/tests/github-live-lab.test.mjs
17 passed

$ node scripts/github-live-lab-manifest.mjs --check
github live lab manifest: rows=957 personal_repo=427 sandbox_org_free=291 github_app_or_marketplace=33 unavailable_entitlement=206

$ node scripts/github-live-bootstrap-probes.mjs --check
github live bootstrap probes: organization_cases=291 app_marketplace_cases=33

$ node scripts/github-live-lab.mjs --check-boundary --boundary .planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-BOUNDARY.json
github live lab boundary: ok allowed_targets=1

$ node --input-type=module  # validates boundary, append-only cleanup, terminal records, and source inventory
{"boundary_targets":1,"fixtures":3,"terminal_records":7,"organization_cases":291,"app_marketplace_cases":33}

$ node --test scripts/tests/github-live-proof-sweep.test.mjs
7 passed

$ git diff --check
```

## Cycle 24 — editable issue and comment lifecycle

**Red 24a — no closed PM issue read-back resolver could prove edit/comment behavior.**
The next planned private-repository slice creates a fresh generated issue rather than reusing the
retained closed issue. Its new test requires a PM `issue list` envelope with exactly one generated
title, immutable node ID and issue number, expected returned body/state, and a minimum returned
comment count. It rejects wrong-command, absent, duplicate, or pre-existing generated issue
results. Before implementation the imports were absent:

```
$ node --test scripts/tests/github-live-lab.test.mjs
SyntaxError: The requested module '../github-live-lab.mjs' does not provide an export named
'assertBoundLabIssueAbsent'
```

The green resolver must retain no provider issue record. It may return only the known immutable ID
and issue number, so the edit/comment proof stays a returned-data assertion and the final close can
use the exact generated issue only.

**Red 24b — the report/cleanup ledger did not yet require the editable issue's full lifecycle.**
Before the provider write, the test was strengthened to require terminal records for `issue edit`
and `issue comment`, and a second issue fixture with create/read-back/edit-read-back/comment-read-back/
neutralize/retain events. It failed against the seven-record report:

```
$ node --test scripts/tests/github-live-lab.test.mjs
✖ records personal-repository cohort results only after immutable target binding, read-back, and neutralization
AssertionError: actual { terminal_records: 7 }; expected { terminal_records: 9 }
```

The green evidence must only follow PM create/edit/comment/close plus returned data assertions;
the comment body must remain process-local and the closed issue must be retained with the existing
issue-delete safety decision.

**Red 24c — a minimum comment count did not prove increase from a fresh issue baseline.**
The resolver test now demands an exact zero returned count after create and a minimum one after the
comment. Before supporting `expectedComments`, the mismatch was silently ignored:

```
$ node --test --test-name-pattern='issue read-back' scripts/tests/github-live-lab.test.mjs
AssertionError: Missing expected exception.
```

The green resolver must validate a caller-declared non-negative exact count without returning the
record, so the live comment proof establishes a strictly increased returned count.

**Red 24d — the PM-only lab runner dropped record flags after planning.**
`runPMPlannedWrite` sent generated `--config`, identity, and record flags only to the initial plan,
then invoked both preview and approval execution with just `--plan`. That cannot support the
runtime's deliberate withheld-field contract: `PreviewConnectorCommandPlan` and
`RunReverseETL` each reconstitute a field withheld from the persisted plan from the caller's flags.
The new deterministic issue-edit lifecycle test requires the exact safe record-argument sequence
on plan, preview, and execution, while keeping the credential on the initial plan only and the
approval grant process-local. It failed before the runner change:

```
$ node --test --test-name-pattern='planned write re-supplies' scripts/tests/github-live-lab.test.mjs
✖ planned write re-supplies the exact record flags for withheld-field preview and execution
AssertionError: actual [ '--root', '/tmp/github-live-lab-root' ]; expected [ '--config', 'owner=lab-owner', ... ]
```

The actual GitHub issue attempt stopped before edit/comment provider execution and was PM-closed
and retained under the existing issue-delete safety rule. This is a lab-harness precondition gap,
not evidence of provider behavior. The green step must preserve the same immutable target guard,
re-supply only `validateRecordArgs`-accepted values, and never serialize the temporary approval
grant, plan ID, or response body.

**Red 24e — immediate PM list read-back can race GitHub's accepted write visibility.**
The repaired PM-only `issue create` lifecycle returned one completed provider mutation, but the
first independent PM `issue list --state all` did not yet meet the exact generated title/body/open/
zero-comment assertion. A later PM-only read found exactly one record with those properties and a
stable immutable ID; the fixture was then PM-closed and retained before any edit or comment was
attempted. The evidence is therefore a read-after-write visibility race, not a credential,
repository-scope, implementation, or provider-write failure.

The new deterministic test requires a bounded six-attempt helper to retry *only* a successful PM
`issue list` envelope that fails the generated-record assertion; it must immediately propagate a
PM read/provider error rather than turning an authentication or entitlement failure into a delay.
It returns only immutable ID, issue number, and attempt count. Before implementation its import
failed:

```
$ node --test --test-name-pattern='PM issue read-back retries' scripts/tests/github-live-lab.test.mjs
SyntaxError: The requested module '../github-live-lab.mjs' does not provide an export named
'waitForBoundLabIssue'
```

The green helper uses no provider tool except the caller's `runPMScopedRead`, holds all envelopes
in memory, caps the visibility wait at five one-second intervals, and leaves provider/credential
errors un-retried for exact classification.

**Red 24f — recovered safety fixtures must remain visible to the final proof.**
The two interrupted generated-issue attempts were each independently PM-read, PM-closed, and
retained, so the append-only ledger now has three short create/read-back/neutralize/retain issue
lifecycles: the original `issue create` proof plus the two recovered safety fixtures. The final
edit/comment proof must add a fourth issue fixture with the longer six-event lifecycle; the test
now requires exactly that partition instead of silently assuming only two issue fixtures. Before
the final fresh lifecycle, the terminal report remains deliberately red:

```
$ node --test --test-name-pattern='records personal-repository cohort results' scripts/tests/github-live-lab.test.mjs
AssertionError: actual { terminal_records: 7 }; expected { terminal_records: 9 }
```

This is an evidence gate, not a request to fabricate result rows. The next provider mutation uses
a new generated title, the immutable-bound repository, the repaired record re-supply lifecycle,
and the bounded PM-only read-back helper; all prior generated issues are closed and retained.

**Green 24 — PM-only issue edit/comment proof with fail-closed visibility handling.**
`runPMPlannedWrite` now sends the same `validateRecordArgs`-accepted argument sequence to plan,
preview, and execution while keeping the credential on the initial plan and all approval material
process-local. `waitForBoundLabIssue` replays only successful PM `issue list` envelopes that have
not yet met the exact generated-record assertion; it returns no provider record, caps retries at
six, and immediately propagates a PM read failure. The deterministic regression coverage verifies
both the re-supply sequence and that a synthetic provider-status failure is never retried.

The final fresh, immutable-bound lab issue completed PM create, edit, comment, and close. Bounded
PM list read-back established the same immutable issue identity after edit, an exact returned
comment count of one after comment, and closed state after cleanup. The append-only ledger now has
three short retained issue lifecycles for prior safe fixtures plus one six-event edit/comment
lifecycle; no open generated issue remains. The report records both historical command rows as
proven only after that evidence.

```
$ node --test scripts/tests/github-live-lab.test.mjs
20 passed

$ node scripts/github-live-lab-manifest.mjs --check
github live lab manifest: rows=957 personal_repo=427 sandbox_org_free=291 github_app_or_marketplace=33 unavailable_entitlement=206

$ node scripts/github-live-bootstrap-probes.mjs --check
github live bootstrap probes: organization_cases=291 app_marketplace_cases=33

$ node scripts/github-live-lab.mjs --check-boundary --boundary .planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-BOUNDARY.json
github live lab boundary: ok allowed_targets=1

$ node --test scripts/tests/github-live-proof-sweep.test.mjs
7 passed

$ git diff --check
```

---

## Cycle 25 — disposable read-only deploy-key lifecycle

**Red 25a — a deploy-key list could not prove a safe, generated, read-only fixture.**
The next personal-repository family is the two historical rows `repo deploy-key add` and
`repo deploy-key delete`. A PM-only preflight proved `repo deploy-key list` returns a scoped
`ConnectorCommandRead` envelope in the immutable-bound private repository, but the lab had no
resolver that could bind a generated title to one immutable deploy-key ID while refusing to retain
the public key material. The new test requires that resolver plus a baseline-absence guard,
rejects wrong command/absent/duplicate/non-integer/non-read-only records, and asserts the returned
value contains no `key` field. Before implementation its import failed:

```
$ node --test --test-name-pattern='deploy-key read-back' scripts/tests/github-live-lab.test.mjs
SyntaxError: The requested module '../github-live-lab.mjs' does not provide an export named
'assertBoundLabDeployKeyAbsent'
```

The report/cleanup gate is deliberately strengthened before any provider write from nine to eleven
terminal records and requires a `deploy_key:<immutable-id>` lifecycle of `created`, `read_back`,
and `cleanup_completed`. The green path must generate the Ed25519 pair only in process memory,
pass only its public line to PM, set `--read-only`, redact/no-persist all key material, and use the
existing typed-confirmed delete contract before writing either terminal result.

**Red 25b — deploy-key visibility must not strand a credential fixture after an accepted write.**
The issue family already demonstrated that a successful GitHub write can briefly precede its PM
list visibility. The new deploy-key retry test applies the same narrow rule: retry a successful
PM `repo deploy-key list` envelope that does not yet contain the generated title, return only
immutable ID/title/attempt count, and never retry a PM provider-status failure. Before the helper
exists, the import is absent:

```
$ node --test --test-name-pattern='PM deploy-key read-back retries' scripts/tests/github-live-lab.test.mjs
SyntaxError: The requested module '../github-live-lab.mjs' does not provide an export named
'waitForBoundLabDeployKey'
```

The green helper is capped at six PM-only reads with five one-second visibility waits and retains
no public-key string from any envelope.

**Red 25c — typed-confirmed deletion needs the same final-absence visibility guard.**
The cleanup contract cannot treat a successful PM delete exit as proof that the public deploy key
is absent. A second deterministic test requires a bounded PM-only list retry until the generated
title is absent and returns only the attempt count; it likewise propagates a PM provider-status
failure immediately. Before implementation the absence helper import is absent:

```
$ node --test --test-name-pattern='PM deploy-key absence read-back' scripts/tests/github-live-lab.test.mjs
SyntaxError: The requested module '../github-live-lab.mjs' does not provide an export named
'waitForBoundLabDeployKeyAbsent'
```

The green path records `cleanup_completed` only after this assertion, never after exit status alone.

---

## Cycle 26 — deploy-key `404` false-success diagnosis gate

**Red 26 — a correctly addressed deploy-key delete can complete after an HTTP `404`.**
The preserved PM-only lab observed two typed-confirmed delete executions reported as completed,
then an independent scoped PM list returned the same generated immutable deploy key. Before any
third provider attempt, the regression reproduces the observable failure locally: a fake GitHub
server receives the expected `DELETE /repos/<owner>/<repo>/keys/<integer-id>` route for the
persisted scoped plan, replies `404`, and still exposes that key from a separate local list
fixture. The caller must receive a failed reverse run rather than a completed one.

This keeps trigger, masking condition, and symptom separate:

- **trigger under test:** the provider returns `404` to the exact typed-confirmed delete;
- **masking condition:** `delete_deploy_key.delete.missing_ok_status` contains `404`, so
  `executeApprovedWrite` increments `RecordsWritten` and suppresses the error;
- **visible symptom:** PM emits a completed reverse run while a later independent list can still
  find the same immutable fixture.

The paired label-delete control uses the same plan/preview/confirmation machinery and a correctly
scoped local path, but its `404` remains a failure. That disconfirms lost `--config` state,
`--key-id` mapping, and typed-confirmation bypass as explanations for the false completion.

Expected initial failure:

```text
--- FAIL: TestGitHubDeployKeyDeleteDoesNotMaskNotFoundForAVisibleBoundFixture
    ... RunReverseETL() unexpectedly completed after provider 404
```

No provider request, key material, credential value, repository slug, or provider body is used in
this red test. The green change must be source-owned/regenerated and preserve both the independent
absence read-back and generic delete semantics outside the demonstrated GitHub declaration.

**Green 26:** `internal/connectors/defs/github/writes.json` is the owned action declaration (the
migration conventions define `writes.json` as the declarative write source). The focused removal of
`delete_deploy_key.delete.missing_ok_status: [404]` restores the same provider-error visibility as
the label control without changing the generic engine. The historical comparison found the semantic
already removed by `48dac5782a` on a non-ancestor line; the full-surface import rooted at
`5bc7465b9` carried the older declaration, so this is a narrow source correction rather than a
runtime-wide counterfactual. `surface-sync --check` found zero derived-field drift.

```text
$ go test -timeout 20m ./internal/app -run 'TestGitHubDeployKeyDelete' -count=1
ok   polymetrics.ai/internal/app
$ go test -timeout 20m ./internal/connectors/engine -run 'TestWriteDelete' -count=1
ok   polymetrics.ai/internal/connectors/engine
```

## Cycle 27 — truthful deploy-key cleanup-failure accounting

**Red 27 — a false-success cleanup cannot be represented as a completed lifecycle.**
The existing report deliberately had nine terminal records because the deploy-key create/delete pair
was withheld until final absence. After the local diagnosis, that omission would hide a real command
result: creation plus immutable-ID read-back succeeded, while deletion is a factual failed operation
because two typed-confirmed PM executions reported completion and independent PM list read-back kept
the same generated key visible. The boundary/read-back test was strengthened to require eleven
terminal records with the add marked proven, the delete marked failed, and an append-only
`cleanup_failed` observation that does **not** terminalize the live fixture.

Expected initial failure:

```text
$ node --test scripts/tests/github-live-lab.test.mjs
... terminal_records: 9
... terminal_records: 11
```

**Green 27:** add `cleanup_failed` as a sanitized, nonterminal ledger observation; append the exact
historical failure fact without a third provider call; and make the report tally 11 = 9 proven + 1
failed + 1 credential blocker. The test continues to require final absence before a later delete can
be promoted to proven, so this records failure rather than weakening the cleanup gate.

```text
$ node --test scripts/tests/github-live-lab.test.mjs
23 passed
$ node --test scripts/tests/github-live-proof-sweep.test.mjs
7 passed
$ node scripts/github-live-lab-manifest.mjs --check
github live lab manifest: rows=957 personal_repo=427 sandbox_org_free=291 github_app_or_marketplace=33 unavailable_entitlement=206
$ node scripts/github-live-lab.mjs --check-boundary --boundary .planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-BOUNDARY.json
github live lab boundary: ok allowed_targets=1
$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 551 connector(s) checked, 0 findings
$ go run ./cmd/connectorgen surface-sync --check
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)
```

## Cycle 28 — combined pinned REST + GraphQL operation ledger

**Red 28a — no importer or source-derived GraphQL root inventory existed.** The preserved parity
surface counted only four fixed GraphQL documents while GitHub's official public schema exposes the
actual `Query` and `Mutation` roots. A new fixture-only test requires a source lock with REST
method/path rows, multiline SDL root-field parsing, schema hash/provenance, explicit
`createEnterpriseOrganization` canary, and one combined row per source operation. Before any
implementation it failed at module resolution:

```text
$ node --test scripts/tests/github-combined-operation-ledger.test.mjs
Error [ERR_MODULE_NOT_FOUND]: Cannot find module '.../scripts/github-combined-operation-ledger.mjs'
```

The official source snapshot used for the following green work is read-only and non-credentialed:
the pinned REST artifact at `github/rest-api-description` commit
`b26c240ded1c8b79cb0fb09dee4a21239061fa23` is 12,920,264 bytes with SHA-256
`80850db290cde4eb487e0efb587cf27f305e77b6bef96933ed8a09b5169d5b1d` and 1,220 method/path
operations. GitHub Docs' public `schema.docs.graphql` snapshot is 1,546,421 bytes with SHA-256
`c09aba9911b08d2aa8a022578edaf256aa040f38d7fb7196656356ea236c249d`; it contains 31 `Query`
and 274 `Mutation` root fields, including exactly one `createEnterpriseOrganization` mutation. A
naive indentation-only probe initially counted a `https` token inside a description as a Query
field; the SDL parser disconfirmed that false positive before the lock was accepted. No GitHub
provider operation or PM fixture call occurred.

**Red 28b — fixed GraphQL bindings were still being mistaken for the complete GraphQL source
denominator.** The new permanent test scans each active source-surface/proof gate and rejects the
legacy `GRAPHQL: 4` / `1224` denominator. It also derives the REST method split from the lock, so a
future source refresh cannot preserve a stale static count. The initial gate was red because the
four fixed-document bindings were the only GraphQL rows the old tests knew how to count.

**Red 28c — a disabled fixed mutation was classified as partially implemented.** The fixture adds
an `unsafe_or_disallowed` command bound to a GraphQL mutation and expects
`declared_not_executable`, an exact mapped-command blocker, and no claim that a GraphQL executor
exists. Before the classifier change it failed with:

```text
actual: 'partially_implemented'
expected: 'declared_not_executable'
```

**Green 28 — the generated source lock and combined ledger are hermetic, complete, and explicit.**
`scripts/github-combined-operation-ledger.mjs` parses the official artifacts only in `--write`
mode; `--check` rebuilds the ledger solely from checked-in lock/bundle data. It emits all 1,525
source operations, rejects duplicate/missing IDs, stale source hashes, the `UNTESTABLE` label, an
absent `createEnterpriseOrganization` canary, or a stale generated ledger. It distinguishes the
four existing fixed GraphQL documents from the complete root inventory, records `node` and `nodes`
as a fixed-projection matrix, and labels `deleteIssue` as
`mapped_command_not_executable` rather than a partial runtime implementation. No provider request
or fixture action occurred in this green step.

```text
$ node --test scripts/tests/github-combined-operation-ledger.test.mjs
5 passed
$ node scripts/github-combined-operation-ledger.mjs --check
github combined operation ledger: ok rest=1220 graphql_query=31 graphql_mutation=274 total=1525
$ go test -timeout 20m ./cmd/connectorgen -run 'TestGitHub(DocumentedRESTSurfaceIsComplete|APISurfaceOperationLedgerMetrics)$' -count=1
ok   polymetrics.ai/cmd/connectorgen
$ go test -timeout 20m ./internal/connectors/certify -run '^TestSurfaceInventoryForGitHubAccountsForAllReviewedEndpoints$' -count=1
ok   polymetrics.ai/internal/connectors/certify
$ node --test scripts/tests/github-parity-proof.test.mjs
6 passed
$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 551 connector(s) checked, 0 findings
$ go run ./cmd/connectorgen surface-sync --check
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)
$ git diff --check
```

## Cycle 17e — final generated-artifact, lint, and provider-boundary repair

**Red 17e — post-repair gates caught stale owned output and an out-of-scope proof file.** The
current `TestGoldenTranscripts` run rejected GitHub inspect/manual output that still described 555
write actions after the generated bundle reached 574. `make lint` also found a dead direct-write
helper and a non-idiomatic output-policy branch. `connectorgen boundary` rejected the exhaustive
provider-double implementation because its provider-specific Go lived in shared non-test scope.

**Green 17e — regenerate, simplify, and constrain the proof.** The golden fixture was regenerated
by the owning test command with `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1`; the implementation now
uses a tagged output-policy switch and removes the unused helper. The provider-double implementation
and contract test are test-only files, so GitHub-specific proof code remains outside shared
production Go. Current checks are green:

```
$ go test -count=1 -timeout 20m ./internal/connectors/conformance/ -run '^TestGitHubExhaustiveProviderDouble$'
ok  polymetrics.ai/internal/connectors/conformance
$ make lint
0 issues.
$ go run ./cmd/connectorgen boundary . --json
{"outcome":"clean","findings":0}
$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 551 connector(s) checked, 0 findings
$ go run ./cmd/connectorgen surface-sync --check
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)
```

---

## Cycle 16 completion — exhaustive provider, limiter, and terminal live accounting

**Red 16c — the exhaustive provider-double entry point was absent.** The first conformance
test failed to compile at `runGitHubExhaustiveProviderDouble`, so a source count could not be
mistaken for execution evidence. The implementation then exercised every declared stream,
write action, and operation through the existing engine paths, with controlled provider capture,
fixture-backed records where available, bounded synthetic records otherwise, and concrete
untestable rows for the GraphQL mutation, local workflow, and sensitive secret operation.

**Green 16c — deterministic provider-double proof is complete.** The report has 37 streams,
574 write actions, 377 operations, 988 rows, 985 exercised, 3 explicitly untestable, and 0
failed. It identifies the 23 streams without a command as `pm etl` generic routes and the 38
write actions without a command as `pm reverse` generic routes. Captured requests retain only
method/path/query-key/header-name/body-key and SHA-256 metadata; no body, credential, or token
value is persisted.

```
$ go test -timeout 20m ./internal/connectors/conformance/ -run TestGitHubExhaustiveProviderDouble -count=1
ok   polymetrics.ai/internal/connectors/conformance

$ jq '{streams,write_actions,operations,generic_streams,generic_write_actions,exercised,untestable,failed}' .planning/phases/github-parity-extract-r1/PROVIDER-DOUBLE-PROOF.json
{
  "streams": 37, "write_actions": 574, "operations": 377,
  "generic_streams": 23, "generic_write_actions": 38,
  "exercised": 985, "untestable": 3, "failed": 0
}
```

**Green 16d — the existing GitHub limiter has an executable provider-boundary proof.** The
GitHub-specific test uses the declared authenticated-user policy with a deterministic one-minute
fixture window, observes a non-secret reset/remaining signal, records one local wait before the
second same-scope request, and proves an independent scope does not inherit the wait. All three
provider-double requests return 200; no provider 429 is used as evidence of local pacing.

```
$ go test -timeout 20m ./internal/connectors/engine/ -run TestGitHubRateLimitAdmissionPrecedesProviderAndIsolatesScope -count=1
ok   polymetrics.ai/internal/connectors/engine
```

**Red 16e — current-head credentialed live acceptance cannot run in this isolated worktree.**
The live harness requires an approved private test repository and a credential scoped to it before
it starts `pm`; neither is available here. The harness therefore has an explicit external-blocker
mode that emits one terminal `untestable` record for every implemented command, never prints or
stores a credential value, and does not claim provider acceptance.

**Green 16e — terminal live accounting is complete but externally blocked.**
`LIVE-PROOF-REPORT.json` contains 1,081/1,081 implemented-command records with
`proven=0`, `untestable=1081`, and `failed=0`; the blocker is the unavailable approved private
credential/repository. A real credentialed run remains a merge gate, not a green substitute.

```
$ node scripts/github-live-proof-sweep.mjs --external-blocker --pm /tmp/pm-github-parity-current \
    --report .planning/phases/github-parity-extract-r1/LIVE-PROOF-REPORT.json \
    --reason 'approved private GitHub credential and dedicated test repository are unavailable in this isolated worktree'
github live proof: external blocker; untestable=1081
```

## Cycle 16 — exhaustive proof inventory and current-head routing evidence

**Red 16a — source-derived proof ledgers did not exist.** The captain-ordered proof contract
requires one row for every endpoint, operation, command, stream, and write action; a summary count
cannot reveal an omitted declaration or an invented binding. The new Node test initially failed
because `scripts/github-parity-proof.mjs` was absent. Its green implementation derives the model
from all five shipped GitHub definition files and validates the exact 1,224 / 37 / 574 / 377 /
1,179 source totals, including 23 generic-only streams and 38 generic-only write actions.

**Green 16a — mechanically generated ledgers account for the complete source bundle.**
`node scripts/github-parity-proof.mjs --write` emits `OPERATION-PROOF-LEDGER.json` and
`COMMAND-PROOF-LEDGER.json`; the validator rejects omitted rows, unknown bindings, and empty
source bundles. The attached command evidence is intentionally not inferred from availability:
it is populated only by the separate built-binary sweep.

```
$ node --test scripts/tests/github-parity-proof.test.mjs
6 passed
$ node scripts/github-parity-proof.mjs --write
{"endpoints":1224,"coveredEndpoints":1147,"blockedEndpoints":77,"streams":37,"writeActions":574,"operations":377,"commands":1179,"implementedCommands":1081,"partialCommands":37} generic_streams=23 generic_writes=38
```

**Red 16b — reachability could not be proven by static metadata.** The command proof test first
had no current-head executor. A fresh `/tmp/pm-github-parity-current` binary and eight isolated
initialized projects are now swept once per declared command; the classifier accepts only the
exact rendered `pm github <path>` name and stores no subprocess output. Before attaching that
report, the generated command ledger correctly showed `binary.state=not_run` rather than
repeating the stale 1,079/1,086 historical claim.

**Green 16b — the current built binary dispatches every declared GitHub command.** The sweep
recorded 1,179/1,179 exact command names, with 1,081 implemented, 37 partial, and all other
declared classifications dispatchable to their declared help surface. The command ledger was
regenerated with the report's binary hash and has no `not_run` rows.

**Red 16c — deterministic provider-double proof was not executable for the full declaration set.**
The new conformance test named the required 37-stream / 574-write-action / 377-operation proof,
but the exhaustive runner was absent:

```
$ go test -timeout 20m ./internal/connectors/conformance/ -run TestGitHubExhaustiveProviderDouble -count=1
undefined: runGitHubExhaustiveProviderDouble
```

The green step must execute every stream against the existing replay engine, every write action
against a controlled capture server (fixture-backed or a bounded synthetic record), and every
executable operation through its real engine executor, while recording blocked operation kinds
with concrete reasons and asserting no request on rejected safety paths.

## Cycle 16 — exhaustive source-derived proof ledgers

**Red 16a — the captain's required terminal proof ledgers did not exist.** The first contract test
was added before the proof generator. It failed at module resolution rather than silently accepting
an absent artifact:

```
$ node --test scripts/tests/github-parity-proof.test.mjs
Error [ERR_MODULE_NOT_FOUND]: Cannot find module
.../scripts/github-parity-proof.mjs
```

The test contract requires the source model to account for every 1,224 endpoint, 37 stream, 574
write-action, 377 operation, and 1,179 command row; it also rejects an omitted endpoint, unknown
covered_by target, and missing operation binding. The 23 stream and 38 write-action members with
no dedicated connector command must carry a generic ETL/reverse-ETL route.

**Green 16a — source-derived accounting is now executable.** `scripts/github-parity-proof.mjs`
loads the five GitHub source bundles, normalizes provider/config path templates, links streams,
write actions, operations, GraphQL operation names, aliases, and command API references, then
validates and writes `OPERATION-PROOF-LEDGER.json` and `COMMAND-PROOF-LEDGER.json`. It computes
counts from the loaded declarations and never accepts a hand-maintained summary.

```
$ node --test scripts/tests/github-parity-proof.test.mjs
ok — 5 tests passed
$ node scripts/github-parity-proof.mjs --check
{"endpoints":1224,"coveredEndpoints":1147,"blockedEndpoints":77,"streams":37,"writeActions":574,"operations":377,"commands":1179,"implementedCommands":1081,"partialCommands":37} generic_streams=23 generic_writes=38
```

No credential, provider request, raw body, or generated source bundle was changed by this cycle.

## Cycle 17 — approved current-head live acceptance and ref-path repair

**Red 17a — stream and binary live envelopes had no HTTP status field.** The live runner initially
treated those successful CLI envelopes as failures because `ConnectorCommandRead` reports returned
record counts and `ConnectorCommandBinaryDownload` reports bounded file accounting, but neither
claims a provider status. Requiring an invented status would make the evidence dishonest.

**Green 17a — returned-data assertions are sufficient for status-less envelope kinds.** The runner
now proves streams with `count` plus `records` and binary downloads with the bounded `record`
metadata; direct reads and write operations retain their real 2xx status where the runtime exposes
one. The report still rejects raw output, response bodies, grants, credentials, and token-shaped
values.

**Red 17b — the first approved credentialed sweep surfaced 92 terminal failures.** The failures
were not converted wholesale: sanitized status triage separated provider-state prerequisites and
credential-scope boundaries from invalid search queries, missing Git-ref path semantics, and the
write harness's plan-token handling. The case generator was corrected for provider-valid queries,
empty-repository content, App-only/account-scope resources, and concrete absent feature resources.
The write runner now obtains non-destructive plan grants from human-readable plan output while
retaining typed confirmation for destructive DELETE actions and performs approved read-backs.

**Red 17c — GitHub's `git ref view` could not reach `heads/main`.** The provider-valid value
`heads/main` was rejected as a single identifier, while `main` reached the API as the wrong ref and
returned no successful result:

```
pm github git ref view --ref heads/main --json
error: path variable ref contains invalid character '/'
```

**Green 17c — slash-bearing generic ref path variables are safely segmented.** The shared engine
now treats only the provider-neutral `ref` path variable like a path of validated segments,
preserving URL escaping and traversal rejection; ordinary identifiers retain their old validation.
The focused engine test proves `/git/ref/heads/main` reaches the fixture endpoint. This is the
only shared production change in this proof slice and is provider-neutral.

**Green 17d — final current-head credentialed acceptance is exact and zero-failure.** The final
rebuilt binary ran the complete case file against the captain-approved private repository and
credential metadata without persisting either secret. The report is tied to the binary and case
hashes and contains one terminal row for every implemented command:

```
$ node scripts/github-live-proof-sweep.mjs --pm /tmp/pm-github-parity-final \
    --root <isolated-live-project> --credential github-live-proof \
    --cases .planning/phases/github-parity-extract-r1/LIVE-PROOF-CASES.json \
    --report .planning/phases/github-parity-extract-r1/LIVE-PROOF-REPORT.json --execute-writes
github live proof: proven=124 untestable=957 failed=0
```

Validation confirms 1,081 unique rows, exact source command-set equality, `status=credentialed_live`,
matching current binary/surface/case hashes, no forbidden report fields, and no credential-shaped
text. `LIVE-RATE-LIMIT-PROOF.json` records the live `rate-limit get` 200 response, zero observed
429s across the bounded live workload, preserved headroom, and the existing same-scope admission /
independent-scope unit proof.

**Red 15c — the required post-rebase CLI gate found stale generated artifacts and an incomplete
test-only rate-limit scope.** The first full `internal/cli` baseline after the clean rebase fails
in two places. `TestGoldenTranscripts` correctly rejects the old GitHub inspect/manual snapshots:
they still describe 555 write actions and omit the newly generated namespaces. Separately,
`TestReverseETLToGitHubCreatesPullRequestAfterApproval` configures `auth_type=token` but omits the
declared non-secret `rate_limit_account`, so the active policy correctly refuses to issue the mock
request rather than inferring a coordination subject from a secret or repository target.

```
$ go test -timeout 20m ./internal/cli/
--- FAIL: TestGoldenTranscripts
    connectors_inspect_github_json: stdout mismatch (-want +got)
    dynamic_connector_bare_json: stdout mismatch (-want +got)
--- FAIL: TestReverseETLToGitHubCreatesPullRequestAfterApproval
    reverse run code = 1 ... rate-limit policy "authenticated-user" requires non-secret config
    "rate_limit_account" for its declared scope
FAIL    polymetrics.ai/internal/cli
```

The green repair must retain the fail-closed policy, add an opaque non-secret account label only to
the isolated test fixture, and regenerate the affected user-facing artifacts through their owning
commands. No credential, token-derived value, or synthetic default belongs in this repair.

**Red 12d — a live write case could replace the dedicated repository owner before `pm` started.**
The proof runner's credential metadata check bound the saved credential to the supplied test
repository, but case arguments were still accepted unchanged.  The new deterministic test provides
the implemented `repos create-using-template` write with `--owner
outside-the-dedicated-repository` and a fake `pm` that leaves a marker if invoked.  Before the
guard, validation reached that fake process instead of rejecting the case:

```
$ node --test scripts/tests/github-live-proof-sweep.test.mjs
✖ rejects a write case that overrides the dedicated repository owner before starting pm
  AssertionError: expected /dedicated repository owner/i
  actual: github live proof: named credential is not a GitHub credential
```

The fake credential error is the behavioral proof of the gap: the case made it past validation and
the runner had already started `pm`.  No credential, provider request, or real write is involved in
the test.

**Green 12d — reject target overrides before credential inspection.** The runner now resolves the
production command metadata while it validates the case file. For an executable `reverse_etl` or
`direct_write` case, any `--owner` or `--repo` argument is interpolated first and must equal the
dedicated repository identity supplied to the runner. The rule accepts both `--flag value` and
`--flag=value`; reads remain able to use separately approved public data. Validation occurs before
the credential inspection subprocess, so a rejected case cannot plan, preview, obtain a grant, or
dispatch a write.

```
$ node --test scripts/tests/github-live-proof-sweep.test.mjs
✔ rejects a write case that overrides the dedicated repository owner before starting pm
ℹ pass 5
ℹ fail 0

$ node scripts/github-live-proof-sweep.mjs --self-test
github live proof self-test: ok
```

**Red 12b — no committed live-proof accounting runner existed.** The deterministic Node test
defines the non-negotiable accounting contract before a runner exists: it enumerates only
`availability: implemented`, rejects an omitted command, rejects fixture coverage as a terminal
state, requires either a returned-data assertion or a concrete untestable reason, and proves that
raw subprocess output cannot enter a report record. With no runner at the specified path, the
first test run failed at module resolution:

```
$ node --test scripts/tests/github-live-proof-sweep.test.mjs
Error [ERR_MODULE_NOT_FOUND]: Cannot find module
'.../scripts/github-live-proof-sweep.mjs' imported from
'.../scripts/tests/github-live-proof-sweep.test.mjs'
✖ scripts/tests/github-live-proof-sweep.test.mjs
```

The fixture string is deliberately token-shaped but non-secret; the assertion rejects it from the
serialized record rather than putting a credential in a test fixture or report.

**Green 12b — committed GitHub-only proof accounting.**
`scripts/github-live-proof-sweep.mjs` now loads the production GitHub `cli_surface.json` itself,
derives the complete `availability: implemented` set, and refuses a case file that omits even one
command. It permits only `proven`, `untestable`, and `failed` records; `proven` requires a 2xx
provider status plus a matched returned-data assertion, while the other two require a concrete
reason. Raw stdout, stderr, bodies, approval grants, credentials, and token-shaped text are
discarded or rejected before a JSON report is written.

Live mode is deliberately GitHub-only and requires an explicit built `pm` path, project root,
saved GitHub credential, test owner/repository, complete case file, report destination, and an
additional `--execute-writes` acknowledgement before it will dispatch a mutation. It inspects the
credential metadata first and refuses a credential not scoped to the supplied dedicated test
repository. Write cases use the runtime's existing plan → preview → caller-supplied confirmation
when declared → single-use grant sequence; transient plan and grant values stay in process memory
and are redacted from the stored invocation.

```
$ node --test scripts/tests/github-live-proof-sweep.test.mjs
✔ enumerates every and only implemented GitHub command
✔ rejects an omitted command instead of treating a sample as a sweep
✔ requires a returned-data assertion or a concrete untestable reason
✔ redacts raw subprocess output before it can become a report record
ℹ pass 4
ℹ fail 0

$ node scripts/github-live-proof-sweep.mjs --self-test
github live proof self-test: ok
```

No test was weakened, skipped, or deleted. The runner intentionally records a missing runtime HTTP
status as a failed live result rather than fabricating success; the full run therefore identifies
which intent paths need observability repairs before they can be counted as proven.

**Red 12c — GitHub archive downloads rejected their documented codeload redirect.** The binary
executor already refuses arbitrary cross-host redirects and strips credentials on a narrow
allowlist; its local redirect tests establish that behavior. The production GitHub `tarball` and
`zipball` declarations, however, named neither `allow_cross_host` nor the known codeload host, so
their live `api.github.com` redirect could only fail before streaming bytes. The focused production
bundle test failed before a declaration change:

```
$ go test -timeout 20m ./internal/connectors/engine/ -run TestGitHubArchiveDownloadsAllowOnlyCodeloadRedirect -count=1
--- FAIL: TestGitHubArchiveDownloadsAllowOnlyCodeloadRedirect
    --- FAIL: .../github.tarball_ref
        allowed_hosts = [], want [codeload.github.com]
    --- FAIL: .../github.zipball_ref
        allowed_hosts = [], want [codeload.github.com]
FAIL
```

This is deliberately a host allowlist, not `allow_cross_host: true`: the existing executor tests
already prove an allowlisted redirect still strips credentials and rejects every other host.

**Green 12c — narrow codeload declaration.** `github.tarball_ref` and `github.zipball_ref` now
declare only `allowed_hosts: ["codeload.github.com"]`. No transport code changed; the existing
binary requester's bounded redirect policy remains the sole redirect implementation, and continues
to strip credentials on the permitted host hop. The production test and the two relevant executor
policy tests passed:

```
$ go test -timeout 20m ./internal/connectors/engine/ -run 'TestGitHubArchiveDownloadsAllowOnlyCodeloadRedirect|TestBinaryDownload(RefusesCrossHostRedirectByDefault|AllowedHostsIsEnforced)' -count=1
ok   polymetrics.ai/internal/connectors/engine  0.804s

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 551 connector(s) checked, 0 findings

$ go run ./cmd/connectorgen surface-sync --check
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)
```

## Cycle 13 — rebase compatibility and guard verification

**Red 13 — current-main changed the preview API.** After rebasing this lane on
`origin/main`, the existing GitHub confirmation tests no longer compiled because
`App.PreviewReversePlan` now requires its `withheldFlags` argument. This was an
API integration failure, not a change to confirmation behavior:

```
$ go test -timeout 20m ./internal/app/ -run 'TestGitHubRepo(DeleteRequiresConfirmationAndSingleUseGrant|ArchiveIsApprovalOnlyAndSendsPinnedField|ApprovalOnlyCreatesCarryNoTypedChallenge)' -count=1
internal/app/reverse_confirmation_test.go:187:50: not enough arguments in call to a.PreviewReversePlan
    have (context.Context, string)
    want (context.Context, string, map[string][]string)
internal/app/reverse_confirmation_test.go:245:44: not enough arguments in call to a.PreviewReversePlan
    have (context.Context, string)
    want (context.Context, string, map[string][]string)
FAIL	polymetrics.ai/internal/app [build failed]
```

**Green 13 — preserve the original no-withheld-flags behavior.** Both test calls
now pass `nil`, which is the explicit current-main representation of no withheld
flags. The owner/repository guard, confirmation contract, rate declaration, and
bundle checks remained green after the rebase:

```
$ node --test scripts/tests/github-live-proof-sweep.test.mjs
✔ 5 tests passed

$ node scripts/github-live-proof-sweep.mjs --self-test
github live proof self-test: ok

$ go test -timeout 20m ./internal/app/ -run 'TestGitHubRepo(DeleteRequiresConfirmationAndSingleUseGrant|ArchiveIsApprovalOnlyAndSendsPinnedField|ApprovalOnlyCreatesCarryNoTypedChallenge)' -count=1
ok	polymetrics.ai/internal/app

$ go test -timeout 20m ./internal/connectors/commandrunner/ -run 'TestGitHub(ApprovalTextMatchesTheEnforcedWriteContract|GeneratedTwinsShareTheirAliasWriteContract|RestoredCommandsAreExecutable|HeldCommandsStayBlocked|NotesDoNotRestateTheTypedConfirmation)|TestEveryImplementedCommandPassesRuntimePreflight' -count=1
ok	polymetrics.ai/internal/connectors/commandrunner

$ go test -timeout 20m ./internal/connectors/engine/ -run 'TestGitHubDeclaredRateLimits|TestGitHubArchiveDownloadsAllowOnlyCodeloadRedirect|TestPreparedWritePreviewDeclaresConfirmationOnlyWhereTheGateDemandsIt|TestBundleLoadEmbeddedGitHub(Operations|CLISurface)' -count=1
ok	polymetrics.ai/internal/connectors/engine
```

---

## Cycle 14 — bounded status/text direct reads (GitHub parity gap closure)

**Red 14a — a successful empty or text response was incorrectly treated as malformed JSON.**
The focused operation-level test defines five declaration-bound cases before any policy or transport
change: a `204 No Content` status check must return `nil`; a nonempty status response must fail
closed; text must be valid UTF-8 and bounded by `rest.max_bytes`. Current code rejects the first
policy check, before issuing an untyped request:

```
$ go test -timeout 20m ./internal/connectors/engine/ -run TestOperationDirectReadSupportsBoundedStatusAndTextResponses -count=1
--- FAIL: TestOperationDirectReadSupportsBoundedStatusAndTextResponses (0.00s)
    --- FAIL: .../status_only_accepts_an_empty_success_body
        OperationDirectRead: direct read output policy "none" is not supported
    --- FAIL: .../text_response_is_bounded_and_valid_UTF-8
        OperationDirectRead: direct read output policy "text" is not supported
    --- FAIL: .../status_only_rejects_a_nonempty_success_body
        ... want nonempty status-only response rejection
    --- FAIL: .../text_response_rejects_invalid_UTF-8
        ... want UTF-8 rejection
    --- FAIL: .../text_response_retains_its_declared_cap
        ... want response cap rejection
FAIL
```

This is a foundation red, not a GitHub fixture assertion: the test uses local `httptest` responses,
an operation ledger row, and the real `OperationDirectRead` path. It proves the executor's JSON-only
assumption is the blocker for GitHub's nine 204 checks plus text endpoints.

**Green 14a — closed response policies, not an alternate transport.** `none` and `text` are now
recognized only by `OperationDirectRead`. The shared read walk keeps its existing bounded
requester and checks the cap before policy decoding: `none` returns `nil` only for an empty 2xx
body, while `text` requires valid UTF-8 and returns its bounded string. JSON policies still take
the existing decoder/redactor path. The legacy `DirectRead` path explicitly refuses both policies
before network, so a command cannot attach `none` to an arbitrary GET merely to discard its result.

The closed policy sets in commandrunner, `connectorgen`, and the CLI schema now agree; an
implemented operation command can use the declarations, while a non-operation command receives a
validator/preflight finding. No provider credential or GitHub request was used.

```
$ go test -timeout 20m ./internal/connectors/engine/ -run 'TestOperationDirectReadSupportsBoundedStatusAndTextResponses|TestDirectReadRejectsOperationOnlyResponsePoliciesBeforeNetwork' -count=1
ok	polymetrics.ai/internal/connectors/engine

$ go test -timeout 20m ./internal/connectors/commandrunner/ -run 'TestCLISurfaceOutputPolicyEnumMatchesRuntimePolicySets|TestRunDirectReadRequiresOutputPolicy' -count=1
ok	polymetrics.ai/internal/connectors/commandrunner

$ go test -timeout 20m ./cmd/connectorgen/ -run 'TestValidate_CLISurface' -count=1
ok	polymetrics.ai/cmd/connectorgen
```

**Red 14b — the typed operation request could not represent GitHub Markdown raw's documented
body.** A local `POST /markdown/raw` fixture asserts literal source bytes and `Content-Type:
text/plain`; its operation declares a root string schema and a text response. Before adding a
raw-body field to the closed operation request contract, the test does not compile:

```
$ go test -timeout 20m ./internal/connectors/engine/ -run TestOperationDirectReadSendsDeclaredPlainTextBody -count=1
# polymetrics.ai/internal/connectors/engine [polymetrics.ai/internal/connectors/engine.test]
internal/connectors/engine/direct_read_test.go:806:3: unknown field RawBody in struct literal of type connectors.OperationDirectReadRequest
FAIL	polymetrics.ai/internal/connectors/engine [build failed]
```

This is intentional red-before-transport evidence: the previous `Body map[string]any` can model
JSON object inputs only, and JSON-marshal would turn a Markdown source string into the wrong wire
format. The forthcoming contract is restricted to a declared plain-text POST; it is not a generic
raw-request field.

**Green 14b — literal text input is declaration-bound and stays on the existing requester.**
`OperationDirectReadRequest.RawBody` is distinct from an absent body, and the engine admits it only
for a `rest_read` `POST` declaring `text/plain`, no static `rest.body`, and a compiled root-string
schema. It rejects missing, mixed JSON/raw, oversized, and JSON-operation raw input before a
request. `Requester.DoTextLimited` has no caller-selected media type and delegates to the same
request core as JSON reads, preserving auth, retries, limiter admission before the request, and
limiter observation from the response.

The command surface has exactly one `maps_to: body` string flag for that contract; dotted JSON body
mappings remain disallowed. Markdown line breaks are accepted only in this literal body position
(not paths, queries, headers, or JSON fields); other control characters remain rejected. Bundle
load validates the new text contract while retaining current-main's established metadata-only POST
schema behavior for unrelated blocked connectors; an implemented command still takes the full
operation preflight.

```
$ go test -timeout 20m ./internal/connectors/engine/ -run 'TestOperationDirectRead(SendsDeclaredPlainTextBody|RejectsUndeclaredOrInvalidPlainTextBodiesBeforeNetwork|SupportsBoundedStatusAndTextResponses)' -count=1
ok	polymetrics.ai/internal/connectors/engine

$ go test -timeout 20m ./internal/connectors/commandrunner/ -run 'TestRun(OperationDirectReadPassesDeclaredPlainTextBody|OperationDirectReadRejectsMixedRawAndJSONBodyMappings|OperationDirectReadPlainTextBodyOnlyAdmitsDocumentWhitespace|ImplementedOperationDirectReadCommand)' -count=1
ok	polymetrics.ai/internal/connectors/commandrunner

$ go test -timeout 20m ./cmd/connectorgen/ -run 'TestValidate_CLISurfaceOperationDirectRead(PlainTextBodyRequiresOneRequiredStringFlag|RequiresBodyMappings|RequiresRequiredBodyFlags)' -count=1
ok	polymetrics.ai/cmd/connectorgen

$ go test -timeout 20m ./internal/connectors/engine/ ./internal/connectors/commandrunner/ ./internal/connectors/connsdk/ ./cmd/connectorgen/
ok	polymetrics.ai/internal/connectors/engine
ok	polymetrics.ai/internal/connectors/commandrunner
ok	polymetrics.ai/internal/connectors/connsdk
ok	polymetrics.ai/cmd/connectorgen
```

**Red 14c — GitHub's classified status/text rows still had no generated executable contracts.**
The generator-contract table names all nine documented `204` checks and the four text/Markdown
endpoints, including `POST /markdown/raw`'s 400 KiB declared plain-text ceiling. It loads the
embedded bundle and calls the real operation preflight, so it cannot be satisfied by changing only
the API ledger. Before the GitHub generator knows these explicit exceptions, every expected command
is absent:

```
$ go test -timeout 20m ./cmd/connectorgen/ -run TestGitHubStatusAndTextOperationContracts -count=1
--- FAIL: TestGitHubStatusAndTextOperationContracts
    generated command "gists star check" is missing
    generated command "orgs blocks check" is missing
    generated command "orgs members check" is missing
    generated command "orgs public-members check" is missing
    generated command "teams members check" is missing
    generated command "user blocks check" is missing
    generated command "user following check" is missing
    generated command "user starred check" is missing
    generated command "users following check" is missing
    generated command "meta zen view" is missing
    generated command "meta octocat view" is missing
    generated command "markdown render" is missing
    generated command "markdown raw render" is missing
FAIL
```

The next green step changes the source generator's narrow table, regenerates only the GitHub bundle
and shared ledger, and then updates the derived inventory snapshot from 1126/98 to 1139/85.

**Green 14c — the generator emits the 13 exact operation contracts and no broad classifier bypass.**
`EXPLICIT_DIRECT_READ_CONTRACTS` promotes only the nine status checks, `/zen`, `/octocat`, and the
two Markdown renderers. The generated commands carry bounded `none` or `text` policies; the JSON
renderer has a required `text` body field plus the documented `mode`/`context` options, and raw
Markdown has the sole exact `body` mapping with its 400 KiB limit. The starred-repository check
uses explicit path `owner`/`repo` flags rather than accidentally reusing the connection's target.

The generator produced 13 GitHub rows, then `surface-sync` regenerated the shared runtime endpoint
ledger. A before/after structural comparison found exactly those 13 GitHub API rows, operations,
and CLI paths changed; the shared ledger changed only its `github` key. Inventory now reads
1139 covered / 85 blocked; no final parity claim follows from that intermediate reduction.

```
$ python3 scripts/gen-github-parity.py
generated: 13 endpoints covered
  new operations: 13 (total 358)
  new write actions: 0 (total 555)
  new cli commands: 13 (total 1160)

$ go test -timeout 20m ./cmd/connectorgen/ -run 'TestGitHub(StatusAndTextOperationContracts|APISurfaceOperationLedgerMetrics|DocumentedRESTSurfaceIsComplete)' -count=1
ok	polymetrics.ai/cmd/connectorgen

$ go test -timeout 20m ./internal/connectors/certify/ -run TestSurfaceInventoryForGitHubAccountsForAllReviewedEndpoints -count=1
ok	polymetrics.ai/internal/connectors/certify

$ go test -timeout 20m ./internal/connectors/commandrunner/ -run 'TestEveryImplementedCommandPassesRuntimePreflight|TestGitHubRestoredCommandsAreExecutable' -count=1
ok	polymetrics.ai/internal/connectors/commandrunner

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 551 connector(s) checked, 0 findings

$ go run ./cmd/connectorgen surface-sync --check
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)
```

---

## Cycle 15 — concrete `oneOf` write arms and bounded structured record input

**Red 15a — a declared structured record field could not yet pass the real write path.**
The new commandrunner contract deliberately leaves `coerceFlagValue`'s generic `json` rejection
intact. It declares a `json` flag only at `record.payload`, backed by a concrete object schema,
then invokes the real `Preflight` and `BuildWriteCommand`. Before the narrow declaration-bound
parser and engine preflight exist, a valid object, malformed JSON, and a top-level array all fail
at the old generic rejection; a scalar target is also incorrectly preflighted as executable:

```
$ go test -timeout 20m ./internal/connectors/commandrunner/ -run TestBuildWriteCommandSupportsOnlyDeclaredStructuredJSONRecordFlags -count=1
--- FAIL: TestBuildWriteCommandSupportsOnlyDeclaredStructuredJSONRecordFlags
    valid object becomes typed planned record: flag --payload has unsupported type "json"
    malformed JSON never produces a plan: want "invalid JSON", got unsupported type "json"
    array cannot satisfy object field: want "does not match type", got unsupported type "json"
    Preflight scalar structured-json mapping error = <nil>, want declared object/array rejection
FAIL
```

This establishes the bounded foundation required by the documented GitHub arms without admitting a
generic request-body escape hatch: only declared `record.<field>` object/array inputs may acquire
JSON parsing, and the action schema must remain the type authority.

**Red 15b — GitHub's documented top-level `oneOf` arms were still classified but absent.**
`TestGitHubOneOfWriteContracts` names all 19 actions across the eight documented endpoints, checks
their required fields and `covered_by.writes` linkage, invokes the real runtime preflight, and
requires destructive acknowledgement only for the four attestation deletion arms. Before the
generator has concrete arm declarations, all 19 commands are absent (for example):

```
$ go test -timeout 20m ./cmd/connectorgen/ -run TestGitHubOneOfWriteContracts -count=1
--- FAIL: TestGitHubOneOfWriteContracts
    orgs_attestations_delete-by-subject-digests: generated command "orgs attestations delete-by-subject-digests" is missing
    orgs_campaigns_create-code-scanning: generated command "orgs campaigns create-code-scanning" is missing
    orgs_projects_fields_create-single-select: generated command "orgs projects fields create-single-select" is missing
    users_projects_items_create-by-repo-number: generated command "users projects items create-by-repo-number" is missing
    codespaces_create-from-pull-request: generated command "codespaces create-from-pull-request" is missing
FAIL
```

The green step must change the generator source, regenerate GitHub's bundle and shared ledger, and
keep the shared ledger delta confined to `github`.

**Green 15 — each documented root `oneOf` arm has one closed, executable contract.**
The engine owns the common structural rule: a `json` CLI flag is admissible only for a declared
top-level `record.<field>` whose record schema declares an `object` or `array`. The generator's
static validation and the commandrunner's real preflight both call that same engine rule. The
parser accepts exactly one bounded (1 MiB) JSON object or array with `UseNumber`; it neither
re-enables generic `json` flags nor permits `body.*`, nested record paths, scalars, or undeclared
fields to escape a write schema.

`EXPLICIT_ONE_OF_WRITE_CONTRACTS` is deliberately a small, source-owned table rather than a broad
classifier bypass. It expands the eight reviewed GitHub endpoints into 19 closed reverse-ETL
actions, with `covered_by.writes` on each source endpoint. Only the four attestation-delete arms
carry the existing destructive caller-supplied intent acknowledgement; the other fifteen are
ordinary approval-gated writes. Regeneration changed only GitHub definition artifacts. This is a
write-only slice, so `surface-sync` was still rerun but correctly made no shared direct-read ledger
change.

The generated inventory is now 1,147 covered / 77 blocked endpoints, with 377 operations, 574
write actions, and 1,179 GitHub CLI commands. The planning source correction also fixes the
secret-scanning campaign arm's property name from `code_scanning_alerts` to
`secret_scanning_alerts`.

```
$ python3 -m py_compile scripts/gen-github-parity.py
$ python3 scripts/gen-github-parity.py
generated: 8 endpoints covered
  new operations: 19 (total 377)
  new write actions: 19 (total 574)
  new cli commands: 19 (total 1179)

$ go test -timeout 20m ./internal/connectors/engine/
ok  polymetrics.ai/internal/connectors/engine

$ go test -timeout 20m ./internal/connectors/commandrunner/
ok  polymetrics.ai/internal/connectors/commandrunner

$ go test -timeout 20m ./internal/connectors/certify/
ok  polymetrics.ai/internal/connectors/certify

$ go test -timeout 20m ./cmd/connectorgen/
ok  polymetrics.ai/cmd/connectorgen

$ go vet ./...
$ go build ./cmd/pm
$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 551 connector(s) checked, 0 findings

$ go run ./cmd/connectorgen surface-sync --check
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)
```
## Cycle 30 — typed GraphQL source import for all-root generation

**Scope:** GitHub-only source tooling and generated GitHub/planning artifacts.  No credential,
`pm`, browser, fixture, or provider request occurs in this cycle.  The public official SDL hash
was independently verified in memory before this cycle; the test suite itself remains hermetic.

**Red 30a — root signatures alone cannot generate a safe typed operation.**
The existing source lock records `Query`/`Mutation` names and formatted signatures only.  Add
fixture coverage that requires a parsed root argument type, return type, nested input-field model,
enum values, and interface/union possible-object facts.  The current v1 parser/lock is expected to
fail this test because those facts are absent.

**Red 30b — type drift must fail before ledger output.**
Add negative fixtures/tests for an undeclared referenced input type, duplicate type/field identity,
and a `createEnterpriseOrganization` canary whose typed `input` contract is missing.  A command
document regex is not evidence of a `node`/`nodes` projection: require the ledger to use the parsed
type graph instead.

```
$ node --test scripts/tests/github-combined-operation-ledger.test.mjs
✖ imports typed GraphQL root contracts and source-derived projection possibilities
  AssertionError: 1 !== 2
✖ fails closed when GraphQL root types or the enterprise typed input contract drift
  AssertionError: Missing expected exception.
✖ records generic node projections and the disabled deleteIssue mutation from the real GitHub bundle
  AssertionError: projection_matrix has no possible_object_types
```

The red result proves three independent absences in the v1/root-signature-only importer: no typed
source graph, no source-type drift rejection, and generic projections inferred only from existing
documents.  It made no network or provider request.

**Green 30 — a dependency-free parser now writes a compact v2 typed source lock.**
`parseGraphQLSchema` retains structured root argument/return type references plus compact input,
enum, object, interface, union, and scalar facts.  It checks unknown references, duplicate
identities, the `createEnterpriseOrganization(input: CreateEnterpriseOrganizationInput!)` canary,
and source-derived interface/union possibilities.  Generic `node`/`nodes` ledger rows now carry
the source possible-object matrix separately from the fixed documents' supported-object list.
The review regression also proves that a v2 lock cannot omit an explicit empty `arguments` array
for a no-argument root; every source root has a typed argument contract as well as a return type.

The checked-in lock and combined ledger were mechanically regenerated from the same official public
SDL provenance already pinned in the lock.  The source bytes/hash remain
`1,546,421` / `c09aba9911b08d2aa8a022578edaf256aa040f38d7fb7196656356ea236c249d`;
the denominator remains `1,220 REST + 31 GraphQL Query + 274 GraphQL Mutation = 1,525`.
The type graph contains 415 input objects, 254 enums, 1,025 objects, 50 interfaces, 49 unions,
and 13 scalars; `Node` has 282 source-declared possible object types.  No raw SDL, secret, or
provider response body is checked in.

```
$ node --test scripts/tests/github-combined-operation-ledger.test.mjs
✔ 8 tests passed

$ node scripts/github-combined-operation-ledger.mjs --check
github combined operation ledger: ok rest=1220 graphql_query=31 graphql_mutation=274 total=1525

$ node --test scripts/tests/github-live-lab.test.mjs
✔ 23 tests passed

$ node scripts/github-live-lab.mjs --check-boundary --boundary .planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-BOUNDARY.json
github live lab boundary: ok allowed_targets=1

$ go test -timeout 20m ./internal/connectors/engine -count=1
ok  polymetrics.ai/internal/connectors/engine

$ go test -timeout 20m ./internal/connectors/commandrunner -count=1
ok  polymetrics.ai/internal/connectors/commandrunner

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 551 connector(s) checked, 0 findings

$ go run ./cmd/connectorgen surface-sync --check
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)
```

**GSD/manual fallback:** all required adapter prompts and command-source resolutions were generated
and read inline in this single-worker lane; no GSD role may be spawned.  Skills: `golang-how-to`,
`golang-cli`, `golang-graphql`, `golang-testing`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-design-patterns`, and `golang-structs-interfaces`.

---

## Cycle 31 — generated fixed GraphQL root contracts

**Scope:** local generated-contract work only.  The deploy-key regression, generic DELETE handling,
PM-only lab boundary, and PM read-back tests were re-run green before this cycle.  No provider
request, fixture creation, cleanup retry, browser action, or credential use is permitted here.

**Red 31a — a closed GraphQL input variable is still rejected as generic JSON.**
The first focused test will instantiate a fixed GraphQL operation with a closed object `input`
variable, invoke the real command preflight/shaper, and prove that only the source-declared
top-level `body.input` can parse one bounded JSON object/array.  The existing reverse-ETL-only
placement rule must reject it before the narrow GraphQL-variable validator is added.  The same test
will retain the refusal for ordinary REST direct operations, nested `body.input.field`, scalar
variables, unknown variables, malformed JSON, and raw documents.

The initial engine boundary test is red because no declaration-owned validator exists yet:

```
$ go test -timeout 20m ./internal/connectors/engine \
  -run TestValidateGraphQLOperationStructuredJSONVariableRequiresClosedTopLevelContainer -count=1
# polymetrics.ai/internal/connectors/engine [polymetrics.ai/internal/connectors/engine.test]
internal/connectors/engine/graphql_operation_test.go: undefined: ValidateGraphQLOperationStructuredJSONVariable
FAIL
```

The runner-level counterpart then reproduces the exact closed-placement symptom with a fixed
`POST /graphql` operation:

```
$ go test -timeout 20m ./internal/connectors/commandrunner \
  -run TestRunOperationDirectReadAdmitsOnlyPreflightedStructuredGraphQLVariable -count=1
--- FAIL: TestRunOperationDirectReadAdmitsOnlyPreflightedStructuredGraphQLVariable
    Run: connector command "graphql query widgets" is blocked: intent=direct_read:
    availability=implemented: structured JSON flag --input is allowed only on a declared
    reverse-ETL record field
FAIL
```

**Red 31b — source graph facts are not yet a command catalog.**
The generator test will use the mini typed SDL to demand exactly one fixed operation/command/physical
transport mapping for every source root, including the typed `createEnterpriseOrganization` canary
and an explicit non-executable `deleteIssue`.  It must reject missing roots, unbounded input arrays,
duplicate operation IDs/paths, and any caller-controlled document or selection field.  Before the
new generator exists there is no generated all-root catalog, so this test is expected to fail by
missing module/contract rather than passing on the legacy four bindings.

```
$ node --test scripts/tests/gen-github-graphql-parity.test.mjs
Error [ERR_MODULE_NOT_FOUND]: Cannot find module 'scripts/gen-github-graphql-parity.mjs'
FAIL
```

**Green 31 — every pinned root now has one fixed, typed PM contract.**
`scripts/gen-github-graphql-parity.mjs` consumes only the checked-in v2 source lock and writes one
operation, one CLI command, and one exact `POST /graphql` operation binding for each of the 31
`Query` plus 274 `Mutation` roots.  The transport is not counted as a REST endpoint: the runtime
ledger binds every query by its operation ID, and the write preflight binds every mutation to that
same named transport list.  All generated documents, operation names, paths, selections, variable
schemas, input-depth/array bounds, and cursor handling are declaration-owned; no raw document,
selection, endpoint, header, or cursor flag exists.

The deterministic source inventory is now `1,220 REST + 31 GraphQL Query + 274 GraphQL Mutation =
1,525`.  The combined ledger reports these separate facts rather than translating local generation
into live acceptance: inventory `1525/1525 (100%)`, exact executable implementation
`1345/1525 (88.2%)`, and current-head live proof `0/1525 (0%)`.  `createEnterpriseOrganization`
is present as the typed source canary; `deleteIssue` remains `unsafe_or_disallowed`.  The scheduled
source-drift workflow is read-only and checks the independently pinned REST and GraphQL source
inventories without altering generated artifacts.

**Red 31c — GraphQL non-2xx responses bypassed the envelope sanitizer.**
The loopback safety tests initially showed that a failed fixed GraphQL mutation, and then a failed
fixed GraphQL query, could include the synthetic provider error-body fixture verbatim.  Both tests
were red before the fix:

```
$ go test -timeout 20m ./internal/connectors/engine \
  -run 'TestOperationDirect(Read|Write)RedactsGraphQLHTTPErrorBody' -count=1
FAIL: raw GraphQL HTTP error body reached the returned error
```

**Green 31c — all fixed GraphQL failure paths use the same redaction boundary.**
The query and mutation HTTP-error paths now apply `safety.RedactErrorText`; successful GraphQL
responses continue to expose only bounded/sanitized `errors[]` metadata.  The loopback tests prove
both paths retain a redaction marker without retaining the fixture value.

**Red/Green 31d — a nominal query could select an appended mutation.**
`TestPreflightOperationDirectReadRejectsMixedFixedGraphQLDocument` was red when a document beginning
with a query appended a named mutation and set `operationName` to that mutation.  Prefix-only
validation accepted the document as a direct read.  The fixed-operation admission check now uses a
small comment/string-aware top-level scanner and requires exactly one named operation of the
declared kind with the declared operation name.  It accepts ordinary fixed fragments and
argument/default syntax, but refuses extra operations, subscriptions, anonymous operations, name
mismatches, and malformed top-level syntax before any credential or request is resolved.

`cmd/connectorgen` now also treats a fixed GraphQL query as an executable read although its physical
transport is POST, matching conformance's capability accounting.  The new regression first failed
with `capabilities.read=false` and no finding, then passed after the shared surface rule was fixed.

```
$ go test -timeout 20m ./internal/connectors/engine -count=1
ok  polymetrics.ai/internal/connectors/engine

$ go test -timeout 20m ./internal/connectors/commandrunner -count=1
ok  polymetrics.ai/internal/connectors/commandrunner

$ go test -timeout 20m ./cmd/connectorgen -count=1
ok  polymetrics.ai/cmd/connectorgen

$ go test -timeout 20m ./internal/connectors/conformance ./internal/connectors/certify -count=1
ok  polymetrics.ai/internal/connectors/conformance
ok  polymetrics.ai/internal/connectors/certify

$ make github-parity-artifacts-check
github graphql parity artifacts: ok
github combined operation ledger: ok rest=1220 graphql_query=31 graphql_mutation=274 total=1525

$ go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 551 connector(s) checked, 0 findings

$ go run ./cmd/connectorgen surface-sync --check
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)

$ node --test scripts/tests/github-live-lab.test.mjs
✔ 23 tests passed

$ node scripts/github-live-lab.mjs --check-boundary \
  --boundary .planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-BOUNDARY.json
github live lab boundary: ok allowed_targets=1
```

Help/manual/site artifacts were regenerated and representative help verified locally for the GraphQL
namespace, `query viewer`, the typed `create-enterprise-organization` mutation, and the blocked
`delete-issue` mutation.  `go vet` for the changed packages, `go build ./cmd/pm`, the full
`internal/app` and `internal/cli` package suites, and the website script suite all passed.  This
cycle made no `pm` provider invocation, browser action, credential read, fixture creation, cleanup,
or provider write.

---

## Cycle 32 — target rebind and authenticated direct-read finding (planned)

**Scope:** GitHub-only live-lab safety and one authenticated PM direct-read defect. The captain
supplied the independently verified immutable target ID `1327549621`; it is not an authorization to
retain the prior target or to bypass PM for provider activity.

**Red 32a:** Current `validateCleanupLedger` rejects every historical event as soon as the current
boundary's run ID/target changes. Add a regression with an archived historical run and a new exact
current target. It must fail until historical validation is explicitly separate from normal target
authorization, and it must prove the archived target cannot reach PM read/write execution.

**Red 32b:** The authenticated PM command returned a JSON `kind: Error` envelope twice with no safe
provider status. Add a red regression requiring a bounded safe capture to preserve all non-sensitive
envelope fields, retain status/null faithfully, reject secrets, and never silently convert the error
to a credential or setup outcome.

**Green criteria:** historical target evidence validates only under an exact archived identity;
current executors allow only the captain target; the committed boundary contains the supplied owner
and repository IDs; the current-head PM failure is captured as a sanitized finding and traced through
the direct-read path before any provider mutation.

**Red 32a — retired cleanup evidence was coupled to the live allowlist.**
`node --test --test-name-pattern='archived cleanup evidence never reauthorizes'
scripts/tests/github-live-lab.test.mjs` first failed with `cleanup ledger entry 1 has a different
run_id`. The old current-run-only validator could neither retain the `8ccd8dd6` history nor prove
that it had become mutation-ineligible.

**Green 32a — history is validation-only.** The boundary now archives the historical run under its
own exact immutable target while the one executable `allowed_targets` entry is the captain-approved
`karthik-sivadas/pm-live-test-direct-read-20260808081515` / repository ID `1327549621` target.
`normalizeHistoricalRun` and `validateCleanupLedger` validate archived entries only against that
archive; `authorizeLabTarget` and every PM execution seam continue to read `allowed_targets` only.
The regression proves an archived target fails before a PM runner can start.

**Red 32b — the PM failure could not be safely preserved as evidence.** The initial capture test
failed to load because `github-live-lab.mjs` exported no `capturePMErrorEnvelope`. The missing
seam would have encouraged a shape-only summary and lost the classified error details the captain
explicitly required.

**Green 32b — bounded full Error envelope capture.** `capturePMErrorEnvelope` now accepts only a
JSON `kind: Error` PM envelope, rejects sensitive fields and credential-shaped values, bounds the
serialized object, preserves every safe field, and reports `provider_status: null` unless an actual
structural HTTP status field exists. The regression includes a message that merely mentions a status
to prove the harness does not invent one.

**Red 32c — the current finding/control had no source-controlled record.** The divergence test
failed twice with the new records absent (then with `current_run_id` absent), preventing a
historical run ID from being silently reused for current-head evidence.

**Green 32c — current-head diagnostic and control are mechanically pinned.**
`GITHUB-LIVE-LAB-DIVERGENCES.json` now carries a distinct
`github-live-lab-20260810-target-rebind` record for each current invocation. The exact PM-only
authenticated repository-list diagnostic exited `1` with the full safe envelope:

```
{"api_version":"polymetrics.ai/v1","error":{"category":"internal","code":"internal_error","message":"rate-limit policy \"authenticated-user\" requires non-secret config \"rate_limit_account\" for its declared scope"},"kind":"Error"}
```

There was no provider status because `newRuntime` resolves GitHub's whole-connector
`authenticated-user` rate-limit policy before a requester dispatches. This is GitHub-specific, and
affects every GitHub runtime request selected by that policy; it is not shared all-connector
direct-read machinery. The GitHub specification/docs describe `rate_limit_account` as optional but
the matching policy requires it, while the CLI serialized the local configuration error as
`internal_error`; that contract gap is recorded as a defect rather than credential friction.

After adding only the approved non-secret rate-limit account subject to the disposable credential
(the owner/repository scope was unchanged), the preserved PM-only `repo view` control exited `0`,
returned `ConnectorCommandRead`, and matched the bound immutable ID, exact slug, and private state
without `--owner`, `--repo`, `--config`, or `--connection` overrides. Its `repository` stream
projection has no `archived` schema field, so it deliberately does **not** claim unarchived state;
the captain's independent target verification remains the authority for that admission property.

Focused safety gates passed before the control and remain green:

```
go test -timeout 20m ./internal/app -run '^TestGitHubDeployKeyDeleteDoesNotMaskNotFoundForAVisibleBoundFixture$' -count=1
node --test scripts/tests/github-live-lab.test.mjs
node scripts/github-live-lab.mjs --check-boundary --boundary .planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-BOUNDARY.json
git diff --check
```

---

## Cycle 33 — captain-ordered complete REST + GraphQL closure

**Scope:** replace the legacy REST-only completion denominator with source-derived terminal
accounting, close every ordinary documented REST/GraphQL command route, and retain only the
captain-held `auth token` and `api` aliases.  The current PM-only lab remains target-bound and no
provider write is authorized by this local classification cycle.

**Red 33a — the source bundle still records nonterminal operation and command rows.**
`TestGitHubCompleteParityLeavesOnlyCaptainHeldRawAliases` was added before changing a source
classification.  It reads the generated GitHub `cli_surface.json` and `api_surface.json`, excludes
only the captain-held token-printing/raw-API aliases, and requires every REST endpoint to have a
fixed executable `covered_by` contract or a machine-checkable named dependency.  The first red run
is the exact reconciliation point for the captain's historical 1,224/1,179/1,081 figures versus
the current generated GraphQL source:

```
$ go test -timeout 20m ./cmd/connectorgen \
  -run '^TestGitHubCompleteParityLeavesOnlyCaptainHeldRawAliases$' -count=1
--- FAIL: TestGitHubCompleteParityLeavesOnlyCaptainHeldRawAliases
    GitHub completion still has 53 nonterminal command rows
    GitHub completion still has 77 REST endpoints without a fixed executable contract or named dependency
FAIL
```

The 53 command rows are exactly 37 `partial`, 8 `planned`, and 8 unsafe rows other than the two
held aliases; the 77 REST rows are 67 former `duplicate`, 9 former `disallowed`, and one former
`deprecated` endpoint.  The test reports all exact paths rather than accepting the contradictory
historical phrase “45 no-command endpoints”.  The planned green state may not hide a source row by
changing the test: it must supply a concrete operation/command contract, or use the forthcoming
schema-validated named-dependency model for a genuine captain-held capability boundary.

**Red/Green 33b(1) — materialized structured record flags.** The former
`materializedWriteFlags` path rejected every required nested object or object-array because it
could only invent scalar flags. That disagreed with the already-green runtime contract
`TestBuildWriteCommandSupportsOnlyDeclaredStructuredJSONRecordFlags`, which permits one
declaration-bound `json` value for a closed top-level record property. The new focused test
`TestMaterializedWriteFlagsUseTopLevelStructuredJSONForRequiredContainers` locks the bridge:
required descendants of `payload` and `items` normalize to exactly `--payload`/`--items` JSON
flags, never flattened child flags. Green evidence:

```
$ go test -timeout 20m ./cmd/connectorgen \
  -run '^TestMaterializedWriteFlagsUseTopLevelStructuredJSONForRequiredContainers$' -count=1
ok   polymetrics.ai/cmd/connectorgen

$ go test -timeout 20m ./cmd/connectorgen -run '^TestBatchMaterialize' -count=1
ok   polymetrics.ai/cmd/connectorgen
```

The materializer still rejects unrepresentable scalar/union shapes; `json` is accepted only for
a top-level object/array validated by `engine.ValidateStructuredJSONRecordField`, so this does
not add an arbitrary request-body channel.

**Red/Green 33b(2) — static validation recognizes the same JSON boundary as runtime.** Adding
one required `payload.kind` child to the already-valid structured-json fixture first produced:

```
implemented reverse ETL command 1 ("widget create") for write "create_widget" lacks flag mappings
for required record fields: payload.kind
```

The runtime had correctly accepted `--payload` after checking the action's schema, but the static
validator treated it as an exact scalar mapping. It now lets a `json` flag satisfy descendants
only when its top-level mapping is the declared container; scalar mappings keep the prior exact or
child-construction behavior. Green evidence:

```
$ go test -timeout 20m ./cmd/connectorgen \
  -run '^(TestValidate_CLISurfaceReverseETLStructuredJSONRequiresDeclaredTopLevelContainer|TestMaterializedWriteFlagsUseTopLevelStructuredJSONForRequiredContainers)$' -count=1
ok   polymetrics.ai/cmd/connectorgen

$ go run ./cmd/connectorgen validate internal/connectors/defs/github
connectorgen validate: 1 connector(s) checked, 0 findings

$ go test -timeout 20m ./internal/connectors/commandrunner \
  -run '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1
ok   polymetrics.ai/internal/connectors/commandrunner
```

---

### Cycle 33 continuation — secret-safe GraphQL closure and zero unsafe aliases

**Scope:** close the remaining generated GraphQL and legacy-alias terminal states without adding a
generic secret or raw-request channel. This continuation records the final source-classification
work in Cycle 33; the older live evidence and prior policy records remain historical evidence.

**Red 34a — a generated sensitive GraphQL mutation had no safe way to receive its required JSON
input.** Before the declaration and CLI changes, the focused test did not compile because
`CommandSurfaceFlag.EnvOnly` and `resolveConnectorCommandEnvironmentOnlyFlags` did not exist. A
secret value could only have been passed on argv, which violates the at-rest/preview/transcript
boundary.

**Green 34a — declaration-bound environment-only input.** `env_only` is accepted only for a
required top-level JSON field on an implemented `graphql_mutation` whose operation is explicitly
`mutation_class: secret`, declares `input_mode: env`, typed confirmation, and an exact redaction
path. The CLI accepts only `--from-env field=ENV` for such a declared flag; it rejects direct argv
values, undeclared fields, duplicate fields, malformed identifiers, and empty environment values
without formatting the value. It resolves the value in memory immediately before the existing typed
coercion and passes the resolved flags through the existing plan withholding path. There is no
generic `--from-env` body mechanism.

```
$ go test -timeout 20m ./internal/cli \\
    -run '^TestResolveConnectorCommandEnvironmentOnlyFlags$' -count=1
ok   polymetrics.ai/internal/cli

$ go test -timeout 20m ./cmd/connectorgen ./internal/connectors/engine ./internal/cli \\
    -run 'TestValidate_CLISurfaceEnvOnlyFlagRequiresDeclaredSecretGraphQLContract|TestSecretOperationTypedConfirmationPolicyReachesSharedWriteGate|TestResolveConnectorCommandEnvironmentOnlyFlags' -count=1
ok   polymetrics.ai/cmd/connectorgen
ok   polymetrics.ai/internal/connectors/engine
ok   polymetrics.ai/internal/cli
```

**Red/Green 34b — source semantics, not a type-name heuristic, identify secret mutations.** The
first generator rule matched any input type containing `Token`, falsely treating
`RegenerateVerifiableDomainTokenInput` as secret. The generator now examines recursively declared
input *field names*. The current pinned schema produces exactly three sensitive mutations:
`createMigrationSource`, `startOrganizationMigration`, and `startRepositoryMigration`. Each has
the environment-only JSON input, `body.input` redaction, and the shared typed confirmation gate.
`deleteIssue`, `transferIssue`, and `revertPullRequest` are ordinary implemented destructive
mutations and receive typed confirmations, not an unsafe classification.

```
$ node --test scripts/tests/gen-github-graphql-parity.test.mjs
✔ generated secret GraphQL mutations are exact field-name matches
✔ destructive GraphQL mutations remain implemented and typed-confirmed
```

**Green 34c — legacy aliases are fixed bindings, not duplicate capability paths.** The generator
now derives `issue`, `pr`, `release`, `workflow`, `run`, `ruleset`, `discussion`, `project`,
`search`, and `status` compatibility paths by cloning an exact existing declared operation or
write action. REST direct-read aliases share the same API-surface endpoint coverage; GraphQL write
aliases use the same plan/preview/approval/execute contract. `auth token` and `api` are
`unsupported_local` safety boundaries with no operation/write/API-surface binding, so the surface
has **zero** `unsafe_or_disallowed`, `partial`, or `planned` rows and neither alias accidentally
becomes executable.

```
$ go run ./cmd/connectorgen validate internal/connectors/defs/github
connectorgen validate: 1 connector(s) checked, 0 findings

$ go run ./cmd/connectorgen surface-sync internal/connectors/defs --check
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)

$ go test -timeout 20m ./cmd/connectorgen \\
    -run '^TestGitHubCompleteParityHasNoNonterminalCommandRows$' -count=1
ok   polymetrics.ai/cmd/connectorgen

$ go test -timeout 20m ./internal/connectors/commandrunner \\
    -run '^TestGitHub(RestoredCommandsAreExecutable|CapabilityEscapesStayNonExecutableWithoutUnsafeClassification|LegacyAliasesPassRuntimePreflight|GraphQLDestructiveAliasesRequireTypedConfirmation)$' -count=1
ok   polymetrics.ai/internal/connectors/commandrunner
```

**Green 33c(1) — 31 partial reverse-ETL aliases now have exact record forms.**
`gen-github-parity.py` derives each missing required top-level field from the action's concrete
`record_schema`: scalars retain scalar flags, declared objects/non-string arrays become required
`json` record flags, and string arrays retain `string_array`. It refuses root `oneOf`/`anyOf`
instead of broadening an input contract. This changed the source inventory from 37 partial rows to
the six true read-alias gaps; the complete-parity red test now reports 22 nonterminal rows
(6 partial + 8 planned + 8 non-held unsafe), down from 53. The shared runtime sweep confirms none
of the 31 promotions stops at preflight.

```
$ python3 scripts/gen-github-parity.py
generated: 0 endpoints covered
  promoted partial structured writes: 31

$ go run ./cmd/connectorgen validate internal/connectors/defs/github
connectorgen validate: 1 connector(s) checked, 0 findings

$ go test -timeout 20m ./internal/connectors/commandrunner \
  -run '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1
ok   polymetrics.ai/internal/connectors/commandrunner
```

**Red/Green 33b — source-pinned REST closure.** The completion test initially
reported 77 REST rows with no executable contract: 67 historical `duplicate`,
9 `disallowed`, and one `deprecated` classifier. Those labels described a
prior generator limitation; they were not runtime contracts. The initial
source-derived generation correctly reduced the red output to the four root
union endpoints below, proving that 73 independently typed REST contracts
were added without weakening the completion test:

```
DELETE /user/emails
POST /user/emails
PATCH /orgs/{org}/secret-scanning/custom-patterns/{pattern_id}
PATCH /repos/{owner}/{repo}/secret-scanning/custom-patterns/{pattern_id}
```

The generator reads only the explicitly supplied, SHA-256-pinned
`github/rest-api-description` artifact
`80850db290cde4eb487e0efb587cf27f305e77b6bef96933ed8a09b5169d5b1d`; it neither
fetches a source nor accepts an unpinned substitution. Each ordinary GET
receives its declared path/query flags and bounded direct-read operation. Each
write receives a closed record schema derived from the source, declared
top-level structured JSON fields where necessary, plan/approval execution,
and destructive confirmation for DELETE. The rerun also repairs an
interrupted generated state by deduplicating only identical `record.*` flag
mappings and retaining the parameter-derived field type.

The four remaining root unions are explicit named actions: object/array email
forms (an array of one is the documented scalar email semantics) and five
independently required custom-pattern update arms for each scope. No generic
raw body was added. JSON-array actions use the existing declaration-bound
`body_field`/`body_schema` executor, and the email deletion actions retain the
typed destructive confirmation gate. After regeneration, the same completion
test reported only the 22 remaining command classifications and **no REST
endpoint finding**, while the real runtime preflight and static validator were
green:

```
$ GITHUB_OPENAPI_PATH=/tmp/github-openapi-parity.4UTT85/api.github.com.json \
    python3 scripts/gen-github-parity.py
generated: 4 endpoints covered
  new operations: 14 (total 769)
  new write actions: 14 (total 607)
  new cli commands: 14 (total 1571)

$ go run ./cmd/connectorgen surface-sync internal/connectors/defs --check
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)

$ go run ./cmd/connectorgen validate internal/connectors/defs/github
connectorgen validate: 1 connector(s) checked, 0 findings

$ go test -timeout 20m ./internal/connectors/commandrunner \
    -run '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1
ok   polymetrics.ai/internal/connectors/commandrunner
```

---

## Cycle 35 — post-`f96a47e80` regenerated certification inventory

**Red — rebase exposed stale fixed inventory totals, not a runtime regression.** The first
post-main focused gate ran the real regenerated GitHub bundle and failed
`TestSurfaceInventoryForGitHubAccountsForAllReviewedEndpoints` with
`Covered = 1225, want legacy coverage plus fixed GraphQL transport 1148`, and
`TestGithubWriteActionInventoryAccountsForAllDeclaredActions` with
`len(items) = 607, want 574`. The generated `api_surface.json` is the
authoritative source for this assertion: it contains 1,225 rows (1,220 REST,
four retained legacy GraphQL bindings, and one shared GraphQL transport), all
covered; its coverage totals are 37 streams, 607 writes, 366 singular
direct-reads, 252 plural direct-reads, and 305 GraphQL operation bindings.

**Green — assert complete parity rather than the superseded blocked baseline.** The certify
test now derives its endpoint denominator from the pinned REST count plus the explicit GraphQL
rows, requires every endpoint to be covered and no blocker models/statuses, and pins the
source-derived coverage/write-action totals above. This strengthens the test: regenerating a
partially covered surface cannot silently retain the former 77 blocked rows.

---

## Cycle 36 — CI repair: bounded registry initialization and safe archive copy

**Delivery context:** scoped no-mistakes CI repair under the outer executor; no pipeline command
was invoked from this phase. Required skills used: `golang-how-to`, `golang-troubleshooting`,
`golang-continuous-integration`, `golang-cli`, `golang-testing`, `golang-error-handling`,
`golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`,
`golang-performance`, `golang-benchmark`, and `golang-lint`.

**Red — the verify suite exhausted its per-package 20-minute deadline.** CI timed out
`internal/cli` at `1200.055s` while `TestGitHubDestructiveCommandRequiresTypedConfirmation`
was constructing an app registry. The stack reached `bundleregistry.New -> engine.LoadAll ->
loadStreamSchemas -> CompileSchema`: every in-process `cli.Run` reparsed all embedded connector
bundles. The GitHub parity bundle makes that repeated immutable work large enough to exceed the
deadline. The first cache regression test also failed to compile before the production seam was
added:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/bundleregistry \
    -run '^TestLoadDefinitionsCachesEmbeddedBundleSnapshot$'
registry_test.go:40:16: undefined: loadDefinitions
registry_test.go:44:17: undefined: loadDefinitions
FAIL  polymetrics.ai/internal/connectors/bundleregistry [build failed]
```

**Red — CodeQL reported a high severity allocation-overflow path.** The annotation identified
`make(connectors.Record, len(rec)+1)` in `pinRepoArchived`: even though ordinary records are
bounded in practice, the size arithmetic is unnecessary and can overflow before allocation.

**Green — cache only immutable definitions, never caller registry state.**
`bundleregistry.loadDefinitions` uses `sync.Once` to compile `defs.FS` once per process.
`bundleregistry.New` still creates a fresh registry, fresh engine connectors, and fresh hook
instances on every call, so app/test callers may register custom connectors without sharing
mutable registry maps. The archive helper now allocates with `len(rec)` and lets normal map growth
add the pinned field, preserving the non-mutating record contract without overflow arithmetic.

```text
$ go test -count=1 -timeout 20m ./internal/connectors/bundleregistry \
    -run '^TestLoadDefinitionsCachesEmbeddedBundleSnapshot$'
ok   polymetrics.ai/internal/connectors/bundleregistry  2.064s

$ go test -count=1 -timeout 20m ./internal/connectors/hooks/github \
    -run '^(TestMapWriteRecord_ArchiveRepoPinsArchivedTrue|TestMapWriteRecord_UnarchiveRepoPinsArchivedFalse|TestMapWriteRecord_DoesNotMutateCallerRecord)$'
ok   polymetrics.ai/internal/connectors/hooks/github  0.432s

$ go test -count=1 -timeout 20m ./internal/cli \
    -run '^TestGitHubDestructiveCommandRequiresTypedConfirmation$'
ok   polymetrics.ai/internal/cli  6.499s

$ go test -count=1 -timeout 20m ./internal/cli
ok   polymetrics.ai/internal/cli

$ go test -race -count=1 -timeout 20m ./internal/connectors/bundleregistry
ok   polymetrics.ai/internal/connectors/bundleregistry

$ go vet ./internal/connectors/bundleregistry ./internal/connectors/hooks/github

$ golangci-lint run ./internal/connectors/bundleregistry ./internal/connectors/hooks/github
0 issues.
```
