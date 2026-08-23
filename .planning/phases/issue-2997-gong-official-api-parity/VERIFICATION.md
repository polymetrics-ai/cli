# Verification — Gong Official API Parity Completion (#2997)

| Command | Purpose | Result |
| --- | --- | --- |
| `gh-axi issue view 2997 --full` / subissues, then `gh-axi issue edit --body-file` | Read parent/subissue contracts; append idempotent captain-policy addendum | Pass; marker present exactly once on #2997-#3004; existing count tables preserved |
| `scripts/gsd doctor` | GSD adapter preflight | Pass |
| `scripts/gsd list` | Confirm adapter registry | Pass; `programming-loop` command absent |
| `scripts/gsd prompt programming-loop init --phase issue-2997-gong-official-api-parity --dry-run` | Required programming-loop path | Failed: `unknown GSD command: programming-loop`; manual-GSD fallback recorded |
| `go test ./cmd/connectorgen -run GongAPISurfaceOperationLedger -count=1` | Red/green operation ledger count test | Red failed at `endpoints = 67, want 69`; green passed (`ok`, 0.310s) |
| `GOCACHE=/tmp/fm-cli-gong-parity-wave01-r1/gocache go test ./cmd/connectorgen -run 'Gong' -count=1` | Connector-owned Gong operation/full-surface tests after 2026-08-01 re-audit | Pass (`ok`, 0.565s) |
| `go test ./cmd/connectorgen -count=1` | Connector-owned connectorgen/Gong full-surface tests | Pass (`ok`, 12.164s) |
| `GOCACHE=/tmp/fm-cli-gong-parity-wave01-r1/gocache go run ./cmd/connectorgen validate internal/connectors/defs` | Bundle validation after 2026-08-01 re-audit | Pass: `548 connector(s) checked, 0 findings` |
| `go run ./cmd/connectorgen validate internal/connectors/defs/gong` | Issue checklist literal command | Current validator treats `schemas/` and `fixtures/` as bundle dirs for this path and fails with missing metadata; root-level validation above is the valid current invocation |
| `GOCACHE=/tmp/fm-cli-gong-parity-wave01-r1/gocache go test ./internal/connectors/conformance -run 'TestConformance/gong' -count=1` | Gong conformance replay after 2026-08-01 write-fixture/path fixes | Pass (`ok`, 7.285s) |
| Official-source re-audit script (`curl -fsSL https://gong.app.gong.io/ajax/settings/api/documentation/specs?version=` + local JSON comparison) | Verify no omissions/stale rows or required schema/flag gaps | Pass: officialOps=69, surface=69, coverage stream=12/write=27/direct_read=30, writeFind=0, directFind=0 |
| `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` | CLI command metadata coverage | Pass (`ok`, 192.036s) |
| `./pm help gong`, `./pm gong`, `./pm gong targets --help`, `./pm gong targets list --help`, `./pm gong targets upload-assignments --help` | CLI help/docs parity smoke | Pass; targets group appears, commands show bounded/direct/reverse-ETL flags and plan/preview/approval/confirm safety text |
| `go vet ./internal/connectors/... ./internal/cli/...` | Focused vet | Pass |
| `go build ./cmd/pm` | CLI build | Pass |
| `make connector-boundary` | Connector Guard boundary | Pass; outcome clean, 0 findings/warnings |
| `git diff --check` | Whitespace/diff hygiene | Pass |
| `/Users/karthiksivadas/karthik-agent-workspace/bin/fm-ensure-agents-md.sh .` | Firstmate project-memory maintenance hook | Ran; exited with `conflict: both AGENTS.md and CLAUDE.md are real files`; no manual edits made because this task produced no new global project instruction and allowed scope is connector-local |

No live credentials, live provider calls, provider writes, certification, or merges were performed.
