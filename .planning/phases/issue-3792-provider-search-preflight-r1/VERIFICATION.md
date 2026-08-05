# Verification — issue #3792 provider-search runtime preflight

Status: blocked; no production change has been made.

## Blocking assessment — 2026-08-06

A strict, no-network candidate was tested and then fully reverted. Its real
runtime sweep was:

```text
go test ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1
178 of 1239 commands marked "implemented" fail runtime Preflight
```

The candidate would correctly make preflight compare the command with the
loaded operation's kind, method, path, cap, and policy. It cannot land as an
independent #3792 change because:

1. The present shared read contract exposes execution only
   (`internal/connectors/connectors.go:159-161`). The analogous typed
   no-network metadata contract exists only for writes
   (`internal/connectors/connectors.go:195-220`). Adding the corresponding
   shared read boundary requires the #3740-overlap files named in #3792:
   `connectors.go` and `engine/connector.go` (the current engine adapter is
   at `internal/connectors/engine/connector.go:132-145`).
2. Making a commandrunner-local contract mandatory instead correctly rejects
   native readers which execute direct reads but cannot safely be treated as a
   declarative bundle: Amazon SQS has its closed native executor at
   `internal/connectors/native/amazon-sqs/direct_read.go:15-45`; Ashby
   exposes the generated command surface at
   `internal/connectors/native/ashby/engine_delegate.go:45-48` while adding
   its own response semantics after engine execution at lines 95-105. Making
   that contract optional would leave claims unproven and violates the issue's
   acceptance criterion.
3. Existing operation and command response policies intentionally drift:
   GitHub's operation declares `"json"`
   (`internal/connectors/defs/github/operations.json:211-222`), while its
   implemented command declares `"json_redacted"`
   (`internal/connectors/defs/github/cli_surface.json:2404-2417`).
   `surface-sync` deliberately preserves a supported direct-read command
   policy rather than matching the operation
   (`cmd/connectorgen/surfacesync.go:277-303`). Exact policy comparison
   therefore produces broad failures. The only current direct-read policies
   are redacting/specialized values
   (`internal/connectors/engine/direct_read.go:329-344`); the required
   full-content policy belongs to #3771/#3852 and must not be invented here.

There are no production `provider_search` declarations yet, but that does
not make a provider-search-only check deliverable: #3792 requires the shared
eligibility contract to accept both `rest_read` and `provider_search`, and
requires the real all-implemented sweep to pass. Narrowing the check would
repeat the generic-reader defect for the rest of the existing fleet.

The committed R1 commandrunner test remains the RED evidence. No candidate
production code or extra provider-search declaration remains in the worktree.

## Completion checklist

- [x] New commandrunner RED test observed against the original generic-reader preflight.
- [ ] Loaded engine operation preflight rejects unsupported kind, mismatched method/path/policy,
  absent operation, and non-positive cap before any request.
- [ ] Existing bounded provider-search `httptest` coverage stays green.
- [ ] Real implemented-command preflight sweep passes. Blocked: strict candidate reports 178
  failures; see assessment above.
- [ ] Focused/package tests, formatting, vet, build, and applicable individual repository gates pass.
- [ ] No declarations, schemas, validators, capabilities, redaction paths/policies, CLI/help/docs,
  #3740 overlap paths, or credentials/live calls changed.
- [ ] GSD verify/code review executed inline and recorded; no-mistakes deferred to firstmate's
  post-commit instruction.
