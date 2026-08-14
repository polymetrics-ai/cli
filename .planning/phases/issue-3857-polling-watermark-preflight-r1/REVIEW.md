# Code review — #3857 polling-watermark preflight

## Method

Manual standard review under the documented inline fallback: this task's
single-worker contract forbids role spawning and no compatible isolated Pi
review runtime is available. Reviewed the complete diff, descriptor validation,
loader/schema path, preflight registry, immutable-corpus admission, eligibility
projection, tests, and authoring documentation after all focused and repository
gates passed.

## Findings

No critical, warning, or informational findings remain.

Reviewed safeguards:

- `polling_watermark.json` is optional, strict-decoded, structurally validated,
  and semantically validated; it is separate from `database.json` and
  `changefeed.json`.
- Runtime eligibility calls `PollingPreflight` for every declared mode, so no
  generator or inspection projection copies the admission decision tree.
- The registry checks exact native-database executor references under an
  `RWMutex`; a resolved declaration/catalog object is defensively copied.
- The preflight exposes neither a read nor a DML operation and calls no source
  executor method, preserving the pre-I/O boundary owned by this issue.
- All negative rows assert their full error string plus zero source/target
  observable counters; happy and empty-result rows assert distinct state
  transitions.
- No changefeed capability, commandrunner REST preflight, query taxonomy,
  generic protocol, credential, raw query, source executor, or target DML was
  introduced.
