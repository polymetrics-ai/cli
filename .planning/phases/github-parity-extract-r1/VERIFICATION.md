# VERIFICATION — github-parity-extract-r1

## 1. The shared artifacts were regenerated, not hand-merged

The cherry-pick of `6fe60991d` auto-merged `operation_endpoint_ledger.json`. That merged blob
was **discarded**: the file was reset to `main`'s content (`08cc41c87`) and regenerated.

```
$ git checkout 08cc41c87 -- internal/connectors/defs/operation_endpoint_ledger.json
$ go run ./cmd/connectorgen surface-sync
runtime operation endpoint ledger: updated 1978 endpoint(s)
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected
```

The regenerated ledger is **byte-identical** to the one the branch carried (`git diff HEAD`
on that path is empty), which is the strongest available evidence that the branch's ledger was
itself correctly derived and that nothing was lost in the extraction.

Other generators, all run from their own entry points:

| artifact | generator |
| --- | --- |
| `docs/connectors/github/{MANUAL,SKILL}.md` | `./pm docs generate --dir docs/cli` |
| `docs/skills/pm-github/SKILL.md` | `./pm skills generate --dir docs/skills` |
| `website/data/connectors.generated.json`, `website/lib/connectors.catalog.data.generated.json` | `node scripts/gen-connector-bundles.mjs`, `gen-connector-catalog.mjs`, `gen-connectors.mjs` |
| `internal/cli/testdata/golden_transcripts.json` | `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli/ -run GoldenTranscript` |

## 2. The delta is confined to github

Every generated artifact was diffed **per connector / per transcript**, not by eyeballing the
patch:

```
website/data/connectors.generated.json:                 ['GitHub']
website/lib/connectors.catalog.data.generated.json:     ['GitHub']
internal/connectors/defs/operation_endpoint_ledger.json:['github']   (162 -> 164 rows)
golden transcripts:  ['connectors_inspect_github_json', 'dynamic_connector_bare_json']
```

`dynamic_connector_bare_json` is not a false positive — its `args` are `["github", "--json"]`.

The doc generators additionally wanted to rewrite all 551 connectors' `MANUAL.md`/`SKILL.md`
plus four `pm-*` skill files. That is **pre-existing drift on main** (field type annotations
the committed docs render as empty parens, e.g. `created_at()` -> `created_at(string)`, and a
missing `## Icon` section), unrelated to github. It was reverted and left for its own PR rather
than smuggled into a github diff.

## 3. Every reachable command was verified by running the binary

Not inherited from the branch. A prior worker found gmail answering `unknown command` for all
79 operations while the records claimed success, so the figure was re-derived by invoking the
built `pm` binary once per command.

Method: for each of the 1086 github commands with `availability` in {`implemented`, `partial`},
run `pm github <path>` inside an initialised project with no credential configured, and check
the output for `unknown command`. That string is the exact discriminator — a dispatchable
command falls through to `error: missing --credential`, an undispatchable one does not:

```
$ pm github issue list          ->  error: missing --credential      (reachable)
$ pm github nosuchgroup cmd     ->  error: unknown command "nosuchgroup cmd"
```

Each of the 8 workers ran in its own project directory, because a shared one produces
state-lock races that are not reachability failures.

**Result (parity extraction, before the unblock):** `probed=1079 unreachable=0`.
**Result (final binary, after the unblock):** recorded in `RUN-STATE.json`.

No network call is made: with no credential configured every command stops at the credential
gate, ahead of any request.

## 4. Runtime honours every `implemented` claim

`TestEveryImplementedCommandPassesRuntimePreflight` sweeps every bundle in `defs.FS` through
the real `commandrunner.Preflight`, so the 7 restored commands are covered by the repo's own
guard and not only by the test added here. It passes.

## 5. Local gates

```
gofmt -l cmd internal                                   clean
go vet ./...                                            clean
go test ./cmd/connectorgen/                             ok
go test ./internal/connectors/engine/                   ok
go test ./internal/connectors/conformance/              ok
go test ./internal/connectors/commandrunner/            ok
go test ./internal/connectors/hooks/github/             ok
go test ./internal/cli/                                 see RUN-STATE.json
go run ./cmd/connectorgen validate internal/connectors/defs   551 checked, 0 findings
go run ./cmd/connectorgen surface-sync --check          no drift
make lint                                               0 issues
make connector-boundary                                 ok
make agent-contract-check                               contract and projections current
make docs-check                                         ok
go build ./cmd/pm                                       ok
```

`go test ./...` and `make verify` were not run as single commands: per `AGENTS.md`, the full
suite spans 551 connectors and routinely exceeds an agent's per-command timeout, where a cutoff
is indistinguishable from a hang. Scoped package runs plus the individual `make verify` gates
were run instead, and CI carries the whole suite.

## 5b. Gates re-run locally on the recovered head (`133d7174a`)

The `no-mistakes` run reached `review: completed, 0 findings` and then failed at the test step on
the agent's monthly spend limit — an external blocker, not a code failure. Its five commits were
recovered with `no-mistakes axi sync --recover` and the gates re-run here:

```
go build ./cmd/pm                                              ok
gofmt -l cmd internal                                          clean
go vet ./...                                                   clean
go test ./internal/app/                                        ok  (251.4s)
go test ./internal/connectors/commandrunner/                   ok
go test ./internal/connectors/hooks/github/                    ok
go test ./cmd/connectorgen/                                    ok
go run ./cmd/connectorgen validate internal/connectors/defs    551 checked, 0 findings
go run ./cmd/connectorgen surface-sync --check                 no drift
make lint                                                      0 issues
make docs-check                                                ok
make connector-boundary                                        ok
make agent-contract-check                                      contract and projections current
bash scripts/verify-gsd-workflow                               exit 0
```

## 5c. The shared-code redaction exposure, re-counted independently

The review reported 170 commands across 14 connectors. That number was **re-derived from the
bundles here rather than taken on trust**, by matching every `implemented` reverse-ETL command's
flag `maps_to` against its own write action's `redact_fields`:

```
commands with a flag feeding a redact_fields field: 170
connectors: 14
  ashby 87, mailchimp 20, asana 10, amazon-sqs 8, recurly 8, google-calendar 7,
  freshchat 5, zendesk-support 5, hubplanner 5, youtube-analytics 5,
  google-search-console 4, bahmni 3, github 2, stripe 1
```

Exact match. Every one of those persisted declared-sensitive values to `state.json` before this
change, permanently, because reverse plans are append-only.

## 6. Surface totals

| | main (`08cc41c87`) | this branch |
| --- | --- | --- |
| github commands | 461 | 1147 |
| `implemented` | — | 1049 |
| `partial` | — | 37 |
| reachable (`implemented` + `partial`) | — | 1086 |
| `unsafe_or_disallowed` | — | 5 |
| REST endpoint rows | 505 | 1224 (1220 documented + 4 synthetic close/reopen) |
| covered endpoints | — | 1126 |
| operation-blocked endpoints | — | 98 |
| ledger rows for github | 162 | 164 |

## 7. Fresh restart validation — 2026-08-08

This recovery run started from the remote-backed branch head
`b756c9c63feae44c91a79ab9e11d27e8c7fffd11`; it did not rebuild, reset, or
rebase the extracted work. It re-ran the delivery checks rather than inheriting
the predecessor's green claims.

The repo-local GSD adapter was healthy (`scripts/gsd doctor`) and the canonical
contract was current (`go run ./cmd/agentcontractgen check`). The following
official command prompts were resolved: `discuss-phase github-parity-extract-r1
--auto`, `plan-phase github-parity-extract-r1 --tdd --skip-research`,
`execute-phase github-parity-extract-r1 --interactive`, `verify-work
github-parity-extract-r1 --auto`, and `code-review github-parity-extract-r1
--depth=standard`. Their work was performed inline: this Codex runtime cannot
provide the Pi-isolated workflow roles and the project contract forbids spawning
separate GSD roles for this job. The existing PLAN/TDD cycles remain the
production-change evidence; this section is the fresh verify-work execution.

### Direct binary reachability

Built a new `pm` binary, initialized an empty throwaway project with no
credential, and invoked every GitHub command whose availability is `implemented`
or `partial`. The harness treats `unknown command` as the only dispatch failure;
credential-gated and declared-partial responses are reachable by design. It made
no provider request.

```
connector=github probed=1086 unreachable=0
```

### Fresh local gates

```
gofmt -l cmd internal                                              clean
go vet ./...                                                       clean
go test -count=1 -timeout 20m ./cmd/connectorgen/                 ok (12.097s)
go test -count=1 -timeout 20m ./internal/connectors/engine/       ok (4.934s)
go test -count=1 -timeout 20m ./internal/connectors/commandrunner/ ok (13.034s)
go test -count=1 -timeout 20m ./internal/connectors/hooks/github/ ok (0.875s)
go test -count=1 -timeout 20m ./internal/connectors/conformance/  ok (21.038s)
go test -count=1 -timeout 20m ./internal/app/                     ok (266.614s)
go test -count=1 -timeout 20m ./internal/cli/                     ok (743.205s)
go run ./cmd/connectorgen validate internal/connectors/defs        551 checked, 0 findings
go run ./cmd/connectorgen surface-sync --check                     no drift
make tidy-check, lint, docs-check, smoke-no-build                  ok
make agent-contract-check, connectorgen-validate, connectorgen-surface-sync  ok
make connector-boundary, release-workflow-check                    ok
bash scripts/verify-gsd-workflow                                  exit 0
go build ./cmd/pm                                                  ok
```

## Review round (cycle 4) re-verification

Five review findings were investigated; all five were legitimate and are fixed
in place. See TDD-LEDGER cycle 4 for the red/green record. Two of them
(4a withheld re-supply, 4c plural write coverage) were shared-code regressions
this branch introduced outside GitHub, so the fixes are at the shared boundary
rather than at the reported line: `withholdRecordFields` now reports what it
actually removed and `ReversePlan.WithheldFields` persists it, and the certify
surface inventory reads `SurfaceCoverage.WriteTargets()` like every other
consumer of `covered_by`.

`internal/connectors/certify/` was absent from the gate list above, which is why
its two stale GitHub counts were red on this branch without being noticed. It is
listed below and stays listed.

```
gofmt -l cmd internal                                              clean
go vet ./internal/app/ ./internal/cli/ ./internal/connectors/certify/ ./internal/connectors/commandrunner/   clean
go build ./cmd/pm                                                  ok
go test -timeout 20m ./internal/app/                               ok (265.262s)
go test -timeout 20m ./internal/connectors/certify/                ok (12.555s)
go test -timeout 20m ./internal/connectors/commandrunner/          ok (13.315s)
go test -timeout 15m -run 'Golden|Reverse|Docs|Manual|Help' ./internal/cli/   ok (174.584s)
pm docs generate                                                   docs/cli/reverse.md only
POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test -run TestGoldenTranscripts ./internal/cli/   3 reverse entries
node website/scripts/gen-docs-data.mjs                             reverse-etl page only
```

CLI/docs/website parity for the re-supply contract: `pm help reverse`,
`pm reverse` and `pm reverse --json` all render the new USAGE, FLAGS,
DESCRIPTION, COMMANDS and SECURITY text (pinned by the regenerated golden
transcripts); `docs/cli/reverse.md` and `website/content/docs/reverse-etl.mdx`
plus its generated data carry the same contract.

## Review round (cycle 5) re-verification

Both review findings were legitimate and are fixed. The ancestor-subtree gap was
the residual of cycle 4's own fix, so it is closed at the same shared boundary
that owns "rebuild the record fragment from the same flags with the same
coercion rules" -- `commandrunner.ReconstituteWithheldFields` -- rather than at
the recurly bundle that exposed it. The docs finding changed no behaviour.

```
gofmt -l cmd internal                                              clean
go vet ./internal/cli/ ./internal/connectors/commandrunner/        clean
go build ./cmd/pm                                                  ok
go test -run 'Withheld|Reconstitute' ./internal/connectors/commandrunner/   3 pass
go test -timeout 20m ./internal/connectors/commandrunner/          ok
go test -run TestGoldenTranscripts ./internal/cli/                 ok (70.267s)
pm docs generate                                                   docs/cli/reverse.md only
node website/scripts/gen-docs-data.mjs                             reverse-etl page only
pm recurly invoices retries create ... --preview                   plan created, subtree withheld
```

## Review round (cycle 10) re-verification

The finding raised for the captain's decision was answered: `repo delete` keeps
its restored, destructive, `implemented` classification, and `repo create`,
`repo archive`, `repo unarchive` and `secret set` stay non-destructive
approval-only writes. `issue delete` stays blocked -- there is no GraphQL
mutation executor to reach. Nothing was reclassified; the four corrections
attached to that decision were implemented instead, each one closing a gap
between what the surface declares and what the runtime enforces.

```
go test ./internal/connectors/engine/                              ok (15.571s)
go test -timeout 20m ./internal/connectors/commandrunner/          ok (221.071s)
go test -timeout 20m ./internal/app/                               ok (335.570s)
go test -timeout 20m ./internal/cli/                               ok (962.621s)
go test -timeout 20m ./internal/connectors/conformance/            ok (18.151s)
go test -timeout 20m ./internal/connectors/                        ok (0.520s)
go test -timeout 20m ./internal/connectors/certify/                ok (17.764s)
go test -timeout 20m ./internal/connectors/hooks/github/           ok (1.258s)
go test -timeout 20m ./internal/connectors/native/amazon-sqs/      ok (1.139s)
go test -timeout 20m ./internal/connectors/boundary/               ok (220.497s)
go build ./cmd/pm                                                  ok
connectorgen validate internal/connectors/defs                     551 checked, 0 findings
connectorgen surface-sync --check                                  551 scanned, 0 drift
pm docs validate --connectors-dir docs/connectors                  validated
```

Product-level proof of the corrected approval metadata, against a
request-logging mock GitHub API in a throwaway project:

```
safe write, preview skipped entirely
  pm github repo create --name demo-repo        -> Approval token issued at plan
  pm github repo create --plan ... --approve ...-> succeeded=1 failed=0
  mock received: POST /user/repos BODY={"name":"demo-repo"}

destructive write, same two steps
  pm github repo delete                         -> Preview required before an approval
                                                   token is issued.
                                                   Confirmation required: --confirm destructive
  pm github repo delete --plan ... --approve ...-> error: ... must be previewed before
                                                   approval  (0 requests dispatched)
  ... --preview                                 -> Approval token issued
  ... --approve <token>                         -> error: requires typed confirmation:
                                                   pass --confirm destructive  (0 requests)
  ... --approve <token> --confirm destructive   -> succeeded=1 failed=0  (1 DELETE)
```

`pm github repo delete --help` and `pm github repo delete-2 --help` render an
identical WRITE/APPROVAL/CONFIRMATION block, and `pm github repo create --help`
renders the corrected APPROVAL line with no CONFIRMATION section -- the
approval-only create the captain classified, unchanged.

Derived artifacts regenerated and confined to github: `docs/connectors/github/{MANUAL,SKILL}.md`,
`docs/skills/pm-github/SKILL.md`, `website/data/connectors.generated.json` and
`website/lib/connectors.catalog.data.generated.json`, each exactly 349 changed
lines. The other 1031 doc files' generator drift is the pre-existing `main`
condition and was reverted, as in every previous cycle.

## Review round (cycle 11) re-verification

Cycle 10 raised one item rather than resolving it: the decision lists
`repo archive` and `repo unarchive` among the commands to keep non-destructive,
and on this branch both were destructive by cycle 6's declared
`confirm: destructive`. Removing a gate that exists and is tested is not
something to do on an ambiguous reading, so it was flagged. The captain
confirmed the non-destructive classification, and cycle 11 applies it: the two
`confirm` declarations are deleted from
`internal/connectors/defs/github/writes.json`, and nothing else is declared.

Red first, then green (full output in `TDD-LEDGER.md` cycle 11):

```
go test ./internal/connectors/commandrunner/ -run TestGitHubRestoredCommandsAreExecutable
  RED   repo archive / repo unarchive reach PATCH ... with confirmation "destructive"
  GREEN ok
go test ./internal/app/ -run GitHubRepoArchive
  RED   repo archive ConfirmationChallenge = "destructive", want none
  GREEN ok
go test ./internal/cli/ -run 'TestConnectorCommandHelp(States|Omits)...'
  RED   pm github repo unarchive --help printed the CONFIRMATION section
  GREEN ok
```

Automated tests do not show the operator-facing contract, so the whole path was
walked with the built binary against a mock GitHub API that logs every dispatched
request, in a scratch project with a mock credential:

```
$ pm github repo archive --help
  APPROVAL      Reverse ETL writes require plan, approval, execute; preview is optional.
  (no CONFIRMATION section)

$ pm github repo archive --credential github-local
  Created connector command plan rplan_17ab35fc3e572543 for repo archive
  Approval token: 81c18f3e...                  <- issued at plan; no preview demanded
  requests dispatched: 0

$ pm github repo archive ... --plan rplan_17ab... --preview --json
  plan.confirmation                          = {"kind": ""}
  write_preview.approval_target.confirmation = {"kind": ""}

$ pm github repo archive ... --plan rplan_17ab... --approve 81c18f3e...   # no --confirm
  Reverse ETL run rrun_00a2... completed: succeeded=1 failed=0
  mock received: PATCH /repos/acme/widgets BODY={"archived":true}

$ pm github repo unarchive ... --plan rplan_f4ff... --approve <wrong token>
  error: approval token is invalid            (requests dispatched: 0)
$ pm github repo unarchive ... --plan rplan_f4ff... --approve 924b380a...
  succeeded=1 failed=0
  mock received: PATCH /repos/acme/widgets BODY={"archived":false}
```

The approval gate is intact — an invalid token is refused with zero requests
dispatched — and the pinned `archived` field still separates the two commands,
so the `WriteRecordHook` keeps doing the job it was built for even though the
confirmation that forced its introduction is gone. The path resolves through
`{{ config.owner }}/{{ config.repo }}` to the configured repository.

`repo delete` is untouched and re-confirmed on the same binary:

```
$ pm github repo delete --credential github-local
  Created connector command plan rplan_af16... for repo delete
  Preview required before an approval token is issued.
  Confirmation required: --confirm destructive
```

Derived artifacts regenerated and confined to github --
`internal/connectors/defs/github/cli_surface.json`,
`docs/connectors/github/{MANUAL,SKILL}.md`, `docs/skills/pm-github/SKILL.md`,
`docs/connectors/catalog/all-connectors.json`,
`website/data/connectors.generated.json`,
`website/lib/connectors.catalog.data.generated.json` and
`internal/cli/testdata/golden_transcripts.json` -- two changed lines each, with
github's catalog entry verified byte-identical to generator output. The golden
transcript was the one gate that caught the artifact it owns:
`TestGoldenTranscripts/connectors_inspect_github_json` failed on the stale
`confirm` lines before being regenerated with
`POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1`, and its diff is exactly those two
lines. The
other 1031 doc files' generator drift is the pre-existing `main` condition and
was reverted, as in every previous cycle. `connectorgen validate` (551 checked,
0 findings), `connectorgen surface-sync --check` (no drift) and
`pm docs validate` are clean.
