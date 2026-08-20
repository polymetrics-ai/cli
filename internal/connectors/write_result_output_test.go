package connectors

import (
	"reflect"
	"testing"
)

func TestSanitizeWriteResultForOutputPreservesOrdinaryFieldsAndMasksOnlyCredentials(t *testing.T) {
	result := WriteResult{
		RecordsWritten: 1,
		ProviderResponses: []WriteProviderResponse{{
			Status: 201,
			Headers: map[string]WriteProviderHeader{
				"X-Request-ID":          {Values: []string{"provider-1"}},
				"Authorization":         {Values: []string{"provider-secret"}},
				"X-Configured-Echo":     {Values: []string{"Bearer client-secret", "second-value"}},
				"X-Paid-Tier-Indicator": {Values: []string{"enterprise"}},
			},
			Body: map[string]any{
				"id":             "provider-1",
				"paid_tier":      "enterprise",
				"rare":           map[string]any{"nested": true},
				"credentialEcho": "client-secret",
			},
		}},
	}

	safe := SanitizeWriteResultForOutput(result, map[string]string{"token": "client-secret"})
	response := safe.ProviderResponses[0]
	if got := response.Headers["X-Request-ID"]; !reflect.DeepEqual(got.Values, []string{"provider-1"}) || got.Masked {
		t.Fatalf("ordinary response header = %#v, want preserved", got)
	}
	if got := response.Headers["X-Paid-Tier-Indicator"]; !reflect.DeepEqual(got.Values, []string{"enterprise"}) || got.Masked {
		t.Fatalf("paid-tier response header = %#v, want preserved", got)
	}
	if got := response.Headers["Authorization"]; got.Masked || !reflect.DeepEqual(got.Values, []string{"provider-secret"}) {
		t.Fatalf("unconfigured authorization response header = %#v, want preserved", got)
	}
	if got := response.Headers["X-Configured-Echo"]; !got.Masked || !reflect.DeepEqual(got.Values, []string{"Bearer [masked]", "second-value"}) {
		t.Fatalf("configured credential response header = %#v, want exact replacement with all values retained", got)
	}
	body := response.Body.(map[string]any)
	if body["id"] != "provider-1" || body["paid_tier"] != "enterprise" || !reflect.DeepEqual(body["rare"], map[string]any{"nested": true}) || body["credentialEcho"] != "[masked]" {
		t.Fatalf("sanitized provider body = %#v, want ordinary fields plus exact credential replacement", body)
	}
	if result.ProviderResponses[0].Headers["Authorization"].Masked || result.ProviderResponses[0].Body.(map[string]any)["credentialEcho"] != "client-secret" {
		t.Fatal("sanitizing output mutated the in-memory provider result")
	}
}

func TestSanitizeOperationDirectWriteResultForOutputMasksOnlyConfiguredCredentialBytes(t *testing.T) {
	result := OperationDirectWriteResult{
		Connector: "fixture", Operation: "fixture.create", Method: "POST", Path: "/fixed", ResponseReceived: true, Status: 201,
		Headers:         map[string]WriteProviderHeader{"X-Request-ID": {Values: []string{"provider-1"}}, "Set-Cookie": {Values: []string{"session=secret"}}, "X-Echo": {Values: []string{"client-secret", "ordinary"}}},
		BodyPresent:     true,
		BodyRaw:         ` { "credential" : "new-provider-credential", "echo" : "client-secret" } `,
		BodyRawEncoding: "text",
		Body: map[string]any{
			"credential": "new-provider-credential",
			"echo":       "client-secret",
			"account":    map[string]any{"tier": "enterprise", "region": "eu"},
		},
		OutputSecretFields: []string{"credential"},
	}

	safe := SanitizeOperationDirectWriteResultForOutput(result, map[string]string{"token": "client-secret"})
	if got := safe.Headers["X-Request-ID"]; got.Masked || !reflect.DeepEqual(got.Values, []string{"provider-1"}) {
		t.Fatalf("ordinary direct-write header = %#v, want preserved", got)
	}
	if got := safe.Headers["Set-Cookie"]; got.Masked || !reflect.DeepEqual(got.Values, []string{"session=secret"}) {
		t.Fatalf("unconfigured direct-write header = %#v, want preserved", got)
	}
	if got := safe.Headers["X-Echo"]; !got.Masked || !reflect.DeepEqual(got.Values, []string{"[masked]", "ordinary"}) {
		t.Fatalf("configured direct-write header = %#v, want exact replacement", got)
	}
	body := safe.Body.(map[string]any)
	if body["credential"] != "new-provider-credential" || body["echo"] != "[masked]" || !reflect.DeepEqual(body["account"], map[string]any{"tier": "enterprise", "region": "eu"}) {
		t.Fatalf("sanitized direct-write body = %#v, want only configured credential replacement", body)
	}
	if safe.BodyRaw != ` { "credential" : "new-provider-credential", "echo" : "[masked]" } ` {
		t.Fatalf("sanitized direct-write raw body = %q, want byte-preserving credential replacement", safe.BodyRaw)
	}
}
