# Plan — Issue #182 Freshchat help renderer

Refs #182, parent #180.

## GSD / skills

- `scripts/gsd prompt plan-phase issue-182-freshchat-help-renderer --skip-research` generated successfully.
- `scripts/gsd prompt programming-loop init --phase issue-182-freshchat-help-renderer --dry-run` is unavailable (`unknown GSD command: programming-loop`); manual programming-loop fallback is active.
- Required skills used: gsd-core, golang-how-to, golang-cli, golang-testing, golang-error-handling, golang-security, golang-safety, golang-design-patterns, golang-structs-interfaces, golang-context, golang-concurrency, golang-documentation, golang-lint, frontend-design, web-design-guidelines, vercel-react-best-practices, vercel-composition-patterns.

## Scope

- Make Freshchat connector command help discoverable without credentials.
- Ensure `pm freshchat` and `pm freshchat --help` render contextual command surface help and exit successfully.
- Update CLI docs and website docs with Freshchat command-surface safety guidance.
- Regenerate Freshchat connector manual/skill artifacts as needed so command-surface metadata appears in docs.

## Non-goals

- No credentialed Freshchat API calls.
- No direct-read executor implementation (#185).
- No binary upload support (#186).
- No reverse ETL execution.
- No generic raw HTTP/shell/SQL escape hatches.

## Implementation slices

1. Red CLI tests for `pm freshchat` and `pm freshchat --help` expecting command-surface help output and exit code 0.
2. Runtime help: route connector namespace help through `connectors.RenderConnectorManual` when a connector exposes a command surface.
3. Docs parity: update `docs/cli/connectors.md`/website docs and generated Freshchat manual/skill artifacts.
4. Verification: run targeted CLI/help tests, docs validation, connectorgen validation, and full handoff gates if practical.

## Safety decisions

- Help rendering must not resolve credentials or read secrets.
- Blocked/planned Freshchat commands must remain visibly non-executable in help metadata.
- Reverse ETL commands must continue to advertise plan/preview/approval/execute and never imply direct mutation.
