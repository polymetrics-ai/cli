# UAT — issue #3715 Pi clean project-only workers

Status: passed by automated execution; no human judgment is required for these structural
properties.

| Deliverable | Evidence | Result |
| --- | --- | --- |
| Clean discovery | `bash scripts/tests/pi-clean-project-agents.sh` imports the real extension with a hostile global fixture, retained project roles, and the historical bundled directory present. | Pass — only the two generated project workers appear. |
| Project trust guard | The same executable test invokes the registered default tool without UI. | Pass — a clean-project worker is refused before spawn until explicit trusted confirmation. |
| Delegation boundary | The same test proves `subagent` stripping, depth rejection, and `--no-extensions` child arguments. | Pass. |
| Canonical generated files | `go test ./internal/agentcontract` and `go run ./cmd/agentcontractgen check`. | Pass — missing required files are generated and whole-file drift is rejected. |

The task's stated isolation property is proven for the actual forked subagent path: ambient roles
are not discoverable in clean mode, and child Pi processes load no extensions. This does not claim
that the intentionally allowed built-in `bash` tool is sandboxed.
