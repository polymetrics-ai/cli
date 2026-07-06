package coordination

import "testing"

func TestDragonflyCloseIsNilSafe(t *testing.T) {
	if err := (*Dragonfly)(nil).Close(); err != nil {
		t.Fatalf("nil Close() = %v", err)
	}
	if err := (&Dragonfly{}).Close(); err != nil {
		t.Fatalf("zero-value Close() = %v", err)
	}
}

func TestOpenDragonflyCreatesClient(t *testing.T) {
	d := OpenDragonfly("127.0.0.1:6379")
	if d == nil || d.client == nil {
		t.Fatal("OpenDragonfly returned nil client")
	}
	if err := d.Close(); err != nil {
		t.Fatalf("Close(): %v", err)
	}
}
