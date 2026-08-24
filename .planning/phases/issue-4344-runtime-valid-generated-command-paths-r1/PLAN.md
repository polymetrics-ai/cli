# Plan — #4344 runtime-valid generated command paths

## Task Delivery Header

- Issue: Refs #4344 — fix(connectorgen): derive runtime-valid generated command paths
- Base branch: `main`
- Merges into: `main`
- Delivery: A committed and pushed `fm/cli-runtime-valid-generated-command-paths-r1` branch with a PR open against `main`, required verification recorded, and the PR base read back from the GitHub API.
- Working branch: `fm/cli-runtime-valid-generated-command-paths-r1`
- Task: Replace raw-source-ID CLI command identity with an operation-derived,
  parser-valid, injective identity; preserve path parameter bindings; migrate
  only previously invalid generated paths; regenerate checked-in dependent
  surfaces where present; prove credential-free reachability.
- Verification: focused source-projection and commandrunner tests; targeted
  binary/credential-bound reachability test; `connectorgen validate`,
  `surface-sync --check`, operation evidence check, connector runtime
  preflight, docs validation, build/vet/lint and the applicable `make verify`
  entry points, all with `GOFLAGS='-p=3'` and one heavy suite at a time.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| New generated paths are parser-valid and operation-derived | live | A projected parameterized mutation is accepted by the real commandrunner preflight; deliberately restoring the raw ID makes the behavioral test fail on retained `{parameter}` text. |
| Operation identity is collision-free and stable | live | Operations differing only in a parameter name or literal/path shape yield distinct valid command paths, and a second projection has no changes. |
| Path bindings are retained | live | The projected command retains the required `record.<field>` path flags, and the real commandrunner builds the expected bound record from those flags. |
| Existing reachable surface is not renamed | live | A valid legacy generated path remains byte-for-byte unchanged after projection; an invalid legacy generated path is changed to the new valid identity. |
| Bitbucket generated surface is reachable | live | Once the source-import descriptor is present on the reviewed batch-1 input branch, a built binary in an isolated project reaches `missing --credential` for all 50 implemented Bitbucket commands, including the 28 formerly rejected paths. This branch must not copy unrelated connector work merely to manufacture the fixture. |

## TDD execution slices

1. **Red:** Add behavioral source-projection coverage for a raw `{parameter}`
   source ID/path that reaches real command preflight, plus collision,
   idempotence, legacy-valid preservation, and legacy-invalid migration cases.
   Confirm it fails with the raw source-ID implementation.
2. **Green:** Add the smallest injective operation identity encoder and narrow
   legacy migration predicate. Keep path-field flags and `maps_to` untouched.
3. **Regenerate:** Run `surface-sync` and operation-evidence generation. Record
   exactly which checked-in surfaces changed; do not absorb unrelated batch-1
   declaration work.
4. **Verify:** Build `pm`; run a credential-free generated-command sweep in an
   isolated project. The complete Bitbucket 50-command sweep is executed
   against its reviewed source-import artifact rather than hand-authored
   commands. Verify CLI docs/help parity for any changed checked-in command
   name; otherwise record why it is not applicable.
5. **Review:** Execute the inline verify-work and code-review prompts, record
   findings/dispositions, commit, push, open the PR, and read back its API base.

## Required skills

`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`,
`golang-security`, `golang-safety`, `golang-design-patterns`,
`golang-structs-interfaces`, and `golang-naming`.

## CLI help/manual/website parity

The generator's output can rename a user CLI surface. For current valid
generated paths, migration policy preserves the existing name, so no help/doc
change is expected. Any generated invalid path has never been executable; when
the reviewed Bitbucket/GitLab artifacts are regenerated, run the connector
help/docs generators and inspect `pm help <connector>`, `pm <connector>`, and
one changed command's `--help` before claiming parity.
