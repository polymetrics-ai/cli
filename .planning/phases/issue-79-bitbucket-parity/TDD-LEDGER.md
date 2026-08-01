# TDD ledger - issue #79 Bitbucket parity

## GSD command path

- GSD adapter preflight: `scripts/gsd doctor` passed in this worktree.
- Required command attempted for this CI repair: `scripts/gsd prompt programming-loop init --phase issue-79-bitbucket-parity --dry-run`.
- Adapter result: `scripts/gsd: unknown GSD command: programming-loop`.
- Manual-GSD fallback is active for this CI repair because the branch-local GSD adapter does not expose `programming-loop`.

## Required skills used

- `.agents/agentic-delivery/references/required-skills-routing.md`
- `golang-how-to`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-lint`
- `gsd-programming-loop`
- `no-mistakes` context only; no no-mistakes control commands were invoked because this is the assigned CI phase inside an active outer run.
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`

## Red evidence

| Check | Result |
| --- | --- |
| `scripts/verify-gsd-workflow dc753087d3ec7cbb7d317869a26550db5de25cd2` | Failed: `cmd/internal changed, but no GSD planning evidence changed`. |
| `make verify` | Failed in `go test -timeout 20m ./...`: `TestNewLoadsDeclarativeBundlesWithHooksAndNativeOverrides` loaded 550 bundles, but the stale assertion expected 549. |

## Green plan

1. Keep the Bitbucket connector implementation unchanged.
2. Update only the stale bundle-count assertion to include the added Bitbucket bundle.
3. Add this `.planning` TDD ledger so the workflow evidence gate can see the existing GSD/manual-GSD evidence.
4. Re-run the focused failing package, the GSD workflow evidence script, and the relevant verify target.

## Green evidence

| Check | Result |
| --- | --- |
| `go test ./internal/connectors/bundleregistry -count=1` | Passed after updating the expected bundle count from 549 to 550. |
| `make verify` | Passed: fmt, tidy-check, vet, full `go test ./...`, build, docs validation, smoke, lint, connectorgen validate, connector-boundary, and release workflow check. |
| `git diff --check` | Passed. |

The official `scripts/verify-gsd-workflow` script diffs `merge-base...HEAD`, so it cannot pass against uncommitted repair files in this CI phase. Once this phase is committed by the outer executor, this changed `.planning/phases/issue-79-bitbucket-parity/TDD-LEDGER.md` file satisfies the gate's accepted path pattern and includes GSD/TDD evidence.

## Safety notes

- No live Bitbucket or credentialed connector checks are part of this CI repair.
- No dependencies, generic write tools, generic HTTP write tools, or generic SQL write tools are introduced.
- DELETE/destructive Bitbucket actions remain represented only as typed reverse-ETL actions with `confirm: "destructive"` and the existing plan -> preview -> explicit approval -> execute path.
