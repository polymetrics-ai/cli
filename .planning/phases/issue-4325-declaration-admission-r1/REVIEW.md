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
- Confirmed authoring admission and the compact production ledger use the same
  provider-citation canonicalizer: public HTTPS, unambiguous lowercase DNS,
  no default `:443`, normalized path escaping, and bounded stable secret-safe
  query encoding. Persisted evidence is compared and rejected, never rewritten.
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
9. **Fleet compatibility red accepted:** the first shared resolver compared
   runtime-relative paths literally with provider-facing command endpoints,
   rejecting 243 valid base-path and placeholder aliases. Binding existence
   and identity remain exact, while the command projection retains the cited
   provider endpoint. The complete implemented-preflight and CLI suites pass.

## Independent R2 audit disposition (2026-08-26)

1. **DA-002 accepted — provenance-only uniqueness.** Runtime binding fields
   were incorrectly part of the source-operation duplicate key, so changing a
   binding could disguise one repeated provider row. Source URL, exact document
   location, protocol, raw operation ID, method, and canonical provider
   operation/endpoint now form the provenance key. Binding uniqueness is
   checked separately in both authoring admission and the compact runtime
   ledger. Regressions cover one duplicate provenance row under different
   bindings and one binding claimed by distinct source rows.
2. **DA-003 accepted — fail-closed endpoint equivalence.** The resolver no
   longer substitutes the command endpoint unconditionally. It keeps canonical
   and transport endpoints separately and accepts only named proofs for exact,
   declared base path, positional placeholders, registered hook transport,
   GraphQL operation-to-`POST /graphql`, absolute URL/query normalization,
   provider suffix, and closed operation annotation transformations. Negative
   stream, write, REST, and binary aliases fail. The clean real-bundle census
   proves all 243 non-GraphQL aliases plus all 4 GraphQL aliases; Notion's hook
   and GitHub GraphQL have explicit positive coverage.
3. **DA-004 accepted — real stale-deferral preflight.** Commandrunner clones
   the exact deferred row to implemented form and invokes the same implemented
   preflight helper used by normal dispatch before considering a missing
   foundation. Runnable GitHub label delete and GraphQL read/write controls are
   rejected as stale for every valid component. Components are executor-
   specific: response descriptors require REST/binary operations, source
   importers require unbound direct operations, and idempotency policy is not a
   missing executor foundation.
4. **DA-012 accepted — exact production inventory.** The one compact
   `declaration_admission_sources.json` root artifact is classified exactly as
   `runtime_declaration_target_ledger`, with deterministic byte attribution.
   Full API surfaces and build-time declaration/source catalogs remain outside
   the runtime embed inventory.
5. **Captain Outreach compatibility seam accepted; integration claim corrected
   by R3.** The test loads the real Outreach bundle and synthesizes its absent
   CLI discovery projection in memory. Its existing stream/write shapes pass
   generic admission, canonical/transport resolution, and no-I/O commandrunner
   preflight. This is not shipped CLI, source-evidence, credential-boundary, or
   zero-transport proof.
6. **Local boundary finding accepted and fixed.** The provider-policy scanner
   matched `cal-com` inside the neutral local name `canonicalComparable`. The
   variable is now `declaredComparable`; no boundary exception was added, and
   the exact-head whole-tree scan reports 0 findings.

No unresolved local or R2 findings remained on code SHA `f97dede07`; the R3
audit below supersedes that exact-head conclusion. The
certificate proves complete source-cited six-lane declaration independently
of runnable count: an unavailable operation must remain present and deferred
with its named exact foundation, while stale or falsely implemented rows fail.

## Independent R3 audit disposition (2026-08-26)

1. **DA-002 accepted — raw citation spelling bypassed provenance uniqueness.**
   One shared `internal/safety` canonicalizer now serves connectorgen authoring
   and the embedded declaration-target ledger. It rejects unsafe or ambiguous
   authorities, returns lowercase DNS/no-default-port/normalized-path/stable-
   query identity, and both consumers require the authored string to match it.
   Binding uniqueness remains a separate invariant. Focused regressions cover
   host case, explicit `:443`, query order, path escaping, trailing-dot/empty
   DNS labels, and the same provider operation under another binding.
2. **Outreach generic-engine defect claim declined; evidence overclaim
   accepted and corrected.** The committed Outreach bundle has no discovery
   surface, so this foundation cannot claim its commands reach credentials.
   The renamed test truthfully proves only resolver compatibility over real
   bundle shapes with a synthetic in-memory projection. No connector-owned
   definition is added here. Final merge validation requires a real combined-
   head Outreach mapping/pilot after #4350 repair with committed commands,
   source evidence, credential-boundary and zero-transport proof, then a fresh
   audit.

No provider I/O, credential access, write/delete execution, schema-version
migration, or Batch-1 connector definition change is part of the R3 repair.
Exact commit `1d3ac8d273235664c92a84c170d1946ce56a3339` passes the
full connectorgen, engine, commandrunner, connectors, definitions, safety, app,
and CLI packages plus vet/build/lint/docs/smoke/generator/boundary/canon/release
gates. No new local review finding remains before the fresh independent audit.
The final inline review additionally rejected plain and percent-escaped path
dot segments; focused safety/authoring/ledger tests and scoped lint pass after
that hardening.

## Automated review route

The direct PR targets `main` and is already open. Its original route was
`claude_auto`; this R3 repair creates a new unreviewed commit, so one deliberate
fresh independent exact-head audit is requested after push under Firstmate's
route. Any actionable response requires a recorded disposition before merge;
this worker will not merge.

## Independent R5 audit disposition (2026-08-26)

1. **Reviewed-source binding accepted.** The independent inventory selects an
   exact operation in a connector-owned reviewed source lock. Admission checks
   protocol, source URL, document location, provider operation ID, method, and
   provider path without fetching or rehashing retained bytes. Unrelated hosts,
   nonexistent operations, semantic aliases, symlink escapes, and mutable count
   escape hatches fail closed.
2. **Provider-evidenced unsupported accepted.** One explicit terminal state is
   available in all six lanes. It remains discoverable and counted but carries
   neither an executor selector nor a missing-foundation claim; commandrunner
   returns typed `system/provider_evidenced_unsupported`.
3. **Public-input ordering accepted.** Help remains first. Shared flag
   validation precedes App and credentials for required, unknown, enum, bound,
   and env-only inputs. An initial full-suite finding that unknown paths became
   validation errors was fixed narrowly; existing usage/suggestion controls and
   the complete CLI suite pass.
4. **Independent denominator accepted.** The compact production ledger and
   authoring admission use the exact inventory selection; adjacent mutable
   counts were removed, and omitting source and declaration rows together fails.

No unresolved local R5 finding remains. Clean exact-head package,
generator/snapshot, boundary, runtime-preflight, canon, lint, build, docs,
smoke, and release checks pass. The next gate is a fresh independent exact-SHA
audit and PR CI; this worker does not merge.
