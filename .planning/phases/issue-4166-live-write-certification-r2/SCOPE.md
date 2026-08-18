# Live GitHub Write Certification Scope

## Revised disposable-identity boundary

The earlier classification incorrectly treated the authenticated identity and
its organisation as a captain-owned personal control plane. The captain has
now established that the certification identity is disposable, the
**Polymetrics-Cert** organisation already exists solely for this programme, and
its GitHub Enterprise Cloud trial is in progress. No category is inherently
non-live on that boundary.

The GitHub bundle declares **607** write actions. The exact, mutually exclusive
path/action selectors below classify all 607 actions. Every deferred row names
the concrete missing prerequisite; every action also retains its own
`writes.json.actions[].path` and `risk` in the eventual report. A future
classification test must fail if the action count changes or a declaration
matches no row.

| Status now | Actions | Exact selector | Concrete current blocker |
| --- | ---: | --- | --- |
| `repository_wave_ready` | 25 | The named list below except commit comments | The existing run-owned `Polymetrics-Cert/pm-cert-3993-20260810-wz0fru` fixture has disposable-token `push` and `admin`, issues enabled, and a captured label baseline. This wave must not create a repository or request repository-creation permission. |
| `repository_item_read_permission_pending` | 3 | `create_commit_comment`, `update_commit_comment`, `delete_commit_comment` | The line-277 disposable fine-grained PAT can create/update/delete a tagged commit comment but GitHub refuses the required `GET /repos/{owner}/{repo}/comments/{comment_id}` read-back with `Resource not accessible by personal access token`. GitHub documents the picker permission as **Metadata — Read-only** for that endpoint. These actions are `blocked`, never `pass`, until that item read succeeds. |
| `repository_fixture_pending` | 234 | Every other `/repos/{owner}/{repo}/...` action | The specific run-owned fixture addressed by the endpoint does not yet exist: branch/PR/review graph; no-op Actions workflow/run/runner; webhook receiver/delivery; release asset; deployment/environment; ruleset; security alert/analysis; Codespace; secret/variable; imported repository; or package/repository policy baseline. The runner must create and restore that exact fixture before its action. |
| `gist_fixture_pending` | 9 | `/gists/...` | A purpose-made private Gist and, where applicable, its comment/fork/star baseline are not yet created. |
| `org_fixture_and_permission_pending` | 217 | `/orgs/...`, `/organizations/...`, and `/teams/...` | Polymetrics-Cert exists, but the disposable certification credential lacks the required organisation authority (`admin:org` on the classic test PAT) and the named team/project/member/runner/feature fixture. No other organisation is eligible. |
| `app_or_oauth_pending` | 14 | `/app/...`, `/applications/...`, `/app-manifests/...`, `/installation/...`, and `/agents/...` | The dedicated Polymetrics-Cert GitHub App, its installation/private key, the private OAuth application/client secret, and the isolated coding-agent fixture have not yet been browser-provisioned. |
| `enterprise_trial_and_token_pending` | 25 | `/enterprises/...` | The existing Enterprise Cloud trial must finish and yield its slug; the disposable certification identity must be an enterprise owner; a classic test PAT with `admin:enterprise` is required for routes GitHub documents as incompatible with fine-grained PATs/App tokens; enterprise team/security/Copilot fixture state must be created only inside that enterprise. |
| `primary_user_fixture_and_permission_pending` | 49 | `/user/...` | The disposable primary account needs the named user-level fixtures (profile baseline, Codespace, user secret, key, package, migration, project, invite, social/follow target) and the dedicated classic-PAT scopes enumerated in `MANUAL-PROVISIONING.md`. |
| `secondary_user_fixture_and_permission_pending` | 25 | `/users/{username}/...` | A second disposable account, its verified email, purpose-made credential, and its own package/project/attestation/Copilot-Space fixtures do not yet exist. It is the only permitted `{username}` target. |
| `notification_token_and_fixture_pending` | 5 | `/notifications/...` | The disposable primary inbox needs a run-owned notification thread **and a classic PAT with `notifications` scope**. GitHub's current endpoint documentation says these writes do not accept fine-grained PATs or App tokens. |
| `sacrificial_credential_pending` | 1 | `/credentials/revoke` | A separately created, purpose-made sacrificial token has not been supplied through secret injection. It must be revoked last and can never be the runner credential. |

The counts sum to 607. `not_live` remains an honest report outcome when a
listed prerequisite is absent, but its reason must be the concrete row above
plus the action's own path/risk—never “no safe boundary.” `--full-parity` must
therefore refuse to claim success until all applicable rows have completed a
real mutation, independent read-back, and verified cleanup/restoration.

### Repository wave ready now: 28 actions

`create_issue`, `update_issue`, `comment_issue`, `close_issue`,
`reopen_issue`, `create_label`, `update_label`, `delete_label`,
`create_milestone`, `update_milestone`, `delete_milestone`, `create_release`,
`update_release`, `delete_release`, `create_commit_comment`,
`update_commit_comment`, `delete_commit_comment`, `update_issue_comment`,
`delete_issue_comment`, `lock_issue`, `unlock_issue`, `set_issue_labels`,
`add_issue_labels`, `remove_issue_label`, `create_ref`, `update_ref`,
`delete_ref`, and `replace_repo_topics`.

Each needs production-path mutation, separate REST read-back, ownership-tag
check, and verified cleanup/restoration against
`Polymetrics-Cert/pm-cert-3993-20260810-wz0fru`. A cleanup action does not
count as primary-action coverage unless it also has its own mutation/read-back
scenario. The close-only issue cleanup is the declared reversible boundary for
that API: its tagged issue is left closed, never deleted or repurposed.
The production runner rejects any other owner/repository pair before it builds
a credential or sends a provider request for this wave.

### Action-to-fixture precision for the 234 deferred repository routes

The `repository_fixture_pending` rule is deliberately a fixture lookup, not a
permission hand-wave. The scenario manifest will select one of these exact
route families before a write is enabled:

| Route family | Missing run-owned fixture |
| --- | --- |
| `/pulls`, `/branches`, `/merges`, `/stacks`, `/issues/*/dependencies` | Base commit, tagged head branch, pull request, review/comment, and the secondary account as reviewer/collaborator. |
| `/actions`, `/check-*`, `/statuses`, `/attestations` | No-op workflow file, completed run/job/artifact/check/status and, only where required, an isolated self-hosted runner. |
| `/hooks`, `/dispatches`, `/pages`, `/deployments`, `/environments` | Controlled HTTPS receiver, tagged hook delivery, dispatch event, deployment/environment and read-back endpoint. |
| `/contents`, `/git/*`, `/import`, `/forks`, `/generate`, `/transfer` | Disposable source/tree/blob/tag/import/template/fork target; transfer stays inside Polymetrics-Cert. |
| `/rulesets`, `/branches/*/protection`, `/autolinks`, `/properties`, `/interaction-limits`, `/immutable-releases` | Captured repository-settings baseline and a dedicated branch/ruleset/property fixture so cleanup restores the exact pre-state. |
| `/code-scanning`, `/dependabot`, `/secret-scanning`, `/dependency-graph`, `/vulnerability-*`, `/security-advisories`, `/code-quality` | Enterprise feature enabled and a deliberately generated test-only alert/analysis/SARIF/custom pattern/advisory fixture. |
| `/codespaces`, `/agents`, `/secrets`, `/variables`, `/packages` | Disposable Codespace/agent task or encrypted test value/package-version fixture, with values held only in secret injection. |
| remaining repository paths | The provider object directly named by the path (asset, key, invitation, reaction, subscription, cache, label, policy, or notification) is first created/tagged and its original state captured. |

## Rate and runtime budget

The 28-action wave remains roughly 84–140 provider calls, with at least 56
content-generating calls. It runs serially with a one-second mutation floor and
a five-minute normal budget. GitHub rate-limit headers and `Retry-After` must
checkpoint and resume rather than restart.

The other 579 actions are a staged, multi-hour programme. The prior
3–5-calls/action lower bound is 1,737–2,895 calls before feature setup; waves
must be serialized by fixture family and split below GitHub's published
content-generation guidance. A provider feature that is unavailable after the
named prerequisite is attempted becomes a concrete `not_live` reason naming
that feature and action path.

## Resumability contract

Before mutation, a durable per-action ledger record holds the connector,
action, stable scenario id, ownership tag, resource identity when known, and
state (`planned`, `mutated`, `read_back`, `cleaned`, `not_live`). On restart,
the runner first reads and cleans incomplete run-owned resources; named
resources are matched only by their exact ownership tag and repository-topic
restoration uses the baseline captured in the write-ahead ledger. A
recovery-only action is `recovered_unverified`, not `pass`, until a later
complete wave executes it afresh. It never repeats a create merely because a
process stopped. A rate-limit wait preserves the checkpoint. A resource
without a verified ownership tag is a terminal safety failure, not a cleanup
candidate.

## Current decision record

The captain's accepted plan is now **all 607 actions**, beginning with the
28-action repository wave. The first execution boundary is currently 25
live-testable actions plus three explicitly blocked commit-comment actions.
The remaining manual/browser prerequisites are one
consolidated list in `MANUAL-PROVISIONING.md`; they reuse the existing
disposable identity, Polymetrics-Cert organisation, and in-progress Enterprise
Cloud trial. No token value, application private key, client secret, or
sacrificial credential may be recorded in this repository, a command line,
or a certification artifact.
