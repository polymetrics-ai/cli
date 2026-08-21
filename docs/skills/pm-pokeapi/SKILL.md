---
name: pm-pokeapi
description: PokeAPI connector knowledge and safe action guide.
---

# pm-pokeapi

## Purpose

Reads the documented public PokeAPI v2 resource catalog, including list and detail endpoints.

## Icon

- id: pokeapi
- asset: icons/pokeapi.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://pokeapi.co/docs/v2

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- No secret authentication is required for this connector.

## Configuration

- ability_id
- base_url
- berry_firmness_id
- berry_flavor_id
- berry_id
- characteristic_id
- contest_effect_id
- contest_type_id
- egg_group_id
- encounter_condition_id
- encounter_condition_value_id
- encounter_method_id
- evolution_chain_id
- evolution_trigger_id
- gender_id
- generation_id
- growth_rate_id
- item_attribute_id
- item_category_id
- item_fling_effect_id
- item_id
- item_pocket_id
- language_id
- location_area_id
- location_id
- machine_id
- mode
- move_ailment_id
- move_battle_style_id
- move_category_id
- move_damage_class_id
- move_id
- move_learn_method_id
- move_target_id
- nature_id
- pal_park_area_id
- pokeathlon_stat_id
- pokedex_id
- pokemon_color_id
- pokemon_form_id
- pokemon_habitat_id
- pokemon_id
- pokemon_shape_id
- pokemon_species_id
- region_id
- stat_id
- super_contest_effect_id
- type_id
- version_group_id
- version_id

## ETL Streams

- pokemon:
  - primary key: name
  - fields: id(string), name(string), url(string)
- types:
  - primary key: name
  - fields: id(string), name(string), url(string)
- abilities:
  - primary key: name
  - fields: id(string), name(string), url(string)
- moves:
  - primary key: name
  - fields: id(string), name(string), url(string)
- berries:
  - primary key: name
  - fields: id(string), name(string), url(string)
- berry_firmnesses:
  - primary key: name
  - fields: id(string), name(string), url(string)
- berry_flavors:
  - primary key: name
  - fields: id(string), name(string), url(string)
- contest_types:
  - primary key: name
  - fields: id(string), name(string), url(string)
- contest_effects:
  - primary key: id
  - fields: id(string), url(string)
- super_contest_effects:
  - primary key: id
  - fields: id(string), url(string)
- encounter_methods:
  - primary key: name
  - fields: id(string), name(string), url(string)
- encounter_conditions:
  - primary key: name
  - fields: id(string), name(string), url(string)
- encounter_condition_values:
  - primary key: name
  - fields: id(string), name(string), url(string)
- evolution_chains:
  - primary key: id
  - fields: id(string), url(string)
- evolution_triggers:
  - primary key: name
  - fields: id(string), name(string), url(string)
- generations:
  - primary key: name
  - fields: id(string), name(string), url(string)
- pokedexes:
  - primary key: name
  - fields: id(string), name(string), url(string)
- versions:
  - primary key: name
  - fields: id(string), name(string), url(string)
- version_groups:
  - primary key: name
  - fields: id(string), name(string), url(string)
- items:
  - primary key: name
  - fields: id(string), name(string), url(string)
- item_attributes:
  - primary key: name
  - fields: id(string), name(string), url(string)
- item_categories:
  - primary key: name
  - fields: id(string), name(string), url(string)
- item_fling_effects:
  - primary key: name
  - fields: id(string), name(string), url(string)
- item_pockets:
  - primary key: name
  - fields: id(string), name(string), url(string)
- locations:
  - primary key: name
  - fields: id(string), name(string), url(string)
- location_areas:
  - primary key: name
  - fields: id(string), name(string), url(string)
- pal_park_areas:
  - primary key: name
  - fields: id(string), name(string), url(string)
- regions:
  - primary key: name
  - fields: id(string), name(string), url(string)
- machines:
  - primary key: id
  - fields: id(string), url(string)
- move_ailments:
  - primary key: name
  - fields: id(string), name(string), url(string)
- move_battle_styles:
  - primary key: name
  - fields: id(string), name(string), url(string)
- move_categories:
  - primary key: name
  - fields: id(string), name(string), url(string)
- move_damage_classes:
  - primary key: name
  - fields: id(string), name(string), url(string)
- move_learn_methods:
  - primary key: name
  - fields: id(string), name(string), url(string)
- move_targets:
  - primary key: name
  - fields: id(string), name(string), url(string)
- characteristics:
  - primary key: id
  - fields: id(string), url(string)
- egg_groups:
  - primary key: name
  - fields: id(string), name(string), url(string)
- genders:
  - primary key: name
  - fields: id(string), name(string), url(string)
- growth_rates:
  - primary key: name
  - fields: id(string), name(string), url(string)
- natures:
  - primary key: name
  - fields: id(string), name(string), url(string)
- pokeathlon_stats:
  - primary key: name
  - fields: id(string), name(string), url(string)
- pokemon_colors:
  - primary key: name
  - fields: id(string), name(string), url(string)
- pokemon_forms:
  - primary key: name
  - fields: id(string), name(string), url(string)
- pokemon_habitats:
  - primary key: name
  - fields: id(string), name(string), url(string)
- pokemon_shapes:
  - primary key: name
  - fields: id(string), name(string), url(string)
- pokemon_species:
  - primary key: name
  - fields: id(string), name(string), url(string)
- stats:
  - primary key: name
  - fields: id(string), name(string), url(string)
- languages:
  - primary key: name
  - fields: id(string), name(string), url(string)
- pokemon_detail:
  - primary key: id
  - fields: id(integer), name(string)
- types_detail:
  - primary key: id
  - fields: id(integer), name(string)
- abilities_detail:
  - primary key: id
  - fields: id(integer), name(string)
- moves_detail:
  - primary key: id
  - fields: id(integer), name(string)
- berries_detail:
  - primary key: id
  - fields: id(integer), name(string)
- berry_firmnesses_detail:
  - primary key: id
  - fields: id(integer), name(string)
- berry_flavors_detail:
  - primary key: id
  - fields: id(integer), name(string)
- contest_types_detail:
  - primary key: id
  - fields: id(integer), name(string)
- contest_effects_detail:
  - primary key: id
  - fields: id(integer)
- super_contest_effects_detail:
  - primary key: id
  - fields: id(integer)
- encounter_methods_detail:
  - primary key: id
  - fields: id(integer), name(string)
- encounter_conditions_detail:
  - primary key: id
  - fields: id(integer), name(string)
- encounter_condition_values_detail:
  - primary key: id
  - fields: id(integer), name(string)
- evolution_chains_detail:
  - primary key: id
  - fields: id(integer)
- evolution_triggers_detail:
  - primary key: id
  - fields: id(integer), name(string)
- generations_detail:
  - primary key: id
  - fields: id(integer), name(string)
- pokedexes_detail:
  - primary key: id
  - fields: id(integer), name(string)
- versions_detail:
  - primary key: id
  - fields: id(integer), name(string)
- version_groups_detail:
  - primary key: id
  - fields: id(integer), name(string)
- items_detail:
  - primary key: id
  - fields: id(integer), name(string)
- item_attributes_detail:
  - primary key: id
  - fields: id(integer), name(string)
- item_categories_detail:
  - primary key: id
  - fields: id(integer), name(string)
- item_fling_effects_detail:
  - primary key: id
  - fields: id(integer), name(string)
- item_pockets_detail:
  - primary key: id
  - fields: id(integer), name(string)
- locations_detail:
  - primary key: id
  - fields: id(integer), name(string)
- location_areas_detail:
  - primary key: id
  - fields: id(integer), name(string)
- pal_park_areas_detail:
  - primary key: id
  - fields: id(integer), name(string)
- regions_detail:
  - primary key: id
  - fields: id(integer), name(string)
- machines_detail:
  - primary key: id
  - fields: id(integer)
- move_ailments_detail:
  - primary key: id
  - fields: id(integer), name(string)
- move_battle_styles_detail:
  - primary key: id
  - fields: id(integer), name(string)
- move_categories_detail:
  - primary key: id
  - fields: id(integer), name(string)
- move_damage_classes_detail:
  - primary key: id
  - fields: id(integer), name(string)
- move_learn_methods_detail:
  - primary key: id
  - fields: id(integer), name(string)
- move_targets_detail:
  - primary key: id
  - fields: id(integer), name(string)
- characteristics_detail:
  - primary key: id
  - fields: id(integer)
- egg_groups_detail:
  - primary key: id
  - fields: id(integer), name(string)
- genders_detail:
  - primary key: id
  - fields: id(integer), name(string)
- growth_rates_detail:
  - primary key: id
  - fields: id(integer), name(string)
- natures_detail:
  - primary key: id
  - fields: id(integer), name(string)
- pokeathlon_stats_detail:
  - primary key: id
  - fields: id(integer), name(string)
- pokemon_colors_detail:
  - primary key: id
  - fields: id(integer), name(string)
- pokemon_forms_detail:
  - primary key: id
  - fields: id(integer), name(string)
- pokemon_habitats_detail:
  - primary key: id
  - fields: id(integer), name(string)
- pokemon_shapes_detail:
  - primary key: id
  - fields: id(integer), name(string)
- pokemon_species_detail:
  - primary key: id
  - fields: id(integer), name(string)
- stats_detail:
  - primary key: id
  - fields: id(integer), name(string)
- languages_detail:
  - primary key: id
  - fields: id(integer), name(string)
- pokemon_location_areas:
  - primary key: id
  - fields: id(string), location_area(object), version_details(array)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Security

- read risk: external PokeAPI read of public Pokemon reference data
- approval: none; read-only public reference API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect pokeapi
```

### Inspect as structured JSON

```bash
pm connectors inspect pokeapi --json
```

## Agent Rules

- Run pm connectors inspect pokeapi before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
