INSTALL_DIR ?= $(HOME)/.local/bin

# go.mod requires Go 1.25 and pins a patched toolchain. Allow the go command to
# fetch the matching toolchain when the ambient one is older.
export GOTOOLCHAIN ?= auto

.PHONY: fmt vet test build icons-generate docs-check install uninstall smoke verify verify-duckdb perf-free perf-runtime runtime-doctor runtime-up runtime-down runtime-reset clean

fmt:
	gofmt -w cmd internal

vet:
	go vet ./...

test:
	go test ./...

build:
	go build ./cmd/pm

icons-generate:
	@test -n "$$PM_ICON_REGISTRY_SOURCE" || (printf 'PM_ICON_REGISTRY_SOURCE is required\n' >&2; exit 1)
	go run ./cmd/iconregistrygen --source "$$PM_ICON_REGISTRY_SOURCE"

docs-check: build
	./pm docs validate --connectors-dir docs/connectors

install: build
	mkdir -p "$(INSTALL_DIR)"
	install -m 0755 pm "$(INSTALL_DIR)/pm"
	printf 'installed pm to %s\n' "$(INSTALL_DIR)/pm"

uninstall:
	rm -f "$(INSTALL_DIR)/pm"
	printf 'removed %s\n' "$(INSTALL_DIR)/pm"

smoke: build
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

verify: fmt vet test build docs-check smoke

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
