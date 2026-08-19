# Review — MySQL container harness R1

`scripts/gsd sources code-review` and its generated prompt were resolved. The canonical contract
forbids spawning the prescribed reviewer role in this task, so this is the required inline review.

## Reviewed surfaces

- `dbtest` scopes every container command to a direct local Unix Podman endpoint, allocates only
  run-named resources, and runs all cleanup stages after failure/interrupt. Endpoint identity and
  image-store capacity are rechecked before every command.
- MySQL's full and incremental reader uses validated identifiers, parameterized keyset boundaries,
  primary-key tie-breaking, and bounded pages. The native reader does not masquerade as #3902
  page-wise direct-read functionality.
- MySQL's internally proven row-event reader retains its row-image fail-closure and checkpoint
  safeguards, while metadata, definition, catalog, and generated projections keep `cdc: false`
  until a production runtime entrypoint exists.
- TLS is fail-closed outside `preferred`; normal MySQL and replication connections share the same
  configuration. PostgreSQL now consumes the same explicit options through `openPool` for Check,
  Catalog, and Read, without touching its write path.
- Definition docs, generated catalog data, website data, and the operation surface were regenerated
  and inspected. The unrelated warehouse catalog wording produced by current-main's docs generator
  is derived current-main drift, not a hand edit.
- Production native wiring is generated from connector packages only; `dbtest` and `sqltls` are
  explicit support-library exclusions and are covered before regenerating the native set.
- The dependency is direct, MIT-licensed, externally maintained, vulnerability-scanned, and has a
  recorded binary-cost decision point.

## Findings and disposition

| Finding | Disposition |
| --- | --- |
| PostgreSQL accepted canonical TLS aliases in runtime code but its definition rejected them before invocation. | Fixed in `41275a450`; red/green tests prove definition/runtime/pool alignment. |
| A native SQL ETL reader could be mistaken for a #3902 page-context direct reader. | Documented and locked by `e36fbebc2`; no `DirectReader` is exposed. |
| An opted-in live startup failure lost its sanitized stage reason. | Fixed locally: the integration test now displays the safe harness stage without exposing endpoint or authentication material. |
| Earlier planning/PR text claimed actions on other Podman machines. | Superseded: the harness no longer creates or removes Podman machines. |
| Recovery of the prior pipeline head showed generated production wiring blank-imported `dbtest`; current-main also placed shared `sqltls` alongside connectors. | Fixed before recovery: source has explicit support-library exclusions, the generated file was regenerated, and a red/green generator test covers both. |
| MySQL advertised CDC without a production operator entrypoint. | Fixed: metadata, definition, catalog, docs, and website projections now keep `cdc: false`; the internal reader has no public changefeed descriptor or executor. |
| A task-owned machine could leave Podman's global default altered, and a remote pull used local free-space reporting. | Fixed: task-owned machine lifecycle was removed. A direct local Unix endpoint is identity-bound to its own reported socket and image-store path before every mutation; remote and named endpoints fail closed, including for cached images. |
| The live TLS matrix omitted `verify-ca`, and public connector documentation contained harness maintenance instructions. | Fixed: the tagged proof copies the container CA for a live `verify-ca` session; maintenance instructions now live in `dbtest/README.md`. |

## Verdict

No unresolved implementation finding remains. `go test -count=1 -timeout 5m
./internal/connectors/native/dbtest` passed for the direct-endpoint correction. The tagged live proof,
automated GitHub review, and CI remain outer PR-stage gates.
