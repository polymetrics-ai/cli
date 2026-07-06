---
name: pm-github
description: GitHub connector knowledge and safe action guide.
---

# pm-github

## Purpose

Reads GitHub repository, issue, pull request, code, release, collaboration, Actions, security (code scanning/dependabot/secret scanning/advisories), webhook, deploy key, environment, and ruleset data, and writes approved reverse ETL actions through the GitHub REST API (Pass B full-surface expansion: 33 streams, 67 write actions).

## Icon

- asset: icons/github.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.github.com/en/rest/about-the-rest-api/breaking-changes

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- app_id
- auth_type
- base_url
- installation_id
- installation_permissions
- installation_repositories
- installation_repository_ids
- owner
- public_access
- repo
- since
- private_key (secret)
- private_key_base64 (secret)
- token (secret)

## ETL Streams

- repository:
  - primary key: node_id
  - cursor: updated_at
  - fields: created_at(), default_branch(), description(), forks_count(), full_name(), html_url(), id(), language(), name(), node_id(), open_issues_count(), private(), pushed_at(), repository(), stargazers_count(), updated_at(), watchers_count()
- issues:
  - primary key: node_id
  - cursor: updated_at
  - fields: author_association(), body(), closed_at(), comments(), created_at(), html_url(), id(), locked(), node_id(), number(), repository(), state(), state_reason(), title(), updated_at(), url(), user_id(), user_login()
- pull_requests:
  - primary key: node_id
  - cursor: updated_at
  - fields: author_association(), base_ref(), base_sha(), body(), closed_at(), comments(), created_at(), draft(), head_ref(), head_sha(), html_url(), id(), locked(), merge_commit_sha(), merged_at(), node_id(), number(), repository(), state(), title(), updated_at(), url(), user_id(), user_login()
- branches:
  - primary key: name
  - fields: commit_sha(), commit_url(), name(), protected(), repository()
- commits:
  - primary key: sha
  - cursor: commit_committer_date
  - fields: author_id(), author_login(), commit_author_date(), commit_author_email(), commit_author_name(), commit_committer_date(), commit_committer_email(), commit_committer_name(), commit_message(), committer_id(), committer_login(), html_url(), node_id(), repository(), sha(), url()
- tags:
  - primary key: name
  - fields: commit_sha(), commit_url(), name(), node_id(), repository(), tarball_url(), zipball_url()
- releases:
  - primary key: id
  - cursor: published_at
  - fields: assets_count(), author_login(), body(), created_at(), draft(), html_url(), id(), name(), node_id(), prerelease(), published_at(), repository(), tag_name(), target_commitish()
- labels:
  - primary key: name
  - fields: color(), default(), description(), id(), name(), node_id(), repository(), url()
- milestones:
  - primary key: number
  - cursor: updated_at
  - fields: closed_at(), closed_issues(), created_at(), creator_login(), description(), due_on(), id(), node_id(), number(), open_issues(), repository(), state(), title(), updated_at()
- issue_comments:
  - primary key: id
  - cursor: updated_at
  - fields: author_association(), body(), created_at(), html_url(), id(), issue_url(), node_id(), repository(), updated_at(), user_id(), user_login()
- pull_request_review_comments:
  - primary key: id
  - cursor: updated_at
  - fields: body(), commit_id(), created_at(), diff_hunk(), html_url(), id(), node_id(), original_commit_id(), original_position(), path(), position(), pull_request_review_id(), pull_request_url(), repository(), updated_at(), user_login()
- collaborators:
  - primary key: id
  - fields: contributions(), html_url(), id(), login(), node_id(), relation(), repository(), role_name(), site_admin(), type()
- contributors:
  - primary key: id
  - fields: contributions(), html_url(), id(), login(), node_id(), relation(), repository(), role_name(), site_admin(), type()
- stargazers:
  - primary key: id
  - fields: contributions(), html_url(), id(), login(), node_id(), relation(), repository(), role_name(), site_admin(), type()
- subscribers:
  - primary key: id
  - fields: contributions(), html_url(), id(), login(), node_id(), relation(), repository(), role_name(), site_admin(), type()
- workflows:
  - primary key: id
  - cursor: updated_at
  - fields: badge_url(), created_at(), html_url(), id(), name(), node_id(), path(), repository(), state(), updated_at()
- workflow_runs:
  - primary key: id
  - cursor: updated_at
  - fields: conclusion(), created_at(), event(), head_branch(), head_sha(), html_url(), id(), name(), node_id(), repository(), run_attempt(), run_number(), status(), updated_at(), workflow_id()
- workflow_artifacts:
  - primary key: id
  - cursor: updated_at
  - fields: archive_download_url(), created_at(), expired(), expires_at(), id(), name(), node_id(), repository(), size_in_bytes(), updated_at(), url(), workflow_run_id()
- deployments:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), creator_login(), description(), environment(), id(), node_id(), ref(), repository(), sha(), task(), updated_at()
- commit_comments:
  - primary key: id
  - cursor: updated_at
  - fields: author_association(), body(), commit_id(), created_at(), html_url(), id(), line(), node_id(), path(), position(), repository(), updated_at(), url(), user_id(), user_login()
- deploy_keys:
  - primary key: id
  - fields: added_by(), created_at(), enabled(), id(), key(), last_used(), read_only(), repository(), title(), url(), verified()
- webhooks:
  - primary key: id
  - cursor: updated_at
  - fields: active(), config_url(), created_at(), deliveries_url(), events(), id(), name(), ping_url(), repository(), test_url(), type(), updated_at(), url()
- environments:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), html_url(), id(), name(), node_id(), repository(), updated_at(), url()
- forks:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), default_branch(), forks_count(), full_name(), html_url(), id(), name(), node_id(), open_issues_count(), owner_login(), private(), pushed_at(), repository(), stargazers_count(), updated_at(), watchers_count()
- invitations:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), expired(), html_url(), id(), invitee_login(), inviter_login(), node_id(), permissions(), repository(), url()
- issue_events:
  - primary key: id
  - cursor: created_at
  - fields: actor_login(), commit_id(), commit_url(), created_at(), event(), id(), lock_reason(), node_id(), repository(), url()
- code_scanning_alerts:
  - primary key: number
  - cursor: updated_at
  - fields: created_at(), dismissed_at(), dismissed_by_login(), dismissed_comment(), dismissed_reason(), fixed_at(), html_url(), number(), repository(), rule_id(), rule_severity(), state(), tool_name(), updated_at(), url()
- dependabot_alerts:
  - primary key: number
  - cursor: updated_at
  - fields: auto_dismissed_at(), created_at(), dismissed_at(), dismissed_by_login(), dismissed_comment(), dismissed_reason(), fixed_at(), html_url(), number(), package_ecosystem(), package_name(), repository(), state(), updated_at(), url()
- secret_scanning_alerts:
  - primary key: number
  - cursor: updated_at
  - fields: created_at(), html_url(), number(), push_protection_bypassed(), repository(), resolution(), resolved_at(), resolved_by_login(), secret_type(), secret_type_display_name(), state(), updated_at(), url(), validity()
- security_advisories:
  - primary key: ghsa_id
  - cursor: updated_at
  - fields: author_login(), closed_at(), created_at(), cve_id(), ghsa_id(), html_url(), published_at(), publisher_login(), repository(), severity(), state(), summary(), updated_at(), url(), withdrawn_at()
- repo_rulesets:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), enforcement(), id(), name(), repository(), source(), source_type(), target(), updated_at()
- autolinks:
  - primary key: id
  - fields: id(), is_alphanumeric(), key_prefix(), repository(), updated_at(), url_template()
- languages:
  - primary key: repository
  - fields: repository()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_issue:
  - endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/issues
  - risk: creates user-visible GitHub issue and may notify watchers
- update_issue:
  - endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number }}
  - required fields: issue_number
  - risk: mutates existing GitHub issue or pull request issue metadata
- comment_issue:
  - endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number }}/comments
  - required fields: issue_number
  - optional fields: body
  - risk: creates user-visible comment and may notify participants
- close_issue:
  - endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number }}
  - required fields: issue_number
  - risk: closes existing GitHub issue
- create_pull_request:
  - endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/pulls
  - risk: creates user-visible pull request and may notify watchers/reviewers
- update_pull_request:
  - endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/pulls/{{ record.pull_number }}
  - required fields: pull_number
  - risk: mutates existing GitHub pull request
- close_pull_request:
  - endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/pulls/{{ record.pull_number }}
  - required fields: pull_number
  - risk: closes existing GitHub pull request
- request_reviewers:
  - endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/pulls/{{ record.pull_number }}/requested_reviewers
  - required fields: pull_number
  - risk: notifies requested GitHub reviewers
- merge_pull_request:
  - endpoint: PUT /repos/{{ config.owner }}/{{ config.repo }}/pulls/{{ record.pull_number }}/merge
  - required fields: pull_number
  - risk: irreversibly changes repository history unless branch protection blocks merge
- create_label:
  - endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/labels
  - risk: changes repository taxonomy used by issues and pull requests
- update_label:
  - endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/labels/{{ record.name }}
  - required fields: name
  - optional fields: new_name, color, description
  - risk: renames or changes labels already used by issues and pull requests
- delete_label:
  - endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/labels/{{ record.name }}
  - required fields: name
  - risk: removes a label from the repository and existing issue metadata
- create_milestone:
  - endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/milestones
  - risk: creates planning metadata visible to repository collaborators
- update_milestone:
  - endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/milestones/{{ record.milestone_number }}
  - required fields: milestone_number
  - risk: changes planning metadata used by issues and pull requests
- delete_milestone:
  - endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/milestones/{{ record.milestone_number }}
  - required fields: milestone_number
  - risk: removes repository planning metadata from GitHub
- create_release:
  - endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/releases
  - risk: publishes release metadata and may notify repository watchers
- update_release:
  - endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/releases/{{ record.release_id }}
  - required fields: release_id
  - risk: changes published release metadata
- delete_release:
  - endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/releases/{{ record.release_id }}
  - required fields: release_id
  - risk: removes release metadata from GitHub; tags are not deleted by this action
- dispatch_workflow:
  - endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/actions/workflows/{{ record.workflow_id }}/dispatches
  - required fields: workflow_id
  - risk: starts CI/CD automation that may deploy, publish, or mutate external systems
- rerun_workflow_run:
  - endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/actions/runs/{{ record.run_id }}/rerun
  - required fields: run_id
  - risk: reruns CI/CD automation and consumes workflow minutes
- cancel_workflow_run:
  - endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/actions/runs/{{ record.run_id }}/cancel
  - required fields: run_id
  - risk: interrupts in-flight CI/CD automation
- delete_workflow_run:
  - endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/actions/runs/{{ record.run_id }}
  - required fields: run_id
  - risk: removes workflow run history from GitHub
- create_pull_request_review:
  - endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/pulls/{{ record.pull_number }}/reviews
  - required fields: pull_number
  - risk: submits reviewer feedback and may approve or request changes on a pull request
- create_or_update_file:
  - endpoint: PUT /repos/{{ config.owner }}/{{ config.repo }}/contents/{{ record.path }}
  - required fields: path
  - risk: writes a commit to the repository and may trigger CI/CD
- delete_file:
  - endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/contents/{{ record.path }}
  - required fields: path
  - optional fields: message, sha, branch, committer, author
  - risk: writes a commit that removes a file from the repository
- create_webhook:
  - endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/hooks
  - risk: registers an outbound webhook that will receive repository event payloads
- update_webhook:
  - endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/hooks/{{ record.hook_id }}
  - required fields: hook_id
  - risk: changes an existing webhook's target URL, secret, or event subscriptions
- delete_webhook:
  - endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/hooks/{{ record.hook_id }}
  - required fields: hook_id
  - risk: removes a webhook; the target will stop receiving repository event payloads
- create_deploy_key:
  - endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/keys
  - risk: grants a new SSH public key deploy access to the repository
- delete_deploy_key:
  - endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/keys/{{ record.key_id }}
  - required fields: key_id
  - risk: revokes an SSH deploy key's access to the repository
- create_or_update_environment:
  - endpoint: PUT /repos/{{ config.owner }}/{{ config.repo }}/environments/{{ record.environment_name }}
  - required fields: environment_name
  - risk: creates or changes a deployment environment's protection rules and reviewers
- delete_environment:
  - endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/environments/{{ record.environment_name }}
  - required fields: environment_name
  - risk: removes a deployment environment and its protection rules
- create_commit_comment:
  - endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/commits/{{ record.commit_sha }}/comments
  - required fields: commit_sha
  - risk: creates a user-visible comment attached to a specific commit
- update_commit_comment:
  - endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/comments/{{ record.comment_id }}
  - required fields: comment_id
  - optional fields: body
  - risk: changes the text of an existing commit comment
- delete_commit_comment:
  - endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/comments/{{ record.comment_id }}
  - required fields: comment_id
  - risk: removes a commit comment
- update_issue_comment:
  - endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/issues/comments/{{ record.comment_id }}
  - required fields: comment_id
  - optional fields: body
  - risk: changes the text of an existing issue or pull request comment
- delete_issue_comment:
  - endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/issues/comments/{{ record.comment_id }}
  - required fields: comment_id
  - risk: removes an issue or pull request comment
- lock_issue:
  - endpoint: PUT /repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number }}/lock
  - required fields: issue_number
  - optional fields: lock_reason
  - risk: prevents further comments from non-collaborators on an issue or pull request
- unlock_issue:
  - endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number }}/lock
  - required fields: issue_number
  - risk: reopens an issue or pull request to comments from non-collaborators
- set_issue_labels:
  - endpoint: PUT /repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number }}/labels
  - required fields: issue_number
  - optional fields: labels
  - risk: replaces every label on an issue or pull request, removing any not listed
- add_issue_labels:
  - endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number }}/labels
  - required fields: issue_number
  - optional fields: labels
  - risk: adds labels to an issue or pull request without removing existing ones
- remove_issue_label:
  - endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number }}/labels/{{ record.name }}
  - required fields: issue_number, name
  - risk: removes a single label from an issue or pull request
- add_issue_assignees:
  - endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number }}/assignees
  - required fields: issue_number
  - optional fields: assignees
  - risk: assigns additional GitHub users to an issue or pull request and may notify them
- remove_issue_assignees:
  - endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number }}/assignees
  - required fields: issue_number
  - optional fields: assignees
  - risk: removes assignees from an issue or pull request
- create_review_comment:
  - endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/pulls/{{ record.pull_number }}/comments
  - required fields: pull_number
  - risk: creates a user-visible inline review comment on a pull request diff
- update_review_comment:
  - endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/pulls/comments/{{ record.comment_id }}
  - required fields: comment_id
  - optional fields: body
  - risk: changes the text of an existing pull request review comment
- delete_review_comment:
  - endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/pulls/comments/{{ record.comment_id }}
  - required fields: comment_id
  - risk: removes a pull request review comment
- submit_pull_request_review:
  - endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/pulls/{{ record.pull_number }}/reviews/{{ record.review_id }}/events
  - required fields: pull_number, review_id
  - optional fields: body, event
  - risk: submits a pending pull request review, which may approve or request changes
- dismiss_pull_request_review:
  - endpoint: PUT /repos/{{ config.owner }}/{{ config.repo }}/pulls/{{ record.pull_number }}/reviews/{{ record.review_id }}/dismissals
  - required fields: pull_number, review_id
  - optional fields: message, event
  - risk: dismisses an existing pull request review, clearing its approval status
- update_pull_request_branch:
  - endpoint: PUT /repos/{{ config.owner }}/{{ config.repo }}/pulls/{{ record.pull_number }}/update-branch
  - required fields: pull_number
  - optional fields: expected_head_sha
  - risk: merges the base branch into the pull request's head branch, adding a merge commit
- update_release_asset:
  - endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/releases/assets/{{ record.asset_id }}
  - required fields: asset_id
  - risk: changes a release asset's file name or label
- delete_release_asset:
  - endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/releases/assets/{{ record.asset_id }}
  - required fields: asset_id
  - risk: removes a downloadable asset from a published release
- replace_repo_topics:
  - endpoint: PUT /repos/{{ config.owner }}/{{ config.repo }}/topics
  - risk: replaces the repository's entire topic list, removing any topic not listed
- add_collaborator:
  - endpoint: PUT /repos/{{ config.owner }}/{{ config.repo }}/collaborators/{{ record.username }}
  - required fields: username
  - optional fields: permission
  - risk: grants a GitHub user access to the repository and may send an invitation email
- remove_collaborator:
  - endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/collaborators/{{ record.username }}
  - required fields: username
  - risk: revokes a collaborator's access to the repository
- create_ref:
  - endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/git/refs
  - risk: creates a new branch or tag ref pointing at the given commit SHA
- update_ref:
  - endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/git/refs/{{ record.ref }}
  - required fields: ref
  - optional fields: sha, force
  - risk: moves an existing branch or tag ref to a different commit SHA, potentially discarding history
- delete_ref:
  - endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/git/refs/{{ record.ref }}
  - required fields: ref
  - risk: permanently deletes a branch or tag ref
- merge_branch:
  - endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/merges
  - risk: creates a merge commit combining the head ref into the base branch
- update_code_scanning_alert:
  - endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/code-scanning/alerts/{{ record.alert_number }}
  - required fields: alert_number
  - risk: changes a code scanning alert's triage state, which can suppress a real security finding
- update_dependabot_alert:
  - endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/dependabot/alerts/{{ record.alert_number }}
  - required fields: alert_number
  - risk: changes a dependabot alert's triage state, which can suppress a real vulnerability finding
- create_deployment:
  - endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/deployments
  - risk: records a new deployment and may trigger CI/CD deployment automation
- create_fork:
  - endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/forks
  - risk: creates a new repository forked from this one, under the caller's account or a target organization
- create_repo_ruleset:
  - endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/rulesets
  - risk: creates a repository ruleset that can block pushes, merges, or deletions repo-wide once active
- update_repo_ruleset:
  - endpoint: PUT /repos/{{ config.owner }}/{{ config.repo }}/rulesets/{{ record.ruleset_id }}
  - required fields: ruleset_id
  - risk: changes an existing repository ruleset's enforcement or rule set, which can block pushes, merges, or deletions repo-wide
- delete_repo_ruleset:
  - endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/rulesets/{{ record.ruleset_id }}
  - required fields: ruleset_id
  - risk: removes a repository ruleset, lifting any push/merge/deletion restrictions it enforced
- update_secret_scanning_alert:
  - endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/secret-scanning/alerts/{{ record.alert_number }}
  - required fields: alert_number
  - risk: changes a secret scanning alert's triage state, which can suppress a real leaked-credential finding

## Security

- read risk: external API read
- write risk: external GitHub API mutation
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
