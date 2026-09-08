package cli

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
)

func TestSourceVisibilityShapePublicRoute165(t *testing.T) {
	raw, err := os.ReadFile("testdata/source-visibility-shape-165.json")
	if err != nil {
		t.Fatal(err)
	}
	var original connectors.SourceVisibilityArtifact
	if err := json.Unmarshal(raw, &original); err != nil {
		t.Fatal(err)
	}
	// This artifact is checked against actual canonical generation by connectorgen.
	for _, tc := range []struct {
		name   string
		mutate func(*connectors.SourceArtifactRef)
	}{
		{"valid", func(*connectors.SourceArtifactRef) {}},
		{"nil_field_target", func(r *connectors.SourceArtifactRef) { r.FieldMappings[0].TargetPointer = nil }},
		{"source_schema_without_role", func(r *connectors.SourceArtifactRef) { r.SchemaRole = "" }},
		{"transport_with_schema", func(r *connectors.SourceArtifactRef) { r.Kind = "sync_transport" }},
		{"duplicate_field_mapping", func(r *connectors.SourceArtifactRef) { r.FieldMappings = append(r.FieldMappings, r.FieldMappings[0]) }},
		{"invalid_parameter_coordinate", func(r *connectors.SourceArtifactRef) {
			r.FieldMappings[0].TargetKind = "parameter"
			p := "/wrong/index"
			r.FieldMappings[0].TargetPointer = &p
		}},
	} {
		for _, role := range []string{"intended", "admitted"} {
			t.Run(tc.name+"/"+role, func(t *testing.T) {
				v, err := connectors.DecodeSourceVisibility(t.Context(), original)
				if err != nil {
					t.Fatal(err)
				}
				cell := &v.Operations[0].Cells[1]
				ref := &cell.Admitted[0]
				if ref.SourceSchema == nil || len(ref.FieldMappings) != 1 {
					t.Fatal("canonical mapped schema control absent")
				}
				if role == "intended" {
					cell.Intended = append(cell.Intended, *ref)
					ref = &cell.Intended[len(cell.Intended)-1]
					ref.Role = role
					ref.FieldMappings = slices.Clone(ref.FieldMappings)
				}
				tc.mutate(ref)
				payload, err := json.Marshal(v)
				if err != nil {
					t.Fatal(err)
				}
				artifact := original
				artifact.Payload = string(payload)
				artifact.Bytes = len(payload)
				sum := sha256.Sum256(payload)
				artifact.SHA256 = hex.EncodeToString(sum[:])
				loads, opens := 0, 0
				registry, err := connectors.NewLazyRegistryWithEntries([]connectors.LazyRegistryEntry{{Metadata: connectors.Metadata{Name: "acme"}, SourceVisibility: artifact}}, func(context.Context, string) (connectors.Connector, error) {
					loads++
					return nil, errors.New("execution boundary reached")
				})
				if err != nil {
					t.Fatal(err)
				}
				open := func(string) (*app.App, error) { opens++; return nil, errors.New("app open boundary reached") }
				approval := &sourceApprovalReader162{}
				openers := appOpeners{open: open, reverse: open, registry: registry, approvalReader: approval, mode: appOpenerTestOverride}
				args := []string{"connectors", "inspect", "acme", "--inventory", "primary", "--source-id", v.Operations[0].Source.ID, "--lane", "direct_write", "--preflight", "--json", "--root", t.TempDir()}
				var out, diag bytes.Buffer
				exit := run(args, &out, &diag, openers)
				var envelope struct {
					Kind  string `json:"kind"`
					Error struct {
						Code     string `json:"code"`
						Category string `json:"category"`
					} `json:"error"`
				}
				if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
					t.Fatalf("malformed error envelope: %v: %s", err, &out)
				}
				expected := "source_visibility_invalid"
				category := "internal"
				if tc.name == "valid" {
					expected = "source_mapping_unproven"
					category = "internal"
				}
				if exit != 1 || envelope.Kind != "Error" || envelope.Error.Code != expected || envelope.Error.Category != category || loads != 0 || opens != 0 || approval.reads != 0 {
					t.Fatalf("wrong selected outcome: exit=%d result=%+v loads=%d opens=%d approvals=%d stderr=%s", exit, envelope, loads, opens, approval.reads, &diag)
				}
			})
		}
	}
}
