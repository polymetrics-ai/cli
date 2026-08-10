# Summary — Issue #3974 typed database foundation recovery

## Delivered

The complete ordered recovery range
`0be21e1293a9e5c913d9b4e23a846237594630c7^..9e763fe07a69b48f8ed88484ccb18580c63ae160`
was replayed from parent `8242bedee0cfa75fc61d1fb0e47d73251605162d`.

It establishes the typed, non-executing database definition foundation: strict policy loading,
logical types, catalog identities, stable bounded read plans, resource policy, exact native
admission, structural warehouse artifact identity, separate source/warehouse and warehouse/target
legs, and PostgreSQL's reference driver seam. It deliberately does not add target execution,
generic SQL/query/write, canonical modes, or CDC v2.

## Semantic replay dispositions

- `docs/architecture/connector-architecture-v2-design.md`: retained #4003 current-canon framing
  and added only the F1 policy/warehouse-mediation declaration.
- `docs/migration/conventions.md`: retained the current authoring recipe and added the optional,
  strict native `database.json` policy recipe.
- `internal/connectors/native/postgres/connector.go`: retained current TLS and fail-closed pglogrepl
  staging behavior; added only typed declaration wording/reference seam support.

## Correction status

One correction round of five is consumed by child #4026. The generated capability matrix drift was
proven to result from #3974's `bundle.go` source-location shift, regenerated through its declared
tool, and verified green. It changes no capability semantics.

## Remaining delivery gates

The source verification, inline code review, and no-mistakes run
`01KZPXF5VNSD89NNBWJT74BD00` are green. The next gates are push, child PR
creation/replacement linkage, and remote CI/automated-review observation. #3995 is recorded as a
bounded non-automatic `RETRY` compatibility verdict because its implementation is outside the
child/parent ancestry and PostgreSQL is intentionally uncertified at F1.
