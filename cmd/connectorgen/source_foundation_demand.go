package main

import (
	"context"
	"fmt"
)

// sourceFoundationCell identifies an authoring assessment, never execution
// authority. In particular it is not a sourceLaneProofCell or a lane target.
type sourceFoundationCell struct {
	Key  sourceOperationKey `json:"key"`
	Lane string             `json:"lane"`
}

// The universe is produced from retained inputs before any assessment is read.
// Its manifest is kept intact so assessments cannot rewrite source/lane facts.
type sourceFoundationUniverse struct {
	manifest    sourceLaneManifest
	cells       []sourceFoundationCell
	atlasOwners map[string]string
	atlasPin    sourceArtifactPin
}

// buildSourceFoundationUniverse is the non-authorizing projection frontier.
// It uses the actual retained-source builder and confined Atlas reader. No
// foundation assertion, assessment, proof reference or capability is admitted
// by this step. Later assessment admission consumes this independently built U.
func buildSourceFoundationUniverse(ctx context.Context, repo string, cohort sourceLaneCohort, annotations []sourceSemanticAnnotation) (sourceFoundationUniverse, error) {
	var result sourceFoundationUniverse
	if err := ctx.Err(); err != nil {
		return result, err
	}
	manifest, err := buildSourceLaneManifest(ctx, repo, cohort, annotations)
	if err != nil {
		return result, fmt.Errorf("foundation source universe: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if manifest.Validation.Status != "valid" {
		return result, fmt.Errorf("foundation source universe: retained inputs invalid")
	}
	owners, pin := loadSourceDemandAtlas(ctx, repo)
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if len(owners) == 0 || pin.Path != sourceDemandAtlasPath {
		return result, fmt.Errorf("foundation source universe: atlas unavailable")
	}
	result = sourceFoundationUniverse{
		manifest: manifest, cells: []sourceFoundationCell{}, atlasOwners: owners, atlasPin: pin,
	}
	for _, row := range manifest.SourceOperations {
		for _, cell := range row.Lanes {
			result.cells = append(result.cells, sourceFoundationCell{Key: row.Source.Key, Lane: cell.Lane})
		}
	}
	return result, nil
}
