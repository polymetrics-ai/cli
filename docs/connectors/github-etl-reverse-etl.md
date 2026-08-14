# GitHub ETL And Reverse ETL Reference

This is a task-oriented setup guide for the `pm` GitHub connector. The generated
connector manual and runtime inspection own the current stream, action, and command
inventory; this guide does not duplicate it. Check the contract with:

```bash
pm connectors inspect github
pm connectors inspect github --json
```

Generated connector references are:

- [generated GitHub connector manual](github/MANUAL.md)
- [generated GitHub connector skill](github/SKILL.md)
- [generated general GitHub skill](../skills/pm-github/SKILL.md)

## Security Model

Agents and humans must use the preview-approval-execute workflow for reverse
ETL writes:

```bash
pm reverse plan <name> --source-table <table> --destination github:<credential> --action <action> --map <source>:<target>
pm reverse preview <plan-id> --json
pm reverse run <plan-id> --approval-token-stdin --json
```

For withheld sensitive fields, including `env_only` values that must be supplied
with `--from-env <flag>=ENV`, follow the [reverse CLI reference](../cli/reverse.md).

Never place GitHub tokens or private keys in chat, docs, command arguments, or
JSON output. Use environment variables or stdin:

```bash
export GITHUB_TOKEN=...
pm credentials add github-token \
  --connector github \
  --config owner=OWNER \
  --config repo=REPO \
  --from-env token=GITHUB_TOKEN
```

For GitHub App private keys:

```bash
pm credentials add github-app \
  --connector github \
  --config owner=OWNER \
  --config repo=REPO \
  --config auth_type=github_app \
  --config app_id=12345 \
  --config installation_id=67890 \
  --value-stdin private_key < app-private-key.pem
```

The generated connector manual owns authentication modes and configuration defaults. Set
`public_access` only for unauthenticated public reads; use a token or GitHub App credential for
private or write access.

## ETL

Use the generated manual or `pm connectors inspect github --json` for the current stream names,
pagination, primary keys, and cursor fields. For example:

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

## Reverse ETL and Direct Writes

All declared GitHub writes use the plan, preview, approval, and execution gates. The generated
manual owns the action and direct-write inventory, required fields, risk labels, and confirmation
requirements. The connector does not expose generic `gh api` or raw GraphQL mutation escape hatches.

## Official References

- [GitHub REST API authentication](https://docs.github.com/en/rest/authentication/authenticating-to-the-rest-api)
- [GitHub App installation authentication](https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/authenticating-as-a-github-app-installation)
- [GitHub REST API](https://docs.github.com/en/rest)
- [GitHub GraphQL API](https://docs.github.com/en/graphql)
