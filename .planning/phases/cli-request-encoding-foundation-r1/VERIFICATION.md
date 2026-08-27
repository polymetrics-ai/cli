# Verification checklist — source-backed request encoding foundation

## Behavior and provenance

- [ ] Source-derived 51-row reconciliation proves 50 multipart / 1 urlencoded
      and all source citations, methods, paths, identities, and media.
- [ ] Closed descriptor rejects malformed or mismatched media/part contracts.
- [ ] Multipart text/binary/request-limit no-network spy coverage passes.
- [ ] Form required/missing/duplicate/unknown/style no-network spy coverage passes.
- [ ] Unrelated schema/media gaps remain exact cited `missing_foundation`.
- [ ] Promotion report names credential-bound runnable count, residual count,
      and every primary residual reason; no method-only classification.

## Required local gates

- [ ] `go run ./cmd/connectorgen source-import --check` for every touched connector.
- [ ] `go run ./cmd/connectorgen validate` and `go run ./cmd/connectorgen surface-sync --check`.
- [ ] Focused `cmd/connectorgen`, engine, commandrunner, and CLI tests.
- [ ] `go test -timeout 20m ./cmd/connectorgen -count=1`.
- [ ] `gofmt`, package `go vet`, `go build ./cmd/pm`, generated-artifact gates,
      individual applicable `make verify` checks, and `git diff --check`.
- [ ] If a command is promoted, a newly initialized isolated project invokes
      built `pm` with valid typed input and no credential and receives exactly
      `error: missing --credential` with zero provider I/O.

## Lifecycle and delivery

- [ ] Execute/verify/code-review prompts executed inline with evidence.
- [ ] Fresh exact-SHA independent review complete; all findings dispositioned.
- [ ] Direct PR pushed/opened, API-reported base equals `main`, no merge.

