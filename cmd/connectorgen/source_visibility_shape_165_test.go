package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"slices"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
)

func TestSourceVisibilityCarriedReferenceShape165(t *testing.T) {
	repo, _, _ := sourceFoundationCombinedRetainedProfile155(t, false, "sibling159")
	projections, err := buildSourceVisibility(t.Context(), repo)
	if err != nil {
		t.Fatal(err)
	}
	original := projections["acme"]
	v, err := connectors.DecodeSourceVisibility(t.Context(), original)
	if err != nil {
		t.Fatal(err)
	}
	if len(v.Operations[0].Cells[1].Admitted) != 2 {
		t.Fatal("real admitted operation and command control absent")
	}
	// Valid sibling representations remain data-valid, including explicit schema roots.
	for _, kind := range []string{"canonical_operation", "schema", "sync_transport", "operation", "write", "stream", "command"} {
		for _, target := range []string{"schema", "config", "parameter", "pagination_parameter"} {
			t.Run("valid/"+kind+"/"+target, func(t *testing.T) {
				current, err := connectors.DecodeSourceVisibility(t.Context(), original)
				if err != nil {
					t.Fatal(err)
				}
				ref := &current.Operations[0].Cells[1].Admitted[0]
				ref.Kind = kind
				pointer := ""
				ref.FieldMappings[0].TargetKind = target
				switch target {
				case "config":
					pointer = "/enabled"
				case "parameter":
					pointer = "/parameters/0"
				case "pagination_parameter":
					pointer = "/pagination_parameters/12"
					ref.FieldMappings[0].TargetKind = "parameter"
				}
				ref.FieldMappings[0].TargetPointer = &pointer
				if kind == "sync_transport" {
					ref.SourceSchema = nil
					ref.SchemaRole = ""
					ref.FieldMappings = nil
				}
				raw, err := json.Marshal(current)
				if err != nil {
					t.Fatal(err)
				}
				artifact := original
				artifact.Payload = sourceFixtureGzip174(t, raw)
				artifact.Bytes = len(raw)
				artifact.SHA256 = sourceBytesHash(raw)
				if _, err := connectors.DecodeSourceVisibility(t.Context(), artifact); err != nil {
					t.Fatal(err)
				}
			})
		}
	}

	for _, tc := range []struct {
		name   string
		mutate func(*connectors.SourceArtifactRef)
	}{
		{"schema_field_without_role", func(r *connectors.SourceArtifactRef) { r.SourceSchema = nil; r.SchemaRole = "" }},
		{"config_root", func(r *connectors.SourceArtifactRef) {
			r.FieldMappings[0].TargetKind = "config"
			p := ""
			r.FieldMappings[0].TargetPointer = &p
		}},
		{"parameter_leading_zero", func(r *connectors.SourceArtifactRef) {
			r.FieldMappings[0].TargetKind = "parameter"
			p := "/parameters/01"
			r.FieldMappings[0].TargetPointer = &p
		}},
		{"parameter_missing_index", func(r *connectors.SourceArtifactRef) {
			r.FieldMappings[0].TargetKind = "parameter"
			p := "/parameters/"
			r.FieldMappings[0].TargetPointer = &p
		}},
		{"parameter_negative_index", func(r *connectors.SourceArtifactRef) {
			r.FieldMappings[0].TargetKind = "parameter"
			p := "/pagination_parameters/-1"
			r.FieldMappings[0].TargetPointer = &p
		}},
		{"excessive_target_depth", func(r *connectors.SourceArtifactRef) {
			p := strings.Repeat("/a", 257)
			r.FieldMappings[0].TargetPointer = &p
		}},
		{"nil_field_target", func(r *connectors.SourceArtifactRef) { r.FieldMappings[0].TargetPointer = nil }},
		{"source_schema_without_role", func(r *connectors.SourceArtifactRef) { r.SchemaRole = "" }},
		{"transport_with_schema", func(r *connectors.SourceArtifactRef) { r.Kind = "sync_transport" }},
		{"duplicate_field_mapping", func(r *connectors.SourceArtifactRef) { r.FieldMappings = append(r.FieldMappings, r.FieldMappings[0]) }},
		{"invalid_parameter_coordinate", func(r *connectors.SourceArtifactRef) {
			r.FieldMappings[0].TargetKind = "parameter"
			x := "/not-parameters/not-an-index"
			r.FieldMappings[0].TargetPointer = &x
		}},
	} {
		for _, role := range []string{"admitted", "intended"} {
			t.Run(tc.name+"/"+role, func(t *testing.T) {
				v, e := connectors.DecodeSourceVisibility(t.Context(), original)
				if e != nil {
					t.Fatal(e)
				}
				ref := &v.Operations[0].Cells[1].Admitted[0]
				if ref.SourceSchema == nil || len(ref.FieldMappings) != 1 {
					t.Fatal("real complete schema lineage control absent")
				}
				if role == "intended" {
					cell := &v.Operations[0].Cells[1]
					cell.Intended = append(cell.Intended, *ref)
					ref = &cell.Intended[len(cell.Intended)-1]
					ref.Role = "intended"
					ref.FieldMappings = slices.Clone(ref.FieldMappings)
				}
				tc.mutate(ref)
				// The existing author's closed representation provides an independent oracle.
				convert := sourceLaneTargetRef{Kind: ref.Kind, Lane: string(ref.Lane), SchemaRole: sourceLaneSchemaRole(ref.SchemaRole)}
				if ref.SourceSchema != nil {
					c := v.Citations[*ref.SourceSchema]
					convert.SourceSchema = &sourceFactRef{DocumentID: c.DocumentID, Pointer: c.Pointer, ValueSHA256: c.ValueSHA256}
				}
				for _, m := range ref.FieldMappings {
					c := v.Citations[m.SourceCitation]
					convert.FieldMappings = append(convert.FieldMappings, sourceLaneFieldMapping{Source: sourceFactRef{DocumentID: c.DocumentID, Pointer: c.Pointer, ValueSHA256: c.ValueSHA256}, Target: sourceLaneFieldTarget{Kind: sourceLaneFieldTargetKind(m.TargetKind), Pointer: m.TargetPointer}})
				}
				if e := sourceLaneTargetRefShape(convert); e == nil {
					t.Fatal("mutation did not violate existing closed representation")
				} else {
					t.Logf("authoring shape rejects %s: %v", tc.name, e)
				}
				raw, e := json.Marshal(v)
				if e != nil {
					t.Fatal(e)
				}
				a := original
				a.Payload = sourceFixtureGzip174(t, raw)
				a.Bytes = len(raw)
				a.SHA256 = sourceBytesHash(raw)
				loads := 0
				registry, e := connectors.NewLazyRegistryWithEntries([]connectors.LazyRegistryEntry{{Metadata: connectors.Metadata{Name: "acme"}, SourceVisibility: a}}, func(context.Context, string) (connectors.Connector, error) {
					loads++
					return nil, errors.New("execution should remain uncalled")
				})
				if e != nil {
					t.Fatal(e)
				}
				selection := connectors.SourceCellSelection{Source: v.Operations[0].Source, Lane: "direct_write"}
				_, e = app.PreflightConnectorSource(t.Context(), registry, selection)
				t.Logf("actual App preflight: err=%v execution_loads=%d", e, loads)
				if loads != 0 {
					t.Fatal("unexpected execution")
				}
				var invalid *connectors.SourceVisibilityDataError
				if !errors.As(e, &invalid) {
					t.Fatalf("selected decoder accepted authoring-invalid reference shape %s: %v", tc.name, e)
				}
			})
		}
	}
}

func TestSourceVisibilityReferenceCoordinates165(t *testing.T) {
	repo, _, _ := sourceFoundationCombinedRetainedProfile155(t, false, "sibling159")
	projections, err := buildSourceVisibility(t.Context(), repo)
	if err != nil {
		t.Fatal(err)
	}
	original := projections["acme"]
	for _, role := range []string{"intended", "admitted"} {
		for _, axis := range []string{"source", "target"} {
			for _, shape := range []string{"duplicate", "ancestor", "root", "sibling", "different_owner"} {
				t.Run(role+"/"+axis+"/"+shape, func(t *testing.T) {
					v, err := connectors.DecodeSourceVisibility(t.Context(), original)
					if err != nil {
						t.Fatal(err)
					}
					cell := &v.Operations[0].Cells[1]
					ref := &cell.Admitted[0]
					if role == "intended" {
						cell.Intended = append(cell.Intended, *ref)
						ref = &cell.Intended[len(cell.Intended)-1]
						ref.Role = role
					}
					// Distinct citation entries deliberately may name the same coordinate: index inequality is insufficient.
					sourceA := v.Citations[ref.FieldMappings[0].SourceCitation]
					sourceB := sourceA
					sourceA.Pointer = "/a"
					sourceB.Pointer = "/b"
					targetA, targetB := "/a", "/b"
					switch shape {
					case "duplicate":
						if axis == "source" {
							sourceB.Pointer = sourceA.Pointer
						} else {
							targetB = targetA
						}
					case "ancestor":
						if axis == "source" {
							sourceB.Pointer = "/a/b"
						} else {
							targetB = "/a/b"
						}
					case "root":
						if axis == "source" {
							sourceA.Pointer = ""
						} else {
							targetA = ""
						}
					case "different_owner":
						if axis == "source" {
							for _, doc := range v.Documents {
								if doc.DocumentID != sourceA.DocumentID {
									sourceB.DocumentID = doc.DocumentID
									break
								}
							}
							if sourceB.DocumentID == sourceA.DocumentID {
								// Synthetic second document isolates coordinate ownership, not source truth.
								sibling := v.Documents[0]
								sibling.DocumentID += ":coordinate-control"
								v.Documents = append(v.Documents, sibling)
								sourceB.DocumentID = sibling.DocumentID
							}
							sourceB.Pointer = sourceA.Pointer
						} else {
							targetB = targetA
						}
					}
					first := len(v.Citations)
					v.Citations = append(v.Citations, sourceA, sourceB)
					ref.FieldMappings = []connectors.SourceFieldMapping{{SourceCitation: first, TargetKind: "schema", TargetPointer: &targetA}, {SourceCitation: first + 1, TargetKind: "schema", TargetPointer: &targetB}}
					if shape == "different_owner" && axis == "target" {
						ref.FieldMappings[1].TargetKind = "config"
					}
					raw, err := json.Marshal(v)
					if err != nil {
						t.Fatal(err)
					}
					a := original
					a.Payload = sourceFixtureGzip174(t, raw)
					a.Bytes = len(raw)
					a.SHA256 = sourceBytesHash(raw)
					_, err = connectors.DecodeSourceVisibility(t.Context(), a)
					valid := shape == "sibling" || shape == "different_owner"
					var invalid *connectors.SourceVisibilityDataError
					if valid && err != nil {
						t.Fatalf("disjoint coordinates rejected: %v", err)
					}
					if !valid && !errors.As(err, &invalid) {
						t.Fatalf("overlapping coordinates accepted: %v", err)
					}
				})
			}
		}
	}
}

func TestSourceVisibilityReferenceCitationShape165(t *testing.T) {
	repo, _, _ := sourceFoundationCombinedRetainedProfile155(t, false, "sibling159")
	projections, err := buildSourceVisibility(t.Context(), repo)
	if err != nil {
		t.Fatal(err)
	}
	original := projections["acme"]
	for _, location := range []string{"schema", "field"} {
		for _, shape := range []string{"section", "part", "depth"} {
			t.Run(location+"/"+shape, func(t *testing.T) {
				v, err := connectors.DecodeSourceVisibility(t.Context(), original)
				if err != nil {
					t.Fatal(err)
				}
				ref := &v.Operations[0].Cells[1].Admitted[0]
				index := *ref.SourceSchema
				if location == "field" {
					index = ref.FieldMappings[0].SourceCitation
				}
				citation := v.Citations[index]
				switch shape {
				case "section":
					citation.Section = "Heading"
				case "part":
					citation.Part = "body"
				case "depth":
					citation.Pointer = strings.Repeat("/a", 257)
				}
				if sourceLaneJSONFactRefShape(sourceFactRef{DocumentID: citation.DocumentID, Pointer: citation.Pointer, ValueSHA256: citation.ValueSHA256, Section: citation.Section, Part: citation.Part}) {
					t.Fatal("authoring oracle accepted malformed JSON citation")
				}
				newIndex := len(v.Citations)
				v.Citations = append(v.Citations, citation)
				if location == "field" {
					ref.FieldMappings[0].SourceCitation = newIndex
				} else {
					ref.SourceSchema = &newIndex
				}
				raw, err := json.Marshal(v)
				if err != nil {
					t.Fatal(err)
				}
				a := original
				a.Payload = sourceFixtureGzip174(t, raw)
				a.Bytes = len(raw)
				a.SHA256 = sourceBytesHash(raw)
				_, err = connectors.DecodeSourceVisibility(t.Context(), a)
				var invalid *connectors.SourceVisibilityDataError
				if !errors.As(err, &invalid) {
					t.Fatalf("JSON reference admitted non-JSON/deep citation: %v", err)
				}
			})
		}
	}
}

// The CLI fixture is generated from the same real canonical two-operation
// authoring fixture, so its admitted-reference control cannot silently drift.
func TestSourceVisibilityCLIFixture165(t *testing.T) {
	repo, _, _ := sourceFoundationCombinedRetainedProfile155(t, false, "sibling159")
	projections, err := buildSourceVisibility(t.Context(), repo)
	if err != nil {
		t.Fatal(err)
	}
	artifact := projections["acme"]
	artifact.Payload = string(sourceFixtureRaw174(t, artifact.Payload))
	artifact.Encoding = ""
	raw, err := json.Marshal(artifact)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := os.ReadFile("../../internal/cli/testdata/source-visibility-shape-165.json")
	if err != nil {
		t.Fatal(err)
	}
	var retained connectors.SourceVisibilityArtifact
	if err := json.Unmarshal(expected, &retained); err != nil {
		t.Fatal(err)
	}
	expected, err = json.Marshal(retained)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bytes.TrimSpace(expected), raw) {
		t.Fatal("CLI fixture differs from real canonical generation")
	}
}
