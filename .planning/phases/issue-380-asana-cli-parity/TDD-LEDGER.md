# TDD Ledger

| Slice | Red evidence | Green evidence | Refactor evidence | Status |
| --- | --- | --- | --- | --- |
| CLI surface metadata | Pending worker test | Pending | Pending | planned |
| Help/docs renderer | `TestDocsGenerateAndValidateSelectedConnectorsDoesNotRewriteCorpus` proved `--connector` was ignored and rewrote a `100ms` sentinel. | Targeted `--connector asana --cli-connector github` generation and exact validation leave unrelated connector manuals unchanged and emit only the planned Asana/GitHub CLI pages plus Asana manual/skill. | Unfiltered repository-wide generation retains its existing behavior and structural validation; targeted output alone removes trailing whitespace. | green |
| Stream runner | Pending worker test | Pending | Pending | dependency-blocked |
| Operation ledger | Pending worker test | Pending | Pending | planned |
| Direct reads | Pending worker test | Pending | Pending | dependency-blocked |
| Attachment downloads | Pending worker test | Pending | Pending | dependency-blocked |
| Reverse ETL policy | Pending worker test | Pending | Pending | dependency-blocked |

Production edits are forbidden until the owning slice records a failing behavior test or a precise
non-code exemption. Existing unrelated failures must be named and must not be converted to passes.
