package connectors

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestSanitizeWriteResultForOutputPreservesProviderResponseExactly(t *testing.T) {
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
				credential:          "provider-key-preserved",
				"credential_echo":   credential,
				"numeric_echo":      json.Number(numericCredential),
				"base64_echo":       base64.StdEncoding.EncodeToString([]byte(credential)),
				"nested":            map[string]any{"value": credential},
				"ordinary_provider": true,
			},
		}},
	}

	got := SanitizeWriteResultForOutput(result, map[string]string{"token": credential, "numeric": numericCredential})
	if !reflect.DeepEqual(got, result) {
		t.Fatal("provider result changed during output serialization")
	}
}

func TestSanitizeOperationDirectWriteResultForOutputPreservesProviderResponseExactly(t *testing.T) {
	credential := "client-secret"
	numericCredential := "12345678901234567890"
	result := OperationDirectWriteResult{
		Connector: "fixture", Operation: "fixture.create", Method: "POST", Path: "/fixed", ResponseReceived: true, Status: 201,
		Headers:         map[string]WriteProviderHeader{"X-Echo": {Values: []string{credential, "ordinary"}}},
		BodyPresent:     true,
		BodyRaw:         `{"credential":"client-secret","numeric":12345678901234567890}`,
		BodyRawEncoding: "text",
		Body: map[string]any{
			credential: "provider-key-preserved",
			"echo":     credential,
			"numeric":  json.Number(numericCredential),
			"base64":   base64.StdEncoding.EncodeToString([]byte(credential)),
		},
		GraphQL:            &GraphQLResponseMetadata{Errors: []GraphQLResultError{{Message: credential}}},
		OutputSecretFields: []string{"credential"},
	}

	got := SanitizeOperationDirectWriteResultForOutput(result, map[string]string{"token": credential, "numeric": numericCredential})
	if !reflect.DeepEqual(got, result) {
		t.Fatal("provider direct-write result changed during output serialization")
	}
}

func TestSanitizeWriteErrorForOutputKeepsSystemDiagnosticsSecretFree(t *testing.T) {
	credential := "client-secret"
	output := SanitizeWriteErrorForOutput(errors.New("system diagnostic "+credential), map[string]string{"token": credential})
	if strings.Contains(output, credential) {
		t.Fatal("system diagnostic leaked a configured credential")
	}
}
