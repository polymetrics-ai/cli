# Source Reports

`data/` holds source-pinned reports imported into this repository so a clean
checkout has the evidence used by the [connector delivery canon](../docs/connector-canon/INDEX.md).

The current sources are listed in `CANON-MANIFEST.sha256`. The PostgreSQL
parity r2 correction is repository-authored from a REST read-back; its r1
baseline remains recoverable under `data/archive/` with an explicit
supersession marker.
`CURRENT-CORRECTIONS.md` makes historical measurement corrections explicit
without rewriting a source-pinned captain record. Material under `data/archive/`
is retained, explicitly marked, and is not current authority.
