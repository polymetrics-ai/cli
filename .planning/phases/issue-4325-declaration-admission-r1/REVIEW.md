# Code review — issue 4325 declaration-admission foundation

## Local review

- Reviewed the declaration schema, deterministic finder, command-surface
  projection, and commandrunner dispatch boundary against the issue contract.
- Confirmed the check performs local filesystem reads only. It neither fetches
  a provider nor accepts credentials, and it does not execute writes/deletes.
- Confirmed a deferred command remains discoverable and is refused before any
  executor, while an `implemented` declaration still requires the existing
  lane-specific binding.
- Confirmed sidecar schema intentionally contains no source bytes, hash, raw
  body, or typed request schema requirement.
- Confirmed no connector-owned Batch-1 definition, generated evidence, or
  live certification record is changed.

## Finding disposition

Two implementation findings were corrected before this review record:

1. The typed failure reference cannot contain the human-readable
   space-separated command path. The dispatch now uses the normalized command
   identity for that reference, and
   `TestPreflightDeferredCommandReturnsNamedFoundationBeforeExecutor` verifies
   the `system/missing_foundation` classification.
2. An implemented write could otherwise name an existing action whose endpoint
   differed from the cited source operation. The admission binding now compares
   method and canonicalized template path as well as action identity; the red
   `false implemented write binding` subtest is now green.

No unresolved local findings remain.

Captain clarification was applied after the initial PR: the regression suite
now loads the actual GitHub bundle and admits/preflights its implemented `label
delete` action. Documentation states that only an endpoint's named missing
foundation can defer it; delete and destructive operations are not generically
deferred.

Captain clarification 007 supersedes a subsequent Stripe conversion request.
This PR now reviews only generic admission semantics: a deferred sidecar row
must name a bounded missing implementation component plus evidence, and a
deferred command may cite its blocked endpoint without claiming executable
coverage. The started Stripe CLI/source declaration files remain deliberately
unstaged for `cli-batch1-repair-r1`; no connector-owned mapping change appears
in this PR's committed diff.

## Automated review route

The direct PR targets `main`. Its route is `claude_auto`: opening a non-draft
PR by the repository author triggers the configured Claude review. No manual
Claude or Copilot request is made before that trigger. Review coverage is
therefore pending PR creation; any actionable response will require a recorded
disposition before merge.
