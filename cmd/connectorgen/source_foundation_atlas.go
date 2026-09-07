package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"strings"
)

type sourceFoundationSymbol struct {
	File string `json:"file"`
	Name string `json:"name"`
}

type sourceFoundationAtlasTest struct {
	Package string `json:"package"`
	File    string `json:"file"`
	Name    string `json:"name"`
}

// Shared Atlas wire declaration: both the authoring binary and publication
// contract tests consume these same guarantee-to-proof relationships.
type vNextPublicationGuaranteeProof struct {
	Guarantees []string `json:"guarantees"`
	Positive   string   `json:"positive"`
	Negative   string   `json:"negative"`
}

type sourceFoundationAtlasEntry struct {
	raw     json.RawMessage
	pointer string
	ID      string `json:"id"`
	Layer   string `json:"layer"`
	Status  string `json:"status"`
	Kind    string `json:"kind"`
	Summary string `json:"summary"`
	Owner   struct {
		PrimaryPackage     string                   `json:"primary_package"`
		SupportingPackages []string                 `json:"supporting_packages"`
		Files              []string                 `json:"files"`
		Symbols            []sourceFoundationSymbol `json:"symbols"`
	} `json:"owner"`
	SupportedContracts map[string][]string `json:"supported_contracts"`
	Selection          struct {
		Mechanisms      []string `json:"mechanisms"`
		DefinitionFiles []string `json:"definition_files"`
		Selectors       []string `json:"selectors"`
		Notes           string   `json:"notes"`
	} `json:"selection"`
	Constraints      []string `json:"constraints"`
	NonGoals         []string `json:"non_goals"`
	ConsumerExamples []struct {
		Name  string   `json:"name"`
		Files []string `json:"files"`
		Notes string   `json:"notes"`
	} `json:"consumer_examples"`
	ProofTests                 []sourceFoundationAtlasTest      `json:"proof_tests"`
	PublicationGuaranteeProofs []vNextPublicationGuaranteeProof `json:"publication_guarantee_proofs,omitempty"`
	KnownGapRefs               []struct {
		ID        string `json:"id"`
		Reference string `json:"reference"`
		Notes     string `json:"notes"`
	} `json:"known_gap_refs"`
	UpdateConditions []string `json:"update_conditions"`
	Supersedes       []string `json:"supersedes"`
}

type sourceFoundationAtlas struct {
	pin     sourceArtifactPin
	entries map[string]sourceFoundationAtlasEntry
}

type sourceFoundationDeclaration struct {
	file  *ast.File
	lines *token.FileSet
}

func validateSourceFoundationProofAtlas(cache *sourceProofFileCache, parsed map[string]sourceFoundationDeclaration, r sourceFoundationProofRecord, entry sourceFoundationAtlasEntry) error {
	if entry.Status != "available" {
		return fmt.Errorf("foundation atlas entry not available")
	}
	value, err := sourceJSONPointer(entry.raw, r.Contract.Pointer)
	if err != nil {
		return fmt.Errorf("foundation contract occurrence unavailable")
	}
	canonical, err := canonicalSourceJSON(value)
	if err != nil || sourceBytesHash(canonical) != r.Contract.ValueSHA256 {
		return fmt.Errorf("foundation contract occurrence changed")
	}
	for _, owner := range r.OwnerSymbols {
		if !slices.Contains(entry.Owner.Files, owner.File) || !slices.Contains(entry.Owner.Symbols, owner) {
			return fmt.Errorf("foundation owner not declared by atlas entry")
		}
		declaration, err := sourceFoundationPinnedDeclaration(cache, parsed, r.Inputs, owner.File, "code")
		if err != nil {
			return err
		}
		if !declaresSymbol(declaration.file, owner.Name) {
			return fmt.Errorf("foundation owner declaration unavailable")
		}
	}
	wantTest := sourceFoundationAtlasTest{File: r.Test.File, Name: r.Test.Symbol, Package: r.Test.Package}
	if !slices.Contains(entry.ProofTests, wantTest) {
		return fmt.Errorf("foundation test not registered by atlas entry")
	}
	declaration, err := sourceFoundationPinnedDeclaration(cache, parsed, r.Inputs, r.Test.File, "test")
	if err != nil {
		return err
	}
	if !declaresTestFunction(declaration.file, r.Test.Symbol) {
		return fmt.Errorf("foundation test declaration unavailable")
	}
	for _, node := range declaration.file.Decls {
		fn, ok := node.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || fn.Name.Name != r.Test.Symbol {
			continue
		}
		if r.Assertion.StartLine < declaration.lines.Position(fn.Pos()).Line || r.Assertion.EndLine > declaration.lines.Position(fn.End()).Line {
			return fmt.Errorf("foundation assertion outside selected test declaration")
		}
		return nil
	}
	return fmt.Errorf("foundation assertion declaration unavailable")
}

func sourceFoundationPinnedDeclaration(cache *sourceProofFileCache, parsed map[string]sourceFoundationDeclaration, inputs []sourceFoundationProofInput, name, role string) (sourceFoundationDeclaration, error) {
	var pin sourceFoundationProofInput
	for _, input := range inputs {
		if input.Path == name && input.Role == role {
			pin = input
			break
		}
	}
	if pin.Path == "" {
		return sourceFoundationDeclaration{}, fmt.Errorf("foundation declaration input missing")
	}
	raw, file := cache.get(name, 4<<20, false)
	if file.code != "" || file.hash != pin.SHA256 || file.size != pin.Bytes {
		return sourceFoundationDeclaration{}, fmt.Errorf("foundation declaration input changed")
	}
	if declaration, ok := parsed[name]; ok {
		return declaration, nil
	}
	lines := token.NewFileSet()
	// An explicit reader prevents a nil cache-hit slice from being interpreted
	// by go/parser as permission to open a process-relative filename.
	fileAST, err := parser.ParseFile(lines, name, bytes.NewReader(raw), 0)
	if err != nil {
		return sourceFoundationDeclaration{}, fmt.Errorf("foundation declaration parse: %w", err)
	}
	declaration := sourceFoundationDeclaration{file: fileAST, lines: lines}
	parsed[name] = declaration
	return declaration, nil
}

// The richer CP13 view is separate from CP12's owner-only projection. Missing
// examples remain reconciliation work and cannot erase a valid Atlas owner.
func readSourceFoundationAtlas(cache *sourceProofFileCache) (sourceFoundationAtlas, error) {
	result := sourceFoundationAtlas{entries: map[string]sourceFoundationAtlasEntry{}}
	raw, file := cache.get(sourceDemandAtlasPath, 64<<20, false)
	if file.code != "" {
		return result, fmt.Errorf("foundation atlas: %s", file.code)
	}
	result.pin = sourceArtifactPin{Path: sourceDemandAtlasPath, SHA256: file.hash, Bytes: file.size}
	var document struct {
		Schema          string `json:"$schema"`
		SchemaVersion   int    `json:"schema_version"`
		CatalogRevision int    `json:"catalog_revision"`
		Authority       struct {
			AuthoringOnly                bool   `json:"authoring_only"`
			RuntimeInput                 bool   `json:"runtime_input"`
			ProviderFactAuthority        string `json:"provider_fact_authority"`
			ConnectorSpecificSelection   string `json:"connector_specific_selection"`
			SharedRuntimeConnectorBranch bool   `json:"shared_runtime_connector_name_branching"`
		} `json:"authority"`
		Foundations []json.RawMessage `json:"foundations"`
	}
	if decodeSourceJSON(raw, &document) != nil || decodeStrictJSON(raw, &document) != nil || document.Schema != "./catalog.schema.json" || document.SchemaVersion != 1 || document.CatalogRevision < 1 || !document.Authority.AuthoringOnly || document.Authority.RuntimeInput || document.Authority.ProviderFactAuthority != "source_lock_vnext" || document.Authority.ConnectorSpecificSelection != "closed_definition_reference" || document.Authority.SharedRuntimeConnectorBranch || len(document.Foundations) == 0 {
		return result, fmt.Errorf("foundation atlas: invalid catalogue")
	}
	for index, rawEntry := range document.Foundations {
		var entry sourceFoundationAtlasEntry
		if decodeStrictJSON(rawEntry, &entry) != nil || !validSourceID(entry.ID) || entry.Owner.PrimaryPackage == "" || len(entry.Owner.Files) == 0 || len(entry.Owner.Symbols) == 0 || len(entry.ProofTests) == 0 {
			return result, fmt.Errorf("foundation atlas: invalid entry")
		}
		if _, duplicate := result.entries[entry.ID]; duplicate {
			return result, fmt.Errorf("foundation atlas: duplicate identity")
		}
		for name := range entry.SupportedContracts {
			switch name {
			case "lanes", "protocols", "source_forms", "request_response_shapes", "sync_modes", "guarantees", "publication_guarantees":
			default:
				return result, fmt.Errorf("foundation atlas: unknown supported contract")
			}
		}
		entry.raw = rawEntry
		entry.pointer = fmt.Sprintf("/foundations/%d", index)
		result.entries[entry.ID] = entry
	}
	return result, nil
}

// The authoring Atlas and its regression use the same declaration resolver.
// Declaration existence alone is never a behavioral proof.
func declaresTestFunction(file *ast.File, name string) bool {
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Recv == nil && function.Name.Name == name && hasTestingTParameter(function) {
			return true
		}
	}
	return false
}

func declaresSymbol(file *ast.File, name string) bool {
	receiverName, declaredName := splitAtlasSymbol(name)
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			if declaration.Name.Name != declaredName {
				continue
			}
			if receiverName == "" || receiverTypeName(declaration.Recv) == receiverName {
				return true
			}
		case *ast.GenDecl:
			if receiverName != "" {
				continue
			}
			for _, spec := range declaration.Specs {
				switch spec := spec.(type) {
				case *ast.TypeSpec:
					if spec.Name.Name == declaredName {
						return true
					}
				case *ast.ValueSpec:
					for _, declared := range spec.Names {
						if declared.Name == declaredName {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func splitAtlasSymbol(name string) (receiverName, declaredName string) {
	receiverEnd := strings.LastIndex(name, ".")
	if receiverEnd < 0 {
		return "", name
	}
	receiverName = strings.Trim(name[:receiverEnd], "()*")
	return receiverName, name[receiverEnd+1:]
}

func receiverTypeName(receivers *ast.FieldList) string {
	if receivers == nil || len(receivers.List) != 1 {
		return ""
	}
	switch typ := receivers.List[0].Type.(type) {
	case *ast.Ident:
		return typ.Name
	case *ast.StarExpr:
		if identifier, ok := typ.X.(*ast.Ident); ok {
			return identifier.Name
		}
	}
	return ""
}

func hasTestingTParameter(function *ast.FuncDecl) bool {
	if function.Type.Params == nil || len(function.Type.Params.List) != 1 {
		return false
	}
	pointer, ok := function.Type.Params.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	selector, ok := pointer.X.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "T" {
		return false
	}
	packageName, ok := selector.X.(*ast.Ident)
	return ok && packageName.Name == "testing"
}
