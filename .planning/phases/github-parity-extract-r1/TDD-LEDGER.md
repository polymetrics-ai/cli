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
