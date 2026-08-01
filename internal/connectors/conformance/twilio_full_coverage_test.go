package conformance

import (
	"io/fs"
	"slices"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestTwilioOfficialParityLedgerAndFixtureCoverage(t *testing.T) {
	bundle, err := engine.Load(realDefsFS(), "twilio")
	if err != nil {
		t.Fatalf("load twilio bundle: %v", err)
	}
	if bundle.Surface == nil {
		t.Fatal("twilio api_surface.json is missing")
	}
	if bundle.Surface.OperationLedgerVersion != 1 {
		t.Fatalf("operation_ledger_version = %d, want 1", bundle.Surface.OperationLedgerVersion)
	}

	laneCounts := map[string]int{}
	for _, endpoint := range bundle.Surface.Endpoints {
		laneCounts[twilioOfficialLane(endpoint.Method, endpoint.Path)]++
	}
	wantLanes := map[string]int{
		"etl_read":                 96,
		"reverse_etl_write":        93,
		"direct_read_query_search": 0,
		"binary_file":              3,
		"cdc_changefeed":           5,
		"excluded_not_applicable":  0,
	}
	for lane, want := range wantLanes {
		if got := laneCounts[lane]; got != want {
			t.Fatalf("lane %s count = %d, want %d (all official Twilio v2010 operations must be classified exactly once)", lane, got, want)
		}
	}
	if got := len(bundle.Surface.Endpoints); got != 197 {
		t.Fatalf("api_surface endpoint rows = %d, want 197", got)
	}
	if got := len(bundle.Streams); got != 103 {
		t.Fatalf("streams = %d, want 103", got)
	}
	if got := len(bundle.Writes); got != 94 {
		t.Fatalf("write actions = %d, want 94", got)
	}

	for _, streamName := range []string{"account", "accounts"} {
		var stream engine.StreamSpec
		found := false
		for _, candidate := range bundle.Streams {
			if candidate.Name == streamName {
				stream = candidate
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("stream %q is missing", streamName)
		}
		if stream.Projection == "passthrough" {
			t.Fatalf("stream %q projection = passthrough, want schema boundary for Twilio auth_token", streamName)
		}
		schema := bundle.Schemas[streamName]
		if schema == nil {
			t.Fatalf("stream %q schema is missing", streamName)
		}
		if slices.Contains(schema.Properties(), "auth_token") {
			t.Fatalf("stream %q schema exposes auth_token", streamName)
		}
	}

	redactionChecks := []struct {
		name   string
		fields []string
	}{
		{name: "create_call", fields: []string{"SipAuthPassword", "CallToken"}},
		{name: "create_participant", fields: []string{"SipAuthPassword", "CallToken"}},
		{name: "create_sip_credential", fields: []string{"Password"}},
		{name: "update_sip_credential", fields: []string{"Password"}},
	}
	for _, check := range redactionChecks {
		var action engine.WriteAction
		found := false
		for _, candidate := range bundle.Writes {
			if candidate.Name == check.name {
				action = candidate
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("write action %q is missing", check.name)
		}
		for _, field := range check.fields {
			if !slices.Contains(action.RedactFields, field) {
				t.Fatalf("write action %q redact_fields = %v, want %q", check.name, action.RedactFields, field)
			}
		}
	}

	assertTwilioCoverage(t, bundle, "GET", "/Accounts/{AccountSid}/Messages/{MessageSid}/Media.json", "stream", "medias")
	assertTwilioCoverage(t, bundle, "GET", "/Accounts/{AccountSid}/Messages/{MessageSid}/Media/{Sid}.json", "stream", "media")
	assertTwilioCoverage(t, bundle, "DELETE", "/Accounts/{AccountSid}/Messages/{MessageSid}/Media/{Sid}.json", "write", "delete_media")
	assertTwilioCoverage(t, bundle, "GET", "/Accounts/{AccountSid}/Calls/{CallSid}/Events.json", "stream", "call_events")
	assertTwilioCoverage(t, bundle, "GET", "/Accounts/{AccountSid}/Notifications.json", "stream", "notifications")

	for _, stream := range bundle.Streams {
		if !twilioFixtureExists(bundle.Fixtures, "streams/"+stream.Name+"/page_1.json") {
			t.Fatalf("stream %q has no page_1 fixture; every executable Twilio stream needs sanitized fixture coverage", stream.Name)
		}
	}
	for _, action := range bundle.Writes {
		if !twilioFixtureExists(bundle.Fixtures, "writes/"+action.Name+".json") {
			t.Fatalf("write action %q has no write fixture; every executable Twilio write needs sanitized request-shape coverage", action.Name)
		}
	}
}

func twilioOfficialLane(method, endpointPath string) string {
	if strings.Contains(endpointPath, "/Messages/{MessageSid}/Media") {
		return "binary_file"
	}
	if strings.Contains(endpointPath, "/Events") || strings.Contains(endpointPath, "/Notifications") {
		return "cdc_changefeed"
	}
	if strings.EqualFold(method, "GET") {
		return "etl_read"
	}
	return "reverse_etl_write"
}

func assertTwilioCoverage(t *testing.T, bundle engine.Bundle, method, endpointPath, kind, name string) {
	t.Helper()
	for _, endpoint := range bundle.Surface.Endpoints {
		if !strings.EqualFold(endpoint.Method, method) || endpoint.Path != endpointPath || endpoint.CoveredBy == nil {
			continue
		}
		switch kind {
		case "stream":
			if endpoint.CoveredBy.Stream == name {
				return
			}
		case "write":
			if endpoint.CoveredBy.Write == name {
				return
			}
		}
		break
	}
	t.Fatalf("%s %s is not covered_by.%s %q", method, endpointPath, kind, name)
}

func twilioFixtureExists(fixtures fs.FS, name string) bool {
	if fixtures == nil {
		return false
	}
	_, err := fs.Stat(fixtures, name)
	return err == nil
}
