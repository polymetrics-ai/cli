package outbox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestProductionGitHubWritesCannotBypassOutboxExecutor(t *testing.T) {
	t.Parallel()
	forbiddenSelectors := map[string]struct{}{
		"SyncDecisionComment": {},
		"SyncQuestionComment": {},
		"NewCLIClient":        {},
		"NewGrant":            {},
		"PutGrant":            {},
		"ClaimEffect":         {},
		"MarkEffectSent":      {},
		"MarkEffectFailed":    {},
	}
	root := filepath.Clean(filepath.Join("..", ".."))
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		slashPath := filepath.ToSlash(path)
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") ||
			strings.Contains(slashPath, "/internal/outbox/") || strings.Contains(slashPath, "/integration/") {
			return nil
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			switch typed := node.(type) {
			case *ast.SelectorExpr:
				if _, forbidden := forbiddenSelectors[typed.Sel.Name]; forbidden {
					t.Errorf("%s uses forbidden external-write/grant selector %s", path, typed.Sel.Name)
				}
			case *ast.FuncDecl:
				if _, forbidden := forbiddenSelectors[typed.Name.Name]; forbidden {
					t.Errorf("%s declares forbidden external-write/grant function %s", path, typed.Name.Name)
				}
			case *ast.BasicLit:
				value, err := strconv.Unquote(typed.Value)
				if err == nil && (value == "POST" || value == "PATCH" || value == "PUT" || value == "DELETE") &&
					!strings.HasSuffix(filepath.ToSlash(path), "/internal/github/outbox_transport.go") {
					t.Errorf("%s contains write method %s outside the outbox GitHub transport", path, value)
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestControllerAuthorizationCallSiteIsCentralized(t *testing.T) {
	t.Parallel()
	root := filepath.Clean(filepath.Join("..", ".."))
	var callSites []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		slashPath := filepath.ToSlash(path)
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") ||
			strings.Contains(slashPath, "/internal/outbox/") || strings.Contains(slashPath, "/integration/") {
			return nil
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(parsed, func(node ast.Node) bool {
			selector, ok := node.(*ast.SelectorExpr)
			if ok && (selector.Sel.Name == "Authorize" || selector.Sel.Name == "Reauthorize") {
				callSites = append(callSites, filepath.ToSlash(path))
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range callSites {
		if !strings.HasSuffix(path, "/cmd/shepherd/effects.go") {
			t.Errorf("authorization call outside protected controller admission: %s", path)
		}
	}
}
