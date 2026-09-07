package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"sort"
)

type sourceFoundationProofIssue struct {
	Code string `json:"code"`
	Path string `json:"path"`
}

type sourceFoundationProofDocumentObservation struct {
	Path   string                       `json:"path"`
	Status string                       `json:"status"`
	Issues []sourceFoundationProofIssue `json:"issues"`
}

type sourceFoundationProofBatch struct {
	document sourceFoundationProofDocumentObservation
	records  []sourceFoundationProofObservation
	pins     []sourceArtifactPin
}

// Preparation is shared by the standalone reader and the command preloader.
// All record roles are known before any selected execution input is observed.
func prepareSourceFoundationProofs(cache *sourceProofFileCache, catalog sourceFoundationProofCatalog) (sourceFoundationAtlas, sourceFoundationProofDocument, sourceFoundationProofDocumentObservation, error) {
	docObservation := sourceFoundationProofDocumentObservation{Path: sourceFoundationProofPath, Status: "present", Issues: []sourceFoundationProofIssue{}}
	var document sourceFoundationProofDocument
	if err := cache.plan(sourceFoundationProofPath, 128<<20, true); err != nil {
		return sourceFoundationAtlas{}, document, docObservation, err
	}
	atlas, err := readSourceFoundationAtlas(cache)
	if err != nil {
		return atlas, document, docObservation, err
	}
	raw, file := cache.getContent(sourceFoundationProofPath, 128<<20)
	if file.code == "missing" {
		docObservation.Status = "proof_unavailable"
		docObservation.Issues = append(docObservation.Issues, sourceFoundationProofIssue{Code: "proof_document_missing", Path: sourceFoundationProofPath})
		return atlas, document, docObservation, nil
	}
	if file.code != "" {
		return atlas, document, docObservation, fmt.Errorf("foundation proof document: %s", file.code)
	}
	document, err = decodeSourceFoundationProofDocument(cache.ctx, raw, len(catalog.Reviews))
	if err != nil {
		return atlas, document, docObservation, err
	}
	if document.Atlas.Path != sourceDemandAtlasPath || !sourceProofDigest(document.Atlas.SHA256) || document.Atlas.Bytes < 0 {
		return atlas, document, docObservation, fmt.Errorf("foundation proof atlas pin invalid")
	}
	if document.Atlas != atlas.pin {
		docObservation.Issues = append(docObservation.Issues, sourceFoundationProofIssue{Code: "proof_atlas_pin_stale", Path: sourceDemandAtlasPath})
	}
	ids := map[string]bool{}
	type selection struct{ atlas, contract, file, test string }
	selections := map[selection]bool{}
	for _, record := range document.Records {
		key := selection{record.AtlasID, record.Contract.Pointer, record.Test.File, record.Test.Selected}
		if !sourceFoundationProofRecordShape(record) || ids[record.ID] || selections[key] {
			return atlas, document, docObservation, fmt.Errorf("foundation proof record invalid or duplicate")
		}
		ids[record.ID], selections[key] = true, true
		entry, exists := atlas.entries[record.AtlasID]
		if !exists {
			return atlas, document, docObservation, fmt.Errorf("foundation proof atlas entry unavailable")
		}
		if err := validateSourceFoundationProofMembership(record, entry); err != nil {
			return atlas, document, docObservation, err
		}
		content := map[string]bool{"go.mod": true, record.Test.File: true}
		for _, owner := range record.OwnerSymbols {
			content[owner.File] = true
		}
		for _, input := range record.Inputs {
			if err := cache.plan(input.Path, 4<<20, content[input.Path]); err != nil {
				return atlas, document, docObservation, err
			}
		}
		if err := cache.plan(record.Capture.Path, 64<<20, true); err != nil {
			return atlas, document, docObservation, err
		}
		if err := cache.plan(record.Output.Path, 4<<20, true); err != nil {
			return atlas, document, docObservation, err
		}
	}
	return atlas, document, docObservation, nil
}

func sourceFoundationInputCurrent(cache *sourceProofFileCache, inputs []sourceFoundationProofInput, name string) bool {
	for _, input := range inputs {
		if input.Path == name {
			return sourceFoundationPinCurrent(cache, sourceArtifactPin{Path: name, SHA256: input.SHA256, Bytes: input.Bytes})
		}
	}
	return false
}

func sourceFoundationPinCurrent(cache *sourceProofFileCache, pin sourceArtifactPin) bool {
	file := cache.files[pin.Path]
	return file != nil && file.code == "" && file.hash == pin.SHA256 && file.size == pin.Bytes
}

func observePreparedSourceFoundationProofs(cache *sourceProofFileCache, atlas sourceFoundationAtlas, document sourceFoundationProofDocument, observed sourceFoundationProofDocumentObservation, catalog sourceFoundationProofCatalog) (sourceFoundationProofBatch, error) {
	batch := sourceFoundationProofBatch{document: observed, records: []sourceFoundationProofObservation{}, pins: []sourceArtifactPin{}}
	reviews := map[string]sourceFoundationProofReview{}
	for _, review := range catalog.Reviews {
		if _, exists := reviews[review.ID]; exists || !validSourceID(review.ID) || !sourceProofDigest(review.RecordSHA256) || !sourceProofDigest(review.InputClosureSHA256) || !sourceProofDigest(review.AssertionSHA256) {
			return batch, fmt.Errorf("foundation review catalogue invalid")
		}
		reviews[review.ID] = review
	}
	paths := map[string]bool{sourceDemandAtlasPath: true, sourceFoundationProofPath: true}
	for _, record := range document.Records {
		paths[record.Capture.Path], paths[record.Output.Path] = true, true
		for _, input := range record.Inputs {
			paths[input.Path] = true
		}
	}
	names := make([]string, 0, len(paths))
	for name := range paths {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		_, file := cache.get(name, cache.plans[name].cap, false)
		if file.code != "" && file.code != "missing" {
			return batch, fmt.Errorf("foundation proof input %s: %s", name, file.code)
		}
	}
	parsed := map[string]sourceFoundationDeclaration{}
	executions := sourceFoundationExecutions{cache: cache, captures: map[string]sourceFoundationCapture{}}
	records := append([]sourceFoundationProofRecord{}, document.Records...)
	sort.Slice(records, func(i, j int) bool { return records[i].ID < records[j].ID })
	for _, record := range records {
		issues := append([]sourceFoundationProofIssue{}, observed.Issues...)
		pins := []sourceArtifactPin{record.Capture, record.Output}
		for _, input := range record.Inputs {
			pins = append(pins, sourceArtifactPin{Path: input.Path, SHA256: input.SHA256, Bytes: input.Bytes})
		}
		for _, pin := range pins {
			file := cache.files[pin.Path]
			if file.code == "missing" {
				issues = append(issues, sourceFoundationProofIssue{Code: "proof_input_missing", Path: pin.Path})
			} else if !sourceFoundationPinCurrent(cache, pin) {
				issues = append(issues, sourceFoundationProofIssue{Code: "proof_input_stale", Path: pin.Path})
			}
		}
		sort.Slice(issues, func(i, j int) bool {
			if issues[i].Code != issues[j].Code {
				return issues[i].Code < issues[j].Code
			}
			return issues[i].Path < issues[j].Path
		})
		issues = slices.Compact(issues)
		entry := atlas.entries[record.AtlasID]
		if err := validateSourceFoundationProofAtlas(cache, parsed, record, entry); err != nil {
			return batch, err
		}
		if err := executions.validate(record); err != nil {
			return batch, err
		}
		status := "unreviewed"
		if review, exists := reviews[record.ID]; exists && sourceFoundationReviewMatches(record, review) {
			status = "current"
		}
		if len(issues) > 0 {
			status = "proof_stale"
		}
		for _, issue := range issues {
			if issue.Code == "proof_input_missing" {
				status = "proof_unavailable"
			}
		}
		batch.records = append(batch.records, sourceFoundationProofObservation{record: record, atlas: entry, status: status, issues: issues})
	}
	for _, name := range names {
		file := cache.files[name]
		if file.code == "" {
			batch.pins = append(batch.pins, sourceArtifactPin{Path: name, SHA256: file.hash, Bytes: file.size})
		}
	}
	sort.Slice(batch.pins, func(i, j int) bool { return batch.pins[i].Path < batch.pins[j].Path })
	return batch, nil
}

func finalizeSourceFoundationProofs(cache *sourceProofFileCache) error {
	cache.finalize()
	if err := cache.ctx.Err(); err != nil {
		return err
	}
	for _, name := range cache.order {
		file := cache.files[name]
		if file.code == "missing" {
			if _, code := sourceProofFileInfo(cache.root, name); code == "missing" {
				continue
			}
		}
		if file.code != "" {
			return fmt.Errorf("foundation proof observation changed: %s", name)
		}
	}
	return nil
}

func readSourceFoundationProofBatch(ctx context.Context, repo string, catalog sourceFoundationProofCatalog, observer func(sourceProofReadEvent)) (batch sourceFoundationProofBatch, err error) {
	defer func() {
		if cause := ctx.Err(); cause != nil {
			batch = sourceFoundationProofBatch{}
			err = errors.Join(err, cause)
		}
	}()
	if err = ctx.Err(); err != nil {
		return batch, err
	}
	root, err := os.OpenRoot(repo)
	if err != nil {
		return batch, err
	}
	defer func() { _ = root.Close() }()
	cache := sourceProofFileCache{ctx: ctx, root: root, limits: sourceProofFileLimits{UniqueBytes: 512 << 20, Files: 65536}, files: map[string]*sourceProofFile{}, afterRead: observer}
	atlas, document, observed, err := prepareSourceFoundationProofs(&cache, catalog)
	if err != nil {
		return batch, err
	}
	batch, err = observePreparedSourceFoundationProofs(&cache, atlas, document, observed, catalog)
	if err != nil {
		return sourceFoundationProofBatch{}, err
	}
	if err = finalizeSourceFoundationProofs(&cache); err != nil {
		return sourceFoundationProofBatch{}, err
	}
	return batch, nil
}
