package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
)

type sourceFoundationExampleFile struct {
	Path   string `json:"path"`
	State  string `json:"state"`
	SHA256 string `json:"sha256"`
	Bytes  int64  `json:"bytes"`
}

// An example reference is not a provider operation identity. Exact declared
// adoption joins are separately derived from admitted requirement bindings.
type sourceFoundationExample struct {
	AtlasID     string                        `json:"atlas_id"`
	Pointer     string                        `json:"pointer"`
	ValueSHA256 string                        `json:"value_sha256"`
	Name        string                        `json:"name"`
	Owner       string                        `json:"owner"`
	Notes       string                        `json:"notes"`
	Files       []sourceFoundationExampleFile `json:"files"`
	Status      string                        `json:"status"`
	Reason      string                        `json:"reason"`
}

func observeSourceFoundationExamples(ctx context.Context, repo string, expected sourceArtifactPin) (result []sourceFoundationExample, resultErr error) {
	defer func() {
		if err := ctx.Err(); err != nil {
			result, resultErr = nil, errors.Join(resultErr, err)
		}
	}()
	root, err := os.OpenRoot(repo)
	if err != nil {
		return nil, fmt.Errorf("foundation examples root: %w", err)
	}
	defer func() { _ = root.Close() }()
	cache := sourceProofFileCache{ctx: ctx, root: root,
		limits: sourceProofFileLimits{UniqueBytes: 512 << 20, Files: 65536}, files: map[string]*sourceProofFile{}}
	atlas, err := readSourceFoundationAtlas(&cache)
	if err != nil {
		return nil, fmt.Errorf("foundation examples Atlas unavailable: %w", err)
	}
	if atlas.pin != expected {
		return nil, fmt.Errorf("foundation examples Atlas changed")
	}
	result = []sourceFoundationExample{}
	for _, entry := range atlas.entries {
		for index, example := range entry.ConsumerExamples {
			pointer := fmt.Sprintf("/consumer_examples/%d", index)
			raw, err := sourceJSONPointer(entry.raw, pointer)
			if err != nil {
				return nil, fmt.Errorf("foundation example occurrence: %w", err)
			}
			canonical, err := canonicalSourceJSON(raw)
			if err != nil {
				return nil, fmt.Errorf("foundation example canonical identity: %w", err)
			}
			row := sourceFoundationExample{AtlasID: entry.ID, Pointer: entry.pointer + pointer, ValueSHA256: sourceBytesHash(canonical),
				Name: example.Name, Owner: entry.Owner.PrimaryPackage, Notes: example.Notes, Files: []sourceFoundationExampleFile{},
				Status: "example_unresolved", Reason: "reference_observed_without_exact_source_and_selector_join"}
			for _, path := range example.Files {
				_, file := cache.get(path, 64<<20, false)
				state := "present_regular"
				if file.code == "missing" {
					state = "missing"
					row.Reason = "referenced_file_missing"
				} else if file.code != "" {
					return nil, fmt.Errorf("foundation example input invalid: %s", path)
				}
				row.Files = append(row.Files, sourceFoundationExampleFile{Path: path, State: state, SHA256: file.hash, Bytes: file.size})
			}
			result = append(result, row)
		}
	}
	cache.finalize()
	for _, name := range cache.order {
		file := cache.files[name]
		if file.code == "missing" {
			// A missing reference is visible evidence too; creation during the
			// observation cannot silently retain yesterday's missing status.
			if _, current := sourceProofFileInfo(root, name); current != "missing" {
				return nil, fmt.Errorf("foundation missing example changed during observation")
			}
		} else if file.code != "" {
			return nil, fmt.Errorf("foundation example changed during observation")
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].AtlasID != result[j].AtlasID {
			return result[i].AtlasID < result[j].AtlasID
		}
		return result[i].Pointer < result[j].Pointer
	})
	return result, nil
}
