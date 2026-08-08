# feat(connectors)!: bring github to documented-operation parity and restore its blocked writes

Extracts GitHub's full documented-operation parity from the paused sweep branch
`fm/cli-top50-sweep-resume2-r1` and lands it on `main` on its own, so live end-to-end testing
can run against the complete surface. Then folds in the captain's order to restore the commands
that shipped `unsafe_or_disallowed` without his authorisation.

**461 commands → 1147.** 505 REST endpoint rows → 1224 (1220 documented + 4 synthetic
close/reopen). 1126 endpoints covered, 98 operation-blocked. **1086 commands reachable, every
one verified by running the binary.**

## What was extracted, and how

Four commits, cherry-picked in TDD order:

| commit | |
| --- | --- |
| `6848cbb2d` | red surface test at the derived 1220 (slice 1/5) |
| `afa4a0fd4` | enumerate github's full documented GET surface (slice 2/5) |
| `60719bfbe` | **`fix(connectors): let one endpoint back several write actions`** — a foundation fix, not github-specific |
| `6fe60991d` | bring github to documented-operation parity |

The foundation fix lands here because github is its first consumer, and it is what lets
`repo update`, `archive_repo` and `unarchive_repo` share one `PATCH /repos/{owner}/{repo}` row.

The sweep's `data/cli-top50-fixed-schema-sweep-r1/PROGRESS.md` was dropped — sweep bookkeeping,
not github parity.

### Shared artifacts were regenerated, not hand-merged

The cherry-pick auto-merged `operation_endpoint_ledger.json`. **That merged blob was
discarded**: the file was reset to `main` and rebuilt with `go run ./cmd/connectorgen
surface-sync`. It came back **byte-identical** to the branch's version, which independently
confirms the branch derived it correctly. Website catalogs, connector docs and golden
transcripts were likewise rebuilt from their own generators.

Every delta was diffed **per connector and per transcript**, not eyeballed:

```
website/data/connectors.generated.json                  -> ['GitHub']
website/lib/connectors.catalog.data.generated.json      -> ['GitHub']
internal/connectors/defs/operation_endpoint_ledger.json -> ['github']  (162 -> 164 rows)
golden transcripts -> ['connectors_inspect_github_json', 'dynamic_connector_bare_json']
```

`dynamic_connector_bare_json` is not a stray — its `args` are `["github", "--json"]`.

The doc generators additionally wanted to rewrite all 551 connectors' `MANUAL.md`/`SKILL.md`
and four `pm-*` skill files. That is **pre-existing drift on `main`** (type annotations
rendering as `created_at()` instead of `created_at(string)`, and a missing `## Icon` section),
not github's. It is reverted here and flagged as a follow-up rather than smuggled into a github
diff.

### Reachability was re-derived, not inherited

A prior worker found gmail answering `unknown command` for all 79 operations while our records
claimed success. So the 1079 figure was not taken on trust. Each of the 1086
`implemented`/`partial` commands was invoked as `pm github <path>` in an initialised project
with **no credential configured** — a dispatchable command stops at `error: missing
--credential`, an undispatchable one answers `error: unknown command "..."`:

```
connector=github probed=1079 unreachable=0     (parity extraction, pre-unblock)
connector=github probed=1086 unreachable=0     (final binary)
```

No network call is made; every command stops at the credential gate ahead of any request.

## The `repo create` classification — and the captain's order

The brief asked whether creating a repository is defensible when `issue create` is already
`implemented`. **That turned out not to be the deciding question.**

The parity enumeration already makes every one of these capabilities reachable under a
generated name: `repos create-for-authenticated-user`, `repo delete-2`, `repo update`,
`secret set-2`, `secret delete-2`, `actions caches delete-2` are all `implemented`. So
`unsafe_or_disallowed` on the gh-familiar name was **not a safety control** — it removed no
capability. It only guaranteed the destructive path is the one an operator reaches by accident,
through a generated name, rather than on purpose.

That is why `repo delete` is restored here despite the brief saying to keep it disallowed:
blocking the name while `repo delete-2` runs the same `DELETE` is a naming accident, not a
safety property. The captain's later order asks for it explicitly, and the typed confirmation
is unchanged either way.

### Restored — 7 commands, on the gate the write path already imposed

All become `intent: reverse_etl`, `availability: implemented`, pointed at the write action
their twin already uses, inheriting **plan → preview → approval → execute** unchanged. No gate
was invented; none was relaxed.

| command | write action | method | mutation class | confirmation now required | source of the challenge |
| --- | --- | --- | --- | --- | --- |
| `repo create` | `repos_create_for_authenticated_user` | POST | create | plan → preview → approval token → execute | none — approval-only create |
| `repo delete` | `repo` | DELETE | delete | + closed typed `--confirm destructive`, digest recomputed at execution | DELETE method **and** declared `confirm: destructive` |
| `repo archive` | `archive_repo` *(new)* | PATCH | update | + closed typed `--confirm destructive` | declared `confirm: destructive` |
| `repo unarchive` | `unarchive_repo` *(new)* | PATCH | update | + closed typed `--confirm destructive` | declared `confirm: destructive` |
| `cache delete` | `actions_caches_cache_id` | DELETE | delete | + closed typed `--confirm destructive` | DELETE method |
| `secret set` | `actions_secrets_secret_name3` | PUT | update | plan → preview → approval token → execute; `encrypted_value` redacted | none — approval-only create |
| `secret delete` | `actions_secrets_secret_name` | DELETE | delete | + closed typed `--confirm destructive` | DELETE method |

`connectors.ConfirmationForWriteAction` returns `destructive` for **any** DELETE regardless of
metadata, and the test asserts through that same resolver — so a future edit that severs the
confirmation fails. It asserts the negative too: `repo create` and `secret set` are the captain's
approval-only creates, and gaining a typed challenge is as much a drift from that decision as
losing one. No declaration was duplicated: the three DELETEs already carried the challenge by
method, and only the two PATCH actions needed one declared.

**The documented surface says so too — from one source.** The bundle used to carry a prose
`notes` marker for this on eight commands, and it was wrong on all eight: it read `destructive;
requires --allow-destructive + typed confirmation`, and there is no `--allow-destructive` flag —
`pm github repo delete --allow-destructive` fails with `unknown flag --allow-destructive for
command "repo delete"`.

Those notes are now gone entirely, along with the two that were briefly added to
`repo archive`/`repo unarchive`. A per-command prose marker cannot carry this fact: it is silent
on every command nobody annotated, so with ten of github's 173 destructive commands marked, the
silence on the other 163 read as "no confirmation needed". The requirement is stated once instead,
by the help, derived from the bound write action — see the `--help` section below. All ten notes
were one byte-identical string whose two facts (the `destructive` classification and the
`--confirm destructive` requirement) the derived line states in full, verified command by command
against the built binary before any deletion, so nothing was lost and no patch was retained.
The 107 github commands whose notes say something genuinely per-command keep them.

Every derived artifact was regenerated from those declarations rather than hand-edited:
`docs/connectors/github/{MANUAL,SKILL}.md` and `docs/skills/pm-github/SKILL.md` via
`./pm docs generate` / `./pm skills generate`, `docs/connectors/catalog/all-connectors.json`
(`archive_repo`/`unarchive_repo` gain `confirm: destructive`, taking github from 181 to 183
declared confirmations), and `website/data/connectors.generated.json` plus
`website/lib/connectors.catalog.data.generated.json` via `npm --prefix website run gen:catalog`.
`body_fields: ["archived"]` now also renders as an `optional fields: archived` line on both
actions, taking github's manual from 9 such lines to 11. The generators again wanted to rewrite
all 551 connectors, plus the catalog's `gorgias` and `warehouse` entries; that is the same
pre-existing drift on `main` the earlier regeneration commit isolated, and it is reverted here
again. Both website catalogs were diffed per connector: the only entry that changes is GitHub.

**`--help` states it for every destructive command.** `pm <connector> <command> --help` renders a
`CONFIRMATION` section resolved by `commandrunner.ConfirmationChallengeForCommand`, the same
declarations `buildWriteCommand`/`buildOperationDirectWriteCommand` read at plan time, so help and
runtime cannot disagree — the boundary `writeConnectorDownloadFlags` already sets for download
flags. It covers all 173 of github's destructive commands and every other connector's, and it is
the only place the fact is stated:

```
pm github release delete --help
-> APPROVAL      Reverse ETL writes require plan, preview, approval, execute.
   CONFIRMATION  execution requires the typed confirmation --confirm destructive
```

`TestGitHubNotesDoNotRestateTheTypedConfirmation` sweeps every github command and fails on any
`notes` that mentions `--confirm`, so the removed copy cannot return; it also asserts the resolver
still answers, so it cannot pass by the derivation silently breaking.

`guide.go` is deliberately untouched: marking all 183 actions there would rewrite all 551
committed connector manuals, which is both the bulk change that was declined and
unrelated-connector churn.

The runtime plan output already carried this signal exhaustively — `pm github release delete`
prints `Confirmation required: --confirm destructive` with no note in its bundle — so the
described failure (plan, then an unexplained error at run) never occurred. `--help` was the one
generated path that could mislead, and it is the one that changed.

**Two needed more than a classification change.** `repo archive`/`repo unarchive` ride the same
endpoint as the generic `repo update`, so the body is the only thing separating them — an
"archive" command that only archives when the caller separately remembers `archived: true` is a
command that lies.

They were first pinned in the github hook's `ExecuteWrite`, as `close_issue`/`reopen_issue` are.
That is incompatible with the typed confirmation the captain ordered: `prepareDeclarativeWrite`
**refuses** to prepare a destructive action whose hook overrides execution, because the preview an
operator approves would not be the request that runs. Declaring `confirm: destructive` alone would
have made both commands unusable rather than gated.

So the pin moved from the request to the record. A new optional `engine.WriteRecordHook` lets a
hook fix the body fields an action's own name implies, before the declarative body is built;
preview and execution both apply it, so the approved body **is** the dispatched body by
construction. `archive_repo`/`unarchive_repo` declare `body_fields: ["archived"]`, so the pinned
field is also the only field sent — the same tightness the hook-executed version had, with an
exact preview it did not. The interface is opt-in, so the other 550 connectors are untouched.

**`secret set` exposed a live defect.** The already-`implemented` `secret set-2` had a
`record_schema` of just the path parameter, so it could only ever send an empty PUT and take a
422. `encrypted_value` and `key_id` are now declared and flagged on both commands.
`pm` does not encrypt and does not store the value — the caller seals it with the repository
public key, as the API requires — and `encrypted_value` is in `redact_fields`.

### Not restored — 3 commands, each needs a runtime capability this PR does not add

Marking any of these `implemented` would fail at runtime. That is the claim-before-establish
defect this project keeps finding, so they stay blocked and are reported instead.

| command | why |
| --- | --- |
| `issue delete` | Declared as `kind: graphql_mutation`; `engine.operationDirectWriteSpec` requires `rest_write`. GitHub has no REST issue-delete endpoint — `deleteIssue` is GraphQL-only and takes a node ID. Needs a GraphQL write executor plus node-ID resolution. |
| `issue transfer` | GraphQL-only. `POST /repos/{owner}/{repo}/transfer` is **repository** transfer — wiring it there would be a different operation wearing this command's name. |
| `pr revert` | No documented REST endpoint at all; absent from github's own `api_surface.json`. `AGENTS.md`: never invent an endpoint to make a command look implemented. |

**bahmni `documents upload`** is the same category and also stays blocked: it has no write
action, and its own notes name the missing capability — "current inline JSON content surface
lacks the claimed file snapshot/SHA-256 approval binding".

### Held for the captain

**`auth token` — held, as ordered.** It prints a credential; runtime output is never stripped,
so a printed token stays printed. That contradicts a rule the captain set, so it waits on him.

**`api` — reported, not acted on.** Unblocking the arbitrary authenticated-request escape hatch
means: no declared parameters, no enum validation, no required-field enforcement, no operation
ledger entry (so no `kind`, no `max_bytes`, nothing bounding the response), and no `api_surface`
row — bypassing the very 1220-endpoint surface this PR exists to enumerate, and every gate keyed
on it, including the destructive-write confirmation. One `pm github api` call reaches every
endpoint the other 1146 commands model, under none of their controls. **The captain's call.**

`TestGitHubHeldCommandsStayBlocked` pins both so neither drifts to `implemented` by accident.

## ⚠️ This PR also contains a shared-code security change

Not a github change. It was found during review of this branch and closed here on the captain's
ruling, because restoring `secret set` is what made it reachable.

**The defect.** A reverse-ETL plan built from a connector command ignored the resolved write
action's `redact_fields`. The declared-sensitive value was therefore both emitted in `--json` and
persisted verbatim to the project state file. `RedactFields` was populated correctly on the
sibling source-table path 130 lines earlier, so this was a competing owner, not an architectural
choice.

**Why it could not wait.** Reverse plans are **append-only** — the only writes to
`a.state.ReversePlans` are two appends — so `ExpiresAt` gates a plan's *usability*, never the
retention of its bytes. The value stayed on disk until the operator deleted project state.
Retention was unbounded.

**How wide.** Independently re-counted from the bundles, not taken from the review:
**170 `implemented` reverse-ETL commands across 14 connectors** have a CLI flag feeding a field
their own write action declares in `redact_fields` — ashby 87, mailchimp 20, asana 10,
amazon-sqs 8 (including `message_body` and `receipt_handle`), recurly 8, google-calendar 7,
freshchat 5, zendesk-support 5, hubplanner 5, youtube-analytics 5, google-search-console 4,
bahmni 3, github 2, stripe 1.

**What changed**, in three commits:

1. `3f78743ad` — populate `RedactFields` from the resolved write action, so the existing
   `RedactReversePlanRecords` boundary stops being a no-op. Closes the `--json` half.
2. `148b763b1` — withhold those fields from the persisted `ConnectorCommandRecord` entirely
   (absent, not the literal `"redacted"`, which is indistinguishable from an operator who typed
   it) and re-supply them at approve/execute. The plan hash already covers the complete record,
   so a missing field, a wrong value and a tampered plan all fail on machinery that already
   existed. No new binding, no new CLI surface — `runConnectorWriteCommandFromPlan` already
   receives the parsed flags. Closes the at-rest half.
3. `133d7174a` — resolve the withholding key **by mode**: operation-backed plans read the
   operation's `sensitive_policy.redact_fields`, write-action plans read the manifest, with
   **no fallback between them**. `buildOperationDirectWriteCommand` sets `Write` to an operation
   ID, and the two namespaces are disjoint in 550 of 551 bundles, so the previous lookup returned
   nil and withholding silently did nothing on that path. asana is the exception and was worse:
   11 names (`delete_task`, `delete_project`, `create_tag`, …) exist in both namespaces, so a
   fallback could apply an unrelated write action's redact list. Mode dispatch makes that
   unreachable rather than merely unlikely.

That third commit is why the 4 OAuth endpoints are **not** implemented in this PR: they take a
live bearer token as a required body field and would be the first `implemented` `direct_write`
command in the repo — the exact path that silently resolved to nil.

**A third declaration site is deliberately left alone.** `CommandSurfaceCommand.RedactFields` is
declared by 217 commands and documented at `connectors/command_surface.go` as not consulted by
`commandrunner`. It stays that way and is filed, not wired: two redaction sources feeding one
path is how this bug class starts.

Tests assert against the **raw persisted bytes** of `state.json`, not an in-memory sub-slice —
asserting on the sub-slice is exactly how the at-rest half survived the first fix round. Coverage
includes the sealed-secret case, a required-and-withheld bearer-token fixture, the asana
same-name collision resolving to the operation, and negative controls (no `redact_fields`
round-trips unchanged; a wrong re-supplied value fails the hash check instead of dispatching).

## ⚠️ The sweep branch needs a rebase after this merges

`fm/cli-top50-sweep-resume2-r1` carries these same four github commits. Once this lands, that
branch **must be rebased onto the new `main` before it resumes** — its github commits will be
already-applied, and its `operation_endpoint_ledger.json` will conflict against the regenerated
one. The sweep is paused at stripe, so this is cheap now. Flagging it so it is not a surprise.

## Follow-ups deliberately not fixed here

1. **Repo-wide doc drift on `main`** — `pm docs generate`/`pm skills generate` want to rewrite
   all 551 connectors' docs plus four `pm-*` skill files. Needs its own PR.
2. **Thin `record_schema`s on generated write actions** — `secret set-2` was `implemented`
   while able only to send an empty body. That is the shape issue #3899 exists to fix; only
   the one blocking this order was fixed. `repo2` and others are worth a sweep.
3. **`CommandSurfaceCommand.RedactFields`** — a third, deliberately unconsulted redaction
   declaration site on 217 commands. Filed, not wired. Needs a deliberate decision before
   anything reads it.
4. **github's 98 blocked endpoints** — a separate stacked PR, scoped and pre-derived. See below.

## The 98 blocked endpoints are a separate PR, and here is why

The captain ordered complete parity over the 98, excluding only genuine duplicates and genuine
missing foundations. That work is scoped, its data is derived and verified, and it is **not** in
this PR: this branch already grew from a github extraction into a shared-code security change
touching 170 commands across 14 connectors, and the remaining work adds ~29 endpoint
implementations plus two new runtime output policies. Splitting costs nothing — the OAuth four
land after the resolver fix they depend on, which is in this PR.

**The foundation question, answered.** `covered_by.writes` (`997d7391b`, in this PR) already
expresses one endpoint as N write actions **with different bodies** — that phrase is from its own
commit message, describing `update_issue`/`close_issue`/`reopen_issue`. It needs no extension for
multi-arm request bodies. `ValidatePromotableRecordSchema` rejects a union root with the literal
instruction *"declare a separate named write action for each arm"*, and `expandRecordSchemaArms`
already merges the wrapper base into each arm. This PR ships the existence proof:
`PATCH /repos/{owner}/{repo}` backs three write actions with `validate` at 0 findings.

But the 12 union-bodied endpoints are not homogeneous, and splitting all of them would be a
modelling error: 8 are genuine alternatives with disjoint `required` sets (→ 19 actions), 2 are
`anyOf`-as-*at-least-one* (all six properties sit on the base; splitting would manufacture five
commands each claiming to be the only way to set one field), and 2 are root-type polymorphism
(`object | array | string`, where a record is an object contract).

**Two genuine foundation gaps** remain, both on the read side: `validateDirectReadOutputPolicy`
has no status-only policy for the 9 `204 No Content` boolean checks, and no text policy for
`text/html`, `text/plain` or `application/octocat-stream` (`/markdown`, `/markdown/raw`, `/zen`,
`/octocat`). `direct_write` already has `none`, so the gap is specifically reads.

**One genuine exclusion:** `POST /app/installations/{installation_id}/access_tokens` mints a
credential and is already consumed internally by the `github_app` AuthHook — same class as the
held `auth token`.

## Verification

### Fresh restart validation (2026-08-08)

This branch was recovered at `b756c9c63feae44c91a79ab9e11d27e8c7fffd11`
without reset or reconstruction. The fresh validation did not inherit the prior
run's result: it built a new `pm` binary and invoked all 1,086
`implemented`/`partial` GitHub paths from an initialized project with no
credential configured. Result: `connector=github probed=1086 unreachable=0`.

The GSD adapter and canonical contract were re-checked; `discuss-phase`,
`plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` prompts
were resolved through `scripts/gsd prompt`. Execution remains an inline/manual
fallback because this runtime cannot provide the required isolated Pi roles and
the project contract forbids spawning them. The existing plan and red/green
ledger are retained, with the new verify evidence in `VERIFICATION.md` section
7 and `RUN-STATE.json`.

Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, `golang-structs-interfaces`, `golang-graphql`,
`golang-lint`, `golang-documentation`, `vercel-react-best-practices`, and
`vercel-composition-patterns`. The routing reference also named
`frontend-design` and `web-design-guidelines`; neither is installed, so the
generated website artifacts were checked through the repository's own docs and
generator gates.

```
gofmt -l cmd internal                                          clean
go vet ./...                                                   clean
go test ./cmd/connectorgen/                                    ok
go test ./internal/connectors/engine/                          ok
go test ./internal/connectors/conformance/                     ok
go test ./internal/connectors/commandrunner/                   ok   (incl. TestEveryImplementedCommandPassesRuntimePreflight)
go test ./internal/connectors/hooks/github/                    ok
go test ./internal/cli/                                        ok
go run ./cmd/connectorgen validate internal/connectors/defs    551 checked, 0 findings
go run ./cmd/connectorgen surface-sync --check                 no drift
make lint                                                      0 issues
make connector-boundary                                        ok
make agent-contract-check                                      contract and projections current
make docs-check                                                ok
go build ./cmd/pm                                              ok
binary reachability sweep                                      1086 probed, 0 unreachable
```

`go test ./...` and `make verify` were not run as single commands: per `AGENTS.md` the full
suite spans 551 connectors and routinely exceeds an agent's per-command timeout, where a cutoff
is indistinguishable from a hang. Scoped package runs plus each `make verify` gate individually
were run instead; CI carries the whole suite.

**No test was weakened, skipped, or deleted. Three were added.**

GSD evidence: `.planning/phases/github-parity-extract-r1/` — `PLAN.md`, `TDD-LEDGER.md`
(both red/green cycles with observed output), `VERIFICATION.md`, `SUMMARY.md`,
`CLASSIFICATION-REPORT.md`, `RUN-STATE.json`.
