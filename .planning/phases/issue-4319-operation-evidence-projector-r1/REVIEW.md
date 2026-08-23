# Issue #4319 — code review

## Result

No unresolved findings.

## Review dispositions

- GraphQL evidence originally risked matching every operation that shared
  `POST /graphql`; it now matches the root field from the operation's fixed
  GraphQL document. A real `createIpAllowListEntry` acronym case guards it.
- Identical repeated source operations now deduplicate deterministically;
  conflicting evidence for the same source identity fails rather than being
  silently chosen.
- Generated artifacts are guarded by the Make target, and the fixed cohort is
  capability-stratified so GraphQL writes cannot crowd out the other executable
  classes.
- Current-main review: v3 source locks are consumed through the existing
  read-only adapter. Revalidation against `cf493b834` passed without changing
  either projector artifact; no source-lock parser or schema file was edited.
