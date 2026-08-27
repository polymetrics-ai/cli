# Cited-only mutation-reference closure — issue 4371

## Task Delivery Header

- Issue: Closes #4371 — fix(connectorgen): keep cited-only mutation references closed; Refs #4291 — Batch 6–7 connector delivery.
- Base branch: `main` (`origin/main@cf29d302c13f7fcd340d31ad6dc27872880ccf42` at task start).
- Merges into: `main`.
- Delivery: Normal direct pull request open against `main`, committed and normally pushed from the working branch after current-main integration, local gates, exact-head audit request, and GitHub API base read-back. Human/captain controls merge.
- Working branch: `fm/cli-cited-only-mutation-reference-foundation-r1`.
- Task: Repair the shared source-import/projector disposition boundary so cited-only mutations remain the exact closed `source_contract_unavailable` reference projection, while preserving ordinary contract-complete mutation dispositions and every provider operation's visible source accounting.
- Verification: Red/green/refactor importer and projector tests; Salesloft/Copper cohort evidence where retained locks are available; source import/projection/evidence/validate/surface-sync checks; real commandrunner preflight for any existing unavailable command; formatter, scoped package tests, vet, builds, docs/generator gates, diff check, rebase/current-main rerun, GSD review, and independent exact-head audit.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Cited-only non-executable disposition remains closed | live | A source-reference mutation with a non-executable disposition either fails before descriptor output or yields byte-identical closed reference projection; test asserts exact gaps, provenance, and no output on refusal. |
| Cited-only partial-coverage disposition remains closed | live | Equivalent partial-disposition regression asserts an actionable source ID and no written descriptor. |
| Normal OpenAPI/Swagger dispositions remain stable | live | Existing complete-contract mutation fixtures apply dispositions and descriptor bytes before/after the guard remain unchanged. |
| Disposition admission remains fail-closed | live | Table cases cover absent, duplicate, unknown, method/path mismatch, and mutating/non-mutating citation misuse. |
| Batch 6–7 source accounting remains truthful | live | Salesloft and Copper source lock/projection/evidence checks account for each cited operation exactly once without an executable claim, when the preserved locks are available in the worktree; otherwise explicit minimal fixtures prove only the shared guard and do not claim cohort regeneration. |
| Any pre-existing unavailable command stops before credentials/I/O | live | Registry/commandrunner test asserts structured `missing_foundation` precedes credential lookup, provider transport, record emission, and mutation. No new command is created by this issue. |
| Usable command surface is not fabricated | live | Diff/registry inspection and commandrunner tests show no newly materialized command or generic provider write path; expected usable-surface delta is zero. |

## TDD plan

1. **Red:** add focused cited-only mutation fixtures for both disposition types
   through the actual source-import CLI path. Record the pre-fix rejection by
   strict descriptor validation after output construction, plus output-byte/no-
   write evidence where applicable. Add normal OpenAPI/Swagger controls and
   invalid citation matrix cases.
2. **Green:** implement one narrow cited-only compatibility guard in each
   disposition application path before mutation/output. Include the exact
   source ID/location in the error and preserve all existing contract-complete
   behavior.
3. **Refactor:** share the closed-reference predicate if it makes both paths
   clearer without widening accepted inputs. Keep defensive copies/sorted gaps
   and strict validator behavior intact.
4. **Cohort reconciliation:** use retained Salesloft/Copper source locks if
   available; regenerate only their canonical derived descriptors/evidence.
   Do not reconstruct, fetch, or re-pin provider material. If unavailable in
   this clean task worktree, record that exact limitation and run tracked
   cross-cohort validation plus the irreducible explicit fixture evidence.
5. **Review and publication:** run `verify-work` and `code-review` inline,
   freeze/audit the final SHA, rebase normally onto newest `origin/main`, rerun
   scoped coverage, commit, push normally, open the direct PR, and API-read
   its base.

## GSD and parity notes

- `scripts/gsd doctor`, `scripts/gsd sources …`, and prompts for all required
  lifecycle stages have been resolved. Inline/manual execution is required by
  the repository's single-worker contract.
- This alters developer generator behavior but does not alter `pm` user-facing
  command syntax/help/manual/website content. Recheck `connectorgen
  source-import --help` and docs only if observed output changes; otherwise
  record CLI/docs/website parity as not applicable.
- Security boundary: no source fetch, secrets, credential lookup, provider
  transport, or reverse-ETL execution is authorized for this repair.
