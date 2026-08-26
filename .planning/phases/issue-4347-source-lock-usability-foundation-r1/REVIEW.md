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
- **Accepted R2 exact-SHA repair:** a populated v3 `source_documents`
  inventory cannot enter the evidence-only skipped/dynamic projection. The
  reader now reaches the strict source-import decoder, which rejects the
  contradictory legacy `state` field before it can suppress REST rows. An
  empty v3 provider-absence record remains deliberately separate.
- **Accepted R2 source classification repair:** `canonical_json` itself is a
  JSON MIME expectation. A non-JSON MIME, malformed canonical JSON, or a
  duplicate member is wrong-source before identity comparison. The canonical
  parser reuses the strict recursive JSON validator, so it never accepts
  implementation-dependent last-member semantics; offline verification calls
  duplicate input invalid/ambiguous rather than asking for a re-pin.
- **Accepted R2 provenance repair:** canonical locks compare strict canonical
  JSON only. A raw size/digest change from formatting or minification remains
  manifest provenance, while byte locks retain exact size/digest and the
  drastic-collapse wrong-source classifier.
- **Accepted R2 citation repair:** a rendered citation fragment must exactly
  name the operation's locked extraction location; any supplied capture
  binding is independently checked even when that fragment is present.
- **No remaining actionable findings:** review of the R2 production diff,
  red/green tests, full changed package suite, generated checks, clean
  structural boundary, and release checks found no route escape, lock rewrite,
  silent re-pin, or source-import relaxation.

## Evidence

The focused R2 command, `go test -timeout 10m ./cmd/connectorgen -run
'^TestSourceRetain' -count=1`, `go test -timeout 20m ./cmd/connectorgen
-count=1` (279.084s), `go vet ./cmd/connectorgen`, all applicable independent
Makefile gates, clean generated checks, and the prior eight built-binary retain
commands passed. The exact commands and the retained-artifact preservation
boundary are recorded in `TDD-LEDGER.md` and `VERIFICATION.md`; no aggregate
`go test ./...` was run.
