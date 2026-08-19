# UAT — PostgreSQL production transport wiring club

Verdict: passed by automated production-composition evidence.

- D1: passed — built `pm` moved exactly 50 authenticated `rails/rails` issues
  through independently counted Parquet into 50 independently queried live
  PostgreSQL rows with the declared logical types and an advanced checkpoint.
- D2: passed — real PostgreSQL → warehouse → live PostgreSQL with 1,001 rows,
  exact replay, receipt, baseline, read-back, and checkpoint assertions.
- D3: passed — bootstrap boundary transaction and process restart resume through
  the built binary; focused live source tests prove failed-snapshot rebootstrap.
- D4: passed with #4158 explicitly excluded — missing approval, consumed replay,
  and provider-request-in-flight cancellation are typed and preserve target rows
  and checkpoint; the remaining required refusal and boundary tests assert the
  same unchanged target/delivery/checkpoint state.

No browser or visual judgment applies to this CLI/database phase.
