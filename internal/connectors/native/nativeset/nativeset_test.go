package nativeset

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/connectors"
	googlecalendardefs "polymetrics.ai/internal/connectors/defs/google-calendar"
	"polymetrics.ai/internal/connectors/engine"
	googlecalendar "polymetrics.ai/internal/connectors/native/google-calendar"
)

func TestFixtureModeEngineConnectorIsNetworkFree(t *testing.T) {
	var requests atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	bundle := loadPromotedBundle("google-calendar")
	fixtures, err := googlecalendardefs.Fixtures()
	if err != nil {
		t.Fatalf("Fixtures: %v", err)
	}
	bundle.Fixtures = fixtures
	bundle.HTTP.URL = srv.URL
	bundle.HTTP.Auth = nil
	connector := &fixtureModeEngineConnector{
		Connector: engine.New(bundle, nil),
		fixture:   googlecalendar.New(),
	}
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	if err := connector.Check(context.Background(), cfg); err != nil {
		t.Fatalf("Check(fixture): %v", err)
	}
	assertFixtureModeReadCatalog(t, connector, cfg)

	writeRequest := connectors.WriteRequest{Action: "insert_calendar", Config: cfg}
	writeRecords := []connectors.Record{{"summary": "Fixture calendar"}}
	if err := connector.ValidateWrite(context.Background(), writeRequest, writeRecords); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("ValidateWrite(fixture) error = %v, want ErrUnsupportedOperation", err)
	}
	if _, err := connector.DryRunWrite(context.Background(), writeRequest, writeRecords); !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("DryRunWrite(fixture) error = %v, want ErrUnsupportedOperation", err)
	}
	result, err := connector.Write(context.Background(), writeRequest, writeRecords)
	if !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("Write(fixture) error = %v, want ErrUnsupportedOperation", err)
	}
	if result.RecordsFailed != len(writeRecords) {
		t.Fatalf("Write(fixture) failed records = %d, want %d", result.RecordsFailed, len(writeRecords))
	}

	_, err = connector.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{
		Operation: "google-calendar.freebusy.query",
		Config:    cfg,
		Body: map[string]any{
			"timeMin": "2030-01-01T00:00:00Z",
			"timeMax": "2030-01-02T00:00:00Z",
			"items":   []any{map[string]any{"id": "primary"}},
		},
	})
	if !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("OperationDirectRead(fixture) error = %v, want ErrUnsupportedOperation", err)
	}
	if got := requests.Load(); got != 0 {
		t.Fatalf("fixture mode made %d network request(s)", got)
	}
}

func TestGoogleCalendarFactoryPreservesFixtureMode(t *testing.T) {
	var connector connectors.Connector
	for _, factory := range Factories() {
		if factory.Name == "google-calendar" {
			connector = factory.New()
			break
		}
	}
	if connector == nil {
		t.Fatal("google-calendar factory not found")
	}

	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}
	if err := connector.Check(context.Background(), cfg); err != nil {
		t.Fatalf("factory Check(fixture): %v", err)
	}
	assertFixtureModeReadCatalog(t, connector, cfg)
	result, err := connector.Write(context.Background(), connectors.WriteRequest{Action: "insert_calendar", Config: cfg}, []connectors.Record{{"summary": "Fixture calendar"}})
	if !errors.Is(err, connectors.ErrUnsupportedOperation) {
		t.Fatalf("factory Write(fixture) error = %v, want ErrUnsupportedOperation", err)
	}
	if result.RecordsFailed != 1 {
		t.Fatalf("factory Write(fixture) failed records = %d, want 1", result.RecordsFailed)
	}
}

func assertFixtureModeReadCatalog(t *testing.T, connector connectors.Connector, cfg connectors.RuntimeConfig) {
	t.Helper()
	catalog, err := connector.Catalog(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Catalog(fixture): %v", err)
	}
	if len(catalog.Streams) != 11 {
		t.Fatalf("Catalog(fixture) streams = %d, want 11", len(catalog.Streams))
	}
	for _, stream := range catalog.Streams {
		t.Run(stream.Name, func(t *testing.T) {
			records := 0
			err := connector.Read(context.Background(), connectors.ReadRequest{Stream: stream.Name, Config: cfg}, func(connectors.Record) error {
				records++
				return nil
			})
			if err != nil {
				t.Fatalf("Read(fixture): %v", err)
			}
			if records == 0 {
				t.Fatal("Read(fixture) emitted no records")
			}
		})
	}
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
		"mysql":                     false,
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

func TestDefinitionConnectorForwardsDeclaredConfigurationConstraints(t *testing.T) {
	spec, err := engine.CompileSchema(json.RawMessage(`{
		"type": "object",
		"properties": {
			"environment": {"type": "string", "enum": ["production", "sandbox"]}
		}
	}`))
	if err != nil {
		t.Fatalf("CompileSchema() error = %v", err)
	}

	wrapped := definitionConnector{
		Connector: connectors.Sample{},
		base:      engine.NewBase(engine.Bundle{Spec: spec}),
	}
	validator, ok := any(wrapped).(connectors.ConfigurationConstraintValidator)
	if !ok {
		t.Fatal("definitionConnector does not expose ConfigurationConstraintValidator")
	}
	if !validator.HasConfigurationConstraints() {
		t.Fatal("HasConfigurationConstraints() = false, want true for the wrapped bundle")
	}
	if err := connectors.ValidateConfiguration(wrapped, map[string]string{"environment": "preview"}); err == nil {
		t.Fatal("ValidateConfiguration() error = nil, want wrapped enum rejection")
	}
}
