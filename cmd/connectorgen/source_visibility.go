package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors"
)

// buildSourceVisibility reads retained inputs only at generation time. The
// independently anchored cohort, rather than report counts, owns membership.
func buildSourceVisibility(ctx context.Context, repo string) (map[string]connectors.SourceVisibilityArtifact, error) {
	root, err := os.OpenRoot(repo)
	if err != nil {
		return nil, err
	}
	defer func() { _ = root.Close() }()
	cohortRaw, err := readSourceInput(root, sourceLaneCohortPath, 64<<20)
	if err != nil {
		return nil, err
	}
	var cohort sourceLaneCohort
	if err = decodeSourceJSON(cohortRaw, &cohort); err != nil {
		return nil, err
	}
	if err = decodeStrictJSON(cohortRaw, &cohort); err != nil {
		return nil, err
	}
	if err = validateSourceLaneCohort(cohort); err != nil {
		return nil, err
	}
	annotationRaw, err := readSourceInput(root, sourceLaneAnnotationsPath, 64<<20)
	if err != nil {
		return nil, err
	}
	var annotations struct {
		SchemaVersion int                        `json:"schema_version"`
		Annotations   []sourceSemanticAnnotation `json:"annotations"`
	}
	if err = decodeSourceJSON(annotationRaw, &annotations); err != nil {
		return nil, err
	}
	if err = decodeStrictJSON(annotationRaw, &annotations); err != nil {
		return nil, err
	}
	if annotations.SchemaVersion != 1 || annotations.Annotations == nil {
		return nil, fmt.Errorf("source annotation document invalid")
	}
	manifest, err := buildSourceLaneManifest(ctx, repo, cohort, annotations.Annotations)
	if err != nil {
		return nil, err
	}
	if manifest.Validation.Errors != 0 {
		return nil, fmt.Errorf("source projection consistency: %d errors", manifest.Validation.Errors)
	}
	cache := sourceProofFileCache{ctx: ctx, root: root, limits: sourceProofFileLimits{UniqueBytes: 64 << 20, Files: 1}, files: map[string]*sourceProofFile{}}
	atlas, err := readSourceFoundationAtlas(&cache)
	if err != nil {
		return nil, err
	}
	result, err := projectSourceVisibility(manifest, cohort, sourceBytesHash(cohortRaw), atlas)
	if err != nil {
		return nil, err
	}
	// Revalidate all observed physical authoring inputs before returning bytes.
	pins := append([]sourceArtifactPin(nil), manifest.Inputs...)
	pins = append(pins, sourceArtifactPin{Path: sourceLaneCohortPath, SHA256: sourceBytesHash(cohortRaw), Bytes: int64(len(cohortRaw))}, sourceArtifactPin{Path: sourceLaneAnnotationsPath, SHA256: sourceBytesHash(annotationRaw), Bytes: int64(len(annotationRaw))}, atlas.pin)
	for _, pin := range pins {
		if pin.Bytes <= 0 {
			continue
		}
		raw, e := readSourceInput(root, pin.Path, pin.Bytes)
		if e != nil || int64(len(raw)) != pin.Bytes || sourceBytesHash(raw) != pin.SHA256 {
			return nil, fmt.Errorf("source projection input changed: %s", pin.Path)
		}
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func projectSourceVisibility(manifest sourceLaneManifest, cohort sourceLaneCohort, cohortHash string, atlas sourceFoundationAtlas) (map[string]connectors.SourceVisibilityArtifact, error) {
	expected := map[sourceOperationKey]bool{}
	for _, inventory := range cohort.Inventories {
		for _, id := range inventory.ExpectedIDs {
			key := sourceOperationKey{Connector: inventory.Connector, Inventory: inventory.Inventory, ID: id}
			if expected[key] {
				return nil, fmt.Errorf("duplicate expected source key: %+v", key)
			}
			expected[key] = true
		}
	}

	seenKeys := map[sourceOperationKey]bool{}
	var identityIssues []string
	for _, row := range manifest.SourceOperations {
		key := row.Source.Key
		if seenKeys[key] {
			identityIssues = append(identityIssues, fmt.Sprintf("duplicate source key: %+v", key))
		}
		if !expected[key] {
			identityIssues = append(identityIssues, fmt.Sprintf("unexpected source key: %+v", key))
		}
		seenKeys[key] = true
	}
	for key := range expected {
		if !seenKeys[key] {
			identityIssues = append(identityIssues, fmt.Sprintf("missing source key: %+v", key))
		}
	}
	if len(identityIssues) > 0 {
		sort.Strings(identityIssues)
		return nil, fmt.Errorf("%s", strings.Join(identityIssues, "; "))
	}
	if findings := validateSourceLaneFactCitations(manifest); len(findings) > 0 {
		return nil, fmt.Errorf("source projection citation invalid: %s %s", findings[0].Code, findings[0].Pointer)
	}
	if err := validateSourceVisibilityLaneCitations(manifest); err != nil {
		return nil, err
	}
	observed := map[sourceOperationKey]bool{}
	catalogs := map[string]*connectors.SourceVisibility{}
	citationIndexes := map[string]map[sourceFactRef]int{}
	for _, row := range manifest.SourceOperations {
		key := row.Source.Key
		if !expected[key] || observed[key] {
			return nil, fmt.Errorf("unexpected or duplicate source key: %+v", key)
		}
		observed[key] = true
		if !row.Source.Observed {
			return nil, fmt.Errorf("unobserved source key: %+v", key)
		}
		v := catalogs[key.Connector]
		if v == nil {
			v = &connectors.SourceVisibility{SchemaVersion: 1, Connector: key.Connector, Coverage: "in_cohort", CohortID: cohort.CohortID, CohortSHA256: cohortHash, AtlasSHA256: atlas.pin.SHA256, Documents: []connectors.SourceDocumentRef{}, Citations: []connectors.SourceFactCitation{}, Operations: []connectors.SourceOperationObservation{}}
			catalogs[key.Connector] = v
			citationIndexes[key.Connector] = map[sourceFactRef]int{}
		}
		cite := func(ref sourceFactRef) int {
			if i, ok := citationIndexes[key.Connector][ref]; ok {
				return i
			}
			i := len(v.Citations)
			v.Citations = append(v.Citations, connectors.SourceFactCitation{DocumentID: ref.DocumentID, Pointer: ref.Pointer, ValueSHA256: ref.ValueSHA256, Section: ref.Section, Part: ref.Part})
			citationIndexes[key.Connector][ref] = i
			return i
		}
		identity, ok := row.Facts.Refs["source_operation"]
		if !ok {
			identity, ok = row.Facts.Refs["rendered_reference"]
		}
		if !ok {
			return nil, fmt.Errorf("source identity citation absent: %+v", key)
		}
		o := connectors.SourceOperationObservation{Source: connectors.SourceOperationKey{Connector: key.Connector, Inventory: key.Inventory, ID: key.ID}, Class: row.Source.Class, DocumentID: row.Source.DocumentID, Pointer: row.Source.Pointer, SourceLocation: row.Source.SourceLocation, Observed: row.Source.Observed, Protocol: row.Facts.Protocol, Method: row.Facts.Method, Path: row.Facts.Path, IdentityCitation: cite(identity), DisplayCitations: []int{}, Cells: []connectors.SourceCellObservation{}}
		for _, name := range []string{"method", "path", "protocol", "operation_id"} {
			if ref, ok := row.Facts.Refs[name]; ok {
				o.DisplayCitations = append(o.DisplayCitations, cite(ref))
			}
		}
		for _, cell := range row.Lanes {
			c := connectors.SourceCellObservation{Lane: connectors.SourceLane(cell.Lane), Applicability: cell.Applicability, SourceState: cell.State, RuleID: cell.RuleID, SourceReason: connectors.SourceReason{Code: cell.Reason.Code, Text: cell.Reason.Text}, LaneFacts: []int{}, Intended: []connectors.SourceArtifactRef{}, Admitted: []connectors.SourceArtifactRef{}, OwnerRefs: append([]string{}, cell.OwnerRefs...), GapRefs: append([]string{}, cell.GapRefs...)}
			c.Capability = connectors.SourceCapabilityObservation{Kind: "unresolved_mapping", Name: cell.Lane + " source-to-capability mapping (unresolved)", CandidateIDs: []string{}, Provenance: "retained_source_lane_observation"}
			for _, ref := range cell.FactRefs {
				c.LaneFacts = append(c.LaneFacts, cite(ref))
			}
			if cell.State == "not_applicable" {
				c.Capability.Kind = "not_required"
				c.Capability.Name = "source contract excludes " + cell.Lane
			}
			if cell.State == "missing_foundation" {
				found := false
				for _, demand := range sourceFoundationDemandCatalog() {
					if demand.Key == key && cell.Lane == "sync_transport" && len(cell.GapRefs) == 1 && cell.GapRefs[0] == demand.Gap {
						entry, ok := atlas.entries[demand.AtlasID]
						if !ok || entry.Owner.PrimaryPackage != demand.AtlasOwner {
							return nil, fmt.Errorf("source gap owner absent")
						}
						id := demand.AtlasID
						c.Capability = connectors.SourceCapabilityObservation{Kind: "known_gap", Name: "webhook event/receiver demand under the sync transport contract", AtlasID: &id, CandidateIDs: []string{}, GapID: demand.Gap, Provenance: "retained_source_demand_under_available_owner"}
						for _, ref := range demand.Citations {
							c.LaneFacts = append(c.LaneFacts, cite(ref))
						}
						found = true
					}
				}
				if !found {
					return nil, fmt.Errorf("unknown source gap")
				}
			}
			for _, group := range []struct {
				role string
				refs []sourceLaneTargetRef
			}{{"intended", cell.IntendedBindings}, {"admitted", cell.References}} {
				for _, ref := range group.refs {
					projected := connectors.SourceArtifactRef{Role: group.role, Kind: ref.Kind, Connector: ref.Connector, ID: ref.ID, Lane: connectors.SourceLane(ref.Lane), Artifact: ref.Artifact, Pointer: ref.Pointer, ArtifactSHA256: ref.ArtifactSHA256, CanonicalID: ref.CanonicalID, CanonicalPointer: ref.CanonicalPointer, Generation: ref.Generation}
					projected.SchemaRole = string(ref.SchemaRole)
					if ref.SourceSchema != nil {
						index := cite(*ref.SourceSchema)
						projected.SourceSchema = &index
					}
					for _, mapping := range ref.FieldMappings {
						projected.FieldMappings = append(projected.FieldMappings, connectors.SourceFieldMapping{SourceCitation: cite(mapping.Source), TargetKind: string(mapping.Target.Kind), TargetPointer: mapping.Target.Pointer})
					}
					if group.role == "intended" {
						c.Intended = append(c.Intended, projected)
					} else {
						c.Admitted = append(c.Admitted, projected)
					}
				}
			}
			o.Cells = append(o.Cells, c)
		}
		v.Operations = append(v.Operations, o)
	}
	for key := range expected {
		if !observed[key] {
			return nil, fmt.Errorf("missing source key: %+v", key)
		}
	}
	result := map[string]connectors.SourceVisibilityArtifact{}
	total := 0
	for name, v := range catalogs {
		documents := map[string]bool{}
		for _, c := range v.Citations {
			documents[c.DocumentID] = true
		}
		for _, o := range v.Operations {
			documents[o.DocumentID] = true
		}
		for _, d := range manifest.Documents {
			if documents[d.ID] {
				v.Documents = append(v.Documents, connectors.SourceDocumentRef{DocumentID: d.ID, Path: d.Path, RetainedFileSHA256: d.RetainedFileSHA256, Bytes: d.Bytes, UpstreamDeclaredSHA256: d.UpstreamDeclaredSHA256, UpstreamDeclaredBytes: d.UpstreamDeclaredBytes, UpstreamBytesVerified: d.UpstreamBytesVerified})
				delete(documents, d.ID)
			}
		}
		if len(documents) != 0 {
			return nil, fmt.Errorf("source projection document absent")
		}
		sort.Slice(v.Documents, func(i, j int) bool { return v.Documents[i].DocumentID < v.Documents[j].DocumentID })
		sort.Slice(v.Operations, func(i, j int) bool {
			a, b := v.Operations[i].Source, v.Operations[j].Source
			if a.Inventory != b.Inventory {
				return a.Inventory < b.Inventory
			}
			return a.ID < b.ID
		})
		keys := make([]connectors.SourceOperationKey, 0, len(v.Operations))
		for _, o := range v.Operations {
			keys = append(keys, o.Source)
		}
		sortSourceVisibilityCitations(v)
		v.KeySHA256 = connectors.SourceKeyDigest(keys)
		v.OperationCount = len(keys)
		v.CellCount = 7 * len(keys)
		if err := connectors.ValidateSourceVisibility(*v); err != nil {
			return nil, fmt.Errorf("source projection %s: %w", name, err)
		}
		raw, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		total += len(raw)
		if len(raw) > connectors.SourceVisibilityConnectorLimit || total > connectors.SourceVisibilityTotalLimit {
			return nil, fmt.Errorf("complete source projection exceeds payload budget: connector=%s bytes=%d total=%d", name, len(raw), total)
		}
		result[name] = connectors.SourceVisibilityArtifact{SchemaVersion: 1, Connector: name, Coverage: "in_cohort", CohortID: cohort.CohortID, Bytes: len(raw), SHA256: sourceBytesHash(raw), KeySHA256: v.KeySHA256, Payload: string(raw)}
	}
	return result, nil
}

// Lane claims must resolve to retained bytes and the selected operation's fact
// closure. A valid hash from a sibling operation is not a source fit witness.
func validateSourceVisibilityLaneCitations(manifest sourceLaneManifest) error {
	documents := map[string]retainedSourceDocument{}
	roots := map[string]any{}
	for _, doc := range manifest.Documents {
		documents[doc.ID] = doc
		if doc.ContentType != "text/html" {
			view, err := sourceDocumentViewFor(doc)
			if err != nil {
				return err
			}
			roots[doc.ID] = view.ReferenceRoot
		}
	}
	checked := map[sourceFactRef]bool{}
	for _, row := range manifest.SourceOperations {
		for _, cell := range row.Lanes {
			for _, ref := range cell.FactRefs {
				belongs := false
				for _, base := range row.Facts.Refs {
					if ref == base || (ref.DocumentID == base.DocumentID && ref.Section == "" && ref.Part == "" && base.Section == "" && base.Part == "" && base.Pointer != "" && strings.HasPrefix(ref.Pointer, base.Pointer+"/")) {
						belongs = true
						break
					}
				}
				if !belongs {
					return fmt.Errorf("source lane citation outside selected fact closure: %+v %s", row.Source.Key, ref.Pointer)
				}
				if checked[ref] {
					continue
				}
				doc, ok := documents[ref.DocumentID]
				if !ok {
					return fmt.Errorf("source lane citation document absent")
				}
				var raw json.RawMessage
				var err error
				if doc.ContentType == "text/html" {
					raw, err = resolveSourceFactValue(doc, ref)
				} else {
					if ref.Section != "" || ref.Part != "" {
						return fmt.Errorf("rendered selector on JSON source")
					}
					raw, err = sourceLaneRetainedPointer(roots[ref.DocumentID], ref.Pointer)
				}
				if err != nil {
					return err
				}
				canonical, err := canonicalSourceJSON(raw)
				if err != nil || sourceBytesHash(canonical) != ref.ValueSHA256 {
					return fmt.Errorf("source lane citation hash mismatch: %+v %s", row.Source.Key, ref.Pointer)
				}
				checked[ref] = true
			}
		}
	}
	return nil
}

func sortSourceVisibilityCitations(v *connectors.SourceVisibility) {
	order := make([]int, len(v.Citations))
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(i, j int) bool {
		a, b := v.Citations[order[i]], v.Citations[order[j]]
		if a.DocumentID != b.DocumentID {
			return a.DocumentID < b.DocumentID
		}
		if a.Pointer != b.Pointer {
			return a.Pointer < b.Pointer
		}
		if a.Section != b.Section {
			return a.Section < b.Section
		}
		if a.Part != b.Part {
			return a.Part < b.Part
		}
		return a.ValueSHA256 < b.ValueSHA256
	})
	refs := make([]connectors.SourceFactCitation, len(order))
	remap := make([]int, len(order))
	for next, old := range order {
		refs[next] = v.Citations[old]
		remap[old] = next
	}
	v.Citations = refs
	for i := range v.Operations {
		o := &v.Operations[i]
		o.IdentityCitation = remap[o.IdentityCitation]
		for j, old := range o.DisplayCitations {
			o.DisplayCitations[j] = remap[old]
		}
		for j := range o.Cells {
			for _, refs := range [][]connectors.SourceArtifactRef{o.Cells[j].Intended, o.Cells[j].Admitted} {
				for k := range refs {
					if refs[k].SourceSchema != nil {
						next := remap[*refs[k].SourceSchema]
						refs[k].SourceSchema = &next
					}
					for n := range refs[k].FieldMappings {
						refs[k].FieldMappings[n].SourceCitation = remap[refs[k].FieldMappings[n].SourceCitation]
					}
				}
			}
			for k, old := range o.Cells[j].LaneFacts {
				o.Cells[j].LaneFacts[k] = remap[old]
			}
		}
	}
}
