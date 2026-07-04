# Overview

PokeAPI is a public, read-only REST API for Pokemon reference data. This bundle covers the
documented PokeAPI v2 REST surface from `https://pokeapi.co/docs/v2`: every concrete resource
list endpoint implied by the generic list/pagination rule, every documented resource detail
endpoint, and the Pokemon location-area subresource.

The four legacy streams (`pokemon`, `types`, `abilities`, and `moves`) keep the legacy record
shape exactly: `name` and `url` pass through from `results[]`, and `id` is computed from the
final path segment of `url` with `{{ record.url | last_path_segment }}`.

## Auth setup

No credentials are required. PokeAPI documents that only HTTP GET is available and that resources
are public. `base_url` defaults to `https://pokeapi.co/api/v2` and may be overridden for tests or
proxies.

## Streams notes

Resource-list streams use PokeAPI's `{count, next, previous, results[]}` envelope with
`offset_limit` pagination (`limit`/`offset`, page size 100, `max_pages: 3`). Named list
resources use `name` as the primary key, matching legacy. The documented unnamed list resources
(`characteristic`, `contest-effect`, `evolution-chain`, `machine`, and `super-contest-effect`)
use the URL-derived `id` as primary key because their list records contain only `url`.

List streams: pokemon, types, abilities, moves, berries, berry_firmnesses, berry_flavors, contest_types, contest_effects, super_contest_effects, encounter_methods, encounter_conditions, encounter_condition_values, evolution_chains, evolution_triggers, generations, pokedexes, versions, version_groups, items, item_attributes, item_categories, item_fling_effects, item_pockets, locations, location_areas, pal_park_areas, regions, machines, move_ailments, move_battle_styles, move_categories, move_damage_classes, move_learn_methods, move_targets, characteristics, egg_groups, genders, growth_rates, natures, pokeathlon_stats, pokemon_colors, pokemon_forms, pokemon_habitats, pokemon_shapes, pokemon_species, stats, languages.

Detail streams are config-backed single-resource reads named `<list_stream>_detail`. Configure
`<resource>_id` (for example `pokemon_id`, `item_id`, or `version_group_id`) with either an ID or
a name where the PokeAPI docs allow `{id or name}`. Detail streams use passthrough projection so
the full resource object returned by PokeAPI is retained while schemas pin the stable `id`/`name`
identifiers. `pokemon_location_areas` reads `GET /pokemon/{id or name}/encounters` and computes
`id` from the returned `location_area.url`.

## Write actions & risks

None. PokeAPI documents the REST API as consumption-only with only HTTP GET available, so
`capabilities.write` is false and this bundle intentionally has no `writes.json`.

## Known limits

Detail and `pokemon_location_areas` streams require the matching config ID/name because PokeAPI
does not expose a bulk detail endpoint for those full resource objects. The list streams keep
legacy's default page size and page cap (100 records per page, 3 pages) as static pagination
settings because the engine's `offset_limit` pagination does not support runtime-configurable
`page_size` or `max_pages`.
