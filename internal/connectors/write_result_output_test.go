package connectors

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestWriteResultOutputMasksConfiguredSecretsAndPreservesOrdinaryProviderTruth(t *testing.T) {
	credential := "client-secret"
	numericCredential := "12345678901234567890"
	result := WriteResult{
		RecordsWritten: 1,
		ProviderResponses: []WriteProviderResponse{{
			Status:          201,
			BodyPresent:     true,
			BodyBytes:       len(`{"credential":"client-secret","numeric":12345678901234567890}`),
			BodyRaw:         `{"credential":"client-secret","numeric":12345678901234567890}`,
			BodyRawEncoding: "text",
			Headers: map[string]WriteProviderHeader{
				"X-Provider-Echo": {Values: []string{credential, "ordinary"}},
			},
			Body: map[string]any{
				"token":             "ordinary-occurrence-id",
				"credential_echo":   credential,
				"numeric_echo":      json.Number(numericCredential),
				"base64_echo":       base64.StdEncoding.EncodeToString([]byte(credential)),
				"nested":            map[string]any{"value": credential},
				"ordinary_provider": true,
			},
		}},
	}

	got := SanitizeWriteResultForOutput(result, map[string]string{"token": credential, "numeric": numericCredential})
	if reflect.DeepEqual(got, result) {
		t.Fatal("public provider result retained configured credentials")
	}
	if got.ProviderResponses[0].Body.(map[string]any)["token"] != "ordinary-occurrence-id" {
		t.Fatal("ordinary token-named provider identifier was changed")
	}
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), credential) || strings.Contains(string(encoded), numericCredential) || strings.Contains(string(encoded), base64.StdEncoding.EncodeToString([]byte(credential))) {
		t.Fatalf("public result leaked a configured secret: %s", encoded)
	}
	if result.ProviderResponses[0].Body.(map[string]any)["credential_echo"] != credential {
		t.Fatal("sanitizer mutated the complete internal receipt")
	}
}

func TestOperationDirectWriteResultOutputMasksConfiguredAndDeclaredSecrets(t *testing.T) {
	credential := "client-secret"
	numericCredential := "12345678901234567890"
	result := OperationDirectWriteResult{
		Connector: "fixture", Operation: "fixture.create", Method: "POST", Path: "/fixed", ResponseReceived: true, Status: 201,
		Headers:         map[string]WriteProviderHeader{"X-Echo": {Values: []string{credential, "ordinary"}}},
		BodyPresent:     true,
		BodyRaw:         `{"credential":"client-secret","numeric":12345678901234567890}`,
		BodyRawEncoding: "text",
		Body: map[string]any{
			"credential": "provider-issued-secret",
			"token":      "ordinary-occurrence-id",
			"echo":       credential,
			"numeric":    json.Number(numericCredential),
			"base64":     base64.StdEncoding.EncodeToString([]byte(credential)),
		},
		GraphQL:            &GraphQLResponseMetadata{Errors: []GraphQLResultError{{Message: credential}}},
		OutputSecretFields: []string{"credential"},
	}

	got := SanitizeOperationDirectWriteResultForOutput(result, map[string]string{"token": credential, "numeric": numericCredential})
	if reflect.DeepEqual(got, result) {
		t.Fatal("public direct-write result retained secret material")
	}
	if got.Body.(map[string]any)["token"] != "ordinary-occurrence-id" {
		t.Fatal("ordinary token-named provider identifier was changed")
	}
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{credential, numericCredential, "provider-issued-secret", base64.StdEncoding.EncodeToString([]byte(credential))} {
		if strings.Contains(string(encoded), secret) {
			t.Fatalf("public result leaked %q: %s", secret, encoded)
		}
	}
	if result.Body.(map[string]any)["credential"] != "provider-issued-secret" {
		t.Fatal("sanitizer mutated the complete internal receipt")
	}
}

func TestSanitizeWriteErrorForOutputKeepsSystemDiagnosticsSecretFree(t *testing.T) {
	credential := "client-secret"
	output := SanitizeWriteErrorForOutput(errors.New("system diagnostic "+credential), map[string]string{"token": credential})
	if strings.Contains(output, credential) {
		t.Fatal("system diagnostic leaked a configured credential")
	}
}

func TestPublicReceiptSanitizationMasksConcreteSecretRepresentationsWithoutChangingProviderNames(t *testing.T) {
	credential := "id"
	encoded := base64.StdEncoding.EncodeToString([]byte(credential))
	escaped, err := json.Marshal(credential)
	if err != nil {
		t.Fatal(err)
	}
	receipt := ProviderResponseReceipt{
		ResponseReceived: true,
		Status:           400,
		Headers: map[string]OperationResponseHeader{
			"X-occurrence-id": {Values: []string{"occurrence_id"}},
			"X-Diagnostic":    {Values: []string{"provider said " + credential + " while retaining occurrence_id"}},
		},
		BodyPresent:     true,
		BodyRaw:         `{"credential":` + string(escaped) + `,"occurrence_id":"occurrence-9007199254740993","token_type":"ordinary"}`,
		BodyRawEncoding: "text",
		Body: map[string]any{
			"credential":    credential,
			"encoded":       encoded,
			"occurrence_id": "occurrence-9007199254740993",
			"token_type":    "ordinary",
		},
	}

	got := SanitizeProviderResponseReceiptForOutput(receipt, map[string]string{"credential": credential})
	if strings.Contains(got.BodyRaw, `"credential":"id"`) || got.Headers["X-Diagnostic"].Values[0] != "provider said [masked] while retaining occurrence_id" {
		t.Fatal("public receipt retained a configured secret representation")
	}
	if _, ok := got.Headers["X-occurrence-id"]; !ok {
		t.Fatal("public receipt changed an ordinary provider header name")
	}
	body, ok := got.Body.(map[string]any)
	if !ok || body["credential"] != "[masked]" || body["encoded"] != "[masked]" || body["occurrence_id"] != "occurrence-9007199254740993" || body["token_type"] != "ordinary" {
		t.Fatal("public receipt changed ordinary provider output")
	}
	if !strings.Contains(got.BodyRaw, "occurrence-9007199254740993") {
		t.Fatal("public JSON receipt lost an ordinary occurrence identifier")
	}

	printable := ProviderResponseReceipt{BodyPresent: true, BodyRaw: "provider diagnostic id for occurrence_id", BodyRawEncoding: "text", Body: "provider diagnostic id for occurrence_id"}
	gotPrintable := SanitizeProviderResponseReceiptForOutput(printable, map[string]string{"credential": credential})
	if gotPrintable.BodyRaw != "provider diagnostic [masked] for occurrence_id" || gotPrintable.Body != "provider diagnostic [masked] for occurrence_id" {
		t.Fatalf("printable receipt = %#v, want concrete secret masking without changing occurrence_id", gotPrintable)
	}
}
