# Bitbucket

## Overview

Bitbucket Cloud connector support covers safe CLI parity for repository, branch, commit, tag, pull request, issue, pipeline, deployment, download metadata, webhook, branch restriction, workspace, project, and snippet workflows. The bundle uses the Bitbucket Cloud REST API 2.0 and the official Swagger at https://api.bitbucket.org/swagger.json as its operation source of truth.

## Auth setup

Use `pm credentials add <name> --connector bitbucket --config workspace=<workspace> --config repo_slug=<repo> --from-env access_token=BITBUCKET_TOKEN` for authenticated reads and writes. Public reads may omit `access_token` when Bitbucket permits anonymous access. Never put token values in prompts, command examples, fixtures, logs, or connector metadata.

## Streams notes

Streams use Bitbucket's paginated `values` envelope and `next` URL pagination. Implemented stream-backed command families include repositories, branches, commits, tags, pull requests, issues, pipelines, deployments, downloads metadata, webhooks, branch restrictions, workspace projects, and snippets. Binary download, patch, diff, log, attachment, and repository clone workflows are deliberately not executed as streams.

## Write actions & risks

Selected mutations are modeled as explicit reverse ETL actions: create repository, create/update issue, create/merge/decline pull request, run/stop pipeline, create/delete webhook, create branch restriction, and create snippet. Reverse ETL remains plan → preview → approval → execute. Destructive/admin/sensitive operation rows in `operations.json` carry typed confirmation policy metadata and remain blocked unless represented by an explicit reviewed write action.

## Known limits

- Full ledger coverage is metadata-first: all 331 Swagger operations are classified, but only reviewed app-intent streams, direct reads, and write actions execute.
- Bitbucket has no GraphQL surface in the official Swagger; #95 is recorded as REST-only and no GraphQL executor is added.
- Local git clone, browser, binary download, raw patch/diff/log, issue import/export archive, attachment, and arbitrary raw API workflows remain unsupported until bounded local-output policies exist.
- Some complex Bitbucket write bodies require callers to provide nested JSON objects through reverse ETL records; generic arbitrary mutation bodies are not exposed.
- Generic raw API calls are explicitly disallowed.
