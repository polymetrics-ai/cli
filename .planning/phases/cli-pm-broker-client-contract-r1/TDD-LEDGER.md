# TDD Ledger

## Slice: PM Broker `/v1` HTTP client/transport contract

Planned red tests before production code:

- `TestHTTPClientLoopbackAndRemoteShareTypedSemantics` should assert loopback and remote/container HTTP endpoints use the same typed compatibility, list, get, and create execution-plan semantics through the synthetic broker round-tripper.
- `TestHTTPClientAuthenticationCorrelationIdempotencyAndDigestTransport` should assert typed requests use an explicit auth seam, safe correlation IDs, `PM-Broker-API-Version`, `Idempotency-Key`, and immutable execution-plan digest transport.
- `TestHTTPClientRejectsUnsafeEndpointHostOriginAndAmbientCookies` should assert credentials/userinfo/query/fragment/unsupported schemes are rejected with redacted errors, fake broker rejects unsafe Host/Origin, and state-changing requests cannot rely on cookies.
- `TestHTTPClientStructuredErrorsRateLimitsAndCompatibilityNegotiation` should assert safe structured errors, rate-limit metadata, compatibility negotiation, and exact HTTP 426 `incompatible_contract_version` behavior.
- `TestHTTPClientSafetySurfaceNoCredentialsNoGRPCNoGenericEscape` should assert diagnostics/log-style fields are redacted, no credentials are accepted in URLs, no public gRPC/socket path exists, and no generic request escape methods are exposed.

Red evidence:

- `go test ./internal/pmbroker/contract/v1` failed before production code with undefined typed HTTP client API (`NewHTTPClient`, `WithClientRoundTripper`, `RoundTripper`, `Pagination`, `Authorization`), proving the HTTP transport lane was absent.

Green evidence:

- `go test ./internal/pmbroker/contract/v1` passed after implementing typed HTTP client, synthetic broker HTTP handler, endpoint/auth/correlation/pagination/idempotency/digest/error/rate-limit coverage.
- Review hardening added coverage and fixes for endpoint path rejection, HTTPS/non-loopback HTTP policy, sanitized diagnostics for unsafe contract-version values, typed nil auth/correlation adapter failures, allowed Host pinning, global cookie rejection, and digest-header mismatch rejection.
- `go test ./internal/pmbroker/...` passed.
- `git diff --check` passed.
- `go test ./...` passed.
- `go vet ./...` passed.
- `go build ./cmd/pm` passed.
- `make verify` passed.

## Review-fix slice

Planned red coverage before production edits:

- `NegotiateCompatibility` should reject a configured client contract version that the broker compatibility response does not support.
- `CreateExecutionPlan` should reject shape-valid broker plans whose identity boundary, idempotency key, or intent connector connection ID differs from the submitted request.
- `ExecutionPlanRequest.Validate` should reject mismatched top-level and intent connector connection IDs.
- `ConnectorConnection.Validate` should accept valid `/v1` enum values beyond the synthetic fixture while still rejecting arbitrary connector kinds, statuses, and write modes.
- `.github/workflows/pr-issue-guard.yml` should keep the linked-issue guard enabled for `fm/*` pull requests unless a stronger validation-mirror marker is added in a later CI design.

Preflight evidence:

- `scripts/gsd doctor` passed.
- `scripts/gsd prompt programming-loop init --phase pmbroker-contract-review-fixes --dry-run` failed with `unknown GSD command: programming-loop`; manual-GSD fallback used.
- `scripts/gsd prompt gsd-quick "PM Broker contract review findings fix round"` generated the repo-local quick-task prompt.

Green evidence:

- Initial focused `go test ./internal/pmbroker/contract/v1` exposed a new parallel-test fixture aliasing bug in the execution-plan mismatch table; the test now deep-copies the nested intent before mutation.
- `gofmt -w internal/pmbroker/contract/v1` passed.
- `go test ./internal/pmbroker/contract/v1` passed.
