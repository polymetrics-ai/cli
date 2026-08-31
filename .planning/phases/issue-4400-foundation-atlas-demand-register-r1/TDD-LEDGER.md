# TDD ledger — issue 4400 Foundation Atlas demand register

This is a documentation-only Atlas reconciliation.  It changes no executable
behavior, so a production-code red/green test is not applicable.  The red/green
checks below make the documentation contract observable instead.

| Slice | Red evidence | Green evidence | Edge/refactor evidence |
| --- | --- | --- | --- |
| Planned adapter records | Before editing, `jq -e` for all four new Atlas IDs exited `1`. | After editing, the same query must exit `0`; all four records must be `planned` and `connector_specific_reference`. | Unique IDs and schema parsing reject duplicate/malformed catalog entries. |
| Demand-register visibility | Before editing, `test -f docs/connector-canon/foundations/BATCH-R1-DEMAND-REGISTER.md` exited `1`. | After editing, the file must exist and name every 12 M-F source/lane identity. | A focused source-accounting report must still report 12 M-F and zero executable declarations; docs must not alter its inputs. |
| Non-execution boundary | Existing connector ledgers and planned GitLab/Stripe entries declare no receiver/executor. | New entries and register repeat that no inbound request, sync transport, worker, executor, CLI command, or provider I/O path is created. | Diff scope excludes source locks, matrices, artifacts, runtime, contract, and certification files. |

## Commands

```sh
jq empty docs/connector-canon/foundations/catalog.schema.json \
  docs/connector-canon/foundations/catalog.json
jq -e '([.foundations[].id] | length) == ([.foundations[].id] | unique | length)' \
  docs/connector-canon/foundations/catalog.json
go run ./cmd/connectorgen deferred-visibility \
  data/connector-canon/batch1-source-operation-mapping-cohort.json --check --json
git diff --check
```

No credentialed, provider-I/O, executor, race, or broad runtime test is allowed
for this documentation-only slice.
