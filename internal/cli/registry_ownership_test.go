package cli

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func TestRunRootRouterReusesPreflightRegistryForNormalHelpAndReverse(t *testing.T) {
	for _, test := range []struct {
		name     string
		intent   string
		args     []string
		want     string
		wantCode int
		approval bool
	}{
		{
			name:     "normal command",
			intent:   "etl",
			args:     []string{"widgets", "list", "--state", "open"},
			want:     "missing --credential",
			wantCode: 1,
		},
		{
			name:     "help",
			intent:   "etl",
			args:     []string{"widgets", "list", "--help"},
			want:     "Input ownership fixture",
			wantCode: 0,
		},
		{
			name:     "reverse approval path",
			intent:   "reverse_etl",
			args:     []string{"widgets", "write", "--plan", "missing", "--approval-token-stdin"},
			want:     `reverse plan "missing" not found`,
			wantCode: 1,
			approval: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			if err := app.InitProject(root); err != nil {
				t.Fatal(err)
			}
			registry, loads := ownershipRegistry(t, test.intent)
			args := append([]string{"input-ownership"}, test.args...)
			var approvalReader *bytes.Buffer
			if test.approval {
				approvalReader = bytes.NewBufferString("approval-token\n")
			}
			args = append(args, "--root", root)
			var stdout, stderr bytes.Buffer
			code := runWithPreflightRegistryAndApprovalReader(args, &stdout, &stderr, registry, approvalReader)
			combined := stdout.String() + stderr.String()
			if code != test.wantCode {
				t.Fatalf("Run() code = %d, want %d; output=%q", code, test.wantCode, combined)
			}
			if !strings.Contains(combined, test.want) {
				t.Fatalf("Run() output = %q, want %q", combined, test.want)
			}
			if strings.Contains(combined, `connector "input-ownership" not found`) {
				t.Fatalf("Run() rebuilt a different registry: %q", combined)
			}
			if got := loads.Load(); got != 1 {
				t.Fatalf("selected connector constructions = %d, want 1 across root/preflight/App path", got)
			}
		})
	}
}

func ownershipRegistry(t *testing.T, intent string) (*connectors.Registry, *atomic.Int32) {
	t.Helper()
	var loads atomic.Int32
	registry, err := connectors.NewLazyRegistry([]connectors.Metadata{{
		Name:            "input-ownership",
		DisplayName:     "Input ownership",
		IntegrationType: "api",
	}}, func(_ context.Context, name string) (connectors.Connector, error) {
		loads.Add(1)
		minimum := connectors.ExactNumber("2")
		commandPath := "widgets list"
		method := http.MethodGet
		stream := "widgets"
		write := ""
		writes := []engine.WriteAction(nil)
		if intent == "reverse_etl" {
			commandPath = "widgets write"
			method = http.MethodPost
			stream = ""
			write = "write_widgets"
			writes = []engine.WriteAction{{
				Name:         write,
				Kind:         "create",
				Method:       http.MethodPost,
				Path:         "/widgets",
				RecordSchema: []byte(`{"type":"object"}`),
			}}
		}
		bundle := engine.Bundle{
			Name: name,
			Metadata: engine.Metadata{
				Name:            name,
				DisplayName:     "Input ownership",
				IntegrationType: "api",
				Capabilities:    engine.Capabilities{Read: true, Write: intent == "reverse_etl"},
			},
			Streams: []engine.StreamSpec{{Name: "widgets", Method: http.MethodGet, Path: "/widgets"}},
			Writes:  writes,
			CLISurface: &engine.CLISurface{
				Tagline: "Input ownership fixture",
				Usage:   "pm input-ownership widgets <command>",
				Commands: []engine.CLICommand{{
					Path:         commandPath,
					Summary:      "Input ownership fixture",
					Intent:       intent,
					Availability: "implemented",
					Stream:       stream,
					Write:        write,
					APISurface:   []engine.CLISurfaceEndpointRef{{Method: method, Path: "/widgets"}},
					Flags: []engine.CLIFlag{{
						Name: "state", Type: "enum", Values: []string{"open", "closed"}, MapsTo: "query.state", Required: intent != "reverse_etl",
					}, {
						Name: "batch", Type: "integer", MapsTo: "query.batch", Minimum: &minimum,
					}},
				}},
			},
		}
		return engine.New(bundle, nil), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return registry, &loads
}
