package registryset

import "testing"

func TestProductionRegistryExcludesStagedConnectors(t *testing.T) {
	r := New()
	for _, name := range []string{"sample", "file", "warehouse", "outbox", "github", "stripe", "source-github", "source-stripe"} {
		if _, ok := r.Get(name); !ok {
			t.Fatalf("production registry missing %q", name)
		}
	}
	for _, staged := range []string{"100ms", "adjust", "freshdesk"} {
		if _, ok := r.Get(staged); ok {
			t.Fatalf("production registry included staged connector %q", staged)
		}
	}
}

func TestStagedRegistryIncludesSelfRegisteredConnectors(t *testing.T) {
	r := NewStaged()
	for _, staged := range []string{"100ms", "adjust", "freshdesk"} {
		if _, ok := r.Get(staged); !ok {
			t.Fatalf("staged registry missing %q", staged)
		}
	}
}
