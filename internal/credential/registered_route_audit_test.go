package credential

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRegisteredCredentialRoutesDoNotTrimSecretBytes(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	tests := []struct {
		path     string
		isSecret func(ast.Expr) bool
	}{
		{
			path: filepath.Join("internal", "connectors", "native", "safetyculture", "safetyculture.go"),
			isSecret: func(expr ast.Expr) bool {
				if ident, ok := expr.(*ast.Ident); ok && ident.Name == "token" {
					return true
				}
				call, ok := expr.(*ast.CallExpr)
				if !ok || len(call.Args) != 2 {
					return false
				}
				name, nameOK := call.Fun.(*ast.Ident)
				key, keyOK := call.Args[1].(*ast.BasicLit)
				return nameOK && keyOK && name.Name == "secret" && key.Value == `"access_token"`
			},
		},
		{
			path: filepath.Join("internal", "connectors", "native", "canny", "canny.go"),
			isSecret: func(expr ast.Expr) bool {
				if ident, ok := expr.(*ast.Ident); ok && ident.Name == "secret" {
					return true
				}
				call, ok := expr.(*ast.CallExpr)
				if !ok {
					return false
				}
				name, ok := call.Fun.(*ast.Ident)
				return ok && name.Name == "cannySecret"
			},
		},
		{
			path: filepath.Join("internal", "connectors", "hooks", "amazon-ads", "hooks.go"),
			isSecret: func(expr ast.Expr) bool {
				if ident, ok := expr.(*ast.Ident); ok {
					return ident.Name == "clientID" || ident.Name == "clientSecret" || ident.Name == "refreshToken"
				}
				if selector, ok := expr.(*ast.SelectorExpr); ok {
					owner, ok := selector.X.(*ast.Ident)
					return ok && owner.Name == "out" && selector.Sel.Name == "AccessToken"
				}
				index, ok := expr.(*ast.IndexExpr)
				if !ok {
					return false
				}
				secrets, ok := index.X.(*ast.SelectorExpr)
				if !ok {
					return false
				}
				owner, ownerOK := secrets.X.(*ast.Ident)
				return ownerOK && owner.Name == "cfg" && secrets.Sel.Name == "Secrets"
			},
		},
	}
	for _, tt := range tests {
		t.Run(filepath.Base(filepath.Dir(tt.path)), func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(root, tt.path), nil, 0)
			if err != nil {
				t.Fatalf("parse route source: %v", err)
			}
			ast.Inspect(file, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok || !isStringsTrimCall(call) || len(call.Args) != 1 || !tt.isSecret(call.Args[0]) {
					return true
				}
				t.Fatalf("registered route trims credential bytes: %s", tt.path)
				return false
			})
		})
	}
}

func isStringsTrimCall(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	owner, ok := selector.X.(*ast.Ident)
	return ok && owner.Name == "strings" && strings.HasPrefix(selector.Sel.Name, "Trim")
}
