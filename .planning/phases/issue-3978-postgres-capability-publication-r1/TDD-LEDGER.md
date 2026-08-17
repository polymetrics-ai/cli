# TDD ledger — Issue #3978: final PostgreSQL certification and publication

| ID | Red guarantee | Green implementation / evidence | Status |
| --- | --- | --- | --- |
| PGFINAL-1 | The original mode-to-capability binding was halted because `sync_mode` cannot satisfy `capability`. | No scope check, baseline, or fixture was weakened. Exact records remain on exact mode cells. | Green: protected certification-gate tests still pass. |
| PGFINAL-2 | `TestCertificationMatrixDoesNotTreatPostgresManagedTransportAsGenericWrite` failed with `PostgreSQL managed destination evidence promoted generic write capability`. | The generic write implementation again follows the direct `Connector.Write` stub, so it is false despite complete closed managed-target evidence. | Green: focused `cmd/connectorgen` test. |
| PGFINAL-3 | The composition guard rejected `write=true`: a closed managed target is not a generic writer. | Metadata, manifest, generated matrix, docs, and website publish `write=false`; the closed `destination_transport` remains declared. | Green: `TestOpenRegistersDefinitionOwnedProductionTransports`. |
| PGFINAL-4 | A complete profile must not disappear merely because it cannot make a generic capability true. | `TestCertificationMatrixRetainsPostgresManagedDestinationEvidenceAtExactModeScope` requires all six `database_write_from_warehouse` cells to be declared, implemented, live-tested, and carry one exact proof each. | Green: focused `cmd/connectorgen` test. |
| PGFINAL-5 | CDC acknowledgement before receipt is invalid. | Receipt-backed PostgreSQL 16 binary evidence keeps `cdc=true` and one `change_capture/database_read_into_warehouse` cell true; API/destination CDC remains non-pass. | Green: existing tagged binary proof and focused matrix tests. |
