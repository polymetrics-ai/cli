package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

const sourceLaneCohortPath = "data/connector-canon/batch1-source-lane-cohort.json"
const sourceLaneAnnotationsPath = "data/connector-canon/batch1-source-lane-annotations.json"
const sourceLaneManifestPath = "data/connector-canon/batch1-source-lane-manifest.json"

func runSourceLanesContext(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	repo, manifest := "", sourceLaneManifestPath
	check, help := false, false
	seen := map[string]bool{}
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if arg == "-h" {
			arg = "--help"
		}
		if seen[arg] {
			logln(stderr, "source-lanes: duplicate argument")
			return 2
		}
		seen[arg] = true
		switch arg {
		case "--check":
			check = true
		case "--help":
			help = true
		case "--repo", "--manifest":
			if i+1 >= len(args) || args[i+1] == "" || strings.HasPrefix(args[i+1], "-") {
				logln(stderr, "source-lanes: missing argument value")
				return 2
			}
			i++
			if arg == "--repo" {
				repo = args[i]
			} else {
				manifest = args[i]
			}
		default:
			logln(stderr, "source-lanes: unknown argument")
			return 2
		}
	}
	if (seen["--manifest"] && !check) || !sourceLaneRelativePath(manifest) {
		logln(stderr, "source-lanes: --manifest requires --check and a confined relative path")
		return 2
	}
	if help {
		return writeSourceLaneBytes(stdout, stderr, []byte("usage: connectorgen source-lanes [--repo <dir>] [--check [--manifest <relative-path>]]\n\nBuild the authoring-only retained source/lane report as JSON on stdout.\n--check reads and validates a saved report without writing files.\nExit 0 means source/report consistency, not executable connector completeness.\n"))
	}
	if repo == "" {
		var err error
		repo, err = repoRoot()
		if err != nil {
			logln(stderr, "source-lanes: repository unavailable")
			return 1
		}
	}
	root, err := os.OpenRoot(repo)
	if err != nil {
		logln(stderr, "source-lanes: repository unavailable")
		return 1
	}
	// This confinement handle is read-only: no data is flushed by Close.
	// Read/validation failures already use the caller's existing result channel;
	// teardown is not an additional source-consistency or proof-authority gate.
	defer func() { _ = root.Close() }()
	cohortRaw, err := readSourceInput(root, sourceLaneCohortPath, 64<<20)
	var cohort sourceLaneCohort
	if err != nil || decodeSourceJSON(cohortRaw, &cohort) != nil || decodeStrictJSON(cohortRaw, &cohort) != nil || validateSourceLaneCohort(cohort) != nil {
		logln(stderr, "source-lanes: cohort_anchor_invalid")
		return 1
	}
	annotationRaw, annotationErr := readSourceInput(root, sourceLaneAnnotationsPath, 64<<20)
	var annotations struct {
		SchemaVersion int                        `json:"schema_version"`
		Annotations   []sourceSemanticAnnotation `json:"annotations"`
	}
	annotationInvalid := annotationErr != nil || decodeSourceJSON(annotationRaw, &annotations) != nil || decodeStrictJSON(annotationRaw, &annotations) != nil || annotations.SchemaVersion != 1 || annotations.Annotations == nil
	if annotationInvalid {
		annotations.Annotations = nil
	}
	result, err := buildSourceLaneManifest(ctx, repo, cohort, annotations.Annotations)
	if err != nil {
		logln(stderr, "source-lanes: cohort_anchor_invalid")
		return 1
	}
	if ctx.Err() != nil {
		logln(stderr, "source-lanes: canceled")
		return 1
	}
	result.Inputs = append(result.Inputs, sourceArtifactPin{Path: sourceLaneCohortPath, SHA256: sourceBytesHash(cohortRaw), Bytes: int64(len(cohortRaw))})
	if annotationErr == nil {
		result.Inputs = append(result.Inputs, sourceArtifactPin{Path: sourceLaneAnnotationsPath, SHA256: sourceBytesHash(annotationRaw), Bytes: int64(len(annotationRaw))})
	}
	if annotationInvalid {
		result.Diagnostics = append(result.Diagnostics, sourceLaneDiagnostic{Lanes: sourceLaneNames(), Stage: "classification", Code: "annotation_document_invalid", Pointer: sourceLaneAnnotationsPath, Owner: cohort.CohortID, Severity: "error"})
		result.Validation.Status = "invalid"
		result.Validation.Errors++
	}
	sort.Slice(result.Inputs, func(i, j int) bool { return result.Inputs[i].Path < result.Inputs[j].Path })
	raw, err := json.Marshal(result)
	if err != nil {
		logln(stderr, "source-lanes: manifest_encoding_invalid")
		return 1
	}
	raw = append(raw, '\n')
	if check {
		candidateRaw, err := readSourceInput(root, manifest, 512<<20)
		var candidate sourceLaneManifest
		if err != nil || decodeSourceJSON(candidateRaw, &candidate) != nil || decodeStrictJSON(candidateRaw, &candidate) != nil {
			logln(stderr, "source-lanes: manifest_input_invalid")
			return 1
		}
		findings := validateSourceLaneManifest(candidate, result)
		for _, finding := range findings {
			logln(stderr, "source-lanes:", finding.Code, finding.Pointer)
		}
		if len(findings) > 0 || !bytes.Equal(raw, candidateRaw) {
			if len(findings) == 0 {
				logln(stderr, "source-lanes: manifest_bytes_drift")
			}
			return 1
		}
		if result.Validation.Status != "valid" {
			logln(stderr, "source-lanes: retained_inputs_invalid")
			return 1
		}
		return 0
	}
	if code := writeSourceLaneBytes(stdout, stderr, raw); code != 0 {
		return code
	}
	if result.Validation.Status != "valid" {
		return 1
	}
	return 0
}

func writeSourceLaneBytes(stdout, stderr io.Writer, raw []byte) int {
	n, err := stdout.Write(raw)
	if err != nil || n != len(raw) {
		_, _ = fmt.Fprintln(stderr, "source-lanes: output_write_failed")
		return 1
	}
	return 0
}
