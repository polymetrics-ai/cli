package commandrunner

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestCommandRunnerPreservesNestedHeuristicContent(t *testing.T) {
	want := connectors.Record{
		"token":              "fixture-token",
		"secret":             "fixture-secret",
		"content":            "complete content",
		"body":               "complete body",
		"download_media_url": "https://media.example.test/fixture.mp4",
		"nested": map[string]any{
			"token":   "nested-token",
			"content": "nested content",
			"key":     "nested-key",
		},
		"items": []any{map[string]any{"content": "array content", "token": "array-token"}},
	}
	connector := &fakeConnector{
		surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{{
			Path:         "records list",
			Intent:       "etl",
			Availability: "implemented",
			Stream:       "records",
			RedactFields: []string{"token", "secret", "content", "body", "download_media_url", "nested", "items"},
		}}},
		readRecords: []connectors.Record{want},
	}

	var got []connectors.Record
	_, err := Run(context.Background(), connector, Request{Path: []string{"records", "list"}, Limit: 1}, func(record connectors.Record) error {
		got = append(got, record)
		return nil
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(got) != 1 || !reflect.DeepEqual(got[0], want) {
		t.Fatalf("ETL record = %#v, want exact complete record %#v", got, want)
	}
}

func TestBuildWriteCommandKeepsRecordAndPreviewEquivalent(t *testing.T) {
	connector := &fakeConnector{
		surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{{
			Path:         "content create",
			Intent:       "reverse_etl",
			Availability: "implemented",
			Write:        "create_content",
			RedactFields: []string{"token", "content", "metadata"},
			Flags: []connectors.CommandSurfaceFlag{
				{Name: "token", Type: "string", MapsTo: "record.token", Required: true},
				{Name: "content", Type: "string", MapsTo: "record.content", Required: true},
				{Name: "nested-token", Type: "string", MapsTo: "record.metadata.token", Required: true},
			},
		}}},
		manifest: connectors.Manifest{WriteActions: []connectors.WriteActionSpec{{Name: "create_content", Method: http.MethodPost, Path: "/content"}}},
	}

	command, err := BuildWriteCommand(context.Background(), connector, Request{
		Path: []string{"content", "create"},
		Flags: map[string][]string{
			"token":        {"fixture-token"},
			"content":      {"complete content"},
			"nested-token": {"nested-token"},
		},
	})
	if err != nil {
		t.Fatalf("BuildWriteCommand: %v", err)
	}
	if !reflect.DeepEqual(command.Record, command.RedactedRecord) {
		t.Fatalf("record/preview = %#v/%#v, want identical complete records", command.Record, command.RedactedRecord)
	}
}

func TestCommandRunnerReturnsOriginalExecutorErrorsAndOmitsRedactFields(t *testing.T) {
	tests := []struct {
		name   string
		run    func(*fakeConnector, error) error
		assert func(*testing.T, *fakeConnector)
	}{
		{
			name: "ETL",
			run: func(connector *fakeConnector, source error) error {
				connector.readErr = source
				_, err := Run(context.Background(), connector, Request{
					Path:  []string{"records", "list"},
					Flags: map[string][]string{"token": {"fixture-token"}, "content": {"fixture-content"}},
				}, func(connectors.Record) error { return nil })
				return err
			},
			assert: func(t *testing.T, connector *fakeConnector) {
				t.Helper()
				if connector.readReq.Query["token"] != "fixture-token" || connector.readReq.Query["content"] != "fixture-content" {
					t.Fatalf("ETL query = %#v, want complete inputs", connector.readReq.Query)
				}
			},
		},
		{
			name: "direct read",
			run: func(connector *fakeConnector, source error) error {
				connector.directReadErr = source
				_, err := Run(context.Background(), connector, Request{
					Path:  []string{"records", "get"},
					Flags: map[string][]string{"token": {"fixture-token"}, "content": {"fixture-content"}},
				}, func(connectors.Record) error { return nil })
				return err
			},
			assert: func(t *testing.T, connector *fakeConnector) {
				t.Helper()
				if len(connector.directReadReq.RedactFields) != 0 {
					t.Fatalf("direct-read RedactFields = %#v, want empty", connector.directReadReq.RedactFields)
				}
			},
		},
		{
			name: "operation direct read",
			run: func(connector *fakeConnector, source error) error {
				connector.operationDirectReadErr = source
				_, err := Run(context.Background(), connector, Request{
					Path:  []string{"records", "lookup"},
					Flags: map[string][]string{"token": {"fixture-token"}, "content": {"fixture-content"}},
				}, func(connectors.Record) error { return nil })
				return err
			},
			assert: func(t *testing.T, connector *fakeConnector) {
				t.Helper()
				if len(connector.operationDirectReadReq.RedactFields) != 0 {
					t.Fatalf("operation direct-read RedactFields = %#v, want empty", connector.operationDirectReadReq.RedactFields)
				}
			},
		},
		{
			name: "binary download",
			run: func(connector *fakeConnector, source error) error {
				connector.binaryDownloadErr = source
				_, err := Run(context.Background(), connector, Request{
					Path:     []string{"records", "download"},
					Flags:    map[string][]string{"token": {"fixture-token"}, "content": {"fixture-content"}},
					DestRoot: "fixture-output",
				}, func(connectors.Record) error { return nil })
				return err
			},
			assert: func(t *testing.T, connector *fakeConnector) {
				t.Helper()
				if len(connector.binaryDownloadReq.RedactFields) != 0 {
					t.Fatalf("binary-download RedactFields = %#v, want empty", connector.binaryDownloadReq.RedactFields)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			connector := contentErrorConnector(test.name)
			source := errors.New(test.name + " rejected token fixture-token and content fixture-content")
			got := test.run(connector, source)
			if got == nil {
				t.Fatal("Run error = nil, want original executor error")
			}
			if got != source {
				t.Fatalf("Run error = %v, want original executor error %v", got, source)
			}
			if strings.Contains(got.Error(), "***") || !strings.Contains(got.Error(), "fixture-token") || !strings.Contains(got.Error(), "fixture-content") {
				t.Fatalf("Run error = %q, want complete executor error", got.Error())
			}
			test.assert(t, connector)
		})
	}
}

func contentErrorConnector(name string) *fakeConnector {
	flags := []connectors.CommandSurfaceFlag{
		{Name: "token", Type: "string", MapsTo: "query.token", Required: true},
		{Name: "content", Type: "string", MapsTo: "query.content", Required: true},
	}
	command := connectors.CommandSurfaceCommand{
		Intent:       "etl",
		Availability: "implemented",
		Path:         "records list",
		Stream:       "records",
		Flags:        flags,
		RedactFields: []string{"token", "content"},
	}
	switch name {
	case "direct read":
		command = connectors.CommandSurfaceCommand{
			Path:         "records get",
			Intent:       "direct_read",
			Availability: "implemented",
			APISurface:   []connectors.CommandSurfaceEndpointRef{{Method: http.MethodGet, Path: "/records"}},
			OutputPolicy: "repository_contents_file_metadata",
			Flags:        flags,
			RedactFields: []string{"token", "content"},
		}
	case "operation direct read":
		command = connectors.CommandSurfaceCommand{
			Path:         "records lookup",
			Intent:       "direct_read",
			Availability: "implemented",
			Operation:    "fixture.records_lookup",
			APISurface:   []connectors.CommandSurfaceEndpointRef{{Method: http.MethodPost, Path: "/records/lookup"}},
			OutputPolicy: "repository_contents_file_metadata",
			Flags: []connectors.CommandSurfaceFlag{
				{Name: "token", Type: "string", MapsTo: "body.token", Required: true},
				{Name: "content", Type: "string", MapsTo: "body.content", Required: true},
			},
			RedactFields: []string{"token", "content"},
		}
	case "binary download":
		command = connectors.CommandSurfaceCommand{
			Path:         "records download",
			Intent:       "binary_download",
			Availability: "implemented",
			Operation:    "fixture.records_download",
			APISurface:   []connectors.CommandSurfaceEndpointRef{{Method: http.MethodGet, Path: "/records/download"}},
			Flags:        flags,
			RedactFields: []string{"token", "content"},
		}
	}
	return &fakeConnector{surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{command}}}
}
