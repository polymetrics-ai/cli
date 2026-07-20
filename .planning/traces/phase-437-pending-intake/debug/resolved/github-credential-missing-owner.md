---
status: resolved
trigger: 'credentials test github-pr466 fails with "github: resolve check path: interpolate: unresolved key owner in config"'
created: 2026-07-20
updated: 2026-07-20
mode: diagnose_only
---

# Symptoms

- Expected: `pm --root <test-root> credentials test github-pr466 --json` validates the stored
  GitHub credential configuration and reaches the connector check when all required non-secret
  configuration is present.
- Actual: the command returns an `internal_error` while resolving the GitHub check path because
  `config.owner` is absent.
- Error: `github: resolve check path: interpolate: unresolved key "owner" in config`.
- Timeline: observed during the current local PR #466 test run; prior success is not established.
- Reproduction: run the reported command against the user's isolated test root.

## Current Focus

- hypothesis: confirmed — the stored credential record lacks GitHub's required `owner` key; the
  checked-in connector-specific examples still use the removed `repository=OWNER/REPO` shape,
  credential creation does not enforce the bundle's required config, and test misclassifies the
  resulting configuration failure as internal.
- test: compare the GitHub bundle spec/templates, credential creation/help examples, sanitized
  credential metadata, and focused tests without reading secret values or contacting GitHub.
- expecting: determine whether `owner` (and likely `repo`) must be supplied when the credential is
  created and whether schema validation should have rejected the incomplete record earlier.
- next_action: user can create a replacement credential with separate `owner` and `repo`; product
  correction remains pending explicit implementation authorization.
- reasoning_checkpoint: root cause reproduced in an isolated root without secrets or network IO;
  no production fix was authorized.
- tdd_checkpoint: TDD mode is enabled; no RED test or production edit will be made during diagnosis.

# Evidence

- timestamp: 2026-07-20T12:37:48+05:30
  observation: GitHub `spec.json` requires `owner` and `repo`; the check request interpolates
    `/repos/{{ config.owner }}/{{ config.repo }}`.
- timestamp: 2026-07-20T12:37:48+05:30
  observation: `pm connectors inspect github` lists `owner` and `repo` under configuration but its
    generated credential examples still pass `repository=OWNER/REPO`.
- timestamp: 2026-07-20T12:37:48+05:30
  observation: a disposable PR #466 root accepted a GitHub credential containing only
    `repository=octocat/Hello-World` plus `auth_type=public` and exited 0.
- timestamp: 2026-07-20T12:37:48+05:30
  observation: testing that disposable credential reproduced the exact
    `resolve check path ... unresolved key "owner"` error, emitted an `internal_error`, and exited
    1 before any network request.
- timestamp: 2026-07-20T12:37:48+05:30
  observation: `runCredentialsAdd` and `App.AddCredential` validate path safety but never validate
    the config object against the connector bundle's required keys; `App.TestCredential` reaches
    `engine.Check`, where path interpolation is the first failing boundary.

# Eliminated

- hypothesis: invalid or expired GitHub PAT.
  reason: failure occurs while interpolating the local request path before authentication or
    network IO.
- hypothesis: GitHub API outage, permission failure, or rate limit.
  reason: no HTTP request is constructed until required path config resolves.
- hypothesis: PR #466 binary was stale.
  reason: reproduction used a freshly built binary at PR head
    `26f98a72419010b961b5b8378ef4a695b0c0a06f`.

# Resolution

- root_cause: The credential's non-secret config does not contain the GitHub bundle's required
    separate `owner` and `repo` keys. Stale checked-in examples direct users to the legacy
    `repository=OWNER/REPO` key, credential creation persists that incomplete config because it
    does not enforce bundle requirements, and credential testing exposes the late interpolation
    failure as an internal error.
- fix: User-side recovery is to create and test a replacement credential using
    `--config owner=OWNER --config repo=REPO` while keeping secrets in an environment reference.
    Product work should update all stale examples, validate required bundle config before vault
    persistence/effects, and map missing configuration to an actionable validation error.
- verification: Exact stale-example reproduction confirmed add exit 0 followed by test exit 1;
    bundle spec, check template, help output, and call chain all agree on the root cause.
- files_changed: diagnosis artifacts only; no production source, generated docs, or tests changed.
