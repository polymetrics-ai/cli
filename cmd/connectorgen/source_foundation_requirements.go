package main

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
)

// Evidence availability, authored assessment and resolved requirement status
// remain separate. A current assertion never becomes a lane proof reference.
type sourceFoundationRequirementResult struct {
	MechanismFits          []sourceFoundationMechanismFit          `json:"mechanism_fits"`
	Identity               sourceFoundationCell                    `json:"identity"`
	ID                     string                                  `json:"id"`
	Statement              string                                  `json:"statement"`
	AuthoredAssessment     string                                  `json:"authored_assessment"`
	Status                 string                                  `json:"status"`
	SourceRefs             []sourceFactRef                         `json:"source_refs"`
	Proofs                 []sourceFoundationAssertionResult       `json:"proofs"`
	RequestedProofIDs      []string                                `json:"requested_proof_ids"`
	ProofIssues            []sourceFoundationRequirementProofIssue `json:"proof_issues"`
	NextOwner              string                                  `json:"next_owner"`
	MissingEvidence        []string                                `json:"missing_evidence"`
	DecisionRefs           []sourceFoundationDecision              `json:"decision_refs"`
	SourceState            string                                  `json:"source_state"`
	SourceApplicability    string                                  `json:"source_applicability"`
	GapRefs                []string                                `json:"gap_refs"`
	RetainedDecisionOwners []string                                `json:"retained_decision_owners"`
	AtlasLookup            sourceFoundationLookup                  `json:"atlas_lookup"`
	AffectedArtifacts      []string                                `json:"affected_artifacts"`
	FitBindings            []sourceLaneTargetRef                   `json:"fit_bindings"`
	ProviderClause         *sourceFactRef                          `json:"provider_clause,omitempty"`
	SourceExclusion        *sourceFactRef                          `json:"source_exclusion,omitempty"`
}

type sourceFoundationRequirementProofIssue struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Path string `json:"path"`
}

type sourceFoundationAssertionResult struct {
	mechanism   string
	ID          string                      `json:"id"`
	AtlasID     string                      `json:"atlas_id"`
	Contract    sourceFoundationContractRef `json:"contract"`
	Assertion   string                      `json:"assertion"`
	Status      string                      `json:"status"`
	Limitations []string                    `json:"limitations"`
}

func buildSourceFoundationRequirements(ctx context.Context, repo string, observed sourceFoundationAssessmentObservations) (result []sourceFoundationRequirementResult, err error) {
	err = sourceFoundationAdmissionWithRoot(ctx, repo, observed.universe.admission, func() error {
		var buildErr error
		result, buildErr = buildSourceFoundationRequirementsCurrent(ctx, repo, observed)
		return buildErr
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func buildSourceFoundationRequirementsCurrent(ctx context.Context, repo string, observed sourceFoundationAssessmentObservations) ([]sourceFoundationRequirementResult, error) {
	batch, err := readSourceFoundationProofBatch(ctx, repo, reviewedSourceFoundationProofs(), nil)
	if err != nil {
		return nil, fmt.Errorf("foundation requirement proof observations: %w", err)
	}
	return buildSourceFoundationRequirementsBatch(ctx, observed, batch)
}

func buildSourceFoundationRequirementsBatch(ctx context.Context, observed sourceFoundationAssessmentObservations, batch sourceFoundationProofBatch) ([]sourceFoundationRequirementResult, error) {
	proofs := batch.records

	var err error
	byID := map[string]sourceFoundationProofObservation{}
	for _, proof := range proofs {
		byID[proof.record.ID] = proof
	}
	sourceRows := map[sourceOperationKey]sourceLaneManifestRow{}
	for _, row := range observed.universe.manifest.SourceOperations {
		sourceRows[row.Source.Key] = row
	}
	result := []sourceFoundationRequirementResult{}
	for _, cell := range observed.authored {
		for _, requirement := range cell.Requirements {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			row := sourceFoundationRequirementResult{
				Identity: sourceFoundationCell{Key: cell.Key, Lane: cell.Lane}, ID: requirement.ID,
				Statement: requirement.Statement, AuthoredAssessment: requirement.Assessment,
				SourceRefs: append([]sourceFactRef{}, requirement.SourceRefs...), Proofs: []sourceFoundationAssertionResult{},
				RequestedProofIDs: append([]string{}, requirement.ProofIDs...), ProofIssues: []sourceFoundationRequirementProofIssue{},
				MechanismFits: []sourceFoundationMechanismFit{},
				NextOwner:     cell.NextOwner, MissingEvidence: append([]string{}, requirement.EvidenceRequirements...),
				DecisionRefs: append([]sourceFoundationDecision{}, requirement.DecisionRefs...), GapRefs: []string{}, RetainedDecisionOwners: []string{},
				AtlasLookup: requirement.AtlasLookup, AffectedArtifacts: append([]string{}, requirement.AffectedArtifacts...),
				FitBindings: append([]sourceLaneTargetRef{}, requirement.FitBindings...), ProviderClause: requirement.ProviderClause, SourceExclusion: requirement.SourceExclusion,
			}
			if err := validateSourceFoundationDecisions(requirement.DecisionRefs); err != nil {
				return nil, err
			}
			for _, sourceCell := range sourceRows[cell.Key].Lanes {
				if sourceCell.Lane != cell.Lane {
					continue
				}
				row.SourceState, row.SourceApplicability = sourceCell.State, sourceCell.Applicability
				row.GapRefs = append(row.GapRefs, sourceCell.GapRefs...)
				for _, owner := range sourceCell.OwnerRefs {
					if sourceFoundationDecisionOwner(owner) {
						row.RetainedDecisionOwners = append(row.RetainedDecisionOwners, owner)
					}
				}
			}
			for _, id := range requirement.ProofIDs {
				proof, exists := byID[id]
				if !exists {
					code := "proof_record_missing"
					if batch.document.Status == "proof_unavailable" {
						code = "proof_document_missing"
					}
					row.ProofIssues = append(row.ProofIssues, sourceFoundationRequirementProofIssue{ID: id, Code: code, Path: sourceFoundationProofPath})
					row.MissingEvidence = append(row.MissingEvidence, code+": "+id+" ("+sourceFoundationProofPath+")")
					continue
				}
				if !slices.ContainsFunc(requirement.AtlasLookup.Candidates, func(candidate sourceFoundationLookupCandidate) bool {
					return candidate.AtlasID == proof.record.AtlasID && candidate.Contract == proof.record.Contract
				}) {
					return nil, fmt.Errorf("foundation requirement proof outside examined contract")
				}
				for _, issue := range proof.issues {
					row.ProofIssues = append(row.ProofIssues, sourceFoundationRequirementProofIssue{ID: id, Code: issue.Code, Path: issue.Path})
					row.MissingEvidence = append(row.MissingEvidence, issue.Code+": "+id+" ("+issue.Path+")")
				}
				row.Proofs = append(row.Proofs, sourceFoundationAssertionResult{ID: id, AtlasID: proof.record.AtlasID,
					Contract: proof.record.Contract, Assertion: proof.record.Assertion.Statement, Status: proof.status, mechanism: proof.mechanism,
					Limitations: append([]string{}, proof.record.Limitations...),
				})
			}
			row.Status, err = resolveSourceFoundationRequirement(requirement, cell, sourceRows[cell.Key], row.Proofs, observed)
			if err != nil {
				return nil, fmt.Errorf("foundation requirement %s: %w", requirement.ID, err)
			}
			if row.Status == "existing_shared_capability" || row.Status == "connector_local_configuration" {
				row.MechanismFits, err = sourceFoundationMechanismFits(requirement, cell, sourceRows[cell.Key], row.Proofs, observed)
				if err != nil {
					return nil, fmt.Errorf("foundation requirement %s: %w", requirement.ID, err)
				}
			}
			result = append(result, row)
		}
	}
	// A canonical-intended shared fit retains the separately validated local
	// command deficit in the same source cell; it cannot erase that local work.
	for _, row := range result {
		if row.Status != "existing_shared_capability" {
			continue
		}
		for _, fit := range row.MechanismFits {
			if fit.DeclarationState != "canonical_intended" {
				continue
			}
			companion := false
			for _, local := range result {
				if local.Identity != row.Identity || local.Status != "connector_local_configuration" {
					continue
				}
				for _, localFit := range local.MechanismFits {
					companion = companion || localFit.DeclarationState == "canonical_intended" && localFit.Binding.Kind == "command" && localFit.Binding.CanonicalID == fit.Binding.CanonicalID && localFit.Binding.Generation == fit.Binding.Generation && sourceFoundationAdmissionFit(observed.universe.admission, local.Identity, localFit.Binding)
				}
			}
			if !companion {
				return nil, fmt.Errorf("canonical intended shared fit lacks its validated local CLI deficit")
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Identity != result[j].Identity {
			return sourceFoundationCellLess(result[i].Identity, result[j].Identity)
		}
		return result[i].ID < result[j].ID
	})
	return result, nil
}

func sourceFoundationDecisionOwner(id string) bool {
	return id == "cli-batch1-vercel-inbound-sync-decision-r1" ||
		id == "cli-plan-webhook-receiver-foundation-r1-decision-webhook-exposure-product-boundary"
}

func validateSourceFoundationDecisions(decisions []sourceFoundationDecision) error {
	seen := map[string]bool{}
	for _, decision := range decisions {
		if !sourceFoundationDecisionOwner(decision.ID) || seen[decision.ID] {
			return fmt.Errorf("foundation decision owner unknown or duplicate")
		}
		seen[decision.ID] = true
		switch decision.ID {
		case "cli-batch1-vercel-inbound-sync-decision-r1":
			if decision.State != "pending" || decision.Condition != "" {
				return fmt.Errorf("foundation receiver decision cannot grant approval")
			}
		case "cli-plan-webhook-receiver-foundation-r1-decision-webhook-exposure-product-boundary":
			if decision.State != "conditional" || strings.TrimSpace(decision.Condition) == "" {
				return fmt.Errorf("foundation exposure decision must retain its condition")
			}
		}
	}
	return nil
}

func resolveSourceFoundationRequirement(requirement sourceFoundationRequirement, assessment sourceFoundationCellAssessment, source sourceLaneManifestRow, proofs []sourceFoundationAssertionResult, observed sourceFoundationAssessmentObservations) (string, error) {
	var lane sourceLaneCell
	for _, current := range source.Lanes {
		if current.Lane == assessment.Lane {
			lane = current
			break
		}
	}
	if lane.Lane == "" {
		return "", fmt.Errorf("source cell unavailable")
	}
	switch requirement.Assessment {
	case "unresolved":
		return "unresolved", nil
	case "existing_shared_capability", "connector_local_configuration":
		if len(proofs) == 0 || len(proofs) != len(requirement.ProofIDs) || len(requirement.FitBindings) == 0 {
			return "", fmt.Errorf("shared assertion or exact source configuration fit missing")
		}
		for _, proof := range proofs {
			// Each requirement is narrow. A broader need must be split into
			// separately evidenced requirements, not inferred from a test name.
			if proof.Status != "current" || proof.Assertion != requirement.Statement {
				return "", fmt.Errorf("reviewed assertion does not cover the stated requirement")
			}
		}
		available := append(append([]sourceLaneTargetRef{}, lane.References...), lane.IntendedBindings...)
		if requirement.Assessment == "connector_local_configuration" {
			available = lane.IntendedBindings
		}
		for i, binding := range requirement.FitBindings {
			matches := func(candidate sourceLaneTargetRef) bool { return sourceLaneTargetRefEqual(binding, candidate) }
			if !slices.ContainsFunc(available, matches) || slices.ContainsFunc(requirement.FitBindings[:i], matches) {
				return "", fmt.Errorf("configuration fit is not an exact independently reconciled binding")
			}
			if requirement.Assessment == "connector_local_configuration" &&
				(!slices.Contains(requirement.AffectedArtifacts, binding.Artifact) || slices.ContainsFunc(lane.References, matches)) {
				return "", fmt.Errorf("local configuration work is absent or already materialized")
			}
		}
		return requirement.Assessment, nil
	case "no_demand_for_this_cell":
		if lane.Applicability != "not_applicable" || lane.Reason.Code != "source_exclusion" ||
			requirement.SourceExclusion == nil || !slices.Contains(lane.FactRefs, *requirement.SourceExclusion) ||
			!slices.Contains(requirement.SourceRefs, *requirement.SourceExclusion) {
			return "", fmt.Errorf("no-demand claim lacks the independently established source exclusion")
		}
		return "no_demand_for_this_cell", nil
	case "absent_shared_foundation":
		if lane.State != "missing_foundation" || len(lane.GapRefs) == 0 || !slices.Contains(lane.OwnerRefs, assessment.NextOwner) {
			return "", fmt.Errorf("foundation absence lacks an existing source-bound gap and next owner")
		}
		for _, candidate := range requirement.AtlasLookup.Candidates {
			if candidate.Disposition != "absent_shared_foundation" {
				return "", fmt.Errorf("foundation absence lacks examined owner/seam mismatches")
			}
		}
		return "absent_shared_foundation", nil
	case "provider_limitation":
		if requirement.ProviderClause == nil || !slices.Contains(requirement.SourceRefs, *requirement.ProviderClause) {
			return "", fmt.Errorf("provider limitation clause missing")
		}
		for _, document := range observed.universe.manifest.Documents {
			if document.ID != requirement.ProviderClause.DocumentID {
				continue
			}
			value, err := resolveSourceFactValue(document, *requirement.ProviderClause)
			var clause string
			if err == nil && json.Unmarshal(value, &clause) == nil && strings.TrimSpace(clause) != "" {
				return "provider_limitation", nil
			}
		}
		return "", fmt.Errorf("provider limitation must cite the retained textual clause")
	default:
		return "", fmt.Errorf("unknown requirement assessment")
	}
}
