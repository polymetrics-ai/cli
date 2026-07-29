package contractv1_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"

	contractv1 "polymetrics.ai/internal/pmbroker/contract/v1"
)

func TestSyntheticClientSuccessPinsAcceptedFixtures(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	fixtures := contractv1.AcceptedSyntheticFixtures()
	broker := contractv1.NewSyntheticBroker()
	client := broker.NewClient()

	compatibility, err := client.Compatibility(ctx)
	if err != nil {
		t.Fatalf("Compatibility() error = %v", err)
	}
	if compatibility.CurrentVersion != contractv1.ContractVersion1 {
		t.Fatalf("CurrentVersion = %q, want %q", compatibility.CurrentVersion, contractv1.ContractVersion1)
	}
	if !reflect.DeepEqual(compatibility, fixtures.Compatibility) {
		t.Fatalf("Compatibility() = %#v, want fixture %#v", compatibility, fixtures.Compatibility)
	}

	connection, err := client.ConnectorConnection(ctx, fixtures.ConnectorConnection.ConnectorConnectionID)
	if err != nil {
		t.Fatalf("ConnectorConnection() error = %v", err)
	}
	if !reflect.DeepEqual(connection, fixtures.ConnectorConnection) {
		t.Fatalf("ConnectorConnection() = %#v, want fixture %#v", connection, fixtures.ConnectorConnection)
	}
	if err := connection.AuthRef.Validate(); err != nil {
		t.Fatalf("AuthRef.Validate() error = %v", err)
	}
	if connection.AuthRef.Exportable {
		t.Fatal("AuthRef.Exportable = true, want false")
	}

	plan, err := client.CreateExecutionPlan(ctx, fixtures.ExecutionPlanRequest)
	if err != nil {
		t.Fatalf("CreateExecutionPlan() error = %v", err)
	}
	if !reflect.DeepEqual(plan, fixtures.ExecutionPlan) {
		t.Fatalf("CreateExecutionPlan() = %#v, want fixture %#v", plan, fixtures.ExecutionPlan)
	}
	if got := broker.ExecutionPlanRequestCount(); got != 1 {
		t.Fatalf("ExecutionPlanRequestCount() = %d, want 1", got)
	}
}

func TestSyntheticClientRefusesIncompatibleContractVersion(t *testing.T) {
	t.Parallel()

	fixtures := contractv1.AcceptedSyntheticFixtures()
	tests := []struct {
		name    string
		version contractv1.ContractVersion
	}{
		{name: "missing version", version: ""},
		{name: "unsupported version", version: contractv1.ContractVersion("0.9")},
		{name: "unsafe version header value", version: contractv1.ContractVersion("1.0\nX-Secret: value")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			broker := contractv1.NewSyntheticBroker()
			client := broker.NewClient(contractv1.WithClientContractVersion(tt.version))

			_, err := client.CreateExecutionPlan(context.Background(), fixtures.ExecutionPlanRequest)
			if err == nil {
				t.Fatal("CreateExecutionPlan() error = nil, want incompatible-version refusal")
			}
			var refusal *contractv1.IncompatibleContractVersionError
			if !errors.As(err, &refusal) {
				t.Fatalf("CreateExecutionPlan() error type = %T, want IncompatibleContractVersionError", err)
			}
			if refusal.StatusCode != http.StatusUpgradeRequired {
				t.Fatalf("StatusCode = %d, want %d", refusal.StatusCode, http.StatusUpgradeRequired)
			}
			if refusal.Response.Error.Code != contractv1.ErrorCodeIncompatibleContractVersion {
				t.Fatalf("error code = %q, want %q", refusal.Response.Error.Code, contractv1.ErrorCodeIncompatibleContractVersion)
			}
			if refusal.Response.Error.Message != contractv1.IncompatibleContractVersionMessage {
				t.Fatalf("message = %q, want %q", refusal.Response.Error.Message, contractv1.IncompatibleContractVersionMessage)
			}
			if !reflect.DeepEqual(refusal.Response, fixtures.IncompatibleVersionError) {
				t.Fatalf("response = %#v, want fixture %#v", refusal.Response, fixtures.IncompatibleVersionError)
			}
			if got := broker.ExecutionPlanRequestCount(); got != 0 {
				t.Fatalf("ExecutionPlanRequestCount() = %d, want 0 after version refusal", got)
			}
		})
	}
}

func TestContractSafetyInvariants(t *testing.T) {
	t.Parallel()

	fixtures := contractv1.AcceptedSyntheticFixtures()
	if !contractv1.CorrelationID("corr_0123456789abcdef").IsSafe() {
		t.Fatal("accepted correlation ID did not validate as safe")
	}
	unsafeCorrelationIDs := []contractv1.CorrelationID{
		"corr_0123456789abcdef\nX-Secret: value",
		"corr_0123456789abcdef/private_key",
		"corr_0123456789abcdef?access_token=token",
		"not-a-correlation-id",
	}
	for _, correlationID := range unsafeCorrelationIDs {
		if correlationID.IsSafe() {
			t.Fatalf("CorrelationID(%q).IsSafe() = true, want false", correlationID)
		}
	}

	exportable := fixtures.ConnectorConnection.AuthRef
	exportable.Exportable = true
	if err := exportable.Validate(); !errors.Is(err, contractv1.ErrRawSecretExportForbidden) {
		t.Fatalf("exportable AuthRef Validate() error = %v, want %v", err, contractv1.ErrRawSecretExportForbidden)
	}

	unsafeHint := fixtures.ConnectorConnection.AuthRef
	unsafeHint.DisplayHint = "access_token abc123"
	if err := unsafeHint.Validate(); !errors.Is(err, contractv1.ErrUnsafeDisplayHint) {
		t.Fatalf("unsafe display hint Validate() error = %v, want %v", err, contractv1.ErrUnsafeDisplayHint)
	}

	controlHint := fixtures.ConnectorConnection.AuthRef
	controlHint.DisplayHint = "managed\nby broker"
	if err := controlHint.Validate(); !errors.Is(err, contractv1.ErrUnsafeDisplayHint) {
		t.Fatalf("control display hint Validate() error = %v, want %v", err, contractv1.ErrUnsafeDisplayHint)
	}

	if err := fixtures.BrokerProfile.Validate(); err != nil {
		t.Fatalf("BrokerProfile.Validate() error = %v", err)
	}
	unsafeProfile := fixtures.BrokerProfile
	unsafeProfile.AllowedConnectorKinds = append(append([]string(nil), unsafeProfile.AllowedConnectorKinds...), "arbitrary-http")
	if err := unsafeProfile.Validate(); !errors.Is(err, contractv1.ErrInvalidIdentityBoundary) {
		t.Fatalf("unsafe BrokerProfile Validate() error = %v, want %v", err, contractv1.ErrInvalidIdentityBoundary)
	}

	payload, err := json.Marshal(fixtures)
	if err != nil {
		t.Fatalf("marshal fixtures: %v", err)
	}
	for _, forbidden := range []string{
		"raw_secret",
		"secret_value",
		"private_key",
		"service_account_json",
		"access_token",
		"refresh_token",
		"client_secret",
		"password",
	} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("fixtures expose forbidden raw-secret marker %q in %s", forbidden, payload)
		}
	}
}

func TestExecutionPlanRequestRejectsMismatchedIntentConnectorConnection(t *testing.T) {
	t.Parallel()

	request := contractv1.AcceptedSyntheticFixtures().ExecutionPlanRequest
	request.Intent.ValidateConnectorConnection.ConnectorConnectionID = contractv1.ConnectorConnectionID("ccn_aaaaaaaaaaaaaaaa")

	if err := request.Validate(); !errors.Is(err, contractv1.ErrInvalidExecutionIntent) {
		t.Fatalf("ExecutionPlanRequest.Validate() error = %v, want %v", err, contractv1.ErrInvalidExecutionIntent)
	}
}

func TestConnectorConnectionValidationAllowsContractEnums(t *testing.T) {
	t.Parallel()

	fixtures := contractv1.AcceptedSyntheticFixtures()
	validGCP := fixtures.ConnectorConnection
	validGCP.ConnectorKind = "gcp"
	validGCP.Status = "not_ready"
	validGCP.WriteMode = "plan_only"
	if err := validGCP.Validate(); err != nil {
		t.Fatalf("ConnectorConnection.Validate() for GCP non-ready plan-only connection error = %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*contractv1.ConnectorConnection)
	}{
		{
			name: "connector kind",
			mutate: func(connection *contractv1.ConnectorConnection) {
				connection.ConnectorKind = "arbitrary-http"
			},
		},
		{
			name: "status",
			mutate: func(connection *contractv1.ConnectorConnection) {
				connection.Status = "compromised"
			},
		},
		{
			name: "write mode",
			mutate: func(connection *contractv1.ConnectorConnection) {
				connection.WriteMode = "raw_http"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			connection := validGCP
			tt.mutate(&connection)
			if err := connection.Validate(); !errors.Is(err, contractv1.ErrInvalidExecutionPlan) {
				t.Fatalf("ConnectorConnection.Validate() error = %v, want %v", err, contractv1.ErrInvalidExecutionPlan)
			}
		})
	}
}

func TestNoGenericRequestEscapeHatches(t *testing.T) {
	t.Parallel()

	clientType := reflect.TypeOf(contractv1.NewSyntheticBroker().NewClient())
	forbiddenMethods := []string{
		"Do",
		"RoundTrip",
		"Request",
		"RawJSON",
		"WithHeader",
		"WithEndpoint",
		"WithURL",
		"WithBody",
		"SQL",
		"Shell",
	}
	for _, method := range forbiddenMethods {
		if _, ok := clientType.MethodByName(method); ok {
			t.Fatalf("Client exposes generic escape-hatch method %q", method)
		}
	}

	checked := map[reflect.Type]bool{}
	for _, value := range []any{
		contractv1.AcceptedSyntheticFixtures(),
		contractv1.ExecutionPlanRequest{},
		contractv1.ExecutionIntent{},
		contractv1.ValidateConnectorConnectionIntent{},
		contractv1.OpaqueSecretReference{},
	} {
		assertNoGenericFields(t, reflect.TypeOf(value), checked)
	}
}

func assertNoGenericFields(t *testing.T, typ reflect.Type, checked map[reflect.Type]bool) {
	t.Helper()
	for typ.Kind() == reflect.Pointer || typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct || checked[typ] {
		return
	}
	checked[typ] = true

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}
		jsonName := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonName == "" {
			jsonName = field.Name
		}
		if isForbiddenContractField(jsonName) || isForbiddenContractField(field.Name) {
			t.Fatalf("%s.%s exposes forbidden generic field %q", typ.Name(), field.Name, jsonName)
		}
		fieldType := field.Type
		if fieldType.Kind() == reflect.Map || fieldType.Kind() == reflect.Interface {
			t.Fatalf("%s.%s uses generic %s type", typ.Name(), field.Name, fieldType.Kind())
		}
		assertNoGenericFields(t, fieldType, checked)
	}
}

func isForbiddenContractField(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, forbidden := range []string{
		"url",
		"uri",
		"endpoint",
		"path",
		"method",
		"header",
		"headers",
		"body",
		"json",
		"payload",
		"raw",
		"raw_secret",
		"sql",
		"shell",
		"command",
		"provider_payload",
		"authenticated_http",
	} {
		if name == forbidden {
			return true
		}
	}
	return false
}
