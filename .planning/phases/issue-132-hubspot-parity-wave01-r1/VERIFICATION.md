# Verification — HubSpot parity wave01 r1 (#132)

## Local gates

| Command | Result |
|---|---|
| `go run ./cmd/connectorgen validate internal/connectors/defs/hubspot` | red before generation: bundle path absent (`validate: read root: open .: no such file or directory`); after generation this single-bundle path checks 0 connectors, so the meaningful gate is full defs validation below |
| `go run ./cmd/connectorgen validate internal/connectors/defs` | pass: 549 connectors checked, 0 findings |
| `go test ./internal/connectors/conformance -run 'TestConformance/hubspot' -count=1` | pass |
| `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` | pass after updating generated golden transcripts and catalog count |
| `./pm docs validate --connectors-dir docs/connectors` | pass |
| `go vet ./internal/connectors/... ./internal/cli/...` | pass |
| `go build ./cmd/pm` | pass |
| `make connector-boundary` | pass: outcome clean |
| `git diff --check` | pass |

## CLI/help/docs parity notes

- `./pm help connectors` exits 0.
- `./pm connectors inspect hubspot --json` exits 0 and reports zero streams/write actions for the ledger-only planned bundle.
- `./pm help hubspot`, `./pm hubspot`, and `./pm hubspot --help` exit 0 and render connector-owned planned command metadata.
- `./pm hubspot --bogus` exits 2; invalid action/flag forms still fail rather than silently succeeding.
- Generated docs retained only HubSpot connector docs plus connector catalog index changes.

## Safety checklist

- [x] No live HubSpot provider calls.
- [x] No credentials requested, printed, summarized, or stored.
- [x] No live writes, destructive provider actions, certification, VPS/Thaalam work, or merges.
- [x] DELETE/destructive operations are included as in-scope typed destructive-confirmation candidates, not blanket exclusions.
- [x] No generic raw HTTP method/path/body, shell, SQL, unrestricted file, arbitrary GraphQL, or raw passthrough command.
- [x] Shared missing foundations recorded without claiming implementation completion (#2985, #2986, #2988).
