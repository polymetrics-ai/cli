# Context — source-import identity-bearing artifact query

## Confirmed diagnosis

`cmd/connectorgen/batch_materialize.go:parseBatchArtifactURL` rejects every
artifact query. `validateSourceImportPublishedURL` strips a copy only to
validate a provenance citation; the citation is never passed to Fetch. The
fetch path is `fetchSourceImportArtifact` → cache fetcher →
`httpSourceImportFetcher`.

## Decision

Use a v3 document-owned artifact declaration:

```json
"artifact": {
  "source_url": "https://provider.example/openapi?version=2026-08-01",
  "identity_query": true
}
```

`identity_query` is an explicit boolean opt-in. It is valid only when the
locked artifact URL has a non-empty query, and that fixed query is validated
with the citation policy's length, key-count, key/value, duplicate-key, and
credential-shaped-key guards. The source URL remains the lock-owned request
identity; no caller can supply a query or URL.

An omitted/false declaration retains the existing no-query artifact policy.
The separate v3 `published_source.source_url` continues to accept a bounded
capture/provenance query, which is never fetched and is stripped only for
base-URL validation.

## Security boundary

The identity opt-in changes only the query component of a lock-owned v3
artifact. HTTPS, no userinfo, no fragment, ordinary-host, literal public-IP,
DNS public-address, request-dial, redirect, byte-limit, and digest checks
remain in force. Generic batch materialization and all legacy source locks
continue to use the no-query policy.

## GSD execution note

This issue is not a numbered roadmap phase and this worker has no compatible
Pi isolated-agent runtime. The generated `discuss-phase`, `plan-phase --tdd`,
`execute-phase`, `verify-work`, and `code-review` prompts are executed inline
and evidenced in this directory.
