package awscloudtrail_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	awscloudtrail "polymetrics.ai/internal/connectors/aws-cloudtrail"
)

// TestReadPaginatesAndSigns is the red-first test: it asserts AWS SigV4 auth on
// the Authorization header, the CloudTrail X-Amz-Target action, NextToken-based
// pagination across two pages, and record mapping of LookupEvents results.
func TestReadPaginatesAndSigns(t *testing.T) {
	var sawAuth, sawTarget, sawContentType string
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		sawAuth = r.Header.Get("Authorization")
		sawTarget = r.Header.Get("X-Amz-Target")
		sawContentType = r.Header.Get("Content-Type")

		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		next, _ := body["NextToken"].(string)

		w.Header().Set("Content-Type", "application/x-amz-json-1.1")
		switch next {
		case "":
			_, _ = w.Write([]byte(`{"Events":[` +
				`{"EventId":"e1","EventName":"ConsoleLogin","EventTime":1700000000,"Username":"alice","EventSource":"signin.amazonaws.com","CloudTrailEvent":"{\"eventVersion\":\"1.08\"}","Resources":[]},` +
				`{"EventId":"e2","EventName":"RunInstances","EventTime":1700000100,"Username":"bob","EventSource":"ec2.amazonaws.com","CloudTrailEvent":"{}","Resources":[]}` +
				`],"NextToken":"page2"}`))
		case "page2":
			_, _ = w.Write([]byte(`{"Events":[` +
				`{"EventId":"e3","EventName":"StopInstances","EventTime":1700000200,"Username":"carol","EventSource":"ec2.amazonaws.com","CloudTrailEvent":"{}","Resources":[]}` +
				`]}`))
		default:
			t.Errorf("unexpected NextToken=%q", next)
		}
	}))
	defer srv.Close()

	c := awscloudtrail.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":        srv.URL,
			"aws_region_name": "us-east-1",
		},
		Secrets: map[string]string{
			"aws_key_id":     "AKIAIOSFODNN7EXAMPLE",
			"aws_secret_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "management_events", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if calls != 2 {
		t.Fatalf("server calls = %d, want 2 (pagination)", calls)
	}
	if !strings.HasPrefix(sawAuth, "AWS4-HMAC-SHA256 ") {
		t.Fatalf("Authorization = %q, want AWS4-HMAC-SHA256 prefix (SigV4)", sawAuth)
	}
	if !strings.Contains(sawAuth, "Credential=AKIAIOSFODNN7EXAMPLE/") {
		t.Fatalf("Authorization missing access key in Credential scope: %q", sawAuth)
	}
	if !strings.HasSuffix(sawTarget, "LookupEvents") {
		t.Fatalf("X-Amz-Target = %q, want ...LookupEvents", sawTarget)
	}
	if sawContentType != "application/x-amz-json-1.1" {
		t.Fatalf("Content-Type = %q, want application/x-amz-json-1.1", sawContentType)
	}
	if len(got) != 3 {
		t.Fatalf("records = %d, want 3 (2 pages)", len(got))
	}
	for _, rec := range got {
		if rec["EventId"] == nil || rec["EventName"] == nil {
			t.Fatalf("record missing EventId/EventName: %+v", rec)
		}
	}
	if got[0]["EventName"] != "ConsoleLogin" {
		t.Fatalf("first record EventName = %v, want ConsoleLogin", got[0]["EventName"])
	}
}

// TestFixtureModeRead exercises the credential-free conformance path.
func TestFixtureModeRead(t *testing.T) {
	c := awscloudtrail.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"mode": "fixture"}}

	var got []connectors.Record
	err := c.Read(context.Background(), connectors.ReadRequest{Stream: "management_events", Config: cfg}, func(rec connectors.Record) error {
		got = append(got, rec)
		return nil
	})
	if err != nil {
		t.Fatalf("fixture Read: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("fixture Read produced no records")
	}
	for _, rec := range got {
		if rec["EventId"] == nil {
			t.Fatalf("fixture record missing EventId: %+v", rec)
		}
	}
	// Check also short-circuits in fixture mode (no network, no creds).
	if err := c.Check(context.Background(), cfg); err != nil {
		t.Fatalf("fixture Check: %v", err)
	}
}

func TestCheckRequiresSecrets(t *testing.T) {
	c := awscloudtrail.New()
	cfg := connectors.RuntimeConfig{Config: map[string]string{"aws_region_name": "us-east-1"}}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check without secrets should fail")
	}
}

func TestCatalogAndMetadata(t *testing.T) {
	c := awscloudtrail.New()
	if c.Name() != "aws-cloudtrail" {
		t.Fatalf("Name = %q, want aws-cloudtrail", c.Name())
	}
	md := c.Metadata()
	if !md.Capabilities.Read || !md.Capabilities.Catalog || !md.Capabilities.Check {
		t.Fatalf("capabilities = %+v, want Read+Catalog+Check", md.Capabilities)
	}
	if md.Capabilities.Write {
		t.Fatal("aws-cloudtrail is a read-only source; Write must be false")
	}
	cat, err := c.Catalog(context.Background(), connectors.RuntimeConfig{})
	if err != nil {
		t.Fatalf("Catalog: %v", err)
	}
	if len(cat.Streams) == 0 {
		t.Fatal("Catalog returned no streams")
	}
	found := false
	for _, s := range cat.Streams {
		if s.Name == "management_events" {
			found = true
			if len(s.PrimaryKey) == 0 || len(s.Fields) == 0 {
				t.Fatalf("management_events stream malformed: %+v", s)
			}
		}
	}
	if !found {
		t.Fatal("Catalog missing management_events stream")
	}
}

func TestBaseURLSSRFValidation(t *testing.T) {
	c := awscloudtrail.New()
	cfg := connectors.RuntimeConfig{
		Config: map[string]string{
			"base_url":        "ftp://evil.example.com",
			"aws_region_name": "us-east-1",
		},
		Secrets: map[string]string{"aws_key_id": "k", "aws_secret_key": "s"},
	}
	if err := c.Check(context.Background(), cfg); err == nil {
		t.Fatal("Check should reject non-http(s) base_url")
	}
}

func TestRegistryResolution(t *testing.T) {
	r := connectors.NewRegistry()
	if _, ok := r.Get("aws-cloudtrail"); !ok {
		t.Fatal("registry did not resolve aws-cloudtrail (self-registration)")
	}
}
