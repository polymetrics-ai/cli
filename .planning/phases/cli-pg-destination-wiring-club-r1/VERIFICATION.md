# Verification checklist — PostgreSQL production transport wiring club

- [ ] All three issue amendments and the captain edge matrix map to tests and PR-body rows.
- [ ] `app.Open` call chain registers and dispatches PostgreSQL source, bootstrap/resume, warehouse stage, PostgreSQL destination, workset delivery, read-back, and checkpoint commit.
- [ ] Real built binary and production composition prove PostgreSQL → warehouse → PostgreSQL with actual state assertions.
- [ ] Missing/stale/expired/replayed approval and every supported refusal assert typed error plus zero/unchanged side effects.
- [ ] Empty, single, large, duplicate, out-of-order, cancellation, connection death, concurrency, resume, replay, schema drift, permission, and authentication cases are covered or explicitly labeled genuinely untestable with rationale.
- [ ] PostgreSQL public `write` remains false and target `change_capture`/history remain unadvertised or typed non-executable.
- [ ] CLI help, bare namespace, manual docs, website docs, generated manuals/skills, catalog, and golden transcripts are regenerated together and drift checks pass.
- [ ] Focused package tests, `internal/cli`, race tests, build/vet, container live tests, and scoped CI gates pass without changing #4125/#4158.
- [ ] Inline `verify-work` and `code-review` have no unresolved finding.
- [ ] Branch is committed/pushed, PR targets `integration/4015-mvp-flat-r1`, and the API-observed base is recorded.
