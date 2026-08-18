package cli

import (
	"errors"
	"os"
	"path/filepath"

	"polymetrics.ai/internal/connectors/database"
)

const maxTransformPlanFileBytes = 64 << 10

// readTransformPlanFile admits one bounded regular JSON file at connection
// creation. The durable connection stores only the parsed canonical form and
// its hash, never this path or a user-authored formatting variant.
func readTransformPlanFile(path string) (database.TransformPlanV1, error) {
	if path == "" || !filepath.IsAbs(path) && filepath.Clean(path) == "." {
		return database.TransformPlanV1{}, database.ErrTransformPlanInvalid
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > maxTransformPlanFileBytes {
		return database.TransformPlanV1{}, database.ErrTransformPlanInvalid
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return database.TransformPlanV1{}, errors.Join(database.ErrTransformPlanInvalid, err)
	}
	return database.ParseTransformPlanV1(raw)
}
