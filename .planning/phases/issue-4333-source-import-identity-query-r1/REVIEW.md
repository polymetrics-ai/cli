# Code review — source-import identity-bearing artifact query

## Scope reviewed

- `cmd/connectorgen/batch_materialize.go`
- `cmd/connectorgen/sourceimport.go`
- `cmd/connectorgen/sourceimport_test.go`
- `docs/migration/conventions.md`

## Security review

- Query admission is default-deny. Only a parsed v3 REST document artifact
  with `identity_query:true` creates the internal allow policy.
- The locked query is bounded with the existing citation key-count, length,
  duplicate-key, control-character, and credential-shaped-key checks before
  a request can be constructed.
- HTTPS, userinfo, fragment, ordinary-host, literal-public-IP, DNS resolution,
  public-dial, redirect, byte, and digest protections are still applied.
- No connector name, runtime URL/query argument, credential input, or generic
  request capability was introduced.

## Correctness review

- The cache previously discarded artifact declaration context by calling
  `source.Fetch(ctx, artifact.SourceURL)` directly. It now routes through
  `fetchSourceImportArtifact`, allowing the HTTP implementation to use the
  lock-owned policy.
- Cache keys remain content-addressed by locked SHA-256; output projection
  does not serialize the declaration, so absent/false default locks remain
  byte-identical.
- Generic batch materialization retains its wrapper functions and therefore
  the existing no-query behavior.

## Findings

No actionable findings.
