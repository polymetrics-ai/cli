package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

const sourceDemandAtlasPath = "docs/connector-canon/foundations/catalog.json"

// This authoring attribution records an existing demand, not receiver approval.
// Its decision owners remain pending. Annotation input cannot create entries.
type sourceFoundationDemandAuthority struct {
	Key                                 sourceOperationKey
	SourcePath, SourceSHA256            string
	Protocol, Method, Path, OperationID string
	Gap, AtlasID, AtlasOwner            string
	DecisionOwners                      []string
	Citations                           []sourceFactRef
}

func sourceFoundationDemandCatalog() []sourceFoundationDemandAuthority {
	const prefix = "/rest/operations/381/source_operation/"
	citation := func(path, hash string) sourceFactRef {
		return sourceFactRef{DocumentID: "vercel:primary", Pointer: prefix + path, ValueSHA256: hash}
	}
	return []sourceFoundationDemandAuthority{{
		Key:          sourceOperationKey{Connector: "vercel", Inventory: "primary", ID: "vercel.rest.createWebhook"},
		SourcePath:   "internal/connectors/defs/vercel/sources/vercel-operation-source-lock.json",
		SourceSHA256: "2eb0130c9357c90cf74e6b47da48eac3768fec3ae3de9b15890b24a2d3f10f4e",
		Protocol:     "rest", Method: "POST", Path: "/v1/webhooks", OperationID: "createWebhook",
		Gap: "cli-webhook-event-surface-foundation-r1", AtlasID: "transport.sync-contract.v1", AtlasOwner: "polymetrics.ai/internal/synctransport",
		DecisionOwners: []string{"cli-batch1-vercel-inbound-sync-decision-r1", "cli-plan-webhook-receiver-foundation-r1-decision-webhook-exposure-product-boundary"},
		Citations: []sourceFactRef{
			citation("summary", "de9630dadd47a37e30684147218c08bddde88fe71a287c7f9f16e00c0aae690f"),
			citation("requestBody/required", "b5bea41b6c623f7c09f1bf24dcae58ebab3c0cdd90ad966bc43a45b44867e12b"),
			citation("requestBody/content/application~1json/schema/required", "67f049164296f2dc1f1d5c6b5d027464216365b43d74d1fb4780cbc16e13d296"),
			citation("requestBody/content/application~1json/schema/properties/url", "7ce701965cb4f5b9b777d0ed12ea7b0a7bb84a44d6982af853efa47ddb6e2741"),
			citation("requestBody/content/application~1json/schema/properties/events", "f24ea042467e970d3e04e8f586a44a96d36d8b320889bb1dc97ac1ab9efa9959"),
		},
	}}
}

// Read the real Atlas once per requesting builder invocation, through the same
// confined regular-file reader as other authoring inputs. No runtime consumer.
func loadSourceDemandAtlas(ctx context.Context, repo string) (map[string]string, sourceArtifactPin) {
	owners := map[string]string{}
	if ctx.Err() != nil {
		return owners, sourceArtifactPin{}
	}
	root, err := os.OpenRoot(repo)
	if err != nil {
		return owners, sourceArtifactPin{}
	}
	defer root.Close()
	raw, err := readSourceInput(root, sourceDemandAtlasPath, 64<<20)
	if err != nil {
		return owners, sourceArtifactPin{}
	}
	pin := sourceArtifactPin{Path: sourceDemandAtlasPath, SHA256: sourceBytesHash(raw), Bytes: int64(len(raw))}
	var document struct {
		Foundations []struct {
			ID    string `json:"id"`
			Owner struct {
				PrimaryPackage string `json:"primary_package"`
			} `json:"owner"`
		} `json:"foundations"`
	}
	// The full Atlas has its own closed schema; this projection still uses the
	// strict duplicate-key/node-bounded source decoder before normal unmarshaling.
	var checked any
	if decodeSourceJSON(raw, &checked) != nil || json.Unmarshal(raw, &document) != nil || ctx.Err() != nil {
		return owners, pin
	}
	for _, entry := range document.Foundations {
		if entry.ID == "" || entry.Owner.PrimaryPackage == "" {
			return map[string]string{}, pin
		}
		if _, exists := owners[entry.ID]; exists {
			return map[string]string{}, pin
		}
		owners[entry.ID] = entry.Owner.PrimaryPackage
	}
	return owners, pin
}

func validateSourceFoundationDemand(key sourceOperationKey, facts sourceFacts, a sourceSemanticAnnotation) error {
	var matched *sourceFoundationDemandAuthority
	for _, entry := range sourceFoundationDemandCatalog() {
		if entry.Key == key {
			if matched != nil {
				return fmt.Errorf("duplicate demand authority")
			}
			copy := entry
			matched = &copy
		}
	}
	if matched == nil {
		return fmt.Errorf("source demand has no established attribution")
	}
	entry := *matched
	if facts.retainedDocument.Path != entry.SourcePath || facts.retainedDocument.SHA256 != entry.SourceSHA256 || sourceBytesHash(facts.Document) != entry.SourceSHA256 || facts.Protocol != entry.Protocol || facts.Method != entry.Method || facts.Path != entry.Path || facts.OperationID != entry.OperationID {
		return fmt.Errorf("source demand identity changed")
	}
	if a.FoundationGap != entry.Gap || a.AtlasID != entry.AtlasID || facts.demandAtlasOwners[entry.AtlasID] != entry.AtlasOwner {
		return fmt.Errorf("source demand Atlas identity unavailable")
	}
	got := append([]string(nil), a.DecisionRefs...)
	want := append([]string(nil), entry.DecisionOwners...)
	sort.Strings(got)
	sort.Strings(want)
	if len(got) != len(want) {
		return fmt.Errorf("source demand owner set differs")
	}
	for i := range want {
		if got[i] != want[i] || (i > 0 && got[i] == got[i-1]) {
			return fmt.Errorf("source demand owner set differs")
		}
	}
	for _, ref := range entry.Citations {
		raw, err := sourceJSONPointer(facts.Document, ref.Pointer)
		if err != nil {
			return fmt.Errorf("source demand citation unavailable")
		}
		canonical, err := canonicalSourceJSON(raw)
		if err != nil || sourceBytesHash(canonical) != ref.ValueSHA256 {
			return fmt.Errorf("source demand citation changed")
		}
	}
	if a.Citation != entry.Citations[0] || a.Clause != sourceFactText(facts, "summary") {
		return fmt.Errorf("source demand registration assertion differs")
	}
	return nil
}
