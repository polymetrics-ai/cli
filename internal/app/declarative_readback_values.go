package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"sort"
	"strconv"

	"polymetrics.ai/internal/connectors"
)

// declarativeReadBackValuesEqual compares only the closed transport value
// vocabulary. Numeric leaves compare as exact rational values: exponent and
// scale are semantic (`42 == 42.0`, `1e3 == 1000`), while strings and booleans
// retain their JSON types. Floats are converted from their exact IEEE value,
// never through a formatted float64 approximation; non-finite values fail
// closed because JSON cannot represent them.
func declarativeReadBackValuesEqual(left, right any) (bool, error) {
	leftCanonical, err := canonicalDeclarativeReadBackValue(left)
	if err != nil {
		return false, err
	}
	rightCanonical, err := canonicalDeclarativeReadBackValue(right)
	if err != nil {
		return false, err
	}
	return bytes.Equal(leftCanonical, rightCanonical), nil
}

func canonicalDeclarativeReadBackValue(value any) ([]byte, error) {
	var out bytes.Buffer
	if err := appendCanonicalDeclarativeReadBackValue(&out, value); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func appendCanonicalDeclarativeReadBackValue(out *bytes.Buffer, value any) error {
	switch typed := value.(type) {
	case nil:
		out.WriteString("null;")
	case bool:
		if typed {
			out.WriteString("bool:1;")
		} else {
			out.WriteString("bool:0;")
		}
	case string:
		appendCanonicalDeclarativeReadBackString(out, "string", typed)
	case json.Number:
		rat, err := declarativeReadBackJSONNumberRat(typed)
		if err != nil {
			return err
		}
		appendCanonicalDeclarativeReadBackRat(out, rat)
	case int:
		appendCanonicalDeclarativeReadBackRat(out, big.NewRat(int64(typed), 1))
	case int8:
		appendCanonicalDeclarativeReadBackRat(out, big.NewRat(int64(typed), 1))
	case int16:
		appendCanonicalDeclarativeReadBackRat(out, big.NewRat(int64(typed), 1))
	case int32:
		appendCanonicalDeclarativeReadBackRat(out, big.NewRat(int64(typed), 1))
	case int64:
		appendCanonicalDeclarativeReadBackRat(out, big.NewRat(typed, 1))
	case uint:
		appendCanonicalDeclarativeReadBackRat(out, new(big.Rat).SetInt(new(big.Int).SetUint64(uint64(typed))))
	case uint8:
		appendCanonicalDeclarativeReadBackRat(out, new(big.Rat).SetInt(new(big.Int).SetUint64(uint64(typed))))
	case uint16:
		appendCanonicalDeclarativeReadBackRat(out, new(big.Rat).SetInt(new(big.Int).SetUint64(uint64(typed))))
	case uint32:
		appendCanonicalDeclarativeReadBackRat(out, new(big.Rat).SetInt(new(big.Int).SetUint64(uint64(typed))))
	case uint64:
		appendCanonicalDeclarativeReadBackRat(out, new(big.Rat).SetInt(new(big.Int).SetUint64(typed)))
	case float32:
		return appendCanonicalDeclarativeReadBackFloat(out, float64(typed))
	case float64:
		return appendCanonicalDeclarativeReadBackFloat(out, typed)
	case connectors.Record:
		return appendCanonicalDeclarativeReadBackMap(out, map[string]any(typed))
	case map[string]any:
		return appendCanonicalDeclarativeReadBackMap(out, typed)
	case map[string]string:
		values := make(map[string]any, len(typed))
		for key, nested := range typed {
			values[key] = nested
		}
		return appendCanonicalDeclarativeReadBackMap(out, values)
	case []any:
		return appendCanonicalDeclarativeReadBackSlice(out, typed)
	case []connectors.Record:
		out.WriteString("records[")
		for _, nested := range typed {
			if err := appendCanonicalDeclarativeReadBackValue(out, nested); err != nil {
				return err
			}
		}
		out.WriteString("];")
	case []byte:
		appendCanonicalDeclarativeReadBackBytes(out, "bytes", typed)
	case json.RawMessage:
		decoder := json.NewDecoder(bytes.NewReader(typed))
		decoder.UseNumber()
		var decoded any
		if err := decoder.Decode(&decoded); err != nil {
			return fmt.Errorf("read-back raw JSON: %w", err)
		}
		if err := requireOnlyOneDeclarativeReadBackJSONValue(decoder); err != nil {
			return err
		}
		out.WriteString("raw:")
		return appendCanonicalDeclarativeReadBackValue(out, decoded)
	default:
		return fmt.Errorf("read-back comparison does not support value type %T", value)
	}
	return nil
}

func declarativeReadBackJSONNumberRat(value json.Number) (*big.Rat, error) {
	raw := value.String()
	if !json.Valid([]byte(raw)) {
		return nil, fmt.Errorf("read-back comparison received invalid JSON number %q", raw)
	}
	rat, ok := new(big.Rat).SetString(raw)
	if !ok {
		return nil, fmt.Errorf("read-back comparison cannot represent JSON number %q", raw)
	}
	return rat, nil
}

func appendCanonicalDeclarativeReadBackFloat(out *bytes.Buffer, value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("read-back comparison rejects non-finite float")
	}
	rat := new(big.Rat).SetFloat64(value)
	if rat == nil {
		return fmt.Errorf("read-back comparison cannot represent finite float")
	}
	appendCanonicalDeclarativeReadBackRat(out, rat)
	return nil
}

func appendCanonicalDeclarativeReadBackRat(out *bytes.Buffer, value *big.Rat) {
	out.WriteString("number:")
	appendCanonicalDeclarativeReadBackString(out, "numerator", value.Num().String())
	appendCanonicalDeclarativeReadBackString(out, "denominator", value.Denom().String())
}

func appendCanonicalDeclarativeReadBackMap(out *bytes.Buffer, values map[string]any) error {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out.WriteString("map{")
	for _, key := range keys {
		appendCanonicalDeclarativeReadBackString(out, "key", key)
		if err := appendCanonicalDeclarativeReadBackValue(out, values[key]); err != nil {
			return err
		}
	}
	out.WriteString("};")
	return nil
}

func appendCanonicalDeclarativeReadBackSlice(out *bytes.Buffer, values []any) error {
	out.WriteString("list[")
	for _, value := range values {
		if err := appendCanonicalDeclarativeReadBackValue(out, value); err != nil {
			return err
		}
	}
	out.WriteString("];")
	return nil
}

func appendCanonicalDeclarativeReadBackString(out *bytes.Buffer, kind, value string) {
	out.WriteString(kind)
	out.WriteByte(':')
	out.WriteString(strconv.Itoa(len(value)))
	out.WriteByte(':')
	out.WriteString(value)
	out.WriteByte(';')
}

func appendCanonicalDeclarativeReadBackBytes(out *bytes.Buffer, kind string, value []byte) {
	out.WriteString(kind)
	out.WriteByte(':')
	out.WriteString(strconv.Itoa(len(value)))
	out.WriteByte(':')
	out.Write(value)
	out.WriteByte(';')
}

func requireOnlyOneDeclarativeReadBackJSONValue(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("read-back raw JSON has trailing value")
		}
		return fmt.Errorf("read-back raw JSON: %w", err)
	}
	return nil
}
