package main

import (
	"go/ast"
	"strings"
)

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
