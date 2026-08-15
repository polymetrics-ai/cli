---
coverage:
  - id: D1
    description: "PostgreSQL is a production-registered managed destination for declared PostgreSQL and authenticated API sources."
    requirement: "#3982"
    verification:
      - kind: e2e
        ref: "internal/cli/postgres_transport_binary_integration_test.go — authenticated real GitHub issue through built pm and warehouse into live PostgreSQL"
        status: pass
      - kind: integration
        ref: "internal/app/transport_composition_test.go — app.Open exact factory registration/preflight and write=false"
        status: pass
    human_judgment: false
  - id: D2
    description: "A real PostgreSQL source reaches the managed PostgreSQL target only through immutable connection-owned warehouse worksets."
    requirement: "#3983"
    verification:
      - kind: e2e
        ref: "internal/cli/postgres_transport_binary_integration_test.go — 1,001-row PostgreSQL→warehouse→PostgreSQL binary proof, replay and refusals"
        status: pass
      - kind: integration
        ref: "internal/connectors/native/postgres/managed_target_integration_test.go — live workset insert/update/unchanged/tombstone/receipt/baseline proof"
        status: pass
    human_judgment: false
  - id: D3
    description: "The production incremental route invokes gap-free bootstrap and resumes acknowledged pgoutput state after process interruption."
    requirement: "#3979"
    verification:
      - kind: e2e
        ref: "internal/cli/postgres_transport_binary_integration_test.go — built-binary bootstrap transaction, target state, LSN checkpoint, kill/restart/resume"
        status: pass
      - kind: integration
        ref: "internal/connectors/native/postgres/cdc_integration_test.go — before/during/after boundary and failed-snapshot explicit rebootstrap"
        status: pass
    human_judgment: false
  - id: D4
    description: "Approval, cancellation, empty/single/large, replay, drift, auth, permission and concurrency boundaries fail closed or preserve durable state."
    verification:
      - kind: unit
        ref: "focused App/database/postgres/synctransport tests plus focused -race run"
        status: pass
      - kind: e2e
        ref: "both production-binary integration tests assert target rows/receipts/artifacts/checkpoints, not exit status"
        status: pass
    human_judgment: false
---

# Summary — PostgreSQL production transport wiring club

## Outcome

`app.Open` now discovers PostgreSQL's definition-owned managed destination,
constructs its connector-local factory, preflights it with either the declared
PostgreSQL snapshot/bootstrap source or the declared GitHub issue source, and
dispatches only after a sealed plan/preview/one-time approval. Both sources pass
through connection-owned WAL, Parquet, and manifest artifacts; the destination
reopens the workset, invokes the native managed driver, reads back durable
evidence, and only then admits the source checkpoint.

The authenticated GitHub→PostgreSQL and decisive PostgreSQL→PostgreSQL built-
binary tests passed against real systems. The latter additionally proves the
bootstrap App bridge, transactional pgoutput delivery, deliberate process death,
and resume from the acknowledged LSN.

The final review removed `full_overwrite` from PostgreSQL's bounded, multi-page
source declaration because applying replace once per page cannot honor a whole-
snapshot overwrite. The destination still declares all five required modes;
the authenticated one-row API route proves its safe production overwrite path.

## Verification

- Focused package, race, vet, lint, and build commands passed.
- Live workset and bootstrap component tests passed through `dbtest`.
- Derived docs/catalog/skills/website/goldens were regenerated together.
- Tidy, docs, smoke, agent contract, surface-sync, GitHub artifact, connector
  canon/boundary, and release drift checks passed.
- Per firstmate load control, no new `go test ./...`, full-parity certification,
  certification matrix, or runtime-preflight sweep was launched; CI owns those
  repository-wide runs.

## Disclosed exclusion

The captain forbade fixing #4158. Its failing assertion reaches the unadvertised
`incremental_dedupe_history` polling route after the five destination modes this
change declares; this PR does not advertise history and does not claim #4158 is
green. The PR edge table records that exact boundary explicitly.
