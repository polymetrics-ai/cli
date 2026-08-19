# Provider-neutral foundation gaps — issue #4291

Machine-readable rows: `FOUNDATION-GAPS.json`. An operation is not enabled for an affected surface
while its stable gap remains open, and it cannot support a merge-ready verdict.

| Stable gap id | Affected operations | Connectors | Owner | Status | Closure proof |
| --- | ---: | ---: | --- | --- | --- |
| `declarative-typed-destination-app-dispatch` | 9 | 9 | #4304 / `fm/cli-reverse-etl-destination-r1` | resolved by `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` | Definition validation, surface sync, runtime preflight, and installed App/CLI fixture path remain required per connector. |
| `declarative-typed-destination-exact-action-selection` | 369 | 9 | #4304 / `fm/cli-reverse-etl-destination-r1` | resolved by `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` | Persisted exact action selection is foundation-proven; each action still needs its connector-owned source mapping and fixture evidence. |
| `declarative-typed-destination-camelcase-input-identifier` | 1 | 1 | #4304 / `fm/cli-reverse-etl-destination-r1` | resolved by `609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57` | The exact schema-known `conversationId` mapping validates; malformed, unknown, generic, and cross-action bindings remain refused before I/O. |
| `declarative-typed-destination-action-specific-source-bindings` | 1 | 1 | #4304 / `fm/cli-reverse-etl-destination-r1` | open | Permit schema-validated, definition-owned bindings per selected action for one source executor; prove Help Scout `conversations.id → conversationId` and `customers.id → customerId`, then run generated and fixture App/CLI checks. |
| `declarative-operation-route-override` | 5 | 1 | `cli-operation-route-override-foundation-r1` | open | Add a closed definition-owned per-operation base/version/route contract across every executor; prove Help Scout's five `/v3` reads never route to `/v2/v3`, while arbitrary caller URLs and undeclared routes fail before I/O. |

The resolved historical fan-out is preserved in the JSON: Batch 6 has 220 operations across five
connectors; Batch 7 has 159 resolved operations across four connectors plus six open Help Scout rows:
one `PATCH /v2/customers/{customerId}` destination-binding row and five source-locked `/v3` direct
reads. The JSON carries every row's exact provider operation, URL, revision/hash, validator/runtime
evidence, owner, and closure verification.

No shared gap is hidden as an unexplained disabled row, `N/A`, or a connector-specific workaround.
An open-gap operation is explicitly foundation-blocked and not enabled; `N/A` is allowed only when a
source-locked provider operation proves the corresponding capability absent.
