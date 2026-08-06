package connectors

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestCoordinationIdentity_DerivesDistinctOpaqueScopes(t *testing.T) {
	binding := CredentialBinding{
		BindingID:      "binding-fixture-001",
		ProviderFamily: "provider-fixture",
		AuthProfile:    "service-profile",
	}
	identity, err := NewCoordinationIdentity([]byte("project-salt-fixture"), binding)
	if err != nil {
		t.Fatalf("NewCoordinationIdentity() error = %v", err)
	}

	linked, err := NewCoordinationIdentity([]byte("project-salt-fixture"), binding)
	if err != nil {
		t.Fatalf("NewCoordinationIdentity() for linked credential error = %v", err)
	}
	if identity.AuthCohortKey() == "" {
		t.Fatal("AuthCohortKey() is empty")
	}
	if identity.AuthCohortKey() != linked.AuthCohortKey() {
		t.Fatal("explicitly linked credentials received different auth cohort projections")
	}

	accountScope := RateLimitScope{
		PolicyID: "core-rest-v1",
		Kind:     RateScopeKindAccount,
		Subject:  "account-fixture-001",
	}
	firstRateKey, err := identity.RateScopeKey(accountScope)
	if err != nil {
		t.Fatalf("RateScopeKey(account) error = %v", err)
	}
	linkedRateKey, err := linked.RateScopeKey(accountScope)
	if err != nil {
		t.Fatalf("RateScopeKey(linked account) error = %v", err)
	}
	if firstRateKey == "" {
		t.Fatal("RateScopeKey(account) is empty")
	}
	if firstRateKey != linkedRateKey {
		t.Fatal("compatible linked credentials received different rate scope projections")
	}

	differentSubject, err := identity.RateScopeKey(RateLimitScope{
		PolicyID: "core-rest-v1",
		Kind:     RateScopeKindAccount,
		Subject:  "account-fixture-002",
	})
	if err != nil {
		t.Fatalf("RateScopeKey(different subject) error = %v", err)
	}
	if firstRateKey == differentSubject {
		t.Fatal("different declared rate subjects shared a budget")
	}

	if _, err := identity.RateScopeKey(RateLimitScope{}); err == nil {
		t.Fatal("RateScopeKey() accepted an undeclared rate scope")
	}
	if _, err := identity.RateScopeKey(RateLimitScope{
		PolicyID: "core-rest-v1",
		Kind:     RateScopeKind("unrecognized"),
		Subject:  "account-fixture-001",
	}); err == nil {
		t.Fatal("RateScopeKey() accepted an unsupported rate scope kind")
	}
}

func TestCoordinationIdentity_ContainsNoBindingOrSecretInput(t *testing.T) {
	binding := CredentialBinding{
		BindingID:      "binding-fixture-hidden",
		ProviderFamily: "provider-fixture",
		AuthProfile:    "service-profile",
	}
	identity, err := NewCoordinationIdentity([]byte("project-salt-fixture"), binding)
	if err != nil {
		t.Fatalf("NewCoordinationIdentity() error = %v", err)
	}
	encoded, err := json.Marshal(identity)
	if err != nil {
		t.Fatalf("marshal identity: %v", err)
	}
	if strings.Contains(string(encoded), binding.BindingID) {
		t.Fatal("identity JSON contains the protected binding preimage")
	}
	if strings.Contains(identity.AuthCohortKey(), binding.BindingID) {
		t.Fatal("auth cohort projection contains the protected binding preimage")
	}

	inputType := reflect.TypeOf(CredentialBinding{})
	for i := 0; i < inputType.NumField(); i++ {
		field := strings.ToLower(inputType.Field(i).Name)
		if strings.Contains(field, "secret") || strings.Contains(field, "revision") {
			t.Fatal("coordination identity input accepts secret or approval revision material")
		}
	}

	invalid := CredentialBinding{BindingID: "binding-fixture-not-for-errors"}
	if _, err := NewCoordinationIdentity([]byte("project-salt-fixture"), invalid); err == nil {
		t.Fatal("NewCoordinationIdentity() accepted incomplete binding metadata")
	} else if strings.Contains(err.Error(), invalid.BindingID) {
		t.Fatal("identity validation error contains the protected binding preimage")
	}
}

func TestCoordinationIdentity_IsIndependentOfApprovalRevision(t *testing.T) {
	identity, err := NewCoordinationIdentity([]byte("project-salt-fixture"), CredentialBinding{
		BindingID:      "binding-fixture-approval-lifetime",
		ProviderFamily: "provider-fixture",
		AuthProfile:    "service-profile",
	})
	if err != nil {
		t.Fatalf("NewCoordinationIdentity() error = %v", err)
	}

	first := RuntimeConfig{CredentialRevision: "approval-evidence-a", CoordinationIdentity: identity}
	second := RuntimeConfig{CredentialRevision: "approval-evidence-b", CoordinationIdentity: identity}
	if first.CoordinationIdentity.AuthCohortKey() != second.CoordinationIdentity.AuthCohortKey() {
		t.Fatal("approval evidence changed the auth cohort projection")
	}
	firstRate, err := first.CoordinationIdentity.RateScopeKey(RateLimitScope{
		PolicyID: "core-rest-v1",
		Kind:     RateScopeKindAccount,
		Subject:  "account-fixture-001",
	})
	if err != nil {
		t.Fatalf("first RateScopeKey() error = %v", err)
	}
	secondRate, err := second.CoordinationIdentity.RateScopeKey(RateLimitScope{
		PolicyID: "core-rest-v1",
		Kind:     RateScopeKindAccount,
		Subject:  "account-fixture-001",
	})
	if err != nil {
		t.Fatalf("second RateScopeKey() error = %v", err)
	}
	if firstRate != secondRate {
		t.Fatal("approval evidence changed the rate scope projection")
	}
}
