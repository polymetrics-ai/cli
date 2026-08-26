INSTALL_DIR ?= $(HOME)/.local/bin
VERIFY_JOBS ?= 2

# go.mod requires Go 1.25 and pins a patched toolchain. Allow the go command to
# fetch the matching toolchain when the ambient one is older.
export GOTOOLCHAIN ?= auto

# DuckDB is the query engine and the only Parquet implementation in the binary,
# so cgo is required rather than optional. Building with CGO_ENABLED=0 no longer
# produces a pm that can read or write a warehouse table.
export CGO_ENABLED ?= 1

.PHONY: fmt vet tidy-check test build icons-generate docs-check docs-check-no-build install uninstall smoke smoke-no-build pinned-build-dependencies-check release-workflow-check verify verify-parallel perf-free perf-runtime runtime-doctor runtime-up runtime-down runtime-reset clean lint agent-contract-check connectorgen-validate connectorgen-surface-sync connectorgen-declaration-admission connectorgen-operation-evidence connectorgen-certification-subject connectorgen-certification-matrix connectorgen-certification-candidates connectorgen-certification-sweep connector-boundary connector-runtime-preflight connector-canon-check certify-timing github-parity-artifacts-check

# Packages covered by `lint` include declarative connector and canonical agent-contract tooling.
# Paths are filtered to existing directories so optional local trees do not hard-fail
# golangci-lint's arg parsing.
LINT_CANDIDATE_DIRS := internal/connectors/engine internal/connectors/defs internal/connectors/hooks internal/connectors/native internal/connectors/conformance internal/connectors/certify internal/connectors/boundary internal/safety internal/agentcontract cmd/connectorgen cmd/agentcontractgen cmd/certifytiming
LINT_PKGS := $(foreach d,$(LINT_CANDIDATE_DIRS),$(if $(wildcard $(d)),./$(d)/...))

fmt:
	gofmt -w cmd internal

vet:
	go vet ./...

tidy-check:
	go mod tidy
	git diff --exit-code -- go.mod go.sum

test:
	go test -timeout 20m ./...

# Emits the raw cold -json streams and a compact timing summary for only the
# certification harness and its CLI route tests. Verify invokes this target in
# a separate step so the diagnostic is visible even when the aggregate suite
# later fails or reaches its job limit. The active topology budgets 25 real
# harness calls and 92 real CLI invocations; its hosted-measurement-derived
# cap remains 3m30s.
CERTIFY_TIMING_MAX_DURATION := 3m30s

certify-timing:
	go run ./cmd/certifytiming --max-duration $(CERTIFY_TIMING_MAX_DURATION)

build:
	go build ./cmd/pm

icons-generate:
	@test -n "$$PM_ICON_REGISTRY_SOURCE" || (printf 'PM_ICON_REGISTRY_SOURCE is required\n' >&2; exit 1)
	go run ./cmd/iconregistrygen --source "$$PM_ICON_REGISTRY_SOURCE"

docs-check: build docs-check-no-build

docs-check-no-build:
	./pm docs validate --connectors-dir docs/connectors

install: build
	mkdir -p "$(INSTALL_DIR)"
	install -m 0755 pm "$(INSTALL_DIR)/pm"
	printf 'installed pm to %s\n' "$(INSTALL_DIR)/pm"

uninstall:
	rm -f "$(INSTALL_DIR)/pm"
	printf 'removed %s\n' "$(INSTALL_DIR)/pm"

smoke: build smoke-no-build

smoke-no-build:
	set -eu; \
	SMOKE_DIR=$$(mktemp -d); \
	export PM_SAMPLE_TOKEN=sample-token; \
	./pm init --root "$$SMOKE_DIR" --json >/dev/null; \
	./pm credentials add sample-local --connector sample --from-env token=PM_SAMPLE_TOKEN --root "$$SMOKE_DIR" --json >/dev/null; \
	./pm credentials add warehouse-local --connector warehouse --config path="$$SMOKE_DIR/.polymetrics/warehouse" --root "$$SMOKE_DIR" --json >/dev/null; \
	./pm credentials add outbox-local --connector outbox --config path="$$SMOKE_DIR/.polymetrics/outbox" --root "$$SMOKE_DIR" --json >/dev/null; \
	./pm connections create sample_to_warehouse --source sample:sample-local --destination warehouse:warehouse-local --stream customers --primary-key id --cursor updated_at --table sample_customers --root "$$SMOKE_DIR" --json >/dev/null; \
	./pm catalog refresh --connection sample_to_warehouse --root "$$SMOKE_DIR" --json >/dev/null; \
	./pm etl run --connection sample_to_warehouse --stream customers --root "$$SMOKE_DIR" --json >/dev/null; \
	PLAN_OUTPUT=$$(./pm reverse plan customers_to_outbox --source-table sample_customers --destination outbox:outbox-local --map id:external_id --map name:full_name --map email:email --root "$$SMOKE_DIR"); \
	PLAN_ID=$$(printf '%s\n' "$$PLAN_OUTPUT" | awk '/Created reverse plan/ {print $$4}'); \
	./pm reverse preview "$$PLAN_ID" --root "$$SMOKE_DIR" --json >/dev/null; \
	APPROVAL=$$(printf '%s\n' "$$PLAN_OUTPUT" | awk '/Approval token:/ {print $$3}'); \
	printf '%s\n' "$$APPROVAL" | ./pm reverse run "$$PLAN_ID" --approval-token-stdin --root "$$SMOKE_DIR" --json >/dev/null; \
	TABLE=$$(ls "$$SMOKE_DIR"/.polymetrics/warehouse/*/*/*/tables/sample_customers.parquet); \
	test -s "$$TABLE"; \
	test -s "$$(dirname "$$(dirname "$$TABLE")")/owner.json"; \
	test -s "$$SMOKE_DIR/.polymetrics/outbox/customers_to_outbox.jsonl"; \
	printf 'smoke ok: %s\n' "$$SMOKE_DIR"

lint:
	@command -v golangci-lint >/dev/null || (echo "golangci-lint not found — brew install golangci-lint" && exit 1)
	golangci-lint run $(LINT_PKGS)

# Validates the canonical delivery source, every referenced GSD command, and each required or
# present registered harness projection.
agent-contract-check:
	go run ./cmd/agentcontractgen check

connectorgen-validate:
	go run ./cmd/connectorgen validate internal/connectors/defs

# Fails when derivable command metadata (api_surface, flag maps_to,
# output_policy, rest.max_bytes) or the compact runtime endpoint ledger drifts.
# Regenerate with `go run ./cmd/connectorgen surface-sync`.
connectorgen-surface-sync:
	go run ./cmd/connectorgen surface-sync --check

# Source declaration admission is intentionally separate from source-lock
# retention, runtime preflight, generated operation evidence, and live proof.
# It is a provider-I/O-free check of the required independent source and
# declaration catalogs.
connectorgen-declaration-admission:
	go run ./cmd/connectorgen declaration-admission

# Fails on generated evidence drift and on any regression in the immutable,
# source-locked 100-operation cohort across the runtime, CLI, website,
# fixtures, conformance, and classification surfaces.
connectorgen-operation-evidence:
	go run ./cmd/connectorgen operation-evidence --check

# The GitHub source lock is the shared REST + GraphQL denominator. These
# hermetic checks reject a stale generated root contract, a stale combined
# ledger, a missing operation classification, or a disappearance of the
# organization-create canary before a connector surface change reaches CI.
github-parity-artifacts-check:
	node --test scripts/tests/github-combined-operation-ledger.test.mjs scripts/tests/gen-github-graphql-parity.test.mjs scripts/tests/github-source-drift.test.mjs
	node scripts/gen-github-graphql-parity.mjs --check
	node scripts/github-combined-operation-ledger.mjs --check

# The checked-in subject is deterministic repository identity. Individual live
# proof records separately bind the pm binary and build that actually ran.
connectorgen-certification-subject:
	go run ./cmd/connectorgen certification-subject --check

# Fails when the allowlisted connector certification shards drift.
# Regenerate one connector with `go run ./cmd/connectorgen certification-matrix --connector <name>`.
connectorgen-certification-matrix: connectorgen-certification-subject
	go run ./cmd/connectorgen certification-matrix --check

# Fails when direct-read candidates generated from the declared connector surface drift.
# Regenerate with `go run ./cmd/connectorgen certification-candidates --connector github`.
connectorgen-certification-candidates:
	go run ./cmd/connectorgen certification-candidates --connector github --check

# Fails when GitHub's source-derived certification candidate ledger drifts.
# Regenerate with `go run ./cmd/connectorgen certification-sweep --connector github`.
connectorgen-certification-sweep:
	go run ./cmd/connectorgen certification-sweep --connector github --check

# Structural runtime proof for every command that claims availability: implemented.
# The test calls commandrunner.Preflight rather than a copied validator so newly
# added executor kinds are covered automatically.
connector-runtime-preflight:
	go test -timeout 20m -run '^TestEveryImplementedCommandPassesRuntimePreflight$$' ./internal/connectors/commandrunner

# Keeps the binding source reports, archive markers, and required delivery
# procedure visible to a clean checkout. Runtime executability is verified by
# connector-runtime-preflight above.
connector-canon-check:
	bash scripts/tests/connector-canon.sh

connector-boundary:
	go run ./cmd/connectorgen boundary . --json

pinned-build-dependencies-check:
	./scripts/tests/pinned-build-dependencies.sh

release-workflow-check: pinned-build-dependencies-check
	./scripts/tests/homebrew-release-notify.sh
	./scripts/tests/release-target-parity.sh
	./scripts/tests/verify-release-tooling.sh
	./scripts/tests/release-size-budget.sh
	./scripts/tests/release-production-layout.sh
	./scripts/tests/release-installed-github-certification.sh

verify: fmt tidy-check vet test build docs-check smoke lint agent-contract-check connectorgen-validate connectorgen-surface-sync connectorgen-declaration-admission connectorgen-operation-evidence github-parity-artifacts-check connectorgen-certification-subject connectorgen-certification-matrix connectorgen-certification-candidates connectorgen-certification-sweep connector-boundary connector-canon-check release-workflow-check

# Opt-in local gate that overlaps independent read/build checks after the
# mutating fmt/tidy steps. CI keeps using serial `verify` for stable logs.
verify-parallel: fmt tidy-check
	$(MAKE) -j$(VERIFY_JOBS) vet test build lint agent-contract-check connectorgen-validate connectorgen-surface-sync connectorgen-declaration-admission connectorgen-operation-evidence github-parity-artifacts-check connectorgen-certification-subject connectorgen-certification-matrix connectorgen-certification-candidates connectorgen-certification-sweep connector-boundary connector-canon-check release-workflow-check
	$(MAKE) -j$(VERIFY_JOBS) docs-check-no-build smoke-no-build

perf-free: build
	./pm perf compare --iterations 50 --json

perf-runtime: build
	./pm perf compare --iterations 50 --runtime --json

runtime-doctor:
	scripts/runtime.sh doctor

runtime-up:
	scripts/runtime.sh up

runtime-down:
	scripts/runtime.sh down

runtime-reset:
	scripts/runtime.sh reset

clean:
	rm -f pm
