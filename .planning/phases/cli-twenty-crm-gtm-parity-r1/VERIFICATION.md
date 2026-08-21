# Verification checklist — Twenty CRM

- [x] GSD command resolution and inline fallback recorded.
- [ ] Recovery assessment precedes bundle import.
- [x] Targeted Twenty test red/green evidence recorded in `TDD-LEDGER.md`.
- [x] Focused structural generator and conformance checks: `go run ./cmd/connectorgen validate internal/connectors/defs/twenty`, `go run ./cmd/connectorgen surface-sync --check`, and `go test -timeout 20m ./internal/connectors/conformance -run 'TestConformance/twenty'`.
- [x] Structural generator, runtime preflight, conformance, boundary, docs/golden,
  lint, build, and verification results recorded with exact commands.
- [x] Historical built-binary live read, pagination, create, update, round-trip,
  delete, and post-delete absence evidence is retained without secret material.
- [x] PR #4298 was retargeted through the GitHub API to
  `fm/cli-reverse-etl-destination-r1`.
- [x] Typed-action eligibility red/green evidence is recorded in
  `TDD-LEDGER.md` and covers every action in `writes.json`.
- [x] #4304 refresh: fetched and normally merged
  `d814875a902be684cb2a38b94f7a8077f66b70b1` without rewriting history;
  `git merge-base --is-ancestor` passed and GitHub API read-back confirms #4298
  now targets that exact branch/SHA.
- [ ] Provider-neutral per-action source binding: #4304's persisted action
  selection is insufficient because its one source binding per executor/stream
  cannot preserve Twenty's 55 distinct candidate mappings. Do not add a
  connector workaround; await a foundation decision before attempting the 56
  record-shaped bindings or installed App/CLI proof.
- [x] Foundation-lane handoff: `FOUNDATION-HANDOFF.{md,json}` source-traces
  the 55/28/28 membership, the current descriptor/App call path, and the
  committed red/green behavioral witness. Its source hashes match
  `FOUNDATION-GAPS.json`; no live credential or runtime action was used.
- [ ] Fresh dedicated-instance certification after that merge: built-binary
  authenticated list/get with pagination, lane-owned create/read-back/update/
  delete/verified cleanup, ETL source, reverse-ETL plan/apply/acknowledgement
  plus independent API read-back, and binary transfer only if documented.

## Historical live proof — superseded for final acceptance

The disposable published-Compose stack used the pinned image
`twentycrm/twenty@sha256:cb80b05bc2619a88a3a83293f45f2be495a55ac77a90946fa1f7d85f0b7fde24`.
The stack's `/healthz` returned 200. GraphQL `__schema` introspection was
disabled, so the current image's supported `workspace:seed:dev --light` command
created the disposable workspace instead; no UI and no captain-owned instance
were used.

Firstmate's former direct-stream decision avoided a Keychain truncation
observation. This historical proof is not reused for the current acceptance
gate. The refreshed proof must source the registered `pm-twenty`
Keychain/secret-store reference only at stdin injection, never from argv,
output, a file, a prompt, commit, or evidence. The immediate built-binary
`pm twenty companies list --credential twenty-live --limit 3 --json` returned
three real seeded records, proving the whole stored credential authenticated.

The actual built binary then passed these bounded live probes (all identifier
and token values withheld):

- `companies list --limit 1000` returned 600 records, proving cursor pagination
  crossed a provider page boundary. A separate `people list --limit 1000`
  returned 1,000 records.
- The 28 declared ETL object commands each completed a bounded `list --limit 1`;
  where seeded data existed, a following operation-backed `get --id` completed.
- A company created through the plan → approval-stdin → run flow was readable
  with its exact input name, then updated through the same flow and readable
  with the changed name.
- The same record was deleted through plan → preview → approval-stdin →
  `--confirm destructive`; the run reported one staged, one succeeded, zero
  failed. It was absent from the following list and its get returned provider
  404.

This is real-data runtime proof only; it does not widen the repository's
separate certification allowlist or claim `CERTIFIED` status for Twenty.

## Current dedicated-runtime discovery — blocked

Captain authorized a read-only local-runtime and protected-tree discovery on
2026-08-20. Docker reported two stopped Polymetrics PostgreSQL integration
containers, zero Compose projects, and no Twenty container. Its local image
store contains `twentycrm/twenty:latest` and `twentycrm/twenty-app-dev:latest`,
but neither image is deployed. Podman is installed but unavailable; both local
machines are stopped. The preserved `.twenty-live.SRsCLr/` tree contains one
non-executable JSON file and no compose, shell, environment, secret, credential,
stdin-handoff, or deployment marker. Its contents and all environment/inspect
values were intentionally not printed.

Therefore no currently running dedicated Twenty identity, isolation proof, or
non-exporting credential handoff is available. Do not create, start, inspect
with environment output, or repurpose any other local container until a
dedicated disposable runtime and an execution-time stdin/FD/approved-secret-store
credential reference are available. The refreshed #4304 head is also still
`c6f03c937`; it remains an ancestor, but it does not provide the required
persisted multi-action selection.

## Official Twenty Compose review — blocked before execution

On 2026-08-20, the lane reviewed the exact current upstream Twenty source at
commit `e14694f4ff9ca51b791ba6b09639fed0944c5ad7` without emitting environment
assignments or secret material. The reviewed source blobs were:

- `packages/twenty-docker/docker-compose.yml`:
  `0b0d172d52e5f41d6f4bee704ee8b7cb91fa5434`
- `packages/twenty-docs/developers/self-host/capabilities/docker-compose.mdx`:
  `650a2c9b632417667969715ac0bd3e92d25fd390`
- `packages/twenty-docs/developers/self-host/capabilities/setup.mdx`:
  `14c4c5291c742eecc09006f06be750749e76bdf4`
- `packages/twenty-docker/twenty/entrypoint.sh`:
  `fd85657f5d0d010a5cea63b007023cdc3be513f9`

The official production recipe declares four services (`server`, `worker`,
`db`, and `redis`), two named data volumes, one published-port mapping, and
health checks for the server and backing data services. It declares no
external volume/network, host networking, privilege escalation, external HTTP
endpoint, Twenty Cloud/production domain, or `env_file`. Its default published
port is not loopback-only, so a future lane-owned project must override it to a
unique loopback port and retain Compose's project-private network and new named
volumes.

The recipe declares image tags rather than immutable digests. The only
recipe-declared Twenty image presently cached locally resolves to
`sha256:cb80b05bc2619a88a3a83293f45f2be495a55ac77a90946fa1f7d85f0b7fde24`
with local image ID
`sha256:e4a44722897aa851ac139ad1693141e382c0480693f0153a4f50cbdc28fd12cc`.
No additional image was pulled and no container, volume, network, or workspace
was created during this review.

The production entrypoint performs database setup/migrations but has no
workspace-seed, admin-account, or API-key creation behavior. The reviewed
official setup documentation describes the first single-workspace creation as
an interactive application flow and does not document a noninteractive
workspace/bootstrap or API-key issuance handoff. A development-image seed
script exists upstream, but that image and script are not declared by the
reviewed production Compose recipe, so this lane will not substitute it.

Therefore the lane must not start the recipe yet: it lacks an authoritative,
supported noninteractive disposable-workspace and API-key creation path that
can feed the built CLI via stdin/FD without recording the credential. Editing
the database, bypassing authentication, using a UI-extracted value, or
repurposing the historical runtime would violate the live-certification safety
contract.

### Captain decision correction — interactive bootstrap authorized

The captain subsequently authorized the official interactive first-workspace
and API-key flow within this new disposable local instance. This resolves only
the noninteractive-bootstrap blocker: it does not authorize database edits,
authentication bypasses, production access, existing Docker-resource reuse, or
token display in a browser snapshot/transcript. The next lane step is a
project-private Compose deployment with a unique project name, new named
volumes/default network, a unique loopback-only override for the one published
port, and redacted per-image digest evidence before startup. The generated API
key must move directly from the approved local application flow to the built
CLI's stdin credential input without printing or persisting the plaintext.

### Isolated Compose inputs resolved before startup

The lane reserved Compose project `pm-twenty-cert-r1` and verified its project
containers, networks, and volumes do not exist. Loopback port `39177` was
unbound and is the only host binding in the lane override
`.twenty-cert-r1/compose.loopback.yml`; the production recipe's original public
mapping is replaced rather than supplemented.

Before any container startup, the recipe-declared images resolved locally to:

- `twentycrm/twenty@sha256:cb80b05bc2619a88a3a83293f45f2be495a55ac77a90946fa1f7d85f0b7fde24`
- `postgres@sha256:95206741a5b214807675e14165369d05b93a9cf692223b616d07cca227e74b0b`
- `redis@sha256:344e3945a0b431c8ff1eecd58c5573538126bd756f02fc7e218ddf1fc2546366`

The two backing images were pulled only because the reviewed recipe declares
them. Docker has not yet created a container, network, volume, or workspace
for this project. Server-internal encryption material and local application URL
will be generated only in the startup process environment and will not be
written to the override, logs, evidence, or source control.

### Initial local-start red checkpoint — no workspace data created

The first isolated startup created only the lane project's network, volumes,
and four containers. PostgreSQL and Redis became healthy, but the Twenty server
restarted during configuration validation before serving the health endpoint;
the worker remained unstarted. Redacted classification identified an unset
storage configuration. A clean second lane-owned recreation reached the same
validation failure because the non-echoing enum lookup used to populate the
local-storage setting was a no-op; a sanitized rendered Compose check confirmed
that the field remained empty. The exact upstream configuration source requires
the normalized local-storage option for this recipe, with its default local
path, and validates the setting before application startup.

This is the second occurrence of the same startup obstacle. Per the lane
contract, do not make a third attempt in this turn. No workspace, account,
API key, or business record was created. The verified lane-owned failed project
must be torn down before pausing; a resumed worker can use the now-established
source-backed local-storage setting directly in the process environment.

### Approved source-backed restart

Firstmate approved one resume after the two-attempt checkpoint. The restart
must parse the normalized local-storage enum from the pinned upstream type
definition at execution time, use it only in the Compose process environment,
and confirm the server and worker render the source-derived setting through
boolean-only configuration checks. The enum value, all environment assignments,
and fresh generated encryption material are intentionally absent from evidence.

### Source-backed local restart and interactive bootstrap — green

The approved restart passed. Sanitized resolved-Compose checks confirmed that
both application services received the source-derived local-storage setting,
that fresh encryption material was present without rendering it, and that the
topology remained four services, two lane-owned volumes, one project-default
private network, and one loopback-only published port. The lane then created
only its own disposable workspace/admin and disposable API key through Twenty's
official interactive UI. The four containers are running under the unique
`pm-twenty-cert-r1` Compose project; database, cache, and server are healthy,
the worker is running, and the loopback health endpoint returned HTTP 200.

The one-time API-key display was copied by the official UI and piped directly
from the local clipboard to `pm credentials add --value-stdin api_key`; it was
then cleared from the clipboard. The newly built CLI stored only the secret
field name in its AES-GCM credential vault. A built-CLI credential test returned
`status: ok`. No key, browser session secret, encryption value, connection
string, or provider token was rendered in the terminal, planning evidence, or
browser snapshot.

This is environment and authentication preparation, not final connector
certification. The mandated built-binary plus persisted-App create/read/update/
delete, ETL, reverse-ETL, and eligible binary proofs remain pending the final
published #4304/#4305 application-dispatch foundation integration.

### Dependency-independent live read smoke — blocked

The approved bounded built-binary read did not authenticate. The actual
`twenty companies list` command, using the encrypted `api_key` field and the
loopback endpoint configuration, failed with HTTP 403. Its redacted diagnostic
classification is `authentication_or_authorization` and its non-secret error
fingerprint is `dd3dcab3ba4132eb599ce2857f643f96b22f85e26c52436a603815cfdb9dc901`.

An independent browser request, made through the official authenticated UI
session to the same lane-local `/rest/companies` route with a bounded window,
returned HTTP 200 and two records. The response schema fingerprint was
`ef6ed57dad7b61f2d7426674d8e2d1ed9a83c5e6eafb38916cca6ee3e23d4e45`.
No payload values, identifier, credential, header, or browser session material
was emitted. The unique Compose project, private volumes/network, local
workspace, and encrypted credential remain intact for authorization-path repair.

### Read-only authorization audit — shared CLI dependency

The authorized read-only audit establishes that this is not a Twenty connector
header/scheme defect. A read-only App open selected exactly one Twenty
credential. Its encrypted secret bundle has one declared field and declares
`api_key`, but the decrypted value has length zero (its SHA-256 is the standard
empty-input digest). The real declarative reader therefore reached a local
in-process receiver with `Authorization: Bearer` and no credential value. The
receiver emitted zero records and returned successfully; no live Twenty request
or state mutation occurred during this probe.

This matches the shared CLI source: `runCredentials` reads
`--value-stdin` with `io.ReadAll`, trims only line endings, and persists the
result without rejecting an empty value. The vault round trip faithfully
preserves that empty input. Existing engine bearer tests pass, and the Twenty
bundle's declared `bearer` mode correctly maps the expected `api_key` field to
that header. The original HTTP 403 is therefore expected and must not be
retried while the stored secret is empty.

Pinned Twenty source confirms that its REST API extracts a bearer token, rejects
user-session tokens as bearer credentials, and validates API keys separately
through the workspace API-key cache, expiry/revocation check, and assigned API
key role. Its current API-key tokens use the database-backed asymmetric signing
key; the legacy `APP_SECRET` path is only a verification fallback. Boolean-only
inspection of the two application containers found current encryption and
fallback-encryption material present, no legacy `APP_SECRET`, and no evidence
of a self-host authorization configuration mismatch. The official UI's API-key
table showed one visible disposable key with Admin role, Never expiry, and no
revoked-state label. No identifier, token, response field, environment value,
or session material was rendered.

The minimal provider-neutral dependency is: reject an empty
`credentials add --value-stdin <field>` value after line-ending normalization
and before `App.AddCredential`/vault persistence, with happy (non-empty), bad
(empty), and edge (newline-only) tests. This is shared CLI credential-ingestion
behavior, not connector-owned code. Per the lane boundary, no generic fix or
Twenty-specific workaround has been added here. A separate approved repair must
land before a newly acquired disposable key can be injected through the
corrected path and the built-CLI HTTP-200 gate can be retried.

Commands and redacted results:

```text
go run ./.twenty-cert-r1/credential-wire-audit
# PASS audit execution: one Twenty credential, expected secret field present,
# zero-length decrypted value, bearer scheme with no token, zero-record local receiver.
go test -count=1 -timeout 20m ./internal/connectors/engine -run 'Test.*Bearer|TestSelectAuth'
# PASS (1.088s)
chrome-devtools-axi eval <boolean-only API-key-table projection>
# PASS: one visible key; Admin + Never labels; no revoked label.
docker inspect <lane server, worker> | jq <boolean-only auth-config projection>
# PASS: both services have encryption and fallback-encryption material; legacy APP_SECRET absent.
```

Redacted command/result evidence for this checkpoint:

```text
docker compose --project-name pm-twenty-cert-r1 ... config --images
# PASS: exactly the three reviewed recipe image references
docker image inspect --format '<id> <repo-digest>' <each declared image>
# PASS: all three immutable digests recorded above before startup
docker compose --project-name pm-twenty-cert-r1 ... config --format json | jq <structural projection>
# PASS: one loopback-only port, four services, project-default network, two named volumes
docker compose --project-name pm-twenty-cert-r1 ... create
docker start pm-twenty-cert-r1-db-1 pm-twenty-cert-r1-redis-1 pm-twenty-cert-r1-server-1
# RED: backing services healthy; Twenty server configuration validation fails before /healthz
docker rm -f <four verified pm-twenty-cert-r1 containers>
docker network rm pm-twenty-cert-r1_default
docker volume rm pm-twenty-cert-r1_db-data pm-twenty-cert-r1_server-local-data
# PASS: verified lane-only teardown; no project resources remain
```

## Local verification results

All of these completed successfully unless explicitly marked as a documented
red-to-green checkpoint:

```text
go test -timeout 20m ./internal/connectors/defs/twenty
go test -timeout 20m ./internal/connectors/conformance -run 'TestConformance/twenty'
go test -timeout 20m ./internal/connectors/commandrunner -run 'TestEveryImplementedCommandPassesRuntimePreflight'
go test -timeout 20m ./internal/connectors
go test -timeout 20m ./cmd/iconregistrygen
go vet ./...
go build -o .twenty-live.SRsCLr/pm ./cmd/pm
go run ./cmd/connectorgen validate internal/connectors/defs/twenty
go run ./cmd/connectorgen surface-sync --check
go run ./cmd/connectorgen certification-sweep --connector twenty
go run ./cmd/connectorgen certification-sweep --connector twenty --check
make fmt
make tidy-check
make docs-check-no-build
make smoke-no-build
make lint
make agent-contract-check
make connectorgen-validate
make connectorgen-surface-sync
make connector-runtime-preflight
make connector-canon-check
make connector-boundary
make github-parity-artifacts-check
make connectorgen-certification-matrix
make connectorgen-certification-candidates
make connectorgen-certification-sweep
make release-workflow-check
scripts/verify-gsd-workflow
git diff --check
```

The expected initial red checkpoint was
`go test -timeout 20m ./internal/cli`: catalog count 556 and root-help golden
fixtures did not yet include Twenty. After regenerating exactly the nine
root-help transcripts from the binary and changing the catalog expectation to
557, the following green results were recorded:

```text
go test -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$' -count=1
# PASS (17.729s)
go test -timeout 20m ./internal/cli
# PASS (1184.608s)
```

The subsequent GitHub `Verify` run `32274541836` supplied one further red
projection check: `internal/connectors/bundleregistry` loaded 553 definitions
but its historical assertion still expected 552. The fixed focused check is:

```text
go test -timeout 20m ./internal/connectors/bundleregistry
# PASS
```

After rebasing the committed branch onto the latest `origin/main`, the final
pre-push run passed again:

```text
go test -timeout 20m ./internal/cli
# PASS (596.926s)
go test -timeout 20m ./internal/connectors/defs/twenty
go run ./cmd/connectorgen validate internal/connectors/defs/twenty
go run ./cmd/connectorgen surface-sync --check
go run ./cmd/connectorgen certification-sweep --connector twenty --check
make docs-check-no-build
make smoke-no-build
make lint
make connector-boundary
go build ./cmd/pm
git diff --check
# PASS; boundary: 553 connectors loaded, zero findings
```

`make verify` was not invoked as a monolithic command because this repository's
agent guidance explicitly forbids it under a per-command timeout: it embeds
the whole suite and a timeout is indistinguishable from a hang. Its named
non-suite entry points above were run individually, and the full
`internal/cli` package was run with the repository-required 20-minute timeout.
The remaining repository-wide suite is CI coverage.

The post-build real-instance authentication check was run from the disposable
proof project:

```text
../pm twenty companies list --credential twenty-live --limit 3 --json | jq ...
# {"status":"passed","record_count":3}
```

The credential value, record values, record identifiers, approval material,
and endpoint configuration are intentionally absent from this evidence.

## Pull request

Created non-draft PR [#4298](https://github.com/polymetrics-ai/cli/pull/4298)
with `Refs #277`. The reconciliation API read-back returned base
`fm/cli-reverse-etl-destination-r1` and head
`fm/cli-twenty-crm-gtm-parity-r1`. A final API read-back is still required
after the fresh #4304 merge and before the next push; the generic
destination's persisted App/CLI dispatch is not yet claimed as deployable.
