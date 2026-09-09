package engine

import (
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

// identity uses ReadFile; optional presence uses sub.Open then file.Stat.
// Separating these interfaces observes the actual Go subFS fallback, rather
// than assuming that outer StatFS survives fs.Sub.
type bundleStatFault167 struct {
	fs.FS
	target                string
	fault                 error
	faults, identityReads int
	phase                 string
}

func (f *bundleStatFault167) ReadFile(name string) ([]byte, error) {
	f.identityReads++
	return fs.ReadFile(f.FS, name)
}
func (f *bundleStatFault167) Open(name string) (fs.File, error) {
	file, err := f.FS.Open(name)
	if name != f.target || f.fault == nil {
		return file, err
	}
	if err != nil {
		f.faults++
		f.phase = "open"
		return nil, f.fault
	}
	return &bundleFaultFile167{File: file, owner: f}, nil
}

type bundleFaultFile167 struct {
	fs.File
	owner *bundleStatFault167
}

func (f *bundleFaultFile167) Stat() (fs.FileInfo, error) {
	f.owner.faults++
	f.owner.phase = "file.Stat"
	return nil, f.owner.fault
}
func TestBundleOptionalStatFailure167(t *testing.T) {
	for _, file := range []string{"changefeed.json", "polling_watermark.json", "sync_transport.json", "database.json", "streams.json", "writes.json", "operations.json", "cli_surface.json", "rate_limits.json"} {
		for _, kind := range []string{"permission", "compound_absence"} {
			t.Run(file+"/"+kind, func(t *testing.T) {
				cause := error(fs.ErrPermission)
				if kind == "compound_absence" {
					cause = errors.Join(fs.ErrNotExist, fs.ErrPermission)
				}
				files := &bundleStatFault167{FS: fullValidBundleFS("acme"), target: "acme/" + file, fault: cause}
				expectedPhase := "file.Stat"
				if _, e := fs.Stat(files.FS, files.target); errors.Is(e, fs.ErrNotExist) {
					expectedPhase = "open"
				}
				control := &bundleStatFault167{FS: fullValidBundleFS("acme"), target: files.target}
				if _, e := Load(control, "acme"); e != nil || control.faults != 0 || control.identityReads == 0 {
					t.Fatalf("healthy topology control failed: %v", e)
				}
				_, err := Load(files, "acme")
				if files.faults != 1 || files.phase != expectedPhase || files.identityReads == 0 {
					t.Fatalf("did not reach selected optional Open/file.Stat: %d error=%v", files.faults, err)
				}
				if !publicBundleMatches167(err, file, "/", "bundle_file_unreadable", "execution file is unreadable") || !errors.Is(err, fs.ErrPermission) {
					t.Errorf("optional I/O failure swallowed or mislabeled: %v", err)
				}
				if kind == "compound_absence" && !errors.Is(err, fs.ErrNotExist) {
					t.Error("lost joined absence cause")
				}
			})
		}
	}
}

func TestBundleSchemaReferenceDiagnostic168(t *testing.T) {
	for _, ref := range []string{"schemas/private-sentinel-167.json", "https://private-sentinel-167.invalid/schema", "../private-sentinel-167"} {
		t.Run(ref, func(t *testing.T) {
			files := fullValidBundleFS("acme")
			files["acme/streams.json"] = &fstest.MapFile{Data: []byte(strings.Replace(validStreams, "schemas/widgets.json", ref, 1))}
			_, err := Load(files, "acme")
			if !publicBundleMatches167(err, "streams.json", "/streams/0/schema", "schema_reference_unreadable", "referenced schema could not be read") {
				t.Errorf("unselected reference became public file identity: %v", err)
			}
			var pathErr *fs.PathError
			if !errors.As(err, &pathErr) {
				t.Errorf("lost original file cause: %v", err)
			}
		})
	}
}

func TestBundlePureJoinedOptionalAbsence169(t *testing.T) {
	for _, file := range []string{"changefeed.json", "polling_watermark.json", "sync_transport.json", "database.json"} {
		for _, nested := range []bool{false, true} {
			name := file + "/single"
			if nested {
				name = file + "/nested"
			}
			t.Run(name, func(t *testing.T) {
				original := fullValidBundleFS("acme")
				if _, e := fs.Stat(original, "acme/"+file); !errors.Is(e, fs.ErrNotExist) {
					t.Fatalf("fixture optional file must actually be absent: %v", e)
				}
				control, e := Load(original, "acme")
				if e != nil {
					t.Fatal(e)
				}
				cause := errors.Join(&fs.PathError{Op: "open", Path: "private-sentinel-167", Err: fs.ErrNotExist}, nil)
				if nested {
					cause = errors.Join(errors.Join(cause, nil), nil)
				}
				files := &bundleStatFault167{FS: original, target: "acme/" + file, fault: cause}
				got, e := Load(files, "acme")
				if files.faults != 1 || files.phase != "open" || files.identityReads == 0 {
					t.Fatalf("did not reach optional loader Open: faults=%d phase=%s", files.faults, files.phase)
				}
				if e != nil || got.Identity != control.Identity || got.Name != "acme" {
					t.Errorf("pure joined absence changed healthy execution: %v", e)
				}
			})
		}
	}
}

type absenceCycle169 struct{}

func (e *absenceCycle169) Error() string { return "absence cycle" }
func (e *absenceCycle169) Unwrap() error { return e }
func TestBundleAbsenceGraphBounds169(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want bool
	}{
		{"single_join", errors.Join(fs.ErrNotExist, nil), true},
		{"nested_join", errors.Join(errors.Join(fs.ErrNotExist, nil), nil), true},
		{"compound_permission", errors.Join(fs.ErrNotExist, fs.ErrPermission), false},
		{"two_absence_causes", errors.Join(fs.ErrNotExist, fs.ErrNotExist), false},
		{"nil", nil, false},
		{"cycle", &absenceCycle169{}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := bundlePureAbsence(tc.err); got != tc.want {
				t.Errorf("pure absence=%t want %t", got, tc.want)
			}
		})
	}
}
