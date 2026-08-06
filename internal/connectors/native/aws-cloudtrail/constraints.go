package awscloudtrail

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

type awsStringConstraints struct {
	minimumLength int
	maximumLength int
	fixedLength   int
	pattern       *regexp.Regexp
	values        map[string]struct{}
	valueList     string
}

type awsFieldConstraints struct {
	minimum      int
	maximum      int
	fixedItems   int
	minimumItems int
	maximumItems int
	strings      awsStringConstraints
	keys         awsStringConstraints
	values       awsStringConstraints
	validKeys    map[string]struct{}
	keyList      string
}

type awsActionCrossFields struct {
	anyOf        []string
	exactlyOneOf [][]string
}

var (
	cloudTrailActionFieldConstraints = buildActionFieldConstraints(cloudTrailActionFields)
	cloudTrailActionCrossFields      = map[string]awsActionCrossFields{
		"DescribeQuery":         {anyOf: []string{"QueryId", "QueryAlias"}},
		"GetEventConfiguration": {anyOf: []string{"EventDataStore", "TrailName"}},
		"GetInsightSelectors":   {exactlyOneOf: [][]string{{"EventDataStore"}, {"TrailName"}}},
		"StartImport":           {exactlyOneOf: [][]string{{"ImportId"}, {"Destinations", "ImportSource"}}},
	}
	integerRangePattern  = regexp.MustCompile(`(?i)Valid Range:\s*Minimum value of ([0-9]+)\.\s*Maximum value of ([0-9]+)\.`)
	fixedItemsPattern    = regexp.MustCompile(`(?i)(?:Array Members|Map Entries):\s*Fixed number of ([0-9]+) items?\.`)
	minimumItemsPattern  = regexp.MustCompile(`(?i)(?:Array Members|Map Entries):\s*Minimum number of ([0-9]+) items?\.`)
	maximumItemsPattern  = regexp.MustCompile(`(?i)(?:Array Members|Map Entries):.*?Maximum number of ([0-9]+) items?\.`)
	fixedLengthPattern   = regexp.MustCompile(`(?i)Fixed length of ([0-9]+)\.`)
	minimumLengthPattern = regexp.MustCompile(`(?i)Minimum length of ([0-9]+)\.`)
	maximumLengthPattern = regexp.MustCompile(`(?i)Maximum length of ([0-9]+)\.`)
)

func buildActionFieldConstraints(actions map[string][]awsActionField) map[string]map[string]awsFieldConstraints {
	constraints := make(map[string]map[string]awsFieldConstraints, len(actions))
	for action, fields := range actions {
		byName := make(map[string]awsFieldConstraints, len(fields))
		for _, field := range fields {
			byName[field.Name] = parseActionFieldConstraints(field.Type)
		}
		constraints[action] = byName
	}
	return constraints
}

func parseActionFieldConstraints(description string) awsFieldConstraints {
	constraints := awsFieldConstraints{
		fixedItems:   firstConstraintNumber(description, fixedItemsPattern),
		minimumItems: firstConstraintNumber(description, minimumItemsPattern),
		maximumItems: firstConstraintNumber(description, maximumItemsPattern),
	}
	if match := integerRangePattern.FindStringSubmatch(description); len(match) == 3 {
		constraints.minimum, _ = strconv.Atoi(match[1])
		constraints.maximum, _ = strconv.Atoi(match[2])
	}
	if strings.Contains(strings.ToLower(description), "map") {
		constraints.keys = parseStringConstraints(sectionAfter(description, "Key Length Constraints:", "Value Length Constraints:"))
		constraints.values = parseStringConstraints(sectionAfter(description, "Value Length Constraints:", ""))
		constraints.validKeys, constraints.keyList = parseDelimitedValues(description, "Valid Keys:", "Value Length Constraints:", "Key Length Constraints:")
		return constraints
	}
	constraints.strings = parseStringConstraints(description)
	return constraints
}

func firstConstraintNumber(description string, pattern *regexp.Regexp) int {
	match := pattern.FindStringSubmatch(description)
	if len(match) != 2 {
		return 0
	}
	value, _ := strconv.Atoi(match[1])
	return value
}

func parseStringConstraints(description string) awsStringConstraints {
	constraints := awsStringConstraints{
		fixedLength:   firstConstraintNumber(description, fixedLengthPattern),
		minimumLength: firstConstraintNumber(description, minimumLengthPattern),
		maximumLength: firstConstraintNumber(description, maximumLengthPattern),
	}
	constraints.values, constraints.valueList = parseDelimitedValues(description, "Valid Values:", "Pattern:", "Key Length Constraints:", "Value Length Constraints:")
	pattern := strings.TrimSpace(sectionAfter(description, "Pattern:", "Key Length Constraints:", "Value Length Constraints:"))
	if pattern != "" {
		if compiled, err := regexp.Compile(pattern); err == nil {
			constraints.pattern = compiled
		}
	}
	return constraints
}

func sectionAfter(description, marker string, stops ...string) string {
	index := strings.Index(strings.ToLower(description), strings.ToLower(marker))
	if index < 0 {
		return ""
	}
	section := description[index+len(marker):]
	lower := strings.ToLower(section)
	for _, stop := range stops {
		if stopIndex := strings.Index(lower, strings.ToLower(stop)); stopIndex >= 0 {
			section = section[:stopIndex]
			lower = lower[:stopIndex]
		}
	}
	return strings.TrimSpace(section)
}

func parseDelimitedValues(description, marker string, stops ...string) (map[string]struct{}, string) {
	section := strings.TrimSpace(sectionAfter(description, marker, stops...))
	if section == "" {
		return nil, ""
	}
	section = strings.TrimSuffix(section, ".")
	values := map[string]struct{}{}
	parts := make([]string, 0)
	for _, part := range strings.Split(section, "|") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		values[part] = struct{}{}
		parts = append(parts, part)
	}
	if len(values) == 0 {
		return nil, ""
	}
	return values, strings.Join(parts, " | ")
}

func validateActionField(action string, field awsActionField, value any) error {
	constraints := cloudTrailActionFieldConstraints[action][field.Name]
	typeName := strings.ToLower(field.Type)
	switch {
	case strings.Contains(typeName, "integer"):
		integer, ok := value.(int)
		if !ok {
			return fmt.Errorf("want integer")
		}
		if constraints.minimum > 0 && integer < constraints.minimum {
			return fmt.Errorf("must be at least %d", constraints.minimum)
		}
		if constraints.maximum > 0 && integer > constraints.maximum {
			return fmt.Errorf("must be at most %d", constraints.maximum)
		}
	case strings.Contains(typeName, "array of"):
		items, ok := actionArray(value)
		if !ok {
			return fmt.Errorf("want array")
		}
		if err := validateActionItemCount(items, constraints); err != nil {
			return err
		}
		if strings.Contains(typeName, "array of strings") {
			for index, item := range items {
				if err := validateActionString(stringifyValue(item), constraints.strings); err != nil {
					return fmt.Errorf("item %d: %w", index, err)
				}
			}
		}
	case strings.Contains(typeName, "map"):
		object, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("want object")
		}
		if err := validateActionObject(object, constraints); err != nil {
			return err
		}
	case strings.Contains(typeName, "object"):
		if _, ok := value.(map[string]any); !ok {
			return fmt.Errorf("want object")
		}
	default:
		if err := validateActionString(stringifyValue(value), constraints.strings); err != nil {
			return err
		}
	}

	switch {
	case action == "LookupEvents" && field.Name == "LookupAttributes":
		return validateLookupAttributes(value)
	case action == "StartImport" && field.Name == "ImportSource":
		return validateImportSource(value)
	}
	return nil
}

func validateActionItemCount(items []any, constraints awsFieldConstraints) error {
	if constraints.fixedItems > 0 && len(items) != constraints.fixedItems {
		return fmt.Errorf("must contain exactly %d items", constraints.fixedItems)
	}
	if constraints.minimumItems > 0 && len(items) < constraints.minimumItems {
		return fmt.Errorf("must contain at least %d items", constraints.minimumItems)
	}
	if constraints.maximumItems > 0 && len(items) > constraints.maximumItems {
		return fmt.Errorf("must contain at most %d items", constraints.maximumItems)
	}
	return nil
}

func validateActionObject(object map[string]any, constraints awsFieldConstraints) error {
	items := make([]any, 0, len(object))
	for _, value := range object {
		items = append(items, value)
	}
	if err := validateActionItemCount(items, constraints); err != nil {
		return err
	}
	for key, value := range object {
		if len(constraints.validKeys) > 0 {
			if _, ok := constraints.validKeys[key]; !ok {
				return fmt.Errorf("key %q must be one of %s", key, constraints.keyList)
			}
		}
		if err := validateActionString(key, constraints.keys); err != nil {
			return fmt.Errorf("key %q: %w", key, err)
		}
		if err := validateActionString(stringifyValue(value), constraints.values); err != nil {
			return fmt.Errorf("value for key %q: %w", key, err)
		}
	}
	return nil
}

func validateActionString(value string, constraints awsStringConstraints) error {
	if constraints.fixedLength > 0 && utf8.RuneCountInString(value) != constraints.fixedLength {
		return fmt.Errorf("must have length %d", constraints.fixedLength)
	}
	if constraints.minimumLength > 0 && utf8.RuneCountInString(value) < constraints.minimumLength {
		return fmt.Errorf("must have length at least %d", constraints.minimumLength)
	}
	if constraints.maximumLength > 0 && utf8.RuneCountInString(value) > constraints.maximumLength {
		return fmt.Errorf("must have length at most %d", constraints.maximumLength)
	}
	if len(constraints.values) > 0 {
		if _, ok := constraints.values[value]; !ok {
			return fmt.Errorf("must be one of %s", constraints.valueList)
		}
	}
	if constraints.pattern != nil {
		match := constraints.pattern.FindStringIndex(value)
		if match == nil || match[0] != 0 || match[1] != len(value) {
			return fmt.Errorf("does not match documented pattern")
		}
	}
	return nil
}

func actionArray(value any) ([]any, bool) {
	switch items := value.(type) {
	case []any:
		return items, true
	case []string:
		out := make([]any, len(items))
		for index, item := range items {
			out[index] = item
		}
		return out, true
	default:
		return nil, false
	}
}

func actionValuePresent(value any) bool {
	switch typed := value.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(typed) != ""
	case []any:
		return len(typed) > 0
	case []string:
		return len(typed) > 0
	case map[string]any:
		return len(typed) > 0
	default:
		return true
	}
}

func requiredActionValuePresent(value any) bool {
	switch typed := value.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(typed) != ""
	default:
		return true
	}
}

func validateActionCrossFields(action string, body map[string]any) error {
	rules := cloudTrailActionCrossFields[action]
	if len(rules.anyOf) > 0 && !hasActionField(body, rules.anyOf) {
		return fmt.Errorf("aws-cloudtrail %s requires one of %s", action, strings.Join(rules.anyOf, " or "))
	}
	if len(rules.exactlyOneOf) > 0 {
		matches := 0
		for _, alternative := range rules.exactlyOneOf {
			if allActionFieldsPresent(body, alternative) {
				matches++
			}
		}
		if matches != 1 {
			parts := make([]string, 0, len(rules.exactlyOneOf))
			for _, alternative := range rules.exactlyOneOf {
				parts = append(parts, strings.Join(alternative, " and "))
			}
			return fmt.Errorf("aws-cloudtrail %s requires exactly one of %s", action, strings.Join(parts, " or "))
		}
	}
	for _, pair := range [][2]string{{"StartTime", "EndTime"}, {"StartEventTime", "EndEventTime"}} {
		if err := validateTimestampOrder(action, body, pair[0], pair[1]); err != nil {
			return err
		}
	}
	return validateActionSpecificCrossFields(action, body)
}

func validateActionSpecificCrossFields(action string, body map[string]any) error {
	if action == "UpdateEventDataStore" && !hasActionField(body, []string{"AdvancedEventSelectors", "BillingMode", "KmsKeyId", "MultiRegionEnabled", "Name", "OrganizationEnabled", "RetentionPeriod", "TerminationProtectionEnabled"}) {
		return fmt.Errorf("aws-cloudtrail UpdateEventDataStore requires at least one field besides EventDataStore")
	}
	switch action {
	case "PutInsightSelectors":
		return validatePutInsightSelectors(body)
	case "CreateEventDataStore", "UpdateEventDataStore":
		if err := validateEventDataStoreRetention(action, body); err != nil {
			return err
		}
	case "ListInsightsMetricData":
		if insightType, _ := actionString(body["InsightType"]); insightType == "ApiErrorRateInsight" && !hasActionField(body, []string{"ErrorCode"}) {
			return fmt.Errorf("aws-cloudtrail ListInsightsMetricData requires ErrorCode for ApiErrorRateInsight")
		}
		if period, ok := body["Period"].(int); ok && period != 60 && period != 300 && period != 3600 {
			return fmt.Errorf("aws-cloudtrail ListInsightsMetricData Period must be one of 60, 300, or 3600")
		}
	case "PutEventConfiguration":
		return validatePutEventConfiguration(body)
	}
	return nil
}

func validatePutInsightSelectors(body map[string]any) error {
	hasEventDataStore := hasActionField(body, []string{"EventDataStore"})
	hasDestination := hasActionField(body, []string{"InsightsDestination"})
	hasTrail := hasActionField(body, []string{"TrailName"})
	if hasTrail {
		if hasEventDataStore || hasDestination {
			return fmt.Errorf("aws-cloudtrail PutInsightSelectors cannot combine TrailName with EventDataStore or InsightsDestination")
		}
		return nil
	}
	if !hasEventDataStore {
		return fmt.Errorf("aws-cloudtrail PutInsightSelectors requires TrailName or EventDataStore")
	}
	if actionValuePresent(body["InsightSelectors"]) && !hasDestination {
		return fmt.Errorf("aws-cloudtrail PutInsightSelectors requires InsightsDestination when enabling Insights on an EventDataStore")
	}
	return nil
}

func validateEventDataStoreRetention(action string, body map[string]any) error {
	billingMode, _ := actionString(body["BillingMode"])
	retentionPeriod, ok := body["RetentionPeriod"].(int)
	if billingMode == "FIXED_RETENTION_PRICING" && ok && retentionPeriod > 2557 {
		return fmt.Errorf("aws-cloudtrail %s RetentionPeriod must be at most 2557 for FIXED_RETENTION_PRICING", action)
	}
	return nil
}

func validatePutEventConfiguration(body map[string]any) error {
	hasEventDataStore := hasActionField(body, []string{"EventDataStore"})
	hasTrail := hasActionField(body, []string{"TrailName"})
	if hasEventDataStore == hasTrail {
		return fmt.Errorf("aws-cloudtrail PutEventConfiguration requires exactly one of EventDataStore or TrailName")
	}
	if hasActionField(body, []string{"ContextKeySelectors"}) {
		if !hasEventDataStore {
			return fmt.Errorf("aws-cloudtrail PutEventConfiguration requires EventDataStore with ContextKeySelectors")
		}
		if maxEventSize, _ := actionString(body["MaxEventSize"]); maxEventSize != "Large" {
			return fmt.Errorf("aws-cloudtrail PutEventConfiguration requires MaxEventSize Large with ContextKeySelectors")
		}
	}
	if hasActionField(body, []string{"AggregationConfigurations"}) && !hasTrail {
		return fmt.Errorf("aws-cloudtrail PutEventConfiguration requires TrailName with AggregationConfigurations")
	}
	if hasActionField(body, []string{"MaxEventSize"}) && !hasEventDataStore {
		return fmt.Errorf("aws-cloudtrail PutEventConfiguration requires EventDataStore with MaxEventSize")
	}
	return nil
}

func hasActionField(body map[string]any, fields []string) bool {
	for _, field := range fields {
		if value, ok := body[field]; ok && actionValuePresent(value) {
			return true
		}
	}
	return false
}

func allActionFieldsPresent(body map[string]any, fields []string) bool {
	for _, field := range fields {
		value, ok := body[field]
		if !ok || !actionValuePresent(value) {
			return false
		}
	}
	return true
}

func validateTimestampOrder(action string, body map[string]any, startField, endField string) error {
	start, startPresent := actionTimestamp(body[startField])
	end, endPresent := actionTimestamp(body[endField])
	if !startPresent || !endPresent {
		return nil
	}
	if end < start {
		return fmt.Errorf("aws-cloudtrail %s requires %s not to precede %s", action, endField, startField)
	}
	return nil
}

func actionTimestamp(value any) (int64, bool) {
	switch typed := value.(type) {
	case int:
		return int64(typed), true
	case int64:
		return typed, true
	default:
		return 0, false
	}
}

func validateLookupAttributes(value any) error {
	attributes, ok := actionArray(value)
	if !ok {
		return fmt.Errorf("want array")
	}
	if len(attributes) != 1 {
		return fmt.Errorf("must contain exactly one LookupAttribute")
	}
	attribute, ok := attributes[0].(map[string]any)
	if !ok {
		return fmt.Errorf("item 0 must be an object")
	}
	if err := requireExactObjectFields(attribute, "LookupAttribute", "AttributeKey", "AttributeValue"); err != nil {
		return err
	}
	key, ok := actionString(attribute["AttributeKey"])
	if !ok {
		return fmt.Errorf("LookupAttribute.AttributeKey must be a string")
	}
	allowedKeys := map[string]struct{}{
		"EventId": {}, "EventName": {}, "ReadOnly": {}, "Username": {}, "ResourceType": {}, "ResourceName": {}, "EventSource": {}, "AccessKeyId": {},
	}
	if _, ok := allowedKeys[key]; !ok {
		return fmt.Errorf("LookupAttribute.AttributeKey must be one of EventId | EventName | ReadOnly | Username | ResourceType | ResourceName | EventSource | AccessKeyId")
	}
	attributeValue, ok := actionString(attribute["AttributeValue"])
	if !ok || strings.TrimSpace(attributeValue) == "" {
		return fmt.Errorf("LookupAttribute.AttributeValue must be a non-empty string")
	}
	if length := lookupAttributeValueLength(attributeValue); length > 2000 {
		return fmt.Errorf("LookupAttribute.AttributeValue must have length at most 2000")
	}
	return nil
}

func lookupAttributeValueLength(value string) int {
	length := 0
	for _, character := range value {
		if character == '_' || character == ' ' || character == ',' || character == '\n' {
			length += 2
			continue
		}
		length++
	}
	return length
}

func validateImportSource(value any) error {
	source, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("must be an object")
	}
	if err := requireExactObjectFields(source, "ImportSource", "S3"); err != nil {
		return err
	}
	s3, ok := source["S3"].(map[string]any)
	if !ok {
		return fmt.Errorf("ImportSource.S3 must be an object")
	}
	if err := requireExactObjectFields(s3, "ImportSource.S3", "S3BucketAccessRoleArn", "S3BucketRegion", "S3LocationUri"); err != nil {
		return err
	}
	for _, field := range []string{"S3BucketAccessRoleArn", "S3BucketRegion", "S3LocationUri"} {
		value, ok := actionString(s3[field])
		if !ok || strings.TrimSpace(value) == "" {
			return fmt.Errorf("ImportSource.S3.%s must be a non-empty string", field)
		}
	}
	return nil
}

func requireExactObjectFields(object map[string]any, objectName string, required ...string) error {
	allowed := make(map[string]struct{}, len(required))
	for _, field := range required {
		allowed[field] = struct{}{}
		value, ok := object[field]
		if !ok || !actionValuePresent(value) {
			return fmt.Errorf("%s requires field %s", objectName, field)
		}
	}
	for field := range object {
		if _, ok := allowed[field]; !ok {
			return fmt.Errorf("%s field %q is not in the official request schema", objectName, field)
		}
	}
	return nil
}

func actionString(value any) (string, bool) {
	stringValue, ok := value.(string)
	return stringValue, ok
}
