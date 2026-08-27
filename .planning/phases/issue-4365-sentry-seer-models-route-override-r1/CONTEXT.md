# Issue #4365 — Sentry Seer Models route override

## Fixed decisions

- Scope is exactly one Sentry direct-read: source operation
  `sentry.rest.listSeerModels`, `GET /api/0/seer/models/`, cited by the
  retained Sentry OpenAPI lock at
  `cmd/connectorgen/testdata/issue4329/sentry-operation-source-lock.json`.
- Use the existing closed `streams.json` named-route mechanism. The only
  selected route is connector-owned; it resolves only `{{ config.base_url }}`
  and its fixed source-traced `/api/0/...` path. No caller receives a route,
  base URL override, HTTP method, or raw path input.
- The stable command identity is `seer list-models`; it binds one
  `rest_read` operation and one `api_surface` endpoint.
- This is a direct PR to `main`, not a certification claim and not a live
  provider test. Credential absence must stop before transport I/O.

## GSD execution note

`scripts/gsd doctor`, every required `sources` lookup, and the generated
`discuss-phase` and `plan-phase --tdd` prompts were run. The canonical
single-worker contract and this non-Pi runtime do not provide compatible
isolated workers, so discussion and planning are executed inline. This file
and the sibling phase artifacts are the durable manual fallback record.
