# #2985 stale api-surface reason report

This is a check-only report from the deterministic reconciler introduced by
this phase. It selects the former issue-#2985 reason text and performs no
bundle write:

```sh
go run ./cmd/connectorgen surface-reconcile internal/connectors/defs --check --reason-contains '#2985'
```

| Connector | Rows scanned | Covered by a runnable command | Retained as blocked | Refused |
| --- | ---: | ---: | ---: | ---: |
| Zendesk Support | 306 | 0 | 306 | 0 |
| HubSpot | 260 | 0 | 260 | 0 |
| Asana | 3 | 0 | 3 | 0 |
| Bitbucket | 3 | 0 | 3 | 0 |
| Freshchat | 1 | 0 | 1 | 0 |
| YouTube Analytics | 1 | 0 | 1 | 0 |
| **Total** | **574** | **0** | **574** | **0** |

The command intentionally exits nonzero in `--check` mode because all 574
rows would receive a newly derived current blocked reason. It does not promote
or edit Zendesk Support, HubSpot, or any other connector. The next wave is
Zendesk Support command adoption: after each command passes the same runtime
preflight, rerun this tool to derive coverage rather than hand-editing a row.
