package amazonsqs

import (
	"encoding/base64"
	"errors"
	"io"
	"strconv"
	"strings"
)

func validateSQSNumberAttributeValue(value string) error {
	mantissa := value
	exponent := int64(0)
	if index := strings.IndexAny(value, "eE"); index >= 0 {
		mantissa = value[:index]
		exponentText := value[index+1:]
		if exponentText == "" || strings.ContainsAny(exponentText, "eE") {
			return errors.New("number string_value must be a valid SQS number")
		}
		parsed, err := strconv.ParseInt(exponentText, 10, 64)
		if err != nil {
			return errors.New("number string_value must be a valid SQS number")
		}
		exponent = parsed
	}
	if strings.HasPrefix(mantissa, "+") || strings.HasPrefix(mantissa, "-") {
		mantissa = mantissa[1:]
	}
	if mantissa == "" {
		return errors.New("number string_value must be a valid SQS number")
	}
	digits := make([]byte, 0, len(mantissa))
	integerDigits := 0
	dotSeen := false
	for index := 0; index < len(mantissa); index++ {
		char := mantissa[index]
		switch {
		case char >= '0' && char <= '9':
			digits = append(digits, char)
			if !dotSeen {
				integerDigits++
			}
		case char == '.' && !dotSeen:
			dotSeen = true
		default:
			return errors.New("number string_value must be a valid SQS number")
		}
	}
	if len(digits) == 0 {
		return errors.New("number string_value must be a valid SQS number")
	}
	firstNonZero := -1
	lastNonZero := -1
	for index, digit := range digits {
		if digit == '0' {
			continue
		}
		if firstNonZero < 0 {
			firstNonZero = index
		}
		lastNonZero = index
	}
	if firstNonZero < 0 {
		return nil
	}
	precision := lastNonZero - firstNonZero + 1
	if precision > 38 {
		return errors.New("number string_value must have at most 38 digits of precision")
	}
	baseExponent := int64(integerDigits - firstNonZero - 1)
	if exponent > 126-baseExponent || exponent < -128-baseExponent {
		return errors.New("number string_value magnitude must be between 1e-128 and 1e126")
	}
	normalizedExponent := baseExponent + exponent
	if normalizedExponent < -128 || normalizedExponent > 126 {
		return errors.New("number string_value magnitude must be between 1e-128 and 1e126")
	}
	if normalizedExponent == 126 && (precision != 1 || digits[firstNonZero] != '1') {
		return errors.New("number string_value magnitude must be between 1e-128 and 1e126")
	}
	return nil
}

func validateSQSBinaryAttributeValue(value string) error {
	decoder := base64.NewDecoder(base64.StdEncoding.Strict(), strings.NewReader(value))
	decodedBytes, err := io.Copy(io.Discard, decoder)
	if err != nil {
		return errors.New("binary_value must use standard base64 encoding")
	}
	if decodedBytes == 0 {
		return errors.New("binary_value must decode to at least one byte")
	}
	return nil
}

func validateSQSXRayTraceHeader(value string) error {
	seen := map[string]bool{}
	for _, field := range strings.Split(value, ";") {
		name, fieldValue, ok := strings.Cut(field, "=")
		if !ok || !validSQSXRayField(name, fieldValue) || seen[name] {
			return errors.New("string_value must be a valid AWS X-Ray trace header")
		}
		seen[name] = true
		switch name {
		case "Root":
			parts := strings.Split(fieldValue, "-")
			if len(parts) != 3 || parts[0] != "1" || !validSQSXRayHex(parts[1], 8) || !validSQSXRayHex(parts[2], 24) {
				return errors.New("string_value must be a valid AWS X-Ray trace header")
			}
		case "Parent":
			if !validSQSXRayHex(fieldValue, 16) {
				return errors.New("string_value must be a valid AWS X-Ray trace header")
			}
		case "Sampled":
			if fieldValue != "0" && fieldValue != "1" && fieldValue != "?" {
				return errors.New("string_value must be a valid AWS X-Ray trace header")
			}
		}
	}
	if !seen["Root"] {
		return errors.New("string_value must be a valid AWS X-Ray trace header")
	}
	return nil
}

func validSQSXRayField(name, value string) bool {
	if name == "" || value == "" {
		return false
	}
	for index := 0; index < len(name); index++ {
		char := name[index]
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_' || char == '-' {
			continue
		}
		return false
	}
	for index := 0; index < len(value); index++ {
		if value[index] <= ' ' || value[index] >= 0x7f || value[index] == ';' {
			return false
		}
	}
	return true
}

func validSQSXRayHex(value string, length int) bool {
	if len(value) != length {
		return false
	}
	for index := 0; index < len(value); index++ {
		char := value[index]
		if (char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F') {
			continue
		}
		return false
	}
	return true
}
