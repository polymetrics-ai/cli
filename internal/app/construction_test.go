package app

import (
	"errors"
	"io/fs"
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestOpenWithRegistryBuildsRegistryBeforeProjectIO(t *testing.T) {
	called := false
	_, err := openWithRegistry(t.TempDir(), false, func() (*connectors.Registry, error) {
		called = true
		return connectors.NewRegistry(), nil
	})
	if err == nil {
		t.Fatal("openWithRegistry accepted a project without .polymetrics")
	}
	if !called {
		t.Fatal("openWithRegistry performed project I/O before building its registry")
	}
}

func TestOpenWithInvalidRegistryStopsBeforeProjectStat(t *testing.T) {
	want := errors.New("invalid generated executor selection")
	constructorCalls := 0
	statCalls := 0

	_, err := openWithRegistryWithStat(t.TempDir(), false, func() (*connectors.Registry, error) {
		constructorCalls++
		return nil, want
	}, func(string) (fs.FileInfo, error) {
		statCalls++
		return nil, fs.ErrNotExist
	})
	if !errors.Is(err, want) {
		t.Fatalf("openWithRegistryWithStat() error = %v, want invalid registry selection", err)
	}
	if constructorCalls != 1 {
		t.Fatalf("registry construction calls = %d, want 1", constructorCalls)
	}
	if statCalls != 0 {
		t.Fatalf("project stat calls = %d, want 0 after invalid construction", statCalls)
	}
}
