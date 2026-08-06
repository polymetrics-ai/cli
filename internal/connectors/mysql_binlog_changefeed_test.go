package connectors

import "testing"

func TestChangefeedDescriptorAcceptsBinlogReplication(t *testing.T) {
	descriptor := ChangefeedDescriptor{
		Status:    ChangefeedStatusUnsupported,
		Mechanism: ChangefeedMechanism("binlog_replication"),
		Source: ChangefeedSource{
			ArtifactURL:     "https://dev.mysql.com/doc/refman/8.4/en/replication-options-binary-log.html",
			ArtifactVersion: "8.4",
			RetrievedAt:     "2026-08-06",
		},
		Reason: "test declaration only",
	}
	if err := descriptor.Validate(); err != nil {
		t.Fatalf("binlog replication declaration rejected: %v", err)
	}
}
