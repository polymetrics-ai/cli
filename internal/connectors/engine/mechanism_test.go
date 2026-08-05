package engine

import (
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
)

func bundleFSWithMetadata(name, metadataJSON string) fstest.MapFS {
	fsys := fullValidBundleFS(name)
	fsys[name+"/metadata.json"] = &fstest.MapFile{Data: []byte(metadataJSON)}
	return fsys
}

func TestMechanismDefaultsToOfficialAPIWhenAbsent(t *testing.T) {
	b, err := Load(bundleFSWithMetadata("acme", validMetadata("acme")), "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.Metadata.Mechanism == nil {
		t.Fatalf("Mechanism = nil, want a defaulted official_api mechanism")
	}
	if b.Metadata.Mechanism.Kind != MechanismOfficialAPI {
		t.Fatalf("Mechanism.Kind = %q, want %q", b.Metadata.Mechanism.Kind, MechanismOfficialAPI)
	}
	if !b.Metadata.Mechanism.SanctionedByProvider {
		t.Fatalf("defaulted mechanism SanctionedByProvider = false, want true")
	}
}

func TestMechanismExplicitOfficialAPILoads(t *testing.T) {
	meta := `{
		"name": "acme",
		"display_name": "Test Connector",
		"description": "a test connector",
		"integration_type": "api",
		"release_stage": "ga",
		"capabilities": { "check": true, "read": true, "write": false, "query": false, "cdc": false, "dynamic_schema": false },
		"mechanism": {
			"kind": "official_api",
			"label": "Acme OAuth API",
			"sanctioned_by_provider": true,
			"provider_terms_url": "https://acme.example/terms",
			"auth_flow": "oauth2_authorization_code_pkce"
		}
	}`
	b, err := Load(bundleFSWithMetadata("acme", meta), "acme")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.Metadata.Mechanism.Label != "Acme OAuth API" {
		t.Fatalf("Mechanism.Label = %q", b.Metadata.Mechanism.Label)
	}
}

func validWebSessionMetadata(name string) string {
	return `{
		"name": "` + name + `",
		"display_name": "Test Web Connector",
		"description": "a test unofficial web session connector",
		"integration_type": "api",
		"release_stage": "alpha",
		"capabilities": { "check": true, "read": true, "write": false, "query": false, "cdc": false, "dynamic_schema": false },
		"risk": { "approval": "none; unofficial mechanism, user-accepted risk" },
		"mechanism": {
			"kind": "web_session",
			"label": "Test provider web session (unofficial, experimental)",
			"sanctioned_by_provider": false,
			"provider_terms_url": "https://provider.example/terms",
			"auth_flow": "browser_session_capture",
			"opt_in_required": true,
			"upstream_pin": {
				"repo": "https://github.com/example/reference-cli",
				"sha": "abcdef1234567890",
				"verified_at": "2026-08-03"
			},
			"breakage_review_cadence_days": 30
		}
	}`
}

func TestMechanismWebSessionLoadsWhenValid(t *testing.T) {
	b, err := Load(bundleFSWithMetadata("acme-web", validWebSessionMetadata("acme-web")), "acme-web")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if b.Metadata.Mechanism.Kind != MechanismWebSession {
		t.Fatalf("Mechanism.Kind = %q, want %q", b.Metadata.Mechanism.Kind, MechanismWebSession)
	}
	if b.Metadata.Mechanism.UpstreamPin == nil || b.Metadata.Mechanism.UpstreamPin.SHA != "abcdef1234567890" {
		t.Fatalf("UpstreamPin = %+v", b.Metadata.Mechanism.UpstreamPin)
	}
}

func TestMechanismSynthesisPreservesWebGovernanceMetadata(t *testing.T) {
	meta := strings.Replace(validWebSessionMetadata("acme-web"), `"breakage_review_cadence_days": 30`, `"breakage_review_cadence_days": 30, "disabled_reason": "upstream contract changed; awaiting review"`, 1)
	b, err := Load(bundleFSWithMetadata("acme-web", meta), "acme-web")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	c := New(b, nil)

	for _, tc := range []struct {
		name string
		mech *connectors.MechanismSpec
	}{
		{name: "metadata", mech: c.Metadata().Mechanism},
		{name: "definition", mech: c.Definition().Mechanism},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mech == nil {
				t.Fatal("Mechanism = nil")
			}
			if tc.mech.UpstreamPin == nil || tc.mech.UpstreamPin.Repo != "https://github.com/example/reference-cli" || tc.mech.UpstreamPin.SHA != "abcdef1234567890" || tc.mech.UpstreamPin.VerifiedAt != "2026-08-03" {
				t.Fatalf("UpstreamPin = %+v, want the declared pin", tc.mech.UpstreamPin)
			}
			if tc.mech.BreakageReviewCadenceDays != 30 {
				t.Fatalf("BreakageReviewCadenceDays = %d, want 30", tc.mech.BreakageReviewCadenceDays)
			}
			if tc.mech.DisabledReason != "upstream contract changed; awaiting review" {
				t.Fatalf("DisabledReason = %q, want projected kill-switch reason", tc.mech.DisabledReason)
			}
		})
	}
}

func TestMechanismWebSessionRequiresUnsanctioned(t *testing.T) {
	meta := strings.Replace(validWebSessionMetadata("acme-web"), `"sanctioned_by_provider": false`, `"sanctioned_by_provider": true`, 1)
	_, err := Load(bundleFSWithMetadata("acme-web", meta), "acme-web")
	if err == nil || !strings.Contains(err.Error(), "sanctioned_by_provider") {
		t.Fatalf("Load error = %v, want a sanctioned_by_provider complaint", err)
	}
}

func TestMechanismWebSessionRequiresOptIn(t *testing.T) {
	meta := strings.Replace(validWebSessionMetadata("acme-web"), `"opt_in_required": true,`, "", 1)
	_, err := Load(bundleFSWithMetadata("acme-web", meta), "acme-web")
	if err == nil || !strings.Contains(err.Error(), "opt_in_required") {
		t.Fatalf("Load error = %v, want an opt_in_required complaint", err)
	}
}

func TestMechanismWebSessionRequiresAlphaReleaseStage(t *testing.T) {
	meta := strings.Replace(validWebSessionMetadata("acme-web"), `"release_stage": "alpha",`, `"release_stage": "ga",`, 1)
	_, err := Load(bundleFSWithMetadata("acme-web", meta), "acme-web")
	if err == nil || !strings.Contains(err.Error(), "release_stage") {
		t.Fatalf("Load error = %v, want a release_stage complaint", err)
	}
}

func TestMechanismWebSessionRequiresRiskApproval(t *testing.T) {
	meta := strings.Replace(validWebSessionMetadata("acme-web"), `"risk": { "approval": "none; unofficial mechanism, user-accepted risk" },`, "", 1)
	_, err := Load(bundleFSWithMetadata("acme-web", meta), "acme-web")
	if err == nil || !strings.Contains(err.Error(), "risk.approval") {
		t.Fatalf("Load error = %v, want a risk.approval complaint", err)
	}
}

func TestMechanismWebSessionRequiresUpstreamPin(t *testing.T) {
	meta := `{
		"name": "acme-web",
		"display_name": "Test Web Connector",
		"description": "a test unofficial web session connector",
		"integration_type": "api",
		"release_stage": "alpha",
		"capabilities": { "check": true, "read": true, "write": false, "query": false, "cdc": false, "dynamic_schema": false },
		"risk": { "approval": "none; unofficial mechanism, user-accepted risk" },
		"mechanism": {
			"kind": "web_session",
			"sanctioned_by_provider": false,
			"opt_in_required": true
		}
	}`
	_, err := Load(bundleFSWithMetadata("acme-web", meta), "acme-web")
	if err == nil || !strings.Contains(err.Error(), "upstream_pin") {
		t.Fatalf("Load error = %v, want an upstream_pin complaint", err)
	}
}

func TestMechanismWebNameSuffixMustDeclareWebSession(t *testing.T) {
	_, err := Load(bundleFSWithMetadata("acme-web", validMetadata("acme-web")), "acme-web")
	if err == nil || !strings.Contains(err.Error(), "-web") {
		t.Fatalf("Load error = %v, want a -web naming complaint", err)
	}
}

func TestMechanismNonWebNameMustNotDeclareWebSession(t *testing.T) {
	meta := strings.ReplaceAll(validWebSessionMetadata("acme-web"), "acme-web", "acme")
	_, err := Load(bundleFSWithMetadata("acme", meta), "acme")
	if err == nil || !strings.Contains(err.Error(), "-web") {
		t.Fatalf("Load error = %v, want a -web naming complaint", err)
	}
}

func TestMechanismDisabledReasonIsValidBundleState(t *testing.T) {
	meta := strings.Replace(validWebSessionMetadata("acme-web"), `"breakage_review_cadence_days": 30`, `"breakage_review_cadence_days": 30, "disabled_reason": "upstream Voyager routes rotated; pending re-verification"`, 1)
	b, err := Load(bundleFSWithMetadata("acme-web", meta), "acme-web")
	if err != nil {
		t.Fatalf("Load with disabled_reason set: %v", err)
	}
	if b.Metadata.Mechanism.DisabledReason != "upstream Voyager routes rotated; pending re-verification" {
		t.Fatalf("DisabledReason = %q", b.Metadata.Mechanism.DisabledReason)
	}
	if !b.IsDisabled() {
		t.Fatal("IsDisabled() = false, want true")
	}
}

func TestMechanismUnknownKindRejected(t *testing.T) {
	meta := strings.Replace(validMetadata("acme"), `"capabilities"`, `"mechanism": { "kind": "carrier_pigeon", "sanctioned_by_provider": true }, "capabilities"`, 1)
	_, err := Load(bundleFSWithMetadata("acme", meta), "acme")
	if err == nil || !strings.Contains(err.Error(), "kind") {
		t.Fatalf("Load error = %v, want a mechanism kind complaint", err)
	}
}

func TestMechanismOfficialAPIRequiresSanctioned(t *testing.T) {
	meta := strings.Replace(validMetadata("acme"), `"capabilities"`, `"mechanism": { "kind": "official_api", "sanctioned_by_provider": false }, "capabilities"`, 1)
	_, err := Load(bundleFSWithMetadata("acme", meta), "acme")
	if err == nil || !strings.Contains(err.Error(), "sanctioned_by_provider") {
		t.Fatalf("Load error = %v, want a sanctioned_by_provider complaint", err)
	}
}
