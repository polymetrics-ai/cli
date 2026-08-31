package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// sourceOperationMappingCohortMatrixReport is source-accounting evidence for
// the real connector-local matrix corpus. It deliberately describes no
// command, descriptor, stream, write, transport, credential, executor, or
// runtime declaration. A deferred cell without a declared artifact link is a
// typed projection deficit, not an admission failure or an executable claim.
type sourceOperationMappingCohortMatrixReport struct {
	SourceRows                  int                                             `json:"source_rows"`
	MatrixCells                 int                                             `json:"matrix_cells"`
	DeferredCells               int                                             `json:"deferred_cells"`
	DeclaredArtifactLinkRecords int                                             `json:"declared_artifact_link_records"`
	DeclaredArtifactLaneLinks   int                                             `json:"declared_artifact_lane_links"`
	DeferredProjectionDeficits  []sourceOperationMappingCohortProjectionDeficit `json:"deferred_projection_deficits"`
	ExecutableDeclarations      int                                             `json:"executable_declarations"`
}

// sourceOperationMappingCohortProjectionDeficit identifies a source-backed
// deferred cell that has no declared connector-local artifact link. It records
// an authoring/projection gap only; it neither rejects the source row nor
// asserts that a runtime artifact ought to exist yet.
type sourceOperationMappingCohortProjectionDeficit struct {
	Kind              string `json:"kind"`
	Connector         string `json:"connector"`
	SourceOperationID string `json:"source_operation_id"`
	Lane              string `json:"lane"`
	State             string `json:"state"`
}

const sourceOperationMappingCohortDeferredProjectionDeficitKind = "deferred_without_declared_artifact_link"

// sourceOperationMappingCohortFullCheck composes the immutable frozen
// denominator checker with the source/citation/hidden-row validation already
// used by deferred-visibility. It adds integrity checking only for explicit
// connector-local artifact links; it never requires a deferred cell to have a
// link, and it never creates an executable declaration.
//
// The denominator checker remains separate because deferred-visibility uses
// it as its non-recursive source-accounting prerequisite.
func sourceOperationMappingCohortFullCheck(root, manifestPath string) (sourceOperationMappingCohortReport, sourceOperationMappingCohortMatrixReport, error) {
	repositoryRoot, err := sourceOperationMappingCohortRoot(root)
	if err != nil {
		return sourceOperationMappingCohortReport{}, sourceOperationMappingCohortMatrixReport{}, err
	}
	cohort, err := sourceOperationMappingCohortPathCheck(repositoryRoot, manifestPath)
	if err != nil {
		return sourceOperationMappingCohortReport{}, sourceOperationMappingCohortMatrixReport{}, err
	}
	if len(cohort.Findings) != 0 {
		return cohort, sourceOperationMappingCohortMatrixReport{}, nil
	}

	visibility, err := deferredVisibilityFromRepository(repositoryRoot, manifestPath)
	if err != nil {
		return cohort, sourceOperationMappingCohortMatrixReport{}, fmt.Errorf("validate source-lane matrices: %w", err)
	}
	manifest, err := sourceOperationMappingCohortReadManifest(manifestPath)
	if err != nil {
		return cohort, sourceOperationMappingCohortMatrixReport{}, err
	}
	laneLinks, artifactLinks, artifactLaneLinks, err := sourceOperationMappingCohortDeclaredArtifactLinks(repositoryRoot, manifest)
	if err != nil {
		return cohort, sourceOperationMappingCohortMatrixReport{}, fmt.Errorf("validate declared artifact links: %w", err)
	}

	report := sourceOperationMappingCohortMatrixReport{
		SourceRows:                  visibility.SourceRows,
		MatrixCells:                 visibility.MatrixCells,
		DeferredCells:               visibility.DeferredCells,
		DeclaredArtifactLinkRecords: artifactLinks,
		DeclaredArtifactLaneLinks:   artifactLaneLinks,
		ExecutableDeclarations:      visibility.ExecutableDeclarations,
		DeferredProjectionDeficits:  make([]sourceOperationMappingCohortProjectionDeficit, 0),
	}
	for _, entry := range visibility.Entries {
		key := sourceOperationMappingCohortArtifactLaneKey(entry.Connector, entry.SourceOperationID, entry.Lane)
		if laneLinks[key] {
			continue
		}
		report.DeferredProjectionDeficits = append(report.DeferredProjectionDeficits, sourceOperationMappingCohortProjectionDeficit{
			Kind:              sourceOperationMappingCohortDeferredProjectionDeficitKind,
			Connector:         entry.Connector,
			SourceOperationID: entry.SourceOperationID,
			Lane:              entry.Lane,
			State:             entry.SourceDisposition,
		})
	}
	sort.Slice(report.DeferredProjectionDeficits, func(i, j int) bool {
		left, right := report.DeferredProjectionDeficits[i], report.DeferredProjectionDeficits[j]
		if left.Connector != right.Connector {
			return left.Connector < right.Connector
		}
		if left.SourceOperationID != right.SourceOperationID {
			return left.SourceOperationID < right.SourceOperationID
		}
		return left.Lane < right.Lane
	})
	return cohort, report, nil
}

// sourceOperationMappingCohortDeclaredArtifactLinks accepts only the two
// explicit matrix dialects already present in the frozen corpus:
// source_id/source_operation_id and lane/lanes. Provider record labels remain
// opaque metadata. The parser deliberately does not infer links from methods,
// schemas, command names, or provider-specific conventions.
func sourceOperationMappingCohortDeclaredArtifactLinks(root string, manifest sourceOperationMappingCohortManifest) (map[string]bool, int, int, error) {
	repositoryRoot, err := sourceOperationMappingCohortRoot(root)
	if err != nil {
		return nil, 0, 0, err
	}
	laneLinks := make(map[string]bool)
	seenArtifactLaneLinks := make(map[string]bool)
	artifactLinks := 0
	artifactLaneLinks := 0
	for _, entry := range manifest.SourceLocks {
		matrixPath, err := deferredVisibilityOwnedMatrixPath(repositoryRoot, entry.Connector, entry.MatrixPath)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("%s matrix path: %w", entry.Connector, err)
		}
		raw, err := os.ReadFile(matrixPath)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("read %s source-lane matrix: %w", entry.Connector, err)
		}
		defaultLockPath, err := deferredVisibilityRelativeSourceLockPath(entry.Connector, entry.Path)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("%s primary source lock: %w", entry.Connector, err)
		}
		matrix, err := decodeDeferredVisibilityMatrix(raw, entry.Connector, defaultLockPath)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("decode %s source-lane matrix: %w", entry.Connector, err)
		}
		rows := make(map[string]deferredVisibilityMatrixRow, len(matrix.Rows))
		for _, row := range matrix.Rows {
			if _, duplicate := rows[row.SourceID]; duplicate {
				return nil, 0, 0, fmt.Errorf("%s source-lane matrix duplicates source ID %q", entry.Connector, row.SourceID)
			}
			rows[row.SourceID] = row
		}

		matrixRoot, err := sourceOperationMappingCohortMatrixRoot(raw)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("decode %s source-lane matrix artifacts: %w", entry.Connector, err)
		}
		rawArtifacts, found := matrixRoot["artifacts"]
		if !found || rawArtifacts == nil {
			continue
		}
		artifacts, err := retainedSourceMappingArray(rawArtifacts, entry.Connector+".artifacts")
		if err != nil {
			return nil, 0, 0, err
		}
		for index, rawArtifact := range artifacts {
			field := fmt.Sprintf("%s.artifacts[%d]", entry.Connector, index)
			artifact, err := retainedSourceMappingObject(rawArtifact, field)
			if err != nil {
				return nil, 0, 0, err
			}
			artifactPath, err := retainedSourceMappingString(artifact["path"], field+".path")
			if err != nil {
				return nil, 0, 0, err
			}
			if err := sourceOperationMappingCohortArtifactPath(repositoryRoot, entry.Connector, artifactPath); err != nil {
				return nil, 0, 0, fmt.Errorf("%s path %q: %w", field, artifactPath, err)
			}
			rawLinks, found := artifact["links"]
			if !found || rawLinks == nil {
				continue
			}
			links, err := retainedSourceMappingArray(rawLinks, field+".links")
			if err != nil {
				return nil, 0, 0, err
			}
			for linkIndex, rawLink := range links {
				linkField := fmt.Sprintf("%s.links[%d]", field, linkIndex)
				link, err := retainedSourceMappingObject(rawLink, linkField)
				if err != nil {
					return nil, 0, 0, err
				}
				sourceID, err := sourceOperationMappingCohortArtifactLinkSourceID(link, linkField)
				if err != nil {
					return nil, 0, 0, err
				}
				row, found := rows[sourceID]
				if !found {
					return nil, 0, 0, fmt.Errorf("%s source operation %q does not resolve to a source-lane matrix row", linkField, sourceID)
				}
				lanes, err := sourceOperationMappingCohortArtifactLinkLanes(link, linkField)
				if err != nil {
					return nil, 0, 0, err
				}
				artifactLinks++
				for _, lane := range lanes {
					if _, found := row.Cells[lane]; !found {
						return nil, 0, 0, fmt.Errorf("%s lane %q does not resolve to source row %q", linkField, lane, sourceID)
					}
					key := sourceOperationMappingCohortArtifactLaneKey(entry.Connector, sourceID, lane)
					duplicateKey := artifactPath + "\x00" + key
					if seenArtifactLaneLinks[duplicateKey] {
						return nil, 0, 0, fmt.Errorf("%s duplicates declared artifact source-lane link %q/%q", linkField, sourceID, lane)
					}
					seenArtifactLaneLinks[duplicateKey] = true
					laneLinks[key] = true
					artifactLaneLinks++
				}
			}
		}
	}
	return laneLinks, artifactLinks, artifactLaneLinks, nil
}

func sourceOperationMappingCohortMatrixRoot(raw []byte) (map[string]any, error) {
	var value any
	if err := decodeSourceJSON(raw, &value); err != nil {
		return nil, err
	}
	return retainedSourceMappingObject(value, "source-lane matrix")
}

func sourceOperationMappingCohortArtifactLinkSourceID(link map[string]any, field string) (string, error) {
	rawSourceID, hasSourceID := link["source_id"]
	rawSourceOperationID, hasSourceOperationID := link["source_operation_id"]
	if hasSourceID == hasSourceOperationID {
		return "", fmt.Errorf("%s must declare exactly one of source_id or source_operation_id", field)
	}
	if hasSourceID {
		return retainedSourceMappingString(rawSourceID, field+".source_id")
	}
	return retainedSourceMappingString(rawSourceOperationID, field+".source_operation_id")
}

func sourceOperationMappingCohortArtifactLinkLanes(link map[string]any, field string) ([]string, error) {
	rawLane, hasLane := link["lane"]
	rawLanes, hasLanes := link["lanes"]
	if hasLane == hasLanes {
		return nil, fmt.Errorf("%s must declare exactly one of lane or lanes", field)
	}
	if hasLane {
		lane, err := retainedSourceMappingString(rawLane, field+".lane")
		if err != nil {
			return nil, err
		}
		if !retainedSourceMappingKnownLane(lane) {
			return nil, fmt.Errorf("%s has unknown lane %q", field, lane)
		}
		return []string{lane}, nil
	}
	lanes, err := retainedSourceMappingStringArray(rawLanes, field+".lanes")
	if err != nil {
		return nil, err
	}
	if len(lanes) == 0 {
		return nil, fmt.Errorf("%s.lanes must not be empty", field)
	}
	seen := make(map[string]bool, len(lanes))
	for _, lane := range lanes {
		if !retainedSourceMappingKnownLane(lane) {
			return nil, fmt.Errorf("%s has unknown lane %q", field, lane)
		}
		if seen[lane] {
			return nil, fmt.Errorf("%s duplicates lane %q", field, lane)
		}
		seen[lane] = true
	}
	return lanes, nil
}

func sourceOperationMappingCohortArtifactPath(root, connector, raw string) error {
	if raw == "" || filepath.IsAbs(raw) || strings.Contains(raw, "\\") || filepath.ToSlash(filepath.Clean(filepath.FromSlash(raw))) != raw || raw == "." || strings.HasPrefix(raw, "../") {
		return fmt.Errorf("must be one canonical connector-relative path")
	}
	connectorRoot := filepath.Join(root, "internal", "connectors", "defs", connector)
	resolvedConnectorRoot, err := filepath.EvalSymlinks(connectorRoot)
	if err != nil {
		return err
	}
	resolved, err := filepath.EvalSymlinks(filepath.Join(connectorRoot, filepath.FromSlash(raw)))
	if err != nil {
		return err
	}
	relative, err := filepath.Rel(resolvedConnectorRoot, resolved)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("resolves outside connector definition")
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("must resolve to a regular file")
	}
	return nil
}

func sourceOperationMappingCohortArtifactLaneKey(connector, sourceOperationID, lane string) string {
	return connector + "\x00" + sourceOperationID + "\x00" + lane
}
