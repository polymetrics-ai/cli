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
