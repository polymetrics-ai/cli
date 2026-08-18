# Research — issue #4015 GitHub declared-command parity

## Task Delivery Header

- Issue: Refs #4015 — Production MVP
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `fm/cli-parity-implement-50` → `integration/4015-mvp-flat-r1` → `main`
- Delivery: Incremental pull request open against `integration/4015-mvp-flat-r1`, with the
  repository gates green and the API-reported PR base verified.
- Working branch: `fm/cli-parity-implement-50`
- Task: Verdict all 50 empty-surface GitHub commands, implement every command expressible through
  the existing typed connector runtime, preserve commands that require a shared or prohibited
  executor with exact evidence, certify safe provider behavior, and leave no disposable fixtures.
- Verification: focused Go tests, runtime preflight sweep, connector validation, surface sync,
  certification matrix, generated docs/help checks, binary reachability, credential scans, live
  GitHub observations, cleanup reads, and the repository's non-monolithic local gates.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| All 50 commands receive a verdict | live | A table-driven test and the final verdict ledger read the supplied command names and assert exactly 50 unique entries, with no omission. |
| Every promoted command is executable | live | The real `commandrunner.Preflight` sweep accepts every `availability: implemented` command; a no-credential binary invocation reaches credential resolution instead of returning unknown/blocked. |
| Provider reads match documented endpoints | live | Safe reads against the certification identity assert non-empty identities, exact known repository data, or returned collection shape and page context. |
| Provider mutations work and clean up | live | Each safe mutation asserts the created/changed state through an independent read, reverts or deletes it, then asserts the prior state or a provider 404. |
| Local/foundation gaps stay truthful | live | Repository tests assert each retained command remains declared and its note names the missing typed capability or binding prohibition. |
| No credential or fixture remains | live | A branch-content secret scan is empty and independent provider reads find no task-created resource after cleanup. |

Every live provider proof must assert returned content or a state transition. Exit status alone is
not evidence.

## Current runtime facts

- `direct_read` executes fixed REST or GraphQL operation declarations.
- `direct_write` executes fixed GraphQL mutations through plan → preview → approval → execute.
- REST mutations execute as declared `reverse_etl` write actions through the same lifecycle.
- `binary_download` supports one declared, bounded response written to an explicit destination and
  deliberately performs no archive extraction.
- The runtime has no `local_workflow`, `config`, `auth`, `raw_api`, composite-command, raw-upload,
  subprocess, browser-launch, SSH, or cryptographic-verification executor.
- GitHub's connector hook is the bounded Tier-2 escape hatch for existing irreducible provider
  writes, not authorization to add generic shell or HTTP behavior.

## Provider confirmation and preliminary verdicts

The links below are GitHub's current official REST, GraphQL, or CLI documentation. `promote` means
the command can be represented by an existing executor. `retain` means the provider may support a
related primitive, but the declared command requires a capability the runtime intentionally does
not have. Final verdicts and observed evidence are recorded in `SUMMARY.md`.

| # | Command | Preliminary verdict | Evidence and implementation mapping |
| ---: | --- | --- | --- |
| 1 | `issue pin` | promote | Fixed `pinIssue` GraphQL mutation already exists as `github.graphql.mutation.pin-issue`; add a compatibility alias with typed `input`. |
| 2 | `issue unpin` | promote | Fixed `unpinIssue` mutation already exists; add a typed alias. |
| 3 | `pr diff` | retain | [Pull request REST media types](https://docs.github.com/en/rest/pulls/pulls#get-a-pull-request) require `Accept: application/vnd.github.diff`; operation declarations cannot set a fixed per-command response media type and JSON direct reads would misreport the body. |
| 4 | `pr ready` | promote | Fixed `markPullRequestReadyForReview` GraphQL mutation already exists; add a typed alias. |
| 5 | `repo list` | promote | [RepositoryOwner](https://docs.github.com/en/graphql/reference/interfaces#repositoryowner) exposes paged repositories for either a user or organization; add a fixed GraphQL query with a required login. |
| 6 | `repo autolink create` | promote | [Create an autolink](https://docs.github.com/en/rest/repos/autolinks#create-an-autolink-reference-for-a-repository) maps to the existing declared REST write after completing its schema. |
| 7 | `repo autolink delete` | promote | [Delete an autolink](https://docs.github.com/en/rest/repos/autolinks#delete-an-autolink-reference-from-a-repository) maps to the existing declared REST write. |
| 8 | `repo license list` | promote | [List commonly used licenses](https://docs.github.com/en/rest/licenses/licenses#get-all-commonly-used-licenses) is fixed `GET /licenses`. |
| 9 | `repo gitignore list` | promote | [List gitignore templates](https://docs.github.com/en/rest/gitignore/gitignore#get-all-gitignore-templates) is fixed `GET /gitignore/templates`. |
| 10 | `workflow enable` | promote | [Enable a workflow](https://docs.github.com/en/rest/actions/workflows#enable-a-workflow) maps to the existing REST write. |
| 11 | `workflow disable` | promote | [Disable a workflow](https://docs.github.com/en/rest/actions/workflows#disable-a-workflow) maps to the existing REST write. |
| 12 | `cache list` | promote | [List repository caches](https://docs.github.com/en/rest/actions/cache#list-github-actions-caches-for-a-repository) maps to `github.actions_caches2`. |
| 13 | `label clone` | retain | GitHub exposes label list/create primitives, but cloning is a multi-page, multi-write composite with partial-failure semantics; no typed composite command executor exists. |
| 14 | `secret list` | promote | [List repository secrets](https://docs.github.com/en/rest/actions/secrets#list-repository-secrets) returns metadata only and maps to `github.actions_secrets`. |
| 15 | `variable list` | promote | [List repository variables](https://docs.github.com/en/rest/actions/variables#list-repository-variables) maps to `github.actions_variables`. |
| 16 | `variable get` | promote | [Get a repository variable](https://docs.github.com/en/rest/actions/variables#get-a-repository-variable) maps to `github.actions_variables_name2`. |
| 17 | `variable set` | retain | GitHub has distinct create and update endpoints; upstream `set` selects between them. The runtime has both typed writes but no conditional composite command executor. Existing `variable create`/`variable update` remain executable. |
| 18 | `variable delete` | promote | [Delete a repository variable](https://docs.github.com/en/rest/actions/variables#delete-a-repository-variable) maps to the existing REST write. |
| 19 | `org list` | promote | [List organizations for the authenticated user](https://docs.github.com/en/rest/orgs/orgs#list-organizations-for-the-authenticated-user) is fixed `GET /user/orgs`. |
| 20 | `gist list` | promote | [List gists for the authenticated user](https://docs.github.com/en/rest/gists/gists#list-gists-for-the-authenticated-user) is fixed `GET /gists`. |
| 21 | `gist create` | retain | [Create a gist](https://docs.github.com/en/rest/gists/gists#create-a-gist) requires a dynamic filename-keyed `files` object. REST structured JSON flags are deliberately refused, and a user-scoped gist violates the authorized repo/org write fixture boundary. |
| 22 | `codespace list` | promote | [List codespaces for the authenticated user](https://docs.github.com/en/rest/codespaces/codespaces#list-codespaces-for-the-authenticated-user) is fixed `GET /user/codespaces`. |
| 23 | `codespace create` | promote | [Create a codespace](https://docs.github.com/en/rest/codespaces/codespaces#create-a-codespace-for-the-authenticated-user) maps to the existing repository-ID typed write. |
| 24 | `gpg-key list` | promote | [List GPG keys for the authenticated user](https://docs.github.com/en/rest/users/gpg-keys#list-gpg-keys-for-the-authenticated-user) is fixed `GET /user/gpg_keys`. |
| 25 | `ssh-key list` | promote | [List public SSH keys for the authenticated user](https://docs.github.com/en/rest/users/keys#list-public-ssh-keys-for-the-authenticated-user) is fixed `GET /user/keys`. |
| 26 | `skill list` | retain | [`gh skill list`](https://cli.github.com/manual/gh_skill_list) scans local project/user agent-host skill directories; GitHub has no provider collection behind this command and the connector runtime has no local-filesystem workflow executor. |
| 27 | `agent-task list` | promote | [List agent tasks](https://docs.github.com/en/rest/agentic-workflows/agent-tasks#list-agent-tasks-for-the-authenticated-user) is fixed `GET /agents/tasks`. |
| 28 | `issue develop` | retain | [`gh issue develop`](https://cli.github.com/manual/gh_issue_develop) creates/manages branches and may check them out; no typed git/local composite executor exists. |
| 29 | `pr checkout` | retain | [`gh pr checkout`](https://cli.github.com/manual/gh_pr_checkout) mutates a local git worktree; no typed git executor exists. |
| 30 | `repo clone` | retain | [`gh repo clone`](https://cli.github.com/manual/gh_repo_clone) invokes local git; `github.repo.clone` is metadata-only and there is no runtime `local_git` executor. |
| 31 | `repo sync` | retain | [`gh repo sync`](https://cli.github.com/manual/gh_repo_sync) performs local git fetch/reset/branch operations; no typed git executor exists. |
| 32 | `repo set-default` | retain | [`gh repo set-default`](https://cli.github.com/manual/gh_repo_set-default) writes gh-local repository config, not provider state; no connector config executor exists. |
| 33 | `release upload` | retain | [Upload a release asset](https://docs.github.com/en/rest/releases/assets#upload-a-release-asset) uses the separate upload host and raw binary request body; `rest_write` only supports typed JSON and no bounded binary-upload executor exists. |
| 34 | `release download` | promote | [Get a release asset](https://docs.github.com/en/rest/releases/assets#get-a-release-asset) maps to existing bounded `github.release.download_assets`; expose the fixed asset-ID form as `binary_download`. |
| 35 | `release verify` | retain | [`gh release verify`](https://cli.github.com/manual/gh_release_verify) performs local cryptographic verification of downloaded assets, not merely a provider read; no verification executor exists. |
| 36 | `run download` | retain | [`gh run download`](https://cli.github.com/manual/gh_run_download) lists/selects multiple run artifacts, downloads, and extracts them. The runtime only supports one fixed bounded binary response and deliberately forbids extraction; `artifact download` remains available. |
| 37 | `run watch` | retain | [`gh run watch`](https://cli.github.com/manual/gh_run_watch) is a polling/terminal workflow; no watch executor exists. |
| 38 | `codespace ssh` | retain | [`gh codespace ssh`](https://cli.github.com/manual/gh_codespace_ssh) opens an interactive SSH/process session; no SSH or subprocess executor exists. |
| 39 | `auth login` | retain | [`gh auth login`](https://cli.github.com/manual/gh_auth_login) is an interactive credential bootstrap workflow; connector credentials are managed by `pm` and no gh-auth executor exists. |
| 40 | `auth status` | retain | [`gh auth status`](https://cli.github.com/manual/gh_auth_status) inspects gh's local credential store, outside the connector runtime's credential abstraction. |
| 41 | `auth token` | retain | [`gh auth token`](https://cli.github.com/manual/gh_auth_token) prints a credential; binding rules prohibit requesting or exposing secret values. |
| 42 | `config get` | retain | [`gh config get`](https://cli.github.com/manual/gh_config_get) reads gh-local config; no connector config executor exists. |
| 43 | `config set` | retain | [`gh config set`](https://cli.github.com/manual/gh_config_set) writes gh-local config; no connector config executor exists. |
| 44 | `browse` | retain | [`gh browse`](https://cli.github.com/manual/gh_browse) launches a browser; no browser-launch executor exists. |
| 45 | `alias list` | retain | [`gh alias list`](https://cli.github.com/manual/gh_alias_list) reads gh-local aliases; no gh-config/local-filesystem executor exists. |
| 46 | `extension list` | retain | [`gh extension list`](https://cli.github.com/manual/gh_extension_list) reads installed local extensions; no local-extension executor exists. |
| 47 | `completion` | retain | [`gh completion`](https://cli.github.com/manual/gh_completion) generates shell-specific text for gh itself; connector commands do not use Cobra and expose no shell-completion contract. |
| 48 | `api` | retain | [`gh api`](https://cli.github.com/manual/gh_api) is caller-selected generic HTTP/GraphQL passthrough. Connector canon expressly prohibits generic HTTP write and generic SQL/shell escape hatches; fixed typed operations are the supported replacement. |
| 49 | `attestation verify` | retain | [Artifact attestations REST](https://docs.github.com/en/rest/repos/repos#list-attestations) only returns bundles; [`gh attestation verify`](https://cli.github.com/manual/gh_attestation_verify) verifies signatures, identity, subject digest, and policy locally. A provider read is not verification. |
| 50 | `copilot` | retain | [`gh copilot`](https://cli.github.com/manual/gh_copilot) runs an external interactive extension; no extension/subprocess executor exists. |

Preliminary sum: **23 promote + 27 retain = 50**. The implementation phase may demote a candidate
only if the real runtime preflight or safe live certification produces concrete contrary evidence.

