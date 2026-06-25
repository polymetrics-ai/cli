package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type SourceSyncMode string

const (
	SourceSyncFullRefresh SourceSyncMode = "full_refresh"
	SourceSyncIncremental SourceSyncMode = "incremental"
)

type DestinationSyncMode string

const (
	DestinationSyncAppend           DestinationSyncMode = "append"
	DestinationSyncOverwrite        DestinationSyncMode = "overwrite"
	DestinationSyncAppendDeduped    DestinationSyncMode = "append_dedup"
	DestinationSyncOverwriteDeduped DestinationSyncMode = "overwrite_dedup"
	DefaultUserFacingSyncMode                           = "full_refresh_overwrite"
)

type SyncMode struct {
	Name        string
	Source      SourceSyncMode
	Destination DestinationSyncMode
}

func ParseSyncMode(raw string) (SyncMode, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		value = DefaultUserFacingSyncMode
	}
	switch value {
	case "full_refresh_append":
		return SyncMode{Name: value, Source: SourceSyncFullRefresh, Destination: DestinationSyncAppend}, nil
	case "full_refresh_overwrite":
		return SyncMode{Name: value, Source: SourceSyncFullRefresh, Destination: DestinationSyncOverwrite}, nil
	case "full_refresh_overwrite_deduped", "full_refresh_overwrite_dedup", "full_refresh_deduped":
		return SyncMode{Name: "full_refresh_overwrite_deduped", Source: SourceSyncFullRefresh, Destination: DestinationSyncOverwriteDeduped}, nil
	case "incremental_append":
		return SyncMode{Name: value, Source: SourceSyncIncremental, Destination: DestinationSyncAppend}, nil
	case "incremental_append_deduped", "incremental_append_dedup":
		return SyncMode{Name: "incremental_append_deduped", Source: SourceSyncIncremental, Destination: DestinationSyncAppendDeduped}, nil
	default:
		return SyncMode{}, fmt.Errorf("unsupported sync mode %q", raw)
	}
}

func MustSyncModeNames() []string {
	return []string{
		"full_refresh_append",
		"full_refresh_overwrite",
		"full_refresh_overwrite_deduped",
		"incremental_append",
		"incremental_append_deduped",
	}
}

func (m SyncMode) RequiresCursor() bool {
	return m.Source == SourceSyncIncremental || m.IsDeduped()
}

func (m SyncMode) RequiresPrimaryKey() bool {
	return m.IsDeduped()
}

func (m SyncMode) IsOverwrite() bool {
	return m.Destination == DestinationSyncOverwrite || m.Destination == DestinationSyncOverwriteDeduped
}

func (m SyncMode) IsDeduped() bool {
	return m.Destination == DestinationSyncAppendDeduped || m.Destination == DestinationSyncOverwriteDeduped
}

func ValidateStreamSyncConfig(stream StreamConfig) error {
	mode, err := ParseSyncMode(stream.SyncMode)
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
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
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
	if strings.TrimSpace(cursor) == "" {
		return "", fmt.Errorf("record has empty cursor field %q", field)
	}
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
