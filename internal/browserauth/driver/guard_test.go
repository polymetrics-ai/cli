package driver_test

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"polymetrics.ai/internal/browserauth/driver"
)

// TestNoTypingSurface is the mechanical enforcement of the captain's
// non-negotiable boundary: "the tool never collects or stores a password."
// It is not enough that Session's documented interface omits a typing
// method — nothing stops a future edit from quietly adding one directly
// against *rod.Page and wiring it up. This test greps every non-test .go
// file in this package for the symbols a password-entry code path would
// need (Input, Keyboard, SetFiles, Type — rod's own names for its
// form-filling API) and fails the build if any appear, so the boundary
// survives a careless future edit rather than depending on a reviewer
// noticing.
func TestNoTypingSurface(t *testing.T) {
	forbidden := regexp.MustCompile(`\b(Input|Keyboard|SetFiles|MustInput|MustSetFiles|MustType)\b`)
	// "Type" alone is too broad (matches Go's own type keyword/identifiers
	// like "ContentType"); rod's actual page-filling method is `Element.Input`
	// / `Page.Keyboard` / `Element.SetFiles`, all covered above. `MustType`
	// (rod's low-level key-event-sequence typer) is covered explicitly since
	// "Type" alone would over-match.

	err := filepath.WalkDir(".", func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, line := range strings.Split(string(raw), "\n") {
			// Skip only whole-line comments — the shape this package's own doc
			// comment uses when it names the forbidden symbols by design.
			// Cutting at the first "//" anywhere would also blind the scan to
			// anything following a URL literal on a line of real code.
			if strings.HasPrefix(strings.TrimSpace(line), "//") {
				continue
			}
			if loc := forbidden.FindString(line); loc != "" {
				t.Errorf("%s: forbidden password-entry-shaped symbol %q — Session must expose only Navigate/WaitFor/GetCookies/Close", path, loc)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk internal/browserauth/driver: %v", err)
	}
}

// TestSessionInterfaceIsMinimal locks the Session interface's method set to
// exactly the four documented methods, via a compile-time-shaped check: any
// addition to the interface changes this test's expected method list, which
// forces a reviewer to consciously widen the boundary rather than have it
// drift silently.
func TestSessionInterfaceIsMinimal(t *testing.T) {
	want := []string{"Close", "GetCookies", "Navigate", "WaitFor"} // reflect returns interface methods sorted by name
	typ := reflect.TypeOf((*driver.Session)(nil)).Elem()
	got := make([]string, typ.NumMethod())
	for i := 0; i < typ.NumMethod(); i++ {
		got[i] = typ.Method(i).Name
	}
	sort.Strings(got)
	if len(got) != len(want) {
		t.Fatalf("Session method set = %v, want %v", got, want)
	}
	for i, name := range want {
		if got[i] != name {
			t.Fatalf("Session method set = %v, want %v", got, want)
		}
	}
}
