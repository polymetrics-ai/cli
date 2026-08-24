package connectors

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestWriteResultOutputPreservesProviderValuesEqualToConfiguredCredentials(t *testing.T) {
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
	if !reflect.DeepEqual(got, result) {
		t.Fatalf("public provider result = %#v, want exact provider values", got)
	}
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{credential, numericCredential, base64.StdEncoding.EncodeToString([]byte(credential))} {
		if !strings.Contains(string(encoded), value) {
			t.Fatalf("public result did not retain provider value %q: %s", value, encoded)
		}
	}
	got.ProviderResponses[0].Body.(map[string]any)["nested"].(map[string]any)["value"] = "changed"
	header := got.ProviderResponses[0].Headers["X-Provider-Echo"]
	header.Values[0] = "changed"
	got.ProviderResponses[0].Headers["X-Provider-Echo"] = header
	if result.ProviderResponses[0].Body.(map[string]any)["nested"].(map[string]any)["value"] != credential || result.ProviderResponses[0].Headers["X-Provider-Echo"].Values[0] != credential {
		t.Fatal("sanitizer mutated the complete internal receipt")
	}
}

func TestOperationDirectWriteResultOutputMasksDeclaredSecretsAndPreservesConfiguredEqualProviderValues(t *testing.T) {
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
	body := got.Body.(map[string]any)
	if body["credential"] != "[masked]" {
		t.Fatalf("declared provider secret = %#v, want masked", body["credential"])
	}
	if body["token"] != "ordinary-occurrence-id" || body["echo"] != credential || body["numeric"] != json.Number(numericCredential) || body["base64"] != base64.StdEncoding.EncodeToString([]byte(credential)) {
		t.Fatalf("provider result did not retain undeclared values: %#v", body)
	}
	if got.Headers["X-Echo"].Values[0] != credential {
		t.Fatalf("provider header = %#v, want configured-equal value", got.Headers)
	}
	if got.GraphQL == nil || got.GraphQL.Errors[0].Message != credential {
		t.Fatalf("GraphQL result = %#v, want configured-equal provider value", got.GraphQL)
	}
	if strings.Contains(got.BodyRaw, `"credential":"client-secret"`) || !strings.Contains(got.BodyRaw, `"credential":"[masked]"`) || !strings.Contains(got.BodyRaw, numericCredential) {
		t.Fatalf("declared raw masking = %q", got.BodyRaw)
	}
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "provider-issued-secret") {
		t.Fatalf("public result retained declared secret: %s", encoded)
	}
	for _, value := range []string{credential, numericCredential, base64.StdEncoding.EncodeToString([]byte(credential))} {
		if !strings.Contains(string(encoded), value) {
			t.Fatalf("public result did not retain provider value %q: %s", value, encoded)
		}
	}
	if result.Body.(map[string]any)["credential"] != "provider-issued-secret" {
		t.Fatal("sanitizer mutated the complete internal receipt")
	}
}

func TestOperationDirectWriteResultOutputPreservesRequestSensitiveEchoWithoutOutputDeclaration(t *testing.T) {
	requestSecret := "declaration-owned-request-secret"
	result := OperationDirectWriteResult{
		Connector: "fixture", Operation: "fixture.create", Method: "POST", Path: "/fixed", ResponseReceived: true, Status: 400,
		BodyPresent: true, BodyRaw: `{"error":"declaration-owned-request-secret"}`, BodyRawEncoding: "text",
		Body:                   map[string]any{"error": requestSecret},
		RequestSensitiveValues: []string{requestSecret},
	}

	got := SanitizeOperationDirectWriteResultForOutput(result, map[string]string{"token": requestSecret})
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), requestSecret) || got.Body.(map[string]any)["error"] != requestSecret || got.BodyRaw != `{"error":"declaration-owned-request-secret"}` {
		t.Fatalf("public direct-write receipt did not retain undeclared provider echo: %s", encoded)
	}
	if result.Body.(map[string]any)["error"] != requestSecret {
		t.Fatal("sanitizer mutated the immutable internal receipt")
	}
}

func TestSanitizeWriteErrorForOutputKeepsSystemDiagnosticsSecretFree(t *testing.T) {
	credential := "client-secret"
	output := SanitizeWriteErrorForOutput(errors.New("system diagnostic "+credential), map[string]string{"token": credential})
	if strings.Contains(output, credential) {
		t.Fatal("system diagnostic leaked a configured credential")
	}
}

func TestPublicReceiptProjectionPreservesProviderValuesEqualToConfiguredCredentials(t *testing.T) {
	// Short values are the regression boundary: a public projection must never
	// turn provider-owned keys such as occurrence_id into fabricated keys merely
	// because a configured credential happens to be "id".
	shortSecret := "id"
	escapedSecret := "pa\"ss<>&日本語"
	receipt := ProviderResponseReceipt{
		ResponseReceived: true,
		Status:           502,
		Headers: map[string]OperationResponseHeader{
			"WWW-Authenticate": {Values: []string{"Unknown token type", "Basic realm=ordinary"}},
			"X-Occurrence-Id":  {Values: []string{"occurrence_id"}},
		},
		BodyPresent:     true,
		BodyBytes:       int64(len(`{"credential":"pa\"ss<>&日本語","occurrence_id":"provider-id","trained_tokens":7}`)),
		BodyRaw:         `{"credential":"pa\"ss<>&日本語","occurrence_id":"provider-id","trained_tokens":7}`,
		BodyRawEncoding: "text",
		Body: map[string]any{
			"credential":     escapedSecret,
			"occurrence_id":  "provider-id",
			"trained_tokens": json.Number("7"),
		},
	}

	public := SanitizeProviderResponseReceiptForOutput(receipt, map[string]string{"short": shortSecret, "escaped": escapedSecret})
	if !reflect.DeepEqual(public, receipt) {
		t.Fatalf("public receipt = %#v, want exact provider values", public)
	}
	if _, ok := public.Body.(map[string]any)["occurrence_id"]; !ok {
		t.Fatalf("public body rewrote provider key: %#v", public.Body)
	}
	if got := public.Body.(map[string]any)["occurrence_id"]; got != "provider-id" {
		t.Fatalf("public occurrence id = %#v, want provider-owned value", got)
	}
	if got := public.Headers["WWW-Authenticate"].Values; !reflect.DeepEqual(got, []string{"Unknown token type", "Basic realm=ordinary"}) {
		t.Fatalf("public WWW-Authenticate = %#v, want exact ordinary provider metadata", got)
	}
	if _, ok := public.Headers["X-Occurrence-Id"]; !ok {
		t.Fatalf("public receipt rewrote provider header name: %#v", public.Headers)
	}
	if public.Body.(map[string]any)["credential"] != escapedSecret || public.BodyRaw != receipt.BodyRaw {
		t.Fatalf("public receipt did not retain configured-equal provider evidence: %#v", public)
	}
}

func TestPublicReceiptProjectionPreservesRawJSONBytesWhenNoMaskApplies(t *testing.T) {
	raw := "{\n  \"occurrence_id\" : \"provider-id-9007199254740993\",\n  \"escaped\" : \"quote\\\\slash\\u003cprovider\\u003e\"\n}"
	receipt := ProviderResponseReceipt{
		ResponseReceived: true,
		Status:           200,
		BodyPresent:      true,
		BodyBytes:        int64(len(raw)),
		BodyRaw:          raw,
		BodyRawEncoding:  "text",
		Body: map[string]any{
			"occurrence_id": "provider-id-9007199254740993",
			"escaped":       "quote\\slash<provider>",
		},
	}

	public := SanitizeProviderResponseReceiptForOutput(receipt, map[string]string{"credential": "unrelated-secret"})
	if public.BodyRaw != raw {
		t.Fatalf("public raw receipt = %q, want byte-for-byte unchanged %q", public.BodyRaw, raw)
	}
	if receipt.BodyRaw != raw {
		t.Fatalf("public projection mutated internal receipt raw bytes: %q", receipt.BodyRaw)
	}

	binaryRaw := []byte{0x00, 'p', 'r', 'o', 'v', 'i', 'd', 'e', 'r', '-', 'i', 'd', 0xff}
	binaryEncoded := base64.StdEncoding.EncodeToString(binaryRaw)
	binaryReceipt := ProviderResponseReceipt{
		ResponseReceived: true,
		Status:           200,
		BodyPresent:      true,
		BodyBytes:        int64(len(binaryRaw)),
		BodyRaw:          binaryEncoded,
		BodyRawEncoding:  "base64",
		Body:             binaryEncoded,
	}
	binaryPublic := SanitizeProviderResponseReceiptForOutput(binaryReceipt, map[string]string{"credential": "id"})
	if binaryPublic.BodyRaw != binaryEncoded || binaryPublic.Body != binaryEncoded {
		t.Fatalf("public opaque bytes = %#v, want exact %q", binaryPublic, binaryEncoded)
	}
}

func TestPublicDirectReadReceiptMasksOnlyDeclaredZendeskValueLocations(t *testing.T) {
	result := DirectReadResult{
		Operation:          "zendesk-support.list_oauth_tokens",
		OutputSecretFields: []string{"tokens.token", "tokens.refresh_token"},
		Receipt: &ProviderResponseReceipt{
			ResponseReceived: true,
			Status:           200,
			BodyPresent:      true,
			BodyRaw:          `{"tokens":[{"token":"provider-access-secret","refresh_token":"provider-refresh-secret","occurrence_id":"provider-id"}]}`,
			BodyRawEncoding:  "text",
			Body: map[string]any{"tokens": []any{map[string]any{
				"token": "provider-access-secret", "refresh_token": "provider-refresh-secret", "occurrence_id": "provider-id",
			}}},
		},
		GraphQL: &GraphQLResponseMetadata{Errors: []GraphQLResultError{{Message: "provider-access-secret"}}},
	}

	public := SanitizeDirectReadResultForOutput(result, nil)
	if public.Receipt == nil {
		t.Fatal("public direct read omitted receipt")
	}
	encoded, err := json.Marshal(public)
	if err != nil {
		t.Fatalf("marshal public direct read: %v", err)
	}
	for _, secret := range []string{"provider-access-secret", "provider-refresh-secret"} {
		if strings.Contains(string(encoded), secret) {
			t.Fatalf("public direct receipt leaked declared secret %q: %s", secret, encoded)
		}
	}
	tokens := public.Receipt.Body.(map[string]any)["tokens"].([]any)
	row := tokens[0].(map[string]any)
	if row["token"] != "[masked]" || row["refresh_token"] != "[masked]" || row["occurrence_id"] != "provider-id" {
		t.Fatalf("declared receipt masking = %#v, want only named scalar values masked", row)
	}
	if public.GraphQL == nil || len(public.GraphQL.Errors) != 1 || public.GraphQL.Errors[0].Message != "[masked]" {
		t.Fatalf("declared GraphQL error masking = %#v, want the exact classified scalar masked", public.GraphQL)
	}
	if result.Receipt.Body.(map[string]any)["tokens"].([]any)[0].(map[string]any)["token"] != "provider-access-secret" {
		t.Fatalf("public projection mutated immutable internal receipt: %#v", result.Receipt)
	}
}
