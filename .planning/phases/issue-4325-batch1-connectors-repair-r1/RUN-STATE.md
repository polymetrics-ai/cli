# Issue 4325 — Execution State

## Held connector slice

- Target: CircleCI, pending #4328. Its structural work is held in the named
  local stash `issue-4325-circleci-secret-input-foundation-hold` and must not
  ship until the shared secret-input foundation resolves.

## Red evidence

The built baseline binary returned `error: unknown command "circleci"` (exit
2) for `pm circleci operations list`; the current bundle has neither
`operations.json` nor `cli_surface.json`.

## Green evidence

- Source lock refreshed from CircleCI's live OpenAPI: 621321 bytes, SHA-256
  `61c6ce11e8de509948aa3d53dcd0169913f52de20920b130b6a85dea41d66d07`,
  111 method/path operations.
- Source import, bundle validation, and global surface-sync check pass.
- The v2 artifact ledger has 111 cited endpoints: 40 executable bindings and
  71 blocked/disallowed rows. The terminal surface contains nine ETL commands
  and 34 reverse-ETL commands.
- Fresh credential-free binary probes of `contexts list` and `delete context
  apply` reach `missing --credential` rather than an unknown command.

## Blocked shared foundation

- #4328 must resolve the secret `signing-secret` field on CircleCI webhook
  writes. The current generated command surface exposes it as an ordinary
  shell flag; `env_only` is unavailable for declarative reverse-ETL writes.
- Sentry needs a read-only source-projection exemption for supported source
  mutations that have no declared write action. The generated four-command
  stream surface is valid and source-bound, but
  `validateSourceExecutableCoverage` in `cmd/connectorgen/sourceprojection.go`
  (the mutation branch at lines 1943-1948) requires an executable action for
  every mutation without an importer-generated runtime gap. The Sentry source
  has 34 such mutations; this makes a read-only connector with any real CLI
  surface unvalidatable. No connector-local source gap is legitimate because
  `source-import --check` regenerates the descriptor. The shared validator must
  recognize an explicit read-only/source-cited operation policy or generate a
  source-bound refusal with its actual foundation location.
- Vercel is also pending #4329. Its checked-in source lock enumerates 400
  operations (163 reads and 237 mutations), while its 18 declared writes match
  only four source method/path identities and no CLI surface exists. A CLI
  surface would trigger the same all-or-nothing source mutation check, so no
  Vercel declaration was modified.

## Current connector slice

- Target: Docker Hub only.
- Allowed production paths: `internal/connectors/defs/dockerhub/**` and generated
  docs/catalog outputs only where the owning generator updates them.
- Ownership guard: `git diff --name-only` before commit must name no other
  connector. The separate Asana, CircleCI, and Sentry attempts remain only in
  their named foundation-hold stashes.

## Current connector slice (2026-08-25)

- Target: Jira only; materialize exactly `resetUserColumns` at manifest-reserved
  path `api op-798e4bdcb516fc99a56c6b35b2bc97e67b65830a72dc867eeab1bb261c01b320`
  as a source-cited, closed, destructive reverse-ETL action.
- Hold `removeGroup`: the provider makes every selector individually optional
  but documents a 400 when a group is not supplied; the current action schema
  has no source-derived conditional-query projection. Hold `addWatcher`: its
  required `application/json` scalar string is unsupported by the object-only
  JSON action executor at `internal/connectors/engine/write.go:674-692`. The
  exact future paths are held in the independent mapping manifest; no local
  conditional-query, scalar-body, or deferred-command projection is permitted.

## Current connector slice (2026-08-25, Asana)

- Target: the 21 source-cited Asana no-body mutations listed in
  `PLAN.md`'s Asana cohort. Each existing planned command retains its current
  CLI path; a connector-owned action can be complete only when source import
  derives its closed path/query record contract.
- Red observed: `go run ./cmd/connectorgen source-import asana --check`
  reports descriptor drift with `writes=0 cli=0`; importing refreshes only the
  source summaries, verifies all 249 provider operations on the rerun, and
  creates no action/CLI contract by itself.
- Boundaries: do not add a schema, executor, idempotency, missing-status, or
  provider-scope approximation. The 24 existing Asana request-schema gaps and
  all other deferred manifest rows remain source-cited as-is.
- Result: source import projects 21 write and 21 CLI updates; all 21 existing
  canonical commands now pass real preflight through a no-body action. The
  interim stale operation references were removed because they named no
  operation executor; no local executor was introduced.

## Current connector slice (2026-08-25, CircleCI reconciliation)

- Target: 24 current CircleCI direct-write commands already marked
  `availability: implemented` and bound to complete actions. The source
  importer verifies all 111 operations and targeted validation returns zero
  findings.
- Scope: correct only their independent manifest classification after the
  table-driven runtime preflight and binary evidence establish that the
  existing command/action/source endpoint triples are real. Do not modify
  CircleCI connector JSON in this reconciliation.
- Outcome: `source-import circleci --check` verified 111 operations and
  targeted validation returned zero findings. The table-driven runtime test
  passes all 24 source ID / CLI path / action / method / endpoint tuples, and
  a built binary in an isolated credential-free project returned exactly
  `error: missing --credential` for every one. The red evidence corrected the
  recorded action spelling to `remove_u_r_l_orb_allow_list_entry`; no CircleCI
  declaration, source lock, executor, or secret-input policy was changed.
- Manifest result: CircleCI is now `40 runnable / 51 declarable / 20 deferred`;
  Batch 1 is `813 runnable / 1,618 declarable / 1,910 deferred` over the same
  4,341 source operations. The 51 reservations and 20 genuine gaps remain.

## Current connector slice (2026-08-25, Sentry reconciliation)

- Target: 33 current Sentry source paths that are already `implemented` in
  `cli_surface.json` but are recorded as deferred in the manifest. Thirty-two
  bind existing write actions; `projects list` is the existing source-bound
  ETL command.
- Red observed: the manifest still cites `cli-surface-missing-sentry`, despite
  source import verifying 223 operations, targeted validation returning zero,
  and the existing 32-write runtime preflight test passing.
- Boundary: prove every exact command/action/source tuple and credential
  boundary first. Do not alter Sentry JSON or promote the other 111 declarable
  rows or 76 genuine gaps.
- Outcome: source import verified all 223 Sentry operations and targeted
  validation returned zero findings. The expanded runtime test passes all 32
  source-bound write actions and `projects list`; a built binary in an
  isolated credential-free project returned exactly `error: missing
  --credential` for all 33 paths.
- Manifest result: Sentry is now `36 runnable / 111 declarable / 76 deferred`;
  Batch 1 is `846 runnable / 1,585 declarable / 1,910 deferred` over the same
  4,341 source operations. No Sentry connector declaration changed.

## Jira red evidence

- `go run ./cmd/connectorgen source-import jira --check` exited 1 because the
  bundle has no `sources/` directory.
- The existing CLI surface has 590 commands (584 implemented): 292 implemented
  direct reads and 286 implemented write commands. Its `operations.json` has
  only 25 typed contracts (22 REST reads and three binary downloads), so 270
  enabled direct reads have no operation contract. The v1 ledger contains 617
  endpoint rows and does not describe the runnable direct-write actions.
- Live refresh measured Jira's current provider OpenAPI at 2456011 bytes,
  SHA-256 `511d0b97390cc47aa0e1367189210a41f32088d9c869e7bb01f43698bdf7e5e8`,
  OpenAPI 3.0.1, and 617 operations. The exact method/path set is preserved,
  but `source-import jira` stops at `POST /rest/api/3/issue/bulkfetch` because
  `sourceReferenceResolver.referenceTargetWithCount` rejects the recursive
  `#/components/schemas/LinkGroup` at `cmd/connectorgen/sourceimport.go:5107-5108`.
  Jira is pending the same cyclic-schema shared foundation as Stripe (#4327);
  its source and ledger work is held in a named stash.

## Notion source hold

- Live Notion OpenAPI refresh measured 1304814 bytes, SHA-256
  `dee5763763b0b9fbad2aa8d5adb173ca350ec26dda557e658c5dbe9d2ea2f258`,
  OpenAPI 3.1.0, and 61 operations (the current source adds 12 to the old
  49-operation lock). `go run ./cmd/connectorgen source-import notion` stops
  at `GET /v1/async_tasks/{task_id}` because the recursive
  `#/components/schemas/publicApiAsyncTaskStatusResultJsonValue` reaches
  `sourceReferenceResolver.referenceTargetWithCount` at
  `cmd/connectorgen/sourceimport.go:5107-5108`. Notion is a further #4327
  consumer; its source refresh is held without a local schema flattening.

## Bitbucket source hold

- Live Bitbucket OpenAPI refresh measured 1359673 bytes, SHA-256
  `3dbfe6a80143511a287e58c21a193d3551ab5d41e8b60e65c1ae121b7000dec3`,
  OpenAPI 3.0.0, and 297 operations. This removes 34 stale rows from the
  former 331-operation lock, but `go run ./cmd/connectorgen source-import
  bitbucket` stops at `GET /repositories` on recursive
  `#/components/schemas/base_commit`, rejected by
  `sourceReferenceResolver.referenceTargetWithCount` at
  `cmd/connectorgen/sourceimport.go:5107-5108`. Bitbucket is another #4327
  consumer; its source refresh is held without a local schema flattening.

## GitLab source hold

- Live GitLab OpenAPI refresh measured 3576860 bytes, SHA-256
  `6b6ad591ff1b54ab429d0502812a2b2955501f1f6bebdae1888ba0bea086cf82`,
  OpenAPI 3.0.0, and 1752 operations after the document-level `/api/v4` base
  normalization. This removes three stale product-analytics rows from the
  former 1755-row lock, but `go run ./cmd/connectorgen source-import gitlab`
  stops at `POST /api/v4/glql`: its 200 response has a `$ref` with a
  `description` sibling. The refusal is
  `sourceReferenceResolver.referenceTargetWithCount` at
  `cmd/connectorgen/sourceimport.go:5088-5091`, so GitLab is another #4326
  consumer and its source refresh is held.

## Docker Hub source hold

- Docker Hub's live source exactly matches its lock: 148322 bytes, SHA-256
  `99d9d53c2d93656a3c66d604885abd153dc5df285abc0ecb13802a3bc53d0756`,
  OpenAPI 3.0.3, and 54 operations. Despite the exact source set,
  `go run ./cmd/connectorgen source-import dockerhub` stops at
  `POST /v2/auth/token`: response 401 has a `$ref` with a `description`
  sibling. The refusal is
  `sourceReferenceResolver.referenceTargetWithCount` at
  `cmd/connectorgen/sourceimport.go:5088-5091`, so Docker Hub is another
  #4326 consumer and its declaration repair is held.

## Post-#4327 revalidation (2026-08-23)

- The branch merged `origin/main` at `02a2201ed`, including #4327
  (`e338cd301`). The former recursive-schema and OpenAPI 3.0 descriptive
  `$ref`-sibling refusals no longer occur.
- `source-import` passed only CircleCI and Sentry (2/10). `connectorgen
  validate` passed CircleCI only (1/10); CircleCI remains security-blocked by
  #4334 because the source-derived `signing-secret` flags remain ordinary CLI
  flags.
- Asana and Jira now reach source projection, which reports respectively 25
  and 16 mutations without complete executable actions at
  `cmd/connectorgen/sourceprojection.go:211`.
- Bitbucket and Notion now fail provider schema-depth enforcement at
  `cmd/connectorgen/sourceimport.go:4271` (Bitbucket's pull-request comment
  response; Notion's meeting-notes response). Stripe fails reference-depth
  enforcement at `cmd/connectorgen/sourceimport.go:5170` for `GET /v1/account`.
- Docker Hub advances to an invalid provider reference
  `#/components/responses/team_repo`, refused by `sourcePointer` at
  `cmd/connectorgen/sourceimport.go:5496`; its existing SCIM operations also
  contain unsupported `example` schema keywords, rejected by the engine at
  `internal/connectors/engine/schema.go:168`.
- GitLab advances to a malformed provider path contract: `epic_issue_id` has
  no required path parameter, refused at
  `cmd/connectorgen/sourceimport.go:6048`.
- Vercel's mutable source refresh measured 10463249 bytes and SHA-256
  `74cb7ff3dc0b89cc344b13ac9c6d5f1d9b7d7a9356cfd6b5a779da51fd43da28`
  (400 operations), then rejects OpenAPI 3.0 `patternProperties` at
  `cmd/connectorgen/sourceimport.go:4314`.
- Sentry still has 34 mutation coverage findings at
  `cmd/connectorgen/sourceprojection.go:1943-1948`, the already-open #4329
  read-only source coverage gap.
- None of these post-#4327 results requests a rendered-reference contract: all
  failures occur after retrieval of the pinned OpenAPI source in the shared
  importer, source projection, or engine schema dialect.

## Sentry source evidence

- Red: the baseline `pm sentry operations list` exited 2 with
  `error: unknown command "sentry"`.
- The provider document at the lock URL measured 3868570 bytes and SHA-256
  `b71216654e44cc18f5e262fbb5075df67f1504a123d4bcb51cc8e8cc74ebd435`.
  It is OpenAPI 3.0.3 with 223 HTTP operations: 120 GET reads and 103
  mutations (35 DELETE, 2 PATCH, 34 POST, and 32 PUT).
- The old lock envelope was normalized to the v3 lock grammar, then
  `go run ./cmd/connectorgen source-import sentry --check` passed with all 223
  source operations verified. The pending production change is a generated
  read-only terminal surface for the four existing Sentry streams.
- The materializer refused the legacy `projects` stream because its old
  `GET /api/0/projects/` path is no longer in the provider artifact. The
  current source-bound equivalent is `GET /api/0/organizations/{organization_id_or_slug}/projects/`;
  the connector-local StreamHook must require `organization` and use that path
  before the stream can be materialized. This is connector-specific migration
  work, not a missing shared foundation.
