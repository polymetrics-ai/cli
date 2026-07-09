# Summary — Issue #80 Linear CLI parity parent

Status: implementation verified locally; parent PR remains draft/human-gated.

Completed:

- Read required repo, GSD, parent-orchestration, review-routing, CLI parity, connector migration, and Go skill references.
- Validated GSD/Pi adapter with `scripts/gsd doctor`, `scripts/gsd verify-pi`, and `scripts/gsd list --json`.
- Generated the parent planning prompt with `scripts/gsd prompt plan-phase issue-80-linear-cli-parity --skip-research`.
- Recorded manual-GSD fallback because `scripts/gsd prompt programming-loop ...` is not available in this checkout.
- Completed #97 CLI surface metadata and opened draft parent PR #131.
- Completed local critical-path slices #98-#103:
  - connector command-surface help for `pm help linear`, bare `pm linear`, and `pm linear --help`;
  - fixed GraphQL Linear streams for list/view reads plus generated fixed-document streams for every SDK query row in the ledger;
  - stream-backed direct-read runner support;
  - fixed GraphQL write actions for `create_issue`, `update_issue`, `comment_issue`, `create_project`, plus generated typed reverse-ETL actions for every non-deprecated live Linear mutation row (369 write actions total, including safe legacy SDK deprecated rows);
  - Linear operation ledger v1 with all 514 official non-deprecated fields covered, 530 covered fixed-document rows overall, and 2 blocked rows (raw arbitrary GraphQL plus deprecated `integrationSettingsUpdate`);
  - docs and website generated data updates.
- Local verification passed: `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `go run ./cmd/connectorgen validate internal/connectors/defs --json`, `./pm docs validate --connectors-dir docs/connectors`, `make verify`, and `git diff --check`.

Safety:

- No credentialed Linear checks were run.
- No live Linear writes were executed.
- No secrets were requested, printed, or stored.
- Raw arbitrary GraphQL remains disallowed.

Next:

1. Commit and push the verified Linear parity slice on `feat/80-linear-cli-parity`.
2. Update PR #131 body with verification and issue mapping.
3. Keep parent PR draft or mark ready only after applying the project review-routing policy.
