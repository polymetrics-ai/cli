---
name: pm-github
description: GitHub connector knowledge and safe action guide.
---

# pm-github

## Purpose

Reads GitHub repository, issue, pull request, code, release, collaboration, and Actions data, and writes approved reverse ETL actions through the GitHub REST API.

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Certification

- Full certification: passed for the current GitHub connector surface.
- Accounted API endpoints: 509 total, 440 covered, 69 explicitly blocked.
- Catalog streams: 37.
- Direct-read command families checked: 2.
- Write actions accounted: 231.
- Live write lifecycle: `create_label` passed with read-back verification and cleanup.
- Remaining write actions are safe untested pairings or blocked by policy; destructive/admin/binary surfaces are not executed blindly.
- Binary download surfaces remain safely blocked until a bounded binary executor and destination policy exist.

## Authentication

- public: Unauthenticated public repository reads. Writes are not allowed.
  - config: repository, base_url
  - supports: read=true write=false
- token: Bearer-token auth for classic PATs, fine-grained PATs, OAuth tokens, GitHub Actions GITHUB_TOKEN, or pre-generated installation tokens.
  - config: repository, base_url, auth_type=token
  - secrets: token, personalAccessToken, oauthToken, accessToken, installationToken, githubToken
  - supports: read=true write=true
- github_app: Server-to-server auth. pm signs a GitHub App JWT and exchanges it for a one-hour installation access token.
  - config: repository, base_url, auth_type=github_app, app_id, installation_id, installation_repositories, installation_repository_ids, installation_permissions
  - secrets: private_key, private_key_base64
  - supports: read=true write=true

## Configuration

- repository (required): Repository in owner/repo format.
- base_url: GitHub API base URL override for tests or GitHub Enterprise.
- auth_type default=auto: Authentication mode: auto, public, token, or github_app.
- app_id: GitHub App ID or client ID for auth_type=github_app.
- installation_id: GitHub App installation ID for auth_type=github_app.
- installation_repositories: Optional comma-separated repository names for a restricted installation token.
- installation_repository_ids: Optional comma-separated repository IDs for a restricted installation token.
- installation_permissions: Optional JSON object of requested GitHub App installation-token permissions.
- per_page default=100: Records per page.
- max_pages default=1: Maximum pages; use 0, all, or unlimited to exhaust the stream.
- state default=all: Issue, pull request, or milestone state filter where the GitHub API supports it.
- sort: Issue, pull request, or milestone sort field where supported.
- direction: Sort direction where supported.
- since: Lower-bound timestamp for issues, issue_comments, pull_request_review_comments, and commits.
- until: Upper-bound timestamp for commits.
- sha: Branch or commit SHA filter for commits.
- path: Repository path filter for commits.
- author: Author filter for commits.
- committer: Committer filter for commits.
- actor: Actor filter for workflow_runs.
- branch: Branch filter for workflow_runs.
- event: Event filter for workflow_runs.
- status: Status filter for workflow_runs.
- created: Created-at filter for workflow_runs.
- head_sha: Head SHA filter for workflow_runs.
- check_suite_id: Check suite ID filter for workflow_runs.
- token (secret): GitHub token. Can be classic PAT, fine-grained PAT, OAuth token, GitHub Actions GITHUB_TOKEN, or installation access token.
- personalAccessToken (secret): Alias for token.
- oauthToken (secret): Alias for token when supplied by an OAuth app.
- accessToken (secret): Alias for token.
- installationToken (secret): Alias for token when an installation token is generated outside pm.
- githubToken (secret): Alias for token when passing GitHub Actions GITHUB_TOKEN.
- private_key (secret): GitHub App PEM private key for auth_type=github_app.
- private_key_base64 (secret): Base64-encoded GitHub App PEM private key for auth_type=github_app.

## ETL Streams

- repository: Repository metadata.
  - primary key: node_id
  - cursor: updated_at
  - fields: repository(string), id(integer), node_id(string), name(string), full_name(string), private(boolean), description(string), html_url(string), default_branch(string), language(string), stargazers_count(integer), watchers_count(integer), forks_count(integer), open_issues_count(integer), created_at(timestamp), updated_at(timestamp), pushed_at(timestamp)
- issues: Repository issues. Pull requests returned by the GitHub issues endpoint are filtered out.
  - primary key: node_id
  - cursor: updated_at
  - fields: repository(string), id(integer), node_id(string), number(integer), state(string), state_reason(string), title(string), body(string), html_url(string), url(string), user_login(string), user_id(integer), author_association(string), comments(integer), locked(boolean), labels_count(integer), assignees_count(integer), is_pull_request(boolean), created_at(timestamp), updated_at(timestamp), closed_at(timestamp)
- pull_requests: Repository pull requests.
  - primary key: node_id
  - cursor: updated_at
  - fields: repository(string), id(integer), node_id(string), number(integer), state(string), title(string), body(string), html_url(string), url(string), user_login(string), user_id(integer), author_association(string), comments(integer), locked(boolean), created_at(timestamp), updated_at(timestamp), closed_at(timestamp), merged_at(timestamp), draft(boolean), merge_commit_sha(string), base_ref(string), base_sha(string), head_ref(string), head_sha(string)
- branches: Repository branches.
  - primary key: name
  - fields: repository(string), name(string), protected(boolean), commit_sha(string), commit_url(string)
- commits: Repository commits.
  - primary key: sha
  - cursor: commit_committer_date
  - fields: repository(string), sha(string), node_id(string), html_url(string), url(string), author_login(string), author_id(integer), committer_login(string), committer_id(integer), commit_message(string), commit_author_name(string), commit_author_email(string), commit_author_date(timestamp), commit_committer_name(string), commit_committer_email(string), commit_committer_date(timestamp)
- tags: Repository tags.
  - primary key: name
  - fields: repository(string), name(string), zipball_url(string), tarball_url(string), commit_sha(string), commit_url(string), node_id(string)
- releases: Repository releases.
  - primary key: id
  - cursor: published_at
  - fields: repository(string), id(integer), node_id(string), tag_name(string), target_commitish(string), name(string), body(string), draft(boolean), prerelease(boolean), html_url(string), author_login(string), assets_count(integer), created_at(timestamp), published_at(timestamp)
- labels: Repository labels.
  - primary key: name
  - fields: repository(string), id(integer), node_id(string), name(string), color(string), description(string), default(boolean), url(string)
- milestones: Repository milestones.
  - primary key: number
  - cursor: updated_at
  - fields: repository(string), id(integer), node_id(string), number(integer), state(string), title(string), description(string), open_issues(integer), closed_issues(integer), creator_login(string), created_at(timestamp), updated_at(timestamp), due_on(timestamp), closed_at(timestamp)
- issue_comments: Issue and pull request timeline comments at repository scope.
  - primary key: id
  - cursor: updated_at
  - fields: repository(string), id(integer), node_id(string), issue_url(string), html_url(string), body(string), user_login(string), user_id(integer), author_association(string), created_at(timestamp), updated_at(timestamp)
- pull_request_review_comments: Pull request diff review comments.
  - primary key: id
  - cursor: updated_at
  - fields: repository(string), id(integer), node_id(string), pull_request_review_id(integer), diff_hunk(string), path(string), position(integer), original_position(integer), commit_id(string), original_commit_id(string), pull_request_url(string), html_url(string), body(string), user_login(string), created_at(timestamp), updated_at(timestamp)
- collaborators: Repository collaborators visible to the token.
  - primary key: id
  - fields: repository(string), relation(string), login(string), id(integer), node_id(string), type(string), site_admin(boolean), html_url(string), contributions(integer), role_name(string)
- contributors: Repository contributors.
  - primary key: id
  - fields: repository(string), relation(string), login(string), id(integer), node_id(string), type(string), site_admin(boolean), html_url(string), contributions(integer), role_name(string)
- stargazers: Users who starred the repository.
  - primary key: id
  - fields: repository(string), relation(string), login(string), id(integer), node_id(string), type(string), site_admin(boolean), html_url(string), contributions(integer), role_name(string)
- subscribers: Users watching repository notifications.
  - primary key: id
  - fields: repository(string), relation(string), login(string), id(integer), node_id(string), type(string), site_admin(boolean), html_url(string), contributions(integer), role_name(string)
- workflows: GitHub Actions workflows.
  - primary key: id
  - cursor: updated_at
  - fields: repository(string), id(integer), node_id(string), name(string), path(string), state(string), badge_url(string), html_url(string), created_at(timestamp), updated_at(timestamp)
- workflow_runs: GitHub Actions workflow runs.
  - primary key: id
  - cursor: updated_at
  - fields: repository(string), id(integer), node_id(string), name(string), head_branch(string), head_sha(string), status(string), conclusion(string), event(string), workflow_id(integer), run_number(integer), run_attempt(integer), html_url(string), created_at(timestamp), updated_at(timestamp)
- workflow_artifacts: GitHub Actions workflow artifacts.
  - primary key: id
  - cursor: updated_at
  - fields: repository(string), id(integer), node_id(string), name(string), size_in_bytes(integer), url(string), archive_download_url(string), expired(boolean), created_at(timestamp), updated_at(timestamp), expires_at(timestamp), workflow_run_id(integer)
- deployments: Repository deployments.
  - primary key: id
  - cursor: updated_at
  - fields: repository(string), id(integer), node_id(string), sha(string), ref(string), task(string), environment(string), description(string), creator_login(string), created_at(timestamp), updated_at(timestamp)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped
- Source modes: full_refresh, incremental

## Reverse ETL Actions

- create_issue: Create a repository issue.
  - endpoint: POST /repos/{owner}/{repo}/issues
  - required fields: title
  - optional fields: body, labels, assignees, milestone, type
  - risk: creates user-visible GitHub issue and may notify watchers
- update_issue: Edit issue title, body, state, labels, assignees, milestone, or type.
  - endpoint: PATCH /repos/{owner}/{repo}/issues/{issue_number}
  - required fields: issue_number or number
  - optional fields: title, body, state, state_reason, labels, assignees, milestone, type
  - risk: mutates existing GitHub issue or pull request issue metadata
- comment_issue: Create a comment on an issue or pull request.
  - endpoint: POST /repos/{owner}/{repo}/issues/{issue_number}/comments
  - required fields: issue_number, pull_number, or number, body
  - risk: creates user-visible comment and may notify participants
- close_issue: Close an issue, optionally with a comment and state reason.
  - endpoint: PATCH /repos/{owner}/{repo}/issues/{issue_number}
  - required fields: issue_number or number
  - optional fields: comment, state_reason
  - risk: closes existing GitHub issue
- create_pull_request: Create a pull request and optionally add labels, assignees, milestone, or reviewers.
  - endpoint: POST /repos/{owner}/{repo}/pulls
  - required fields: title, head, base
  - optional fields: body, draft, maintainer_can_modify, issue, labels, assignees, milestone, reviewers, team_reviewers
  - risk: creates user-visible pull request and may notify watchers/reviewers
- update_pull_request: Edit pull request fields and optionally add issue metadata or reviewers.
  - endpoint: PATCH /repos/{owner}/{repo}/pulls/{pull_number}
  - required fields: pull_number or number
  - optional fields: title, body, state, base, maintainer_can_modify, labels, assignees, milestone, reviewers, team_reviewers
  - risk: mutates existing GitHub pull request
- close_pull_request: Close a pull request, optionally with a comment.
  - endpoint: PATCH /repos/{owner}/{repo}/pulls/{pull_number}
  - required fields: pull_number or number
  - optional fields: comment
  - risk: closes existing GitHub pull request
- request_reviewers: Request user or team reviewers for a pull request.
  - endpoint: POST /repos/{owner}/{repo}/pulls/{pull_number}/requested_reviewers
  - required fields: pull_number or number, reviewers or team_reviewers
  - optional fields: reviewers, team_reviewers
  - risk: notifies requested GitHub reviewers
- merge_pull_request: Merge a pull request with optional commit title, message, SHA guard, and method.
  - endpoint: PUT /repos/{owner}/{repo}/pulls/{pull_number}/merge
  - required fields: pull_number or number
  - optional fields: commit_title, commit_message, sha, merge_method
  - risk: irreversibly changes repository history unless branch protection blocks merge
- create_label: Create a repository label.
  - endpoint: POST /repos/{owner}/{repo}/labels
  - required fields: name, color
  - optional fields: description
  - risk: changes repository taxonomy used by issues and pull requests
- update_label: Update a repository label name, color, or description.
  - endpoint: PATCH /repos/{owner}/{repo}/labels/{name}
  - required fields: name
  - optional fields: new_name, color, description
  - risk: renames or changes labels already used by issues and pull requests
- delete_label: Delete a repository label.
  - endpoint: DELETE /repos/{owner}/{repo}/labels/{name}
  - required fields: name
  - risk: removes a label from the repository and existing issue metadata
- create_milestone: Create a repository milestone.
  - endpoint: POST /repos/{owner}/{repo}/milestones
  - required fields: title
  - optional fields: state, description, due_on
  - risk: creates planning metadata visible to repository collaborators
- update_milestone: Update milestone title, state, description, or due date.
  - endpoint: PATCH /repos/{owner}/{repo}/milestones/{milestone_number}
  - required fields: milestone_number or number
  - optional fields: title, state, description, due_on
  - risk: changes planning metadata used by issues and pull requests
- delete_milestone: Delete a repository milestone.
  - endpoint: DELETE /repos/{owner}/{repo}/milestones/{milestone_number}
  - required fields: milestone_number or number
  - risk: removes repository planning metadata from GitHub
- create_release: Create a repository release for a tag.
  - endpoint: POST /repos/{owner}/{repo}/releases
  - required fields: tag_name
  - optional fields: target_commitish, name, body, draft, prerelease, generate_release_notes, make_latest
  - risk: publishes release metadata and may notify repository watchers
- update_release: Update release metadata.
  - endpoint: PATCH /repos/{owner}/{repo}/releases/{release_id}
  - required fields: release_id or id
  - optional fields: tag_name, target_commitish, name, body, draft, prerelease, generate_release_notes, make_latest
  - risk: changes published release metadata
- delete_release: Delete a repository release.
  - endpoint: DELETE /repos/{owner}/{repo}/releases/{release_id}
  - required fields: release_id or id
  - risk: removes release metadata from GitHub; tags are not deleted by this action
- dispatch_workflow: Trigger a GitHub Actions workflow dispatch event.
  - endpoint: POST /repos/{owner}/{repo}/actions/workflows/{workflow_id}/dispatches
  - required fields: workflow_id, ref
  - optional fields: inputs
  - risk: starts CI/CD automation that may deploy, publish, or mutate external systems
- rerun_workflow_run: Rerun a GitHub Actions workflow run.
  - endpoint: POST /repos/{owner}/{repo}/actions/runs/{run_id}/rerun
  - required fields: run_id, workflow_run_id, or id
  - risk: reruns CI/CD automation and consumes workflow minutes
- cancel_workflow_run: Cancel a GitHub Actions workflow run.
  - endpoint: POST /repos/{owner}/{repo}/actions/runs/{run_id}/cancel
  - required fields: run_id, workflow_run_id, or id
  - risk: interrupts in-flight CI/CD automation
- delete_workflow_run: Delete a GitHub Actions workflow run record.
  - endpoint: DELETE /repos/{owner}/{repo}/actions/runs/{run_id}
  - required fields: run_id, workflow_run_id, or id
  - risk: removes workflow run history from GitHub
- create_pull_request_review: Create a pull request review with optional review comments.
  - endpoint: POST /repos/{owner}/{repo}/pulls/{pull_number}/reviews
  - required fields: pull_number or number
  - optional fields: event, body, commit_id, comments
  - risk: submits reviewer feedback and may approve or request changes on a pull request
- create_or_update_file: Create or update repository file contents.
  - endpoint: PUT /repos/{owner}/{repo}/contents/{path}
  - required fields: path, message, content or content_base64
  - optional fields: sha, branch, committer, author
  - risk: writes a commit to the repository and may trigger CI/CD
- delete_file: Delete a repository file through the contents API.
  - endpoint: DELETE /repos/{owner}/{repo}/contents/{path}
  - required fields: path, message, sha
  - optional fields: branch, committer, author
  - risk: writes a commit that removes a file from the repository

## Pagination

- type: page
- page size field: per_page
- page limit field: max_pages
- default limit: 1

## Security

- read risk: external API read
- write risk: external GitHub API mutation
- mutation risk: creates or changes issues, pull requests, comments, reviewers, labels, milestones, releases, workflow runs, or repository contents
- approval: reverse ETL plan approval required before writes
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect github
```

### Inspect as structured JSON

```bash
pm connectors inspect github --json
```

### Public repository credential

```bash
pm credentials add github-public --connector github --config repository=octocat/Hello-World
```

### Token credential

```bash
export GITHUB_TOKEN=...
pm credentials add github-token --connector github --config repository=OWNER/REPO --from-env token=GITHUB_TOKEN
```

### GitHub App credential

```bash
pm credentials add github-app --connector github --config repository=OWNER/REPO --config auth_type=github_app --config app_id=12345 --config installation_id=67890 --value-stdin private_key < app-private-key.pem
```

### Pull request ETL

```bash
pm connections create github_prs_to_warehouse --source github:github-token --destination warehouse:warehouse-local --stream pull_requests --primary-key node_id --cursor updated_at --table github_pull_requests
pm etl run --connection github_prs_to_warehouse --stream pull_requests --batch-size 100 --json
```

### Approved pull request creation

```bash
pm reverse plan prs_to_github --source-table github_pr_candidates --destination github:github-token --action create_pull_request --map title:title --map body:body --map head:head --map base:base --map reviewers:reviewers
pm reverse preview <plan-id> --json
pm reverse run <plan-id> --approve <approval-token> --json
```

## Agent Rules

- Run pm connectors inspect github before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

## References

- [GitHub REST authentication](https://docs.github.com/en/rest/authentication/authenticating-to-the-rest-api)
- [GitHub App installation auth](https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/authenticating-as-a-github-app-installation)
- [GitHub pull requests REST API](https://docs.github.com/en/rest/pulls/pulls)
- [GitHub issues REST API](https://docs.github.com/en/rest/issues/issues)
- [GitHub issue comments REST API](https://docs.github.com/en/rest/issues/comments)
- [GitHub labels REST API](https://docs.github.com/en/rest/issues/labels)
- [GitHub commits REST API](https://docs.github.com/en/rest/commits/commits)
- [GitHub branches REST API](https://docs.github.com/en/rest/branches/branches)
- [GitHub releases REST API](https://docs.github.com/en/rest/releases/releases)
- [GitHub Actions workflows REST API](https://docs.github.com/en/rest/actions/workflows)
- [GitHub Actions workflow runs REST API](https://docs.github.com/en/rest/actions/workflow-runs)
- [GitHub Actions artifacts REST API](https://docs.github.com/en/rest/actions/artifacts)
- [GitHub repository contents REST API](https://docs.github.com/en/rest/repos/contents)
