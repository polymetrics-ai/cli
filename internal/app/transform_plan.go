package app

import (
	"context"
	"errors"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
)

// validateConnectionTransformPlan is deliberately called while a connection
// is created, before it becomes durable state. A path is parsed by the CLI,
// but every API caller reaches this same typed, catalog-bound refusal.
func validateConnectionTransformPlan(ctx context.Context, source connectors.Connector, runtime connectors.RuntimeConfig, streamName string, stream *StreamConfig) error {
	if stream == nil {
		return database.ErrTransformPlanInvalid
	}
	hasPlan := strings.TrimSpace(stream.TransformPlan) != ""
	hasHash := strings.TrimSpace(stream.TransformPlanHash) != ""
	if !hasPlan && !hasHash {
		return nil
	}
	if !hasPlan || !hasHash {
		return database.ErrTransformPlanInvalid
	}
	plan, err := database.ParseTransformPlanV1([]byte(stream.TransformPlan))
	if err != nil || plan.Hash() != stream.TransformPlanHash {
		return database.ErrTransformPlanInvalid
	}
	catalog, err := database.CatalogForManagedTargetSource(ctx, source, runtime, streamName)
	if err != nil {
		return errors.Join(database.ErrTransformPlanInvalid, err)
	}
	relations := catalog.Relations()
	if len(relations) != 1 {
		return database.ErrTransformPlanInvalid
	}
	if err := plan.ValidateAgainstRelation(relations[0]); err != nil {
		return err
	}
	stream.TransformPlan = string(plan.NormalizedJSON())
	stream.TransformPlanHash = plan.Hash()
	return nil
}
