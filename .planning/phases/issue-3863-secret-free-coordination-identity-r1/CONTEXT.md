# CONTEXT — issue #3863 secret-free credential coordination identity

## Phase mapping

Issue #3863 is the first dependency-ordered child of parent #3862. It supplies only the
identity input consumed later by #3754 (rate registry), #3865 (authentication cohort fence),
#3867 (rate-limit park/resume), and #3708 (optional authenticator cache). It does not implement
their registries, request admission, fencing, scheduler, cache, connector executor, or sync
transport behavior.

## Evidence that a new identity is required

- `CredentialMeta.ID` is stable but credential/connector-owned at
  `internal/app/types.go:16-25`; `resolveEndpoint` rejects a credential for another connector at
  `internal/app/app.go:1736-1754`.
- `connectors.RuntimeConfig` currently carries config, secrets, approval revision, configuration
  digest, approval scope, payload hashes, and secret-store plumbing, but no coordination identity
  at `internal/connectors/connectors.go:101-111`.
- `CredentialRevision` serializes the secret map before authenticating it at
  `internal/connectors/approval.go:181-205`. It remains approval/rotation evidence and must never
  become a cohort or rate key.

## Locked design

- Each credential receives one durable, randomly generated non-secret binding ID on creation.
  The protected binding is stored in project state keyed by credential ID, not in ordinary
  credential JSON, list output, inspect output, logs, events, or runtime state.
- Credential metadata records declared `provider_family` and `auth_profile`. A new unlinked
  credential gets an isolated binding. Compatibility is exact equality of both declarations;
  a cross-connector link additionally requires both declarations to be supplied explicitly.
- Explicit linking is by credential name, never by showing or accepting a raw binding ID. It is
  available when adding a credential and for existing credentials. Linking rejects a mismatched
  provider family or auth profile before changing state.
- `connectors.CoordinationIdentity` has no binding or secret accessor. It exposes only an opaque
  authentication-cohort projection and a method that derives an opaque rate-scope projection from
  an explicit policy ID, supported scope kind, and declared non-secret subject. Domain-separated
  keyed hashes make the two projections non-interchangeable.
- No rate scope is inferred. An absent policy declaration yields no rate-scope key; an incomplete
  or unsupported scope is rejected without a fallback. The rate projection includes provider
  family, policy ID, kind, and subject so #3754 can use it alongside its `(connector, policy ID,
  opaque scope)` registry shape without assuming that a credential/token is the budget.
- Existing credentials are migrated on open to a fresh isolated binding and connector-scoped
  default declarations. Migration never reads the vault and never changes approval revisions.
- `RuntimeConfig` carries only the opaque identity. `CredentialRevision` remains unchanged and
  secret rotation cannot reset the resulting auth/rate identity.

## Scope and ownership guard

- Owned paths: `internal/app/{types,app,coordination_identity}*.go`,
  `internal/connectors/{connectors,coordination_identity}*.go`, targeted `internal/cli` credential
  handling/tests/help, `docs/cli/credentials.md`, the CLI-reference website source and its generated
  data, and this phase directory.
- Do not edit `internal/connectors/commandrunner/runner.go`, connector bundle schemas,
  `cmd/connectorgen`, engine policy/requester behavior, #3754 registries, or #3865/#3867 runtime
  behavior. No live provider or credentialed check is permitted.

## GSD execution note

`scripts/gsd doctor`, all required `scripts/gsd sources` commands, and
`go run ./cmd/agentcontractgen check` passed. The generated discuss and plan prompts are executed
inline: the issue and cited research fix the decisions, while the canonical worker contract forbids
role spawning. This fallback preserves TDD, verification, review, and human gates.

## CLI help/docs/website parity

The explicit safe link and declaration inputs change `pm credentials`; runtime help, bare namespace
help, `docs/cli/credentials.md`, `website/content/docs/cli-reference.mdx`, generated website data,
and targeted CLI tests are in scope. JSON output must expose non-secret family/profile declarations
only; it must not expose a binding preimage or either opaque key.
