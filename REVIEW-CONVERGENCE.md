# Review Convergence — Current Foundations Main Integration r1

## Review target

Review the unpushed provisional composite `808896a28873c5f0479fa10e2f798da56f885b5e` on `fm/cli-current-foundations-main-integration-r1`, based on `114a67727f2ef60b132054091c73987be4118a9b`.

This is one rollup checkpoint, not component-PR review. Do not push, open a pull request, merge, retarget, close, or mutate any component branch or pipeline worktree as part of this review.

## Exact ancestry to audit

| Foundation | Exact head | Provenance | Preserving merge |
| --- | --- | --- | --- |
| source import / #4306 | `19a32bd0bc08faf217be8f45b39841b5ff589a92` | published #4312, prior qualified intake | `223f7b126eb4039f5c940a3cf233b15d6a18eff6` |
| closed operation runtime / #4307 | `3c768cade6703426afd2272fbc01bfd60583e04f` | published #4311, prior qualified intake | `1cb9cdb31c2fc446fe1da0b176b7422f04e81111` |
| status/export / #4302 | `fe5b8e18788538c4fcce34969da7ff88a7fa66d6` | published #4308, Firstmate terminal qualification | `9e3cd99b7ebd2ebac2303ad8770e50fee85c92c6` |
| structured bodies / #4305 | `55ddb650aa5594ddd156b0939cb1df6027a31d56` | captain-authorized local pipeline `01M0DY0HM9HNZVNKJ2J9Z9SCG7` | `0eb98d3844da7b48d0ca27f51ba7deb46d8f5d1b` |
| declarative reverse ETL / #4303 | `e7f474375af969555efd82f684ad6d0b8a26cfc0` | captain-authorized local pipeline `01M0DYNQ9HSJBYS9YQ4MJR4JGR` | `808896a28873c5f0479fa10e2f798da56f885b5e` |

The local worktrees were read only. `data/cli-current-foundations-main-integration-r1/input-manifest.json` is the machine-readable provenance record.

## Required convergence findings

- Typed direct write accepts only declaration-owned path, query, structured body, and header bindings. It must reject malformed, unknown, oversized, duplicate, CR/LF, and cross-bound values before I/O; no raw HTTP method/path/header/body/action escape hatch is present.
- Terminal direct-write results preserve provider status, repeatable headers, exact text/binary body bytes, response presence, operation ID, and path, including terminal errors. Generated diagnostics must not copy response bodies or secrets.
- A fixed GraphQL operation parses JSON when its content type is JSON or omitted; a provider-declared text/binary response remains byte-exact rather than fabricated as a GraphQL envelope.
- `rest_status` keeps the final non-2xx HEAD response and typed declaration-owned headers after retries. Binary/text GET failures remain ordinary errors and do not create an output file.
- Persisted App and installed CLI reach multiple independently selectable reverse-ETL actions with plan, approval, apply, durable acknowledgement, provider result persistence, and provider readback; no connector-name branch narrows an action.
- Source-locked provider declarations remain lossless/bounded and reach the generated command/help surface; generated documentation is synchronized.
- Existing specialized GitHub, scalar, form, SCIM, multipart, and binary behavior remain covered rather than selected away in a conflict.

## Focused evidence

| Command | Result |
| --- | --- |
| `go test -timeout 20m ./internal/connectors/connsdk ./internal/connectors/engine` | pass after the combined direct-write/status resolution |
| `go test -timeout 20m ./internal/app -run '^TestFoundationRollupPreservesMultiActionReverseETLComposition$' -count=1` | pass |
| App, CLI, commandrunner, sync-transport, source-import, generator, build, docs, boundary, and website-data checks | pass; exact commands and assertions recorded in `data/cli-current-foundations-main-integration-r1/evidence-manifest.json` |

## Out of scope for this provisional checkpoint

No full verifier, real-provider operation, credential, temporary declaration/download, PR, CI, or no-mistakes stage is claimed complete. Those are later Firstmate-controlled gates. Any later component commits are additive follow-ups and must not replace the exact reviewed heads above.
