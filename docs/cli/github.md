# pm github

```text
NAME
  pm github - Work with GitHub repositories from the command line.

SYNOPSIS
  pm github <command> <subcommand> [flags]

DESCRIPTION
  Work with GitHub repositories from the command line.

COMMANDS
  Core Commands
    issue list - List issues [intent=etl availability=implemented stream=issues]
    issue view - View issue details [intent=etl availability=partial stream=issues]
    issue create - Create an issue [intent=reverse_etl availability=implemented write=create_issue]; risk: Creates a visible issue in the configured repository.; approval: Reverse ETL writes require plan, preview, approval, execute.
    issue edit - Edit an issue [intent=reverse_etl availability=implemented write=update_issue]; risk: Mutates title, body, labels, assignees, milestone, or state on an existing issue.; approval: Reverse ETL writes require plan, preview, approval, execute.
    issue close - Close an issue [intent=reverse_etl availability=implemented write=close_issue]; risk: Closes an existing issue.; approval: Reverse ETL writes require plan, preview, approval, execute.
    issue reopen - Reopen an issue [intent=reverse_etl availability=implemented write=reopen_issue]; risk: Reopens a previously closed issue.; approval: Reverse ETL writes require plan, preview, approval, execute.
    issue comment - Comment on an issue [intent=reverse_etl availability=implemented write=comment_issue]; risk: Adds a visible issue comment.; approval: Reverse ETL writes require plan, preview, approval, execute.
    issue lock - Lock issue conversation [intent=reverse_etl availability=implemented write=lock_issue]; risk: Locks issue conversation for repository users.; approval: Reverse ETL writes require plan, preview, approval, execute.
    issue unlock - Unlock issue conversation [intent=reverse_etl availability=implemented write=unlock_issue]; risk: Unlocks issue conversation.; approval: Reverse ETL writes require plan, preview, approval, execute.
    issue delete - Delete an issue [intent=direct_write availability=unsafe_or_disallowed operation=github.issue.delete]
    issue develop - Manage development branches for an issue [intent=local_workflow availability=unsupported_local]
    issue status - Show relevant issues [intent=direct_read availability=planned]
    issue pin - Pin an issue [intent=direct_write availability=unsupported_api]
    issue unpin - Unpin an issue [intent=direct_write availability=unsupported_api]
    issue transfer - Transfer an issue [intent=direct_write availability=unsafe_or_disallowed]
    pr list - List pull requests [intent=etl availability=implemented stream=pull_requests]
    pr view - View pull request details [intent=etl availability=partial stream=pull_requests]
    pr create - Create a pull request [intent=reverse_etl availability=implemented write=create_pull_request]; risk: Creates a visible pull request and may request reviewers.; approval: Reverse ETL writes require plan, preview, approval, execute.
    pr edit - Edit a pull request [intent=reverse_etl availability=implemented write=update_pull_request]; risk: Mutates an existing pull request.; approval: Reverse ETL writes require plan, preview, approval, execute.
    pr close - Close a pull request [intent=reverse_etl availability=implemented write=close_pull_request]; risk: Closes an existing pull request.; approval: Reverse ETL writes require plan, preview, approval, execute.
    pr reopen - Reopen a pull request [intent=reverse_etl availability=implemented write=reopen_pull_request]; risk: Reopens a previously closed pull request.; approval: Reverse ETL writes require plan, preview, approval, execute.
    pr comment - Comment on a pull request [intent=reverse_etl availability=implemented write=comment_issue]; risk: Comments on a pull request (PRs are issues in GitHub's data model).; approval: Reverse ETL writes require plan, preview, approval, execute.
    pr merge - Merge a pull request [intent=reverse_etl availability=implemented write=merge_pull_request]; risk: Merges code into the pull request base branch.; approval: Reverse ETL writes require plan, preview, approval, execute.
    pr review - Add a pull request review [intent=reverse_etl availability=implemented write=create_pull_request_review]; risk: Adds a visible pull request review.; approval: Reverse ETL writes require plan, preview, approval, execute.
    pr checks - Show pull request checks [intent=direct_read availability=planned]
    pr diff - Show pull request diff [intent=direct_read availability=unsupported_api]
    pr checkout - Check out a pull request locally [intent=local_workflow availability=unsupported_local]
    pr ready - Mark a draft pull request ready [intent=direct_write availability=unsupported_api]
    pr update-branch - Update a pull request branch [intent=reverse_etl availability=implemented write=update_pull_request_branch]; risk: Updates a pull request branch from its base branch.; approval: Reverse ETL writes require plan, preview, approval, execute.
    pr status - Show relevant pull requests [intent=direct_read availability=planned]
    pr lock - Lock pull request conversation [intent=reverse_etl availability=implemented write=lock_issue]; risk: Locks a pull request's conversation.; approval: Reverse ETL writes require plan, preview, approval, execute.
    pr unlock - Unlock pull request conversation [intent=reverse_etl availability=implemented write=unlock_issue]; risk: Unlocks a pull request's conversation.; approval: Reverse ETL writes require plan, preview, approval, execute.
    pr revert - Revert a pull request [intent=direct_write availability=unsafe_or_disallowed]
    repo view - View repository metadata [intent=etl availability=implemented stream=repository]
    repo list - List repositories for an owner [intent=direct_read availability=unsupported_api]
    repo create - Create a repository [intent=direct_write availability=unsafe_or_disallowed]
    repo delete - Delete a repository [intent=direct_write availability=unsafe_or_disallowed]
    repo archive - Archive a repository [intent=direct_write availability=unsafe_or_disallowed]
    repo unarchive - Unarchive a repository [intent=direct_write availability=unsafe_or_disallowed]
    repo fork - Fork a repository [intent=reverse_etl availability=implemented write=create_fork]; risk: Creates a fork of the configured repository.; approval: Reverse ETL writes require plan, preview, approval, execute.
    repo clone - Clone a repository locally [intent=local_workflow availability=unsupported_local operation=github.repo.clone]
    repo sync - Sync a local repository [intent=local_workflow availability=unsupported_local]
    repo set-default - Set the default local repository [intent=config availability=unsupported_local]
    repo read-file - Read repository file metadata [intent=direct_read availability=implemented]; output: github_contents_file_metadata
    repo read-dir - Read repository directory contents [intent=direct_read availability=implemented]; output: github_contents_directory
    repo autolink list - List repository autolinks [intent=etl availability=implemented stream=autolinks]
    repo autolink create - Create a repository autolink [intent=direct_write availability=unsupported_api]
    repo autolink delete - Delete a repository autolink [intent=direct_write availability=unsupported_api]
    repo deploy-key list - List deploy keys [intent=etl availability=implemented stream=deploy_keys]
    repo deploy-key add - Add a deploy key [intent=reverse_etl availability=implemented write=create_deploy_key]; risk: Adds a deploy key to the repository.; approval: Reverse ETL writes require plan, preview, approval, execute.
    repo deploy-key delete - Delete a deploy key [intent=reverse_etl availability=implemented write=delete_deploy_key]; risk: Deletes a deploy key from the repository.; approval: Reverse ETL writes require plan, preview, approval, execute.
    repo license list - List license templates [intent=direct_read availability=unsupported_api]
    repo gitignore list - List gitignore templates [intent=direct_read availability=unsupported_api]
    repo ruleset create - Create a repository ruleset [intent=reverse_etl availability=implemented write=create_repo_ruleset]; risk: Creates repository rules that can affect contribution workflows.; approval: Reverse ETL writes require plan, preview, approval, execute.
    repo ruleset update - Update a repository ruleset [intent=reverse_etl availability=implemented write=update_repo_ruleset]; risk: Updates repository rules that can affect contribution workflows.; approval: Reverse ETL writes require plan, preview, approval, execute.
    repo ruleset delete - Delete a repository ruleset [intent=reverse_etl availability=implemented write=delete_repo_ruleset]; risk: Deletes repository rules that can affect contribution workflows.; approval: Reverse ETL writes require plan, preview, approval, execute.
    repo delete-2 - DELETE /repos/{owner}/{repo} [intent=reverse_etl availability=implemented write=repo]; risk: critical; approval: Reverse ETL writes require plan, preview, approval, execute.
    repo update - PATCH /repos/{owner}/{repo} [intent=reverse_etl availability=implemented write=repo2]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
    repo sbom view - Download /repos/{owner}/{repo}/dependency-graph/sbom [intent=direct_read availability=implemented operation=github.dependency_graph_sbom]
    repo sbom fetch - Download /repos/{owner}/{repo}/dependency-graph/sbom/fetch-report/{sbom_uuid} [intent=direct_read availability=implemented operation=github.dependency_graph_sbom_fetch_report_sbom_uuid]
    repo sbom generate - Download /repos/{owner}/{repo}/dependency-graph/sbom/generate-report [intent=direct_read availability=implemented operation=github.dependency_graph_sbom_generate_report]
    repo archive tarball - Download /repos/{owner}/{repo}/tarball/{ref} [intent=direct_read availability=implemented operation=github.tarball_ref]
    repo archive zipball - Download /repos/{owner}/{repo}/zipball/{ref} [intent=direct_read availability=implemented operation=github.zipball_ref]
    release list - List releases [intent=etl availability=implemented stream=releases]
    release view - View a release [intent=etl availability=partial stream=releases]
    release create - Create a release [intent=reverse_etl availability=implemented write=create_release]; risk: Creates a visible release.; approval: Reverse ETL writes require plan, preview, approval, execute.
    release edit - Edit a release [intent=reverse_etl availability=implemented write=update_release]; risk: Mutates an existing release.; approval: Reverse ETL writes require plan, preview, approval, execute.
    release delete - Delete a release [intent=reverse_etl availability=implemented write=delete_release]; risk: Deletes an existing release.; approval: Reverse ETL writes require plan, preview, approval, execute.
    release upload - Upload release assets [intent=direct_write availability=unsupported_local]
    release download - Download release assets [intent=local_workflow availability=unsupported_local operation=github.release.download_assets]
    release delete-asset - Delete a release asset [intent=reverse_etl availability=implemented write=delete_release_asset]; risk: Deletes an existing release asset.; approval: Reverse ETL writes require plan, preview, approval, execute.
    release verify - Verify release assets [intent=local_workflow availability=unsupported_local]
  GitHub Actions Commands
    workflow list - List workflows [intent=etl availability=implemented stream=workflows]
    workflow view - View workflow details [intent=etl availability=partial stream=workflows]
    workflow run - Dispatch a workflow [intent=reverse_etl availability=implemented write=dispatch_workflow]; risk: Triggers a GitHub Actions workflow run.; approval: Reverse ETL writes require plan, preview, approval, execute.
    workflow enable - Enable a workflow [intent=direct_write availability=unsupported_api]
    workflow disable - Disable a workflow [intent=direct_write availability=unsupported_api]
    run list - List workflow runs [intent=etl availability=implemented stream=workflow_runs]
    run view - View workflow run details [intent=etl availability=partial stream=workflow_runs]
    run rerun - Rerun a workflow run [intent=reverse_etl availability=implemented write=rerun_workflow_run]; risk: Reruns a GitHub Actions workflow run.; approval: Reverse ETL writes require plan, preview, approval, execute.
    run cancel - Cancel a workflow run [intent=reverse_etl availability=implemented write=cancel_workflow_run]; risk: Cancels a GitHub Actions workflow run.; approval: Reverse ETL writes require plan, preview, approval, execute.
    run delete - Delete a workflow run [intent=reverse_etl availability=implemented write=delete_workflow_run]; risk: Deletes workflow run data.; approval: Reverse ETL writes require plan, preview, approval, execute.
    run download - Download workflow artifacts [intent=local_workflow availability=unsupported_local]
    run watch - Watch a workflow run [intent=local_workflow availability=unsupported_local]
    run logs view - Read /repos/{owner}/{repo}/actions/jobs/{job_id}/logs [intent=direct_read availability=implemented operation=github.actions_jobs_job_id_logs]
    run logs view-2 - Download /repos/{owner}/{repo}/actions/runs/{run_id}/attempts/{attempt_number}/logs [intent=direct_read availability=implemented operation=github.actions_runs_run_id_attempts_attempt_number_logs]
    run logs view-3 - Download /repos/{owner}/{repo}/actions/runs/{run_id}/logs [intent=direct_read availability=implemented operation=github.actions_runs_run_id_logs2]
    cache list - List GitHub Actions caches [intent=direct_read availability=unsupported_api]
    cache delete - Delete GitHub Actions caches [intent=direct_write availability=unsafe_or_disallowed]
  Collaboration Commands
    label list - List labels [intent=etl availability=implemented stream=labels]
    label create - Create a label [intent=reverse_etl availability=implemented write=create_label]; risk: Creates a repository label.; approval: Reverse ETL writes require plan, preview, approval, execute.
    label edit - Edit a label [intent=reverse_etl availability=implemented write=update_label]; risk: Mutates a repository label.; approval: Reverse ETL writes require plan, preview, approval, execute.
    label delete - Delete a label [intent=reverse_etl availability=implemented write=delete_label]; risk: Deletes a repository label.; approval: Reverse ETL writes require plan, preview, approval, execute.
    label clone - Clone labels between repositories [intent=direct_write availability=unsupported_api]
    ruleset list - List repository rulesets [intent=etl availability=implemented stream=repo_rulesets]
    ruleset view - View repository ruleset details [intent=etl availability=partial stream=repo_rulesets]
    ruleset check - Check rules that apply to a branch [intent=direct_read availability=planned]
    org list - List organizations for the authenticated user [intent=direct_read availability=unsupported_api]
    project list - List projects [intent=etl availability=implemented stream=projects]
    project create - Create a project [intent=direct_write availability=planned]
    project item-list - List project items [intent=etl availability=implemented stream=project_items]
    discussion list - List discussions [intent=etl availability=implemented stream=discussions]
    discussion view - View a discussion [intent=etl availability=implemented stream=discussion]
    discussion create - Create a discussion [intent=direct_write availability=planned]
  Security And Configuration Commands
    secret list - List repository secrets [intent=direct_read availability=unsupported_api]
    secret set - Create or update a secret [intent=direct_write availability=unsafe_or_disallowed]
    secret delete - Delete a secret [intent=direct_write availability=unsafe_or_disallowed]
    secret delete-2 - DELETE /repos/{owner}/{repo}/actions/secrets/{secret_name} [intent=reverse_etl availability=implemented write=actions_secrets_secret_name]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
    secret set-2 - PUT /repos/{owner}/{repo}/actions/secrets/{secret_name} [intent=reverse_etl availability=implemented write=actions_secrets_secret_name3]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
    secret delete-3 - DELETE /repos/{owner}/{repo}/codespaces/secrets/{secret_name} [intent=reverse_etl availability=implemented write=codespaces_secrets_secret_name]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
    secret set-3 - PUT /repos/{owner}/{repo}/codespaces/secrets/{secret_name} [intent=reverse_etl availability=implemented write=codespaces_secrets_secret_name3]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
    secret delete-4 - DELETE /repos/{owner}/{repo}/dependabot/secrets/{secret_name} [intent=reverse_etl availability=implemented write=dependabot_secrets_secret_name]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
    secret set-4 - PUT /repos/{owner}/{repo}/dependabot/secrets/{secret_name} [intent=reverse_etl availability=implemented write=dependabot_secrets_secret_name3]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
    secret delete-5 - DELETE /repos/{owner}/{repo}/environments/{environment_name}/secrets/{secret_name} [intent=reverse_etl availability=implemented write=environments_environment_name_secrets_secret_name]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
    secret set-5 - PUT /repos/{owner}/{repo}/environments/{environment_name}/secrets/{secret_name} [intent=reverse_etl availability=implemented write=environments_environment_name_secrets_secret_name3]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
    variable list - List repository variables [intent=direct_read availability=unsupported_api]
    variable get - Get a repository variable [intent=direct_read availability=unsupported_api]
    variable set - Create or update a repository variable [intent=direct_write availability=unsupported_api]
    variable delete - Delete a repository variable [intent=direct_write availability=unsupported_api]
    variable create - POST /repos/{owner}/{repo}/actions/variables [intent=reverse_etl availability=implemented write=actions_variables2]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
    variable delete-2 - DELETE /repos/{owner}/{repo}/actions/variables/{name} [intent=reverse_etl availability=implemented write=actions_variables_name]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
    variable update - PATCH /repos/{owner}/{repo}/actions/variables/{name} [intent=reverse_etl availability=implemented write=actions_variables_name3]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
    variable create-2 - POST /repos/{owner}/{repo}/environments/{environment_name}/variables [intent=reverse_etl availability=implemented write=environments_environment_name_variables2]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
    variable delete-3 - DELETE /repos/{owner}/{repo}/environments/{environment_name}/variables/{name} [intent=reverse_etl availability=implemented write=environments_environment_name_variables_name]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
    variable update-2 - PATCH /repos/{owner}/{repo}/environments/{environment_name}/variables/{name} [intent=reverse_etl availability=implemented write=environments_environment_name_variables_name3]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
    gpg-key list - List GPG keys [intent=direct_read availability=unsupported_api]
    ssh-key list - List SSH keys [intent=direct_read availability=unsupported_api]
    attestation verify - Verify artifact attestations [intent=direct_read availability=unsupported_local]
  Local Workflow Commands
    auth login - Authenticate gh [intent=auth availability=unsupported_local]
    auth status - View gh authentication status [intent=auth availability=unsupported_local]
    auth token - Print gh token [intent=auth availability=unsafe_or_disallowed]
    config get - Read gh local config [intent=config availability=unsupported_local]
    config set - Write gh local config [intent=config availability=unsupported_local]
    browse - Open GitHub in a browser [intent=local_workflow availability=unsupported_local]
    alias list - List gh aliases [intent=local_workflow availability=unsupported_local]
    extension list - List gh extensions [intent=local_workflow availability=unsupported_local]
    completion - Generate shell completion [intent=local_workflow availability=unsupported_local]
  Additional Commands
    api - Make an authenticated GitHub API request [intent=raw_api availability=unsafe_or_disallowed]
    search repos - Search repositories [intent=direct_read availability=planned]
    search issues - Search issues [intent=direct_read availability=planned]
    search prs - Search pull requests [intent=direct_read availability=planned]
    search code - Search code [intent=direct_read availability=planned]
    search commits - Search commits [intent=direct_read availability=planned]
    gist list - List gists [intent=direct_read availability=unsupported_api]
    gist create - Create a gist [intent=direct_write availability=unsupported_api]
    codespace list - List codespaces [intent=direct_read availability=unsupported_api]
    codespace create - Create a codespace [intent=direct_write availability=unsupported_api]
    codespace ssh - SSH into a codespace [intent=local_workflow availability=unsupported_local]
    status - Print GitHub status [intent=direct_read availability=planned]
    copilot - Use GitHub Copilot CLI [intent=local_workflow availability=unsupported_local]
    copilot configuration view - Read /repos/{owner}/{repo}/copilot/cloud-agent/configuration [intent=direct_read availability=implemented operation=github.copilot_cloud_agent_configuration]
    skill list - List GitHub Skills [intent=direct_read availability=unsupported_api]
    agent-task list - List GitHub agent tasks [intent=direct_read availability=unsupported_api]
  artifact download - Download /repos/{owner}/{repo}/actions/artifacts/{artifact_id}/{archive_format} [intent=direct_read availability=implemented operation=github.actions_artifacts_artifact_id_archive_format]
  actions retention-limit view - Read /repos/{owner}/{repo}/actions/cache/retention-limit [intent=direct_read availability=implemented operation=github.actions_cache_retention_limit]
  actions retention-limit set - PUT /repos/{owner}/{repo}/actions/cache/retention-limit [intent=reverse_etl availability=implemented write=actions_cache_retention_limit2]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions storage-limit view - Read /repos/{owner}/{repo}/actions/cache/storage-limit [intent=direct_read availability=implemented operation=github.actions_cache_storage_limit]
  actions storage-limit set - PUT /repos/{owner}/{repo}/actions/cache/storage-limit [intent=reverse_etl availability=implemented write=actions_cache_storage_limit2]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions usage view - Read /repos/{owner}/{repo}/actions/cache/usage [intent=direct_read availability=implemented operation=github.actions_cache_usage]
  actions caches delete - DELETE /repos/{owner}/{repo}/actions/caches [intent=reverse_etl availability=implemented write=actions_caches]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions caches view - Read /repos/{owner}/{repo}/actions/caches [intent=direct_read availability=implemented operation=github.actions_caches2]
  actions caches delete-2 - DELETE /repos/{owner}/{repo}/actions/caches/{cache_id} [intent=reverse_etl availability=implemented write=actions_caches_cache_id]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions concurrency_groups view - Read /repos/{owner}/{repo}/actions/concurrency_groups [intent=direct_read availability=implemented operation=github.actions_concurrency_groups]
  actions concurrency_groups view-2 - Read /repos/{owner}/{repo}/actions/concurrency_groups/{concurrency_group_name} [intent=direct_read availability=implemented operation=github.actions_concurrency_groups_concurrency_group_name]
  actions jobs view - Read /repos/{owner}/{repo}/actions/jobs/{job_id} [intent=direct_read availability=implemented operation=github.actions_jobs_job_id]
  actions rerun create - POST /repos/{owner}/{repo}/actions/jobs/{job_id}/rerun [intent=reverse_etl availability=implemented write=actions_jobs_job_id_rerun]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions sub view - Read /repos/{owner}/{repo}/actions/oidc/customization/sub [intent=direct_read availability=implemented operation=github.actions_oidc_customization_sub]
  actions sub set - PUT /repos/{owner}/{repo}/actions/oidc/customization/sub [intent=reverse_etl availability=implemented write=actions_oidc_customization_sub2]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions organization-secrets view - Read /repos/{owner}/{repo}/actions/organization-secrets [intent=direct_read availability=implemented operation=github.actions_organization_secrets]
  actions organization-variables view - Read /repos/{owner}/{repo}/actions/organization-variables [intent=direct_read availability=implemented operation=github.actions_organization_variables]
  actions permissions view - Read /repos/{owner}/{repo}/actions/permissions [intent=direct_read availability=implemented operation=github.actions_permissions]
  actions permissions set - PUT /repos/{owner}/{repo}/actions/permissions [intent=reverse_etl availability=implemented write=actions_permissions2]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions access view - Read /repos/{owner}/{repo}/actions/permissions/access [intent=direct_read availability=implemented operation=github.actions_permissions_access]
  actions permissions set-2 - PUT /repos/{owner}/{repo}/actions/permissions/access [intent=reverse_etl availability=implemented write=actions_permissions_access2]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions artifact-and-log-retention view - Read /repos/{owner}/{repo}/actions/permissions/artifact-and-log-retention [intent=direct_read availability=implemented operation=github.actions_permissions_artifact_and_log_retention]
  actions permissions set-3 - PUT /repos/{owner}/{repo}/actions/permissions/artifact-and-log-retention [intent=reverse_etl availability=implemented write=actions_permissions_artifact_and_log_retention2]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions fork-pr-contributor-approval view - Read /repos/{owner}/{repo}/actions/permissions/fork-pr-contributor-approval [intent=direct_read availability=implemented operation=github.actions_permissions_fork_pr_contributor_approval]
  actions permissions set-4 - PUT /repos/{owner}/{repo}/actions/permissions/fork-pr-contributor-approval [intent=reverse_etl availability=implemented write=actions_permissions_fork_pr_contributor_approval2]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions fork-pr-workflows-private-repos view - Read /repos/{owner}/{repo}/actions/permissions/fork-pr-workflows-private-repos [intent=direct_read availability=implemented operation=github.actions_permissions_fork_pr_workflows_private_repos]
  actions permissions set-5 - PUT /repos/{owner}/{repo}/actions/permissions/fork-pr-workflows-private-repos [intent=reverse_etl availability=implemented write=actions_permissions_fork_pr_workflows_private_repos2]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions selected-actions view - Read /repos/{owner}/{repo}/actions/permissions/selected-actions [intent=direct_read availability=implemented operation=github.actions_permissions_selected_actions]
  actions permissions set-6 - PUT /repos/{owner}/{repo}/actions/permissions/selected-actions [intent=reverse_etl availability=implemented write=actions_permissions_selected_actions2]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions workflow view - Read /repos/{owner}/{repo}/actions/permissions/workflow [intent=direct_read availability=implemented operation=github.actions_permissions_workflow]
  actions permissions set-7 - PUT /repos/{owner}/{repo}/actions/permissions/workflow [intent=reverse_etl availability=implemented write=actions_permissions_workflow2]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions runners view - Read /repos/{owner}/{repo}/actions/runners [intent=direct_read availability=implemented operation=github.actions_runners]
  actions downloads view - Read /repos/{owner}/{repo}/actions/runners/downloads [intent=direct_read availability=implemented operation=github.actions_runners_downloads]
  actions generate-jitconfig create - POST /repos/{owner}/{repo}/actions/runners/generate-jitconfig [intent=reverse_etl availability=implemented write=actions_runners_generate_jitconfig]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions registration-token create - POST /repos/{owner}/{repo}/actions/runners/registration-token [intent=reverse_etl availability=implemented write=actions_runners_registration_token]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions remove-token create - POST /repos/{owner}/{repo}/actions/runners/remove-token [intent=reverse_etl availability=implemented write=actions_runners_remove_token]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions runners delete - DELETE /repos/{owner}/{repo}/actions/runners/{runner_id} [intent=reverse_etl availability=implemented write=actions_runners_runner_id]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions runners view-2 - Read /repos/{owner}/{repo}/actions/runners/{runner_id} [intent=direct_read availability=implemented operation=github.actions_runners_runner_id2]
  actions labels delete - DELETE /repos/{owner}/{repo}/actions/runners/{runner_id}/labels [intent=reverse_etl availability=implemented write=actions_runners_runner_id_labels]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions labels view - Read /repos/{owner}/{repo}/actions/runners/{runner_id}/labels [intent=direct_read availability=implemented operation=github.actions_runners_runner_id_labels2]
  actions labels create - POST /repos/{owner}/{repo}/actions/runners/{runner_id}/labels [intent=reverse_etl availability=implemented write=actions_runners_runner_id_labels3]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions labels set - PUT /repos/{owner}/{repo}/actions/runners/{runner_id}/labels [intent=reverse_etl availability=implemented write=actions_runners_runner_id_labels4]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions labels delete-2 - DELETE /repos/{owner}/{repo}/actions/runners/{runner_id}/labels/{name} [intent=reverse_etl availability=implemented write=actions_runners_runner_id_labels_name]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions approvals view - Read /repos/{owner}/{repo}/actions/runs/{run_id}/approvals [intent=direct_read availability=implemented operation=github.actions_runs_run_id_approvals]
  actions approve create - POST /repos/{owner}/{repo}/actions/runs/{run_id}/approve [intent=reverse_etl availability=implemented write=actions_runs_run_id_approve]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions attempts view - Read /repos/{owner}/{repo}/actions/runs/{run_id}/attempts/{attempt_number} [intent=direct_read availability=implemented operation=github.actions_runs_run_id_attempts_attempt_number]
  actions jobs view-2 - Read /repos/{owner}/{repo}/actions/runs/{run_id}/attempts/{attempt_number}/jobs [intent=direct_read availability=implemented operation=github.actions_runs_run_id_attempts_attempt_number_jobs]
  actions concurrency_groups view-3 - Read /repos/{owner}/{repo}/actions/runs/{run_id}/concurrency_groups [intent=direct_read availability=implemented operation=github.actions_runs_run_id_concurrency_groups]
  actions deployment_protection_rule create - POST /repos/{owner}/{repo}/actions/runs/{run_id}/deployment_protection_rule [intent=reverse_etl availability=implemented write=actions_runs_run_id_deployment_protection_rule]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions jobs view-3 - Read /repos/{owner}/{repo}/actions/runs/{run_id}/jobs [intent=direct_read availability=implemented operation=github.actions_runs_run_id_jobs]
  actions logs delete - DELETE /repos/{owner}/{repo}/actions/runs/{run_id}/logs [intent=reverse_etl availability=implemented write=actions_runs_run_id_logs]; risk: critical; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions pending_deployments view - Read /repos/{owner}/{repo}/actions/runs/{run_id}/pending_deployments [intent=direct_read availability=implemented operation=github.actions_runs_run_id_pending_deployments]
  actions pending_deployments create - POST /repos/{owner}/{repo}/actions/runs/{run_id}/pending_deployments [intent=reverse_etl availability=implemented write=actions_runs_run_id_pending_deployments2]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions timing view - Read /repos/{owner}/{repo}/actions/runs/{run_id}/timing [intent=direct_read availability=implemented operation=github.actions_runs_run_id_timing]
  actions secrets view - Read /repos/{owner}/{repo}/actions/secrets [intent=direct_read availability=implemented operation=github.actions_secrets]
  actions public-key view - Read /repos/{owner}/{repo}/actions/secrets/public-key [intent=direct_read availability=implemented operation=github.actions_secrets_public_key]
  actions secrets view-2 - Read /repos/{owner}/{repo}/actions/secrets/{secret_name} [intent=direct_read availability=implemented operation=github.actions_secrets_secret_name2]
  actions variables view - Read /repos/{owner}/{repo}/actions/variables [intent=direct_read availability=implemented operation=github.actions_variables]
  actions variables view-2 - Read /repos/{owner}/{repo}/actions/variables/{name} [intent=direct_read availability=implemented operation=github.actions_variables_name2]
  actions disable set - PUT /repos/{owner}/{repo}/actions/workflows/{workflow_id}/disable [intent=reverse_etl availability=implemented write=actions_workflows_workflow_id_disable]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions enable set - PUT /repos/{owner}/{repo}/actions/workflows/{workflow_id}/enable [intent=reverse_etl availability=implemented write=actions_workflows_workflow_id_enable]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  actions timing view-2 - Read /repos/{owner}/{repo}/actions/workflows/{workflow_id}/timing [intent=direct_read availability=implemented operation=github.actions_workflows_workflow_id_timing]
  assignees view - Read /repos/{owner}/{repo}/assignees/{assignee} [intent=direct_read availability=implemented operation=github.assignees_assignee]
  attestations create - POST /repos/{owner}/{repo}/attestations [intent=reverse_etl availability=implemented write=attestations]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  attestations view - Read /repos/{owner}/{repo}/attestations/{subject_digest} [intent=direct_read availability=implemented operation=github.attestations_subject_digest]
  autolinks create - POST /repos/{owner}/{repo}/autolinks [intent=reverse_etl availability=implemented write=autolinks]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  autolinks delete - DELETE /repos/{owner}/{repo}/autolinks/{autolink_id} [intent=reverse_etl availability=implemented write=autolinks_autolink_id]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  automated-security-fixes delete - DELETE /repos/{owner}/{repo}/automated-security-fixes [intent=reverse_etl availability=implemented write=automated_security_fixes]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  automated-security-fixes view - Read /repos/{owner}/{repo}/automated-security-fixes [intent=direct_read availability=implemented operation=github.automated_security_fixes2]
  automated-security-fixes set - PUT /repos/{owner}/{repo}/automated-security-fixes [intent=reverse_etl availability=implemented write=automated_security_fixes3]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches protection delete - DELETE /repos/{owner}/{repo}/branches/{branch}/protection [intent=reverse_etl availability=implemented write=branches_branch_protection]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches protection view - Read /repos/{owner}/{repo}/branches/{branch}/protection [intent=direct_read availability=implemented operation=github.branches_branch_protection2]
  branches protection set - PUT /repos/{owner}/{repo}/branches/{branch}/protection [intent=reverse_etl availability=implemented write=branches_branch_protection3]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches enforce_admins delete - DELETE /repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins [intent=reverse_etl availability=implemented write=branches_branch_protection_enforce_admins]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches enforce_admins view - Read /repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins [intent=direct_read availability=implemented operation=github.branches_branch_protection_enforce_admins2]
  branches enforce_admins create - POST /repos/{owner}/{repo}/branches/{branch}/protection/enforce_admins [intent=reverse_etl availability=implemented write=branches_branch_protection_enforce_admins3]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches required_pull_request_reviews delete - DELETE /repos/{owner}/{repo}/branches/{branch}/protection/required_pull_request_reviews [intent=reverse_etl availability=implemented write=branches_branch_protection_required_pull_request_reviews]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches required_pull_request_reviews view - Read /repos/{owner}/{repo}/branches/{branch}/protection/required_pull_request_reviews [intent=direct_read availability=implemented operation=github.branches_branch_protection_required_pull_request_reviews2]
  branches required_pull_request_reviews update - PATCH /repos/{owner}/{repo}/branches/{branch}/protection/required_pull_request_reviews [intent=reverse_etl availability=implemented write=branches_branch_protection_required_pull_request_reviews3]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches required_signatures delete - DELETE /repos/{owner}/{repo}/branches/{branch}/protection/required_signatures [intent=reverse_etl availability=implemented write=branches_branch_protection_required_signatures]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches required_signatures view - Read /repos/{owner}/{repo}/branches/{branch}/protection/required_signatures [intent=direct_read availability=implemented operation=github.branches_branch_protection_required_signatures2]
  branches required_signatures create - POST /repos/{owner}/{repo}/branches/{branch}/protection/required_signatures [intent=reverse_etl availability=implemented write=branches_branch_protection_required_signatures3]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches required_status_checks delete - DELETE /repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks [intent=reverse_etl availability=implemented write=branches_branch_protection_required_status_checks]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches required_status_checks view - Read /repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks [intent=direct_read availability=implemented operation=github.branches_branch_protection_required_status_checks2]
  branches required_status_checks update - PATCH /repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks [intent=reverse_etl availability=implemented write=branches_branch_protection_required_status_checks3]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches contexts delete - DELETE /repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts [intent=reverse_etl availability=implemented write=branches_branch_protection_required_status_checks_contexts]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches contexts view - Read /repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts [intent=direct_read availability=implemented operation=github.branches_branch_protection_required_status_checks_contexts2]
  branches contexts create - POST /repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts [intent=reverse_etl availability=implemented write=branches_branch_protection_required_status_checks_contexts3]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches contexts set - PUT /repos/{owner}/{repo}/branches/{branch}/protection/required_status_checks/contexts [intent=reverse_etl availability=implemented write=branches_branch_protection_required_status_checks_contexts4]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches restrictions delete - DELETE /repos/{owner}/{repo}/branches/{branch}/protection/restrictions [intent=reverse_etl availability=implemented write=branches_branch_protection_restrictions]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches restrictions view - Read /repos/{owner}/{repo}/branches/{branch}/protection/restrictions [intent=direct_read availability=implemented operation=github.branches_branch_protection_restrictions2]
  branches apps delete - DELETE /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/apps [intent=reverse_etl availability=implemented write=branches_branch_protection_restrictions_apps]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches apps view - Read /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/apps [intent=direct_read availability=implemented operation=github.branches_branch_protection_restrictions_apps2]
  branches apps create - POST /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/apps [intent=reverse_etl availability=implemented write=branches_branch_protection_restrictions_apps3]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches apps set - PUT /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/apps [intent=reverse_etl availability=implemented write=branches_branch_protection_restrictions_apps4]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches teams delete - DELETE /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams [intent=reverse_etl availability=implemented write=branches_branch_protection_restrictions_teams]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches teams view - Read /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams [intent=direct_read availability=implemented operation=github.branches_branch_protection_restrictions_teams2]
  branches teams create - POST /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams [intent=reverse_etl availability=implemented write=branches_branch_protection_restrictions_teams3]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches teams set - PUT /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/teams [intent=reverse_etl availability=implemented write=branches_branch_protection_restrictions_teams4]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches users delete - DELETE /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users [intent=reverse_etl availability=implemented write=branches_branch_protection_restrictions_users]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches users view - Read /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users [intent=direct_read availability=implemented operation=github.branches_branch_protection_restrictions_users2]
  branches users create - POST /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users [intent=reverse_etl availability=implemented write=branches_branch_protection_restrictions_users3]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches users set - PUT /repos/{owner}/{repo}/branches/{branch}/protection/restrictions/users [intent=reverse_etl availability=implemented write=branches_branch_protection_restrictions_users4]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  branches rename create - POST /repos/{owner}/{repo}/branches/{branch}/rename [intent=reverse_etl availability=implemented write=branches_branch_rename]; risk: critical; approval: Reverse ETL writes require plan, preview, approval, execute.
  check-runs create - POST /repos/{owner}/{repo}/check-runs [intent=reverse_etl availability=implemented write=check_runs]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  check-runs view - Read /repos/{owner}/{repo}/check-runs/{check_run_id} [intent=direct_read availability=implemented operation=github.check_runs_check_run_id]
  check-runs update - PATCH /repos/{owner}/{repo}/check-runs/{check_run_id} [intent=reverse_etl availability=implemented write=check_runs_check_run_id2]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  check-runs annotations view - Read /repos/{owner}/{repo}/check-runs/{check_run_id}/annotations [intent=direct_read availability=implemented operation=github.check_runs_check_run_id_annotations]
  check-runs rerequest create - POST /repos/{owner}/{repo}/check-runs/{check_run_id}/rerequest [intent=reverse_etl availability=implemented write=check_runs_check_run_id_rerequest]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  check-suites create - POST /repos/{owner}/{repo}/check-suites [intent=reverse_etl availability=implemented write=check_suites]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  check-suites preferences update - PATCH /repos/{owner}/{repo}/check-suites/preferences [intent=reverse_etl availability=implemented write=check_suites_preferences]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  check-suites view - Read /repos/{owner}/{repo}/check-suites/{check_suite_id} [intent=direct_read availability=implemented operation=github.check_suites_check_suite_id]
  check-suites check-runs view - Read /repos/{owner}/{repo}/check-suites/{check_suite_id}/check-runs [intent=direct_read availability=implemented operation=github.check_suites_check_suite_id_check_runs]
  check-suites rerequest create - POST /repos/{owner}/{repo}/check-suites/{check_suite_id}/rerequest [intent=reverse_etl availability=implemented write=check_suites_check_suite_id_rerequest]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  code-quality findings view - Read /repos/{owner}/{repo}/code-quality/findings [intent=direct_read availability=implemented operation=github.code_quality_findings]
  code-quality findings view-2 - Read /repos/{owner}/{repo}/code-quality/findings/{finding_number} [intent=direct_read availability=implemented operation=github.code_quality_findings_finding_number]
  code-quality setup view - Read /repos/{owner}/{repo}/code-quality/setup [intent=direct_read availability=implemented operation=github.code_quality_setup]
  code-quality setup update - PATCH /repos/{owner}/{repo}/code-quality/setup [intent=reverse_etl availability=implemented write=code_quality_setup2]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  code-scanning autofix view - Read /repos/{owner}/{repo}/code-scanning/alerts/{alert_number}/autofix [intent=direct_read availability=implemented operation=github.code_scanning_alerts_alert_number_autofix]
  code-scanning autofix create - POST /repos/{owner}/{repo}/code-scanning/alerts/{alert_number}/autofix [intent=reverse_etl availability=implemented write=code_scanning_alerts_alert_number_autofix2]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  code-scanning commits create - POST /repos/{owner}/{repo}/code-scanning/alerts/{alert_number}/autofix/commits [intent=reverse_etl availability=implemented write=code_scanning_alerts_alert_number_autofix_commits]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  code-scanning instances view - Read /repos/{owner}/{repo}/code-scanning/alerts/{alert_number}/instances [intent=direct_read availability=implemented operation=github.code_scanning_alerts_alert_number_instances]
  code-scanning analyses view - Read /repos/{owner}/{repo}/code-scanning/analyses [intent=direct_read availability=implemented operation=github.code_scanning_analyses]
  code-scanning analyses delete - DELETE /repos/{owner}/{repo}/code-scanning/analyses/{analysis_id} [intent=reverse_etl availability=implemented write=code_scanning_analyses_analysis_id]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  code-scanning analyses view-2 - Read /repos/{owner}/{repo}/code-scanning/analyses/{analysis_id} [intent=direct_read availability=implemented operation=github.code_scanning_analyses_analysis_id2]
  code-scanning databases view - Read /repos/{owner}/{repo}/code-scanning/codeql/databases [intent=direct_read availability=implemented operation=github.code_scanning_codeql_databases]
  code-scanning databases delete - DELETE /repos/{owner}/{repo}/code-scanning/codeql/databases/{language} [intent=reverse_etl availability=implemented write=code_scanning_codeql_databases_language]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  code-scanning databases view-2 - Read /repos/{owner}/{repo}/code-scanning/codeql/databases/{language} [intent=direct_read availability=implemented operation=github.code_scanning_codeql_databases_language2]
  code-scanning variant-analyses create - POST /repos/{owner}/{repo}/code-scanning/codeql/variant-analyses [intent=reverse_etl availability=implemented write=code_scanning_codeql_variant_analyses]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  code-scanning variant-analyses view - Read /repos/{owner}/{repo}/code-scanning/codeql/variant-analyses/{codeql_variant_analysis_id} [intent=direct_read availability=implemented operation=github.code_scanning_codeql_variant_analyses_codeql_variant_analysis_id]
  code-scanning repos view - Read /repos/{owner}/{repo}/code-scanning/codeql/variant-analyses/{codeql_variant_analysis_id}/repos/{repo_owner}/{repo_name} [intent=direct_read availability=implemented operation=github.code_scanning_codeql_variant_analyses_codeql_variant_analysis_id_repos_repo_owner_repo_name]
  code-scanning default-setup view - Read /repos/{owner}/{repo}/code-scanning/default-setup [intent=direct_read availability=implemented operation=github.code_scanning_default_setup]
  code-scanning default-setup update - PATCH /repos/{owner}/{repo}/code-scanning/default-setup [intent=reverse_etl availability=implemented write=code_scanning_default_setup2]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  code-sanning upload - POST /repos/{owner}/{repo}/code-scanning/sarifs [intent=reverse_etl availability=implemented write=code_scanning_sarifs]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  code-scanning sarifs view - Read /repos/{owner}/{repo}/code-scanning/sarifs/{sarif_id} [intent=direct_read availability=implemented operation=github.code_scanning_sarifs_sarif_id]
  code-security-configuration view - Read /repos/{owner}/{repo}/code-security-configuration [intent=direct_read availability=implemented operation=github.code_security_configuration]
  codeowners errors view - Read /repos/{owner}/{repo}/codeowners/errors [intent=direct_read availability=implemented operation=github.codeowners_errors]
  codespaces view - Read /repos/{owner}/{repo}/codespaces [intent=direct_read availability=implemented operation=github.codespaces]
  codespaces create - POST /repos/{owner}/{repo}/codespaces [intent=reverse_etl availability=implemented write=codespaces2]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  codespaces devcontainers view - Read /repos/{owner}/{repo}/codespaces/devcontainers [intent=direct_read availability=implemented operation=github.codespaces_devcontainers]
  codespaces machines view - Read /repos/{owner}/{repo}/codespaces/machines [intent=direct_read availability=implemented operation=github.codespaces_machines]
  codespaces new view - Read /repos/{owner}/{repo}/codespaces/new [intent=direct_read availability=implemented operation=github.codespaces_new]
  codespaces permissions_check view - Read /repos/{owner}/{repo}/codespaces/permissions_check [intent=direct_read availability=implemented operation=github.codespaces_permissions_check]
  codespaces secrets view - Read /repos/{owner}/{repo}/codespaces/secrets [intent=direct_read availability=implemented operation=github.codespaces_secrets]
  codespaces public-key view - Read /repos/{owner}/{repo}/codespaces/secrets/public-key [intent=direct_read availability=implemented operation=github.codespaces_secrets_public_key]
  codespaces secrets view-2 - Read /repos/{owner}/{repo}/codespaces/secrets/{secret_name} [intent=direct_read availability=implemented operation=github.codespaces_secrets_secret_name2]
  comments reactions view - Read /repos/{owner}/{repo}/comments/{comment_id}/reactions [intent=direct_read availability=implemented operation=github.comments_comment_id_reactions]
  comments reactions create - POST /repos/{owner}/{repo}/comments/{comment_id}/reactions [intent=reverse_etl availability=implemented write=comments_comment_id_reactions2]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  comments reactions delete - DELETE /repos/{owner}/{repo}/comments/{comment_id}/reactions/{reaction_id} [intent=reverse_etl availability=implemented write=comments_comment_id_reactions_reaction_id]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  commits branches-where-head view - Read /repos/{owner}/{repo}/commits/{commit_sha}/branches-where-head [intent=direct_read availability=implemented operation=github.commits_commit_sha_branches_where_head]
  commits pulls view - Read /repos/{owner}/{repo}/commits/{commit_sha}/pulls [intent=direct_read availability=implemented operation=github.commits_commit_sha_pulls]
  commits check-runs view - Read /repos/{owner}/{repo}/commits/{ref}/check-runs [intent=direct_read availability=implemented operation=github.commits_ref_check_runs]
  commits check-suites view - Read /repos/{owner}/{repo}/commits/{ref}/check-suites [intent=direct_read availability=implemented operation=github.commits_ref_check_suites]
  commits status view - Read /repos/{owner}/{repo}/commits/{ref}/status [intent=direct_read availability=implemented operation=github.commits_ref_status]
  commits statuses view - Read /repos/{owner}/{repo}/commits/{ref}/statuses [intent=direct_read availability=implemented operation=github.commits_ref_statuses]
  community profile view - Read /repos/{owner}/{repo}/community/profile [intent=direct_read availability=implemented operation=github.community_profile]
  compare view - Read /repos/{owner}/{repo}/compare/{basehead} [intent=direct_read availability=implemented operation=github.compare_basehead]
  dependabot secrets view - Read /repos/{owner}/{repo}/dependabot/secrets [intent=direct_read availability=implemented operation=github.dependabot_secrets]
  dependabot public-key view - Read /repos/{owner}/{repo}/dependabot/secrets/public-key [intent=direct_read availability=implemented operation=github.dependabot_secrets_public_key]
  dependabot secrets view-2 - Read /repos/{owner}/{repo}/dependabot/secrets/{secret_name} [intent=direct_read availability=implemented operation=github.dependabot_secrets_secret_name2]
  dependency-graph compare view - Read /repos/{owner}/{repo}/dependency-graph/compare/{basehead} [intent=direct_read availability=implemented operation=github.dependency_graph_compare_basehead]
  dependency-graph snapshots create - POST /repos/{owner}/{repo}/dependency-graph/snapshots [intent=reverse_etl availability=implemented write=dependency_graph_snapshots]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  deployments delete - DELETE /repos/{owner}/{repo}/deployments/{deployment_id} [intent=reverse_etl availability=implemented write=deployments_deployment_id]; risk: critical; approval: Reverse ETL writes require plan, preview, approval, execute.
  deployments statuses view - Read /repos/{owner}/{repo}/deployments/{deployment_id}/statuses [intent=direct_read availability=implemented operation=github.deployments_deployment_id_statuses]
  deployments statuses create - POST /repos/{owner}/{repo}/deployments/{deployment_id}/statuses [intent=reverse_etl availability=implemented write=deployments_deployment_id_statuses2]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  dispatches create - POST /repos/{owner}/{repo}/dispatches [intent=reverse_etl availability=implemented write=dispatches]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  environments deployment-branch-policies view - Read /repos/{owner}/{repo}/environments/{environment_name}/deployment-branch-policies [intent=direct_read availability=implemented operation=github.environments_environment_name_deployment_branch_policies]
  environments deployment-branch-policies create - POST /repos/{owner}/{repo}/environments/{environment_name}/deployment-branch-policies [intent=reverse_etl availability=implemented write=environments_environment_name_deployment_branch_policies2]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  environments deployment-branch-policies delete - DELETE /repos/{owner}/{repo}/environments/{environment_name}/deployment-branch-policies/{branch_policy_id} [intent=reverse_etl availability=implemented write=environments_environment_name_deployment_branch_policies_branch_policy_id]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  environments deployment-branch-policies view-2 - Read /repos/{owner}/{repo}/environments/{environment_name}/deployment-branch-policies/{branch_policy_id} [intent=direct_read availability=implemented operation=github.environments_environment_name_deployment_branch_policies_branch_policy_id2]
  environments deployment-branch-policies set - PUT /repos/{owner}/{repo}/environments/{environment_name}/deployment-branch-policies/{branch_policy_id} [intent=reverse_etl availability=implemented write=environments_environment_name_deployment_branch_policies_branch_policy_id3]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  environments deployment_protection_rules view - Read /repos/{owner}/{repo}/environments/{environment_name}/deployment_protection_rules [intent=direct_read availability=implemented operation=github.environments_environment_name_deployment_protection_rules]
  environments deployment_protection_rules create - POST /repos/{owner}/{repo}/environments/{environment_name}/deployment_protection_rules [intent=reverse_etl availability=implemented write=environments_environment_name_deployment_protection_rules2]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  environments apps view - Read /repos/{owner}/{repo}/environments/{environment_name}/deployment_protection_rules/apps [intent=direct_read availability=implemented operation=github.environments_environment_name_deployment_protection_rules_apps]
  environments deployment_protection_rules delete - DELETE /repos/{owner}/{repo}/environments/{environment_name}/deployment_protection_rules/{protection_rule_id} [intent=reverse_etl availability=implemented write=environments_environment_name_deployment_protection_rules_protection_rule_id]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  environments deployment_protection_rules view-2 - Read /repos/{owner}/{repo}/environments/{environment_name}/deployment_protection_rules/{protection_rule_id} [intent=direct_read availability=implemented operation=github.environments_environment_name_deployment_protection_rules_protection_rule_id2]
  environments secrets view - Read /repos/{owner}/{repo}/environments/{environment_name}/secrets [intent=direct_read availability=implemented operation=github.environments_environment_name_secrets]
  environments public-key view - Read /repos/{owner}/{repo}/environments/{environment_name}/secrets/public-key [intent=direct_read availability=implemented operation=github.environments_environment_name_secrets_public_key]
  environments secrets view-2 - Read /repos/{owner}/{repo}/environments/{environment_name}/secrets/{secret_name} [intent=direct_read availability=implemented operation=github.environments_environment_name_secrets_secret_name2]
  environments variables view - Read /repos/{owner}/{repo}/environments/{environment_name}/variables [intent=direct_read availability=implemented operation=github.environments_environment_name_variables]
  environments variables view-2 - Read /repos/{owner}/{repo}/environments/{environment_name}/variables/{name} [intent=direct_read availability=implemented operation=github.environments_environment_name_variables_name2]
  git blobs create - POST /repos/{owner}/{repo}/git/blobs [intent=reverse_etl availability=implemented write=git_blobs]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  git blobs view - Read /repos/{owner}/{repo}/git/blobs/{file_sha} [intent=direct_read availability=implemented operation=github.git_blobs_file_sha]
  git commits create - POST /repos/{owner}/{repo}/git/commits [intent=reverse_etl availability=implemented write=git_commits]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  git ref view - Read /repos/{owner}/{repo}/git/ref/{ref} [intent=direct_read availability=implemented operation=github.git_ref_ref]
  git tags view - Read /repos/{owner}/{repo}/git/tags/{tag_sha} [intent=direct_read availability=implemented operation=github.git_tags_tag_sha]
  git trees create - POST /repos/{owner}/{repo}/git/trees [intent=reverse_etl availability=implemented write=git_trees]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  git trees view - Read /repos/{owner}/{repo}/git/trees/{tree_sha} [intent=direct_read availability=implemented operation=github.git_trees_tree_sha]
  hash-algorithm view - Read /repos/{owner}/{repo}/hash-algorithm [intent=direct_read availability=implemented operation=github.hash_algorithm]
  hooks deliveries view - Read /repos/{owner}/{repo}/hooks/{hook_id}/deliveries [intent=direct_read availability=implemented operation=github.hooks_hook_id_deliveries]
  hooks deliveries view-2 - Read /repos/{owner}/{repo}/hooks/{hook_id}/deliveries/{delivery_id} [intent=direct_read availability=implemented operation=github.hooks_hook_id_deliveries_delivery_id]
  webhook create - POST /repos/{owner}/{repo}/hooks/{hook_id}/deliveries/{delivery_id}/attempts [intent=reverse_etl availability=implemented write=hooks_hook_id_deliveries_delivery_id_attempts]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  webhook create-2 - POST /repos/{owner}/{repo}/hooks/{hook_id}/pings [intent=reverse_etl availability=implemented write=hooks_hook_id_pings]; risk: low; approval: Reverse ETL writes require plan, preview, approval, execute.
  webhook create-3 - POST /repos/{owner}/{repo}/hooks/{hook_id}/tests [intent=reverse_etl availability=implemented write=hooks_hook_id_tests]; risk: low; approval: Reverse ETL writes require plan, preview, approval, execute.
  immutable-releases delete - DELETE /repos/{owner}/{repo}/immutable-releases [intent=reverse_etl availability=implemented write=immutable_releases]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  immutable-releases view - Read /repos/{owner}/{repo}/immutable-releases [intent=direct_read availability=implemented operation=github.immutable_releases2]
  immutable-releases set - PUT /repos/{owner}/{repo}/immutable-releases [intent=reverse_etl availability=implemented write=immutable_releases3]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  import delete - DELETE /repos/{owner}/{repo}/import [intent=reverse_etl availability=implemented write=import]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  import view - Read /repos/{owner}/{repo}/import [intent=direct_read availability=implemented operation=github.import2]
  import update - PATCH /repos/{owner}/{repo}/import [intent=reverse_etl availability=implemented write=import3]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  import set - PUT /repos/{owner}/{repo}/import [intent=reverse_etl availability=implemented write=import4]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  import authors view - Read /repos/{owner}/{repo}/import/authors [intent=direct_read availability=implemented operation=github.import_authors]
  import authors update - PATCH /repos/{owner}/{repo}/import/authors/{author_id} [intent=reverse_etl availability=implemented write=import_authors_author_id]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  import large_files view - Read /repos/{owner}/{repo}/import/large_files [intent=direct_read availability=implemented operation=github.import_large_files]
  import lfs update - PATCH /repos/{owner}/{repo}/import/lfs [intent=reverse_etl availability=implemented write=import_lfs]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  installation view - Read /repos/{owner}/{repo}/installation [intent=direct_read availability=implemented operation=github.installation]
  interaction-limits delete - DELETE /repos/{owner}/{repo}/interaction-limits [intent=reverse_etl availability=implemented write=interaction_limits]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  interaction-limits view - Read /repos/{owner}/{repo}/interaction-limits [intent=direct_read availability=implemented operation=github.interaction_limits2]
  interaction-limits set - PUT /repos/{owner}/{repo}/interaction-limits [intent=reverse_etl availability=implemented write=interaction_limits3]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  interaction-limits bypass-list delete - DELETE /repos/{owner}/{repo}/interaction-limits/pulls/bypass-list [intent=reverse_etl availability=implemented write=interaction_limits_pulls_bypass_list]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  interaction-limits bypass-list view - Read /repos/{owner}/{repo}/interaction-limits/pulls/bypass-list [intent=direct_read availability=implemented operation=github.interaction_limits_pulls_bypass_list2]
  interaction-limits bypass-list set - PUT /repos/{owner}/{repo}/interaction-limits/pulls/bypass-list [intent=reverse_etl availability=implemented write=interaction_limits_pulls_bypass_list3]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  invitations delete - DELETE /repos/{owner}/{repo}/invitations/{invitation_id} [intent=reverse_etl availability=implemented write=invitations_invitation_id]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  invitations update - PATCH /repos/{owner}/{repo}/invitations/{invitation_id} [intent=reverse_etl availability=implemented write=invitations_invitation_id2]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  issue-types view - Read /repos/{owner}/{repo}/issue-types [intent=direct_read availability=implemented operation=github.issue_types]
  issues pin delete - DELETE /repos/{owner}/{repo}/issues/comments/{comment_id}/pin [intent=reverse_etl availability=implemented write=issues_comments_comment_id_pin]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  issues pin set - PUT /repos/{owner}/{repo}/issues/comments/{comment_id}/pin [intent=reverse_etl availability=implemented write=issues_comments_comment_id_pin2]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  issues reactions view - Read /repos/{owner}/{repo}/issues/comments/{comment_id}/reactions [intent=direct_read availability=implemented operation=github.issues_comments_comment_id_reactions]
  issues reactions create - POST /repos/{owner}/{repo}/issues/comments/{comment_id}/reactions [intent=reverse_etl availability=implemented write=issues_comments_comment_id_reactions2]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  issues reactions delete - DELETE /repos/{owner}/{repo}/issues/comments/{comment_id}/reactions/{reaction_id} [intent=reverse_etl availability=implemented write=issues_comments_comment_id_reactions_reaction_id]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  issues assignees view - Read /repos/{owner}/{repo}/issues/{issue_number}/assignees/{assignee} [intent=direct_read availability=implemented operation=github.issues_issue_number_assignees_assignee]
  issues blocked_by view - Read /repos/{owner}/{repo}/issues/{issue_number}/dependencies/blocked_by [intent=direct_read availability=implemented operation=github.issues_issue_number_dependencies_blocked_by]
  issues blocked_by create - POST /repos/{owner}/{repo}/issues/{issue_number}/dependencies/blocked_by [intent=reverse_etl availability=implemented write=issues_issue_number_dependencies_blocked_by2]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  issues blocked_by delete - DELETE /repos/{owner}/{repo}/issues/{issue_number}/dependencies/blocked_by/{issue_id} [intent=reverse_etl availability=implemented write=issues_issue_number_dependencies_blocked_by_issue_id]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  issues blocking view - Read /repos/{owner}/{repo}/issues/{issue_number}/dependencies/blocking [intent=direct_read availability=implemented operation=github.issues_issue_number_dependencies_blocking]
  issues issue-field-values view - Read /repos/{owner}/{repo}/issues/{issue_number}/issue-field-values [intent=direct_read availability=implemented operation=github.issues_issue_number_issue_field_values]
  issues issue-field-values create - POST /repos/{owner}/{repo}/issues/{issue_number}/issue-field-values [intent=reverse_etl availability=implemented write=issues_issue_number_issue_field_values2]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  issues issue-field-values set - PUT /repos/{owner}/{repo}/issues/{issue_number}/issue-field-values [intent=reverse_etl availability=implemented write=issues_issue_number_issue_field_values3]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  issues issue-field-values delete - DELETE /repos/{owner}/{repo}/issues/{issue_number}/issue-field-values/{issue_field_id} [intent=reverse_etl availability=implemented write=issues_issue_number_issue_field_values_issue_field_id]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  issues parent view - Read /repos/{owner}/{repo}/issues/{issue_number}/parent [intent=direct_read availability=implemented operation=github.issues_issue_number_parent]
  issues reactions view-2 - Read /repos/{owner}/{repo}/issues/{issue_number}/reactions [intent=direct_read availability=implemented operation=github.issues_issue_number_reactions]
  issues reactions create-2 - POST /repos/{owner}/{repo}/issues/{issue_number}/reactions [intent=reverse_etl availability=implemented write=issues_issue_number_reactions2]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  issues reactions delete-2 - DELETE /repos/{owner}/{repo}/issues/{issue_number}/reactions/{reaction_id} [intent=reverse_etl availability=implemented write=issues_issue_number_reactions_reaction_id]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  issues sub_issue delete - DELETE /repos/{owner}/{repo}/issues/{issue_number}/sub_issue [intent=reverse_etl availability=implemented write=issues_issue_number_sub_issue]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  issues sub_issues view - Read /repos/{owner}/{repo}/issues/{issue_number}/sub_issues [intent=direct_read availability=implemented operation=github.issues_issue_number_sub_issues]
  issues sub_issues create - POST /repos/{owner}/{repo}/issues/{issue_number}/sub_issues [intent=reverse_etl availability=implemented write=issues_issue_number_sub_issues2]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  issues priority update - PATCH /repos/{owner}/{repo}/issues/{issue_number}/sub_issues/priority [intent=reverse_etl availability=implemented write=issues_issue_number_sub_issues_priority]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  license view - Read /repos/{owner}/{repo}/license [intent=direct_read availability=implemented operation=github.license]
  merge-upstream create - POST /repos/{owner}/{repo}/merge-upstream [intent=reverse_etl availability=implemented write=merge_upstream]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  milestones labels view - Read /repos/{owner}/{repo}/milestones/{milestone_number}/labels [intent=direct_read availability=implemented operation=github.milestones_milestone_number_labels]
  notifications view - Read /repos/{owner}/{repo}/notifications [intent=direct_read availability=implemented operation=github.notifications]
  notifications set - PUT /repos/{owner}/{repo}/notifications [intent=reverse_etl availability=implemented write=notifications2]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  pages delete - DELETE /repos/{owner}/{repo}/pages [intent=reverse_etl availability=implemented write=pages]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  pages view - Read /repos/{owner}/{repo}/pages [intent=direct_read availability=implemented operation=github.pages2]
  pages create - POST /repos/{owner}/{repo}/pages [intent=reverse_etl availability=implemented write=pages3]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  pages set - PUT /repos/{owner}/{repo}/pages [intent=reverse_etl availability=implemented write=pages4]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  pages builds view - Read /repos/{owner}/{repo}/pages/builds [intent=direct_read availability=implemented operation=github.pages_builds]
  pages builds create - POST /repos/{owner}/{repo}/pages/builds [intent=reverse_etl availability=implemented write=pages_builds2]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  pages latest view - Read /repos/{owner}/{repo}/pages/builds/latest [intent=direct_read availability=implemented operation=github.pages_builds_latest]
  pages builds view-2 - Read /repos/{owner}/{repo}/pages/builds/{build_id} [intent=direct_read availability=implemented operation=github.pages_builds_build_id]
  pages deployments create - POST /repos/{owner}/{repo}/pages/deployments [intent=reverse_etl availability=implemented write=pages_deployments]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  pages deployments view - Read /repos/{owner}/{repo}/pages/deployments/{pages_deployment_id} [intent=direct_read availability=implemented operation=github.pages_deployments_pages_deployment_id]
  pages cancel create - POST /repos/{owner}/{repo}/pages/deployments/{pages_deployment_id}/cancel [intent=reverse_etl availability=implemented write=pages_deployments_pages_deployment_id_cancel]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  pages health view - Read /repos/{owner}/{repo}/pages/health [intent=direct_read availability=implemented operation=github.pages_health]
  private-vulnerability-reporting delete - DELETE /repos/{owner}/{repo}/private-vulnerability-reporting [intent=reverse_etl availability=implemented write=private_vulnerability_reporting]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  private-vulnerability-reporting view - Read /repos/{owner}/{repo}/private-vulnerability-reporting [intent=direct_read availability=implemented operation=github.private_vulnerability_reporting2]
  private-vulnerability-reporting set - PUT /repos/{owner}/{repo}/private-vulnerability-reporting [intent=reverse_etl availability=implemented write=private_vulnerability_reporting3]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  properties values view - Read /repos/{owner}/{repo}/properties/values [intent=direct_read availability=implemented operation=github.properties_values]
  properties values update - PATCH /repos/{owner}/{repo}/properties/values [intent=reverse_etl availability=implemented write=properties_values2]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  pulls reactions view - Read /repos/{owner}/{repo}/pulls/comments/{comment_id}/reactions [intent=direct_read availability=implemented operation=github.pulls_comments_comment_id_reactions]
  pulls reactions create - POST /repos/{owner}/{repo}/pulls/comments/{comment_id}/reactions [intent=reverse_etl availability=implemented write=pulls_comments_comment_id_reactions2]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  pulls reactions delete - DELETE /repos/{owner}/{repo}/pulls/comments/{comment_id}/reactions/{reaction_id} [intent=reverse_etl availability=implemented write=pulls_comments_comment_id_reactions_reaction_id]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  pulls codespaces create - POST /repos/{owner}/{repo}/pulls/{pull_number}/codespaces [intent=reverse_etl availability=implemented write=pulls_pull_number_codespaces]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  pulls commits view - Read /repos/{owner}/{repo}/pulls/{pull_number}/commits [intent=direct_read availability=implemented operation=github.pulls_pull_number_commits]
  pulls files view - Read /repos/{owner}/{repo}/pulls/{pull_number}/files [intent=direct_read availability=implemented operation=github.pulls_pull_number_files]
  pulls merge view - Read /repos/{owner}/{repo}/pulls/{pull_number}/merge [intent=direct_read availability=implemented operation=github.pulls_pull_number_merge]
  pulls requested_reviewers delete - DELETE /repos/{owner}/{repo}/pulls/{pull_number}/requested_reviewers [intent=reverse_etl availability=implemented write=pulls_pull_number_requested_reviewers]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  pulls reviews view - Read /repos/{owner}/{repo}/pulls/{pull_number}/reviews [intent=direct_read availability=implemented operation=github.pulls_pull_number_reviews]
  pulls comments view - Read /repos/{owner}/{repo}/pulls/{pull_number}/reviews/{review_id}/comments [intent=direct_read availability=implemented operation=github.pulls_pull_number_reviews_review_id_comments]
  readme view - Read /repos/{owner}/{repo}/readme [intent=direct_read availability=implemented operation=github.readme]
  readme view-2 - Read /repos/{owner}/{repo}/readme/{dir} [intent=direct_read availability=implemented operation=github.readme_dir]
  releases assets view - Read /repos/{owner}/{repo}/releases/assets/{asset_id} [intent=direct_read availability=implemented operation=github.releases_assets_asset_id]
  releases generate-notes view - POST /repos/{owner}/{repo}/releases/generate-notes [intent=reverse_etl availability=implemented write=releases_generate_notes]; risk: low; approval: Reverse ETL writes require plan, preview, approval, execute.
  releases assets view-2 - Read /repos/{owner}/{repo}/releases/{release_id}/assets [intent=direct_read availability=implemented operation=github.releases_release_id_assets]
  releases assets view-3 - POST /repos/{owner}/{repo}/releases/{release_id}/assets [intent=reverse_etl availability=implemented write=releases_release_id_assets2]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  releases reactions view - Read /repos/{owner}/{repo}/releases/{release_id}/reactions [intent=direct_read availability=implemented operation=github.releases_release_id_reactions]
  releases reactions create - POST /repos/{owner}/{repo}/releases/{release_id}/reactions [intent=reverse_etl availability=implemented write=releases_release_id_reactions2]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  releases reactions delete - DELETE /repos/{owner}/{repo}/releases/{release_id}/reactions/{reaction_id} [intent=reverse_etl availability=implemented write=releases_release_id_reactions_reaction_id]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  rulesets rule-suites view - Read /repos/{owner}/{repo}/rulesets/rule-suites [intent=direct_read availability=implemented operation=github.rulesets_rule_suites]
  rulesets rule-suites view-2 - Read /repos/{owner}/{repo}/rulesets/rule-suites/{rule_suite_id} [intent=direct_read availability=implemented operation=github.rulesets_rule_suites_rule_suite_id]
  rulesets history view - Read /repos/{owner}/{repo}/rulesets/{ruleset_id}/history [intent=direct_read availability=implemented operation=github.rulesets_ruleset_id_history]
  rulesets history view-2 - Read /repos/{owner}/{repo}/rulesets/{ruleset_id}/history/{version_id} [intent=direct_read availability=implemented operation=github.rulesets_ruleset_id_history_version_id]
  secret-scanning locations view - Read /repos/{owner}/{repo}/secret-scanning/alerts/{alert_number}/locations [intent=direct_read availability=implemented operation=github.secret_scanning_alerts_alert_number_locations]
  secret-scanning push-protection-bypasses create - POST /repos/{owner}/{repo}/secret-scanning/push-protection-bypasses [intent=reverse_etl availability=implemented write=secret_scanning_push_protection_bypasses]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  secret-scanning scan-history view - Read /repos/{owner}/{repo}/secret-scanning/scan-history [intent=direct_read availability=implemented operation=github.secret_scanning_scan_history]
  security-advisories create - POST /repos/{owner}/{repo}/security-advisories [intent=reverse_etl availability=implemented write=security_advisories]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  security-advisories reports create - POST /repos/{owner}/{repo}/security-advisories/reports [intent=reverse_etl availability=implemented write=security_advisories_reports]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  security-advisories update - PATCH /repos/{owner}/{repo}/security-advisories/{ghsa_id} [intent=reverse_etl availability=implemented write=security_advisories_ghsa_id]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  security-advisories cve create - POST /repos/{owner}/{repo}/security-advisories/{ghsa_id}/cve [intent=reverse_etl availability=implemented write=security_advisories_ghsa_id_cve]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  security-advisories forks create - POST /repos/{owner}/{repo}/security-advisories/{ghsa_id}/forks [intent=reverse_etl availability=implemented write=security_advisories_ghsa_id_forks]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  stats code_frequency view - Read /repos/{owner}/{repo}/stats/code_frequency [intent=direct_read availability=implemented operation=github.stats_code_frequency]
  stats commit_activity view - Read /repos/{owner}/{repo}/stats/commit_activity [intent=direct_read availability=implemented operation=github.stats_commit_activity]
  stats contributors view - Read /repos/{owner}/{repo}/stats/contributors [intent=direct_read availability=implemented operation=github.stats_contributors]
  stats participation view - Read /repos/{owner}/{repo}/stats/participation [intent=direct_read availability=implemented operation=github.stats_participation]
  stats punch_card view - Read /repos/{owner}/{repo}/stats/punch_card [intent=direct_read availability=implemented operation=github.stats_punch_card]
  statuses create - POST /repos/{owner}/{repo}/statuses/{sha} [intent=reverse_etl availability=implemented write=statuses_sha]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  subscription delete - DELETE /repos/{owner}/{repo}/subscription [intent=reverse_etl availability=implemented write=subscription]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  subscription view - Read /repos/{owner}/{repo}/subscription [intent=direct_read availability=implemented operation=github.subscription2]
  subscription set - PUT /repos/{owner}/{repo}/subscription [intent=reverse_etl availability=implemented write=subscription3]; risk: medium; approval: Reverse ETL writes require plan, preview, approval, execute.
  teams view - Read /repos/{owner}/{repo}/teams [intent=direct_read availability=implemented operation=github.teams]
  traffic clones view - Read /repos/{owner}/{repo}/traffic/clones [intent=direct_read availability=implemented operation=github.traffic_clones]
  traffic paths view - Read /repos/{owner}/{repo}/traffic/popular/paths [intent=direct_read availability=implemented operation=github.traffic_popular_paths]
  traffic referrers view - Read /repos/{owner}/{repo}/traffic/popular/referrers [intent=direct_read availability=implemented operation=github.traffic_popular_referrers]
  traffic views view - Read /repos/{owner}/{repo}/traffic/views [intent=direct_read availability=implemented operation=github.traffic_views]
  transfer create - POST /repos/{owner}/{repo}/transfer [intent=reverse_etl availability=implemented write=transfer]; risk: critical; approval: Reverse ETL writes require plan, preview, approval, execute.
  vulnerability-alerts delete - DELETE /repos/{owner}/{repo}/vulnerability-alerts [intent=reverse_etl availability=implemented write=vulnerability_alerts]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.
  vulnerability-alerts view - Read /repos/{owner}/{repo}/vulnerability-alerts [intent=direct_read availability=implemented operation=github.vulnerability_alerts2]
  vulnerability-alerts set - PUT /repos/{owner}/{repo}/vulnerability-alerts [intent=reverse_etl availability=implemented write=vulnerability_alerts3]; risk: high; approval: Reverse ETL writes require plan, preview, approval, execute.

GLOBAL FLAGS
  --json - Write machine-readable JSON output.
  --connection <string> - Use a saved GitHub connector credential and repository scope. - maps_to=connection

SAFETY
  Help is generated from validated connector metadata. It does not read credentials, open project state, contact connector APIs, or execute commands.
  Mutation commands retain their declared plan, preview, approval, and execution policy.

EXIT STATUS
  0  Help rendered successfully.
  2  The connector help path is invalid.
```
