# Code review — issue 4325 declaration-admission foundation

## Local review

- Reviewed the declaration schema, deterministic finder, command-surface
  projection, and commandrunner dispatch boundary against the issue contract.
- Confirmed the check performs local filesystem reads only. It neither fetches
  a provider nor accepts credentials, and it does not execute writes/deletes.
- Confirmed a deferred command remains discoverable and is refused before any
  executor only after resolving one exact non-excluded, non-policy target,
  while an `implemented` declaration still requires the existing lane-specific
  binding and real no-I/O preflight.
- Confirmed the two required catalog schemas intentionally contain no source bytes, hash, raw
  body, or typed request schema requirement.
- Confirmed the admission URL is checked by the same public HTTPS,
  no-userinfo/no-fragment, bounded secret-safe query policy as source import.
- Confirmed no connector-owned Batch-1 definition, concurrent Docker Hub
  mapping, generated connector evidence, or live certification record is
  staged for this repair.

## Finding disposition

Two implementation findings were corrected before this review record:

1. The typed failure reference cannot contain the human-readable
   space-separated command path. The dispatch now uses the normalized command
   identity for that reference, and
   `TestPreflightDeferredCommandReturnsNamedFoundationAfterExactTargetValidation`
   verifies the `system/missing_foundation` classification.
2. An implemented write could otherwise name an existing action whose endpoint
   differed from the cited source operation. The admission binding now compares
   method and canonicalized template path as well as action identity; the red
   `false implemented write binding` subtest is now green.

No unresolved local findings remain.

The post-`3d39cc1fc` independent audit findings were also dispositioned:

1. **Accepted — deferred resolver bypass.** The old path returned typed
   `missing_foundation` before resolving its API-surface target. Deferred
   preflight now fails closed unless the connector proves one exact blocked
   operation and the declared foundation absence. Excluded, `disallowed`,
   duplicate, mismatched, and stale targets fail before typed classification.
2. **Accepted — weaker citation policy.** Declaration admission now reuses
   `validateSourceImportPublishedURL`; HTTP, userinfo, fragments, private
   literals, and credential-shaped query keys are rejected without DNS or
   provider I/O.
3. **Accepted — semantic destructive applicability.** Metadata is required for
   every `DELETE` and every other exact target classified
   `destructive_action`; a non-destructive POST cannot be relabelled delete.
4. **Accepted — false implemented surface mapping.** A valid runtime binding
   no longer overrides an excluded, policy-only, or duplicated canonical
   surface row.

Focused red tests failed on each old behavior before the shared fixes and the
complete admission, commandrunner, and engine focused suites are green.

Captain clarification was applied after the initial PR: the regression suite
now loads the actual GitHub bundle and admits/preflights its implemented `label
delete` action. Documentation states that only an endpoint's named missing
foundation can defer it; delete and destructive operations are not generically
deferred.

Captain clarification 007 supersedes a subsequent Stripe conversion request.
This PR now reviews only generic admission semantics: a deferred catalog row
must name a bounded missing implementation component plus evidence, and a
deferred command may cite its blocked endpoint without claiming executable
coverage. The started Stripe CLI/source declaration files remain deliberately
unstaged for `cli-batch1-repair-r1`; no connector-owned mapping change appears
in this PR's committed diff.

## Exact-SHA Codex audit disposition (2026-08-26)

1. **DA-001 accepted:** optional sidecars allowed a successful zero-work gate.
   Two required root catalogs, nonzero expected counts, and exact count checks
   now fail missing catalogs and omitted rows.
2. **DA-002 accepted:** method/path and a provider-native operation ID were not
   sufficient identities. Source and declaration rows now carry an exact
   binding identity; runtime must match it, same-endpoint swaps fail, and an
   empty provider-native ID remains valid.
3. **DA-003 accepted:** admission copied an incomplete runtime switch. The
   engine now owns the shared implemented resolver used by runtime preflight
   and admission, covering templated ETL, operation-free direct read, REST,
   binary, and GraphQL operations.
4. **DA-004 accepted:** deferred staleness was endpoint-first and missed
   GraphQL identity. The identity-aware resolver rejects both an existing exact
   GraphQL binding and another operation occupying its shared transport.
5. **DA-005 accepted:** destructive applicability was self-authored through
   the surface model. The independent source cohort owns `none`, `delete`, or
   `destructive`; declaration and implemented runtime semantics must match.
6. **DA-006 accepted:** typed missing-foundation could be lost through large
   metadata, App credential ordering, and the CLI wrapper. Text/references are
   bounded deterministically, App preflights first, and the public error keeps
   code `missing_foundation`.
7. **DA-010 accepted:** production omits `api_surface.json`. The compact source
   cohort is embedded into `defs.FS` and loaded per connector; a production-
   layout CLI test proves a deferred admitted command returns typed
   missing-foundation with `Surface == nil`.
8. **DA-011 accepted:** source and deferred targets now reject noncanonical
   methods, absolute URLs, queries/fragments, controls, traversal, backslashes,
   and repeated separators. Legacy implemented provider-specific identities are
   not silently reclassified by this new admission-only constraint.

## Automated review route

The direct PR targets `main` and is already open. Its original route was
`claude_auto`; this audit repair creates a new unreviewed commit, so one
deliberate fresh review request is required after push under the repository's
review routing policy. Firstmate also requested an independent-Codex re-audit
handoff on the clean commit. Any actionable response requires a recorded
disposition before merge; this worker will not merge.
