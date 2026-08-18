# Verification — GitHub mutation certification slice 4 writes-d

## Live result

- [x] All 145 manifest commands are assigned exactly once using the manifest's zero-based command index: `certified=30`, `no_object=18`, `wrong_credential=0`, `entitlement=13`, `not_implemented=0`, `product_defect=69`, `escape_needs_captain=15` (total 145).
  - `certified` (30): 24, 29–30, 47–48, 54, 57–62, 78, 80, 83, 85, 87–89, 92–94, 116–120, 132–134.
  - `no_object` (18): 64–75, 82, 91, 101, 111, 113, 128.
  - `wrong_credential` (0): none. Every 403/404 was retried or classified through the measured credential route before absence was accepted.
  - `entitlement` (13): 51–53, 104–109, 127, 139–140, 142.
  - `not_implemented` (0): none.
  - `product_defect` (69): 0–23, 25–28, 31–46, 49–50, 55–56, 63, 76, 90, 95, 97–100, 102–103, 125–126, 129–131, 135–138, 141, 144.
  - `escape_needs_captain` (15): 77, 79, 81, 84, 86, 96, 110, 112, 114–115, 121–124, 143.
- [x] The branch adds 30 schema-v2 passed evidence records over `origin/integration/4015-mvp-flat-r1`, one for every `certified` command and no record for any other bucket.
- [x] Each certified record contains a provider-visible produced-value or absence assertion, an `agent_derived` plausible-wrong rejection, and cleanup proven by a direct provider delete followed by an independent 404 or empty-collection read-back.
- [x] Batch-1 recovery replaced every provisional `no_object` that lacked a fixture with an honest product-defect result after real branch/user/repository fixtures and raw `api.github.com` controls. Its final split is `certified=5`, `product_defect=45` (total 50).
- [x] The resumed 50-command batch (indices 51–102 excluding already-certified 54 and 87) finishes `certified=15`, `no_object=15`, `wrong_credential=0`, `entitlement=3`, `product_defect=11`, `escape_needs_captain=6` (total 50).
- [x] The five enterprise configuration refusals at 77, 79, 81, 84, and 86 were recovered through the matrix-prescribed classic PAT: a direct read of the real `polymetrics-cert` enterprise collection returned 200. They are escapes, not credential failures, because the captain's real-account enterprise is outside the disposable boundary.
- [x] Tail indices 103–144 finish `certified=8`, `no_object=3`, `entitlement=10`, `product_defect=12`, `escape_needs_captain=9` (total 42). App-only check controls used the installed certification App; its task-local installation token was revoked directly and independently read back as 401.
- [x] The tail repository was independently readable before cleanup (200), deleted directly through GitHub (204), and independently read back as absent (404). Task-local `cert-tail`, `cert-tail-app`, and `cert-app-jwt` saved credentials were removed.
- [x] No GitHub credential value is present in the worktree, evidence records, status file, or planned PR body.
- [x] `go run ./cmd/connectorgen certification-matrix --check` passes after every retained tail record and at handoff.
- [ ] Repository verification and `scripts/verify-gsd-workflow` are run before the PR.
- [ ] The opened PR base is read from GitHub's API and equals `integration/4015-mvp-flat-r1`.
