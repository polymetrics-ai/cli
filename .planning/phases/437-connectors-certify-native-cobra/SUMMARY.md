# Phase 437 Summary

Status: third accepted safety/correctness correction in progress; verification pending.

## Identity

- Session: `issue-437-third-safety-correction-20260719`
- Profile: Sol/high
- Branch: `refactor/437-connectors-certify-native-cobra`
- Original exact start/base: `6c038bb4ab4a5497fca28a0cab42d0a7fa4eb22b` / `feat/cli-architecture-v2`
- Third-correction exact start: `437d13cf`
- Parent #397, umbrella #407, draft parent PR #438

## Delivered before this correction

`connectors` is a native Cobra subtree with `list`, `catalog`, `inspect`, `man`, `docs`, hidden positional help, and nested `certify`. Single, batch, and sweep certification use invocation-local runtime seams while preserving in-process execution, report rendering, telemetry, cancellation, bounded workers, events, and exit mapping. Canonical help, generated CLI docs, golden transcripts, and website data describe the bounded certification surface.

Two prior review corrections were completed, verified, committed, and pushed. They made unsupported controls fail closed, restored execution ordering and help behavior, made batch write-disable controls dominate credential-file writes, gated writes on sandbox, rejected unsupported credential-file limits, constrained skip values, and corrected stale docs. Their recorded full gates passed at their respective heads.

## Third accepted correction

All findings in `/tmp/pm-397-rereview2-437.log` are accepted:

- certify subtrees must reject every unknown flag before credential loading, runners, sweep, or effects, including write-like typos;
- sweep age must be strictly positive and reasonably bounded;
- ordinary completed prior reports must be reusable with `--resume`, without fabricated future timestamps, while incomplete reports rerun;
- credential-file `exec` must remain prohibited and reject before effects; generic external execution code and claims must be removed;
- usage exit docs, release-stage token examples (`ga`), flag/docs audit, and terminal planning state must be accurate.

The plan, TDD ledger, verification checklist, prompt snapshot, summary, and run state are reopened before RED tests or production changes. Full verification is intentionally marked pending until `make verify` and the declared phase gates actually pass.

## GSD / skills / execution decision

`scripts/gsd doctor` and `scripts/gsd list` passed. `scripts/gsd prompt programming-loop ...` is absent from the adapter registry, so the manual universal runtime loop is the recorded fallback. Execution decision is `local_critical_path`: one bounded correction in the existing isolated issue worktree, no subagent tool, and no credentials, external commands, services, dependencies, PR, or review.

Loaded skills: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-concurrency`, `golang-context`, `golang-code-style`, `golang-naming`, `golang-documentation`, `golang-spf13-cobra`, and `golang-lint`.

## Safety

Implementation and verification are fixture/temp/in-process only. No credential values, credentialed connector checks, external credential commands, live services, external writes or sweeps, new dependencies, generic tools, destructive/admin actions, production changes, PR, or review are permitted.
