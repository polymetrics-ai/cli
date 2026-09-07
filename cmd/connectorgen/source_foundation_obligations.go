package main

import (
	"context"
	"fmt"
	"os"
	"slices"
)

const sourceFoundationObligationsPath = "data/connector-canon/batch1-foundation-obligations.json"

// This pin retains the source-bound 140A seed independently of editable
// assessments. It is authoring policy, not a list of approved receivers.
func sourceFoundationBaselinePin() sourceArtifactPin {
	return sourceArtifactPin{Path: sourceFoundationObligationsPath,
		SHA256: "5c7b972ad6eccb01a6b2633ac7505d2467d5667c7cd7a873b018e0cef48a5b05", Bytes: 27715}
}

type sourceFoundationKnownObligation struct {
	Identity   sourceFoundationCell    `json:"identity"`
	Source     retainedSourceOperation `json:"source"`
	SourceRefs []sourceFactRef         `json:"source_refs"`
}

type sourceFoundationObligationDocument struct {
	SchemaVersion                int                               `json:"schema_version"`
	Kind                         string                            `json:"kind"`
	BaselineManifest             sourceArtifactPin                 `json:"baseline_manifest"`
	SourceEvidenceSHA256         string                            `json:"source_evidence_sha256"`
	Obligations                  []sourceFoundationKnownObligation `json:"obligations"`
	HistoricalReceiverCandidates []sourceFoundationCell            `json:"historical_receiver_candidates"`
	SentryRegistration           []sourceFoundationCell            `json:"sentry_registration"`
}

type sourceFoundationDemandInputs struct {
	assessments sourceFoundationAssessmentObservations
	baseline    sourceFoundationObligationDocument
	baselinePin sourceArtifactPin
}

// The expected pin is supplied by the source-owned policy, never by the
// authored assessment document or command-line input.
func observeSourceFoundationDemandInputs(ctx context.Context, repo, assessments string, universe sourceFoundationUniverse, expected sourceArtifactPin) (sourceFoundationDemandInputs, error) {
	var result sourceFoundationDemandInputs
	observed, err := observeSourceFoundationAssessments(ctx, repo, assessments, universe)
	if err != nil {
		return result, err
	}
	root, err := os.OpenRoot(repo)
	if err != nil {
		return result, fmt.Errorf("foundation obligations root: %w", err)
	}
	defer func() { _ = root.Close() }()
	cache := sourceProofFileCache{ctx: ctx, root: root,
		limits: sourceProofFileLimits{UniqueBytes: 4 << 20, Files: 1}, files: map[string]*sourceProofFile{}}
	raw, file := cache.get(expected.Path, 4<<20, false)
	if file.code != "" || file.hash != expected.SHA256 || file.size != expected.Bytes {
		return result, fmt.Errorf("foundation obligation baseline pin mismatch")
	}
	var document sourceFoundationObligationDocument
	if decodeSourceJSON(raw, &document) != nil || decodeStrictJSON(raw, &document) != nil ||
		document.SchemaVersion != 1 || document.Kind != "foundation_known_obligations" || document.Obligations == nil ||
		document.HistoricalReceiverCandidates == nil || document.SentryRegistration == nil {
		return result, fmt.Errorf("foundation obligation baseline invalid")
	}
	if err := validateSourceFoundationObligations(ctx, document, observed); err != nil {
		return result, err
	}
	cache.finalize()
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if file.code != "" {
		return result, fmt.Errorf("foundation obligation baseline changed")
	}
	return sourceFoundationDemandInputs{assessments: observed, baseline: document, baselinePin: expected}, nil
}

func validateSourceFoundationObligations(ctx context.Context, baseline sourceFoundationObligationDocument, observed sourceFoundationAssessmentObservations) error {
	rows := map[sourceOperationKey]sourceLaneManifestRow{}
	for _, row := range observed.universe.manifest.SourceOperations {
		rows[row.Source.Key] = row
	}
	documents := map[string]retainedSourceDocument{}
	for _, document := range observed.universe.manifest.Documents {
		documents[document.ID] = document
	}
	assessed := map[sourceFoundationCell]bool{}
	for _, row := range observed.authored {
		assessed[sourceFoundationCell{Key: row.Key, Lane: row.Lane}] = true
	}
	known := map[sourceFoundationCell]bool{}
	for _, obligation := range baseline.Obligations {
		if err := ctx.Err(); err != nil {
			return err
		}
		cell := obligation.Identity
		row, exists := rows[cell.Key]
		if !exists || known[cell] || !slices.Contains(sourceLaneNames(), cell.Lane) || !assessed[cell] {
			return fmt.Errorf("foundation independent obligation missing, duplicate or outside source universe: %s/%s", cell.Key.ID, cell.Lane)
		}
		known[cell] = true
		before, current := obligation.Source, row.Source
		if before.Key != cell.Key || before.Key != current.Key || before.Class != current.Class ||
			before.DocumentID != current.DocumentID || before.RawDocumentID != current.RawDocumentID ||
			before.Pointer != current.Pointer || before.SourceLocation != current.SourceLocation ||
			!before.Observed || !current.Observed || len(obligation.SourceRefs) == 0 {
			return fmt.Errorf("foundation independent obligation source changed without retained disposition")
		}
		refs := map[sourceFactRef]bool{}
		for _, ref := range obligation.SourceRefs {
			if refs[ref] || !sourceFoundationRequirementCitation(row.Facts, documents[ref.DocumentID], ref) {
				return fmt.Errorf("foundation independent obligation citation changed without retained disposition")
			}
			refs[ref] = true
		}
	}
	// These historical contracts remain separate even while they overlap K.
	for _, group := range [][]sourceFoundationCell{baseline.HistoricalReceiverCandidates, baseline.SentryRegistration} {
		seen := map[sourceFoundationCell]bool{}
		for _, cell := range group {
			if seen[cell] || !known[cell] {
				return fmt.Errorf("foundation historical obligation is duplicate or missing from baseline")
			}
			seen[cell] = true
		}
	}
	return nil
}
