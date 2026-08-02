package awscloudtrail

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

var cloudTrailJSONTimestampFields = map[string][]string{
	"get_dashboard": {
		"CreatedTimestamp",
		"UpdatedTimestamp",
	},
	"get_event_data_store": {
		"CreatedTimestamp",
		"UpdatedTimestamp",
	},
	"get_import": {
		"CreatedTimestamp",
		"EndEventTime",
		"StartEventTime",
		"UpdatedTimestamp",
	},
	"get_trail_status": {
		"LatestCloudWatchLogsDeliveryTime",
		"LatestDeliveryTime",
		"LatestDigestDeliveryTime",
		"LatestNotificationTime",
		"StartLoggingTime",
		"StopLoggingTime",
	},
	"list_event_data_stores": {
		"CreatedTimestamp",
		"UpdatedTimestamp",
	},
	"list_import_failures": {
		"LastUpdatedTime",
	},
	"list_imports": {
		"CreatedTimestamp",
		"EndEventTime",
		"StartEventTime",
		"UpdatedTimestamp",
	},
	"list_public_keys": {
		"ValidityStartTime",
		"ValidityEndTime",
	},
}

func TestCloudTrailTimestampSchemasUseJSONNumbers(t *testing.T) {
	for stream, fields := range cloudTrailJSONTimestampFields {
		stream, fields := stream, fields
		t.Run(stream, func(t *testing.T) {
			schema := readCloudTrailSchemaDoc(t, stream)
			for _, field := range fields {
				prop, ok := schema.Properties[field]
				if !ok {
					t.Fatalf("schema %s missing timestamp field %s", stream, field)
				}
				types, ok := schemaTypeList(prop.Type)
				if !ok {
					t.Fatalf("schema %s.%s has invalid type shape %#v", stream, field, prop.Type)
				}
				if !hasSchemaType(types, "number") || !hasSchemaType(types, "null") || hasSchemaType(types, "string") || hasSchemaType(types, "integer") {
					t.Fatalf("schema %s.%s type = %v, want nullable JSON number", stream, field, types)
				}
			}
		})
	}
}

func TestCloudTrailTimestampFixturesUseFractionalJSONNumbers(t *testing.T) {
	for stream, fields := range cloudTrailJSONTimestampFields {
		stream, fields := stream, fields
		t.Run(stream, func(t *testing.T) {
			body := readCloudTrailFixtureBody(t, stream)
			for _, field := range fields {
				values := collectCloudTrailFieldValues(body, field, nil)
				if len(values) == 0 {
					t.Fatalf("fixture %s missing timestamp field %s", stream, field)
				}
				for _, value := range values {
					number, ok := value.(json.Number)
					if !ok {
						t.Fatalf("fixture %s.%s = %#v (%T), want JSON number", stream, field, value, value)
					}
					if _, err := number.Float64(); err != nil {
						t.Fatalf("fixture %s.%s = %q, want valid JSON number: %v", stream, field, number.String(), err)
					}
					if !strings.Contains(number.String(), ".") {
						t.Fatalf("fixture %s.%s = %q, want fractional JSON number", stream, field, number.String())
					}
				}
			}
		})
	}
}

func TestNativeCloudTrailReadPreservesFractionalTimestamp(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		action := strings.TrimPrefix(r.Header.Get("X-Amz-Target"), cloudTrailTargetPrefix)
		if action != "GetDashboard" {
			t.Fatalf("target = %s, want GetDashboard", action)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"DashboardArn":"arn:aws:cloudtrail:us-east-1:123456789012:dashboard/fixture-dashboard","CreatedTimestamp":1767225600.25}`))
	}))
	defer srv.Close()

	c := Connector{Client: srv.Client()}
	var records []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{
		Stream: "get_dashboard",
		Config: fixtureRuntimeConfig(srv.URL),
		Query:  map[string]string{"DashboardId": "arn:aws:cloudtrail:us-east-1:123456789012:dashboard/fixture-dashboard"},
	}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("records = %#v, want one record", records)
	}
	number, ok := records[0]["CreatedTimestamp"].(json.Number)
	if !ok || number.String() != "1767225600.25" {
		t.Fatalf("CreatedTimestamp = %#v (%T), want preserved JSON number", records[0]["CreatedTimestamp"], records[0]["CreatedTimestamp"])
	}
}

type cloudTrailSchemaDoc struct {
	Properties map[string]struct {
		Type any `json:"type"`
	} `json:"properties"`
}

func readCloudTrailSchemaDoc(t *testing.T, stream string) cloudTrailSchemaDoc {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("../../defs/aws-cloudtrail/schemas", stream+".json"))
	if err != nil {
		t.Fatalf("read schema %s: %v", stream, err)
	}
	var schema cloudTrailSchemaDoc
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("parse schema %s: %v", stream, err)
	}
	return schema
}

func readCloudTrailFixtureBody(t *testing.T, stream string) any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("../../defs/aws-cloudtrail/fixtures/streams", stream, "page_1.json"))
	if err != nil {
		t.Fatalf("read fixture %s: %v", stream, err)
	}
	var page struct {
		Response struct {
			Body any `json:"body"`
		} `json:"response"`
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&page); err != nil {
		t.Fatalf("parse fixture %s: %v", stream, err)
	}
	return page.Response.Body
}

func schemaTypeList(value any) ([]string, bool) {
	switch typed := value.(type) {
	case string:
		return []string{typed}, true
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			name, ok := item.(string)
			if !ok {
				return nil, false
			}
			out = append(out, name)
		}
		return out, true
	default:
		return nil, false
	}
}

func hasSchemaType(types []string, want string) bool {
	for _, typ := range types {
		if typ == want {
			return true
		}
	}
	return false
}

func collectCloudTrailFieldValues(value any, field string, out []any) []any {
	switch typed := value.(type) {
	case map[string]any:
		if item, ok := typed[field]; ok {
			out = append(out, item)
		}
		for _, item := range typed {
			out = collectCloudTrailFieldValues(item, field, out)
		}
	case []any:
		for _, item := range typed {
			out = collectCloudTrailFieldValues(item, field, out)
		}
	}
	return out
}
