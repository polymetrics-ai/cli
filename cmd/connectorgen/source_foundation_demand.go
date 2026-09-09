package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
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
	producer    sourceLaneManifest
	admission   *sourceFoundationAdmissionCustody
	cohort      sourceLaneCohort
	annotations []sourceSemanticAnnotation
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
	return buildSourceFoundationUniverseWithCache(ctx, repo, cohort, annotations, nil)
}

func buildSourceFoundationUniverseWithCache(ctx context.Context, repo string, cohort sourceLaneCohort, annotations []sourceSemanticAnnotation, cache *sourceProofFileCache) (sourceFoundationUniverse, error) {
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
	var admission *sourceFoundationAdmissionCustody
	if manifest.Validation.Status != "valid" {
		manifest, admission, err = buildSourceFoundationAdmissionObserved(ctx, repo, manifest, cohort, annotations, cache)
		if err != nil {
			return result, fmt.Errorf("foundation source universe: %w", err)
		}
	}
	owners, pin := loadSourceDemandAtlas(ctx, repo)
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if len(owners) == 0 || pin.Path != sourceDemandAtlasPath {
		return result, fmt.Errorf("foundation source universe: atlas unavailable")
	}
	result = sourceFoundationUniverse{
		producer: manifest, manifest: manifest, cells: []sourceFoundationCell{}, atlasOwners: owners, atlasPin: pin, admission: admission,
	}
	// Keep independent copies of arguments for a later valid-manifest
	// intended-fit witness. A caller cannot alter retained admission inputs.
	cohortRaw, _ := json.Marshal(cohort)
	annotationRaw, _ := json.Marshal(annotations)
	if json.Unmarshal(cohortRaw, &result.cohort) != nil || json.Unmarshal(annotationRaw, &result.annotations) != nil {
		return sourceFoundationUniverse{}, fmt.Errorf("foundation admission arguments invalid")
	}
	for _, row := range manifest.SourceOperations {
		for _, cell := range row.Lanes {
			result.cells = append(result.cells, sourceFoundationCell{Key: row.Source.Key, Lane: cell.Lane})
		}
	}
	return result, nil
}

func buildSourceFoundationAdmissionObserved(ctx context.Context, repo string, manifest sourceLaneManifest, cohort sourceLaneCohort, annotations []sourceSemanticAnnotation, cache *sourceProofFileCache) (actual sourceLaneManifest, custody *sourceFoundationAdmissionCustody, err error) {
	defer func() { err = errors.Join(err, ctx.Err()) }()
	if cache != nil {
		return buildSourceFoundationAdmission(ctx, repo, manifest, cohort, annotations, cache)
	}
	root, err := os.OpenRoot(repo)
	if err != nil {
		return manifest, nil, err
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	owned := sourceProofFileCache{ctx: ctx, root: root, limits: sourceProofFileLimits{UniqueBytes: 512 << 20, Files: 65536}, files: map[string]*sourceProofFile{}}
	actual, custody, err = buildSourceFoundationAdmission(ctx, repo, manifest, cohort, annotations, &owned)
	if err != nil {
		return manifest, nil, err
	}
	owned.finalize()
	if err = revalidateSourceFoundationAdmission(&owned, custody); err != nil {
		return manifest, nil, err
	}
	return actual, custody, nil
}
