# Native Go All Connectors Implementation Prompt

Use the `gsd-programming-loop` and `go-engineering` skills.

Implement the full catalog connector native binding program in the Polymetrics `pm` Go CLI:

- Every catalog connector must have a Go runtime binding.
- No connector images, Python, Java, Ruby, shell plugins, or untrusted dynamic code may be executed.
- Keep GitHub as the reference hand-written live SaaS connector.
- Use fixture-backed conformance for connectors without live credentials.
- Preserve reverse ETL plan, preview, approval, run, and receipt boundaries.
- Keep secrets out of logs, JSON, docs, skills, errors, previews, and test fixtures.
- Verify with `go test ./...`, `go build ./cmd/pm`, docs validation, and `make verify`.
