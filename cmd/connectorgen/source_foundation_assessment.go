package main

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
)

const sourceFoundationAssessmentsPath = "data/connector-canon/batch1-foundation-assessments.json"

// Authored requirements describe work to assess. They are not proof records,
// execution targets or permission to alter the independent source manifest.
type sourceFoundationRequirement struct {
	ID                   string                     `json:"id"`
	SourceRefs           []sourceFactRef            `json:"source_refs"`
	Statement            string                     `json:"statement"`
	AtlasLookup          sourceFoundationLookup     `json:"atlas_lookup"`
	Assessment           string                     `json:"assessment"`
	ProofIDs             []string                   `json:"proof_ids"`
	AffectedArtifacts    []string                   `json:"affected_artifacts"`
	EvidenceRequirements []string                   `json:"evidence_requirements"`
	DecisionRefs         []sourceFoundationDecision `json:"decision_refs"`
}

type sourceFoundationLookup struct {
	Atlas      sourceArtifactPin                 `json:"atlas"`
	Candidates []sourceFoundationLookupCandidate `json:"candidates"`
}

type sourceFoundationLookupCandidate struct {
	AtlasID     string                      `json:"atlas_id"`
	Contract    sourceFoundationContractRef `json:"contract"`
	Disposition string                      `json:"disposition"`
	Rationale   string                      `json:"rationale"`
}

type sourceFoundationDecision struct {
	ID        string `json:"id"`
	State     string `json:"state"`
	Condition string `json:"condition"`
}

type sourceFoundationCellAssessment struct {
	Key          sourceOperationKey            `json:"key"`
	Lane         string                        `json:"lane"`
	Requirements []sourceFoundationRequirement `json:"requirements"`
	NextOwner    string                        `json:"next_owner"`
}

type sourceFoundationAssessmentDocument struct {
	SchemaVersion int                              `json:"schema_version"`
	Kind          string                           `json:"kind"`
	Atlas         sourceArtifactPin                `json:"atlas"`
	Assessments   []sourceFoundationCellAssessment `json:"assessments"`
}

// An observation retains what the author supplied, separately from the source
// and Atlas actually read. Only the subsequent requirement consumer determines
// whether an authored claim is supported; no resolved result is exposed here.
type sourceFoundationAssessmentObservations struct {
	universe sourceFoundationUniverse
	atlas    sourceFoundationAtlas
	pin      sourceArtifactPin
	authored []sourceFoundationCellAssessment
}

func observeSourceFoundationAssessments(ctx context.Context, repo, name string, universe sourceFoundationUniverse) (sourceFoundationAssessmentObservations, error) {
	var result sourceFoundationAssessmentObservations
	if err := ctx.Err(); err != nil {
		return result, err
	}
	root, err := os.OpenRoot(repo)
	if err != nil {
		return result, fmt.Errorf("foundation assessment root: %w", err)
	}
	defer func() { _ = root.Close() }() // Read-only confinement handle.
	cache := sourceProofFileCache{ctx: ctx, root: root,
		limits: sourceProofFileLimits{UniqueBytes: 512 << 20, Files: 65536},
		files:  map[string]*sourceProofFile{}}
	atlas, err := readSourceFoundationAtlas(&cache)
	if err != nil {
		return result, err
	}
	raw, file := cache.get(name, 64<<20, false)
	if file.code != "" {
		return result, fmt.Errorf("foundation assessments: %s", file.code)
	}
	var document sourceFoundationAssessmentDocument
	if decodeSourceJSON(raw, &document) != nil || decodeStrictJSON(raw, &document) != nil ||
		document.SchemaVersion != 1 || document.Kind != "foundation_cell_assessments" ||
		document.Assessments == nil || document.Atlas != atlas.pin || atlas.pin != universe.atlasPin {
		return result, fmt.Errorf("foundation assessment document invalid")
	}
	if err := validateSourceFoundationAssessmentObservations(ctx, document.Assessments, universe, atlas); err != nil {
		return result, err
	}
	cache.finalize()
	if err := ctx.Err(); err != nil {
		return result, err
	}
	for _, path := range cache.order {
		if cache.files[path].code != "" {
			return result, fmt.Errorf("foundation assessment input changed")
		}
	}
	return sourceFoundationAssessmentObservations{universe: universe, atlas: atlas,
		pin: sourceArtifactPin{Path: name, SHA256: file.hash, Bytes: file.size}, authored: document.Assessments}, nil
}

func validateSourceFoundationAssessmentObservations(ctx context.Context, assessments []sourceFoundationCellAssessment, universe sourceFoundationUniverse, atlas sourceFoundationAtlas) error {
	cells := make(map[sourceFoundationCell]sourceLaneManifestRow, len(universe.cells))
	for _, row := range universe.manifest.SourceOperations {
		for _, lane := range row.Lanes {
			cells[sourceFoundationCell{Key: row.Source.Key, Lane: lane.Lane}] = row
		}
	}
	if len(assessments) > len(cells) {
		return fmt.Errorf("foundation assessment cell capacity exceeded")
	}
	documents := map[string]retainedSourceDocument{}
	for _, document := range universe.manifest.Documents {
		documents[document.ID] = document
	}
	seen := map[sourceFoundationCell]bool{}
	for _, assessment := range assessments {
		if err := ctx.Err(); err != nil {
			return err
		}
		cell := sourceFoundationCell{Key: assessment.Key, Lane: assessment.Lane}
		row, exists := cells[cell]
		if !exists || seen[cell] || len(assessment.Requirements) == 0 || len(assessment.Requirements) > 4096 {
			return fmt.Errorf("foundation assessment unknown or duplicate cell or invalid requirements")
		}
		seen[cell] = true
		if !sourceFoundationKnownOwner(assessment.NextOwner, atlas) {
			return fmt.Errorf("foundation assessment next owner unknown")
		}
		ids := map[string]bool{}
		for _, requirement := range assessment.Requirements {
			if !validSourceID(requirement.ID) || ids[requirement.ID] || strings.TrimSpace(requirement.Statement) == "" ||
				len(requirement.SourceRefs) == 0 || !sourceFoundationAssessmentStatus(requirement.Assessment) ||
				!sourceFoundationUniqueText(requirement.EvidenceRequirements, true) ||
				!sourceFoundationUniqueText(requirement.ProofIDs, false) || !sourceFoundationUniqueText(requirement.AffectedArtifacts, false) {
				return fmt.Errorf("foundation requirement identity or evidence invalid")
			}
			ids[requirement.ID] = true
			refs := map[sourceFactRef]bool{}
			for _, ref := range requirement.SourceRefs {
				if refs[ref] || !sourceFoundationRequirementCitation(row.Facts, documents[ref.DocumentID], ref) {
					return fmt.Errorf("foundation requirement source citation invalid or outside operation")
				}
				refs[ref] = true
			}
			if err := validateSourceFoundationLookup(requirement.AtlasLookup, atlas); err != nil {
				return err
			}
			for _, artifact := range requirement.AffectedArtifacts {
				if !sourceProofSafePath(artifact) {
					return fmt.Errorf("foundation affected artifact path invalid")
				}
			}
		}
	}
	// Known source requirements are independent of the editable assessment
	// array. Preserve registration/update facts even when a candidate omits
	// both its assessment and a copied owner label.
	for _, row := range universe.manifest.SourceOperations {
		_, registration := sourceRegistrationDemand(row.Facts)
		for _, lane := range row.Lanes {
			required := slices.Contains(lane.OwnerRefs, "CP13") || len(lane.GapRefs) != 0 ||
				registration && lane.Lane == "sync_transport"
			if required && !seen[sourceFoundationCell{Key: row.Source.Key, Lane: lane.Lane}] {
				return fmt.Errorf("foundation known source requirement missing: %s/%s", row.Source.Key.ID, lane.Lane)
			}
		}
	}
	return nil
}

func sourceFoundationKnownOwner(owner string, atlas sourceFoundationAtlas) bool {
	if owner == "cli-batch1-vercel-inbound-sync-decision-r1" || owner == "cli-plan-webhook-receiver-foundation-r1-decision-webhook-exposure-product-boundary" {
		return true
	}
	for _, entry := range atlas.entries {
		if owner == entry.Owner.PrimaryPackage {
			return true
		}
	}
	return false
}

func sourceFoundationAssessmentStatus(status string) bool {
	switch status {
	case "existing_shared_capability", "connector_local_configuration", "provider_limitation", "absent_shared_foundation", "unresolved", "no_demand_for_this_cell":
		return true
	default:
		return false
	}
}

func sourceFoundationUniqueText(values []string, required bool) bool {
	if values == nil || required && len(values) == 0 {
		return false
	}
	seen := map[string]bool{}
	for _, value := range values {
		if strings.TrimSpace(value) == "" || seen[value] {
			return false
		}
		seen[value] = true
	}
	return true
}

func sourceFoundationRequirementCitation(facts sourceFacts, document retainedSourceDocument, ref sourceFactRef) bool {
	// A value elsewhere in the same provider document is not evidence for this
	// operation. Admit its normalized references or descendants of its exact
	// operation occurrence; global normalized auth/paging refs remain available.
	owned := false
	for _, current := range facts.Refs {
		owned = owned || current == ref
	}
	operation, exists := facts.Refs["source_operation"]
	if exists && ref.Section == "" && operation.Section == "" && operation.DocumentID == ref.DocumentID &&
		(ref.Pointer == operation.Pointer || strings.HasPrefix(ref.Pointer, operation.Pointer+"/")) {
		owned = true
	}
	if !owned || !sourceProofDigest(ref.ValueSHA256) {
		return false
	}
	value, err := resolveSourceFactValue(document, ref)
	if err != nil {
		return false
	}
	canonical, err := canonicalSourceJSON(value)
	return err == nil && sourceBytesHash(canonical) == ref.ValueSHA256
}

func validateSourceFoundationLookup(lookup sourceFoundationLookup, atlas sourceFoundationAtlas) error {
	if lookup.Atlas != atlas.pin || len(lookup.Candidates) == 0 {
		return fmt.Errorf("foundation lookup missing or stale")
	}
	type candidateKey struct{ id, pointer string }
	seen := map[candidateKey]bool{}
	for _, candidate := range lookup.Candidates {
		entry, exists := atlas.entries[candidate.AtlasID]
		key := candidateKey{candidate.AtlasID, candidate.Contract.Pointer}
		if !exists || seen[key] || !strings.HasPrefix(candidate.Contract.Pointer, "/supported_contracts/") ||
			strings.TrimSpace(candidate.Rationale) == "" || !sourceFoundationAssessmentStatus(candidate.Disposition) {
			return fmt.Errorf("foundation lookup candidate invalid or duplicate")
		}
		seen[key] = true
		value, err := sourceJSONPointer(entry.raw, candidate.Contract.Pointer)
		if err != nil {
			return fmt.Errorf("foundation lookup contract occurrence unavailable")
		}
		canonical, err := canonicalSourceJSON(value)
		if err != nil || sourceBytesHash(canonical) != candidate.Contract.ValueSHA256 {
			return fmt.Errorf("foundation lookup contract occurrence changed")
		}
	}
	return nil
}
