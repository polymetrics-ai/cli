# Review — #3859 native database apply strategies

Manual `code-review` fallback: the canonical generated reviewer expects a
numeric phase and an isolated runtime agent, neither applies to this
issue-named direct-PR phase. The worker performed the standard cross-file
review required by the generated prompt.

Reviewed files: the sealed polling apply bridge, database plan/input boundary,
PostgreSQL managed-target DDL/assertions/value helpers/session implementation,
definition change, and live assertions.

## Findings

No Critical, Warning, or Info findings remain.

Review checks confirmed that:

- the bridge has no arbitrary SQL, relation, or per-record write surface;
- page bounds and cancellation are checked before preview/session mutation;
- only explicit typed tombstones invoke non-history physical deletion;
- history mode closes validity windows in the same transaction as its source
  order-fence update; and
- all dynamic identifiers are definition/control-derived and quoted, while
  values remain PostgreSQL parameters.

The local regression suite, race tests, explicit live dbtest, lint, vet,
build, and individual repository gates are recorded in `VERIFICATION.md`.
