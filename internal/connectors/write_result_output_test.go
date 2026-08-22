package connectors

import (
	"bytes"
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

func TestWriteResultOutputMasksConfiguredTextInsidePrintableReceiptBody(t *testing.T) {
	const credential = "opaque-provider-value-77"
	encodedCredential := base64.RawURLEncoding.EncodeToString([]byte(credential))
	diagnostic := "provider diagnostic " + credential + " encoded " + encodedCredential + " occurrence_id=occurrence-9007199254740993"
	const expectedDiagnostic = "provider diagnostic [masked] encoded [masked] occurrence_id=occurrence-9007199254740993"
	result := WriteResult{
		RecordsWritten: 1,
		ProviderResponses: []WriteProviderResponse{{
			Status:          201,
			BodyPresent:     true,
			BodyBytes:       len(diagnostic),
			BodyRaw:         diagnostic,
			BodyRawEncoding: "text",
			Body:            diagnostic,
			BodyEncoding:    "text",
		}},
	}

	got := SanitizeWriteResultForOutput(result, map[string]string{"credential": credential})
	public, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{credential, encodedCredential} {
		if strings.Contains(string(public), value) {
			t.Fatal("public write receipt exposed configured material")
		}
	}
	response := got.ProviderResponses[0]
	if response.BodyRaw != expectedDiagnostic {
		t.Fatal("public write receipt raw body did not mask configured material")
	}
	body, ok := response.Body.(string)
	if !ok || body != expectedDiagnostic {
		t.Fatal("public write receipt body did not mask configured material")
	}
	if result.ProviderResponses[0].Body != diagnostic {
		t.Fatal("sanitizer mutated the complete internal receipt")
	}
}

func TestOperationDirectWriteResultOutputMasksConfiguredTextInsideJSONDiagnostic(t *testing.T) {
	const credential = "opaque-provider-value-77"
	encodedCredential := base64.RawURLEncoding.EncodeToString([]byte(credential))
	diagnostic := "provider diagnostic " + credential + " encoded " + encodedCredential + " occurrence_id=occurrence-9007199254740993"
	const expectedDiagnostic = "provider diagnostic [masked] encoded [masked] occurrence_id=occurrence-9007199254740993"
	body := map[string]any{
		"diagnostic":    diagnostic,
		"occurrence_id": "occurrence-9007199254740993",
		"token_type":    "unconfigured-provider-token",
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	result := OperationDirectWriteResult{
		Connector:        "fixture",
		Operation:        "fixture.create",
		Method:           "POST",
		Path:             "/fixed",
		ResponseReceived: true,
		Status:           201,
		BodyPresent:      true,
		BodyRaw:          string(raw),
		BodyRawEncoding:  "text",
		Body:             body,
	}

	got := SanitizeOperationDirectWriteResultForOutput(result, map[string]string{"credential": credential})
	public, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{credential, encodedCredential} {
		if strings.Contains(string(public), value) {
			t.Fatal("public direct-write result exposed configured material")
		}
	}
	gotBody, ok := got.Body.(map[string]any)
	if !ok || gotBody["diagnostic"] != expectedDiagnostic {
		t.Fatal("public direct-write diagnostic did not mask configured material")
	}
	if gotBody["occurrence_id"] != "occurrence-9007199254740993" || gotBody["token_type"] != "unconfigured-provider-token" {
		t.Fatal("public direct-write result changed ordinary provider output")
	}
	if result.Body.(map[string]any)["diagnostic"] != diagnostic {
		t.Fatal("sanitizer mutated the complete internal receipt")
	}
}

func TestPublicWriteReceiptsMaskConfiguredBinaryBodies(t *testing.T) {
	const credential = "configured-binary-material"
	raw := []byte(`{"id":"ordinary-occurrence-id","occurrence_id":"occurrence-9007199254740993","duplicate":"first","duplicate":"second","credential":"`)
	raw = append(raw, 0xff)
	raw = append(raw, []byte(credential)...)
	raw = append(raw, 0xfe)
	raw = append(raw, []byte(`","token_type":"unconfigured-provider-token"}`)...)
	expected := []byte(`{"id":"ordinary-occurrence-id","occurrence_id":"occurrence-9007199254740993","duplicate":"first","duplicate":"second","credential":"`)
	expected = append(expected, 0xff)
	expected = append(expected, []byte("[masked]")...)
	expected = append(expected, 0xfe)
	expected = append(expected, []byte(`","token_type":"unconfigured-provider-token"}`)...)
	encoded := base64.StdEncoding.EncodeToString(raw)

	tests := []struct {
		name     string
		sanitize func() (string, any)
	}{
		{
			name: "write receipt",
			sanitize: func() (string, any) {
				result := WriteResult{ProviderResponses: []WriteProviderResponse{{
					BodyPresent:     true,
					BodyBytes:       len(raw),
					BodyRaw:         encoded,
					BodyRawEncoding: "base64",
					Body:            encoded,
					BodyEncoding:    "base64",
				}}}
				got := SanitizeWriteResultForOutput(result, map[string]string{"key_collision": "id", "credential": credential})
				return got.ProviderResponses[0].BodyRaw, got.ProviderResponses[0].Body
			},
		},
		{
			name: "direct operation receipt",
			sanitize: func() (string, any) {
				result := OperationDirectWriteResult{
					Connector:        "fixture",
					Operation:        "fixture.create",
					Method:           "POST",
					Path:             "/fixed",
					ResponseReceived: true,
					Status:           201,
					BodyPresent:      true,
					BodyBytes:        len(raw),
					BodyRaw:          encoded,
					BodyRawEncoding:  "base64",
					Body:             encoded,
				}
				got := SanitizeOperationDirectWriteResultForOutput(result, map[string]string{"key_collision": "id", "credential": credential})
				return got.BodyRaw, got.Body
			},
		},
		{
			name: "direct read receipt",
			sanitize: func() (string, any) {
				receipt := ProviderResponseReceipt{
					ResponseReceived: true,
					Status:           200,
					BodyPresent:      true,
					BodyBytes:        int64(len(raw)),
					BodyRaw:          encoded,
					BodyRawEncoding:  "base64",
					Body:             encoded,
				}
				got := SanitizeProviderResponseReceiptForOutput(receipt, map[string]string{"key_collision": "id", "credential": credential})
				return got.BodyRaw, got.Body
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			bodyRaw, body := testCase.sanitize()
			decodedRaw, err := base64.StdEncoding.DecodeString(bodyRaw)
			if err != nil || !bytes.Equal(decodedRaw, expected) || bytes.Contains(decodedRaw, []byte(credential)) {
				t.Fatal("public receipt raw body did not preserve non-secret bytes while masking configured material")
			}
			bodyString, ok := body.(string)
			if !ok {
				t.Fatal("public receipt body did not retain its base64 representation")
			}
			decodedBody, err := base64.StdEncoding.DecodeString(bodyString)
			if err != nil || !bytes.Equal(decodedBody, expected) || bytes.Contains(decodedBody, []byte(credential)) {
				t.Fatal("public receipt body did not preserve non-secret bytes while masking configured material")
			}
		})
	}
}

func TestPublicWriteReceiptsPreserveRawJSONFieldNames(t *testing.T) {
	const credential = "configured-text-material"
	const raw = "{\n  \"id\" : \"ordinary-occurrence-id\",\n\t\"id\":\"ordinary-occurrence-id\",\n  \"diagnostic\" : \"provider " + credential + "\",\n  \"token_type\" : \"unconfigured-provider-token\"\n}"
	const expected = "{\n  \"id\" : \"ordinary-occurrence-id\",\n\t\"id\":\"ordinary-occurrence-id\",\n  \"diagnostic\" : \"provider [masked]\",\n  \"token_type\" : \"unconfigured-provider-token\"\n}"

	tests := []struct {
		name     string
		sanitize func() (string, any)
	}{
		{
			name: "write receipt",
			sanitize: func() (string, any) {
				result := WriteResult{ProviderResponses: []WriteProviderResponse{{
					BodyPresent:     true,
					BodyBytes:       len(raw),
					BodyRaw:         raw,
					BodyRawEncoding: "text",
					Body:            raw,
					BodyEncoding:    "text",
				}}}
				got := SanitizeWriteResultForOutput(result, map[string]string{"key_collision": "id", "credential": credential})
				return got.ProviderResponses[0].BodyRaw, got.ProviderResponses[0].Body
			},
		},
		{
			name: "direct operation receipt",
			sanitize: func() (string, any) {
				result := OperationDirectWriteResult{
					Connector:        "fixture",
					Operation:        "fixture.create",
					Method:           "POST",
					Path:             "/fixed",
					ResponseReceived: true,
					Status:           201,
					BodyPresent:      true,
					BodyBytes:        len(raw),
					BodyRaw:          raw,
					BodyRawEncoding:  "text",
					Body:             raw,
				}
				got := SanitizeOperationDirectWriteResultForOutput(result, map[string]string{"key_collision": "id", "credential": credential})
				return got.BodyRaw, got.Body
			},
		},
		{
			name: "direct read receipt",
			sanitize: func() (string, any) {
				receipt := ProviderResponseReceipt{
					ResponseReceived: true,
					Status:           200,
					BodyPresent:      true,
					BodyBytes:        int64(len(raw)),
					BodyRaw:          raw,
					BodyRawEncoding:  "text",
					Body:             raw,
				}
				got := SanitizeProviderResponseReceiptForOutput(receipt, map[string]string{"key_collision": "id", "credential": credential})
				return got.BodyRaw, got.Body
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			bodyRaw, body := testCase.sanitize()
			if bodyRaw != expected || strings.Contains(bodyRaw, credential) {
				t.Fatal("public receipt raw JSON did not preserve field names while masking configured material")
			}
			bodyString, ok := body.(string)
			if !ok || bodyString != expected || strings.Contains(bodyString, credential) {
				t.Fatal("public receipt JSON body did not preserve field names while masking configured material")
			}
		})
	}
}

func TestPublicWriteReceiptsMaskEscapedJSONValuesWithoutChangingOtherBytes(t *testing.T) {
	const credential = "secret"
	const escapedCredential = `\u0073\u0065\u0063\u0072\u0065\u0074`
	const raw = "{\n  \"id\" : \"ordinary-occurrence-id\",\n  \"diagnostic\" : \"provider " + escapedCredential + " occurrence_id=occurrence-9007199254740993\",\n  \"token_type\" : \"unconfigured-provider-token\"\n}"
	const expected = "{\n  \"id\" : \"ordinary-occurrence-id\",\n  \"diagnostic\" : \"provider [masked] occurrence_id=occurrence-9007199254740993\",\n  \"token_type\" : \"unconfigured-provider-token\"\n}"

	tests := []struct {
		name     string
		raw      string
		encoding string
		expected string
		sanitize func(string, string) (string, any)
	}{
		{
			name:     "text write receipt",
			raw:      raw,
			encoding: "text",
			expected: expected,
			sanitize: func(value, encoding string) (string, any) {
				result := WriteResult{ProviderResponses: []WriteProviderResponse{{BodyPresent: true, BodyBytes: len(raw), BodyRaw: value, BodyRawEncoding: encoding, Body: value, BodyEncoding: encoding}}}
				got := SanitizeWriteResultForOutput(result, map[string]string{"key_collision": "id", "credential": credential})
				return got.ProviderResponses[0].BodyRaw, got.ProviderResponses[0].Body
			},
		},
		{
			name:     "base64 write receipt",
			raw:      base64.StdEncoding.EncodeToString([]byte(raw)),
			encoding: "base64",
			expected: base64.StdEncoding.EncodeToString([]byte(expected)),
			sanitize: func(value, encoding string) (string, any) {
				result := WriteResult{ProviderResponses: []WriteProviderResponse{{BodyPresent: true, BodyBytes: len(raw), BodyRaw: value, BodyRawEncoding: encoding, Body: value, BodyEncoding: encoding}}}
				got := SanitizeWriteResultForOutput(result, map[string]string{"key_collision": "id", "credential": credential})
				return got.ProviderResponses[0].BodyRaw, got.ProviderResponses[0].Body
			},
		},
		{
			name:     "text direct operation",
			raw:      raw,
			encoding: "text",
			expected: expected,
			sanitize: func(value, encoding string) (string, any) {
				result := OperationDirectWriteResult{BodyPresent: true, BodyBytes: len(raw), BodyRaw: value, BodyRawEncoding: encoding, Body: value}
				got := SanitizeOperationDirectWriteResultForOutput(result, map[string]string{"key_collision": "id", "credential": credential})
				return got.BodyRaw, got.Body
			},
		},
		{
			name:     "base64 direct operation",
			raw:      base64.StdEncoding.EncodeToString([]byte(raw)),
			encoding: "base64",
			expected: base64.StdEncoding.EncodeToString([]byte(expected)),
			sanitize: func(value, encoding string) (string, any) {
				result := OperationDirectWriteResult{BodyPresent: true, BodyBytes: len(raw), BodyRaw: value, BodyRawEncoding: encoding, Body: value}
				got := SanitizeOperationDirectWriteResultForOutput(result, map[string]string{"key_collision": "id", "credential": credential})
				return got.BodyRaw, got.Body
			},
		},
		{
			name:     "text direct read receipt",
			raw:      raw,
			encoding: "text",
			expected: expected,
			sanitize: func(value, encoding string) (string, any) {
				receipt := ProviderResponseReceipt{BodyPresent: true, BodyBytes: int64(len(raw)), BodyRaw: value, BodyRawEncoding: encoding, Body: value}
				got := SanitizeProviderResponseReceiptForOutput(receipt, map[string]string{"key_collision": "id", "credential": credential})
				return got.BodyRaw, got.Body
			},
		},
		{
			name:     "base64 direct read receipt",
			raw:      base64.StdEncoding.EncodeToString([]byte(raw)),
			encoding: "base64",
			expected: base64.StdEncoding.EncodeToString([]byte(expected)),
			sanitize: func(value, encoding string) (string, any) {
				receipt := ProviderResponseReceipt{BodyPresent: true, BodyBytes: int64(len(raw)), BodyRaw: value, BodyRawEncoding: encoding, Body: value}
				got := SanitizeProviderResponseReceiptForOutput(receipt, map[string]string{"key_collision": "id", "credential": credential})
				return got.BodyRaw, got.Body
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bodyRaw, body := test.sanitize(test.raw, test.encoding)
			bodyText, ok := body.(string)
			if !ok || bodyRaw != test.expected || bodyText != test.expected {
				t.Fatalf("sanitized receipt = raw %q body %#v, want byte-preserving %q", bodyRaw, body, test.expected)
			}
			if strings.Contains(bodyRaw, escapedCredential) {
				t.Fatal("sanitized receipt retained the reversible escaped credential")
			}
		})
	}
}

func TestPublicReceiptsMaskEscapedValuesInMalformedOrMultiValueJSON(t *testing.T) {
	const credential = "secret"
	const escapedCredential = `\u0073\u0065\u0063\u0072\u0065\u0074`
	tests := []struct {
		name     string
		raw      string
		expected string
	}{
		{
			name:     "malformed object",
			raw:      `{"secret":"ordinary-occurrence-id","diagnostic":"provider ` + escapedCredential + ` occurrence_id=occurrence-9007199254740993"`,
			expected: `{"secret":"ordinary-occurrence-id","diagnostic":"provider [masked] occurrence_id=occurrence-9007199254740993"`,
		},
		{
			name:     "multiple objects",
			raw:      `{"secret":"ordinary-occurrence-id"}` + "\n" + `{"diagnostic":"provider ` + escapedCredential + ` occurrence_id=occurrence-9007199254740993"}`,
			expected: `{"secret":"ordinary-occurrence-id"}` + "\n" + `{"diagnostic":"provider [masked] occurrence_id=occurrence-9007199254740993"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := SanitizeProviderResponseReceiptForOutput(ProviderResponseReceipt{
				BodyPresent:     true,
				BodyBytes:       int64(len(test.raw)),
				BodyRaw:         test.raw,
				BodyRawEncoding: "text",
				Body:            test.raw,
			}, map[string]string{"credential": credential})
			body, ok := got.Body.(string)
			if !ok || got.BodyRaw != test.expected || body != test.expected {
				t.Fatalf("sanitized receipt = raw %q body %#v, want %q", got.BodyRaw, got.Body, test.expected)
			}
			if strings.Contains(got.BodyRaw, escapedCredential) {
				t.Fatal("sanitized receipt retained the escaped configured value")
			}
		})
	}
}

func TestPublicReceiptsMaskNonJSONQuotedSecretsAndPreserveJSONKeys(t *testing.T) {
	const credential = "configured-secret"
	const occurrenceID = "occurrence-9007199254740993"
	tests := []struct {
		name     string
		raw      string
		expected string
		mustMask bool
	}{
		{
			name:     "non JSON diagnostic and escaped authorization envelope",
			raw:      `provider said "configured-secret": denied; Authorization: Bearer \u003cconfigured-secret\u003e; occurrence_id=occurrence-9007199254740993`,
			expected: `provider said "[masked]": denied; Authorization: Bearer \u003c[masked]\u003e; occurrence_id=occurrence-9007199254740993`,
			mustMask: true,
		},
		{
			name:     "non JSON object lookalike",
			raw:      `provider said { not-json, "configured-secret": denied; occurrence_id=occurrence-9007199254740993`,
			expected: `provider said { not-json, "[masked]": denied; occurrence_id=occurrence-9007199254740993`,
			mustMask: true,
		},
		{
			name:     "true object member key",
			raw:      `{"configured-secret":"ordinary-occurrence-id","occurrence_id":"occurrence-9007199254740993","token_type":"unconfigured-provider-token"}`,
			expected: `{"configured-secret":"ordinary-occurrence-id","occurrence_id":"occurrence-9007199254740993","token_type":"unconfigured-provider-token"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := SanitizeProviderResponseReceiptForOutput(ProviderResponseReceipt{
				BodyPresent:     true,
				BodyBytes:       int64(len(test.raw)),
				BodyRaw:         test.raw,
				BodyRawEncoding: "text",
				Body:            test.raw,
			}, map[string]string{"credential": credential})
			body, ok := got.Body.(string)
			if !ok || got.BodyRaw != test.expected || body != test.expected {
				t.Fatalf("sanitized receipt = raw %q body %#v, want %q", got.BodyRaw, got.Body, test.expected)
			}
			if test.mustMask && strings.Contains(got.BodyRaw, credential) {
				t.Fatalf("sanitized non-JSON diagnostic retained configured material: %q", got.BodyRaw)
			}
			if !strings.Contains(got.BodyRaw, occurrenceID) {
				t.Fatalf("sanitized receipt changed ordinary occurrence ID: %q", got.BodyRaw)
			}
		})
	}
}

func TestPublicResponseHeaderSanitizationMasksOnlyExactConfiguredValues(t *testing.T) {
	const credential = "configured-header-material"
	const occurrenceID = "occurrence-9007199254740993"
	const unconfiguredToken = "ghp_unconfigured_provider_token"
	encoded := base64.StdEncoding.EncodeToString([]byte(credential))
	newHeaders := func() map[string]OperationResponseHeader {
		return map[string]OperationResponseHeader{
			credential: {Values: []string{occurrenceID}},
			"X-Provider-Metadata": {Values: []string{
				"first",
				credential,
				encoded,
				occurrenceID,
				unconfiguredToken,
				"prefix-" + credential,
				credential,
			}},
		}
	}
	want := []string{"first", "[masked]", "[masked]", occurrenceID, unconfiguredToken, "prefix-" + credential, "[masked]"}
	tests := []struct {
		name     string
		sanitize func(map[string]OperationResponseHeader) map[string]OperationResponseHeader
	}{
		{
			name: "write receipt",
			sanitize: func(headers map[string]OperationResponseHeader) map[string]OperationResponseHeader {
				result := WriteResult{ProviderResponses: []WriteProviderResponse{{Headers: headers}}}
				return SanitizeWriteResultForOutput(result, map[string]string{"credential": credential}).ProviderResponses[0].Headers
			},
		},
		{
			name: "direct operation",
			sanitize: func(headers map[string]OperationResponseHeader) map[string]OperationResponseHeader {
				return SanitizeOperationDirectWriteResultForOutput(OperationDirectWriteResult{Headers: headers}, map[string]string{"credential": credential}).Headers
			},
		},
		{
			name: "provider receipt",
			sanitize: func(headers map[string]OperationResponseHeader) map[string]OperationResponseHeader {
				return SanitizeProviderResponseReceiptForOutput(ProviderResponseReceipt{Headers: headers}, map[string]string{"credential": credential}).Headers
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.sanitize(newHeaders())
			if !reflect.DeepEqual(got["X-Provider-Metadata"].Values, want) {
				t.Fatalf("header values = %#v, want %#v", got["X-Provider-Metadata"].Values, want)
			}
			if got[credential].Values[0] != occurrenceID {
				t.Fatalf("configured header name was changed: %#v", got)
			}
		})
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
			"string_map":    map[string]string{"credential": credential, "occurrence_id": "occurrence-9007199254740993"},
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
	stringMap, ok := body["string_map"].(map[string]string)
	if !ok || stringMap["credential"] != "[masked]" || stringMap["occurrence_id"] != "occurrence-9007199254740993" {
		t.Fatalf("public receipt string map = %#v, want concrete masking without changing provider IDs", body["string_map"])
	}
	if !strings.Contains(got.BodyRaw, "occurrence-9007199254740993") {
		t.Fatal("public JSON receipt lost an ordinary occurrence identifier")
	}

	printable := ProviderResponseReceipt{BodyPresent: true, BodyRaw: "provider diagnostic id for occurrence_id", BodyRawEncoding: "text", Body: "provider diagnostic id for occurrence_id"}
	gotPrintable := SanitizeProviderResponseReceiptForOutput(printable, map[string]string{"credential": credential})
	if gotPrintable.BodyRaw != "provider diagnostic [masked] for occurrence_id" || gotPrintable.Body != "provider diagnostic [masked] for occurrence_id" {
		t.Fatalf("printable receipt = %#v, want concrete secret masking without changing occurrence_id", gotPrintable)
	}

	page := SanitizeDirectReadPageForOutput(DirectReadPage{NextCursor: credential}, map[string]string{"credential": credential})
	ordinaryPage := SanitizeDirectReadPageForOutput(DirectReadPage{NextCursor: "occurrence_id"}, map[string]string{"credential": credential})
	if page.NextCursor != "[masked]" || ordinaryPage.NextCursor != "occurrence_id" {
		t.Fatalf("public pages = %#v, %#v, want exact secret masking with ordinary cursor preservation", page, ordinaryPage)
	}
}

func TestPublicReceiptSanitizationMasksConfiguredTextInsideStructuredDiagnostics(t *testing.T) {
	const credential = "opaque-provider-value-77"
	encoded := base64.RawURLEncoding.EncodeToString([]byte(credential))
	diagnostic := "provider diagnostic " + credential + " encoded " + encoded + " occurrence_id=occurrence-9007199254740993"
	body := map[string]any{
		"diagnostic":    diagnostic,
		"occurrence_id": "occurrence-9007199254740993",
		"token_type":    "unconfigured-provider-token",
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	got := SanitizeProviderResponseReceiptForOutput(ProviderResponseReceipt{
		BodyPresent:     true,
		BodyRaw:         string(raw),
		BodyRawEncoding: "text",
		Body:            body,
	}, map[string]string{"credential": credential})
	public, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{credential, encoded} {
		if strings.Contains(string(public), value) {
			t.Fatal("public receipt exposed configured material")
		}
	}
	gotBody, ok := got.Body.(map[string]any)
	if !ok {
		t.Fatalf("receipt body type = %T, want map", got.Body)
	}
	if gotBody["diagnostic"] != "provider diagnostic [masked] encoded [masked] occurrence_id=occurrence-9007199254740993" {
		t.Fatal("receipt diagnostic did not mask configured material")
	}
	if gotBody["occurrence_id"] != "occurrence-9007199254740993" || gotBody["token_type"] != "unconfigured-provider-token" {
		t.Fatalf("receipt body changed ordinary provider values: %#v", gotBody)
	}
}
