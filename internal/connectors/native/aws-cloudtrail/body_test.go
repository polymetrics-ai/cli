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
			name:    "cancel query event data store",
			action:  "CancelQuery",
			raw:     map[string]any{"QueryId": "11111111-1111-1111-1111-111111111111"},
			wantErr: "requires field EventDataStore",
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
		{
			name:   "start import retry mode cannot mix new import fields",
			action: "StartImport",
			raw: map[string]any{
				"ImportId":     "11111111-1111-1111-1111-111111111111",
				"Destinations": []string{"arn:aws:cloudtrail:us-east-1:123456789012:eventdatastore/import-destination"},
			},
			wantErr: "cannot combine ImportId with Destinations",
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

	channelBody, err := buildActionBody("CreateChannel", map[string]any{
		"Destinations": []any{map[string]any{
			"Location": "arn:aws:cloudtrail:us-east-1:123456789012:eventdatastore/channel-destination",
			"Type":     "EVENT_DATA_STORE",
		}},
		"Name":   "example-channel",
		"Source": "Custom",
	}, true)
	if err != nil {
		t.Fatalf("build CreateChannel body: %v", err)
	}
	if got := channelBody["Destinations"].([]any)[0].(map[string]any)["Type"]; got != "EVENT_DATA_STORE" {
		t.Fatalf("Destination.Type = %v, want EVENT_DATA_STORE", got)
	}
	if _, err := buildActionBody("CreateChannel", map[string]any{
		"Destinations": []any{map[string]any{"Location": "arn:aws:cloudtrail:us-east-1:123456789012:eventdatastore/channel-destination"}},
		"Name":         "example-channel",
		"Source":       "Custom",
	}, true); err == nil {
		t.Fatal("CreateChannel accepted a Destination without Type")
	}
	if _, err := buildActionBody("UpdateChannel", map[string]any{
		"Channel": "arn:aws:cloudtrail:us-east-1:123456789012:channel/example",
		"Destinations": []any{map[string]any{
			"Location": "arn:aws:cloudtrail:us-east-1:123456789012:eventdatastore/channel-destination",
			"Type":     "AWS::SNS::Topic",
		}},
	}, true); err == nil {
		t.Fatal("UpdateChannel accepted an invalid Destination.Type")
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

func TestBuildActionBodyEnforcesClassicTrailContracts(t *testing.T) {
	tests := []struct {
		name    string
		action  string
		raw     map[string]any
		wantErr string
	}{
		{
			name:   "create trail cloudwatch role needs log group",
			action: "CreateTrail",
			raw: map[string]any{
				"Name":                  "example-trail",
				"S3BucketName":          "example-cloudtrail-bucket",
				"CloudWatchLogsRoleArn": "arn:aws:iam::123456789012:role/CloudTrailLogs",
			},
			wantErr: "requires CloudWatchLogsLogGroupArn",
		},
		{
			name:   "update trail cloudwatch role needs log group",
			action: "UpdateTrail",
			raw: map[string]any{
				"Name":                  "example-trail",
				"CloudWatchLogsRoleArn": "arn:aws:iam::123456789012:role/CloudTrailLogs",
			},
			wantErr: "requires CloudWatchLogsLogGroupArn",
		},
		{
			name:   "create trail name minimum",
			action: "CreateTrail",
			raw: map[string]any{
				"Name":         "a",
				"S3BucketName": "example-cloudtrail-bucket",
			},
			wantErr: "length from 3 to 128",
		},
		{
			name:   "create trail s3 prefix maximum",
			action: "CreateTrail",
			raw: map[string]any{
				"Name":         "example-trail",
				"S3BucketName": "example-cloudtrail-bucket",
				"S3KeyPrefix":  strings.Repeat("x", 201),
			},
			wantErr: "length at most 200",
		},
		{
			name:   "create trail sns topic maximum",
			action: "CreateTrail",
			raw: map[string]any{
				"Name":         "example-trail",
				"S3BucketName": "example-cloudtrail-bucket",
				"SnsTopicName": strings.Repeat("x", 257),
			},
			wantErr: "length at most 256",
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

	body, err := buildActionBody("UpdateTrail", map[string]any{
		"Name": "arn:aws:cloudtrail:us-east-1:123456789012:trail/example-trail",
	}, true)
	if err != nil {
		t.Fatalf("build UpdateTrail body with trail ARN: %v", err)
	}
	if body["Name"] != "arn:aws:cloudtrail:us-east-1:123456789012:trail/example-trail" {
		t.Fatalf("UpdateTrail Name = %#v, want trail ARN", body["Name"])
	}
}

func TestBuildActionBodyValidatesTypedCloudTrailNestedPayloads(t *testing.T) {
	if _, err := buildActionBody("AddTags", map[string]any{
		"ResourceId": "arn:aws:cloudtrail:us-east-1:123456789012:trail/example",
		"TagsList":   []any{map[string]any{}},
	}, true); err == nil || !strings.Contains(err.Error(), "Tag requires field Key") {
		t.Fatalf("AddTags missing key error = %v, want closed Tag error", err)
	}
	if _, err := buildActionBody("PutInsightSelectors", map[string]any{
		"TrailName":        "example-trail",
		"InsightSelectors": []any{map[string]any{"InsightType": "UnknownInsight"}},
	}, true); err == nil || !strings.Contains(err.Error(), "InsightSelector.InsightType") {
		t.Fatalf("PutInsightSelectors invalid type error = %v, want enum error", err)
	}
	if _, err := buildActionBody("PutEventSelectors", map[string]any{
		"TrailName": "example-trail",
		"AdvancedEventSelectors": []any{map[string]any{
			"FieldSelectors": []any{map[string]any{"Field": "eventCategory", "Equals": []string{"Data"}}},
		}},
		"EventSelectors": []any{map[string]any{"ReadWriteType": "All"}},
	}, true); err == nil || !strings.Contains(err.Error(), "cannot combine AdvancedEventSelectors") {
		t.Fatalf("PutEventSelectors mixed selectors error = %v, want exclusivity error", err)
	}
	if _, err := buildActionBody("PutEventConfiguration", map[string]any{
		"EventDataStore": "arn:aws:cloudtrail:us-east-1:123456789012:eventdatastore/example",
		"MaxEventSize":   "Large",
		"ContextKeySelectors": []any{map[string]any{
			"Type": "RequestContext",
		}},
	}, true); err == nil || !strings.Contains(err.Error(), "ContextKeySelector requires field Equals") {
		t.Fatalf("PutEventConfiguration incomplete context selector error = %v, want closed selector error", err)
	}

	for _, test := range []struct {
		name   string
		action string
		raw    map[string]any
	}{
		{
			name:   "tag",
			action: "AddTags",
			raw: map[string]any{
				"ResourceId": "arn:aws:cloudtrail:us-east-1:123456789012:trail/example",
				"TagsList":   []any{map[string]any{"Key": "Environment", "Value": "preview"}},
			},
		},
		{
			name:   "insight selector",
			action: "PutInsightSelectors",
			raw: map[string]any{
				"TrailName":        "example-trail",
				"InsightSelectors": []any{map[string]any{"InsightType": "ApiCallRateInsight", "EventCategories": []string{"Management"}}},
			},
		},
		{
			name:   "advanced event selector",
			action: "PutEventSelectors",
			raw: map[string]any{
				"TrailName": "example-trail",
				"AdvancedEventSelectors": []any{map[string]any{
					"Name":           "s3-object-events",
					"FieldSelectors": []any{map[string]any{"Field": "eventCategory", "Equals": []string{"Data"}}},
				}},
			},
		},
		{
			name:   "event selector",
			action: "PutEventSelectors",
			raw: map[string]any{
				"TrailName": "example-trail",
				"EventSelectors": []any{map[string]any{
					"ReadWriteType": "All",
					"DataResources": []any{map[string]any{
						"Type":   "AWS::S3::Object",
						"Values": []string{"arn:aws:s3:::example-bucket/"},
					}},
				}},
			},
		},
		{
			name:   "aggregation configuration",
			action: "PutEventConfiguration",
			raw: map[string]any{
				"TrailName": "example-trail",
				"AggregationConfigurations": []any{map[string]any{
					"EventCategory": "Data",
					"Templates":     []string{"API_ACTIVITY"},
				}},
			},
		},
		{
			name:   "context key selector",
			action: "PutEventConfiguration",
			raw: map[string]any{
				"EventDataStore": "arn:aws:cloudtrail:us-east-1:123456789012:eventdatastore/example",
				"MaxEventSize":   "Large",
				"ContextKeySelectors": []any{map[string]any{
					"Type":   "RequestContext",
					"Equals": []string{"aws:PrincipalArn"},
				}},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := buildActionBody(test.action, test.raw, true); err != nil {
				t.Fatalf("buildActionBody(%s): %v", test.action, err)
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
