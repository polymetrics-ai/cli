# Phase 437 Pending Intake Preservation

Status: preserved for triage; implementation not authorized.
Recorded: 2026-07-20T19:28:58Z
Source worktree: `/Users/karthiksivadas/Development/polymetrics-cli-agents/wt-437-connectors-certify-native-cobra`
Source PR: #466 (`refactor/437-connectors-certify-native-cobra`)
Source head: `26f98a72419010b961b5b8378ef4a695b0c0a06f`
Parent checkpoint at preservation: `c3d8a7573bfaf661bdcab737db84e3497929cdff`

## Rule

These files preserve untracked Phase 437 user-request/research intake without changing PR #466's tested head. Do not implement, regenerate, commit to PR #466, or edit the referenced GitHub issues until the human coordinator explicitly authorizes implementation or planning synchronization.

## Preserved files

- `PENDING-USER-REQUESTS.md` — mixed backlog intake; future owners include #437, #411, #417, and one connections-conflict follow-up needing issue assignment.
- `PENDING-HELP-MANUAL-TREE-PLAN.md` — #417 help/manual tree plan; dependency-blocked.
- `research/CONNECTOR-CATALOG-TABLE-RESEARCH.md` — #411 connector list/catalog/table/browser acceptance research.
- `research/CONNECTOR-INSPECT-EXPERIENCE.md` — #411/#412/#417 progressive connector inspection research.
- `research/CONNECTOR-INSPECT-PRIMARY-SOURCES.md` — primary-source evidence for connector inspection design.
- `research/HIERARCHICAL-HELP-MANUAL-ARCHITECTURE.md` — #417 command manual architecture research.
- `research/HIERARCHICAL-HELP-MANUAL-SOURCES.md` — primary-source evidence for #417.
- `debug/resolved/github-credential-missing-owner.md` — diagnose-only #437 credential-config trace; no production fix authorized.

## Triage

- #437 candidate follow-ups: connectors catalog required-value parsing, GitHub credential owner/repo validation, certify-specific help, duplicate connection conflict classification.
- #411 candidate follow-ups: readable connector list/catalog tables, connector browser, progressive connector inspect overview/focused selectors.
- #412 consumer: full connector/manual pager over the canonical presentation model.
- #417 candidate follow-ups: individual command manuals, hierarchical help tree, generated man/docs/website parity, aggregate connectors help example cleanup.

## Safety

- No secrets or credential values are preserved here.
- No live connector checks, PAT use, Keychain reads, reverse ETL execution, external writes, sweeps, dependencies, or GitHub issue edits are authorized by this preservation step.
