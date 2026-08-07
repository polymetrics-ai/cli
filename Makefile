INSTALL_DIR ?= $(HOME)/.local/bin
VERIFY_JOBS ?= 2

# go.mod requires Go 1.25 and pins a patched toolchain. Allow the go command to
# fetch the matching toolchain when the ambient one is older.
export GOTOOLCHAIN ?= auto

.PHONY: fmt vet tidy-check test build icons-generate docs-check docs-check-no-build install uninstall smoke smoke-no-build release-workflow-check verify verify-parallel verify-duckdb perf-free perf-runtime runtime-doctor runtime-up runtime-down runtime-reset clean lint agent-contract-check connectorgen-validate connectorgen-surface-sync connector-boundary certify-timing

# Packages covered by `lint` include declarative connector and canonical agent-contract tooling.
# Paths are filtered to existing directories so optional local trees do not hard-fail
# golangci-lint's arg parsing.
LINT_CANDIDATE_DIRS := internal/connectors/engine internal/connectors/defs internal/connectors/hooks internal/connectors/native internal/connectors/conformance internal/connectors/certify internal/connectors/boundary internal/agentcontract cmd/connectorgen cmd/agentcontractgen cmd/certifytiming
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
	./pm reverse run "$$PLAN_ID" --approve "$$APPROVAL" --root "$$SMOKE_DIR" --json >/dev/null; \
	TABLE=$$(ls "$$SMOKE_DIR"/.polymetrics/warehouse/*/*/*/tables/sample_customers.jsonl); \
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

connector-boundary:
	go run ./cmd/connectorgen boundary . --json

release-workflow-check:
	./scripts/tests/homebrew-release-notify.sh

verify: fmt tidy-check vet test build docs-check smoke lint agent-contract-check connectorgen-validate connectorgen-surface-sync connector-boundary release-workflow-check

# Opt-in local gate that overlaps independent read/build checks after the
# mutating fmt/tidy steps. CI keeps using serial `verify` for stable logs.
verify-parallel: fmt tidy-check
	$(MAKE) -j$(VERIFY_JOBS) vet test build lint agent-contract-check connectorgen-validate connectorgen-surface-sync connector-boundary release-workflow-check
	$(MAKE) -j$(VERIFY_JOBS) docs-check-no-build smoke-no-build

verify-duckdb:
	CGO_ENABLED=1 go build -tags duckdb ./cmd/pm
	CGO_ENABLED=1 go test -tags duckdb ./...

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
