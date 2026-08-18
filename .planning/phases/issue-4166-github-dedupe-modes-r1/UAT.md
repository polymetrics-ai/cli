# UAT: GitHub dedupe modes r1

Manual inline verify-work fallback: this task runs as a single autonomous
worker and the repository contract forbids GSD-role spawning.

| Deliverable | Result | Evidence |
| --- | --- | --- |
| Generator truthfulness | Pass | Focused red/green test and two byte-identical `certification-matrix --all` runs. |
| GitHub current dedupe | Pass | Built `pm` read the retained private PR after an update and a replay; the independent warehouse query remained one row. |
| GitHub dedupe history | Pass | Built `pm` recorded an earlier closed and later current source version; a third replay remained two rows. |
| ETL sync-mode help | Pass | The existing `TestETLHelpListsAllSyncModes` assertion now finds the complete history-mode description in rendered `pm help etl`; regenerated manual and golden transcript match it. |
| Bad input | Pass | Built `pm connections create` rejected `incremental_dedupe_not_a_mode` before any provider request; `UnsupportedSyncModeError` is asserted by the parser unit test. |
| Independent provider read | Pass | `pm github pr list --connection github-live --json` independently reported the source key and current update timestamp. |

Verdict: pass. The live proof is reproducible only by an authorized maintainer
with access to the retained private repository; no secret is persisted.
