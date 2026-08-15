# Inline code review — #4093

Reviewed the final diff against the phase scope and the closed transport
contract.

- Loader validation is fail-closed at the schema, strict-decode, version, and
  domain-descriptor layers; it has no registration effect.
- Composition verifies all discovered definitions before it builds adapters or
  mutates a registry; the batch registration path validates the entire batch
  while locked before its first map write.
- Factory selection is exact family-plus-ID matching, with no connector-name
  or capability fallback. Connector-local factory providers keep native
  PostgreSQL wiring out of App; provider adapters remain typed
  GitHub/PostgreSQL implementations.
- The production bundle embed inventory includes `sync_transport.json`; the
  binary lifecycle test and PostgreSQL definition tests prove the declaration
  does not disappear between source-tree validation and runtime composition.
- The CLI inspection projection, help/manual, generated connector docs, and
  website guide consistently expose GitHub's declared roles without implying
  that a declaration is certified or executable before preflight.
- The production test observes PostgreSQL and both GitHub roles through
  `Preflight`; refusal tests assert zero construction, registry, read, plan,
  and apply state changes.
- The local warehouse primitive now owns an additive destination declaration
  and exact native-database executor reference. Its app-composed adapter
  writes only beneath the connection-owned location, fsyncs and materializes
  the reopened workset before acknowledging, and hashes/row-counts the durable
  table again before a checkpoint can advance. A post-acknowledgement table
  mutation is observed and refused by read-back.
- `change_capture` is admitted only with `change_apply`; the local warehouse
  applies validated delete tombstones through keyed materialization. It does
  not advertise dedupe-history semantics it does not implement.
- The generic route predicate now requires both endpoint declarations. This is
  not a warehouse exception: a newly declared destination cannot turn any
  legacy source into a half-transport, while a malformed two-sided pair still
  reaches fail-closed preflight.
- No changes address #4125, #4136, #4090, or #4154.

`scripts/gsd sources verify-work` and `code-review`, plus `scripts/gsd doctor`,
were run. The official `code-review` command requires a numeric GSD phase, but
this issue uses the non-numeric phase directory above; the compatible inline
manual review fallback is recorded here. Result: no unresolved actionable
findings. Automated review is not requested: the task's direct-PR brief makes
CI the review gate.
