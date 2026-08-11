# Verification — Issue #4031

Status: in progress.

## Required evidence

- [x] Correction owner #4031 created and attached to parent #4015 before any project edit.
- [x] Child branch based on `origin/docs/4015-connector-release-certification`.
- [x] Parent ancestry confirmed: the parent branch is one commit ahead of `origin/main` and is an ancestor of this child.
- [x] Source proof inspected in `internal/connectors/native/postgres/reader.go` and #3855.
- [x] Red documentation assertion captured: focused PostgreSQL test exited 1 because the positional composite predicate was absent.
- [x] Green documentation assertion captured: focused PostgreSQL test passed with `$1`/`$2` predicate.
- [x] Focused PostgreSQL package test, vet, tidy check, lint, docs validation, and hygiene complete.
- [x] Inline/manual GSD verify and code review complete; no findings.
- [ ] Final no-mistakes run complete with `--skip=push,pr,ci` and no `--yes`.
- [ ] Child branch pushed and draft PR opened against `docs/4015-connector-release-certification`.

## Safety confirmation

No credentials, provider calls, runtime-backed database checks, SQL writes,
reverse ETL, certification execution, generated broad rewrites, or #3855
branch changes are in scope.

## Pipeline-scope recovery

The first no-mistakes run (`01KZQ0E0C56MYBS3TKC2VYRTAT`) passed its review,
test, documentation, and lint steps with push/PR/CI skipped. Its documentation
agent nevertheless created `524542d2` with an unrelated Parquet-wording edit.
That edit was outside #4031's scope and was immediately restored with
`3d5188784` after guarded custody recovery. A final no-mistakes run must now
validate the scoped net branch; the unrelated wording awaits a separate
issue-first owner.
