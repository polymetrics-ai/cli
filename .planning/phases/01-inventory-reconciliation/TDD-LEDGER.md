# TDD Ledger — Phase 1 Inventory and Surface Reconciliation

## Classification

Planning-only issue. No production behavior changes and no Go source edits are allowed.

## Red / Initial Evidence

| Evidence | Command / Source | Result |
|---|---|---|
| Legacy planning existed | `find .planning/phases -maxdepth 1 -mindepth 1 -type d | wc -l` before replacement | 26 phase directories |
| Canonical codebase maps absent | `test -d .planning/codebase || echo NO_CODEBASE_MAP` before replacement | `NO_CODEBASE_MAP` |
| GSD preflight recognized brownfield planning state | `node ~/.claude/get-shit-done/bin/gsd-tools.cjs init new-project` | `project_exists: true`, `is_brownfield: true`, `needs_codebase_map: true` |
| No Go source edits at baseline | `git diff --name-only -- cmd internal` | no output |

## Green Evidence Targets

| Target | Verification |
|---|---|
| Config parses | `node -e "JSON.parse(...)"` |
| Required active planning files exist | `test -f .planning/{PROJECT,REQUIREMENTS,ROADMAP,STATE}.md` |
| Codebase maps exist | `test -d .planning/codebase` and file checks |
| Multi-surface connector parity language present | `rg` for REST/GraphQL/XML/SOAP/binary/direct-read/CDC/queue/etc. |
| No Go source changed | `git diff --name-only -- cmd internal` returns no output |

## Notes

No red unit test is appropriate because issue #122 changes planning artifacts only. The red/green evidence is state validation and scope-guard verification.
