package gitlab

import "testing"

const gitLabDescriptorCorrectionProvenancePath = "sources/gitlab-operation-descriptor-correction-provenance.json"

func TestGitLabDescriptorCorrectionProvenanceIsDeclared(t *testing.T) {
	correction := loadGitLabObject(t, gitLabDescriptorCorrectionProvenancePath)
	if stringAt(correction, "connector") != "gitlab" {
		t.Fatalf("descriptor correction connector = %q, want gitlab", stringAt(correction, "connector"))
	}
}
