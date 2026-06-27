# TDD-LEDGER — Action Step (Phase 1)

Format: test ID | file | red-evidence (compile/test failure output) | green-at (commit/timestamp)

| Test ID | File | Red Evidence | Green At |
|---------|------|-------------|----------|
| T-10 | internal/flow/action_test.go | `undefined: KindAction; unknown field ActionCfg; undefined: ActionConfig` | GREEN 2026-06-27 |
| T-11 | internal/flow/action_test.go | `undefined: HTTPActionRunner` | GREEN 2026-06-27 |
| T-12 | internal/flow/action_test.go | `undefined: HTTPActionRunner` | GREEN 2026-06-27 |
| T-13 | internal/flow/action_test.go | `undefined: HTTPActionRunner` | GREEN 2026-06-27 |
| T-14 | internal/flow/action_test.go | `undefined: HTTPActionRunner` | GREEN 2026-06-27 |
| T-15 | internal/flow/action_test.go | `undefined: HTTPActionRunner; undefined: deterministicRecordID` | GREEN 2026-06-27 |
| T-16 | internal/flow/action_test.go | `undefined: HTTPActionRunner; undefined: ErrSchemaDrift` | GREEN 2026-06-27 |
| T-16b | internal/flow/action_test.go | `undefined: HTTPActionRunner` | GREEN 2026-06-27 |
| T-17 | internal/flow/action_test.go | `undefined: HTTPActionRunner` | GREEN 2026-06-27 |
| T-18 | internal/flow/action_test.go | `undefined: ActionResult; undefined: ErrApprovalRequired` | GREEN 2026-06-27 |
| T-18b | internal/flow/action_test.go | `undefined: ActionResult` | GREEN 2026-06-27 |
| T-19 | internal/cli/flow_cli_test.go | deferred — CLI test for --token flag | deferred |

(Red evidence column will be populated with actual compiler/test output as tests are written before implementation.)
