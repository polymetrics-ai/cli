package failures

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

type typedNilCause struct{}

func (*typedNilCause) Error() string {
	return "typed nil cause"
}

func TestClassificationValidatesDomainsRetryabilityAndSafeJSON(t *testing.T) {
	cause := errors.New("internal parser rejected declared ssl mode")
	tests := []struct {
		name      string
		domain    Domain
		retryable bool
	}{
		{name: "configuration", domain: DomainConfiguration, retryable: false},
		{name: "system", domain: DomainSystem, retryable: false},
		{name: "transient", domain: DomainTransient, retryable: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			classification, err := New(Input{
				Domain:     tt.domain,
				Code:       "invalid_sslmode",
				Message:    "sslmode is not accepted",
				FieldPath:  "/connection/sslmode",
				References: []Reference{{Kind: ReferenceKindConnector, Value: "postgres"}},
			}, cause)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			if got := classification.Retryable(); got != tt.retryable {
				t.Fatalf("Retryable() = %v, want %v", got, tt.retryable)
			}
			if !errors.Is(classification, cause) {
				t.Fatalf("classification does not retain cause %v", cause)
			}

			raw, err := json.Marshal(classification)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}
			if strings.Contains(string(raw), cause.Error()) || strings.Contains(string(raw), "cause") {
				t.Fatalf("serialized classification leaked cause: %s", raw)
			}
			var decoded Classification
			if err := json.Unmarshal(raw, &decoded); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}
			if got := decoded.Cause(); got != nil {
				t.Fatalf("decoded Cause() = %v, want nil", got)
			}
			if got := decoded.FieldPath(); got != "/connection/sslmode" {
				t.Fatalf("FieldPath() = %q", got)
			}
		})
	}
}

func TestClassificationRejectsInvalidStructuredValues(t *testing.T) {
	valid := Input{
		Domain:    DomainSystem,
		Code:      "dispatch_refused",
		Message:   "the command is not routable",
		FieldPath: "/operations/0",
	}
	tests := []struct {
		name  string
		input Input
	}{
		{name: "empty domain", input: Input{Code: valid.Code, Message: valid.Message}},
		{name: "unknown domain", input: Input{Domain: Domain("other"), Code: valid.Code, Message: valid.Message}},
		{name: "empty code", input: Input{Domain: valid.Domain, Message: valid.Message}},
		{name: "unsafe code", input: Input{Domain: valid.Domain, Code: "dispatch-refused", Message: valid.Message}},
		{name: "empty message", input: Input{Domain: valid.Domain, Code: valid.Code}},
		{name: "control message", input: Input{Domain: valid.Domain, Code: valid.Code, Message: "bad\nmessage"}},
		{name: "malformed pointer", input: Input{Domain: valid.Domain, Code: valid.Code, Message: valid.Message, FieldPath: "operations/0"}},
		{name: "malformed pointer escape", input: Input{Domain: valid.Domain, Code: valid.Code, Message: valid.Message, FieldPath: "/operations/~2"}},
		{name: "invalid UTF-8 pointer", input: Input{Domain: valid.Domain, Code: valid.Code, Message: valid.Message, FieldPath: "/\xff"}},
		{name: "unknown dispatch kind", input: Input{Domain: valid.Domain, Code: valid.Code, Message: valid.Message, DispatchKind: DispatchKind("unknown")}},
		{name: "dispatch kind outside system", input: Input{Domain: DomainConfiguration, Code: valid.Code, Message: valid.Message, DispatchKind: DispatchKindDirectStub}},
		{name: "unsafe reference", input: Input{Domain: valid.Domain, Code: valid.Code, Message: valid.Message, References: []Reference{{Kind: ReferenceKindSource, Value: "line\nbreak"}}}},
		{name: "request-shaped reference", input: Input{Domain: valid.Domain, Code: valid.Code, Message: valid.Message, References: []Reference{{Kind: ReferenceKindOperation, Value: "widgets?token=value"}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := New(tt.input, nil); err == nil {
				t.Fatal("New() error = nil, want invalid classification rejection")
			}
		})
	}
}

func TestClassificationAcceptsEveryDispatchKind(t *testing.T) {
	kinds := []DispatchKind{
		DispatchKindDirectStub,
		DispatchKindHelperDelegatedRefusal,
		DispatchKindWrappedTypedUnsupported,
		DispatchKindDeclaredButUnroutableCommand,
		DispatchKindUnresolvedDynamicTarget,
	}
	for _, kind := range kinds {
		t.Run(string(kind), func(t *testing.T) {
			classification, err := New(Input{
				Domain:       DomainSystem,
				Code:         "dispatch_not_executable",
				Message:      "the declared command does not reach an executor",
				DispatchKind: kind,
			}, nil)
			if err != nil {
				t.Fatalf("New(%q) error = %v", kind, err)
			}
			if got := classification.DispatchKind(); got != kind {
				t.Fatalf("DispatchKind() = %q, want %q", got, kind)
			}
		})
	}
}

func TestClassificationNormalizesTypedNilCause(t *testing.T) {
	var cause *typedNilCause
	classification, err := New(Input{
		Domain:  DomainSystem,
		Code:    "dispatch_refused",
		Message: "the command cannot be dispatched",
	}, cause)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if got := classification.Cause(); got != nil {
		t.Fatalf("Cause() = %T, want nil", got)
	}
	if got := classification.Unwrap(); got != nil {
		t.Fatalf("Unwrap() = %T, want nil", got)
	}
	var extracted *typedNilCause
	if errors.As(classification, &extracted) {
		t.Fatalf("errors.As() extracted = %T, want no typed nil cause", extracted)
	}
}

func TestClassificationUnmarshalRejectsInvalidWireCodes(t *testing.T) {
	for _, raw := range []string{
		`{"domain":"other","code":"valid_code","message":"invalid"}`,
		`{"domain":"configuration","code":"","message":"invalid"}`,
		`{"domain":"system","code":"valid_code","message":"invalid","dispatch_kind":"other"}`,
	} {
		var classification Classification
		if err := json.Unmarshal([]byte(raw), &classification); err == nil {
			t.Fatalf("json.Unmarshal(%s) error = nil, want invalid wire-code rejection", raw)
		}
	}
}

func TestClassificationUnmarshalRejectsInvalidUTF8FieldPath(t *testing.T) {
	raw := []byte("{\"domain\":\"system\",\"code\":\"dispatch_refused\",\"message\":\"the field path is invalid\",\"field_path\":\"/\xff\"}")
	var classification Classification
	if err := json.Unmarshal(raw, &classification); err == nil {
		t.Fatal("json.Unmarshal() error = nil, want invalid UTF-8 rejection")
	}
}

func TestClassificationUnmarshalValidatesJSONPointerSurrogateEscapes(t *testing.T) {
	tests := []struct {
		name      string
		fieldPath string
		wantPath  string
		wantError bool
	}{
		{name: "unpaired high surrogate", fieldPath: `"/\uD800"`, wantError: true},
		{name: "unpaired low surrogate", fieldPath: `"/\uDC00"`, wantError: true},
		{name: "paired surrogate", fieldPath: `"/\uD83D\uDE00"`, wantPath: "/😀"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := []byte(`{"domain":"system","code":"dispatch_refused","message":"the field path is invalid","field_path":` + tt.fieldPath + `}`)
			var classification Classification
			err := json.Unmarshal(raw, &classification)
			if tt.wantError {
				if err == nil {
					t.Fatal("json.Unmarshal() error = nil, want surrogate rejection")
				}
				return
			}
			if err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}
			if got := classification.FieldPath(); got != tt.wantPath {
				t.Fatalf("FieldPath() = %q, want %q", got, tt.wantPath)
			}
		})
	}
}
