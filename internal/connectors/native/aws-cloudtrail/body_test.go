package awscloudtrail

import (
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestBuildActionBodyEnforcesDocumentedConstraints(t *testing.T) {
	tests := []struct {
		name    string
		action  string
		raw     map[string]any
		wantErr string
	}{
		{
			name:    "lookup event max results",
			action:  "LookupEvents",
			raw:     map[string]any{"MaxResults": 51},
			wantErr: "must be at most 50",
		},
		{
			name:    "lookup event time order",
			action:  "LookupEvents",
			raw:     map[string]any{"StartTime": "2026-01-02T00:00:00Z", "EndTime": "2026-01-01T00:00:00Z"},
			wantErr: "EndTime not to precede StartTime",
		},
		{
			name:    "describe query selector",
			action:  "DescribeQuery",
			raw:     map[string]any{},
			wantErr: "QueryId or QueryAlias",
		},
		{
			name:   "get insight selectors target",
			action: "GetInsightSelectors",
			raw: map[string]any{
				"EventDataStore": "arn:aws:cloudtrail:us-east-1:123456789012:eventdatastore/example",
				"TrailName":      "example-trail",
			},
			wantErr: "requires exactly one",
		},
		{
			name:    "start import mode",
			action:  "StartImport",
			raw:     map[string]any{},
			wantErr: "ImportId or Destinations and ImportSource",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := buildActionBody(test.action, test.raw, true)
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("buildActionBody(%s) error = %v, want %q", test.action, err, test.wantErr)
			}
		})
	}
}

func TestBuildActionBodyValidatesClosedNestedRequests(t *testing.T) {
	lookupBody, err := buildActionBody("LookupEvents", map[string]any{
		"LookupAttributes": []any{map[string]any{"AttributeKey": "EventName", "AttributeValue": "ConsoleLogin"}},
	}, true)
	if err != nil {
		t.Fatalf("build LookupEvents body: %v", err)
	}
	if got := lookupBody["LookupAttributes"].([]any)[0].(map[string]any)["AttributeKey"]; got != "EventName" {
		t.Fatalf("LookupAttributes key = %v, want EventName", got)
	}
	if _, err := buildActionBody("LookupEvents", map[string]any{
		"LookupAttributes": []any{map[string]any{"AttributeKey": "Unknown", "AttributeValue": "ConsoleLogin"}},
	}, true); err == nil {
		t.Fatal("LookupEvents accepted an unknown LookupAttribute key")
	}

	importBody, err := buildActionBody("StartImport", map[string]any{
		"Destinations": []string{"arn:aws:cloudtrail:us-east-1:123456789012:eventdatastore/import-destination"},
		"ImportSource": map[string]any{
			"S3": map[string]any{
				"S3BucketAccessRoleArn": "arn:aws:iam::123456789012:role/CloudTrailImport",
				"S3BucketRegion":        "us-east-1",
				"S3LocationUri":         "s3://ct-import/source/",
			},
		},
	}, true)
	if err != nil {
		t.Fatalf("build StartImport body: %v", err)
	}
	if _, ok := importBody["ImportSource"].(map[string]any)["S3"].(map[string]any); !ok {
		t.Fatalf("ImportSource = %#v, want closed S3 object", importBody["ImportSource"])
	}
	if _, err := buildActionBody("StartImport", map[string]any{
		"Destinations": []string{"arn:aws:cloudtrail:us-east-1:123456789012:eventdatastore/import-destination"},
		"ImportSource": map[string]any{"S3": map[string]any{"S3BucketRegion": "us-east-1"}},
	}, true); err == nil {
		t.Fatal("StartImport accepted an incomplete ImportSource.S3 object")
	}
}

func TestBuildActionBodyAllowsEmptyRequiredInsightSelectors(t *testing.T) {
	body, err := buildActionBody("PutInsightSelectors", map[string]any{
		"InsightSelectors": []any{},
		"TrailName":        "example-trail",
	}, true)
	if err != nil {
		t.Fatalf("build PutInsightSelectors body: %v", err)
	}
	if selectors, ok := body["InsightSelectors"].([]any); !ok || len(selectors) != 0 {
		t.Fatalf("InsightSelectors = %#v, want an empty selector list", body["InsightSelectors"])
	}
}

func TestBuildActionBodyEnforcesConditionalRules(t *testing.T) {
	tests := []struct {
		name    string
		action  string
		raw     map[string]any
		wantErr string
	}{
		{
			name:   "event data store insights need a destination to enable",
			action: "PutInsightSelectors",
			raw: map[string]any{
				"EventDataStore":   "arn:aws:cloudtrail:us-east-1:123456789012:eventdatastore/example",
				"InsightSelectors": []any{map[string]any{"InsightType": "ApiCallRateInsight"}},
			},
			wantErr: "requires InsightsDestination",
		},
		{
			name:   "event data store update needs a change",
			action: "UpdateEventDataStore",
			raw: map[string]any{
				"EventDataStore": "arn:aws:cloudtrail:us-east-1:123456789012:eventdatastore/example",
			},
			wantErr: "at least one field besides EventDataStore",
		},
		{
			name:   "fixed retention is bounded",
			action: "CreateEventDataStore",
			raw: map[string]any{
				"Name":            "example-store",
				"BillingMode":     "FIXED_RETENTION_PRICING",
				"RetentionPeriod": 2558,
			},
			wantErr: "must be at most 2557",
		},
		{
			name:   "error rate metrics need error code",
			action: "ListInsightsMetricData",
			raw: map[string]any{
				"EventName":   "ConsoleLogin",
				"EventSource": "signin.amazonaws.com",
				"InsightType": "ApiErrorRateInsight",
			},
			wantErr: "requires ErrorCode",
		},
		{
			name:   "metric period is discrete",
			action: "ListInsightsMetricData",
			raw: map[string]any{
				"EventName":   "ConsoleLogin",
				"EventSource": "signin.amazonaws.com",
				"InsightType": "ApiCallRateInsight",
				"Period":      120,
			},
			wantErr: "must be one of 60, 300, or 3600",
		},
		{
			name:   "context selectors require large events",
			action: "PutEventConfiguration",
			raw: map[string]any{
				"EventDataStore":      "arn:aws:cloudtrail:us-east-1:123456789012:eventdatastore/example",
				"MaxEventSize":        "Standard",
				"ContextKeySelectors": []any{map[string]any{"Type": "RequestContext"}},
			},
			wantErr: "requires MaxEventSize Large",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := buildActionBody(test.action, test.raw, true)
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("buildActionBody(%s) error = %v, want %q", test.action, err, test.wantErr)
			}
		})
	}
}

func TestPageSizeForActionClampsToDocumentedLimit(t *testing.T) {
	value, err := pageSizeForAction("LookupEvents", connectors.RuntimeConfig{Config: map[string]string{"page_size": "1000"}})
	if err != nil {
		t.Fatalf("pageSizeForAction: %v", err)
	}
	if value != 50 {
		t.Fatalf("pageSizeForAction = %d, want 50", value)
	}
}
