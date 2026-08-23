# Issue #4288 discussion log

| Area | Decision | Source |
| --- | --- | --- |
| Connector scope | Jira → Asana → Notion; no other connector is touched. | Issue #4288 and launch brief |
| Missing generator scope | Add only the three targets to the reviewable certification allowlist, then regenerate the normal scope artifacts. | Red `certification-matrix --check` observations |
| Capability truthfulness | Certify only existing declared/implemented cells; persist concrete non-pass reasons for every other cell. | Issue #4288 |
| Live-account safety | Use separate provider mailboxes and secret-safe stdin/keychain paths; stop for the runbook's human gates. | Provisioning runbook §§0–16 |
| Live writes | Read first; use scratch-owned records only; independently prove cleanup. | Issue #4288 and AGENTS.md |

No interactive question is outstanding. The GSD `discuss-phase` prompt was run inline because
this direct-PR lane cannot create compatible lifecycle roles.
