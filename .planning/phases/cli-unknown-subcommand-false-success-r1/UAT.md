# UAT — cli-unknown-subcommand-false-success-r1

## Automated acceptance evidence

| Deliverable | Evidence | Result |
| --- | --- | --- |
| Unknown deep connector help does not silently succeed | `TestDynamicConnectorDeepHelpPathsResolveOrReportUsage/unknown_deep_command` | Pass: exit is non-zero, output names `calls definitely-not-real`, and root help is absent. |
| JSON preserves the usage-error contract | `TestDynamicConnectorDeepHelpPathsResolveOrReportUsage/unknown_deep_command_JSON` and golden transcript `dynamic_connector_unknown_deep_help_json` | Pass: exit 2 and JSON `error.category=usage`, `error.code=usage_error`. |
| Valid deep help remains valid | `TestDynamicConnectorDeepHelpPathsResolveOrReportUsage/real_deep_command` | Pass: `pm gong calls transcript --help` exits 0 and renders its command manual. |
| Connector root/help behavior remains valid | `TestDynamicConnectorHelpAndBareNamespace` | Pass. |
| Existing unknown-path handling remains valid | `TestDynamicConnectorUnknownPathIsUsageError` and `TestBahmniBareCommandGroupInvalidMultiPartPathIsNotHelp` | Pass. |

No human judgment is required: each acceptance criterion is deterministic and covered by the CLI
test harness and the built binary probe recorded in `TDD-LEDGER.md`.
