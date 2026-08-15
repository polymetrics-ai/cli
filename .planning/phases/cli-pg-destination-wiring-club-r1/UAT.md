# UAT — PostgreSQL production transport wiring club

Verdict: passed by automated production-composition evidence.

- D1: passed — authenticated real GitHub API → warehouse → live PostgreSQL.
- D2: passed — real PostgreSQL → warehouse → live PostgreSQL with 1,001 rows,
  exact replay, receipt, baseline, read-back, and checkpoint assertions.
- D3: passed — bootstrap boundary transaction and process restart resume through
  the built binary; focused live source tests prove failed-snapshot rebootstrap.
- D4: passed with #4158 explicitly excluded — required refusal and boundary
  tests assert typed errors plus unchanged target/delivery/checkpoint state.

No browser or visual judgment applies to this CLI/database phase.
