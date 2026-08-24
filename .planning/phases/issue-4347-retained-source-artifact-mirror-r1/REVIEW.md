# Issue #4347 — inline code review

## Scope reviewed

- `cmd/connectorgen/sourceimport.go`
- `cmd/connectorgen/sourceimport_test.go`
- `cmd/connectorgen/sourceartifact.go`
- `cmd/connectorgen/sourceartifact_test.go`
- `internal/connectors/defs/github/sources/`
- `docs/migration/conventions.md`
- Issue planning/TDD/verification artifacts

## Method

Manual standard review, following the resolved `scripts/gsd prompt code-review
4347` contract. The task forbids GSD role spawning, so no reviewer agent was
started. Reviewed retained-path derivation, source/manifest/file symlink
handling, lock-versus-manifest identity checks, byte limits, command dispatch,
and regression coverage; then ran the focused and complete `sourceimport` test
suites, generator package suite, vet, build, lint, and repository gates listed
in `VERIFICATION.md`.

## Findings and dispositions

- **Resolved — persistent network cache:** removed the former cache
  implementation rather than leaving an unreferenced fallback. `--cache-dir`
  now exits before a lock or fetcher is invoked.
- **Resolved — identity-query provenance drift:** the manifest's
  `identity_query` now participates in the match against the immutable lock;
  malformed identity-query provenance is rejected during manifest validation.
- **Resolved — v2 GraphQL gap:** an embedded GraphQL projection no longer lets
  a separately pinned raw schema go unverified. Import verifies both retained
  raw artifact identities before it reports the lock verified.
- **Resolved — maintenance write safety:** `source-retain` validates every
  fetched artifact before a write, derives content-addressed paths from the
  lock, does not overwrite a changed retained file, and leaves the lock bytes
  untouched. Its network path is explicit maintenance only, never build code.
- **Resolved — source-directory symlink:** a `sources` symlink is refused even
  when it resolves inside the connector bundle.
- **No remaining Critical/Warning findings in the foundation code.** A final
  inline review after the last lint/test repairs re-read the artifact writer,
  retained reader, lock/manifest match, source-directory path guard, and
  command dispatch; it found no additional issue.
- **Lane handoff, not a foundation blocker:** Elasticsearch must perform its
  own Firstmate-authorized re-pin/retain; Zoom must record the accounts source
  as irrecoverable rather than pin its 404 body. See `LANE-ADOPTION.md`.
- **Resolved after main rollup — unavailable source terminal order:** a v3
  all-unavailable lock no longer reports a missing retained manifest before its
  immutable unavailable reason. The new red/green regression proves that no
  retained-reader or provider path runs first; artifact-bearing locks retain
  the mandatory reader path.
- **Resolved after CI regression — optional source schema gap wording:** the
  read projection previously recognized only the old `oneOf` gap text as
  non-blocking when the parameter was optional. Current-main calls the same
  optional input non-scalar serialization, which wrongly downgraded 18 GitHub
  reads. Requiredness is the actual safety boundary, so the implementation now
  keys on the declaration and the new red/green test proves it. The complete
  fresh-binary 633-route fixture and source inventory are green.
