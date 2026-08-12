package app

import (
	"sort"

	"polymetrics.ai/internal/warehouse"
)

// duckDBIdentifierKey is the one deterministic identifier comparison used by
// local-warehouse admission and DuckDB view policy. DuckDB folds ASCII letters
// for unquoted and quoted identifiers alike; do not broaden this to Unicode
// folding, path normalization, or a SQL-text rewrite.
func duckDBIdentifierKey(name string) string {
	for i := 0; i < len(name); i++ {
		if name[i] < 'A' || name[i] > 'Z' {
			continue
		}
		key := []byte(name)
		key[i] += 'a' - 'A'
		for j := i + 1; j < len(key); j++ {
			if key[j] >= 'A' && key[j] <= 'Z' {
				key[j] += 'a' - 'A'
			}
		}
		return string(key)
	}
	return name
}

// warehouseDestinationCollision is the deterministic configuration inventory
// for one local-warehouse connection. It preserves exact spellings because
// direct reads remain exact-name resolver lookups even when SQL cannot safely
// bind either spelling.
type warehouseDestinationCollision struct {
	connection   string
	connectionID string
	key          string
	tables       []string
	streams      []string
}

func (c warehouseDestinationCollision) err() error {
	return &warehouse.SameOwnerCaseEquivalentTableError{
		Connection: c.connection,
		Tables:     append([]string(nil), c.tables...),
		Streams:    append([]string(nil), c.streams...),
	}
}

func effectiveDestinationTable(streamName string, stream StreamConfig) string {
	if stream.DestinationTable != "" {
		return stream.DestinationTable
	}
	return streamName
}

// sameOwnerCaseEquivalentDestinationCollisions groups each supplied local
// warehouse connection independently. Different connections are a distinct,
// existing ambiguity contract; this helper only identifies two exact table
// spellings one owner declared that DuckDB treats as one identifier.
func sameOwnerCaseEquivalentDestinationCollisions(connections []Connection) []warehouseDestinationCollision {
	sortedConnections := append([]Connection(nil), connections...)
	sort.Slice(sortedConnections, func(i, j int) bool {
		if sortedConnections[i].Name != sortedConnections[j].Name {
			return sortedConnections[i].Name < sortedConnections[j].Name
		}
		return sortedConnections[i].ID < sortedConnections[j].ID
	})

	collisions := make([]warehouseDestinationCollision, 0)
	for _, connection := range sortedConnections {
		streamNames := make([]string, 0, len(connection.Streams))
		for streamName := range connection.Streams {
			streamNames = append(streamNames, streamName)
		}
		sort.Strings(streamNames)

		byKey := make(map[string]map[string][]string, len(streamNames))
		for _, streamName := range streamNames {
			table := effectiveDestinationTable(streamName, connection.Streams[streamName])
			key := duckDBIdentifierKey(table)
			byTable := byKey[key]
			if byTable == nil {
				byTable = make(map[string][]string)
				byKey[key] = byTable
			}
			byTable[table] = append(byTable[table], streamName)
		}

		keys := make([]string, 0, len(byKey))
		for key := range byKey {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			byTable := byKey[key]
			if len(byTable) < 2 {
				continue
			}
			tables := make([]string, 0, len(byTable))
			streams := make([]string, 0)
			for table, names := range byTable {
				tables = append(tables, table)
				streams = append(streams, names...)
			}
			sort.Strings(tables)
			sort.Strings(streams)
			collisions = append(collisions, warehouseDestinationCollision{
				connection:   connection.Name,
				connectionID: connection.ID,
				key:          key,
				tables:       tables,
				streams:      streams,
			})
		}
	}
	return collisions
}
