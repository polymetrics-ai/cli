# Summary — issue 4329 source-cited read-only coverage

`api_surface` now has one additional closed operation model: `read_only`.
It is effective only when a source-locked, non-mutating operation has the same
method/path and an exact declaration-owned policy note. A POST, PUT, PATCH, or
DELETE marker is rejected; the absent-action mutation case stays visible for
`cli-mutation-disposition-foundation-r1`.

The operation-evidence projector emits an explicit `read_only` row and an
`intentionally_read_only` rollup keyed by connector and policy. An executable
runtime/CLI/website route, a source foundation gap, or malformed declaration
is a contradiction rather than an escape hatch.

No Sentry or Vercel connector definition file changed. Their mutation handoff
was copied to the firstmate workspace, outside the product repository.
