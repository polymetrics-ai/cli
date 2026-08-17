# GitHub not-applicable audit — 2026-08-17

## Method

This audit enumerated every GitHub `cli_surface.json` command declared
`unsupported_api` or `unsupported_local` (50 total) and compared it with the
pinned provider source lock at
`internal/connectors/defs/github/sources/github-operation-source-lock.json`.
That lock is a GitHub REST OpenAPI capture plus GitHub GraphQL root fields; it
is the right evidence for provider availability, not a claim that the current
runtime already has a safe executable command contract.

**Finding:** 26 of the 27 `unsupported_api` declarations contradict the pinned
provider source: their provider operation exists. They remain non-executable
in this branch because no safe operation binding exists, but they must not be
counted as provider-unsupported or excluded from the certification denominator
on that basis. `skill list` is the single `unsupported_api` declaration with no
matching REST operation or GraphQL root field in the pinned source lock.

All 23 `unsupported_local` declarations have a genuine local-workflow,
credential-vault, browser, filesystem, or deliberately forbidden generic API
boundary. Four have a supporting provider endpoint, noted below; that does not
make their complete local `gh` behavior executable through the current typed
connector runtime.

## `unsupported_api` declarations (27)

| Command | Source-lock evidence | Verdict |
| --- | --- | --- |
| `issue pin` | GraphQL mutation `pinIssue` | **Provider-supported; report as wrong unsupported declaration.** |
| `issue unpin` | GraphQL mutation `unpinIssue` | **Provider-supported; report as wrong unsupported declaration.** |
| `pr diff` | REST `pulls/get` `GET /repos/{owner}/{repo}/pulls/{pull_number}`; diff needs a non-JSON representation policy | **Provider-supported; report as wrong unsupported declaration.** |
| `pr ready` | GraphQL mutation `markPullRequestReadyForReview` | **Provider-supported; report as wrong unsupported declaration.** |
| `repo list` | REST `repos/list-for-authenticated-user` `GET /user/repos` | **Provider-supported; report as wrong unsupported declaration.** |
| `repo autolink create` | REST `repos/create-autolink` `POST /repos/{owner}/{repo}/autolinks` | **Provider-supported; report as wrong unsupported declaration.** |
| `repo autolink delete` | REST `repos/delete-autolink` `DELETE /repos/{owner}/{repo}/autolinks/{autolink_id}` | **Provider-supported; report as wrong unsupported declaration.** |
| `repo license list` | REST `licenses/get-all-commonly-used` `GET /licenses` | **Provider-supported; report as wrong unsupported declaration.** |
| `repo gitignore list` | REST `gitignore/get-all-templates` `GET /gitignore/templates` | **Provider-supported; report as wrong unsupported declaration.** |
| `workflow enable` | REST `actions/enable-workflow` `PUT /repos/{owner}/{repo}/actions/workflows/{workflow_id}/enable` | **Provider-supported; report as wrong unsupported declaration.** |
| `workflow disable` | REST `actions/disable-workflow` `PUT /repos/{owner}/{repo}/actions/workflows/{workflow_id}/disable` | **Provider-supported; report as wrong unsupported declaration.** |
| `cache list` | REST `actions/get-actions-cache-list` `GET /repos/{owner}/{repo}/actions/caches` | **Provider-supported; report as wrong unsupported declaration.** |
| `label clone` | REST pair `issues/list-labels-for-repo` + `issues/create-label` | **Provider-supported; report as wrong unsupported declaration.** |
| `secret list` | REST `actions/list-repo-secrets` `GET /repos/{owner}/{repo}/actions/secrets` | **Provider-supported; report as wrong unsupported declaration.** |
| `variable list` | REST `actions/list-repo-variables` `GET /repos/{owner}/{repo}/actions/variables` | **Provider-supported; report as wrong unsupported declaration.** |
| `variable get` | REST `actions/get-repo-variable` `GET /repos/{owner}/{repo}/actions/variables/{name}` | **Provider-supported; report as wrong unsupported declaration.** |
| `variable set` | REST `actions/create-repo-variable` / `actions/update-repo-variable` | **Provider-supported; report as wrong unsupported declaration.** |
| `variable delete` | REST `actions/delete-repo-variable` `DELETE /repos/{owner}/{repo}/actions/variables/{name}` | **Provider-supported; report as wrong unsupported declaration.** |
| `org list` | REST `orgs/list-for-authenticated-user` `GET /user/orgs` | **Provider-supported; report as wrong unsupported declaration.** |
| `gist list` | REST `gists/list` `GET /gists` | **Provider-supported; report as wrong unsupported declaration.** |
| `gist create` | REST `gists/create` `POST /gists` | **Provider-supported; report as wrong unsupported declaration.** |
| `codespace list` | REST `codespaces/list-for-authenticated-user` `GET /user/codespaces` | **Provider-supported; report as wrong unsupported declaration.** |
| `codespace create` | REST `codespaces/create-for-authenticated-user` `POST /user/codespaces` | **Provider-supported; report as wrong unsupported declaration.** |
| `gpg-key list` | REST `users/list-gpg-keys-for-authenticated-user` `GET /user/gpg_keys` | **Provider-supported; report as wrong unsupported declaration.** |
| `ssh-key list` | REST `users/list-public-ssh-keys-for-authenticated-user` `GET /user/keys` | **Provider-supported; report as wrong unsupported declaration.** |
| `skill list` | No matching REST operation or GraphQL root field in the pinned source lock | Supported provider operation is **not evidenced**; declaration is retained. |
| `agent-task list` | REST `agent-tasks/list-tasks` `GET /agents/tasks` | **Provider-supported; report as wrong unsupported declaration.** |

## `unsupported_local` declarations (23)

| Command | Verification | Verdict |
| --- | --- | --- |
| `issue develop` | GraphQL `createLinkedBranch` exists, but the command's declared behavior also creates/checks out a local branch. | Correctly non-executable as a local workflow. |
| `pr checkout` | Requires local git checkout state. | Correctly non-executable as a local workflow. |
| `repo clone` | Requires local git and filesystem mutation. | Correctly non-executable as a local workflow. |
| `repo sync` | Requires local git state. | Correctly non-executable as a local workflow. |
| `repo set-default` | Mutates `gh` local configuration rather than connector metadata. | Correctly non-executable as a local workflow. |
| `release upload` | REST `repos/upload-release-asset` exists, but the declared behavior needs a local file and a typed binary-upload contract. | Correctly non-executable locally; provider endpoint is not an implementation claim. |
| `release download` | REST `repos/get-release-asset` exists, but the declared behavior writes an asset to local disk. | Correctly non-executable locally; provider endpoint is not an implementation claim. |
| `release verify` | Requires local asset files and signature verification. | Correctly non-executable as a local workflow. |
| `run download` | REST artifact metadata exists (`actions/list-workflow-run-artifacts`), but the command downloads to local disk. | Correctly non-executable locally. |
| `run watch` | Interactive polling/watch behavior is not a fixed connector operation. | Correctly non-executable as a local workflow. |
| `codespace ssh` | Requires local SSH and terminal behavior. | Correctly non-executable as a local workflow. |
| `auth login` | PM's credential vault intentionally does not manage `gh` sessions. | Correctly non-executable locally. |
| `auth status` | PM credential inspection is distinct from a `gh` session. | Correctly non-executable locally. |
| `auth token` | Exporting a token conflicts with PM's no-secret-disclosure boundary. | Correctly non-executable locally. |
| `config get` | Reads local `gh` configuration. | Correctly non-executable locally. |
| `config set` | Writes local `gh` configuration. | Correctly non-executable locally. |
| `browse` | Opens a local browser. | Correctly non-executable as a local workflow. |
| `alias list` | Reads `gh` alias configuration. | Correctly non-executable locally. |
| `extension list` | Depends on local extension installation/execution. | Correctly non-executable locally. |
| `completion` | Belongs to PM's shell-integration surface, not a connector operation. | Correctly non-executable locally. |
| `api` | Would introduce forbidden generic authenticated API dispatch. | Correctly non-executable locally. |
| `attestation verify` | REST attestation records exist, but verification requires a local artifact and verifier. | Correctly non-executable locally. |
| `copilot` | Is Copilot CLI behavior rather than a GitHub connector operation. | Correctly non-executable locally. |

## Follow-up boundary

This P1 lane does not change any availability declaration. The 26
provider-supported `unsupported_api` classifications need a separate owner to
choose a truthful implemented/partial/planned contract with fixtures,
entitlements, write plan-preview-approval-execute behavior, and certification
evidence. The task brief explicitly requires reporting, not silent
reclassification.
