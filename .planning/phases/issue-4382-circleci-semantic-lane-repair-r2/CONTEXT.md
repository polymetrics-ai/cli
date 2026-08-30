# CircleCI semantic lane repair R2 — context

## Task Delivery Header

- Issue: Refs #4382 — CircleCI — source-to-seven-lane matrix
- Frozen repair target: `a74c13b6e42c8d6a0e2dca3033571ad5941c8fc2`
- Current upstream merge-base: `origin/main` at `813f457a925f7ee3fe3bea101a43e445992c8552`
- Delivery: A scoped repair branch is committed, locally verified, pushed, and has a #4382 completion-proof/re-review comment; no pull request or merge is opened by this task.
- Working branch: `fix/4382-circleci-semantic-repair-r2`
- Task: Correct CircleCI source-to-seven-lane classification from retained request/response semantics. Preserve all 111 source rows and seven lane cells per row, without changing source locks, runtime code, importer, certification, shared controls, transport, Atlas, or other connector artifacts.
- Verification: Focused CircleCI matrix tests (including deliberate red probes), package tests and race tests, `go vet`, strict JSON/count reconciliation, `agentcontractgen check`, `git diff --check`, and explicitly recorded residual admission results.

## Scope and fixed decisions

The frozen target `a74c13b6e42c8d6a0e2dca3033571ad5941c8fc2` contains a CircleCI source matrix whose original classification relied too heavily on HTTP method and inline response inspection. This repair changes only:

- `internal/connectors/defs/circleci/sources/circleci-source-lane-matrix.json`;
- `internal/connectors/defs/circleci/source_lane_matrix_test.go`;
- this repair phase's planning/review evidence.

Every provider mutation remains independently visible as both `direct_write` and `reverse_etl`; this is source mapping only and does not claim runtime execution. Non-GET requests that are source-documented bounded reads must remain direct reads instead of being promoted as mutations. GET is not enough to establish ETL. ETL requires source-backed collection extraction plus continuation semantics; an array, limit, or request-only paging control alone is insufficient. Sync remains non-applicable unless a cited event/webhook contract supports it; pagination never implies sync.

Five retained source operations carry a documented `page-token` request parameter and response `$ref` whose referenced schema provides a string `next_page_token`: `listContexts`, `listEnvironmentVariablesFromContext`, `getOrganizationGroups`, `listEnvironments`, and `listComponents`. The repair resolves this retained `$ref` evidence before classifying their ETL cells.

The retained source also documents `createWebhook` and `updateWebhook` as outbound-webhook registrations. Their typed request bodies carry an event array, URL, and signing secret. They are not represented as executable sync transport in this Track A repair: their `sync_transport` cells are source-backed `missing_foundation` records. All other CircleCI sync cells, including paginated lists and webhook inspection/deletion endpoints, remain `not_applicable`.

## Evidence and safety

- Source authority remains `internal/connectors/defs/circleci/sources/circleci-operation-source-lock.json` unchanged.
- No provider I/O, credential, secret, destructive operation, or generic request path is used.
- The Foundation Atlas was consulted read-only. It records generic `transport.sync-contract.v1`, but this repair neither changes it nor claims executable CircleCI ingress. The retained signing-secret/event contract identifies a real provider-specific inbound-webhook receiver/conformance gap; it remains a named mapping gap for later captain-approved runtime work.
- CodeGraph was checked at the target repository root and is absent; repository-local `rg`/targeted source inspection is used instead.

## GSD and skill record

CodeGraph is absent at this target repository root, so targeted `rg` inspection was used. Skills used for this scoped repair: `connector-lane-build-order` and `go-engineering`. The repair follows red-green-refactor with explicitly recorded negative probes in `TDD-LEDGER.md`.
