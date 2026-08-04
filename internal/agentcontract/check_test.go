package agentcontract

import (
	"context"
	"path/filepath"
	"testing"
)

func TestReferencedGSDCommandsResolve(t *testing.T) {
	script := filepath.Join("..", "..", "scripts", "gsd")
	commands := []string{"discuss-phase", "programming-loop"}

	if err := CheckGSDCommands(context.Background(), script, commands); err != nil {
		t.Fatalf("referenced GSD commands must resolve: %v", err)
	}
}

func TestCheckProjectionRejectsDivergence(t *testing.T) {
	want := []byte("canonical generated block\n")
	got := []byte("diverged generated block\n")

	if err := CheckProjection(want, got); err == nil {
		t.Fatal("CheckProjection accepted a diverged projection")
	}
}
