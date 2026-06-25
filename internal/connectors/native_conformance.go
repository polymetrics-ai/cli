package connectors

import (
	"context"
	"fmt"
)

type NativeConformanceTest struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Error  string `json:"error,omitempty"`
}

type NativeConformanceReport struct {
	Slug           string                  `json:"slug"`
	Name           string                  `json:"name"`
	Type           ConnectorType           `json:"type"`
	RuntimeKind    RuntimeKind             `json:"runtime_kind"`
	FixtureBacked  bool                    `json:"fixture_backed"`
	Passed         bool                    `json:"passed"`
	Tests          []NativeConformanceTest `json:"tests"`
	BenchmarkHooks []string                `json:"benchmark_hooks"`
}

func NativeConformanceReports(ctx context.Context, defs []ConnectorDefinition) []NativeConformanceReport {
	reports := make([]NativeConformanceReport, 0, len(defs))
	for _, def := range defs {
		reports = append(reports, NativeConformanceForDefinition(ctx, def))
	}
	return reports
}

func NativeConformanceForDefinition(ctx context.Context, def ConnectorDefinition) NativeConformanceReport {
	def = normalizeConnectorDefinition(def)
	connector := NewNativeCatalogConnector(def)
	fixtureCaps := enabledRuntimeCapabilitiesForDefinition(def)
	report := NativeConformanceReport{
		Slug:           def.Slug,
		Name:           def.Name,
		Type:           def.Type,
		RuntimeKind:    def.RuntimeKind,
		FixtureBacked:  true,
		Passed:         true,
		BenchmarkHooks: nativeBenchmarkHooks(def),
	}
	add := func(name string, err error) {
		test := NativeConformanceTest{Name: name, Passed: err == nil}
		if err != nil {
			test.Error = err.Error()
			report.Passed = false
		}
		report.Tests = append(report.Tests, test)
	}

	cfg := RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	add("spec", validateNativeSpec(def))
	add("check", connector.Check(ctx, cfg))
	catalog, err := connector.Catalog(ctx, cfg)
	add("catalog", err)
	if err == nil && len(catalog.Streams) == 0 {
		add("catalog_nonempty", fmt.Errorf("native connector %s returned no streams", def.Slug))
	}
	if fixtureCaps.Read {
		stream := ""
		if len(catalog.Streams) > 0 {
			stream = catalog.Streams[0].Name
		}
		count := 0
		add("read_fixture", connector.Read(ctx, ReadRequest{Stream: stream, Config: cfg}, func(record Record) error {
			count++
			if record["connector_slug"] != def.Slug {
				return fmt.Errorf("record connector_slug = %v, want %s", record["connector_slug"], def.Slug)
			}
			return nil
		}))
		if count == 0 {
			add("read_nonempty", fmt.Errorf("native connector %s emitted no fixture records", def.Slug))
		}
		add("state_checkpoint", validateNativeState(ctx, connector))
	}
	if fixtureCaps.Write {
		records := []Record{connector.fixtureRecord("records")}
		add("write_validate", connector.ValidateWrite(ctx, WriteRequest{Action: "upsert", Config: cfg}, records))
		_, err := connector.DryRunWrite(ctx, WriteRequest{Action: "upsert", Config: cfg}, records)
		add("write_dry_run", err)
		result, err := connector.Write(ctx, WriteRequest{Action: "upsert", Config: cfg}, records)
		if err == nil && result.RecordsWritten != len(records) {
			err = fmt.Errorf("records_written = %d, want %d", result.RecordsWritten, len(records))
		}
		add("write_fixture", err)
	}
	if fixtureCaps.Query {
		result, err := connector.Query(ctx, QueryRequest{SQL: "select * from records", Limit: 1, Config: cfg})
		if err == nil && len(result.Rows) != 1 {
			err = fmt.Errorf("query rows = %d, want 1", len(result.Rows))
		}
		add("query_safety", err)
	}
	if CDCPlanForDefinition(def).Supported {
		count := 0
		add("cdc_fixture", connector.ReadCDC(ctx, CDCReadRequest{Stream: "records", Config: cfg}, func(event CDCEvent) error {
			count++
			if event.Operation == "" {
				return fmt.Errorf("cdc event missing operation")
			}
			return nil
		}))
		if count == 0 {
			add("cdc_nonempty", fmt.Errorf("native connector %s emitted no cdc events", def.Slug))
		}
	}
	add("docs_skill", ValidateConnectorDefinitionGuide(def))
	add("secret_redaction", validateNativeRedaction(def))
	return report
}

func validateNativeSpec(def ConnectorDefinition) error {
	if def.Slug == "" || def.Name == "" || def.Type == "" || def.RuntimeKind == "" {
		return fmt.Errorf("native spec missing required fields")
	}
	if def.DocumentationURL == "" {
		return fmt.Errorf("native spec missing documentation URL")
	}
	switch def.ImplementationStatus {
	case ImplementationEnabled, ImplementationPlannedNativePort, ImplementationUnsupported:
	default:
		return fmt.Errorf("native spec has unknown implementation_status %q", def.ImplementationStatus)
	}
	return nil
}

func validateNativeState(ctx context.Context, connector NativeCatalogConnector) error {
	state, err := connector.InitialState(ctx, "records", RuntimeConfig{})
	if err != nil {
		return err
	}
	if _, ok := state["snapshot_completed"]; !ok {
		return fmt.Errorf("initial state missing snapshot_completed")
	}
	return nil
}

func validateNativeRedaction(def ConnectorDefinition) error {
	manual := RenderConnectorDefinitionManual(def)
	for _, forbidden := range []string{"ghp_", "secret-token", "private_key-----"} {
		if containsCaseInsensitive(manual, forbidden) {
			return fmt.Errorf("manual contains secret-like value %q", forbidden)
		}
	}
	return nil
}

func nativeBenchmarkHooks(def ConnectorDefinition) []string {
	hooks := []string{"bench_check", "bench_catalog"}
	if def.RuntimeCapabilities.Read {
		hooks = append(hooks, "bench_read_fixture")
	}
	if def.RuntimeCapabilities.Write {
		hooks = append(hooks, "bench_write_fixture")
	}
	if def.RuntimeCapabilities.Query {
		hooks = append(hooks, "bench_query_fixture")
	}
	if CDCPlanForDefinition(def).Supported {
		hooks = append(hooks, "bench_cdc_fixture")
	}
	return hooks
}

func containsCaseInsensitive(value, substr string) bool {
	return len(substr) == 0 || stringsContainsFold(value, substr)
}

func stringsContainsFold(value, substr string) bool {
	valueRunes := []rune(value)
	substrRunes := []rune(substr)
	if len(substrRunes) > len(valueRunes) {
		return false
	}
	for i := 0; i <= len(valueRunes)-len(substrRunes); i++ {
		matched := true
		for j := range substrRunes {
			a := valueRunes[i+j]
			b := substrRunes[j]
			if a == b {
				continue
			}
			if 'A' <= a && a <= 'Z' {
				a += 'a' - 'A'
			}
			if 'A' <= b && b <= 'Z' {
				b += 'a' - 'A'
			}
			if a != b {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
}
