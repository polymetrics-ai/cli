# pm connectors inspect github

```text
NAME
  pm connectors inspect github - GitHub connector manual

SYNOPSIS
  pm connectors inspect github
  pm connectors inspect github --json
  pm credentials add <name> --connector github [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads GitHub repository, issue, pull request, code, release, collaboration, Actions, security (code scanning/dependabot/secret scanning/advisories), webhook, deploy key, environment, and ruleset data; tracks the full GitHub REST, GraphQL, and webhook parity ledger; and writes approved reverse ETL actions through the safety-gated connector path.

ICON
  asset: icons/github.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.github.com/en/rest/about-the-rest-api/breaking-changes

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  app_id
  auth_type
  base_url
  installation_id
  installation_permissions
  installation_repositories
  installation_repository_ids
  owner
  public_access
  repo
  since
  private_key (secret)
  private_key_base64 (secret)
  token (secret)

ETL STREAMS
  repository:
    primary key: node_id
    cursor: updated_at
    fields: created_at(), default_branch(), description(), forks_count(), full_name(), html_url(), id(), language(), name(), node_id(), open_issues_count(), private(), pushed_at(), repository(), stargazers_count(), updated_at(), watchers_count()
  issues:
    primary key: node_id
    cursor: updated_at
    fields: author_association(), body(), closed_at(), comments(), created_at(), html_url(), id(), locked(), node_id(), number(), repository(), state(), state_reason(), title(), updated_at(), url(), user_id(), user_login()
  pull_requests:
    primary key: node_id
    cursor: updated_at
    fields: author_association(), base_ref(), base_sha(), body(), closed_at(), comments(), created_at(), draft(), head_ref(), head_sha(), html_url(), id(), locked(), merge_commit_sha(), merged_at(), node_id(), number(), repository(), state(), title(), updated_at(), url(), user_id(), user_login()
  branches:
    primary key: name
    fields: commit_sha(), commit_url(), name(), protected(), repository()
  commits:
    primary key: sha
    cursor: commit_committer_date
    fields: author_id(), author_login(), commit_author_date(), commit_author_email(), commit_author_name(), commit_committer_date(), commit_committer_email(), commit_committer_name(), commit_message(), committer_id(), committer_login(), html_url(), node_id(), repository(), sha(), url()
  tags:
    primary key: name
    fields: commit_sha(), commit_url(), name(), node_id(), repository(), tarball_url(), zipball_url()
  releases:
    primary key: id
    cursor: published_at
    fields: assets_count(), author_login(), body(), created_at(), draft(), html_url(), id(), name(), node_id(), prerelease(), published_at(), repository(), tag_name(), target_commitish()
  labels:
    primary key: name
    fields: color(), default(), description(), id(), name(), node_id(), repository(), url()
  milestones:
    primary key: number
    cursor: updated_at
    fields: closed_at(), closed_issues(), created_at(), creator_login(), description(), due_on(), id(), node_id(), number(), open_issues(), repository(), state(), title(), updated_at()
  issue_comments:
    primary key: id
    cursor: updated_at
    fields: author_association(), body(), created_at(), html_url(), id(), issue_url(), node_id(), repository(), updated_at(), user_id(), user_login()
  pull_request_review_comments:
    primary key: id
    cursor: updated_at
    fields: body(), commit_id(), created_at(), diff_hunk(), html_url(), id(), node_id(), original_commit_id(), original_position(), path(), position(), pull_request_review_id(), pull_request_url(), repository(), updated_at(), user_login()
  collaborators:
    primary key: id
    fields: contributions(), html_url(), id(), login(), node_id(), relation(), repository(), role_name(), site_admin(), type()
  contributors:
    primary key: id
    fields: contributions(), html_url(), id(), login(), node_id(), relation(), repository(), role_name(), site_admin(), type()
  stargazers:
    primary key: id
    fields: contributions(), html_url(), id(), login(), node_id(), relation(), repository(), role_name(), site_admin(), type()
  subscribers:
    primary key: id
    fields: contributions(), html_url(), id(), login(), node_id(), relation(), repository(), role_name(), site_admin(), type()
  workflows:
    primary key: id
    cursor: updated_at
    fields: badge_url(), created_at(), html_url(), id(), name(), node_id(), path(), repository(), state(), updated_at()
  workflow_runs:
    primary key: id
    cursor: updated_at
    fields: conclusion(), created_at(), event(), head_branch(), head_sha(), html_url(), id(), name(), node_id(), repository(), run_attempt(), run_number(), status(), updated_at(), workflow_id()
  workflow_artifacts:
    primary key: id
    cursor: updated_at
    fields: archive_download_url(), created_at(), expired(), expires_at(), id(), name(), node_id(), repository(), size_in_bytes(), updated_at(), url(), workflow_run_id()
  deployments:
    primary key: id
    cursor: updated_at
    fields: created_at(), creator_login(), description(), environment(), id(), node_id(), ref(), repository(), sha(), task(), updated_at()
  commit_comments:
    primary key: id
    cursor: updated_at
    fields: author_association(), body(), commit_id(), created_at(), html_url(), id(), line(), node_id(), path(), position(), repository(), updated_at(), url(), user_id(), user_login()
  deploy_keys:
    primary key: id
    fields: added_by(), created_at(), enabled(), id(), key(), last_used(), read_only(), repository(), title(), url(), verified()
  webhooks:
    primary key: id
    cursor: updated_at
    fields: active(), config_url(), created_at(), deliveries_url(), events(), id(), name(), ping_url(), repository(), test_url(), type(), updated_at(), url()
  environments:
    primary key: id
    cursor: updated_at
    fields: created_at(), html_url(), id(), name(), node_id(), repository(), updated_at(), url()
  forks:
    primary key: id
    cursor: updated_at
    fields: created_at(), default_branch(), forks_count(), full_name(), html_url(), id(), name(), node_id(), open_issues_count(), owner_login(), private(), pushed_at(), repository(), stargazers_count(), updated_at(), watchers_count()
  invitations:
    primary key: id
    cursor: created_at
    fields: created_at(), expired(), html_url(), id(), invitee_login(), inviter_login(), node_id(), permissions(), repository(), url()
  issue_events:
    primary key: id
    cursor: created_at
    fields: actor_login(), commit_id(), commit_url(), created_at(), event(), id(), lock_reason(), node_id(), repository(), url()
  code_scanning_alerts:
    primary key: number
    cursor: updated_at
    fields: created_at(), dismissed_at(), dismissed_by_login(), dismissed_comment(), dismissed_reason(), fixed_at(), html_url(), number(), repository(), rule_id(), rule_severity(), state(), tool_name(), updated_at(), url()
  dependabot_alerts:
    primary key: number
    cursor: updated_at
    fields: auto_dismissed_at(), created_at(), dismissed_at(), dismissed_by_login(), dismissed_comment(), dismissed_reason(), fixed_at(), html_url(), number(), package_ecosystem(), package_name(), repository(), state(), updated_at(), url()
  secret_scanning_alerts:
    primary key: number
    cursor: updated_at
    fields: created_at(), html_url(), number(), push_protection_bypassed(), repository(), resolution(), resolved_at(), resolved_by_login(), secret_type(), secret_type_display_name(), state(), updated_at(), url(), validity()
  security_advisories:
    primary key: ghsa_id
    cursor: updated_at
    fields: author_login(), closed_at(), created_at(), cve_id(), ghsa_id(), html_url(), published_at(), publisher_login(), repository(), severity(), state(), summary(), updated_at(), url(), withdrawn_at()
  repo_rulesets:
    primary key: id
    cursor: updated_at
    fields: created_at(), enforcement(), id(), name(), repository(), source(), source_type(), target(), updated_at()
  autolinks:
    primary key: id
    fields: id(), is_alphanumeric(), key_prefix(), repository(), updated_at(), url_template()
  languages:
    primary key: repository
    fields: repository()
  projects:
    primary key: id
    fields: closed(), id(), number(), owner(), repository(), title(), updated_at(), url()
  project_items:
    primary key: id
    fields: content_id(), content_number(), content_state(), content_title(), content_type(), content_url(), created_at(), id(), project_id(), project_number(), project_title(), repository(), type(), updated_at()
  discussions:
    primary key: id
    fields: answer_chosen_at(), author_login(), category_id(), category_name(), category_slug(), created_at(), id(), is_answered(), number(), repository(), title(), updated_at(), url()
  discussion:
    primary key: id
    fields: answer_chosen_at(), author_login(), body(), category_id(), category_name(), category_slug(), comments(), created_at(), id(), is_answered(), number(), repository(), title(), updated_at(), url()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_issue:
    endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/issues
    required fields: title
    risk: creates user-visible GitHub issue and may notify watchers
  update_issue:
    endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number }}
    required fields: issue_number
    risk: mutates existing GitHub issue or pull request issue metadata
  comment_issue:
    endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number }}/comments
    required fields: issue_number, body
    risk: creates user-visible comment and may notify participants
  close_issue:
    endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number }}
    required fields: issue_number
    risk: closes existing GitHub issue
  create_pull_request:
    endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/pulls
    required fields: head, base
    risk: creates user-visible pull request and may notify watchers/reviewers
  update_pull_request:
    endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/pulls/{{ record.pull_number }}
    required fields: pull_number
    risk: mutates existing GitHub pull request
  close_pull_request:
    endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/pulls/{{ record.pull_number }}
    required fields: pull_number
    risk: closes existing GitHub pull request
  request_reviewers:
    endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/pulls/{{ record.pull_number }}/requested_reviewers
    required fields: pull_number
    risk: notifies requested GitHub reviewers
  merge_pull_request:
    endpoint: PUT /repos/{{ config.owner }}/{{ config.repo }}/pulls/{{ record.pull_number }}/merge
    required fields: pull_number
    risk: irreversibly changes repository history unless branch protection blocks merge
  create_label:
    endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/labels
    required fields: name, color
    risk: changes repository taxonomy used by issues and pull requests
  update_label:
    endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/labels/{{ record.name }}
    required fields: name
    optional fields: new_name, color, description
    risk: renames or changes labels already used by issues and pull requests
  delete_label:
    endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/labels/{{ record.name }}
    required fields: name
    risk: removes a label from the repository and existing issue metadata
  create_milestone:
    endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/milestones
    required fields: title
    risk: creates planning metadata visible to repository collaborators
  update_milestone:
    endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/milestones/{{ record.milestone_number }}
    required fields: milestone_number
    risk: changes planning metadata used by issues and pull requests
  delete_milestone:
    endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/milestones/{{ record.milestone_number }}
    required fields: milestone_number
    risk: removes repository planning metadata from GitHub
  create_release:
    endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/releases
    required fields: tag_name
    risk: publishes release metadata and may notify repository watchers
  update_release:
    endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/releases/{{ record.release_id }}
    required fields: release_id
    risk: changes published release metadata
  delete_release:
    endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/releases/{{ record.release_id }}
    required fields: release_id
    risk: removes release metadata from GitHub; tags are not deleted by this action
  dispatch_workflow:
    endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/actions/workflows/{{ record.workflow_id }}/dispatches
    required fields: workflow_id, ref
    risk: starts CI/CD automation that may deploy, publish, or mutate external systems
  rerun_workflow_run:
    endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/actions/runs/{{ record.run_id }}/rerun
    required fields: run_id
    risk: reruns CI/CD automation and consumes workflow minutes
  cancel_workflow_run:
    endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/actions/runs/{{ record.run_id }}/cancel
    required fields: run_id
    risk: interrupts in-flight CI/CD automation
  delete_workflow_run:
    endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/actions/runs/{{ record.run_id }}
    required fields: run_id
    risk: removes workflow run history from GitHub
  create_pull_request_review:
    endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/pulls/{{ record.pull_number }}/reviews
    required fields: pull_number
    risk: submits reviewer feedback and may approve or request changes on a pull request
  create_webhook:
    endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/hooks
    required fields: config
    risk: registers an outbound webhook that will receive repository event payloads
  update_webhook:
    endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/hooks/{{ record.hook_id }}
    required fields: hook_id
    risk: changes an existing webhook's target URL, secret, or event subscriptions
  delete_webhook:
    endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/hooks/{{ record.hook_id }}
    required fields: hook_id
    risk: removes a webhook; the target will stop receiving repository event payloads
  create_deploy_key:
    endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/keys
    required fields: key
    risk: grants a new SSH public key deploy access to the repository
  delete_deploy_key:
    endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/keys/{{ record.key_id }}
    required fields: key_id
    risk: revokes an SSH deploy key's access to the repository
  create_or_update_environment:
    endpoint: PUT /repos/{{ config.owner }}/{{ config.repo }}/environments/{{ record.environment_name }}
    required fields: environment_name
    risk: creates or changes a deployment environment's protection rules and reviewers
  delete_environment:
    endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/environments/{{ record.environment_name }}
    required fields: environment_name
    risk: removes a deployment environment and its protection rules
  create_commit_comment:
    endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/commits/{{ record.commit_sha }}/comments
    required fields: commit_sha, body
    risk: creates a user-visible comment attached to a specific commit
  update_commit_comment:
    endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/comments/{{ record.comment_id }}
    required fields: comment_id, body
    risk: changes the text of an existing commit comment
  delete_commit_comment:
    endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/comments/{{ record.comment_id }}
    required fields: comment_id
    risk: removes a commit comment
  update_issue_comment:
    endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/issues/comments/{{ record.comment_id }}
    required fields: comment_id, body
    risk: changes the text of an existing issue or pull request comment
  delete_issue_comment:
    endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/issues/comments/{{ record.comment_id }}
    required fields: comment_id
    risk: removes an issue or pull request comment
  lock_issue:
    endpoint: PUT /repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number }}/lock
    required fields: issue_number
    optional fields: lock_reason
    risk: prevents further comments from non-collaborators on an issue or pull request
  unlock_issue:
    endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number }}/lock
    required fields: issue_number
    risk: reopens an issue or pull request to comments from non-collaborators
  set_issue_labels:
    endpoint: PUT /repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number }}/labels
    required fields: issue_number
    optional fields: labels
    risk: replaces every label on an issue or pull request, removing any not listed
  add_issue_labels:
    endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number }}/labels
    required fields: issue_number, labels
    risk: adds labels to an issue or pull request without removing existing ones
  remove_issue_label:
    endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number }}/labels/{{ record.name }}
    required fields: issue_number, name
    risk: removes a single label from an issue or pull request
  add_issue_assignees:
    endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number }}/assignees
    required fields: issue_number, assignees
    risk: assigns additional GitHub users to an issue or pull request and may notify them
  remove_issue_assignees:
    endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/issues/{{ record.issue_number }}/assignees
    required fields: issue_number, assignees
    risk: removes assignees from an issue or pull request
  create_review_comment:
    endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/pulls/{{ record.pull_number }}/comments
    required fields: pull_number, body, commit_id, path
    risk: creates a user-visible inline review comment on a pull request diff
  update_review_comment:
    endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/pulls/comments/{{ record.comment_id }}
    required fields: comment_id, body
    risk: changes the text of an existing pull request review comment
  delete_review_comment:
    endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/pulls/comments/{{ record.comment_id }}
    required fields: comment_id
    risk: removes a pull request review comment
  submit_pull_request_review:
    endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/pulls/{{ record.pull_number }}/reviews/{{ record.review_id }}/events
    required fields: pull_number, review_id, event
    optional fields: body
    risk: submits a pending pull request review, which may approve or request changes
  dismiss_pull_request_review:
    endpoint: PUT /repos/{{ config.owner }}/{{ config.repo }}/pulls/{{ record.pull_number }}/reviews/{{ record.review_id }}/dismissals
    required fields: pull_number, review_id, message
    optional fields: event
    risk: dismisses an existing pull request review, clearing its approval status
  update_pull_request_branch:
    endpoint: PUT /repos/{{ config.owner }}/{{ config.repo }}/pulls/{{ record.pull_number }}/update-branch
    required fields: pull_number
    optional fields: expected_head_sha
    risk: merges the base branch into the pull request's head branch, adding a merge commit
  update_release_asset:
    endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/releases/assets/{{ record.asset_id }}
    required fields: asset_id
    risk: changes a release asset's file name or label
  delete_release_asset:
    endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/releases/assets/{{ record.asset_id }}
    required fields: asset_id
    risk: removes a downloadable asset from a published release
  replace_repo_topics:
    endpoint: PUT /repos/{{ config.owner }}/{{ config.repo }}/topics
    required fields: names
    risk: replaces the repository's entire topic list, removing any topic not listed
  add_collaborator:
    endpoint: PUT /repos/{{ config.owner }}/{{ config.repo }}/collaborators/{{ record.username }}
    required fields: username
    optional fields: permission
    risk: grants a GitHub user access to the repository and may send an invitation email
  remove_collaborator:
    endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/collaborators/{{ record.username }}
    required fields: username
    risk: revokes a collaborator's access to the repository
  create_ref:
    endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/git/refs
    required fields: ref, sha
    risk: creates a new branch or tag ref pointing at the given commit SHA
  merge_branch:
    endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/merges
    required fields: base, head
    risk: creates a merge commit combining the head ref into the base branch
  update_code_scanning_alert:
    endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/code-scanning/alerts/{{ record.alert_number }}
    required fields: alert_number, state
    risk: changes a code scanning alert's triage state, which can suppress a real security finding
  update_dependabot_alert:
    endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/dependabot/alerts/{{ record.alert_number }}
    required fields: alert_number, state
    risk: changes a dependabot alert's triage state, which can suppress a real vulnerability finding
  create_deployment:
    endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/deployments
    required fields: ref
    risk: records a new deployment and may trigger CI/CD deployment automation
  create_fork:
    endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/forks
    risk: creates a new repository forked from this one, under the caller's account or a target organization
  create_repo_ruleset:
    endpoint: POST /repos/{{ config.owner }}/{{ config.repo }}/rulesets
    required fields: name, enforcement
    risk: creates a repository ruleset that can block pushes, merges, or deletions repo-wide once active
  update_repo_ruleset:
    endpoint: PUT /repos/{{ config.owner }}/{{ config.repo }}/rulesets/{{ record.ruleset_id }}
    required fields: ruleset_id
    risk: changes an existing repository ruleset's enforcement or rule set, which can block pushes, merges, or deletions repo-wide
  delete_repo_ruleset:
    endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}/rulesets/{{ record.ruleset_id }}
    required fields: ruleset_id
    risk: removes a repository ruleset, lifting any push/merge/deletion restrictions it enforced
  update_secret_scanning_alert:
    endpoint: PATCH /repos/{{ config.owner }}/{{ config.repo }}/secret-scanning/alerts/{{ record.alert_number }}
    required fields: alert_number, state
    risk: changes a secret scanning alert's triage state, which can suppress a real leaked-credential finding
  delete_repo:
    endpoint: DELETE /repos/{{ config.owner }}/{{ config.repo }}
    risk: critical

SECURITY
  read risk: external API read
  write risk: external GitHub API mutation
  approval: reverse ETL plan approval required before writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Work with GitHub repositories from the command line.
  Usage: pm github <command> <subcommand> [flags]
  Source CLI: gh (https://cli.github.com/manual/gh_help_reference)
  Global flags:
    --json (boolean): Write machine-readable JSON output.
    --connection (string): Use a saved GitHub connector credential and repository scope.: maps_to=connection
  Core Commands
    issue list - List issues [intent=etl availability=implemented stream=issues]; flags: --state
    issue view - View issue details [intent=etl availability=partial stream=issues]; notes: The connector exposes issue records as an ETL stream; single issue lookup and gh-style field selection are planned for a later direct-read slice.
    issue create - Create an issue [intent=reverse_etl availability=implemented write=create_issue]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a visible issue in the configured repository.; flags: --title, --body, --label, --assignee, --milestone
    issue edit - Edit an issue [intent=reverse_etl availability=implemented write=update_issue]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Mutates title, body, labels, assignees, milestone, or state on an existing issue.; flags: --issue-number, --title, --body, --state, --state-reason, --label, --assignee, --milestone, --type
    issue close - Close an issue [intent=reverse_etl availability=implemented write=close_issue]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Closes an existing issue.; flags: --issue-number, --comment, --state-reason
    issue reopen - Reopen an issue [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: Reopens a previously closed issue.; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --issue-number
    issue comment - Comment on an issue [intent=reverse_etl availability=implemented write=comment_issue]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Adds a visible issue comment.; flags: --issue-number, --body
    issue lock - Lock issue conversation [intent=reverse_etl availability=implemented write=lock_issue]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Locks issue conversation for repository users.; flags: --issue-number, --lock-reason
    issue unlock - Unlock issue conversation [intent=reverse_etl availability=implemented write=unlock_issue]; approval: reverse ETL writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation.; risk: Unlocks issue conversation.; flags: --issue-number
    issue delete - Delete an issue [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; notes: In scope under the captain policy, but not implemented by this gh-style command path yet; the bounded operation must use typed schemas, plan -> preview -> explicit approval -> execute, and typed `destructive` confirmation when destructive.
    issue develop - Manage development branches for an issue [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Depends on local git branch workflow and checkout state.
    issue status - Show relevant issues [intent=direct_read availability=planned]; notes: Requires viewer-centric queries that are not modeled by the repository-scoped stream set yet.
    issue pin - Pin an issue [intent=direct_write availability=unsupported_api]; notes: GitHub issue pinning is not modeled in the current REST-backed connector surface.
    issue unpin - Unpin an issue [intent=direct_write availability=unsupported_api]; notes: GitHub issue unpinning is not modeled in the current REST-backed connector surface.
    issue transfer - Transfer an issue [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; notes: In scope under the captain policy, but not implemented by this gh-style command path yet; the bounded operation must use typed schemas, plan -> preview -> explicit approval -> execute, and typed `destructive` confirmation when destructive.
    pr list - List pull requests [intent=etl availability=implemented stream=pull_requests]
    pr view - View pull request details [intent=etl availability=partial stream=pull_requests]; notes: The connector exposes PR records as an ETL stream; single PR lookup and gh-style field selection are planned for a direct-read slice.
    pr create - Create a pull request [intent=reverse_etl availability=implemented write=create_pull_request]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a visible pull request and may request reviewers.; flags: --head, --base, --title, --body, --issue, --draft, --maintainer-can-modify, --label, --assignee, --milestone, --reviewer, --team-reviewer
    pr edit - Edit a pull request [intent=reverse_etl availability=implemented write=update_pull_request]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Mutates an existing pull request.; flags: --pull-number, --title, --body, --state, --base, --maintainer-can-modify, --label, --assignee, --milestone, --reviewer, --team-reviewer
    pr close - Close a pull request [intent=reverse_etl availability=implemented write=close_pull_request]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Closes an existing pull request.; flags: --pull-number, --comment
    pr reopen - Reopen a pull request [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: Reopens a previously closed pull request.; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --pull-number
    pr comment - Comment on a pull request [intent=reverse_etl availability=implemented write=comment_issue]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Comments on a pull request (PRs are issues in GitHub's data model).; flags: --pull-number, --body
    pr merge - Merge a pull request [intent=reverse_etl availability=implemented write=merge_pull_request]; approval: reverse ETL writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation.; risk: Merges code into the pull request base branch.; flags: --pull-number, --commit-title, --commit-message, --sha, --merge-method
    pr review - Add a pull request review [intent=reverse_etl availability=implemented write=create_pull_request_review]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Adds a visible pull request review.; flags: --pull-number, --body, --commit-id, --event
    pr checks - Show pull request checks [intent=direct_read availability=planned]; notes: Requires check-run/status aggregation not modeled by the current repository streams.
    pr diff - Show pull request diff [intent=direct_read availability=unsupported_api]; notes: Diff output is a patch/binary-like representation rather than a JSON ETL stream.
    pr checkout - Check out a pull request locally [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Depends on local git checkout state.
    pr ready - Mark a draft pull request ready [intent=direct_write availability=unsupported_api]; notes: Not modeled by the current REST write actions.
    pr update-branch - Update a pull request branch [intent=reverse_etl availability=implemented write=update_pull_request_branch]; approval: reverse ETL writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation.; risk: Updates a pull request branch from its base branch.; flags: --pull-number, --expected-head-sha
    pr status - Show relevant pull requests [intent=direct_read availability=planned]; notes: Requires viewer-centric and branch-aware filtering not modeled by repository streams yet.
    pr lock - Lock pull request conversation [intent=reverse_etl availability=implemented write=lock_issue]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Locks a pull request's conversation.; flags: --pull-number, --lock-reason
    pr unlock - Unlock pull request conversation [intent=reverse_etl availability=implemented write=unlock_issue]; approval: reverse ETL writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation.; risk: Unlocks a pull request's conversation.; flags: --pull-number
    pr revert - Revert a pull request [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; notes: In scope under the captain policy, but not implemented by this gh-style command path yet; the bounded operation must use typed schemas, plan -> preview -> explicit approval -> execute, and typed `destructive` confirmation when destructive.
    repo view - View repository metadata [intent=etl availability=implemented stream=repository]
    repo list - List repositories for an owner [intent=direct_read availability=unsupported_api]; notes: The current connector is scoped to one configured repository.
    repo create - Create a repository [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; notes: In scope under the captain policy, but not implemented by this gh-style command path yet; the bounded operation must use typed schemas, plan -> preview -> explicit approval -> execute, and typed `destructive` confirmation when destructive.
    repo delete - Delete a repository [intent=reverse_etl availability=implemented write=delete_repo]; approval: reverse ETL writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation.; risk: Permanently deletes the configured repository.; notes: Executes the typed repository deletion write through the reverse ETL safety path; use `--confirm destructive` only after approving the preview.
    repo archive - Archive a repository [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; notes: In scope under the captain policy, but not implemented by this gh-style command path yet; the bounded operation must use typed schemas, plan -> preview -> explicit approval -> execute, and typed `destructive` confirmation when destructive.
    repo unarchive - Unarchive a repository [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; notes: In scope under the captain policy, but not implemented by this gh-style command path yet; the bounded operation must use typed schemas, plan -> preview -> explicit approval -> execute, and typed `destructive` confirmation when destructive.
    repo fork - Fork a repository [intent=reverse_etl availability=implemented write=create_fork]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a fork of the configured repository.; flags: --organization, --name, --default-branch-only
    repo clone - Clone a repository locally [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Depends on local git and filesystem state.
    repo sync - Sync a local repository [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Depends on local git state.
    repo set-default - Set the default local repository [intent=config availability=unsupported_local unsupported local workflow]; notes: Local gh configuration is outside connector metadata.
    repo read-file - Read repository file metadata [intent=direct_read availability=implemented]; notes: Executes the fixed GitHub repository contents read endpoint with file content and raw download URLs redacted.; flags: --path, --ref
    repo read-dir - Read repository directory contents [intent=direct_read availability=implemented]; notes: Executes the fixed GitHub repository contents read endpoint for directory listings; file responses are rejected.; flags: --path, --ref
    repo autolink list - List repository autolinks [intent=etl availability=implemented stream=autolinks]
    repo autolink create - Create a repository autolink [intent=direct_write availability=unsupported_api]; notes: Autolink writes are not modeled by the current write set.
    repo autolink delete - Delete a repository autolink [intent=direct_write availability=unsupported_api]; notes: Autolink writes are not modeled by the current write set.
    repo deploy-key list - List deploy keys [intent=etl availability=implemented stream=deploy_keys]
    repo deploy-key add - Add a deploy key [intent=reverse_etl availability=implemented write=create_deploy_key]; approval: reverse ETL writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation.; risk: Adds a deploy key to the repository.; flags: --title, --key, --read-only
    repo deploy-key delete - Delete a deploy key [intent=reverse_etl availability=implemented write=delete_deploy_key]; approval: reverse ETL writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation.; risk: Deletes a deploy key from the repository.; flags: --key-id
    repo license list - List license templates [intent=direct_read availability=unsupported_api]; notes: Global license template APIs are not repository-scoped connector streams.
    repo gitignore list - List gitignore templates [intent=direct_read availability=unsupported_api]; notes: Global gitignore template APIs are not repository-scoped connector streams.
    repo ruleset create - Create a repository ruleset [intent=reverse_etl availability=implemented write=create_repo_ruleset]; approval: reverse ETL writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation.; risk: Creates repository rules that can affect contribution workflows.; notes: Connector-native write action; the current gh ruleset surface documents check, list, and view, but not create.; flags: --name, --target, --enforcement
    repo ruleset update - Update a repository ruleset [intent=reverse_etl availability=implemented write=update_repo_ruleset]; approval: reverse ETL writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation.; risk: Updates repository rules that can affect contribution workflows.; notes: Connector-native write action; the current gh ruleset surface documents check, list, and view, but not update.; flags: --ruleset-id, --name, --target, --enforcement
    repo ruleset delete - Delete a repository ruleset [intent=reverse_etl availability=implemented write=delete_repo_ruleset]; approval: reverse ETL writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation.; risk: Deletes repository rules that can affect contribution workflows.; notes: Connector-native write action; the current gh ruleset surface documents check, list, and view, but not delete.; flags: --ruleset-id
    repo update - PATCH /repos/{owner}/{repo} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    repo sbom view - Download /repos/{owner}/{repo}/dependency-graph/sbom [intent=direct_read availability=planned]; notes: Binary direct downloads remain planned until the command runner supports bounded binary output policies.
    repo sbom fetch - Download /repos/{owner}/{repo}/dependency-graph/sbom/fetch-report/{sbom_uuid} [intent=direct_read availability=planned]; notes: Binary direct downloads remain planned until the command runner supports bounded binary output policies.; flags: --sbom-uuid
    repo sbom generate - Download /repos/{owner}/{repo}/dependency-graph/sbom/generate-report [intent=direct_read availability=planned]; notes: Binary direct downloads remain planned until the command runner supports bounded binary output policies.
    repo archive tarball - Download /repos/{owner}/{repo}/tarball/{ref} [intent=direct_read availability=planned]; notes: Binary direct downloads remain planned until the command runner supports bounded binary output policies.; flags: --ref
    repo archive zipball - Download /repos/{owner}/{repo}/zipball/{ref} [intent=direct_read availability=planned]; notes: Binary direct downloads remain planned until the command runner supports bounded binary output policies.; flags: --ref
    release list - List releases [intent=etl availability=implemented stream=releases]
    release view - View a release [intent=etl availability=partial stream=releases]; notes: Release records are available as an ETL stream; single-release lookup is planned for direct reads.
    release create - Create a release [intent=reverse_etl availability=implemented write=create_release]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a visible release.; flags: --tag-name, --target-commitish, --name, --body, --draft, --prerelease, --generate-release-notes, --make-latest
    release edit - Edit a release [intent=reverse_etl availability=implemented write=update_release]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Mutates an existing release.; flags: --release-id, --tag-name, --target-commitish, --name, --body, --draft, --prerelease, --generate-release-notes, --make-latest
    release delete - Delete a release [intent=reverse_etl availability=implemented write=delete_release]; approval: reverse ETL writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation.; risk: Deletes an existing release.; flags: --release-id
    release upload - Upload release assets [intent=direct_write availability=unsupported_local unsupported local workflow]; notes: Depends on local files and binary upload semantics.
    release download - Download release assets [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Downloads binary assets to the local filesystem.
    release delete-asset - Delete a release asset [intent=reverse_etl availability=implemented write=delete_release_asset]; approval: reverse ETL writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation.; risk: Deletes an existing release asset.; flags: --asset-id
    release verify - Verify release assets [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Depends on local artifact files and signature verification behavior.
  GitHub Actions Commands
    workflow list - List workflows [intent=etl availability=implemented stream=workflows]
    workflow view - View workflow details [intent=etl availability=partial stream=workflows]; notes: Workflow records are available as a stream; single workflow lookup is planned for direct reads.
    workflow run - Dispatch a workflow [intent=reverse_etl availability=implemented write=dispatch_workflow]; approval: reverse ETL writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation.; risk: Triggers a GitHub Actions workflow run.; flags: --workflow-id, --ref
    workflow enable - Enable a workflow [intent=direct_write availability=unsupported_api]; notes: Workflow enablement is not modeled by the current write set.
    workflow disable - Disable a workflow [intent=direct_write availability=unsupported_api]; notes: Workflow disablement is not modeled by the current write set.
    run list - List workflow runs [intent=etl availability=implemented stream=workflow_runs]
    run view - View workflow run details [intent=etl availability=partial stream=workflow_runs]; notes: Workflow run records are available as a stream; log and job expansion are not modeled yet.
    run rerun - Rerun a workflow run [intent=reverse_etl availability=implemented write=rerun_workflow_run]; approval: reverse ETL writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation.; risk: Reruns a GitHub Actions workflow run.; flags: --run-id
    run cancel - Cancel a workflow run [intent=reverse_etl availability=implemented write=cancel_workflow_run]; approval: reverse ETL writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation.; risk: Cancels a GitHub Actions workflow run.; flags: --run-id
    run delete - Delete a workflow run [intent=reverse_etl availability=implemented write=delete_workflow_run]; approval: reverse ETL writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation.; risk: Deletes workflow run data.; flags: --run-id
    run download - Download workflow artifacts [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Downloads artifacts to the local filesystem; artifact metadata is available through the workflow_artifacts stream.
    run watch - Watch a workflow run [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Interactive watch behavior is outside connector command metadata.
    run logs view - Download /repos/{owner}/{repo}/actions/jobs/{job_id}/logs [intent=direct_read availability=planned]; notes: Binary direct downloads remain planned until the command runner supports bounded binary output policies.; flags: --job-id
    run logs view-2 - Download /repos/{owner}/{repo}/actions/runs/{run_id}/attempts/{attempt_number}/logs [intent=direct_read availability=planned]; notes: Binary direct downloads remain planned until the command runner supports bounded binary output policies.; flags: --run-id, --attempt-number
    run logs view-3 - Download /repos/{owner}/{repo}/actions/runs/{run_id}/logs [intent=direct_read availability=planned]; notes: Binary direct downloads remain planned until the command runner supports bounded binary output policies.; flags: --run-id
    cache list - List GitHub Actions caches [intent=direct_read availability=unsupported_api]; notes: Actions cache endpoints are tracked but excluded from the current repository connector surface.
    cache delete - Delete GitHub Actions caches [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; notes: In scope under the captain policy, but not implemented by this gh-style command path yet; the bounded operation must use typed schemas, plan -> preview -> explicit approval -> execute, and typed `destructive` confirmation when destructive.
  Collaboration Commands
    label list - List labels [intent=etl availability=implemented stream=labels]
    label create - Create a label [intent=reverse_etl availability=implemented write=create_label]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a repository label.; flags: --name, --color, --description
    label edit - Edit a label [intent=reverse_etl availability=implemented write=update_label]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Mutates a repository label.; flags: --name, --new-name, --color, --description
    label delete - Delete a label [intent=reverse_etl availability=implemented write=delete_label]; approval: reverse ETL writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation.; risk: Deletes a repository label.; flags: --name
    label clone - Clone labels between repositories [intent=direct_write availability=unsupported_api]; notes: Cross-repository copy is outside the configured repository connector scope.
    ruleset list - List repository rulesets [intent=etl availability=implemented stream=repo_rulesets]
    ruleset view - View repository ruleset details [intent=etl availability=partial stream=repo_rulesets]; notes: Rulesets are available as a stream; single ruleset detail is planned for direct reads.
    ruleset check - Check rules that apply to a branch [intent=direct_read availability=planned]; notes: Requires a constrained direct-read operation with branch input.
    org list - List organizations for the authenticated user [intent=direct_read availability=unsupported_api]; notes: Viewer/org-scoped APIs are outside the configured repository connector.
    project list - List projects [intent=etl availability=implemented stream=projects]; notes: Lists Projects v2 for the configured repository owner using a fixed GraphQL query. Private projects require read:project scope.
    project create - Create a project [intent=direct_write availability=planned]; notes: Requires fixed GraphQL operations and explicit project policy.
    project item-list - List project items [intent=etl availability=implemented stream=project_items]; notes: Lists items for a Project v2 node ID using a fixed GraphQL query. Private projects require read:project scope.; flags: --project-id
    discussion list - List discussions [intent=etl availability=implemented stream=discussions]; notes: Lists repository discussions using a fixed GraphQL query. Private repositories require repo scope; public repositories require public_repo scope.
    discussion view - View a discussion [intent=etl availability=implemented stream=discussion]; notes: Views one repository discussion using a fixed GraphQL query. Private repositories require repo scope; public repositories require public_repo scope.; flags: --number
    discussion create - Create a discussion [intent=direct_write availability=planned]; notes: Requires fixed GraphQL mutations and approval policy.
  Security And Configuration Commands
    secret list - List repository secrets [intent=direct_read availability=unsupported_api]; notes: Secret metadata endpoints require elevated scopes and are explicitly excluded from the current connector surface.
    secret set - Create or update a secret [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; notes: In scope under the captain policy, but not implemented by this gh-style command path yet; the bounded operation must use typed schemas, plan -> preview -> explicit approval -> execute, and typed `destructive` confirmation when destructive.
    secret delete - Delete a secret [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; notes: In scope under the captain policy, but not implemented by this gh-style command path yet; the bounded operation must use typed schemas, plan -> preview -> explicit approval -> execute, and typed `destructive` confirmation when destructive.
    secret delete-2 - DELETE /repos/{owner}/{repo}/actions/secrets/{secret_name} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --secret-name
    secret set-2 - PUT /repos/{owner}/{repo}/actions/secrets/{secret_name} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --secret-name
    secret delete-3 - DELETE /repos/{owner}/{repo}/codespaces/secrets/{secret_name} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --secret-name
    secret set-3 - PUT /repos/{owner}/{repo}/codespaces/secrets/{secret_name} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --secret-name
    secret delete-4 - DELETE /repos/{owner}/{repo}/dependabot/secrets/{secret_name} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --secret-name
    secret set-4 - PUT /repos/{owner}/{repo}/dependabot/secrets/{secret_name} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --secret-name
    secret delete-5 - DELETE /repos/{owner}/{repo}/environments/{environment_name}/secrets/{secret_name} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --environment-name, --secret-name
    secret set-5 - PUT /repos/{owner}/{repo}/environments/{environment_name}/secrets/{secret_name} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --environment-name, --secret-name
    variable list - List repository variables [intent=direct_read availability=unsupported_api]; notes: Actions variable endpoints require elevated scopes and are excluded from the current connector surface.
    variable get - Get a repository variable [intent=direct_read availability=unsupported_api]; notes: Actions variable endpoints require elevated scopes and are excluded from the current connector surface.
    variable set - Create or update a repository variable [intent=direct_write availability=unsupported_api]; notes: Actions variable writes are not modeled by the current write set.
    variable delete - Delete a repository variable [intent=direct_write availability=unsupported_api]; notes: Actions variable deletion is not modeled by the current write set.
    variable create - POST /repos/{owner}/{repo}/actions/variables [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    variable delete-2 - DELETE /repos/{owner}/{repo}/actions/variables/{name} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --name
    variable update - PATCH /repos/{owner}/{repo}/actions/variables/{name} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --name
    variable create-2 - POST /repos/{owner}/{repo}/environments/{environment_name}/variables [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --environment-name
    variable delete-3 - DELETE /repos/{owner}/{repo}/environments/{environment_name}/variables/{name} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --environment-name, --name
    variable update-2 - PATCH /repos/{owner}/{repo}/environments/{environment_name}/variables/{name} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --environment-name, --name
    gpg-key list - List GPG keys [intent=direct_read availability=unsupported_api]; notes: Account key APIs are outside repository-scoped connector metadata.
    ssh-key list - List SSH keys [intent=direct_read availability=unsupported_api]; notes: Account key APIs are outside repository-scoped connector metadata.
    attestation verify - Verify artifact attestations [intent=direct_read availability=unsupported_local unsupported local workflow]; notes: Attestation verification depends on local artifact files.
  Local Workflow Commands
    auth login - Authenticate gh [intent=auth availability=unsupported_local unsupported local workflow]; notes: Polymetrics uses its own credential vault and does not manage gh sessions.
    auth status - View gh authentication status [intent=auth availability=unsupported_local unsupported local workflow]; notes: Polymetrics credential inspection is separate from gh session management.
    auth token - Print gh token [intent=auth availability=unsafe_or_disallowed]; notes: Disallowed because commands must never print stored secret values; this is a secret disclosure escape hatch, not a destructive-operation exclusion.
    config get - Read gh local config [intent=config availability=unsupported_local unsupported local workflow]; notes: Local gh configuration is outside connector metadata.
    config set - Write gh local config [intent=config availability=unsupported_local unsupported local workflow]; notes: Local gh configuration is outside connector metadata.
    browse - Open GitHub in a browser [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Browser-opening behavior is local workflow, not connector API behavior.
    alias list - List gh aliases [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: gh alias management is local configuration.
    extension list - List gh extensions [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Extension installation and execution are local gh workflows.
    completion - Generate shell completion [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Completion generation belongs to pm CLI shell integration, not connector metadata.
  Additional Commands
    api - Make an authenticated GitHub API request [intent=raw_api availability=unsafe_or_disallowed]; notes: Unrestricted raw GitHub API escape hatches remain disallowed; bounded documented operations are represented individually in the connector ledgers instead.
    search repos - Search repositories [intent=direct_read availability=planned]; notes: Search APIs are useful direct-read candidates but not stream-backed in this connector yet.
    search issues - Search issues [intent=direct_read availability=planned]; notes: Search APIs are useful direct-read candidates but not stream-backed in this connector yet.
    search prs - Search pull requests [intent=direct_read availability=planned]; notes: Search APIs are useful direct-read candidates but not stream-backed in this connector yet.
    search code - Search code [intent=direct_read availability=planned]; notes: Search APIs are useful direct-read candidates but need pagination/rate-limit policy.
    search commits - Search commits [intent=direct_read availability=planned]; notes: Search APIs are useful direct-read candidates but not stream-backed in this connector yet.
    gist list - List gists [intent=direct_read availability=unsupported_api]; notes: Gists are account-scoped, not repository-scoped.
    gist create - Create a gist [intent=direct_write availability=unsupported_api]; notes: Gist writes are outside the repository connector scope.
    codespace list - List codespaces [intent=direct_read availability=unsupported_api]; notes: Codespaces endpoints are tracked but excluded from the current connector surface.
    codespace create - Create a codespace [intent=direct_write availability=unsupported_api]; notes: Codespaces creation is outside the connector write surface.
    codespace ssh - SSH into a codespace [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Requires local SSH and interactive terminal behavior.
    status - Print GitHub status [intent=direct_read availability=planned]; notes: Viewer-centric dashboard data is not modeled by the repository connector yet.
    copilot - Use GitHub Copilot CLI [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Copilot CLI behavior is not a GitHub connector API surface.
    copilot configuration view - Read /repos/{owner}/{repo}/copilot/cloud-agent/configuration [intent=direct_read availability=implemented]
    skill list - List GitHub Skills [intent=direct_read availability=unsupported_api]; notes: GitHub Skills are not modeled by the repository connector surface.
    agent-task list - List GitHub agent tasks [intent=direct_read availability=unsupported_api]; notes: Agent task APIs are not modeled by the repository connector surface.
  Other Commands
    artifact download - Download /repos/{owner}/{repo}/actions/artifacts/{artifact_id}/{archive_format} [intent=direct_read availability=planned]; notes: Binary direct downloads remain planned until the command runner supports bounded binary output policies.; flags: --artifact-id, --archive-format
    actions retention-limit view - Read /repos/{owner}/{repo}/actions/cache/retention-limit [intent=direct_read availability=implemented]
    actions retention-limit set - PUT /repos/{owner}/{repo}/actions/cache/retention-limit [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    actions storage-limit view - Read /repos/{owner}/{repo}/actions/cache/storage-limit [intent=direct_read availability=implemented]
    actions storage-limit set - PUT /repos/{owner}/{repo}/actions/cache/storage-limit [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    actions usage view - Read /repos/{owner}/{repo}/actions/cache/usage [intent=direct_read availability=implemented]
    actions caches delete - DELETE /repos/{owner}/{repo}/actions/caches [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    actions caches view - Read /repos/{owner}/{repo}/actions/caches [intent=direct_read availability=implemented]
    actions caches delete-2 - DELETE /repos/{owner}/{repo}/actions/caches/{cache_id} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --cache-id
    actions concurrency_groups view - Read /repos/{owner}/{repo}/actions/concurrency_groups [intent=direct_read availability=implemented]
    actions concurrency_groups view-2 - Read /repos/{owner}/{repo}/actions/concurrency_groups/{concurrency_group_name} [intent=direct_read availability=planned]; notes: Concurrency group names may contain refs or expressions with slashes/spaces; direct reads remain planned until arbitrary string path parameter encoding is supported.; flags: --concurrency-group-name
    actions jobs view - Read /repos/{owner}/{repo}/actions/jobs/{job_id} [intent=direct_read availability=implemented]; flags: --job-id
    actions rerun create - POST /repos/{owner}/{repo}/actions/jobs/{job_id}/rerun [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --job-id
    actions sub view - Read /repos/{owner}/{repo}/actions/oidc/customization/sub [intent=direct_read availability=implemented]
    actions sub set - PUT /repos/{owner}/{repo}/actions/oidc/customization/sub [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    actions organization-secrets view - Read /repos/{owner}/{repo}/actions/organization-secrets [intent=direct_read availability=implemented]
    actions organization-variables view - Read /repos/{owner}/{repo}/actions/organization-variables [intent=direct_read availability=implemented]
    actions permissions view - Read /repos/{owner}/{repo}/actions/permissions [intent=direct_read availability=implemented]
    actions permissions set - PUT /repos/{owner}/{repo}/actions/permissions [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    actions access view - Read /repos/{owner}/{repo}/actions/permissions/access [intent=direct_read availability=implemented]
    actions permissions set-2 - PUT /repos/{owner}/{repo}/actions/permissions/access [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    actions artifact-and-log-retention view - Read /repos/{owner}/{repo}/actions/permissions/artifact-and-log-retention [intent=direct_read availability=implemented]
    actions permissions set-3 - PUT /repos/{owner}/{repo}/actions/permissions/artifact-and-log-retention [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    actions fork-pr-contributor-approval view - Read /repos/{owner}/{repo}/actions/permissions/fork-pr-contributor-approval [intent=direct_read availability=implemented]
    actions permissions set-4 - PUT /repos/{owner}/{repo}/actions/permissions/fork-pr-contributor-approval [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    actions fork-pr-workflows-private-repos view - Read /repos/{owner}/{repo}/actions/permissions/fork-pr-workflows-private-repos [intent=direct_read availability=implemented]
    actions permissions set-5 - PUT /repos/{owner}/{repo}/actions/permissions/fork-pr-workflows-private-repos [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    actions selected-actions view - Read /repos/{owner}/{repo}/actions/permissions/selected-actions [intent=direct_read availability=implemented]
    actions permissions set-6 - PUT /repos/{owner}/{repo}/actions/permissions/selected-actions [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    actions workflow view - Read /repos/{owner}/{repo}/actions/permissions/workflow [intent=direct_read availability=implemented]
    actions permissions set-7 - PUT /repos/{owner}/{repo}/actions/permissions/workflow [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    actions runners view - Read /repos/{owner}/{repo}/actions/runners [intent=direct_read availability=implemented]
    actions downloads view - Read /repos/{owner}/{repo}/actions/runners/downloads [intent=direct_read availability=implemented]
    actions generate-jitconfig create - POST /repos/{owner}/{repo}/actions/runners/generate-jitconfig [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    actions registration-token create - POST /repos/{owner}/{repo}/actions/runners/registration-token [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    actions remove-token create - POST /repos/{owner}/{repo}/actions/runners/remove-token [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    actions runners delete - DELETE /repos/{owner}/{repo}/actions/runners/{runner_id} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --runner-id
    actions runners view-2 - Read /repos/{owner}/{repo}/actions/runners/{runner_id} [intent=direct_read availability=implemented]; flags: --runner-id
    actions labels delete - DELETE /repos/{owner}/{repo}/actions/runners/{runner_id}/labels [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --runner-id
    actions labels view - Read /repos/{owner}/{repo}/actions/runners/{runner_id}/labels [intent=direct_read availability=implemented]; flags: --runner-id
    actions labels create - POST /repos/{owner}/{repo}/actions/runners/{runner_id}/labels [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --runner-id
    actions labels set - PUT /repos/{owner}/{repo}/actions/runners/{runner_id}/labels [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --runner-id
    actions labels delete-2 - DELETE /repos/{owner}/{repo}/actions/runners/{runner_id}/labels/{name} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --runner-id, --name
    actions approvals view - Read /repos/{owner}/{repo}/actions/runs/{run_id}/approvals [intent=direct_read availability=implemented]; flags: --run-id
    actions approve create - POST /repos/{owner}/{repo}/actions/runs/{run_id}/approve [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --run-id
    actions attempts view - Read /repos/{owner}/{repo}/actions/runs/{run_id}/attempts/{attempt_number} [intent=direct_read availability=implemented]; flags: --run-id, --attempt-number
    actions jobs view-2 - Read /repos/{owner}/{repo}/actions/runs/{run_id}/attempts/{attempt_number}/jobs [intent=direct_read availability=implemented]; flags: --run-id, --attempt-number
    actions concurrency_groups view-3 - Read /repos/{owner}/{repo}/actions/runs/{run_id}/concurrency_groups [intent=direct_read availability=implemented]; flags: --run-id
    actions deployment_protection_rule create - POST /repos/{owner}/{repo}/actions/runs/{run_id}/deployment_protection_rule [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --run-id
    actions jobs view-3 - Read /repos/{owner}/{repo}/actions/runs/{run_id}/jobs [intent=direct_read availability=implemented]; flags: --run-id
    actions logs delete - DELETE /repos/{owner}/{repo}/actions/runs/{run_id}/logs [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: critical; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --run-id
    actions pending_deployments view - Read /repos/{owner}/{repo}/actions/runs/{run_id}/pending_deployments [intent=direct_read availability=implemented]; flags: --run-id
    actions pending_deployments create - POST /repos/{owner}/{repo}/actions/runs/{run_id}/pending_deployments [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --run-id
    actions timing view - Read /repos/{owner}/{repo}/actions/runs/{run_id}/timing [intent=direct_read availability=implemented]; flags: --run-id
    actions secrets view - Read /repos/{owner}/{repo}/actions/secrets [intent=direct_read availability=implemented]
    actions public-key view - Read /repos/{owner}/{repo}/actions/secrets/public-key [intent=direct_read availability=implemented]
    actions secrets view-2 - Read /repos/{owner}/{repo}/actions/secrets/{secret_name} [intent=direct_read availability=implemented]; flags: --secret-name
    actions variables view - Read /repos/{owner}/{repo}/actions/variables [intent=direct_read availability=implemented]
    actions variables view-2 - Read /repos/{owner}/{repo}/actions/variables/{name} [intent=direct_read availability=implemented]; flags: --name
    actions disable set - PUT /repos/{owner}/{repo}/actions/workflows/{workflow_id}/disable [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --workflow-id
    actions enable set - PUT /repos/{owner}/{repo}/actions/workflows/{workflow_id}/enable [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --workflow-id
    actions timing view-2 - Read /repos/{owner}/{repo}/actions/workflows/{workflow_id}/timing [intent=direct_read availability=implemented]; flags: --workflow-id
    assignees view - Read /repos/{owner}/{repo}/assignees/{assignee} [intent=direct_read availability=planned]; notes: GitHub documents a 204 no-content success response for this endpoint; direct reads remain planned until no-content direct-read handling is supported.; flags: --assignee
    attestations create - POST /repos/{owner}/{repo}/attestations [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    attestations view - Read /repos/{owner}/{repo}/attestations/{subject_digest} [intent=direct_read availability=planned]; notes: Subject digests require values such as sha256:<digest>; direct reads remain planned until digest path parameter encoding is supported.; flags: --subject-digest
    autolinks create - POST /repos/{owner}/{repo}/autolinks [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    autolinks delete - DELETE /repos/{owner}/{repo}/autolinks/{autolink_id} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --autolink-id
    automated-security-fixes delete - DELETE /repos/{owner}/{repo}/automated-security-fixes [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    automated-security-fixes view - Read /repos/{owner}/{repo}/automated-security-fixes [intent=direct_read availability=implemented]
    automated-security-fixes set - PUT /repos/{owner}/{repo}/automated-security-fixes [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    branches protection delete - DELETE /repos/{owner}/{repo}/branches/{branch}/protection [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    branches protection view - Read /repos/{owner}/{repo}/branches/{branch}/protection [intent=direct_read availability=planned]; notes: Branch names may contain slashes; direct reads remain planned until GitHub ref-like path parameter encoding is supported.; flags: --branch
    branches protection set - PUT /repos/{owner}/{repo}/branches/{branch}/protection [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    branches enforce_admins delete - DELETE /repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    branches enforce_admins view - Read /repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins [intent=direct_read availability=planned]; notes: Branch names may contain slashes; direct reads remain planned until GitHub ref-like path parameter encoding is supported.; flags: --branch
    branches enforce_admins create - POST /repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    branches required_pull_request_reviews delete - DELETE /repos/{owner}/{repo}/branches/{branch}/protection/required_pull_request_reviews [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    branches required_pull_request_reviews view - Read /repos/{owner}/{repo}/branches/{branch}/protection/required_pull_request_reviews [intent=direct_read availability=planned]; notes: Branch names may contain slashes; direct reads remain planned until GitHub ref-like path parameter encoding is supported.; flags: --branch
    branches required_pull_request_reviews update - PATCH /repos/{owner}/{repo}/branches/{branch}/protection/required_pull_request_reviews [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    branches required_signatures delete - DELETE /repos/{owner}/{repo}/branches/{branch}/protection/required_signatures [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    branches required_signatures view - Read /repos/{owner}/{repo}/branches/{branch}/protection/required_signatures [intent=direct_read availability=planned]; notes: Branch names may contain slashes; direct reads remain planned until GitHub ref-like path parameter encoding is supported.; flags: --branch
    branches required_signatures create - POST /repos/{owner}/{repo}/branches/{branch}/protection/required_signatures [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    branches required_status_checks delete - DELETE /repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    branches required_status_checks view - Read /repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks [intent=direct_read availability=planned]; notes: Branch names may contain slashes; direct reads remain planned until GitHub ref-like path parameter encoding is supported.; flags: --branch
    branches required_status_checks update - PATCH /repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    branches contexts delete - DELETE /repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    branches contexts view - Read /repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts [intent=direct_read availability=planned]; notes: Branch names may contain slashes; direct reads remain planned until GitHub ref-like path parameter encoding is supported.; flags: --branch
    branches contexts create - POST /repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    branches contexts set - PUT /repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    branches restrictions delete - DELETE /repos/{owner}/{repo}/branches/{branch}/protection/restrictions [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    branches restrictions view - Read /repos/{owner}/{repo}/branches/{branch}/protection/restrictions [intent=direct_read availability=planned]; notes: Branch names may contain slashes; direct reads remain planned until GitHub ref-like path parameter encoding is supported.; flags: --branch
    branches apps delete - DELETE /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/apps [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    branches apps view - Read /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/apps [intent=direct_read availability=planned]; notes: Branch names may contain slashes; direct reads remain planned until GitHub ref-like path parameter encoding is supported.; flags: --branch
    branches apps create - POST /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/apps [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    branches apps set - PUT /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/apps [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    branches teams delete - DELETE /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    branches teams view - Read /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams [intent=direct_read availability=planned]; notes: Branch names may contain slashes; direct reads remain planned until GitHub ref-like path parameter encoding is supported.; flags: --branch
    branches teams create - POST /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    branches teams set - PUT /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    branches users delete - DELETE /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    branches users view - Read /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users [intent=direct_read availability=planned]; notes: Branch names may contain slashes; direct reads remain planned until GitHub ref-like path parameter encoding is supported.; flags: --branch
    branches users create - POST /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    branches users set - PUT /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    branches rename create - POST /repos/{owner}/{repo}/branches/{branch}/rename [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: critical; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --branch
    check-runs create - POST /repos/{owner}/{repo}/check-runs [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    check-runs view - Read /repos/{owner}/{repo}/check-runs/{check_run_id} [intent=direct_read availability=implemented]; flags: --check-run-id
    check-runs update - PATCH /repos/{owner}/{repo}/check-runs/{check_run_id} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --check-run-id
    check-runs annotations view - Read /repos/{owner}/{repo}/check-runs/{check_run_id}/annotations [intent=direct_read availability=implemented]; flags: --check-run-id
    check-runs rerequest create - POST /repos/{owner}/{repo}/check-runs/{check_run_id}/rerequest [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --check-run-id
    check-suites create - POST /repos/{owner}/{repo}/check-suites [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    check-suites preferences update - PATCH /repos/{owner}/{repo}/check-suites/preferences [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    check-suites view - Read /repos/{owner}/{repo}/check-suites/{check_suite_id} [intent=direct_read availability=implemented]; flags: --check-suite-id
    check-suites check-runs view - Read /repos/{owner}/{repo}/check-suites/{check_suite_id}/check-runs [intent=direct_read availability=implemented]; flags: --check-suite-id
    check-suites rerequest create - POST /repos/{owner}/{repo}/check-suites/{check_suite_id}/rerequest [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --check-suite-id
    code-quality findings view - Read /repos/{owner}/{repo}/code-quality/findings [intent=direct_read availability=implemented]
    code-quality findings view-2 - Read /repos/{owner}/{repo}/code-quality/findings/{finding_number} [intent=direct_read availability=implemented]; flags: --finding-number
    code-quality setup view - Read /repos/{owner}/{repo}/code-quality/setup [intent=direct_read availability=implemented]
    code-quality setup update - PATCH /repos/{owner}/{repo}/code-quality/setup [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    code-scanning autofix view - Read /repos/{owner}/{repo}/code-scanning/alerts/{alert_number}/autofix [intent=direct_read availability=implemented]; flags: --alert-number
    code-scanning autofix create - POST /repos/{owner}/{repo}/code-scanning/alerts/{alert_number}/autofix [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --alert-number
    code-scanning commits create - POST /repos/{owner}/{repo}/code-scanning/alerts/{alert_number}/autofix/commits [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --alert-number
    code-scanning instances view - Read /repos/{owner}/{repo}/code-scanning/alerts/{alert_number}/instances [intent=direct_read availability=implemented]; flags: --alert-number
    code-scanning analyses view - Read /repos/{owner}/{repo}/code-scanning/analyses [intent=direct_read availability=implemented]
    code-scanning analyses delete - DELETE /repos/{owner}/{repo}/code-scanning/analyses/{analysis_id} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --analysis-id
    code-scanning analyses view-2 - Read /repos/{owner}/{repo}/code-scanning/analyses/{analysis_id} [intent=direct_read availability=implemented]; flags: --analysis-id
    code-scanning databases view - Read /repos/{owner}/{repo}/code-scanning/codeql/databases [intent=direct_read availability=implemented]
    code-scanning databases delete - DELETE /repos/{owner}/{repo}/code-scanning/codeql/databases/{language} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --language
    code-scanning databases view-2 - Read /repos/{owner}/{repo}/code-scanning/codeql/databases/{language} [intent=direct_read availability=implemented]; flags: --language
    code-scanning variant-analyses create - POST /repos/{owner}/{repo}/code-scanning/codeql/variant-analyses [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    code-scanning variant-analyses view - Read /repos/{owner}/{repo}/code-scanning/codeql/variant-analyses/{codeql_variant_analysis_id} [intent=direct_read availability=implemented]; flags: --codeql-variant-analysis-id
    code-scanning repos view - Read /repos/{owner}/{repo}/code-scanning/codeql/variant-analyses/{codeql_variant_analysis_id}/repos/{repo_owner}/{repo_name} [intent=direct_read availability=implemented]; flags: --codeql-variant-analysis-id, --repo-owner, --repo-name
    code-scanning default-setup view - Read /repos/{owner}/{repo}/code-scanning/default-setup [intent=direct_read availability=implemented]
    code-scanning default-setup update - PATCH /repos/{owner}/{repo}/code-scanning/default-setup [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    code-sanning upload - POST /repos/{owner}/{repo}/code-scanning/sarifs [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    code-scanning sarifs view - Read /repos/{owner}/{repo}/code-scanning/sarifs/{sarif_id} [intent=direct_read availability=implemented]; flags: --sarif-id
    code-security-configuration view - Read /repos/{owner}/{repo}/code-security-configuration [intent=direct_read availability=planned]; notes: GitHub may return a 204 no-content success response for this endpoint; direct reads remain planned until no-content direct-read handling is supported.
    codeowners errors view - Read /repos/{owner}/{repo}/codeowners/errors [intent=direct_read availability=implemented]
    codespaces view - Read /repos/{owner}/{repo}/codespaces [intent=direct_read availability=implemented]
    codespaces create - POST /repos/{owner}/{repo}/codespaces [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    codespaces devcontainers view - Read /repos/{owner}/{repo}/codespaces/devcontainers [intent=direct_read availability=implemented]
    codespaces machines view - Read /repos/{owner}/{repo}/codespaces/machines [intent=direct_read availability=implemented]
    codespaces new view - Read /repos/{owner}/{repo}/codespaces/new [intent=direct_read availability=implemented]
    codespaces permissions_check view - Read /repos/{owner}/{repo}/codespaces/permissions_check [intent=direct_read availability=implemented]
    codespaces secrets view - Read /repos/{owner}/{repo}/codespaces/secrets [intent=direct_read availability=implemented]
    codespaces public-key view - Read /repos/{owner}/{repo}/codespaces/secrets/public-key [intent=direct_read availability=implemented]
    codespaces secrets view-2 - Read /repos/{owner}/{repo}/codespaces/secrets/{secret_name} [intent=direct_read availability=implemented]; flags: --secret-name
    comments reactions view - Read /repos/{owner}/{repo}/comments/{comment_id}/reactions [intent=direct_read availability=implemented]; flags: --comment-id
    comments reactions create - POST /repos/{owner}/{repo}/comments/{comment_id}/reactions [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --comment-id
    comments reactions delete - DELETE /repos/{owner}/{repo}/comments/{comment_id}/reactions/{reaction_id} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --comment-id, --reaction-id
    commits branches-where-head view - Read /repos/{owner}/{repo}/commits/{commit_sha}/branches-where-head [intent=direct_read availability=implemented]; flags: --commit-sha
    commits pulls view - Read /repos/{owner}/{repo}/commits/{commit_sha}/pulls [intent=direct_read availability=implemented]; flags: --commit-sha
    commits check-runs view - Read /repos/{owner}/{repo}/commits/{ref}/check-runs [intent=direct_read availability=planned]; notes: Commit refs may contain branch-style slashes; direct reads remain planned until GitHub ref-like path parameter encoding is supported.; flags: --ref
    commits check-suites view - Read /repos/{owner}/{repo}/commits/{ref}/check-suites [intent=direct_read availability=planned]; notes: Commit refs may contain branch-style slashes; direct reads remain planned until GitHub ref-like path parameter encoding is supported.; flags: --ref
    commits status view - Read /repos/{owner}/{repo}/commits/{ref}/status [intent=direct_read availability=planned]; notes: Commit refs may contain branch-style slashes; direct reads remain planned until GitHub ref-like path parameter encoding is supported.; flags: --ref
    commits statuses view - Read /repos/{owner}/{repo}/commits/{ref}/statuses [intent=direct_read availability=planned]; notes: Commit refs may contain branch-style slashes; direct reads remain planned until GitHub ref-like path parameter encoding is supported.; flags: --ref
    community profile view - Read /repos/{owner}/{repo}/community/profile [intent=direct_read availability=implemented]
    compare view - Read /repos/{owner}/{repo}/compare/{basehead} [intent=direct_read availability=planned]; notes: Compare refs require a base...head path segment; direct reads remain planned until GitHub ref-like path parameter encoding is supported.; flags: --basehead
    dependabot secrets view - Read /repos/{owner}/{repo}/dependabot/secrets [intent=direct_read availability=implemented]
    dependabot public-key view - Read /repos/{owner}/{repo}/dependabot/secrets/public-key [intent=direct_read availability=implemented]
    dependabot secrets view-2 - Read /repos/{owner}/{repo}/dependabot/secrets/{secret_name} [intent=direct_read availability=implemented]; flags: --secret-name
    dependency-graph compare view - Read /repos/{owner}/{repo}/dependency-graph/compare/{basehead} [intent=direct_read availability=planned]; notes: Compare refs require a base...head path segment; direct reads remain planned until GitHub ref-like path parameter encoding is supported.; flags: --basehead
    dependency-graph snapshots create - POST /repos/{owner}/{repo}/dependency-graph/snapshots [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    deployments delete - DELETE /repos/{owner}/{repo}/deployments/{deployment_id} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: critical; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --deployment-id
    deployments statuses view - Read /repos/{owner}/{repo}/deployments/{deployment_id}/statuses [intent=direct_read availability=implemented]; flags: --deployment-id
    deployments statuses create - POST /repos/{owner}/{repo}/deployments/{deployment_id}/statuses [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --deployment-id
    dispatches create - POST /repos/{owner}/{repo}/dispatches [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    environments deployment-branch-policies view - Read /repos/{owner}/{repo}/environments/{environment_name}/deployment-branch-policies [intent=direct_read availability=planned]; notes: Environment names may contain slashes and must be URL encoded; direct reads remain planned until GitHub environment-name path parameter encoding is supported.; flags: --environment-name
    environments deployment-branch-policies create - POST /repos/{owner}/{repo}/environments/{environment_name}/deployment-branch-policies [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --environment-name
    environments deployment-branch-policies delete - DELETE /repos/{owner}/{repo}/environments/{environment_name}/deployment-branch-policies/{branch_policy_id} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --environment-name, --branch-policy-id
    environments deployment-branch-policies view-2 - Read /repos/{owner}/{repo}/environments/{environment_name}/deployment-branch-policies/{branch_policy_id} [intent=direct_read availability=planned]; notes: Environment names may contain slashes and must be URL encoded; direct reads remain planned until GitHub environment-name path parameter encoding is supported.; flags: --environment-name, --branch-policy-id
    environments deployment-branch-policies set - PUT /repos/{owner}/{repo}/environments/{environment_name}/deployment-branch-policies/{branch_policy_id} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --environment-name, --branch-policy-id
    environments deployment_protection_rules view - Read /repos/{owner}/{repo}/environments/{environment_name}/deployment_protection_rules [intent=direct_read availability=planned]; notes: Environment names may contain slashes and must be URL encoded; direct reads remain planned until GitHub environment-name path parameter encoding is supported.; flags: --environment-name
    environments deployment_protection_rules create - POST /repos/{owner}/{repo}/environments/{environment_name}/deployment_protection_rules [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --environment-name
    environments apps view - Read /repos/{owner}/{repo}/environments/{environment_name}/deployment_protection_rules/apps [intent=direct_read availability=planned]; notes: Environment names may contain slashes and must be URL encoded; direct reads remain planned until GitHub environment-name path parameter encoding is supported.; flags: --environment-name
    environments deployment_protection_rules delete - DELETE /repos/{owner}/{repo}/environments/{environment_name}/deployment_protection_rules/{protection_rule_id} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --environment-name, --protection-rule-id
    environments deployment_protection_rules view-2 - Read /repos/{owner}/{repo}/environments/{environment_name}/deployment_protection_rules/{protection_rule_id} [intent=direct_read availability=planned]; notes: Environment names may contain slashes and must be URL encoded; direct reads remain planned until GitHub environment-name path parameter encoding is supported.; flags: --environment-name, --protection-rule-id
    environments secrets view - Read /repos/{owner}/{repo}/environments/{environment_name}/secrets [intent=direct_read availability=planned]; notes: Environment names may contain slashes and must be URL encoded; direct reads remain planned until GitHub environment-name path parameter encoding is supported.; flags: --environment-name
    environments public-key view - Read /repos/{owner}/{repo}/environments/{environment_name}/secrets/public-key [intent=direct_read availability=planned]; notes: Environment names may contain slashes and must be URL encoded; direct reads remain planned until GitHub environment-name path parameter encoding is supported.; flags: --environment-name
    environments secrets view-2 - Read /repos/{owner}/{repo}/environments/{environment_name}/secrets/{secret_name} [intent=direct_read availability=planned]; notes: Environment names may contain slashes and must be URL encoded; direct reads remain planned until GitHub environment-name path parameter encoding is supported.; flags: --environment-name, --secret-name
    environments variables view - Read /repos/{owner}/{repo}/environments/{environment_name}/variables [intent=direct_read availability=planned]; notes: Environment names may contain slashes and must be URL encoded; direct reads remain planned until GitHub environment-name path parameter encoding is supported.; flags: --environment-name
    environments variables view-2 - Read /repos/{owner}/{repo}/environments/{environment_name}/variables/{name} [intent=direct_read availability=planned]; notes: Environment names may contain slashes and must be URL encoded; direct reads remain planned until GitHub environment-name path parameter encoding is supported.; flags: --environment-name, --name
    git blobs create - POST /repos/{owner}/{repo}/git/blobs [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    git blobs view - Read /repos/{owner}/{repo}/git/blobs/{file_sha} [intent=direct_read availability=implemented]; flags: --file-sha
    git commits create - POST /repos/{owner}/{repo}/git/commits [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    git ref view - Read /repos/{owner}/{repo}/git/ref/{ref} [intent=direct_read availability=planned]; notes: Git refs require heads/<branch> or tags/<tag> path segments; direct reads remain planned until GitHub ref-like path parameter encoding is supported.; flags: --ref
    git tags view - Read /repos/{owner}/{repo}/git/tags/{tag_sha} [intent=direct_read availability=implemented]; flags: --tag-sha
    git trees create - POST /repos/{owner}/{repo}/git/trees [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    git trees view - Read /repos/{owner}/{repo}/git/trees/{tree_sha} [intent=direct_read availability=planned]; notes: Tree refs may contain branch or tag names with slashes; direct reads remain planned until GitHub ref-like path parameter encoding is supported.; flags: --tree-sha
    hash-algorithm view - Read /repos/{owner}/{repo}/hash-algorithm [intent=direct_read availability=implemented]
    hooks deliveries view - Read /repos/{owner}/{repo}/hooks/{hook_id}/deliveries [intent=direct_read availability=implemented]; flags: --hook-id
    hooks deliveries view-2 - Read /repos/{owner}/{repo}/hooks/{hook_id}/deliveries/{delivery_id} [intent=direct_read availability=implemented]; flags: --hook-id, --delivery-id
    webhook create - POST /repos/{owner}/{repo}/hooks/{hook_id}/deliveries/{delivery_id}/attempts [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --hook-id, --delivery-id
    webhook create-2 - POST /repos/{owner}/{repo}/hooks/{hook_id}/pings [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: low; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --hook-id
    webhook create-3 - POST /repos/{owner}/{repo}/hooks/{hook_id}/tests [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: low; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --hook-id
    immutable-releases delete - DELETE /repos/{owner}/{repo}/immutable-releases [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    immutable-releases view - Read /repos/{owner}/{repo}/immutable-releases [intent=direct_read availability=implemented]
    immutable-releases set - PUT /repos/{owner}/{repo}/immutable-releases [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    import delete - DELETE /repos/{owner}/{repo}/import [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    import view - Read /repos/{owner}/{repo}/import [intent=direct_read availability=implemented]
    import update - PATCH /repos/{owner}/{repo}/import [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    import set - PUT /repos/{owner}/{repo}/import [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    import authors view - Read /repos/{owner}/{repo}/import/authors [intent=direct_read availability=implemented]
    import authors update - PATCH /repos/{owner}/{repo}/import/authors/{author_id} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --author-id
    import large_files view - Read /repos/{owner}/{repo}/import/large_files [intent=direct_read availability=implemented]
    import lfs update - PATCH /repos/{owner}/{repo}/import/lfs [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    installation view - Read /repos/{owner}/{repo}/installation [intent=direct_read availability=implemented]
    interaction-limits delete - DELETE /repos/{owner}/{repo}/interaction-limits [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    interaction-limits view - Read /repos/{owner}/{repo}/interaction-limits [intent=direct_read availability=implemented]
    interaction-limits set - PUT /repos/{owner}/{repo}/interaction-limits [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    interaction-limits bypass-list delete - DELETE /repos/{owner}/{repo}/interaction-limits/pulls/bypass-list [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    interaction-limits bypass-list view - Read /repos/{owner}/{repo}/interaction-limits/pulls/bypass-list [intent=direct_read availability=implemented]
    interaction-limits bypass-list set - PUT /repos/{owner}/{repo}/interaction-limits/pulls/bypass-list [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    invitations delete - DELETE /repos/{owner}/{repo}/invitations/{invitation_id} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --invitation-id
    invitations update - PATCH /repos/{owner}/{repo}/invitations/{invitation_id} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --invitation-id
    issue-types view - Read /repos/{owner}/{repo}/issue-types [intent=direct_read availability=implemented]
    issues pin delete - DELETE /repos/{owner}/{repo}/issues/comments/{comment_id}/pin [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --comment-id
    issues pin set - PUT /repos/{owner}/{repo}/issues/comments/{comment_id}/pin [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --comment-id
    issues reactions view - Read /repos/{owner}/{repo}/issues/comments/{comment_id}/reactions [intent=direct_read availability=implemented]; flags: --comment-id
    issues reactions create - POST /repos/{owner}/{repo}/issues/comments/{comment_id}/reactions [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --comment-id
    issues reactions delete - DELETE /repos/{owner}/{repo}/issues/comments/{comment_id}/reactions/{reaction_id} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --comment-id, --reaction-id
    issues assignees view - Read /repos/{owner}/{repo}/issues/{issue_number}/assignees/{assignee} [intent=direct_read availability=planned]; notes: GitHub documents a 204 no-content success response for this endpoint; direct reads remain planned until no-content direct-read handling is supported.; flags: --issue-number, --assignee
    issues blocked_by view - Read /repos/{owner}/{repo}/issues/{issue_number}/dependencies/blocked_by [intent=direct_read availability=implemented]; flags: --issue-number
    issues blocked_by create - POST /repos/{owner}/{repo}/issues/{issue_number}/dependencies/blocked_by [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --issue-number
    issues blocked_by delete - DELETE /repos/{owner}/{repo}/issues/{issue_number}/dependencies/blocked_by/{issue_id} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --issue-number, --issue-id
    issues blocking view - Read /repos/{owner}/{repo}/issues/{issue_number}/dependencies/blocking [intent=direct_read availability=implemented]; flags: --issue-number
    issues issue-field-values view - Read /repos/{owner}/{repo}/issues/{issue_number}/issue-field-values [intent=direct_read availability=implemented]; flags: --issue-number
    issues issue-field-values create - POST /repos/{owner}/{repo}/issues/{issue_number}/issue-field-values [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --issue-number
    issues issue-field-values set - PUT /repos/{owner}/{repo}/issues/{issue_number}/issue-field-values [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --issue-number
    issues issue-field-values delete - DELETE /repos/{owner}/{repo}/issues/{issue_number}/issue-field-values/{issue_field_id} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --issue-number, --issue-field-id
    issues parent view - Read /repos/{owner}/{repo}/issues/{issue_number}/parent [intent=direct_read availability=implemented]; flags: --issue-number
    issues reactions view-2 - Read /repos/{owner}/{repo}/issues/{issue_number}/reactions [intent=direct_read availability=implemented]; flags: --issue-number
    issues reactions create-2 - POST /repos/{owner}/{repo}/issues/{issue_number}/reactions [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --issue-number
    issues reactions delete-2 - DELETE /repos/{owner}/{repo}/issues/{issue_number}/reactions/{reaction_id} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --issue-number, --reaction-id
    issues sub_issue delete - DELETE /repos/{owner}/{repo}/issues/{issue_number}/sub_issue [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --issue-number
    issues sub_issues view - Read /repos/{owner}/{repo}/issues/{issue_number}/sub_issues [intent=direct_read availability=implemented]; flags: --issue-number
    issues sub_issues create - POST /repos/{owner}/{repo}/issues/{issue_number}/sub_issues [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --issue-number
    issues priority update - PATCH /repos/{owner}/{repo}/issues/{issue_number}/sub_issues/priority [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --issue-number
    license view - Read /repos/{owner}/{repo}/license [intent=direct_read availability=implemented]
    merge-upstream create - POST /repos/{owner}/{repo}/merge-upstream [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    milestones labels view - Read /repos/{owner}/{repo}/milestones/{milestone_number}/labels [intent=direct_read availability=implemented]; flags: --milestone-number
    notifications view - Read /repos/{owner}/{repo}/notifications [intent=direct_read availability=implemented]
    notifications set - PUT /repos/{owner}/{repo}/notifications [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    pages delete - DELETE /repos/{owner}/{repo}/pages [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    pages view - Read /repos/{owner}/{repo}/pages [intent=direct_read availability=implemented]
    pages create - POST /repos/{owner}/{repo}/pages [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    pages set - PUT /repos/{owner}/{repo}/pages [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    pages builds view - Read /repos/{owner}/{repo}/pages/builds [intent=direct_read availability=implemented]
    pages builds create - POST /repos/{owner}/{repo}/pages/builds [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    pages latest view - Read /repos/{owner}/{repo}/pages/builds/latest [intent=direct_read availability=implemented]
    pages builds view-2 - Read /repos/{owner}/{repo}/pages/builds/{build_id} [intent=direct_read availability=implemented]; flags: --build-id
    pages deployments create - POST /repos/{owner}/{repo}/pages/deployments [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    pages deployments view - Read /repos/{owner}/{repo}/pages/deployments/{pages_deployment_id} [intent=direct_read availability=implemented]; flags: --pages-deployment-id
    pages cancel create - POST /repos/{owner}/{repo}/pages/deployments/{pages_deployment_id}/cancel [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --pages-deployment-id
    pages health view - Read /repos/{owner}/{repo}/pages/health [intent=direct_read availability=planned]; notes: GitHub may return a 202 pending response for this endpoint; direct reads remain planned until pending/no-content direct-read handling is supported.
    private-vulnerability-reporting delete - DELETE /repos/{owner}/{repo}/private-vulnerability-reporting [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    private-vulnerability-reporting view - Read /repos/{owner}/{repo}/private-vulnerability-reporting [intent=direct_read availability=implemented]
    private-vulnerability-reporting set - PUT /repos/{owner}/{repo}/private-vulnerability-reporting [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    properties values view - Read /repos/{owner}/{repo}/properties/values [intent=direct_read availability=implemented]
    properties values update - PATCH /repos/{owner}/{repo}/properties/values [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    pulls reactions view - Read /repos/{owner}/{repo}/pulls/comments/{comment_id}/reactions [intent=direct_read availability=implemented]; flags: --comment-id
    pulls reactions create - POST /repos/{owner}/{repo}/pulls/comments/{comment_id}/reactions [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --comment-id
    pulls reactions delete - DELETE /repos/{owner}/{repo}/pulls/comments/{comment_id}/reactions/{reaction_id} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --comment-id, --reaction-id
    pulls codespaces create - POST /repos/{owner}/{repo}/pulls/{pull_number}/codespaces [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --pull-number
    pulls commits view - Read /repos/{owner}/{repo}/pulls/{pull_number}/commits [intent=direct_read availability=implemented]; flags: --pull-number
    pulls files view - Read /repos/{owner}/{repo}/pulls/{pull_number}/files [intent=direct_read availability=implemented]; flags: --pull-number
    pulls merge view - Read /repos/{owner}/{repo}/pulls/{pull_number}/merge [intent=direct_read availability=planned]; notes: GitHub documents a 204 no-content success response for this endpoint; direct reads remain planned until no-content direct-read handling is supported.; flags: --pull-number
    pulls requested_reviewers delete - DELETE /repos/{owner}/{repo}/pulls/{pull_number}/requested_reviewers [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --pull-number
    pulls reviews view - Read /repos/{owner}/{repo}/pulls/{pull_number}/reviews [intent=direct_read availability=implemented]; flags: --pull-number
    pulls comments view - Read /repos/{owner}/{repo}/pulls/{pull_number}/reviews/{review_id}/comments [intent=direct_read availability=implemented]; flags: --pull-number, --review-id
    readme view - Read /repos/{owner}/{repo}/readme [intent=direct_read availability=implemented]
    readme view-2 - Read /repos/{owner}/{repo}/readme/{dir} [intent=direct_read availability=planned]; notes: Directory paths may contain slashes; direct reads remain planned until nested repository path parameter encoding is supported.; flags: --dir
    releases assets view - Read /repos/{owner}/{repo}/releases/assets/{asset_id} [intent=direct_read availability=implemented]; flags: --asset-id
    releases generate-notes view - POST /repos/{owner}/{repo}/releases/generate-notes [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: low; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    releases assets view-2 - Read /repos/{owner}/{repo}/releases/{release_id}/assets [intent=direct_read availability=implemented]; flags: --release-id
    releases assets view-3 - POST /repos/{owner}/{repo}/releases/{release_id}/assets [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --release-id
    releases reactions view - Read /repos/{owner}/{repo}/releases/{release_id}/reactions [intent=direct_read availability=implemented]; flags: --release-id
    releases reactions create - POST /repos/{owner}/{repo}/releases/{release_id}/reactions [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --release-id
    releases reactions delete - DELETE /repos/{owner}/{repo}/releases/{release_id}/reactions/{reaction_id} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --release-id, --reaction-id
    rulesets rule-suites view - Read /repos/{owner}/{repo}/rulesets/rule-suites [intent=direct_read availability=implemented]
    rulesets rule-suites view-2 - Read /repos/{owner}/{repo}/rulesets/rule-suites/{rule_suite_id} [intent=direct_read availability=implemented]; flags: --rule-suite-id
    rulesets history view - Read /repos/{owner}/{repo}/rulesets/{ruleset_id}/history [intent=direct_read availability=implemented]; flags: --ruleset-id
    rulesets history view-2 - Read /repos/{owner}/{repo}/rulesets/{ruleset_id}/history/{version_id} [intent=direct_read availability=implemented]; flags: --ruleset-id, --version-id
    secret-scanning locations view - Read /repos/{owner}/{repo}/secret-scanning/alerts/{alert_number}/locations [intent=direct_read availability=implemented]; flags: --alert-number
    secret-scanning push-protection-bypasses create - POST /repos/{owner}/{repo}/secret-scanning/push-protection-bypasses [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    secret-scanning scan-history view - Read /repos/{owner}/{repo}/secret-scanning/scan-history [intent=direct_read availability=implemented]
    security-advisories create - POST /repos/{owner}/{repo}/security-advisories [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    security-advisories reports create - POST /repos/{owner}/{repo}/security-advisories/reports [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    security-advisories update - PATCH /repos/{owner}/{repo}/security-advisories/{ghsa_id} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --ghsa-id
    security-advisories cve create - POST /repos/{owner}/{repo}/security-advisories/{ghsa_id}/cve [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --ghsa-id
    security-advisories forks create - POST /repos/{owner}/{repo}/security-advisories/{ghsa_id}/forks [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --ghsa-id
    stats code_frequency view - Read /repos/{owner}/{repo}/stats/code_frequency [intent=direct_read availability=planned]; notes: GitHub may return 202/204 no-content responses while repository statistics are generated; direct reads remain planned until pending/no-content direct-read handling is supported.
    stats commit_activity view - Read /repos/{owner}/{repo}/stats/commit_activity [intent=direct_read availability=planned]; notes: GitHub may return 202/204 no-content responses while repository statistics are generated; direct reads remain planned until pending/no-content direct-read handling is supported.
    stats contributors view - Read /repos/{owner}/{repo}/stats/contributors [intent=direct_read availability=planned]; notes: GitHub may return 202/204 no-content responses while repository statistics are generated; direct reads remain planned until pending/no-content direct-read handling is supported.
    stats participation view - Read /repos/{owner}/{repo}/stats/participation [intent=direct_read availability=implemented]
    stats punch_card view - Read /repos/{owner}/{repo}/stats/punch_card [intent=direct_read availability=planned]; notes: GitHub may return a 204 no-content response for repository statistics; direct reads remain planned until no-content direct-read handling is supported.
    statuses create - POST /repos/{owner}/{repo}/statuses/{sha} [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.; flags: --sha
    subscription delete - DELETE /repos/{owner}/{repo}/subscription [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    subscription view - Read /repos/{owner}/{repo}/subscription [intent=direct_read availability=implemented]
    subscription set - PUT /repos/{owner}/{repo}/subscription [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: medium; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    teams view - Read /repos/{owner}/{repo}/teams [intent=direct_read availability=implemented]
    traffic clones view - Read /repos/{owner}/{repo}/traffic/clones [intent=direct_read availability=implemented]
    traffic paths view - Read /repos/{owner}/{repo}/traffic/popular/paths [intent=direct_read availability=implemented]
    traffic referrers view - Read /repos/{owner}/{repo}/traffic/popular/referrers [intent=direct_read availability=implemented]
    traffic views view - Read /repos/{owner}/{repo}/traffic/views [intent=direct_read availability=implemented]
    transfer create - POST /repos/{owner}/{repo}/transfer [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: critical; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    vulnerability-alerts delete - DELETE /repos/{owner}/{repo}/vulnerability-alerts [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
    vulnerability-alerts view - Read /repos/{owner}/{repo}/vulnerability-alerts [intent=direct_read availability=planned]; notes: GitHub documents a 204 no-content success response for this endpoint; direct reads remain planned until no-content direct-read handling is supported.
    vulnerability-alerts set - PUT /repos/{owner}/{repo}/vulnerability-alerts [intent=direct_write availability=planned]; approval: Planned destructive/sensitive writes require plan, preview, explicit approval, execute, and typed `destructive` confirmation when destructive.; risk: high; notes: Blocked until endpoint-specific typed request schema, fixture coverage, and safety metadata are authored.
  Help topics:
    authentication - Use pm credentials for public, token, or GitHub App repository access. Never print stored tokens.
    execution-model - ETL commands map to streams. Reverse ETL commands map to approved write actions and keep plan, preview, approval, execute.
    local-workflows - Commands that depend on local git, browser, shell completion, extensions, or gh config are documented but not connector-dispatched.
    known-gaps - Projects, Discussions, Search, direct reads, and sensitive/admin surfaces are planned follow-up slices.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect github

  # Inspect as structured JSON
  pm connectors inspect github --json

  # Public repository credential
  pm credentials add github-public --connector github --config owner=octocat --config repo=Hello-World

  # Token credential
  export GITHUB_TOKEN=...
  pm credentials add github-token --connector github --config owner=OWNER --config repo=REPO --from-env token=GITHUB_TOKEN

  # GitHub App credential
  pm credentials add github-app --connector github --config owner=OWNER --config repo=REPO --config auth_type=github_app --config app_id=12345 --config installation_id=67890 --value-stdin private_key < app-private-key.pem

  # Pull request ETL
  pm connections create github_prs_to_warehouse --source github:github-token --destination warehouse:warehouse-local --stream pull_requests --primary-key node_id --cursor updated_at --table github_pull_requests
  pm etl run --connection github_prs_to_warehouse --stream pull_requests --batch-size 100 --json

  # Approved pull request creation
  pm reverse plan prs_to_github --source-table github_pr_candidates --destination github:github-token --action create_pull_request --map title:title --map body:body --map head:head --map base:base --map reviewers:reviewers
  pm reverse preview <plan-id> --json
  pm reverse run <plan-id> --approve <approval-token> --json

AGENT WORKFLOW
  - Run pm connectors inspect github before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

SEE ALSO
  GitHub REST authentication: https://docs.github.com/en/rest/authentication/authenticating-to-the-rest-api
  GitHub App installation auth: https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/authenticating-as-a-github-app-installation
  GitHub pull requests REST API: https://docs.github.com/en/rest/pulls/pulls
  GitHub issues REST API: https://docs.github.com/en/rest/issues/issues
  GitHub issue comments REST API: https://docs.github.com/en/rest/issues/comments
  GitHub labels REST API: https://docs.github.com/en/rest/issues/labels
  GitHub commits REST API: https://docs.github.com/en/rest/commits/commits
  GitHub branches REST API: https://docs.github.com/en/rest/branches/branches
  GitHub releases REST API: https://docs.github.com/en/rest/releases/releases
  GitHub Actions workflows REST API: https://docs.github.com/en/rest/actions/workflows
  GitHub Actions workflow runs REST API: https://docs.github.com/en/rest/actions/workflow-runs
  GitHub Actions artifacts REST API: https://docs.github.com/en/rest/actions/artifacts
  GitHub repository contents REST API: https://docs.github.com/en/rest/repos/contents

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
