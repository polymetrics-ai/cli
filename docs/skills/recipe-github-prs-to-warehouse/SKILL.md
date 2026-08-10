---
name: recipe-github-prs-to-warehouse
description: Sync GitHub pull requests into the local warehouse.
---

# recipe-github-prs-to-warehouse

- Create a GitHub credential with required config `owner` and `repo`, optional `auth_type`, and a token from the environment unless you explicitly opt into public reads.
- Create a warehouse credential with a local path.
- Create a connection with stream `pull_requests` and table `github_pull_requests`.
- Run `pm etl run --connection github_to_warehouse --stream pull_requests --batch-size 100 --json`.
