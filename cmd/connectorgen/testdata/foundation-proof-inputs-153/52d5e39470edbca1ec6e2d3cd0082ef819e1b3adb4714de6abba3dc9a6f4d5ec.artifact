package engine

import (
	"strings"
	"testing"
)

func TestOperationKindContractsAlignLoaderAndParameterImport(t *testing.T) {
	ops := []OperationSpec{
		{ID: "acme.status", Kind: "rest_status", Summary: "status", Risk: "low", Approval: "none", OutputPolicy: "status", REST: &RESTOperationSpec{Method: "HEAD", Path: "/status", MaxBytes: 1}},
		{ID: "acme.text", Kind: "text_export", Summary: "export", Risk: "low", Approval: "none", OutputPolicy: "file_manifest", Binary: &BinaryOperationSpec{Method: "GET", Path: "/export", MaxBytes: 1, Accept: "text/csv", ContentTypes: []string{"text/csv"}, Charset: "utf-8", Response: &OperationResponseSpec{SuccessStatuses: []string{"200"}}}},
	}
	if err := validateOperations(ops); err != nil {
		t.Fatalf("validateOperations: %v", err)
	}
	for _, kind := range []string{"rest_status", "text_export"} {
		block, ok := OperationRequestParameterBlock(kind)
		if !ok || block == "" {
			t.Fatalf("parameter import contract for %q = %q/%t, want registered block", kind, block, ok)
		}
	}
}

func TestOperationHeaderDeclarationsCannotOverrideRuntimeHeaders(t *testing.T) {
	err := validateOperationRuntimeHeaderIsolation(HTTPBase{Headers: map[string]string{"X-Runtime-Token": "fixed"}}, []OperationSpec{{
		ID:   "acme.read",
		REST: &RESTOperationSpec{Parameters: []OperationParameter{{Name: "x_runtime_token", In: "header"}}},
	}})
	if err == nil || !strings.Contains(err.Error(), "runtime-owned") {
		t.Fatalf("validateOperationRuntimeHeaderIsolation error = %v, want runtime header collision", err)
	}
}

func TestOperationHeaderRepeatabilityExcludesPaginationParameters(t *testing.T) {
	err := validateOperationHeaderParameters(OperationSpec{REST: &RESTOperationSpec{
		PaginationParameters: []OperationParameter{{Name: "page", In: "query", Repeatable: true}},
	}})
	if err == nil || !strings.Contains(err.Error(), "pagination parameter") {
		t.Fatalf("validateOperationHeaderParameters error = %v, want pagination repeatability refusal", err)
	}
}

// TestOperationHeaderDeclarationsRejectRuntimeOwnedIdempotencyNames proves a
// preview-bound operation cannot publish a declaration-owned idempotency
// header that the retry transport later silently removes. These names stay
// runtime-owned until their retry semantics are deliberately modelled.
func TestOperationHeaderDeclarationsRejectRuntimeOwnedIdempotencyNames(t *testing.T) {
	for _, name := range []string{"Idempotency-Key", "X-Idempotency-Key"} {
		t.Run(name, func(t *testing.T) {
			err := validateOperationHeaderParameters(OperationSpec{REST: &RESTOperationSpec{Parameters: []OperationParameter{{
				Name:     name,
				In:       "header",
				Type:     "string",
				MaxBytes: 128,
				Schema:   []byte(`{"type":"string","maxLength":128}`),
			}}}})
			if err == nil || !strings.Contains(err.Error(), "runtime-owned") {
				t.Fatalf("validateOperationHeaderParameters(%q) error = %v, want runtime-owned refusal", name, err)
			}
		})
	}
}

func TestReadSurfaceContracts(t *testing.T) {
	for _, intent := range []string{"direct_read", "binary_download", "text_export", "status_check"} {
		if !IsReadSurfaceIntent(intent) {
			t.Fatalf("IsReadSurfaceIntent(%q) = false, want true", intent)
		}
	}
	if IsReadSurfaceIntent("direct_write") {
		t.Fatal("IsReadSurfaceIntent(direct_write) = true, want false")
	}

	for _, method := range []string{"GET", "HEAD", "POST"} {
		if !IsReadSurfaceMethod(method) {
			t.Fatalf("IsReadSurfaceMethod(%q) = false, want true", method)
		}
	}
	if IsReadSurfaceMethod("PATCH") {
		t.Fatal("IsReadSurfaceMethod(PATCH) = true, want false")
	}
}
