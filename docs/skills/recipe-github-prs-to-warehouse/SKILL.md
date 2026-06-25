---
name: recipe-github-prs-to-warehouse
description: Sync GitHub pull requests into the local warehouse.
---

# recipe-github-prs-to-warehouse

- Create a GitHub credential with config `repository=owner/repo` and optional token from environment.
- Create a warehouse credential with a local path.
- Create a connection with stream `pull_requests` and table `github_pull_requests`.
- Run `pm etl run --connection github_to_warehouse --stream pull_requests --batch-size 100 --json`.
