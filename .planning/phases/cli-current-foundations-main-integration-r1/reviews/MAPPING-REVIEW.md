---
reviewed_sha: 9e5329f34e015e39160bb8e951452bbd071a698a
depth: deep
scope_count: 25
status: issues_found
finding_count: 12
---

# Foundation Rollup Mapping Review

## Review contract and result

This was a discovery-only adversarial review of the immutable foundation-rollup snapshot. The initial preflight resolved `HEAD` to `9e5329f34e015e39160bb8e951452bbd071a698a` and found a clean worktree. CodeGraph was attempted first as required, but this worktree has no `.codegraph/` index, so the review used read-only source inspection, exact locked artifacts, and executable checks. No source, canonical `REVIEW.md`, or `REVIEW-CONVERGENCE.md` was changed.

The reviewed mapping is not shippable. Eleven BLOCKER findings show broken source import/projection, currently unreachable provider inputs, a nonfunctional binary-upload declaration, severe GraphQL and response-output narrowing, an ineffective destination read-back, a reusable-authorization post-effect failure, credential disclosure, and lost failure receipts. One WARNING identifies a hard-coded connector allowlist in generated agent skills.

## Exact scope

All 25 requested files were read in full or in complete logical sections, and their cross-file call paths were followed read-only:

1. `cmd/connectorgen/main_test.go`
2. `cmd/connectorgen/validate.go`
3. `internal/app/app.go`
4. `internal/app/authorization.go`
5. `internal/app/declarative_typed_destination_approval.go`
6. `internal/app/durable_coordination.go`
7. `internal/app/etl_mode_dispatch.go`
8. `internal/app/foundations_integration_test.go`
9. `internal/app/issue_label_transport_approval.go`
10. `internal/app/issue_label_warehouse_transport.go`
11. `internal/app/local_warehouse.go`
12. `internal/app/rest_write_command_test.go`
13. `internal/app/transport_composition_test.go`
14. `internal/app/transport_dispatch.go`
15. `internal/app/transport_dispatch_test.go`
16. `internal/app/types.go`
17. `internal/cli/cli.go`
18. `internal/cli/cli_test.go`
19. `internal/cli/docs.go`
20. `internal/cli/etl_transport.go`
21. `internal/cli/etl_transport_test.go`
22. `internal/cli/golden_transcript_test.go`
23. `internal/cli/skills.go`
24. `internal/cli/structured_rest_body_help_test.go`
25. `internal/cli/testdata/golden_transcripts.json`

The required planning and evidence inputs were also read: `REVIEW-CONVERGENCE.md`, `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `input-manifest.json`, and `evidence-manifest.json`. The evidence manifest still names composite SHA `808896a28873c5f0479fa10e2f798da56f885b5e`; the reviewed source at `9e5329f...` is the same source composite plus later review/planning material, but all conclusions below are tied to the requested SHA.

## Provider-to-installed-path call and mapping matrix

The checked-in GitHub source lock declares 1,525 provider operations: 1,220 REST, 31 GraphQL queries, and 274 GraphQL mutations (`internal/connectors/defs/github/sources/github-operation-source-lock.json:121883-121887`). `operations.json` contains 770 operation declarations, while `cli_surface.json` contains 1,571 commands, 1,546 marked implemented. Exact method/path comparison found all 1,220 locked REST method/path pairs in `api_surface.json` (with five local additions), but endpoint presence is not field-complete execution.

| Surface / stage | Source-locked or imported declaration | Validator / generated command and help | App planner / approval | Runner and terminal result | Verdict |
|---|---|---|---|---|---|
| Source import and projection | The locked REST artifact is SHA- and byte-pinned; the same lock contains the full GraphQL root/type system. `sourceImportLock` models only `rest` (`cmd/connectorgen/sourceimport.go:109-113`). | `source-import` emits an arbitrary `--out` descriptor file (`cmd/connectorgen/sourceimport.go:6230-6283`). `validateDir` loads bundle directories only (`cmd/connectorgen/validate.go:225-301`); surface sync reads existing CLI/operation/API documents and cannot create an omitted operation. | None. The imported descriptor has no App consumer. | The real checked-in lock fails import with `source grammar position byte limit exceeded`; even a successful descriptor has no installed consumer. | **FAIL — MAP-BL-01, MAP-BL-02.** |
| Ordinary ETL / local warehouse | A saved connection and exact stream select the source declaration. | `pm etl run` accepts connection, stream, bounds, and declaration-owned transport approval only. | `RunETL` resolves source/destination and `dispatchETLMode` chooses a closed transport or local warehouse (`internal/app/etl_mode_dispatch.go:33-114`). | Local warehouse writes and syncs its WAL/table before producing pending stream state (`internal/app/local_warehouse.go:255-310`). | **PASS for the configured stream path.** It does not compensate for source-to-bundle gaps. |
| Declarative reverse ETL / generated write commands | REST write actions and operation declarations are separately generated. The locked artifact exposes query and JSON-body fields that many actions omit. | Reverse-ETL validation checks only fields already named by `record_schema`/`path_fields` (`cmd/connectorgen/validate.go:1664-1753`), not the provider request. Generated commands therefore advertise implemented operations with required provider inputs absent. | Connector-command writes use the plan/preview/approval lifecycle; generic typed destinations bind the saved `destination_action`. | Declarative write sends record fields minus path fields when `body_fields` is absent. Sixty-five mapped REST endpoints omit at least one required body field; two more omit required query fields. | **FAIL — MAP-BL-03.** |
| Generic typed destination ETL | `sync_transport.json` binds the exact action, source executor, record mapping, mode, and conformance. | `pm etl transport declarative-typed-destination` accepts no connector/action/URL/method/body/mapping override (`internal/cli/etl_transport.go:129-160`). | The App seals the action definition and creates a durable authorization (`internal/app/declarative_typed_destination_approval.go:142-216`). | The write executes, but read-back checks only local values, and a reused authorization errors after a later successful write/checkpoint. | **FAIL — MAP-BL-06, MAP-BL-07.** |
| Specialized issue-label transport | The connector is selected by an exact, unique transport capability, not by connector name (`internal/app/issue_label_warehouse_transport.go:542-579`). | The leaf help and goldens are current. | Planning/approval binds source issue, target issue, label, mode, action, and definition digest. | Unlike the generic destination, its read-back performs a provider read. The one-page declared limit is explicit. | **PASS for name-independent selection and provider read-back.** |
| Direct REST read | An implemented command points to one `rest_read` operation, exact API endpoint, typed path/query/header/body mappings, cap, and output policy. | Static direct-operation mappings match the exact pinned REST artifact for commands that actually reference an operation. CLI dispatches through `commandrunner`, not raw HTTP (`internal/cli/cli.go:1263-1351`). | No write approval is needed. | Success emits only decoded/shaped body and admitted headers; provider HTTP failures return a zero result. Operation identity/raw body/presence and undeclared ordinary headers are lost (`internal/cli/cli.go:1397-1421`). | **FAIL — MAP-BL-11.** |
| GraphQL direct read | All 305 locked root operations have generated operations, but 303 documents select only `__typename` for the root result; five queries also omit the provider's `before`/`last` arguments. | Remaining variables are exact and typed; validation does not compare the complete argument/result surface with the locked schema. | No write approval is needed. | Runtime further reduces GraphQL errors to messages and drops error paths, locations, extensions, and top-level extensions. | **FAIL — MAP-BL-04, MAP-BL-11.** |
| Direct REST/GraphQL write | Direct operation commands have exact current path/query/header/body mappings and use a sealed plan, preview digest, single-use approval, and typed confirmation. | `validate.go` blocks raw API and checks operation intent/method/path/output policy plus required query/body mappings (`cmd/connectorgen/validate.go:1141-1495`). | `runOperationDirectWritePlan` revalidates the preview and consumes approval immediately before dispatch (`internal/app/app.go:2895-2933`). | Received responses are retained, but configured credential bytes are emitted verbatim; a no-response failure loses all result identity; CLI emits no failed run/result. | **FAIL — MAP-BL-08, MAP-BL-09, MAP-BL-10.** |
| Binary download / text export | Typed `binary_download`/`text_export` operation; GET-only, bounded, destination-root-confined. | CLI exposes only declared operation flags and required `--dest-root`. | No write approval; local destination is explicitly selected. | Successful bytes land exactly in a confined file with size/SHA-256 metadata. Failure/provider metadata is narrowed to an error and admitted headers only. | **PASS for successful content integrity and confinement; FAIL for complete provider failure/metadata output — MAP-BL-11.** |
| Binary upload | Locked REST operation `repos/upload-release-asset` requires `uploads.github.com`, query `name`, optional `label`, and an `application/octet-stream` body. | Generated command is named `releases assets view-3`, exposes only `release-id`, and is marked implemented (`internal/connectors/defs/github/cli_surface.json:10448-10469`). | It goes through ordinary reverse-ETL approval, but the approved request is the wrong request shape. | The action sends JSON to the default API origin with no file or required name (`internal/connectors/defs/github/writes.json:5750-5770`). | **FAIL — MAP-BL-05.** |
| Help, manuals, goldens, skills | Help is generated from the command surface; transport manuals are static. | Transport and structured-body focused tests pass; structured help uses a synthetic in-memory connector rather than an installed GitHub leaf (`internal/cli/structured_rest_body_help_test.go:13-60`). Golden cases cover namespaces and transport leaves but not real generated write leaves (`internal/cli/golden_transcript_test.go:24-136`). | N/A. | Generated skills arbitrarily include only five connector names. | **PASS for current transport/golden text; WARNING — MAP-WR-01.** |

## Frozen findings

### MAP-BL-01 — The authoritative source lock cannot be imported and its GraphQL half is silently ignored

**Severity:** BLOCKER

**Evidence:**

- The locked REST artifact is 12,920,264 bytes with exact SHA-256 `80850d...d5b1d` (`internal/connectors/defs/github/sources/github-operation-source-lock.json:5-10`), while the source lock also declares 31 GraphQL queries and 274 mutations (`internal/connectors/defs/github/sources/github-operation-source-lock.json:121883-121887`).
- `sourceImportLock` has only a `Rest` field (`cmd/connectorgen/sourceimport.go:109-113`). `parseSourceImportLock` decodes without rejecting unknown members and validates only `lock.Rest` (`cmd/connectorgen/sourceimport.go:304-324`), so the authoritative `graphql` member is ignored.
- The grammar-position index incorrectly reuses the artifact-byte ceiling (`cmd/connectorgen/sourceimport.go:1103-1121`). Running the built snapshot against the checked-in lock failed before descriptor emission: `connectorgen source-import: index source grammar positions: source grammar position byte limit exceeded`.
- The focused source-import tests passed because they use small fixture fetchers; they never run the real checked-in lock. The scoped additions in `cmd/connectorgen/main_test.go:662-760` cover hand-built direct-write declarations, not installed source import.

**Downstream impact:** No descriptor can currently be produced from the authoritative GitHub source lock. Even after raising the REST index limit, 305 locked GraphQL operations and their type system remain outside the importer. A “source-import passed” fixture result therefore cannot establish the first link in any installed-operation proof.

**Root cause:** The importer's lock schema predates the enriched schema-v2 REST+GraphQL lock, unknown lock members are accepted, and a memory/index accounting limit was coupled to raw artifact byte size rather than the actual indexed representation.

**Exact proposed changes:**

1. Replace `sourceImportLock` with a strict, versioned schema that models both REST and GraphQL and rejects unknown top-level sections.
2. Import the locked GraphQL root fields, arguments/input objects, return types, deprecation/preview state, and type-system references into the descriptor document.
3. Give the grammar-position index a separately named/measured bound based on the locked artifact's indexed structure; retain the raw artifact limit solely for downloaded bytes.
4. Add a hermetic checked-in-lock certification that runs `source-import --check` against exact cached/embedded locked bytes and asserts all 1,525 source identities are represented.

**Behavioral tests:**

- **Happy:** The exact 12,920,264-byte artifact plus locked GraphQL schema imports 1,220 REST + 31 query + 274 mutation descriptors with the expected SHA/source locations.
- **Bad:** An unknown source-lock section, changed SHA/byte count, duplicate source identity, or unmodeled GraphQL root causes a deterministic hard failure before output.
- **Edge:** A legal maximum-sized schema with many long JSON pointers stays within the independent index budget; exceeding that budget reports measured count/bytes without confusing it with artifact size.

### MAP-BL-02 — Imported descriptors are orphaned from bundle generation and validation

**Severity:** BLOCKER

**Evidence:**

- `source-import` requires an arbitrary `--out`, compares/writes that file, and exits; it never updates a connector bundle (`cmd/connectorgen/sourceimport.go:6230-6283`).
- Production references to `sourceImportDescriptorDocument` and `sourceOperationDescriptor` are confined to `cmd/connectorgen/sourceimport.go`; there is no bundle/App/CLI consumer.
- `validateDir` enumerates bundle directories and checks only the already-authored bundle (`cmd/connectorgen/validate.go:225-301`). It never loads the source lock or a source descriptor artifact.
- `surface-sync` reads existing `cli_surface.json`, `operations.json`, and `api_surface.json`, then iterates existing commands/operations (`cmd/connectorgen/surfacesync.go:268-380`); it cannot create an omitted provider operation.
- `params-import` independently reads raw OpenAPI and mutates only existing operations; a missing path or method silently continues (`cmd/connectorgen/paramsimport.go:157-217`). Against the exact locked artifact, `params-import --check` reported **211 operations differ**, yet a temporary bundle with those updates still passed `connectorgen validate` with zero findings.

**Downstream impact:** A missing operation, request field, server override, response field, or entire protocol can pass both validation and surface-sync. The existing equality of REST method/path sets is accidental state, not a maintained invariant. Every later App/CLI reachability assertion is disconnected from the authoritative source.

**Root cause:** Source import was implemented as a terminal reporting tool rather than the canonical input of generation; validation is internally referential and proves only that hand-authored bundle files agree with one another.

**Exact proposed changes:**

1. Check in one canonical descriptor artifact per locked provider source and make it the input to generation.
2. Generate/update `operations.json`, `writes.json`/streams, `api_surface.json`, `cli_surface.json`, and help inputs from every non-gap source descriptor.
3. Add one-to-one validation for source identity, method/path/server, all request parameters and schemas, response schemas/headers, and operation availability.
4. Represent intentional non-executable operations/fields in a typed, source-identity-bound gap ledger with a reason and owner; never satisfy coverage through mere endpoint presence.
5. Make both `validate` and `surface-sync --check` fail on a missing or stale canonical source projection.

**Behavioral tests:**

- **Happy:** Add a field and operation to a fixture source, regenerate, and observe matching operation, CLI/help, and exact App runner mappings.
- **Bad:** Delete an operation/field from any derived file and prove both validator and drift check fail with the source identity and missing projection.
- **Edge:** Multiple semantic actions may cover one provider endpoint only when their union covers every provider input and each action remains closed; aliases must not double-count coverage.

### MAP-BL-03 — Reverse-ETL commands omit hundreds of provider inputs; 65 advertised operations lack required bodies

**Severity:** BLOCKER

**Evidence:**

- Reverse-ETL validation inspects only the action's already-authored `record_schema` and `path_fields` (`cmd/connectorgen/validate.go:1664-1753`). It never compares either with the locked provider request.
- Example: `actions_permissions_access2` declares `PUT /actions/permissions/access` with an empty record schema (`internal/connectors/defs/github/writes.json:2687-2699`), and its implemented generated command has zero flags (`internal/connectors/defs/github/cli_surface.json:3829`). The exact locked OpenAPI requires body field `access_level`.
- Example: `actions_runners_generate_jitconfig` also has an empty record schema (`internal/connectors/defs/github/writes.json:2777-2789`) and zero generated flags (`internal/connectors/defs/github/cli_surface.json:4050`), while the provider requires `name`, `runner_group_id`, and `labels`.
- Exact method/path plus resolved JSON-schema audit against the SHA-pinned artifact found **281 omitted top-level JSON body fields across 99 write-covered endpoints**. **99 of those fields are required across 65 endpoints**, making the advertised commands incapable of a valid request. It also found five omitted write query parameters; required `key` for cache deletion is absent from the empty action/command (`writes.json:2596-2608`, `cli_surface.json:3498-3512`).
- The new validator tests are self-constructed direct-operation fixtures (`cmd/connectorgen/main_test.go:662-760`), so they cannot detect write-action drift from the installed provider source.

**Downstream impact:** Many commands marked `availability=implemented` deterministically send an empty or incomplete JSON body and receive a provider validation error. Optional provider capabilities are also unreachable. The generic reverse planner cannot repair this because its mapping validator accepts only fields named in `record_schema`; callers are correctly denied an undeclared raw-body escape hatch.

**Root cause:** The write-action derivation narrowed provider schemas, and the validator treats the narrowed declaration as authority instead of checking it against source. Endpoint-level coverage is counted even when its action union is field-incomplete.

**Exact proposed changes:**

1. Project every provider query and request-body field into closed action schemas and generated typed flags, preserving exact provider names and structured schemas.
2. Where semantic actions intentionally fix a subset (for example `close_issue`), add a complete provider-operation action/command or prove that the union of closed semantic actions covers every input.
3. Extend reverse-ETL validation to compare the union of actions covering each source endpoint against the imported request contract, including requiredness, types, enums, nullability, query/header placement, and nested fields.
4. Mark any still-incomplete command unavailable with a source-bound gap instead of `implemented`.

**Behavioral tests:**

- **Happy:** Installed commands for permissions access and JIT runner config expose all required typed fields, generate the exact provider JSON, and pass a fake-provider request assertion.
- **Bad:** Removing required `access_level`, `labels`, or cache `key` makes source projection/validation fail before build; an installed invocation never reaches I/O with an incomplete request.
- **Edge:** Optional fields, nullable unions, structured arrays/objects, and multiple semantic actions are all counted exactly once without allowing a raw body or undeclared property.

### MAP-BL-04 — Generated GraphQL operations narrow declared result fields and pagination inputs

**Severity:** BLOCKER

**Evidence:**

- The source lock's `CodeOfConduct` object declares ordinary fields including `body`, `id`, `key`, `name`, `resourcePath`, and `url` (`internal/connectors/defs/github/sources/github-operation-source-lock.json:40543-40547` and following fields).
- The installed operation selects only `codeOfConduct { __typename }` plus rate-limit metadata (`internal/connectors/defs/github/operations.json:8877-8886`).
- Exact audit found 305 source-generated GraphQL operations and 303 documents containing only a `__typename` selection for the provider root result; the only scalar/non-placeholder exceptions are root `id` and the explicit `rateLimit` query. Mutations likewise return only `__typename` (for example `github.graphql.mutation.abort-queued-migrations`).
- Exact root-argument comparison also found five queries (`marketplaceListings`, `search`, `securityAdvisories`, `securityVulnerabilities`, and `sponsorables`) omit the provider-declared `before` and `last` inputs, exposing only forward pagination. For example, the installed `search` operation declares only `after`, `first`, `query`, and `type` in its fixed document/schema.
- GraphQL command validation checks operation kind and top-level variable mappings (`cmd/connectorgen/validate.go:998-1080`) but never validates result selection against the locked return type. CLI faithfully prints the narrowed body (`internal/cli/cli.go:1397-1418`).

**Downstream impact:** The operations are nominally reachable but practically useless: object IDs, occurrence IDs, created/updated entities, payload errors, and ordinary provider fields cannot be requested. Five list/search roots cannot use the provider's backwards-pagination contract. This violates the complete, user-authorized provider surface while giving a false “implemented” signal.

**Root cause:** GraphQL generation created minimum syntactically valid placeholder documents rather than source-derived selection sets, and the ignored GraphQL lock has no validator-to-document link.

**Exact proposed changes:**

1. Generate source-derived, bounded typed selections for every result type, including scalar fields and recursively selectable object fields under explicit depth/page bounds.
2. Expose every exact typed root argument. Model forward/backward pagination as declaration-owned alternatives (`first`+`after` and `last`+`before`) and validate their mutually exclusive constraints rather than deleting one direction.
3. Validate every fixed document against the locked schema and fail when a declared ordinary field has no reachable selection path or typed gap.
4. Preserve the full returned GraphQL envelope subject only to exact credential-value masking (see MAP-BL-08/MAP-BL-11).

**Behavioral tests:**

- **Happy:** `graphql query code-of-conduct` returns exact `id`, `key`, `name`, `body`, `resourcePath`, and `url`; a mutation returns its declared payload IDs/entities.
- **Bad:** Replacing a generated selection with only `__typename`, selecting a removed field, or omitting a source field fails generation/validation.
- **Edge:** Interfaces/unions use exhaustive source-derived fragments, connections support bounded forward and backward pagination without mixing incompatible arguments, deprecated/preview fields retain their declared status, and scalar roots remain scalar.

### MAP-BL-05 — The advertised GitHub binary upload is a JSON call to the wrong origin with no file or required name

**Severity:** BLOCKER

**Evidence:**

- The source lock identifies `repos/upload-release-asset` as `POST /repos/{owner}/{repo}/releases/{release_id}/assets` (`internal/connectors/defs/github/sources/github-operation-source-lock.json:8717-8723`). The exact SHA-pinned OpenAPI operation declares server `https://uploads.github.com`, required query `name`, optional `label`, and an `application/octet-stream` binary body.
- The checked-in action instead declares ordinary `body_type: json`, only `release_id`, no query, no file field, and no operation-specific origin (`internal/connectors/defs/github/writes.json:5750-5770`).
- The installed command is misleadingly named `releases assets view-3`, is marked implemented, and exposes only `--release-id` (`internal/connectors/defs/github/cli_surface.json:10448-10469`).
- Scoped direct-write tests prove typed multipart and reject legacy `file_upload` (`internal/app/rest_write_command_test.go:710-735`), but they do not execute this installed GitHub upload declaration.

**Downstream impact:** Binary upload cannot upload a byte. It targets the default API origin, omits a required query parameter, and sends the wrong media/body format. Users receive a provider failure from a command advertised as implemented.

**Root cause:** The generator flattened a provider-specific binary upload into the generic reverse-ETL JSON action model and discarded per-operation server, query, media type, and file-source semantics.

**Exact proposed changes:**

1. Declare this as a bounded typed binary-upload operation/action with a confined local file field, required `name`, optional `label`, exact `application/octet-stream`, byte cap, and pinned uploads origin.
2. Route it through the existing plan/preview/single-use approval lifecycle and bind file identity/digest before approval.
3. Generate an accurate command name/help and refuse symlinks, traversal, non-regular files, changed files, oversize content, redirects, and retries without idempotency proof.
4. Add source validation for operation-level server overrides and binary request bodies.

**Behavioral tests:**

- **Happy:** The installed command sends byte-exact file content to `uploads.github.com/.../assets?name=...&label=...`, and preserves the provider asset receipt.
- **Bad:** Missing name/file, wrong media type/origin, changed post-preview file, traversal, symlink, or oversize file is refused before provider I/O.
- **Edge:** Empty file, non-ASCII URL-encoded label/name, network failure before response, and provider 4xx with a complete masked receipt.

### MAP-BL-06 — Reusing the advertised durable typed-destination authorization fails after the provider write and checkpoint

**Severity:** BLOCKER

**Evidence:**

- Once `AuthorizationReference` exists, the App intentionally accepts the plan again without a token and revalidates its durable authorization (`internal/app/declarative_typed_destination_approval.go:165-172`). CLI parsing explicitly allows this (`internal/cli/etl_transport_test.go:214-232`), and help promises later identical-scope reuse (`internal/cli/etl_transport.go:149-154`; the specialized help makes the same promise at `internal/cli/etl_transport.go:57-60`).
- After every successful definition-owned transport, `runTransportETL` calls `markDeclarativeTypedDestinationPlanExecuted` (`internal/app/transport_dispatch.go:351-359`), after the checkpoint commit callback has already persisted acknowledged stream state (`internal/app/transport_dispatch.go:252-283`).
- The marker accepts only `approval_consumption_uncertain`; an already `executed` plan returns an error (`internal/app/declarative_typed_destination_approval.go:339-359`). The analogous PostgreSQL marker correctly treats executed + durable authorization as idempotent (`internal/app/postgres_transport_approval.go:283-300`).

**Downstream impact:** The first run succeeds. A second authorized run can write provider data, pass read-back, and persist the source checkpoint, then return a failed run because final plan bookkeeping rejects `executed`. Callers may retry a mutation that already happened.

**Root cause:** The generic marker copied one-time plan semantics but the authorization path was made reusable; the state machine and post-effect transition were not updated together.

**Exact proposed changes:**

1. Make the generic marker idempotently accept `executed` only when the same non-empty authorization reference and exact plan mode/binding are present, matching the PostgreSQL pattern.
2. Separate reusable authorization state from per-run completion state so a plan need not transition on every authorized execution.
3. Ensure no bookkeeping operation after durable checkpoint commit can convert a successful external write into a failed command; reconcile such state on reopen.

**Behavioral tests:**

- **Happy:** Two sequential installed `pm etl run` invocations use the same plan, token only on the first, make exactly one write per run, and both finish successfully with distinct run/checkpoint identities.
- **Bad:** Replayed token, changed action/runtime/credential revision, expired/revoked authorization, or foreign plan fails before source/provider I/O.
- **Edge:** Zero-record second run, rate-limit park/rearm, process interruption after checkpoint but before marker, and concurrent same-plan runs reconcile without duplicate writes or false failure.

### MAP-BL-07 — Generic destination “read-back” never reads the provider but still authorizes checkpoint advancement

**Severity:** BLOCKER

**Evidence:**

- `ApplyDestination` performs the provider write and creates an acknowledgement (`internal/app/issue_label_warehouse_transport.go:195-266`).
- `ReadBackDestination` only recomputes local plan/action digests and checks workset ID, acknowledgement sink, and timestamp; it makes no connector/provider call (`internal/app/issue_label_warehouse_transport.go:297-321`).
- The orchestrator calls this method as the “independent read-back” immediately before committing a checkpoint (`internal/synctransport/orchestrator.go:249-283`).
- The composition test's fake server supports one source GET and one destination POST only, then asserts read/write/commit `1/1/1` (`internal/app/transport_composition_test.go:402-475`). Therefore the claimed installed path is proven without any destination GET/read-back. Help nevertheless says checkpoint follows declared durable acknowledgement and read-back (`internal/cli/etl_transport.go:149-154`).

**Downstream impact:** A provider can acknowledge a request without durably applying it, or apply a different value, and the source checkpoint still advances. The source record may never be retried, causing silent downstream data loss/divergence.

**Root cause:** The generic transport declaration models write/acknowledgement but no typed provider read-back operation or receipt-to-state matcher; local receipt validation was mislabeled as independent read-back.

**Exact proposed changes:**

1. Extend destination transport declarations with an exact read-back operation, typed identity projection from the workset/write receipt, expected-state matcher, bounds, and conformance evidence.
2. Execute that provider-owned read under its own timeout after acknowledgement and compare durable provider IDs/state before checkpoint commit.
3. If a provider cannot offer independent verification, declare that transport unavailable or explicitly weaker; do not expose it under the durable read-back contract.

**Behavioral tests:**

- **Happy:** POST persists an object, provider GET returns the exact ID/state, and only then does checkpoint commit occur.
- **Bad:** POST returns success but GET is missing/mismatched; run fails with no checkpoint advancement.
- **Edge:** Bounded eventual consistency, read-back rate limits, duplicated provider IDs, partial batches, and cancellation between write and read-back resume without reclassifying a local receipt check as provider proof.

### MAP-BL-08 — Configured credentials are deliberately persisted and printed when a provider echoes them

**Severity:** BLOCKER

**Evidence:**

- The output “sanitizers” ignore the supplied secret map and return provider results unchanged (`internal/connectors/connectors.go:921-931`).
- Generic typed destination output calls the identity sanitizer and attaches it to durable acknowledgement/output (`internal/app/issue_label_warehouse_transport.go:247-264`). Direct writes likewise retain the identity-sanitized result (`internal/app/app.go:2924-2929`).
- Scoped tests demand that configured `destination-secret` remain in provider headers/body/raw bytes and serialized runs (`internal/app/transport_composition_test.go:498-512`, `639-657`, `677-681`, `1089-1097`, `1171-1176`).
- User-facing docs explicitly promise provider values remain verbatim “even when they equal configured credential bytes” (`internal/cli/docs.go:923-927`, `1401-1407`), and generated agent skills repeat the unsafe promise (`internal/cli/skills.go:94-95`).

**Downstream impact:** A malicious, misconfigured, or diagnostic provider that reflects an Authorization token/API key causes the credential to be persisted in project state and emitted in CLI JSON. Encryption of the credential store does not protect the copied plaintext receipt.

**Root cause:** Complete provider output was incorrectly defined as byte-for-byte credential echo preservation. The implementation distinguishes system-generated diagnostics from provider output, but not ordinary provider values from exact known credential occurrences.

**Exact proposed changes:**

1. Preserve every provider field/key, orderable receipt element, ordinary value, large numeric occurrence ID, status, header presence, and raw-body length/encoding.
2. Mask exact configured credential values (and robust encoded forms where the value could be transported) in response headers, decoded JSON, raw text, and base64 projections; never drop the surrounding field.
3. Always mask declared secret response fields/headers. Represent masking with explicit markers while retaining presence/byte-count metadata.
4. Rewrite docs/skills/tests to require complete ordinary output **and** credential non-disclosure.

**Behavioral tests:**

- **Happy:** Rare fields, unfamiliar keys, paid-tier fields, duplicate headers, and a `9007199254740993` occurrence ID are preserved exactly.
- **Bad:** An exact configured token echoed in header, nested JSON, text raw body, or base64 form is absent from state/stdout/stderr and represented as masked.
- **Edge:** Overlapping credentials, very short values, repeated multi-value headers, non-UTF-8 bodies, substrings inside ordinary values, and declared secret fields whose value does not equal the configured credential.

### MAP-BL-09 — Installed CLI commands discard persisted failed runs and complete provider receipts

**Severity:** BLOCKER

**Evidence:**

- Approved ETL returns immediately on `RunETL` error before writing its `ETLRun` JSON (`internal/cli/etl_transport.go:486-506`).
- Both generated connector-command writes and `pm reverse run` return immediately on `RunReverseETL` error before writing `ReverseRun` (`internal/cli/cli.go:1808-1814`, `2146-2152`).
- The App deliberately returns a persisted failed run containing a complete HTTP 500 direct-write response (`internal/app/rest_write_command_test.go:738-786`) and partial destination results (`internal/app/transport_composition_test.go:1070-1178`).
- The rollup integration test merely calls another App test (`internal/app/foundations_integration_test.go:5-12`), and the composition test manually marshals an App run into a “CLI-shaped” envelope (`internal/app/transport_composition_test.go:677-681`); neither executes the installed CLI failure path.

**Downstream impact:** On the failures where provider receipts and run IDs matter most, the CLI exits nonzero with no typed terminal JSON containing the run. The user cannot see the receipt or even discover the persisted run ID for `status`, so App-level completeness is not an installed behavior.

**Root cause:** The App uses Go's `(value, error)` to return a meaningful terminal run alongside failure, while CLI helpers follow the conventional but incorrect “discard value on error” pattern.

**Exact proposed changes:**

1. Introduce a typed execution error carrying the persisted terminal run (or explicitly inspect the non-zero returned run).
2. In `--json` mode, write the complete credential-masked terminal envelope before returning the categorized nonzero error; keep stdout valid JSON.
3. In human mode, print at least run ID/status and a safe status lookup hint; never duplicate provider body onto stderr.

**Behavioral tests:**

- **Happy:** Successful ETL/reverse/direct write emits the unchanged terminal envelope and exits zero.
- **Bad:** Fake-provider HTTP 500 exits nonzero but emits one valid JSON envelope with run ID, failed status, and complete masked ordered receipt; stderr contains no provider body/credential.
- **Edge:** Partial multi-record failure, state-persist failure, no-response failure, and JSON writer failure preserve a deterministic single terminal-output contract.

### MAP-BL-10 — A direct-write transport failure before a response loses operation identity and response presence

**Severity:** BLOCKER

**Evidence:**

- `OperationDirectWrite` initializes a zero result and populates identity only inside `if response != nil` (`internal/connectors/engine/direct_write.go:92-99`, `155-160`). Network errors and a nil response therefore return an entirely zero `OperationDirectWriteResult` (`internal/connectors/engine/direct_write.go:161-185`).
- The result type is designed to carry connector, operation, method, path, and `response_received` (`internal/connectors/connectors.go:641-659`).
- App persists the operation result only when `ResponseReceived` is true (`internal/app/app.go:2924-2929`), so even an identity-only no-response receipt would currently be discarded.

**Downstream impact:** After approval has been consumed and dispatch attempted, the terminal run cannot distinguish “request attempted, no provider response” from “no direct operation result existed.” Connector/operation/method/path provenance is lost, weakening auditability and safe retry decisions.

**Root cause:** Result identity is constructed from a response rather than from the already sealed/prepared request, and App equates result presence with response presence.

**Exact proposed changes:**

1. Initialize the result from the prepared request before executing I/O with `ResponseReceived:false` and exact connector/operation/method/path.
2. Populate response status/headers/body only when a response exists.
3. Persist the result whenever sealed operation identity exists; use `ResponseReceived` solely to distinguish the response half.

**Behavioral tests:**

- **Happy:** A normal response retains the current complete receipt with `response_received=true`.
- **Bad:** Connection refusal/timeout after dispatch produces a persisted failed run with exact identity and `response_received=false`.
- **Edge:** Preflight refusal before approval/I/O has no attempted-operation receipt; context cancellation racing with dispatch is classified deterministically.

### MAP-BL-11 — Direct-read and binary-download results omit ordinary provider output and all provider error receipts

**Severity:** BLOCKER

**Evidence:**

- `DirectReadResult` has no operation ID, response-presence, body-presence/bytes/raw representation, and explicitly carries only response headers admitted by the declaration (`internal/connectors/connectors.go:458-477`).
- REST direct read returns a zero result for provider HTTP errors and otherwise returns only policy-shaped decoded JSON plus admitted headers (`internal/connectors/engine/direct_read.go:145-174`).
- GraphQL response parsing retains only each error message and drops error `path`, `locations`, `extensions`, and top-level `extensions` (`internal/connectors/engine/graphql_operation.go:707-753`).
- CLI emits only the narrowed decoded response/headers (`internal/cli/cli.go:1397-1421`).
- Binary download returns a zero result on provider failure and includes only admitted headers on success (`internal/connectors/engine/binary_read.go:160-200`); CLI can therefore expose no complete provider error receipt (`internal/cli/cli.go:1354-1372`).

**Downstream impact:** Ordinary provider request IDs, deprecation/sunset/rate headers, non-2xx bodies, raw invalid-JSON/text, GraphQL error occurrence IDs/extensions, and exact operation identity can be lost. Credential safety is achieved by omission rather than value masking, contrary to the required complete-provider-output contract.

**Root cause:** Read and download results use an older curated-output model, while writes introduced a complete receipt model. Output policy, credential redaction, and transport receipt preservation are conflated.

**Exact proposed changes:**

1. Use a shared complete response receipt for read/write/download metadata: operation identity, response presence, status, all headers with exact credential-value masking, body presence/bytes/raw encoding, decoded body when valid, and full GraphQL envelope.
2. Keep output-policy projection as an additional convenience view; never replace the complete receipt with it.
3. Return and emit response receipts alongside typed nonzero provider errors. For binary success, keep bytes in the confined file and report exact size/digest rather than inlining them.

**Behavioral tests:**

- **Happy:** REST/GraphQL reads preserve unfamiliar fields, all ordinary headers, exact large IDs, full GraphQL errors/extensions, and selected convenience output; binary download preserves byte-exact file+digest and complete metadata.
- **Bad:** Provider 4xx/5xx and invalid JSON return nonzero with a complete credential-masked receipt instead of a zero result.
- **Edge:** Header duplicates/casing, empty vs absent body, invalid UTF-8/base64 raw encoding, 200 GraphQL partial data, and binary media mismatch/oversize without a truncated success artifact.

### MAP-WR-01 — Generated connector skills are hard-coded to five connector names

**Severity:** WARNING

**Evidence:**

- `baseSkillDocs` iterates every manifest but appends a connector skill only when its computed name equals one of `pm-warehouse`, `pm-outbox`, `pm-file`, `pm-sample`, or `pm-github` (`internal/cli/skills.go:137-145`).
- `connectorSkill` itself is generic and can render any registered connector guide (`internal/cli/skills.go:149-160`), so the name branch is the only gate.
- Golden transcripts cover the skills namespace but not a non-allowlisted connector's generated skill (`internal/cli/golden_transcript_test.go:52`, `77`, `94`, `123`).

**Downstream impact:** Agent-facing generated documentation treats GitHub and four local connectors as special and silently omits every other connector, even when its manifest/guide/command surface exists. This is stale as the registry grows and makes API operations harder to discover through the advertised skills surface.

**Root cause:** A bootstrap allowlist became permanent product logic instead of a manifest/guide capability declaration.

**Exact proposed changes:** Remove the name allowlist. Generate a skill for each manifest that exposes a valid guide/command capability, or add an explicit manifest property controlling skill publication and validate it. Skip only with a typed reason, never connector-name literals.

**Behavioral tests:**

- **Happy:** A synthetic non-GitHub connector with a guide produces `pm-<name>` and an index entry.
- **Bad:** A manifest requesting skill publication without a registered guide fails generation rather than producing an empty/omitted file.
- **Edge:** Duplicate/sanitization-colliding names, local-only connectors, and a connector intentionally opting out are handled deterministically.

## Explicit PASS evidence

The following audited requirements had no independent finding beyond the failures already listed:

- **Snapshot integrity at review start:** exact requested SHA and clean status were confirmed before source inspection. Later concurrent review output did not change source.
- **REST endpoint identity coverage:** set comparison found zero locked REST method/path pairs missing from `api_surface.json`; five API-surface pairs are local additions. This is endpoint identity only, not field completeness (MAP-BL-02/MAP-BL-03).
- **Direct-operation request mappings:** after applying the exact locked artifact's `params-import` result in a temporary copy, every operation-backed installed command still had mappings for all imported path/query/header parameters, and direct operation JSON body top-level properties had zero omissions. This PASS does not cover write-action commands.
- **GraphQL input-object fields:** all nested fields of source-declared GraphQL input objects map exactly into the installed variable schemas. Five root queries still omit `before`/`last`, which is part of MAP-BL-04 rather than a PASS.
- **No generic raw HTTP/action flag:** implemented `raw_api` is rejected; direct commands dispatch through declaration-backed `commandrunner`; generic typed ETL accepts no connector/action/URL/method/body/mapping/evidence override (`cmd/connectorgen/validate.go:1141-1177`, `internal/cli/etl_transport.go:137-160`).
- **Direct-write approval gate:** request identity is previewed, sealed, revalidated, and single-use approval is consumed immediately before provider I/O (`internal/app/app.go:2895-2933`, `3073-3120`). No approval bypass was found.
- **Binary download success safety:** download is GET-only, capped, root-confined with `os.Root`, rejects truncation, and returns exact file size/SHA-256 metadata (`internal/connectors/engine/binary_read.go:87-105`, `184-215`).
- **Specialized issue-label transport selection:** App selection is based on an exact unique declaration capability and fails on ambiguity, with no connector-name/GitHub branch (`internal/app/issue_label_warehouse_transport.go:542-579`). Its provider read-back is real, unlike MAP-BL-07's generic adapter.
- **Caller action rejection:** persisted `destination_action` selects generic typed destination actions; `pm etl run` cannot substitute an action at runtime. Scoped ETL tests confirm unexpected shaping inputs are rejected before App/provider I/O.
- **Local warehouse durability:** WAL and final Parquet materialization are synced before acknowledgement/pending state (`internal/app/local_warehouse.go:255-310`). No scoped data-loss transition was found there.
- **Help/golden parity for changed static surfaces:** focused golden/transport/structured-help tests passed, and the new declarative transport cases are present in `internal/cli/golden_transcript_test.go:55-64`. The installed dynamic-leaf coverage gap is cited where it masks findings.
- **Generated diagnostic secrecy:** plan, approval, and synthetic error paths use redaction/sanitization and do not accept approval tokens in argv/environment/files. The finding is specifically provider output echoing configured credentials (MAP-BL-08).

## Verification record

- `go test -timeout 5m ./cmd/connectorgen -run '^(TestSourceImport.*|TestValidate_CLISurfaceReverseETLRequiresRiskAndApproval|TestValidate_CLISurfaceDirectWriteStructuredRESTBodyRequiresClosedBoundedDeclaration|TestCheckCLISurfaceOperationHeaderMappingsRequiresExactDeclaredHeader|TestSyncBundleDirectWriteDerivesOperationContract)$' -count=1` — **PASS**.
- `go test -timeout 5m ./internal/cli -run '^(TestETLRunTransportApprovalAllowsDurablePlanReferenceWithoutTokenCarrier|TestStructuredRESTBodyCommandHelpAndManualExposeOnlyDeclaredTypedFields|TestETLTransportBareAndLeafHelpAreContextual|TestGoldenTranscripts)$' -count=1` — **PASS**.
- `go test ./internal/app` as part of the three-package run — **PASS**.
- Real checked-in `source-import github` — **FAIL**, grammar-position byte limit exceeded (MAP-BL-01).
- Exact artifact SHA/byte verification — **PASS**; `params-import github --check` — **FAIL**, 211 operation declarations drifted (MAP-BL-02).
- Temporary updated bundle `connectorgen validate` — **PASS with zero findings**, demonstrating the missing source-to-validator invariant (MAP-BL-02).
- Full `go test ./cmd/connectorgen ./internal/app ./internal/cli` was not green: `cmd/connectorgen` failed because the current GitHub certification shard duplicates `operation:text_export`; `internal/cli` reached the default 10-minute timeout amid repeated unavailable Redis test endpoints while decoding the large runtime operation ledger. These are recorded as verification anomalies, not added to the frozen mapping findings because their owning sources are outside the explicit 25-file scope and they do not change the proven mapping defects above.
