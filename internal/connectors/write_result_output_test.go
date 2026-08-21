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
				"X-Configured-Echo":     {Values: []string{"client-secret"}},
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
	for _, name := range []string{"Authorization", "X-Configured-Echo"} {
		if got := response.Headers[name]; !got.Masked || len(got.Values) != 0 {
			t.Fatalf("credential response header %q = %#v, want explicit marker", name, got)
		}
	}
	body := response.Body.(map[string]any)
	if body["id"] != "provider-1" || body["paid_tier"] != "enterprise" || !reflect.DeepEqual(body["rare"], map[string]any{"nested": true}) || !reflect.DeepEqual(body["credentialEcho"], map[string]bool{"masked": true}) {
		t.Fatalf("sanitized provider body = %#v, want ordinary fields plus explicit credential marker", body)
	}
	if result.ProviderResponses[0].Headers["Authorization"].Masked || result.ProviderResponses[0].Body.(map[string]any)["credentialEcho"] != "client-secret" {
		t.Fatal("sanitizing output mutated the in-memory provider result")
	}
}

func TestSanitizeOperationDirectWriteResultForOutputMasksDeclaredResponseSecretInPlace(t *testing.T) {
	result := OperationDirectWriteResult{
		Connector: "fixture", Operation: "fixture.create", Method: "POST", Path: "/fixed", Status: 201,
		Headers: map[string]WriteProviderHeader{"X-Request-ID": {Values: []string{"provider-1"}}, "Set-Cookie": {Values: []string{"session=secret"}}},
		Body: map[string]any{
			"credential": "new-provider-credential",
			"account":    map[string]any{"tier": "enterprise", "region": "eu"},
		},
		OutputSecretFields: []string{"credential"},
	}

	safe := SanitizeOperationDirectWriteResultForOutput(result, nil)
	if got := safe.Headers["X-Request-ID"]; got.Masked || !reflect.DeepEqual(got.Values, []string{"provider-1"}) {
		t.Fatalf("ordinary direct-write header = %#v, want preserved", got)
	}
	if got := safe.Headers["Set-Cookie"]; !got.Masked || len(got.Values) != 0 {
		t.Fatalf("credential direct-write header = %#v, want explicit marker", got)
	}
	body := safe.Body.(map[string]any)
	if !reflect.DeepEqual(body["credential"], map[string]bool{"masked": true}) || !reflect.DeepEqual(body["account"], map[string]any{"tier": "enterprise", "region": "eu"}) {
		t.Fatalf("sanitized direct-write body = %#v, want preserved account and masked credential", body)
	}
}
