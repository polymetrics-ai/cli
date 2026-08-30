# Issue #4365 — Sentry exact Seer Models route override plan

## Task Delivery Header

- Issue: Closes #4365 — feat(sentry): add exact Seer Models route override
- Base branch: `main` (`origin/main@cf29d302c13f7fcd340d31ad6dc27872880ccf42`)
- Merges into: `main`
- Delivery: One ordinary PR open against `main`, with local gates recorded,
  CI terminal green, an independent exact-head audit requested, and no merge.
- Working branch: `feat/4365-sentry-seer-models-route-override`
- Task: Materialize only `sentry.rest.listSeerModels` as a typed Sentry
  direct-read, bound to one connector-declared route/base identity and the
  stable `seer list-models` CLI path. It must preserve `GET
  /api/0/seer/models/`, reject all identity drift, and preflight for a missing
  credential before any transport I/O.
- Verification: Focused red/green/refactor tests; source/bundle validation;
  `surface-sync`, admission/evidence/reconcile and connector-boundary gates;
  generated docs/help checks; `go vet`, build, and scoped tests. Build and
  invoke the exact CLI against an empty credential context while a local spy
  asserts zero provider requests.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Exact source operation projects to one CLI path and one endpoint | live | A focused test loads the Sentry bundle and asserts `seer list-models` resolves exactly once with the exact source ID, GET method, `/api/0/seer/models/`, operation binding, and endpoint ledger row. |
| The route/base contract is closed | live | Table-driven mutations of source ID, route, base, method, and path are rejected before provider I/O; no generic command value can supply any of them. |
| Slash joining preserves source identity | live | A local route test uses base URLs with and without trailing slash and asserts the actual received request remains exactly `/api/0/seer/models/`. |
| Credential boundary prevents provider I/O | live | The built `pm sentry seer list-models` command in a fresh initialized project returns exactly `error: missing --credential`; a transport spy records zero requests. |
| Source/declaration/generated artifacts agree | live | `connectorgen validate`, `surface-sync --check`, declaration admission, operation evidence, surface reconcile, and the embedded endpoint ledger check read the actual declarations. |
| Documentation/help remain coherent | live | Runtime command and namespace help, generated Sentry manual/skill references, docs/website grep, and docs check show the one command or record a generated-surface exemption. |

## Foundation Check

| Need | Required proof | Status |
| --- | --- | --- |
| Typed direct-read executor | Existing `OperationDirectRead` route resolver and commandrunner preflight execute the declared REST operation. | Verify before promotion. |
| Connector-owned route | `streams.json.base.routes` contains the selected named Sentry base and the operation references only that name. | Implement in this connector slice. |
| Source provenance | `source_operation` repeats exact ID/method/path from the retained Sentry source lock; the provider citation remains on the operation. | Implement and test. |
| Warehouse/reverse/certification flow | Not claimed by this direct-read-only slice. | Not applicable. |

## TDD plan

1. **Red — happy projection:** add a Sentry-focused bundle/command test that
   expects the one source-bound operation, stable CLI path, exact endpoint,
   route selection, and generated endpoint ledger. It must fail because
   Sentry currently has none.
2. **Red — closed mismatches:** add named source ID, route, base, method, and
   path mutation cases plus a no-I/O transport assertion. They must show no
   source-bound Sentry route contract exists yet.
3. **Red — slash edge:** add a local transport test for trailing/non-trailing
   declared bases, asserting the received path remains exactly the provider
   path rather than changing prefix or slash semantics.
4. **Green:** add only Sentry `operations.json`, `api_surface.json`,
   `cli_surface.json`, and the connector-owned named route needed for the
   closed binding. Run the canonical generators; do not change shared engine,
   command parsing, or hooks.
5. **Refactor/verify:** keep declaration order stable, regenerate derived
   endpoint/evidence/manual artifacts, run the credential-boundary executable
   proof, docs/help parity, review, and final audit.

## Required skills and parity

Loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, `golang-structs-interfaces`, and
`golang-documentation`.

The CLI parity checklist applies: exact command help, `pm sentry`,
`pm sentry seer`, `pm help sentry`, generated manual/docs, website mirrors or
an explicit generated-surface exemption, machine-readable behavior, and an
invalid action. No new generic command, route flag, or raw HTTP feature is
permitted.
