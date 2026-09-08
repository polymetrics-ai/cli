package engine

import (
	"context"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

// fakeHooks implements every hook interface so a single fake can be
// dispatched through each of the 5 hook-point contracts.
type fakeHooks struct {
	name string

	authenticator connsdk.Authenticator
	authErr       error
	authCalls     int

	mapRecordOut      connsdk.Record
	mapRecordKeep     bool
	mapRecordErr      error
	mapRecordCalls    int
	lastMapRecordRaw  connsdk.Record
	lastMapRecordProj connsdk.Record
	lastMapRecordName string

	readStreamHandled bool
	readStreamErr     error
	readStreamCalls   int

	executeWriteHandled bool
	executeWriteErr     error
	executeWriteCalls   int

	checkHandled bool
	checkErr     error
	checkCalls   int
}

func (f *fakeHooks) ConnectorName() string { return f.name }

func (f *fakeHooks) Authenticator(_ context.Context, _ connectors.RuntimeConfig, _ AuthSpec) (connsdk.Authenticator, error) {
	f.authCalls++
	return f.authenticator, f.authErr
}

func (f *fakeHooks) MapRecord(stream string, raw, projected connsdk.Record) (connsdk.Record, bool, error) {
	f.mapRecordCalls++
	f.lastMapRecordName = stream
	f.lastMapRecordRaw = raw
	f.lastMapRecordProj = projected
	return f.mapRecordOut, f.mapRecordKeep, f.mapRecordErr
}

func (f *fakeHooks) ReadStream(_ context.Context, _ StreamSpec, _ connectors.ReadRequest, _ *Runtime, _ func(connectors.Record) error) (bool, error) {
	f.readStreamCalls++
	return f.readStreamHandled, f.readStreamErr
}

func (f *fakeHooks) ExecuteWrite(_ context.Context, _ WriteAction, _ connectors.Record, _ *Runtime) (bool, []*connsdk.Response, error) {
	f.executeWriteCalls++
	return f.executeWriteHandled, nil, f.executeWriteErr
}

func (f *fakeHooks) Check(_ context.Context, _ connectors.RuntimeConfig, _ *Runtime) (bool, error) {
	f.checkCalls++
	return f.checkHandled, f.checkErr
}

// compile-time assertions that fakeHooks satisfies every hook interface.
var (
	_ Hooks      = (*fakeHooks)(nil)
	_ AuthHook   = (*fakeHooks)(nil)
	_ RecordHook = (*fakeHooks)(nil)
	_ StreamHook = (*fakeHooks)(nil)
	_ WriteHook  = (*fakeHooks)(nil)
	_ CheckHook  = (*fakeHooks)(nil)
)


func TestAuthHookDispatch(t *testing.T) {
	wantAuth := connsdk.Bearer("tok")
	fh := &fakeHooks{authenticator: wantAuth}

	auth, err := fh.Authenticator(context.Background(), connectors.RuntimeConfig{}, AuthSpec{Mode: "custom", Hook: "acme"})
	if err != nil {
		t.Fatalf("Authenticator() error = %v", err)
	}
	if auth != wantAuth {
		t.Fatalf("Authenticator() returned a different authenticator than the fake provided")
	}
	if fh.authCalls != 1 {
		t.Fatalf("authCalls = %d, want 1", fh.authCalls)
	}
}

func TestRecordHookDispatch(t *testing.T) {
	fh := &fakeHooks{mapRecordOut: connsdk.Record{"a": 1}, mapRecordKeep: true}
	raw := connsdk.Record{"a": 1, "b": 2}
	proj := connsdk.Record{"a": 1}

	out, keep, err := fh.MapRecord("widgets", raw, proj)
	if err != nil {
		t.Fatalf("MapRecord() error = %v", err)
	}
	if !keep {
		t.Fatalf("MapRecord() keep = false, want true")
	}
	if out["a"] != 1 {
		t.Fatalf("MapRecord() out = %#v", out)
	}
	if fh.lastMapRecordName != "widgets" {
		t.Fatalf("MapRecord() stream = %q, want widgets", fh.lastMapRecordName)
	}
}

func TestStreamHookDispatch(t *testing.T) {
	fh := &fakeHooks{readStreamHandled: true}
	handled, err := fh.ReadStream(context.Background(), StreamSpec{Name: "widgets"}, connectors.ReadRequest{}, &Runtime{}, func(connectors.Record) error { return nil })
	if err != nil {
		t.Fatalf("ReadStream() error = %v", err)
	}
	if !handled {
		t.Fatalf("ReadStream() handled = false, want true")
	}
	if fh.readStreamCalls != 1 {
		t.Fatalf("readStreamCalls = %d, want 1", fh.readStreamCalls)
	}
}

func TestWriteHookDispatch(t *testing.T) {
	fh := &fakeHooks{executeWriteHandled: true}
	handled, _, err := fh.ExecuteWrite(context.Background(), WriteAction{Name: "create_widget"}, connectors.Record{}, &Runtime{})
	if err != nil {
		t.Fatalf("ExecuteWrite() error = %v", err)
	}
	if !handled {
		t.Fatalf("ExecuteWrite() handled = false, want true")
	}
	if fh.executeWriteCalls != 1 {
		t.Fatalf("executeWriteCalls = %d, want 1", fh.executeWriteCalls)
	}
}

func TestCheckHookDispatch(t *testing.T) {
	fh := &fakeHooks{checkHandled: true}
	handled, err := fh.Check(context.Background(), connectors.RuntimeConfig{}, &Runtime{})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !handled {
		t.Fatalf("Check() handled = false, want true")
	}
	if fh.checkCalls != 1 {
		t.Fatalf("checkCalls = %d, want 1", fh.checkCalls)
	}
}
