# Current Foundations Main Integration r1 — Provisional Composite

**Composite:** `808896a28873c5f0479fa10e2f798da56f885b5e` on `fm/cli-current-foundations-main-integration-r1`.

**Base:** `114a67727f2ef60b132054091c73987be4118a9b`, whose preserved parent is the existing #4308 integration merge. It already retains the #4310 non-empty credential-input foundation and #4309 source-lock embed.

## Exact provenance and ancestry

The existing published #4312, #4311, and #4308 component heads remain preserved through their prior merge commits recorded in `input-manifest.json`. The captain additionally authorized these exact, provisional local no-mistakes pipeline heads:

| Issue | Exact component head | Local source (read-only) | Preserving merge |
| --- | --- | --- | --- |
| #4305 structured bodies | `55ddb650aa5594ddd156b0939cb1df6027a31d56` | `.../e56f7e7b3cf6/01M0DY0HM9HNZVNKJ2J9Z9SCG7` | `0eb98d3844da7b48d0ca27f51ba7deb46d8f5d1b` |
| #4303 reverse ETL | `e7f474375af969555efd82f684ad6d0b8a26cfc0` | `.../e56f7e7b3cf6/01M0DYNQ9HSJBYS9YQ4MJR4JGR` | `808896a28873c5f0479fa10e2f798da56f885b5e` |

Each SHA was fetched directly from the specified local worktree without modifying either worktree. No older remote ref was used as a substitute. The two heads are merge parents of the listed commits, preserving their complete histories.

## Converged overlap

The shared direct-write/requester seam now retains complete typed provider receipts (status, repeatable headers, raw text or base64 body bytes, decoded body, operation, and path) for declared reverse-ETL actions, including terminal errors. Printable diagnostics remain secret-safe. Declaration-owned headers remain typed and preview-bound; structured body, query, and path bindings are materialized together. GraphQL treats omitted or JSON content type as its declared JSON protocol and preserves explicit text/binary results byte-for-byte.

The status-only `HEAD` path continues to return its final terminal 4xx/5xx metadata after normal retry behavior. Ordinary binary/text GET errors retain their existing error/no-output behavior. No generic raw HTTP method, path, header, body, or action authority was added.

## Checkpoint state

Focused combined checks passed from the composite: full `internal/app`, `internal/connectors/engine`, `connsdk`, `commandrunner`, focused persisted reverse/structured/status CLI cases, focused `synctransport`, focused source-import/generator cases, `go build ./cmd/pm`, `connectorgen validate`, `surface-sync --check`, documentation validation through the built binary, `connector-boundary`, and the generated website data test. The exact commands and assertions are in `evidence-manifest.json`.

The composite is intentionally **unpublished**: no push, pull request, merge to `main`, component-PR mutation, or no-mistakes run was performed. `REVIEW-CONVERGENCE.md` is the independent deep-review charter. The machine-readable input/evidence manifests identify the exact inputs, local provenance, and focused command results without credentials. The temporary locally-built `pm` binary was moved recoverably to Trash after the checks.
