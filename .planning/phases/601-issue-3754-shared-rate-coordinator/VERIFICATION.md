# Phase 601 verification checklist — #3754

**Status:** planned; no production verification is claimed yet.

## Required behavior

- [ ] Existing `RateLimitRegistry`/`RateLimiter` behavior remains a
      dependency-free `process_local` backend.
- [ ] Matching policies are reserved atomically with one in-flight lease; a
      non-grant consumes neither policy capacity nor lease occupancy.
- [ ] Policy identity contains only a deterministic typed-contract fingerprint
      and #3863 opaque rate scope; no raw credential, revision, subject, URL,
      request body/header/variables, endpoint, or epoch reaches evidence.
- [ ] UDS owner directory/socket modes are 0700/0600; protocol is versioned,
      closed, length-bounded, context/deadline-aware, and has no generic RPC.
- [ ] `Finish` is opaque-lease-idempotent, releases concurrency, retains
      indeterminate consumption, and applies only stricter observations.
- [ ] `require_shared` has no local fallback; missing/dead/old/incompatible
      owners refuse before a transport request.
- [ ] Deadline-too-short refusal and owner crash have zero unapproved sends.
- [ ] Eight actual helpers produce shared 3 grants/5 blocks and local 8
      grants; normal close proves zero socket/run-directory residue.

## Local commands to record after GREEN

- [ ] Focused mandatory suite and focused requester/engine coverage.
- [ ] `go test -race -count=1 -timeout 20m ./internal/coordination`.
- [ ] `gofmt` check, targeted `go vet`, and `go build ./cmd/pm`.
- [ ] Individual non-monolithic gates required by AGENTS.md: tidy-check,
      lint, docs-check, smoke-no-build, agent-contract-check,
      connectorgen-validate, connectorgen-surface-sync, connector-boundary,
      and release-workflow-check.
- [ ] `scripts/verify-gsd-workflow` against the parent base after production
      evidence lands.
- [ ] Generated inline `verify-work`, `code-review`, #3995 equivalent bounded
      supervisor evidence, and terminal child no-mistakes gate.

## Explicit non-applicability

No CLI command/flag/help/manual/website output changes, no provider/GraphQL
code, no connector bundle/generated surface, no external service, and no live
credentialed check is in this phase. The PR records that scope fence rather
than claiming parity work that did not apply.
