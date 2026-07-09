# Bitbucket

## Overview

Bitbucket Cloud connector metadata is being introduced through the Bitbucket CLI parity parent issue (#79). This seed bundle records a safe command map for repository, pull request, issue, pipeline, deployment, download, webhook, branch-restriction, workspace, project, and snippet workflows.

This slice does not enable live Bitbucket reads or writes. Later lanes add the full operation ledger, streams, direct reads, and approval-gated reverse ETL actions.

## Auth setup

Future Bitbucket execution will use credentials supplied through the Polymetrics credential vault, such as OAuth or app-password style access tokens. Do not place token values in prompts, command examples, fixtures, logs, or connector metadata.

## Streams notes

No Bitbucket streams are implemented in this metadata seed. `streams.json` is intentionally empty so the bundle can load while #90 focuses on `cli_surface.json`. Planned stream-backed command groups include repositories, branches, commits, tags, pull requests, issues, pipelines, deployments, downloads metadata, webhooks, and branch restrictions.

## Write actions & risks

No Bitbucket write actions are implemented in this metadata seed. Future writes must be modeled as explicit reverse ETL actions with record schemas, risk text, and plan → preview → approval → execute semantics. Destructive/admin operations such as repository deletion, branch protection changes, and sensitive webhook or snippet content workflows remain blocked by default until #96 defines the policy.

## Known limits

- The full official Bitbucket Swagger has 331 operations (GET 179, POST 50, PUT 48, DELETE 54); #93 owns the complete operation ledger.
- Direct-read execution and output policies are deferred to #94.
- GraphQL or advanced fixed-body support is deferred to #95 if Bitbucket needs it.
- Binary downloads and local git workflows are metadata-only in this slice and require bounded destination/output policy before execution.
- Generic raw API calls are explicitly disallowed.
