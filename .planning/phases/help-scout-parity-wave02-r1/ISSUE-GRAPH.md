# Help Scout parent/subissue graph

Parent: #212 — `feat(connectors): Help Scout official API parity parent`

Subissues:

| Issue | Lane | Count disposition from issue body | Dependency note |
| --- | --- | --- | --- |
| #213 | ledger | cross-cutting, count 0 | no serialized foundation beyond parent readiness |
| #214 | etl_cdc | total 48, implemented 4, blocked 44 | #2986, #2988 for CDC truth/state; this worker kept `cdc=false` |
| #215 | direct_binary | total 36, blocked 36 | #2985 provider search/query; binary download foundation still needed |
| #216 | reverse | total 60, blocked 60 | no serialized foundation beyond parent readiness |
| #217 | cli_config | cross-cutting, count 0 | #2985 provider search/query |
| #218 | fixtures_guard | cross-cutting, count 0 | no serialized foundation beyond parent readiness |
| #219 | cert_release | cross-cutting, count 0 | no serialized foundation beyond parent readiness |

Captain-policy addendum:

- Marker: `help-scout-captain-policy-addendum-wave02-r1`.
- Applied with `gh-axi issue edit --body-file` to #212-#219 on 2026-07-31.
- Verification fetched each issue with `gh-axi issue view --full --json body` and confirmed marker presence.
- Existing issue body count tables were preserved; the addendum was appended after the original body.
