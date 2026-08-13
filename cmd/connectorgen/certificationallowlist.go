package main

// certificationConnectorAllowlist is the reviewable set of connectors for
// which this repository currently makes a proof-bearing certification claim.
// Adding a connector is intentionally one line here plus its generated shard.
var certificationConnectorAllowlist = []string{
	"github",
	"postgres",
}

func certificationConnectorAllowed(name string) bool {
	for _, allowed := range certificationConnectorAllowlist {
		if name == allowed {
			return true
		}
	}
	return false
}
