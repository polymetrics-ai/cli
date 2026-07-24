# Summary — CLI Architecture v2 Cobra/Viper release split

Status: reconstructed; verification pending.

The approved five source squashes are reconstructed on latest `main`. The single `internal/cli/cli.go` conflict retained current Gong behavior, Cobra was adapted to the current JSON-help signature, and the two affected golden cases were regenerated. Focused Cobra/golden/config tests pass.

The candidate explicitly excludes TUI, events, logging, OpenTelemetry, PR #493 routing work, and PM review-system implementation. ADR 0002 and an additive release-split state record describe the exact boundary and parent consequences.

Final exact head/tree, full compatibility evidence, dependency delta, review/security/Snyk/no-mistakes results, PR URL, RC version decision, and release publication status remain pending.
