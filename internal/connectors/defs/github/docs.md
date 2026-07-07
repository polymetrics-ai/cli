# Overview

GitHub reads 33 stream(s), and writes through 67 action(s).

The connector now declares a JSON-first command surface in `cli_surface.json`. This surface is a
docs, validation, and safe dispatch contract for gh-inspired GitHub commands. Commands mapped to
existing streams are ETL reads, commands mapped to existing write actions are reverse ETL command
plans, and unsupported local workflow commands such as clone, checkout, browser, completion, alias,
extension, and local config are labeled separately.

Readable streams: `repository`, `issues`, `pull_requests`, `branches`, `commits`, `tags`,
`releases`, `labels`, `milestones`, `issue_comments`, `pull_request_review_comments`,
`collaborators`, `contributors`, `stargazers`, `subscribers`, `workflows`, `workflow_runs`,
`workflow_artifacts`, `deployments`, `commit_comments`, `deploy_keys`, `webhooks`, `environments`,
`forks`, `invitations`, `issue_events`, `code_scanning_alerts`, `dependabot_alerts`,
`secret_scanning_alerts`, `security_advisories`, `repo_rulesets`, `autolinks`, `languages`.

Write actions: `create_issue`, `update_issue`, `comment_issue`, `close_issue`,
`create_pull_request`, `update_pull_request`, `close_pull_request`, `request_reviewers`,
`merge_pull_request`, `create_label`, `update_label`, `delete_label`, `create_milestone`,
`update_milestone`, `delete_milestone`, `create_release`, `update_release`, `delete_release`,
`dispatch_workflow`, `rerun_workflow_run`, `cancel_workflow_run`, `delete_workflow_run`,
`create_pull_request_review`, `create_or_update_file`, `delete_file`, `create_webhook`,
`update_webhook`, `delete_webhook`, `create_deploy_key`, `delete_deploy_key`,
`create_or_update_environment`, `delete_environment`, `create_commit_comment`,
`update_commit_comment`, `delete_commit_comment`, `update_issue_comment`, `delete_issue_comment`,
`lock_issue`, `unlock_issue`, `set_issue_labels`, `add_issue_labels`, `remove_issue_label`,
`add_issue_assignees`, `remove_issue_assignees`, `create_review_comment`, `update_review_comment`,
`delete_review_comment`, `submit_pull_request_review`, `dismiss_pull_request_review`,
`update_pull_request_branch`, `update_release_asset`, `delete_release_asset`, `replace_repo_topics`,
`add_collaborator`, `remove_collaborator`, `create_ref`, `update_ref`, `delete_ref`, `merge_branch`,
`update_code_scanning_alert`, `update_dependabot_alert`, `create_deployment`, `create_fork`,
`create_repo_ruleset`, `update_repo_ruleset`, `delete_repo_ruleset`, `update_secret_scanning_alert`.

Service API documentation: https://docs.github.com/en/rest.

## Auth setup

Connection fields:

- `app_id` (optional, string); GitHub App ID for auth_type=github_app.
- `auth_type` (optional, string).
- `base_url` (optional, string); default `https://api.github.com`; format `uri`; GitHub API base URL
  override for tests or GitHub Enterprise.
- `installation_id` (optional, string); GitHub App installation ID for auth_type=github_app.
- `installation_permissions` (optional, string); Optional JSON object of requested GitHub App
  installation-token permissions.
- `installation_repositories` (optional, string); Optional comma-separated repository names for a
  restricted installation token.
- `installation_repository_ids` (optional, string); Optional comma-separated repository IDs for a
  restricted installation token.
- `owner` (required, string); Repository owner (user or organization login).
- `private_key` (optional, secret, string); GitHub App PEM private key for auth_type=github_app.
  Never logged.
- `private_key_base64` (optional, secret, string); Base64-encoded GitHub App PEM private key for
  auth_type=github_app (alternative to private_key). Never logged.
- `public_access` (optional, string); Explicit opt-in for unauthenticated (public) reads. Set to any
  non-empty value (e.g. 'true') to allow reads with no token/app credentials configured.
- `repo` (required, string); Repository name (without the owner prefix).
- `since` (optional, string); format `date-time`; Lower-bound timestamp for issues, issue_comments,
  and pull_request_review_comments (incremental start; also usable as a fresh-sync start_config_key
  value).
- `token` (optional, secret, string); GitHub token: classic PAT, fine-grained PAT, OAuth token,
  GitHub Actions GITHUB_TOKEN, or a pre-generated installation access token. Used only for Bearer
  auth; never logged.

Secret fields are redacted in logs and write previews: `private_key`, `private_key_base64`, `token`.

Default configuration values: `base_url=https://api.github.com`.

Authentication behavior:

- Bearer token authentication using `secrets.token` when `{{ secrets.token }}`.
- Connector-specific authentication when `{{ config.app_id }}`.
- No authentication when `{{ config.public_access }}`.
- No authentication when `{{ config.auth_type in ['public', 'none', 'anonymous', 'unauthenticated']
  }}`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/repos/{{ config.owner }}/{{ config.repo }}`.

## Connector command writes

Mapped `reverse_etl` commands never mutate GitHub directly from a plain command invocation. The
provider-style command creates a stored reverse plan with an approval token, optional preview uses
the connector write dry-run path, and execution requires the same stored plan plus `--approve`.

Example:

```bash
pm github issue close --issue-number 101 --credential github-token
pm github issue close --plan <plan-id> --preview --json
pm github issue close --plan <plan-id> --approve <approval-token> --json
```

JSON plan and preview output redacts approval tokens, approval token hashes, and raw command payload
records. Commands without explicit `record.*` flag mappings remain blocked until their input model is
declared.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

Pagination by stream: none: `languages`; page_number: `repository`, `issues`, `pull_requests`,
`branches`, `commits`, `tags`, `releases`, `labels`, `milestones`, `issue_comments`,
`pull_request_review_comments`, `collaborators`, `contributors`, `stargazers`, `subscribers`,
`workflows`, `workflow_runs`, `workflow_artifacts`, `deployments`, `commit_comments`, `deploy_keys`,
`webhooks`, `environments`, `forks`, `invitations`, `issue_events`, `code_scanning_alerts`,
`dependabot_alerts`, `secret_scanning_alerts`, `security_advisories`, `repo_rulesets`, `autolinks`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `repository`: GET `/repos/{{ config.owner }}/{{ config.repo }}` - single-object response; records
  path `.`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 100; computed output fields `repository`.
- `issues`: GET `/repos/{{ config.owner }}/{{ config.repo }}/issues` - records path `.`; drops
  records where `pull_request` is present; query `state`=`all`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; incremental cursor
  `updated_at`; sent as `since`; formatted as `rfc3339`; initial lower bound from `since`; computed
  output fields `repository`, `user_id`, `user_login`.
- `pull_requests`: GET `/repos/{{ config.owner }}/{{ config.repo }}/pulls` - records path `.`; query
  `state`=`all`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 100; incremental cursor `updated_at`; formatted as `rfc3339`; computed output fields
  `base_ref`, `base_sha`, `head_ref`, `head_sha`, `repository`, `user_id`, `user_login`.
- `branches`: GET `/repos/{{ config.owner }}/{{ config.repo }}/branches` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `commit_sha`, `commit_url`, `repository`.
- `commits`: GET `/repos/{{ config.owner }}/{{ config.repo }}/commits` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `author_id`, `author_login`, `commit_author_date`,
  `commit_author_email`, `commit_author_name`, `commit_committer_date`, `commit_committer_email`,
  `commit_committer_name`, `commit_message`, `committer_id`, `committer_login`, `repository`.
- `tags`: GET `/repos/{{ config.owner }}/{{ config.repo }}/tags` - records path `.`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; computed
  output fields `commit_sha`, `commit_url`, `repository`.
- `releases`: GET `/repos/{{ config.owner }}/{{ config.repo }}/releases` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; incremental cursor `published_at`; formatted as `rfc3339`; computed output fields
  `assets_count`, `author_login`, `repository`.
- `labels`: GET `/repos/{{ config.owner }}/{{ config.repo }}/labels` - records path `.`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; computed
  output fields `repository`.
- `milestones`: GET `/repos/{{ config.owner }}/{{ config.repo }}/milestones` - records path `.`;
  query `state`=`all`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 100; incremental cursor `updated_at`; formatted as `rfc3339`; computed
  output fields `creator_login`, `repository`.
- `issue_comments`: GET `/repos/{{ config.owner }}/{{ config.repo }}/issues/comments` - records path
  `.`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page
  size 100; incremental cursor `updated_at`; sent as `since`; formatted as `rfc3339`; initial lower
  bound from `since`; computed output fields `repository`, `user_id`, `user_login`.
- `pull_request_review_comments`: GET `/repos/{{ config.owner }}/{{ config.repo }}/pulls/comments` -
  records path `.`; page-number pagination; page parameter `page`; size parameter `per_page`; starts
  at 1; page size 100; incremental cursor `updated_at`; sent as `since`; formatted as `rfc3339`;
  initial lower bound from `since`; computed output fields `repository`, `user_login`.
- `collaborators`: GET `/repos/{{ config.owner }}/{{ config.repo }}/collaborators` - records path
  `.`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page
  size 100; computed output fields `relation`, `repository`.
- `contributors`: GET `/repos/{{ config.owner }}/{{ config.repo }}/contributors` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `relation`, `repository`.
- `stargazers`: GET `/repos/{{ config.owner }}/{{ config.repo }}/stargazers` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `relation`, `repository`.
- `subscribers`: GET `/repos/{{ config.owner }}/{{ config.repo }}/subscribers` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `relation`, `repository`.
- `workflows`: GET `/repos/{{ config.owner }}/{{ config.repo }}/actions/workflows` - records path
  `workflows`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at
  1; page size 100; computed output fields `repository`.
- `workflow_runs`: GET `/repos/{{ config.owner }}/{{ config.repo }}/actions/runs` - records path
  `workflow_runs`; page-number pagination; page parameter `page`; size parameter `per_page`; starts
  at 1; page size 100; computed output fields `repository`.
- `workflow_artifacts`: GET `/repos/{{ config.owner }}/{{ config.repo }}/actions/artifacts` -
  records path `artifacts`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100; computed output fields `repository`, `workflow_run_id`.
- `deployments`: GET `/repos/{{ config.owner }}/{{ config.repo }}/deployments` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `creator_login`, `repository`.
- `commit_comments`: GET `/repos/{{ config.owner }}/{{ config.repo }}/comments` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; incremental cursor `updated_at`; formatted as `rfc3339`; computed output fields `repository`,
  `user_id`, `user_login`.
- `deploy_keys`: GET `/repos/{{ config.owner }}/{{ config.repo }}/keys` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `repository`.
- `webhooks`: GET `/repos/{{ config.owner }}/{{ config.repo }}/hooks` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `config_url`, `repository`.
- `environments`: GET `/repos/{{ config.owner }}/{{ config.repo }}/environments` - records path
  `environments`; page-number pagination; page parameter `page`; size parameter `per_page`; starts
  at 1; page size 100; computed output fields `repository`.
- `forks`: GET `/repos/{{ config.owner }}/{{ config.repo }}/forks` - records path `.`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; computed
  output fields `owner_login`, `repository`.
- `invitations`: GET `/repos/{{ config.owner }}/{{ config.repo }}/invitations` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `invitee_login`, `inviter_login`, `repository`.
- `issue_events`: GET `/repos/{{ config.owner }}/{{ config.repo }}/issues/events` - records path
  `.`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page
  size 100; computed output fields `actor_login`, `repository`.
- `code_scanning_alerts`: GET `/repos/{{ config.owner }}/{{ config.repo }}/code-scanning/alerts` -
  records path `.`; page-number pagination; page parameter `page`; size parameter `per_page`; starts
  at 1; page size 100; incremental cursor `updated_at`; formatted as `rfc3339`; computed output
  fields `dismissed_by_login`, `repository`, `rule_id`, `rule_severity`, `tool_name`.
- `dependabot_alerts`: GET `/repos/{{ config.owner }}/{{ config.repo }}/dependabot/alerts` - records
  path `.`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 100; incremental cursor `updated_at`; formatted as `rfc3339`; computed output fields
  `dismissed_by_login`, `package_ecosystem`, `package_name`, `repository`.
- `secret_scanning_alerts`: GET `/repos/{{ config.owner }}/{{ config.repo }}/secret-scanning/alerts`
  - records path `.`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 100; incremental cursor `updated_at`; formatted as `rfc3339`; computed
  output fields `repository`, `resolved_by_login`.
- `security_advisories`: GET `/repos/{{ config.owner }}/{{ config.repo }}/security-advisories` -
  records path `.`; page-number pagination; page parameter `page`; size parameter `per_page`; starts
  at 1; page size 100; incremental cursor `updated_at`; formatted as `rfc3339`; computed output
  fields `author_login`, `publisher_login`, `repository`.
- `repo_rulesets`: GET `/repos/{{ config.owner }}/{{ config.repo }}/rulesets` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `repository`.
- `autolinks`: GET `/repos/{{ config.owner }}/{{ config.repo }}/autolinks` - records path `.`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `repository`.
- `languages`: GET `/repos/{{ config.owner }}/{{ config.repo }}/languages` - single-object response;
  records path `.`; computed output fields `repository`; emits passthrough records.

## Write actions & risks

Overall write risk: external GitHub API mutation.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_issue`: POST `/repos/{{ config.owner }}/{{ config.repo }}/issues` - kind `create`; body
  type `json`; required record fields `title`; accepted fields `assignees`, `body`, `labels`,
  `milestone`, `title`, `type`; risk: creates user-visible GitHub issue and may notify watchers.
- `update_issue`: PATCH `/repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number
  }}` - kind `update`; body type `json`; path fields `issue_number`; required record fields
  `issue_number`; accepted fields `assignees`, `body`, `issue_number`, `labels`, `milestone`,
  `state`, `state_reason`, `title`, `type`; risk: mutates existing GitHub issue or pull request
  issue metadata.
- `comment_issue`: POST `/repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number
  }}/comments` - kind `create`; body type `json`; path fields `issue_number`; body fields `body`;
  required record fields `issue_number`, `body`; accepted fields `body`, `issue_number`; risk:
  creates user-visible comment and may notify participants.
- `close_issue`: PATCH `/repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number
  }}` - kind `update`; body type `json`; path fields `issue_number`; required record fields
  `issue_number`; accepted fields `comment`, `issue_number`, `state_reason`; risk: closes existing
  GitHub issue.
- `create_pull_request`: POST `/repos/{{ config.owner }}/{{ config.repo }}/pulls` - kind `create`;
  body type `json`; required record fields `head`, `base`; accepted fields `assignees`, `base`,
  `body`, `draft`, `head`, `issue`, `labels`, `maintainer_can_modify`, `milestone`, `reviewers`,
  `team_reviewers`, `title`; risk: creates user-visible pull request and may notify
  watchers/reviewers.
- `update_pull_request`: PATCH `/repos/{{ config.owner }}/{{ config.repo }}/pulls/{{
  record.pull_number }}` - kind `update`; body type `json`; path fields `pull_number`; required
  record fields `pull_number`; accepted fields `assignees`, `base`, `body`, `labels`,
  `maintainer_can_modify`, `milestone`, `pull_number`, `reviewers`, `state`, `team_reviewers`,
  `title`; risk: mutates existing GitHub pull request.
- `close_pull_request`: PATCH `/repos/{{ config.owner }}/{{ config.repo }}/pulls/{{
  record.pull_number }}` - kind `update`; body type `json`; path fields `pull_number`; required
  record fields `pull_number`; accepted fields `comment`, `pull_number`; risk: closes existing
  GitHub pull request.
- `request_reviewers`: POST `/repos/{{ config.owner }}/{{ config.repo }}/pulls/{{ record.pull_number
  }}/requested_reviewers` - kind `create`; body type `json`; path fields `pull_number`; required
  record fields `pull_number`; accepted fields `pull_number`, `reviewers`, `team_reviewers`; risk:
  notifies requested GitHub reviewers.
- `merge_pull_request`: PUT `/repos/{{ config.owner }}/{{ config.repo }}/pulls/{{ record.pull_number
  }}/merge` - kind `update`; body type `json`; path fields `pull_number`; required record fields
  `pull_number`; accepted fields `commit_message`, `commit_title`, `merge_method`, `pull_number`,
  `sha`; risk: irreversibly changes repository history unless branch protection blocks merge.
- `create_label`: POST `/repos/{{ config.owner }}/{{ config.repo }}/labels` - kind `create`; body
  type `json`; required record fields `name`, `color`; accepted fields `color`, `description`,
  `name`; risk: changes repository taxonomy used by issues and pull requests.
- `update_label`: PATCH `/repos/{{ config.owner }}/{{ config.repo }}/labels/{{ record.name }}` -
  kind `update`; body type `json`; path fields `name`; body fields `new_name`, `color`,
  `description`; required record fields `name`; accepted fields `color`, `description`, `name`,
  `new_name`; risk: renames or changes labels already used by issues and pull requests.
- `delete_label`: DELETE `/repos/{{ config.owner }}/{{ config.repo }}/labels/{{ record.name }}` -
  kind `delete`; body type `none`; path fields `name`; required record fields `name`; accepted
  fields `name`; risk: removes a label from the repository and existing issue metadata.
- `create_milestone`: POST `/repos/{{ config.owner }}/{{ config.repo }}/milestones` - kind `create`;
  body type `json`; required record fields `title`; accepted fields `description`, `due_on`,
  `state`, `title`; risk: creates planning metadata visible to repository collaborators.
- `update_milestone`: PATCH `/repos/{{ config.owner }}/{{ config.repo }}/milestones/{{
  record.milestone_number }}` - kind `update`; body type `json`; path fields `milestone_number`;
  required record fields `milestone_number`; accepted fields `description`, `due_on`,
  `milestone_number`, `state`, `title`; risk: changes planning metadata used by issues and pull
  requests.
- `delete_milestone`: DELETE `/repos/{{ config.owner }}/{{ config.repo }}/milestones/{{
  record.milestone_number }}` - kind `delete`; body type `none`; path fields `milestone_number`;
  required record fields `milestone_number`; accepted fields `milestone_number`; risk: removes
  repository planning metadata from GitHub.
- `create_release`: POST `/repos/{{ config.owner }}/{{ config.repo }}/releases` - kind `create`;
  body type `json`; required record fields `tag_name`; accepted fields `body`, `draft`,
  `generate_release_notes`, `make_latest`, `name`, `prerelease`, `tag_name`, `target_commitish`;
  risk: publishes release metadata and may notify repository watchers.
- `update_release`: PATCH `/repos/{{ config.owner }}/{{ config.repo }}/releases/{{ record.release_id
  }}` - kind `update`; body type `json`; path fields `release_id`; required record fields
  `release_id`; accepted fields `body`, `draft`, `generate_release_notes`, `make_latest`, `name`,
  `prerelease`, `release_id`, `tag_name`, `target_commitish`; risk: changes published release
  metadata.
- `delete_release`: DELETE `/repos/{{ config.owner }}/{{ config.repo }}/releases/{{
  record.release_id }}` - kind `delete`; body type `none`; path fields `release_id`; required record
  fields `release_id`; accepted fields `release_id`; risk: removes release metadata from GitHub;
  tags are not deleted by this action.
- `dispatch_workflow`: POST `/repos/{{ config.owner }}/{{ config.repo }}/actions/workflows/{{
  record.workflow_id }}/dispatches` - kind `create`; body type `json`; path fields `workflow_id`;
  required record fields `workflow_id`, `ref`; accepted fields `inputs`, `ref`, `workflow_id`; risk:
  starts CI/CD automation that may deploy, publish, or mutate external systems.
- `rerun_workflow_run`: POST `/repos/{{ config.owner }}/{{ config.repo }}/actions/runs/{{
  record.run_id }}/rerun` - kind `custom`; body type `none`; path fields `run_id`; required record
  fields `run_id`; accepted fields `run_id`; risk: reruns CI/CD automation and consumes workflow
  minutes.
- `cancel_workflow_run`: POST `/repos/{{ config.owner }}/{{ config.repo }}/actions/runs/{{
  record.run_id }}/cancel` - kind `custom`; body type `none`; path fields `run_id`; required record
  fields `run_id`; accepted fields `run_id`; risk: interrupts in-flight CI/CD automation.
- `delete_workflow_run`: DELETE `/repos/{{ config.owner }}/{{ config.repo }}/actions/runs/{{
  record.run_id }}` - kind `delete`; body type `none`; path fields `run_id`; required record fields
  `run_id`; accepted fields `run_id`; risk: removes workflow run history from GitHub.
- `create_pull_request_review`: POST `/repos/{{ config.owner }}/{{ config.repo }}/pulls/{{
  record.pull_number }}/reviews` - kind `create`; body type `json`; path fields `pull_number`;
  required record fields `pull_number`; accepted fields `body`, `comments`, `commit_id`, `event`,
  `pull_number`; risk: submits reviewer feedback and may approve or request changes on a pull
  request.
- `create_or_update_file`: PUT `/repos/{{ config.owner }}/{{ config.repo }}/contents/{{ record.path
  }}` - kind `upsert`; body type `json`; path fields `path`; required record fields `path`,
  `message`, `content`; accepted fields `author`, `branch`, `committer`, `content`, `message`,
  `path`, `sha`; risk: writes a commit to the repository and may trigger CI/CD.
- `delete_file`: DELETE `/repos/{{ config.owner }}/{{ config.repo }}/contents/{{ record.path }}` -
  kind `delete`; body type `json`; path fields `path`; body fields `message`, `sha`, `branch`,
  `committer`, `author`; required record fields `path`, `message`, `sha`; accepted fields `author`,
  `branch`, `committer`, `message`, `path`, `sha`; risk: writes a commit that removes a file from
  the repository.
- `create_webhook`: POST `/repos/{{ config.owner }}/{{ config.repo }}/hooks` - kind `create`; body
  type `json`; required record fields `config`; accepted fields `active`, `config`, `events`,
  `name`; risk: registers an outbound webhook that will receive repository event payloads.
- `update_webhook`: PATCH `/repos/{{ config.owner }}/{{ config.repo }}/hooks/{{ record.hook_id }}` -
  kind `update`; body type `json`; path fields `hook_id`; required record fields `hook_id`; accepted
  fields `active`, `add_events`, `config`, `events`, `hook_id`, `remove_events`; risk: changes an
  existing webhook's target URL, secret, or event subscriptions.
- `delete_webhook`: DELETE `/repos/{{ config.owner }}/{{ config.repo }}/hooks/{{ record.hook_id }}`
  - kind `delete`; body type `none`; path fields `hook_id`; required record fields `hook_id`;
  accepted fields `hook_id`; missing records treated as success for status `404`; risk: removes a
  webhook; the target will stop receiving repository event payloads.
- `create_deploy_key`: POST `/repos/{{ config.owner }}/{{ config.repo }}/keys` - kind `create`; body
  type `json`; required record fields `key`; accepted fields `key`, `read_only`, `title`; risk:
  grants a new SSH public key deploy access to the repository.
- `delete_deploy_key`: DELETE `/repos/{{ config.owner }}/{{ config.repo }}/keys/{{ record.key_id }}`
  - kind `delete`; body type `none`; path fields `key_id`; required record fields `key_id`; accepted
  fields `key_id`; missing records treated as success for status `404`; risk: revokes an SSH deploy
  key's access to the repository.
- `create_or_update_environment`: PUT `/repos/{{ config.owner }}/{{ config.repo }}/environments/{{
  record.environment_name }}` - kind `upsert`; body type `json`; path fields `environment_name`;
  required record fields `environment_name`; accepted fields `deployment_branch_policy`,
  `environment_name`, `prevent_self_review`, `reviewers`, `wait_timer`; risk: creates or changes a
  deployment environment's protection rules and reviewers.
- `delete_environment`: DELETE `/repos/{{ config.owner }}/{{ config.repo }}/environments/{{
  record.environment_name }}` - kind `delete`; body type `none`; path fields `environment_name`;
  required record fields `environment_name`; accepted fields `environment_name`; missing records
  treated as success for status `404`; risk: removes a deployment environment and its protection
  rules.
- `create_commit_comment`: POST `/repos/{{ config.owner }}/{{ config.repo }}/commits/{{
  record.commit_sha }}/comments` - kind `create`; body type `json`; path fields `commit_sha`;
  required record fields `commit_sha`, `body`; accepted fields `body`, `commit_sha`, `line`, `path`,
  `position`; risk: creates a user-visible comment attached to a specific commit.
- `update_commit_comment`: PATCH `/repos/{{ config.owner }}/{{ config.repo }}/comments/{{
  record.comment_id }}` - kind `update`; body type `json`; path fields `comment_id`; body fields
  `body`; required record fields `comment_id`, `body`; accepted fields `body`, `comment_id`; risk:
  changes the text of an existing commit comment.
- `delete_commit_comment`: DELETE `/repos/{{ config.owner }}/{{ config.repo }}/comments/{{
  record.comment_id }}` - kind `delete`; body type `none`; path fields `comment_id`; required record
  fields `comment_id`; accepted fields `comment_id`; missing records treated as success for status
  `404`; risk: removes a commit comment.
- `update_issue_comment`: PATCH `/repos/{{ config.owner }}/{{ config.repo }}/issues/comments/{{
  record.comment_id }}` - kind `update`; body type `json`; path fields `comment_id`; body fields
  `body`; required record fields `comment_id`, `body`; accepted fields `body`, `comment_id`; risk:
  changes the text of an existing issue or pull request comment.
- `delete_issue_comment`: DELETE `/repos/{{ config.owner }}/{{ config.repo }}/issues/comments/{{
  record.comment_id }}` - kind `delete`; body type `none`; path fields `comment_id`; required record
  fields `comment_id`; accepted fields `comment_id`; missing records treated as success for status
  `404`; risk: removes an issue or pull request comment.
- `lock_issue`: PUT `/repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number
  }}/lock` - kind `update`; body type `json`; path fields `issue_number`; body fields `lock_reason`;
  required record fields `issue_number`; accepted fields `issue_number`, `lock_reason`; risk:
  prevents further comments from non-collaborators on an issue or pull request.
- `unlock_issue`: DELETE `/repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number
  }}/lock` - kind `update`; body type `none`; path fields `issue_number`; required record fields
  `issue_number`; accepted fields `issue_number`; risk: reopens an issue or pull request to comments
  from non-collaborators.
- `set_issue_labels`: PUT `/repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number
  }}/labels` - kind `update`; body type `json`; path fields `issue_number`; body fields `labels`;
  required record fields `issue_number`; accepted fields `issue_number`, `labels`; risk: replaces
  every label on an issue or pull request, removing any not listed.
- `add_issue_labels`: POST `/repos/{{ config.owner }}/{{ config.repo }}/issues/{{
  record.issue_number }}/labels` - kind `update`; body type `json`; path fields `issue_number`; body
  fields `labels`; required record fields `issue_number`, `labels`; accepted fields `issue_number`,
  `labels`; risk: adds labels to an issue or pull request without removing existing ones.
- `remove_issue_label`: DELETE `/repos/{{ config.owner }}/{{ config.repo }}/issues/{{
  record.issue_number }}/labels/{{ record.name }}` - kind `delete`; body type `none`; path fields
  `issue_number`, `name`; required record fields `issue_number`, `name`; accepted fields
  `issue_number`, `name`; missing records treated as success for status `404`; risk: removes a
  single label from an issue or pull request.
- `add_issue_assignees`: POST `/repos/{{ config.owner }}/{{ config.repo }}/issues/{{
  record.issue_number }}/assignees` - kind `update`; body type `json`; path fields `issue_number`;
  body fields `assignees`; required record fields `issue_number`, `assignees`; accepted fields
  `assignees`, `issue_number`; risk: assigns additional GitHub users to an issue or pull request and
  may notify them.
- `remove_issue_assignees`: DELETE `/repos/{{ config.owner }}/{{ config.repo }}/issues/{{
  record.issue_number }}/assignees` - kind `update`; body type `json`; path fields `issue_number`;
  body fields `assignees`; required record fields `issue_number`, `assignees`; accepted fields
  `assignees`, `issue_number`; risk: removes assignees from an issue or pull request.
- `create_review_comment`: POST `/repos/{{ config.owner }}/{{ config.repo }}/pulls/{{
  record.pull_number }}/comments` - kind `create`; body type `json`; path fields `pull_number`;
  required record fields `pull_number`, `body`, `commit_id`, `path`; accepted fields `body`,
  `commit_id`, `in_reply_to`, `line`, `path`, `pull_number`, `side`, `start_line`, `start_side`;
  risk: creates a user-visible inline review comment on a pull request diff.
- `update_review_comment`: PATCH `/repos/{{ config.owner }}/{{ config.repo }}/pulls/comments/{{
  record.comment_id }}` - kind `update`; body type `json`; path fields `comment_id`; body fields
  `body`; required record fields `comment_id`, `body`; accepted fields `body`, `comment_id`; risk:
  changes the text of an existing pull request review comment.
- `delete_review_comment`: DELETE `/repos/{{ config.owner }}/{{ config.repo }}/pulls/comments/{{
  record.comment_id }}` - kind `delete`; body type `none`; path fields `comment_id`; required record
  fields `comment_id`; accepted fields `comment_id`; missing records treated as success for status
  `404`; risk: removes a pull request review comment.
- `submit_pull_request_review`: POST `/repos/{{ config.owner }}/{{ config.repo }}/pulls/{{
  record.pull_number }}/reviews/{{ record.review_id }}/events` - kind `update`; body type `json`;
  path fields `pull_number`, `review_id`; body fields `body`, `event`; required record fields
  `pull_number`, `review_id`, `event`; accepted fields `body`, `event`, `pull_number`, `review_id`;
  risk: submits a pending pull request review, which may approve or request changes.
- `dismiss_pull_request_review`: PUT `/repos/{{ config.owner }}/{{ config.repo }}/pulls/{{
  record.pull_number }}/reviews/{{ record.review_id }}/dismissals` - kind `update`; body type
  `json`; path fields `pull_number`, `review_id`; body fields `message`, `event`; required record
  fields `pull_number`, `review_id`, `message`; accepted fields `event`, `message`, `pull_number`,
  `review_id`; risk: dismisses an existing pull request review, clearing its approval status.
- `update_pull_request_branch`: PUT `/repos/{{ config.owner }}/{{ config.repo }}/pulls/{{
  record.pull_number }}/update-branch` - kind `update`; body type `json`; path fields `pull_number`;
  body fields `expected_head_sha`; required record fields `pull_number`; accepted fields
  `expected_head_sha`, `pull_number`; risk: merges the base branch into the pull request's head
  branch, adding a merge commit.
- `update_release_asset`: PATCH `/repos/{{ config.owner }}/{{ config.repo }}/releases/assets/{{
  record.asset_id }}` - kind `update`; body type `json`; path fields `asset_id`; required record
  fields `asset_id`; accepted fields `asset_id`, `label`, `name`, `state`; risk: changes a release
  asset's file name or label.
- `delete_release_asset`: DELETE `/repos/{{ config.owner }}/{{ config.repo }}/releases/assets/{{
  record.asset_id }}` - kind `delete`; body type `none`; path fields `asset_id`; required record
  fields `asset_id`; accepted fields `asset_id`; missing records treated as success for status
  `404`; risk: removes a downloadable asset from a published release.
- `replace_repo_topics`: PUT `/repos/{{ config.owner }}/{{ config.repo }}/topics` - kind `update`;
  body type `json`; required record fields `names`; accepted fields `names`; risk: replaces the
  repository's entire topic list, removing any topic not listed.
- `add_collaborator`: PUT `/repos/{{ config.owner }}/{{ config.repo }}/collaborators/{{
  record.username }}` - kind `upsert`; body type `json`; path fields `username`; body fields
  `permission`; required record fields `username`; accepted fields `permission`, `username`; risk:
  grants a GitHub user access to the repository and may send an invitation email.
- `remove_collaborator`: DELETE `/repos/{{ config.owner }}/{{ config.repo }}/collaborators/{{
  record.username }}` - kind `delete`; body type `none`; path fields `username`; required record
  fields `username`; accepted fields `username`; missing records treated as success for status
  `404`; risk: revokes a collaborator's access to the repository.
- `create_ref`: POST `/repos/{{ config.owner }}/{{ config.repo }}/git/refs` - kind `create`; body
  type `json`; required record fields `ref`, `sha`; accepted fields `ref`, `sha`; risk: creates a
  new branch or tag ref pointing at the given commit SHA.
- `update_ref`: PATCH `/repos/{{ config.owner }}/{{ config.repo }}/git/refs/{{ record.ref }}` - kind
  `update`; body type `json`; path fields `ref`; body fields `sha`, `force`; required record fields
  `ref`, `sha`; accepted fields `force`, `ref`, `sha`; risk: moves an existing branch or tag ref to
  a different commit SHA, potentially discarding history.
- `delete_ref`: DELETE `/repos/{{ config.owner }}/{{ config.repo }}/git/refs/{{ record.ref }}` -
  kind `delete`; body type `none`; path fields `ref`; required record fields `ref`; accepted fields
  `ref`; missing records treated as success for status `404`, `422`; confirmation `destructive`;
  risk: permanently deletes a branch or tag ref.
- `merge_branch`: POST `/repos/{{ config.owner }}/{{ config.repo }}/merges` - kind `create`; body
  type `json`; required record fields `base`, `head`; accepted fields `base`, `commit_message`,
  `head`; risk: creates a merge commit combining the head ref into the base branch.
- `update_code_scanning_alert`: PATCH `/repos/{{ config.owner }}/{{ config.repo
  }}/code-scanning/alerts/{{ record.alert_number }}` - kind `update`; body type `json`; path fields
  `alert_number`; required record fields `alert_number`, `state`; accepted fields `alert_number`,
  `dismissed_comment`, `dismissed_reason`, `state`; risk: changes a code scanning alert's triage
  state, which can suppress a real security finding.
- `update_dependabot_alert`: PATCH `/repos/{{ config.owner }}/{{ config.repo }}/dependabot/alerts/{{
  record.alert_number }}` - kind `update`; body type `json`; path fields `alert_number`; required
  record fields `alert_number`, `state`; accepted fields `alert_number`, `dismissed_comment`,
  `dismissed_reason`, `state`; risk: changes a dependabot alert's triage state, which can suppress a
  real vulnerability finding.
- `create_deployment`: POST `/repos/{{ config.owner }}/{{ config.repo }}/deployments` - kind
  `create`; body type `json`; required record fields `ref`; accepted fields `auto_merge`,
  `description`, `environment`, `payload`, `production_environment`, `ref`, `required_contexts`,
  `task`, `transient_environment`; risk: records a new deployment and may trigger CI/CD deployment
  automation.
- `create_fork`: POST `/repos/{{ config.owner }}/{{ config.repo }}/forks` - kind `create`; body type
  `json`; accepted fields `default_branch_only`, `name`, `organization`; risk: creates a new
  repository forked from this one, under the caller's account or a target organization.
- `create_repo_ruleset`: POST `/repos/{{ config.owner }}/{{ config.repo }}/rulesets` - kind
  `create`; body type `json`; required record fields `name`, `enforcement`; accepted fields
  `bypass_actors`, `conditions`, `enforcement`, `name`, `rules`, `target`; risk: creates a
  repository ruleset that can block pushes, merges, or deletions repo-wide once active.
- `update_repo_ruleset`: PUT `/repos/{{ config.owner }}/{{ config.repo }}/rulesets/{{
  record.ruleset_id }}` - kind `update`; body type `json`; path fields `ruleset_id`; required record
  fields `ruleset_id`; accepted fields `bypass_actors`, `conditions`, `enforcement`, `name`,
  `rules`, `ruleset_id`, `target`; risk: changes an existing repository ruleset's enforcement or
  rule set, which can block pushes, merges, or deletions repo-wide.
- `delete_repo_ruleset`: DELETE `/repos/{{ config.owner }}/{{ config.repo }}/rulesets/{{
  record.ruleset_id }}` - kind `delete`; body type `none`; path fields `ruleset_id`; required record
  fields `ruleset_id`; accepted fields `ruleset_id`; missing records treated as success for status
  `404`; risk: removes a repository ruleset, lifting any push/merge/deletion restrictions it
  enforced.
- `update_secret_scanning_alert`: PATCH `/repos/{{ config.owner }}/{{ config.repo
  }}/secret-scanning/alerts/{{ record.alert_number }}` - kind `update`; body type `json`; path
  fields `alert_number`; required record fields `alert_number`, `state`; accepted fields
  `alert_number`, `resolution`, `resolution_comment`, `state`; risk: changes a secret scanning
  alert's triage state, which can suppress a real leaked-credential finding.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 33 stream-backed endpoint group(s), 67 write-backed endpoint group(s).
- GitHub CLI parity is intentionally staged. The current metadata covers selected `gh` command
  families modeled in this slice and maps implemented commands to current stream/write names. Runtime
  dispatch is limited to stream reads, guarded direct reads, and reverse ETL write commands with
  explicit `record.*` flag mappings.
- GitHub Projects v2, discussions, gist, codespaces, organization-wide views, and several status or
  search commands require GraphQL or mixed REST/GraphQL coverage that is not modeled by this REST
  bundle yet.
- Secret and variable write commands are not exposed as reverse ETL actions until encryption,
  redaction, scope, and approval semantics are modeled explicitly.
- Raw `gh api` and `gh api graphql` style escape hatches are classified as unsafe unless constrained
  to connector auth, connector base URLs, allowlisted methods, mutation approval, and secret
  redaction.
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=10, deprecated=1, destructive_admin=5, duplicate_of=67, non_data_endpoint=9,
  out_of_scope=143, requires_elevated_scope=168.
