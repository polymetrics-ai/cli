# Asana complete source-backed parity R1

## Delivery binding

- Remote integration base: `7212f14bf8b602c317f30c6e0addcfb6655d88c4`.
- Final code/artifact commit: `04fa06e99c815a8f88f8999c067ed22cd75b92c4`.
- Final Asana docs/Foundation Atlas commit:
  `0a83a1f2a34c4f5047f3d849cf358f9981faeb47`.
- Delivery state: local commits only; no push, merge, force update, provider call,
  credential access, or live-certification claim.
- The pre-existing dirty files
  `.planning/phases/issue-380-asana-parity/{PLAN.md,TDD-LEDGER.md}` are not part
  of either final commit and are intentionally excluded from this trace commit.

## Ordered atomic chain

The complete local chain from the remote integration base is:

| Commit | Contract delivered |
| --- | --- |
| `f28f5a05d` | materialize the source-backed membership direct read |
| `b9b892c81` | coerce closed action-backed direct-write JSON flags |
| `7ef0a41c6` | preserve connector JSON strings through the warehouse |
| `faf3c5627` | preserve saved ETL/reverse-ETL evidence beside direct lanes |
| `05eeb1497` | validate action-backed direct commands |
| `4b801a98a` | materialize the 34 deferred Asana mutations |
| `aa4893221` | apply one aggregate request budget to interactive stream reads |
| `dbb0e5071` | admit source-declared unrestricted attachment media safely |
| `40d449ad3` | execute the closed declared-action batch adapter |
| `7497e753a` | encode source-declared query arrays |
| `f05ad48a6` | retain all-empty-object warehouse batches |
| `ca44eec8f` | exclude OpenAPI read-only request fields |
| `e7feebf3e` | recognize the implemented batch lanes in evidence |
| `5c45f123e` | record the shared connector foundations/procedure |
| `50abc95ab` | decouple source mapping from certification admission |
| `459bab956` | trace the authoring-admission RGR |
| `cb16f77dc` | bind rendered-reference citations |
| `1b40b270f` | trace rendered-citation RGR |
| `f32177160` | admit closed enriched Asana source fields |
| `2f02b9f91` | expose mapping-wire variant drift |
| `668298e82` | restore structural mapping validation |
| `d77fd5b24` | close mapping-reader wire variants |
| `a83776bba` | bind mapping-reader R3 evidence |
| `845b00f5f` | bind retained Asana source-operation identities |
| `a7c28d832` | trace source-backed mapping/multipart repair |
| `efa6d8a0e` | migrate the Asana source lock to strict schema v3 |
| `514446894` | close the v3 event-inventory mapping wire |
| `9c32a2bdd` | trace v3 event-inventory mapping |
| `04fa06e99` | preserve request requiredness, query encoding, phase-qualified gaps, and the closed source-backed batch inventory through projection |
| `0a83a1f2a` | synchronize Asana docs, resolve the obsolete parser gap, and update the Foundation Atlas/procedure |

## Final source authority and lane census

The retained provider artifact remains the pinned Asana OpenAPI 3.0.0 document
at commit `56796a67a3c093eedf55fd9682357957a2ebfd85`, SHA-256
`cb3b90f4e0af56035eab0c648974f625b942a28a7144aa6c2326e38ca0bb3d56`,
3,066,750 bytes. The schema-v3 connector source lock contains the same 249
operation identity tuples plus closed event-schema and batch-action selector
inventories. Its final file SHA-256 is
`eb5517f0f1456e4cacb03d4f705fd2244a1ccf6113a9c0d67e29ed2417e407c4`.

| Surface | Final source-backed disposition |
| --- | --- |
| Locked operations | 249: GET 119, POST 81, PUT 26, DELETE 23 |
| Direct read | 119 implemented commands: 107 operation-backed and 12 stream-backed with one aggregate interactive request budget |
| Direct write | 131 implemented one-record actions across all 130 mutation endpoints; the attachment endpoint has two closed request variants |
| Binary upload | 1 implemented alias over the bounded multipart file action and source-declared `provider_unrestricted` media policy |
| Binary download | 0; the retained source contains no binary response contract, so this lane is provider-evidenced not applicable |
| ETL | all 12 streams support saved exhaustive full-refresh overwrite/append; project-scoped `tasks` additionally supports event-token `incremental_append`, `incremental_upsert`, and `incremental_dedupe` |
| Reverse ETL | all 131 actions are available to warehouse-table reverse delivery; interactive commands remain one-record direct writes |
| CLI accounting | 252 rows: 119 implemented direct reads, 131 implemented direct writes, one implemented binary-upload alias, and one planned legacy attachment operation alias |

`POST /batch` is implemented without generic HTTP. The source lock binds the
provider operation, canonical request/action/response schemas, maximum of 10,
method enum, and envelope fields. The connector definition allow-lists named
existing actions, while the engine derives each method, relative path, and typed
body. Caller-authored methods, paths, headers, raw bodies, nested batches,
uploads, query-bearing subrequests, and non-allow-listed actions fail before
provider I/O. The source-backed outer `opt_fields` and `opt_pretty` parameters
remain ordinary projected query fields.

## Red / green evidence for the final source-first wave

| Witness | Red | Green |
| --- | --- | --- |
| Single `data` envelope requiredness | required `requestBody` did not make the sole record envelope required | `TestSourceProjectionSingleDataEnvelopeUsesRequestBodyRequiredness` proves required and optional bodies independently |
| Form/explode-false arrays | source arrays were emitted as an unencoded record value | `TestSourceProjectionQueryArrayUsesSourceFormEncoding` proves `join:,`, absence, and scalar preservation |
| Existing CLI presentation | regenerated flags sorted away declaration-owned names/order | `TestSourceProjectCommandPreservesDeclaredFlagOrderAndNames` proves stable presentation with source semantics |
| OAS reference phase | a response-only schema sibling blocked a complete bounded request | request/response phase tests admit the membership read while retaining its response diagnostic |
| Batch inventory | an unknown lock field failed strict parsing | the closed happy/bad inventory matrix resolves exact selectors/max/methods/envelopes and rejects unbound or open variants |
| Executable batch coverage | `sourceActionCoversOperation` dropped `DeclaredBatch` and `RecordSchema`; focused RED failed with `source-backed declared batch lost its closed spec or record schema during executable-coverage validation` | `TestSourceProjectionPreservesOnlySourceBackedDeclaredBatch` validates the closed spec/schema, source-aware outer queries, and negative wrong-query/schema/batch cases |
| Retained full coverage | the first full package run failed only on `asana.rest.createBatchRequest` as an unresolved source-bound gap | the retained 249-operation coverage test and final full package run are green with no coverage finding |

## Exact verification

```text
go test ./cmd/connectorgen -run '^(TestSourceProjectionPreservesOnlySourceBackedDeclaredBatch|TestRetainedAsanaMutationDispositionsCoverEveryDeferredSourceOperation)$' -count=1
# PASS: ok polymetrics.ai/cmd/connectorgen 7.582s

go test ./internal/connectors/defs/asana -run '^TestBatchActionUsesClosedSourceBackedActionSelection$' -count=1
# PASS: ok polymetrics.ai/internal/connectors/defs/asana 5.835s

go run ./cmd/connectorgen source-import asana --defs internal/connectors/defs --check
go run ./cmd/connectorgen source-import asana --defs internal/connectors/defs --read-projection-only --check
# PASS: both verify 249 operation(s), 0 inbound event(s)

go test ./internal/connectors/defs/asana -count=1
# PASS: ok polymetrics.ai/internal/connectors/defs/asana 41.391s

go test ./cmd/connectorgen -count=1
# PASS: ok polymetrics.ai/cmd/connectorgen 332.205s

bash scripts/tests/connector-canon.sh
# PASS: connector canon check: ok

go run ./cmd/pm docs validate --connectors-dir /tmp/cli-asana-docs.ET7QkZ/connectors
# PASS: isolated generated connector-doc tree validates

go run ./cmd/pm docs generate --dir /tmp/cli-asana-docs-review.BkxvaI/cli --connectors-dir /tmp/cli-asana-docs-review.BkxvaI/connectors
# PASS: canonical generator changed only the Asana entries in the aggregate
# connector README, Markdown catalog, and JSON catalog

go run ./cmd/pm docs validate --connectors-dir docs/connectors
# PASS: checked-in connector docs and aggregate catalogs validate

jq empty docs/connector-canon/foundations/catalog.schema.json docs/connector-canon/foundations/catalog.json internal/connectors/defs/asana/missing-foundation.json
jq -e '([.foundations[].id] | length) == ([.foundations[].id] | unique | length)' docs/connector-canon/foundations/catalog.json
git diff --check
# PASS
```

The canonical `pm docs generate` command generated a full isolated tree. Byte
comparison against the checked-in tree found Asana-entry differences only in
the three shared generated docs, `docs/connectors/README.md` and
`docs/connectors/catalog/all-connectors.{json,md}`; only those generated bytes
were copied into the checked-in tree, and the full-tree validator now passes.
The global generated `operation-evidence.json`, declaration-admission ledgers,
and endpoint ledger remain untouched; their multi-connector regeneration is
not hand-edited inside an Asana parity slice.

## Remaining honest limitations

There is no remaining mapping or runtime-foundation blocker for any of the 249
locked Asana operations or for the direct-read, direct-write, saved full-refresh
ETL, warehouse reverse-ETL, or binary-upload claims above. The remaining limits
are narrower and stay explicit:

- Provider evidence proves event-token incremental semantics only for
  project-scoped `tasks`. The other 11 streams remain full-refresh-only for
  incremental purposes; this is missing provider scope evidence, not a mapping
  or certification block.
- Asana documents no total event order, so ordered history/change-capture mode
  is not claimed.
- The legacy `attachments create-attachment-for-object` operation alias remains
  planned because the generic operation executor has no typed multipart route;
  the provider operation itself is fully executable through two closed direct
  actions and the binary-upload alias.
- Saved bulk multipart file paths still resolve relative to the app-owned
  `.polymetrics` runtime directory, while one-record direct writes use the
  project root. This is a shared bulk-path policy discrepancy, not missing
  Asana endpoint coverage.
- No provider idempotency header is fabricated. Initial writes execute, but an
  ambiguous provider result remains a no-auto-retry condition unless the
  provider contract supplies a safe key/readback rule.
- Live certification remains proof-only. Absence of accepted live evidence does
  not retract the source mapping or independently tested runtime admission.
