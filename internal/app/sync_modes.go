package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/synccontract"
)

type SourceSyncMode string

const (
	SourceSyncFullRefresh   SourceSyncMode = "full_refresh"
	SourceSyncIncremental   SourceSyncMode = "incremental"
	SourceSyncChangeCapture SourceSyncMode = "change_capture"
)

type DestinationSyncMode string

const (
	DestinationSyncAppend           DestinationSyncMode = "append"
	DestinationSyncOverwrite        DestinationSyncMode = "overwrite"
	DestinationSyncAppendDeduped    DestinationSyncMode = "append_dedup"
	DestinationSyncOverwriteDeduped DestinationSyncMode = "overwrite_dedup"
	DestinationSyncUpsert           DestinationSyncMode = "upsert"
	DestinationSyncDedupeHistory    DestinationSyncMode = "dedupe_history"
	DefaultUserFacingSyncMode                           = "full_refresh_overwrite"
)

type SyncMode struct {
	Name                string
	Source              SourceSyncMode
	Destination         DestinationSyncMode
	ContractMode        synccontract.Mode
	LegacyCompatibility bool
}

const syncModeCompatibilityVersion uint = 1

type syncModeDefinition struct {
	mode                         SyncMode
	legacyInput                  bool
	persistedLegacyCompatibility bool
}

func (d syncModeDefinition) persistedLegacyMode() SyncMode {
	mode := d.mode
	mode.LegacyCompatibility = d.persistedLegacyCompatibility
	return mode
}

var syncModeDefinitions = map[string]syncModeDefinition{
	"full_refresh_append": {
		mode:                         SyncMode{Name: "full_refresh_append", Source: SourceSyncFullRefresh, Destination: DestinationSyncAppend, LegacyCompatibility: true},
		legacyInput:                  true,
		persistedLegacyCompatibility: true,
	},
	"full_refresh_overwrite": {
		mode:                         SyncMode{Name: "full_refresh_overwrite", Source: SourceSyncFullRefresh, Destination: DestinationSyncOverwrite, LegacyCompatibility: true},
		legacyInput:                  true,
		persistedLegacyCompatibility: true,
	},
	"full_refresh_overwrite_deduped": {
		mode:        SyncMode{Name: "full_refresh_overwrite_deduped", Source: SourceSyncFullRefresh, Destination: DestinationSyncOverwriteDeduped},
		legacyInput: true,
	},
	string(synccontract.ModeFullOverwrite): {
		mode: SyncMode{Name: string(synccontract.ModeFullOverwrite), Source: SourceSyncFullRefresh, Destination: DestinationSyncOverwrite, ContractMode: synccontract.ModeFullOverwrite},
	},
	string(synccontract.ModeFullAppend): {
		mode: SyncMode{Name: string(synccontract.ModeFullAppend), Source: SourceSyncFullRefresh, Destination: DestinationSyncAppend, ContractMode: synccontract.ModeFullAppend},
	},
	"incremental_append": {
		mode:                         SyncMode{Name: "incremental_append", Source: SourceSyncIncremental, Destination: DestinationSyncAppend},
		legacyInput:                  true,
		persistedLegacyCompatibility: true,
	},
	"incremental_append_deduped": {
		mode:        SyncMode{Name: "incremental_append_deduped", Source: SourceSyncIncremental, Destination: DestinationSyncAppendDeduped},
		legacyInput: true,
	},
	string(synccontract.ModeIncrementalUpsert): {
		mode: SyncMode{Name: string(synccontract.ModeIncrementalUpsert), Source: SourceSyncIncremental, Destination: DestinationSyncUpsert, ContractMode: synccontract.ModeIncrementalUpsert},
	},
	string(synccontract.ModeIncrementalDedupe): {
		mode: SyncMode{Name: string(synccontract.ModeIncrementalDedupe), Source: SourceSyncIncremental, Destination: DestinationSyncAppendDeduped, ContractMode: synccontract.ModeIncrementalDedupe},
	},
	string(synccontract.ModeIncrementalDedupeHistory): {
		mode: SyncMode{Name: string(synccontract.ModeIncrementalDedupeHistory), Source: SourceSyncIncremental, Destination: DestinationSyncDedupeHistory, ContractMode: synccontract.ModeIncrementalDedupeHistory},
	},
	string(synccontract.ModeChangeCapture): {
		mode: SyncMode{Name: string(synccontract.ModeChangeCapture), Source: SourceSyncChangeCapture, Destination: DestinationSyncUpsert, ContractMode: synccontract.ModeChangeCapture},
	},
}

func ParseSyncMode(raw string) (SyncMode, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		value = DefaultUserFacingSyncMode
	}
	if mode, _, ok := publicSyncMode(value); ok {
		return mode, nil
	}
	if definition, ok := syncModeDefinitions[value]; ok {
		return definition.mode, nil
	}
	return SyncMode{}, fmt.Errorf("unsupported sync mode %q", raw)
}

func publicSyncMode(raw string) (SyncMode, synccontract.PublicMode, bool) {
	public, ok := synccontract.LookupPublicMode(raw)
	if !ok {
		return SyncMode{}, synccontract.PublicMode{}, false
	}
	definition, ok := syncModeDefinitions[public.Name]
	if !ok {
		return SyncMode{}, synccontract.PublicMode{}, false
	}
	mode := definition.mode
	mode.ContractMode = public.ContractMode
	return mode, public, true
}

func ParseStreamSyncMode(stream StreamConfig) (SyncMode, error) {
	if stream.LegacyCompatibility {
		return parseLegacySyncMode(stream.SyncMode)
	}
	return ParseSyncMode(stream.SyncMode)
}

func parseLegacySyncMode(raw string) (SyncMode, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		value = DefaultUserFacingSyncMode
	}
	mode, public, ok := publicSyncMode(value)
	if !ok {
		return SyncMode{}, fmt.Errorf("sync mode %q is not an explicit legacy compatibility adapter", raw)
	}
	definition, ok := syncModeDefinitions[public.Name]
	if ok && definition.legacyInput {
		mode = definition.persistedLegacyMode()
		mode.ContractMode = public.ContractMode
		return mode, nil
	}
	return SyncMode{}, fmt.Errorf("sync mode %q is not an explicit legacy compatibility adapter", raw)
}

func isLegacySyncModeName(raw string) bool {
	if strings.TrimSpace(raw) == "" {
		return false
	}
	_, err := parseLegacySyncMode(raw)
	return err == nil
}

func MustSyncModeNames() []string {
	return synccontract.PublicModeNames()
}

func (m SyncMode) RequiresCursor() bool {
	return m.Source == SourceSyncIncremental || m.IsDeduped()
}

func (m SyncMode) RequiresPrimaryKey() bool {
	return m.IsDeduped() || m.Destination == DestinationSyncUpsert || m.Destination == DestinationSyncDedupeHistory
}

func (m SyncMode) IsOverwrite() bool {
	return m.Destination == DestinationSyncOverwrite || m.Destination == DestinationSyncOverwriteDeduped
}

func (m SyncMode) IsDeduped() bool {
	return m.Destination == DestinationSyncAppendDeduped || m.Destination == DestinationSyncOverwriteDeduped || m.Destination == DestinationSyncDedupeHistory
}

// IsContractMode reports a new closed vocabulary entry. It may parse and be
// persisted, but RunETL refuses to execute it until a native executor and the
// shared conformance evidence have been admitted.
func (m SyncMode) IsContractMode() bool {
	return m.ContractMode != "" && !m.LegacyCompatibility
}

func ValidateStreamSyncConfig(stream StreamConfig) error {
	mode, err := ParseStreamSyncMode(stream)
	if err != nil {
		return err
	}
	if mode.RequiresCursor() && strings.TrimSpace(stream.CursorField) == "" {
		return fmt.Errorf("sync mode %s requires a cursor field", mode.Name)
	}
	if mode.RequiresPrimaryKey() && len(stream.PrimaryKey) == 0 {
		return fmt.Errorf("sync mode %s requires at least one primary key field", mode.Name)
	}
	return nil
}

func streamStateKey(connection, stream string) string {
	return connection + ":" + stream
}

func compareCursor(a, b string) int {
	if a == b {
		return 0
	}
	if at, aerr := time.Parse(time.RFC3339Nano, a); aerr == nil {
		if bt, berr := time.Parse(time.RFC3339Nano, b); berr == nil {
			switch {
			case at.Before(bt):
				return -1
			case at.After(bt):
				return 1
			default:
				return 0
			}
		}
	}
	if af, aerr := strconv.ParseFloat(a, 64); aerr == nil {
		if bf, berr := strconv.ParseFloat(b, 64); berr == nil {
			switch {
			case af < bf:
				return -1
			case af > bf:
				return 1
			default:
				return 0
			}
		}
	}
	return strings.Compare(a, b)
}

func toComparableString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case json.Number:
		return v.String()
	case float64:
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case bool:
		return strconv.FormatBool(v)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprint(v)
		}
		return string(b)
	}
}

func recordCursor(record map[string]any, field string) (string, error) {
	field = strings.TrimSpace(field)
	if field == "" {
		return "", nil
	}
	value, ok := record[field]
	if !ok || value == nil {
		return "", fmt.Errorf("record is missing cursor field %q", field)
	}
	cursor := toComparableString(value)
	return cursor, nil
}

func primaryKeyTuple(record map[string]any, fields []string) (string, error) {
	if len(fields) == 0 {
		return "", errors.New("primary key fields are required")
	}
	values := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			return "", errors.New("primary key field cannot be empty")
		}
		value, ok := record[field]
		if !ok || value == nil {
			return "", fmt.Errorf("record is missing primary key field %q", field)
		}
		values = append(values, toComparableString(value))
	}
	b, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
