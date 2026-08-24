# Inline code review — issue 4347 source-lock usability

## Route

The repository's canonical contract forbids spawning workflow roles in this
environment, so the `verify-work` and `code-review` prompts were resolved and
executed inline. Claude automatic review remains the primary external route
once the main-targeted PR opens.

## Findings and disposition

- **Accepted before review completion:** the initial retain-only loader still
  reached `httpSourceImportFetcher.FetchArtifact`, whose import-form validator
  could reject an unknown future form. `FetchRetainArtifact` now uses only
  retain validation, and `TestSourceRetainHTTPFetchDoesNotRequireImportFormValidation`
  proves the separation.
- **Accepted follow-up evidence:** a 403 or TLS refusal can mean provider
  automation blocking, not source absence. `BOT-BLOCK` now directs the reader
  to a browser capture or provider-owned repository, with red/green coverage
  in `TestSourceRetainReportsBotBlockBeforeWrongSourceOrDrift`.
- **Accepted CI root fix:** CI's
  `TestSourceImportRetainedArtifactRejectsMissingAndMismatchedCopies` passes
  unchanged on clean `060bb7864`, proving the failure was branch-owned. Byte
  identity now retains its legacy `locked bytes and SHA-256` diagnostic, while
  canonical identity retains the distinct fetched-byte provenance diagnostic.
- **No remaining actionable findings:** lock loading stays connector-owned;
  redirects, denial, and drastic source collapse are classified before drift;
  canonical identity preserves fetched-byte provenance; no source lock is
  rewritten.

## Evidence

`go test -timeout 10m ./cmd/connectorgen -run '^TestSourceRetain'`,
`go vet ./cmd/connectorgen`, the independent Makefile gates, and the eight
built-binary retain commands passed. The full package command was not used as
final evidence because its unrelated certification matrix fixture recompiles
the generator under a fresh cache; see `VERIFICATION.md`.
