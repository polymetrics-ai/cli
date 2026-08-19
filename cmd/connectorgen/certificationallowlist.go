package main

// certificationConnectorAllowlist is the reviewable set of connectors for
// which this repository currently makes a proof-bearing certification claim.
// Adding a connector is intentionally one line here; --all regenerates its
// shard and the compact runtime status projection.
var certificationConnectorAllowlist = []string{
	"github",
	"postgres",
	"zoom",
}

func certificationConnectorAllowed(name string) bool {
	for _, allowed := range certificationConnectorAllowlist {
		if name == allowed {
			return true
		}
	}
	return false
}
