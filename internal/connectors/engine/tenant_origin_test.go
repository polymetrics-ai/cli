package engine

import (
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestResolveTenantOriginLocksHTTPSOriginAndAPIPath(t *testing.T) {
	origin, err := resolveTenantOrigin(TenantOriginSpec{ConfigKey: "instance_api_url", AppendPath: "/api"}, map[string]string{"instance_api_url": "https://tenant.example"})
	if err != nil {
		t.Fatalf("resolveTenantOrigin: %v", err)
	}
	if origin != "https://tenant.example/api" {
		t.Fatalf("origin = %q, want locked API path", origin)
	}
}

func TestResolveTenantOriginPermitsOnlyDeclaredLoopbackHTTP(t *testing.T) {
	spec := TenantOriginSpec{ConfigKey: "origin", AppendPath: "/api", AllowLoopbackHTTP: true}
	origin, err := resolveTenantOrigin(spec, map[string]string{"origin": "http://127.0.0.1:8080"})
	if err != nil || origin != "http://127.0.0.1:8080/api" {
		t.Fatalf("loopback origin = %q, %v; want bounded local test origin", origin, err)
	}
	for _, raw := range []string{
		"http://tenant.example",
		"https://user@tenant.example",
		"https://tenant.example?query=1",
		"https://tenant.example#fragment",
	} {
		if _, err := resolveTenantOrigin(spec, map[string]string{"origin": raw}); err == nil {
			t.Fatalf("unsafe origin %q was accepted", raw)
		}
	}
	if _, err := resolveTenantOrigin(TenantOriginSpec{ConfigKey: "origin", AppendPath: "/../admin"}, map[string]string{"origin": "https://tenant.example"}); err == nil {
		t.Fatal("unsafe append path was accepted")
	}
	origin, err = resolveTenantOrigin(spec, map[string]string{"origin": "https://tenant.example/api"})
	if err != nil || origin != "https://tenant.example/api" {
		t.Fatalf("already-normalized origin = %q, %v", origin, err)
	}
}

func TestTenantOriginSchemaRejectsMissingOrUnknownDescriptorFields(t *testing.T) {
	for _, raw := range []string{
		`{"base":{"url":"https://unused.example","tenant_origin":{}},"streams":[]}`,
		`{"base":{"url":"https://unused.example","tenant_origin":{"config_key":"origin","unexpected":true}},"streams":[]}`,
	} {
		if err := metaSchemas.streams.Validate(mustDecodeAny([]byte(raw))); err == nil {
			t.Fatalf("tenant-origin descriptor %s was accepted", raw)
		}
	}
}

func TestTenantOriginOverridesDefaultOperationRoute(t *testing.T) {
	bundle := Bundle{
		Name: "tenant-connector",
		HTTP: HTTPBase{
			URL:          "https://unused.invalid",
			TenantOrigin: &TenantOriginSpec{ConfigKey: "instance_api_url", AppendPath: "/api", AllowLoopbackHTTP: true},
		},
	}
	baseURL, err := resolveOperationRoute(bundle, connectors.RuntimeConfig{Config: map[string]string{"instance_api_url": "http://127.0.0.1:8080"}}, "", "cards", "/card")
	if err != nil {
		t.Fatalf("resolveOperationRoute: %v", err)
	}
	if baseURL != "http://127.0.0.1:8080/api" {
		t.Fatalf("operation route base = %q, want declaration-bound tenant origin", baseURL)
	}
}
