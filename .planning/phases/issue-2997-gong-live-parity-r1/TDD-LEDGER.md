# TDD ledger: Gong release-0.3.0 live parity reconciliation

## Red → green slices

| Slice | Red evidence | Green evidence | Refactor boundary |
| --- | --- | --- | --- |
| Current source inventory | Fresh official OpenAPI inventory had not been compared to the Batch 2/3 source lock. | Credential-free fetch proves all 69 semantic operation rows match the refreshed strict lock exactly. | Lock the current exact artifact and compare a sorted semantic fingerprint so serialization-only changes cannot erase operations. |
| Foundation reconciliation | Preserved branch predates current main, typed destinations, and Batch 2/3 source disposition evidence. | Merge ancestry proves retained branch history plus the exact published foundation heads. | Resolve connector-owned declarations, not provider-named engine conditions. |
| Direct-read exact endpoint binding | Reproduce any Gong command that preflights with an implemented operation but no matching `api_surface` row. | Real `commandrunner.Preflight`, `surface-reconcile`, and a built CLI preflight sweep accept each declared direct read up to missing credentials. | Let `surface-sync` derive operation-owned metadata; do not hand-author it. |
| Typed write and reverse-ETL declarations | The historical CLI marks 24 exact write actions partial; the three multipart actions were recorded as an F4 foundation gap. | All 27 named actions are implemented, their CLI field shapes pass runtime preflight, and focused Gong multipart conformance passes through the shared approval-digest path. | No generic writer, raw body, arbitrary endpoint, or Gong-specific shared branch. |
| Provider output preservation | Gong command metadata described ordinary provider response fields as redacted, and the shared result boundary masks an undeclared provider value that happens to equal a credential. | A focused Gong surface test fails on direct-read redaction language or read-field declarations; the provider-neutral foundation test for #4321 fails until an undeclared matching scalar, header, raw body, and cursor remain exact while an explicitly declared secret field stays visibly masked. | Keep declared-secret masking at the generic output boundary; do not create field-name heuristics or a Gong branch. |
| Eight-surface enabled parity | A source-locked operation or application workflow can be structurally declared yet be absent from CLI/App dispatch, generated docs, or a supported ETL, reverse-ETL, direct, binary, flow, or schedule path. | Generated inventory-to-surface evidence and built-CLI/App checks classify each of ETL, reverse ETL, direct read, direct write, binary download, binary upload, flow, and schedule as proven or exact-source/application-contract `not_applicable`. | Safety, scope, tier, and destructive metadata can add confirmation; they cannot disable a provider-defined operation or workflow. |
| External-proof configuration privacy | The external-proof artifact stores its outer process command verbatim. An account-scoped connector `--config` value would therefore become generated evidence even when every secret is fingerprinted. | Provider-neutral foundation #4337 must retain only safe argument structure/fingerprints before a tenant-scoped base URL is used in a proof-producing run. | Do not bypass this by using a Gong branch, a generic public endpoint, or an untracked proof. |
| Source-document identity query | The legacy v2 Gong lock reaches `artifact URL must not include a query`, so its required fixed `?version=` document cannot be retrieved. | The v3 `gong-v2` document declares `identity_query: true`; the real scoped importer fetches and parses the locked artifact before reaching an unrelated generic request-schema limit. | Retain the exact 69 rows, digest, bytes, and fixed query; do not invent a queryless mirror, a Gong exception, or a provider schema bound. |
| Certification evidence | No credential reference means live stages cannot assert persisted provider state. | Credential-free gates are green and the remaining external block is explicit and secret-free. | Do not substitute browser authentication or fixtures for live certification. |

## Recorded red evidence

- Source-lock import and Batch 2/3 declarations are absent from the preserved branch; current
  `origin/main` also does not contain the Batch 2/3 source-lock files.
- The historical branch's phase records 67 operations. The current official OpenAPI has 69,
  confirming that historical completion evidence is insufficient for this release certification.
- Direct-read runtime coverage must be re-proven after reconciliation because prior audits found
  declaration rows that validated structurally but lacked exact executable `api_surface` bindings.
- The first output-preservation assertion did not compile until the focused test model was extended
  with CLI risk metadata; it then failed on `calls get` claiming its fields were redacted. This
  established a declaration-level red case without provider I/O.
- Captain policy audit found collision masking in the shared public provider-result projections:
  `SanitizeProviderOutputForOutput`, receipt/header/write projections, and direct-read cursors
  mask a scalar solely because it equals configured credential material. Foundation issue #4321
  records the required provider-neutral red test and fix; this Gong lane does not add an exception.
- `node .../verify-parity-maps.mjs` failed after the shared multipart foundation was merged because
  its generated ledger still carried the now-obsolete Gong F4 rows. Removing that stale special
  case and regenerating the ledger made the 19-connector/5,127-operation check pass with zero
  Gong gap rows.
- Before #4335, `go run ./cmd/connectorgen source-import gong --check` rejected the legacy v2
  lock before fetch because the official fixed URL contains `?version=`. The red case establishes
  why merely retaining the historical v2 ledger was insufficient.
- `go run ./cmd/connectorgen params-import gong --artifact /tmp/gong-openapi-20260823.json --check`
  reported three declaration drifts. The generator identified the three multipart operations, not
  an inferred provider policy; its first green run is recorded below.
- A live `pm gong targets list` reached Gong but returned only the safe HTTP-400 classification.
  The current official contract says `workspaceId` is required for `GET /v2/targets`, while the
  generated direct-read flag is still optional. The canonical source descriptor needed for
  `surface-sync` cannot be generated until generic source-import common-input preflight retains
  the provider contract, so this is visible as a source-projection dependency rather than
  hand-authoring a flag.
- The initial certification declaration named that invalid target-list call. The focused
  certification test went red until the declaration selected the ordinary, bounded
  `users extensive` typed read instead.
- The eight-surface audit found that binary upload, flow, and schedule must be separate cells:
  Gong has three fixed multipart uploads; the generic flow and schedule roundtrips use Gong as
  the declared source. The latter two are application workflows, not undocumented Gong provider
  endpoints. The current external-proof serializer also retains raw outer command arguments, so
  tenant-scoped configuration cannot be passed through it until the provider-neutral privacy
  boundary is corrected.

## Green evidence recorded during execution

- `go test -timeout 20m ./cmd/connectorgen -run '^TestGong(FullSurfaceCommandAndOperationCoverage|MetadataEnablesWriteCapability)$' -count=1` passed after the reverse-ETL and output-preservation declarations were corrected.
- `go test -timeout 20m ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1` passed for every implemented bundle command.
- `go test -timeout 20m ./internal/connectors/conformance -run 'TestConformance/gong' -count=1` passed, including the three declaration-owned multipart actions.
- In one freshly initialized project with no configured credential, the built binary classified all 30 direct-read commands, all 27 reverse-ETL write commands, and all 12 ETL stream commands as `missing --credential`. Each command was invoked without provider credentials; zero classified as unknown, partial, or unbound.
- `node .planning/phases/issue-4289-parity-map-batches-2-3-r1/traces/verify-parity-maps.mjs` passed: 19 connectors / 5,127 documented operations, with zero remaining Gong foundation-gap rows.
- `go run ./cmd/connectorgen params-import gong --artifact /tmp/gong-openapi-20260823.json`
  updated exactly three multipart operation parameter declarations; the immediate `--check` then
  reported 17 scanned / 0 updated.
- `go test -timeout 20m ./cmd/connectorgen -run '^TestGong(CertificationDeclarationUsesOnlyOrdinaryRESTLiveCandidates|FullSurfaceCommandAndOperationCoverage|MetadataEnablesWriteCapability)$' -count=1`
  passed. `certification-candidates` and `certification-sweep` generated and then passed their
  Gong `--check` modes (71 rows / 69 CLI commands).
- Gong declarations now state the captain rule: all ordinary provider values, including a value
  equal to configured credential material, must be preserved; only explicitly declared secret
  fields may be masked. The shared runtime implementation is tracked as an open #4321 dependency.
- The strengthened Gong declaration test first failed because three direct-read descriptions
  (`meetings integration-status`, `flows steps`, and `flows prospects`) omitted the policy. After
  declaration-only corrections, `TestGongFullSurfaceCommandAndOperationCoverage` passed and its
  JSON assertion proved all 30 direct-read descriptions carry the policy.
- The rebuilt CLI authenticated through the persisted App credential path. A bounded `users list`
  ETL read returned one record, and the bounded ordinary `users extensive` direct read returned a
  successful provider response with page context. No provider fields or identifiers were retained
  in this ledger.
- `pm connectors certify gong --direct-read-only --external-proof` produced a passing scoped
  report: preflight, credential test, and `gong_ordinary_rest_users_extensive` passed; the report
  had six pass stages, one documented skip, and zero leaks. The local external-proof artifact is
  not committed; its non-secret SHA-256 is
  `3c1a3e16480a139029dc2220f760b4726a0c8c71fb2fc35f97fd06e7a85d76e9`.
- A fresh built-binary preflight sweep of all 30 direct reads produced 30 credential gates, zero
  unknown commands, zero unexpected successes, and no provider request. With the persisted
  credential, a missing typed `--call-id` was rejected before provider I/O, and cursor-based
  `users extensive --page 2` was rejected before provider I/O.
- The same rebuilt binary swept the complete current declaration surface: 69 commands (30 direct
  reads, 12 ETL streams, and 27 reverse-ETL writes) all reached the credential gate; unknown,
  partial, unbound, and unexpected-success counts were zero. This was credential-free and made no
  provider request.
- After merging #4335 at `8127de418`, the v3 Gong lock preserved all 69 rows and declared its
  sole fixed artifact query as `identity_query: true`. A real `source-import --check` traversed
  that query and parsed the source, then stopped at the generic
  `/v2/all-permission-profiles` parameter-0 `maxLength` preflight limit. This is green evidence
  for query identity retrieval, not a false source-projection or validation success.

## Green evidence to record during execution

- exact inventory diff result, generated source-map result, and source/disposition arithmetic;
- focused Gong test names/results and direct-read built-binary classifications;
- generator, docs, boundary, and static gate results;
- an explicit live-certification result or the one non-secret credential-reference blocker.
- eight-surface inventory, CLI/help/manual/website reachability, output-preservation, and App-path
  classifications for every supported provider operation; any `not_applicable` status cites the
  exact source-audit row(s), never a safety or tier label.
- Complete full-parity live certification only after Gong has declaration-owned, self-cleaning
  write pairings and the canonical source descriptor can project every provider-required direct
  input. The current full harness evidence remains bounded and explicitly partial; it is not a
  full-parity claim.
