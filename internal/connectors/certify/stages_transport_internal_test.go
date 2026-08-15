package certify

import (
	"context"
	"strings"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/synctransport"
)

func TestCertificationDeclaredTransportPairResolvesAndExecutes(t *testing.T) {
	var proof certificationTransportPairProof
	rc := &runContext{
		ctx:  context.Background(),
		opts: Options{Connector: "github", Full: true},
		root: t.TempDir(),
		transportPairProbe: func(ctx context.Context, root string) (certificationTransportPairProof, error) {
			var err error
			proof, err = app.ProbeDeclaredTransportForCertification(ctx, root, "github")
			return proof, err
		},
	}
	report := Report{Passed: true}
	if err := stageDeclaredTransportPair(rc, &report); err != nil {
		t.Fatalf("stageDeclaredTransportPair() = %v", err)
	}
	report.Passed = allStagesPassed(report.Stages)

	if !report.Passed || ExitCodeFor(report) != 0 {
		t.Fatalf("registered declared pair report = %+v, want terminal pass", report)
	}
	if len(report.Stages) != 1 || report.Stages[0].Name != declaredTransportPairStage || !report.Stages[0].Passed {
		t.Fatalf("registered declared pair stage = %+v, want named pass", report.Stages)
	}
	if proof.SourceReference != "issue_label_source" || proof.DestinationReference != "issue_label_destination" {
		t.Fatalf("resolved references = %q -> %q", proof.SourceReference, proof.DestinationReference)
	}
	if proof.ProviderReads < 2 || proof.ProviderWrites != 1 {
		t.Fatalf("provider calls = reads:%d writes:%d, want source/readback and one apply", proof.ProviderReads, proof.ProviderWrites)
	}
	if proof.RecordsRead != 1 || proof.RecordsLoaded != 1 || !proof.CheckpointCommitted || proof.WarehouseManifests != 1 || proof.WarehouseParquet != 1 {
		t.Fatalf("durable proof = %+v, want one staged/reopened record and committed checkpoint", proof)
	}
}

func TestCertificationDeclaredTransportPairFailsWhenRegistrationIsMissing(t *testing.T) {
	var proof certificationTransportPairProof
	rc := &runContext{
		ctx:  context.Background(),
		opts: Options{Connector: "github", Full: true},
		root: t.TempDir(),
		transportPairProbe: func(context.Context, string) (certificationTransportPairProof, error) {
			bundle, err := engine.Load(defs.FS, "github")
			if err != nil {
				return certificationTransportPairProof{}, err
			}
			registry := connectors.NewEmptyRegistry()
			registry.Register(engine.New(bundle, nil))
			verifier, err := synctransport.NewDefinitionConformanceVerifier(nil)
			if err != nil {
				return certificationTransportPairProof{}, err
			}
			transports := synctransport.NewRegistry(verifier)
			// Passing no factories deliberately removes registration while
			// preserving the real bundle declaration and registration code.
			err = synctransport.RegisterDeclaredTransports(registry, transports, nil)
			return proof, err
		},
	}
	report := Report{Passed: true}
	if err := stageDeclaredTransportPair(rc, &report); err != nil {
		t.Fatalf("stageDeclaredTransportPair() = %v", err)
	}
	report.Passed = allStagesPassed(report.Stages)

	if report.Passed || ExitCodeFor(report) != 2 {
		t.Fatalf("unregistered declared pair report = %+v, want terminal certification failure", report)
	}
	if proof.ProviderReads != 0 || proof.ProviderWrites != 0 {
		t.Fatalf("provider calls = reads:%d writes:%d, want zero before registration failure", proof.ProviderReads, proof.ProviderWrites)
	}
	if len(report.Stages) != 1 || report.Stages[0].Name != declaredTransportPairStage || report.Stages[0].Passed {
		t.Fatalf("unregistered declared pair stage = %+v, want named failure", report.Stages)
	}
	if !strings.Contains(report.Stages[0].Error, `declared source transport executor "issue_label_source" is not registered`) {
		t.Fatalf("unregistered declared pair error = %q", report.Stages[0].Error)
	}
}

func TestCertificationDeclaredTransportPairSkipsConnectorWithoutAdapter(t *testing.T) {
	rc := &runContext{
		ctx:  context.Background(),
		opts: Options{Connector: "sample", Full: true},
		root: t.TempDir(),
		transportPairProbe: func(context.Context, string) (certificationTransportPairProof, error) {
			return certificationTransportPairProof{}, nil
		},
	}
	report := Report{Passed: true}
	if err := stageDeclaredTransportPair(rc, &report); err != nil {
		t.Fatalf("stageDeclaredTransportPair() = %v", err)
	}
	if len(report.Stages) != 1 || report.Stages[0].Passed || !strings.HasPrefix(report.Stages[0].Error, "skipped:") {
		t.Fatalf("unadapted connector stage = %+v, want explicit skip", report.Stages)
	}
}
