# Verification — issue #4158 / Production MVP verify green

| Requirement | Evidence | Status |
| --- | --- | --- |
| Fresh externally built binary completes GitHub → warehouse flow | Original fixture: repeatably fails exit 3. Smallest job-only fixture counterfactual: passes and proves flow receipt / durable warehouse behavior. | Needs authority to select the intended contract |
| Valid managed-target control path reaches durable acknowledgement | Live PostgreSQL test correctly skipped without its explicit container opt-in; independent issue #4158 requires a separately provisioned lane. | Not established in this worktree |
| Non-PostgreSQL history route remains typed and pre-I/O | New bad regression; fake driver zero side effects | Pending |
| Causal explanation distinguishes trigger, mask, symptom | `SUMMARY.md` §1–§5 and `TDD-LEDGER.md` T1–T5 | Complete |
| Smallest counterfactual and falsifier checks are retained | `TDD-LEDGER.md` T3–T4 | Complete |
| CLI help/docs/website parity | The decision may change the public flow-manifest contract; re-evaluate after authority selects fixture migration or compatibility. | Pending decision |
