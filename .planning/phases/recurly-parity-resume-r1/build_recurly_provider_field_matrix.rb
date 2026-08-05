#!/usr/bin/env ruby
# frozen_string_literal: true

# Generates raw provider evidence for the Recurly v2021-02-25 API. This is a
# research artifact, deliberately not connector citation metadata: the shared
# citation convention is being authored on a separate lane.

require "digest"
require "json"
require "open-uri"
require "pathname"
require "yaml"

SOURCE_URL = "https://recurly.com/developers/api/spec/v2021-02-25.yaml"
REFERENCE_URL = "https://recurly.com/developers/api/v2021-02-25/"
ROOT = Pathname.new(__dir__).join("../../..").realpath
BUNDLE_DIR = ROOT.join("internal/connectors/defs/recurly")
OUTPUT = Pathname.new(__dir__).join("RECURLY-PROVIDER-FIELD-RESEARCH.json")

def resolve_ref(document, reference)
  return nil unless reference.is_a?(String) && reference.start_with?("#/")

  reference.delete_prefix("#/").split("/").reduce(document) do |node, part|
    node.is_a?(Hash) ? node[part.gsub("~1", "/").gsub("~0", "~")] : nil
  end
end

def dereference(document, node)
  refs = []
  current = node
  while current.is_a?(Hash) && current["$ref"]
    reference = current.fetch("$ref")
    break if refs.include?(reference)

    refs << reference
    current = resolve_ref(document, reference)
  end
  [current || node, refs]
end

def schema_summary(schema)
  return {} unless schema.is_a?(Hash)

  %w[type format nullable enum description readOnly writeOnly deprecated maxLength pattern].each_with_object({}) do |key, out|
    out[key] = schema[key] if schema.key?(key)
  end
end

def required_rationale(location, required)
  return "OpenAPI requires all path parameters." if location == "path"
  if location == "body"
    return "OpenAPI request schema's required array lists this field." if required

    return "OpenAPI request schema does not list this field as required."
  end
  return "OpenAPI parameter has required: true." if required

  "OpenAPI does not mark this field required."
end

def field_row(operation_id:, method:, route:, source_path:, location:, name:, required:, schema:, refs:, content_type: nil, recursive: false)
  field = location == "body" ? "body.#{name}" : "#{location}.#{name}"
  field = field.sub(/\Abody\.$/, "body")
  {
    "operation_id" => operation_id,
    "method" => method,
    "path" => route,
    "field" => field,
    "location" => location,
    "content_type" => content_type,
    "required" => required,
    "requiredness_rationale" => required_rationale(location, required),
    "source_url" => "#{REFERENCE_URL}#operation/#{operation_id}",
    "evidence_type" => location == "body" ? "openapi.requestBody" : "openapi.parameter",
    "evidence_path" => source_path,
    "schema" => schema_summary(schema),
    "schema_refs" => refs,
    "confidence" => "high",
    "recursive_schema_reference" => recursive
  }.compact
end

def flatten_schema(document, schema, prefix:, required:, operation_id:, method:, route:, source_path:, content_type:, refs: [], seen_refs: [], inherited_required_names: [])
  resolved, new_refs = dereference(document, schema)
  schema_refs = refs + new_refs
  return [] unless resolved.is_a?(Hash)

  # Preserve a finite record for an intentionally recursive schema instead of
  # silently omitting it from research coverage.
  if new_refs.any? { |reference| seen_refs.include?(reference) }
    return [field_row(
      operation_id: operation_id,
      method: method,
      route: route,
      source_path: source_path,
      location: "body",
      name: prefix,
      required: required,
      schema: resolved,
      refs: schema_refs,
      content_type: content_type,
      recursive: true
    )]
  end

  rows = []
  # `allOf` combines object constraints. A required name can therefore live in
  # one member while its property is declared in a sibling member (CouponCreate
  # is one such Recurly shape), so apply those names across the composed object.
  all_of_required_names = Array(resolved["allOf"]).flat_map do |member|
    member_resolved, = dereference(document, member)
    member_resolved.is_a?(Hash) ? Array(member_resolved["required"]) : []
  end
  required_names = (Array(resolved["required"]) + inherited_required_names + all_of_required_names).uniq
  composites = %w[allOf anyOf oneOf].flat_map { |key| Array(resolved[key]) }
  unless composites.empty?
    composites.each_with_index do |member, index|
      rows.concat(flatten_schema(
        document,
        member,
        prefix: prefix,
        required: required,
        operation_id: operation_id,
        method: method,
        route: route,
        source_path: "#{source_path}.#{index}",
        content_type: content_type,
        refs: schema_refs,
        seen_refs: seen_refs + new_refs,
        inherited_required_names: required_names
      ))
    end
  end

  properties = resolved["properties"]
  if properties.is_a?(Hash)
    properties.each do |property, child_schema|
      child_prefix = prefix.empty? ? property : "#{prefix}.#{property}"
      child_resolved, child_refs = dereference(document, child_schema)
      child_required = required_names.include?(property)
      rows << field_row(
        operation_id: operation_id,
        method: method,
        route: route,
        source_path: "#{source_path}.properties.#{property}",
        location: "body",
        name: child_prefix,
        required: child_required,
        schema: child_resolved,
        refs: schema_refs + child_refs,
        content_type: content_type
      )
      rows.concat(flatten_schema(
        document,
        child_schema,
        prefix: child_prefix,
        required: child_required,
        operation_id: operation_id,
        method: method,
        route: route,
        source_path: "#{source_path}.properties.#{property}",
        content_type: content_type,
        refs: schema_refs,
        seen_refs: seen_refs + new_refs
      ))
    end
  elsif resolved["items"]
    rows.concat(flatten_schema(
      document,
      resolved["items"],
      prefix: "#{prefix}[]",
      required: required,
      operation_id: operation_id,
      method: method,
      route: route,
      source_path: "#{source_path}.items",
      content_type: content_type,
      refs: schema_refs,
      seen_refs: seen_refs + new_refs
    ))
  end
  rows
end

def bundle_coverage(bundle_dir, method, route)
  surface = JSON.parse(bundle_dir.join("api_surface.json").read)
  cli = JSON.parse(bundle_dir.join("cli_surface.json").read)
  endpoint = surface.fetch("endpoints").find do |candidate|
    candidate["method"].to_s.upcase == method && candidate["path"] == route
  end
  return { "present" => false } unless endpoint

  covered_by = endpoint["covered_by"] || {}
  command_paths = Array(covered_by["direct_reads"]) + [covered_by["direct_read"]].compact
  commands = cli.fetch("commands").select { |command| command_paths.include?(command["path"]) }
  {
    "present" => true,
    "covered_by" => covered_by,
    "blocked_operation" => endpoint["operation"],
    "commands" => commands.map { |command| { "path" => command["path"], "intent" => command["intent"], "availability" => command["availability"], "operation" => command["operation"] } }
  }.compact
end

source = URI.open(SOURCE_URL, "User-Agent" => "Polymetrics Recurly connector field research").read
document = YAML.safe_load(source, aliases: true)
operations = []

document.fetch("paths").each do |route, path_item|
  path_item.each do |verb, operation|
    next unless operation.is_a?(Hash) && operation["operationId"]

    method = verb.upcase
    operation_id = operation.fetch("operationId")
    inputs = []
    (Array(path_item["parameters"]) + Array(operation["parameters"])).each_with_index do |parameter, index|
      resolved, refs = dereference(document, parameter)
      next unless resolved.is_a?(Hash)

      location = resolved.fetch("in")
      name = resolved.fetch("name")
      required = location == "path" || resolved["required"] == true
      inputs << field_row(
        operation_id: operation_id,
        method: method,
        route: route,
        source_path: "paths.#{route}.#{verb}.parameters.#{index}",
        location: location,
        name: name,
        required: required,
        schema: resolved["schema"] || {},
        refs: refs
      )
    end

    request_body, request_refs = dereference(document, operation["requestBody"])
    if request_body.is_a?(Hash)
      request_body.fetch("content", {}).each do |content_type, media_type|
        schema = media_type.is_a?(Hash) ? media_type["schema"] : nil
        next unless schema

        inputs.concat(flatten_schema(
          document,
          schema,
          prefix: "",
          required: request_body["required"] == true,
          operation_id: operation_id,
          method: method,
          route: route,
          source_path: "paths.#{route}.#{verb}.requestBody.content.#{content_type}.schema",
          content_type: content_type,
          refs: request_refs
        ))
      end
    end

    unique_inputs = inputs.uniq { |entry| [entry["field"], entry["content_type"], entry["evidence_path"]] }
    operations << {
      "operation_id" => operation_id,
      "method" => method,
      "path" => route,
      "source_url" => "#{REFERENCE_URL}#operation/#{operation_id}",
      "summary" => operation["summary"],
      "provider_input_count" => unique_inputs.length,
      "bundle_coverage" => bundle_coverage(BUNDLE_DIR, method, route),
      "fields" => unique_inputs
    }.compact
  end
end

matrix = {
  "purpose" => "Raw provider evidence only; do not treat this as final citation metadata until the shared citation convention is merged.",
  "provider" => "Recurly",
  "provider_reference_url" => REFERENCE_URL,
  "openapi_source_url" => SOURCE_URL,
  "openapi_sha256" => Digest::SHA256.hexdigest(source),
  "openapi_version" => document["openapi"],
  "documented_operation_count" => operations.length,
  "method_counts" => operations.group_by { |operation| operation["method"] }.transform_values(&:length),
  "operations" => operations
}

OUTPUT.write(JSON.pretty_generate(matrix) + "\n")
puts "wrote #{OUTPUT} (#{operations.length} operations, #{operations.sum { |operation| operation["provider_input_count"] }} raw provider input rows)"
