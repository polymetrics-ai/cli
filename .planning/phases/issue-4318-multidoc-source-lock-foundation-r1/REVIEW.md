# Code review — issue #4318 multi-document source-lock foundation

## Route and fallback

`scripts/gsd prompt code-review issue-4318-multidoc-source-lock-foundation-r1` was resolved.
The normal GSD reviewer role is not compatible with this repository's single-worker contract, which
forbids lifecycle role spawning here, so this is the required standard-depth inline/manual fallback.

## Review scope

- `cmd/connectorgen/sourceimport.go`: versioned strict decoding, inventory/bounds, artifact
  fetch/caching, source identity, provenance, and YAML normalization.
- `cmd/connectorgen/sourceprojection.go` and `surfacesync.go`: schema-3 consumer handling.
- Focused behavioral tests, migration documentation, and delivery evidence.

## Findings

No Critical, Warning, or Info findings.

The manual review verified that v1/v2 retain their original parser/import/descriptor path; v3
retrieves only queryless artifacts; published query URLs are citation-only and bounded; duplicate
digests have a mutex/channel single-flight with verified bytes on every consumer; and projection
comparison includes the new document and published provenance fields. The YAML correction performs
duplicate detection on normalized scalar keys and still rejects compound/custom-tagged keys.

## Review evidence

- `git diff --check` passed.
- `go test -race -count=1 -timeout 20m ./cmd/connectorgen -run '^TestSourceImportVersion3SynchronizesDuplicateArtifactDigests$'` passed.
- `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestSourceImportPreservesFrozenGitHubArtifacts$'` passed.
- Full `make verify` passed with zero lint findings.

## External review routing

PR #4320 is non-draft, targets `main`, and its API-reported author association is `MEMBER`.
No Claude automatic workflow run was created after the open trigger, so the repository's required
fallback is one manual `@claude review` request. That request is pending; Copilot is not a parallel
route and has not been requested.
