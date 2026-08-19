# GitHub read certification scope — r1

## Decision and accounting rule

This is a read-side certification scope, not a claim that every GitHub command
or stream has executed.  A row may contribute to a `pass` roll-up only after it
has a definition-owned produced-value assertion and a recorded live result.
Until a candidate has a matching live result, its status is
`eligible_pending_live`, `blocked`, `fixture_required`,
`blocked_identity_or_entitlement`, or `not_applicable`; none is a pass.

The inventory source is `internal/connectors/defs/github/cli_surface.json` and
`streams.json`.  Counts below are mutually exclusive and reconcile separately:

| Declared command bucket | Status | Count | Exact classification / reason |
| --- | --- | ---: | --- |
| Selected fixture-root direct reads | `pass` | 23 | The exact declaration-owned candidate set in `certification.json`: `repo read-file`, `readme view`, `branches view`, `git ref view`, `commits view`, `commits status view`, `commits statuses view`, `activity view`, `assignees view-2`, `events view`, `topics view`, `community profile view`, `hash-algorithm view`, `teams view`, `actions access view`, `actions caches view`, `actions downloads view`, `actions permissions view`, `actions workflow view`, `actions runners view`, `actions secrets view`, `actions variables view`, and `rulesets rule-suites view`. The focused live run passed all 23, each with a direct-read result-kind check, one or more produced-value/type assertions, and a secret-output scan. |
| Other direct reads rooted only in the run-owned repository or organization | `blocked: no r1 stable output assertion` | 100 | Their only target is the disposable `owner`/`repo` (or declared organization ID), but this r1 candidate set has no stable provider value/type invariant for them. They are not counted as pass merely because they could be invoked. |
| Direct reads requiring a typed, disposable fixture resource | `fixture_required` | 372 | 287 repository-root plus 82 organization-root operations need a stable value such as an issue, run, asset, workflow, branch, team, or other resource ID. `repo read-dir` returned a live `Error` for both root spellings and requires a known nonempty fixture directory; the two check-run/check-suite commands likewise need a compatible disposable check fixture and permission contract before they can pass. |
| Direct reads requiring unavailable identity or product entitlement | `blocked_identity_or_entitlement` | 144 | 23 GitHub App/installation-authenticated operations cannot use the disposable PAT identity; 120 feature/enterprise/security-product operations have no declared entitlement in this fixture; `subscription view` additionally needs a user-specific watch relationship and returned a live `Error` without one. |
| Direct reads unavailable in the declared CLI surface | `not_applicable` | 16 | 15 `unsupported_api` plus 1 `unsupported_local`; no executable direct-read contract exists. |
| Binary-download commands | `not_applicable` | 11 | Read-like commands, but the current binary stage deliberately classifies its executor as blocked and cannot honestly produce a live-pass result. |
| ETL read entrypoints | `delegated_to_stream_scope` | 15 | These are stream execution entrypoints, not direct-read candidates; their runtime read boundary is accounted for in the separate stream ledger below. |
| All other command intents | `out_of_read_scope` | 890 | 292 direct-write, 577 reverse-ETL, 14 local workflow, 3 auth, 3 config, and 1 raw API command.  They are not evidence for this read wave. |
| **Declared commands** |  | **1,571** | **23 + 100 + 372 + 144 + 16 + 11 + 15 + 890** |

The 639 implemented direct reads derive from the previously reviewed canonical
GitHub live classifier's target cohorts, which remain at 639 in the current
surface: 64 run-owned repository roots, 63 run-owned organization roots, 38
feature roots, 4 App-installation roots, 287 repository typed-resource reads,
82 organization typed-resource reads, 82 feature/enterprise reads, and 19
App-installation reads. This scope moves the 38 + 82 feature rows and 4 + 19
App rows to non-pass statuses rather than treating a provider 403/404 or an
ambient identity as a certification success. The 23 + 100 + 372 + 144 split
reconciles that same direct-read total after real candidate results changed
four initially-root-only commands into fixture/identity requirements.

| Declared stream bucket | Status | Count | Streams / reason |
| --- | --- | ---: | --- |
| Repository-root collection streams | `blocked: no r1 stream output assertion` | 29 | `repository`, `issues`, `pull_requests`, `branches`, `commits`, `tags`, `releases`, `labels`, `milestones`, `issue_comments`, `pull_request_review_comments`, `collaborators`, `contributors`, `stargazers`, `subscribers`, `workflows`, `workflow_runs`, `workflow_artifacts`, `deployments`, `commit_comments`, `deploy_keys`, `webhooks`, `environments`, `forks`, `invitations`, `issue_events`, `repo_rulesets`, `autolinks`, `languages`. They remain non-pass until a stream candidate can assert a stable produced record rather than only ETL exit status. |
| GraphQL streams requiring typed fixture values | `fixture_required` | 4 | `projects`, `project_items`, `discussions`, `discussion`; these require a fixture project/discussion or its opaque node ID. |
| Product-security streams without the declared entitlement | `blocked_identity_or_entitlement` | 4 | `code_scanning_alerts`, `dependabot_alerts`, `secret_scanning_alerts`, `security_advisories`. |
| **Declared streams** |  | **37** | **29 + 4 + 4** |

## Early escalation and resolved foundation gap

The safely targetable set is 127 direct commands and 29 streams, substantially
larger than the two former candidates but well short of the complete surface.
The early review found that `stageDirectReadSweep` proved only result kind and
the absence of leaked secrets. Firstmate then authorized one direct PR to add
the shared declaration-owned output assertion foundation under #4191 together
with this GitHub work. It is evaluated by the generic stage, not hard-coded for
GitHub, and is covered by a post-schema red/green test.

The first full live attempt proved 23 of 27 proposed commands and left four
red. `repo read-dir` remained red for both available root spellings and needs
a known nonempty fixture directory. The check-run/check-suite endpoints need a
typed compatible fixture and permission contract, and `subscription view`
needs a user-specific watch relationship; they remain non-pass. A focused final
live run passed all 23 resulting candidates against the final declaration. It
is not a whole-surface claim.

## Execution contract after the foundation lands

- Run one candidate at a time with the disposable certification identity loaded
  by command substitution into an exported environment variable.  Never use
  the ambient `gh` login and never emit the token, its length, or derived text.
- Persist only candidate names, pass status, and a declaration/configuration
  fingerprint in `.polymetrics/certifications/progress/<connector>-direct-read.json`.
  It contains no credential, provider response, URL, header, or rate-limit
  scope. On `Retry-After` or exhausted rate-limit headers, rerun with
  `--resume`; matching prior live candidates are marked `resumed` instead of
  implying a new provider call, and the next candidate runs serially.
- Each candidate asserts its direct-read result kind, one declaration-owned
  produced value, and secret-free output.  An expected provider non-pass stays
  out of the pass count with its concrete reason.
- A scratch post-schema assertion mismatch must turn certification red before
  live proof is accepted; restore the valid declaration immediately after the
  demonstration.
