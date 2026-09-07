package main

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
)

// The register is an authoring observation. None of its types can be supplied
// to the lane reducer as accepted execution or proof authority.
type sourceFoundationRegister struct {
	SourceAdmission sourceFoundationAdmission                `json:"source_admission"`
	ProofDocument   sourceFoundationProofDocumentObservation `json:"foundation_proof_document"`
	SchemaVersion   int                                      `json:"schema_version"`
	Kind            string                                   `json:"kind"`
	Atlas           sourceArtifactPin                        `json:"atlas"`
	Assessments     sourceArtifactPin                        `json:"assessments"`
	Baseline        sourceArtifactPin                        `json:"baseline"`
	SourceSHA256    string                                   `json:"source_manifest_content_sha256"`
	Coverage        sourceFoundationCoverage                 `json:"coverage"`
	Known           []sourceFoundationKnownState             `json:"known_obligations"`
	Requirements    []sourceFoundationRequirementResult      `json:"requirements"`
	Facets          []sourceFoundationFacet                  `json:"source_fit"`
	Examples        []sourceFoundationExample                `json:"atlas_examples"`
	Adopters        []sourceFoundationAdoption               `json:"adopter_relations"`
	Inputs          []sourceArtifactPin                      `json:"inputs"`
	Proofs          []sourceFoundationObservedProof          `json:"foundation_proof_observations"`
	Checks          sourceFoundationRegisterChecks           `json:"checks"`
}

type sourceFoundationObservedProof struct {
	Record sourceFoundationProofRecord  `json:"record"`
	Status string                       `json:"status"`
	Issues []sourceFoundationProofIssue `json:"issues"`
}

type sourceFoundationRegisterChecks struct {
	StructuralValid            bool   `json:"structural_valid"`
	CoverageAccounted          bool   `json:"coverage_accounted"`
	RequiredReconciliation     bool   `json:"required_reconciliation_complete"`
	SelectedReuseProofComplete bool   `json:"selected_reuse_proof_complete"`
	SelectedReuseRequirements  int    `json:"selected_reuse_requirements"`
	UnresolvedRequirements     int    `json:"unresolved_requirements"`
	Authority                  string `json:"authority"`
}

type sourceFoundationKnownState struct {
	Identity           sourceFoundationCell `json:"identity"`
	SourceRefs         []sourceFactRef      `json:"source_refs"`
	SourceDocumentPins []sourceArtifactPin  `json:"source_document_pins"`
	SourceState        string               `json:"source_state"`
	Applicability      string               `json:"source_applicability"`
	HistoricalReceiver bool                 `json:"historical_receiver_candidate"`
	SentryRegistration bool                 `json:"sentry_registration"`
}

type sourceFoundationAdoption struct {
	ProofID          string                      `json:"proof_id"`
	Mechanism        string                      `json:"mechanism"`
	DeclarationState string                      `json:"declaration_state"`
	Selectors        []sourceFoundationSelector  `json:"selectors"`
	SourceRefs       []sourceFactRef             `json:"source_refs"`
	AtlasID          string                      `json:"atlas_id"`
	Contract         sourceFoundationContractRef `json:"contract"`
	Identity         sourceFoundationCell        `json:"identity"`
	Requirement      string                      `json:"requirement_id"`
	Binding          sourceLaneTargetRef         `json:"binding"`
	Relationship     string                      `json:"relationship"`
}

func buildSourceFoundationRegister(ctx context.Context, repo string, inputs sourceFoundationDemandInputs) (result sourceFoundationRegister, err error) {
	build := func() error {
		var buildErr error
		result, buildErr = buildSourceFoundationRegisterCurrent(ctx, repo, inputs)
		return buildErr
	}
	if inputs.admissionCache != nil {
		// The command owns final content/identity/namespace validation before its
		// one output write. Nested reconstruction shares that bounded observation.
		err = revalidateSourceFoundationAdmission(inputs.admissionCache, inputs.assessments.universe.admission)
		if err == nil {
			err = build()
		}
		if err == nil {
			err = revalidateSourceFoundationAdmission(inputs.admissionCache, inputs.assessments.universe.admission)
		}
	} else {
		err = sourceFoundationAdmissionWithRoot(ctx, repo, inputs.assessments.universe.admission, build)
	}
	if err != nil {
		return sourceFoundationRegister{}, err
	}
	return result, nil
}

func buildSourceFoundationRegisterCurrent(ctx context.Context, repo string, inputs sourceFoundationDemandInputs) (sourceFoundationRegister, error) {
	var result sourceFoundationRegister
	observed := inputs.assessments
	if err := validateSourceFoundationObligations(ctx, inputs.baseline, observed); err != nil {
		return result, err
	}
	coverage, err := buildSourceFoundationCoverage(observed)
	if err != nil {
		return result, err
	}
	if err := validateSourceFoundationCoverage(coverage, observed); err != nil {
		return result, err
	}
	batch, err := readSourceFoundationProofBatch(ctx, repo, reviewedSourceFoundationProofs(), nil)
	if err != nil {
		return result, err
	}
	proofs := batch.records
	requirements, err := buildSourceFoundationRequirementsBatch(ctx, observed, batch)
	if err != nil {
		return result, err
	}
	facets, err := buildSourceFoundationFacets(ctx, observed)
	if err != nil {
		return result, err
	}
	examples, err := observeSourceFoundationExamples(ctx, repo, observed.atlas.pin)
	if err != nil {
		return result, err
	}
	source, err := json.Marshal(observed.universe.manifest)
	if err != nil {
		return result, fmt.Errorf("foundation source manifest encoding: %w", err)
	}
	result = sourceFoundationRegister{SchemaVersion: 1, Kind: "foundation_demand_register",
		ProofDocument: batch.document, Atlas: observed.atlas.pin, Assessments: observed.pin, Baseline: inputs.baselinePin,
		SourceSHA256: sourceBytesHash(source), Coverage: coverage, Requirements: requirements, Facets: facets, Examples: examples,
		Known: []sourceFoundationKnownState{}, Adopters: []sourceFoundationAdoption{},
		Inputs: append([]sourceArtifactPin{}, inputs.commandPins...), Proofs: []sourceFoundationObservedProof{}}
	result.SourceAdmission = sourceFoundationCurrentAdmission(observed.universe.manifest)
	if observed.universe.admission != nil {
		result.SourceAdmission = observed.universe.admission.observation
	}
	for _, pin := range batch.pins {
		found := false
		for _, prior := range result.Inputs {
			if prior.Path == pin.Path {
				if prior != pin {
					return sourceFoundationRegister{}, fmt.Errorf("foundation proof input changed between readers")
				}
				found = true
			}
		}
		if !found {
			result.Inputs = append(result.Inputs, pin)
		}
	}
	sort.Slice(result.Inputs, func(i, j int) bool { return result.Inputs[i].Path < result.Inputs[j].Path })
	for _, proof := range proofs {
		result.Proofs = append(result.Proofs, sourceFoundationObservedProof{Record: proof.record, Status: proof.status, Issues: append([]sourceFoundationProofIssue{}, proof.issues...)})
	}
	sort.Slice(result.Proofs, func(i, j int) bool { return result.Proofs[i].Record.ID < result.Proofs[j].Record.ID })
	historical, sentry := map[sourceFoundationCell]bool{}, map[sourceFoundationCell]bool{}
	for _, cell := range inputs.baseline.HistoricalReceiverCandidates {
		historical[cell] = true
	}
	for _, cell := range inputs.baseline.SentryRegistration {
		sentry[cell] = true
	}
	rows := map[sourceOperationKey]sourceLaneManifestRow{}
	for _, row := range observed.universe.manifest.SourceOperations {
		rows[row.Source.Key] = row
	}
	for _, obligation := range inputs.baseline.Obligations {
		row := sourceFoundationKnownState{Identity: obligation.Identity,
			SourceRefs: append([]sourceFactRef{}, obligation.SourceRefs...), SourceDocumentPins: append([]sourceArtifactPin{}, obligation.SourceDocumentPins...),
			HistoricalReceiver: historical[obligation.Identity], SentryRegistration: sentry[obligation.Identity]}
		for _, cell := range rows[obligation.Identity.Key].Lanes {
			if cell.Lane == obligation.Identity.Lane {
				row.SourceState, row.Applicability = cell.State, cell.Applicability
			}
		}
		result.Known = append(result.Known, row)
	}
	knownCells := map[sourceFoundationCell]bool{}
	for _, row := range result.Known {
		knownCells[row.Identity] = true
	}
	// Completed admission independently requires local work. Its affected cells
	// belong to K even when the pinned historical baseline predates them.
	admissionRequired := map[sourceFoundationCell]bool{}
	if observed.universe.admission != nil {
		for _, group := range observed.universe.admission.observation.MissingCLI {
			for _, affected := range group.Affected {
				admissionRequired[affected.Identity] = true
			}
		}
	}
	for _, source := range observed.universe.manifest.SourceOperations {
		registrationRef, registration := sourceRegistrationDemand(source.Facts)
		for _, lane := range source.Lanes {
			identity := sourceFoundationCell{Key: source.Source.Key, Lane: lane.Lane}
			required := admissionRequired[identity] || slices.Contains(lane.OwnerRefs, "CP13") || len(lane.GapRefs) != 0 || registration && lane.Lane == "sync_transport"
			if !required || knownCells[identity] {
				continue
			}
			row := sourceFoundationKnownState{Identity: identity, SourceState: lane.State, Applicability: lane.Applicability,
				SourceRefs: []sourceFactRef{}, SourceDocumentPins: []sourceArtifactPin{}}
			if registration && lane.Lane == "sync_transport" {
				row.SourceRefs = append(row.SourceRefs, registrationRef)
			} else {
				row.SourceRefs = append(row.SourceRefs, lane.FactRefs...)
			}
			documentIDs := map[string]bool{source.Source.DocumentID: true}
			if source.Source.RawDocumentID != "" {
				documentIDs[source.Source.RawDocumentID] = true
			}
			for _, ref := range row.SourceRefs {
				documentIDs[ref.DocumentID] = true
			}
			pins := map[sourceArtifactPin]bool{}
			for _, document := range observed.universe.manifest.Documents {
				pin := sourceArtifactPin{Path: document.Path, SHA256: document.RetainedFileSHA256, Bytes: document.Bytes}
				if documentIDs[document.ID] && !pins[pin] {
					row.SourceDocumentPins = append(row.SourceDocumentPins, pin)
					pins[pin] = true
				}
			}
			sort.Slice(row.SourceDocumentPins, func(i, j int) bool { return row.SourceDocumentPins[i].Path < row.SourceDocumentPins[j].Path })
			knownCells[identity] = true
			result.Known = append(result.Known, row)
		}
	}
	sort.Slice(result.Known, func(i, j int) bool {
		return sourceFoundationCellLess(result.Known[i].Identity, result.Known[j].Identity)
	})
	// Derive relations from actual admitted requirement/proof/selector tuples,
	// not Atlas example names or a separately authored adopter list.
	seen := map[string]bool{}
	for _, requirement := range requirements {
		if requirement.Status != "existing_shared_capability" && requirement.Status != "connector_local_configuration" {
			continue
		}
		for _, fit := range requirement.MechanismFits {
			row := sourceFoundationAdoption{ProofID: fit.ProofID, Mechanism: fit.Mechanism, DeclarationState: fit.DeclarationState, Selectors: append([]sourceFoundationSelector{}, fit.Selectors...), SourceRefs: append([]sourceFactRef{}, fit.SourceRefs...), AtlasID: fit.AtlasID, Contract: fit.Contract, Identity: requirement.Identity, Requirement: requirement.ID, Binding: fit.Binding, Relationship: requirement.Status}
			encoded, err := json.Marshal(row)
			if err != nil {
				return sourceFoundationRegister{}, err
			}
			if !seen[string(encoded)] {
				seen[string(encoded)] = true
				result.Adopters = append(result.Adopters, row)
			}
		}
	}
	if err := ctx.Err(); err != nil {
		return sourceFoundationRegister{}, err
	}
	result.Checks = sourceFoundationRegisterCompletion(result, observed.atlas)
	return result, nil
}

func sourceFoundationRegisterCompletion(register sourceFoundationRegister, atlas sourceFoundationAtlas) sourceFoundationRegisterChecks {
	result := sourceFoundationRegisterChecks{StructuralValid: true, CoverageAccounted: true,
		RequiredReconciliation: true, SelectedReuseProofComplete: true,
		Authority: "authoring_consistency_only; no lane authority or checkpoint acceptance"}
	requirements := map[sourceFoundationCell]int{}
	for _, requirement := range register.Requirements {
		requirements[requirement.Identity]++
		if requirement.Status == "unresolved" {
			result.UnresolvedRequirements++
		}
		if requirement.Status == "existing_shared_capability" || requirement.Status == "connector_local_configuration" {
			result.SelectedReuseRequirements++
			if len(requirement.Proofs) == 0 || len(requirement.FitBindings) == 0 || len(requirement.MechanismFits) != len(requirement.Proofs)*len(requirement.FitBindings) {
				result.SelectedReuseProofComplete = false
			}
			for _, proof := range requirement.Proofs {
				if proof.Status != "current" || proof.Assertion != requirement.Statement {
					result.SelectedReuseProofComplete = false
				}
			}
		}
	}
	for _, known := range register.Known {
		if requirements[known.Identity] == 0 {
			result.RequiredReconciliation = false
		}
	}
	for _, cell := range register.Coverage.Assessed {
		kinds := map[string]bool{}
		for _, facet := range register.Facets {
			if facet.Identity == cell {
				kinds[facet.Kind] = true
			}
		}
		if !kinds["mime"] || !kinds["auth"] || !kinds["body"] || !kinds["paging"] {
			result.RequiredReconciliation = false
		}
	}
	examples := 0
	for _, entry := range atlas.entries {
		examples += len(entry.ConsumerExamples)
	}
	if len(register.Examples) != examples {
		result.RequiredReconciliation = false
	}
	return result
}
