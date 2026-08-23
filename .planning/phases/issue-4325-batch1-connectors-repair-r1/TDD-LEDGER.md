# Issue 4325 — TDD Ledger

| Slice | Red evidence to record before production edit | Green assertion | Refactor/quality gate | Status |
| --- | --- | --- | --- | --- |
| Baseline source drift | `go run ./cmd/connectorgen source-import asana --check` and the CircleCI equivalent each exited 1 because current main lacks `defs/<connector>/sources/` | Re-pinned import and descriptor check pass with provider-derived exact set | `connectorgen validate` and `surface-sync --check` | CircleCI source green / security hold #4328; remaining locks pending |
| CircleCI/Sentry/Vercel surfaces | A built baseline binary returned exit 2 / `unknown command` for all three `operations list` probes | The same command returns `missing --credential` and preflight accepts it | Generated command/help checks | CircleCI pending #4328; Sentry/Vercel pending |
| Jira reachability | Enabled row lacks typed operation or disabled write row contradicts implemented command | Every enabled row has the exact typed operation and direct writes report their actual state | Real preflight sweep | pending |
| Docker Hub/Notion truth | Contradictory status/citation or wrong covered-by action is detectable | Metadata points to existing source/action and current refusal line | Bundle validation | pending |
| Stripe semantics | Four JSON response routes classify as binary | The four routes classify as JSON direct reads and command output policy matches | Semantic regression check | pending |
| Evidence reasons | Report scan finds forbidden scope reason or an uncited foundation gap | Scan finds none; all remaining citations resolve to the stated runtime refusal | Independent Gate B rerun | pending |
| Final gate | Independent report returns NO-GO | Independent report returns GO | Full `make verify` and review | pending |

## Jira red/green evidence

- Red: `go run ./cmd/connectorgen source-import jira --check` exited 1 because
  no connector-owned source directory exists. The pre-change surface has 590
  commands (584 implemented), 292 implemented direct reads, 286 implemented
  write commands, but only 25 typed operations (22 REST reads and three binary
  downloads). Its 617-row v1 ledger therefore cannot truthfully describe the
  legacy direct-read and direct-write reachability.
- Green assertion: a refreshed source lock and v2 artifact ledger must retain
  the full provider method/path set, bind every implemented direct read to an
  explicit typed operation, and classify direct-write actions according to the
  actual CLI command state. No enabled endpoint may remain operation-less.
- Blocked foundation: after live refresh (2456011 bytes; SHA-256
  `511d0b97390cc47aa0e1367189210a41f32088d9c869e7bb01f43698bdf7e5e8`;
  OpenAPI 3.0.1; 617 operations), `go run ./cmd/connectorgen source-import
  jira` stops on `POST /rest/api/3/issue/bulkfetch`: the recursive
  `#/components/schemas/LinkGroup` reaches
  `sourceReferenceResolver.referenceTargetWithCount` at
  `cmd/connectorgen/sourceimport.go:5107-5108`. Jira is pending the shared
  cyclic-schema importer repair #4327; no local flattening or fabricated gap
  is valid.

Every red/green command and its result is appended when it executes. No test is
weakened or skipped to advance a row.

## CircleCI red/green evidence

- Red: the baseline `pm circleci operations list` exited 2 with
  `error: unknown command "circleci"`.
- Green source: the 2026-08-23 provider retrieval measured 621321 bytes,
  SHA-256 `61c6ce11e8de509948aa3d53dcd0169913f52de20920b130b6a85dea41d66d07`,
  and 111 exact method/path identities. `go run ./cmd/connectorgen source-import
  circleci --check` passed.
- Green ledger: the temporary batch materializer accepted the same live artifact
  with 111 artifact operations and no dropped candidate. The retained v2
  ledger has 40 executable bindings and 71 source-cited blocked/disallowed
  rows. `go run ./cmd/connectorgen validate internal/connectors/defs/circleci`
  and `go run ./cmd/connectorgen surface-sync --check` both passed.
- Green runtime: from a fresh initialized project with no credential,
  `pm circleci contexts list` and `pm circleci delete context apply
  --context-id fixture-context` each returned exactly `error: missing
  --credential` (exit 1), not `unknown command`. No provider request was made.

## Foundation holds

- Asana is pending #4326, not partial. The real source-import failure is
  `sourceReferenceResolver.referenceTargetWithCount` at
  `cmd/connectorgen/sourceimport.go:5088-5091`; its OpenAPI 3.0 response `$ref`
  has a `description` sibling. The shared importer must retain the source-bound
  contract before this slice can continue.
- Stripe is pending #4323’s source-import cyclic-schema repair and stays last.
- CircleCI webhook writes are pending #4328. The locked source declares
  `signing-secret`, but `sourceProjectCommand` at
  `cmd/connectorgen/sourceprojection.go:1550-1592` generates an ordinary
  reverse-ETL CLI flag. `checkCLISurfaceEnvOnlyFlags` at
  `cmd/connectorgen/validate.go:896-928` permits secure env-only input only
  for operation-backed GraphQL contracts. The shared write-field secret-input
  capability or a source-bound gap is required; no connector-local shim or
  partial label is valid.
- Sentry is pending a shared read-only source-projection capability. With the
  generated four-command CLI surface present, `connectorgen validate
  internal/connectors/defs/sentry` reports 34 `source operation has no
  executable action` findings. The enforcement is
  `validateSourceExecutableCoverage` at
  `cmd/connectorgen/sourceprojection.go:1943-1948`: a source mutation with no
  action is valid only when the importer itself generated a blocking runtime
  gap. These are unsupported reverse-ETL operations in an intentionally
  read-only connector, not a source grammar failure, so an invented connector
  gap, action, or `partial` command would be false evidence. A shared,
  source-cited read-only refusal model is required.
- Vercel is pending the same #4329 foundation. The checked-in source lock has
  400 operations (163 reads and 237 mutations). Its 18 writes match only four
  source method/path identities, and its CLI surface is absent. Adding a
  surface before #4329 would produce the identical dishonest all-or-nothing
  mutation demand, so the connector remains untouched.

## Notion red/green evidence

- Red source: after importing the prior source envelope, live retrieval found
  61 OpenAPI operations, not the 49 locked rows (1304814 bytes, SHA-256
  `dee5763763b0b9fbad2aa8d5adb173ca350ec26dda557e658c5dbe9d2ea2f258`).
- Blocked foundation: `go run ./cmd/connectorgen source-import notion` stops
  at `GET /v1/async_tasks/{task_id}` on the recursive
  `#/components/schemas/publicApiAsyncTaskStatusResultJsonValue`, refused in
  `sourceReferenceResolver.referenceTargetWithCount` at
  `cmd/connectorgen/sourceimport.go:5107-5108`. The shared #4327 cycle-gap
  importer repair is needed; no connector-local flattening or invented gap is
  honest.

## Bitbucket red/green evidence

- Red source: live retrieval found a 297-operation OpenAPI document, replacing
  the former 331-row lock (1359673 bytes, SHA-256
  `3dbfe6a80143511a287e58c21a193d3551ab5d41e8b60e65c1ae121b7000dec3`).
- Blocked foundation: `go run ./cmd/connectorgen source-import bitbucket`
  stops at `GET /repositories` on the recursive
  `#/components/schemas/base_commit`, refused at
  `cmd/connectorgen/sourceimport.go:5107-5108`. The shared #4327 cycle-gap
  importer repair is required before the stale source rows can be truthfully
  removed and the source projections regenerated.

## GitLab red/green evidence

- Red source: live retrieval found 1752 operations after the documented
  `/api/v4` base normalization, instead of the former 1755 lock rows
  (3576860 bytes, SHA-256
  `6b6ad591ff1b54ab429d0502812a2b2955501f1f6bebdae1888ba0bea086cf82`).
- Blocked foundation: `go run ./cmd/connectorgen source-import gitlab` stops
  at `POST /api/v4/glql`: response 200 has a `$ref` with a `description`
  sibling. The importer refusal is
  `sourceReferenceResolver.referenceTargetWithCount` at
  `cmd/connectorgen/sourceimport.go:5088-5091`. GitLab needs #4326 before
  stale source rows and its 981 foundation-gap citations can be regenerated.

## Docker Hub red/green evidence

- Green source retrieval: Docker Hub remains exact at 54 operations, 148322
  bytes, and SHA-256
  `99d9d53c2d93656a3c66d604885abd153dc5df285abc0ecb13802a3bc53d0756`.
- Blocked foundation: `go run ./cmd/connectorgen source-import dockerhub`
  stops at `POST /v2/auth/token`, response 401, because its `$ref` has a
  `description` sibling. The failure is
  `sourceReferenceResolver.referenceTargetWithCount` at
  `cmd/connectorgen/sourceimport.go:5088-5091`. Docker Hub needs #4326 before
  the source projection can truthfully repair stale citations and contradictory
  enabled/blocked ledger metadata.

## Post-#4327 red/green revalidation

- Green foundation observation: after merging `origin/main` at `02a2201ed`
  (including #4327 at `e338cd301`), none of the seven former cycle or
  descriptive `$ref` sibling errors recurred.
- Re-run: `source-import` passed CircleCI and Sentry only (2/10). The full
  `connectorgen validate` sweep passed CircleCI only (1/10).
- New red: Asana (25) and Jira (16) reach source projection but have incomplete
  executable action contracts (`cmd/connectorgen/sourceprojection.go:211`).
  Bitbucket and Notion reach `schema depth limit exceeded`
  (`cmd/connectorgen/sourceimport.go:4271`); Stripe reaches `reference depth
  limit exceeded` (`cmd/connectorgen/sourceimport.go:5170`). Docker Hub has an
  unresolved source reference (`cmd/connectorgen/sourceimport.go:5496`) and
  existing SCIM `example` keywords rejected by the engine dialect
  (`internal/connectors/engine/schema.go:168`). GitLab's provider path has a
  placeholder with no required parameter (`cmd/connectorgen/sourceimport.go:6048`).
  Vercel's live 400-operation source rejects OAS 3.0 `patternProperties`
  (`cmd/connectorgen/sourceimport.go:4314`). Sentry remains blocked by its 34
  read-only mutation coverage findings (`cmd/connectorgen/sourceprojection.go:1943-1948`).
- No failure is a rendered-reference contract request: every observed error is
  in the OpenAPI importer/projection/engine path after source retrieval.

## Sentry red/green evidence

- Red: the baseline `pm sentry operations list` exited 2 with
  `error: unknown command "sentry"`.
- Red source: `go run ./cmd/connectorgen source-import sentry --check` first
  exited 1 because the imported lock still carried the retired
  `rest.operation_counts` field.
- Green source: after normalizing the lock envelope, the live provider
  document measured 3868570 bytes and SHA-256
  `b71216654e44cc18f5e262fbb5075df67f1504a123d4bcb51cc8e8cc74ebd435`.
  It has 223 operations (120 GET, 35 DELETE, 2 PATCH, 34 POST, and 32 PUT),
  and `go run ./cmd/connectorgen source-import sentry --check` passed.
- Next green assertion: batch materialization must preserve all 223 cited
  endpoint rows while exposing only the four existing read streams. The absent
  `writes.json` is intentional: no reverse-ETL action exists or is claimed.
- Red hook contract: a new focused test must require `projects` to resolve to
  the provider's organization-scoped path and reject a missing organization;
  it must fail against the former legacy-wide `/api/0/projects/` implementation.
