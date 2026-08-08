# TDD Ledger — Zoom SCIM2 documented-operation parity, R1

## Planned RED contract

Before any production engine or Zoom bundle change, the RED checkpoint will contain only tests and
planning evidence. It must fail against the current branch because:

- Zoom remains at `27` executable / `1,815` local implementable rows, with `17` direct reads and
  `5` direct writes; the target requires `38` / `1,804` / `21` / `12`.
- All eleven SCIM2 paths are absent from the real commandrunner preflight, so a compiled `pm zoom
  scim2 …` route cannot yet resolve.
- A declared named `json_object` input targeted at exact root `body` is rejected, despite SCIM's
  documented extensible resource and PatchOp bodies. The existing generic `json` type remains
  deliberately unsupported.
- A declared rest-read operation still uses the ordinary bundle origin/auth and therefore cannot
  honor a provider root endpoint such as `/scim2/...` independently of ordinary `/v2` calls.

The pending RED run is test-only and will be captured verbatim below before any production change.
It contains no provider credential or token value.

## Planned GREEN foundation — operation-scoped direct-read origin/auth

The foundation will extend the existing paired `rest.base_url` / `rest.auth` transport contract to
bounded direct reads. A rest-read with the override must issue only to its declared origin, use only
its declared auth, clear unrelated global headers, and retain the fixed relative endpoint and
response cap. It is necessary for SCIM2 because Zoom documents the API server at the host root while
ordinary Zoom operations use `/v2`; it also unblocks similarly documented distinct-origin reads.

## Planned GREEN foundation — named root JSON-object body

The foundation will permit exact `maps_to: "body"` only for a named `json_object` flag on an
operation-declared object body schema. It will reject path/query mappings, all generic `json` input,
and mixed root/field body mappings. The declared operation remains responsible for schema and byte
limits. This is needed for Zoom SCIM2's documented extensible User/Group/PatchOp object bodies,
including custom extension attributes, without creating raw transport capability.

## Planned GREEN connector contract

- Four SCIM2 reads use bounded `json_redacted` output with declared PII/account field redaction.
- Seven SCIM2 mutations use declared typed direct writes and the plan lifecycle; both DELETEs use
  destructive confirmation, and 204 status-only actions return `none` rather than an invented body.
- All group/user resource and patch inputs are synthetically fixture-tested for exact method, root
  path, declared Bearer header, JSON body, redaction, and no paging input.
- The ledger changes only the eleven `provider_module=scim2` endpoints; zero Zoom rows become
  `unsafe_or_disallowed`.
