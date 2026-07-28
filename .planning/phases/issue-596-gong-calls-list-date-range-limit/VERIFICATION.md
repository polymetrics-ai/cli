# Verification — issue 596 Gong calls list correction

## Targeted

- [ ] `go test ./internal/connectors/commandrunner -run 'Test.*Gong.*CallsList' -count=1`
- [ ] `go test ./internal/cli -run 'Test.*Gong.*CallsList|TestGong' -count=1`
- [ ] `go test ./cmd/connectorgen -run TestGong -count=1`
- [ ] `go test ./internal/connectors/conformance -run 'TestConformance/gong' -count=1`
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`

## Help/docs parity

- [ ] `go run ./cmd/pm help gong`
- [ ] `go run ./cmd/pm gong calls list --help`
- [ ] `go run ./cmd/pm docs validate --dir docs/cli --connectors-dir docs/connectors --website-dir website/content/docs`
- [ ] grep docs/website generated surfaces for `calls list`, `--from`, `--to`, and output-limit wording.

## Broader local gates

- [ ] `gofmt -w cmd internal`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`

## no-mistakes / PR

- [ ] Commit implementation on feature branch.
- [ ] Run `no-mistakes axi` home view.
- [ ] Run `no-mistakes axi run --intent "..."` and drive gates through `checks-passed` or terminal outcome.
- [ ] Push branch and open PR with `Closes #596`.
- [ ] Confirm automated review route status; do not merge or release.
