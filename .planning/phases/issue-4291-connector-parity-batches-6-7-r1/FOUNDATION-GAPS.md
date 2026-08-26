# Provider-neutral foundation gaps — issue #4291

Machine-readable rows: `FOUNDATION-GAPS.json`. An operation is not enabled for an affected surface
while its stable gap remains open, and it cannot support a merge-ready verdict.

| Stable gap id | Affected operations | Connectors | Owner | Status | Closure proof |
| --- | ---: | ---: | --- | --- | --- |
| `declarative-typed-destination-app-dispatch` | 9 | 9 | #4304 / `fm/cli-reverse-etl-destination-r1` | resolved by `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` | Definition validation, surface sync, runtime preflight, and installed App/CLI fixture path remain required per connector. |
| `declarative-typed-destination-exact-action-selection` | 369 | 9 | #4304 / `fm/cli-reverse-etl-destination-r1` | resolved by `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` | Persisted exact action selection is foundation-proven; each action still needs its connector-owned source mapping and fixture evidence. |
| `declarative-typed-destination-camelcase-input-identifier` | 1 | 1 | #4304 / `fm/cli-reverse-etl-destination-r1` | resolved by `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` | The exact schema-known `conversationId` mapping validates; malformed, unknown, generic, and cross-action bindings remain refused before I/O. |
| `declarative-typed-destination-action-specific-source-bindings` | 1 | 1 | #4304 / `fm/cli-reverse-etl-destination-r1` | resolved by `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` | Help Scout now declares `customers.id → customerId` for `update_customer`; connector conformance remains a separate declaration task. |
| `declarative-operation-route-override` | 5 | 1 | `cli-operation-route-override-foundation-r1` | resolved by `6410fe59c` | `mailbox_v3` resolves all five source-locked paths before I/O; the focused engine route test passes. |
| `scalar_json_write_body` | 2 | 1 | #4291 foundation handoff | open | Add a closed schema-validated bare scalar JSON write body; `write.go:674-692` currently materializes an object. |
| `structured_recursive_filter_input` | 1 | 1 | #4291 foundation handoff | open | Add a closed validated builder for Gorgias's recursive statistics filter; do not permit a generic JSON-body escape hatch. |
| `post_binary_text_export` | 1 | 1 | #4291 foundation handoff | open | Add a bounded POST binary/text-export executor with exact request/response declarations. |
| `put_operation_direct_read` | 1 | 1 | #4291 foundation handoff | open | Add a bounded PUT-capable operation direct-read executor with current route/body validation. |

The resolved historical fan-out is preserved in the JSON. The former six Help Scout rows are
resolved by their shipped foundations and connector declarations. The current open fan-out is five
Gorgias rows across four provider-neutral executor capabilities; the separately documented signed
redirect download is `provider-contract-unavailable`, not a foundation gap. The JSON carries every
row's exact provider operation, URL, revision/hash, validator/runtime evidence, owner, and closure
verification.

No shared gap is hidden as an unexplained disabled row, `N/A`, or a connector-specific workaround.
An open-gap operation is explicitly foundation-blocked and not enabled; `N/A` is allowed only when a
source-locked provider operation proves the corresponding capability absent.
