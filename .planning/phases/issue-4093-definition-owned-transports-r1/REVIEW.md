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
- No changes address #4125, #4136, #4090, or #4154.

Result: no unresolved actionable findings. Automated review is not requested:
the task's direct-PR brief makes CI the review gate.
