---
coverage:
  - id: D1
    description: Exact Sentry Seer Models source projection has one stable CLI identity and endpoint.
    verification:
      - kind: unit
        ref: internal/connectors/engine/sentry_seer_models_route_test.go TestSentrySeerModelsSourceBoundRoute
        status: pass
      - kind: unit
        ref: cmd/connectorgen/sentry_seer_models_source_binding_test.go TestSentrySeerModelsSourceProjectionKeepsExactRouteBinding
        status: pass
    human_judgment: false
  - id: D2
    description: Identity drift is rejected and base/path joining preserves the exact endpoint.
    verification:
      - kind: unit
        ref: internal/connectors/engine/sentry_seer_models_route_test.go TestSentrySeerModelsRouteRejectsIdentityDriftBeforeProviderIO
        status: pass
      - kind: unit
        ref: internal/connectors/engine/sentry_seer_models_route_test.go TestSentrySeerModelsRoutePreservesPathAcrossBaseSlashForms
        status: pass
    human_judgment: false
  - id: D3
    description: The generated CLI stops before provider I/O when no credential is configured.
    verification:
      - kind: unit
        ref: internal/cli/cli_test.go TestSentrySeerModelsCommandStopsBeforeProviderIOWithoutCredential
        status: pass
      - kind: other
        ref: ./pm sentry seer list-models --root <fresh-project>
        status: pass
    human_judgment: false
---

# Issue #4365 execution summary

Implemented exactly one new Sentry direct-read projection: the source-locked
`sentry.rest.listSeerModels` operation, `GET /api/0/seer/models/`, is now the
typed `sentry.seer_models_list` operation and stable `pm sentry seer list-models`
command. It uses the connector-declared `sentry_api_v0` route only.

The implementation is declaration-only: Sentry operations, CLI surface, route,
API-surface coverage, and generated artifacts changed. No generic provider route,
caller-selected method/path/base, raw HTTP command, shared engine code, or hook was
introduced.

## Lifecycle fallback

`discuss-phase`, `plan-phase --tdd`, and `execute-phase` prompts were resolved via
the project adapter. The available runtime does not supply compatible isolated Pi
workers, and the Firstmate assignment forbids delegation, so execution, verification,
and review use the documented inline/manual fallback. Red and green evidence is in
`TDD-LEDGER.md`; verification and review records are siblings of this file.

## Generated parity

`surface-sync`, declaration admission, operation evidence, endpoint ledger,
certification subject, Sentry manual/skills/catalog, website connector data, and the
root help transcripts were regenerated from the final declaration.
