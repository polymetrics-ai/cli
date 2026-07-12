INSTALL_DIR ?= $(HOME)/.local/bin
VERIFY_JOBS ?= 2

# go.mod requires Go 1.25 and pins a patched toolchain. Allow the go command to
# fetch the matching toolchain when the ambient one is older.
export GOTOOLCHAIN ?= auto

.PHONY: fmt vet tidy-check test build icons-generate docs-check docs-check-no-build install uninstall smoke smoke-no-build verify verify-parallel verify-duckdb perf-free perf-runtime runtime-doctor runtime-up runtime-down runtime-reset clean lint connectorgen-validate agent-loop-test

# Packages covered by `lint`: the declarative connector architecture packages.
# Paths are filtered to existing directories so optional local trees do not
# hard-fail golangci-lint's arg parsing.
LINT_CANDIDATE_DIRS := internal/connectors/engine internal/connectors/defs internal/connectors/hooks internal/connectors/native internal/connectors/conformance internal/connectors/certify cmd/connectorgen
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
	APPROVAL=$$(printf '%s\n' "$$PLAN_OUTPUT" | awk '/Approval token:/ {print $$3}'); \
	./pm reverse run "$$PLAN_ID" --approve "$$APPROVAL" --root "$$SMOKE_DIR" --json >/dev/null; \
	test -s "$$SMOKE_DIR/.polymetrics/warehouse/sample_customers.jsonl"; \
	test -s "$$SMOKE_DIR/.polymetrics/outbox/customers_to_outbox.jsonl"; \
	printf 'smoke ok: %s\n' "$$SMOKE_DIR"

lint:
	@command -v golangci-lint >/dev/null || (echo "golangci-lint not found — brew install golangci-lint" && exit 1)
	golangci-lint run $(LINT_PKGS)

connectorgen-validate:
	go run ./cmd/connectorgen validate internal/connectors/defs

agent-loop-test:
	go test ./internal/agentloop/... -count=1
	go test -race ./internal/agentloop/... -count=1
	go test ./cmd/loopctl/... -count=1
	bash scripts/tests/auto-loop-control.sh

verify: fmt tidy-check vet test build docs-check smoke lint connectorgen-validate agent-loop-test

# Opt-in local gate that overlaps independent read/build checks after the
# mutating fmt/tidy steps. CI keeps using serial `verify` for stable logs.
verify-parallel: fmt tidy-check
	$(MAKE) -j$(VERIFY_JOBS) vet test build lint connectorgen-validate
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
