package awscloudtrail

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func TestOperationLedgerCounts(t *testing.T) {
	bundle, err := engine.Load(os.DirFS("../../defs"), "aws-cloudtrail")
	if err != nil {
		t.Fatalf("load bundle: %v", err)
	}
	if got, want := len(bundle.Surface.Endpoints), 60; got != want {
		t.Fatalf("api_surface rows = %d, want %d", got, want)
	}
	if got, want := len(bundle.Streams), 19; got != want {
		t.Fatalf("streams = %d, want %d", got, want)
	}
	if got, want := len(bundle.Operations), 0; got != want {
		t.Fatalf("implemented direct operations = %d, want %d", got, want)
	}
	if got, want := len(bundle.Writes), 0; got != want {
		t.Fatalf("implemented write actions = %d, want %d", got, want)
	}
	blocked := 0
	coveredStreams := 0
	for _, endpoint := range bundle.Surface.Endpoints {
		if endpoint.CoveredBy != nil && endpoint.CoveredBy.Stream != "" {
			coveredStreams++
		}
		if endpoint.Operation != nil && endpoint.Operation.Status == "blocked" {
			blocked++
		}
	}
	if got, want := coveredStreams, 19; got != want {
		t.Fatalf("stream-covered operations = %d, want %d", got, want)
	}
	if got, want := blocked, 41; got != want {
		t.Fatalf("blocked/planned operations = %d, want %d", got, want)
	}
}

func TestPublishedStreamsHaveClosedRuntimeDispatch(t *testing.T) {
	published := map[string]bool{}
	for _, stream := range cloudTrailPublishedStreams {
		published[stream] = true
		action, ok := cloudTrailStreamActions[stream]
		if !ok {
			t.Fatalf("published stream %s missing action", stream)
		}
		_, err := buildActionBodyFromStrings(action, nil, true)
		if err != nil && !cloudTrailCanDeriveActionBody(action) {
			t.Fatalf("published stream %s action %s has no default request body or derived body: %v", stream, action, err)
		}
	}
	if len(cloudTrailStreamActions) != len(cloudTrailPublishedStreams) {
		t.Fatalf("stream action map has %d entries, want %d", len(cloudTrailStreamActions), len(cloudTrailPublishedStreams))
	}
	for stream := range cloudTrailStreamActions {
		if !published[stream] {
			t.Fatalf("unpublished stream %s is dispatchable", stream)
		}
	}
}

func TestGetInsightSelectorsUsesTypedTrailNameBody(t *testing.T) {
	if _, err := buildActionBodyFromStrings("GetInsightSelectors", nil, true); err == nil {
		t.Fatal("GetInsightSelectors unexpectedly accepted an empty request body")
	} else if !strings.Contains(err.Error(), "EventDataStore or TrailName") {
		t.Fatalf("GetInsightSelectors error = %v, want alternate request field requirement", err)
	}
	body, err := buildActionBodyFromStrings("GetInsightSelectors", map[string]string{"TrailName": "trail-fixture"}, true)
	if err != nil {
		t.Fatalf("GetInsightSelectors with TrailName: %v", err)
	}
	if got := body["TrailName"]; got != "trail-fixture" {
		t.Fatalf("TrailName = %#v, want fixture", got)
	}
}

func TestOperationLedgerNoRawEscapes(t *testing.T) {
	bundle, err := engine.Load(os.DirFS("../../defs"), "aws-cloudtrail")
	if err != nil {
		t.Fatalf("load bundle: %v", err)
	}
	for _, stream := range bundle.Streams {
		if stream.Path != "/" || stream.Method != http.MethodPost {
			t.Fatalf("stream %s request = %s %s, want fixed POST /", stream.Name, stream.Method, stream.Path)
		}
	}
	for _, endpoint := range bundle.Surface.Endpoints {
		if endpoint.CoveredBy != nil && endpoint.CoveredBy.Stream != "" && endpoint.Method != "READ_POST" {
			t.Fatalf("stream ledger endpoint %s method = %q, want READ_POST logical read-over-POST marker", endpoint.Path, endpoint.Method)
		}
	}
	for _, endpoint := range bundle.Surface.Endpoints {
		if endpoint.Operation == nil {
			continue
		}
		if endpoint.Operation.Status != "blocked" || !endpoint.Operation.BlockedByDefault {
			t.Fatalf("blocked endpoint %+v is not blocked by default", endpoint)
		}
	}
}

func TestNativeCloudTrailCheckUsesImplementedTarget(t *testing.T) {
	var gotTarget string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTarget = r.Header.Get("X-Amz-Target")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"trailList":[]}`))
	}))
	defer srv.Close()

	c := Connector{Client: srv.Client()}
	if err := c.Check(context.Background(), fixtureRuntimeConfig(srv.URL)); err != nil {
		t.Fatalf("Check: %v", err)
	}
	if got, want := gotTarget, cloudTrailTarget("DescribeTrails"); got != want {
		t.Fatalf("X-Amz-Target = %q, want %q", got, want)
	}
	if len(gotBody) != 0 {
		t.Fatalf("body = %#v, want empty DescribeTrails body by default", gotBody)
	}
}

func TestNativeCloudTrailReadDispatchesStreamTarget(t *testing.T) {
	var gotTarget string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTarget = r.Header.Get("X-Amz-Target")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"trailList":[{"Name":"trail-fixture"}]}`))
	}))
	defer srv.Close()

	c := Connector{Client: srv.Client()}
	var records []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{
		Stream: "describe_trails",
		Config: fixtureRuntimeConfig(srv.URL),
	}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	wantTarget := cloudTrailTarget("DescribeTrails")
	if gotTarget != wantTarget {
		t.Fatalf("X-Amz-Target = %q, want %q", gotTarget, wantTarget)
	}
	if len(records) != 1 || records[0]["Name"] != "trail-fixture" {
		t.Fatalf("records = %#v", records)
	}
	if len(gotBody) != 0 {
		t.Fatalf("body = %#v, want empty DescribeTrails body by default", gotBody)
	}
}

func TestNativeCloudTrailReadDerivesRequiredStreamFields(t *testing.T) {
	tests := []struct {
		name         string
		stream       string
		detailAction string
		field        string
		response     string
	}{
		{
			name:         "trail status uses discovered trail name",
			stream:       "get_trail_status",
			detailAction: "GetTrailStatus",
			field:        "Name",
			response:     `{"IsLogging":true}`,
		},
		{
			name:         "insight selectors use discovered trail name",
			stream:       "get_insight_selectors",
			detailAction: "GetInsightSelectors",
			field:        "TrailName",
			response:     `{"InsightSelectors":[{"InsightType":"ApiCallRateInsight"}]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var targets []string
			var detailBody map[string]any
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				action := strings.TrimPrefix(r.Header.Get("X-Amz-Target"), cloudTrailTargetPrefix)
				targets = append(targets, action)
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				w.Header().Set("Content-Type", "application/json")
				switch action {
				case "DescribeTrails":
					if len(body) != 0 {
						t.Fatalf("DescribeTrails body = %#v, want empty discovery request", body)
					}
					_, _ = w.Write([]byte(`{"trailList":[{"Name":"trail-fixture","TrailARN":"arn:aws:cloudtrail:us-east-1:123456789012:trail/fixture-trail"}]}`))
				case tt.detailAction:
					detailBody = body
					_, _ = w.Write([]byte(tt.response))
				default:
					t.Fatalf("unexpected target %s", action)
				}
			}))
			defer srv.Close()

			c := Connector{Client: srv.Client()}
			var records []connectors.Record
			err := c.Read(context.Background(), connectors.ReadRequest{Stream: tt.stream, Config: fixtureRuntimeConfig(srv.URL)}, func(record connectors.Record) error {
				records = append(records, record)
				return nil
			})
			if err != nil {
				t.Fatalf("Read: %v", err)
			}
			if got, want := strings.Join(targets, ","), "DescribeTrails,"+tt.detailAction; got != want {
				t.Fatalf("targets = %s, want %s", got, want)
			}
			if got := detailBody[tt.field]; got != "trail-fixture" {
				t.Fatalf("%s body = %#v, want trail-fixture", tt.detailAction, detailBody)
			}
			if len(records) != 1 || records[0]["operation"] != tt.detailAction {
				t.Fatalf("records = %#v", records)
			}
		})
	}
}

func TestNativeCloudTrailReadStampsDerivedRequestIdentity(t *testing.T) {
	t.Run("trail status uses request name", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			action := strings.TrimPrefix(r.Header.Get("X-Amz-Target"), cloudTrailTargetPrefix)
			w.Header().Set("Content-Type", "application/json")
			switch action {
			case "DescribeTrails":
				_, _ = w.Write([]byte(`{"trailList":[{"Name":"trail-fixture"}]}`))
			case "GetTrailStatus":
				_, _ = w.Write([]byte(`{"IsLogging":true}`))
			default:
				t.Fatalf("unexpected target %s", action)
			}
		}))
		defer srv.Close()

		c := Connector{Client: srv.Client()}
		var records []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: "get_trail_status", Config: fixtureRuntimeConfig(srv.URL)}, func(record connectors.Record) error {
			records = append(records, record)
			return nil
		})
		if err != nil {
			t.Fatalf("Read: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("records = %#v, want 1", records)
		}
		if got, want := records[0]["pm_record_id"], "trail-fixture"; got != want {
			t.Fatalf("pm_record_id = %#v, want %q", got, want)
		}
	})

	t.Run("resource policy uses response arn", func(t *testing.T) {
		resourceArn := "arn:aws:cloudtrail:us-east-1:123456789012:eventdatastore/fixture-store"
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			action := strings.TrimPrefix(r.Header.Get("X-Amz-Target"), cloudTrailTargetPrefix)
			w.Header().Set("Content-Type", "application/json")
			switch action {
			case "ListChannels":
				_, _ = w.Write([]byte(`{"Channels":[]}`))
			case "ListDashboards":
				_, _ = w.Write([]byte(`{"Dashboards":[]}`))
			case "ListEventDataStores":
				_, _ = w.Write([]byte(`{"EventDataStores":[{"EventDataStoreArn":"` + resourceArn + `"}]}`))
			case "GetResourcePolicy":
				_, _ = w.Write([]byte(`{"ResourceArn":"` + resourceArn + `","ResourcePolicy":"{}"}`))
			default:
				t.Fatalf("unexpected target %s", action)
			}
		}))
		defer srv.Close()

		c := Connector{Client: srv.Client()}
		var records []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: "get_resource_policy", Config: fixtureRuntimeConfig(srv.URL)}, func(record connectors.Record) error {
			records = append(records, record)
			return nil
		})
		if err != nil {
			t.Fatalf("Read: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("records = %#v, want 1", records)
		}
		if got := records[0]["pm_record_id"]; got != resourceArn {
			t.Fatalf("pm_record_id = %#v, want %q", got, resourceArn)
		}
	})

	t.Run("import failures include request import id", func(t *testing.T) {
		importIDs := []string{"11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222"}
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			action := strings.TrimPrefix(r.Header.Get("X-Amz-Target"), cloudTrailTargetPrefix)
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			switch action {
			case "ListImports":
				_, _ = w.Write([]byte(`{"Imports":[{"ImportId":"` + importIDs[0] + `"},{"ImportId":"` + importIDs[1] + `"}]}`))
			case "ListImportFailures":
				if body["ImportId"] != importIDs[0] && body["ImportId"] != importIDs[1] {
					t.Fatalf("ImportId body = %#v", body)
				}
				_, _ = w.Write([]byte(`{"Failures":[{"Location":"s3://example/failure","Status":"FAILED"}]}`))
			default:
				t.Fatalf("unexpected target %s", action)
			}
		}))
		defer srv.Close()

		c := Connector{Client: srv.Client()}
		var records []connectors.Record
		err := c.Read(context.Background(), connectors.ReadRequest{Stream: "list_import_failures", Config: fixtureRuntimeConfig(srv.URL)}, func(record connectors.Record) error {
			records = append(records, record)
			return nil
		})
		if err != nil {
			t.Fatalf("Read: %v", err)
		}
		if len(records) != 2 {
			t.Fatalf("records = %#v, want 2", records)
		}
		wantIDs := map[string]bool{
			"ImportId=" + importIDs[0] + "|Location=s3://example/failure": true,
			"ImportId=" + importIDs[1] + "|Location=s3://example/failure": true,
		}
		for _, record := range records {
			gotID, _ := record["pm_record_id"].(string)
			if !wantIDs[gotID] {
				t.Fatalf("pm_record_id = %#v, want import-scoped failure id", record["pm_record_id"])
			}
		}
	})
}

func TestNativeCloudTrailReadSkipsInsightSelectorsNotEnabledTrails(t *testing.T) {
	var insightRequests []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		action := strings.TrimPrefix(r.Header.Get("X-Amz-Target"), cloudTrailTargetPrefix)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		switch action {
		case "DescribeTrails":
			_, _ = w.Write([]byte(`{"trailList":[{"Name":"plain-trail"},{"Name":"enabled-trail"}]}`))
		case "GetInsightSelectors":
			trailName, _ := body["TrailName"].(string)
			insightRequests = append(insightRequests, trailName)
			if trailName == "plain-trail" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"__type":"InsightNotEnabledException","Message":"Insights not enabled for trail"}`))
				return
			}
			_, _ = w.Write([]byte(`{"InsightSelectors":[{"InsightType":"ApiCallRateInsight"}]}`))
		default:
			t.Fatalf("unexpected target %s", action)
		}
	}))
	defer srv.Close()

	c := Connector{Client: srv.Client()}
	var records []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "get_insight_selectors", Config: fixtureRuntimeConfig(srv.URL)}, func(record connectors.Record) error {
		records = append(records, record)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got, want := strings.Join(insightRequests, ","), "plain-trail,enabled-trail"; got != want {
		t.Fatalf("insight requests = %s, want %s", got, want)
	}
	if len(records) != 1 || records[0]["pm_record_id"] != "enabled-trail" {
		t.Fatalf("records = %#v", records)
	}
}

func TestNativeCloudTrailRejectsInvalidMaxPages(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{name: "malformed", value: "ten"},
		{name: "negative", value: "-1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"Channels":[{"Name":"unexpected"}]}`))
			}))
			defer srv.Close()

			cfg := fixtureRuntimeConfig(srv.URL)
			cfg.Config["max_pages"] = tt.value
			c := Connector{Client: srv.Client()}
			err := c.Read(context.Background(), connectors.ReadRequest{Stream: "list_channels", Config: cfg}, func(connectors.Record) error {
				t.Fatal("emit called for invalid max_pages")
				return nil
			})
			if err == nil {
				t.Fatalf("Read unexpectedly accepted max_pages=%s", tt.value)
			}
			if !strings.Contains(err.Error(), "max_pages") {
				t.Fatalf("Read error = %v, want max_pages validation", err)
			}
			if requests != 0 {
				t.Fatalf("requests = %d, want 0", requests)
			}
		})
	}
}

func TestNativeCloudTrailRejectsBlockedReadStreams(t *testing.T) {
	blocked := []string{
		"management_events",
		"read_only_events",
		"write_only_events",
		"console_logins",
	}
	c := Connector{}
	for _, stream := range blocked {
		t.Run(stream, func(t *testing.T) {
			err := c.Read(context.Background(), connectors.ReadRequest{Stream: stream}, func(connectors.Record) error {
				t.Fatalf("emit called for blocked stream %s", stream)
				return nil
			})
			if err == nil {
				t.Fatalf("Read(%s) unexpectedly succeeded", stream)
			}
		})
	}
}

func TestNativeCloudTrailDirectAndWritesAreBlockedInScopeCorrectedSurface(t *testing.T) {
	c := Connector{}
	if _, err := c.OperationDirectRead(context.Background(), connectors.OperationDirectReadRequest{Operation: "aws-cloudtrail.lookup_events"}); err == nil {
		t.Fatal("OperationDirectRead unexpectedly succeeded for blocked direct operation")
	}
	if err := c.ValidateWrite(context.Background(), connectors.WriteRequest{Action: "start_logging"}, []connectors.Record{{"Name": "trail-fixture"}}); err == nil {
		t.Fatal("ValidateWrite unexpectedly accepted blocked write action")
	}
	if _, err := c.DryRunWrite(context.Background(), connectors.WriteRequest{Action: "start_logging"}, []connectors.Record{{"Name": "trail-fixture"}}); err == nil {
		t.Fatal("DryRunWrite unexpectedly accepted blocked write action")
	}
	if got, err := c.Write(context.Background(), connectors.WriteRequest{Action: "start_logging"}, []connectors.Record{{"Name": "trail-fixture"}}); err == nil || got.RecordsFailed != 1 {
		t.Fatalf("Write result = %+v err = %v, want blocked failure", got, err)
	}
}

func fixtureRuntimeConfig(baseURL string) connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config: map[string]string{
			"aws_region_name": "us-east-1",
			"base_url":        baseURL,
		},
		Secrets: map[string]string{
			"aws_key_id":     "fixture-access-key",
			"aws_secret_key": "fixture-secret-key",
		},
	}
}
