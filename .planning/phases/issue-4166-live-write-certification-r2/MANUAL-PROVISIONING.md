# Captain one-sitting browser checklist — Issue 4166

This replaces the earlier parallel-organisation proposal. Reuse the existing
**disposable certification account**, the existing **Polymetrics-Cert**
organisation, and its **in-progress GitHub Enterprise Cloud trial**. Do not
create another organisation, transfer a repository, widen a personal/captain
credential, or install anything outside this boundary.

Nothing in this document asks for a token, private key, client secret, webhook
secret, or OAuth code to be pasted into chat, a command line, a repository
file, an issue, or a certification artifact. Deliver all secrets only through
the approved secret-injection channel.

## 1. Finish the existing enterprise boundary

1. Complete the Enterprise Cloud trial already in progress for
   Polymetrics-Cert and record its enterprise **slug** in the provisioning
   response (never a credential). Keep Polymetrics-Cert as the only
   organisation in that enterprise until the enterprise scenarios are green.
2. Make the disposable certification account an **enterprise owner** and an
   organisation owner. Set a low spend limit and billing alert before enabling
   hosted compute.
3. Enable, only in Polymetrics-Cert, Actions, self-hosted runners, Codespaces,
   Projects, Pages, Packages, Dependabot, code scanning, secret scanning,
   GitHub Advanced Security/Secret Protection, and the Enterprise/Copilot
   features the trial exposes. Record a feature that GitHub does not expose as
   unavailable; do not substitute a shared organisation.

This resolves the concrete enterprise prerequisite for 25 `/enterprises/...`
actions. Some enterprise endpoints require a classic PAT even when a
fine-grained/App permission appears related: GitHub documents the enterprise
team and code-security configuration routes as incompatible with fine-grained
PATs/App tokens.

## 2. Create exactly one secondary disposable user

Create a second, verified-email GitHub user from one of the spare addresses.
It must have no personal repositories, organisations, packages, Apps, or
production identity. Invite and accept it into Polymetrics-Cert and the trial
enterprise.

It is the sole target for every `/users/{username}/...` scenario and the sole
counterparty for the primary account's block/follow/reviewer/invitation tests.
The runner will create its tagged project, package, attestation, and Copilot
Space fixture state; no uninvolved user may be targeted.

Provide a purpose-made **secondary classic PAT** through secret injection for
operations that must authenticate as the target account. It needs:

`repo`, `gist`, `workflow`, `user`, `admin:public_key`, `write:gpg_key`,
`codespace`, `project`, `write:packages`, and `delete:packages`.

## 3. Add the certification credentials (three separate roles)

### Primary classic PAT — required

The existing active credential is insufficient. Create or replace it only for
the disposable certification account with a **classic PAT** carrying exactly:

`repo`, `admin:org`, `admin:enterprise`, `workflow`, `gist`, `notifications`,
`user`, `admin:public_key`, `write:gpg_key`, `codespace`, `project`,
`write:packages`, and `delete:packages`.

Why each addition is needed:

| Scope | Routes it unlocks |
| --- | --- |
| `admin:org` | Polymetrics-Cert's 217 organisation/team writes and organisation fixtures. |
| `admin:enterprise` | The 25 Enterprise-trial writes, including routes that reject fine-grained/App tokens. |
| `workflow` | Fixture workflow-file writes used by Actions scenarios. |
| `notifications` | All five `/notifications` writes. This is a classic-PAT scope, not a fine-grained permission checkbox. |
| `user` | Disposable profile, email visibility, follow/social, and other `/user` fixtures. |
| `admin:public_key` | Disposable SSH-key lifecycle. |
| `write:gpg_key` | Disposable GPG-key lifecycle. |
| `codespace` | User and repository Codespaces lifecycle/secret fixtures. |
| `project` | Disposable user/organisation Projects v2 fixture operations. |
| `write:packages`, `delete:packages` | Tagged package/package-version delete/restore scenarios. |
| `repo`, `gist` | Existing required private repository and private-Gist scopes. |

Keep the current `read:org` only if GitHub's UI requires it; `admin:org`
subsumes the programme's organisation write need. Do not use a personal/captain
token as a fallback.

**Verified correction:** the current GitHub documentation for “Mark
notifications as read,” “Mark a thread as read/done,” and subscription changes
says they do *not* accept fine-grained PATs or GitHub App tokens. Their concrete
blocker is therefore this classic `notifications` scope plus a tagged inbox
thread, not a missing fine-grained permission.

### Fine-grained/App credentials — required for the App-only surfaces

Register a private GitHub App owned by Polymetrics-Cert, install it only in
Polymetrics-Cert/the test enterprise and only on runner-created tagged private
repositories. Configure a controlled callback and webhook receiver (below),
and provide its private key only by secret injection. Give the App these
write permissions (Metadata remains read-only):

| Scope | Exact GitHub App permission names |
| --- | --- |
| Enterprise | Enterprise administration; Enterprise teams; Enterprise organization installations; Enterprise organization installation repositories. |
| Organisation | Administration; Agent secrets; Agent variables; Blocking users; Campaigns; Copilot Spaces; Copilot agent settings; Copilot content exclusion; Custom organization roles; Custom properties; Events; GitHub Copilot Business; Hosted runner custom images; Issue Fields; Issue Types; Members; Network configurations; Organization codespaces secrets; Organization codespaces settings; Organization codespaces; Organization dependabot secrets; Organization private registries; Personal access token requests; Personal access tokens; Projects; Secrets; Self-hosted runners; Variables; Webhooks. |
| Repository | Actions; Administration; Agent secrets; Agent variables; Artifact metadata; Attestations; Checks; Code quality; Code scanning alerts; Codespaces lifecycle admin; Codespaces metadata; Codespaces secrets; Codespaces; Commit statuses; Contents; Copilot agent settings; Custom properties; Dependabot alerts; Dependabot secrets; Deployments; Environments; Issues; Metadata (**read-only**); Pages; Pull requests; Repository security advisories; Secret scanning alerts; Secrets; Variables; Webhooks; Workflows. |

Create a separate private OAuth application, also owned by Polymetrics-Cert,
with a controlled callback. Its client ID/secret are needed only for the four
`/applications/{client_id}/...` token/grant routes. The GitHub App/private key
is needed for `/app/...`, `/installation/token`, and app-installation routes;
the App manifest route is exercised against this disposable App only. The
coding-agent task route additionally needs a trial-enabled, tagged agent task
fixture in Polymetrics-Cert.

### Sacrificial credential — required and intentionally revoked

Create one short-lived, purpose-made sacrificial PAT for the disposable primary
account. Supply it only through secret injection, separate from all runner
credentials. The `/credentials/revoke` scenario is last, submits only that
sacrificial secret, checks that it is unusable, and records no secret value.
GitHub documents this endpoint as accepting a revocation request without a
special token permission; it cannot be retried with the operational credential.

## 4. Controlled fixture endpoints

Provide one disposable controlled HTTPS service, preferably with distinct
callback and webhook paths. It must validate GitHub webhook signatures, retain
only test-event metadata/body hashes, offer a bounded read-back endpoint, and
delete tagged deliveries after verification. It must not log credentials,
OAuth codes, or secret values, and must not point at a third-party receiver.

## 5. What the agent creates after this checklist

The agent creates and retains only tagged private repositories, Gists, teams,
projects, runners, webhook fixtures, enterprise teams, secondary-account
fixtures, and per-action resources inside this boundary. Every action receives
a durable lifecycle ledger entry, provider read-back, and verified cleanup or
baseline restoration. The organisation, enterprise, and identities are never
deleted by the harness.

The immediate 28 repository-safe scenarios use the existing run-owned
`Polymetrics-Cert/pm-cert-3993-20260810-wz0fru` fixture. They need none of the
browser steps above and specifically do **not** need repository-creation
permission. The 579 remaining actions are scheduled in bounded resumable waves
once their named prerequisite is available. Add repository-creation permission
only if a later wave identifies the exact action family that needs a second
repository; an unavailable trial feature remains a path- and risk-specific
`not_live` result and makes `--full-parity` refuse its claim.

### Immediate repository-wave repair: commit-comment read-back

In the disposable certification account's **existing fine-grained PAT** for
`Polymetrics-Cert/pm-cert-3993-20260810-wz0fru`, set the repository permission
shown by GitHub's picker as **Metadata — Read-only**. This is the documented
fine-grained permission for `GET /repos/{owner}/{repo}/comments/{comment_id}`
(`Get a commit comment`). The current line-277 token returns **`Resource not
accessible by personal access token`** for that exact endpoint; the line-283
token returns **`Not Found`** because it is not scoped to this repository.

This is a concrete three-action gap only: `create_commit_comment`,
`update_commit_comment`, and `delete_commit_comment`. Until an item read
succeeds, the harness records those actions as `blocked`, not `pass`, even if
GitHub accepted their write. It uses the permitted commit-comment collection
read solely to verify cleanup of a run-owned tagged comment; cleanup does not
upgrade the blocked mutation to coverage. No repository-creation permission is
needed for this repair. Once the permission is present, the bounded re-run is
explicitly enabled with `--config certification_commit_comment_item_read=enabled`;
the default deliberately avoids creating another comment that cannot yet be
certified.
