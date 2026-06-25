# GitHub ETL And Reverse ETL Reference

This is the stable human reference for the `pm` GitHub connector. The canonical
machine-readable contract is generated from Go connector metadata and should be
checked with:

```bash
pm connectors inspect github
pm connectors inspect github --json
```

Generated connector docs are also written to:

- `docs/connectors/github/MANUAL.md`
- `docs/connectors/github/SKILL.md`
- `docs/skills/pm-github/SKILL.md`

## Security Model

Agents and humans must use the preview-approval-execute workflow for reverse
ETL writes:

```bash
pm reverse plan <name> --source-table <table> --destination github:<credential> --action <action> --map <source>:<target>
pm reverse preview <plan-id> --json
pm reverse run <plan-id> --approve <approval-token> --json
```

Never place GitHub tokens or private keys in chat, docs, command arguments, or
JSON output. Use environment variables or stdin:

```bash
export GITHUB_TOKEN=...
pm credentials add github-token \
  --connector github \
  --config repository=OWNER/REPO \
  --from-env token=GITHUB_TOKEN
```

For GitHub App private keys:

```bash
pm credentials add github-app \
  --connector github \
  --config repository=OWNER/REPO \
  --config auth_type=github_app \
  --config app_id=12345 \
  --config installation_id=67890 \
  --value-stdin private_key < app-private-key.pem
```

## Authentication

The connector supports:

- `public`: unauthenticated public repository reads. Writes are blocked.
- `token`: classic PAT, fine-grained PAT, OAuth token, GitHub Actions
  `GITHUB_TOKEN`, or pre-generated installation access token.
- `github_app`: server-to-server authentication by signing a GitHub App JWT and
  exchanging it for an installation token.

`auth_type=auto` is the default. It selects token auth when a token secret is
present, GitHub App auth when app config and private key material are present,
and public auth otherwise.

## ETL Streams

All GitHub streams use bounded page reads. Set `per_page` up to `100`, and set
`max_pages=all`, `max_pages=unlimited`, or `max_pages=0` to exhaust a stream.

| Stream | GitHub API family | Primary key | Cursor |
| --- | --- | --- | --- |
| `repository` | Repository metadata | `node_id` | `updated_at` |
| `issues` | Issues | `node_id` | `updated_at` |
| `pull_requests` | Pull requests | `node_id` | `updated_at` |
| `branches` | Branches | `name` | none |
| `commits` | Commits | `sha` | `commit_committer_date` |
| `tags` | Tags | `name` | none |
| `releases` | Releases | `id` | `published_at` |
| `labels` | Labels | `name` | none |
| `milestones` | Milestones | `number` | `updated_at` |
| `issue_comments` | Issue comments | `id` | `updated_at` |
| `pull_request_review_comments` | Pull request review comments | `id` | `updated_at` |
| `collaborators` | Collaborators | `id` | none |
| `contributors` | Contributors | `id` | none |
| `stargazers` | Stargazers | `id` | none |
| `subscribers` | Watchers/subscribers | `id` | none |
| `workflows` | GitHub Actions workflows | `id` | `updated_at` |
| `workflow_runs` | GitHub Actions workflow runs | `id` | `updated_at` |
| `workflow_artifacts` | GitHub Actions artifacts | `id` | `updated_at` |
| `deployments` | Deployments | `id` | `updated_at` |

Example ETL:

```bash
pm connections create github_prs_to_warehouse \
  --source github:github-token \
  --destination warehouse:warehouse-local \
  --stream pull_requests \
  --primary-key node_id \
  --cursor updated_at \
  --table github_pull_requests

pm etl run --connection github_prs_to_warehouse --stream pull_requests --batch-size 100 --json
```

## Sync Modes

The GitHub connector advertises all source-to-warehouse sync modes supported by
the local ETL runtime:

- `full_refresh_append`
- `full_refresh_overwrite`
- `full_refresh_overwrite_deduped`
- `incremental_append`
- `incremental_append_deduped`

## Reverse ETL Actions

All GitHub reverse ETL actions require token or GitHub App auth and approval
before execution.

| Action | Endpoint shape | Required fields |
| --- | --- | --- |
| `create_issue` | `POST /repos/{owner}/{repo}/issues` | `title` |
| `update_issue` | `PATCH /repos/{owner}/{repo}/issues/{issue_number}` | `issue_number` or `number` |
| `comment_issue` | `POST /repos/{owner}/{repo}/issues/{issue_number}/comments` | `issue_number`, `pull_number`, or `number`; `body` |
| `close_issue` | `PATCH /repos/{owner}/{repo}/issues/{issue_number}` | `issue_number` or `number` |
| `create_pull_request` | `POST /repos/{owner}/{repo}/pulls` | `title`, `head`, `base` |
| `update_pull_request` | `PATCH /repos/{owner}/{repo}/pulls/{pull_number}` | `pull_number` or `number` |
| `close_pull_request` | `PATCH /repos/{owner}/{repo}/pulls/{pull_number}` | `pull_number` or `number` |
| `request_reviewers` | `POST /repos/{owner}/{repo}/pulls/{pull_number}/requested_reviewers` | `pull_number` or `number`; `reviewers` or `team_reviewers` |
| `merge_pull_request` | `PUT /repos/{owner}/{repo}/pulls/{pull_number}/merge` | `pull_number` or `number` |
| `create_label` | `POST /repos/{owner}/{repo}/labels` | `name`, `color` |
| `update_label` | `PATCH /repos/{owner}/{repo}/labels/{name}` | `name` |
| `delete_label` | `DELETE /repos/{owner}/{repo}/labels/{name}` | `name` |
| `create_milestone` | `POST /repos/{owner}/{repo}/milestones` | `title` |
| `update_milestone` | `PATCH /repos/{owner}/{repo}/milestones/{milestone_number}` | `milestone_number` or `number` |
| `delete_milestone` | `DELETE /repos/{owner}/{repo}/milestones/{milestone_number}` | `milestone_number` or `number` |
| `create_release` | `POST /repos/{owner}/{repo}/releases` | `tag_name` |
| `update_release` | `PATCH /repos/{owner}/{repo}/releases/{release_id}` | `release_id` or `id` |
| `delete_release` | `DELETE /repos/{owner}/{repo}/releases/{release_id}` | `release_id` or `id` |
| `dispatch_workflow` | `POST /repos/{owner}/{repo}/actions/workflows/{workflow_id}/dispatches` | `workflow_id`, `ref` |
| `rerun_workflow_run` | `POST /repos/{owner}/{repo}/actions/runs/{run_id}/rerun` | `run_id`, `workflow_run_id`, or `id` |
| `cancel_workflow_run` | `POST /repos/{owner}/{repo}/actions/runs/{run_id}/cancel` | `run_id`, `workflow_run_id`, or `id` |
| `delete_workflow_run` | `DELETE /repos/{owner}/{repo}/actions/runs/{run_id}` | `run_id`, `workflow_run_id`, or `id` |
| `create_pull_request_review` | `POST /repos/{owner}/{repo}/pulls/{pull_number}/reviews` | `pull_number` or `number` |
| `create_or_update_file` | `PUT /repos/{owner}/{repo}/contents/{path}` | `path`, `message`, `content` or `content_base64` |
| `delete_file` | `DELETE /repos/{owner}/{repo}/contents/{path}` | `path`, `message`, `sha` |

For optional fields and per-action risk text, use:

```bash
pm connectors inspect github
```

## Official References

- [GitHub REST API authentication](https://docs.github.com/en/rest/authentication/authenticating-to-the-rest-api)
- [GitHub App installation authentication](https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/authenticating-as-a-github-app-installation)
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
