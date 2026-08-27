# Summary — #4352 source-bound read execution foundation

## Delivered

- Added a closed `source_operation` declaration edge: one locked provider ID,
  exact method, and connector-relative path on the operation and canonical
  command. There is no arbitrary URL, method, route, header, body, curl, or
  connector-local request shim.
- Source projection now evaluates every non-mutating locked GET from capability,
  not its historical `planned` label. A source-complete bounded JSON REST
  operation becomes a `direct_read`; ETL is selected only by an exact declared
  stream with record path, schema, and matching pagination semantics.
- The direct executor imports typed source-owned scalar path/query inputs and
  retains the response cap/output policy. Missing typed contracts and named
  foundation gaps stop before credential/auth/request construction.
- Source-bound operations reject an alternate caller `base_url` before auth or
  I/O. Retained raw artifact integrity remains strict, but mapping admission no
  longer depends on capture hashes, byte counts, adapters, or certification.
- Regenerated the V3-source operation-evidence outputs, tracked Asana manuals
  and skills, root-help goldens (no delta after regeneration), and website
  catalog data from the final bundle.

## Asana reconciliation from the pinned lock

The retained V3 Asana source descriptor has 249 operations and 119 GETs.
Capability materialization is exact and source-backed:

| Lane | Count | Proof |
| --- | ---: | --- |
| Bounded direct read | 106 | Exact `source_operation` ID/method/path, typed scalar input contract, bounded JSON operation. |
| ETL stream | 12 | Exact source route plus declared records/schema/pagination. Includes `getProjectStatusesForProject`, `getSectionsForProject`, and `getStoriesForTask` fan-out streams. |
| Deferred GET | 1 | `asana.rest.getMembership` remains explicitly blocked by `cli-openapi30-reference-sibling-foundation-r1`. |
| Mutations | 130 | 21 source-cited absent actions, 65 implemented partial request-schema contracts, and 4 implemented partial path-alias delete contracts retain their prior lanes. |
| Batch wrapper | 1 | `/batch` remains `unsupported_api`; it would be an arbitrary nested request escape hatch. |

The drift regression changes a complete direct read or a fan-out ETL command
back to `planned` and requires `source-import --check` to fail. Thus a
historical label cannot silently override a currently proven shared capability.

## Batch 1 materialization checklist

The historical 100 planned Asana GET declarations are no longer a status
partition. On each pinned-source refresh, Batch 1 must:

1. Run `connectorgen source-import asana --read-projection-only --check`; it
   considers every locked GET and detects stale command/operation binding.
2. Keep each direct read tied to an existing bounded `rest_read`, exact locked
   identity/method/path, and source-derived typed scalar flags. Do not promote
   a header/body/non-scalar contract until its shared foundation exists.
3. Keep a collection in ETL only when its existing stream proves records,
   schema, and provider pagination. The three fan-out streams above are the
   source-backed examples; all other eligible GETs use one-page direct reads.
4. Leave `getMembership` deferred until
   `cli-openapi30-reference-sibling-foundation-r1` supplies its missing typed
   contract; preserve the stable `missing_foundation` note in the command.
5. Re-run source import, `validate`, `surface-sync --check`, operation evidence,
   generated docs/catalog, and credential-bound binary preflight before any
   future provider-lock update.

For another connector batch, retain its source lock/descriptor, provide an
existing fixed REST operation or semantically complete stream, then run the
same generator and checks. A source gap remains source-cited and named; it does
not need a hash, live certificate, provider call, or connector-local shim to be
admitted as a declaration.

## Integration

PR #4356 remains open and unmerged. It requires a fresh independent Codex audit
of the new exact head, normal automated review coverage, and the human merge
gate.

## r4 repair delta for PR #4356

This continuation repairs the six frozen audit findings without changing the
Asana source lock or retained provider bytes. The final executable source map
is 106 direct reads, 12 ETL streams, and 94 reverse-ETL actions (51 create, 20
update, 23 delete); 19 source-complete DELETEs and two no-body POST mutations
now use the established action/approval path. Remaining source rows are either
an exact `missing_foundation` or `/batch`, explicitly not applicable as a
generic wrapper. A fresh independent audit is required after the repair SHA is
pushed; this PR remains open and unmerged.
