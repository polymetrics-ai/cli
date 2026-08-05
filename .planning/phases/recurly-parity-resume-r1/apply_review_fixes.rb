#!/usr/bin/env ruby

require "digest"
require "fileutils"
require "json"
require "open-uri"
require "uri"
require "yaml"

OAS_URL = "https://recurly.com/developers/api/spec/v2021-02-25.yaml"
OAS_SHA256 = "b98a3f85d0a1190c2c8e11f57fa5ec13b841665e658596dcb5d7f3ddce70baca"
LEGACY_STREAMS = %w[accounts invoices plans subscriptions transactions].freeze
SENSITIVE_FIELDS = %w[
  account_number adyen_risk_profile_reference_id amazon_billing_agreement_id
  cvv fraud_session_id gateway_attributes gateway_code gateway_token iban
  network_transaction_id number paypal_billing_agreement_id
  roku_billing_agreement_id routing_number sort_code tax_identifier
  tax_identifier_type three_d_secure_action_result_token_id token token_id
  verification_value
].freeze
REMOVED_GENERATED_CLI_MAPPINGS = {
  "update_subscription" => ["record.custom_fields.0.name"]
}.freeze
CALLER_DECISION_QUERY_FIELDS = {
  "deactivate_account" => %w[redact],
  "terminate_subscription" => %w[refund charge]
}.freeze
MUTATION_QUERY_FIXTURE_VALUES = {
  "deactivate_account" => {"redact" => true},
  "terminate_subscription" => {"refund" => "partial", "charge" => false}
}.freeze

WORKTREE = File.expand_path("../../..", __dir__)
BUNDLE = File.join(WORKTREE, "internal/connectors/defs/recurly")
RETRY_RESEARCH_PATH = File.join(__dir__, "RECURLY-WRITE-RETRY-RESEARCH.json")

def read_json(path)
  JSON.parse(File.read(path))
end

def write_json(path, value)
  FileUtils.mkdir_p(File.dirname(path))
  File.write(path, JSON.pretty_generate(value) + "\n")
end

oas_raw = URI.open(OAS_URL, &:read)
actual_sha = Digest::SHA256.hexdigest(oas_raw)
abort "unexpected Recurly OAS digest #{actual_sha}" unless actual_sha == OAS_SHA256
OAS = YAML.safe_load(oas_raw, aliases: true)
SCHEMAS = OAS.fetch("components").fetch("schemas")
IDEMPOTENCY_EVIDENCE_PATH = "info.description[section=Idempotent Requests]"
idempotency_description = OAS.dig("info", "description").to_s
abort "Recurly OAS no longer documents keyed DELETE idempotency" unless
  idempotency_description.include?("Idempotent Requests") &&
  idempotency_description.match?(/Idempotency-Key.*POST.*PUT.*PATCH.*DELETE/m)

def deep_copy(value)
  Marshal.load(Marshal.dump(value))
end

def local_ref(value, section)
  return value unless value.is_a?(Hash) && value["$ref"]

  prefix = "#/components/#{section}/"
  ref = value.fetch("$ref")
  abort "unsupported OAS reference #{ref}" unless ref.start_with?(prefix)
  resolved = ref.delete_prefix("#/").split("/").reduce(OAS) do |current, segment|
    current&.[](segment.gsub("~1", "/").gsub("~0", "~"))
  end
  resolved || abort("missing OAS reference #{ref}")
end

def schema_ref(value)
  local_ref(value, "schemas")
end

def schema_annotation(value, name, seen = [])
  return nil unless value.is_a?(Hash)
  if value["$ref"]
    ref = value.fetch("$ref")
    return nil if seen.include?(ref)
    return schema_annotation(schema_ref(value), name, seen + [ref])
  end
  return value[name] if value.key?(name)
  Array(value["allOf"]).each do |part|
    annotation = schema_annotation(part, name, seen)
    return annotation unless annotation.nil?
  end
  nil
end

def merge_schemas(left, right)
  result = deep_copy(left)
  right.each do |key, value|
    case key
    when "properties"
      result[key] = (result[key] || {}).merge(deep_copy(value))
    when "required"
      result[key] = (Array(result[key]) + Array(value)).uniq
    when "type"
      result[key] ||= deep_copy(value)
    when "description", "title", "format", "default", "pattern", "minProperties", "minItems", "maxItems"
      result[key] = deep_copy(value) unless result.key?(key)
    when "additionalProperties"
      result[key] = false if result[key] == false || value == false
      result[key] = true unless result.key?(key)
    else
      result[key] = deep_copy(value)
    end
  end
  result
end

def normalize_schema(value, mode, stack = [])
  abort "schema must be an object: #{value.inspect}" unless value.is_a?(Hash)

  if value["$ref"]
    ref = value.fetch("$ref")
    abort "cyclic Recurly schema reference #{ref}" if stack.include?(ref)
    return normalize_schema(schema_ref(value), mode, stack + [ref])
  end

  if value["allOf"]
    merged = {}
    value.fetch("allOf").each do |part|
      merged = merge_schemas(merged, normalize_schema(part, mode, stack))
    end
    siblings = value.reject { |key, _| key == "allOf" }
    merged = merge_schemas(merged, normalize_schema(siblings, mode, stack)) unless siblings.empty?
    if merged["properties"] && merged["required"]
      merged["required"] = Array(merged["required"]).select { |name| merged.fetch("properties").key?(name) }
      merged.delete("required") if merged["required"].empty?
    end
    return merged
  end

  result = {}
  type = value["type"]
  type = "object" if type.nil? && value["properties"]
  type = "array" if type.nil? && value["items"]
  if value["nullable"]
    types = Array(type).compact
    types << "null"
    type = types.uniq
  end
  result["type"] = deep_copy(type) unless type.nil?

  if value["properties"]
    properties = {}
    value.fetch("properties").each do |name, property|
      next if mode == :request && schema_annotation(property, "readOnly") == true
      next if mode == :response && schema_annotation(property, "writeOnly") == true
      properties[name] = normalize_schema(property, mode, stack)
    end
    result["properties"] = properties
    required = Array(value["required"]).select { |name| properties.key?(name) }
    result["required"] = required unless required.empty?
    additional = value["additionalProperties"]
    result["additionalProperties"] = if additional.is_a?(Hash)
                                       normalize_schema(additional, mode, stack)
                                     else
                                       additional == true
                                     end
  elsif type == "object" || Array(type).include?("object")
    additional = value["additionalProperties"]
    result["additionalProperties"] = if additional.is_a?(Hash)
                                       normalize_schema(additional, mode, stack)
                                     else
                                       additional != false
                                     end
    required = Array(value["required"])
    result["required"] = required unless required.empty?
  end

  result["items"] = normalize_schema(value.fetch("items"), mode, stack) if value["items"]
  %w[enum minProperties minItems maxItems format default title description].each do |keyword|
    result[keyword] = deep_copy(value[keyword]) if value.key?(keyword)
  end
  if value["pattern"]
    pattern = value.fetch("pattern")
    if pattern.start_with?("/") && pattern.match?(%r{/i?\z})
      insensitive = pattern.end_with?("/i")
      pattern = pattern.sub(%r{\A/}, "").sub(%r{/i?\z}, "")
      pattern = "(?i)#{pattern}" if insensitive
    end
    result["pattern"] = pattern
  end
  if value["nullable"] && result["enum"] && !result["enum"].include?(nil)
    result["enum"] << nil
  end
  result
end

def operation_for(endpoint)
  OAS.dig("paths", endpoint.fetch("path"), endpoint.fetch("method").downcase) ||
    abort("missing OAS operation #{endpoint.fetch('method')} #{endpoint.fetch('path')}")
end

def request_body_for(operation)
  body = operation["requestBody"]
  return nil unless body
  body = local_ref(body, "requestBodies")
  schema = body.dig("content", "application/json", "schema")
  abort "Recurly request body lacks application/json schema" unless schema
  [body, schema]
end

def response_for(operation)
  status, response = operation.fetch("responses").sort_by { |code, _| code.to_i }.find do |code, candidate|
    code.to_s.start_with?("2") && begin
      resolved = local_ref(candidate, "responses")
      resolved.dig("content", "application/json", "schema")
    end
  end
  return [200, nil] unless response
  response = local_ref(response, "responses")
  [status.to_i, response.dig("content", "application/json", "schema")]
end

def operation_parameters(path, operation)
  values = Array(OAS.dig("paths", path, "parameters")) + Array(operation["parameters"])
  values.map { |parameter| local_ref(parameter, "parameters") }
end

def parameter_schema(path, operation, location, name)
  parameter = operation_parameters(path, operation).find do |candidate|
    candidate["in"] == location && candidate["name"] == name
  end
  abort "missing OAS #{location} parameter #{name} for #{path}" unless parameter
  normalize_schema(parameter.fetch("schema"), :request)
end

def query_parameters(path, operation)
  operation_parameters(path, operation).select { |parameter| parameter["in"] == "query" }
end

def operation_documents_intrinsic_idempotency?(operation)
  [operation["summary"], operation["description"]].compact.join("\n").match?(/\bidempoten/i)
end

def query_fixture_value(value)
  value.to_s
end

def schema_types(schema)
  Array(schema["type"]).compact
end

def object_schema?(schema)
  types = schema_types(schema)
  return types.include?("object") unless types.empty?
  schema.key?("properties")
end

def array_schema?(schema)
  types = schema_types(schema)
  return types.include?("array") unless types.empty?
  schema.key?("items")
end

def required_mapping_paths(schema, prefix = "")
  Array(schema["required"]).flat_map do |name|
    child = schema.fetch("properties").fetch(name)
    path = prefix.empty? ? name : "#{prefix}.#{name}"
    descendants = required_node_paths(child, path)
    descendants.empty? ? [path] : descendants
  end
end

def required_node_paths(schema, prefix)
  if array_schema?(schema)
    return [] unless schema["items"]
    descendants = required_node_paths(schema.fetch("items"), "#{prefix}.0")
    return descendants.empty? ? [prefix] : descendants
  end
  return required_mapping_paths(schema, prefix) if object_schema?(schema)
  []
end

def leaf_paths(schema, prefix = "")
  if object_schema?(schema)
    return schema.fetch("properties", {}).flat_map do |name, child|
      child_prefix = prefix.empty? ? name : "#{prefix}.#{name}"
      leaf_paths(child, child_prefix)
    end
  end
  if array_schema?(schema)
    items = schema["items"]
    return [] unless items
    if object_schema?(items) || array_schema?(items)
      return leaf_paths(items, "#{prefix}.0")
    end
    return [[prefix, schema]] if schema_types(items).include?("string")
    return []
  end
  [[prefix, schema]]
end

def schema_at_path(schema, path)
  current = schema
  path.split(".").each do |segment|
    if array_schema?(current)
      return nil unless segment.match?(/\A\d+\z/)
      current = current["items"]
    else
      return nil unless object_schema?(current)
      current = current.fetch("properties", {})[segment]
    end
    return nil unless current
  end
  current
end

def flag_type(schema)
  if array_schema?(schema)
    return "string_array" if schema["items"] && schema_types(schema.fetch("items")).include?("string")
    return nil
  end
  types = schema_types(schema) - ["null"]
  type = types.first
  return "enum" if type == "string" && Array(schema["enum"]).compact.any?
  return type if %w[string integer number boolean].include?(type)
  nil
end

def flag_name(path)
  path.tr("_", "-").tr(".", "-")
end

def sample_string(name, schema)
  value = case schema["format"]
          when "date-time" then "2026-01-02T03:04:05Z"
          when "date" then "2026-01-02"
          when "email" then "fixture@example.com"
          when "decimal" then "12.34"
          else case name
          when /harmonized_system_code/ then "1"
          when /currency/ then "USD"
          when /country/ then "US"
          when /email/ then "fixture@example.com"
          when /date$|_date$/ then "2026-01-02"
          when /_at$|timestamp/ then "2026-01-02T03:04:05Z"
          when /postal/ then "94105"
          when /phone/ then "+14155550100"
          when /month/ then "2026-01"
          when /year/ then "2026"
          else "fixture_#{name.gsub(/[^A-Za-z0-9_]/, '_')}"
          end
          end
  max_length = schema["maxLength"]
  value = value[0, max_length] if max_length && value.length > max_length
  value
end

def bounded_number(schema, candidate)
  candidate = schema["minimum"] if schema.key?("minimum") && candidate < schema["minimum"]
  candidate = schema["maximum"] if schema.key?("maximum") && candidate > schema["maximum"]
  candidate
end

def sample_value(schema, name, include_optional: false)
  enum_value = Array(schema["enum"]).find { |value| !value.nil? }
  return deep_copy(enum_value) unless enum_value.nil?
  return deep_copy(schema["default"]) if schema.key?("default") && !schema["default"].nil?

  types = schema_types(schema) - ["null"]
  type = types.first
  type ||= "object" if object_schema?(schema)
  type ||= "array" if array_schema?(schema)
  case type
  when "object"
    properties = schema.fetch("properties", {})
    names = Array(schema["required"])
    if include_optional
      preferred = %w[id object code name state created_at updated_at]
      names += preferred.select { |candidate| properties.key?(candidate) }
      names << properties.keys.first if names.empty? && properties.any?
    end
    names.uniq.to_h do |property|
      [property, sample_value(properties.fetch(property), property, include_optional: false)]
    end
  when "array"
    [sample_value(schema.fetch("items"), name.sub(/s\z/, ""), include_optional: true)]
  when "integer" then bounded_number(schema, 1).to_i
  when "number" then bounded_number(schema, 12.34)
  when "boolean" then false
  when "null" then nil
  else sample_string(name, schema)
  end
end

def set_nested(object, path, value)
  segments = path.split(".")
  current = object
  segments.each_with_index do |segment, index|
    last = index == segments.length - 1
    if segment.match?(/\A\d+\z/)
      abort "unexpected root array sample path #{path}" unless current.is_a?(Array)
      array_index = segment.to_i
      current[array_index] ||= last ? value : (segments[index + 1].match?(/\A\d+\z/) ? [] : {})
      current = current[array_index] unless last
    else
      abort "unexpected scalar sample path #{path}" unless current.is_a?(Hash)
      current[segment] ||= last ? value : (segments[index + 1]&.match?(/\A\d+\z/) ? [] : {})
      current = current[segment] unless last
    end
  end
end

def require_schema_path(schema, path)
  current = schema
  path.split(".").each do |segment|
    abort "derived key path #{path} crosses an array" if array_schema?(current)
    properties = current.fetch("properties")
    child = properties[segment] || abort("missing derived key schema path #{path}")
    current["required"] = (Array(current["required"]) + [segment]).uniq
    current = child
  end
end

def sensitive_paths(schema, prefix = "")
  return [] unless object_schema?(schema)
  schema.fetch("properties", {}).flat_map do |name, child|
    path = prefix.empty? ? name : "#{prefix}.#{name}"
    if SENSITIVE_FIELDS.include?(name)
      [path]
    elsif array_schema?(child) && child["items"] && object_schema?(child.fetch("items")) && sensitive_paths(child.fetch("items")).any?
      [path]
    elsif object_schema?(child)
      sensitive_paths(child, path)
    else
      []
    end
  end
end

def interpolate_fixture_path(path, record)
  path.gsub(/\{([^}]+)\}/) do
    name = Regexp.last_match(1)
    URI.encode_www_form_component(record.fetch(name).to_s).tr("+", "%20")
  end
end

api_surface = read_json(File.join(BUNDLE, "api_surface.json"))
endpoints = api_surface.fetch("endpoints")
write_endpoints = endpoints.filter_map do |endpoint|
  name = endpoint.dig("covered_by", "write")
  [name, endpoint] if name
end.to_h
stream_endpoints = endpoints.filter_map do |endpoint|
  name = endpoint.dig("covered_by", "stream")
  [name, endpoint] if name
end.to_h

writes_path = File.join(BUNDLE, "writes.json")
writes_document = read_json(writes_path)
delete_retry_evidence = []
mutation_query_evidence = []
writes_document.fetch("actions").each do |action|
  endpoint = write_endpoints[action.fetch("name")] || abort("missing surface endpoint for #{action.fetch('name')}")
  operation = operation_for(endpoint)
  operation_id = operation.fetch("operationId")
  body_info = request_body_for(operation)
  operation_query_parameters = query_parameters(endpoint.fetch("path"), operation)

  path_fields = Array(action["path_fields"])
  path_properties = path_fields.to_h do |field|
    [field, parameter_schema(endpoint.fetch("path"), operation, "path", field)]
  end
  query_properties = operation_query_parameters.to_h do |parameter|
    schema = normalize_schema(parameter.fetch("schema"), :request)
    schema["description"] = parameter["description"] if parameter["description"]
    [parameter.fetch("name"), schema]
  end
  properties = deep_copy(path_properties)
  required = path_fields.dup

  if body_info
    body, body_schema_ref = body_info
    body_schema = normalize_schema(body_schema_ref, :request)
    abort "request schema for #{action.fetch('name')} is not an object" unless object_schema?(body_schema)
    overlap = path_fields & body_schema.fetch("properties", {}).keys
    abort "request/path field collision for #{action.fetch('name')}: #{overlap.join(', ')}" unless overlap.empty?
    properties.merge!(deep_copy(body_schema.fetch("properties", {})))
    required.concat(Array(body_schema["required"]))
    action["body_type"] = "json"
    action["body_required"] = true if body["required"] == true
    action.delete("body_required") unless body["required"] == true
    action["body_fields"] = body_schema.fetch("properties", {}).keys
  else
    action["body_type"] = "none"
    action.delete("body_required")
    action.delete("body_fields")
  end

  overlap = properties.keys & query_properties.keys
  abort "request/query field collision for #{action.fetch('name')}: #{overlap.join(', ')}" unless overlap.empty?
  properties.merge!(deep_copy(query_properties))
  required.concat(operation_query_parameters.filter_map do |parameter|
    parameter.fetch("name") if parameter["required"] == true
  end)
  required.concat(Array(CALLER_DECISION_QUERY_FIELDS[action.fetch("name")]))

  if operation_query_parameters.empty?
    action.delete("query")
  else
    action["query"] = operation_query_parameters.to_h do |parameter|
      name = parameter.fetch("name")
      template = "{{ record.#{name} }}"
      value = required.include?(name) ? template : {"template" => template, "omit_when_absent" => true}
      [name, value]
    end
  end

  action["record_schema"] = {
    "$schema" => "http://json-schema.org/draft-07/schema#",
    "title" => "Recurly #{action.fetch('name')} record",
    "type" => "object",
    "properties" => properties,
    "required" => required.uniq,
    "additionalProperties" => false
  }
  action["idempotency_key_header"] = "Idempotency-Key"
  if action["kind"] == "delete"
    intrinsic_idempotency = operation_documents_intrinsic_idempotency?(operation)
    if intrinsic_idempotency
      action["delete"] ||= {}
      action["delete"]["idempotent"] = true
    else
      action.delete("delete")
    end
    operation_path = endpoint.fetch("path")
    operation_method = endpoint.fetch("method")
    delete_retry_evidence << {
      "action" => action.fetch("name"),
      "operation_id" => operation_id,
      "method" => operation_method,
      "path" => operation_path,
      "operation_source_url" => "https://recurly.com/developers/api/v2021-02-25/#operation/#{operation_id}",
      "operation_evidence_path" => "paths.#{operation_path}.#{operation_method.downcase}",
      "intrinsic_idempotency_documented" => intrinsic_idempotency,
      "retry_mode" => "provider_idempotency_key",
      "idempotency_key_header" => "Idempotency-Key",
      "idempotency_source_url" => OAS_URL,
      "idempotency_evidence_path" => IDEMPOTENCY_EVIDENCE_PATH
    }
  end
  combined_parameters = operation_parameters(endpoint.fetch("path"), operation)
  operation_query_parameters.each do |parameter|
    name = parameter.fetch("name")
    parameter_index = combined_parameters.index(parameter) || abort("missing query parameter index for #{action.fetch('name')}.#{name}")
    mutation_query_evidence << {
      "action" => action.fetch("name"),
      "operation_id" => operation_id,
      "method" => endpoint.fetch("method"),
      "path" => endpoint.fetch("path"),
      "field" => "query.#{name}",
      "local_source_path" => "writes.#{action.fetch('name')}.query.#{name}",
      "provider_required" => parameter["required"] == true,
      "local_required" => required.include?(name),
      "source_url" => "https://recurly.com/developers/api/v2021-02-25/#operation/#{operation_id}",
      "evidence_type" => "openapi.parameter",
      "evidence_path" => "paths.#{endpoint.fetch('path')}.#{endpoint.fetch('method').downcase}.parameters.#{parameter_index}",
      "schema" => deep_copy(query_properties.fetch(name))
    }
  end
  action["redact_fields"] = (Array(action["redact_fields"]) + sensitive_paths(action.fetch("record_schema"))).uniq.sort
  action.delete("redact_fields") if action["redact_fields"].empty?
  action["risk"] = action.fetch("risk").sub(
    "do not reuse idempotency keys across different records.",
    "the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries."
  )

  record = {}
  path_fields.each do |field|
    record[field] = sample_value(path_properties.fetch(field), field)
  end
  if body_info
    body_schema = normalize_schema(body_info.last, :request)
    body_sample = sample_value(body_schema, "body", include_optional: false)
    if body_sample.empty? && body_schema.fetch("properties", {}).any?
      preferred = %w[email first_name plan_code amount code name currency]
      selected = preferred.find { |name| body_schema.fetch("properties").key?(name) } || body_schema.fetch("properties").keys.first
      body_sample[selected] = sample_value(body_schema.fetch("properties").fetch(selected), selected, include_optional: true)
    end
    record.merge!(body_sample)
  end
  operation_query_parameters.each do |parameter|
    name = parameter.fetch("name")
    value = MUTATION_QUERY_FIXTURE_VALUES.dig(action.fetch("name"), name)
    value = sample_value(query_properties.fetch(name), name) if value.nil?
    record[name] = value
  end

  expected = {
    "method" => endpoint.fetch("method"),
    "path" => interpolate_fixture_path(endpoint.fetch("path"), record)
  }
  body_fields = Array(action["body_fields"])
  expected_body = record.select { |field, _| body_fields.include?(field) }
  expected["body"] = expected_body if body_info
  unless operation_query_parameters.empty?
    expected["query"] = operation_query_parameters.to_h do |parameter|
      name = parameter.fetch("name")
      [name, query_fixture_value(record.fetch(name))]
    end
  end

  response_status, response_schema_ref = response_for(operation)
  fixture_response = {"status" => response_status}
  if response_schema_ref
    response_schema = normalize_schema(response_schema_ref, :response)
    fixture_response["body"] = sample_value(response_schema, "response", include_optional: true)
  end
  fixture_response["body"] ||= {}
  fixture = {"record" => record, "expect" => expected, "response" => fixture_response}
  write_json(File.join(BUNDLE, "fixtures", "writes", "#{action.fetch('name')}.json"), fixture)
end
write_json(writes_path, writes_document)
write_json(RETRY_RESEARCH_PATH, {
  "purpose" => "Raw Recurly operation retry and mutation-query evidence; not final citation metadata until the shared citation convention lands.",
  "provider" => "Recurly",
  "provider_reference_url" => "https://recurly.com/developers/api/v2021-02-25/",
  "openapi_source_url" => OAS_URL,
  "openapi_sha256" => OAS_SHA256,
  "retry_policy" => "Unkeyed writes run once unless an action explicitly declares documented idempotent delete semantics. Recurly mutations send a stable per-record Idempotency-Key across automatic retries.",
  "operations" => delete_retry_evidence,
  "mutation_query_controls" => mutation_query_evidence
})

actions = writes_document.fetch("actions").to_h { |action| [action.fetch("name"), action] }
cli_path = File.join(BUNDLE, "cli_surface.json")
cli_document = read_json(cli_path)
cli_document.fetch("commands").each do |command|
  if command["write"]
    action = actions.fetch(command.fetch("write"))
    schema = action.fetch("record_schema")
    required_paths = required_mapping_paths(schema)
    removed_mappings = Array(REMOVED_GENERATED_CLI_MAPPINGS[action.fetch("name")])
    existing = Array(command["flags"]).reject { |flag| removed_mappings.include?(flag["maps_to"]) }.to_h do |flag|
      [flag["maps_to"], flag]
    end
    desired_paths = existing.keys.filter_map do |mapping|
      path = mapping&.delete_prefix("record.")
      path if mapping&.start_with?("record.") && schema_at_path(schema, path) && flag_type(schema_at_path(schema, path))
    end
    desired_paths.concat(required_paths)
    desired_paths.concat(Array(action["path_fields"]))
    desired_paths.concat(action.fetch("query", {}).keys)

    required_body_paths = required_paths - Array(action["path_fields"])
    if action["body_required"] && required_body_paths.empty?
      candidates = leaf_paths(schema).reject { |path, _| Array(action["path_fields"]).include?(path) }
      preferred = %w[email first_name plan_code amount code name currency]
      selected = candidates.find { |path, _| !path.include?(".") && preferred.include?(path) }
      selected ||= candidates.find { |path, _| !path.include?(".") }
      selected ||= candidates.find { |path, _| preferred.include?(path.split(".").last) }
      selected ||= candidates.first
      abort "required-body action #{action.fetch('name')} has no CLI-compatible body field" unless selected
      desired_paths << selected.first
    end

    used_names = {}
    command["flags"] = desired_paths.uniq.filter_map do |path|
      node = schema_at_path(schema, path)
      type = node && flag_type(node)
      next unless type
      mapping = "record.#{path}"
      old = existing[mapping] || {}
      name = old["name"] || flag_name(path)
      if used_names[name] && used_names[name] != path
        name = "#{name}-#{Digest::SHA256.hexdigest(path)[0, 6]}"
      end
      used_names[name] = path
      flag = {
        "name" => name,
        "type" => type,
        "summary" => old["summary"] || node["description"] || "Typed Recurly request field `#{path}`.",
        "maps_to" => mapping
      }
      values = Array(node["enum"]).compact
      flag["values"] = values if type == "enum"
      flag["format"] = "date-time" if node["format"] == "date-time"
      required_mapping = required_paths.any? { |required| path == required || path.start_with?("#{required}.") }
      flag["required"] = true if required_mapping || Array(action["path_fields"]).include?(path)
      flag["max_items"] = node["maxItems"] if type == "string_array" && node["maxItems"]
      flag["min_items"] = node["minItems"] if type == "string_array" && node["minItems"]
      flag
    end
    required_flag_names = command.fetch("flags").select { |flag| flag["required"] }.map { |flag| flag.fetch("name") }
    if action["body_required"]
      path_names = Array(action["path_fields"])
      body_flags = command.fetch("flags").select do |flag|
        mapped = flag.fetch("maps_to").delete_prefix("record.")
        !path_names.include?(mapped)
      end
      body_flag = body_flags.find { |flag| !flag.fetch("maps_to").delete_prefix("record.").include?(".") } || body_flags.first
      required_flag_names << body_flag.fetch("name") if body_flag && !required_flag_names.include?(body_flag.fetch("name"))
    end
    invocation = "pm recurly #{command.fetch('path')}"
    query_flag_names = command.fetch("flags").filter_map do |flag|
      field = flag.fetch("maps_to").delete_prefix("record.")
      flag.fetch("name") if action.fetch("query", {}).key?(field)
    end
    (required_flag_names + query_flag_names).uniq.each { |name| invocation += " --#{name} \"<value>\"" }
    invocation += " --json"
    command["examples"] = [invocation]
    command["risk"] = action["risk"] if command.key?("risk")
  elsif command["operation"] == "preview_gift_card"
    flag = command.fetch("flags").find { |candidate| candidate["maps_to"] == "body.unit_amount" } || abort("missing gift-card unit_amount flag")
    flag["type"] = "number"
    flag["summary"] = "Decimal gift-card amount to preview; mapped to Recurly unit_amount."
  end
end
write_json(cli_path, cli_document)

streams_path = File.join(BUNDLE, "streams.json")
streams_document = read_json(streams_path)
streams_document.fetch("base").delete("rate_limit")

streams_document.fetch("streams").each do |stream|
  next if LEGACY_STREAMS.include?(stream.fetch("name"))

  endpoint = stream_endpoints[stream.fetch("name")] || abort("missing surface endpoint for stream #{stream.fetch('name')}")
  operation = operation_for(endpoint)
  _status, response_schema_ref = response_for(operation)
  abort "stream #{stream.fetch('name')} lacks JSON response schema" unless response_schema_ref
  response_schema = normalize_schema(response_schema_ref, :response)
  data_property = response_schema.fetch("properties", {})["data"]
  list = data_property && array_schema?(data_property)
  item_schema = list ? data_property.fetch("items") : response_schema
  abort "stream #{stream.fetch('name')} record schema is not an object" unless object_schema?(item_schema)
  item_schema = deep_copy(item_schema)
  item_schema["additionalProperties"] = false

  computed = {}
  source_key_path = nil
  primary_key = case stream.fetch("name")
                when "list_entitlements"
                  computed["customer_permission_id"] = "{{ record.customer_permission.id }}"
                  source_key_path = "customer_permission.id"
                  item_schema.fetch("properties")["customer_permission_id"] = {"type" => "string"}
                  "customer_permission_id"
                when "get_account_balance"
                  computed["account_id"] = "{{ record.account.id }}"
                  source_key_path = "account.id"
                  item_schema.fetch("properties")["account_id"] = {"type" => "string"}
                  "account_id"
                else
                  abort "stream #{stream.fetch('name')} response has no durable id" unless item_schema.fetch("properties", {}).key?("id")
                  "id"
                end
  require_schema_path(item_schema, source_key_path) if source_key_path
  item_schema["required"] = (Array(item_schema["required"]) + [primary_key]).uniq
  schema_document = {
    "$schema" => "http://json-schema.org/draft-07/schema#",
    "title" => "Recurly #{stream.fetch('name')} record",
    "type" => "object",
    "x-primary-key" => [primary_key],
    "properties" => item_schema.fetch("properties"),
    "required" => item_schema.fetch("required"),
    "additionalProperties" => false
  }
  schema_path = File.join(BUNDLE, stream.fetch("schema"))
  write_json(schema_path, schema_document)

  stream["records"] = list ? {"path" => "data"} : {"path" => ".", "single_object" => true}
  stream["projection"] = "schema"
  if list
    stream.delete("pagination")
  else
    stream["pagination"] = {"type" => "none"}
  end
  if computed.empty?
    stream.delete("computed_fields")
  else
    stream["computed_fields"] = computed
  end

  fixture_path = File.join(BUNDLE, "fixtures", "streams", stream.fetch("name"), "page_1.json")
  fixture = read_json(fixture_path)
  record = sample_value(item_schema, stream.fetch("name"), include_optional: true)
  if source_key_path
    source_schema = schema_at_path(item_schema, source_key_path) || abort("missing derived source #{source_key_path}")
    set_nested(record, source_key_path, sample_value(source_schema, source_key_path.split(".").last))
    record.delete(primary_key)
  end
  body = list ? {"data" => [record]} : record
  fixture["response"] = {
    "status" => 200,
    "headers" => {"Content-Type" => ["application/json"]},
    "body" => body
  }

  if stream.fetch("name") == "list_sites"
    fixture.fetch("response").fetch("headers")["Link"] = ["<{{ base_url }}/sites?cursor=fixture_next>; rel=\"next\""]
    page_two = deep_copy(fixture)
    page_two.fetch("request")["query"] = {"cursor" => "fixture_next"}
    second_record = deep_copy(record)
    second_record["id"] = sample_string("id_2", item_schema.fetch("properties").fetch("id"))
    page_two["response"] = {
      "status" => 200,
      "headers" => {"Content-Type" => ["application/json"]},
      "body" => {"data" => [second_record]}
    }
    write_json(File.join(BUNDLE, "fixtures", "streams", "list_sites", "page_2.json"), page_two)
  end
  write_json(fixture_path, fixture)
end
write_json(streams_path, streams_document)
