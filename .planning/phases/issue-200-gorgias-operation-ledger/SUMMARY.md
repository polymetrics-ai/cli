# Summary: Gorgias Operation Ledger

Parent issue: #196  
Sub-issue: #200  
Branch: `feat/200-gorgias-operation-ledger`

## Result

- Converted `internal/connectors/defs/gorgias/api_surface.json` to operation-ledger mode (`operation_ledger_version: 1`).
- Accounted for all 114 public Gorgias operations captured from official `llms.txt` plus linked ReadMe/OpenAPI markdown pages.
- Preserved executable coverage only for the four existing streams:
  - `GET /tickets` -> `tickets`
  - `GET /customers` -> `customers`
  - `GET /messages` -> `messages`
  - `GET /satisfaction-surveys` -> `satisfaction_surveys`
- Classified every other operation as blocked-by-default metadata:
  - `direct_read`: 42
  - `binary_read`: 5
  - `admin_reverse_etl`: 27
  - `sensitive_reverse_etl`: 15
  - `destructive_action`: 20
  - `disallowed`: 1
- Removed legacy `excluded` rows from the Gorgias API surface.
- Added `cmd/connectorgen/gorgias_api_surface_test.go` to lock the 114-operation ledger counts and method/model splits.
- Updated Gorgias docs wording so it says the ledger accounts for all 114 operations while runtime execution remains limited to existing streams.

## Safety posture

- No new runtime commands were made executable.
- No credentialed Gorgias checks were run.
- No write actions were declared.
- All non-stream operations remain blocked by default pending typed implementation lanes and approval workflows.
