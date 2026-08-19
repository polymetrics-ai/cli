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
	if string(firstRateKey) == string(identity.AuthCohortKey()) {
		t.Fatal("rate scope projection reused the authentication cohort key")
	}
	if reflect.TypeOf(firstRateKey) == reflect.TypeOf(identity.AuthCohortKey()) {
		t.Fatal("rate scope and auth cohort projections have the same Go type")
	}
	if strings.Contains(string(firstRateKey), accountScope.Subject) {
		t.Fatal("rate scope projection contains its subject preimage")
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
	differentKind, err := identity.RateScopeKey(RateLimitScope{
		PolicyID: "core-rest-v1",
		Kind:     RateScopeKindApplication,
		Subject:  "account-fixture-001",
	})
	if err != nil {
		t.Fatalf("RateScopeKey(different kind) error = %v", err)
	}
	if firstRateKey == differentKind {
		t.Fatal("different declared rate scope kinds shared a budget")
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
	encoded, err := json.Marshal(struct {
		Identity any `json:"identity"`
	}{Identity: identity})
	if err != nil {
		t.Fatalf("marshal identity: %v", err)
	}
	for _, preimage := range []string{binding.BindingID, binding.ProviderFamily, binding.AuthProfile} {
		if strings.Contains(string(encoded), preimage) {
			t.Fatal("identity JSON contains a protected coordination preimage")
		}
	}
	if strings.Contains(string(identity.AuthCohortKey()), binding.BindingID) {
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
