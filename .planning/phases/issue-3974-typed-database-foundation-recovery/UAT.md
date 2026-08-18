# UAT — Issue #3974 typed database foundation recovery

## Automated acceptance result

**PASS for this non-interactive foundation slice.** No human-judgment UI or provider mutation is
in scope. The built binary was inspected without credentials:

- PostgreSQL reports `write=false` and `query=false`.
- Its changefeed is `planned`, and the CDC catalog contains no PostgreSQL entry.
- `metadata.json` remains `write:false`, `query:false`, `cdc:false`; `database.json` has
  `admitted_modes=[]`.

This accepts only the intended typed policy/admission foundation. It does not represent a
certification, a PostgreSQL write/query/CDC release, or a permission to integrate/merge the child PR.
