# TDD ledger — Zoom ETL certification parity

| Slice | Red | Green | Refactor | Status |
| --- | --- | --- | --- | --- |
| Generated sweep artifact | `go test -timeout 20m ./internal/connectors/defs/zoom -run TestCertificationSweepProjectsExistingETL -count=1` fails because `certification-sweep.json` is absent (`open certification-sweep.json: no such file or directory`). | Generate with `go run ./cmd/connectorgen certification-sweep --connector zoom`; the same test passes and asserts three ETL and one capability-read projection. | Run `surface-sync --check` and ensure the resulting diff is only the generated Zoom artifact plus connector-local test/evidence. | green |

Live proof is intentionally absent: there is no approved server-to-server OAuth consumption path or accepted live evidence. The 2026-08-19 captain decision also keeps Zoom outside the central certification scope. This wave therefore records fixture-required, implemented-and-pending-certification status only; the expected `capability/zoom/missing` HALT is not an implementation blocker and cannot be presented as certification.
