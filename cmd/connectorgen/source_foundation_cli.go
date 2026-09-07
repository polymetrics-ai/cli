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

const sourceFoundationRegisterPath = "data/connector-canon/batch1-foundation-demand-register.json"

type sourceFoundationCommandOptions struct {
	repo, cohort, manifest, assessments, checkPath string
	check, help                                    bool
}

func parseSourceFoundationCommand(args []string) (sourceFoundationCommandOptions, error) {
	o := sourceFoundationCommandOptions{cohort: sourceLaneCohortPath, manifest: sourceLaneManifestPath,
		assessments: sourceFoundationAssessmentsPath, checkPath: sourceFoundationRegisterPath}
	seen := map[string]bool{}
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if arg == "-h" {
			arg = "--help"
		}
		if seen[arg] {
			return o, fmt.Errorf("duplicate argument")
		}
		seen[arg] = true
		switch arg {
		case "--help":
			o.help = true
		case "--check":
			o.check = true
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				o.checkPath = args[i]
			}
		case "--repo", "--cohort", "--manifest", "--assessments":
			if i+1 >= len(args) || args[i+1] == "" || strings.HasPrefix(args[i+1], "-") {
				return o, fmt.Errorf("missing argument value")
			}
			i++
			switch arg {
			case "--repo":
				o.repo = args[i]
			case "--cohort":
				o.cohort = args[i]
			case "--manifest":
				o.manifest = args[i]
			case "--assessments":
				o.assessments = args[i]
			}
		default:
			return o, fmt.Errorf("unknown argument")
		}
	}
	for _, name := range []string{o.cohort, o.manifest, o.assessments, o.checkPath} {
		if !sourceLaneRelativePath(name) {
			return o, fmt.Errorf("input paths must be confined relative paths")
		}
	}
	return o, nil
}

func runSourceDemandsContext(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	return runSourceDemandsPolicy(ctx, args, stdout, stderr, sourceFoundationBaselinePin())
}

// Policy is source-owned; there is no CLI baseline/proof/Atlas trust override.
// Hermetic tests supply their independently retained fixture baseline here.
func runSourceDemandsPolicy(ctx context.Context, args []string, stdout, stderr io.Writer, baseline sourceArtifactPin) int {
	return runSourceDemandsObserved(ctx, args, stdout, stderr, baseline, nil)
}

func runSourceDemandsObserved(ctx context.Context, args []string, stdout, stderr io.Writer, baseline sourceArtifactPin, observer func(sourceProofReadEvent)) int {
	o, err := parseSourceFoundationCommand(args)
	if err != nil {
		logln(stderr, "source-demands:", err)
		return 2
	}
	if o.help {
		return writeSourceFoundationBytes(stdout, stderr, []byte("usage: connectorgen source-demands [--repo <dir>] [--cohort <relative-path>] [--manifest <relative-source-manifest>] [--assessments <relative-path>] [--check [<relative-register>]]\n\nGenerate the authoring-only foundation demand register as complete JSON on stdout.\n--manifest reads a retained source/lane report and verifies it against current inputs.\n--check validates a saved register without writing files; its default is data/connector-canon/batch1-foundation-demand-register.json.\nAtlas, proof review and independent baseline authority are fixed by the source-owned policy.\nConsistency is not executable capability, receiver approval or checkpoint acceptance.\n"))
	}
	if o.repo == "" {
		o.repo, err = repoRoot()
		if err != nil {
			logln(stderr, "source-demands: repository unavailable")
			return 1
		}
	}
	root, err := os.OpenRoot(o.repo)
	if err != nil {
		logln(stderr, "source-demands: repository unavailable")
		return 1
	}
	defer func() { _ = root.Close() }()
	cache := sourceProofFileCache{ctx: ctx, root: root,
		limits: sourceProofFileLimits{UniqueBytes: 512 << 20, Files: 65536},
		files:  map[string]*sourceProofFile{}, afterRead: observer}
	// Plan command byte roles before proof preloading can observe any overlap.
	for _, role := range []struct {
		name string
		cap  int64
	}{{o.cohort, 64 << 20}, {sourceLaneAnnotationsPath, 64 << 20}, {o.manifest, 512 << 20}} {
		if err := cache.plan(role.name, role.cap, true); err != nil {
			logln(stderr, "source-demands:", err)
			return 1
		}
	}
	if o.check {
		if err := cache.plan(o.checkPath, 512<<20, true); err != nil {
			logln(stderr, "source-demands:", err)
			return 1
		}
	}
	catalog := reviewedSourceFoundationProofs()
	atlas, document, documentObservation, err := prepareSourceFoundationProofs(&cache, catalog)
	if err != nil {
		logln(stderr, "source-demands:", err)
		return 1
	}
	preloaded, err := observePreparedSourceFoundationProofs(&cache, atlas, document, documentObservation, catalog)
	if err != nil {
		logln(stderr, "source-demands:", err)
		return 1
	}
	missingProofs := map[string]bool{}
	for _, name := range cache.order {
		if cache.files[name].code == "missing" {
			missingProofs[name] = true
		}
	}
	inputs, err := loadSourceFoundationCommandInputs(ctx, o, baseline, &cache)
	if err != nil {
		logln(stderr, "source-demands:", err)
		return 1
	}
	missingExamples := map[string]bool{}
	for _, entry := range inputs.assessments.atlas.entries {
		for _, example := range entry.ConsumerExamples {
			for _, name := range example.Files {
				_, file := cache.get(name, 64<<20, false)
				if file.code == "missing" {
					missingExamples[name] = true
				} else if file.code != "" {
					logln(stderr, "source-demands: example input invalid")
					return 1
				}
			}
		}
	}
	for _, name := range cache.order {
		file := cache.files[name]
		if file.code == "" {
			inputs.commandPins = append(inputs.commandPins, sourceArtifactPin{Path: name, SHA256: file.hash, Bytes: file.size})
		}
	}
	sort.Slice(inputs.commandPins, func(i, j int) bool { return inputs.commandPins[i].Path < inputs.commandPins[j].Path })
	register, err := buildSourceFoundationRegister(ctx, o.repo, inputs)
	if err != nil {
		logln(stderr, "source-demands:", err)
		return 1
	}
	expectedProofs := []sourceFoundationObservedProof{}
	for _, proof := range preloaded.records {
		expectedProofs = append(expectedProofs, sourceFoundationObservedProof{Record: proof.record, Status: proof.status, Issues: proof.issues})
	}
	wanted, _ := json.Marshal(struct {
		Document sourceFoundationProofDocumentObservation
		Proofs   []sourceFoundationObservedProof
	}{preloaded.document, expectedProofs})
	actual, _ := json.Marshal(struct {
		Document sourceFoundationProofDocumentObservation
		Proofs   []sourceFoundationObservedProof
	}{register.ProofDocument, register.Proofs})
	if !bytes.Equal(wanted, actual) {
		logln(stderr, "source-demands: proof observations changed between readers")
		return 1
	}
	raw, err := json.Marshal(register)
	if err != nil {
		logln(stderr, "source-demands: register encoding failed")
		return 1
	}
	raw = append(raw, '\n')
	if err := validateSourceFoundationRegister(ctx, o.repo, raw, inputs); err != nil {
		logln(stderr, "source-demands:", err)
		return 1
	}
	if o.check {
		candidate, file := cache.getContent(o.checkPath, 512<<20)
		if file.code != "" || validateSourceFoundationRegister(ctx, o.repo, candidate, inputs) != nil || !bytes.Equal(candidate, raw) {
			logln(stderr, "source-demands: saved register invalid or changed")
			return 1
		}
	}
	// Spend the shared physical-read allowance on final content/identity
	// verification before either successful check return or the single write.
	cache.finalize()
	if ctx.Err() != nil {
		logln(stderr, "source-demands: canceled")
		return 1
	}
	for _, name := range cache.order {
		file := cache.files[name]
		if (missingExamples[name] || missingProofs[name]) && file.code == "missing" {
			if _, code := sourceProofFileInfo(root, name); code == "missing" {
				continue
			}
		}
		if file.code != "" {
			logln(stderr, "source-demands: input changed before result", name)
			return 1
		}
	}
	if o.check {
		return 0
	}
	return writeSourceFoundationBytes(stdout, stderr, raw)
}

func loadSourceFoundationCommandInputs(ctx context.Context, o sourceFoundationCommandOptions, baseline sourceArtifactPin, cache *sourceProofFileCache) (sourceFoundationDemandInputs, error) {
	var result sourceFoundationDemandInputs
	cohortRaw, cohortFile := cache.getContent(o.cohort, 64<<20)
	var cohort sourceLaneCohort
	if cohortFile.code != "" || decodeSourceJSON(cohortRaw, &cohort) != nil || decodeStrictJSON(cohortRaw, &cohort) != nil || validateSourceLaneCohort(cohort) != nil {
		return result, fmt.Errorf("cohort anchor invalid")
	}
	annotationRaw, annotationFile := cache.getContent(sourceLaneAnnotationsPath, 64<<20)
	var annotations struct {
		SchemaVersion int                        `json:"schema_version"`
		Annotations   []sourceSemanticAnnotation `json:"annotations"`
	}
	if annotationFile.code != "" || decodeSourceJSON(annotationRaw, &annotations) != nil || decodeStrictJSON(annotationRaw, &annotations) != nil || annotations.SchemaVersion != 1 || annotations.Annotations == nil {
		return result, fmt.Errorf("source annotation document invalid")
	}
	universe, err := buildSourceFoundationUniverse(ctx, o.repo, cohort, annotations.Annotations)
	if err != nil {
		return result, err
	}
	universe.manifest.Inputs = append(universe.manifest.Inputs,
		sourceArtifactPin{Path: o.cohort, SHA256: sourceBytesHash(cohortRaw), Bytes: int64(len(cohortRaw))},
		sourceArtifactPin{Path: sourceLaneAnnotationsPath, SHA256: sourceBytesHash(annotationRaw), Bytes: int64(len(annotationRaw))})
	sort.Slice(universe.manifest.Inputs, func(i, j int) bool { return universe.manifest.Inputs[i].Path < universe.manifest.Inputs[j].Path })
	manifestRaw, manifestFile := cache.getContent(o.manifest, 512<<20)
	var manifest sourceLaneManifest
	if manifestFile.code != "" || decodeSourceJSON(manifestRaw, &manifest) != nil || decodeStrictJSON(manifestRaw, &manifest) != nil {
		return result, fmt.Errorf("source manifest invalid")
	}
	if findings := validateSourceLaneManifest(manifest, universe.manifest); len(findings) != 0 {
		return result, fmt.Errorf("source manifest differs from current retained inputs")
	}
	for _, pin := range universe.manifest.Inputs {
		if err := observeSourceFoundationCommandPin(cache, pin, 64<<20); err != nil {
			return result, err
		}
	}
	// Source/lane reconciliation also consumed canonical locks and execution
	// declarations. Retain their actual observed bytes, including HTTP/schema
	// siblings used by a fit, rather than pinning only a selected target file.
	bindingPins := map[string]sourceArtifactPin{}
	seenBindings := map[*sourceLaneBindingInputs]bool{}
	for _, row := range universe.manifest.SourceOperations {
		bindings := row.Facts.bindings
		if bindings == nil || seenBindings[bindings] {
			continue
		}
		seenBindings[bindings] = true
		for _, files := range []map[string][]byte{bindings.Authoring, bindings.Artifacts} {
			for name, raw := range files {
				pin := sourceArtifactPin{Path: name, SHA256: sourceBytesHash(raw), Bytes: int64(len(raw))}
				if previous, exists := bindingPins[name]; exists && previous != pin {
					return result, fmt.Errorf("source binding input observations disagree")
				}
				bindingPins[name] = pin
			}
		}
	}
	names := make([]string, 0, len(bindingPins))
	for name := range bindingPins {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if err := observeSourceFoundationCommandPin(cache, bindingPins[name], 64<<20); err != nil {
			return result, err
		}
	}
	_, assessmentFile := cache.get(o.assessments, 64<<20, false)
	if assessmentFile.code != "" {
		return result, fmt.Errorf("assessment input unavailable")
	}
	if err := observeSourceFoundationCommandPin(cache, baseline, 4<<20); err != nil {
		return result, err
	}
	result, err = observeSourceFoundationDemandInputs(ctx, o.repo, o.assessments, universe, baseline)
	if err != nil {
		return result, err
	}
	if result.assessments.pin.SHA256 != assessmentFile.hash || result.assessments.pin.Bytes != assessmentFile.size {
		return sourceFoundationDemandInputs{}, fmt.Errorf("assessment input changed between readers")
	}
	return result, nil
}

func writeSourceFoundationBytes(stdout, stderr io.Writer, raw []byte) int {
	n, err := stdout.Write(raw)
	if err != nil || n != len(raw) {
		logln(stderr, "source-demands: output_write_failed")
		return 1
	}
	return 0
}

func observeSourceFoundationCommandPin(cache *sourceProofFileCache, pin sourceArtifactPin, limit int64) error {
	if err := cache.plan(pin.Path, limit, false); err != nil {
		return err
	}
	_, file := cache.get(pin.Path, limit, false)
	if file.code != "" || file.hash != pin.SHA256 || file.size != pin.Bytes {
		return fmt.Errorf("foundation command input pin mismatch: %s", pin.Path)
	}
	return nil
}
