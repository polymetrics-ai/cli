# Context — source-backed request encoding foundation

## Task Delivery Header

- Issue: Closes #4367 — add source-backed request encoding foundation.
- Base branch: `main` at `cf29d302c13f7fcd340d31ad6dc27872880ccf42`.
- Merges into: `main`.
- Delivery: Direct PR open against `main`, committed and locally verified; read
  the API-reported PR base after opening. The merge remains human-gated.
- Working branch: `fm/cli-request-encoding-foundation-r1`.
- Task: Add only the shared closed source-backed encoder admission for the
  Batch 1 request-encoding cohort: 50 multipart/form-data operations and one
  application/x-www-form-urlencoded operation. Preserve source ID, source URL
  and location, method/path, media type, part names, requiredness, and binary
  semantics. Keep all non-encoding gaps deferred before provider I/O.
- Verification: Source importer/projection tests, engine no-network request
  spy tests, per-connector source-import/check/validate/surface-sync checks,
  focused CLI checks, full `cmd/connectorgen` tests, generator checks, direct
  PR checks, diff check, and frozen exact-SHA review.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Only declared request encoders are admitted | live | Source descriptor and bundle validation accept JSON/form/multipart only; unknown media or malformed metadata fails before transport creation. |
| Multipart preserves closed text/file wire contract and bounds | live local HTTP | A loopback spy parses the request and observes exact named text/file parts, declared content type, and no request on over-limit or invalid input. |
| URL-encoded values use declared serialization | live local HTTP | A loopback spy receives provider-declared key/value serialization, including repeatability; missing, duplicate, and unknown inputs stop locally. |
| Batch 1 census stays exact and cited | fake — retained-source fixture | The fixture is a hermetic, source-derived reimport of the 47 GitLab, 2 Sentry, 1 Asana, and 1 Jira records. Network retrieval is deliberately excluded; it asserts all source identities, URLs, locations, method/path, and selected media. |
| Unrelated typed schema gaps stay deferred | fake — importer fixture | A Batch 1 form operation with a separately unsupported typed schema retains its cited schema foundation despite encoder admission. |
| No arbitrary HTTP/body escape hatch is exposed | live | Bundle, projection, command metadata, and engine tests reject arbitrary content types, paths, methods, fields, and raw bodies. |

## Decisions and boundaries

- The authoritative future manifest belongs to #4364 and is not copied,
  cherry-picked, or used as a generated input in this branch. The reconciliation
  test independently derives the 51 rows from retained source-shaped fixtures
  and source-lock citations.
- This is a shared foundation issue, not a connector adoption lane. It will not
  invent `operations.json`, `writes.json`, `streams.json`, binary, ETL, or
  reverse-ETL declarations. Exact command materialization/admission remains
  #4364 and provider adoption work.
- Existing direct-operation form/multipart support is reused, not generalized.
  New admission is explicit and closed. A method never implies a six-lane
  classification.
- Source schema composition, dynamic objects, selection among multiple media,
  and other unsupported parts remain their existing source-cited
  `missing_foundation` reasons. No raw JSON or arbitrary body flag substitutes
  for those contracts.

## GSD execution note

The adapter prompts for `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
`verify-work`, and `code-review` are executed inline. This issue-scoped phase
is outside the numeric roadmap and the worker brief prohibits role spawning.
No no-mistakes interaction is authorized.

