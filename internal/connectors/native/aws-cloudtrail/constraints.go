package awscloudtrail

import (
	"fmt"
	"net"
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
	cloudTrailActionFieldConstraints         = buildActionFieldConstraints(cloudTrailActionFields)
	cloudTrailAdditionalRequiredActionFields = map[string]map[string]bool{"CancelQuery": {"EventDataStore": true}}
	cloudTrailActionCrossFields              = map[string]awsActionCrossFields{
		"DescribeQuery":         {anyOf: []string{"QueryId", "QueryAlias"}},
		"GetEventConfiguration": {anyOf: []string{"EventDataStore", "TrailName"}},
		"GetInsightSelectors":   {exactlyOneOf: [][]string{{"EventDataStore"}, {"TrailName"}}},
	}
	integerRangePattern                      = regexp.MustCompile(`(?i)Valid Range:\s*Minimum value of ([0-9]+)\.\s*Maximum value of ([0-9]+)\.`)
	fixedItemsPattern                        = regexp.MustCompile(`(?i)(?:Array Members|Map Entries):\s*Fixed number of ([0-9]+) items?\.`)
	minimumItemsPattern                      = regexp.MustCompile(`(?i)(?:Array Members|Map Entries):\s*Minimum number of ([0-9]+) items?\.`)
	maximumItemsPattern                      = regexp.MustCompile(`(?i)(?:Array Members|Map Entries):.*?Maximum number of ([0-9]+) items?\.`)
	fixedLengthPattern                       = regexp.MustCompile(`(?i)Fixed length of ([0-9]+)\.`)
	minimumLengthPattern                     = regexp.MustCompile(`(?i)Minimum length of ([0-9]+)\.`)
	maximumLengthPattern                     = regexp.MustCompile(`(?i)Maximum length of ([0-9]+)\.`)
	cloudTrailDestinationLocationConstraints = awsStringConstraints{minimumLength: 3, maximumLength: 1024, pattern: regexp.MustCompile(`^[a-zA-Z0-9._/\-:*]+$`)}
	cloudTrailDestinationTypeConstraints     = awsStringConstraints{values: map[string]struct{}{"EVENT_DATA_STORE": {}, "AWS_SERVICE": {}}, valueList: "EVENT_DATA_STORE | AWS_SERVICE"}
	cloudTrailTagKeyConstraints              = awsStringConstraints{minimumLength: 1, maximumLength: 128}
	cloudTrailTagValueConstraints            = awsStringConstraints{minimumLength: 1, maximumLength: 256}
	cloudTrailInsightTypeConstraints         = awsStringConstraints{values: map[string]struct{}{"ApiCallRateInsight": {}, "ApiErrorRateInsight": {}}, valueList: "ApiCallRateInsight | ApiErrorRateInsight"}
	cloudTrailInsightCategoryConstraints     = awsStringConstraints{values: map[string]struct{}{"Management": {}, "Data": {}}, valueList: "Management | Data"}
	cloudTrailReadWriteTypeConstraints       = awsStringConstraints{values: map[string]struct{}{"ReadOnly": {}, "WriteOnly": {}, "All": {}}, valueList: "ReadOnly | WriteOnly | All"}
	cloudTrailDataResourceTypeConstraints    = awsStringConstraints{values: map[string]struct{}{"AWS::S3::Object": {}, "AWS::Lambda::Function": {}, "AWS::DynamoDB::Table": {}}, valueList: "AWS::S3::Object | AWS::Lambda::Function | AWS::DynamoDB::Table"}
	cloudTrailManagementSourceConstraints    = awsStringConstraints{values: map[string]struct{}{"kms.amazonaws.com": {}, "rdsdata.amazonaws.com": {}}, valueList: "kms.amazonaws.com | rdsdata.amazonaws.com"}
	cloudTrailAggregationCategoryConstraints = awsStringConstraints{values: map[string]struct{}{"Data": {}}, valueList: "Data"}
	cloudTrailAggregationTemplateConstraints = awsStringConstraints{values: map[string]struct{}{"API_ACTIVITY": {}, "RESOURCE_ACCESS": {}, "USER_ACTIONS": {}}, valueList: "API_ACTIVITY | RESOURCE_ACCESS | USER_ACTIONS"}
	cloudTrailContextKeyTypeConstraints      = awsStringConstraints{values: map[string]struct{}{"TagContext": {}, "RequestContext": {}}, valueList: "TagContext | RequestContext"}
	cloudTrailTrailNamePattern               = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	cloudTrailARNPartitionPattern            = regexp.MustCompile(`^aws(?:-[a-z0-9-]+)?$`)
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
	case (action == "CreateChannel" || action == "UpdateChannel") && field.Name == "Destinations":
		return validateDestinations(value)
	case field.Name == "TagsList":
		return validateTagsList(value)
	case field.Name == "AdvancedEventSelectors":
		return validateAdvancedEventSelectors(value)
	case action == "PutEventSelectors" && field.Name == "EventSelectors":
		return validateEventSelectors(value)
	case action == "PutInsightSelectors" && field.Name == "InsightSelectors":
		return validateInsightSelectors(value)
	case action == "PutEventConfiguration" && field.Name == "AggregationConfigurations":
		return validateAggregationConfigurations(value)
	case action == "PutEventConfiguration" && field.Name == "ContextKeySelectors":
		return validateContextKeySelectors(value)
	case action == "CreateTrail" && field.Name == "Name":
		return validateTrailName(value, false)
	case action == "UpdateTrail" && field.Name == "Name":
		return validateTrailName(value, true)
	case (action == "CreateTrail" || action == "UpdateTrail") && (field.Name == "S3KeyPrefix" || field.Name == "SnsTopicName"):
		return validateTrailOptionalString(field.Name, value)
	}
	return nil
}

func requiredActionField(action string, field awsActionField) bool {
	return field.Required || cloudTrailAdditionalRequiredActionFields[action][field.Name]
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
	case "CreateTrail", "UpdateTrail":
		return validateTrailCloudWatchLogs(action, body)
	case "StartImport":
		return validateStartImport(body)
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
	case "PutEventSelectors":
		return validatePutEventSelectors(body)
	}
	return nil
}

func validateTrailCloudWatchLogs(action string, body map[string]any) error {
	if hasActionField(body, []string{"CloudWatchLogsRoleArn"}) && !hasActionField(body, []string{"CloudWatchLogsLogGroupArn"}) {
		return fmt.Errorf("aws-cloudtrail %s requires CloudWatchLogsLogGroupArn with CloudWatchLogsRoleArn", action)
	}
	return nil
}

func validateTrailName(value any, allowARN bool) error {
	name, ok := actionString(value)
	if !ok {
		return fmt.Errorf("must be a string")
	}
	if allowARN && validTrailARN(name) {
		return nil
	}
	return validateClassicTrailName(name)
}

func validateClassicTrailName(name string) error {
	if length := utf8.RuneCountInString(name); length < 3 || length > 128 {
		return fmt.Errorf("must have length from 3 to 128")
	}
	if !cloudTrailTrailNamePattern.MatchString(name) {
		return fmt.Errorf("does not match documented trail name pattern")
	}
	characters := []rune(name)
	if !trailNameAlphaNumeric(characters[0]) || !trailNameAlphaNumeric(characters[len(characters)-1]) {
		return fmt.Errorf("must start and end with an alphanumeric character")
	}
	for index := 1; index < len(characters); index++ {
		if trailNamePunctuation(characters[index-1]) && trailNamePunctuation(characters[index]) {
			return fmt.Errorf("cannot contain adjacent punctuation")
		}
	}
	if parsed := net.ParseIP(name); parsed != nil && parsed.To4() != nil {
		return fmt.Errorf("cannot be an IP address")
	}
	return nil
}

func validTrailARN(value string) bool {
	parts := strings.SplitN(value, ":", 6)
	if len(parts) != 6 || parts[0] != "arn" || !cloudTrailARNPartitionPattern.MatchString(parts[1]) || parts[2] != "cloudtrail" || strings.TrimSpace(parts[3]) == "" || len(parts[4]) != 12 {
		return false
	}
	for _, character := range parts[4] {
		if character < '0' || character > '9' {
			return false
		}
	}
	trailName, ok := strings.CutPrefix(parts[5], "trail/")
	if !ok {
		return false
	}
	return validateClassicTrailName(trailName) == nil
}

func trailNameAlphaNumeric(character rune) bool {
	return character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9'
}

func trailNamePunctuation(character rune) bool {
	return character == '.' || character == '_' || character == '-'
}

func validateTrailOptionalString(field string, value any) error {
	stringValue, ok := actionString(value)
	if !ok {
		return fmt.Errorf("must be a string")
	}
	maximumLength := 200
	if field == "SnsTopicName" {
		maximumLength = 256
	}
	if utf8.RuneCountInString(stringValue) > maximumLength {
		return fmt.Errorf("must have length at most %d", maximumLength)
	}
	return nil
}

func validateStartImport(body map[string]any) error {
	if hasActionField(body, []string{"ImportId"}) {
		for _, field := range []string{"Destinations", "ImportSource", "StartEventTime", "EndEventTime"} {
			if hasActionField(body, []string{field}) {
				return fmt.Errorf("aws-cloudtrail StartImport cannot combine ImportId with %s", field)
			}
		}
		return nil
	}
	if !allActionFieldsPresent(body, []string{"Destinations", "ImportSource"}) {
		return fmt.Errorf("aws-cloudtrail StartImport requires ImportId or Destinations and ImportSource")
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

func validatePutEventSelectors(body map[string]any) error {
	if hasActionField(body, []string{"AdvancedEventSelectors"}) && hasActionField(body, []string{"EventSelectors"}) {
		return fmt.Errorf("aws-cloudtrail PutEventSelectors cannot combine AdvancedEventSelectors with EventSelectors")
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

func validateDestinations(value any) error {
	destinations, ok := actionArray(value)
	if !ok {
		return fmt.Errorf("want array")
	}
	for index, item := range destinations {
		destination, ok := item.(map[string]any)
		if !ok {
			return fmt.Errorf("item %d must be an object", index)
		}
		if err := requireExactObjectFields(destination, "Destination", "Location", "Type"); err != nil {
			return fmt.Errorf("item %d: %w", index, err)
		}
		location, ok := actionString(destination["Location"])
		if !ok {
			return fmt.Errorf("item %d: Destination.Location must be a string", index)
		}
		if err := validateActionString(location, cloudTrailDestinationLocationConstraints); err != nil {
			return fmt.Errorf("item %d: Destination.Location: %w", index, err)
		}
		destinationType, ok := actionString(destination["Type"])
		if !ok {
			return fmt.Errorf("item %d: Destination.Type must be a string", index)
		}
		if err := validateActionString(destinationType, cloudTrailDestinationTypeConstraints); err != nil {
			return fmt.Errorf("item %d: Destination.Type: %w", index, err)
		}
	}
	return nil
}

func validateTagsList(value any) error {
	tags, ok := actionArray(value)
	if !ok {
		return fmt.Errorf("want array")
	}
	for index, item := range tags {
		tag, ok := item.(map[string]any)
		if !ok {
			return fmt.Errorf("item %d must be an object", index)
		}
		if err := validateClosedObjectFields(tag, "Tag", []string{"Key"}, []string{"Value"}); err != nil {
			return fmt.Errorf("item %d: %w", index, err)
		}
		key, ok := actionString(tag["Key"])
		if !ok {
			return fmt.Errorf("item %d: Tag.Key must be a string", index)
		}
		if err := validateActionString(key, cloudTrailTagKeyConstraints); err != nil {
			return fmt.Errorf("item %d: Tag.Key: %w", index, err)
		}
		if rawValue, present := tag["Value"]; present {
			stringValue, ok := actionString(rawValue)
			if !ok {
				return fmt.Errorf("item %d: Tag.Value must be a string", index)
			}
			if err := validateActionString(stringValue, cloudTrailTagValueConstraints); err != nil {
				return fmt.Errorf("item %d: Tag.Value: %w", index, err)
			}
		}
	}
	return nil
}

func validateInsightSelectors(value any) error {
	selectors, ok := actionArray(value)
	if !ok {
		return fmt.Errorf("want array")
	}
	for index, item := range selectors {
		selector, ok := item.(map[string]any)
		if !ok {
			return fmt.Errorf("item %d must be an object", index)
		}
		if err := validateClosedObjectFields(selector, "InsightSelector", []string{"InsightType"}, []string{"EventCategories"}); err != nil {
			return fmt.Errorf("item %d: %w", index, err)
		}
		insightType, ok := actionString(selector["InsightType"])
		if !ok {
			return fmt.Errorf("item %d: InsightSelector.InsightType must be a string", index)
		}
		if err := validateActionString(insightType, cloudTrailInsightTypeConstraints); err != nil {
			return fmt.Errorf("item %d: InsightSelector.InsightType: %w", index, err)
		}
		if categories, present := selector["EventCategories"]; present {
			if err := validateClosedStringArray(categories, "InsightSelector.EventCategories", 0, 0, cloudTrailInsightCategoryConstraints); err != nil {
				return fmt.Errorf("item %d: %w", index, err)
			}
		}
	}
	return nil
}

func validateEventSelectors(value any) error {
	selectors, ok := actionArray(value)
	if !ok {
		return fmt.Errorf("want array")
	}
	if err := validateNestedItemCount(len(selectors), 1, 5); err != nil {
		return err
	}
	dataResourceValues := 0
	for index, item := range selectors {
		selector, ok := item.(map[string]any)
		if !ok {
			return fmt.Errorf("item %d must be an object", index)
		}
		if err := validateClosedObjectFields(selector, "EventSelector", nil, []string{"DataResources", "ExcludeManagementEventSources", "IncludeManagementEvents", "ReadWriteType"}); err != nil {
			return fmt.Errorf("item %d: %w", index, err)
		}
		if readWriteType, present := selector["ReadWriteType"]; present {
			stringValue, ok := actionString(readWriteType)
			if !ok {
				return fmt.Errorf("item %d: EventSelector.ReadWriteType must be a string", index)
			}
			if err := validateActionString(stringValue, cloudTrailReadWriteTypeConstraints); err != nil {
				return fmt.Errorf("item %d: EventSelector.ReadWriteType: %w", index, err)
			}
		}
		if includeManagementEvents, present := selector["IncludeManagementEvents"]; present {
			if _, ok := includeManagementEvents.(bool); !ok {
				return fmt.Errorf("item %d: EventSelector.IncludeManagementEvents must be a boolean", index)
			}
		}
		if excludedSources, present := selector["ExcludeManagementEventSources"]; present {
			if err := validateClosedStringArray(excludedSources, "EventSelector.ExcludeManagementEventSources", 1, 0, cloudTrailManagementSourceConstraints); err != nil {
				return fmt.Errorf("item %d: %w", index, err)
			}
		}
		if rawResources, present := selector["DataResources"]; present {
			resources, ok := actionArray(rawResources)
			if !ok {
				return fmt.Errorf("item %d: EventSelector.DataResources must be an array", index)
			}
			if err := validateNestedItemCount(len(resources), 1, 250); err != nil {
				return fmt.Errorf("item %d: EventSelector.DataResources: %w", index, err)
			}
			for resourceIndex, rawResource := range resources {
				resource, ok := rawResource.(map[string]any)
				if !ok {
					return fmt.Errorf("item %d: EventSelector.DataResources item %d must be an object", index, resourceIndex)
				}
				if err := requireExactObjectFields(resource, "DataResource", "Type", "Values"); err != nil {
					return fmt.Errorf("item %d: EventSelector.DataResources item %d: %w", index, resourceIndex, err)
				}
				resourceType, ok := actionString(resource["Type"])
				if !ok {
					return fmt.Errorf("item %d: EventSelector.DataResources item %d: DataResource.Type must be a string", index, resourceIndex)
				}
				if err := validateActionString(resourceType, cloudTrailDataResourceTypeConstraints); err != nil {
					return fmt.Errorf("item %d: EventSelector.DataResources item %d: DataResource.Type: %w", index, resourceIndex, err)
				}
				values, ok := actionArray(resource["Values"])
				if !ok {
					return fmt.Errorf("item %d: EventSelector.DataResources item %d: DataResource.Values must be an array", index, resourceIndex)
				}
				if err := validateClosedStringArray(resource["Values"], "DataResource.Values", 1, 250, awsStringConstraints{minimumLength: 1}); err != nil {
					return fmt.Errorf("item %d: EventSelector.DataResources item %d: %w", index, resourceIndex, err)
				}
				dataResourceValues += len(values)
			}
		}
	}
	if dataResourceValues > 250 {
		return fmt.Errorf("EventSelector.DataResources must contain at most 250 values")
	}
	return nil
}

func validateAdvancedEventSelectors(value any) error {
	selectors, ok := actionArray(value)
	if !ok {
		return fmt.Errorf("want array")
	}
	conditionValues := 0
	for index, item := range selectors {
		selector, ok := item.(map[string]any)
		if !ok {
			return fmt.Errorf("item %d must be an object", index)
		}
		if err := validateClosedObjectFields(selector, "AdvancedEventSelector", []string{"FieldSelectors"}, []string{"Name"}); err != nil {
			return fmt.Errorf("item %d: %w", index, err)
		}
		if name, present := selector["Name"]; present {
			stringValue, ok := actionString(name)
			if !ok {
				return fmt.Errorf("item %d: AdvancedEventSelector.Name must be a string", index)
			}
			if err := validateActionString(stringValue, awsStringConstraints{minimumLength: 1, maximumLength: 1000}); err != nil {
				return fmt.Errorf("item %d: AdvancedEventSelector.Name: %w", index, err)
			}
		}
		fields, ok := actionArray(selector["FieldSelectors"])
		if !ok {
			return fmt.Errorf("item %d: AdvancedEventSelector.FieldSelectors must be an array", index)
		}
		if err := validateNestedItemCount(len(fields), 1, 500); err != nil {
			return fmt.Errorf("item %d: AdvancedEventSelector.FieldSelectors: %w", index, err)
		}
		for fieldIndex, item := range fields {
			fieldSelector, ok := item.(map[string]any)
			if !ok {
				return fmt.Errorf("item %d: AdvancedEventSelector.FieldSelectors item %d must be an object", index, fieldIndex)
			}
			if err := validateClosedObjectFields(fieldSelector, "AdvancedFieldSelector", []string{"Field"}, []string{"EndsWith", "Equals", "NotEndsWith", "NotEquals", "NotStartsWith", "StartsWith"}); err != nil {
				return fmt.Errorf("item %d: AdvancedEventSelector.FieldSelectors item %d: %w", index, fieldIndex, err)
			}
			fieldName, ok := actionString(fieldSelector["Field"])
			if !ok {
				return fmt.Errorf("item %d: AdvancedEventSelector.FieldSelectors item %d: AdvancedFieldSelector.Field must be a string", index, fieldIndex)
			}
			if err := validateActionString(fieldName, awsStringConstraints{minimumLength: 1, maximumLength: 1000}); err != nil {
				return fmt.Errorf("item %d: AdvancedEventSelector.FieldSelectors item %d: AdvancedFieldSelector.Field: %w", index, fieldIndex, err)
			}
			selectorsWithValues := 0
			for _, operator := range []string{"EndsWith", "Equals", "NotEndsWith", "NotEquals", "NotStartsWith", "StartsWith"} {
				rawValues, present := fieldSelector[operator]
				if !present {
					continue
				}
				values, ok := actionArray(rawValues)
				if !ok {
					return fmt.Errorf("item %d: AdvancedEventSelector.FieldSelectors item %d: AdvancedFieldSelector.%s must be an array", index, fieldIndex, operator)
				}
				if err := validateClosedStringArray(rawValues, "AdvancedFieldSelector."+operator, 1, 500, awsStringConstraints{minimumLength: 1, maximumLength: 2048}); err != nil {
					return fmt.Errorf("item %d: AdvancedEventSelector.FieldSelectors item %d: %w", index, fieldIndex, err)
				}
				selectorsWithValues += len(values)
				conditionValues += len(values)
			}
			if selectorsWithValues == 0 {
				return fmt.Errorf("item %d: AdvancedEventSelector.FieldSelectors item %d: AdvancedFieldSelector requires at least one condition", index, fieldIndex)
			}
		}
	}
	if conditionValues > 500 {
		return fmt.Errorf("AdvancedEventSelectors must contain at most 500 condition values")
	}
	return nil
}

func validateAggregationConfigurations(value any) error {
	configurations, ok := actionArray(value)
	if !ok {
		return fmt.Errorf("want array")
	}
	if err := validateNestedItemCount(len(configurations), 1, 1); err != nil {
		return err
	}
	for index, item := range configurations {
		configuration, ok := item.(map[string]any)
		if !ok {
			return fmt.Errorf("item %d must be an object", index)
		}
		if err := requireExactObjectFields(configuration, "AggregationConfiguration", "EventCategory", "Templates"); err != nil {
			return fmt.Errorf("item %d: %w", index, err)
		}
		category, ok := actionString(configuration["EventCategory"])
		if !ok {
			return fmt.Errorf("item %d: AggregationConfiguration.EventCategory must be a string", index)
		}
		if err := validateActionString(category, cloudTrailAggregationCategoryConstraints); err != nil {
			return fmt.Errorf("item %d: AggregationConfiguration.EventCategory: %w", index, err)
		}
		if err := validateClosedStringArray(configuration["Templates"], "AggregationConfiguration.Templates", 1, 50, cloudTrailAggregationTemplateConstraints); err != nil {
			return fmt.Errorf("item %d: %w", index, err)
		}
	}
	return nil
}

func validateContextKeySelectors(value any) error {
	selectors, ok := actionArray(value)
	if !ok {
		return fmt.Errorf("want array")
	}
	if err := validateNestedItemCount(len(selectors), 1, 2); err != nil {
		return err
	}
	for index, item := range selectors {
		selector, ok := item.(map[string]any)
		if !ok {
			return fmt.Errorf("item %d must be an object", index)
		}
		if err := requireExactObjectFields(selector, "ContextKeySelector", "Equals", "Type"); err != nil {
			return fmt.Errorf("item %d: %w", index, err)
		}
		selectorType, ok := actionString(selector["Type"])
		if !ok {
			return fmt.Errorf("item %d: ContextKeySelector.Type must be a string", index)
		}
		if err := validateActionString(selectorType, cloudTrailContextKeyTypeConstraints); err != nil {
			return fmt.Errorf("item %d: ContextKeySelector.Type: %w", index, err)
		}
		if err := validateClosedStringArray(selector["Equals"], "ContextKeySelector.Equals", 1, 50, awsStringConstraints{minimumLength: 1, maximumLength: 128}); err != nil {
			return fmt.Errorf("item %d: %w", index, err)
		}
	}
	return nil
}

func validateClosedStringArray(value any, name string, minimumItems, maximumItems int, constraints awsStringConstraints) error {
	items, ok := actionArray(value)
	if !ok {
		return fmt.Errorf("%s must be an array", name)
	}
	if err := validateNestedItemCount(len(items), minimumItems, maximumItems); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	for index, item := range items {
		stringValue, ok := actionString(item)
		if !ok {
			return fmt.Errorf("%s item %d must be a string", name, index)
		}
		if err := validateActionString(stringValue, constraints); err != nil {
			return fmt.Errorf("%s item %d: %w", name, index, err)
		}
	}
	return nil
}

func validateNestedItemCount(length, minimumItems, maximumItems int) error {
	if minimumItems > 0 && length < minimumItems {
		return fmt.Errorf("must contain at least %d items", minimumItems)
	}
	if maximumItems > 0 && length > maximumItems {
		return fmt.Errorf("must contain at most %d items", maximumItems)
	}
	return nil
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
	return validateClosedObjectFields(object, objectName, required, nil)
}

func validateClosedObjectFields(object map[string]any, objectName string, required, optional []string) error {
	allowed := make(map[string]struct{}, len(required)+len(optional))
	for _, field := range required {
		allowed[field] = struct{}{}
		value, ok := object[field]
		if !ok || !actionValuePresent(value) {
			return fmt.Errorf("%s requires field %s", objectName, field)
		}
	}
	for _, field := range optional {
		allowed[field] = struct{}{}
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
