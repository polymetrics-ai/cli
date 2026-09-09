package main

import (
	"testing"
)

func TestVNextExistingFlagGrammar185(t *testing.T) {
	for _, tc := range []struct {
		name  string
		valid bool
	}{
		{"completed-at.after", true}, {"due_on.before", true}, {"ProviderOption", true}, {"limit", true},
		{" completed-at.after", false}, {"a..b", false}, {"a/b", false}, {"a=b", false}, {"a\n", false}, {"", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, err := canonicalizeVNextSourceLock(flagNamespaceLock182(t, []string{tc.name}))
			if !tc.valid {
				if err == nil {
					t.Fatal("unsafe flag admitted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want := tc.name
			if want == "limit" {
				want = "provider-limit"
			}
			if d.Graph.Operations[0].Commands[0].Spec.Flags[0].Name != want {
				t.Fatal("safe spelling changed")
			}
			if _, err := admitVNextCanonicalDescriptor(d, vNextSemanticAdmissionInput{}); err != nil {
				t.Fatal(err)
			}
		})
	}
}
