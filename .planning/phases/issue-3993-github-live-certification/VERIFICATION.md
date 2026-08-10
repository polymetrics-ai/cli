# Verification checklist — Issue #3993

- [ ] Deterministic harness tests show that the supplied `Polymetrics-Cert` boundary controls both emitted cases and classification.
- [ ] Deterministic harness tests show a common barrier release and terminal records with read-back evidence.
- [ ] The two artifact self-checks pass after regeneration.
- [ ] A real `pm` binary is built and its hash recorded without secrets.
- [ ] The App installation credential is used without disclosure; the revoked fine-grained token is untouched.
- [ ] The whole applicable surface has a complete/failed/unavailable tally and failures are grouped by actual cause and quota bucket.
- [ ] Every created resource has read-back, inverse cleanup, and final empty-residue proof.
- [ ] GitHub → Parquet warehouse → DuckDB inbound flow is independently proven.
- [ ] The outbound workflow refusal is attributed to #3994/#3992, with no duplicate action-path implementation.
- [ ] Targeted local checks and required non-full-suite verification gates pass.

