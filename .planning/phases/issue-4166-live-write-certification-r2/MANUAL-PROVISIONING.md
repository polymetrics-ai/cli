# Captain one-sitting provisioning checklist — Issue 4166

This is the complete browser/manual prerequisite list for the **474 reachable
actions that are not in the 28-action repository-safe wave**. It deliberately
does not authorize or require any mutation of a shared organisation,
third-party repository, existing application, personal account control plane,
or enterprise control plane. The certification runner will create and retain
only its tagged fixtures after these prerequisites exist.

## Do once: isolated ownership and cost boundary

1. Create a brand-new organisation solely for this programme, for example
   `pm-cert-live-writes-r1` (choose an available unique name). It must have no
   shared repositories, no production data, no members except the owner and
   the dedicated fixture account below, and no pre-existing Apps or OAuth
   applications. Do **not** transfer an existing organisation.
2. Choose a **GitHub Enterprise Cloud trial** for that new organisation. This
   is the required plan, not Free or Team: the 217 organisation/team actions
   include security, hosted-compute, policy, runner, Codespaces, and
   organisation-administration surfaces which are otherwise feature-gated.
   This plan does *not* make the 105 `/enterprises`, `/user`, `/users`,
   `/notifications`, or `/credentials` routes eligible; those remain
   permanently `not_live`.
3. In the organisation settings, enable only the services required for the
   fixtures: Actions, self-hosted runners, hosted runners/hosted compute when
   offered, Codespaces, organisation Projects, Pages, Packages, Dependabot,
   code scanning, secret scanning, and GitHub Advanced Security/Secret
   Protection. Enable GitHub Copilot Business/Enterprise and Copilot Spaces
   only if the organisation's trial makes them available; record an unavailable
   product as a feature-gated wave instead of substituting shared resources.
4. Set a hard low spend limit and billing alert before enabling hosted compute
   or Codespaces. The certification schedule will serialize content-generating
   writes and stop/resume on provider limits; the cap is the extra protection
   against accidental fixture cost.
5. Create one additional **dedicated disposable GitHub user** with a verified
   email, no personal repositories or organisations, and no credentials used
   elsewhere. Invite and accept it into this test organisation in the browser.
   It is required only for organisation/team membership, invitation, blocking,
   and Codespaces-access fixtures. Never use an uninvolved human as a fixture.

## Do once: test GitHub App (installation-authenticated paths)

Create a private GitHub App owned by the new organisation, named for example
`pm-cert-live-writes-r1-app`. Register and install it **only** in this
organisation, initially on the two later agent-created, private, tagged fixture
repositories. Do not install it in any other organisation or repository.

Set these GitHub App permissions to **Read and write**, except `Metadata`,
which GitHub requires as **Read-only**. These are the current GitHub App
permission-picker names required by the declared reachable paths; do not
substitute a personal-account permission, because the 105 personal/control-
plane actions are intentionally out of scope.

| Scope | Exact permission names |
| --- | --- |
| Repository | Actions; Administration; Agent secrets; Agent variables; Artifact metadata; Attestations; Checks; Code scanning alerts; Codespaces lifecycle admin; Codespaces metadata; Codespaces secrets; Codespaces; Commit statuses; Contents; Copilot agent settings; Custom properties; Dependabot alerts; Dependabot secrets; Deployments; Environments; Issues; Metadata (**read-only**); Pages; Pull requests; Repository security advisories; Secret scanning alerts; Secrets; Variables; Webhooks; Workflows. |
| Organisation | Administration; Agent secrets; Agent variables; Blocking users; Campaigns; Copilot Spaces; Copilot agent settings; Copilot content exclusion; Custom organization roles; Custom properties; Events; GitHub Copilot Business; Hosted runner custom images; Issue Fields; Issue Types; Members; Network configurations; Organization codespaces secrets; Organization codespaces settings; Organization codespaces; Organization dependabot secrets; Organization private registries; Personal access token requests; Personal access tokens; Projects; Secrets; Self-hosted runners; Variables; Webhooks. |

If the current GitHub App registration UI uses a renamed permission, choose the
documented equivalent with the same resource noun and write access, and record
the exact displayed name in the provisioning response. Do not broaden access
to “all organisations” or “all repositories.”

Enable **Device Flow** only if the app-based wave uses it. Set a dedicated
HTTPS callback URL and webhook URL that point to the controlled receiver below;
turn on webhook delivery only after the receiver is ready. Generate one private
key and provide it to the runner only through the approved secret-injection
environment. Do not paste a key, client secret, token, webhook secret, or
callback code into chat, a command line, a repository file, an issue, or a
certification artifact.

## Do once: test OAuth application (OAuth-application paths)

Register a separate private OAuth application owned by the same disposable
organisation, for example `pm-cert-live-writes-r1-oauth`, with the same
controlled HTTPS callback endpoint. It must not reuse any production OAuth
application or client secret. Enable device flow if offered and needed by the
fixture. Keep its client secret solely in approved secret injection.

Authorize its purpose-made test token with exactly these OAuth scopes:

| Scope | Why it is required |
| --- | --- |
| `repo` | Private test-repository writes, repository hooks, deployments, releases, commit statuses, invitations, and security-event routes. |
| `admin:org` | Disposable organisation/team administration, membership, projects, runners, Actions policy, organisation hooks, and org-level security/variables/secrets. |
| `gist` | The nine private-Gist lifecycle actions. |
| `workflow` | Create/update a fixture workflow file for Actions-dependent repository scenarios. |

The current personal token already has `repo` and `gist`. Its only required
scope additions are **`admin:org` and `workflow`**. Keep its existing
`admin:public_key` unchanged; do not add it for this programme. Do **not** add
`admin:enterprise`, `user`, `notifications`, or `codespace`: each would widen
personal/enterprise control-plane authority for the 105 actions that remain
non-live. A scope does not make an action safe outside the disposable boundary.

## Do once: controlled external fixture endpoints

1. Provide one disposable, controlled HTTPS webhook receiver. It must retain
   only test-event metadata/body hashes, validate the GitHub signature with a
   secret held by approved secret injection, expose a read-back endpoint for
   the runner, and support deletion of run-tagged deliveries. Do not point a
   certification webhook at a third-party service.
2. Provide the publicly reachable HTTPS callback endpoint used by the GitHub
   App/OAuth application (it may be the same controlled service, but must keep
   callback and webhook records separate). It must log no credentials or OAuth
   codes.
3. Where the plan exposes a browser-only feature enrolment or preview, enable
   it only for the disposable organisation and record the feature’s availability
   result. A missing paid/preview feature is a documented wave prerequisite,
   never a reason to send the route to another organisation.

## What the agent will do after this checklist

The agent, not the captain, will create the retained private repositories,
private Gists, tagged teams/projects/runners/webhooks, and all per-run fixture
resources. Each mutation will carry a `pm-cert-` ownership tag, use a durable
per-action ledger, independently read back the provider state, and verify
cleanup/restoration. Repositories and the organisation will not be deleted by
the certification harness.

The schedule is 28 repository-safe actions first, then bounded resumable waves
for the 234 additional repository actions, 9 Gist actions, 217
organisation/team actions, and 14 app/OAuth/agent actions. The report will
retain every action's declared path and bundle `risk` text; the 105 genuine
control-plane actions are explicitly `not_live`, never `pass`, `live`, or full
parity coverage.
