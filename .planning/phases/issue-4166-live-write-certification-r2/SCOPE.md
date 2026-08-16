# Live GitHub Write Certification Scope

## Inventory and classification

The GitHub bundle currently declares **607** write actions. This inventory is
complete: every `writes.json.actions[]` entry is classified by the ordered
rules below. A planned invariant test must fail if the action count changes or
an action matches no rule.

| Classification | Actions | Exact selector | Why this is the boundary |
| --- | ---: | --- | --- |
| `safe_repository` | 28 | The named action list below | Each action can change only run-owned state in one private retained repository, has a bounded provider read-back, and can be restored or removed without touching shared account, organisation, or enterprise state. |
| `dedicated_repository_infrastructure` | 234 | Every other `/repos/{owner}/{repo}/...` action | A private repository alone is not sufficient: these endpoints need a deliberately controlled workflow, runner, webhook receiver, key, branch/PR graph, deployment/security fixture, or settings sandbox. Their risk field in `writes.json` is retained as the per-action reason. |
| `dedicated_user_infrastructure` | 9 | `/gists/...` | Requires run-owned private Gists and account-scoped cleanup. The retained repository does not isolate Gist state. |
| `dedicated_organization_infrastructure` | 217 | `/orgs/...`, `/organizations/...`, and `/teams/...` | Requires an independently owned disposable organisation and, depending on the action, members, teams, projects, runners, billing/security features, or a supported plan. Never mutate a shared organisation. |
| `dedicated_application_infrastructure` | 14 | `/app/...`, `/applications/...`, `/app-manifests/...`, `/installation/...`, and `/agents/...` | Requires a separately registered disposable GitHub App/OAuth application, isolated installation, or coding-agent environment. No shared application or installation is eligible. |
| `genuinely_untestable` | 105 | `/enterprises/...`, `/user/...`, `/users/...`, `/notifications/...`, and `/credentials/...` | These mutate enterprise-wide policy or the authenticated/named user’s identity, credentials, notification inbox, memberships, keys, packages, Codespaces, social graph, or personal settings. There is no disposable-resource boundary that makes the existing credential safe to mutate. |

The counts are mutually exclusive and sum to 607. The selector table is an
action-by-action classification: resolving an action’s declared path selects
one row, while the 28 named safe actions override the generic repository row.
The source bundle's own `risk` text remains the specific reason for every
non-safe action; the eventual report must carry both the action name/path and
that reason, rather than a generic aggregate label.

### Safely executable now: 28 actions

`create_issue`, `update_issue`, `comment_issue`, `close_issue`,
`reopen_issue`, `create_label`, `update_label`, `delete_label`,
`create_milestone`, `update_milestone`, `delete_milestone`, `create_release`,
`update_release`, `delete_release`, `create_commit_comment`,
`update_commit_comment`, `delete_commit_comment`, `update_issue_comment`,
`delete_issue_comment`, `lock_issue`, `unlock_issue`, `set_issue_labels`,
`add_issue_labels`, `remove_issue_label`, `create_ref`, `update_ref`,
`delete_ref`, and `replace_repo_topics`.

Each needs deterministic fixture state in one uniquely named private
repository, a production-path mutation, a separate REST read-back, and
verified restoration or removal. `create_*` actions may use their paired
cleanup action only as cleanup; cleanup does not count as evidence for the
primary action without its own mutation/read-back scenario.

### Infrastructure not currently available: 474 actions

- **234 repository-scoped** actions need fixtures that cannot safely be
  inferred (no-op Actions workflow and disposable runner; a verified webhook
  sink; non-secret deploy key; controlled branch/PR/review graph; release
  asset; deployment/environment; security alert; or reversible repository
  policy). Some are destructive even inside a test repository and must begin
  from an explicitly created resource plus a restoration proof.
- **9 Gist** actions need private Gist lifecycle support and a rule forbidding
  use of pre-existing Gists.
- **217 organisation/team** actions need a disposable organisation and feature
  support. Organisation billing, policies, security, runners, and membership
  operations require separate sub-boundaries rather than one broad admin token.
- **14 application/install/agent** actions need isolated application
  registration, keys, installation, and agent environment. The current
  user-token credential is not evidence that these routes are safe.

### Genuine non-live residue: 105 actions

Enterprise and user/control-plane endpoints are not candidates for this
certification credential. They must appear as a distinct `not_live` outcome
with their path and exact bundle risk, never as `pass`, `live`, or a full-parity
success condition.

## Rate and runtime budget

The safe 28-action set needs roughly 3–5 provider calls per action (fixture
setup when needed, mutation, independent read-back, cleanup/read-back), or
about **84–140 requests** and **at least 56 content-generating requests**.
Run them serially with a one-second inter-mutation floor, budget **5 minutes**
wall clock for normal API latency and cleanup retries, and no concurrent
mutations. GitHub’s current REST guidance says to serialize requests and wait
at least one second between `POST`, `PATCH`, `PUT`, and `DELETE`; its published
secondary-limit guidance is generally 80 content-generating requests/minute
and 500/hour, with endpoint-specific limits possible. The runner must honor
`Retry-After`, `X-RateLimit-Remaining`, and `X-RateLimit-Reset`, otherwise back
off for at least one minute with bounded exponential retries.

If all 474 infrastructure-gated actions were eventually enabled, even the
optimistic 3–5 calls/action model is 1,422–2,370 requests and likely exceeds
the 500 content-generating-requests/hour guidance once setup and cleanup are
included. That work needs a staged, resumable multi-hour schedule and isolated
credentials/infrastructure, not a single `--full-parity` invocation.

## Resumability contract

Before any provider mutation, a durable per-action ledger record must contain
the connector, action, stable scenario id, ownership tag, resource identity
when known, and state (`planned`, `mutated`, `read_back`, `cleaned`,
`not_live`). On restart, the runner first read-backs and cleans incomplete
run-owned resources; it never repeats a create merely because a process
stopped. A rate-limit wait preserves the checkpoint and resumes the next
unfinished action. A resource without a verified ownership tag is a terminal
safety failure, not a cleanup candidate.

## Honest end state without additional infrastructure

The defensible live claim is **28/607 safe repository actions executed** with
read-back, alongside **474 infrastructure-gated** and **105 genuinely
non-live** outcomes. `prepared_only` is not a passing or live status. A full
parity proof must either require the provisioned actions for its declared
surface or refuse to claim full parity while either applicable stage was
skipped or the report contains a non-live action.

## Decision required before implementation

The captain’s direction is to make the 607-action claim true through execution,
but the existing private repository can safely cover only 28 actions. Choose
one of the following before the harness is built:

1. Provision and authorize separately owned disposable organisation, app,
   installation, Gist, Actions, webhook, deployment/security, and account
   fixtures, then schedule the 474 infrastructure-gated actions in bounded
   waves. Enterprise/user/control-plane endpoints remain explicitly non-live.
2. Authorize the 28-action repository-safe boundary only. Certification then
   reports every remaining action as a non-pass, non-live classification and
   `--full-parity` must refuse whole-surface/full-write language.

No option permits writing to a third-party repository or organisation, or
mutating user/enterprise control-plane state with the existing credential.
