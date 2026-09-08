package connectors

import "testing"

func TestSharedCommandControlTuple182(t *testing.T) {
	for _, tc := range []struct {
		name, intent, stream string
		flag                 CommandSurfaceFlag
		want                 bool
	}{
		{"twenty", "etl", "companies", CommandSurfaceFlag{Name: "limit", MapsTo: "limit", Type: "integer"}, true},
		{"direct_stream", "direct_read", "companies", CommandSurfaceFlag{Name: "limit", MapsTo: "limit", Type: "integer"}, true},
		{"provider_query", "etl", "companies", CommandSurfaceFlag{Name: "limit", MapsTo: "query.limit", Type: "integer"}, false},
		{"provider_body", "direct_read", "", CommandSurfaceFlag{Name: "limit", MapsTo: "body.limit", Type: "integer"}, false},
		{"required", "etl", "companies", CommandSurfaceFlag{Name: "limit", MapsTo: "limit", Type: "integer", Required: true}, false},
		{"repeatable", "etl", "companies", CommandSurfaceFlag{Name: "limit", MapsTo: "limit", Type: "integer", Repeatable: true}, false},
		{"env_only", "etl", "companies", CommandSurfaceFlag{Name: "limit", MapsTo: "limit", Type: "integer", EnvOnly: true}, false},
		{"wrong_type", "etl", "companies", CommandSurfaceFlag{Name: "limit", MapsTo: "limit", Type: "string"}, false},
		{"write", "reverse_etl", "companies", CommandSurfaceFlag{Name: "limit", MapsTo: "limit", Type: "integer"}, false},
		{"no_stream", "etl", "", CommandSurfaceFlag{Name: "limit", MapsTo: "limit", Type: "integer"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsSharedCommandControl(tc.intent, tc.stream, tc.flag); got != tc.want {
				t.Fatalf("compatibility=%v want%v", got, tc.want)
			}
		})
	}
}

func TestControlConsumerExceptions182(t *testing.T) {
	for _, name := range []string{"page", "page-cursor"} {
		if ConnectorControlFlag(name) != PMNavigationControl {
			t.Fatal("page ownership changed")
		}
	}
	for _, name := range []string{"root", "json", "help"} {
		if ConnectorControlFlag(name) != PMGlobalControl {
			t.Fatal("bootstrap ownership changed")
		}
	}
	for _, name := range []string{"config", "plan", "limit", "credential", "from-env", "file-name"} {
		if ConnectorControlFlag(name) != PMConnectorControl {
			t.Fatal("lifecycle ownership changed")
		}
	}
	for _, name := range []string{"provider-plan", "provider-limit", "ordinary"} {
		if ConnectorControlFlag(name) != ProviderCommandFlag {
			t.Fatal("provider alias consumed as PM control")
		}
	}
}
