package nativeset

import (
	"context"
	"errors"
	"testing"

	"polymetrics.ai/internal/connectors"
)

var errOptionalRuntimeForwarded = errors.New("optional runtime call forwarded")

type optionalRuntimeConnector struct{}

type coreRuntimeConnector struct{}

func (optionalRuntimeConnector) Name() string { return "optional-runtime" }

func (optionalRuntimeConnector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "optional-runtime"}
}

func (optionalRuntimeConnector) Check(context.Context, connectors.RuntimeConfig) error { return nil }

func (optionalRuntimeConnector) Catalog(context.Context, connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{Connector: "optional-runtime"}, nil
}

func (optionalRuntimeConnector) Read(context.Context, connectors.ReadRequest, func(connectors.Record) error) error {
	return nil
}

func (optionalRuntimeConnector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, nil
}

func (optionalRuntimeConnector) DirectRead(context.Context, connectors.DirectReadRequest) (connectors.DirectReadResult, error) {
	return connectors.DirectReadResult{Connector: "direct-read"}, nil
}

func (optionalRuntimeConnector) OperationDirectRead(context.Context, connectors.OperationDirectReadRequest) (connectors.DirectReadResult, error) {
	return connectors.DirectReadResult{Connector: "operation-direct-read"}, nil
}

func (optionalRuntimeConnector) OperationBinaryDownload(context.Context, connectors.OperationBinaryDownloadRequest) (connectors.OperationBinaryDownloadResult, error) {
	return connectors.OperationBinaryDownloadResult{Connector: "binary-download", Operation: "fixture"}, nil
}

func (optionalRuntimeConnector) ValidateWrite(context.Context, connectors.WriteRequest, []connectors.Record) error {
	return errOptionalRuntimeForwarded
}

func (optionalRuntimeConnector) DryRunWrite(_ context.Context, req connectors.WriteRequest, _ []connectors.Record) (connectors.WritePreview, error) {
	return connectors.WritePreview{Action: req.Action}, nil
}

func (optionalRuntimeConnector) Query(context.Context, connectors.QueryRequest) (connectors.QueryResult, error) {
	return connectors.QueryResult{Rows: []connectors.Record{{"forwarded": true}}}, nil
}

func (optionalRuntimeConnector) ReadCDC(_ context.Context, _ connectors.CDCReadRequest, emit func(connectors.CDCEvent) error) error {
	return emit(connectors.CDCEvent{Operation: "forwarded"})
}

func (optionalRuntimeConnector) InitialState(context.Context, string, connectors.RuntimeConfig) (map[string]string, error) {
	return map[string]string{"forwarded": "true"}, nil
}

func (optionalRuntimeConnector) MapSchema(_ context.Context, stream connectors.Stream) (connectors.Stream, error) {
	stream.Description = "forwarded"
	return stream, nil
}

func (optionalRuntimeConnector) LiveConformanceConfig(context.Context) (connectors.RuntimeConfig, bool, error) {
	return connectors.RuntimeConfig{Config: map[string]string{"forwarded": "true"}}, true, nil
}

func (optionalRuntimeConnector) MaterializesLocalWarehouse() bool { return true }

func (coreRuntimeConnector) Name() string { return "core-runtime" }

func (coreRuntimeConnector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: "core-runtime"}
}

func (coreRuntimeConnector) Check(context.Context, connectors.RuntimeConfig) error { return nil }

func (coreRuntimeConnector) Catalog(context.Context, connectors.RuntimeConfig) (connectors.Catalog, error) {
	return connectors.Catalog{Connector: "core-runtime"}, nil
}

func (coreRuntimeConnector) Read(context.Context, connectors.ReadRequest, func(connectors.Record) error) error {
	return nil
}

func (coreRuntimeConnector) Write(context.Context, connectors.WriteRequest, []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, nil
}

func TestFactoriesExposeDefinitions(t *testing.T) {
	want := map[string]bool{
		"alpha-vantage":             false,
		"amazon-sqs":                false,
		"apify-dataset":             false,
		"ashby":                     false,
		"aws-cloudtrail":            false,
		"babelforce":                false,
		"basecamp":                  false,
		"bing-ads":                  false,
		"bunny-inc":                 false,
		"canny":                     false,
		"copper":                    false,
		"dixa":                      false,
		"dynamodb":                  false,
		"faker":                     false,
		"fastbill":                  false,
		"feishu":                    false,
		"free-agent-connector":      false,
		"freightview":               false,
		"google-analytics-data-api": false,
		"google-calendar":           false,
		"google-classroom":          false,
		"google-pagespeed-insights": false,
		"less-annoying-crm":         false,
		"lokalise":                  false,
		"mendeley":                  false,
		"mercado-ads":               false,
		"metabase":                  false,
		"mode":                      false,
		"my-hours":                  false,
		"pocket":                    false,
		"postgres":                  false,
		"prestashop":                false,
		"rootly":                    false,
		"safetyculture":             false,
		"tally-prime":               false,
		"yahoo-finance-price":       false,
	}

	for _, factory := range Factories() {
		if factory.New == nil {
			t.Fatalf("factory %q New = nil", factory.Name)
		}
		c := factory.New()
		if c.Name() != factory.Name {
			t.Fatalf("factory %q New().Name() = %q", factory.Name, c.Name())
		}
		def, ok := connectors.DefinitionOf(c)
		if !ok {
			t.Fatalf("factory %q connector does not implement DefinitionProvider", factory.Name)
		}
		if def.Name != factory.Name {
			t.Fatalf("factory %q Definition().Name = %q", factory.Name, def.Name)
		}
		if _, tracked := want[factory.Name]; tracked {
			want[factory.Name] = true
		}
	}

	for name, seen := range want {
		if !seen {
			t.Fatalf("Factories() missing %q", name)
		}
	}
}

func TestBundleDefinitionForwardsNativeCommandSurface(t *testing.T) {
	var cloudTrail connectors.Connector
	for _, factory := range Factories() {
		if factory.Name == "aws-cloudtrail" {
			cloudTrail = factory.New()
			break
		}
	}
	if cloudTrail == nil {
		t.Fatal("aws-cloudtrail factory not found")
	}
	provider, ok := cloudTrail.(connectors.CommandSurfaceProvider)
	if !ok {
		t.Fatal("aws-cloudtrail factory does not implement connectors.CommandSurfaceProvider")
	}
	surface := provider.CommandSurface()
	if surface == nil {
		t.Fatal("aws-cloudtrail CommandSurface() = nil")
	}
	if got, want := len(surface.Commands), 60; got != want {
		t.Fatalf("aws-cloudtrail command rows = %d, want %d", got, want)
	}
}

func TestDefinitionConnectorForwardsEveryOptionalRuntimeInterface(t *testing.T) {
	wrapped := definitionConnector{Connector: optionalRuntimeConnector{}}

	t.Run("DirectReader", func(t *testing.T) {
		reader, ok := any(wrapped).(connectors.DirectReader)
		if !ok {
			t.Fatal("definitionConnector does not implement connectors.DirectReader")
		}
		result, err := reader.DirectRead(context.Background(), connectors.DirectReadRequest{})
		if err != nil || result.Connector != "direct-read" {
			t.Fatalf("DirectRead() = (%+v, %v), want forwarded result", result, err)
		}
	})
	t.Run("OperationDirectReader", func(t *testing.T) {
		reader, ok := any(wrapped).(connectors.OperationDirectReader)
		if !ok {
			t.Fatal("definitionConnector does not implement connectors.OperationDirectReader")
		}
		result, err := reader.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{})
		if err != nil || result.Connector != "operation-direct-read" {
			t.Fatalf("OperationDirectRead() = (%+v, %v), want forwarded result", result, err)
		}
	})
	t.Run("OperationBinaryDownloader", func(t *testing.T) {
		downloader, ok := any(wrapped).(connectors.OperationBinaryDownloader)
		if !ok {
			t.Fatal("definitionConnector does not implement connectors.OperationBinaryDownloader")
		}
		result, err := downloader.OperationBinaryDownload(context.Background(), connectors.OperationBinaryDownloadRequest{})
		if err != nil || result.Connector != "binary-download" {
			t.Fatalf("OperationBinaryDownload() = (%+v, %v), want forwarded result", result, err)
		}
	})
	t.Run("WriteValidator", func(t *testing.T) {
		validator, ok := any(wrapped).(connectors.WriteValidator)
		if !ok {
			t.Fatal("definitionConnector does not implement connectors.WriteValidator")
		}
		if err := validator.ValidateWrite(context.Background(), connectors.WriteRequest{}, nil); !errors.Is(err, errOptionalRuntimeForwarded) {
			t.Fatalf("ValidateWrite() error = %v, want forwarded error", err)
		}
	})
	t.Run("DryRunWriter", func(t *testing.T) {
		dryRunner, ok := any(wrapped).(connectors.DryRunWriter)
		if !ok {
			t.Fatal("definitionConnector does not implement connectors.DryRunWriter")
		}
		preview, err := dryRunner.DryRunWrite(context.Background(), connectors.WriteRequest{Action: "fixture"}, nil)
		if err != nil || preview.Action != "fixture" {
			t.Fatalf("DryRunWrite() = (%+v, %v), want forwarded preview", preview, err)
		}
	})
	t.Run("Querier", func(t *testing.T) {
		querier, ok := any(wrapped).(connectors.Querier)
		if !ok {
			t.Fatal("definitionConnector does not implement connectors.Querier")
		}
		result, err := querier.Query(context.Background(), connectors.QueryRequest{})
		if err != nil || len(result.Rows) != 1 || result.Rows[0]["forwarded"] != true {
			t.Fatalf("Query() = (%+v, %v), want forwarded result", result, err)
		}
	})
	t.Run("CDCReader", func(t *testing.T) {
		reader, ok := any(wrapped).(connectors.CDCReader)
		if !ok {
			t.Fatal("definitionConnector does not implement connectors.CDCReader")
		}
		var event connectors.CDCEvent
		if err := reader.ReadCDC(context.Background(), connectors.CDCReadRequest{}, func(got connectors.CDCEvent) error { event = got; return nil }); err != nil {
			t.Fatalf("ReadCDC() error = %v", err)
		}
		if event.Operation != "forwarded" {
			t.Fatalf("ReadCDC event = %+v, want forwarded", event)
		}
	})
	t.Run("StatefulReader", func(t *testing.T) {
		reader, ok := any(wrapped).(connectors.StatefulReader)
		if !ok {
			t.Fatal("definitionConnector does not implement connectors.StatefulReader")
		}
		state, err := reader.InitialState(context.Background(), "fixture", connectors.RuntimeConfig{})
		if err != nil || state["forwarded"] != "true" {
			t.Fatalf("InitialState() = (%+v, %v), want forwarded state", state, err)
		}
	})
	t.Run("SchemaMapper", func(t *testing.T) {
		mapper, ok := any(wrapped).(connectors.SchemaMapper)
		if !ok {
			t.Fatal("definitionConnector does not implement connectors.SchemaMapper")
		}
		stream, err := mapper.MapSchema(context.Background(), connectors.Stream{})
		if err != nil || stream.Description != "forwarded" {
			t.Fatalf("MapSchema() = (%+v, %v), want forwarded stream", stream, err)
		}
	})
	t.Run("LiveConformanceProvider", func(t *testing.T) {
		provider, ok := any(wrapped).(connectors.LiveConformanceProvider)
		if !ok {
			t.Fatal("definitionConnector does not implement connectors.LiveConformanceProvider")
		}
		config, enabled, err := provider.LiveConformanceConfig(context.Background())
		if err != nil || !enabled || config.Config["forwarded"] != "true" {
			t.Fatalf("LiveConformanceConfig() = (%+v, %t, %v), want forwarded config", config, enabled, err)
		}
	})
	t.Run("LocalWarehouseMaterializer", func(t *testing.T) {
		materializer, ok := any(wrapped).(connectors.LocalWarehouseMaterializer)
		if !ok {
			t.Fatal("definitionConnector does not implement connectors.LocalWarehouseMaterializer")
		}
		if !materializer.MaterializesLocalWarehouse() {
			t.Fatal("MaterializesLocalWarehouse() = false, want forwarded true")
		}
	})
}

func TestDefinitionConnectorRefusesAbsentOptionalRuntimeInterfaces(t *testing.T) {
	wrapped := definitionConnector{Connector: coreRuntimeConnector{}}

	assertUnsupported := func(t *testing.T, err error) {
		t.Helper()
		if !errors.Is(err, connectors.ErrUnsupportedOperation) {
			t.Fatalf("error = %v, want ErrUnsupportedOperation", err)
		}
	}

	t.Run("DirectReader", func(t *testing.T) {
		_, err := wrapped.DirectRead(context.Background(), connectors.DirectReadRequest{})
		assertUnsupported(t, err)
	})
	t.Run("OperationDirectReader", func(t *testing.T) {
		_, err := wrapped.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{})
		assertUnsupported(t, err)
	})
	t.Run("OperationBinaryDownloader", func(t *testing.T) {
		_, err := wrapped.OperationBinaryDownload(context.Background(), connectors.OperationBinaryDownloadRequest{})
		assertUnsupported(t, err)
	})
	t.Run("WriteValidator", func(t *testing.T) {
		assertUnsupported(t, wrapped.ValidateWrite(context.Background(), connectors.WriteRequest{}, nil))
	})
	t.Run("DryRunWriter", func(t *testing.T) {
		_, err := wrapped.DryRunWrite(context.Background(), connectors.WriteRequest{}, nil)
		assertUnsupported(t, err)
	})
	t.Run("Querier", func(t *testing.T) {
		_, err := wrapped.Query(context.Background(), connectors.QueryRequest{})
		assertUnsupported(t, err)
	})
	t.Run("CDCReader", func(t *testing.T) {
		assertUnsupported(t, wrapped.ReadCDC(context.Background(), connectors.CDCReadRequest{}, func(connectors.CDCEvent) error { return nil }))
	})
	t.Run("StatefulReader", func(t *testing.T) {
		_, err := wrapped.InitialState(context.Background(), "fixture", connectors.RuntimeConfig{})
		assertUnsupported(t, err)
	})
	t.Run("SchemaMapper", func(t *testing.T) {
		_, err := wrapped.MapSchema(context.Background(), connectors.Stream{})
		assertUnsupported(t, err)
	})
	t.Run("LiveConformanceProvider", func(t *testing.T) {
		_, enabled, err := wrapped.LiveConformanceConfig(context.Background())
		if err != nil || enabled {
			t.Fatalf("LiveConformanceConfig() = (_, %t, %v), want (_, false, nil)", enabled, err)
		}
	})
	t.Run("LocalWarehouseMaterializer", func(t *testing.T) {
		if wrapped.MaterializesLocalWarehouse() {
			t.Fatal("MaterializesLocalWarehouse() = true, want false")
		}
	})
}
